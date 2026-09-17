package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/ent/account"
	"github.com/AsukaCC/EasySub2api/ent/apikey"
	"github.com/AsukaCC/EasySub2api/ent/group"
	"github.com/AsukaCC/EasySub2api/ent/proxy"
	"github.com/AsukaCC/EasySub2api/ent/redeemcode"
	"github.com/AsukaCC/EasySub2api/ent/usagelog"
)

type AdminGroupStats struct {
	TotalAPIKeys  int     `json:"total_api_keys"`
	ActiveAPIKeys int     `json:"active_api_keys"`
	TotalRequests int     `json:"total_requests"`
	TotalCost     float64 `json:"total_cost"`
}

type AdminProxyStats struct {
	TotalAccounts  int      `json:"total_accounts"`
	ActiveAccounts int      `json:"active_accounts"`
	TotalRequests  *int64   `json:"total_requests"`
	SuccessRate    *float64 `json:"success_rate"`
	AverageLatency *float64 `json:"average_latency"`
}

type AdminRedeemStats struct {
	TotalCodes            int            `json:"total_codes"`
	ActiveCodes           int            `json:"active_codes"`
	UsedCodes             int            `json:"used_codes"`
	ExpiredCodes          int            `json:"expired_codes"`
	TotalValueDistributed float64        `json:"total_value_distributed"`
	ByType                map[string]int `json:"by_type"`
}

func (s *adminServiceImpl) GetGroupStats(ctx context.Context, id string) (*AdminGroupStats, error) {
	if s.entClient == nil {
		return nil, fmt.Errorf("group statistics database unavailable")
	}
	exists, err := s.entClient.Group.Query().Where(group.IDEQ(id), group.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrGroupNotFound
	}
	stats := &AdminGroupStats{}
	var keys []struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	err = s.entClient.APIKey.Query().Where(apikey.GroupIDEQ(id), apikey.DeletedAtIsNil()).
		GroupBy(apikey.FieldStatus).Aggregate(dbent.As(dbent.Count(), "count")).Scan(ctx, &keys)
	if err != nil {
		return nil, err
	}
	for _, row := range keys {
		stats.TotalAPIKeys += row.Count
		if row.Status == StatusActive {
			stats.ActiveAPIKeys += row.Count
		}
	}
	var usage []struct {
		Requests int             `json:"requests"`
		Cost     sql.NullFloat64 `json:"cost"`
	}
	err = s.entClient.UsageLog.Query().Where(usagelog.GroupIDEQ(id)).Aggregate(
		dbent.As(dbent.Count(), "requests"),
		dbent.As(dbent.Sum(usagelog.FieldTotalCost), "cost"),
	).Scan(ctx, &usage)
	if err != nil {
		return nil, err
	}
	if len(usage) > 0 {
		stats.TotalRequests = usage[0].Requests
		stats.TotalCost = usage[0].Cost.Float64
	}
	return stats, nil
}

func (s *adminServiceImpl) GetProxyStats(ctx context.Context, id string) (*AdminProxyStats, error) {
	if s.entClient == nil {
		return nil, fmt.Errorf("proxy statistics database unavailable")
	}
	exists, err := s.entClient.Proxy.Query().Where(proxy.IDEQ(id), proxy.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrProxyNotFound
	}
	// Usage logs do not snapshot proxy IDs. Current bindings cannot establish
	// historical request attribution, so request metrics remain explicitly null.
	stats := &AdminProxyStats{}
	var accounts []struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	err = s.entClient.Account.Query().Where(account.ProxyIDEQ(id), account.DeletedAtIsNil()).
		GroupBy(account.FieldStatus).Aggregate(dbent.As(dbent.Count(), "count")).Scan(ctx, &accounts)
	if err != nil {
		return nil, err
	}
	for _, row := range accounts {
		stats.TotalAccounts += row.Count
		if row.Status == StatusActive {
			stats.ActiveAccounts += row.Count
		}
	}
	return stats, nil
}

func (s *adminServiceImpl) GetRedeemStats(ctx context.Context) (*AdminRedeemStats, error) {
	if s.entClient == nil {
		return nil, fmt.Errorf("redeem statistics database unavailable")
	}
	// All counts must see the same rows while codes are redeemed or expired.
	tx, err := s.entClient.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	client := tx.Client()
	stats := &AdminRedeemStats{ByType: map[string]int{
		RedeemTypeBalance: 0, RedeemTypeConcurrency: 0, RedeemTypeSubscription: 0, RedeemTypeInvitation: 0, "trial": 0,
	}}
	var counts []struct {
		Type   string `json:"type"`
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	err = client.RedeemCode.Query().GroupBy(redeemcode.FieldType, redeemcode.FieldStatus).
		Aggregate(dbent.As(dbent.Count(), "count")).Scan(ctx, &counts)
	if err != nil {
		return nil, err
	}
	for _, row := range counts {
		stats.TotalCodes += row.Count
		stats.ByType[row.Type] += row.Count
		switch row.Status {
		case StatusUnused:
			stats.ActiveCodes += row.Count
		case StatusUsed:
			stats.UsedCodes += row.Count
		case StatusExpired:
			stats.ExpiredCodes += row.Count
		}
	}
	// Expiration can be effective before a background job persists the status.
	expiredUnused, err := client.RedeemCode.Query().Where(
		redeemcode.StatusEQ(StatusUnused), redeemcode.ExpiresAtLTE(time.Now()),
	).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ActiveCodes -= expiredUnused
	stats.ExpiredCodes += expiredUnused
	var values []struct {
		Value sql.NullFloat64 `json:"value"`
	}
	// Only redeemed balance has a monetary value; do not sum concurrency or days.
	err = client.RedeemCode.Query().Where(
		redeemcode.TypeEQ(RedeemTypeBalance), redeemcode.StatusEQ(StatusUsed), redeemcode.ValueGT(0),
	).Aggregate(dbent.As(dbent.Sum(redeemcode.FieldValue), "value")).Scan(ctx, &values)
	if err != nil {
		return nil, err
	}
	if len(values) > 0 {
		stats.TotalValueDistributed = values[0].Value.Float64
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return stats, nil
}
