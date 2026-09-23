package migrations

import (
	"strings"
	"testing"
)

func TestWalletBalanceBucketSplitMigrationPreservesBuckets(t *testing.T) {
	sql, err := FS.ReadFile("280_wallet_balance_bucket_split.sql")
	if err != nil {
		t.Fatal(err)
	}
	text := string(sql)
	for _, want := range []string{
		"ADD COLUMN IF NOT EXISTS recharge_balance",
		"ADD COLUMN IF NOT EXISTS frozen_recharge_balance",
		"recharge_balance = balance - bonus_balance",
		"frozen_recharge_balance = frozen_balance - frozen_bonus_balance",
		"DROP COLUMN IF EXISTS balance",
		"DROP COLUMN IF EXISTS frozen_balance",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("migration missing %q", want)
		}
	}
}
