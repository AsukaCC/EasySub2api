package service

import (
	"context"
	"errors"
	"testing"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/stretchr/testify/require"
)

func TestReconcilePaymentRefundBatchConvergesByRefundID(t *testing.T) {
	refunds := []*dbent.PaymentRefund{
		{ID: "refund-success", OrderID: "shared-order"},
		{ID: "refund-failed", OrderID: "shared-order"},
		{ID: "refund-pending", OrderID: "shared-order"},
		{ID: "refund-error", OrderID: "shared-order"},
	}
	seen := make([]string, 0, len(refunds))
	attempted, resolved, err := reconcilePaymentRefundBatch(context.Background(), refunds,
		func(_ context.Context, refund *dbent.PaymentRefund) (*PaymentRefundResponse, error) {
			seen = append(seen, refund.ID)
			switch refund.ID {
			case "refund-success":
				return &PaymentRefundResponse{ID: refund.ID, Status: RefundStatusSucceeded}, nil
			case "refund-failed":
				return &PaymentRefundResponse{ID: refund.ID, Status: RefundStatusFailed}, nil
			case "refund-pending":
				return &PaymentRefundResponse{ID: refund.ID, Status: RefundStatusPending}, nil
			default:
				return nil, errors.New("provider query unavailable")
			}
		})

	require.Equal(t, 4, attempted)
	require.Equal(t, 2, resolved)
	require.ErrorContains(t, err, "reconcile refund refund-error")
	require.Equal(t, []string{"refund-success", "refund-failed", "refund-pending", "refund-error"}, seen)
}

func TestReconcilePaymentRefundBatchStopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	attempted, resolved, err := reconcilePaymentRefundBatch(ctx,
		[]*dbent.PaymentRefund{{ID: "refund-1"}},
		func(context.Context, *dbent.PaymentRefund) (*PaymentRefundResponse, error) {
			t.Fatal("reconciler must not run after cancellation")
			return nil, nil
		})

	require.Zero(t, attempted)
	require.Zero(t, resolved)
	require.ErrorIs(t, err, context.Canceled)
}
