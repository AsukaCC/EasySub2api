package handler

import (
	"errors"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
)

func TestOpenAIPartialImageResultCanBeBilled(t *testing.T) {
	result := &service.OpenAIForwardResult{ImageCount: 1}

	if !openAIPartialImageResultCanBeBilled(result, errors.New("stream read error")) {
		t.Fatal("non-failover partial usage should remain billable")
	}
	if openAIPartialImageResultCanBeBilled(result, &service.UpstreamFailoverError{StatusCode: 503}) {
		t.Fatal("failover partial usage must not be billed before retry")
	}
	if !openAIPartialImageResultCanBeBilled(result, &service.UpstreamFailoverError{StatusCode: 403}) {
		t.Fatal("non-retryable failover partial usage should remain billable")
	}
}

func TestOpenAIImagesPartialResultCanBeBilled(t *testing.T) {
	result := &service.OpenAIForwardResult{ImageCount: 1}

	if openAIImagesPartialResultCanBeBilled(result, &service.OpenAIImagesUpstreamError{StatusCode: 503}) {
		t.Fatal("retryable Images 5xx partial usage must not be billed")
	}
	if !openAIImagesPartialResultCanBeBilled(result, &service.OpenAIImagesUpstreamError{StatusCode: 400}) {
		t.Fatal("non-retryable Images partial usage should remain billable")
	}
}
