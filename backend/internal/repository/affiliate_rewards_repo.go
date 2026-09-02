package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

func (r *affiliateRepository) BindInviterWithRewards(ctx context.Context, userID, inviterID string, config service.AffiliateBindingRewardConfig) (*service.AffiliateBindingResult, error) {
	result := &service.AffiliateBindingResult{}
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		if _, err := ensureUserAffiliateWithClient(txCtx, txClient, userID); err != nil {
			return err
		}
		if _, err := ensureUserAffiliateWithClient(txCtx, txClient, inviterID); err != nil {
			return err
		}

		rows, err := txClient.QueryContext(txCtx, `
SELECT inviter_id::text
FROM user_affiliates
WHERE user_id = $1
FOR UPDATE`, userID)
		if err != nil {
			return fmt.Errorf("lock affiliate relationship: %w", err)
		}
		var current sql.NullString
		if !rows.Next() {
			_ = rows.Close()
			return service.ErrAffiliateProfileNotFound
		}
		if err := rows.Scan(&current); err != nil {
			_ = rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		if current.Valid {
			if current.String != inviterID {
				return service.ErrAffiliateAlreadyBound
			}
			result.Bound = true
			result.AlreadyBound = true
			return nil
		}

		availableRows, err := txClient.QueryContext(txCtx, `
SELECT EXISTS(
    SELECT 1 FROM users
    WHERE id = $1 AND deleted_at IS NULL AND status = 'active'
)`, inviterID)
		if err != nil {
			return err
		}
		var inviterAvailable bool
		if availableRows.Next() {
			err = availableRows.Scan(&inviterAvailable)
		}
		_ = availableRows.Close()
		if err != nil {
			return err
		}
		if !inviterAvailable {
			return service.ErrAffiliateInviterUnavailable
		}

		cycleRows, err := txClient.QueryContext(txCtx, `
WITH RECURSIVE inviter_chain AS (
    SELECT ua.user_id, ua.inviter_id
    FROM user_affiliates ua
    WHERE ua.user_id = $1
    UNION ALL
    SELECT parent.user_id, parent.inviter_id
    FROM user_affiliates parent
    JOIN inviter_chain child ON parent.user_id = child.inviter_id
    WHERE child.inviter_id IS NOT NULL
)
SELECT EXISTS(SELECT 1 FROM inviter_chain WHERE user_id = $2)`, inviterID, userID)
		if err != nil {
			return err
		}
		var createsCycle bool
		if cycleRows.Next() {
			err = cycleRows.Scan(&createsCycle)
		}
		_ = cycleRows.Close()
		if err != nil {
			return err
		}
		if createsCycle || userID == inviterID {
			return service.ErrAffiliateInviteCycle
		}

		res, err := txClient.ExecContext(txCtx, `
UPDATE user_affiliates
SET inviter_id = $1,
    inviter_bound_at = NOW(),
    binding_reward_version = $2,
    updated_at = NOW()
WHERE user_id = $3 AND inviter_id IS NULL`, inviterID, service.AffiliateBindingRewardVersion, userID)
		if err != nil {
			return fmt.Errorf("bind inviter with rewards: %w", err)
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return service.ErrAffiliateAlreadyBound
		}
		if _, err := txClient.ExecContext(txCtx, `
UPDATE user_affiliates SET aff_count = aff_count + 1, updated_at = NOW() WHERE user_id = $1`, inviterID); err != nil {
			return fmt.Errorf("increment inviter affiliate count: %w", err)
		}

		result.InviterReward, err = creditAffiliateBindingRewardTx(txCtx, txClient, inviterID,
			"affiliate_binding_inviter_reward", userID, config.InviterPoints, config.InviterValidityDays)
		if err != nil {
			return err
		}
		result.InviteeReward, err = creditAffiliateBindingRewardTx(txCtx, txClient, userID,
			"affiliate_binding_invitee_reward", userID, config.InviteePoints, config.InviteeValidityDays)
		if err != nil {
			return err
		}
		result.Bound = true
		return nil
	})
	return result, err
}

func creditAffiliateBindingRewardTx(ctx context.Context, exec sqlQueryExecutor, userID, sourceType, sourceID string, rawAmount float64, validityDays int) (service.AffiliateBindingReward, error) {
	amount := service.ClampAffiliateBindingRewardPoints(rawAmount)
	if amount <= 0 {
		return service.AffiliateBindingReward{}, nil
	}
	validityDays = service.ClampAffiliateBindingRewardValidity(validityDays)
	expiresAt := time.Now().UTC().Add(time.Duration(validityDays) * 24 * time.Hour)
	key := "wallet-credit:" + sourceType + ":" + sourceID
	exists, err := walletTransactionExists(ctx, exec, key)
	if err != nil {
		return service.AffiliateBindingReward{}, err
	}
	if exists {
		return service.AffiliateBindingReward{Points: amount, ExpiresAt: &expiresAt, Applied: false}, nil
	}
	row, err := lockWalletUser(ctx, exec, userID)
	if err != nil {
		return service.AffiliateBindingReward{}, err
	}
	if _, err := expireUserBonusTx(ctx, exec, userID, &row); err != nil {
		return service.AffiliateBindingReward{}, err
	}
	before := row
	bonusAvailable := walletMoney(math.Min(amount, math.Max(row.balance+amount, 0)))
	if _, err := exec.ExecContext(ctx, `
UPDATE users
SET balance = balance + $1, bonus_balance = bonus_balance + $2, updated_at = NOW()
WHERE id = $3 AND deleted_at IS NULL`, amount, bonusAvailable, userID); err != nil {
		return service.AffiliateBindingReward{}, err
	}
	row.balance = walletMoney(row.balance + amount)
	row.bonus = walletMoney(row.bonus + bonusAvailable)
	status := "active"
	if bonusAvailable <= 0 {
		status = "spent"
	}
	grantID := newWalletUUID()
	if _, err := exec.ExecContext(ctx, `
INSERT INTO wallet_bonus_grants (
    id, user_id, original_amount, remaining_amount, spent_amount, expires_at,
    source_type, source_id, status
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		grantID, userID, amount, bonusAvailable, walletMoney(amount-bonusAvailable), expiresAt,
		sourceType, sourceID, status); err != nil {
		return service.AffiliateBindingReward{}, err
	}
	if err := insertWalletTransaction(ctx, exec, userID, "credit", amount, bonusAvailable, 0, 0,
		before, row, sourceType, sourceID, key, "affiliate invitation binding reward"); err != nil {
		return service.AffiliateBindingReward{}, err
	}
	return service.AffiliateBindingReward{Points: amount, ExpiresAt: &expiresAt, Applied: true}, nil
}

func (r *affiliateRepository) LegacyBindingRewardScope(ctx context.Context) (int, string, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
SELECT COUNT(*),
       COALESCE(MD5(STRING_AGG(
           ua.user_id::text || ':' || ua.inviter_id::text,
           ',' ORDER BY ua.user_id
       )), '')
FROM user_affiliates ua
JOIN users invitee ON invitee.id = ua.user_id AND invitee.deleted_at IS NULL
JOIN users inviter ON inviter.id = ua.inviter_id AND inviter.deleted_at IS NULL
WHERE ua.inviter_id IS NOT NULL AND ua.binding_reward_version = 0`)
	if err != nil {
		return 0, "", err
	}
	defer rows.Close() //nolint:errcheck
	var count int
	var scopeHash string
	if rows.Next() {
		err = rows.Scan(&count, &scopeHash)
	}
	return count, scopeHash, err
}

func (r *affiliateRepository) CreateRewardBackfillRun(ctx context.Context, actorID, previewToken string, config service.AffiliateBindingRewardConfig, eligible int) (*service.AffiliateRewardBackfillRun, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
INSERT INTO affiliate_reward_backfill_runs (
    preview_token, inviter_points, inviter_validity_days, invitee_points,
    invitee_validity_days, eligible_relations, created_by
) VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,'')::uuid)
RETURNING id::text`, previewToken, config.InviterPoints, config.InviterValidityDays,
		config.InviteePoints, config.InviteeValidityDays, eligible, actorID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "idx_affiliate_reward_backfill_active") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, service.ErrAffiliateBackfillRunning
		}
		return nil, err
	}
	var id string
	if !rows.Next() {
		_ = rows.Close()
		return nil, errors.New("create affiliate reward backfill run returned no id")
	}
	if err := rows.Scan(&id); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return r.GetRewardBackfillRun(ctx, id)
}

func (r *affiliateRepository) GetRewardBackfillRun(ctx context.Context, runID string) (*service.AffiliateRewardBackfillRun, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
SELECT id::text, status, inviter_points::double precision, inviter_validity_days,
       invitee_points::double precision, invitee_validity_days, eligible_relations,
       processed_relations, inviter_grants, invitee_grants,
       inviter_points_granted::double precision, invitee_points_granted::double precision,
       error_message, created_at, started_at, completed_at, updated_at
FROM affiliate_reward_backfill_runs WHERE id = $1`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	if !rows.Next() {
		return nil, infraerrors.NotFound("AFFILIATE_BACKFILL_NOT_FOUND", "affiliate reward backfill not found")
	}
	return scanAffiliateRewardBackfillRun(rows)
}

func (r *affiliateRepository) GetActiveRewardBackfillRun(ctx context.Context) (*service.AffiliateRewardBackfillRun, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
SELECT id::text, status, inviter_points::double precision, inviter_validity_days,
       invitee_points::double precision, invitee_validity_days, eligible_relations,
       processed_relations, inviter_grants, invitee_grants,
       inviter_points_granted::double precision, invitee_points_granted::double precision,
       error_message, created_at, started_at, completed_at, updated_at
FROM affiliate_reward_backfill_runs
WHERE status IN ('pending', 'running')
ORDER BY created_at
LIMIT 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	if !rows.Next() {
		return nil, nil
	}
	return scanAffiliateRewardBackfillRun(rows)
}

func scanAffiliateRewardBackfillRun(rows *sql.Rows) (*service.AffiliateRewardBackfillRun, error) {
	var run service.AffiliateRewardBackfillRun
	var startedAt, completedAt sql.NullTime
	err := rows.Scan(&run.ID, &run.Status, &run.Config.InviterPoints, &run.Config.InviterValidityDays,
		&run.Config.InviteePoints, &run.Config.InviteeValidityDays, &run.EligibleRelations,
		&run.ProcessedRelations, &run.InviterGrants, &run.InviteeGrants,
		&run.InviterPointsGranted, &run.InviteePointsGranted, &run.ErrorMessage,
		&run.CreatedAt, &startedAt, &completedAt, &run.UpdatedAt)
	if startedAt.Valid {
		run.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		run.CompletedAt = &completedAt.Time
	}
	return &run, err
}

func (r *affiliateRepository) ProcessRewardBackfillBatch(ctx context.Context, runID string, limit int) (int, bool, error) {
	if limit < 1 || limit > 500 {
		limit = 100
	}
	processed := 0
	done := false
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		runRows, err := txClient.QueryContext(txCtx, `
UPDATE affiliate_reward_backfill_runs
SET status = 'running', started_at = COALESCE(started_at, NOW()), updated_at = NOW()
WHERE id = $1 AND status IN ('pending','running')
RETURNING inviter_points::double precision, inviter_validity_days,
          invitee_points::double precision, invitee_validity_days`, runID)
		if err != nil {
			return err
		}
		var config service.AffiliateBindingRewardConfig
		if !runRows.Next() {
			_ = runRows.Close()
			done = true
			return nil
		}
		if err := runRows.Scan(&config.InviterPoints, &config.InviterValidityDays, &config.InviteePoints, &config.InviteeValidityDays); err != nil {
			_ = runRows.Close()
			return err
		}
		_ = runRows.Close()

		candidateRows, err := txClient.QueryContext(txCtx, `
SELECT ua.user_id::text, ua.inviter_id::text
FROM user_affiliates ua
JOIN users invitee ON invitee.id = ua.user_id AND invitee.deleted_at IS NULL
JOIN users inviter ON inviter.id = ua.inviter_id AND inviter.deleted_at IS NULL
WHERE ua.inviter_id IS NOT NULL AND ua.binding_reward_version = 0
ORDER BY ua.user_id
FOR UPDATE OF ua SKIP LOCKED
LIMIT $1`, limit)
		if err != nil {
			return err
		}
		type relation struct{ inviteeID, inviterID string }
		var relations []relation
		for candidateRows.Next() {
			var item relation
			if err := candidateRows.Scan(&item.inviteeID, &item.inviterID); err != nil {
				_ = candidateRows.Close()
				return err
			}
			relations = append(relations, item)
		}
		_ = candidateRows.Close()

		inviterGrants, inviteeGrants := 0, 0
		inviterPoints, inviteePoints := 0.0, 0.0
		for _, item := range relations {
			inviterReward, err := creditAffiliateBindingRewardTx(txCtx, txClient, item.inviterID,
				"affiliate_binding_inviter_reward", item.inviteeID, config.InviterPoints, config.InviterValidityDays)
			if err != nil {
				return err
			}
			inviteeReward, err := creditAffiliateBindingRewardTx(txCtx, txClient, item.inviteeID,
				"affiliate_binding_invitee_reward", item.inviteeID, config.InviteePoints, config.InviteeValidityDays)
			if err != nil {
				return err
			}
			if inviterReward.Applied {
				inviterGrants++
				inviterPoints += inviterReward.Points
			}
			if inviteeReward.Applied {
				inviteeGrants++
				inviteePoints += inviteeReward.Points
			}
			if _, err := txClient.ExecContext(txCtx, `
UPDATE user_affiliates SET binding_reward_version = $1, updated_at = NOW()
WHERE user_id = $2 AND binding_reward_version = 0`, service.AffiliateBindingRewardVersion, item.inviteeID); err != nil {
				return err
			}
		}
		processed = len(relations)
		if _, err := txClient.ExecContext(txCtx, `
UPDATE affiliate_reward_backfill_runs
SET processed_relations = processed_relations + $2,
    inviter_grants = inviter_grants + $3,
    invitee_grants = invitee_grants + $4,
    inviter_points_granted = inviter_points_granted + $5,
    invitee_points_granted = invitee_points_granted + $6,
    updated_at = NOW()
WHERE id = $1`, runID, processed, inviterGrants, inviteeGrants,
			walletMoney(inviterPoints), walletMoney(inviteePoints)); err != nil {
			return err
		}

		remainingRows, err := txClient.QueryContext(txCtx, `
SELECT EXISTS(
    SELECT 1 FROM user_affiliates ua
    JOIN users invitee ON invitee.id = ua.user_id AND invitee.deleted_at IS NULL
    JOIN users inviter ON inviter.id = ua.inviter_id AND inviter.deleted_at IS NULL
    WHERE ua.inviter_id IS NOT NULL AND ua.binding_reward_version = 0
)`)
		if err != nil {
			return err
		}
		var remaining bool
		if remainingRows.Next() {
			err = remainingRows.Scan(&remaining)
		}
		_ = remainingRows.Close()
		if err != nil {
			return err
		}
		done = !remaining
		if done {
			_, err = txClient.ExecContext(txCtx, `
UPDATE affiliate_reward_backfill_runs
SET status = 'completed', completed_at = NOW(), updated_at = NOW()
WHERE id = $1`, runID)
		}
		return err
	})
	return processed, done, err
}

func (r *affiliateRepository) FailRewardBackfillRun(ctx context.Context, runID, message string) error {
	if len(message) > 2000 {
		message = message[:2000]
	}
	_, err := clientFromContext(ctx, r.client).ExecContext(ctx, `
UPDATE affiliate_reward_backfill_runs
SET status = 'failed', error_message = $2, completed_at = NOW(), updated_at = NOW()
WHERE id = $1 AND status IN ('pending','running')`, runID, message)
	return err
}
