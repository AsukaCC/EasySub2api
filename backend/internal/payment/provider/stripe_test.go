//go:build unit

package provider

import (
	"bytes"
	"context"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/payment"
	"github.com/stretchr/testify/require"
	stripe "github.com/stripe/stripe-go/v85"
)

type stripeRefundBackend struct {
	params []*stripe.RefundCreateParams
	status stripe.RefundStatus
}

func (b *stripeRefundBackend) Call(_ string, _ string, _ string, params stripe.ParamsContainer, v stripe.LastResponseSetter) error {
	b.params = append(b.params, params.(*stripe.RefundCreateParams))
	refund := v.(*stripe.Refund)
	refund.ID = "re_123"
	refund.Status = b.status
	if refund.Status == "" {
		refund.Status = stripe.RefundStatusSucceeded
	}
	return nil
}

func TestStripeRefundMapsDefinitiveProviderFailure(t *testing.T) {
	backend := &stripeRefundBackend{status: stripe.RefundStatusFailed}
	client := stripe.NewClient("sk_test", stripe.WithBackends(&stripe.Backends{API: backend}))
	provider := &Stripe{
		config: map[string]string{"currency": "CNY"}, initialized: true, sc: client,
	}

	result, err := provider.Refund(context.Background(), payment.RefundRequest{
		TradeNo: "pi_123", OrderID: "sub2_order_456", RequestID: "refund-failed", Amount: "12.34",
	})
	require.NoError(t, err)
	require.Equal(t, payment.ProviderStatusFailed, result.Status)
}

func (*stripeRefundBackend) CallStreaming(string, string, string, stripe.ParamsContainer, stripe.StreamingLastResponseSetter) error {
	return nil
}

func (*stripeRefundBackend) CallRaw(string, string, string, []byte, *stripe.Params, stripe.LastResponseSetter) error {
	return nil
}

func (*stripeRefundBackend) CallMultipart(string, string, string, string, *bytes.Buffer, *stripe.Params, stripe.LastResponseSetter) error {
	return nil
}

func (*stripeRefundBackend) SetMaxNetworkRetries(int64) {}

func TestStripeRefundUsesStableRequestIDIdempotencyKey(t *testing.T) {
	backend := &stripeRefundBackend{}
	client := stripe.NewClient("sk_test", stripe.WithBackends(&stripe.Backends{API: backend}))
	provider := &Stripe{
		config:      map[string]string{"currency": "CNY"},
		initialized: true,
		sc:          client,
	}

	refund := func(requestID, amount string) {
		_, err := provider.Refund(context.Background(), payment.RefundRequest{
			TradeNo:   "pi_123",
			OrderID:   "sub2_order_456",
			RequestID: requestID,
			Amount:    amount,
		})
		require.NoError(t, err)
	}

	refund("refund-ticket-1", "12.34")
	refund("refund-ticket-1", "12.34")
	refund("refund-ticket-2", "12.34")

	require.Len(t, backend.params, 3)
	require.Equal(t, int64(1234), *backend.params[0].Amount)
	require.Equal(t, "re-refund-ticket-1", *backend.params[0].IdempotencyKey)
	require.Equal(t, backend.params[0].IdempotencyKey, backend.params[1].IdempotencyKey)
	require.Equal(t, int64(1234), *backend.params[2].Amount)
	require.Equal(t, "re-refund-ticket-2", *backend.params[2].IdempotencyKey)
	require.NotEqual(t, *backend.params[0].IdempotencyKey, *backend.params[2].IdempotencyKey)
}

func TestStripeQueryRefundWithoutProviderIDReplaysStableCreate(t *testing.T) {
	backend := &stripeRefundBackend{}
	client := stripe.NewClient("sk_test", stripe.WithBackends(&stripe.Backends{API: backend}))
	provider := &Stripe{
		config:      map[string]string{"currency": "CNY"},
		initialized: true,
		sc:          client,
	}

	result, err := provider.QueryRefund(context.Background(), payment.RefundQueryRequest{
		TradeNo: "pi_123", OrderID: "sub2_order_456",
		RequestID: "refund-attempt-2", Amount: "12.34",
	})
	require.NoError(t, err)
	require.Equal(t, "re_123", result.RefundID)
	require.Equal(t, payment.ProviderStatusSuccess, result.Status)
	require.Len(t, backend.params, 1)
	require.Equal(t, "re-refund-attempt-2", *backend.params[0].IdempotencyKey)
}
