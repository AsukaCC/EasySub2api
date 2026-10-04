package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"time"
)

const UpstreamModelCatalogExtraKey = "upstream_model_catalog"

// The model list is independent of capability completeness and account allowlists.
type upstreamModelCatalogSnapshot struct {
	ConfigHash string               `json:"config_hash"`
	SyncedAt   time.Time            `json:"synced_at"`
	Catalog    UpstreamModelCatalog `json:"catalog"`
	Configured bool                 `json:"configured,omitempty"`
	Mapping    []string             `json:"mapping,omitempty"`
}

func upstreamModelCatalogConfigHash(account *Account) string {
	credentials := maps.Clone(account.Credentials)
	delete(credentials, "model_mapping")
	body, err := json.Marshal(struct {
		Platform    string
		Type        string
		Credentials map[string]any
	}{account.Platform, account.Type, credentials})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func cachedUpstreamModelCatalog(account *Account) *UpstreamModelCatalog {
	if account == nil {
		return nil
	}
	body, err := json.Marshal(account.Extra[UpstreamModelCatalogExtraKey])
	if err != nil {
		return nil
	}
	var snapshot upstreamModelCatalogSnapshot
	if json.Unmarshal(body, &snapshot) != nil || snapshot.ConfigHash == "" || snapshot.SyncedAt.IsZero() || snapshot.Catalog.Models == nil || snapshot.ConfigHash != upstreamModelCatalogConfigHash(account) {
		return nil
	}
	if snapshot.Configured && !slices.Equal(snapshot.Mapping, configuredUpstreamModelsForCapabilitySync(account)) {
		return nil
	}
	snapshot.Catalog.Models = dedupeAndSortModelIDs(snapshot.Catalog.Models)
	return &snapshot.Catalog
}

func (s *AccountTestService) saveUpstreamModelCatalog(ctx context.Context, account *Account, catalog *UpstreamModelCatalog, configured bool, metadata *UpstreamModelMetadataSnapshot) error {
	if account == nil || account.ID == "" || s.accountRepo == nil {
		return nil
	}
	snapshot := upstreamModelCatalogSnapshot{ConfigHash: upstreamModelCatalogConfigHash(account), SyncedAt: time.Now().UTC(), Catalog: *catalog}
	snapshot.Configured = configured
	if configured {
		snapshot.Mapping = configuredUpstreamModelsForCapabilitySync(account)
	}
	if snapshot.ConfigHash == "" {
		return newUpstreamModelSyncConfigError("Invalid model catalog configuration", nil)
	}
	updates := map[string]any{UpstreamModelCatalogExtraKey: snapshot}
	if metadata != nil {
		updates[UpstreamModelMetadataExtraKey] = *metadata
	}
	if err := s.accountRepo.UpdateExtra(ctx, account.ID, updates); err != nil {
		return newUpstreamModelSyncInternalError("Failed to save upstream model list", err)
	}
	return nil
}

// Explicit sync bypasses this cache; ordinary queries and test validation reuse it.
func (s *AccountTestService) accountModelCatalog(ctx context.Context, account *Account) (*UpstreamModelCatalog, error) {
	if s == nil || account == nil {
		return nil, newUpstreamModelSyncConfigError("Account model discovery is unavailable", nil)
	}
	if catalog := cachedUpstreamModelCatalog(account); catalog != nil {
		return catalog, nil
	}
	key := account.ID + ":" + upstreamModelCatalogConfigHash(account)
	result := s.upstreamModelCatalogLoads.DoChan(key, func() (any, error) {
		loadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 20*time.Second)
		defer cancel()
		if account.ID != "" && s.accountRepo != nil {
			current, err := s.accountRepo.GetByID(loadCtx, account.ID)
			if err != nil {
				return nil, err
			}
			if current == nil || upstreamModelCatalogConfigHash(current) != upstreamModelCatalogConfigHash(account) {
				return nil, newUpstreamModelSyncConfigError("Account configuration changed; reload the model list", nil)
			}
			if catalog := cachedUpstreamModelCatalog(current); catalog != nil {
				return catalog, nil
			}
		}
		requestAccount := account
		if account.IsCNProvider() && strings.TrimSpace(account.GetCredential("base_url")) != "" {
			requestAccount = DirectModelTestAccount(account)
			requestAccount.Credentials["api_protocol"] = APIProtocolChatCompletions
		}
		ids, body, err := s.fetchUpstreamModelList(loadCtx, requestAccount)
		configured := err != nil
		if err != nil {
			if !upstreamModelListEndpointUnsupported(err) {
				return nil, err
			}
			ids = configuredUpstreamModelsForCapabilitySync(account)
			if len(ids) == 0 {
				return nil, err
			}
		}
		_, metadata, _ := extractUpstreamModelCatalog(body, account.IsGrok())
		catalog := &UpstreamModelCatalog{Models: ids, Metadata: metadata}
		if err := s.saveUpstreamModelCatalog(loadCtx, account, catalog, configured, nil); err != nil {
			return nil, err
		}
		return catalog, nil
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case loaded := <-result:
		if loaded.Err != nil {
			return nil, loaded.Err
		}
		catalog, ok := loaded.Val.(*UpstreamModelCatalog)
		if !ok || catalog == nil {
			return nil, newUpstreamModelSyncInternalError("Invalid model catalog result", nil)
		}
		return catalog, nil
	}
}
