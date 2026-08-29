package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/lib/pq"
)

type userLevelRepository struct {
	sql *sql.DB
}

func NewUserLevelRepository(sqlDB *sql.DB) service.UserLevelRepository {
	return &userLevelRepository{sql: sqlDB}
}

func (r *userLevelRepository) GetRollingSpend(ctx context.Context, userID string, since, until time.Time) (float64, error) {
	var spend float64
	err := r.sql.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(actual_cost), 0)
		FROM usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3 AND actual_cost > 0
	`, userID, since, until).Scan(&spend)
	return spend, err
}

func (r *userLevelRepository) GetRollingSpendBatch(ctx context.Context, userIDs []string, since, until time.Time) (map[string]float64, error) {
	out := make(map[string]float64, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	rows, err := r.sql.QueryContext(ctx, `
		SELECT user_id, COALESCE(SUM(actual_cost), 0)
		FROM usage_logs
		WHERE user_id = ANY($1) AND created_at >= $2 AND created_at < $3 AND actual_cost > 0
		GROUP BY user_id
	`, pq.Array(userIDs), since, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	for rows.Next() {
		var userID string
		var spend float64
		if err := rows.Scan(&userID, &spend); err != nil {
			return nil, err
		}
		out[userID] = spend
	}
	return out, rows.Err()
}

func (r *userLevelRepository) GetDynamicRateUsage(ctx context.Context, userID, groupID string, keys []service.DynamicRateUsageKey) (map[service.DynamicRateUsageKey]float64, error) {
	out := make(map[service.DynamicRateUsageKey]float64, len(keys))
	if len(keys) == 0 {
		return out, nil
	}
	ruleIDs := make([]string, 0, len(keys))
	dates := make([]string, 0, len(keys))
	for _, key := range keys {
		ruleIDs = append(ruleIDs, key.RuleID)
		dates = append(dates, key.BucketDate)
	}
	rows, err := r.sql.QueryContext(ctx, `
		SELECT rule_id::text, bucket_date::text, used_amount
		FROM user_dynamic_rate_usage
		WHERE user_id = $1 AND group_id = $2
		  AND (rule_id, bucket_date) IN (
			SELECT data.rule_id::uuid, data.bucket_date::date
			FROM unnest($3::text[], $4::text[]) AS data(rule_id, bucket_date)
		  )
	`, userID, groupID, pq.Array(ruleIDs), pq.Array(dates))
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	for rows.Next() {
		var key service.DynamicRateUsageKey
		var used float64
		if err := rows.Scan(&key.RuleID, &key.BucketDate, &used); err != nil {
			return nil, err
		}
		out[key] = used
	}
	return out, rows.Err()
}
