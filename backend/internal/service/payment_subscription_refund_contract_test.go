package service

import (
	"context"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/payment"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestLegacyPaymentRefundEntryPointsRejectSubscriptionOrders(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("subscription-refund-boundary@example.com").
		SetPasswordHash("hash").
		SetUsername("subscription-refund-boundary").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(25).
		SetPayAmount(0).
		SetWalletAmount(25).
		SetWalletOnly(true).
		SetRechargeCode("SUBSCRIPTION-REFUND-BOUNDARY").
		SetOutTradeNo("subscription_refund_boundary").
		SetPaymentType("wallet").
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeSubscription).
		SetStatus(OrderStatusCompleted).
		SetPaidAt(time.Now()).
		SetCompletedAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	plan, early, err := svc.PrepareRefund(ctx, order.ID, 25, "cancel", false, false)
	require.Nil(t, plan)
	require.Nil(t, early)
	require.Equal(t, "INVALID_ORDER_TYPE", infraerrors.Reason(err))

	_, err = client.PaymentOrder.UpdateOneID(order.ID).SetStatus(OrderStatusRefundPending).Save(ctx)
	require.NoError(t, err)
	result, err := svc.QueryAndFinalizeRefund(ctx, order.ID)
	require.Nil(t, result)
	require.Equal(t, "INVALID_ORDER_TYPE", infraerrors.Reason(err))
}
