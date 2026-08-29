package service

import (
	"context"
	"time"
)

func (a *Account) IsSchedulableForModel(requestedModel string) bool {
	return a.IsSchedulableForModelWithContext(context.Background(), requestedModel)
}

func (a *Account) IsSchedulableForModelWithContext(ctx context.Context, requestedModel string) bool {
	if a == nil || !a.IsSchedulable() {
		return false
	}
	if !a.isModelRateLimitedWithContext(ctx, requestedModel) {
		return true
	}
	return a.Platform == PlatformAntigravity && a.IsOveragesEnabled() && !a.isCreditsExhausted()
}

func (a *Account) GetRateLimitRemainingTime(requestedModel string) time.Duration {
	return a.GetRateLimitRemainingTimeWithContext(context.Background(), requestedModel)
}

func (a *Account) GetRateLimitRemainingTimeWithContext(ctx context.Context, requestedModel string) time.Duration {
	if a == nil {
		return 0
	}
	return a.GetModelRateLimitRemainingTimeWithContext(ctx, requestedModel)
}
