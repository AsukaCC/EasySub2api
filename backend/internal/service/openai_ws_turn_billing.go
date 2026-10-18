package service

import (
	"context"
	"fmt"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/timezone"
)

// RefreshWSTurnBilling retains the selected group and subscription while freezing
// current prices and the user's rate plan for both admission and settlement.
func (s *OpenAIGatewayService) RefreshWSTurnBilling(ctx context.Context, key *APIKey, selected *AccountSelectionResult) (context.Context, *AccountSelectionResult, time.Time, error) {
	at := timezone.Now()
	if key == nil || key.Group == nil {
		return ctx, nil, at, fmt.Errorf("missing websocket billing identity")
	}
	group := key.Group
	if selected != nil && selected.BillingGroup != nil {
		group = selected.BillingGroup
	}
	var latest *Group
	var err error
	if s.schedulerSnapshot != nil {
		latest, err = s.schedulerSnapshot.GetGroupByIDLite(ctx, group.ID)
	} else if s.userLevelService != nil && s.userLevelService.groupRepo != nil {
		latest, err = s.userLevelService.groupRepo.GetByIDLite(ctx, group.ID)
	}
	if err != nil || latest == nil {
		return ctx, nil, at, fmt.Errorf("websocket billing group unavailable")
	}
	if latest.ID != group.ID || latest.Platform != group.Platform || latest.SubscriptionType != group.SubscriptionType || latest.Status != StatusActive {
		return ctx, nil, at, fmt.Errorf("websocket billing group changed; reconnect required")
	}
	group = latest
	snapshot := AccountSelectionResult{}
	if selected != nil {
		snapshot = *selected
	}
	snapshot.BillingGroup = group
	if plan, ok := s.ResolveUserRatePlan(ctx, key.UserID, group, at); ok {
		snapshot.RatePlan = plan
	} else {
		return ctx, nil, at, fmt.Errorf("websocket billing rate plan unavailable")
	}
	ctx = contextWithUserRatePlan(ctx, group, snapshot.RatePlan)
	ctx = context.WithValue(ctx, openAIPricingAtCtxKey{}, at)
	gateGroupID := group.ID
	if selected != nil && selected.profitGate != nil {
		gateGroupID = selected.profitGate.groupID
	}
	snapshot.profitGate = nil
	if _, suppressed := ctx.Value(openAIProfitControlSuppressCtxKey{}).(struct{}); !suppressed {
		snapshot.profitGate = s.resolveOpenAIProfitControlGate(ctx, &gateGroupID)
	}
	ctx = ContextWithSelectionProfitGate(ctx, &snapshot)
	return ctx, &snapshot, at, nil
}
