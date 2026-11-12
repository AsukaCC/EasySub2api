package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBufferedResponsePreservesSealedNamespacedCalls(t *testing.T) {
	for _, marker := range []string{`null`, `[]`, `"sealed"`, `{"cipher":"value"}`} {
		t.Run(marker, func(t *testing.T) {
			acc := NewBufferedResponseAccumulator()
			acc.ProcessEvent(&ResponsesStreamEvent{
				Type: "response.output_item.added", OutputIndex: 2,
				Item: &ResponsesOutput{Type: "function_call", Name: "js", CallID: "call_1", Namespace: "mcp__node_repl", EncryptedFunctionArgs: json.RawMessage(marker)},
			})
			acc.ProcessEvent(&ResponsesStreamEvent{Type: "response.function_call_arguments.delta", OutputIndex: 2, Delta: `{"code":"1"}`})
			out := acc.BuildOutput()
			require.Len(t, out, 1)
			require.Equal(t, "mcp__node_repl", out[0].Namespace)
			require.Equal(t, `{"code":"1"}`, out[0].Arguments)
			wire, err := json.Marshal(out[0])
			require.NoError(t, err)
			var decoded ResponsesOutput
			require.NoError(t, json.Unmarshal(wire, &decoded))
			require.JSONEq(t, marker, string(decoded.EncryptedFunctionArgs))
			lifecycle := responsesItemWire(&decoded)
			require.Equal(t, "mcp__node_repl", lifecycle["namespace"])
			require.JSONEq(t, marker, string(lifecycle["encrypted_function_args"].(json.RawMessage)))
		})
	}
}
