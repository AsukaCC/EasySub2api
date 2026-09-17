package repository

import (
	"context"
	"encoding/json"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

func lockAccountProtection(ctx context.Context, client *dbent.Client, account *service.Account) error {
	rows, err := client.QueryContext(ctx, "SELECT extra, updated_at FROM accounts WHERE id = $1 AND deleted_at IS NULL FOR NO KEY UPDATE", account.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return service.ErrAccountNotFound
	}
	var raw []byte
	var revision time.Time
	if err := rows.Scan(&raw, &revision); err != nil {
		return err
	}
	current := &service.Account{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &current.Extra); err != nil {
			return err
		}
	}
	if expected, ok := service.GetProtectionWriteExpectation(ctx); ok && (expected.AccountID != account.ID || !expected.UpdatedAt.Equal(revision)) {
		return service.ErrProtectionConflict
	}
	account.Extra = service.PreserveAccountProtection(ctx, current, account.Extra)
	service.BoundAccountProtectionConcurrency(account)
	return service.ValidateAccountProtectionConfiguration(account)
}

// JSONB preservation runs inside the same write as background and bulk updates.
func preserveProtectionExtraSQL(expr string) string {
	keys := "ARRAY['anti_degrade','anti_degradation','protection_scope','codex_fingerprint_seed'] || CASE WHEN extra #>> '{anti_degrade,enabled}' = 'true' THEN ARRAY['codex_fingerprint_mode','enable_tls_fingerprint','tls_fingerprint_builtin','tls_fingerprint_profile_id','proxy_mode'] ELSE ARRAY[]::text[] END"
	return "((" + expr + ") - (" + keys + ")) || COALESCE((SELECT jsonb_object_agg(key,value) FROM jsonb_each(COALESCE(extra,'{}'::jsonb)) WHERE key = ANY(" + keys + ")), '{}'::jsonb)"
}

func (r *accountRepository) EnsureCodexIdentitySeed(ctx context.Context, id string) (*service.Account, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	client := tx.Client()
	_, err = client.ExecContext(ctx, "UPDATE accounts SET extra = "+ensureCodexFingerprintSeedSQL("COALESCE(extra, '{}'::jsonb)")+", updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL AND NOT COALESCE(("+codexFingerprintSeedValidSQL("extra")+"), false)", id)
	if err != nil {
		return nil, err
	}
	if err := enqueueSchedulerOutbox(ctx, client, service.SchedulerOutboxEventAccountChanged, &id, nil, nil); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	r.syncSchedulerAccountSnapshot(ctx, id)
	return r.GetByID(ctx, id)
}
