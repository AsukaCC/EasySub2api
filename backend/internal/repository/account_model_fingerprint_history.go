package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
)

func (r *accountRepository) ListModelFingerprintHistory(ctx context.Context, id string, page, size int) (*service.ModelFingerprintHistory, error) {
	page = max(1, page)
	size = min(100, max(1, size))
	result := &service.ModelFingerprintHistory{Items: []*service.ModelFingerprintSnapshot{}, Page: page, PageSize: size}
	cutoff := time.Now().Add(-service.ModelFingerprintRetention)
	rows, err := r.sql.QueryContext(ctx, `WITH filtered AS (
 SELECT id, started_at, snapshot FROM model_fingerprint_runs WHERE account_id = $1 AND started_at > $2
 ), paged AS (SELECT * FROM filtered ORDER BY started_at DESC, id DESC LIMIT $3 OFFSET $4)
 SELECT (SELECT count(*) FROM filtered), COALESCE(jsonb_agg(snapshot ORDER BY started_at DESC, id DESC), '[]'::jsonb) FROM paged`, id, cutoff, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var body []byte
		if err := rows.Scan(&result.Total, &body); err != nil {
			return nil, err
		}
		var snapshots []*service.ModelFingerprintSnapshot
		if err := json.Unmarshal(body, &snapshots); err != nil {
			return nil, err
		}
		for _, snapshot := range snapshots {
			parsed, err := service.ParseModelFingerprintSnapshot(snapshot, time.Now())
			if err != nil {
				return nil, err
			}
			if parsed != nil {
				result.Items = append(result.Items, parsed)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *accountRepository) GetModelFingerprintSchedule(ctx context.Context, id string) (*service.ModelFingerprintSchedule, error) {
	result := &service.ModelFingerprintSchedule{AccountID: id}
	var body []byte
	rows, err := r.sql.QueryContext(ctx, `SELECT user_id, enabled, options, next_run_at FROM model_fingerprint_schedules WHERE account_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return result, rows.Err()
	}
	if err := rows.Scan(&result.UserID, &result.Enabled, &body, &result.NextRunAt); err != nil {
		return nil, err
	}
	return result, json.Unmarshal(body, &result.Options)
}

func (r *accountRepository) SetModelFingerprintSchedule(ctx context.Context, plan *service.ModelFingerprintSchedule) error {
	body, err := json.Marshal(plan.Options)
	if err != nil {
		return err
	}
	_, err = r.sql.ExecContext(ctx, `INSERT INTO model_fingerprint_schedules(account_id, user_id, enabled, options, next_run_at)
 VALUES ($1,$2,$3,$4::jsonb,$5) ON CONFLICT(account_id) DO UPDATE
 SET user_id=EXCLUDED.user_id, enabled=EXCLUDED.enabled, options=EXCLUDED.options, next_run_at=EXCLUDED.next_run_at`, plan.AccountID, plan.UserID, plan.Enabled, string(body), plan.NextRunAt)
	return err
}

func (r *accountRepository) ListDueModelFingerprints(ctx context.Context, now time.Time) ([]service.ModelFingerprintSchedule, error) {
	rows, err := r.sql.QueryContext(ctx, `SELECT s.account_id, s.user_id, s.options, s.next_run_at FROM model_fingerprint_schedules s
 JOIN accounts a ON a.id=s.account_id JOIN users u ON u.id=s.user_id
 WHERE s.enabled AND s.next_run_at <= $1 AND a.deleted_at IS NULL AND a.status='active'
 AND (a.expires_at IS NULL OR a.expires_at > $1)
 AND u.deleted_at IS NULL AND u.status='active' AND u.role='admin'
 ORDER BY s.next_run_at, s.account_id LIMIT 100`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	plans := []service.ModelFingerprintSchedule{}
	for rows.Next() {
		plan := service.ModelFingerprintSchedule{Enabled: true}
		var body []byte
		if err := rows.Scan(&plan.AccountID, &plan.UserID, &body, &plan.NextRunAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &plan.Options); err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (r *accountRepository) CleanupModelFingerprintHistory(ctx context.Context, now time.Time) error {
	cutoff := now.Add(-service.ModelFingerprintRetention)
	// Limit each sweep so cleanup cannot monopolize the database after downtime.
	if _, err := r.sql.ExecContext(ctx, `DELETE FROM usage_logs WHERE id IN (
 SELECT id FROM usage_logs WHERE request_type=6 AND request_id LIKE 'fingerprint:%' AND created_at <= $1
 ORDER BY created_at LIMIT 5000)`, cutoff); err != nil {
		return err
	}
	_, err := r.sql.ExecContext(ctx, `DELETE FROM model_fingerprint_runs WHERE id IN (
 SELECT id FROM model_fingerprint_runs WHERE started_at <= $1 ORDER BY started_at LIMIT 1000)`, cutoff)
	return err
}
