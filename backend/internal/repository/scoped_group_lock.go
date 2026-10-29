package repository

import (
	"context"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/lib/pq"
)

func lockLiveGroups(ctx context.Context, exec sqlExecutor, groupIDs []string) error {
	if len(groupIDs) == 0 {
		return nil
	}
	unique := make(map[string]struct{}, len(groupIDs))
	for _, id := range groupIDs {
		unique[id] = struct{}{}
	}
	rows, err := exec.QueryContext(ctx, `/* account_group_live_group_lock */
		SELECT id FROM groups
		WHERE id = ANY($1) AND deleted_at IS NULL
		ORDER BY id
		FOR SHARE`, pq.Array(groupIDs))
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	locked := 0
	for rows.Next() {
		locked++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if locked != len(unique) {
		return service.ErrGroupNotFound
	}
	return nil
}
