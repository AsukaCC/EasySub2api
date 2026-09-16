package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortedUniqueAccountIDs(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{name: "unsorted duplicates", input: []string{"account-12", "account-3", "account-12", "account-8", "account-3"}, want: []string{"account-12", "account-3", "account-8"}},
		{name: "already sorted", input: []string{"account-3", "account-8", "account-12"}, want: []string{"account-12", "account-3", "account-8"}},
		{name: "single", input: []string{"account-3"}, want: []string{"account-3"}},
		{name: "empty", input: []string{}, want: []string{}},
		{name: "nil", input: nil, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, sortedUniqueAccountIDs(tt.input))
		})
	}
}
