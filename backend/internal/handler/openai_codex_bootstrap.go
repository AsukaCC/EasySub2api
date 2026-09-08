package handler

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	codexBootstrapKindDelegation = "delegation"
	codexBootstrapKindAutomation = "scheduled_automation"
)

// normalizeCodexCallOutputBootstrap converts the two call-less Codex bootstrap
// envelopes into user messages. Ordinary function outputs remain subject to the
// regular call_id/item_reference validation in the gateway handler.
//
// Delegation（create_thread / send_message_to_thread）：已有任务通过
// send_message_to_thread 唤醒时会携带 previous_response_id；完整历史回放还会带有
// 已配对的调用项。delegation 仍是客户端注入的用户输入，不属于这些历史调用的结果，
// 因此允许它与可明确配对（带 id / call_id）的历史上下文共存。
// Automation（automation_update，含 heartbeat 形态）保持 bootstrap-only 边界。
func normalizeCodexCallOutputBootstrap(body []byte) ([]byte, string, bool) {
	if normalized, changed := normalizeCodexBootstrapByKind(body, isCodexDelegationCandidate, true); changed {
		return normalized, codexBootstrapKindDelegation, true
	}
	if normalized, changed := normalizeCodexBootstrapByKind(body, isCodexAutomationCandidate, false); changed {
		return normalized, codexBootstrapKindAutomation, true
	}
	return body, "", false
}

func normalizeCodexBootstrapByKind(body []byte, isCandidate func(map[string]any) bool, allowHistoricalContext bool) ([]byte, bool) {
	if !hasUniqueCodexBootstrapJSONMembers(body) {
		return body, false
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var request map[string]any
	if err := decoder.Decode(&request); err != nil {
		return body, false
	}
	if previousResponseID, exists := request["previous_response_id"]; exists {
		value, ok := previousResponseID.(string)
		if !ok || (!allowHistoricalContext && strings.TrimSpace(value) != "") {
			return body, false
		}
	}

	input, ok := request["input"].([]any)
	if !ok {
		return body, false
	}

	// Responses built-ins follow the *_call / *_call_output naming convention,
	// so classify by the wire type shape instead of maintaining an incomplete
	// allowlist. Delegation may coexist with historical anchors only when their
	// IDs make them unambiguous; automation retains the bootstrap-only boundary.
	for _, raw := range input {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		typ := codexBootstrapStringField(item, "type")
		if isCandidate(item) {
			callIDValue, exists := item["call_id"]
			callID, isString := callIDValue.(string)
			if exists && (!isString || strings.TrimSpace(callID) != "") {
				return body, false
			}
			continue
		}
		if typ == "item_reference" {
			if allowHistoricalContext && strings.TrimSpace(codexBootstrapStringField(item, "id")) != "" {
				continue
			}
			return body, false
		}
		if strings.HasSuffix(typ, "_call") || isCodexResponsesCallOutputType(typ) {
			if allowHistoricalContext && strings.TrimSpace(codexBootstrapStringField(item, "call_id")) != "" {
				continue
			}
			return body, false
		}
	}

	changed := false
	for i, raw := range input {
		item, ok := raw.(map[string]any)
		if !ok || !isCandidate(item) {
			continue
		}
		output, ok := item["output"].(string)
		if !ok {
			continue
		}
		input[i] = map[string]any{
			"type": "message",
			"role": "user",
			"content": []any{map[string]any{
				"type": "input_text",
				"text": output,
			}},
		}
		changed = true
	}
	if !changed {
		return body, false
	}
	normalized, err := json.Marshal(request)
	if err != nil {
		return body, false
	}
	return normalized, true
}

func isCodexResponsesCallOutputType(typ string) bool {
	return strings.HasSuffix(typ, "_call_output") || typ == "tool_search_output"
}

func isCodexDelegationCandidate(item map[string]any) bool {
	if codexBootstrapStringField(item, "type") != "function_call_output" || !isCodexDelegationBootstrapItem(item) {
		return false
	}
	output, ok := item["output"].(string)
	return ok && validCodexDelegationEnvelope(output)
}

func isCodexAutomationCandidate(item map[string]any) bool {
	if codexBootstrapStringField(item, "type") != "function_call_output" || !isCodexAutomationBootstrapItem(item) {
		return false
	}
	output, ok := item["output"].(string)
	return ok && (validCodexAutomationBootstrap(output) || validCodexAutomationHeartbeat(output))
}

func isCodexDelegationBootstrapItem(item map[string]any) bool {
	namespace := codexBootstrapStringField(item, "namespace")
	name := codexBootstrapStringField(item, "name")
	return (namespace == "codex_app" || namespace == "codex_tui") &&
		(name == "create_thread" || name == "send_message_to_thread")
}

func isCodexAutomationBootstrapItem(item map[string]any) bool {
	return codexBootstrapStringField(item, "namespace") == "codex_app" &&
		codexBootstrapStringField(item, "name") == "automation_update"
}

// validCodexAutomationHeartbeat 校验严格的 <heartbeat><automation_id>…</automation_id></heartbeat>
// 形态：无命名空间/属性/注释，automation_id 为唯一子元素且值无首尾空白。
func validCodexAutomationHeartbeat(value string) bool {
	decoder := xml.NewDecoder(strings.NewReader(value))
	var rootSeen, automationIDSeen bool
	var automationID bytes.Buffer
	depth := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			id := automationID.String()
			return rootSeen && automationIDSeen && depth == 0 &&
				strings.TrimSpace(id) == id && validCodexAutomationID(id)
		}
		if err != nil {
			return false
		}
		switch current := token.(type) {
		case xml.StartElement:
			depth++
			if current.Name.Space != "" || len(current.Attr) != 0 || depth > 2 {
				return false
			}
			if depth == 1 {
				if rootSeen || current.Name.Local != "heartbeat" {
					return false
				}
				rootSeen = true
			} else if automationIDSeen || current.Name.Local != "automation_id" {
				return false
			}
			automationIDSeen = depth == 2
		case xml.EndElement:
			if current.Name.Space != "" {
				return false
			}
			depth--
			if depth < 0 {
				return false
			}
		case xml.CharData:
			if depth == 2 {
				_, _ = automationID.Write(current)
			} else if len(bytes.TrimSpace(current)) != 0 {
				return false
			}
		case xml.Comment, xml.ProcInst, xml.Directive:
			return false
		}
	}
}

func codexBootstrapStringField(item map[string]any, key string) string {
	value, _ := item[key].(string)
	return value
}

func validCodexAutomationBootstrap(value string) bool {
	normalized := strings.ReplaceAll(value, "\r\n", "\n")
	if strings.ContainsRune(normalized, '\r') {
		return false
	}
	lines := strings.Split(normalized, "\n")
	if len(lines) < 6 {
		return false
	}
	if _, ok := codexAutomationHeaderValue(lines[0], "Automation: "); !ok {
		return false
	}
	automationID, ok := codexAutomationHeaderValue(lines[1], "Automation ID: ")
	if !ok || !validCodexAutomationID(automationID) {
		return false
	}
	if lines[2] != "Automation memory: $CODEX_HOME/automations/"+automationID+"/memory.md" {
		return false
	}
	lastRun, ok := codexAutomationHeaderValue(lines[3], "Last run: ")
	if !ok || !validCodexAutomationLastRun(lastRun) || lines[4] != "" {
		return false
	}
	return strings.TrimSpace(strings.Join(lines[5:], "\n")) != ""
}

func codexAutomationHeaderValue(line, prefix string) (string, bool) {
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	value := strings.TrimPrefix(line, prefix)
	return value, value != "" && strings.TrimSpace(value) == value
}

func validCodexAutomationID(value string) bool {
	if len(value) == 0 || len(value) > 128 || value == "." || value == ".." {
		return false
	}
	for i := 0; i < len(value); i++ {
		char := value[i]
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func validCodexAutomationLastRun(value string) bool {
	if value == "never" {
		return true
	}
	separator := strings.LastIndex(value, " (")
	if separator <= 0 || !strings.HasSuffix(value, ")") {
		return false
	}
	runAt, err := time.Parse(time.RFC3339Nano, value[:separator])
	if err != nil {
		return false
	}
	epochMillis, err := strconv.ParseInt(value[separator+2:len(value)-1], 10, 64)
	return err == nil && runAt.UnixMilli() == epochMillis
}

func validCodexDelegationEnvelope(value string) bool {
	decoder := xml.NewDecoder(strings.NewReader(value))
	var rootSeen, sourceSeen, inputSeen bool
	var childName string
	var childText bytes.Buffer
	depth := 0

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return rootSeen && depth == 0 && sourceSeen && inputSeen
		}
		if err != nil {
			return false
		}

		switch current := token.(type) {
		case xml.StartElement:
			depth++
			if current.Name.Space != "" || len(current.Attr) != 0 ||
				(depth == 1 && current.Name.Local != "codex_delegation") || depth > 2 {
				return false
			}
			if depth == 1 {
				if rootSeen {
					return false
				}
				rootSeen = true
				continue
			}
			if current.Name.Local != "source_thread_id" && current.Name.Local != "input" {
				return false
			}
			childName = current.Name.Local
			childText.Reset()
		case xml.EndElement:
			if current.Name.Space != "" {
				return false
			}
			if depth == 2 {
				if current.Name.Local != childName || strings.TrimSpace(childText.String()) == "" {
					return false
				}
				if childName == "source_thread_id" {
					if sourceSeen {
						return false
					}
					sourceSeen = true
				} else {
					if inputSeen {
						return false
					}
					inputSeen = true
				}
				childName = ""
			}
			depth--
			if depth < 0 {
				return false
			}
		case xml.CharData:
			if depth == 2 {
				_, _ = childText.Write(current)
			} else if len(bytes.TrimSpace(current)) != 0 {
				return false
			}
		case xml.Comment, xml.ProcInst, xml.Directive:
			return false
		}
	}
}

func hasUniqueCodexBootstrapJSONMembers(body []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(body))
	if !consumeUniqueCodexBootstrapJSONValue(decoder) {
		return false
	}
	_, err := decoder.Token()
	return err == io.EOF
}

func consumeUniqueCodexBootstrapJSONValue(decoder *json.Decoder) bool {
	token, err := decoder.Token()
	if err != nil {
		return false
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return true
	}

	switch delim {
	case '{':
		members := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return false
			}
			key, ok := keyToken.(string)
			if !ok {
				return false
			}
			if _, duplicate := members[key]; duplicate {
				return false
			}
			members[key] = struct{}{}
			if !consumeUniqueCodexBootstrapJSONValue(decoder) {
				return false
			}
		}
		end, err := decoder.Token()
		return err == nil && end == json.Delim('}')
	case '[':
		for decoder.More() {
			if !consumeUniqueCodexBootstrapJSONValue(decoder) {
				return false
			}
		}
		end, err := decoder.Token()
		return err == nil && end == json.Delim(']')
	default:
		return false
	}
}
