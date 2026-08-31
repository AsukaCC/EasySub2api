package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/ent/group"
	"github.com/AsukaCC/EasySub2api/ent/pendingsubscription"
	"github.com/AsukaCC/EasySub2api/ent/user"
	"github.com/AsukaCC/EasySub2api/ent/usersubscription"
	"github.com/AsukaCC/EasySub2api/internal/payment"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

const (
	SubscriptionActivationActive  = "active"
	SubscriptionActivationPending = "pending"

	PendingSubscriptionStatusPending   = "PENDING"
	PendingSubscriptionStatusActivated = "ACTIVATED"
	PendingSubscriptionStatusCancelled = "CANCELLED"

	SubscriptionActivationModeScheduled = "scheduled"
	SubscriptionActivationModeImmediate = "immediate"
	SubscriptionActivationModeDirect    = "direct"
)

var (
	ErrSubscriptionPlatformPendingExists = infraerrors.Conflict(
		"SUBSCRIPTION_PLATFORM_PENDING_EXISTS",
		"a pending subscription already exists for this platform",
	)
	ErrSubscriptionActivationConfirmation = infraerrors.BadRequest(
		"CONFIRMATION_REQUIRED",
		"immediate activation requires explicit confirmation",
	)
	ErrSubscriptionActivationTargetUnavailable = infraerrors.Conflict(
		"SUBSCRIPTION_ACTIVATION_TARGET_UNAVAILABLE",
		"target subscription group is unavailable",
	)
)

type SubscriptionGrantInput struct {
	AssignSubscriptionInput
	SourceType string
	SourceID   string
}

type PendingSubscription struct {
	ID                       string     `json:"id"`
	UserID                   string     `json:"user_id"`
	GroupID                  string     `json:"group_id"`
	Platform                 string     `json:"platform"`
	ValidityDays             int        `json:"validity_days"`
	SourceType               string     `json:"source_type"`
	SourceID                 string     `json:"source_id,omitempty"`
	BlockedBySubscriptionID  *string    `json:"blocked_by_subscription_id,omitempty"`
	ExpectedActivationAt     *time.Time `json:"expected_activation_at,omitempty"`
	Status                   string     `json:"status"`
	ActivatedSubscriptionID  *string    `json:"activated_subscription_id,omitempty"`
	ActivationMode           string     `json:"activation_mode,omitempty"`
	ForfeitedSubscriptionIDs []string   `json:"forfeited_subscription_ids,omitempty"`
	ActivatedAt              *time.Time `json:"activated_at,omitempty"`
	LastError                string     `json:"last_error,omitempty"`
	Notes                    string     `json:"notes,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	Group                    *Group     `json:"group,omitempty"`
}

type SubscriptionGrantResult struct {
	ActivationStatus    string               `json:"activation_status"`
	Subscription        *UserSubscription    `json:"subscription,omitempty"`
	PendingSubscription *PendingSubscription `json:"pending_subscription,omitempty"`
}

func pendingSubscriptionFromEnt(p *dbent.PendingSubscription) *PendingSubscription {
	if p == nil {
		return nil
	}
	return &PendingSubscription{
		ID: p.ID, UserID: p.UserID, GroupID: p.GroupID, Platform: p.Platform,
		ValidityDays: p.ValidityDays, SourceType: p.SourceType, SourceID: p.SourceID,
		BlockedBySubscriptionID: p.BlockedBySubscriptionID,
		ExpectedActivationAt:    p.ExpectedActivationAt, Status: p.Status,
		ActivatedSubscriptionID: p.ActivatedSubscriptionID, ActivationMode: p.ActivationMode,
		ForfeitedSubscriptionIDs: p.ForfeitedSubscriptionIds, ActivatedAt: p.ActivatedAt,
		LastError: p.LastError, Notes: p.Notes, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

// GrantOrQueueSubscription is the single entry point for every new subscription
// entitlement. It serializes grants per user and queues a second entitlement on
// the same platform instead of extending the current term.
func (s *SubscriptionService) GrantOrQueueSubscription(ctx context.Context, input *SubscriptionGrantInput) (*SubscriptionGrantResult, error) {
	if input == nil {
		return nil, ErrSubscriptionNilInput
	}
	if strings.TrimSpace(input.UserID) == "" || strings.TrimSpace(input.GroupID) == "" {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "user and group are required")
	}
	if input.ValidityDays <= 0 {
		input.ValidityDays = 30
	}
	if input.ValidityDays > MaxValidityDays {
		input.ValidityDays = MaxValidityDays
	}

	if existingTx := dbent.TxFromContext(ctx); existingTx != nil {
		return s.grantOrQueueSubscriptionTx(ctx, existingTx.Client(), input)
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin subscription grant transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	result, err := s.grantOrQueueSubscriptionTx(txCtx, tx.Client(), input)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit subscription grant transaction: %w", err)
	}
	if result.Subscription != nil {
		_ = s.invalidateSubscriptionCaches(input.UserID, input.GroupID)
	}
	return result, nil
}

func (s *SubscriptionService) grantOrQueueSubscriptionTx(ctx context.Context, client *dbent.Client, input *SubscriptionGrantInput) (*SubscriptionGrantResult, error) {
	grp, err := s.groupRepo.GetByID(ctx, input.GroupID)
	if err != nil || grp == nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}
	if !grp.IsSubscriptionType() {
		return nil, ErrGroupNotSubscriptionType
	}
	if grp.Status != payment.EntityStatusActive {
		return nil, infraerrors.Conflict("SUBSCRIPTION_GROUP_INACTIVE", "subscription group is inactive")
	}
	platform := strings.ToLower(strings.TrimSpace(grp.Platform))
	if platform == "" {
		return nil, infraerrors.BadRequest("SUBSCRIPTION_PLATFORM_REQUIRED", "subscription group platform is required")
	}
	if _, err := client.User.Query().Where(user.IDEQ(input.UserID)).ForUpdate().Only(ctx); err != nil {
		return nil, fmt.Errorf("lock subscription user: %w", err)
	}

	if sourceID := strings.TrimSpace(input.SourceID); sourceID != "" {
		existing, err := client.PendingSubscription.Query().Where(
			pendingsubscription.SourceTypeEQ(strings.TrimSpace(input.SourceType)),
			pendingsubscription.SourceIDEQ(sourceID),
		).Only(ctx)
		if err == nil {
			switch existing.Status {
			case PendingSubscriptionStatusPending:
				return &SubscriptionGrantResult{ActivationStatus: SubscriptionActivationPending, PendingSubscription: pendingSubscriptionFromEnt(existing)}, nil
			case PendingSubscriptionStatusActivated:
				if existing.ActivatedSubscriptionID != nil {
					sub, subErr := s.userSubRepo.GetByID(ctx, *existing.ActivatedSubscriptionID)
					if subErr == nil {
						return &SubscriptionGrantResult{ActivationStatus: SubscriptionActivationActive, Subscription: sub, PendingSubscription: pendingSubscriptionFromEnt(existing)}, nil
					}
				}
				return nil, infraerrors.Conflict("SUBSCRIPTION_SOURCE_ALREADY_USED", "subscription grant source has already been activated")
			default:
				return nil, infraerrors.Conflict("SUBSCRIPTION_SOURCE_ALREADY_USED", "subscription grant source has already been used")
			}
		}
		if !dbent.IsNotFound(err) {
			return nil, fmt.Errorf("query subscription grant source: %w", err)
		}
	}

	now := time.Now().UTC()
	blockers, err := currentPlatformSubscriptions(ctx, client, input.UserID, platform, now, true)
	if err != nil {
		return nil, err
	}
	if len(blockers) == 0 {
		sub, _, err := s.assignOrExtendSubscription(ctx, &input.AssignSubscriptionInput, true)
		if err != nil {
			return nil, err
		}
		result := &SubscriptionGrantResult{ActivationStatus: SubscriptionActivationActive, Subscription: sub}
		if sourceID := strings.TrimSpace(input.SourceID); sourceID != "" {
			now := time.Now().UTC()
			recordBuilder := client.PendingSubscription.Create().
				SetUserID(input.UserID).
				SetGroupID(input.GroupID).
				SetPlatform(platform).
				SetValidityDays(input.ValidityDays).
				SetSourceType(firstNonEmpty(strings.TrimSpace(input.SourceType), "manual")).
				SetSourceID(sourceID).
				SetStatus(PendingSubscriptionStatusActivated).
				SetActivatedSubscriptionID(sub.ID).
				SetActivationMode(SubscriptionActivationModeDirect).
				SetActivatedAt(now).
				SetNotes(input.Notes)
			if strings.TrimSpace(input.AssignedBy) != "" {
				recordBuilder.SetAssignedBy(input.AssignedBy)
			}
			record, recordErr := recordBuilder.Save(ctx)
			if recordErr != nil {
				return nil, fmt.Errorf("record direct subscription grant: %w", recordErr)
			}
			result.PendingSubscription = pendingSubscriptionFromEnt(record)
		}
		return result, nil
	}

	exists, err := client.PendingSubscription.Query().Where(
		pendingsubscription.UserIDEQ(input.UserID),
		pendingsubscription.PlatformEQ(platform),
		pendingsubscription.StatusEQ(PendingSubscriptionStatusPending),
	).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("query pending subscription: %w", err)
	}
	if exists {
		return nil, ErrSubscriptionPlatformPendingExists
	}
	expected := blockers[0].ExpiresAt
	blockedBy := blockers[0].ID
	for _, blocker := range blockers[1:] {
		if blocker.ExpiresAt.After(expected) {
			expected = blocker.ExpiresAt
			blockedBy = blocker.ID
		}
	}
	builder := client.PendingSubscription.Create().
		SetUserID(input.UserID).
		SetGroupID(input.GroupID).
		SetPlatform(platform).
		SetValidityDays(input.ValidityDays).
		SetSourceType(firstNonEmpty(strings.TrimSpace(input.SourceType), "manual")).
		SetSourceID(strings.TrimSpace(input.SourceID)).
		SetBlockedBySubscriptionID(blockedBy).
		SetExpectedActivationAt(expected).
		SetStatus(PendingSubscriptionStatusPending).
		SetNotes(input.Notes)
	if strings.TrimSpace(input.AssignedBy) != "" {
		builder.SetAssignedBy(input.AssignedBy)
	}
	pending, err := builder.Save(ctx)
	if err != nil {
		if dbent.IsConstraintError(err) {
			return nil, ErrSubscriptionPlatformPendingExists
		}
		return nil, fmt.Errorf("create pending subscription: %w", err)
	}
	return &SubscriptionGrantResult{ActivationStatus: SubscriptionActivationPending, PendingSubscription: pendingSubscriptionFromEnt(pending)}, nil
}

func currentPlatformSubscriptions(ctx context.Context, client *dbent.Client, userID, platform string, now time.Time, lock bool) ([]*dbent.UserSubscription, error) {
	query := client.UserSubscription.Query().Where(
		usersubscription.UserIDEQ(userID),
		usersubscription.StatusIn(SubscriptionStatusActive, SubscriptionStatusSuspended),
		usersubscription.ExpiresAtGT(now),
		usersubscription.HasGroupWith(group.PlatformEqualFold(platform)),
	)
	if lock {
		query.ForUpdate()
	}
	items, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query current platform subscriptions: %w", err)
	}
	return items, nil
}

func (s *SubscriptionService) ListPendingSubscriptions(ctx context.Context, userID string) ([]PendingSubscription, error) {
	items, err := s.entClient.PendingSubscription.Query().Where(
		pendingsubscription.UserIDEQ(userID),
		pendingsubscription.StatusEQ(PendingSubscriptionStatusPending),
	).Order(dbent.Asc(pendingsubscription.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pending subscriptions: %w", err)
	}
	result := make([]PendingSubscription, 0, len(items))
	for _, item := range items {
		mapped := pendingSubscriptionFromEnt(item)
		if grp, groupErr := s.groupRepo.GetByID(ctx, item.GroupID); groupErr == nil {
			mapped.Group = grp
		}
		result = append(result, *mapped)
	}
	return result, nil
}

func (s *SubscriptionService) ListPendingSubscriptionsAdmin(ctx context.Context, userID, platform, groupID string) ([]PendingSubscription, error) {
	query := s.entClient.PendingSubscription.Query().Where(
		pendingsubscription.StatusEQ(PendingSubscriptionStatusPending),
	)
	if strings.TrimSpace(userID) != "" {
		query.Where(pendingsubscription.UserIDEQ(strings.TrimSpace(userID)))
	}
	if strings.TrimSpace(platform) != "" {
		query.Where(pendingsubscription.PlatformEqualFold(strings.TrimSpace(platform)))
	}
	if strings.TrimSpace(groupID) != "" {
		query.Where(pendingsubscription.GroupIDEQ(strings.TrimSpace(groupID)))
	}
	items, err := query.Order(dbent.Asc(pendingsubscription.FieldExpectedActivationAt)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list admin pending subscriptions: %w", err)
	}
	result := make([]PendingSubscription, 0, len(items))
	for _, item := range items {
		mapped := pendingSubscriptionFromEnt(item)
		if grp, groupErr := s.groupRepo.GetByID(ctx, item.GroupID); groupErr == nil {
			mapped.Group = grp
		}
		result = append(result, *mapped)
	}
	return result, nil
}

func (s *SubscriptionService) ActivatePendingNow(ctx context.Context, userID, pendingID string, confirmed bool) (*SubscriptionGrantResult, error) {
	if !confirmed {
		return nil, ErrSubscriptionActivationConfirmation
	}
	return s.activatePending(ctx, pendingID, userID, true)
}

func (s *SubscriptionService) activatePending(ctx context.Context, pendingID, ownerID string, immediate bool) (*SubscriptionGrantResult, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin pending activation transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	result, cacheGroupIDs, err := s.activatePendingTx(txCtx, tx.Client(), pendingID, ownerID, immediate)
	if err != nil {
		if err == ErrSubscriptionActivationTargetUnavailable {
			if commitErr := tx.Commit(); commitErr != nil {
				return nil, fmt.Errorf("persist pending activation error: %w", commitErr)
			}
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit pending subscription activation: %w", err)
	}
	if result != nil && result.Subscription != nil {
		for _, groupID := range cacheGroupIDs {
			_ = s.invalidateSubscriptionCaches(result.Subscription.UserID, groupID)
		}
	}
	return result, nil
}

func (s *SubscriptionService) activatePendingTx(ctx context.Context, client *dbent.Client, pendingID, ownerID string, immediate bool) (*SubscriptionGrantResult, []string, error) {
	preview, err := client.PendingSubscription.Query().Where(pendingsubscription.IDEQ(pendingID)).Only(ctx)
	if err != nil {
		return nil, nil, infraerrors.NotFound("PENDING_SUBSCRIPTION_NOT_FOUND", "pending subscription not found")
	}
	if ownerID != "" && preview.UserID != ownerID {
		return nil, nil, infraerrors.Forbidden("FORBIDDEN", "no permission for this pending subscription")
	}
	if _, err := client.User.Query().Where(user.IDEQ(preview.UserID)).ForUpdate().Only(ctx); err != nil {
		return nil, nil, fmt.Errorf("lock pending subscription user: %w", err)
	}
	pending, err := client.PendingSubscription.Query().Where(pendingsubscription.IDEQ(pendingID)).ForUpdate().Only(ctx)
	if err != nil {
		return nil, nil, infraerrors.NotFound("PENDING_SUBSCRIPTION_NOT_FOUND", "pending subscription not found")
	}
	if ownerID != "" && pending.UserID != ownerID {
		return nil, nil, infraerrors.Forbidden("FORBIDDEN", "no permission for this pending subscription")
	}
	if pending.Status != PendingSubscriptionStatusPending {
		return nil, nil, infraerrors.Conflict("PENDING_SUBSCRIPTION_NOT_PENDING", "subscription is no longer pending")
	}
	grp, err := s.groupRepo.GetByID(ctx, pending.GroupID)
	if err != nil || grp == nil || !grp.IsSubscriptionType() || grp.Status != payment.EntityStatusActive {
		message := "target subscription group is unavailable"
		_, _ = client.PendingSubscription.UpdateOneID(pending.ID).
			SetLastError(message).
			SetExpectedActivationAt(time.Now().UTC().Add(5 * time.Minute)).
			Save(ctx)
		return nil, nil, ErrSubscriptionActivationTargetUnavailable
	}
	now := time.Now().UTC()
	blockers, err := currentPlatformSubscriptions(ctx, client, pending.UserID, pending.Platform, now, true)
	if err != nil {
		return nil, nil, err
	}
	if !immediate && len(blockers) > 0 {
		expected := blockers[0].ExpiresAt
		blockedBy := blockers[0].ID
		for _, blocker := range blockers[1:] {
			if blocker.ExpiresAt.After(expected) {
				expected, blockedBy = blocker.ExpiresAt, blocker.ID
			}
		}
		_, err = client.PendingSubscription.UpdateOneID(pending.ID).
			SetExpectedActivationAt(expected).SetBlockedBySubscriptionID(blockedBy).SetLastError("").Save(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("reschedule pending subscription: %w", err)
		}
		pending.ExpectedActivationAt = &expected
		pending.BlockedBySubscriptionID = &blockedBy
		return &SubscriptionGrantResult{ActivationStatus: SubscriptionActivationPending, PendingSubscription: pendingSubscriptionFromEnt(pending)}, nil, nil
	}

	forfeited := make([]string, 0, len(blockers))
	if immediate {
		for _, blocker := range blockers {
			forfeited = append(forfeited, blocker.ID)
			if _, err := client.UserSubscription.UpdateOneID(blocker.ID).
				SetStatus(SubscriptionStatusExpired).SetExpiresAt(now).Save(ctx); err != nil {
				return nil, nil, fmt.Errorf("expire current subscription: %w", err)
			}
		}
	}
	grantInput := &AssignSubscriptionInput{
		UserID: pending.UserID, GroupID: pending.GroupID, ValidityDays: pending.ValidityDays,
		Notes: pending.Notes,
	}
	if pending.AssignedBy != nil {
		grantInput.AssignedBy = *pending.AssignedBy
	}
	sub, _, err := s.assignOrExtendSubscription(ctx, grantInput, true)
	if err != nil {
		return nil, nil, err
	}
	mode := SubscriptionActivationModeScheduled
	if immediate {
		mode = SubscriptionActivationModeImmediate
	}
	updated, err := client.PendingSubscription.UpdateOneID(pending.ID).
		SetStatus(PendingSubscriptionStatusActivated).
		SetActivatedSubscriptionID(sub.ID).
		SetActivationMode(mode).
		SetForfeitedSubscriptionIds(forfeited).
		SetActivatedAt(now).
		SetLastError("").
		Save(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("complete pending subscription activation: %w", err)
	}
	cacheGroupIDs := make([]string, 0, len(blockers)+1)
	for _, blocker := range blockers {
		cacheGroupIDs = append(cacheGroupIDs, blocker.GroupID)
	}
	cacheGroupIDs = append(cacheGroupIDs, pending.GroupID)
	return &SubscriptionGrantResult{ActivationStatus: SubscriptionActivationActive, Subscription: sub, PendingSubscription: pendingSubscriptionFromEnt(updated)}, cacheGroupIDs, nil
}

func (s *SubscriptionService) ActivateDuePendingSubscriptions(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 100
	}
	activated := 0
	for processed := 0; processed < limit; processed++ {
		tx, err := s.entClient.Tx(ctx)
		if err != nil {
			return activated, fmt.Errorf("begin pending activation scan: %w", err)
		}
		txCtx := dbent.NewTxContext(ctx, tx)
		item, err := tx.PendingSubscription.Query().Where(
			pendingsubscription.StatusEQ(PendingSubscriptionStatusPending),
			pendingsubscription.Or(
				pendingsubscription.ExpectedActivationAtIsNil(),
				pendingsubscription.ExpectedActivationAtLTE(time.Now().UTC()),
			),
		).Order(dbent.Asc(pendingsubscription.FieldExpectedActivationAt)).
			ForUpdate(entsql.WithLockAction(entsql.SkipLocked)).First(txCtx)
		if dbent.IsNotFound(err) {
			_ = tx.Rollback()
			break
		}
		if err != nil {
			_ = tx.Rollback()
			return activated, fmt.Errorf("scan pending subscriptions: %w", err)
		}
		result, cacheGroupIDs, activationErr := s.activatePendingTx(txCtx, tx.Client(), item.ID, "", false)
		if activationErr != nil {
			if activationErr == ErrSubscriptionActivationTargetUnavailable {
				_ = tx.Commit()
			} else {
				_ = tx.Rollback()
				_, _ = s.entClient.PendingSubscription.UpdateOneID(item.ID).
					SetLastError(activationErr.Error()).
					SetExpectedActivationAt(time.Now().UTC().Add(time.Minute)).
					Save(ctx)
			}
			continue
		}
		if err := tx.Commit(); err != nil {
			return activated, fmt.Errorf("commit pending subscription activation: %w", err)
		}
		if result != nil && result.ActivationStatus == SubscriptionActivationActive {
			activated++
			for _, groupID := range cacheGroupIDs {
				_ = s.invalidateSubscriptionCaches(item.UserID, groupID)
			}
		}
	}
	return activated, nil
}
