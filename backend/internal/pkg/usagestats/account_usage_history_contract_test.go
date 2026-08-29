package usagestats

import (
	"encoding/json"
	"testing"
)

func TestAccountUsageHistorySeparatesUSDAndPoints(t *testing.T) {
	body, err := json.Marshal(AccountUsageHistory{
		Cost: 1.25, AccountCost: 1.5, ActualCost: 2.75, UserCost: 2.75,
	})
	if err != nil {
		t.Fatalf("marshal account usage history: %v", err)
	}
	var payload struct {
		Cost        float64 `json:"cost"`
		AccountCost float64 `json:"account_cost"`
		ActualCost  float64 `json:"actual_cost"`
		UserCost    float64 `json:"user_cost"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode account usage history: %v", err)
	}
	if payload.Cost != 1.25 || payload.AccountCost != 1.5 {
		t.Fatalf("USD fields = cost:%v account_cost:%v", payload.Cost, payload.AccountCost)
	}
	if payload.ActualCost != 2.75 || payload.UserCost != 2.75 {
		t.Fatalf("point fields = actual_cost:%v user_cost:%v", payload.ActualCost, payload.UserCost)
	}
}
