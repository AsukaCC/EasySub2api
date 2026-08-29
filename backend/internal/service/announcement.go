package service

import (
	"context"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/domain"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/pkg/pagination"
)

const (
	AnnouncementStatusDraft    = domain.AnnouncementStatusDraft
	AnnouncementStatusActive   = domain.AnnouncementStatusActive
	AnnouncementStatusArchived = domain.AnnouncementStatusArchived
)

const (
	AnnouncementNotifyModeSilent = domain.AnnouncementNotifyModeSilent
	AnnouncementNotifyModePopup  = domain.AnnouncementNotifyModePopup
	AnnouncementNotifyModeBanner = domain.AnnouncementNotifyModeBanner
)

const (
	AnnouncementConditionTypeSubscription = domain.AnnouncementConditionTypeSubscription
	AnnouncementConditionTypeBalance      = domain.AnnouncementConditionTypeBalance
	AnnouncementConditionTypeUser         = domain.AnnouncementConditionTypeUser
	AnnouncementConditionTypeLevel        = domain.AnnouncementConditionTypeLevel
)

const (
	AnnouncementOperatorIn  = domain.AnnouncementOperatorIn
	AnnouncementOperatorGT  = domain.AnnouncementOperatorGT
	AnnouncementOperatorGTE = domain.AnnouncementOperatorGTE
	AnnouncementOperatorLT  = domain.AnnouncementOperatorLT
	AnnouncementOperatorLTE = domain.AnnouncementOperatorLTE
	AnnouncementOperatorEQ  = domain.AnnouncementOperatorEQ
)

var (
	ErrAnnouncementNotFound        = domain.ErrAnnouncementNotFound
	ErrAnnouncementInvalidTarget   = domain.ErrAnnouncementInvalidTarget
	ErrAnnouncementNilInput        = infraerrors.BadRequest("ANNOUNCEMENT_INPUT_REQUIRED", "announcement input is required")
	ErrAnnouncementInvalidTitle    = infraerrors.BadRequest("ANNOUNCEMENT_TITLE_INVALID", "announcement title is invalid")
	ErrAnnouncementContentRequired = infraerrors.BadRequest(
		"ANNOUNCEMENT_CONTENT_REQUIRED",
		"announcement content is required",
	)
	ErrAnnouncementInvalidStatus     = infraerrors.BadRequest("ANNOUNCEMENT_STATUS_INVALID", "announcement status is invalid")
	ErrAnnouncementInvalidNotifyMode = infraerrors.BadRequest(
		"ANNOUNCEMENT_NOTIFY_MODE_INVALID",
		"announcement notify_mode is invalid",
	)
	ErrAnnouncementInvalidSchedule = infraerrors.BadRequest(
		"ANNOUNCEMENT_TIME_RANGE_INVALID",
		"starts_at must be before ends_at",
	)
)

type AnnouncementTargeting = domain.AnnouncementTargeting

type AnnouncementConditionGroup = domain.AnnouncementConditionGroup

type AnnouncementCondition = domain.AnnouncementCondition

type Announcement = domain.Announcement

type AnnouncementListFilters struct {
	Status string
	Search string
}

type AnnouncementRepository interface {
	Create(ctx context.Context, a *Announcement) error
	GetByID(ctx context.Context, id string) (*Announcement, error)
	Update(ctx context.Context, a *Announcement) error
	Delete(ctx context.Context, id string) error

	List(ctx context.Context, params pagination.PaginationParams, filters AnnouncementListFilters) ([]Announcement, *pagination.PaginationResult, error)
	ListActive(ctx context.Context, now time.Time) ([]Announcement, error)
}

type AnnouncementReadRepository interface {
	MarkRead(ctx context.Context, announcementID, userID string, readAt time.Time) error
	GetReadMapByUser(ctx context.Context, userID string, announcementIDs []string) (map[string]time.Time, error)
	GetReadMapByUsers(ctx context.Context, announcementID string, userIDs []string) (map[string]time.Time, error)
	CountByAnnouncementID(ctx context.Context, announcementID string) (int64, error)
}
