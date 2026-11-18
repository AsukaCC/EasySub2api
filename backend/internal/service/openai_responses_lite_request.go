package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type responsesLiteValidationError struct{ error }

// ResponsesLiteClientErrorMessage distinguishes a client opt-in from an account
// header override. I/O and encoding errors must not become client errors.
func ResponsesLiteClientErrorMessage(err error, headers http.Header, body []byte) (string, bool) {
	var invalid *responsesLiteValidationError
	if !errors.As(err, &invalid) {
		return "", false
	}
	if !isOpenAIResponsesLiteHeader(headers.Get(responsesLiteHeader)) && !isOpenAIResponsesLiteWebSocketPayload(body) {
		return "", false
	}
	return invalid.Error() + "; use full Responses or send a Lite-compatible request", true
}

// Normalize the final body after header simulation without changing the selected protocol.
func finalizeOpenAIResponsesLiteRequest(req *http.Request) error {
	replay, err := req.GetBody()
	if err != nil {
		return fmt.Errorf("read final Responses Lite body: %w", err)
	}
	defer func() { _ = replay.Close() }()
	body, err := io.ReadAll(replay)
	if err != nil {
		return fmt.Errorf("read final Responses Lite body: %w", err)
	}
	lite := isOpenAIResponsesLiteHeader(req.Header.Get(responsesLiteHeader))
	if !lite && !isOpenAIResponsesLiteWebSocketPayload(body) {
		return nil
	}
	body, changed, err := normalizeOpenAIResponsesLiteToolsPayload(body)
	if err != nil {
		return fmt.Errorf("finalize Responses Lite body: %w", err)
	}
	if !changed {
		return nil
	}
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	req.Body, _ = req.GetBody()
	req.ContentLength = int64(len(body))
	return nil
}
