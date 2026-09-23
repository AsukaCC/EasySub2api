package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/lib/pq"
)

func scanUserLevelRule(ctx context.Context, q sqlQueryer, ruleID string) (*service.UserLevelRule, error) {
	var rule service.UserLevelRule
	err := scanSingleRow(ctx, q, `
		SELECT id::text, name, window_days, enabled, created_at, updated_at
		FROM user_level_rules
		WHERE id = $1::uuid
	`, []any{ruleID}, &rule.ID, &rule.Name, &rule.WindowDays, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrUserLevelRuleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *userLevelRepository) SetDefaultLevelRule(ctx context.Context, ruleID string) (err error) {
	defer func() { err = wrapUserLevelRuleDeleteError(err) }()
	tx, err := r.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(276, 1)"); err != nil {
		return err
	}
	var enabled bool
	if err := tx.QueryRowContext(ctx, `SELECT enabled FROM user_level_rules WHERE id = $1::uuid`, ruleID).Scan(&enabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.ErrUserLevelRuleNotFound
		}
		return err
	}
	if !enabled {
		return service.ErrUserLevelRuleDisabled
	}
	if _, err := tx.ExecContext(ctx, `UPDATE settings SET value = jsonb_build_array($1::text)::text, updated_at = NOW() WHERE key = 'default_user_level_rule_ids'`, ruleID); err != nil {
		return err
	}
	return tx.Commit()
}

func loadUserLevelTiers(ctx context.Context, q sqlQueryer, ruleIDs []string) (map[string][]service.UserLevelTier, error) {
	out := make(map[string][]service.UserLevelTier, len(ruleIDs))
	if len(ruleIDs) == 0 {
		return out, nil
	}
	rows, err := q.QueryContext(ctx, `
		SELECT id::text, rule_id::text, name, sort_order, min_spend::double precision,
		       default_multiplier::double precision, created_at, updated_at
		FROM user_level_rule_tiers
		WHERE rule_id = ANY($1::uuid[])
		ORDER BY rule_id, sort_order
	`, pq.Array(ruleIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var tier service.UserLevelTier
		var multiplier sql.NullFloat64
		if err := rows.Scan(&tier.ID, &tier.RuleID, &tier.Name, &tier.SortOrder, &tier.MinSpend, &multiplier, &tier.CreatedAt, &tier.UpdatedAt); err != nil {
			return nil, err
		}
		if multiplier.Valid {
			value := multiplier.Float64
			tier.DefaultMultiplier = &value
		}
		out[tier.RuleID] = append(out[tier.RuleID], tier)
	}
	return out, rows.Err()
}

func attachUserLevelTiers(ctx context.Context, q sqlQueryer, rules []service.UserLevelRule) error {
	ids := make([]string, 0, len(rules))
	for i := range rules {
		ids = append(ids, rules[i].ID)
	}
	tiers, err := loadUserLevelTiers(ctx, q, ids)
	if err != nil {
		return err
	}
	for i := range rules {
		rules[i].Tiers = tiers[rules[i].ID]
	}
	return nil
}

func (r *userLevelRepository) ListLevelRules(ctx context.Context) ([]service.UserLevelRule, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id::text, name, window_days, enabled, created_at, updated_at
		FROM user_level_rules
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	rules := make([]service.UserLevelRule, 0)
	for rows.Next() {
		var rule service.UserLevelRule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.WindowDays, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := attachUserLevelTiers(ctx, r.sql, rules); err != nil {
		return nil, err
	}
	for i := range rules {
		if err := scanSingleRow(ctx, r.sql, `SELECT COUNT(*) FROM user_level_rule_assignments WHERE rule_id = $1::uuid`, []any{rules[i].ID}, &rules[i].AssignedUserCount); err != nil {
			return nil, err
		}
		if rules[i].ReferenceCount, err = r.GetLevelRuleReferenceCount(ctx, rules[i].ID); err != nil {
			return nil, err
		}
	}
	var defaultID string
	if err := scanSingleRow(ctx, r.sql, `SELECT required_default_user_level_rule()::text`, nil, &defaultID); err != nil {
		return nil, err
	}
	for i := range rules {
		rules[i].IsDefault = rules[i].ID == defaultID
	}
	return rules, nil
}

func (r *userLevelRepository) GetLevelRule(ctx context.Context, ruleID string) (*service.UserLevelRule, error) {
	rule, err := scanUserLevelRule(ctx, r.sql, ruleID)
	if err != nil {
		return nil, err
	}
	tiers, err := loadUserLevelTiers(ctx, r.sql, []string{rule.ID})
	if err != nil {
		return nil, err
	}
	rule.Tiers = tiers[rule.ID]
	if err := scanSingleRow(ctx, r.sql, `SELECT COUNT(*) FROM user_level_rule_assignments WHERE rule_id = $1::uuid`, []any{rule.ID}, &rule.AssignedUserCount); err != nil {
		return nil, err
	}
	rule.ReferenceCount, err = r.GetLevelRuleReferenceCount(ctx, rule.ID)
	if err != nil {
		return nil, err
	}
	var defaultID string
	if err := scanSingleRow(ctx, r.sql, `SELECT required_default_user_level_rule()::text`, nil, &defaultID); err != nil {
		return nil, err
	}
	rule.IsDefault = rule.ID == defaultID
	return rule, nil
}

func (r *userLevelRepository) CreateLevelRule(ctx context.Context, rule *service.UserLevelRule) error {
	if rule == nil {
		return fmt.Errorf("user level rule is nil")
	}
	tx, err := r.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if strings.TrimSpace(rule.ID) == "" {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO user_level_rules (name, window_days, enabled)
			VALUES ($1, $2, $3)
			RETURNING id::text, created_at, updated_at
		`, rule.Name, rule.WindowDays, rule.Enabled).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
	} else {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO user_level_rules (id, name, window_days, enabled)
			VALUES ($1::uuid, $2, $3, $4)
			RETURNING created_at, updated_at
		`, rule.ID, rule.Name, rule.WindowDays, rule.Enabled).Scan(&rule.CreatedAt, &rule.UpdatedAt)
	}
	if err != nil {
		return err
	}
	for i := range rule.Tiers {
		tier := &rule.Tiers[i]
		tier.RuleID = rule.ID
		if strings.TrimSpace(tier.ID) == "" {
			err = tx.QueryRowContext(ctx, `
				INSERT INTO user_level_rule_tiers (rule_id, name, sort_order, min_spend, default_multiplier)
				VALUES ($1::uuid, $2, $3, $4, $5)
				RETURNING id::text, created_at, updated_at
			`, rule.ID, tier.Name, tier.SortOrder, tier.MinSpend, tier.DefaultMultiplier).Scan(&tier.ID, &tier.CreatedAt, &tier.UpdatedAt)
		} else {
			err = tx.QueryRowContext(ctx, `
				INSERT INTO user_level_rule_tiers (id, rule_id, name, sort_order, min_spend, default_multiplier)
				VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)
				RETURNING created_at, updated_at
			`, tier.ID, rule.ID, tier.Name, tier.SortOrder, tier.MinSpend, tier.DefaultMultiplier).Scan(&tier.CreatedAt, &tier.UpdatedAt)
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *userLevelRepository) UpdateLevelRule(ctx context.Context, rule *service.UserLevelRule) (err error) {
	defer func() { err = wrapUserLevelRuleDeleteError(err) }()
	if rule == nil {
		return fmt.Errorf("user level rule is nil")
	}
	tx, err := r.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var existingID string
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(276, 1)"); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT id::text FROM user_level_rules WHERE id = $1::uuid FOR UPDATE`, rule.ID).Scan(&existingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.ErrUserLevelRuleNotFound
		}
		return err
	}
	oldRows, err := tx.QueryContext(ctx, `SELECT id::text FROM user_level_rule_tiers WHERE rule_id = $1::uuid`, rule.ID)
	if err != nil {
		return err
	}
	oldIDs := make([]string, 0)
	for oldRows.Next() {
		var id string
		if err := oldRows.Scan(&id); err != nil {
			_ = oldRows.Close()
			return err
		}
		oldIDs = append(oldIDs, id)
	}
	if err := oldRows.Close(); err != nil {
		return err
	}
	newIDs := make(map[string]struct{}, len(rule.Tiers))
	for _, tier := range rule.Tiers {
		newIDs[tier.ID] = struct{}{}
	}
	for _, oldID := range oldIDs {
		if _, keep := newIDs[oldID]; keep {
			continue
		}
		refs, err := countUserLevelTierReferences(ctx, tx, oldID)
		if err != nil {
			return err
		}
		if refs > 0 {
			return service.ErrUserLevelRuleReferenced
		}
	}
	if err := tx.QueryRowContext(ctx, `
		UPDATE user_level_rules
		SET name = $2, window_days = $3, enabled = $4, updated_at = NOW()
		WHERE id = $1::uuid
		RETURNING updated_at
	`, rule.ID, rule.Name, rule.WindowDays, rule.Enabled).Scan(&rule.UpdatedAt); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_level_rule_tiers WHERE rule_id = $1::uuid`, rule.ID); err != nil {
		return err
	}
	for i := range rule.Tiers {
		tier := &rule.Tiers[i]
		tier.RuleID = rule.ID
		if strings.TrimSpace(tier.ID) == "" {
			if err := tx.QueryRowContext(ctx, `
				INSERT INTO user_level_rule_tiers (rule_id, name, sort_order, min_spend, default_multiplier)
				VALUES ($1::uuid, $2, $3, $4, $5)
				RETURNING id::text, created_at, updated_at
			`, rule.ID, tier.Name, tier.SortOrder, tier.MinSpend, tier.DefaultMultiplier).Scan(&tier.ID, &tier.CreatedAt, &tier.UpdatedAt); err != nil {
				return err
			}
		} else if err := tx.QueryRowContext(ctx, `
			INSERT INTO user_level_rule_tiers (id, rule_id, name, sort_order, min_spend, default_multiplier)
			VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)
			RETURNING created_at, updated_at
		`, tier.ID, rule.ID, tier.Name, tier.SortOrder, tier.MinSpend, tier.DefaultMultiplier).Scan(&tier.CreatedAt, &tier.UpdatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *userLevelRepository) DeleteLevelRule(ctx context.Context, ruleID string) (err error) {
	defer func() { err = wrapUserLevelRuleDeleteError(err) }()
	tx, err := r.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var exists string
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(276, 1)"); err != nil {
		return err
	}
	var isDefault, hasMembers bool
	if err := tx.QueryRowContext(ctx, `SELECT $1::uuid = required_default_user_level_rule(), EXISTS(SELECT 1 FROM user_level_rule_assignments WHERE rule_id = $1::uuid)`, ruleID).Scan(&isDefault, &hasMembers); err != nil {
		return err
	}
	if isDefault {
		return service.ErrUserLevelDefaultProtected
	}
	if hasMembers {
		return service.ErrUserLevelRuleHasMembers
	}
	if err := tx.QueryRowContext(ctx, `SELECT id::text FROM user_level_rules WHERE id = $1::uuid FOR UPDATE`, ruleID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.ErrUserLevelRuleNotFound
		}
		return err
	}
	refs, err := countUserLevelRuleReferences(ctx, tx, ruleID)
	if err != nil {
		return err
	}
	if refs > 0 {
		return service.ErrUserLevelRuleReferenced
	}
	// Tiers and leftover assignments use ON DELETE RESTRICT, so children must
	// be removed first. The reference check above already refused deletion
	// when assignments or JSON consumers still point at this rule.
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_level_rule_assignments WHERE rule_id = $1::uuid`, ruleID); err != nil {
		return wrapUserLevelRuleDeleteError(err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_level_rule_tiers WHERE rule_id = $1::uuid`, ruleID); err != nil {
		return wrapUserLevelRuleDeleteError(err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_level_rules WHERE id = $1::uuid`, ruleID); err != nil {
		return wrapUserLevelRuleDeleteError(err)
	}
	return tx.Commit()
}

func wrapUserLevelRuleDeleteError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Message {
		case "USER_LEVEL_DEFAULT_INVALID":
			return service.ErrUserLevelRuleAssignmentInput
		case "USER_LEVEL_DEFAULT_PROTECTED":
			return service.ErrUserLevelDefaultProtected
		case "USER_LEVEL_RULE_HAS_MEMBERS":
			return service.ErrUserLevelRuleHasMembers
		case "USER_LEVEL_RULE_DISABLED":
			return service.ErrUserLevelRuleDisabled
		}
	}
	if errors.As(err, &pqErr) && pqErr.Code == "23503" {
		return service.ErrUserLevelRuleReferenced
	}
	return err
}

func (r *userLevelRepository) ListUserLevelRules(ctx context.Context, userID string) ([]service.UserLevelRule, error) {
	byUser, err := r.GetAssignedLevelRulesBatch(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	return byUser[userID], nil
}

func (r *userLevelRepository) GetAssignedLevelRulesBatch(ctx context.Context, userIDs []string) (map[string][]service.UserLevelRule, error) {
	out := make(map[string][]service.UserLevelRule, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	tx, err := r.sql.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `
		SELECT a.user_id::text, r.id::text, r.name, r.window_days, r.enabled, r.created_at, r.updated_at
		FROM user_level_rule_assignments a
	JOIN user_level_rules r ON r.id = a.rule_id
	WHERE a.user_id = ANY($1::uuid[])
	ORDER BY a.user_id, r.created_at DESC, r.id DESC
	`, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	ruleIDs := make([]string, 0)
	seen := make(map[string]struct{})
	for rows.Next() {
		var userID string
		var rule service.UserLevelRule
		if err := rows.Scan(&userID, &rule.ID, &rule.Name, &rule.WindowDays, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		out[userID] = append(out[userID], rule)
		if _, ok := seen[rule.ID]; !ok {
			seen[rule.ID] = struct{}{}
			ruleIDs = append(ruleIDs, rule.ID)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	tiers, err := loadUserLevelTiers(ctx, tx, ruleIDs)
	if err != nil {
		return nil, err
	}
	for userID, rules := range out {
		for i := range rules {
			rules[i].Tiers = tiers[rules[i].ID]
		}
		out[userID] = rules
	}
	return out, tx.Commit()
}

func (r *userLevelRepository) ReplaceUserLevelRules(ctx context.Context, userID string, ruleIDs []string) error {
	_, err := r.BatchAssignUserLevelRules(ctx, []string{userID}, ruleIDs, service.UserLevelRuleAssignmentReplace)
	return err
}

func (r *userLevelRepository) BatchAssignUserLevelRules(ctx context.Context, userIDs, ruleIDs []string, operation string) (affected int64, err error) {
	defer func() { err = wrapUserLevelRuleDeleteError(err) }()
	if len(ruleIDs) > 1 || len(userIDs) == 0 {
		return 0, service.ErrUserLevelRuleAssignmentInput
	}
	if operation != service.UserLevelRuleAssignmentAdd && operation != service.UserLevelRuleAssignmentReplace && operation != service.UserLevelRuleAssignmentRemove {
		return 0, service.ErrUserLevelRuleAssignmentInput
	}
	tx, err := r.sql.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(276, 1)"); err != nil {
		return 0, err
	}
	if err := validateAssignmentTargets(ctx, tx, userIDs, ruleIDs); err != nil {
		return 0, err
	}
	var target string
	if len(ruleIDs) == 0 || operation == service.UserLevelRuleAssignmentRemove {
		if err := tx.QueryRowContext(ctx, "SELECT required_default_user_level_rule()::text").Scan(&target); err != nil {
			return 0, err
		}
	} else {
		target = ruleIDs[0]
	}
	var result sql.Result
	if operation == service.UserLevelRuleAssignmentRemove && len(ruleIDs) == 1 {
		result, err = tx.ExecContext(ctx, `
			UPDATE user_level_rule_assignments SET rule_id = $3::uuid
			WHERE user_id = ANY($1::uuid[]) AND rule_id = $2::uuid AND rule_id <> $3::uuid
		`, pq.Array(userIDs), ruleIDs[0], target)
	} else {
		result, err = tx.ExecContext(ctx, `
			INSERT INTO user_level_rule_assignments (user_id, rule_id)
			SELECT u, $2::uuid FROM unnest($1::uuid[]) AS users(u)
			ON CONFLICT (user_id) DO UPDATE SET rule_id = EXCLUDED.rule_id
			WHERE user_level_rule_assignments.rule_id <> EXCLUDED.rule_id
		`, pq.Array(userIDs), target)
	}
	if err != nil {
		return 0, err
	}
	affected, err = result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, tx.Commit()
}

func validateAssignmentTargets(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, userIDs, ruleIDs []string) error {
	var userCount int
	if err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE id = ANY($1::uuid[]) AND deleted_at IS NULL`, pq.Array(userIDs)).Scan(&userCount); err != nil {
		return err
	}
	if userCount != len(userIDs) {
		return service.ErrUserNotFound
	}
	if len(ruleIDs) == 0 {
		return nil
	}
	var ruleCount int
	if err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_level_rules WHERE id = ANY($1::uuid[])`, pq.Array(ruleIDs)).Scan(&ruleCount); err != nil {
		return err
	}
	if ruleCount != len(ruleIDs) {
		return service.ErrUserLevelRuleNotFound
	}
	return nil
}

func (r *userLevelRepository) ListLevelRuleMembers(ctx context.Context, ruleID string, page, pageSize int, search ...string) ([]service.User, int64, error) {
	query := ""
	if len(search) > 0 {
		query = strings.TrimSpace(search[0])
	}
	var total int64
	if err := scanSingleRow(ctx, r.sql, `
		SELECT COUNT(*)
		FROM user_level_rule_assignments a
		JOIN users u ON u.id = a.user_id
		WHERE a.rule_id = $1::uuid AND u.deleted_at IS NULL
		AND ($2 = '' OR STRPOS(LOWER(u.email), LOWER($2)) > 0 OR STRPOS(LOWER(COALESCE(u.username, '')), LOWER($2)) > 0)
	`, []any{ruleID, query}, &total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.sql.QueryContext(ctx, `
		SELECT u.id::text, u.email, COALESCE(u.username, ''), COALESCE(u.notes, ''),
		       u.role, u.recharge_balance::double precision, u.status, u.concurrency,
		       u.created_at, u.updated_at, u.deleted_at
		FROM user_level_rule_assignments a
		JOIN users u ON u.id = a.user_id
		WHERE a.rule_id = $1::uuid AND u.deleted_at IS NULL
		AND ($4 = '' OR STRPOS(LOWER(u.email), LOWER($4)) > 0 OR STRPOS(LOWER(COALESCE(u.username, '')), LOWER($4)) > 0)
		ORDER BY u.created_at DESC, u.id DESC
		LIMIT $2 OFFSET $3
	`, ruleID, pageSize, offset, query)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	users := make([]service.User, 0)
	for rows.Next() {
		var user service.User
		var deletedAt sql.NullTime
		if err := rows.Scan(&user.ID, &user.Email, &user.Username, &user.Notes, &user.Role, &user.RechargeBalance, &user.Status, &user.Concurrency, &user.CreatedAt, &user.UpdatedAt, &deletedAt); err != nil {
			return nil, 0, err
		}
		if deletedAt.Valid {
			value := deletedAt.Time
			user.DeletedAt = &value
		}
		users = append(users, user)
	}
	return users, total, rows.Err()
}

func (r *userLevelRepository) GetLevelRuleReferenceCount(ctx context.Context, ruleID string) (int64, error) {
	return countUserLevelRuleReferences(ctx, r.sql, ruleID)
}

func countUserLevelRuleReferences(ctx context.Context, q sqlQueryer, ruleID string) (int64, error) {
	var assignments int64
	if err := scanSingleRow(ctx, q, `SELECT COUNT(*) FROM user_level_rule_assignments WHERE rule_id = $1::uuid`, []any{ruleID}, &assignments); err != nil {
		return 0, err
	}
	// Drain the tier IDs before issuing more queries. When q is a *sql.Tx the
	// connection is shared, and nesting a query inside an open rows cursor
	// desyncs the lib/pq protocol ("unexpected Parse response 'D'").
	tierIDs, err := listUserLevelTierIDs(ctx, q, ruleID)
	if err != nil {
		return 0, err
	}
	var defaultSettings int64
	if err := scanSingleRow(ctx, q, `
		SELECT COUNT(*)
		FROM settings s
		WHERE s.key = 'default_user_level_rule_ids'
		  AND CASE
		        WHEN jsonb_typeof(COALESCE(NULLIF(s.value, ''), '[]')::jsonb) = 'array'
		        THEN COALESCE(NULLIF(s.value, ''), '[]')::jsonb
		        ELSE '[]'::jsonb
		      END ? $1
	`, []any{ruleID}, &defaultSettings); err != nil {
		return 0, err
	}
	var references int64
	for _, tierID := range tierIDs {
		count, err := countUserLevelTierReferences(ctx, q, tierID)
		if err != nil {
			return 0, err
		}
		references += count
	}
	return assignments + defaultSettings + references, nil
}

func listUserLevelTierIDs(ctx context.Context, q sqlQueryer, ruleID string) ([]string, error) {
	rows, err := q.QueryContext(ctx, `SELECT id::text FROM user_level_rule_tiers WHERE rule_id = $1::uuid`, ruleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	ids := make([]string, 0)
	for rows.Next() {
		var tierID string
		if err := rows.Scan(&tierID); err != nil {
			return nil, err
		}
		ids = append(ids, tierID)
	}
	return ids, rows.Err()
}

func countUserLevelTierReferences(ctx context.Context, q sqlQueryer, tierID string) (int64, error) {
	var count int64
	err := scanSingleRow(ctx, q, `
		SELECT
		  (SELECT COUNT(*) FROM groups g
		   WHERE COALESCE(g.level_rate_multipliers, '{}'::jsonb) ? $1)
		+ (SELECT COUNT(DISTINCT g.id) FROM groups g
		   CROSS JOIN LATERAL jsonb_array_elements(
		     CASE WHEN jsonb_typeof(COALESCE(g.dynamic_rate_rules, '[]'::jsonb)) = 'array'
		          THEN COALESCE(g.dynamic_rate_rules, '[]'::jsonb) ELSE '[]'::jsonb END
		   ) AS item(rule)
		   WHERE jsonb_typeof(item.rule->'level_tier_ids') = 'array'
		     AND (item.rule->'level_tier_ids') ? $1)
		+ (SELECT COUNT(DISTINCT a.id) FROM announcements a
		   CROSS JOIN LATERAL jsonb_array_elements(
		     CASE WHEN jsonb_typeof(COALESCE(a.targeting->'any_of', '[]'::jsonb)) = 'array'
		          THEN COALESCE(a.targeting->'any_of', '[]'::jsonb) ELSE '[]'::jsonb END
		   ) AS og(value)
		   CROSS JOIN LATERAL jsonb_array_elements(
		     CASE WHEN jsonb_typeof(COALESCE(og.value->'all_of', '[]'::jsonb)) = 'array'
		          THEN COALESCE(og.value->'all_of', '[]'::jsonb) ELSE '[]'::jsonb END
		   ) AS cond(value)
		   WHERE jsonb_typeof(cond.value->'level_tier_ids') = 'array'
		     AND (cond.value->'level_tier_ids') ? $1)
	`, []any{tierID}, &count)
	return count, err
}
