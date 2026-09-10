package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/ent/group"
	"github.com/AsukaCC/EasySub2api/ent/schema/mixins"
	"github.com/AsukaCC/EasySub2api/ent/subscriptionresetcard"
	"github.com/AsukaCC/EasySub2api/ent/usersubscription"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/google/uuid"
)

const (
	ResetCardStatusAvailable = "available"
	ResetCardStatusConsumed  = "consumed"
	ResetCardStatusExpired   = "expired"
	ResetCardStatusRevoked   = "revoked"

	DefaultResetCardValidityDays = 30
	MaxResetCardIssueQuantity    = 1000
	MaxResetCardValidityDays     = 3650
)

func normalizeResetCardValidityDays(days int) int {
	if days <= 0 {
		return DefaultResetCardValidityDays
	}
	if days > MaxResetCardValidityDays {
		return MaxResetCardValidityDays
	}
	return days
}

var (
	ErrResetCardNotAvailable = infraerrors.Conflict(
		"RESET_CARD_NOT_AVAILABLE", "no usable reset card is available for this subscription",
	)
	ErrResetCardSubscriptionInactive = infraerrors.Forbidden(
		"SUBSCRIPTION_INACTIVE", "subscription is not active",
	)
	ErrWeeklyLimitNotConfigured = infraerrors.BadRequest(
		"WEEKLY_LIMIT_NOT_CONFIGURED", "subscription group has no weekly limit configured",
	)
	ErrResetCardForbidden = infraerrors.Forbidden(
		"RESET_CARD_FORBIDDEN", "reset card cannot be used for this subscription",
	)
	ErrResetCardInvalidInput = infraerrors.BadRequest(
		"RESET_CARD_INVALID_INPUT", "invalid reset card request",
	)
	ErrResetCardServiceUnavailable = infraerrors.InternalServer(
		"RESET_CARD_SERVICE_UNAVAILABLE", "reset card service is unavailable",
	)
)

// ResetCardExpiry groups available cards by their expiration timestamp. The
// timestamp is kept at full precision so the UI can accurately show batches
// issued at different times.
type ResetCardExpiry struct {
	ExpiresAt time.Time `json:"expires_at"`
	Count     int       `json:"count"`
}

type ResetCardSummary struct {
	AvailableCount  int               `json:"available_count"`
	ExpiredCount    int               `json:"expired_count"`
	ConsumedCount   int               `json:"consumed_count"`
	NextExpiryAt    *time.Time        `json:"next_expiry_at"`
	ExpiryBreakdown []ResetCardExpiry `json:"expiry_breakdown"`
}

type ResetCardIssueItem struct {
	SubscriptionID string `json:"subscription_id"`
	IssuedCount    int    `json:"issued_count"`
	Success        bool   `json:"success"`
	ErrorCode      string `json:"error_code,omitempty"`
	Error          string `json:"error,omitempty"`
}

type ResetCardIssueResult struct {
	Items        []ResetCardIssueItem `json:"items"`
	SuccessCount int                  `json:"success_count"`
	FailedCount  int                  `json:"failed_count"`
	TotalIssued  int                  `json:"total_issued"`
}

type WeeklyResetItem struct {
	SubscriptionID  string     `json:"subscription_id"`
	Success         bool       `json:"success"`
	ResetAt         *time.Time `json:"reset_at,omitempty"`
	WeeklyWindowEnd *time.Time `json:"weekly_window_end,omitempty"`
	ErrorCode       string     `json:"error_code,omitempty"`
	Error           string     `json:"error,omitempty"`
}

type WeeklyResetResult struct {
	Items        []WeeklyResetItem `json:"items"`
	SuccessCount int               `json:"success_count"`
	FailedCount  int               `json:"failed_count"`
}

type ResetCardConsumeResult struct {
	Subscription    *UserSubscription `json:"subscription"`
	ResetAt         time.Time         `json:"reset_at"`
	WeeklyWindowEnd time.Time         `json:"weekly_window_end"`
	ResetCards      ResetCardSummary  `json:"reset_cards"`
}

func subscriptionCardClient(ctx context.Context, fallback *dbent.Client) *dbent.Client {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	return fallback
}

func (s *SubscriptionService) resetCardNow() time.Time {
	if s != nil && s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}

// attachResetCardSummaries enriches subscription responses in one query. A
// service without an Ent client (for example, an existing repository-only unit
// test) gets an empty ledger; production startup applies the card migration
// before these reads are served.
func (s *SubscriptionService) attachResetCardSummaries(ctx context.Context, subs []UserSubscription) error {
	if s == nil || len(subs) == 0 {
		return nil
	}
	cardClient := subscriptionCardClient(ctx, s.entClient)
	if cardClient == nil {
		return nil
	}
	ids := make([]string, 0, len(subs))
	seen := make(map[string]struct{}, len(subs))
	for _, sub := range subs {
		if sub.ID == "" {
			continue
		}
		if _, ok := seen[sub.ID]; ok {
			continue
		}
		seen[sub.ID] = struct{}{}
		ids = append(ids, sub.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	cards, err := cardClient.SubscriptionResetCard.Query().
		Where(subscriptionresetcard.SubscriptionIDIn(ids...)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query subscription reset cards: %w", err)
	}

	now := s.resetCardNow()
	type aggregate struct {
		ResetCardSummary
		byExpiry map[string]ResetCardExpiry
	}
	type subscriptionBinding struct {
		userID  string
		groupID string
	}
	aggs := make(map[string]*aggregate, len(ids))
	bindings := make(map[string]subscriptionBinding, len(ids))
	for _, id := range ids {
		aggs[id] = &aggregate{byExpiry: make(map[string]ResetCardExpiry)}
	}
	for _, sub := range subs {
		if _, ok := aggs[sub.ID]; !ok {
			continue
		}
		bindings[sub.ID] = subscriptionBinding{userID: sub.UserID, groupID: sub.GroupID}
	}
	for _, card := range cards {
		agg := aggs[card.SubscriptionID]
		if agg == nil {
			continue
		}
		binding := bindings[card.SubscriptionID]
		// user_id and group_id are denormalized on the ledger for audit and
		// transfer-prevention checks.  Ignore a corrupt/stale row rather than
		// exposing it in a summary for a different subscription binding.
		if binding.userID != "" && card.UserID != binding.userID {
			continue
		}
		if binding.groupID != "" && card.GroupID != binding.groupID {
			continue
		}
		switch card.Status {
		case ResetCardStatusConsumed:
			agg.ConsumedCount++
		case ResetCardStatusExpired, ResetCardStatusRevoked:
			if card.Status == ResetCardStatusExpired {
				agg.ExpiredCount++
			}
		case ResetCardStatusAvailable:
			if !card.ExpiresAt.After(now) {
				agg.ExpiredCount++
				continue
			}
			agg.AvailableCount++
			if agg.NextExpiryAt == nil || card.ExpiresAt.Before(*agg.NextExpiryAt) {
				exp := card.ExpiresAt
				agg.NextExpiryAt = &exp
			}
			key := card.ExpiresAt.UTC().Format(time.RFC3339Nano)
			item := agg.byExpiry[key]
			item.ExpiresAt = card.ExpiresAt
			item.Count++
			agg.byExpiry[key] = item
		}
	}
	for i := range subs {
		if agg := aggs[subs[i].ID]; agg != nil {
			summary := agg.ResetCardSummary
			summary.ExpiryBreakdown = make([]ResetCardExpiry, 0, len(agg.byExpiry))
			for _, item := range agg.byExpiry {
				summary.ExpiryBreakdown = append(summary.ExpiryBreakdown, item)
			}
			sort.Slice(summary.ExpiryBreakdown, func(i, j int) bool {
				return summary.ExpiryBreakdown[i].ExpiresAt.Before(summary.ExpiryBreakdown[j].ExpiresAt)
			})
			subs[i].ResetCards = summary
		}
	}
	return nil
}

func (s *SubscriptionService) attachResetCardSummary(ctx context.Context, sub *UserSubscription) error {
	if sub == nil {
		return nil
	}
	items := []UserSubscription{*sub}
	if err := s.attachResetCardSummaries(ctx, items); err != nil {
		return err
	}
	sub.ResetCards = items[0].ResetCards
	return nil
}

func (s *SubscriptionService) subscriptionHasWeeklyLimit(ctx context.Context, client *dbent.Client, sub *dbent.UserSubscription) (bool, error) {
	if sub == nil {
		return false, ErrSubscriptionNotFound
	}
	// Prefer the transaction client when this operation is running inside an
	// Ent transaction. The reset operation locks the subscription and must
	// validate the group from the same snapshot; using the repository first
	// would bypass that transaction in production. Outside a transaction,
	// retain the repository dependency so lightweight service stubs remain
	// usable in callers that do not configure an Ent client.
	if client != nil && dbent.TxFromContext(ctx) != nil {
		grp, err := client.Group.Query().Where(group.IDEQ(sub.GroupID)).Only(ctx)
		if err != nil {
			if dbent.IsNotFound(err) {
				return false, ErrResetCardForbidden
			}
			return false, fmt.Errorf("load subscription group: %w", err)
		}
		if grp.Status != "" && grp.Status != StatusActive {
			return false, ErrResetCardForbidden
		}
		return grp.WeeklyLimitUsd != nil && *grp.WeeklyLimitUsd > 0, nil
	}
	if s.groupRepo != nil {
		grp, err := s.groupRepo.GetByID(ctx, sub.GroupID)
		if err != nil || grp == nil {
			return false, ErrResetCardForbidden
		}
		if grp.Status != "" && grp.Status != StatusActive {
			return false, ErrResetCardForbidden
		}
		return grp.HasWeeklyLimit(), nil
	}
	if client != nil {
		grp, err := client.Group.Query().Where(group.IDEQ(sub.GroupID)).Only(ctx)
		if err != nil {
			if dbent.IsNotFound(err) {
				return false, ErrResetCardForbidden
			}
			return false, fmt.Errorf("load subscription group: %w", err)
		}
		if grp.Status != "" && grp.Status != StatusActive {
			return false, ErrResetCardForbidden
		}
		return grp.WeeklyLimitUsd != nil && *grp.WeeklyLimitUsd > 0, nil
	}
	return false, ErrResetCardServiceUnavailable
}

func (s *SubscriptionService) resetWeeklyUsageTx(ctx context.Context, client *dbent.Client, subID string, now time.Time) error {
	if client == nil {
		return ErrResetCardServiceUnavailable
	}
	if _, err := client.UserSubscription.UpdateOneID(subID).
		SetWeeklyUsageUsd(0).
		SetWeeklyWindowStart(now).
		Save(ctx); err != nil {
		if dbent.IsNotFound(err) {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("reset weekly subscription usage: %w", err)
	}
	return nil
}

func issueResetCardsTx(
	ctx context.Context,
	client *dbent.Client,
	sub *dbent.UserSubscription,
	quantity, validityDays int,
	issuedBy, sourceType, sourceID string,
	now time.Time,
) (int, error) {
	if client == nil || sub == nil {
		return 0, ErrResetCardServiceUnavailable
	}
	if quantity <= 0 {
		return 0, nil
	}
	if quantity > MaxResetCardIssueQuantity {
		return 0, ErrResetCardInvalidInput
	}
	if validityDays <= 0 {
		validityDays = DefaultResetCardValidityDays
	}
	if validityDays > MaxResetCardValidityDays {
		return 0, ErrResetCardInvalidInput
	}
	sourceType = strings.TrimSpace(sourceType)
	if sourceType == "" {
		sourceType = "manual"
	}
	if len(sourceType) > 40 {
		sourceType = uuid.NewSHA1(uuid.Nil, []byte(sourceType)).String()
	}
	sourceID = strings.TrimSpace(sourceID)
	if sourceID == "" {
		sourceID = uuid.NewString()
	} else if len(sourceID) > 128 {
		// Idempotency-Key accepts a wider range than the ledger's source_id
		// column. Keep long caller-provided keys deterministic and bounded so a
		// valid retry cannot fail merely because the audit key is too long.
		sourceID = uuid.NewSHA1(uuid.Nil, []byte(sourceID)).String()
	}
	// Use a stable source sequence so retries with the same idempotency/source
	// cannot create another copy of an already issued card.
	existing, err := client.SubscriptionResetCard.Query().Where(
		subscriptionresetcard.SourceTypeEQ(sourceType),
		subscriptionresetcard.SourceIDEQ(sourceID),
		subscriptionresetcard.SubscriptionIDEQ(sub.ID),
	).All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query existing reset cards: %w", err)
	}
	seen := make(map[int]struct{}, len(existing))
	for _, card := range existing {
		seen[card.GrantIndex] = struct{}{}
	}
	batchID := uuid.NewString()
	expiresAt := now.AddDate(0, 0, validityDays)
	issued := 0
	for index := 0; index < quantity; index++ {
		if _, ok := seen[index]; ok {
			continue
		}
		builder := client.SubscriptionResetCard.Create().
			SetSubscriptionID(sub.ID).
			SetUserID(sub.UserID).
			SetGroupID(sub.GroupID).
			SetStatus(ResetCardStatusAvailable).
			SetIssuedAt(now).
			SetExpiresAt(expiresAt).
			SetSourceType(sourceType).
			SetSourceID(sourceID).
			SetGrantBatchID(batchID).
			SetGrantIndex(index)
		if strings.TrimSpace(issuedBy) != "" {
			builder.SetIssuedBy(strings.TrimSpace(issuedBy))
		}
		// Use a database-level no-op conflict handler.  Catching a unique
		// violation after a plain INSERT would abort the surrounding Postgres
		// transaction, making the remaining cards fail and breaking retry
		// safety under concurrent fulfillment.
		if _, err := builder.OnConflictColumns(
			subscriptionresetcard.FieldSourceType,
			subscriptionresetcard.FieldSourceID,
			subscriptionresetcard.FieldSubscriptionID,
			subscriptionresetcard.FieldGrantIndex,
		).DoNothing().ID(ctx); err != nil {
			// Another fulfillment attempt may have inserted this exact card
			// after the preflight query above. PostgreSQL returns no row for
			// ON CONFLICT DO NOTHING ... RETURNING; treat that as an idempotent
			// duplicate rather than counting a card we did not create.
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return issued, fmt.Errorf("issue reset card: %w", err)
		}
		issued++
	}
	return issued, nil
}

// grantResetCardsTx is used by payment and pending-subscription activation.
// It deliberately accepts the transaction client so card issuance commits
// atomically with the entitlement itself.
func (s *SubscriptionService) grantResetCardsTx(ctx context.Context, client *dbent.Client, sub *UserSubscription, count, validityDays int, sourceType, sourceID string, now time.Time) (int, error) {
	if count <= 0 {
		return 0, nil
	}
	if sub == nil {
		return 0, ErrSubscriptionNotFound
	}
	entSub, err := client.UserSubscription.Query().Where(usersubscription.IDEQ(sub.ID)).Only(ctx)
	if err != nil {
		return 0, fmt.Errorf("reload subscription for reset cards: %w", err)
	}
	return issueResetCardsTx(ctx, client, entSub, count, validityDays, "", sourceType, sourceID, now)
}

func (s *SubscriptionService) ConsumeResetCard(ctx context.Context, userID, subscriptionID string) (*ResetCardConsumeResult, error) {
	if s == nil || s.entClient == nil {
		return nil, ErrResetCardServiceUnavailable
	}
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(subscriptionID) == "" {
		return nil, ErrResetCardInvalidInput
	}
	if existingTx := dbent.TxFromContext(ctx); existingTx != nil {
		return s.consumeResetCardTx(ctx, existingTx.Client(), userID, subscriptionID)
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin reset card transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	result, err := s.consumeResetCardTx(txCtx, tx.Client(), userID, subscriptionID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit reset card transaction: %w", err)
	}
	if result != nil && result.Subscription != nil {
		_ = s.invalidateSubscriptionCaches(result.Subscription.UserID, result.Subscription.GroupID)
		// Return a fully refreshed subscription and card ledger after the
		// transaction commits. The in-transaction snapshot intentionally only
		// contains the fields needed to complete the atomic operation, while the
		// API contract promises the latest usage windows and card summary.
		if refreshed, refreshErr := s.GetByID(ctx, result.Subscription.ID); refreshErr == nil && refreshed != nil {
			result.Subscription = refreshed
			result.ResetCards = refreshed.ResetCards
		}
	}
	return result, nil
}

func (s *SubscriptionService) consumeResetCardTx(ctx context.Context, client *dbent.Client, userID, subscriptionID string) (*ResetCardConsumeResult, error) {
	// Include soft-deleted rows long enough to distinguish a revoked
	// subscription from a missing one. Revoked subscriptions retain their card
	// ledger for audit/history, but must never be allowed to consume a card.
	queryCtx := mixins.SkipSoftDelete(ctx)
	sub, err := client.UserSubscription.Query().Where(usersubscription.IDEQ(subscriptionID)).ForUpdate().Only(queryCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("lock subscription for reset card: %w", err)
	}
	if sub.UserID != userID {
		return nil, ErrResetCardForbidden
	}
	// Take the timestamp after the subscription lock is acquired so a request
	// that waited on a concurrent reset cannot consume a card using an already
	// stale expiry comparison.
	now := s.resetCardNow()
	if sub.DeletedAt != nil || sub.Status != SubscriptionStatusActive || !sub.ExpiresAt.After(now) {
		return nil, ErrResetCardSubscriptionInactive
	}
	hasWeekly, err := s.subscriptionHasWeeklyLimit(ctx, client, sub)
	if err != nil {
		return nil, err
	}
	if !hasWeekly {
		return nil, ErrWeeklyLimitNotConfigured
	}
	card, err := client.SubscriptionResetCard.Query().Where(
		subscriptionresetcard.SubscriptionIDEQ(subscriptionID),
		subscriptionresetcard.StatusEQ(ResetCardStatusAvailable),
		subscriptionresetcard.ExpiresAtGT(now),
	).Order(
		subscriptionresetcard.ByExpiresAt(),
		subscriptionresetcard.ByIssuedAt(),
		subscriptionresetcard.ByID(),
	).ForUpdate().First(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrResetCardNotAvailable
		}
		return nil, fmt.Errorf("lock reset card: %w", err)
	}
	if card.UserID != sub.UserID || card.GroupID != sub.GroupID {
		return nil, ErrResetCardForbidden
	}
	if _, err := client.SubscriptionResetCard.UpdateOneID(card.ID).
		SetStatus(ResetCardStatusConsumed).
		SetConsumedAt(now).
		SetConsumedBy(userID).
		Save(ctx); err != nil {
		return nil, fmt.Errorf("consume reset card: %w", err)
	}
	if err := s.resetWeeklyUsageTx(ctx, client, subscriptionID, now); err != nil {
		return nil, err
	}
	resultSub := userSubscriptionEntityToServiceForResetCard(sub)
	// The locked Ent entity contains the old window values. Reload through the
	// repository after commit in the public method; this fallback keeps nested
	// transaction callers useful without pretending the old snapshot is fresh.
	resultSub.WeeklyUsageUSD = 0
	resultSub.WeeklyWindowStart = &now
	windowEnd := now.Add(7 * 24 * time.Hour)
	resultSub.ResetCards = ResetCardSummary{ConsumedCount: 1}
	return &ResetCardConsumeResult{
		Subscription:    resultSub,
		ResetAt:         now,
		WeeklyWindowEnd: windowEnd,
		ResetCards:      resultSub.ResetCards,
	}, nil
}

// userSubscriptionEntityToServiceForResetCard maps the fields needed before
// the outer repository reload. Keeping this local avoids coupling the service
// package to repository-only mappers.
func userSubscriptionEntityToServiceForResetCard(sub *dbent.UserSubscription) *UserSubscription {
	if sub == nil {
		return nil
	}
	return &UserSubscription{
		ID: sub.ID, UserID: sub.UserID, GroupID: sub.GroupID,
		StartsAt: sub.StartsAt, ExpiresAt: sub.ExpiresAt, Status: sub.Status,
		DailyWindowStart: sub.DailyWindowStart, WeeklyWindowStart: sub.WeeklyWindowStart,
		MonthlyWindowStart: sub.MonthlyWindowStart,
		DailyUsageUSD:      sub.DailyUsageUsd, WeeklyUsageUSD: sub.WeeklyUsageUsd,
		MonthlyUsageUSD: sub.MonthlyUsageUsd, AssignedAt: sub.AssignedAt,
		CreatedAt: sub.CreatedAt, UpdatedAt: sub.UpdatedAt,
	}
}

func (s *SubscriptionService) IssueResetCards(ctx context.Context, subscriptionIDs []string, quantity, validityDays int, issuedBy, operationID string) *ResetCardIssueResult {
	result := &ResetCardIssueResult{Items: make([]ResetCardIssueItem, 0, len(subscriptionIDs))}
	if s == nil || s.entClient == nil {
		for _, id := range uniqueResetCardIDs(subscriptionIDs) {
			result.Items = append(result.Items, ResetCardIssueItem{SubscriptionID: id, ErrorCode: infraerrors.Reason(ErrResetCardServiceUnavailable), Error: ErrResetCardServiceUnavailable.Error()})
		}
		result.FailedCount = len(result.Items)
		return result
	}
	if quantity <= 0 || quantity > MaxResetCardIssueQuantity {
		result.Items = append(result.Items, ResetCardIssueItem{ErrorCode: infraerrors.Reason(ErrResetCardInvalidInput), Error: ErrResetCardInvalidInput.Error()})
		result.FailedCount = 1
		return result
	}
	if validityDays == 0 {
		validityDays = DefaultResetCardValidityDays
	}
	if validityDays < 1 || validityDays > MaxResetCardValidityDays {
		result.Items = append(result.Items, ResetCardIssueItem{ErrorCode: infraerrors.Reason(ErrResetCardInvalidInput), Error: ErrResetCardInvalidInput.Error()})
		result.FailedCount = 1
		return result
	}
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		operationID = uuid.NewString()
	}
	for _, id := range uniqueResetCardIDs(subscriptionIDs) {
		item := ResetCardIssueItem{SubscriptionID: id}
		tx, err := s.entClient.Tx(ctx)
		if err != nil {
			item.ErrorCode = infraerrors.Reason(ErrResetCardServiceUnavailable)
			item.Error = err.Error()
			result.FailedCount++
			result.Items = append(result.Items, item)
			continue
		}
		txCtx := dbent.NewTxContext(ctx, tx)
		sub, lockErr := tx.UserSubscription.Query().Where(usersubscription.IDEQ(id)).ForUpdate().Only(mixins.SkipSoftDelete(txCtx))
		if dbent.IsNotFound(lockErr) {
			lockErr = ErrSubscriptionNotFound
		}
		if lockErr == nil && (sub.DeletedAt != nil || sub.Status == SubscriptionStatusRevoked) {
			lockErr = ErrResetCardForbidden
		}
		if lockErr != nil {
			_ = tx.Rollback()
			item.ErrorCode = infraerrors.Reason(lockErr)
			item.Error = lockErr.Error()
			result.FailedCount++
			result.Items = append(result.Items, item)
			continue
		}
		issued, issueErr := issueResetCardsTx(txCtx, tx.Client(), sub, quantity, validityDays, issuedBy, "admin_issue", operationID, s.resetCardNow())
		if issueErr == nil {
			issueErr = tx.Commit()
			if issueErr != nil {
				_ = tx.Rollback()
			}
		} else {
			_ = tx.Rollback()
		}
		if issueErr != nil {
			item.ErrorCode = infraerrors.Reason(issueErr)
			item.Error = issueErr.Error()
			result.FailedCount++
		} else {
			item.IssuedCount = issued
			item.Success = true
			result.SuccessCount++
			result.TotalIssued += issued
			_ = s.invalidateSubscriptionCaches(sub.UserID, sub.GroupID)
		}
		result.Items = append(result.Items, item)
	}
	return result
}

func uniqueResetCardIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *SubscriptionService) ResetWeeklyQuotas(ctx context.Context, subscriptionIDs []string) *WeeklyResetResult {
	result := &WeeklyResetResult{Items: make([]WeeklyResetItem, 0, len(subscriptionIDs))}
	if s == nil || s.entClient == nil {
		for _, id := range uniqueResetCardIDs(subscriptionIDs) {
			result.Items = append(result.Items, WeeklyResetItem{SubscriptionID: id, ErrorCode: infraerrors.Reason(ErrResetCardServiceUnavailable), Error: ErrResetCardServiceUnavailable.Error()})
		}
		result.FailedCount = len(result.Items)
		return result
	}
	for _, id := range uniqueResetCardIDs(subscriptionIDs) {
		item := WeeklyResetItem{SubscriptionID: id}
		tx, err := s.entClient.Tx(ctx)
		if err != nil {
			item.ErrorCode = infraerrors.Reason(ErrResetCardServiceUnavailable)
			item.Error = err.Error()
			result.FailedCount++
			result.Items = append(result.Items, item)
			continue
		}
		txCtx := dbent.NewTxContext(ctx, tx)
		sub, lockErr := tx.UserSubscription.Query().Where(usersubscription.IDEQ(id)).ForUpdate().Only(mixins.SkipSoftDelete(txCtx))
		if dbent.IsNotFound(lockErr) {
			lockErr = ErrSubscriptionNotFound
		}
		if lockErr == nil {
			now := s.resetCardNow()
			if sub.DeletedAt != nil || sub.Status == SubscriptionStatusRevoked || sub.Status != SubscriptionStatusActive || !sub.ExpiresAt.After(now) {
				lockErr = ErrResetCardSubscriptionInactive
			} else {
				hasWeekly, weeklyErr := s.subscriptionHasWeeklyLimit(txCtx, tx.Client(), sub)
				if weeklyErr != nil {
					lockErr = weeklyErr
				} else if !hasWeekly {
					lockErr = ErrWeeklyLimitNotConfigured
				} else {
					lockErr = s.resetWeeklyUsageTx(txCtx, tx.Client(), id, now)
					if lockErr == nil {
						item.ResetAt = &now
						end := now.Add(7 * 24 * time.Hour)
						item.WeeklyWindowEnd = &end
					}
				}
			}
		}
		if lockErr != nil {
			_ = tx.Rollback()
			item.ErrorCode = infraerrors.Reason(lockErr)
			item.Error = lockErr.Error()
			result.FailedCount++
		} else if commitErr := tx.Commit(); commitErr != nil {
			_ = tx.Rollback()
			item.ErrorCode = infraerrors.Reason(commitErr)
			item.Error = commitErr.Error()
			result.FailedCount++
		} else {
			item.Success = true
			result.SuccessCount++
			_ = s.invalidateSubscriptionCaches(sub.UserID, sub.GroupID)
		}
		result.Items = append(result.Items, item)
	}
	return result
}
