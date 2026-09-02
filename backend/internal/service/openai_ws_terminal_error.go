package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/google/uuid"
)

const openAIWSFallbackFailureMessage = "Upstream WebSocket disconnected before completing the response. Please retry."

// buildOpenAIWSResponseFailedEvent builds a Responses API terminal event for a
// turn that cannot continue. A WebSocket close frame alone is not a Responses
// terminal event; strict clients otherwise report that the stream disconnected
// before response.completed.
func buildOpenAIWSResponseFailedEvent(responseID, code, message string) []byte {
	responseID = strings.TrimSpace(responseID)
	if responseID == "" {
		responseID = "resp_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	code = strings.TrimSpace(code)
	if code == "" {
		code = "server_error"
	}
	message = strings.TrimSpace(message)
	if message == "" {
		message = openAIWSFallbackFailureMessage
	}

	payload, err := json.Marshal(map[string]any{
		"type": "response.failed",
		"response": map[string]any{
			"id":         responseID,
			"object":     "response",
			"created_at": time.Now().Unix(),
			"status":     "failed",
			"output":     []any{},
			"error": map[string]any{
				"type":    "server_error",
				"code":    code,
				"message": message,
			},
		},
	})
	if err != nil {
		return []byte(`{"type":"response.failed","response":{"id":"resp_gateway_error","object":"response","created_at":1,"status":"failed","output":[],"error":{"type":"server_error","code":"server_error","message":"Upstream WebSocket disconnected before completing the response. Please retry."}}}`)
	}
	return payload
}

// writeOpenAIWSResponseFailed writes the semantic terminal event before the
// transport close frame. The write intentionally outlives an upstream timeout
// or cancellation for a short bounded period so the still-connected client can
// receive the actual failure reason.
func writeOpenAIWSResponseFailed(
	ctx context.Context,
	conn *coderws.Conn,
	responseID string,
	code string,
	message string,
	timeout time.Duration,
) error {
	if conn == nil {
		return nil
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	parent := context.Background()
	if ctx != nil {
		parent = context.WithoutCancel(ctx)
	}
	writeCtx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	return conn.Write(writeCtx, coderws.MessageText, buildOpenAIWSResponseFailedEvent(responseID, code, message))
}
