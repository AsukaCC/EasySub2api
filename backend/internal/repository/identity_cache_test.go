//go:build unit

package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFingerprintKey(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		expected  string
	}{
		{
			name:      "normal_account_id",
			accountID: "123",
			expected:  "fingerprint:123",
		},
		{
			name:      "zero_account_id",
			accountID: "0",
			expected:  "fingerprint:0",
		},
		{
			name:      "negative_account_id",
			accountID: "-1",
			expected:  "fingerprint:-1",
		},
		{
			name:      "max_int64",
			accountID: "9223372036854775807",
			expected:  "fingerprint:9223372036854775807",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := fingerprintKey(tc.accountID)
			require.Equal(t, tc.expected, got)
		})
	}
}
