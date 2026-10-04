package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
)

// Claims are atomic across server replicas and expire after a crashed worker.
func (r *accountRepository) ClaimModelFingerprint(ctx context.Context, id string, snapshot *service.ModelFingerprintSnapshot) (bool, error) {
	body, err := json.Marshal(snapshot)
	if err != nil {
		return false, err
	}
	result, err := r.client.ExecContext(ctx, `WITH claimed AS (UPDATE accounts
		SET extra = jsonb_set(COALESCE(extra, '{}'::jsonb), '{model_fingerprint}', $1::jsonb), updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND (
			COALESCE(extra->'model_fingerprint'->>'status', '') <> 'running'
			OR COALESCE((extra->'model_fingerprint'->>'expires_at')::timestamptz, '-infinity'::timestamptz) <= $3
		) AND ($1::jsonb->>'source' IS DISTINCT FROM 'scheduled' OR EXISTS (
			SELECT 1 FROM model_fingerprint_schedules s JOIN users u ON u.id = s.user_id
			WHERE s.account_id = accounts.id AND s.enabled AND s.next_run_at <= $3
			AND s.user_id::text = $1::jsonb->>'user_id'
			AND s.options->>'model_id' = $1::jsonb->>'model'
			AND s.options->>'protocol' = $1::jsonb->>'protocol'
			AND COALESCE(s.options->>'reasoning_effort', '') = COALESCE($1::jsonb->>'reasoning_effort', '')
			AND u.role = 'admin' AND u.status = 'active' AND u.deleted_at IS NULL
			FOR UPDATE OF s
		)) RETURNING id), advanced AS (
		UPDATE model_fingerprint_schedules SET next_run_at = date_trunc('hour', $3::timestamptz)
			+ (floor(extract(minute from $3::timestamptz) / 30) + 1) * interval '30 minutes'
		WHERE account_id IN (SELECT id FROM claimed) AND $1::jsonb->>'source' = 'scheduled'
		RETURNING account_id
		) INSERT INTO model_fingerprint_runs (id, account_id, started_at, snapshot)
		SELECT ($1::jsonb->>'id')::uuid, id, $3, $1::jsonb FROM claimed`, string(body), id, snapshot.StartedAt)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err == nil && n > 0 {
		r.syncSchedulerAccountSnapshot(ctx, id)
	}
	return n > 0, err
}

// A worker that has lost its lease must never overwrite a newer test result.
func (r *accountRepository) SaveModelFingerprint(ctx context.Context, id string, snapshot *service.ModelFingerprintSnapshot) error {
	body, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	result, err := r.client.ExecContext(ctx, `WITH saved AS (UPDATE accounts
		SET extra = jsonb_set(COALESCE(extra, '{}'::jsonb), '{model_fingerprint}', $1::jsonb), updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND extra->'model_fingerprint'->>'id' = $3 RETURNING id)
		INSERT INTO model_fingerprint_runs (id, account_id, started_at, snapshot)
		SELECT $3::uuid, id, ($1::jsonb->>'started_at')::timestamptz, $1::jsonb FROM saved
		ON CONFLICT (id) DO UPDATE SET snapshot = EXCLUDED.snapshot`, string(body), id, snapshot.ID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("model fingerprint lease lost")
	}
	r.syncSchedulerAccountSnapshot(ctx, id)
	return nil
}

// Evaluate expiration in the UPDATE so cleanup cannot remove a newer test.
func (r *accountRepository) DeleteExpiredModelFingerprints(ctx context.Context, now time.Time) (int64, error) {
	rows, err := r.sql.QueryContext(ctx, `UPDATE accounts
		SET extra = extra - 'model_fingerprint', updated_at = NOW()
		WHERE extra ? 'model_fingerprint' AND COALESCE(
			CASE WHEN extra->'model_fingerprint'->>'sampling_mode' = 'independent'
			THEN NULLIF(extra->'model_fingerprint'->>'started_at', '')::timestamptz END,
			NULLIF(extra->'model_fingerprint'->>'finished_at', '')::timestamptz,
			NULLIF(extra->'model_fingerprint'->>'expires_at', '')::timestamptz,
			NULLIF(extra->'model_fingerprint'->>'started_at', '')::timestamptz,
			'-infinity'::timestamptz
		) <= $1
		RETURNING id`, now.Add(-service.ModelFingerprintRetention))
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	r.syncSchedulerAccountSnapshots(ctx, ids)
	return int64(len(ids)), nil
}
