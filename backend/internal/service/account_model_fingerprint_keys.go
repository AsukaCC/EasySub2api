package service

import (
	"context"
	"errors"
	"slices"
	"strings"

	apperrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/pkg/pagination"
)

type ModelFingerprintKey struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *AccountTestService) fingerprintKey(ctx context.Context, account *Account, userID, keyID string) (*APIKey, error) {
	if strings.TrimSpace(keyID) == "" {
		return nil, apperrors.BadRequest("FINGERPRINT_API_KEY_REQUIRED", "Select an API key")
	}
	if s.modelFingerprintKeys == nil || s.modelFingerprintGroups == nil {
		return nil, errors.New("fingerprint API key lookup unavailable")
	}
	key, err := s.modelFingerprintKeys.GetByID(ctx, keyID)
	if err != nil {
		return nil, err
	}
	if key == nil || key.UserID != userID {
		return nil, apperrors.Forbidden("FINGERPRINT_API_KEY_FORBIDDEN", "API key does not belong to the current user")
	}
	if !key.IsActive() || key.IsExpired() || key.IsQuotaExhausted() {
		return nil, apperrors.BadRequest("FINGERPRINT_API_KEY_INACTIVE", "API key is disabled, expired or exhausted")
	}
	groupID, err := s.fingerprintKeyGroup(ctx, account, key)
	if err != nil {
		return nil, err
	}
	copy := *key
	copy.GroupID = &groupID
	return &copy, nil
}

func (s *AccountTestService) fingerprintKeyGroup(ctx context.Context, account *Account, key *APIKey) (string, error) {
	for _, id := range key.BoundGroupIDs() {
		if !slices.Contains(account.GroupIDs, id) {
			continue
		}
		group, err := s.modelFingerprintGroups.GetByID(ctx, id)
		if err != nil {
			return "", err
		}
		if group != nil && group.Status == StatusActive && (group.Platform == account.Platform || group.Platform == PlatformComposite) {
			return id, nil
		}
	}
	return "", apperrors.BadRequest("FINGERPRINT_API_KEY_PLATFORM", "API key has no active group containing this account and platform")
}

// The selection response includes identifiers and names, never credential values.
func (s *AccountTestService) ListFingerprintKeys(ctx context.Context, accountID, userID, search string) ([]ModelFingerprintKey, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if s.modelFingerprintKeys == nil || s.modelFingerprintGroups == nil || userID == "" {
		return nil, errors.New("fingerprint API key lookup unavailable")
	}
	result := []ModelFingerprintKey{}
	for page := 1; ; page++ {
		keys, paging, err := s.modelFingerprintKeys.ListByUserID(ctx, userID, pagination.PaginationParams{Page: page, PageSize: 100}, APIKeyListFilters{Search: search, Status: StatusActive})
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			if key.UserID != userID || !key.IsActive() || key.IsExpired() || key.IsQuotaExhausted() {
				continue
			}
			if _, err := s.fingerprintKeyGroup(ctx, account, &key); err != nil {
				if apperrors.Reason(err) == "FINGERPRINT_API_KEY_PLATFORM" {
					continue
				}
				return nil, err
			}
			result = append(result, ModelFingerprintKey{ID: key.ID, Name: key.Name})
		}
		if len(keys) < 100 || paging == nil || page >= paging.Pages {
			return result, nil
		}
	}
}

func (s *AccountTestService) GetFingerprintModelsForKey(ctx context.Context, accountID, userID, keyID string) ([]ModelFingerprintModel, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if _, err := s.fingerprintKey(ctx, account, userID, keyID); err != nil {
		return nil, err
	}
	return s.fingerprintModels(ctx, account)
}
