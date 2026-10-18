package service

import (
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

func TestResponsesCreatedAtStamper_StampsMissingAndKeepsStreamStable(t *testing.T) {
	var stamper responsesCreatedAtStamper

	created, changed := stamper.Stamp([]byte(`{"type":"response.created","response":{"id":"resp_1","status":"in_progress"}}`), "response.created")
	require.True(t, changed)
	first := gjson.GetBytes(created, "response.created_at")
	require.Equal(t, gjson.Number, first.Type)
	require.Positive(t, first.Int())

	completed, changed := stamper.Stamp([]byte(`{"type":"response.completed","response":{"id":"resp_1","status":"completed","output":[]}}`), "response.completed")
	require.True(t, changed)
	require.Equal(t, first.Int(), gjson.GetBytes(completed, "response.created_at").Int(), "同一条流的 created_at 必须一致")
}

func TestResponsesCreatedAtStamper_PreservesUpstreamValue(t *testing.T) {
	var stamper responsesCreatedAtStamper
	upstream := []byte(`{"type":"response.created","response":{"id":"resp_1","created_at":1700000000}}`)

	out, changed := stamper.Stamp(upstream, "response.created")
	require.False(t, changed)
	require.Equal(t, string(upstream), string(out), "上游已带 created_at 时字节不变")

	completed, changed := stamper.Stamp([]byte(`{"type":"response.completed","response":{"id":"resp_1"}}`), "response.completed")
	require.True(t, changed)
	require.EqualValues(t, 1700000000, gjson.GetBytes(completed, "response.created_at").Int(), "后续缺失时沿用上游值")
	require.EqualValues(t, 1700000000, stamper.Value())
}

func TestResponsesCreatedAtStamper_NullIsTreatedAsMissing(t *testing.T) {
	var stamper responsesCreatedAtStamper
	out, changed := stamper.Stamp([]byte(`{"type":"response.created","response":{"id":"resp_1","created_at":null}}`), "response.created")
	require.True(t, changed)
	require.Equal(t, gjson.Number, gjson.GetBytes(out, "response.created_at").Type)
}

func TestResponsesCreatedAtStamper_LeavesOtherEventsUntouched(t *testing.T) {
	cases := []struct {
		name      string
		eventType string
		payload   string
	}{
		{"delta", "response.output_text.delta", `{"type":"response.output_text.delta","delta":"hi"}`},
		{"bare_error", "error", `{"type":"error","error":{"message":"boom"}}`},
		{"error_with_response_field", "error", `{"type":"error","response":{"id":"resp_1"}}`},
		{"response_not_object", "response.created", `{"type":"response.created","response":"resp_1"}`},
		{"non_numeric_created_at", "response.created", `{"type":"response.created","response":{"created_at":"2026-10-16T00:00:00Z"}}`},
		{"invalid_json", "response.created", `not json`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stamper responsesCreatedAtStamper
			out, changed := stamper.Stamp([]byte(tc.payload), tc.eventType)
			require.False(t, changed)
			require.Equal(t, tc.payload, string(out))
		})
	}
}

func TestResponsesCreatedAtStamper_ObserveLearnsWithoutStamping(t *testing.T) {
	var stamper responsesCreatedAtStamper
	stamper.Observe([]byte(`{"type":"error","response":{"created_at":1600000000}}`), "error")
	stamper.Observe([]byte(`{"type":"response.created","response":{"created_at":1700000000}}`), "response.created")
	stamper.Observe([]byte(`{"type":"response.in_progress","response":{"created_at":1800000000}}`), "response.in_progress")
	require.EqualValues(t, 1700000000, stamper.Value(), "记首个生命周期事件的上游值")
}

func TestEnsureOpenAIResponseObjectCreatedAt(t *testing.T) {
	stamped := ensureOpenAIResponseObjectCreatedAt([]byte(`{"id":"resp_1","object":"response","status":"completed"}`))
	require.Equal(t, gjson.Number, gjson.GetBytes(stamped, "created_at").Type)
	require.Positive(t, gjson.GetBytes(stamped, "created_at").Int())

	for _, body := range []string{
		`{"id":"resp_1","object":"response","created_at":1700000000}`,
		`{"id":"cmp_1","object":"response.compaction","output":[]}`,
		`{"id":"resp_1","output":[]}`,
		`not json`,
	} {
		require.Equal(t, body, string(ensureOpenAIResponseObjectCreatedAt([]byte(body))))
	}
}
