package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/AsukaCC/EasySub2api/internal/service"
)

// Claims are atomic across server replicas and expire after a crashed worker.
func (r *accountRepository) ClaimModelFingerprint(ctx context.Context, id string, snapshot *service.ModelFingerprintSnapshot) (bool, error) {
	body, err := json.Marshal(snapshot)
	if err != nil {
		return false, err
	}
	result, err := r.client.ExecContext(ctx, `UPDATE accounts
		SET extra = jsonb_set(COALESCE(extra, '{}'::jsonb), '{model_fingerprint}', $1::jsonb), updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND (
			COALESCE(extra->'model_fingerprint'->>'status', '') <> 'running'
			OR COALESCE((extra->'model_fingerprint'->>'expires_at')::timestamptz, '-infinity'::timestamptz) <= $3
		)`, string(body), id, snapshot.StartedAt)
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
	result, err := r.client.ExecContext(ctx, `UPDATE accounts
		SET extra = jsonb_set(COALESCE(extra, '{}'::jsonb), '{model_fingerprint}', $1::jsonb), updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND extra->'model_fingerprint'->>'id' = $3`, string(body), id, snapshot.ID)
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
