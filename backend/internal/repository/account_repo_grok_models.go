package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

var errGrokCatalogInvalidJSON = errors.New("grok model catalog: invalid json")

var _ service.GrokCatalogSnapshotRepository = (*accountRepository)(nil)

const compareAndSwapGrokObservedModelsSQL = `UPDATE accounts
SET extra = jsonb_set(COALESCE(extra, '{}'::jsonb), '{grok_observed_models}', $3::jsonb, true),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND platform = 'grok'
  AND COALESCE(extra -> 'grok_observed_models', 'null'::jsonb) = $2::jsonb`

const updateGrokCatalogExtraIfObservedSQL = `UPDATE accounts
SET extra = COALESCE(extra, '{}'::jsonb) || $3::jsonb,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND platform = 'grok'
  AND COALESCE(extra -> 'grok_observed_models', 'null'::jsonb) = $2::jsonb`

// Keep the local discovery/capability cache tied to the observation that
// produced it, including when metadata enrichment finishes out of order.
func (r *accountRepository) UpdateGrokCatalogExtraIfObserved(ctx context.Context, id string, expected json.RawMessage, updates map[string]any) (bool, error) {
	if scoped := r.accountTransactionScope(ctx); scoped != r {
		return scoped.UpdateGrokCatalogExtraIfObserved(ctx, id, expected, updates)
	}
	expectedArg, err := grokCatalogJSONArg(expected)
	if err != nil {
		return false, err
	}
	safe := make(map[string]any, 2)
	for _, key := range []string{service.UpstreamModelCatalogExtraKey, service.UpstreamModelMetadataExtraKey} {
		if value, ok := updates[key]; ok {
			safe[key] = value
		}
	}
	raw, err := json.Marshal(safe)
	if err != nil {
		return false, err
	}
	result, err := clientFromContext(ctx, r.client).ExecContext(ctx, updateGrokCatalogExtraIfObservedSQL, id, expectedArg, string(raw))
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return false, err
	}
	if dbent.TxFromContext(ctx) == nil {
		r.syncSchedulerAccountSnapshot(ctx, id)
	}
	return true, nil
}

func (r *accountRepository) CompareAndSwapGrokObservedModels(ctx context.Context, id string, expected, next json.RawMessage) (bool, error) {
	if scoped := r.accountTransactionScope(ctx); scoped != r {
		return scoped.CompareAndSwapGrokObservedModels(ctx, id, expected, next)
	}
	expectedArg, err := grokCatalogJSONArg(expected)
	if err != nil {
		return false, err
	}
	nextArg, err := grokCatalogJSONArg(next)
	if err != nil {
		return false, err
	}
	result, err := clientFromContext(ctx, r.client).ExecContext(ctx, compareAndSwapGrokObservedModelsSQL, id, expectedArg, nextArg)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected == 0 {
		return false, nil
	}
	if dbent.TxFromContext(ctx) == nil {
		r.syncSchedulerAccountSnapshot(ctx, id)
	}
	return true, nil
}

func grokCatalogJSONArg(raw json.RawMessage) (string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return "null", nil
	}
	if !json.Valid(trimmed) {
		return "", errGrokCatalogInvalidJSON
	}
	return string(trimmed), nil
}
