package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
)

func (r *accountRepository) ListQueuedModelFingerprints(ctx context.Context) ([]service.ModelFingerprintQueuedJob, error) {
	rows, err := r.sql.QueryContext(ctx, `SELECT id, extra->'model_fingerprint' FROM accounts
		WHERE deleted_at IS NULL AND extra->'model_fingerprint'->>'status' = 'queued'
		ORDER BY extra->'model_fingerprint'->>'started_at', id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	jobs := []service.ModelFingerprintQueuedJob{}
	for rows.Next() {
		var job service.ModelFingerprintQueuedJob
		var body []byte
		if err := rows.Scan(&job.AccountID, &body); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &job.Snapshot); err != nil {
			return nil, err
		}
		if job.Snapshot != nil {
			jobs = append(jobs, job)
		}
	}
	return jobs, rows.Err()
}

func (r *accountRepository) ActivateModelFingerprint(ctx context.Context, id string, snapshot *service.ModelFingerprintSnapshot) (bool, error) {
	body, err := json.Marshal(snapshot)
	if err != nil {
		return false, err
	}
	result, err := r.client.ExecContext(ctx, `WITH activated AS (UPDATE accounts
		SET extra = jsonb_set(extra, '{model_fingerprint}', $1::jsonb), updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND extra->'model_fingerprint'->>'id' = $3
		AND extra->'model_fingerprint'->>'status' = 'queued' RETURNING id)
		INSERT INTO model_fingerprint_runs (id, account_id, started_at, snapshot)
		SELECT $3::uuid, id, ($1::jsonb->>'started_at')::timestamptz, $1::jsonb FROM activated
		ON CONFLICT (id) DO UPDATE SET snapshot = EXCLUDED.snapshot`, string(body), id, snapshot.ID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err == nil && n > 0 {
		r.syncSchedulerAccountSnapshot(ctx, id)
	}
	return n > 0, err
}

func (r *accountRepository) CancelModelFingerprint(ctx context.Context, id, jobID string) (bool, error) {
	result, err := r.client.ExecContext(ctx, `WITH cancelled AS (UPDATE accounts
		SET extra = jsonb_set(extra, '{model_fingerprint}', (extra->'model_fingerprint') ||
		jsonb_build_object('status', 'cancelled', 'error', '', 'queue_reason', '', 'finished_at', $3::text)), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL AND extra->'model_fingerprint'->>'id' = $2
		AND extra->'model_fingerprint'->>'status' IN ('queued', 'running', 'waiting')
		RETURNING id, extra->'model_fingerprint' AS snapshot)
		INSERT INTO model_fingerprint_runs (id, account_id, started_at, snapshot)
		SELECT $2::uuid, id, (snapshot->>'started_at')::timestamptz, snapshot FROM cancelled
		ON CONFLICT (id) DO UPDATE SET snapshot = EXCLUDED.snapshot`, id, jobID, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err == nil && n > 0 {
		r.syncSchedulerAccountSnapshot(ctx, id)
	}
	return n > 0, err
}
