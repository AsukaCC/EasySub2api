//go:build unit

package service

import (
	"context"
	"math"
	"testing"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/payment"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestReviewRefundTicketClaimsApprovalAmountOnce(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	ticket, err := client.RefundTicket.Create().
		SetOrderID("01991f17-b421-7e42-884f-e66c2c6d4331").
		SetUserID("01991f18-c16b-74fd-8684-22f77d5f9f2a").
		SetStatus(RefundTicketStatusPending).
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	firstAmount := 25.0
	_, err = svc.ReviewRefundTicket(ctx, ReviewRefundTicketInput{
		TicketID: ticket.ID, ReviewerID: "01991f1a-e3bd-762b-abfb-0cb600ff4513",
		Decision: "APPROVE", ApprovedPrincipalAmount: &firstAmount, ReviewNote: "first approval",
	})
	require.Error(t, err)
	require.NotEqual(t, "REFUND_TICKET_APPROVAL_CONFLICT", infraerrors.Reason(err), "the first approval must claim the amount before processing continues")

	claimed, err := client.RefundTicket.Get(ctx, ticket.ID)
	require.NoError(t, err)
	require.Equal(t, RefundTicketStatusApproved, claimed.Status)
	require.NotNil(t, claimed.ApprovedPrincipalAmount)
	require.Equal(t, firstAmount, *claimed.ApprovedPrincipalAmount)
	require.Equal(t, "first approval", claimed.ReviewNote)

	conflictingAmount := 30.0
	_, err = svc.ReviewRefundTicket(ctx, ReviewRefundTicketInput{
		TicketID: ticket.ID, ReviewerID: "01991f1b-99bc-7651-852e-6f03f6e989ba",
		Decision: "APPROVE", ApprovedPrincipalAmount: &conflictingAmount, ReviewNote: "second approval",
	})
	require.Error(t, err)
	require.Equal(t, "REFUND_TICKET_APPROVAL_CONFLICT", infraerrors.Reason(err))

	replayed, err := svc.ReviewRefundTicket(ctx, ReviewRefundTicketInput{
		TicketID: ticket.ID, ReviewerID: "01991f1b-99bc-7651-852e-6f03f6e989ba",
		Decision: "APPROVE", ApprovedPrincipalAmount: &firstAmount, ReviewNote: "ignored replay note",
	})
	require.Nil(t, replayed)
	require.Error(t, err)
	require.NotEqual(t, "REFUND_TICKET_APPROVAL_CONFLICT", infraerrors.Reason(err), "same-amount replay must pass approval compatibility and resume processing")

	claimed, err = client.RefundTicket.Get(ctx, ticket.ID)
	require.NoError(t, err)
	require.Equal(t, firstAmount, *claimed.ApprovedPrincipalAmount)
	require.Equal(t, "first approval", claimed.ReviewNote)
}

func TestFailPaymentRefundKeepsExplicitFailureStatusAfterPartialRefund(t *testing.T) {
	client := newPaymentConfigServiceTestClient(t)
	now := time.Now().UTC()
	update := paymentRefundFailedOrderUpdate(
		client.PaymentOrder.UpdateOneID("01991f31-e3b5-791c-8af6-2f46c66053d9"),
		now,
		"provider rejected refund",
	)
	mutation := update.Mutation()
	status, statusSet := mutation.Status()
	require.True(t, statusSet)
	require.Equal(t, OrderStatusRefundFailed, status)
	_, refundAmountSet := mutation.RefundAmount()
	require.False(t, refundAmountSet, "a failed attempt must preserve the cumulative successful refund amount")
	failedAt, failedAtSet := mutation.FailedAt()
	require.True(t, failedAtSet)
	require.Equal(t, now, failedAt)
	failedReason, failedReasonSet := mutation.FailedReason()
	require.True(t, failedReasonSet)
	require.Equal(t, "provider rejected refund", failedReason)
}

func TestMarkRefundPendingPreservesSuccessfulRefundAggregate(t *testing.T) {
	client := newPaymentConfigServiceTestClient(t)
	update := paymentRefundPendingOrderUpdate(client.PaymentOrder.UpdateOneID("01991f32-3a24-7a40-b4f7-cf6904950e96"))
	mutation := update.Mutation()
	status, statusSet := mutation.Status()
	require.True(t, statusSet)
	require.Equal(t, OrderStatusRefundPending, status)
	_, refundAmountSet := mutation.RefundAmount()
	require.False(t, refundAmountSet, "pending target must not become a successful refund aggregate")
	_, refundReasonSet := mutation.RefundReason()
	require.False(t, refundReasonSet, "pending metadata must not overwrite the last successful refund reason")
	require.False(t, mutation.RefundAtCleared(), "pending state must preserve the last successful refund timestamp")
}

func TestProviderErrorWithExplicitFailedStatusIsDefinitive(t *testing.T) {
	require.True(t, isDefinitiveProviderRefundFailure(&payment.RefundResponse{Status: payment.ProviderStatusFailed}))
	require.False(t, isDefinitiveProviderRefundFailure(&payment.RefundResponse{Status: payment.ProviderStatusPending}))
	require.False(t, isDefinitiveProviderRefundFailure(nil))
}

func TestRefundPrincipalRequiresCNYPrecision(t *testing.T) {
	require.True(t, hasRefundMoneyPrecision(12.34))
	require.True(t, hasRefundMoneyPrecision(12))
	require.False(t, hasRefundMoneyPrecision(12.345))
}

func TestPreparePaymentRefundValidatesPrincipalBeforeFingerprinting(t *testing.T) {
	client := newPaymentConfigServiceTestClient(t)
	svc := &PaymentService{entClient: client}

	for _, principal := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), -1, 0, 12.345} {
		principal := principal
		require.NotPanics(t, func() {
			_, _, err := svc.preparePaymentRefund(context.Background(), CreatePaymentRefundInput{
				OrderID: "01991f35-72a0-7aad-93d7-0d005d42e0be", IdempotencyKey: "non-finite-principal",
				Principal: &principal,
			})
			require.Error(t, err)
			require.Equal(t, "INVALID_AMOUNT", infraerrors.Reason(err))
		})
	}
}

func TestFailedTicketRefundConvergesFromApprovedOrProcessing(t *testing.T) {
	require.ElementsMatch(t,
		[]string{RefundTicketStatusApproved, RefundTicketStatusProcessing},
		refundTicketFailureSourceStatuses(),
	)
}

func TestAdminDirectRefundUsesSameSevenDayWindowAsSelfService(t *testing.T) {
	now := time.Now().UTC()
	withinWindow := now.Add(-time.Hour)
	expired := now.Add(-selfServiceRefundWindow - time.Minute)

	require.True(t, refundSelfServiceEligible(&dbent.PaymentOrder{CompletedAt: &withinWindow}, now))
	require.False(t, refundSelfServiceEligible(&dbent.PaymentOrder{CompletedAt: &expired}, now))
}

func TestNormalizeAdminRefundInputRecoversAffiliateAndUsesOrderOwner(t *testing.T) {
	input := normalizeAdminRefundInput(CreatePaymentRefundInput{UserID: "caller-supplied", Source: "other"})
	require.Empty(t, input.UserID)
	require.Equal(t, RefundSourceAdmin, input.Source)
	require.True(t, input.AutoAffiliate)
}

func TestPrepareAdminRefundIdempotencyDoesNotCompareEmptyCallerUserID(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	input := normalizeAdminRefundInput(CreatePaymentRefundInput{
		OrderID: "01991f45-16a4-7780-9495-a03e789bdc4d", IdempotencyKey: "admin-refund-replay", Reason: "admin replay",
	})
	fingerprint := refundRequestFingerprint(input)
	existing, err := client.PaymentRefund.Create().
		SetOrderID(input.OrderID).
		SetUserID("01991f46-98eb-7748-b34a-1bf442587404").
		SetSource(RefundSourceAdmin).
		SetStatus(RefundStatusFailed).
		SetIdempotencyKey(input.IdempotencyKey).
		SetRequestFingerprint(fingerprint).
		SetProviderRequestID("admin-refund-replay-request").
		Save(ctx)
	require.NoError(t, err)

	refund, replayed, err := (&PaymentService{entClient: client}).preparePaymentRefund(ctx, input)
	require.NoError(t, err)
	require.True(t, replayed)
	require.Equal(t, existing.ID, refund.ID)
}

func TestCreateAdminPaymentRefundRequiresTicketAfterSevenDays(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("expired-admin-refund@example.com").
		SetPasswordHash("hash").
		SetUsername("expired-admin-refund").
		Save(ctx)
	require.NoError(t, err)
	completedAt := time.Now().UTC().Add(-selfServiceRefundWindow - time.Minute)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetRechargeCode("EXPIRED-ADMIN-REFUND").
		SetOutTradeNo("sub2_expired_admin_refund").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-expired-admin-refund").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetCompletedAt(completedAt).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	_, err = (&PaymentService{entClient: client}).CreateAdminPaymentRefund(ctx, CreatePaymentRefundInput{
		OrderID: order.ID, IdempotencyKey: "expired-admin-refund",
	})
	require.Error(t, err)
	require.Equal(t, "REFUND_TICKET_REQUIRED", infraerrors.Reason(err))
}

type affiliateWalletHoldSpy struct {
	WalletRepository
	summary WalletSummary
	holds   []WalletHoldInput
}

type partialFrozenRefundCapacityRepo struct {
	UserRepository
	capacity RefundPointCapacity
}

func (r *partialFrozenRefundCapacityRepo) GetRefundPointCapacity(context.Context, string, string) (RefundPointCapacity, error) {
	return r.capacity, nil
}

func (*partialFrozenRefundCapacityRepo) HoldRefundPoints(context.Context, RefundPointHoldInput) (WalletHoldResult, error) {
	return WalletHoldResult{}, nil
}

func TestPartialFrozenSourceBonusUsesRechargePointsForShortfall(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	repo := &partialFrozenRefundCapacityRepo{capacity: RefundPointCapacity{
		RechargeAvailable: 110, SourceBonusAvailable: 10, SourceBonusFrozen: 5,
	}}
	completedAt := time.Now().UTC().Add(-time.Hour)
	order := &dbent.PaymentOrder{
		ID: "01991f8e-e2e0-7b01-a679-c115339810f8", UserID: "01991f8f-60d2-7945-80b6-34bfa8447405",
		PrincipalAmount: 100, BasePoints: 100, BonusPoints: 20, CompletedAt: &completedAt,
	}
	requested := 100.0

	calculation, err := (&PaymentService{entClient: client, userRepo: repo}).calculateRefundQuote(ctx, client, order, &requested, false)
	require.NoError(t, err)
	require.Equal(t, 100.0, calculation.quote.MaxRefundablePrincipalAmount)
	require.Equal(t, 100.0, calculation.quote.PrincipalAmount)
	require.Equal(t, 120.0, calculation.quote.PointsToHold)
	require.Empty(t, calculation.quote.BlockedReason)
}

func TestRefundReconciliationIncludesStaleReservedRecords(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	now := time.Now().UTC()
	staleBefore := now.Add(-refundSubmittingRetryAfter)

	create := func(id, orderID, status string, createdAt time.Time) {
		t.Helper()
		builder := client.PaymentRefund.Create().
			SetID(id).
			SetOrderID(orderID).
			SetUserID("01991f9e-36ee-7afb-ae16-33d25c6b2b05").
			SetStatus(status).
			SetIdempotencyKey("key-" + id).
			SetProviderRequestID("request-" + id).
			SetCreatedAt(createdAt)
		if status == RefundStatusSubmitting {
			builder.SetSubmittedAt(createdAt)
		}
		_, err := builder.Save(ctx)
		require.NoError(t, err)
	}

	create("01991fa1-541d-75ae-b8bf-496dff1e04b4", "01991fa2-188f-77d3-b7e1-242a79f5327a", RefundStatusReserved, staleBefore.Add(-time.Second))
	create("01991fa3-6c57-7a52-a5f6-b40610668a16", "01991fa4-594b-7180-9864-c97b1b626080", RefundStatusReserved, staleBefore.Add(time.Second))
	create("01991fa5-bba4-73e1-a3a8-8878752fe8d2", "01991fa6-3dd9-7af6-94bf-05f14e25c92a", RefundStatusSubmitting, staleBefore.Add(-time.Second))
	create("01991fa7-b5bf-7797-a132-c3ef2a3b7370", "01991fa8-776f-79ea-81fd-5262a17ed29d", RefundStatusPending, now)

	items, err := client.PaymentRefund.Query().Where(paymentRefundReconciliationPredicate(staleBefore)).All(ctx)
	require.NoError(t, err)
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	require.ElementsMatch(t, []string{
		"01991fa1-541d-75ae-b8bf-496dff1e04b4",
		"01991fa5-bba4-73e1-a3a8-8878752fe8d2",
		"01991fa7-b5bf-7797-a132-c3ef2a3b7370",
	}, ids)
}

func (s *affiliateWalletHoldSpy) GetWalletSummary(context.Context, string) (WalletSummary, error) {
	return s.summary, nil
}

func (s *affiliateWalletHoldSpy) HoldWallet(_ context.Context, input WalletHoldInput) (WalletHoldResult, error) {
	s.holds = append(s.holds, input)
	return WalletHoldResult{Applied: true, HoldID: "affiliate-hold", Amount: input.Amount}, nil
}

func TestAffiliateRefundFreezesTransferredWalletPointsBeforeProviderCall(t *testing.T) {
	target := affiliateWalletReservationTarget(10, 6, 0, 10)
	require.Equal(t, 4.0, target)
	require.Equal(t, 10.0, affiliateWalletReservationTarget(10, 10, 0, 20), "prior refund reservations must get pool priority")

	spy := &affiliateWalletHoldSpy{summary: WalletSummary{AvailableBalance: 2}}
	hold, err := holdAffiliateWalletReservation(context.Background(), spy, affiliateReversalReservation{
		InviterID: "inviter-1", WalletTarget: target,
	}, "refund-1", "fingerprint-1")
	require.NoError(t, err)
	require.True(t, hold.Applied)
	require.Len(t, spy.holds, 1)
	require.Equal(t, 2.0, spy.holds[0].Amount, "freeze the available portion and leave only the residual for overdraft recovery")
	require.Equal(t, affiliateRefundHoldPurpose, spy.holds[0].Purpose)
	require.Equal(t, "refund-1", spy.holds[0].ReferenceID)

	settlement, err := affiliateReversalSettlement(10, 2, 4, 2)
	require.NoError(t, err)
	require.Equal(t, 2.0, settlement.Frozen)
	require.Equal(t, 4.0, settlement.Available)
	require.Equal(t, 2.0, settlement.Wallet)
	_, err = affiliateReversalSettlement(10, 11, 0, 0)
	require.Error(t, err)
}
