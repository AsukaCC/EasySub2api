package repository

import (
	"context"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/ent/announcementread"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

type announcementReadRepository struct {
	client *dbent.Client
}

func NewAnnouncementReadRepository(client *dbent.Client) service.AnnouncementReadRepository {
	return &announcementReadRepository{client: client}
}

func (r *announcementReadRepository) MarkRead(ctx context.Context, announcementID, userID string, readAt time.Time) error {
	client := clientFromContext(ctx, r.client)
	err := client.AnnouncementRead.Create().
		SetAnnouncementID(announcementID).
		SetUserID(userID).
		SetReadAt(readAt).
		OnConflictColumns(announcementread.FieldAnnouncementID, announcementread.FieldUserID).
		DoNothing().
		Exec(ctx)
	if isSQLNoRowsError(err) {
		return nil
	}
	return err
}

func (r *announcementReadRepository) GetReadMapByUser(ctx context.Context, userID string, announcementIDs []string) (map[string]time.Time, error) {
	if len(announcementIDs) == 0 {
		return map[string]time.Time{}, nil
	}

	rows, err := r.client.AnnouncementRead.Query().
		Where(
			announcementread.UserIDEQ(userID),
			announcementread.AnnouncementIDIn(announcementIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make(map[string]time.Time, len(rows))
	for i := range rows {
		out[rows[i].AnnouncementID] = rows[i].ReadAt
	}
	return out, nil
}

func (r *announcementReadRepository) GetReadMapByUsers(ctx context.Context, announcementID string, userIDs []string) (map[string]time.Time, error) {
	if len(userIDs) == 0 {
		return map[string]time.Time{}, nil
	}

	rows, err := r.client.AnnouncementRead.Query().
		Where(
			announcementread.AnnouncementIDEQ(announcementID),
			announcementread.UserIDIn(userIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make(map[string]time.Time, len(rows))
	for i := range rows {
		out[rows[i].UserID] = rows[i].ReadAt
	}
	return out, nil
}

func (r *announcementReadRepository) CountByAnnouncementID(ctx context.Context, announcementID string) (int64, error) {
	count, err := r.client.AnnouncementRead.Query().
		Where(announcementread.AnnouncementIDEQ(announcementID)).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}
