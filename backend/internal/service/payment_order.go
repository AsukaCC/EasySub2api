package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/ent/paymentorder"
	"github.com/AsukaCC/EasySub2api/internal/payment"
	"github.com/AsukaCC/EasySub2api/internal/payment/provider"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/pkg/servertiming"
	"github.com/shopspring/decimal"
)

// --- Order Creation ---

func (s *PaymentService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	if req.OrderType == "" {
		req.OrderType = payment.OrderTypeBalance
	}
	if normalized := NormalizeVisibleMethod(req.PaymentType); normalized != "" {
		req.PaymentType = normalized
	}
	cfg, err := s.configService.GetPaymentConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get payment config: %w", err)
	}
	if !cfg.Enabled {
		return nil, infraerrors.Forbidden("PAYMENT_DISABLED", "payment system is disabled")
	}
	plan, err := s.validateOrderInput(ctx, req, cfg)
	if err != nil {
		return nil, err
	}
	if err := s.checkCancelRateLimit(ctx, req.UserID, cfg); err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Status != payment.EntityStatusActive {
		return nil, infraerrors.Forbidden("USER_INACTIVE", "user account is disabled")
	}
	if s.notificationEmailService != nil {
		s.notificationEmailService.RememberRecipientLocale(ctx, req.UserID, user.Email, req.Locale)
	}
	orderAmount := req.Amount
	limitAmount := req.Amount
	walletAmount := 0.0
	walletBonusAmount := 0.0
	walletRechargeAmount := 0.0
	var rechargePricing *rechargeOrderPricing
	if plan != nil {
		orderAmount = plan.Price
		limitAmount = 0
		available := math.Max(user.Balance, 0)
		if available+0.00000001 < plan.Price {
			return nil, infraerrors.BadRequest("INSUFFICIENT_PLATFORM_POINTS", "insufficient platform points for subscription")
		}
		walletBonusAmount = math.Min(math.Max(user.BonusBalance, 0), math.Min(available, plan.Price))
		walletRechargeAmount = math.Max(plan.Price-walletBonusAmount, 0)
		walletAmount = plan.Price
		req.PaymentType = "wallet"
	} else if req.OrderType == payment.OrderTypeBalance {
		rechargePricing = buildRechargeOrderPricing(req.Amount, cfg.RechargeBonusTiers)
		orderAmount = rechargePricing.CreditedPoints
		limitAmount = rechargePricing.PrincipalAmount
	}
	feeRate := cfg.RechargeFeeRate
	walletOnly := plan != nil
	if walletOnly {
		feeRate = 0
	}
	methodCurrency := payment.DefaultPaymentCurrency
	if !walletOnly && strings.TrimSpace(req.PaymentType) == "" {
		return nil, infraerrors.BadRequest("PAYMENT_METHOD_REQUIRED", "payment method is required for the external payment amount")
	}
	if !walletOnly && s.configService != nil {
		methodCurrency, err = s.configService.ValidateMethodCurrencyConsistency(ctx, req.PaymentType)
		if err != nil {
			return nil, err
		}
		if req.OrderType == payment.OrderTypeBalance && methodCurrency != payment.DefaultPaymentCurrency {
			return nil, invalidRechargeCurrency(methodCurrency)
		}
	}
	if !walletOnly && walletRechargeAmount > 0 {
		limits, limitsErr := s.configService.GetAvailableMethodLimits(ctx)
		if limitsErr == nil {
			if methodLimit, ok := limits.Methods[req.PaymentType]; ok && methodLimit.SingleMin > 0 {
				minBase := methodLimit.SingleMin / (1 + feeRate/100)
				if strings.EqualFold(methodCurrency, "CNY") && cfg.SubscriptionUSDToCNYRate > 0 {
					minBase /= cfg.SubscriptionUSDToCNYRate
				}
				if limitAmount < minBase {
					restore := math.Min(walletRechargeAmount, minBase-limitAmount)
					walletAmount -= restore
					limitAmount += restore
				}
				if limitAmount+0.00000001 < minBase {
					return nil, infraerrors.BadRequest("PAYMENT_AMOUNT_BELOW_METHOD_MINIMUM", "remaining external payment amount is below the selected method minimum")
				}
			}
		}
	}
	payAmountStr := "0"
	payAmount := 0.0
	if !walletOnly {
		payAmountStr, payAmount, err = calculateCreateOrderPayAmountForOrderType(limitAmount, feeRate, methodCurrency, req.OrderType, cfg.SubscriptionUSDToCNYRate)
		if err != nil {
			return nil, err
		}
	}
	var sel *payment.InstanceSelection
	if !walletOnly {
		sel, err = s.selectCreateOrderInstance(ctx, req, cfg, payAmount)
		if err != nil {
			return nil, err
		}
		if err := s.validateSelectedCreateOrderInstance(ctx, req, sel); err != nil {
			return nil, err
		}
	}
	selectedCurrency := payment.DefaultPaymentCurrency
	if sel != nil {
		selectedCurrency = paymentProviderConfigCurrency(sel.ProviderKey, sel.Config)
	}
	if !walletOnly && req.OrderType == payment.OrderTypeBalance && selectedCurrency != payment.DefaultPaymentCurrency {
		return nil, invalidRechargeCurrency(selectedCurrency)
	}
	if !walletOnly && selectedCurrency != methodCurrency {
		payAmountStr, payAmount, err = calculateCreateOrderPayAmountForOrderType(limitAmount, feeRate, selectedCurrency, req.OrderType, cfg.SubscriptionUSDToCNYRate)
		if err != nil {
			return nil, err
		}
	}
	if !walletOnly {
		if err := validateSelectedCreateOrderAmountCurrency(payAmountStr, sel); err != nil {
			return nil, err
		}
	}
	var oauthResp *CreateOrderResponse
	if !walletOnly {
		oauthResp, err = s.maybeBuildWeChatOAuthRequiredResponseForSelection(ctx, req, limitAmount, payAmount, feeRate, sel)
	}
	if err != nil {
		return nil, err
	}
	if oauthResp != nil {
		applyRechargePricingToCreateOrderResponse(oauthResp, rechargePricing, payAmount, methodCurrency)
		return oauthResp, nil
	}
	order, err := s.createOrderInTx(ctx, req, user, plan, cfg, orderAmount, limitAmount, feeRate, payAmount, walletAmount, walletOnly, sel, rechargePricing)
	if err != nil {
		return nil, err
	}
	if walletOnly {
		if err := s.ExecuteSubscriptionFulfillment(ctx, order.ID); err != nil {
			return nil, err
		}
		if completed, getErr := s.entClient.PaymentOrder.Get(ctx, order.ID); getErr == nil {
			order = completed
		}
		return &CreateOrderResponse{
			OrderID: order.ID, Amount: order.Amount, PayAmount: 0, WalletAmount: order.WalletAmount,
			WalletBonusAmount: order.WalletBonusAmount, WalletRechargeAmount: order.WalletRechargeAmount,
			GatewayBaseAmount: 0, WalletOnly: true, FeeRate: 0, Status: order.Status,
			PrincipalAmount: order.PrincipalAmount, FeeAmount: order.FeeAmount, Currency: PaymentOrderCurrency(order),
			BasePoints: order.BasePoints, BonusPoints: order.BonusPoints, CreditedPoints: order.CreditedPoints,
			BonusTierSnapshot: order.BonusTierSnapshot, BonusExpiresAt: order.BonusExpiresAt,
			BonusGrantID: order.BonusGrantID, AffiliateRebatePoints: order.AffiliateRebatePoints,
			ResultType: payment.CreatePaymentResultOrderCreated, PaymentType: "wallet", OutTradeNo: order.OutTradeNo, ExpiresAt: order.ExpiresAt,
		}, nil
	}
	resp, err := s.invokeProvider(ctx, order, req, cfg, limitAmount, payAmountStr, payAmount, plan, sel)
	if err != nil {
		if order.WalletHoldID != nil {
			_, _ = s.userRepo.ReleaseWalletHold(ctx, *order.WalletHoldID, "subscription_provider_create_failed")
		}
		_, _ = s.entClient.PaymentOrder.UpdateOneID(order.ID).
			SetStatus(OrderStatusFailed).
			Save(ctx)
		return nil, err
	}
	return resp, nil
}

func (s *PaymentService) validateOrderInput(ctx context.Context, req CreateOrderRequest, cfg *PaymentConfig) (*dbent.SubscriptionPlan, error) {
	switch req.OrderType {
	case payment.OrderTypeBalance:
		if cfg.BalanceDisabled {
			return nil, infraerrors.Forbidden("BALANCE_PAYMENT_DISABLED", "balance recharge has been disabled")
		}
	case payment.OrderTypeSubscription:
		return s.validateSubOrder(ctx, req)
	default:
		return nil, infraerrors.BadRequest("INVALID_ORDER_TYPE", "order type must be balance or subscription")
	}
	if math.IsNaN(req.Amount) || math.IsInf(req.Amount, 0) || req.Amount <= 0 {
		return nil, infraerrors.BadRequest("INVALID_AMOUNT", "amount must be a positive number")
	}
	if !decimal.NewFromFloat(req.Amount).Equal(decimal.NewFromFloat(req.Amount).Round(2)) {
		return nil, infraerrors.BadRequest("INVALID_AMOUNT", "CNY recharge amount allows at most 2 decimal places")
	}
	if (cfg.MinAmount > 0 && req.Amount < cfg.MinAmount) || (cfg.MaxAmount > 0 && req.Amount > cfg.MaxAmount) {
		return nil, infraerrors.BadRequest("INVALID_AMOUNT", "amount out of range").
			WithMetadata(map[string]string{"min": fmt.Sprintf("%.2f", cfg.MinAmount), "max": fmt.Sprintf("%.2f", cfg.MaxAmount)})
	}
	return nil, nil
}

func (s *PaymentService) validateSubOrder(ctx context.Context, req CreateOrderRequest) (*dbent.SubscriptionPlan, error) {
	if !req.UseBalance {
		return nil, infraerrors.BadRequest("SUBSCRIPTION_POINTS_PAYMENT_REQUIRED", "subscription orders must be paid fully with platform points")
	}
	if paymentType := strings.TrimSpace(req.PaymentType); paymentType != "" && !strings.EqualFold(paymentType, "wallet") {
		return nil, infraerrors.BadRequest("SUBSCRIPTION_EXTERNAL_PAYMENT_DISABLED", "external or mixed payment is not supported for subscriptions")
	}
	if req.PlanID == "" {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "subscription order requires a plan")
	}
	plan, err := s.configService.GetPlan(ctx, req.PlanID)
	if err != nil || !plan.ForSale {
		return nil, infraerrors.NotFound("PLAN_NOT_AVAILABLE", "plan not found or not for sale")
	}
	group, err := s.groupRepo.GetByID(ctx, plan.GroupID)
	if err != nil || group.Status != payment.EntityStatusActive {
		return nil, infraerrors.NotFound("GROUP_NOT_FOUND", "subscription group is no longer available")
	}
	if !group.IsSubscriptionType() {
		return nil, infraerrors.BadRequest("GROUP_TYPE_MISMATCH", "group is not a subscription type")
	}
	return plan, nil
}

func (s *PaymentService) createOrderInTx(ctx context.Context, req CreateOrderRequest, user *User, plan *dbent.SubscriptionPlan, cfg *PaymentConfig, orderAmount, limitAmount, feeRate, payAmount, walletAmount float64, walletOnly bool, sel *payment.InstanceSelection, rechargePricing *rechargeOrderPricing) (*dbent.PaymentOrder, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.checkRechargeOrderLimits(ctx, tx, req.OrderType, req.UserID, limitAmount, cfg.MaxPendingOrders, cfg.DailyLimit); err != nil {
		return nil, err
	}
	tm := cfg.OrderTimeoutMin
	if tm <= 0 {
		tm = defaultOrderTimeoutMin
	}
	createdAt := time.Now().UTC()
	exp := createdAt.Add(time.Duration(tm) * time.Minute)
	outTradeNo, err := s.allocateOutTradeNo(ctx, tx)
	if err != nil {
		return nil, err
	}
	providerSnapshot := buildPaymentOrderProviderSnapshot(sel, req)
	selectedInstanceID := ""
	selectedProviderKey := ""
	if sel != nil {
		selectedInstanceID = strings.TrimSpace(sel.InstanceID)
		selectedProviderKey = strings.TrimSpace(sel.ProviderKey)
	}
	b := tx.PaymentOrder.Create().
		SetUserID(req.UserID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetNillableUserNotes(psNilIfEmpty(user.Notes)).
		SetAmount(orderAmount).
		SetPayAmount(payAmount).
		SetWalletAmount(walletAmount).
		SetWalletOnly(walletOnly).
		SetGatewayBaseAmount(limitAmount).
		SetCreatedAt(createdAt).
		SetFeeRate(feeRate).
		SetRechargeCode("").
		SetOutTradeNo(outTradeNo).
		SetPaymentType(req.PaymentType).
		SetPaymentTradeNo("").
		SetOrderType(req.OrderType).
		SetStatus(OrderStatusPending).
		SetExpiresAt(exp).
		SetClientIP(req.ClientIP).
		SetSrcHost(req.SrcHost)
	if rechargePricing != nil {
		b.SetPrincipalAmount(rechargePricing.PrincipalAmount).
			SetFeeAmount(rechargeFeeAmount(payAmount, rechargePricing.PrincipalAmount)).
			SetCurrency(payment.DefaultPaymentCurrency).
			SetBasePoints(rechargePricing.BasePoints).
			SetBonusPoints(rechargePricing.BonusPoints).
			SetCreditedPoints(rechargePricing.CreditedPoints)
		if rechargePricing.BonusTier != nil && rechargePricing.BonusPoints > 0 {
			bonusExpiresAt := createdAt.Add(rechargeBonusValidity)
			rechargePricing.BonusExpiresAt = &bonusExpiresAt
			b.SetBonusTierSnapshot(map[string]any{
				"threshold_cny": rechargePricing.BonusTier.MinAmount,
				"bonus_points":  rechargePricing.BonusTier.BonusPoints,
			}).SetBonusExpiresAt(bonusExpiresAt)
		}
	}
	if req.SrcURL != "" {
		b.SetSrcURL(req.SrcURL)
	}
	if selectedInstanceID != "" {
		b.SetProviderInstanceID(selectedInstanceID)
	}
	if selectedProviderKey != "" {
		b.SetProviderKey(selectedProviderKey)
	}
	if providerSnapshot != nil {
		b.SetProviderSnapshot(providerSnapshot)
	}
	if plan != nil {
		b.SetPlanID(plan.ID).SetSubscriptionGroupID(plan.GroupID).SetSubscriptionDays(psComputeValidityDays(plan.ValidityDays, plan.ValidityUnit))
	}
	if walletOnly {
		b.SetStatus(OrderStatusPaid).SetPaidAt(createdAt)
	}
	order, err := b.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	code := fmt.Sprintf("PAY-%v-%v", order.ID, time.Now().UnixNano()%100000)
	order, err = tx.PaymentOrder.UpdateOneID(order.ID).SetRechargeCode(code).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("set recharge code: %w", err)
	}
	if walletAmount > 0 {
		hold, holdErr := s.userRepo.HoldWallet(dbent.NewTxContext(ctx, tx), WalletHoldInput{
			UserID: req.UserID, Amount: walletAmount, Purpose: "subscription_order", ReferenceID: order.ID,
			RequestFingerprint: order.OutTradeNo, Notes: "subscription balance offset",
		})
		if holdErr != nil {
			return nil, fmt.Errorf("hold subscription wallet balance: %w", holdErr)
		}
		order, err = tx.PaymentOrder.UpdateOneID(order.ID).
			SetWalletHoldID(hold.HoldID).
			SetWalletAmount(hold.Amount).
			SetWalletBonusAmount(hold.BonusAmount).
			SetWalletRechargeAmount(hold.RechargeAmount).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("save subscription wallet hold: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit order transaction: %w", err)
	}
	return order, nil
}

func (s *PaymentService) allocateOutTradeNo(ctx context.Context, tx *dbent.Tx) (string, error) {
	const maxAttempts = 5
	for attempt := 0; attempt < maxAttempts; attempt++ {
		candidate := generateOutTradeNo()
		exists, err := tx.PaymentOrder.Query().Where(paymentorder.OutTradeNo(candidate)).Exist(ctx)
		if err != nil {
			return "", fmt.Errorf("check out_trade_no uniqueness: %w", err)
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("generate unique out_trade_no: exhausted %v attempts", maxAttempts)
}

func (s *PaymentService) checkPendingLimit(ctx context.Context, tx *dbent.Tx, userID string, max int) error {
	if max <= 0 {
		max = defaultMaxPendingOrders
	}
	c, err := tx.PaymentOrder.Query().Where(
		paymentorder.UserIDEQ(userID),
		paymentorder.OrderTypeEQ(payment.OrderTypeBalance),
		paymentorder.StatusEQ(OrderStatusPending),
	).Count(ctx)
	if err != nil {
		return fmt.Errorf("count pending orders: %w", err)
	}
	if c >= max {
		return infraerrors.TooManyRequests("TOO_MANY_PENDING", "too_many_pending").
			WithMetadata(map[string]string{"max": strconv.Itoa(max)})
	}
	return nil
}

func (s *PaymentService) checkRechargeOrderLimits(ctx context.Context, tx *dbent.Tx, orderType, userID string, amount float64, maxPending int, dailyLimit float64) error {
	if orderType != payment.OrderTypeBalance {
		return nil
	}
	if err := s.checkPendingLimit(ctx, tx, userID, maxPending); err != nil {
		return err
	}
	return s.checkDailyLimit(ctx, tx, userID, amount, dailyLimit)
}

func buildPaymentOrderProviderSnapshot(sel *payment.InstanceSelection, req CreateOrderRequest) map[string]any {
	if sel == nil {
		return nil
	}

	snapshot := map[string]any{}
	snapshot["schema_version"] = 2

	instanceID := strings.TrimSpace(sel.InstanceID)
	if instanceID != "" {
		snapshot["provider_instance_id"] = instanceID
	}

	providerKey := strings.TrimSpace(sel.ProviderKey)
	if providerKey != "" {
		snapshot["provider_key"] = providerKey
	}

	paymentMode := strings.TrimSpace(sel.PaymentMode)
	if paymentMode != "" {
		snapshot["payment_mode"] = paymentMode
	}

	if providerKey == payment.TypeWxpay {
		if merchantAppID := paymentOrderSnapshotWxpayAppID(sel, req); merchantAppID != "" {
			snapshot["merchant_app_id"] = merchantAppID
		}
		if merchantID := strings.TrimSpace(sel.Config["mchId"]); merchantID != "" {
			snapshot["merchant_id"] = merchantID
		}
		snapshot["currency"] = payment.DefaultPaymentCurrency
	}
	if providerKey == payment.TypeAlipay {
		if merchantAppID := strings.TrimSpace(sel.Config["appId"]); merchantAppID != "" {
			snapshot["merchant_app_id"] = merchantAppID
		}
	}
	if providerKey == payment.TypeEasyPay {
		if merchantID := strings.TrimSpace(sel.Config["pid"]); merchantID != "" {
			snapshot["merchant_id"] = merchantID
		}
	}
	if providerKey == payment.TypeStripe {
		snapshot["currency"] = paymentProviderConfigCurrency(providerKey, sel.Config)
	}
	if providerKey == payment.TypeAirwallex {
		if accountID := strings.TrimSpace(sel.Config["accountId"]); accountID != "" {
			snapshot["merchant_id"] = accountID
		}
		snapshot["currency"] = paymentProviderConfigCurrency(providerKey, sel.Config)
	}

	if len(snapshot) == 1 {
		return nil
	}
	return snapshot
}

func paymentOrderSnapshotWxpayAppID(sel *payment.InstanceSelection, req CreateOrderRequest) string {
	if sel == nil || strings.TrimSpace(sel.ProviderKey) != payment.TypeWxpay {
		return ""
	}
	if strings.TrimSpace(req.OpenID) != "" {
		return strings.TrimSpace(provider.ResolveWxpayJSAPIAppID(sel.Config))
	}
	return strings.TrimSpace(sel.Config["appId"])
}

func (s *PaymentService) checkDailyLimit(ctx context.Context, tx *dbent.Tx, userID string, amount, limit float64) error {
	if limit <= 0 {
		return nil
	}
	ts := psStartOfDayUTC(time.Now())
	orders, err := tx.PaymentOrder.Query().Where(
		paymentorder.UserIDEQ(userID),
		paymentorder.OrderTypeEQ(payment.OrderTypeBalance),
		paymentorder.PaidAtGTE(ts),
	).All(ctx)
	if err != nil {
		return fmt.Errorf("query daily usage: %w", err)
	}
	used := decimal.Zero
	for _, o := range orders {
		used = used.Add(decimal.NewFromFloat(o.PrincipalAmount).Round(2))
	}
	requested := decimal.NewFromFloat(amount).Round(2)
	limitAmount := decimal.NewFromFloat(limit).Round(2)
	if used.Add(requested).GreaterThan(limitAmount) {
		remaining := limitAmount.Sub(used)
		if remaining.IsNegative() {
			remaining = decimal.Zero
		}
		return infraerrors.TooManyRequests("DAILY_LIMIT_EXCEEDED", "daily_limit_exceeded").
			WithMetadata(map[string]string{"remaining": remaining.StringFixed(2)})
	}
	return nil
}

func (s *PaymentService) selectCreateOrderInstance(ctx context.Context, req CreateOrderRequest, cfg *PaymentConfig, payAmount float64) (*payment.InstanceSelection, error) {
	selectCtx, err := s.prepareCreateOrderSelectionContext(ctx, req)
	if err != nil {
		return nil, err
	}
	sel, err := s.loadBalancer.SelectInstance(selectCtx, "", req.PaymentType, payment.Strategy(cfg.LoadBalanceStrategy), payAmount)
	if err != nil {
		return nil, infraerrors.ServiceUnavailable("PAYMENT_GATEWAY_ERROR", "method_not_configured").
			WithMetadata(map[string]string{"payment_type": req.PaymentType})
	}
	if sel == nil {
		return nil, infraerrors.TooManyRequests("NO_AVAILABLE_INSTANCE", "no_available_instance")
	}
	return sel, nil
}

func (s *PaymentService) prepareCreateOrderSelectionContext(ctx context.Context, req CreateOrderRequest) (context.Context, error) {
	if !requestNeedsWeChatJSAPICompatibility(req) {
		return ctx, nil
	}
	if !s.usesOfficialWxpayVisibleMethod(ctx) {
		return ctx, nil
	}
	expectedAppID, _, err := s.getWeChatPaymentOAuthCredential(ctx)
	if err != nil {
		return nil, err
	}
	return payment.WithWxpayJSAPIAppID(ctx, expectedAppID), nil
}

func requestNeedsWeChatJSAPICompatibility(req CreateOrderRequest) bool {
	if payment.GetBasePaymentType(req.PaymentType) != payment.TypeWxpay {
		return false
	}
	return req.IsWeChatBrowser || strings.TrimSpace(req.OpenID) != ""
}

func (s *PaymentService) usesOfficialWxpayVisibleMethod(ctx context.Context) bool {
	if s == nil || s.configService == nil {
		return false
	}
	inst, err := s.configService.resolveEnabledVisibleMethodInstance(ctx, payment.TypeWxpay)
	if err != nil {
		return false
	}
	if inst == nil {
		return false
	}
	return inst.ProviderKey == payment.TypeWxpay
}

func (s *PaymentService) invokeProvider(ctx context.Context, order *dbent.PaymentOrder, req CreateOrderRequest, cfg *PaymentConfig, limitAmount float64, payAmountStr string, payAmount float64, plan *dbent.SubscriptionPlan, sel *payment.InstanceSelection) (*CreateOrderResponse, error) {
	prov, err := provider.CreateProvider(sel.ProviderKey, sel.InstanceID, sel.Config)
	if err != nil {
		slog.Error("[PaymentService] CreateProvider failed", "provider", sel.ProviderKey, "instance", sel.InstanceID, "error", err)
		// If the provider returned a structured ApplicationError (e.g. WXPAY_CONFIG_MISSING_KEY),
		// pass it through with provider context added to metadata. Otherwise wrap as PAYMENT_PROVIDER_MISCONFIGURED.
		if appErr := new(infraerrors.ApplicationError); errors.As(err, &appErr) {
			md := map[string]string{"provider": sel.ProviderKey, "instance_id": sel.InstanceID}
			for k, v := range appErr.Metadata {
				md[k] = v
			}
			return nil, appErr.WithMetadata(md)
		}
		return nil, infraerrors.ServiceUnavailable("PAYMENT_PROVIDER_MISCONFIGURED", "provider_misconfigured").
			WithMetadata(map[string]string{"provider": sel.ProviderKey, "instance_id": sel.InstanceID})
	}
	subject := s.buildPaymentSubject(plan, limitAmount, cfg, sel)
	outTradeNo := order.OutTradeNo
	canonicalReturnURL, err := CanonicalizeReturnURL(req.ReturnURL, req.SrcHost, req.SrcURL)
	if err != nil {
		return nil, err
	}
	resumeToken := ""
	if resume := s.paymentResume(); resume != nil {
		if canonicalReturnURL != "" && resume.isSigningConfigured() {
			resumeToken, err = resume.CreateToken(ResumeTokenClaims{
				OrderID:            order.ID,
				UserID:             order.UserID,
				ProviderInstanceID: sel.InstanceID,
				ProviderKey:        sel.ProviderKey,
				PaymentType:        req.PaymentType,
				CanonicalReturnURL: canonicalReturnURL,
			})
			if err != nil {
				return nil, fmt.Errorf("create payment resume token: %w", err)
			}
		}
	}
	providerReturnURL, err := buildProviderPaymentReturnURL(
		canonicalReturnURL,
		order.ID,
		outTradeNo,
		resumeToken,
		sel.ProviderKey,
	)
	if err != nil {
		return nil, err
	}
	providerReq := buildProviderCreatePaymentRequest(CreateOrderRequest{
		PaymentType: req.PaymentType,
		OpenID:      req.OpenID,
		ClientIP:    req.ClientIP,
		IsMobile:    req.IsMobile,
		ReturnURL:   providerReturnURL,
	}, sel, outTradeNo, payAmountStr, subject)
	providerReq.AlipayMobilePrecreate = shouldUseAlipayMobilePrecreate(req, cfg, sel)
	finishProviderCall := servertiming.ObserveDependency(ctx, "payment")
	pr, err := prov.CreatePayment(ctx, providerReq)
	finishProviderCall()
	if err != nil {
		slog.Error("[PaymentService] CreatePayment failed", "provider", sel.ProviderKey, "instance", sel.InstanceID, "error", err)
		if appErr := new(infraerrors.ApplicationError); errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, classifyCreatePaymentError(req, sel.ProviderKey, err)
	}
	sanitizeCreatePaymentResponseDetails(pr)
	_, err = s.entClient.PaymentOrder.UpdateOneID(order.ID).
		SetNillablePaymentTradeNo(psNilIfEmpty(pr.TradeNo)).
		SetNillablePayURL(psNilIfEmpty(pr.PayURL)).
		SetNillableQrCode(psNilIfEmpty(pr.QRCode)).
		SetNillableProviderInstanceID(psNilIfEmpty(sel.InstanceID)).
		SetNillableProviderKey(psNilIfEmpty(sel.ProviderKey)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update order with payment details: %w", err)
	}
	s.writeAuditLog(ctx, order.ID, "ORDER_CREATED", fmt.Sprintf("user:%v", req.UserID), map[string]any{
		"paymentAmount":   req.Amount,
		"creditedAmount":  order.Amount,
		"principalAmount": order.PrincipalAmount,
		"feeAmount":       order.FeeAmount,
		"basePoints":      order.BasePoints,
		"bonusPoints":     order.BonusPoints,
		"creditedPoints":  order.CreditedPoints,
		"currency":        PaymentOrderCurrency(order),
		"payAmount":       order.PayAmount,
		"paymentType":     req.PaymentType,
		"orderType":       req.OrderType,
		"paymentSource":   NormalizePaymentSource(req.PaymentSource),
	})
	resultType := pr.ResultType
	if resultType == "" {
		resultType = payment.CreatePaymentResultOrderCreated
	}
	resp := buildCreateOrderResponse(order, req, payAmount, sel, pr, resultType)
	resp.ResumeToken = resumeToken
	resp.AlipayMobilePrecreateDeepLink = providerReq.AlipayMobilePrecreate && strings.TrimSpace(pr.QRCode) != ""
	return resp, nil
}

func shouldUseAlipayMobilePrecreate(req CreateOrderRequest, cfg *PaymentConfig, sel *payment.InstanceSelection) bool {
	return cfg != nil &&
		cfg.AlipayMobilePrecreateDeepLink &&
		req.IsMobile &&
		sel != nil &&
		strings.EqualFold(strings.TrimSpace(sel.ProviderKey), payment.TypeAlipay)
}

func sanitizeCreatePaymentResponseDetails(pr *payment.CreatePaymentResponse) {
	if pr == nil {
		return
	}
	pr.TradeNo = removePostgresTextNUL(pr.TradeNo)
	pr.PayURL = removePostgresTextNUL(pr.PayURL)
	pr.QRCode = removePostgresTextNUL(pr.QRCode)
}

func removePostgresTextNUL(value string) string {
	if !strings.ContainsRune(value, 0) {
		return value
	}
	return strings.ReplaceAll(value, "\x00", "")
}

func buildProviderCreatePaymentRequest(req CreateOrderRequest, sel *payment.InstanceSelection, orderID, amount, subject string) payment.CreatePaymentRequest {
	return payment.CreatePaymentRequest{
		OrderID:            orderID,
		Amount:             amount,
		PaymentType:        req.PaymentType,
		Subject:            subject,
		ReturnURL:          req.ReturnURL,
		OpenID:             strings.TrimSpace(req.OpenID),
		ClientIP:           req.ClientIP,
		IsMobile:           req.IsMobile,
		InstanceSubMethods: selectedInstanceSupportedTypes(sel),
	}
}

func selectedInstanceSupportedTypes(sel *payment.InstanceSelection) string {
	if sel == nil {
		return ""
	}
	return sel.SupportedTypes
}

func (s *PaymentService) buildPaymentSubject(plan *dbent.SubscriptionPlan, limitAmount float64, cfg *PaymentConfig, sel *payment.InstanceSelection) string {
	if plan != nil {
		productName := plan.ProductName
		if productName == "" {
			productName = "EasySub2api Subscription " + plan.Name
		}
		return applyPaymentProductNameAffix(productName, cfg)
	}
	currency := payment.DefaultPaymentCurrency
	if sel != nil {
		currency = paymentProviderConfigCurrency(sel.ProviderKey, sel.Config)
	}
	amountStr := payment.FormatAmountForCurrency(limitAmount, currency)
	if hasPaymentProductNameAffix(cfg) {
		return applyPaymentProductNameAffix(amountStr, cfg)
	}
	return "EasySub2api " + amountStr + " " + currency
}

func hasPaymentProductNameAffix(cfg *PaymentConfig) bool {
	if cfg == nil {
		return false
	}
	pf := strings.TrimSpace(cfg.ProductNamePrefix)
	sf := strings.TrimSpace(cfg.ProductNameSuffix)
	return pf != "" || sf != ""
}

func applyPaymentProductNameAffix(productName string, cfg *PaymentConfig) string {
	if !hasPaymentProductNameAffix(cfg) {
		return productName
	}
	pf := strings.TrimSpace(cfg.ProductNamePrefix)
	sf := strings.TrimSpace(cfg.ProductNameSuffix)
	return strings.TrimSpace(pf + " " + productName + " " + sf)
}

func (s *PaymentService) maybeBuildWeChatOAuthRequiredResponse(ctx context.Context, req CreateOrderRequest, amount, payAmount, feeRate float64) (*CreateOrderResponse, error) {
	return s.maybeBuildWeChatOAuthRequiredResponseForSelection(ctx, req, amount, payAmount, feeRate, nil)
}

func (s *PaymentService) maybeBuildWeChatOAuthRequiredResponseForSelection(ctx context.Context, req CreateOrderRequest, amount, payAmount, feeRate float64, sel *payment.InstanceSelection) (*CreateOrderResponse, error) {
	if sel != nil && sel.ProviderKey != "" && sel.ProviderKey != payment.TypeWxpay {
		return nil, nil
	}
	if strings.TrimSpace(req.OpenID) != "" || !req.IsWeChatBrowser || payment.GetBasePaymentType(req.PaymentType) != payment.TypeWxpay {
		return nil, nil
	}
	return s.buildWeChatOAuthRequiredResponse(ctx, req, amount, payAmount, feeRate)
}

func (s *PaymentService) buildWeChatOAuthRequiredResponse(ctx context.Context, req CreateOrderRequest, amount, payAmount, feeRate float64) (*CreateOrderResponse, error) {
	appID, _, err := s.getWeChatPaymentOAuthCredential(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.paymentResume().ensureSigningKey(); err != nil {
		return nil, err
	}

	authorizeURL, err := buildWeChatPaymentOAuthStartURL(req, "snsapi_base")
	if err != nil {
		return nil, err
	}

	return &CreateOrderResponse{
		Amount:      amount,
		PayAmount:   payAmount,
		FeeRate:     feeRate,
		ResultType:  payment.CreatePaymentResultOAuthRequired,
		PaymentType: req.PaymentType,
		OAuth: &payment.WechatOAuthInfo{
			AuthorizeURL: authorizeURL,
			AppID:        appID,
			Scope:        "snsapi_base",
			RedirectURL:  "/auth/wechat/payment/callback",
		},
	}, nil
}

func (s *PaymentService) validateSelectedCreateOrderInstance(ctx context.Context, req CreateOrderRequest, sel *payment.InstanceSelection) error {
	if !requiresWeChatJSAPICompatibleSelection(req, sel) {
		return nil
	}
	expectedAppID, _, err := s.getWeChatPaymentOAuthCredential(ctx)
	if err != nil {
		return err
	}
	selectedAppID := provider.ResolveWxpayJSAPIAppID(sel.Config)
	if selectedAppID == "" || selectedAppID != expectedAppID {
		return infraerrors.TooManyRequests("NO_AVAILABLE_INSTANCE", "selected payment instance is not compatible with the current WeChat OAuth app")
	}
	return nil
}

func calculateCreateOrderPayAmount(limitAmount, feeRate float64, currency string) (string, float64, error) {
	if err := validateCreateOrderAmountCurrency(limitAmount, currency); err != nil {
		return "", 0, err
	}
	payAmountStr := payment.CalculatePayAmountForCurrency(limitAmount, feeRate, currency)
	if _, err := payment.AmountToMinorUnit(payAmountStr, currency); err != nil {
		return "", 0, infraerrors.BadRequest("INVALID_AMOUNT", err.Error()).
			WithMetadata(map[string]string{"currency": currency})
	}
	payAmount, err := strconv.ParseFloat(payAmountStr, 64)
	if err != nil {
		return "", 0, infraerrors.BadRequest("INVALID_AMOUNT", "invalid payment amount").
			WithMetadata(map[string]string{"currency": currency})
	}
	return payAmountStr, payAmount, nil
}

func calculateCreateOrderPayAmountForOrderType(limitAmount, feeRate float64, currency, orderType string, usdToCnyRate float64) (string, float64, error) {
	// orderType and usdToCnyRate remain in this internal compatibility helper so
	// staggered callers still compile. Subscription checkout is wallet-only, and
	// the legacy exchange-rate setting must never alter a charge amount.
	_ = orderType
	_ = usdToCnyRate
	return calculateCreateOrderPayAmount(limitAmount, feeRate, currency)
}

func validateCreateOrderAmountCurrency(amount float64, currency string) error {
	amountStr := strconv.FormatFloat(amount, 'f', -1, 64)
	if _, err := payment.AmountToMinorUnit(amountStr, currency); err != nil {
		return infraerrors.BadRequest("INVALID_AMOUNT", err.Error()).
			WithMetadata(map[string]string{"currency": currency})
	}
	return nil
}

func validateSelectedCreateOrderAmountCurrency(payAmount string, sel *payment.InstanceSelection) error {
	if sel == nil {
		return nil
	}
	currency := paymentProviderConfigCurrency(sel.ProviderKey, sel.Config)
	if _, err := payment.AmountToMinorUnit(payAmount, currency); err != nil {
		return infraerrors.BadRequest("INVALID_AMOUNT", err.Error()).
			WithMetadata(map[string]string{"currency": currency})
	}
	return nil
}

func invalidRechargeCurrency(currency string) error {
	return infraerrors.BadRequest("UNSUPPORTED_RECHARGE_CURRENCY", "balance recharge only supports CNY").
		WithMetadata(map[string]string{"currency": strings.ToUpper(strings.TrimSpace(currency))})
}

func applyRechargePricingToCreateOrderResponse(resp *CreateOrderResponse, pricing *rechargeOrderPricing, payAmount float64, currency string) {
	if resp == nil || pricing == nil {
		return
	}
	resp.Amount = pricing.CreditedPoints
	resp.PrincipalAmount = pricing.PrincipalAmount
	resp.FeeAmount = rechargeFeeAmount(payAmount, pricing.PrincipalAmount)
	resp.BasePoints = pricing.BasePoints
	resp.BonusPoints = pricing.BonusPoints
	resp.CreditedPoints = pricing.CreditedPoints
	resp.Currency = currency
	if pricing.BonusTier != nil && pricing.BonusPoints > 0 {
		resp.BonusTierSnapshot = map[string]any{
			"threshold_cny": pricing.BonusTier.MinAmount,
			"bonus_points":  pricing.BonusTier.BonusPoints,
		}
	}
}

func requiresWeChatJSAPICompatibleSelection(req CreateOrderRequest, sel *payment.InstanceSelection) bool {
	if sel == nil || sel.ProviderKey != payment.TypeWxpay || payment.GetBasePaymentType(req.PaymentType) != payment.TypeWxpay {
		return false
	}
	return req.IsWeChatBrowser || strings.TrimSpace(req.OpenID) != ""
}

func (s *PaymentService) getWeChatPaymentOAuthCredential(ctx context.Context) (string, string, error) {
	if s == nil || s.configService == nil || s.configService.settingRepo == nil {
		return "", "", infraerrors.ServiceUnavailable(
			"WECHAT_PAYMENT_MP_NOT_CONFIGURED",
			"wechat in-app payment requires a complete WeChat MP OAuth credential",
		)
	}
	cfg, err := (&SettingService{settingRepo: s.configService.settingRepo}).GetWeChatConnectOAuthConfig(ctx)
	appID := strings.TrimSpace(cfg.AppIDForMode("mp"))
	appSecret := strings.TrimSpace(cfg.AppSecretForMode("mp"))
	if err != nil || !cfg.SupportsMode("mp") || appID == "" || appSecret == "" {
		return "", "", infraerrors.ServiceUnavailable(
			"WECHAT_PAYMENT_MP_NOT_CONFIGURED",
			"wechat in-app payment requires a complete WeChat MP OAuth credential",
		)
	}
	return appID, appSecret, nil
}

func classifyCreatePaymentError(req CreateOrderRequest, providerKey string, err error) error {
	if err == nil {
		return nil
	}
	if providerKey == payment.TypeWxpay &&
		payment.GetBasePaymentType(req.PaymentType) == payment.TypeWxpay &&
		strings.Contains(err.Error(), "wxpay h5 payments are not authorized for this merchant") {
		return infraerrors.ServiceUnavailable(
			"WECHAT_H5_NOT_AUTHORIZED",
			"wechat h5 payment is not available for this merchant",
		).WithMetadata(map[string]string{
			"action": "open_in_wechat_or_scan_qr",
		})
	}
	return infraerrors.ServiceUnavailable("PAYMENT_GATEWAY_ERROR", fmt.Sprintf("payment gateway error: %s", err.Error()))
}

func buildCreateOrderResponse(order *dbent.PaymentOrder, req CreateOrderRequest, payAmount float64, sel *payment.InstanceSelection, pr *payment.CreatePaymentResponse, resultType payment.CreatePaymentResultType) *CreateOrderResponse {
	return &CreateOrderResponse{
		OrderID:               order.ID,
		Amount:                order.Amount,
		PayAmount:             payAmount,
		WalletAmount:          order.WalletAmount,
		WalletBonusAmount:     order.WalletBonusAmount,
		WalletRechargeAmount:  order.WalletRechargeAmount,
		GatewayBaseAmount:     order.GatewayBaseAmount,
		WalletOnly:            order.WalletOnly,
		FeeRate:               order.FeeRate,
		PrincipalAmount:       order.PrincipalAmount,
		FeeAmount:             order.FeeAmount,
		BasePoints:            order.BasePoints,
		BonusPoints:           order.BonusPoints,
		CreditedPoints:        order.CreditedPoints,
		BonusTierSnapshot:     order.BonusTierSnapshot,
		BonusExpiresAt:        order.BonusExpiresAt,
		BonusGrantID:          order.BonusGrantID,
		AffiliateRebatePoints: order.AffiliateRebatePoints,
		Status:                OrderStatusPending,
		ResultType:            resultType,
		PaymentType:           req.PaymentType,
		OutTradeNo:            order.OutTradeNo,
		PayURL:                pr.PayURL,
		QRCode:                pr.QRCode,
		ClientSecret:          pr.ClientSecret,
		IntentID:              pr.IntentID,
		Currency:              PaymentOrderCurrency(order),
		CountryCode:           pr.CountryCode,
		PaymentEnv:            pr.PaymentEnv,
		OAuth:                 pr.OAuth,
		JSAPI:                 pr.JSAPI,
		JSAPIPayload:          pr.JSAPI,
		ExpiresAt:             order.ExpiresAt,
		PaymentMode:           sel.PaymentMode,
	}
}

func buildWeChatPaymentOAuthStartURL(req CreateOrderRequest, scope string) (string, error) {
	u, err := url.Parse("/api/v1/auth/oauth/wechat/payment/start")
	if err != nil {
		return "", fmt.Errorf("build wechat payment oauth start url: %w", err)
	}
	q := u.Query()
	q.Set("payment_type", strings.TrimSpace(req.PaymentType))
	if req.Amount > 0 {
		q.Set("amount", strconv.FormatFloat(req.Amount, 'f', -1, 64))
	}
	if orderType := strings.TrimSpace(req.OrderType); orderType != "" {
		q.Set("order_type", orderType)
	}
	if req.PlanID != "" {
		q.Set("plan_id", req.PlanID)
	}
	if scope = strings.TrimSpace(scope); scope != "" {
		q.Set("scope", scope)
	}
	if redirectTo := paymentRedirectPathFromURL(req.SrcURL); redirectTo != "" {
		q.Set("redirect", redirectTo)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func paymentRedirectPathFromURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "/purchase"
	}
	if strings.HasPrefix(rawURL, "/") && !strings.HasPrefix(rawURL, "//") {
		return normalizePaymentRedirectPath(rawURL)
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "/purchase"
	}
	path := strings.TrimSpace(u.EscapedPath())
	if path == "" {
		path = strings.TrimSpace(u.Path)
	}
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return "/purchase"
	}
	if strings.TrimSpace(u.RawQuery) != "" {
		path += "?" + u.RawQuery
	}
	return normalizePaymentRedirectPath(path)
}

func normalizePaymentRedirectPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/purchase"
	}
	if path == "/payment" {
		return "/purchase"
	}
	if strings.HasPrefix(path, "/payment?") {
		return "/purchase" + strings.TrimPrefix(path, "/payment")
	}
	return path
}

// --- Order Queries ---

func (s *PaymentService) GetOrder(ctx context.Context, orderID, userID string) (*dbent.PaymentOrder, error) {
	o, err := s.entClient.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if o.UserID != userID {
		return nil, infraerrors.Forbidden("FORBIDDEN", "no permission for this order")
	}
	return o, nil
}

func (s *PaymentService) GetOrderByID(ctx context.Context, orderID string) (*dbent.PaymentOrder, error) {
	o, err := s.entClient.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	return o, nil
}

func (s *PaymentService) GetUserOrders(ctx context.Context, userID string, p OrderListParams) ([]*dbent.PaymentOrder, int, error) {
	q := s.entClient.PaymentOrder.Query().Where(paymentorder.UserIDEQ(userID))
	if p.Status != "" {
		q = q.Where(paymentorder.StatusEQ(p.Status))
	}
	if p.OrderType != "" {
		q = q.Where(paymentorder.OrderTypeEQ(p.OrderType))
	}
	if p.PaymentType != "" {
		q = q.Where(paymentorder.PaymentTypeEQ(p.PaymentType))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count user orders: %w", err)
	}
	ps, pg := applyPagination(p.PageSize, p.Page)
	orders, err := q.Order(dbent.Desc(paymentorder.FieldCreatedAt)).Limit(ps).Offset((pg - 1) * ps).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("query user orders: %w", err)
	}
	return orders, total, nil
}

// AdminListOrders returns a paginated list of orders. If userID != "", filters by user.
func (s *PaymentService) AdminListOrders(ctx context.Context, userID string, p OrderListParams) ([]*dbent.PaymentOrder, int, error) {
	q := s.entClient.PaymentOrder.Query()
	if userID != "" {
		q = q.Where(paymentorder.UserIDEQ(userID))
	}
	if p.Status != "" {
		q = q.Where(paymentorder.StatusEQ(p.Status))
	}
	if p.OrderType != "" {
		q = q.Where(paymentorder.OrderTypeEQ(p.OrderType))
	}
	if p.PaymentType != "" {
		q = q.Where(paymentorder.PaymentTypeEQ(p.PaymentType))
	}
	if p.Keyword != "" {
		q = q.Where(paymentorder.Or(
			paymentorder.OutTradeNoContainsFold(p.Keyword),
			paymentorder.UserEmailContainsFold(p.Keyword),
			paymentorder.UserNameContainsFold(p.Keyword),
		))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count admin orders: %w", err)
	}
	ps, pg := applyPagination(p.PageSize, p.Page)
	orders, err := q.Order(dbent.Desc(paymentorder.FieldCreatedAt)).Limit(ps).Offset((pg - 1) * ps).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("query admin orders: %w", err)
	}
	return orders, total, nil
}
