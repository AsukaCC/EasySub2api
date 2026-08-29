package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

type rechargeWalletCreditSpy struct {
	UserRepository
	inputs       []WalletCreditInput
	bonusGrantID *string
}

type subscriptionHoldLifecycleSpy struct {
	UserRepository
	releasedHoldIDs []string
	captureErr      error
	debitCalls      int
}

func (s *subscriptionHoldLifecycleSpy) ReleaseWalletHold(_ context.Context, holdID, _ string) (WalletHoldResult, error) {
	s.releasedHoldIDs = append(s.releasedHoldIDs, holdID)
	return WalletHoldResult{HoldID: holdID, Status: "released", Amount: 25}, nil
}

func (s *subscriptionHoldLifecycleSpy) CaptureWalletHold(context.Context, string, string) (WalletHoldResult, error) {
	return WalletHoldResult{}, s.captureErr
}

func (s *subscriptionHoldLifecycleSpy) DebitWallet(context.Context, WalletDebitInput) (WalletMutationResult, error) {
	s.debitCalls++
	return WalletMutationResult{}, nil
}

type unavailableSubscriptionGroupRepo struct {
	GroupRepository
}

func TestValidateOrderInputRejectsUnknownOrderType(t *testing.T) {
	t.Parallel()

	_, err := (&PaymentService{}).validateOrderInput(context.Background(), CreateOrderRequest{
		OrderType: "gift",
		Amount:    10,
	}, &PaymentConfig{})
	require.ErrorContains(t, err, "order type must be balance or subscription")
}

func (unavailableSubscriptionGroupRepo) GetByID(context.Context, string) (*Group, error) {
	return nil, ErrGroupNotFound
}

func (s *rechargeWalletCreditSpy) CreditWallet(_ context.Context, input WalletCreditInput) (WalletMutationResult, error) {
	s.inputs = append(s.inputs, input)
	result := WalletMutationResult{}
	if input.Kind == WalletKindBonus {
		result.BonusGrantID = s.bonusGrantID
	}
	return result, nil
}

func TestCreditBalanceRechargePointsCountsBaseAndBonusAsRecharged(t *testing.T) {
	t.Parallel()
	expiresAt := time.Now().Add(rechargeBonusValidity)
	repo := &rechargeWalletCreditSpy{}
	svc := &PaymentService{userRepo: repo}
	order := &dbent.PaymentOrder{
		ID: "0198-test-order", UserID: "0198-test-user",
		BasePoints: 100, BonusPoints: 25, BonusExpiresAt: &expiresAt,
	}

	require.NoError(t, svc.creditBalanceRechargePoints(context.Background(), order))
	require.Len(t, repo.inputs, 2)
	require.Equal(t, WalletKindRecharge, repo.inputs[0].Kind)
	require.Equal(t, WalletKindBonus, repo.inputs[1].Kind)
	require.True(t, repo.inputs[0].CountAsRecharged)
	require.True(t, repo.inputs[1].CountAsRecharged)
	require.Equal(t, "wallet-payment-base:"+order.ID, repo.inputs[0].IdempotencyKey)
	require.Equal(t, "wallet-payment-bonus:"+order.ID, repo.inputs[1].IdempotencyKey)
	require.Equal(t, &expiresAt, repo.inputs[1].ExpiresAt)

	totalRecharged := 0.0
	for _, input := range repo.inputs {
		if input.CountAsRecharged {
			totalRecharged += input.Amount
		}
	}
	require.Equal(t, 125.0, totalRecharged)
}

func TestRechargeBonusWritebackAdvancesLeaseAndAllowsCompletion(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("bonus-lease@example.com").
		SetPasswordHash("hash").
		SetUsername("bonus-lease-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(110).
		SetPayAmount(100).
		SetFeeRate(0).
		SetPrincipalAmount(100).
		SetCurrency("CNY").
		SetBasePoints(100).
		SetBonusPoints(10).
		SetCreditedPoints(110).
		SetRechargeCode("BONUS-LEASE-CODE").
		SetOutTradeNo("bonus_lease_order").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("bonus-lease-trade").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPaid).
		SetPaidAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	bonusGrantID := "00000000-0000-0000-0000-000000000401"
	svc := &PaymentService{
		entClient: client,
		userRepo:  &rechargeWalletCreditSpy{bonusGrantID: &bonusGrantID},
	}
	lease, err := svc.acquirePaymentFulfillmentLease(ctx, order)
	require.NoError(t, err)
	require.NotNil(t, lease)
	initialVersion := lease.version

	require.NoError(t, svc.creditBalanceRechargePoints(ctx, order, lease))
	require.False(t, lease.version.Equal(initialVersion), "bonus writeback must advance the active lease version")
	afterBonus, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, afterBonus.UpdatedAt, lease.version)
	require.Equal(t, OrderStatusRecharging, afterBonus.Status)
	require.NoError(t, svc.markCompleted(ctx, order, lease, "RECHARGE_SUCCESS"))

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
	require.NotNil(t, reloaded.BonusGrantID)
	require.Equal(t, bonusGrantID, *reloaded.BonusGrantID)
}

func TestBuildRechargeOrderPricingSelectsOnlyHighestEligibleTier(t *testing.T) {
	t.Parallel()
	pricing := buildRechargeOrderPricing(100, []RechargeBonusTier{
		{MinAmount: 50, BonusPoints: 5},
		{MinAmount: 100, BonusPoints: 12},
		{MinAmount: 200, BonusPoints: 30},
	})

	require.Equal(t, 100.0, pricing.BasePoints)
	require.Equal(t, 12.0, pricing.BonusPoints)
	require.Equal(t, 112.0, pricing.CreditedPoints)
	require.NotNil(t, pricing.BonusTier)
	require.Equal(t, 100.0, pricing.BonusTier.MinAmount)
}

func TestRechargeFeeNeverProducesPlatformPoints(t *testing.T) {
	t.Parallel()
	pricing := buildRechargeOrderPricing(100, []RechargeBonusTier{{MinAmount: 100, BonusPoints: 10}})
	payAmountText, payAmount, err := calculateCreateOrderPayAmount(100, 3, payment.DefaultPaymentCurrency)
	require.NoError(t, err)
	require.Equal(t, "103.00", payAmountText)
	require.Equal(t, 3.0, rechargeFeeAmount(payAmount, pricing.PrincipalAmount))
	require.Equal(t, 100.0, pricing.BasePoints)
	require.Equal(t, 10.0, pricing.BonusPoints)
	require.Equal(t, 110.0, pricing.CreditedPoints)
}

func TestSubscriptionValidationRejectsExternalAndMixedPaymentBeforeOrderCreation(t *testing.T) {
	t.Parallel()
	svc := &PaymentService{}
	_, err := svc.validateSubOrder(context.Background(), CreateOrderRequest{
		OrderType: payment.OrderTypeSubscription, UseBalance: false,
	})
	require.ErrorContains(t, err, "paid fully with platform points")

	_, err = svc.validateSubOrder(context.Background(), CreateOrderRequest{
		OrderType: payment.OrderTypeSubscription, UseBalance: true, PaymentType: payment.TypeStripe,
	})
	require.ErrorContains(t, err, "external or mixed payment is not supported")
}

func TestRechargeOrderLimitsIgnoreSubscriptionsAndUsePrincipalCNY(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("recharge-limit-units@example.com").
		SetPasswordHash("hash").
		SetUsername("recharge-limit-units").
		Save(ctx)
	require.NoError(t, err)

	createOrder := func(tradeNo, orderType, status string, amount, payAmount, principal float64) {
		t.Helper()
		builder := client.PaymentOrder.Create().
			SetUserID(user.ID).
			SetUserEmail(user.Email).
			SetUserName(user.Username).
			SetAmount(amount).
			SetPayAmount(payAmount).
			SetPrincipalAmount(principal).
			SetRechargeCode("LIMIT-" + tradeNo).
			SetOutTradeNo(tradeNo).
			SetPaymentType("wallet").
			SetPaymentTradeNo("").
			SetOrderType(orderType).
			SetStatus(status).
			SetExpiresAt(time.Now().Add(time.Hour)).
			SetClientIP("127.0.0.1").
			SetSrcHost("api.example.com")
		if status != OrderStatusPending {
			builder.SetPaidAt(time.Now())
		}
		_, createErr := builder.Save(ctx)
		require.NoError(t, createErr)
	}

	// Subscription points and recharge fees must not consume the CNY principal limit.
	createOrder("limit_subscription_paid", payment.OrderTypeSubscription, OrderStatusPaid, 1000, 0, 0)
	createOrder("limit_balance_completed", payment.OrderTypeBalance, OrderStatusCompleted, 50, 42, 40)
	for index, status := range []string{
		OrderStatusRefunding,
		OrderStatusRefundPending,
		OrderStatusRefundFailed,
		OrderStatusPartiallyRefunded,
		OrderStatusRefunded,
	} {
		createOrder(fmt.Sprintf("limit_balance_refund_%d", index), payment.OrderTypeBalance, status, 1, 1, 1)
	}
	// A pending subscription must not consume the pending recharge-order allowance.
	createOrder("limit_subscription_pending", payment.OrderTypeSubscription, OrderStatusPending, 500, 0, 0)
	createOrder("limit_balance_pending", payment.OrderTypeBalance, OrderStatusPending, 10, 10, 10)

	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	svc := &PaymentService{}
	require.NoError(t, svc.checkRechargeOrderLimits(ctx, tx, payment.OrderTypeSubscription, user.ID, 0, 1, 1))
	require.NoError(t, svc.checkRechargeOrderLimits(ctx, tx, payment.OrderTypeBalance, user.ID, 55, 2, 100))
	require.ErrorContains(t, svc.checkRechargeOrderLimits(ctx, tx, payment.OrderTypeBalance, user.ID, 55.01, 2, 100), "daily_limit_exceeded")
	require.ErrorContains(t, svc.checkRechargeOrderLimits(ctx, tx, payment.OrderTypeBalance, user.ID, 1, 1, 100), "too_many_pending")
}

func TestRechargeDailyLimitUsesExactCNYCentBoundaries(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("recharge-limit-cents@example.com").
		SetPasswordHash("hash").
		SetUsername("recharge-limit-cents").
		Save(ctx)
	require.NoError(t, err)
	_, err = client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(0.1).
		SetPayAmount(0.1).
		SetPrincipalAmount(0.1).
		SetRechargeCode("LIMIT-CENTS-PAID").
		SetOutTradeNo("limit_cents_paid").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetPaidAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	svc := &PaymentService{}
	require.NoError(t, svc.checkDailyLimit(ctx, tx, user.ID, 0.2, 0.3))
	require.ErrorContains(t, svc.checkDailyLimit(ctx, tx, user.ID, 0.21, 0.3), "daily_limit_exceeded")
}

func TestPaymentDashboardUsesNetSuccessfulGatewayRevenueAcrossRefundStates(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("payment-stats-refunds@example.com").
		SetPasswordHash("hash").
		SetUsername("payment-stats-refunds").
		Save(ctx)
	require.NoError(t, err)

	createOrder := func(tradeNo, orderType, status string, payAmount, refundedGateway float64, paid bool) {
		t.Helper()
		builder := client.PaymentOrder.Create().
			SetUserID(user.ID).
			SetUserEmail(user.Email).
			SetUserName(user.Username).
			SetAmount(payAmount).
			SetPayAmount(payAmount).
			SetPrincipalAmount(payAmount).
			SetRefundedGatewayAmount(refundedGateway).
			SetCurrency("CNY").
			SetRechargeCode("STATS-" + tradeNo).
			SetOutTradeNo(tradeNo).
			SetPaymentType(payment.TypeAlipay).
			SetPaymentTradeNo("").
			SetOrderType(orderType).
			SetStatus(status).
			SetExpiresAt(time.Now().Add(time.Hour)).
			SetClientIP("127.0.0.1").
			SetSrcHost("api.example.com")
		if paid {
			builder.SetPaidAt(time.Now())
		}
		_, createErr := builder.Save(ctx)
		require.NoError(t, createErr)
	}

	createOrder("stats_completed", payment.OrderTypeBalance, OrderStatusCompleted, 103, 0, true)
	createOrder("stats_partial", payment.OrderTypeBalance, OrderStatusPartiallyRefunded, 103, 51.5, true)
	createOrder("stats_refund_pending", payment.OrderTypeBalance, OrderStatusRefundPending, 103, 20.6, true)
	createOrder("stats_refund_failed", payment.OrderTypeBalance, OrderStatusRefundFailed, 103, 0, true)
	createOrder("stats_refunded", payment.OrderTypeBalance, OrderStatusRefunded, 103, 103, true)
	// Neither paid subscriptions nor pending subscriptions belong to recharge revenue or pending recharge counts.
	createOrder("stats_subscription_paid", payment.OrderTypeSubscription, OrderStatusPaid, 999, 0, true)
	createOrder("stats_subscription_pending", payment.OrderTypeSubscription, OrderStatusPending, 999, 0, false)
	createOrder("stats_balance_pending", payment.OrderTypeBalance, OrderStatusPending, 99, 0, false)

	stats, err := (&PaymentService{entClient: client}).GetDashboardStats(ctx, 30)
	require.NoError(t, err)
	require.Equal(t, 5, stats.TotalCount, "all paid recharge orders remain counted across refund states")
	require.Equal(t, 5, stats.TodayCount)
	require.Equal(t, 1, stats.PendingOrders)
	require.InDelta(t, 339.9, stats.TotalAmount["CNY"], 1e-9)
	require.InDelta(t, 339.9, stats.TodayAmount["CNY"], 1e-9)
	require.InDelta(t, 67.98, stats.AvgAmount["CNY"], 1e-9)
	require.Equal(t, []PaymentMethodStat{{
		Type: payment.TypeAlipay, Amount: CurrencyAmounts{"CNY": 339.9}, Count: 5,
	}}, stats.PaymentMethods)
	require.Equal(t, TopUsersByCurrency{"CNY": {{
		UserID: user.ID, Email: user.Email, Amount: 339.9,
	}}}, stats.TopUsers)
}

func TestRechargeBonusTierUsesThresholdCNYAndDecimalPrecision(t *testing.T) {
	t.Parallel()
	tiers, err := normalizeRechargeBonusTiers([]RechargeBonusTier{
		{MinAmount: 0.29, BonusPoints: 0.12345678},
	})
	require.NoError(t, err)
	require.Equal(t, 0.29, tiers[0].MinAmount)

	body, err := json.Marshal(tiers[0])
	require.NoError(t, err)
	require.Contains(t, string(body), `"threshold_cny":0.29`)
	require.False(t, strings.Contains(string(body), `"min_amount"`))

	legacy := parseRechargeBonusTiers(`[{"min_amount":10,"bonus_points":1.25}]`)
	require.Len(t, legacy, 1)
	require.Equal(t, 10.0, legacy[0].MinAmount)
}

func TestRechargeBonusTierRejectsJSONPrecisionBeforeFloatConversion(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		body string
	}{
		{name: "threshold exceeds CNY precision", body: `{"threshold_cny":0.001,"bonus_points":1}`},
		{name: "bonus exceeds points precision", body: `{"threshold_cny":1,"bonus_points":0.000000001}`},
		{name: "legacy threshold exceeds CNY precision", body: `{"min_amount":10.001,"bonus_points":1}`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var tier RechargeBonusTier
			require.Error(t, json.Unmarshal([]byte(testCase.body), &tier))
		})
	}
}

func TestCompletedRechargeRefundDeadlineIsFixedAtSevenDays(t *testing.T) {
	t.Parallel()
	completedAt := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	deadline := completedRechargeRefundDeadline(&dbent.PaymentOrder{OrderType: "balance"}, completedAt)
	require.NotNil(t, deadline)
	require.Equal(t, completedAt.Add(168*time.Hour), *deadline)
	require.Nil(t, completedRechargeRefundDeadline(&dbent.PaymentOrder{OrderType: "subscription"}, completedAt))
}

type fixedFreezeSettingRepo struct{}

func (fixedFreezeSettingRepo) Get(context.Context, string) (*Setting, error)    { return nil, nil }
func (fixedFreezeSettingRepo) GetValue(context.Context, string) (string, error) { return "0", nil }
func (fixedFreezeSettingRepo) Set(context.Context, string, string) error        { return nil }
func (fixedFreezeSettingRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	return nil, nil
}
func (fixedFreezeSettingRepo) SetMultiple(context.Context, map[string]string) error { return nil }
func (fixedFreezeSettingRepo) GetAll(context.Context) (map[string]string, error)    { return nil, nil }
func (fixedFreezeSettingRepo) Delete(context.Context, string) error                 { return nil }

func TestAffiliateRebateFreezeHoursIgnoresLegacyConfiguredValue(t *testing.T) {
	t.Parallel()
	svc := NewSettingService(fixedFreezeSettingRepo{}, nil)
	require.Equal(t, 168, svc.GetAffiliateRebateFreezeHours(context.Background()))
}

func TestAffiliateRebateBaseIncludesRechargeBonusButExcludesSubscriptions(t *testing.T) {
	t.Parallel()
	require.Equal(t, 110.0, affiliateRebateBaseAmount(&dbent.PaymentOrder{
		OrderType: payment.OrderTypeBalance, Amount: 110, CreditedPoints: 110,
	}))
	require.Zero(t, affiliateRebateBaseAmount(&dbent.PaymentOrder{
		OrderType: payment.OrderTypeSubscription, Amount: 110, CreditedPoints: 110,
	}))
}

func TestRefundProviderEligibilitySeparatesDirectAndTicketPaths(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	direct, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("direct refund").
		SetConfig("{}").
		SetSupportedTypes("alipay").
		SetRefundEnabled(true).
		SetAllowUserRefund(true).
		Save(ctx)
	require.NoError(t, err)
	ticketOnly, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeWxpay).
		SetName("ticket refund").
		SetConfig("{}").
		SetSupportedTypes("wxpay").
		SetRefundEnabled(true).
		SetAllowUserRefund(false).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeStripe).
		SetName("refund disabled").
		SetConfig("{}").
		SetSupportedTypes("stripe").
		SetRefundEnabled(false).
		SetAllowUserRefund(false).
		Save(ctx)
	require.NoError(t, err)

	userIDs, enabledIDs, err := (&PaymentConfigService{entClient: client}).GetRefundEligibleInstanceIDs(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{direct.ID}, userIDs)
	require.ElementsMatch(t, []string{direct.ID, ticketOnly.ID}, enabledIDs)
}

type completedRechargeCodeRepo struct {
	RedeemCodeRepository
	code *RedeemCode
}

func (r completedRechargeCodeRepo) GetByCode(context.Context, string) (*RedeemCode, error) {
	return r.code, nil
}

func TestPaymentNotificationPreservesRechargePayAmountSnapshot(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("pay-snapshot@example.com").
		SetPasswordHash("hash").
		SetUsername("pay-snapshot-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(110).
		SetPayAmount(103).
		SetFeeRate(3).
		SetGatewayBaseAmount(100).
		SetPrincipalAmount(100).
		SetFeeAmount(3).
		SetCurrency("CNY").
		SetBasePoints(100).
		SetBonusPoints(10).
		SetCreditedPoints(110).
		SetRechargeCode("PAY-SNAPSHOT-CODE").
		SetOutTradeNo("pay_snapshot_order").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	code := &RedeemCode{ID: "pay-snapshot-code-id", Code: order.RechargeCode, Status: StatusUsed}
	svc := &PaymentService{
		entClient: client,
		redeemService: &RedeemService{
			redeemRepo: completedRechargeCodeRepo{code: code},
		},
	}
	providerReportedAmount := 102.999
	require.NoError(t, svc.HandlePaymentNotification(ctx, &payment.PaymentNotification{
		TradeNo: "provider-trade-snapshot",
		OrderID: order.OutTradeNo,
		Amount:  providerReportedAmount,
		Status:  payment.NotificationStatusSuccess,
	}, payment.TypeAlipay))

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
	require.Equal(t, 103.0, reloaded.PayAmount)
	require.Equal(t, 100.0, reloaded.PrincipalAmount)
	require.Equal(t, 3.0, reloaded.FeeAmount)

	logs, err := svc.GetOrderAuditLogs(ctx, order.ID)
	require.NoError(t, err)
	var paidDetail map[string]any
	for _, log := range logs {
		if log.Action == "ORDER_PAID" {
			require.NoError(t, json.Unmarshal([]byte(log.Detail), &paidDetail))
			break
		}
	}
	require.Equal(t, 103.0, paidDetail["expectedPayAmount"])
	require.Equal(t, providerReportedAmount, paidDetail["providerReportedAmount"])
}

func TestFailedWalletOnlySubscriptionReleasesHeldPoints(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("subscription-release@example.com").
		SetPasswordHash("hash").
		SetUsername("subscription-release-user").
		Save(ctx)
	require.NoError(t, err)

	holdID := "00000000-0000-0000-0000-000000000101"
	groupID := "00000000-0000-0000-0000-000000000102"
	planID := "00000000-0000-0000-0000-000000000103"
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(25).
		SetPayAmount(0).
		SetWalletAmount(25).
		SetWalletOnly(true).
		SetWalletHoldID(holdID).
		SetFeeRate(0).
		SetRechargeCode("SUBSCRIPTION-RELEASE-CODE").
		SetOutTradeNo("subscription_release_order").
		SetPaymentType("wallet").
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeSubscription).
		SetPlanID(planID).
		SetSubscriptionGroupID(groupID).
		SetSubscriptionDays(30).
		SetStatus(OrderStatusPaid).
		SetPaidAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	wallet := &subscriptionHoldLifecycleSpy{}
	svc := &PaymentService{entClient: client, userRepo: wallet, groupRepo: unavailableSubscriptionGroupRepo{}}
	require.Error(t, svc.ExecuteSubscriptionFulfillment(ctx, order.ID))
	require.Equal(t, []string{holdID}, wallet.releasedHoldIDs)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusFailed, reloaded.Status)
}

func TestMalformedWalletOnlySubscriptionAlsoReleasesHeldPoints(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("subscription-malformed@example.com").
		SetPasswordHash("hash").
		SetUsername("subscription-malformed-user").
		Save(ctx)
	require.NoError(t, err)

	holdID := "00000000-0000-0000-0000-000000000301"
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(25).
		SetPayAmount(0).
		SetWalletAmount(25).
		SetWalletOnly(true).
		SetWalletHoldID(holdID).
		SetFeeRate(0).
		SetRechargeCode("SUBSCRIPTION-MALFORMED-CODE").
		SetOutTradeNo("subscription_malformed_order").
		SetPaymentType("wallet").
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeSubscription).
		SetStatus(OrderStatusPaid).
		SetPaidAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	wallet := &subscriptionHoldLifecycleSpy{}
	svc := &PaymentService{entClient: client, userRepo: wallet}
	require.ErrorContains(t, svc.ExecuteSubscriptionFulfillment(ctx, order.ID), "missing subscription info")
	require.Equal(t, []string{holdID}, wallet.releasedHoldIDs)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusFailed, reloaded.Status)
}

func TestReleasedWalletOnlySubscriptionHoldNeverFallsBackToOverdraft(t *testing.T) {
	client := newPaymentConfigServiceTestClient(t)
	wallet := &subscriptionHoldLifecycleSpy{captureErr: ErrWalletHoldState}
	svc := &PaymentService{
		entClient:       client,
		userRepo:        wallet,
		subscriptionSvc: &SubscriptionService{},
	}
	holdID := "00000000-0000-0000-0000-000000000201"
	groupID := "00000000-0000-0000-0000-000000000202"
	order := &dbent.PaymentOrder{
		ID: "00000000-0000-0000-0000-000000000203", UserID: "00000000-0000-0000-0000-000000000204",
		WalletOnly: true, WalletHoldID: &holdID, WalletAmount: 25,
	}

	err := svc.ensurePaymentSubscriptionAssigned(context.Background(), order, groupID, 30)
	require.ErrorContains(t, err, "already been released")
	require.Zero(t, wallet.debitCalls)
}
