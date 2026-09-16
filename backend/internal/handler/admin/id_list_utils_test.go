//go:build unit

package admin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeInt64IDList(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"nil input", nil, nil},
		{"empty input", []string{}, nil},
		{"single element", []string{"id-5"}, []string{"id-5"}},
		{"already sorted unique", []string{"id-1", "id-2", "id-3"}, []string{"id-1", "id-2", "id-3"}},
		{"duplicates removed", []string{"id-3", "id-1", "id-3", "id-2", "id-1"}, []string{"id-1", "id-2", "id-3"}},
		{"empty filtered", []string{"", "id-1", "id-2"}, []string{"id-1", "id-2"}},
		{"all invalid", []string{"", ""}, []string{}},
		{"sorted output", []string{"id-9", "id-3", "id-7", "id-1"}, []string{"id-1", "id-3", "id-7", "id-9"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeInt64IDList(tc.in)
			if tc.want == nil {
				require.Nil(t, got)
			} else {
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestBuildAccountTodayStatsBatchCacheKey(t *testing.T) {
	tests := []struct {
		name string
		ids  []string
		want string
	}{
		{"empty", nil, "accounts_today_stats_empty"},
		{"single", []string{"account-42"}, "accounts_today_stats:account-42"},
		{"multiple", []string{"account-1", "account-2", "account-3"}, "accounts_today_stats:account-1,account-2,account-3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildAccountTodayStatsBatchCacheKey(tc.ids)
			require.Equal(t, tc.want, got)
		})
	}
}
