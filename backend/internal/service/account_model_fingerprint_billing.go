package service

import (
	"context"
	"errors"
	"time"
)

// Fingerprint probes use the same pricing and atomic settlement as API requests.
type modelFingerprintBilling struct {
	gateway  *OpenAIGatewayService
	keyQuota APIKeyQuotaUpdater
}

func (b *modelFingerprintBilling) Prepare(ctx context.Context, account *Account, key *APIKey) (*OpenAIRecordUsageInput, error) {
	g := b.gateway
	if g == nil || g.usageBillingRepo == nil || g.billingCacheService == nil || g.userRepo == nil {
		return nil, errors.New("model fingerprint billing unavailable")
	}
	if account == nil || key == nil || key.GroupID == nil || key.Group == nil {
		return nil, errors.New("model fingerprint billing requires an API key and group")
	}
	user, err := g.userRepo.GetByID(ctx, key.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive() {
		return nil, errors.New("model fingerprint billing user unavailable")
	}
	var subscription *UserSubscription
	if key.Group.IsSubscriptionType() {
		if g.userSubRepo == nil {
			return nil, errors.New("model fingerprint subscription lookup unavailable")
		}
		subscription, err = g.userSubRepo.GetActiveByUserIDAndGroupID(ctx, user.ID, *key.GroupID)
		if err != nil {
			return nil, err
		}
		if subscription == nil || !subscription.IsActive() {
			return nil, errors.New("model fingerprint subscription unavailable")
		}
	}
	if err := g.billingCacheService.CheckBillingEligibility(ctx, user, key, key.Group, subscription, account.Platform); err != nil {
		return nil, err
	}
	return &OpenAIRecordUsageInput{
		APIKey: key, User: user, Account: account, Subscription: subscription,
		APIKeyService: b.keyQuota, QuotaPlatform: account.Platform,
		RequestTypeOverride: RequestTypeTest,
	}, nil
}

func (b *modelFingerprintBilling) Record(ctx context.Context, log *UsageLog, prepared *OpenAIRecordUsageInput) error {
	if prepared == nil || log == nil {
		return errors.New("model fingerprint billing snapshot missing")
	}
	input := *prepared
	input.Result = &OpenAIForwardResult{
		RequestID: log.RequestID, Model: log.Model,
		// The collector stores exclusive token buckets; the OpenAI billing entry
		// expects total input and subtracts cache reads/writes itself.
		Usage: OpenAIUsage{
			InputTokens:  log.InputTokens + log.CacheReadTokens + log.CacheCreationTokens,
			OutputTokens: log.OutputTokens, CacheReadInputTokens: log.CacheReadTokens,
			CacheCreationInputTokens: log.CacheCreationTokens,
		},
		Stream: log.Stream, FirstTokenMs: log.FirstTokenMs,
		ReasoningEffort: log.ReasoningEffort, RequestedReasoningEffort: log.RequestedReasoningEffort,
	}
	if log.DurationMs != nil {
		input.Result.Duration = time.Duration(*log.DurationMs) * time.Millisecond
	}
	if log.UpstreamModel != nil {
		input.Result.UpstreamModel = *log.UpstreamModel
	}
	if log.UpstreamResponseModel != nil {
		input.Result.UpstreamResponseModel = *log.UpstreamResponseModel
	}
	if log.InboundEndpoint != nil {
		input.InboundEndpoint = *log.InboundEndpoint
	}
	if log.UpstreamEndpoint != nil {
		input.UpstreamEndpoint = *log.UpstreamEndpoint
	}
	if log.SessionID != nil {
		input.SessionID = *log.SessionID
	}
	input.PricingAt = log.CreatedAt
	return b.gateway.RecordUsage(ctx, &input)
}
