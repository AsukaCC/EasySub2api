package admin

import (
	"testing"
	"time"

	appTimezone "github.com/AsukaCC/EasySub2api/internal/pkg/timezone"
	"github.com/stretchr/testify/require"
)

func TestParseInactiveUserCutoffUsesServerTimezoneForDate(t *testing.T) {
	require.NoError(t, appTimezone.Init("Asia/Shanghai"))
	t.Cleanup(func() { _ = appTimezone.Init("UTC") })

	got, err := parseInactiveUserCutoff("2026-08-01")
	require.NoError(t, err)
	require.Equal(t, "Asia/Shanghai", got.Location().String())
	require.Equal(t, time.Date(2026, 8, 1, 0, 0, 0, 0, appTimezone.Location()), got)
	require.Equal(t, "2026-07-31T16:00:00Z", got.UTC().Format(time.RFC3339))
}

func TestParseInactiveUserCutoffKeepsRFC3339Compatibility(t *testing.T) {
	got, err := parseInactiveUserCutoff("2026-08-01T12:34:56+09:00")
	require.NoError(t, err)
	require.Equal(t, "2026-08-01T03:34:56Z", got.UTC().Format(time.RFC3339))
}

func TestParseInactiveUserCutoffRejectsInvalidInput(t *testing.T) {
	for _, value := range []string{"", "2026-02-30", "08/01/2026", "2026-08-01T25:00:00+08:00"} {
		_, err := parseInactiveUserCutoff(value)
		require.Error(t, err, value)
	}
}
