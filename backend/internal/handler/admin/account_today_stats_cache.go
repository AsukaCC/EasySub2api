package admin

import (
	"strings"
	"time"
)

var accountTodayStatsBatchCache = newSnapshotCache(30 * time.Second)

func buildAccountTodayStatsBatchCacheKey(accountIDs []string) string {
	if len(accountIDs) == 0 {
		return "accounts_today_stats_empty"
	}
	var b strings.Builder
	b.Grow(len(accountIDs) * 6)
	_, _ = b.WriteString("accounts_today_stats:")
	for i, id := range accountIDs {
		if i > 0 {
			_ = b.WriteByte(',')
		}
		_, _ = b.WriteString(id)
	}
	return b.String()
}
