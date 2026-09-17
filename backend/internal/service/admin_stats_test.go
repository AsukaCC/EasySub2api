//go:build unit

package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdminStatsGroupCountsAndHistoricalUsage(t *testing.T) {
	client := newAdminServiceAuthIdentityBindingTestClient(t)
	ctx := context.Background()
	svc := &adminServiceImpl{entClient: client}
	g, err := client.Group.Create().SetName("stats").Save(ctx)
	require.NoError(t, err)
	other, err := client.Group.Create().SetName("other").Save(ctx)
	require.NoError(t, err)
	empty, err := svc.GetGroupStats(ctx, g.ID)
	require.NoError(t, err)
	require.Equal(t, &AdminGroupStats{}, empty)
	u, err := client.User.Create().SetEmail("stats@example.test").SetPasswordHash("hash").Save(ctx)
	require.NoError(t, err)
	a, err := client.Account.Create().SetName("stats").SetPlatform(PlatformOpenAI).SetType(AccountTypeAPIKey).Save(ctx)
	require.NoError(t, err)
	for i, status := range []string{StatusActive, StatusDisabled, StatusActive} {
		key, err := client.APIKey.Create().SetUserID(u.ID).SetGroupID(g.ID).
			SetName(fmt.Sprint(i)).SetKey(fmt.Sprintf("stats-key-%d", i)).SetStatus(status).Save(ctx)
		require.NoError(t, err)
		_, err = client.UsageLog.Create().SetUserID(u.ID).SetAPIKeyID(key.ID).SetAccountID(a.ID).
			SetGroupID(g.ID).SetRequestID(fmt.Sprint(i)).SetModel("test").SetTotalCost(1.25).Save(ctx)
		require.NoError(t, err)
		if i == 2 {
			require.NoError(t, client.APIKey.UpdateOneID(key.ID).SetDeletedAt(time.Now()).Exec(ctx))
		}
	}
	_, err = client.APIKey.Create().SetUserID(u.ID).SetGroupID(other.ID).SetName("other").SetKey("other-key").Save(ctx)
	require.NoError(t, err)
	stats, err := svc.GetGroupStats(ctx, g.ID)
	require.NoError(t, err)
	require.Equal(t, &AdminGroupStats{TotalAPIKeys: 2, ActiveAPIKeys: 1, TotalRequests: 3, TotalCost: 3.75}, stats)
	require.NoError(t, client.Group.UpdateOneID(g.ID).SetDeletedAt(time.Now()).Exec(ctx))
	_, err = svc.GetGroupStats(ctx, g.ID)
	require.ErrorIs(t, err, ErrGroupNotFound)
	_, err = svc.GetGroupStats(ctx, "missing")
	require.ErrorIs(t, err, ErrGroupNotFound)
}

func TestAdminStatsProxyDoesNotInventHistoricalMetrics(t *testing.T) {
	client := newAdminServiceAuthIdentityBindingTestClient(t)
	ctx := context.Background()
	svc := &adminServiceImpl{entClient: client}
	p, err := client.Proxy.Create().SetName("stats").SetProtocol("http").SetHost("localhost").SetPort(8080).Save(ctx)
	require.NoError(t, err)
	for i, status := range []string{StatusActive, StatusDisabled, StatusActive} {
		a, err := client.Account.Create().SetName(fmt.Sprint(i)).SetPlatform(PlatformOpenAI).
			SetType(AccountTypeAPIKey).SetProxyID(p.ID).SetStatus(status).Save(ctx)
		require.NoError(t, err)
		if i == 2 {
			require.NoError(t, client.Account.UpdateOneID(a.ID).SetDeletedAt(time.Now()).Exec(ctx))
		}
	}
	stats, err := svc.GetProxyStats(ctx, p.ID)
	require.NoError(t, err)
	require.Equal(t, &AdminProxyStats{TotalAccounts: 2, ActiveAccounts: 1}, stats)
	require.NoError(t, client.Proxy.UpdateOneID(p.ID).SetDeletedAt(time.Now()).Exec(ctx))
	_, err = svc.GetProxyStats(ctx, p.ID)
	require.ErrorIs(t, err, ErrProxyNotFound)
	_, err = svc.GetProxyStats(ctx, "missing")
	require.ErrorIs(t, err, ErrProxyNotFound)
}

func TestAdminStatsRedeemExpiryAndValueUnits(t *testing.T) {
	client := newAdminServiceAuthIdentityBindingTestClient(t)
	ctx := context.Background()
	svc := &adminServiceImpl{entClient: client}
	empty, err := svc.GetRedeemStats(ctx)
	require.NoError(t, err)
	require.Zero(t, empty.TotalCodes)
	require.Zero(t, empty.TotalValueDistributed)
	require.NotNil(t, empty.ByType)
	past, future := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	for i, row := range []struct {
		kind, status string
		value        float64
		expiry       *time.Time
	}{
		{RedeemTypeBalance, StatusUnused, 10, nil},
		{RedeemTypeBalance, StatusUnused, 20, &future},
		{RedeemTypeBalance, StatusUnused, 30, &past},
		{RedeemTypeBalance, StatusExpired, 40, nil},
		{RedeemTypeBalance, StatusUsed, 50, &past},
		{RedeemTypeConcurrency, StatusUsed, 100, nil},
		{RedeemTypeSubscription, StatusUsed, 30, nil},
		{RedeemTypeInvitation, StatusUnused, 0, nil},
	} {
		_, err = client.RedeemCode.Create().SetCode(fmt.Sprint(i)).SetType(row.kind).
			SetStatus(row.status).SetValue(row.value).SetNillableExpiresAt(row.expiry).Save(ctx)
		require.NoError(t, err)
	}
	stats, err := svc.GetRedeemStats(ctx)
	require.NoError(t, err)
	require.Equal(t, 8, stats.TotalCodes)
	require.Equal(t, 3, stats.ActiveCodes)
	require.Equal(t, 3, stats.UsedCodes)
	require.Equal(t, 2, stats.ExpiredCodes)
	require.Equal(t, 50.0, stats.TotalValueDistributed)
	require.Equal(t, 5, stats.ByType[RedeemTypeBalance])
	require.Equal(t, 1, stats.ByType[RedeemTypeSubscription])
}

func TestAdminStatsDatabaseUnavailable(t *testing.T) {
	svc := &adminServiceImpl{}
	_, err := svc.GetGroupStats(context.Background(), "group")
	require.Error(t, err)
	_, err = svc.GetProxyStats(context.Background(), "proxy")
	require.Error(t, err)
	_, err = svc.GetRedeemStats(context.Background())
	require.Error(t, err)
}
