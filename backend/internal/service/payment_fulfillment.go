package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/ent/paymentauditlog"
	"github.com/AsukaCC/EasySub2api/ent/paymentorder"
	"github.com/AsukaCC/EasySub2api/internal/payment"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

// ErrOrderNotFound is returned by HandlePaymentNotification when the webhook
// references an out_trade_no that does not exist in our DB. Callers (webhook
// handlers) should treat this as a terminal, non-retryable condition and still
// respond with a 2xx success to the provider — otherwise the provider will keep
// retrying forever (e.g. when a foreign environment's webhook endpoint is
// misconfigured to point at us, or when our orders table has been wiped).
var ErrOrderNotFound = errors.New("payment order not found")

const paymentFulfillmentLeaseDuration = 5 * time.Minute

type paymentFulfillmentLease struct {
	version time.Time
}

func (l *paymentFulfillmentLease) nextVersion() time.Time {
	next := time.Now().UTC().Truncate(time.Microsecond)
	if !next.After(l.version) {
		next = l.version.UTC().Add(time.Microsecond)
	}
	return next
}

// --- Payment Notification & Fulfillment ---

func (s *PaymentService) HandlePaymentNotification(ctx context.Context, n *payment.PaymentNotification, pk string) error {
	if n.Status != payment.NotificationStatusSuccess {
		return nil
	}
	// Look up order by out_trade_no (the external order ID we sent to the provider)
	order, err := s.entClient.PaymentOrder.Query().Where(paymentorder.OutTradeNo(n.OrderID)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return fmt.Errorf("%w: out_trade_no=%s", ErrOrderNotFound, n.OrderID)
		}
		return fmt.Errorf("lookup order failed for out_trade_no %s: %w", n.OrderID, err)
	}
	return s.confirmPayment(ctx, order.ID, n.TradeNo, n.Amount, pk, n.Metadata)
}

func (s *PaymentService) confirmPayment(ctx context.Context, oid string, tradeNo string, paid float64, pk string, metadata map[string]string) error {
	o, err := s.entClient.PaymentOrder.Get(ctx, oid)
	if err != nil {
		slog.Error("order not found", "orderID", oid)
		return nil
	}
	instanceProviderKey := ""
	if inst, instErr := s.getOrderProviderInstance(ctx, o); instErr == nil && inst != nil {
		instanceProviderKey = inst.ProviderKey
	}
	expectedProviderKey := expectedNotificationProviderKeyForOrder(s.registry, o, instanceProviderKey)
	if expectedProviderKey != "" && strings.TrimSpace(pk) != "" && !strings.EqualFold(expectedProviderKey, strings.TrimSpace(pk)) {
		s.writeAuditLog(ctx, o.ID, "PAYMENT_PROVIDER_MISMATCH", pk, map[string]any{
			"expectedProvider": expectedProviderKey,
			"actualProvider":   pk,
			"tradeNo":          tradeNo,
		})
		return fmt.Errorf("provider mismatch: expected %s, got %s", expectedProviderKey, pk)
	}
	if err := validateProviderNotificationMetadata(o, pk, metadata); err != nil {
		s.writeAuditLog(ctx, o.ID, "PAYMENT_PROVIDER_METADATA_MISMATCH", pk, map[string]any{
			"detail":  err.Error(),
			"tradeNo": tradeNo,
		})
		return err
	}
	if !isValidProviderAmount(paid) {
		s.writeAuditLog(ctx, o.ID, "PAYMENT_INVALID_AMOUNT", pk, map[string]any{
			"expected": o.PayAmount,
			"paid":     paid,
			"tradeNo":  tradeNo,
		})
		return fmt.Errorf("invalid paid amount from provider: %v", paid)
	}
	if math.Abs(paid-o.PayAmount) > paymentAmountToleranceForCurrency(PaymentOrderCurrency(o)) {
		s.writeAuditLog(ctx, o.ID, "PAYMENT_AMOUNT_MISMATCH", pk, map[string]any{"expected": o.PayAmount, "paid": paid, "tradeNo": tradeNo})
		return fmt.Errorf("amount mismatch: expected %s, got %s", strconv.FormatFloat(o.PayAmount, 'f', -1, 64), strconv.FormatFloat(paid, 'f', -1, 64))
	}
	return s.toPaid(ctx, o, tradeNo, paid, pk)
}

func paymentAmountToleranceForCurrency(currency string) float64 {
	minorUnit := payment.CurrencyMinorUnit(currency)
	if minorUnit <= 2 {
		return amountToleranceCNY
	}
	return math.Pow10(-minorUnit) / 2
}

func isValidProviderAmount(amount float64) bool {
	return amount > 0 && !math.IsNaN(amount) && !math.IsInf(amount, 0)
}

func validateProviderNotificationMetadata(order *dbent.PaymentOrder, providerKey string, metadata map[string]string) error {
	return validateProviderSnapshotMetadata(order, providerKey, metadata)
}

func expectedNotificationProviderKey(registry *payment.Registry, orderPaymentType string, orderProviderKey string, instanceProviderKey string) string {
	if key := strings.TrimSpace(instanceProviderKey); key != "" {
		return key
	}
	if key := strings.TrimSpace(orderProviderKey); key != "" {
		return key
	}
	if registry != nil {
		if key := strings.TrimSpace(registry.GetProviderKey(payment.PaymentType(orderPaymentType))); key != "" {
			return key
		}
	}
	return strings.TrimSpace(orderPaymentType)
}

func (s *PaymentService) toPaid(ctx context.Context, o *dbent.PaymentOrder, tradeNo string, paid float64, pk string) error {
	previousStatus := o.Status
	now := time.Now()
	grace := now.Add(-paymentGraceMinutes * time.Minute)
	c, err := s.entClient.PaymentOrder.Update().Where(
		paymentorder.IDEQ(o.ID),
		paymentorder.Or(
			paymentorder.StatusEQ(OrderStatusPending),
			paymentorder.StatusEQ(OrderStatusCancelled),
			paymentorder.And(
				paymentorder.StatusEQ(OrderStatusExpired),
				paymentorder.UpdatedAtGTE(grace),
			),
		),
	).SetStatus(OrderStatusPaid).SetPaymentTradeNo(tradeNo).SetPaidAt(now).ClearFailedAt().ClearFailedReason().Save(ctx)
	if err != nil {
		return fmt.Errorf("update to PAID: %w", err)
	}
	if c == 0 {
		return s.alreadyProcessed(ctx, o)
	}
	if previousStatus == OrderStatusCancelled || previousStatus == OrderStatusExpired {
		slog.Info("order recovered from webhook payment success",
			"orderID", o.ID,
			"previousStatus", previousStatus,
			"tradeNo", tradeNo,
			"provider", pk,
		)
		s.writeAuditLog(ctx, o.ID, "ORDER_RECOVERED", pk, map[string]any{
			"previous_status": previousStatus,
			"tradeNo":         tradeNo,
			"paidAmount":      paid,
			"reason":          "webhook payment success received after order " + previousStatus,
		})
	}
	s.writeAuditLog(ctx, o.ID, "ORDER_PAID", pk, map[string]any{
		"tradeNo":                tradeNo,
		"expectedPayAmount":      o.PayAmount,
		"providerReportedAmount": paid,
	})
	return s.executeFulfillment(ctx, o.ID)
}

func (s *PaymentService) alreadyProcessed(ctx context.Context, o *dbent.PaymentOrder) error {
	cur, err := s.entClient.PaymentOrder.Get(ctx, o.ID)
	if err != nil {
		return nil
	}
	switch cur.Status {
	case OrderStatusCompleted, OrderStatusRefunded:
		return nil
	case OrderStatusFailed, OrderStatusPaid, OrderStatusRecharging:
		return s.executeFulfillment(ctx, o.ID)
	case OrderStatusExpired:
		slog.Warn("webhook payment success for expired order beyond grace period",
			"orderID", o.ID,
			"status", cur.Status,
			"updatedAt", cur.UpdatedAt,
		)
		s.writeAuditLog(ctx, o.ID, "PAYMENT_AFTER_EXPIRY", "system", map[string]any{
			"status":    cur.Status,
			"updatedAt": cur.UpdatedAt,
			"reason":    "payment arrived after expiry grace period",
		})
		return nil
	default:
		return nil
	}
}

func (s *PaymentService) executeFulfillment(ctx context.Context, oid string) error {
	o, err := s.entClient.PaymentOrder.Get(ctx, oid)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}
	if o.OrderType == payment.OrderTypeSubscription {
		return s.ExecuteSubscriptionFulfillment(ctx, oid)
	}
	return s.ExecuteBalanceFulfillment(ctx, oid)
}

func (s *PaymentService) ExecuteBalanceFulfillment(ctx context.Context, oid string) error {
	o, err := s.entClient.PaymentOrder.Get(ctx, oid)
	if err != nil {
		return infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if o.Status == OrderStatusCompleted {
		return nil
	}
	if psIsRefundStatus(o.Status) {
		return infraerrors.BadRequest("INVALID_STATUS", "refund-related order cannot fulfill")
	}
	if o.Status != OrderStatusPaid && o.Status != OrderStatusFailed && o.Status != OrderStatusRecharging {
		return infraerrors.BadRequest("INVALID_STATUS", "order cannot fulfill in status "+o.Status)
	}
	lease, err := s.acquirePaymentFulfillmentLease(ctx, o)
	if err != nil {
		return err
	}
	if lease == nil {
		return nil
	}
	if err := s.doBalance(ctx, o, lease); err != nil {
		s.markFailed(ctx, oid, lease, err)
		return err
	}
	return nil
}

func (s *PaymentService) acquirePaymentFulfillmentLease(ctx context.Context, o *dbent.PaymentOrder) (*paymentFulfillmentLease, error) {
	if o == nil {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "nil payment order")
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	staleBefore := now.Add(-paymentFulfillmentLeaseDuration)
	updated, err := s.entClient.PaymentOrder.Update().
		Where(
			paymentorder.IDEQ(o.ID),
			paymentorder.Or(
				paymentorder.StatusIn(OrderStatusPaid, OrderStatusFailed),
				paymentorder.And(
					paymentorder.StatusEQ(OrderStatusRecharging),
					paymentorder.UpdatedAtLTE(staleBefore),
				),
			),
		).
		SetStatus(OrderStatusRecharging).
		SetUpdatedAt(now).
		ClearFailedAt().
		ClearFailedReason().
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire fulfillment lease: %w", err)
	}
	if updated == 0 {
		current, getErr := s.entClient.PaymentOrder.Get(ctx, o.ID)
		if getErr != nil {
			return nil, fmt.Errorf("reload fulfillment lease: %w", getErr)
		}
		if current.Status == OrderStatusCompleted {
			return nil, nil
		}
		if current.Status == OrderStatusRecharging {
			return nil, infraerrors.Conflict("CONFLICT", "order is being processed")
		}
		return nil, infraerrors.Conflict("CONFLICT", "order status changed while acquiring fulfillment lease")
	}

	// Reload the persisted timestamp instead of trusting application clock precision.
	claimed, err := s.entClient.PaymentOrder.Get(ctx, o.ID)
	if err != nil {
		return nil, fmt.Errorf("reload acquired fulfillment lease: %w", err)
	}
	if claimed.Status != OrderStatusRecharging {
		return nil, infraerrors.Conflict("CONFLICT", "fulfillment lease was lost")
	}
	return &paymentFulfillmentLease{version: claimed.UpdatedAt}, nil
}

// redeemAction represents the idempotency decision for balance fulfillment.
type redeemAction int

const (
	// redeemActionCreate: code does not exist — create it, then redeem.
	redeemActionCreate redeemAction = iota
	// redeemActionRedeem: code exists but is unused — skip creation, redeem only.
	redeemActionRedeem
	// redeemActionSkipCompleted: code exists and is already used — skip to mark completed.
	redeemActionSkipCompleted
)

// resolveRedeemAction decides the idempotency action based on an existing redeem code lookup.
// existing is the result of GetByCode; lookupErr is the error from that call.
func resolveRedeemAction(existing *RedeemCode, lookupErr error) redeemAction {
	if existing == nil || lookupErr != nil {
		return redeemActionCreate
	}
	if existing.IsUsed() {
		return redeemActionSkipCompleted
	}
	return redeemActionRedeem
}

func (s *PaymentService) doBalance(ctx context.Context, o *dbent.PaymentOrder, lease *paymentFulfillmentLease) error {
	// Idempotency: check if redeem code already exists (from a previous partial run)
	existing, lookupErr := s.redeemService.GetByCode(ctx, o.RechargeCode)
	action := resolveRedeemAction(existing, lookupErr)

	if action == redeemActionSkipCompleted {
		if err := s.applyAffiliateRebateForOrder(ctx, o, lease); err != nil {
			return err
		}
		return s.markCompleted(ctx, o, lease, "RECHARGE_SUCCESS")
	}
	if err := s.creditBalanceRechargePoints(ctx, o, lease); err != nil {
		return err
	}
	if err := s.applyAffiliateRebateForOrder(ctx, o, lease); err != nil {
		return err
	}
	return s.markCompleted(ctx, o, lease, "RECHARGE_SUCCESS")
}

func (s *PaymentService) creditBalanceRechargePoints(ctx context.Context, o *dbent.PaymentOrder, leases ...*paymentFulfillmentLease) error {
	var lease *paymentFulfillmentLease
	if len(leases) > 0 {
		lease = leases[0]
	}
	basePoints := o.BasePoints
	bonusPoints := o.BonusPoints
	if basePoints <= 0 && bonusPoints <= 0 {
		// Compatibility for legacy recharge orders created before the explicit
		// points split was introduced.
		basePoints = o.Amount
	}
	if _, err := s.userRepo.CreditWallet(ctx, WalletCreditInput{
		UserID: o.UserID, Amount: basePoints, Kind: WalletKindRecharge,
		SourceType: "payment_order", SourceID: o.ID,
		IdempotencyKey: "wallet-payment-base:" + o.ID,
		Notes:          "balance recharge base points", CountAsRecharged: true,
	}); err != nil {
		return fmt.Errorf("credit recharge base points: %w", err)
	}
	if bonusPoints > 0 {
		bonusResult, err := s.userRepo.CreditWallet(ctx, WalletCreditInput{
			UserID: o.UserID, Amount: bonusPoints, Kind: WalletKindBonus,
			ExpiresAt: o.BonusExpiresAt, SourceType: "payment_order", SourceID: o.ID,
			IdempotencyKey: "wallet-payment-bonus:" + o.ID,
			Notes:          "balance recharge bonus points", CountAsRecharged: true,
		})
		if err != nil {
			return fmt.Errorf("credit recharge bonus points: %w", err)
		}
		if bonusResult.BonusGrantID != nil && strings.TrimSpace(*bonusResult.BonusGrantID) != "" {
			if lease == nil {
				return errors.New("save recharge bonus grant id: missing payment fulfillment lease")
			}
			nextLeaseVersion := lease.nextVersion()
			updatedOrder, err := s.entClient.PaymentOrder.UpdateOneID(o.ID).
				Where(
					paymentorder.StatusEQ(OrderStatusRecharging),
					paymentorder.UpdatedAtEQ(lease.version),
				).
				SetBonusGrantID(*bonusResult.BonusGrantID).
				SetUpdatedAt(nextLeaseVersion).
				Save(ctx)
			if err != nil {
				if dbent.IsNotFound(err) {
					return infraerrors.Conflict("CONFLICT", "fulfillment lease was lost while saving recharge bonus grant")
				}
				return fmt.Errorf("save recharge bonus grant id: %w", err)
			}
			lease.version = updatedOrder.UpdatedAt
			o.BonusGrantID = bonusResult.BonusGrantID
		}
	}
	return nil
}

func (s *PaymentService) markCompleted(ctx context.Context, o *dbent.PaymentOrder, lease *paymentFulfillmentLease, auditAction string) error {
	if lease == nil {
		return errors.New("missing payment fulfillment lease")
	}
	now := time.Now().UTC()
	refundDeadline := completedRechargeRefundDeadline(o, now)
	update := s.entClient.PaymentOrder.Update().Where(
		paymentorder.IDEQ(o.ID),
		paymentorder.StatusEQ(OrderStatusRecharging),
		paymentorder.UpdatedAtEQ(lease.version),
	).SetStatus(OrderStatusCompleted).SetCompletedAt(now)
	if refundDeadline != nil {
		update.SetRefundDeadline(*refundDeadline)
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return fmt.Errorf("mark completed: %w", err)
	}
	if updated == 0 {
		current, getErr := s.entClient.PaymentOrder.Get(ctx, o.ID)
		if getErr == nil && current.Status == OrderStatusCompleted {
			return nil
		}
		return infraerrors.Conflict("CONFLICT", "fulfillment lease was lost before completion")
	}
	o.CompletedAt = &now
	o.RefundDeadline = refundDeadline
	if !s.hasAuditLog(ctx, o.ID, auditAction) {
		s.writeAuditLog(ctx, o.ID, auditAction, "system", map[string]any{
			"rechargeCode":          o.RechargeCode,
			"creditedAmount":        o.Amount,
			"principalAmount":       o.PrincipalAmount,
			"basePoints":            o.BasePoints,
			"bonusPoints":           o.BonusPoints,
			"creditedPoints":        o.CreditedPoints,
			"affiliateRebatePoints": o.AffiliateRebatePoints,
			"payAmount":             o.PayAmount,
		})
		s.dispatchPaymentFulfillmentNotification(o, auditAction)
	}
	return nil
}

func completedRechargeRefundDeadline(order *dbent.PaymentOrder, completedAt time.Time) *time.Time {
	if order == nil || order.OrderType != payment.OrderTypeBalance {
		return nil
	}
	deadline := completedAt.UTC().Add(rechargeRefundWindow)
	return &deadline
}

func (s *PaymentService) dispatchPaymentFulfillmentNotification(o *dbent.PaymentOrder, auditAction string) {
	if s == nil || s.notificationEmailService == nil || o == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), emailSendTimeout)
		defer cancel()
		var err error
		switch auditAction {
		case "RECHARGE_SUCCESS":
			err = s.sendBalanceRechargeSuccessNotification(ctx, o)
		case "SUBSCRIPTION_SUCCESS":
			err = s.sendSubscriptionPurchaseSuccessNotification(ctx, o)
		default:
			return
		}
		if err != nil {
			slog.Warn("payment fulfillment notification email failed", "order_id", o.ID, "action", auditAction, "err", err.Error())
		}
	}()
}

func (s *PaymentService) sendBalanceRechargeSuccessNotification(ctx context.Context, o *dbent.PaymentOrder) error {
	currentBalance := ""
	if s.userRepo != nil {
		if user, err := s.userRepo.GetByID(ctx, o.UserID); err == nil && user != nil {
			currentBalance = formatPlatformPoints(user.Balance)
		}
	}
	return s.notificationEmailService.Send(ctx, NotificationEmailSendInput{
		Event:          NotificationEmailEventBalanceRechargeSuccess,
		RecipientEmail: o.UserEmail,
		RecipientName:  firstNonEmpty(o.UserName, o.UserEmail),
		UserID:         o.UserID,
		SourceType:     "payment_order",
		SourceID:       o.ID,
		Variables: map[string]string{
			"recharge_amount":      formatPlatformPoints(o.Amount),
			"current_balance":      currentBalance,
			"principal_amount_cny": fmt.Sprintf("%.2f", o.PrincipalAmount),
			"base_points":          formatPlatformPoints(o.BasePoints),
			"bonus_points":         formatPlatformPoints(o.BonusPoints),
			"credited_points":      formatPlatformPoints(o.CreditedPoints),
			"current_points":       currentBalance,
			"order_id":             o.ID,
		},
	})
}

func (s *PaymentService) sendSubscriptionPurchaseSuccessNotification(ctx context.Context, o *dbent.PaymentOrder) error {
	variables := map[string]string{
		"subscription_group": "Subscription",
		"subscription_days":  "",
		"expiry_time":        "",
		"order_id":           o.ID,
	}
	if o.SubscriptionDays != nil {
		variables["subscription_days"] = strconv.Itoa(*o.SubscriptionDays)
	}
	if o.SubscriptionGroupID != nil {
		if s.groupRepo != nil {
			if group, err := s.groupRepo.GetByID(ctx, *o.SubscriptionGroupID); err == nil && group != nil && strings.TrimSpace(group.Name) != "" {
				variables["subscription_group"] = group.Name
			}
		}
		if s.subscriptionSvc != nil {
			if sub, err := s.subscriptionSvc.GetActiveSubscription(ctx, o.UserID, *o.SubscriptionGroupID); err == nil && sub != nil {
				variables["expiry_time"] = sub.ExpiresAt.Format("2006-01-02 15:04")
			}
		}
	}
	return s.notificationEmailService.Send(ctx, NotificationEmailSendInput{
		Event:          NotificationEmailEventSubscriptionPurchaseSuccess,
		RecipientEmail: o.UserEmail,
		RecipientName:  firstNonEmpty(o.UserName, o.UserEmail),
		UserID:         o.UserID,
		SourceType:     "payment_order",
		SourceID:       o.ID,
		Variables:      variables,
	})
}

func (s *PaymentService) ExecuteSubscriptionFulfillment(ctx context.Context, oid string) error {
	o, err := s.entClient.PaymentOrder.Get(ctx, oid)
	if err != nil {
		return infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if o.Status == OrderStatusCompleted {
		return nil
	}
	if psIsRefundStatus(o.Status) {
		return infraerrors.BadRequest("INVALID_STATUS", "refund-related order cannot fulfill")
	}
	if o.Status != OrderStatusPaid && o.Status != OrderStatusFailed && o.Status != OrderStatusRecharging {
		return infraerrors.BadRequest("INVALID_STATUS", "order cannot fulfill in status "+o.Status)
	}
	lease, err := s.acquirePaymentFulfillmentLease(ctx, o)
	if err != nil {
		return err
	}
	if lease == nil {
		return nil
	}
	if err := s.doSub(ctx, o, lease); err != nil {
		s.releaseFailedSubscriptionWalletHold(ctx, o)
		s.markFailed(ctx, oid, lease, err)
		return err
	}
	return nil
}

func (s *PaymentService) releaseFailedSubscriptionWalletHold(ctx context.Context, o *dbent.PaymentOrder) {
	if o == nil || !o.WalletOnly || o.WalletHoldID == nil || strings.TrimSpace(*o.WalletHoldID) == "" {
		return
	}
	result, err := s.userRepo.ReleaseWalletHold(ctx, *o.WalletHoldID, "wallet-subscription-failed-release:"+o.ID)
	if err != nil {
		// A captured hold is deliberately left untouched. Capture and subscription
		// assignment share one transaction, so a captured result can be recovered
		// idempotently without refunding an already-provisioned subscription.
		if !errors.Is(err, ErrWalletHoldState) {
			slog.Error("release failed subscription wallet hold", "orderID", o.ID, "holdID", *o.WalletHoldID, "error", err)
		}
		return
	}
	if result.Status == "released" {
		s.writeAuditLog(ctx, o.ID, "SUBSCRIPTION_WALLET_HOLD_RELEASED", "system", map[string]any{
			"holdID": *o.WalletHoldID,
			"amount": result.Amount,
		})
	}
}

func (s *PaymentService) doSub(ctx context.Context, o *dbent.PaymentOrder, lease *paymentFulfillmentLease) error {
	if o.SubscriptionGroupID == nil || o.SubscriptionDays == nil {
		return infraerrors.BadRequest("INVALID_STATUS", "missing subscription info")
	}
	gid := *o.SubscriptionGroupID
	days := *o.SubscriptionDays
	g, err := s.groupRepo.GetByID(ctx, gid)
	if err != nil || g.Status != payment.EntityStatusActive {
		return fmt.Errorf("group %v no longer exists or inactive", gid)
	}
	if err := s.ensurePaymentSubscriptionAssigned(ctx, o, gid, days); err != nil {
		return err
	}
	return s.markCompleted(ctx, o, lease, "SUBSCRIPTION_SUCCESS")
}

func (s *PaymentService) ensurePaymentSubscriptionAssigned(ctx context.Context, o *dbent.PaymentOrder, groupID string, days int) error {
	if s.subscriptionSvc == nil {
		return errors.New("subscription service is unavailable")
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin subscription fulfillment tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	txClient := tx.Client()
	if o.WalletHoldID != nil && strings.TrimSpace(*o.WalletHoldID) != "" {
		if _, err := s.userRepo.CaptureWalletHold(txCtx, *o.WalletHoldID, "wallet-subscription-capture:"+o.ID); err != nil {
			if !errors.Is(err, ErrWalletHoldState) {
				return fmt.Errorf("capture subscription wallet hold: %w", err)
			}
			if o.WalletOnly {
				return fmt.Errorf("wallet-only subscription hold has already been released: %w", err)
			}
			debit, debitErr := s.userRepo.DebitWallet(txCtx, WalletDebitInput{
				UserID: o.UserID, Amount: o.WalletAmount, AllowOverdraft: true,
				SourceType: "subscription_late_payment", SourceID: o.ID,
				IdempotencyKey: "wallet-subscription-late-debit:" + o.ID,
				Notes:          "payment arrived after wallet hold was released",
			})
			if debitErr != nil {
				return fmt.Errorf("recover released subscription wallet hold: %w", debitErr)
			}
			detail, _ := json.Marshal(map[string]any{"walletAmount": o.WalletAmount, "overdraftAmount": debit.Summary.OverdraftAmount})
			if _, auditErr := txClient.PaymentAuditLog.Create().SetOrderID(o.ID).SetAction("LATE_PAYMENT_WALLET_REDEBIT").SetDetail(string(detail)).SetOperator("system").Save(txCtx); auditErr != nil {
				return fmt.Errorf("record late payment wallet recovery: %w", auditErr)
			}
		}
	}
	alreadyAssigned, err := hasPaymentSubscriptionAssignmentAudit(txCtx, txClient, o.ID)
	if err != nil {
		return fmt.Errorf("check subscription assignment audit: %w", err)
	}

	recoveredFromNote := false
	if !alreadyAssigned {
		orderNote := paymentSubscriptionOrderNote(o.ID)
		existing, lookupErr := s.subscriptionSvc.userSubRepo.GetByUserIDAndGroupID(txCtx, o.UserID, groupID)
		switch {
		case lookupErr == nil && existing != nil && hasPaymentSubscriptionOrderNote(existing.Notes, orderNote):
			recoveredFromNote = true
		case lookupErr != nil && !errors.Is(lookupErr, ErrSubscriptionNotFound):
			return fmt.Errorf("check existing subscription assignment: %w", lookupErr)
		default:
			if _, _, err := s.subscriptionSvc.assignOrExtendSubscription(txCtx, &AssignSubscriptionInput{
				UserID:       o.UserID,
				GroupID:      groupID,
				ValidityDays: days,
				AssignedBy:   "",
				Notes:        orderNote,
			}, true); err != nil {
				return fmt.Errorf("assign subscription: %w", err)
			}
		}

		detail, _ := json.Marshal(map[string]any{
			"groupID":           groupID,
			"validityDays":      days,
			"recoveredFromNote": recoveredFromNote,
		})
		if _, err := txClient.PaymentAuditLog.Create().
			SetOrderID(o.ID).
			SetAction("SUBSCRIPTION_ASSIGNED").
			SetDetail(string(detail)).
			SetOperator("system").
			Save(txCtx); err != nil {
			if dbent.IsConstraintError(err) {
				_ = tx.Rollback()
				claimed, checkErr := hasPaymentSubscriptionAssignmentAudit(ctx, s.entClient, o.ID)
				if checkErr == nil && claimed {
					s.invalidateSubscriptionCachesAfterFulfillment(ctx, o, groupID)
					return nil
				}
			}
			return fmt.Errorf("record subscription assignment audit: %w", err)
		}
	} else {
		slog.Info("subscription already assigned for order, skipping", "orderID", o.ID, "groupID", groupID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subscription fulfillment tx: %w", err)
	}
	// Assignment cache invalidation is deferred while this transaction is open,
	// then performed synchronously against the committed subscription.
	s.invalidateSubscriptionCachesAfterFulfillment(ctx, o, groupID)
	return nil
}

func (s *PaymentService) invalidateSubscriptionCachesAfterFulfillment(ctx context.Context, o *dbent.PaymentOrder, groupID string) {
	if err := s.subscriptionSvc.invalidateSubscriptionCaches(o.UserID, groupID); err != nil {
		// The wallet capture and subscription assignment are already committed.
		// Cache invalidation is advisory here and must not turn a completed purchase
		// into a FAILED order whose captured points cannot be released.
		slog.Error("invalidate subscription cache after fulfillment", "orderID", o.ID, "userID", o.UserID, "groupID", groupID, "error", err)
		s.writeAuditLog(ctx, o.ID, "SUBSCRIPTION_CACHE_INVALIDATION_FAILED", "system", map[string]any{"error": err.Error()})
	}
}

func hasPaymentSubscriptionAssignmentAudit(ctx context.Context, client *dbent.Client, orderID string) (bool, error) {
	count, err := client.PaymentAuditLog.Query().
		Where(
			paymentauditlog.OrderIDEQ(orderID),
			paymentauditlog.ActionIn("SUBSCRIPTION_ASSIGNED", "SUBSCRIPTION_SUCCESS"),
		).
		Limit(1).
		Count(ctx)
	return count > 0, err
}

func paymentSubscriptionOrderNote(orderID string) string {
	return fmt.Sprintf("payment order %v", orderID)
}

func hasPaymentSubscriptionOrderNote(notes string, orderNote string) bool {
	for _, line := range strings.Split(strings.ReplaceAll(notes, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == orderNote {
			return true
		}
	}
	return false
}

func (s *PaymentService) hasAuditLog(ctx context.Context, orderID string, action string) bool {
	oid := orderID
	c, _ := s.entClient.PaymentAuditLog.Query().
		Where(paymentauditlog.OrderIDEQ(oid), paymentauditlog.ActionEQ(action)).
		Limit(1).Count(ctx)
	return c > 0
}

func (s *PaymentService) applyAffiliateRebateForOrder(ctx context.Context, o *dbent.PaymentOrder, lease *paymentFulfillmentLease) error {
	baseAmount := affiliateRebateBaseAmount(o)
	if o == nil || baseAmount <= 0 {
		return nil
	}
	if s.affiliateService == nil {
		return nil
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		s.writeAuditLog(ctx, o.ID, "AFFILIATE_REBATE_FAILED", "system", map[string]any{
			"error": fmt.Sprintf("begin affiliate rebate tx: %v", err),
		})
		return fmt.Errorf("begin affiliate rebate tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	claimed, err := s.tryClaimAffiliateRebateAudit(txCtx, tx.Client(), o.ID, baseAmount)
	if err != nil {
		s.writeAuditLog(ctx, o.ID, "AFFILIATE_REBATE_FAILED", "system", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("claim affiliate rebate audit: %w", err)
	}
	if !claimed {
		return nil
	}

	sourceOrderID := o.ID
	rebateAmount, err := s.affiliateService.AccrueInviteRebateForOrder(txCtx, o.UserID, baseAmount, &sourceOrderID)
	if err != nil {
		s.writeAuditLog(ctx, o.ID, "AFFILIATE_REBATE_FAILED", "system", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("accrue affiliate rebate: %w", err)
	}

	if rebateAmount <= 0 {
		if err := s.updateClaimedAffiliateRebateAudit(txCtx, tx.Client(), o.ID, "AFFILIATE_REBATE_SKIPPED", map[string]any{
			"baseAmount": baseAmount,
			"reason":     "no inviter bound or rebate amount <= 0",
		}); err != nil {
			s.writeAuditLog(ctx, o.ID, "AFFILIATE_REBATE_FAILED", "system", map[string]any{
				"error": err.Error(),
			})
			return fmt.Errorf("update affiliate rebate skipped audit: %w", err)
		}
		if err := tx.Commit(); err != nil {
			s.writeAuditLog(ctx, o.ID, "AFFILIATE_REBATE_FAILED", "system", map[string]any{
				"error": fmt.Sprintf("commit affiliate rebate tx: %v", err),
			})
			return fmt.Errorf("commit affiliate rebate tx: %w", err)
		}
		return nil
	}

	if err := s.updateClaimedAffiliateRebateAudit(txCtx, tx.Client(), o.ID, "AFFILIATE_REBATE_APPLIED", map[string]any{
		"baseAmount":   baseAmount,
		"rebateAmount": rebateAmount,
	}); err != nil {
		s.writeAuditLog(ctx, o.ID, "AFFILIATE_REBATE_FAILED", "system", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("update affiliate rebate applied audit: %w", err)
	}
	if lease == nil {
		return errors.New("save affiliate rebate points: missing payment fulfillment lease")
	}
	nextLeaseVersion := lease.nextVersion()
	updatedOrder, err := tx.Client().PaymentOrder.UpdateOneID(o.ID).
		Where(
			paymentorder.StatusEQ(OrderStatusRecharging),
			paymentorder.UpdatedAtEQ(lease.version),
		).
		SetAffiliateRebatePoints(rebateAmount).
		SetUpdatedAt(nextLeaseVersion).
		Save(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return infraerrors.Conflict("CONFLICT", "fulfillment lease was lost while saving affiliate rebate")
		}
		return fmt.Errorf("save affiliate rebate points: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.writeAuditLog(ctx, o.ID, "AFFILIATE_REBATE_FAILED", "system", map[string]any{
			"error": fmt.Sprintf("commit affiliate rebate tx: %v", err),
		})
		return fmt.Errorf("commit affiliate rebate tx: %w", err)
	}
	lease.version = updatedOrder.UpdatedAt
	o.AffiliateRebatePoints = rebateAmount
	return nil
}

func affiliateRebateBaseAmount(o *dbent.PaymentOrder) float64 {
	if o == nil {
		return 0
	}
	if o.OrderType != payment.OrderTypeBalance {
		return 0
	}
	if o.CreditedPoints > 0 {
		return o.CreditedPoints
	}
	return o.Amount
}

func (s *PaymentService) tryClaimAffiliateRebateAudit(ctx context.Context, client *dbent.Client, orderID string, baseAmount float64) (bool, error) {
	if client == nil {
		return false, errors.New("nil payment client")
	}
	oid := orderID
	detail, _ := json.Marshal(map[string]any{
		"baseAmount": baseAmount,
		"status":     "reserved",
	})
	query, args := buildAffiliateRebateAuditClaimQuery(client, oid, string(detail))
	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return false, err
		}
		return false, nil
	}
	var claimID string
	if err := rows.Scan(&claimID); err != nil {
		return false, err
	}
	return true, nil
}

func buildAffiliateRebateAuditClaimQuery(client *dbent.Client, orderID, detail string) (string, []any) {
	nowExpr := paymentAuditCurrentTimestampExpr(client)
	if paymentAuditDialect(client) == dialect.Postgres {
		return fmt.Sprintf(`
INSERT INTO payment_audit_logs (order_id, action, detail, operator, created_at)
SELECT $1::uuid, 'AFFILIATE_REBATE_APPLIED', $2::text, 'system', %s
WHERE NOT EXISTS (
	SELECT 1
	FROM payment_audit_logs
	WHERE order_id = $1::uuid
	  AND action IN ('AFFILIATE_REBATE_APPLIED', 'AFFILIATE_REBATE_SKIPPED')
)
ON CONFLICT (order_id, action) DO NOTHING
RETURNING id`, nowExpr), []any{orderID, detail}
	}
	return fmt.Sprintf(`
INSERT INTO payment_audit_logs (order_id, action, detail, operator, created_at)
SELECT ?, 'AFFILIATE_REBATE_APPLIED', ?, 'system', %s
WHERE NOT EXISTS (
	SELECT 1
	FROM payment_audit_logs
	WHERE order_id = ?
	  AND action IN ('AFFILIATE_REBATE_APPLIED', 'AFFILIATE_REBATE_SKIPPED')
)
ON CONFLICT (order_id, action) DO NOTHING
RETURNING id`, nowExpr), []any{orderID, detail, orderID}
}

func paymentAuditCurrentTimestampExpr(client *dbent.Client) string {
	if paymentAuditDialect(client) == dialect.Postgres {
		return "NOW()"
	}
	return "CURRENT_TIMESTAMP"
}

func paymentAuditDialect(client *dbent.Client) string {
	if client == nil || client.Driver() == nil {
		return ""
	}
	return client.Driver().Dialect()
}

func (s *PaymentService) updateClaimedAffiliateRebateAudit(ctx context.Context, client *dbent.Client, orderID string, action string, detail map[string]any) error {
	if client == nil {
		return errors.New("nil payment client")
	}
	oid := orderID
	detailJSON, _ := json.Marshal(detail)
	updated, err := client.PaymentAuditLog.Update().
		Where(
			paymentauditlog.OrderIDEQ(oid),
			paymentauditlog.ActionEQ("AFFILIATE_REBATE_APPLIED"),
		).
		SetAction(action).
		SetDetail(string(detailJSON)).
		SetOperator("system").
		Save(ctx)
	if err != nil {
		return err
	}
	if updated == 0 {
		return errors.New("affiliate rebate claim log not found")
	}
	return nil
}

func (s *PaymentService) markFailed(ctx context.Context, oid string, lease *paymentFulfillmentLease, cause error) {
	if lease == nil {
		slog.Error("mark FAILED without fulfillment lease", "orderID", oid)
		return
	}
	now := time.Now()
	r := psErrMsg(cause)
	// The lease version prevents a stale worker from overwriting a newer owner.
	c, e := s.entClient.PaymentOrder.Update().
		Where(
			paymentorder.IDEQ(oid),
			paymentorder.StatusEQ(OrderStatusRecharging),
			paymentorder.UpdatedAtEQ(lease.version),
		).
		SetStatus(OrderStatusFailed).SetFailedAt(now).SetFailedReason(r).Save(ctx)
	if e != nil {
		slog.Error("mark FAILED", "orderID", oid, "error", e)
	}
	if c > 0 {
		s.writeAuditLog(ctx, oid, "FULFILLMENT_FAILED", "system", map[string]any{"reason": r})
	}
}

func (s *PaymentService) RetryFulfillment(ctx context.Context, oid string) error {
	o, err := s.entClient.PaymentOrder.Get(ctx, oid)
	if err != nil {
		return infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if o.PaidAt == nil {
		return infraerrors.BadRequest("INVALID_STATUS", "order is not paid")
	}
	if psIsRefundStatus(o.Status) {
		return infraerrors.BadRequest("INVALID_STATUS", "refund-related order cannot retry")
	}
	if o.Status == OrderStatusCompleted {
		return infraerrors.BadRequest("INVALID_STATUS", "order already completed")
	}
	if o.Status != OrderStatusFailed && o.Status != OrderStatusPaid && o.Status != OrderStatusRecharging {
		return infraerrors.BadRequest("INVALID_STATUS", "only paid, failed, and recoverable recharging orders can retry")
	}
	s.writeAuditLog(ctx, oid, "RECHARGE_RETRY", "admin", map[string]any{"detail": "admin manual retry"})
	return s.executeFulfillment(ctx, oid)
}
