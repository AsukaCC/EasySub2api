package payment

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func newLoadBalancerQueryCapture(t *testing.T) (*DefaultLoadBalancer, sqlmock.Sqlmock, *[]string) {
	t.Helper()
	queries := make([]string, 0, 2)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(
		func(_, actual string) error {
			queries = append(queries, strings.Join(strings.Fields(actual), " "))
			return nil
		},
	)))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	return NewDefaultLoadBalancer(client, nil), mock, &queries
}

func assertRechargeGatewayUsageQuery(t *testing.T, query string) {
	t.Helper()
	require.Contains(t, query, `SUM("payment_orders"."pay_amount")`)
	require.Contains(t, query, `"payment_orders"."order_type" =`)
	require.Contains(t, query, `"payment_orders"."payment_type" IN`)
}

func TestLoadBalancerDailyUsageScopesMultiChannelInstanceToRechargeMethod(t *testing.T) {
	for _, testCase := range []struct {
		name, method, alias string
	}{
		{name: "alipay", method: TypeAlipay, alias: TypeAlipayDirect},
		{name: "wxpay", method: TypeWxpay, alias: TypeWxpayDirect},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			lb, mock, queries := newLoadBalancerQueryCapture(t)
			mock.ExpectQuery("ignored").
				WithArgs("0198-provider", OrderTypeBalance, testCase.method, testCase.alias, OrderStatusPending, sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"provider_instance_id", "sum"}).
					AddRow("0198-provider", 42.0))

			candidates := lb.attachDailyUsage(context.Background(), []*dbent.PaymentProviderInstance{{ID: "0198-provider"}}, testCase.method)
			require.Len(t, candidates, 1)
			require.Equal(t, 42.0, candidates[0].dailyUsed)
			require.Len(t, *queries, 1)
			assertRechargeGatewayUsageQuery(t, (*queries)[0])
			require.Contains(t, (*queries)[0], `"payment_orders"."status" =`)
			require.Contains(t, (*queries)[0], `"payment_orders"."created_at" >=`)
			require.Contains(t, (*queries)[0], `"payment_orders"."paid_at" >=`)
			require.Contains(t, (*queries)[0], " OR ")
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetInstanceDailyAmountUsesRechargeGatewayAmountOnly(t *testing.T) {
	lb, mock, queries := newLoadBalancerQueryCapture(t)
	mock.ExpectQuery("ignored").
		WithArgs("0198-provider", OrderTypeBalance, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(42.0))

	amount, err := lb.GetInstanceDailyAmount(context.Background(), "0198-provider")
	require.NoError(t, err)
	require.Equal(t, 42.0, amount)
	require.Len(t, *queries, 1)
	require.Contains(t, (*queries)[0], `SUM("payment_orders"."pay_amount")`)
	require.Contains(t, (*queries)[0], `"payment_orders"."order_type" =`)
	require.Contains(t, (*queries)[0], `"payment_orders"."paid_at" >=`)
	require.NotContains(t, (*queries)[0], `"payment_orders"."payment_type"`)
	require.NotContains(t, (*queries)[0], `"payment_orders"."status"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFilterByLimitsUsesExactCNYCentBoundaries(t *testing.T) {
	limits, err := json.Marshal(InstanceLimits{
		TypeAlipay: {SingleMin: 0.1, SingleMax: 0.3, DailyLimit: 0.3},
	})
	require.NoError(t, err)
	candidate := instanceCandidate{
		inst:      &dbent.PaymentProviderInstance{ID: "0198-provider", ProviderKey: TypeAlipay, Limits: string(limits)},
		dailyUsed: 0.1,
	}

	require.Len(t, filterByLimits([]instanceCandidate{candidate}, TypeAlipay, 0.2), 1,
		"0.10 used + 0.20 order must equal, not exceed, a 0.30 CNY limit")
	require.Empty(t, filterByLimits([]instanceCandidate{candidate}, TypeAlipay, 0.21),
		"one cent above the CNY limit must be rejected")
}
