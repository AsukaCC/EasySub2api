package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/ent/paymentorder"
	"github.com/AsukaCC/EasySub2api/ent/paymentrefund"
	"github.com/AsukaCC/EasySub2api/ent/predicate"
	"github.com/AsukaCC/EasySub2api/ent/supportticket"
	"github.com/AsukaCC/EasySub2api/internal/payment"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	RefundStatusRequested  = "REQUESTED"
	RefundStatusReserved   = "RESERVED"
	RefundStatusSubmitting = "SUBMITTING"
	RefundStatusPending    = "PENDING"
	RefundStatusSucceeded  = "SUCCEEDED"
	RefundStatusFailed     = "FAILED"

	RefundSourceSelfService = "SELF_SERVICE"
	RefundSourceTicket      = "TICKET"
	RefundSourceAdmin       = "ADMIN"

	selfServiceRefundWindow    = 168 * time.Hour
	refundSubmittingRetryAfter = 2 * time.Minute
	refundSettlementTimeout    = 15 * time.Second
	affiliateRefundHoldPurpose = "affiliate_refund"
)

type RefundQuote struct {
	OrderID                      string     `json:"order_id"`
	Currency                     string     `json:"currency"`
	RequestedPrincipalAmount     float64    `json:"requested_principal_amount"`
	PrincipalAmount              float64    `json:"principal_amount"`
	FeeAmount                    float64    `json:"fee_amount"`
	RefundFeeRate                float64    `json:"refund_fee_rate"`
	RefundFeeAmount              float64    `json:"refund_fee_amount"`
	GatewayAmount                float64    `json:"gateway_amount"`
	BasePoints                   float64    `json:"base_points"`
	BonusPoints                  float64    `json:"bonus_points"`
	PointsToHold                 float64    `json:"points_to_hold"`
	BonusExpiredOffset           float64    `json:"bonus_expired_offset"`
	AffiliateRebatePoints        float64    `json:"affiliate_rebate_points"`
	RemainingPrincipalAmount     float64    `json:"remaining_principal_amount"`
	MaxRefundablePrincipalAmount float64    `json:"max_refundable_principal_amount"`
	RefundDeadline               *time.Time `json:"refund_deadline,omitempty"`
	SelfServiceEligible          bool       `json:"self_service_eligible"`
	RequiresTicket               bool       `json:"requires_ticket"`
	BlockedReason                string     `json:"blocked_reason,omitempty"`
}

type CreatePaymentRefundInput struct {
	OrderID        string
	UserID         string
	RequestedBy    string
	IdempotencyKey string
	Principal      *float64
	Reason         string
	Source         string
	TicketID       string
	AutoAffiliate  bool
}

type PaymentRefundResponse struct {
	ID                    string     `json:"id"`
	OrderID               string     `json:"order_id"`
	TicketID              *string    `json:"ticket_id,omitempty"`
	Source                string     `json:"source"`
	Status                string     `json:"status"`
	Currency              string     `json:"currency"`
	RequestedPrincipal    float64    `json:"requested_principal_amount"`
	PrincipalAmount       float64    `json:"principal_amount"`
	FeeAmount             float64    `json:"fee_amount"`
	RefundFeeRate         float64    `json:"refund_fee_rate"`
	RefundFeeAmount       float64    `json:"refund_fee_amount"`
	GatewayAmount         float64    `json:"gateway_amount"`
	BasePoints            float64    `json:"base_points"`
	BonusPoints           float64    `json:"bonus_points"`
	BonusExpiredOffset    float64    `json:"bonus_expired_offset"`
	AffiliateRebatePoints float64    `json:"affiliate_rebate_points"`
	ProviderRefundID      *string    `json:"provider_refund_id,omitempty"`
	Reason                string     `json:"reason"`
	ErrorCode             string     `json:"error_code,omitempty"`
	ErrorMessage          string     `json:"error_message,omitempty"`
	SubmittedAt           *time.Time `json:"submitted_at,omitempty"`
	SettledAt             *time.Time `json:"settled_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
}

type refundCalculation struct {
	quote     RefundQuote
	amounts   CumulativeRefundAmounts
	affTarget decimal.Decimal
	affDelta  decimal.Decimal
}

func paymentRefundResponse(refund *dbent.PaymentRefund) *PaymentRefundResponse {
	if refund == nil {
		return nil
	}
	return &PaymentRefundResponse{
		ID: refund.ID, OrderID: refund.OrderID, TicketID: refund.TicketID,
		Source: refund.Source, Status: refund.Status, Currency: refund.Currency,
		RequestedPrincipal: refund.RequestedPrincipalAmount,
		PrincipalAmount:    refund.PrincipalAmount, FeeAmount: refund.FeeAmount,
		RefundFeeRate: refund.RefundFeeRate, RefundFeeAmount: refund.RefundFeeAmount,
		GatewayAmount: refund.GatewayAmount, BasePoints: refund.BasePoints,
		BonusPoints: refund.BonusPoints, BonusExpiredOffset: refund.BonusExpiredOffset,
		AffiliateRebatePoints: refund.AffiliateRebatePoints,
		ProviderRefundID:      refund.ProviderRefundID, Reason: refund.Reason,
		ErrorCode: refund.ErrorCode, ErrorMessage: refund.ErrorMessage,
		SubmittedAt: refund.SubmittedAt, SettledAt: refund.SettledAt, CreatedAt: refund.CreatedAt,
	}
}

func (s *PaymentService) GetRefundQuote(ctx context.Context, orderID, userID string, principal *float64) (*RefundQuote, error) {
	order, err := s.entClient.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if order.UserID != userID {
		return nil, infraerrors.Forbidden("FORBIDDEN", "no permission")
	}
	if err := s.validateRechargeRefundOrder(ctx, order, false); err != nil {
		return nil, err
	}
	instance, err := s.getRefundOrderProviderInstance(ctx, order)
	if err != nil || instance == nil {
		return nil, infraerrors.Forbidden("REFUND_DISABLED", "refund is not available for this payment provider")
	}
	directEligible := instance.AllowUserRefund && refundSelfServiceEligible(order, time.Now())
	calculation, err := s.calculateRefundQuote(ctx, s.entClient, order, principal, directEligible)
	if err != nil {
		return nil, err
	}
	calculation.quote.SelfServiceEligible = directEligible
	calculation.quote.RequiresTicket = calculation.quote.RequiresTicket || !directEligible
	return &calculation.quote, nil
}

func (s *PaymentService) CreateSelfServiceRefund(ctx context.Context, input CreatePaymentRefundInput) (*PaymentRefundResponse, error) {
	input.Source = RefundSourceSelfService
	input.AutoAffiliate = true
	refund, existing, err := s.preparePaymentRefund(ctx, input)
	if err != nil {
		return nil, err
	}
	if existing && refund.Status != RefundStatusReserved && refund.Status != RefundStatusSubmitting {
		return paymentRefundResponse(refund), nil
	}
	refund, err = s.executePaymentRefund(ctx, refund.ID)
	if err != nil {
		return nil, err
	}
	return paymentRefundResponse(refund), nil
}

func (s *PaymentService) CreateAdminPaymentRefund(ctx context.Context, input CreatePaymentRefundInput) (*PaymentRefundResponse, error) {
	order, err := s.entClient.PaymentOrder.Get(ctx, input.OrderID)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if order.OrderType == payment.OrderTypeSubscription {
		return nil, infraerrors.BadRequest(
			"SUBSCRIPTION_REFUND_UNSUPPORTED",
			"subscription orders do not use payment-channel refunds; revoke the subscription and adjust platform points instead",
		)
	}
	if !refundSelfServiceEligible(order, time.Now()) {
		return nil, infraerrors.BadRequest(
			"REFUND_TICKET_REQUIRED",
			"the direct refund window has expired; review an approved refund ticket instead",
		)
	}
	input = normalizeAdminRefundInput(input)
	refund, existing, err := s.preparePaymentRefund(ctx, input)
	if err != nil {
		return nil, err
	}
	if existing && refund.Status != RefundStatusReserved && refund.Status != RefundStatusSubmitting {
		return paymentRefundResponse(refund), nil
	}
	refund, err = s.executePaymentRefund(ctx, refund.ID)
	if err != nil {
		return nil, err
	}
	return paymentRefundResponse(refund), nil
}

func normalizeAdminRefundInput(input CreatePaymentRefundInput) CreatePaymentRefundInput {
	input.Source = RefundSourceAdmin
	input.AutoAffiliate = true
	input.UserID = ""
	return input
}

func (s *PaymentService) preparePaymentRefund(ctx context.Context, input CreatePaymentRefundInput) (_ *dbent.PaymentRefund, existing bool, err error) {
	key := strings.TrimSpace(input.IdempotencyKey)
	if key == "" || len(key) > 160 {
		return nil, false, infraerrors.BadRequest("INVALID_IDEMPOTENCY_KEY", "Idempotency-Key is required and must be at most 160 characters")
	}
	if input.Principal != nil {
		if math.IsNaN(*input.Principal) || math.IsInf(*input.Principal, 0) || *input.Principal <= 0 {
			return nil, false, infraerrors.BadRequest("INVALID_AMOUNT", "principal_amount must be a positive finite number")
		}
		if !hasRefundMoneyPrecision(*input.Principal) {
			return nil, false, infraerrors.BadRequest("INVALID_AMOUNT", "principal_amount allows at most 2 decimal places")
		}
	}
	fingerprint := refundRequestFingerprint(input)
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("begin refund: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	txCtx := dbent.NewTxContext(ctx, tx)

	previous, queryErr := tx.PaymentRefund.Query().Where(
		paymentrefund.OrderIDEQ(input.OrderID),
		paymentrefund.IdempotencyKeyEQ(key),
	).Only(txCtx)
	if queryErr == nil {
		if (input.UserID != "" && previous.UserID != input.UserID) || previous.RequestFingerprint != fingerprint {
			return nil, false, infraerrors.Conflict("IDEMPOTENCY_CONFLICT", "Idempotency-Key was already used with different refund parameters")
		}
		if err = tx.Commit(); err != nil {
			return nil, false, fmt.Errorf("commit idempotent refund lookup: %w", err)
		}
		return previous, true, nil
	}
	if !dbent.IsNotFound(queryErr) {
		return nil, false, fmt.Errorf("query idempotent refund: %w", queryErr)
	}

	order, err := tx.PaymentOrder.Query().Where(paymentorder.IDEQ(input.OrderID)).ForUpdate().Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, false, infraerrors.NotFound("NOT_FOUND", "order not found")
		}
		return nil, false, fmt.Errorf("lock refund order: %w", err)
	}
	// The first lookup runs before the order lock. A concurrent request can
	// create the same idempotent refund while this transaction waits, so repeat
	// the lookup after acquiring the lock before validating the order state.
	previous, queryErr = tx.PaymentRefund.Query().Where(
		paymentrefund.OrderIDEQ(input.OrderID),
		paymentrefund.IdempotencyKeyEQ(key),
	).Only(txCtx)
	if queryErr == nil {
		if (input.UserID != "" && previous.UserID != input.UserID) || previous.RequestFingerprint != fingerprint {
			return nil, false, infraerrors.Conflict("IDEMPOTENCY_CONFLICT", "Idempotency-Key was already used with different refund parameters")
		}
		if err = tx.Commit(); err != nil {
			return nil, false, fmt.Errorf("commit locked idempotent refund lookup: %w", err)
		}
		return previous, true, nil
	}
	if !dbent.IsNotFound(queryErr) {
		return nil, false, fmt.Errorf("repeat idempotent refund lookup: %w", queryErr)
	}
	if input.UserID != "" && order.UserID != input.UserID {
		return nil, false, infraerrors.Forbidden("FORBIDDEN", "no permission")
	}
	selfService := input.Source == RefundSourceSelfService
	if err := s.validateRechargeRefundOrder(txCtx, order, selfService); err != nil {
		return nil, false, err
	}

	calculation, err := s.calculateRefundQuote(txCtx, tx.Client(), order, input.Principal, input.AutoAffiliate)
	if err != nil {
		return nil, false, err
	}
	if selfService && !calculation.quote.SelfServiceEligible {
		return nil, false, infraerrors.BadRequest("REFUND_TICKET_REQUIRED", "self-service refund window has expired; submit a refund ticket")
	}
	if calculation.amounts.PrincipalDelta.LessThan(decimal.New(1, -refundMoneyScale)) {
		return nil, false, infraerrors.BadRequest("NO_REFUNDABLE_AMOUNT", "no refundable principal is available for the current point balance")
	}

	refundID, uuidErr := uuid.NewV7()
	if uuidErr != nil {
		return nil, false, fmt.Errorf("generate refund id: %w", uuidErr)
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = "refund order:" + order.ID
	}
	requestedPrincipal := calculation.quote.RequestedPrincipalAmount
	requestedBy := strings.TrimSpace(input.RequestedBy)
	ticketID := strings.TrimSpace(input.TicketID)
	refund, err := tx.PaymentRefund.Create().
		SetID(refundID.String()).
		SetOrderID(order.ID).
		SetUserID(order.UserID).
		SetNillableTicketID(optionalString(ticketID)).
		SetSource(input.Source).
		SetStatus(RefundStatusRequested).
		SetIdempotencyKey(key).
		SetRequestFingerprint(fingerprint).
		SetProviderRequestID("refund-" + refundID.String()).
		SetCurrency("CNY").
		SetRequestedPrincipalAmount(requestedPrincipal).
		SetPrincipalAmount(decimalFloat(calculation.amounts.PrincipalDelta, refundMoneyScale)).
		SetFeeAmount(decimalFloat(calculation.amounts.FeeDelta, refundMoneyScale)).
		SetRefundFeeRate(calculation.quote.RefundFeeRate).
		SetRefundFeeAmount(decimalFloat(calculation.amounts.RefundFeeDelta, refundMoneyScale)).
		SetGatewayAmount(decimalFloat(calculation.amounts.GatewayDelta, refundMoneyScale)).
		SetBasePoints(decimalFloat(calculation.amounts.RechargePointsDelta, refundPointsScale)).
		SetBonusPoints(decimalFloat(calculation.amounts.BonusPointsDelta, refundPointsScale)).
		SetAffiliateRebatePoints(decimalFloat(calculation.affDelta, refundPointsScale)).
		SetBonusExpiredOffset(calculation.quote.BonusExpiredOffset).
		SetTargetPrincipalAmount(decimalFloat(calculation.amounts.TargetPrincipal, refundMoneyScale)).
		SetTargetFeeAmount(decimalFloat(calculation.amounts.TargetFee, refundMoneyScale)).
		SetTargetRefundFeeAmount(decimalFloat(calculation.amounts.TargetRefundFee, refundMoneyScale)).
		SetTargetBasePoints(decimalFloat(calculation.amounts.TargetRechargePoints, refundPointsScale)).
		SetTargetBonusPoints(decimalFloat(calculation.amounts.TargetBonusPoints, refundPointsScale)).
		SetTargetAffiliatePoints(decimalFloat(calculation.affTarget, refundPointsScale)).
		SetNillableRequestedBy(optionalString(requestedBy)).
		SetReason(reason).
		Save(txCtx)
	if err != nil {
		return nil, false, fmt.Errorf("create refund: %w", err)
	}

	walletRepo, ok := s.userRepo.(RefundWalletRepository)
	if !ok {
		return nil, false, infraerrors.InternalServer("REFUND_WALLET_UNAVAILABLE", "refund point reservation is unavailable")
	}
	hold, err := walletRepo.HoldRefundPoints(txCtx, RefundPointHoldInput{
		UserID: order.UserID, RefundID: refund.ID,
		BonusGrantID: psStringValue(order.BonusGrantID),
		BasePoints:   refund.BasePoints, BonusPoints: refund.BonusPoints,
		BonusExpiredOffset: refund.BonusExpiredOffset,
		RequestFingerprint: fingerprint, Notes: "reserve points for payment refund",
	})
	if err != nil {
		if errors.Is(err, ErrInsufficientBalance) {
			return nil, false, infraerrors.Conflict("REFUND_POINT_BALANCE_CHANGED", "available points changed while preparing the refund; request a new quote")
		}
		if errors.Is(err, ErrWalletRefundBonusFrozen) {
			return nil, false, infraerrors.Conflict("REFUND_BONUS_FROZEN", "recharge bonus points are frozen by another operation")
		}
		return nil, false, fmt.Errorf("reserve refund points: %w", err)
	}
	if input.AutoAffiliate && refund.AffiliateRebatePoints > 0 {
		reservation, reserveErr := reserveAffiliateReversal(txCtx, tx.Client(), order.ID, refund.AffiliateRebatePoints)
		if reserveErr != nil {
			return nil, false, fmt.Errorf("reserve affiliate rebate reversal: %w", reserveErr)
		}
		if _, holdErr := holdAffiliateWalletReservation(txCtx, s.userRepo, reservation, refund.ID, fingerprint); holdErr != nil {
			return nil, false, fmt.Errorf("hold transferred affiliate rebate: %w", holdErr)
		}
	}
	update := tx.PaymentRefund.UpdateOneID(refund.ID).SetStatus(RefundStatusReserved)
	if hold.HoldID != "" {
		update.SetWalletHoldID(hold.HoldID)
	}
	refund, err = update.Save(txCtx)
	if err != nil {
		return nil, false, fmt.Errorf("mark refund reserved: %w", err)
	}
	if _, err := tx.PaymentOrder.UpdateOneID(order.ID).
		SetStatus(OrderStatusRefunding).
		SetRefundRequestedAt(time.Now()).
		SetRefundRequestReason(reason).
		SetNillableRefundRequestedBy(optionalString(requestedBy)).
		Save(txCtx); err != nil {
		return nil, false, fmt.Errorf("mark order refunding: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit refund reservation: %w", err)
	}
	return refund, false, nil
}

func (s *PaymentService) validateRechargeRefundOrder(ctx context.Context, order *dbent.PaymentOrder, selfService bool) error {
	if order == nil {
		return infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if order.OrderType != payment.OrderTypeBalance {
		return infraerrors.BadRequest("INVALID_ORDER_TYPE", "only point recharge orders can be refunded")
	}
	allowed := order.Status == OrderStatusCompleted || order.Status == OrderStatusPartiallyRefunded || order.Status == OrderStatusRefundFailed
	if !allowed {
		return infraerrors.BadRequest("INVALID_STATUS", "order status does not allow a new refund")
	}
	if selfService {
		instance, err := s.getRefundOrderProviderInstance(ctx, order)
		if err != nil || instance == nil || !instance.AllowUserRefund {
			return infraerrors.Forbidden("USER_REFUND_DISABLED", "self-service refund is not enabled for this payment provider")
		}
	} else if !order.WalletOnly {
		instance, err := s.getRefundOrderProviderInstance(ctx, order)
		if err != nil || instance == nil || !instance.RefundEnabled {
			return infraerrors.Forbidden("REFUND_DISABLED", "refund is not enabled for this payment provider")
		}
	}
	return nil
}

func (s *PaymentService) calculateRefundQuote(ctx context.Context, client *dbent.Client, order *dbent.PaymentOrder, requested *float64, autoAffiliate bool) (*refundCalculation, error) {
	input, err := cumulativeRefundInputForOrder(order)
	if err != nil {
		return nil, infraerrors.BadRequest("REFUND_SNAPSHOT_MISSING", err.Error())
	}
	refundFeeRate := defaultRefundFeeRate
	if s.configService != nil {
		paymentConfig, configErr := s.configService.GetPaymentConfig(ctx)
		if configErr != nil {
			return nil, fmt.Errorf("get refund fee configuration: %w", configErr)
		}
		refundFeeRate = paymentConfig.RefundFeeRate
	}
	input.RefundFeeRate = decimal.NewFromFloat(refundFeeRate).Round(2)
	remaining := input.PrincipalAmount.Sub(input.PreviousPrincipal).Round(refundMoneyScale)
	requestedAmount := remaining
	if requested != nil {
		if math.IsNaN(*requested) || math.IsInf(*requested, 0) || *requested <= 0 {
			return nil, infraerrors.BadRequest("INVALID_AMOUNT", "principal_amount must be positive")
		}
		if !hasRefundMoneyPrecision(*requested) {
			return nil, infraerrors.BadRequest("INVALID_AMOUNT", "principal_amount allows at most 2 decimal places")
		}
		requestedAmount = decimal.NewFromFloat(*requested).Round(refundMoneyScale)
		if requestedAmount.GreaterThan(remaining) {
			requestedAmount = remaining
		}
	}
	if !remaining.IsPositive() {
		return nil, infraerrors.BadRequest("ALREADY_REFUNDED", "order principal has already been fully refunded")
	}

	walletRepo, ok := s.userRepo.(RefundWalletRepository)
	if !ok {
		return nil, infraerrors.InternalServer("REFUND_WALLET_UNAVAILABLE", "refund point reservation is unavailable")
	}
	capacity, err := walletRepo.GetRefundPointCapacity(ctx, order.UserID, psStringValue(order.BonusGrantID))
	if err != nil && !errors.Is(err, ErrWalletBonusGrantNotFound) {
		return nil, fmt.Errorf("get refundable point capacity: %w", err)
	}
	if errors.Is(err, ErrWalletBonusGrantNotFound) {
		capacity = RefundPointCapacity{RechargeAvailable: walletRechargeBalance(ctx, s.userRepo, order.UserID)}
	}
	expiredUsed, err := successfulRefundExpiredOffset(ctx, client, order.ID)
	if err != nil {
		return nil, fmt.Errorf("query prior expired bonus offsets: %w", err)
	}
	expiredAvailable := decimal.NewFromFloat(math.Max(capacity.SourceBonusExpired-expiredUsed, 0)).Round(refundPointsScale)
	sourceAvailable := decimal.NewFromFloat(capacity.SourceBonusAvailable).Round(refundPointsScale)
	rechargeAvailable := decimal.NewFromFloat(capacity.RechargeAvailable).Round(refundPointsScale)

	// remaining_amount already excludes the frozen part of the source grant.
	// A partial freeze therefore reduces sourceAvailable, while recharge points
	// may still cover the bonus shortfall.
	maxPrincipal := maxAffordableRefundPrincipal(input, remaining, rechargeAvailable, sourceAvailable, expiredAvailable, true)
	actual := requestedAmount
	if actual.GreaterThan(maxPrincipal) {
		actual = maxPrincipal
	}
	if !actual.IsPositive() {
		return &refundCalculation{quote: RefundQuote{
			OrderID: order.ID, Currency: "CNY",
			RequestedPrincipalAmount:     decimalFloat(requestedAmount, refundMoneyScale),
			RemainingPrincipalAmount:     decimalFloat(remaining, refundMoneyScale),
			MaxRefundablePrincipalAmount: 0, RefundDeadline: refundDeadline(order),
			SelfServiceEligible: refundSelfServiceEligible(order, time.Now()), RequiresTicket: true,
			BlockedReason: "INSUFFICIENT_POINTS",
		}}, nil
	}
	input.RequestedPrincipal = actual
	amounts, err := CalculateCumulativeRefundAmounts(input)
	if err != nil {
		return nil, infraerrors.BadRequest("INVALID_REFUND_STATE", err.Error())
	}
	expiredOffset := decimal.Min(amounts.BonusPointsDelta, expiredAvailable).Round(refundPointsScale)
	pointsToHold := amounts.PointsDelta.Sub(expiredOffset).Round(refundPointsScale)
	previousAffiliate := decimal.NewFromFloat(order.ReversedAffiliatePoints).Round(refundPointsScale)
	affTarget := previousAffiliate
	if autoAffiliate {
		affOriginal := decimal.NewFromFloat(affiliateRebateForOrder(ctx, client, order)).Round(refundPointsScale)
		affTarget = proportionalRefundTarget(affOriginal, amounts.TargetPrincipal, input.PrincipalAmount, refundPointsScale)
	}
	affDelta := affTarget.Sub(previousAffiliate)
	if affDelta.IsNegative() {
		affDelta = decimal.Zero
	}
	deadline := refundDeadline(order)
	eligible := refundSelfServiceEligible(order, time.Now())
	return &refundCalculation{
		quote: RefundQuote{
			OrderID: order.ID, Currency: "CNY",
			RequestedPrincipalAmount:     decimalFloat(requestedAmount, refundMoneyScale),
			PrincipalAmount:              decimalFloat(amounts.PrincipalDelta, refundMoneyScale),
			FeeAmount:                    decimalFloat(amounts.FeeDelta, refundMoneyScale),
			RefundFeeRate:                refundFeeRate,
			RefundFeeAmount:              decimalFloat(amounts.RefundFeeDelta, refundMoneyScale),
			GatewayAmount:                decimalFloat(amounts.GatewayDelta, refundMoneyScale),
			BasePoints:                   decimalFloat(amounts.RechargePointsDelta, refundPointsScale),
			BonusPoints:                  decimalFloat(amounts.BonusPointsDelta, refundPointsScale),
			PointsToHold:                 decimalFloat(pointsToHold, refundPointsScale),
			BonusExpiredOffset:           decimalFloat(expiredOffset, refundPointsScale),
			AffiliateRebatePoints:        decimalFloat(affDelta, refundPointsScale),
			RemainingPrincipalAmount:     decimalFloat(remaining, refundMoneyScale),
			MaxRefundablePrincipalAmount: decimalFloat(maxPrincipal, refundMoneyScale),
			RefundDeadline:               deadline, SelfServiceEligible: eligible, RequiresTicket: !eligible,
		},
		amounts: amounts, affTarget: affTarget, affDelta: affDelta,
	}, nil
}

func hasRefundMoneyPrecision(value float64) bool {
	amount := decimal.NewFromFloat(value)
	return amount.Equal(amount.Round(refundMoneyScale))
}

func cumulativeRefundInputForOrder(order *dbent.PaymentOrder) (CumulativeRefundInput, error) {
	if order == nil {
		return CumulativeRefundInput{}, fmt.Errorf("order is missing")
	}
	principal := decimal.NewFromFloat(order.PrincipalAmount).Round(refundMoneyScale)
	fee := decimal.NewFromFloat(order.FeeAmount).Round(refundMoneyScale)
	basePoints := decimal.NewFromFloat(order.BasePoints).Round(refundPointsScale)
	bonusPoints := decimal.NewFromFloat(order.BonusPoints).Round(refundPointsScale)
	if !principal.IsPositive() {
		principal = decimal.NewFromFloat(order.GatewayBaseAmount).Round(refundMoneyScale)
		if !principal.IsPositive() {
			return CumulativeRefundInput{}, fmt.Errorf("order does not contain an immutable RMB principal snapshot")
		}
		fee = decimal.NewFromFloat(math.Max(order.PayAmount-order.GatewayBaseAmount, 0)).Round(refundMoneyScale)
		basePoints = decimal.NewFromFloat(order.Amount).Round(refundPointsScale)
		bonusPoints = decimal.Zero
	}
	return CumulativeRefundInput{
		PrincipalAmount: principal, FeeAmount: fee,
		RechargePoints: basePoints, BonusPoints: bonusPoints,
		PreviousPrincipal:      decimal.NewFromFloat(order.RefundedPrincipalAmount).Round(refundMoneyScale),
		PreviousFee:            decimal.NewFromFloat(order.RefundedFeeAmount).Round(refundMoneyScale),
		PreviousRefundFee:      decimal.NewFromFloat(order.RefundedPrincipalAmount + order.RefundedFeeAmount - order.RefundedGatewayAmount).Round(refundMoneyScale),
		PreviousGateway:        decimal.NewFromFloat(order.RefundedGatewayAmount).Round(refundMoneyScale),
		PreviousRechargePoints: decimal.NewFromFloat(order.ReversedBasePoints).Round(refundPointsScale),
		PreviousBonusPoints:    decimal.NewFromFloat(order.ReversedBonusPoints).Round(refundPointsScale),
	}, nil
}

func successfulRefundExpiredOffset(ctx context.Context, client *dbent.Client, orderID string) (float64, error) {
	refunds, err := client.PaymentRefund.Query().Where(
		paymentrefund.OrderIDEQ(orderID), paymentrefund.StatusEQ(RefundStatusSucceeded),
	).All(ctx)
	if err != nil {
		return 0, err
	}
	var total decimal.Decimal
	for _, refund := range refunds {
		total = total.Add(decimal.NewFromFloat(refund.BonusExpiredOffset))
	}
	return decimalFloat(total, refundPointsScale), nil
}

func refundDeadline(order *dbent.PaymentOrder) *time.Time {
	if order == nil {
		return nil
	}
	if order.RefundDeadline != nil {
		deadline := order.RefundDeadline.UTC()
		return &deadline
	}
	if order.CompletedAt == nil {
		return nil
	}
	deadline := order.CompletedAt.UTC().Add(selfServiceRefundWindow)
	return &deadline
}

func refundSelfServiceEligible(order *dbent.PaymentOrder, now time.Time) bool {
	if order == nil || order.CompletedAt == nil {
		return false
	}
	deadline := refundDeadline(order)
	return deadline != nil && withinSelfServiceRefundWindow(order.CompletedAt.UTC(), *deadline, now.UTC())
}

func refundRequestFingerprint(input CreatePaymentRefundInput) string {
	principal := "ALL"
	if input.Principal != nil {
		principal = decimal.NewFromFloat(*input.Principal).Round(refundMoneyScale).StringFixed(refundMoneyScale)
	}
	payload := strings.Join([]string{
		strings.TrimSpace(input.OrderID), strings.TrimSpace(input.UserID), strings.TrimSpace(input.Source),
		strings.TrimSpace(input.TicketID), principal, strings.TrimSpace(input.Reason),
	}, "\x00")
	digest := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(digest[:])
}

func decimalFloat(value decimal.Decimal, scale int32) float64 {
	result, _ := value.Round(scale).Float64()
	return result
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func walletRechargeBalance(ctx context.Context, repo UserRepository, userID string) float64 {
	summary, err := repo.GetWalletSummary(ctx, userID)
	if err != nil {
		return 0
	}
	return summary.RechargeBalance
}

func affiliateRebateForOrder(ctx context.Context, client *dbent.Client, order *dbent.PaymentOrder) float64 {
	if order == nil {
		return 0
	}
	total := order.AffiliateRebatePoints
	rows, err := client.QueryContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM user_affiliate_ledger
		WHERE source_order_id = $1 AND action = 'accrue'
	`, order.ID)
	if err != nil {
		return total
	}
	defer rows.Close() //nolint:errcheck
	var stored float64
	if rows.Next() && rows.Scan(&stored) == nil && stored > total {
		total = stored
	}
	return total
}

type affiliateReversalReservation struct {
	InviterID    string
	WalletTarget float64
}

func affiliateWalletReservationTarget(amount, available, frozen, totalReserved float64) float64 {
	requested := decimal.NewFromFloat(amount).Round(refundPointsScale)
	reservedBefore := decimal.NewFromFloat(totalReserved).Round(refundPointsScale).Sub(requested)
	if reservedBefore.IsNegative() {
		reservedBefore = decimal.Zero
	}
	poolBalance := decimal.NewFromFloat(math.Max(available, 0) + math.Max(frozen, 0)).Round(refundPointsScale)
	poolAvailable := poolBalance.Sub(reservedBefore)
	if poolAvailable.IsNegative() {
		poolAvailable = decimal.Zero
	}
	return decimalFloat(requested.Sub(decimal.Min(requested, poolAvailable)), refundPointsScale)
}

func holdAffiliateWalletReservation(ctx context.Context, repo WalletRepository, reservation affiliateReversalReservation, refundID, fingerprint string) (WalletHoldResult, error) {
	walletTarget := decimal.NewFromFloat(reservation.WalletTarget).Round(refundPointsScale)
	if reservation.InviterID == "" || !walletTarget.IsPositive() {
		return WalletHoldResult{Status: "not_required"}, nil
	}
	summary, err := repo.GetWalletSummary(ctx, reservation.InviterID)
	if err != nil {
		return WalletHoldResult{}, fmt.Errorf("load affiliate wallet capacity: %w", err)
	}
	holdAmount := decimal.Min(walletTarget, decimal.NewFromFloat(math.Max(summary.AvailableBalance, 0)).Round(refundPointsScale))
	if !holdAmount.IsPositive() {
		return WalletHoldResult{Status: "not_required"}, nil
	}
	return repo.HoldWallet(ctx, WalletHoldInput{
		UserID: reservation.InviterID, Amount: decimalFloat(holdAmount, refundPointsScale),
		Purpose: affiliateRefundHoldPurpose, ReferenceID: refundID,
		RequestFingerprint: fingerprint,
		Notes:              "reserve transferred affiliate rebate for payment refund",
	})
}

func reserveAffiliateReversal(ctx context.Context, client *dbent.Client, orderID string, amount float64) (affiliateReversalReservation, error) {
	var reservation affiliateReversalReservation
	remaining := decimal.NewFromFloat(amount).Round(refundPointsScale)
	if !remaining.IsPositive() {
		return reservation, nil
	}
	rows, err := client.QueryContext(ctx, `
		SELECT id, user_id, amount - reserved_reversal_amount - reversed_amount
		FROM user_affiliate_ledger
		WHERE source_order_id = $1 AND action = 'accrue'
		ORDER BY created_at, id
		FOR UPDATE
	`, orderID)
	if err != nil {
		return reservation, err
	}
	type allocation struct {
		id, inviterID string
		available     float64
	}
	allocations := make([]allocation, 0)
	for rows.Next() {
		var item allocation
		if err := rows.Scan(&item.id, &item.inviterID, &item.available); err != nil {
			_ = rows.Close()
			return reservation, err
		}
		if reservation.InviterID == "" {
			reservation.InviterID = item.inviterID
		} else if reservation.InviterID != item.inviterID {
			_ = rows.Close()
			return reservation, fmt.Errorf("affiliate rebate accrual has multiple inviters")
		}
		allocations = append(allocations, item)
	}
	if err := rows.Close(); err != nil {
		return reservation, err
	}
	for _, item := range allocations {
		if !remaining.IsPositive() {
			break
		}
		part := decimal.Min(remaining, decimal.NewFromFloat(math.Max(item.available, 0)).Round(refundPointsScale))
		if !part.IsPositive() {
			continue
		}
		if _, err := client.ExecContext(ctx, `
			UPDATE user_affiliate_ledger
			SET reserved_reversal_amount = reserved_reversal_amount + $1, updated_at = NOW()
			WHERE id = $2
		`, decimalFloat(part, refundPointsScale), item.id); err != nil {
			return reservation, err
		}
		remaining = remaining.Sub(part)
	}
	if remaining.IsPositive() {
		return reservation, fmt.Errorf("affiliate rebate reversal exceeds the unreversed accrual")
	}
	if reservation.InviterID == "" {
		return reservation, fmt.Errorf("affiliate rebate reversal has no inviter")
	}

	profileRows, err := client.QueryContext(ctx, `
		SELECT aff_quota, aff_frozen_quota
		FROM user_affiliates
		WHERE user_id = $1
		FOR UPDATE
	`, reservation.InviterID)
	if err != nil {
		return reservation, err
	}
	var available, frozen float64
	if !profileRows.Next() {
		_ = profileRows.Close()
		return reservation, fmt.Errorf("affiliate profile not found for reversal")
	}
	if err := profileRows.Scan(&available, &frozen); err != nil {
		_ = profileRows.Close()
		return reservation, err
	}
	if err := profileRows.Close(); err != nil {
		return reservation, err
	}

	reservedRows, err := client.QueryContext(ctx, `
		SELECT COALESCE(SUM(reserved_reversal_amount), 0)
		FROM user_affiliate_ledger
		WHERE user_id = $1 AND action = 'accrue'
	`, reservation.InviterID)
	if err != nil {
		return reservation, err
	}
	var totalReserved float64
	if !reservedRows.Next() {
		_ = reservedRows.Close()
		return reservation, fmt.Errorf("query affiliate reversal reservations returned no row")
	}
	if err := reservedRows.Scan(&totalReserved); err != nil {
		_ = reservedRows.Close()
		return reservation, err
	}
	if err := reservedRows.Close(); err != nil {
		return reservation, err
	}

	reservation.WalletTarget = affiliateWalletReservationTarget(amount, available, frozen, totalReserved)
	return reservation, nil
}

func (s *PaymentService) executePaymentRefund(ctx context.Context, refundID string) (*dbent.PaymentRefund, error) {
	refund, err := s.entClient.PaymentRefund.Get(ctx, refundID)
	if err != nil {
		return nil, infraerrors.NotFound("REFUND_NOT_FOUND", "refund not found")
	}
	switch refund.Status {
	case RefundStatusSucceeded, RefundStatusPending, RefundStatusFailed:
		return refund, nil
	case RefundStatusSubmitting:
		// Another caller already owns the provider submission. Returning the
		// current state keeps idempotent retries from issuing a duplicate call.
		return refund, nil
	case RefundStatusReserved:
	default:
		return nil, infraerrors.Conflict("REFUND_STATE_CONFLICT", "refund is not ready for submission")
	}
	claimed, err := s.entClient.PaymentRefund.Update().Where(
		paymentrefund.IDEQ(refund.ID),
		paymentrefund.StatusEQ(RefundStatusReserved),
	).SetStatus(RefundStatusSubmitting).SetSubmittedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("claim refund submission: %w", err)
	}
	if claimed == 0 {
		return s.entClient.PaymentRefund.Get(ctx, refund.ID)
	}
	refund, err = s.entClient.PaymentRefund.Get(ctx, refund.ID)
	if err != nil {
		return nil, fmt.Errorf("reload refund submission: %w", err)
	}
	return s.submitClaimedPaymentRefund(ctx, refund)
}

func (s *PaymentService) retryStaleSubmittingPaymentRefund(ctx context.Context, refundID string, staleBefore time.Time) (*dbent.PaymentRefund, error) {
	claimed, err := s.entClient.PaymentRefund.Update().Where(
		paymentrefund.IDEQ(refundID),
		paymentrefund.StatusEQ(RefundStatusSubmitting),
		paymentrefund.Or(paymentrefund.SubmittedAtIsNil(), paymentrefund.SubmittedAtLTE(staleBefore)),
	).SetSubmittedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("claim stale refund submission: %w", err)
	}
	refund, err := s.entClient.PaymentRefund.Get(ctx, refundID)
	if err != nil {
		return nil, fmt.Errorf("reload stale refund submission: %w", err)
	}
	if claimed == 0 {
		return refund, nil
	}
	// A stale SUBMITTING record may represent a provider call whose response was
	// lost. Recover it through QueryRefund, whose provider contract uses the
	// stable RequestID, instead of issuing another raw refund request here.
	response, err := s.queryPaymentRefundByID(ctx, refund.ID)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return refund, nil
	}
	return s.entClient.PaymentRefund.Get(ctx, response.ID)
}

func (s *PaymentService) submitClaimedPaymentRefund(ctx context.Context, refund *dbent.PaymentRefund) (*dbent.PaymentRefund, error) {
	order, err := s.entClient.PaymentOrder.Get(ctx, refund.OrderID)
	if err != nil {
		return nil, fmt.Errorf("load refund order: %w", err)
	}
	if strings.TrimSpace(order.PaymentTradeNo) == "" {
		return s.failPaymentRefund(ctx, refund.ID, "MISSING_TRADE_NO", "payment order has no provider trade number")
	}
	provider, err := s.getRefundProvider(ctx, order)
	if err != nil {
		// Provider resolution happens before any external refund request, so this
		// failure is definitive and all point reservations can be released.
		return s.failPaymentRefund(ctx, refund.ID, "PROVIDER_UNAVAILABLE", err.Error())
	}
	if err := validateProviderSnapshotMetadata(order, provider.ProviderKey(), providerMerchantIdentityMetadata(provider)); err != nil {
		return s.failPaymentRefund(ctx, refund.ID, "PROVIDER_METADATA_MISMATCH", err.Error())
	}
	if _, ok := provider.(payment.RefundQueryProvider); !ok {
		// A provider can enter the uncertain state whenever the request reaches
		// upstream but its response is lost. Only providers that can reconcile the
		// stable RequestID are allowed to receive a refund request.
		return s.failPaymentRefund(ctx, refund.ID, "REFUND_RECOVERY_UNSUPPORTED", "payment provider does not support safe refund reconciliation")
	}
	response, callErr := provider.Refund(ctx, payment.RefundRequest{
		TradeNo: order.PaymentTradeNo, OrderID: order.OutTradeNo,
		RequestID: refund.ProviderRequestID,
		Amount:    payment.FormatAmountForCurrency(refund.GatewayAmount, refund.Currency),
		Reason:    refund.Reason,
	})
	if callErr != nil {
		if isDefinitiveProviderRefundFailure(response) {
			return s.failPaymentRefund(ctx, refund.ID, "PROVIDER_REFUND_FAILED", callErr.Error())
		}
		return s.markRefundSubmissionUncertain(ctx, refund.ID, response, "PROVIDER_SUBMISSION_UNCERTAIN", callErr.Error())
	}
	if response == nil {
		return s.markRefundSubmissionUncertain(ctx, refund.ID, nil, "PROVIDER_SUBMISSION_UNCERTAIN", "payment provider returned an empty refund response")
	}
	switch strings.TrimSpace(response.Status) {
	case payment.ProviderStatusSuccess, payment.ProviderStatusRefunded:
		return s.settleProviderConfirmedPaymentRefund(ctx, refund.ID, strings.TrimSpace(response.RefundID))
	case payment.ProviderStatusPending:
		return s.markRefundPendingRecord(ctx, refund.ID, strings.TrimSpace(response.RefundID), "", "")
	case payment.ProviderStatusFailed:
		return s.failPaymentRefund(ctx, refund.ID, "PROVIDER_REFUND_FAILED", "payment provider rejected the refund")
	default:
		return s.markRefundSubmissionUncertain(ctx, refund.ID, response, "PROVIDER_STATUS_UNKNOWN", "payment provider returned an unknown refund status")
	}
}

func isDefinitiveProviderRefundFailure(response *payment.RefundResponse) bool {
	return response != nil && strings.TrimSpace(response.Status) == payment.ProviderStatusFailed
}

func (s *PaymentService) markRefundSubmissionUncertain(ctx context.Context, refundID string, response *payment.RefundResponse, code, message string) (*dbent.PaymentRefund, error) {
	providerRefundID := ""
	if response != nil {
		providerRefundID = strings.TrimSpace(response.RefundID)
	}
	return s.markRefundPendingRecord(ctx, refundID, providerRefundID, code, message)
}

func paymentRefundPendingOrderUpdate(update *dbent.PaymentOrderUpdateOne) *dbent.PaymentOrderUpdateOne {
	// Successful refund aggregates remain untouched until settlement succeeds.
	return update.SetStatus(OrderStatusRefundPending)
}

func (s *PaymentService) markRefundPendingRecord(ctx context.Context, refundID, providerRefundID, code, message string) (*dbent.PaymentRefund, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin pending refund update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	refund, err := tx.PaymentRefund.Query().Where(paymentrefund.IDEQ(refundID)).ForUpdate().Only(txCtx)
	if err != nil {
		return nil, fmt.Errorf("lock pending refund: %w", err)
	}
	if refund.Status == RefundStatusSucceeded || refund.Status == RefundStatusFailed {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return refund, nil
	}
	update := tx.PaymentRefund.UpdateOneID(refund.ID).
		SetStatus(RefundStatusPending).
		SetErrorCode(code).
		SetErrorMessage(message)
	if providerRefundID != "" {
		update.SetProviderRefundID(providerRefundID)
	}
	refund, err = update.Save(txCtx)
	if err != nil {
		return nil, fmt.Errorf("mark refund pending: %w", err)
	}
	if _, err := paymentRefundPendingOrderUpdate(tx.PaymentOrder.UpdateOneID(refund.OrderID)).Save(txCtx); err != nil {
		return nil, fmt.Errorf("mark order refund pending: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit pending refund: %w", err)
	}
	return refund, nil
}

func (s *PaymentService) settleProviderConfirmedPaymentRefund(ctx context.Context, refundID, providerRefundID string) (*dbent.PaymentRefund, error) {
	settlementCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), refundSettlementTimeout)
	defer cancel()
	refund, err := s.settlePaymentRefund(settlementCtx, refundID, providerRefundID)
	if err == nil {
		return refund, nil
	}
	current, loadErr := s.entClient.PaymentRefund.Get(settlementCtx, refundID)
	if loadErr == nil && current.Status == RefundStatusSucceeded {
		return current, nil
	}
	pending, pendingErr := s.markRefundPendingRecord(settlementCtx, refundID, providerRefundID, "REFUND_SETTLEMENT_RETRY", err.Error())
	if pendingErr != nil {
		return nil, errors.Join(
			fmt.Errorf("settle provider-confirmed refund: %w", err),
			fmt.Errorf("preserve provider-confirmed refund for reconciliation: %w", pendingErr),
		)
	}
	if pending.Status == RefundStatusFailed {
		return nil, infraerrors.Conflict("REFUND_STATE_CONFLICT", "provider confirmed a refund that is already marked failed").WithCause(err)
	}
	slog.Warn("payment provider confirmed refund before local settlement completed",
		"refund_id", refundID,
		"provider_refund_id", providerRefundID,
		"error", err,
	)
	return pending, nil
}

func settledPaymentOrderUpdate(update *dbent.PaymentOrderUpdateOne, order *dbent.PaymentOrder, refund *dbent.PaymentRefund, now time.Time) *dbent.PaymentOrderUpdateOne {
	status := OrderStatusPartiallyRefunded
	principalTotal, snapshotErr := cumulativeRefundInputForOrder(order)
	if snapshotErr == nil && decimal.NewFromFloat(refund.TargetPrincipalAmount).Round(refundMoneyScale).Equal(principalTotal.PrincipalAmount) {
		status = OrderStatusRefunded
	}
	return update.
		SetStatus(status).
		SetRefundedPrincipalAmount(refund.TargetPrincipalAmount).
		SetRefundedFeeAmount(refund.TargetFeeAmount).
		SetRefundedGatewayAmount(refund.TargetPrincipalAmount + refund.TargetFeeAmount - refund.TargetRefundFeeAmount).
		SetReversedBasePoints(refund.TargetBasePoints).
		SetReversedBonusPoints(refund.TargetBonusPoints).
		SetReversedAffiliatePoints(refund.TargetAffiliatePoints).
		SetRefundAmount(refund.TargetPrincipalAmount).
		SetRefundReason(refund.Reason).
		SetRefundAt(now).
		ClearFailedAt().
		ClearFailedReason()
}

func (s *PaymentService) completeSettledRefundTicket(ctx context.Context, refund *dbent.PaymentRefund, now time.Time) {
	if refund == nil || refund.TicketID == nil || strings.TrimSpace(*refund.TicketID) == "" {
		return
	}
	ticket, err := s.entClient.SupportTicket.Get(ctx, *refund.TicketID)
	if err != nil {
		if !dbent.IsNotFound(err) {
			slog.Warn("failed to load settled refund ticket", "refund_id", refund.ID, "ticket_id", *refund.TicketID, "error", err)
		}
		return
	}
	update := s.entClient.SupportTicket.UpdateOneID(ticket.ID).
		SetRefundID(refund.ID).
		SetRefundDecision(SupportTicketRefundApproved)
	if ticket.Status != SupportTicketStatusClosed && ticket.Status != SupportTicketStatusCancelled {
		update.SetStatus(SupportTicketStatusResolved).SetResolvedAt(now)
	}
	if _, err := update.Save(ctx); err != nil {
		slog.Warn("failed to complete settled refund ticket", "refund_id", refund.ID, "ticket_id", ticket.ID, "error", err)
	}
}

func (s *PaymentService) settlePaymentRefund(ctx context.Context, refundID, providerRefundID string) (_ *dbent.PaymentRefund, err error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin refund settlement: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	txCtx := dbent.NewTxContext(ctx, tx)
	refund, err := tx.PaymentRefund.Query().Where(paymentrefund.IDEQ(refundID)).ForUpdate().Only(txCtx)
	if err != nil {
		return nil, fmt.Errorf("lock refund settlement: %w", err)
	}
	if refund.Status == RefundStatusFailed {
		return nil, infraerrors.Conflict("REFUND_STATE_CONFLICT", "failed refund cannot be settled")
	}
	order, err := tx.PaymentOrder.Query().Where(paymentorder.IDEQ(refund.OrderID)).ForUpdate().Only(txCtx)
	if err != nil {
		return nil, fmt.Errorf("lock refund order settlement: %w", err)
	}
	now := time.Now()
	if refund.Status == RefundStatusSucceeded {
		active, queryErr := tx.PaymentRefund.Query().Where(
			paymentrefund.OrderIDEQ(order.ID),
			paymentrefund.StatusIn(RefundStatusRequested, RefundStatusReserved, RefundStatusSubmitting, RefundStatusPending),
		).Exist(txCtx)
		if queryErr != nil {
			return nil, fmt.Errorf("query active refund during repair: %w", queryErr)
		}
		if !active {
			projection, projectionErr := tx.PaymentRefund.Query().Where(
				paymentrefund.OrderIDEQ(order.ID),
				paymentrefund.StatusEQ(RefundStatusSucceeded),
			).Order(paymentrefund.ByCreatedAt(sql.OrderDesc())).First(txCtx)
			if projectionErr != nil {
				return nil, fmt.Errorf("load latest settled refund projection: %w", projectionErr)
			}
			if _, updateErr := settledPaymentOrderUpdate(tx.PaymentOrder.UpdateOneID(order.ID), order, projection, now).Save(txCtx); updateErr != nil {
				return nil, fmt.Errorf("repair refunded order totals: %w", updateErr)
			}
		}
		if providerRefundID != "" && psStringValue(refund.ProviderRefundID) != providerRefundID {
			refund, err = tx.PaymentRefund.UpdateOneID(refund.ID).SetProviderRefundID(providerRefundID).Save(txCtx)
			if err != nil {
				return nil, fmt.Errorf("update settled refund provider id: %w", err)
			}
		}
		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit settled refund repair: %w", err)
		}
		s.completeSettledRefundTicket(ctx, refund, now)
		return refund, nil
	}
	if refund.WalletHoldID != nil && strings.TrimSpace(*refund.WalletHoldID) != "" {
		if _, err := s.userRepo.CaptureWalletHold(txCtx, *refund.WalletHoldID, "wallet-refund-capture:"+refund.ID); err != nil {
			return nil, fmt.Errorf("capture refund point hold: %w", err)
		}
	}
	affiliateWalletUserID := ""
	if refund.AffiliateRebatePoints > 0 {
		affiliateWalletUserID, err = s.captureAffiliateReversal(txCtx, tx.Client(), refund, order)
		if err != nil {
			return nil, fmt.Errorf("capture affiliate rebate reversal: %w", err)
		}
	}

	if _, err := settledPaymentOrderUpdate(tx.PaymentOrder.UpdateOneID(order.ID), order, refund, now).Save(txCtx); err != nil {
		return nil, fmt.Errorf("update refunded order totals: %w", err)
	}
	update := tx.PaymentRefund.UpdateOneID(refund.ID).
		SetStatus(RefundStatusSucceeded).
		SetSettledAt(now).
		SetErrorCode("").
		SetErrorMessage("")
	if providerRefundID != "" {
		update.SetProviderRefundID(providerRefundID)
	}
	refund, err = update.Save(txCtx)
	if err != nil {
		return nil, fmt.Errorf("mark refund succeeded: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit refund settlement: %w", err)
	}
	s.invalidateRefundWalletCaches(ctx, order.UserID)
	if affiliateWalletUserID != "" && affiliateWalletUserID != order.UserID {
		s.invalidateRefundWalletCaches(ctx, affiliateWalletUserID)
	}
	s.completeSettledRefundTicket(ctx, refund, now)
	return refund, nil
}

func (s *PaymentService) failPaymentRefund(ctx context.Context, refundID, code, message string) (_ *dbent.PaymentRefund, err error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin refund failure: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	txCtx := dbent.NewTxContext(ctx, tx)
	refund, err := tx.PaymentRefund.Query().Where(paymentrefund.IDEQ(refundID)).ForUpdate().Only(txCtx)
	if err != nil {
		return nil, fmt.Errorf("lock failed refund: %w", err)
	}
	if refund.Status == RefundStatusSucceeded || refund.Status == RefundStatusFailed {
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return refund, nil
	}
	if refund.WalletHoldID != nil && strings.TrimSpace(*refund.WalletHoldID) != "" {
		if _, err := s.userRepo.ReleaseWalletHold(txCtx, *refund.WalletHoldID, "wallet-refund-release:"+refund.ID); err != nil {
			return nil, fmt.Errorf("release failed refund point hold: %w", err)
		}
	}
	affiliateWalletUserID := ""
	if refund.AffiliateRebatePoints > 0 {
		if err := releaseAffiliateReversal(txCtx, tx.Client(), refund.OrderID, refund.AffiliateRebatePoints); err != nil {
			return nil, fmt.Errorf("release affiliate rebate reversal: %w", err)
		}
		affiliateHold, holdErr := loadAffiliateRefundWalletHold(txCtx, tx.Client(), refund.ID)
		if holdErr != nil {
			return nil, fmt.Errorf("load transferred affiliate rebate hold: %w", holdErr)
		}
		if affiliateHold != nil {
			if _, releaseErr := s.userRepo.ReleaseWalletHold(txCtx, affiliateHold.ID, "wallet-affiliate-refund-release:"+refund.ID); releaseErr != nil {
				return nil, fmt.Errorf("release transferred affiliate rebate hold: %w", releaseErr)
			}
			affiliateWalletUserID = affiliateHold.UserID
		}
	}
	now := time.Now()
	refund, err = tx.PaymentRefund.UpdateOneID(refund.ID).
		SetStatus(RefundStatusFailed).
		SetErrorCode(code).
		SetErrorMessage(message).
		Save(txCtx)
	if err != nil {
		return nil, fmt.Errorf("mark refund failed: %w", err)
	}
	order, err := tx.PaymentOrder.Get(txCtx, refund.OrderID)
	if err != nil {
		return nil, fmt.Errorf("load failed refund order: %w", err)
	}
	if _, err := paymentRefundFailedOrderUpdate(tx.PaymentOrder.UpdateOneID(order.ID), now, message).
		Save(txCtx); err != nil {
		return nil, fmt.Errorf("restore order after failed refund: %w", err)
	}
	if refund.TicketID != nil {
		if _, updateErr := tx.SupportTicket.Update().Where(
			supportticket.IDEQ(*refund.TicketID),
			supportticket.StatusIn(SupportTicketStatusInProgress, SupportTicketStatusPendingAdmin),
		).
			SetStatus(SupportTicketStatusPendingAdmin).
			SetRefundID(refund.ID).
			Save(txCtx); updateErr != nil {
			return nil, fmt.Errorf("fail support ticket refund: %w", updateErr)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit refund failure: %w", err)
	}
	s.invalidateRefundWalletCaches(ctx, order.UserID)
	if affiliateWalletUserID != "" && affiliateWalletUserID != order.UserID {
		s.invalidateRefundWalletCaches(ctx, affiliateWalletUserID)
	}
	return refund, nil
}

func paymentRefundFailedOrderUpdate(update *dbent.PaymentOrderUpdateOne, now time.Time, message string) *dbent.PaymentOrderUpdateOne {
	return update.SetStatus(OrderStatusRefundFailed).
		SetFailedAt(now).
		SetFailedReason(message)
}

func (s *PaymentService) QueryPaymentRefund(ctx context.Context, orderID string) (*PaymentRefundResponse, error) {
	refund, err := s.entClient.PaymentRefund.Query().Where(
		paymentrefund.OrderIDEQ(orderID), paymentrefund.StatusIn(RefundStatusSubmitting, RefundStatusPending),
	).Order(paymentrefund.ByCreatedAt(sql.OrderDesc())).First(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, infraerrors.NotFound("PENDING_REFUND_NOT_FOUND", "pending refund not found")
		}
		return nil, err
	}
	return s.queryPaymentRefundByID(ctx, refund.ID)
}

func (s *PaymentService) ReconcileRefundForTicket(ctx context.Context, ticketID, orderID string) (*PaymentRefundResponse, error) {
	ticketID = strings.TrimSpace(ticketID)
	orderID = strings.TrimSpace(orderID)
	if ticketID == "" && orderID == "" {
		return nil, nil
	}
	var refund *dbent.PaymentRefund
	var err error
	if ticketID != "" {
		refund, err = s.entClient.PaymentRefund.Query().Where(paymentrefund.TicketIDEQ(ticketID)).
			Order(paymentrefund.ByCreatedAt(sql.OrderDesc())).First(ctx)
	}
	if (ticketID == "" || dbent.IsNotFound(err)) && orderID != "" {
		refund, err = s.entClient.PaymentRefund.Query().Where(paymentrefund.OrderIDEQ(orderID)).
			Order(paymentrefund.ByCreatedAt(sql.OrderDesc())).First(ctx)
	}
	if err != nil {
		if !dbent.IsNotFound(err) {
			return nil, err
		}
		if orderID == "" {
			return nil, nil
		}
		order, orderErr := s.entClient.PaymentOrder.Get(ctx, orderID)
		if orderErr != nil {
			if dbent.IsNotFound(orderErr) {
				return nil, nil
			}
			return nil, orderErr
		}
		switch order.Status {
		case OrderStatusRefundRequested, OrderStatusRefunding, OrderStatusRefundPending:
			return nil, infraerrors.Conflict("REFUND_RECORD_MISSING", "the order is still refunding but has no recoverable refund record")
		default:
			return nil, nil
		}
	}

	switch refund.Status {
	case RefundStatusSucceeded:
		refund, err = s.settlePaymentRefund(ctx, refund.ID, psStringValue(refund.ProviderRefundID))
		if err != nil {
			return nil, err
		}
		return paymentRefundResponse(refund), nil
	case RefundStatusFailed:
		return paymentRefundResponse(refund), nil
	case RefundStatusReserved:
		refund, err = s.executePaymentRefund(ctx, refund.ID)
		if err != nil {
			return nil, err
		}
		return paymentRefundResponse(refund), nil
	case RefundStatusSubmitting, RefundStatusPending:
		return s.queryPaymentRefundByID(ctx, refund.ID)
	default:
		return nil, infraerrors.Conflict("REFUND_STATE_CONFLICT", "refund is not ready for reconciliation")
	}
}

func (s *PaymentService) queryPaymentRefundByID(ctx context.Context, refundID string) (*PaymentRefundResponse, error) {
	refund, err := s.entClient.PaymentRefund.Get(ctx, refundID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, infraerrors.NotFound("REFUND_NOT_FOUND", "refund not found")
		}
		return nil, err
	}
	if refund.Status != RefundStatusPending && refund.Status != RefundStatusSubmitting {
		return paymentRefundResponse(refund), nil
	}
	order, err := s.entClient.PaymentOrder.Get(ctx, refund.OrderID)
	if err != nil {
		return nil, fmt.Errorf("load pending refund order: %w", err)
	}
	provider, err := s.getRefundProvider(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("get refund provider: %w", err)
	}
	queryProvider, ok := provider.(payment.RefundQueryProvider)
	if !ok {
		return nil, infraerrors.BadRequest("REFUND_QUERY_UNSUPPORTED", "this payment provider does not support refund status query")
	}
	// RefundQueryProvider is the recovery boundary: implementations query the
	// exact refund identity and, when no upstream record exists yet, safely replay
	// creation with the same stable RequestID.
	response, err := queryProvider.QueryRefund(ctx, payment.RefundQueryRequest{
		TradeNo: order.PaymentTradeNo, OrderID: order.OutTradeNo,
		RefundID: psStringValue(refund.ProviderRefundID), RequestID: refund.ProviderRequestID,
		Amount: payment.FormatAmountForCurrency(refund.GatewayAmount, refund.Currency),
	})
	if err != nil {
		return nil, fmt.Errorf("query refund: %w", err)
	}
	if response == nil {
		return paymentRefundResponse(refund), nil
	}
	switch strings.TrimSpace(response.Status) {
	case payment.ProviderStatusSuccess, payment.ProviderStatusRefunded:
		refund, err = s.settleProviderConfirmedPaymentRefund(ctx, refund.ID, strings.TrimSpace(response.RefundID))
	case payment.ProviderStatusPending:
		refund, err = s.markRefundPendingRecord(ctx, refund.ID, strings.TrimSpace(response.RefundID), "", "")
	case payment.ProviderStatusFailed:
		refund, err = s.failPaymentRefund(ctx, refund.ID, "PROVIDER_REFUND_FAILED", "payment provider reported a failed refund")
	default:
		return nil, infraerrors.BadRequest("PROVIDER_STATUS_UNKNOWN", "payment provider returned an unknown refund status")
	}
	if err != nil {
		return nil, err
	}
	return paymentRefundResponse(refund), nil
}

func paymentRefundReconciliationPredicate(staleBefore time.Time) predicate.PaymentRefund {
	return paymentrefund.Or(
		paymentrefund.StatusEQ(RefundStatusPending),
		paymentrefund.And(
			paymentrefund.StatusEQ(RefundStatusReserved),
			paymentrefund.CreatedAtLTE(staleBefore),
		),
		paymentrefund.And(
			paymentrefund.StatusEQ(RefundStatusSubmitting),
			paymentrefund.Or(paymentrefund.SubmittedAtIsNil(), paymentrefund.SubmittedAtLTE(staleBefore)),
		),
	)
}

// ReconcilePendingPaymentRefunds resumes a bounded batch of abandoned local
// reservations and uncertain provider submissions. Query failures leave point
// and affiliate holds intact for a later retry with the same provider identity.
func (s *PaymentService) ReconcilePendingPaymentRefunds(ctx context.Context, limit int) (attempted, resolved int, err error) {
	if limit <= 0 {
		limit = 25
	}
	staleBefore := time.Now().Add(-refundSubmittingRetryAfter)
	pending, err := s.entClient.PaymentRefund.Query().Where(
		paymentRefundReconciliationPredicate(staleBefore),
	).Order(paymentrefund.ByCreatedAt(sql.OrderAsc())).Limit(limit).All(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("list pending payment refunds: %w", err)
	}
	return reconcilePaymentRefundBatch(ctx, pending, func(ctx context.Context, refund *dbent.PaymentRefund) (*PaymentRefundResponse, error) {
		switch refund.Status {
		case RefundStatusReserved:
			resumed, resumeErr := s.executePaymentRefund(ctx, refund.ID)
			if resumeErr != nil {
				return nil, resumeErr
			}
			return paymentRefundResponse(resumed), nil
		case RefundStatusSubmitting:
			retried, retryErr := s.retryStaleSubmittingPaymentRefund(ctx, refund.ID, staleBefore)
			if retryErr != nil {
				return nil, retryErr
			}
			return paymentRefundResponse(retried), nil
		default:
			return s.queryPaymentRefundByID(ctx, refund.ID)
		}
	})
}

func reconcilePaymentRefundBatch(
	ctx context.Context,
	refunds []*dbent.PaymentRefund,
	reconcile func(context.Context, *dbent.PaymentRefund) (*PaymentRefundResponse, error),
) (attempted, resolved int, err error) {
	errs := make([]error, 0)
	for _, refund := range refunds {
		if ctxErr := ctx.Err(); ctxErr != nil {
			errs = append(errs, ctxErr)
			break
		}
		if refund == nil {
			continue
		}
		attempted++
		result, queryErr := reconcile(ctx, refund)
		if queryErr != nil {
			errs = append(errs, fmt.Errorf("reconcile refund %s: %w", refund.ID, queryErr))
			continue
		}
		if result != nil && (result.Status == RefundStatusSucceeded || result.Status == RefundStatusFailed) {
			resolved++
		}
	}
	return attempted, resolved, errors.Join(errs...)
}

type affiliateRefundWalletHold struct {
	ID, UserID string
	Amount     float64
}

type affiliateReversalSettlementAllocation struct {
	Frozen, Available, Wallet float64
}

func affiliateReversalSettlement(amount, held, available, frozen float64) (affiliateReversalSettlementAllocation, error) {
	requested := decimal.NewFromFloat(amount).Round(refundPointsScale)
	heldAmount := decimal.NewFromFloat(held).Round(refundPointsScale)
	if requested.IsNegative() || heldAmount.IsNegative() || heldAmount.GreaterThan(requested) {
		return affiliateReversalSettlementAllocation{}, fmt.Errorf("affiliate refund wallet hold amount is invalid")
	}
	poolTarget := requested.Sub(heldAmount)
	frozenDeduction := decimal.Min(poolTarget, decimal.NewFromFloat(math.Max(frozen, 0)).Round(refundPointsScale))
	afterFrozen := poolTarget.Sub(frozenDeduction)
	availableDeduction := decimal.Min(afterFrozen, decimal.NewFromFloat(math.Max(available, 0)).Round(refundPointsScale))
	return affiliateReversalSettlementAllocation{
		Frozen:    decimalFloat(frozenDeduction, refundPointsScale),
		Available: decimalFloat(availableDeduction, refundPointsScale),
		Wallet:    decimalFloat(afterFrozen.Sub(availableDeduction), refundPointsScale),
	}, nil
}

func loadAffiliateRefundWalletHold(ctx context.Context, client *dbent.Client, refundID string) (*affiliateRefundWalletHold, error) {
	rows, err := client.QueryContext(ctx, `
		SELECT id, user_id, amount
		FROM wallet_holds
		WHERE purpose = $1 AND reference_id = $2
		FOR UPDATE
	`, affiliateRefundHoldPurpose, refundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	var hold affiliateRefundWalletHold
	if err := rows.Scan(&hold.ID, &hold.UserID, &hold.Amount); err != nil {
		return nil, err
	}
	return &hold, rows.Err()
}

func releaseAffiliateReversal(ctx context.Context, client *dbent.Client, orderID string, amount float64) error {
	remaining := decimal.NewFromFloat(amount).Round(refundPointsScale)
	rows, err := client.QueryContext(ctx, `
		SELECT id, reserved_reversal_amount
		FROM user_affiliate_ledger
		WHERE source_order_id = $1 AND action = 'accrue' AND reserved_reversal_amount > 0
		ORDER BY created_at, id
		FOR UPDATE
	`, orderID)
	if err != nil {
		return err
	}
	type allocation struct {
		id       string
		reserved float64
	}
	items := make([]allocation, 0)
	for rows.Next() {
		var item allocation
		if err := rows.Scan(&item.id, &item.reserved); err != nil {
			_ = rows.Close()
			return err
		}
		items = append(items, item)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, item := range items {
		if !remaining.IsPositive() {
			break
		}
		part := decimal.Min(remaining, decimal.NewFromFloat(item.reserved).Round(refundPointsScale))
		if _, err := client.ExecContext(ctx, `
			UPDATE user_affiliate_ledger
			SET reserved_reversal_amount = reserved_reversal_amount - $1, updated_at = NOW()
			WHERE id = $2
		`, decimalFloat(part, refundPointsScale), item.id); err != nil {
			return err
		}
		remaining = remaining.Sub(part)
	}
	if remaining.IsPositive() {
		return fmt.Errorf("affiliate reversal reservation is incomplete")
	}
	return nil
}

func (s *PaymentService) captureAffiliateReversal(ctx context.Context, client *dbent.Client, refund *dbent.PaymentRefund, order *dbent.PaymentOrder) (string, error) {
	amount := decimal.NewFromFloat(refund.AffiliateRebatePoints).Round(refundPointsScale)
	if !amount.IsPositive() {
		return "", nil
	}
	rows, err := client.QueryContext(ctx, `
		SELECT id, user_id, source_user_id, reserved_reversal_amount
		FROM user_affiliate_ledger
		WHERE source_order_id = $1 AND action = 'accrue' AND reserved_reversal_amount > 0
		ORDER BY created_at, id
		FOR UPDATE
	`, order.ID)
	if err != nil {
		return "", err
	}
	type allocation struct {
		id, inviterID, inviteeID string
		reserved                 float64
	}
	items := make([]allocation, 0)
	for rows.Next() {
		var item allocation
		if err := rows.Scan(&item.id, &item.inviterID, &item.inviteeID, &item.reserved); err != nil {
			_ = rows.Close()
			return "", err
		}
		items = append(items, item)
	}
	if err := rows.Close(); err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", fmt.Errorf("affiliate reversal reservation not found")
	}
	inviterID := items[0].inviterID
	profileRows, err := client.QueryContext(ctx, `
		SELECT aff_quota, aff_frozen_quota
		FROM user_affiliates WHERE user_id = $1 FOR UPDATE
	`, inviterID)
	if err != nil {
		return "", err
	}
	var available, frozen float64
	if !profileRows.Next() {
		_ = profileRows.Close()
		return "", fmt.Errorf("affiliate profile not found for reversal")
	}
	if err := profileRows.Scan(&available, &frozen); err != nil {
		_ = profileRows.Close()
		return "", err
	}
	if err := profileRows.Close(); err != nil {
		return "", err
	}
	affiliateHold, err := loadAffiliateRefundWalletHold(ctx, client, refund.ID)
	if err != nil {
		return "", err
	}
	heldAmount := 0.0
	if affiliateHold != nil {
		if affiliateHold.UserID != inviterID {
			return "", fmt.Errorf("affiliate refund wallet hold belongs to another user")
		}
		heldAmount = affiliateHold.Amount
	}
	settlement, err := affiliateReversalSettlement(refund.AffiliateRebatePoints, heldAmount, available, frozen)
	if err != nil {
		return "", err
	}
	if _, err := client.ExecContext(ctx, `
		UPDATE user_affiliates SET
			aff_frozen_quota = GREATEST(aff_frozen_quota - $1, 0),
			aff_quota = GREATEST(aff_quota - $2, 0),
			updated_at = NOW()
		WHERE user_id = $3
	`, settlement.Frozen, settlement.Available, inviterID); err != nil {
		return "", err
	}
	if affiliateHold != nil {
		if _, err := s.userRepo.CaptureWalletHold(ctx, affiliateHold.ID, "wallet-affiliate-refund-capture:"+refund.ID); err != nil {
			return "", err
		}
	}
	if settlement.Wallet > 0 {
		if _, err := s.userRepo.DebitWallet(ctx, WalletDebitInput{
			UserID: inviterID, Amount: settlement.Wallet, AllowOverdraft: true,
			SourceType: "affiliate_refund", SourceID: refund.ID,
			IdempotencyKey: "wallet-affiliate-refund:" + refund.ID,
			Notes:          "reverse transferred affiliate rebate after recharge refund",
		}); err != nil {
			return "", err
		}
	}
	remaining := amount
	for _, item := range items {
		if !remaining.IsPositive() {
			break
		}
		part := decimal.Min(remaining, decimal.NewFromFloat(item.reserved).Round(refundPointsScale))
		if _, err := client.ExecContext(ctx, `
			UPDATE user_affiliate_ledger SET
				reserved_reversal_amount = reserved_reversal_amount - $1,
				reversed_amount = reversed_amount + $1,
				updated_at = NOW()
			WHERE id = $2
		`, decimalFloat(part, refundPointsScale), item.id); err != nil {
			return "", err
		}
		remaining = remaining.Sub(part)
	}
	if remaining.IsPositive() {
		return "", fmt.Errorf("affiliate reversal reservation is incomplete")
	}
	_, err = client.ExecContext(ctx, `
		INSERT INTO user_affiliate_ledger
			(user_id, action, amount, source_user_id, source_order_id, created_at, updated_at)
		VALUES ($1, 'reverse', $2, $3, $4, NOW(), NOW())
	`, inviterID, refund.AffiliateRebatePoints, order.UserID, order.ID)
	return inviterID, err
}

func (s *PaymentService) invalidateRefundWalletCaches(ctx context.Context, userID string) {
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}
	if s.billingCacheService != nil {
		_ = s.billingCacheService.InvalidateUserBalance(ctx, userID)
	}
}
