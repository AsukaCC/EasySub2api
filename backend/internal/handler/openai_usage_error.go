package handler

import (
	"errors"

	"github.com/AsukaCC/EasySub2api/internal/service"
)

// openAIPartialImageResultCanBeBilled keeps an upstream 5xx retry attempt out
// of the normal usage path. A protocol bridge can return partial image usage
// together with UpstreamFailoverError after an upstream 5xx; billing that
// attempt and then retrying the request can charge twice. Other upstream
// statuses (for example 401/403) remain billable when usage was produced.
func openAIPartialImageResultCanBeBilled(result *service.OpenAIForwardResult, err error) bool {
	if result == nil || result.ImageCount <= 0 || err == nil {
		return false
	}
	var failoverErr *service.UpstreamFailoverError
	if !errors.As(err, &failoverErr) || failoverErr == nil {
		return true
	}
	return failoverErr.StatusCode < 500 || failoverErr.StatusCode >= 600
}

// openAIImagesPartialResultCanBeBilled also treats retryable Images protocol
// errors (including a 5xx response.failed event) as failover attempts.
func openAIImagesPartialResultCanBeBilled(result *service.OpenAIForwardResult, err error) bool {
	if !openAIPartialImageResultCanBeBilled(result, err) {
		return false
	}
	var upstreamErr *service.OpenAIImagesUpstreamError
	return !errors.As(err, &upstreamErr) || !service.IsOpenAIImagesRetryableUpstreamError(upstreamErr)
}
