package service

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
)

type upstreamModelMetadataRepoStub struct {
	AccountRepository
	accountID string
	updates   map[string]any
	err       error
	account   *Account
	casCalls  int
}

func headerValuesEqualFold(header http.Header, name string) []string {
	for key, values := range header {
		if strings.EqualFold(key, name) {
			return values
		}
	}
	return nil
}

func (r *upstreamModelMetadataRepoStub) UpdateExtra(_ context.Context, id string, updates map[string]any) error {
	r.accountID = id
	r.updates = updates
	return r.err
}

func (r *upstreamModelMetadataRepoStub) GetByID(_ context.Context, id string) (*Account, error) {
	if r == nil || r.account == nil || r.account.ID != id {
		return nil, ErrAccountNotFound
	}
	return r.account, nil
}

func (r *upstreamModelMetadataRepoStub) CompareAndSwapGrokObservedModels(ctx context.Context, id string, expected, next json.RawMessage) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	r.casCalls++
	if r.account == nil || r.account.ID != id || r.account.Platform != PlatformGrok {
		return false, nil
	}
	var current any
	if r.account.Extra != nil {
		current = r.account.Extra[grokObservedModelsExtraKey]
	}
	currentJSON, err := json.Marshal(current)
	if err != nil {
		return false, err
	}
	if !jsonValuesEqual(currentJSON, expected) {
		return false, nil
	}
	var decoded any
	if err := json.Unmarshal(next, &decoded); err != nil {
		return false, err
	}
	if r.account.Extra == nil {
		r.account.Extra = map[string]any{}
	}
	r.account.Extra[grokObservedModelsExtraKey] = decoded
	return true, nil
}

func jsonValuesEqual(left, right []byte) bool {
	var lhs, rhs any
	if len(left) == 0 {
		left = []byte("null")
	}
	if len(right) == 0 {
		right = []byte("null")
	}
	if err := json.Unmarshal(left, &lhs); err != nil {
		return false
	}
	if err := json.Unmarshal(right, &rhs); err != nil {
		return false
	}
	return reflect.DeepEqual(lhs, rhs)
}
