package service

import (
	"context"
	"time"
)

// OAuthTokenCache stores short-lived access tokens and coordinates refreshes.
type OAuthTokenCache interface {
	GetAccessToken(ctx context.Context, cacheKey string) (string, error)
	SetAccessToken(ctx context.Context, cacheKey string, token string, ttl time.Duration) error
	DeleteAccessToken(ctx context.Context, cacheKey string) error

	AcquireRefreshLock(ctx context.Context, cacheKey string, ttl time.Duration) (bool, error)
	ReleaseRefreshLock(ctx context.Context, cacheKey string) error
}

// GeminiTokenCache is kept as a compatibility name for Gemini and Antigravity
// providers while all OAuth platforms share the same cache implementation.
type GeminiTokenCache = OAuthTokenCache
