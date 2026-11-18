package service

import (
	"slices"
	"strings"
	"unicode"

	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const GroupForwardingSourceKey = "group_forwarding_source"

func normalizeGroupModelForwarding(rules map[string]string) (map[string]string, error) {
	normalized := make(map[string]string, len(rules))
	for source, target := range rules {
		source, target = strings.TrimSpace(source), strings.TrimSpace(target)
		if source == "" || target == "" || strings.IndexFunc(source+target, unicode.IsSpace) >= 0 ||
			strings.Contains(strings.TrimSuffix(source, "*"), "*") || strings.Contains(target, "*") {
			return nil, infraerrors.BadRequest("INVALID_MODEL_FORWARDING", "invalid forwarding source or target")
		}
		if _, exists := normalized[source]; exists {
			return nil, infraerrors.BadRequest("INVALID_MODEL_FORWARDING", "duplicate forwarding source")
		}
		if err := CheckActiveModel(target); err != nil {
			return nil, err
		}
		normalized[source] = target
	}
	return normalized, nil
}

// ResolveForwardedModel applies one exact or longest-prefix rule before account selection.
func (g *Group) ResolveForwardedModel(model string) (string, bool) {
	if g == nil || len(g.ModelForwarding) == 0 {
		return model, false
	}
	target, matched := resolveRequestedModelInMapping(g.ModelForwarding, strings.TrimSpace(model))
	return target, matched && target != model
}

// ForwardRequestModel rewrites a client frame once, before validating or scheduling its model.
func (g *Group) ForwardRequestModel(body []byte) []byte {
	if g == nil || len(g.ModelForwarding) == 0 {
		return body
	}
	model := gjson.GetBytes(body, "model")
	if model.Type == gjson.String {
		if target, matched := g.ResolveForwardedModel(model.String()); matched {
			return ReplaceModelInBody(body, target)
		}
	}
	return body
}

func (g *Group) VisibleModelIDs(ids []string) []string {
	if g == nil || len(g.ModelForwarding) == 0 {
		return ids
	}
	return slices.DeleteFunc(slices.Clone(ids), func(id string) bool {
		_, hidden := g.ResolveForwardedModel(id)
		return hidden
	})
}

// Session updates name the model inside the session object.
func (g *Group) forwardSessionUpdate(body []byte) ([]byte, error) {
	if g == nil || len(g.ModelForwarding) == 0 || gjson.GetBytes(body, "type").String() != "session.update" {
		return body, nil
	}
	if model := gjson.GetBytes(body, "session.model"); model.Type == gjson.String {
		if target, matched := g.ResolveForwardedModel(model.String()); matched {
			return sjson.SetBytes(body, "session.model", target)
		}
	}
	return body, nil
}
