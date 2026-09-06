package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/timezone"
	"github.com/AsukaCC/EasySub2api/internal/pkg/usagestats"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/lib/pq"
)

var _ service.AccountProfitRepository = (*usageLogRepository)(nil)

func (r *usageLogRepository) GetAccountProfit(ctx context.Context, accountID string, from, to time.Time, page, pageSize int) (*usagestats.AccountProfitResponse, error) {
	var accountCreatedAt time.Time
	var accountExpiresAt sql.NullTime
	if err := scanSingleRow(ctx, r.sql, `SELECT created_at, expires_at FROM accounts WHERE id = $1`, []any{accountID}, &accountCreatedAt, &accountExpiresAt); err != nil {
		return nil, err
	}
	accountCreatedDate := timezone.StartOfDay(accountCreatedAt)
	if from.IsZero() {
		from = accountCreatedDate
	} else {
		from = timezone.StartOfDay(from)
	}
	to = timezone.StartOfDay(to)
	if !from.Before(to) {
		return nil, fmt.Errorf("profit date range must be non-empty")
	}
	if page < 1 || pageSize < 1 {
		return nil, fmt.Errorf("invalid profit pagination")
	}

	today := timezone.Today()
	tomorrow := today.AddDate(0, 0, 1)
	ensureFrom := from
	if accountCreatedDate.Before(ensureFrom) {
		ensureFrom = accountCreatedDate
	}
	if accountExpiresAt.Valid {
		expiryStart := timezone.StartOfDay(accountExpiresAt.Time).AddDate(0, 0, -30)
		if expiryStart.Before(ensureFrom) {
			ensureFrom = expiryStart
		}
	}
	if month := timezone.StartOfMonth(today); month.Before(ensureFrom) {
		ensureFrom = month
	}
	ensureTo := to
	if tomorrow.After(ensureTo) {
		ensureTo = tomorrow
	}
	if err := r.ensureAccountProfitDailyRollups(ctx, []string{accountID}, ensureFrom, ensureTo); err != nil {
		return nil, err
	}

	result := &usagestats.AccountProfitResponse{Page: page, PageSize: pageSize}
	var err error
	if result.Today, err = r.sumAccountProfitPeriod(ctx, accountID, today, tomorrow); err != nil {
		return nil, err
	}
	if result.Week, err = r.sumAccountProfitPeriod(ctx, accountID, timezone.StartOfWeek(today), tomorrow); err != nil {
		return nil, err
	}
	if result.Month, err = r.sumAccountProfitPeriod(ctx, accountID, timezone.StartOfMonth(today), tomorrow); err != nil {
		return nil, err
	}
	if result.Period7d, err = r.sumAccountProfitPeriod(ctx, accountID, today.AddDate(0, 0, -6), tomorrow); err != nil {
		return nil, err
	}
	if result.Lifetime, err = r.sumAccountProfitPeriod(ctx, accountID, accountCreatedDate, tomorrow); err != nil {
		return nil, err
	}
	if accountExpiresAt.Valid {
		expiryDay := timezone.StartOfDay(accountExpiresAt.Time)
		expiryStart := expiryDay.AddDate(0, 0, -30)
		expirySummary, summaryErr := r.sumAccountProfitPeriod(ctx, accountID, expiryStart, expiryDay)
		if summaryErr != nil {
			return nil, summaryErr
		}
		result.Expiry30d = &expirySummary
	}

	total, history, err := r.listAccountProfitHistory(ctx, accountID, from, to, page, pageSize)
	if err != nil {
		return nil, err
	}
	result.Total = total
	result.History = history
	result.HasMore = int64(page*pageSize) < total
	return result, nil
}

func (r *usageLogRepository) ListAccountProfit(ctx context.Context, params usagestats.AccountProfitListParams) (*usagestats.AccountProfitListResponse, error) {
	if params.Page < 1 || params.PageSize < 1 || params.PageSize > 100 {
		return nil, fmt.Errorf("invalid account profit pagination")
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}
	params.SortOrder = strings.ToLower(strings.TrimSpace(params.SortOrder))
	if params.SortOrder != "asc" && params.SortOrder != "desc" {
		return nil, fmt.Errorf("invalid account profit sort order")
	}

	where, filterArgs := buildAccountProfitListWhere(params)
	accountRows, err := r.sql.QueryContext(ctx, `
		SELECT a.id::text, a.created_at
		FROM accounts a
		`+where, filterArgs...)
	if err != nil {
		return nil, err
	}
	accountIDs := make([]string, 0)
	var earliestCreatedAt time.Time
	for accountRows.Next() {
		var id string
		var createdAt time.Time
		if err := accountRows.Scan(&id, &createdAt); err != nil {
			_ = accountRows.Close()
			return nil, err
		}
		accountIDs = append(accountIDs, id)
		if earliestCreatedAt.IsZero() || createdAt.Before(earliestCreatedAt) {
			earliestCreatedAt = createdAt
		}
	}
	if err := accountRows.Err(); err != nil {
		_ = accountRows.Close()
		return nil, err
	}
	if err := accountRows.Close(); err != nil {
		return nil, err
	}

	result := &usagestats.AccountProfitListResponse{
		Items:    []usagestats.AccountProfitListItem{},
		Total:    int64(len(accountIDs)),
		Page:     params.Page,
		PageSize: params.PageSize,
	}
	if len(accountIDs) == 0 {
		return result, nil
	}

	today := timezone.Today()
	tomorrow := today.AddDate(0, 0, 1)
	ensureFrom := timezone.StartOfDay(earliestCreatedAt)
	if err := r.ensureAccountProfitDailyRollups(ctx, accountIDs, ensureFrom, tomorrow); err != nil {
		return nil, err
	}

	quotaExpression := accountProfitQuotaUtilizationSQL("a")
	sortColumn := accountProfitSortColumn(params.SortBy, quotaExpression)
	nullExpirySort := ""
	if strings.HasPrefix(params.SortBy, "expiry_30d_") {
		nullExpirySort = "a.expires_at IS NULL ASC, "
	}
	queryArgs := append([]any{}, filterArgs...)
	tzArg := len(queryArgs) + 1
	todayArg := tzArg + 1
	tomorrowArg := tzArg + 2
	limitArg := tzArg + 3
	offsetArg := tzArg + 4
	queryArgs = append(queryArgs,
		resolveUsageStatsTimezone(),
		today.Format("2006-01-02"),
		tomorrow.Format("2006-01-02"),
		params.PageSize,
		(params.Page-1)*params.PageSize,
	)

	query := `
		SELECT a.id::text,
		       a.name,
		       a.platform,
		       COALESCE(a.subscription_tier, ''),
		       a.status,
		       a.created_at,
		       a.expires_at,
		       a.updated_at,
		       a.extra,
		       ` + quotaExpression + ` AS quota_7d_utilization,
		       COALESCE(SUM(r.revenue_points) FILTER (
		           WHERE r.bucket_date >= ($` + itoa(todayArg) + `::date - 6)
		             AND r.bucket_date < $` + itoa(tomorrowArg) + `::date
		       ), 0),
		       COALESCE(SUM(r.cost_usd) FILTER (
		           WHERE r.bucket_date >= ($` + itoa(todayArg) + `::date - 6)
		             AND r.bucket_date < $` + itoa(tomorrowArg) + `::date
		       ), 0),
		       COALESCE(SUM(r.profit_points) FILTER (
		           WHERE r.bucket_date >= ($` + itoa(todayArg) + `::date - 6)
		             AND r.bucket_date < $` + itoa(tomorrowArg) + `::date
		       ), 0),
		       COALESCE(SUM(r.request_count) FILTER (
		           WHERE r.bucket_date >= ($` + itoa(todayArg) + `::date - 6)
		             AND r.bucket_date < $` + itoa(tomorrowArg) + `::date
		       ), 0),
		       COALESCE(SUM(r.total_tokens) FILTER (
		           WHERE r.bucket_date >= ($` + itoa(todayArg) + `::date - 6)
		             AND r.bucket_date < $` + itoa(tomorrowArg) + `::date
		       ), 0),
		       COALESCE(SUM(r.revenue_points) FILTER (
		           WHERE a.expires_at IS NOT NULL
					AND r.bucket_date >= ((a.expires_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date - 30)
					AND r.bucket_date < (a.expires_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date
		       ), 0),
		       COALESCE(SUM(r.cost_usd) FILTER (
		           WHERE a.expires_at IS NOT NULL
					AND r.bucket_date >= ((a.expires_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date - 30)
					AND r.bucket_date < (a.expires_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date
		       ), 0),
		       COALESCE(SUM(r.profit_points) FILTER (
		           WHERE a.expires_at IS NOT NULL
					AND r.bucket_date >= ((a.expires_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date - 30)
					AND r.bucket_date < (a.expires_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date
		       ), 0),
		       COALESCE(SUM(r.request_count) FILTER (
		           WHERE a.expires_at IS NOT NULL
					AND r.bucket_date >= ((a.expires_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date - 30)
					AND r.bucket_date < (a.expires_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date
		       ), 0),
		       COALESCE(SUM(r.total_tokens) FILTER (
		           WHERE a.expires_at IS NOT NULL
					AND r.bucket_date >= ((a.expires_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date - 30)
					AND r.bucket_date < (a.expires_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date
		       ), 0),
		       COALESCE(SUM(r.revenue_points) FILTER (
				   WHERE r.bucket_date >= (a.created_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date
		             AND r.bucket_date < $` + itoa(tomorrowArg) + `::date
		       ), 0),
		       COALESCE(SUM(r.cost_usd) FILTER (
				   WHERE r.bucket_date >= (a.created_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date
		             AND r.bucket_date < $` + itoa(tomorrowArg) + `::date
		       ), 0),
		       COALESCE(SUM(r.profit_points) FILTER (
				   WHERE r.bucket_date >= (a.created_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date
		             AND r.bucket_date < $` + itoa(tomorrowArg) + `::date
		       ), 0),
		       COALESCE(SUM(r.request_count) FILTER (
				   WHERE r.bucket_date >= (a.created_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date
		             AND r.bucket_date < $` + itoa(tomorrowArg) + `::date
		       ), 0),
		       COALESCE(SUM(r.total_tokens) FILTER (
				   WHERE r.bucket_date >= (a.created_at AT TIME ZONE $` + itoa(tzArg) + `::text)::date
		             AND r.bucket_date < $` + itoa(tomorrowArg) + `::date
		       ), 0)
		FROM accounts a
		LEFT JOIN account_profit_daily_rollups r ON r.account_id = a.id::text
		` + where + `
		GROUP BY a.id, a.name, a.platform, a.subscription_tier, a.status,
		         a.created_at, a.expires_at, a.updated_at, a.extra
		ORDER BY ` + nullExpirySort + sortColumn + ` ` + params.SortOrder + ` NULLS LAST, a.name ASC, a.id ASC
		LIMIT $` + itoa(limitArg) + ` OFFSET $` + itoa(offsetArg)

	rows, err := r.sql.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			id, name, platform, subscriptionTier, status string
			createdAt, updatedAt                         time.Time
			expiresAt                                    sql.NullTime
			extraRaw                                     []byte
			quotaUtilization                             sql.NullFloat64
			period7d, expiry30d, lifetime                usagestats.AccountProfitPeriod
		)
		if err := rows.Scan(
			&id, &name, &platform, &subscriptionTier, &status,
			&createdAt, &expiresAt, &updatedAt, &extraRaw, &quotaUtilization,
			&period7d.RevenuePoints, &period7d.CostUSD, &period7d.ProfitPoints, &period7d.Requests, &period7d.Tokens,
			&expiry30d.RevenuePoints, &expiry30d.CostUSD, &expiry30d.ProfitPoints, &expiry30d.Requests, &expiry30d.Tokens,
			&lifetime.RevenuePoints, &lifetime.CostUSD, &lifetime.ProfitPoints, &lifetime.Requests, &lifetime.Tokens,
		); err != nil {
			return nil, err
		}

		extra := make(map[string]any)
		if len(extraRaw) > 0 {
			if err := json.Unmarshal(extraRaw, &extra); err != nil {
				return nil, fmt.Errorf("decode account quota snapshot %s: %w", id, err)
			}
		}
		quota := service.PassiveWeeklyQuotaSnapshot(&service.Account{
			ID:        id,
			Extra:     extra,
			UpdatedAt: updatedAt,
		}, time.Now())

		item := usagestats.AccountProfitListItem{
			ID:               id,
			Name:             name,
			Platform:         platform,
			SubscriptionTier: subscriptionTier,
			Status:           status,
			CreatedAt:        createdAt,
			Period7d:         period7d,
			Lifetime:         lifetime,
			Quota7d: usagestats.AccountProfitQuota7d{
				Known:            quota.Known,
				UsedPercent:      quota.UsedPercent,
				RemainingPercent: quota.RemainingPct,
				ResetAt:          quota.ResetAt,
				ObservedAt:       quota.ObservedAt,
				Source:           quota.Source,
			},
		}
		if expiresAt.Valid {
			expiresUnix := expiresAt.Time.Unix()
			item.ExpiresAt = &expiresUnix
			item.Expiry30d = &expiry30d
		}
		result.Items = append(result.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func buildAccountProfitListWhere(params usagestats.AccountProfitListParams) (string, []any) {
	conditions := []string{"a.deleted_at IS NULL"}
	args := make([]any, 0, 6)
	addArg := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	if platform := strings.TrimSpace(params.Platform); platform != "" {
		addArg("a.platform = $%d", platform)
	}
	if search := strings.TrimSpace(params.Search); search != "" {
		args = append(args, "%"+search+"%")
		conditions = append(conditions, fmt.Sprintf("(a.name ILIKE $%d OR a.id::text ILIKE $%d)", len(args), len(args)))
	}
	if tier := strings.TrimSpace(params.SubscriptionTier); tier != "" {
		if tier == service.AccountSubscriptionTierUnrecognizedFilter {
			conditions = append(conditions, "a.subscription_tier IS NULL")
		} else {
			addArg("a.subscription_tier = $%d", tier)
		}
	}
	if status := strings.TrimSpace(params.Status); status != "" {
		switch status {
		case service.StatusActive:
			conditions = append(conditions, "a.status = 'active' AND a.schedulable = TRUE AND (a.rate_limit_reset_at IS NULL OR a.rate_limit_reset_at <= NOW()) AND (a.temp_unschedulable_until IS NULL OR a.temp_unschedulable_until <= NOW())")
		case "rate_limited":
			conditions = append(conditions, "a.status = 'active' AND a.rate_limit_reset_at > NOW() AND (a.temp_unschedulable_until IS NULL OR a.temp_unschedulable_until <= NOW())")
		case "temp_unschedulable":
			conditions = append(conditions, "a.status = 'active' AND a.temp_unschedulable_until > NOW()")
		case "unschedulable":
			conditions = append(conditions, "a.status = 'active' AND a.schedulable = FALSE AND (a.rate_limit_reset_at IS NULL OR a.rate_limit_reset_at <= NOW()) AND (a.temp_unschedulable_until IS NULL OR a.temp_unschedulable_until <= NOW())")
		default:
			addArg("a.status = $%d", status)
		}
	}
	if expiryStatus := strings.TrimSpace(params.ExpiryStatus); expiryStatus != "" {
		switch expiryStatus {
		case service.AccountExpiryStatusExpiring:
			conditions = append(conditions, "a.expires_at > NOW() AND a.expires_at <= NOW() + INTERVAL '7 days'")
		case service.AccountExpiryStatusExpired:
			conditions = append(conditions, "a.expires_at IS NOT NULL AND a.expires_at <= NOW()")
		}
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

func accountProfitQuotaUtilizationSQL(alias string) string {
	return fmt.Sprintf(`COALESCE(
		CASE WHEN %s.extra #>> '{passive_weekly_quota,used_percent}' ~ '^[0-9]+(\.[0-9]+)?$'
			THEN (%s.extra #>> '{passive_weekly_quota,used_percent}')::double precision END,
		CASE WHEN %s.extra->>'passive_usage_7d_utilization' ~ '^[0-9]+(\.[0-9]+)?$'
			THEN (%s.extra->>'passive_usage_7d_utilization')::double precision * 100 END,
		CASE WHEN %s.extra->>'codex_7d_used_percent' ~ '^[0-9]+(\.[0-9]+)?$'
			THEN (%s.extra->>'codex_7d_used_percent')::double precision END,
		CASE WHEN %s.extra->'grok_billing_snapshot'->>'usage_percent' ~ '^[0-9]+(\.[0-9]+)?$'
			THEN (%s.extra->'grok_billing_snapshot'->>'usage_percent')::double precision END,
		CASE WHEN %s.extra->'grok_billing_snapshot'->>'used_percent' ~ '^[0-9]+(\.[0-9]+)?$'
			THEN (%s.extra->'grok_billing_snapshot'->>'used_percent')::double precision END,
		CASE WHEN %s.extra->'ollama_cloud_usage_snapshot'->'data'->'seven_day'->>'used_percent' ~ '^[0-9]+(\.[0-9]+)?$'
			THEN (%s.extra->'ollama_cloud_usage_snapshot'->'data'->'seven_day'->>'used_percent')::double precision END
	)`, alias, alias, alias, alias, alias, alias, alias, alias, alias, alias, alias, alias)
}

func accountProfitSortColumn(sortBy, quotaExpression string) string {
	if sortBy == "" {
		return "a.created_at"
	}
	columns := map[string]string{
		"period_7d_revenue":    "period_7d_revenue",
		"period_7d_cost":       "period_7d_cost",
		"period_7d_profit":     "period_7d_profit",
		"period_7d_tokens":     "period_7d_tokens",
		"expiry_30d_revenue":   "expiry_30d_revenue",
		"expiry_30d_cost":      "expiry_30d_cost",
		"expiry_30d_profit":    "expiry_30d_profit",
		"expiry_30d_tokens":    "expiry_30d_tokens",
		"lifetime_revenue":     "lifetime_revenue",
		"lifetime_cost":        "lifetime_cost",
		"lifetime_profit":      "lifetime_profit",
		"lifetime_tokens":      "lifetime_tokens",
		"quota_7d_utilization": quotaExpression,
		"expires_at":           "a.expires_at",
		"created_at":           "a.created_at",
	}
	if column, ok := columns[sortBy]; ok {
		return column
	}
	return "a.created_at"
}

func (r *usageLogRepository) ensureAccountProfitDailyRollups(ctx context.Context, accountIDs []string, from, to time.Time) error {
	if len(accountIDs) == 0 || !from.Before(to) {
		return nil
	}
	query := `
		WITH aggregated AS (
			SELECT ul.account_id,
				       (ul.created_at AT TIME ZONE $2::text)::date AS bucket_date,
			       COUNT(*)::bigint AS request_count,
			       COALESCE(SUM(ul.input_tokens + ul.output_tokens + ul.cache_creation_tokens + ul.cache_read_tokens), 0)::bigint AS total_tokens,
			       COALESCE(SUM(ul.actual_cost), 0)::numeric AS revenue_points,
			       COALESCE(SUM(COALESCE(ul.account_stats_cost, ul.total_cost) * COALESCE(ul.account_rate_multiplier, 1)), 0)::numeric AS cost_usd
			FROM usage_logs ul
			WHERE ul.account_id::text = ANY($1::text[])
			  AND ul.created_at >= $3 AND ul.created_at < $4
			GROUP BY ul.account_id, (ul.created_at AT TIME ZONE $2::text)::date
		)
		INSERT INTO account_profit_daily_rollups
			(account_id, account_name_snapshot, bucket_date, request_count, total_tokens, revenue_points, cost_usd, profit_points, computed_at, updated_at)
		SELECT a.account_id,
		       COALESCE(acc.name, ''),
		       a.bucket_date,
		       a.request_count,
		       a.total_tokens,
		       a.revenue_points,
		       a.cost_usd,
		       a.revenue_points - a.cost_usd,
		       NOW(), NOW()
		FROM aggregated a
		LEFT JOIN accounts acc ON acc.id = a.account_id
		ON CONFLICT (account_id, bucket_date) DO UPDATE SET
			account_name_snapshot = CASE
				WHEN EXCLUDED.account_name_snapshot <> '' THEN EXCLUDED.account_name_snapshot
				ELSE account_profit_daily_rollups.account_name_snapshot
			END,
			request_count = EXCLUDED.request_count,
			total_tokens = EXCLUDED.total_tokens,
			revenue_points = EXCLUDED.revenue_points,
			cost_usd = EXCLUDED.cost_usd,
			profit_points = EXCLUDED.profit_points,
			computed_at = EXCLUDED.computed_at,
			updated_at = NOW()
	`
	_, err := r.sql.ExecContext(ctx, query, pq.Array(accountIDs), resolveUsageStatsTimezone(), from, to)
	return err
}

func (r *usageLogRepository) sumAccountProfitPeriod(ctx context.Context, accountID string, from, to time.Time) (usagestats.AccountProfitPeriod, error) {
	var result usagestats.AccountProfitPeriod
	query := `
		SELECT COALESCE(SUM(revenue_points), 0),
		       COALESCE(SUM(cost_usd), 0),
		       COALESCE(SUM(profit_points), 0),
		       COALESCE(SUM(request_count), 0),
		       COALESCE(SUM(total_tokens), 0)
		FROM account_profit_daily_rollups
		WHERE account_id = $1 AND bucket_date >= $2::date AND bucket_date < $3::date
	`
	if err := scanSingleRow(ctx, r.sql, query, []any{accountID, from.Format("2006-01-02"), to.Format("2006-01-02")}, &result.RevenuePoints, &result.CostUSD, &result.ProfitPoints, &result.Requests, &result.Tokens); err != nil {
		return result, err
	}
	return result, nil
}

func (r *usageLogRepository) listAccountProfitHistory(ctx context.Context, accountID string, from, to time.Time, page, pageSize int) (int64, []usagestats.AccountProfitDailyRecord, error) {
	total := int64(dateOnlyUTC(to).Sub(dateOnlyUTC(from)).Hours() / 24)
	if total < 0 {
		total = 0
	}
	offset := (page - 1) * pageSize
	query := `
		SELECT TO_CHAR(days.bucket_date, 'YYYY-MM-DD'),
		       COALESCE(r.revenue_points, 0),
		       COALESCE(r.cost_usd, 0),
		       COALESCE(r.profit_points, 0),
		       COALESCE(r.request_count, 0),
		       COALESCE(r.total_tokens, 0)
		FROM generate_series($2::date, ($3::date - 1), interval '1 day') AS days(bucket_date)
		LEFT JOIN account_profit_daily_rollups r
			ON r.account_id = $1 AND r.bucket_date = days.bucket_date::date
		ORDER BY days.bucket_date DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.sql.QueryContext(ctx, query, accountID, from.Format("2006-01-02"), to.Format("2006-01-02"), pageSize, offset)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = rows.Close() }()
	history := make([]usagestats.AccountProfitDailyRecord, 0, pageSize)
	for rows.Next() {
		var item usagestats.AccountProfitDailyRecord
		if err := rows.Scan(&item.Date, &item.RevenuePoints, &item.CostUSD, &item.ProfitPoints, &item.Requests, &item.Tokens); err != nil {
			return 0, nil, err
		}
		item.Label = item.Date
		history = append(history, item)
	}
	if err := rows.Err(); err != nil {
		return 0, nil, err
	}
	return total, history, nil
}

func dateOnlyUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
