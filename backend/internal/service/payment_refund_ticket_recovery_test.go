//go:build integration

package service

import (
	"context"
	"os"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	_ "github.com/AsukaCC/EasySub2api/ent/runtime"
	"github.com/AsukaCC/EasySub2api/internal/payment"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestProviderConfirmedRefundSettlementFailureRemainsPending(t *testing.T) {
	ctx := context.Background()
	client := newRefundRecoveryTestClient(t)
	user := createRefundRecoveryUser(t, ctx, client, "settlement-pending")
	order := createRefundRecoveryOrder(t, ctx, client, user, OrderStatusRefunding, "settlement-pending")
	refund := createRefundRecoveryRecord(t, ctx, client, order, "", RefundStatusSubmitting, "settlement-pending")

	_, err := client.ExecContext(ctx, `ALTER TABLE payment_orders ADD CONSTRAINT refund_recovery_test_block_settlement CHECK (refunded_gateway_amount = 0)`)
	require.NoError(t, err)

	result, err := (&PaymentService{entClient: client}).settleProviderConfirmedPaymentRefund(ctx, refund.ID, "rf_upstream_success")
	require.NoError(t, err)
	require.Equal(t, RefundStatusPending, result.Status)
	require.Equal(t, "REFUND_SETTLEMENT_RETRY", result.ErrorCode)
	require.NotNil(t, result.ProviderRefundID)
	require.Equal(t, "rf_upstream_success", *result.ProviderRefundID)

	order, err = client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefundPending, order.Status)
}

func TestAdminResolveRefundTicketRepairsSucceededOrder(t *testing.T) {
	ctx := context.Background()
	client := newRefundRecoveryTestClient(t)
	user := createRefundRecoveryUser(t, ctx, client, "ticket-repair")
	order := createRefundRecoveryOrder(t, ctx, client, user, OrderStatusRefunding, "ticket-repair")
	ticket, err := client.SupportTicket.Create().
		SetUserID(user.ID).
		SetCategory(SupportTicketCategoryRefund).
		SetStatus(SupportTicketStatusInProgress).
		SetOrigin(SupportTicketOriginUser).
		SetTitle("refund recovery").
		SetOrderID(order.ID).
		SetRefundDecision(SupportTicketRefundApproved).
		Save(ctx)
	require.NoError(t, err)
	refund := createRefundRecoveryRecord(t, ctx, client, order, ticket.ID, RefundStatusSucceeded, "ticket-repair")

	paymentService := &PaymentService{entClient: client}
	detail, err := NewSupportTicketService(client, paymentService, nil).
		AdminSetStatus(ctx, ticket.ID, "admin-id", SupportTicketStatusResolved, "")
	require.NoError(t, err)
	require.Equal(t, SupportTicketStatusResolved, detail.Ticket.Status)
	require.NotNil(t, detail.Ticket.RefundID)
	require.Equal(t, refund.ID, *detail.Ticket.RefundID)

	order, err = client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefunded, order.Status)
	require.Equal(t, 100.0, order.RefundedPrincipalAmount)
	require.Equal(t, 3.0, order.RefundedFeeAmount)
	require.Equal(t, 100.0, order.RefundedGatewayAmount)
}

func TestAdminResolveRefundTicketReconcilesProviderSuccess(t *testing.T) {
	ctx := context.Background()
	client := newRefundRecoveryTestClient(t)
	user := createRefundRecoveryUser(t, ctx, client, "ticket-provider-success")
	instance, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeStripe).
		SetName("refund recovery provider").
		SetConfig("{}").
		SetSupportedTypes(string(payment.TypeStripe)).
		SetEnabled(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)
	order := createRefundRecoveryOrder(t, ctx, client, user, OrderStatusRefunding, "ticket-provider-success")
	order, err = client.PaymentOrder.UpdateOneID(order.ID).
		SetProviderInstanceID(instance.ID).
		SetProviderKey(payment.TypeStripe).
		Save(ctx)
	require.NoError(t, err)
	ticket, err := client.SupportTicket.Create().
		SetUserID(user.ID).
		SetCategory(SupportTicketCategoryRefund).
		SetStatus(SupportTicketStatusInProgress).
		SetOrigin(SupportTicketOriginUser).
		SetTitle("provider refund recovery").
		SetOrderID(order.ID).
		SetRefundDecision(SupportTicketRefundApproved).
		Save(ctx)
	require.NoError(t, err)
	refund := createRefundRecoveryRecord(t, ctx, client, order, ticket.ID, RefundStatusSubmitting, "ticket-provider-success")

	originalFactory := createPaymentProviderFromInstance
	createPaymentProviderFromInstance = func(string, string, map[string]string) (payment.Provider, error) {
		return refundRecoveryProvider{status: payment.ProviderStatusSuccess}, nil
	}
	t.Cleanup(func() { createPaymentProviderFromInstance = originalFactory })

	paymentService := &PaymentService{entClient: client, loadBalancer: refundRecoveryLoadBalancer{}}
	detail, err := NewSupportTicketService(client, paymentService, nil).
		AdminSetStatus(ctx, ticket.ID, "admin-id", SupportTicketStatusResolved, "")
	require.NoError(t, err)
	require.Equal(t, SupportTicketStatusResolved, detail.Ticket.Status)

	refund, err = client.PaymentRefund.Get(ctx, refund.ID)
	require.NoError(t, err)
	require.Equal(t, RefundStatusSucceeded, refund.Status)
	require.NotNil(t, refund.ProviderRefundID)
	require.Equal(t, "rf_provider_success", *refund.ProviderRefundID)
	order, err = client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefunded, order.Status)
}

type refundRecoveryLoadBalancer struct{}

func (refundRecoveryLoadBalancer) GetInstanceConfig(context.Context, string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (refundRecoveryLoadBalancer) SelectInstance(context.Context, string, payment.PaymentType, payment.Strategy, float64) (*payment.InstanceSelection, error) {
	return nil, nil
}

type refundRecoveryProvider struct {
	status string
}

func (refundRecoveryProvider) Name() string { return "refund-recovery" }

func (refundRecoveryProvider) ProviderKey() string { return payment.TypeStripe }

func (refundRecoveryProvider) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeStripe}
}

func (refundRecoveryProvider) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	return nil, nil
}

func (refundRecoveryProvider) QueryOrder(context.Context, string) (*payment.QueryOrderResponse, error) {
	return nil, nil
}

func (refundRecoveryProvider) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	return nil, nil
}

func (p refundRecoveryProvider) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return &payment.RefundResponse{RefundID: "rf_provider_success", Status: p.status}, nil
}

func (p refundRecoveryProvider) QueryRefund(context.Context, payment.RefundQueryRequest) (*payment.RefundResponse, error) {
	return &payment.RefundResponse{RefundID: "rf_provider_success", Status: p.status}, nil
}

func newRefundRecoveryTestClient(t *testing.T) *dbent.Client {
	t.Helper()
	dsn := os.Getenv("REFUND_RECOVERY_TEST_DSN")
	require.NotEmpty(t, dsn)
	client, err := dbent.Open(dialect.Postgres, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Schema.Create(context.Background()))
	return client
}

func createRefundRecoveryUser(t *testing.T, ctx context.Context, client *dbent.Client, suffix string) *dbent.User {
	t.Helper()
	user, err := client.User.Create().
		SetEmail(suffix + "@example.com").
		SetPasswordHash("hash").
		SetUsername(suffix).
		Save(ctx)
	require.NoError(t, err)
	return user
}

func createRefundRecoveryOrder(t *testing.T, ctx context.Context, client *dbent.Client, user *dbent.User, status, suffix string) *dbent.PaymentOrder {
	t.Helper()
	now := time.Now()
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(103).
		SetFeeRate(3).
		SetGatewayBaseAmount(100).
		SetPrincipalAmount(100).
		SetFeeAmount(3).
		SetCurrency("CNY").
		SetBasePoints(100).
		SetCreditedPoints(100).
		SetRechargeCode("RECOVERY-" + suffix).
		SetOutTradeNo("sub2_" + suffix).
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("pi_" + suffix).
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(status).
		SetExpiresAt(now.Add(time.Hour)).
		SetPaidAt(now).
		SetCompletedAt(now).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)
	return order
}

func createRefundRecoveryRecord(t *testing.T, ctx context.Context, client *dbent.Client, order *dbent.PaymentOrder, ticketID, status, suffix string) *dbent.PaymentRefund {
	t.Helper()
	builder := client.PaymentRefund.Create().
		SetOrderID(order.ID).
		SetUserID(order.UserID).
		SetSource(RefundSourceSelfService).
		SetStatus(status).
		SetIdempotencyKey("recovery-" + suffix).
		SetRequestFingerprint("fingerprint-" + suffix).
		SetProviderRequestID("request-" + suffix).
		SetCurrency("CNY").
		SetRequestedPrincipalAmount(100).
		SetPrincipalAmount(100).
		SetFeeAmount(3).
		SetRefundFeeRate(3).
		SetRefundFeeAmount(3).
		SetGatewayAmount(100).
		SetBasePoints(100).
		SetTargetPrincipalAmount(100).
		SetTargetFeeAmount(3).
		SetTargetRefundFeeAmount(3).
		SetTargetBasePoints(100).
		SetReason("refund recovery")
	if ticketID != "" {
		builder.SetTicketID(ticketID)
	}
	if status == RefundStatusSucceeded {
		builder.SetProviderRefundID("rf_" + suffix).SetSettledAt(time.Now())
	}
	refund, err := builder.Save(ctx)
	require.NoError(t, err)
	return refund
}
