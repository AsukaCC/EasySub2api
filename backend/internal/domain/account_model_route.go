package domain

// AccountModelRoute describes one client-model to upstream-model route.
// An identical request/upstream pair acts as an allow-list entry.
type AccountModelRoute struct {
	RequestModel    string `json:"request_model"`
	UpstreamModel   string `json:"upstream_model"`
	ReasoningEffort string `json:"reasoning_effort,omitempty"`
}
