package openai

import (
	"encoding/json"
	"regexp"
	"strings"
)

var retiredGPTFamily = regexp.MustCompile(`^gpt-?5[.-]?[45](?:$|[^0-9])`)

// IsRetiredModel applies to new traffic and public catalogs, never stored usage.
func IsRetiredModel(model string) bool {
	for _, segment := range strings.Split(strings.ToLower(strings.TrimSpace(model)), "/") {
		segment = strings.ReplaceAll(segment, "_", "-")
		segment = strings.Join(strings.Fields(segment), "-")
		for strings.Contains(segment, "--") {
			segment = strings.ReplaceAll(segment, "--", "-")
		}
		if retiredGPTFamily.MatchString(segment) {
			return true
		}
	}
	return false
}

func FilterActiveModelIDs(ids []string) []string {
	active := make([]string, 0, len(ids))
	for _, id := range ids {
		if !IsRetiredModel(id) {
			active = append(active, id)
		}
	}
	return active
}

// FilterRetiredManifest preserves all unrelated fields and opaque descriptors.
func FilterRetiredManifest(body []byte) ([]byte, bool, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, false, err
	}
	changed := false
	for _, field := range []string{"models", "data"} {
		raw, ok := envelope[field]
		if !ok {
			continue
		}
		var entries []json.RawMessage
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, false, err
		}
		active := make([]json.RawMessage, 0, len(entries))
		for _, entry := range entries {
			var descriptor struct {
				ID   string `json:"id"`
				Slug string `json:"slug"`
			}
			if err := json.Unmarshal(entry, &descriptor); err != nil {
				return nil, false, err
			}
			if IsRetiredModel(descriptor.ID) || IsRetiredModel(descriptor.Slug) {
				changed = true
				continue
			}
			active = append(active, entry)
		}
		if len(active) != len(entries) {
			encoded, err := json.Marshal(active)
			if err != nil {
				return nil, false, err
			}
			envelope[field] = encoded
		}
	}
	if !changed {
		return body, false, nil
	}
	encoded, err := json.Marshal(envelope)
	return encoded, true, err
}
