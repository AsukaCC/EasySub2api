package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelImageCacheReadPriceMigrationIsIdempotent(t *testing.T) {
	content, err := FS.ReadFile("275_channel_image_cache_read_price.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "ALTER TABLE channel_model_pricing")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS image_cache_read_price NUMERIC(20,12)")
}
