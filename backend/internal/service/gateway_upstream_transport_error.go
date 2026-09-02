package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const gatewayTransportErrorTempUnschedDuration = 10 * time.Minute

var gatewayTransportFailoverBody = []byte(`{"type":"error","error":{"type":"upstream_error","message":"Upstream request failed"}}`)

// handleUpstreamTransportError records a transport failure without committing
// a response. The handler remains responsible for failover exhaustion and for
// rendering the protocol-specific final error.
func (s *GatewayService) handleUpstreamTransportError(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	err error,
	event OpsUpstreamErrorEvent,
) error {
	safeErr := sanitizeUpstreamErrorMessage(err.Error())
	setOpsUpstreamError(c, 0, safeErr, "")
	event.Platform = account.Platform
	event.AccountID = account.ID
	event.AccountName = account.Name
	event.UpstreamStatusCode = 0
	event.Kind = "request_error"
	event.Message = safeErr
	appendOpsUpstreamError(c, event)

	// A canceled/timed-out client request or an already committed response must
	// never be replayed on another account and must not penalize this account.
	if errors.Is(err, context.Canceled) ||
		(ctx != nil && ctx.Err() != nil) ||
		(c != nil && c.Writer != nil && (c.Writer.Written() || IsResponseCommitted(c))) {
		return err
	}

	scheduleOllamaCloudUsageActivity(s.deferredService, account)
	if classifyOpenAITransportError(err).Persistent {
		s.tempUnscheduleTransportError(ctx, account, safeErr)
	}

	return &UpstreamFailoverError{
		StatusCode:   http.StatusBadGateway,
		ResponseBody: gatewayTransportFailoverBody,
	}
}

func (s *GatewayService) tempUnscheduleTransportError(ctx context.Context, account *Account, safeErr string) {
	if s == nil || account == nil || s.accountRepo == nil {
		return
	}
	until := time.Now().Add(gatewayTransportErrorTempUnschedDuration)
	reason := "upstream transport error (proxy/network): " + safeErr

	parent := ctx
	if parent == nil {
		parent = context.Background()
	}
	bgCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), openAIAccountStateUpdateTimeout)
	defer cancel()
	if err := s.accountRepo.SetTempUnschedulable(bgCtx, account.ID, until, reason); err != nil {
		logger.L().With(zap.String("component", "service.gateway")).Warn(
			"gateway.account_temp_unschedule_transport_failed",
			zap.String("account_id", account.ID),
			zap.Error(err),
		)
		return
	}
	logger.L().With(zap.String("component", "service.gateway")).Warn(
		"gateway.account_temp_unscheduled_transport",
		zap.String("account_id", account.ID),
		zap.String("account_name", account.Name),
		zap.String("platform", account.Platform),
		zap.Time("until", until),
		zap.String("reason", reason),
	)
}
