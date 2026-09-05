package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// CompleteAPIKeyCodexModelsManifestForClient fills the complete ModelInfo
// contract after copying a shared upstream entry and before group filtering.
func (s *OpenAIGatewayService) CompleteAPIKeyCodexModelsManifestForClient(manifest *CodexModelsManifest, account *Account) error {
	if manifest == nil || account == nil || !account.IsOpenAIApiKey() || manifest.NotModified || len(manifest.Body) == 0 {
		return nil
	}
	body := manifest.Body
	if len(manifest.upstreamSourceBody) > 0 {
		body = append([]byte(nil), manifest.upstreamSourceBody...)
		if manifest.convertedFromOpenAIModelList {
			body = convertOpenAIModelListToCodexManifestForAccount(body, account)
		}
	}
	var err error
	body, err = applySyncedAPIKeyCodexModelMetadata(body, account, manifest.convertedFromOpenAIModelList)
	if err != nil {
		return err
	}
	body, err = completeAPIKeyCodexModelsManifestMetadata(body, true, account)
	if err != nil {
		return err
	}
	body, err = adjustAPIKeyCodexModelsManifest(body, account)
	if err != nil {
		return err
	}
	manifest.Body = body
	manifest.ETag = codexModelsManifestBodyETag(body)
	return nil
}

func applySyncedAPIKeyCodexModelMetadata(body []byte, account *Account, overwriteLocalDefaults bool) ([]byte, error) {
	snapshot := account.GetUpstreamModelMetadataSnapshot()
	if snapshot == nil || len(snapshot.Models) == 0 {
		return body, nil
	}
	envelope, models, err := decodeCodexManifestModels(body)
	if err != nil {
		return nil, err
	}
	changed := false
	for i, rawModel := range models {
		var model map[string]json.RawMessage
		if err := json.Unmarshal(rawModel, &model); err != nil || model == nil {
			continue
		}
		var slug string
		if err := json.Unmarshal(model["slug"], &slug); err != nil {
			continue
		}
		slug = strings.TrimSpace(slug)
		lookupModel := slug
		if account != nil {
			lookupModel = account.GetMappedModel(slug)
		}
		metadata, ok := snapshot.Models[lookupModel]
		if !ok {
			continue
		}
		if lookupModel != slug {
			metadata.DisplayName = ""
			metadata.Description = ""
		}
		descriptor := newConfiguredCodexModelDescriptor(slug)
		applyUpstreamModelMetadataToCodexDescriptor(&descriptor, codexModelMetadataOverride{UpstreamModelMetadata: metadata})
		enforceGPT6AstraCodexDescriptor(&descriptor, lookupModel)
		descriptorBody, err := json.Marshal(descriptor)
		if err != nil {
			return nil, fmt.Errorf("encode synced model %q: %w", slug, err)
		}
		var syncedFields map[string]json.RawMessage
		if err := json.Unmarshal(descriptorBody, &syncedFields); err != nil {
			return nil, fmt.Errorf("decode synced model %q: %w", slug, err)
		}
		fields := make([]string, 0, 12)
		if strings.TrimSpace(metadata.DisplayName) != "" {
			fields = append(fields, "display_name")
		}
		if strings.TrimSpace(metadata.Description) != "" {
			fields = append(fields, "description")
		}
		if metadata.Reasoning != nil {
			fields = append(fields, "default_reasoning_level", "supported_reasoning_levels")
		}
		if len(normalizeCodexInputModalities(metadata.InputModalities)) > 0 {
			fields = append(fields, "input_modalities")
		}
		if metadata.ContextWindow > 0 {
			fields = append(fields, "context_window", "max_context_window")
		}
		for field := range metadata.CodexToolCapabilities {
			if _, exists := syncedFields[field]; exists {
				fields = append(fields, field)
			}
		}
		// Converted model lists already contain live capability fields; do not
		// overwrite explicit false/null declarations from the provider.
		modelChanged := applyCodexToolCapabilities(model, metadata.CodexToolCapabilities, false)
		for _, field := range fields {
			value, exists := syncedFields[field]
			if !exists {
				continue
			}
			current, currentExists := model[field]
			current = bytes.TrimSpace(current)
			if !overwriteLocalDefaults && currentExists && len(current) > 0 && !bytes.Equal(current, []byte("null")) {
				continue
			}
			if bytes.Equal(current, bytes.TrimSpace(value)) {
				continue
			}
			model[field] = value
			modelChanged = true
		}
		if !modelChanged {
			continue
		}
		encoded, err := json.Marshal(model)
		if err != nil {
			return nil, fmt.Errorf("encode model %q with synced metadata: %w", slug, err)
		}
		models[i] = encoded
		changed = true
	}
	if !changed {
		return body, nil
	}
	return encodeCodexManifestModels(envelope, models)
}

func completeAPIKeyCodexModelsManifestMetadata(body []byte, completeAll bool, account *Account) ([]byte, error) {
	envelope, models, err := decodeCodexManifestModels(body)
	if err != nil {
		return nil, err
	}
	officialOpenAI := account != nil && isOfficialOpenAIModelsBaseURL(account.GetOpenAIBaseURL())
	changed := false
	if officialOpenAI {
		filtered := make([]json.RawMessage, 0, len(models))
		for _, rawModel := range models {
			var model struct {
				Slug string `json:"slug"`
			}
			if err := json.Unmarshal(rawModel, &model); err != nil || strings.TrimSpace(model.Slug) == "" {
				filtered = append(filtered, rawModel)
				continue
			}
			if !isOfficialOpenAICodexCatalogModel(model.Slug) {
				changed = true
				continue
			}
			filtered = append(filtered, rawModel)
		}
		models = filtered
	}
	for i, rawModel := range models {
		var model map[string]json.RawMessage
		if err := json.Unmarshal(rawModel, &model); err != nil || model == nil {
			continue
		}
		var slug string
		if err := json.Unmarshal(model["slug"], &slug); err != nil {
			continue
		}
		slug = strings.TrimSpace(slug)
		if slug == "" {
			continue
		}
		completeDescriptor := completeAll || isDeepSeekCodexModel(slug)
		capabilityModel := slug
		if account != nil {
			capabilityModel = account.GetMappedModel(slug)
		}
		forceOfficialImage := officialOpenAI && (isOpenAICodexImageInputModel(slug) || isOpenAICodexImageInputModel(capabilityModel))
		if !completeDescriptor && !forceOfficialImage {
			continue
		}
		descriptor := newConfiguredCodexModelDescriptor(slug)
		enforceGPT6AstraCodexDescriptor(&descriptor, capabilityModel)
		if accountCodexModelSupportsImageInput(account, slug) {
			descriptor.InputModalities = []string{"text", "image"}
		}
		if forceOfficialImage {
			descriptor.InputModalities = []string{"text", "image"}
			descriptor.SupportsImageDetailOriginal = true
		}
		defaultBody, err := json.Marshal(descriptor)
		if err != nil {
			return nil, fmt.Errorf("encode default model %q: %w", slug, err)
		}
		var defaults map[string]json.RawMessage
		if err := json.Unmarshal(defaultBody, &defaults); err != nil {
			return nil, fmt.Errorf("decode default model %q: %w", slug, err)
		}
		capabilities := accountCodexToolCapabilities(account, capabilityModel)
		modelChanged := applyCodexToolCapabilities(model, capabilities, false)
		if completeDescriptor {
			merged, err := mergeMissingCodexModelFields(model, defaults)
			if err != nil {
				return nil, fmt.Errorf("complete model %q: %w", slug, err)
			}
			modelChanged = merged || modelChanged
		}
		if forceOfficialImage {
			modalities, _ := json.Marshal([]string{"text", "image"})
			if !bytes.Equal(bytes.TrimSpace(model["input_modalities"]), modalities) {
				model["input_modalities"] = modalities
				modelChanged = true
			}
			imageDetailOriginal := json.RawMessage("true")
			if !bytes.Equal(bytes.TrimSpace(model["supports_image_detail_original"]), imageDetailOriginal) {
				model["supports_image_detail_original"] = imageDetailOriginal
				modelChanged = true
			}
		}
		if !modelChanged {
			continue
		}
		encoded, err := json.Marshal(model)
		if err != nil {
			return nil, fmt.Errorf("encode completed model %q: %w", slug, err)
		}
		models[i] = encoded
		changed = true
	}
	if !changed {
		return body, nil
	}
	return encodeCodexManifestModels(envelope, models)
}

func mergeMissingCodexModelFields(current, defaults map[string]json.RawMessage) (bool, error) {
	changed := false
	for key, defaultValue := range defaults {
		currentValue, exists := current[key]
		if exists && stringSliceContains(codexToolCapabilityFields, key) {
			continue
		}
		if !exists || (bytes.Equal(bytes.TrimSpace(currentValue), []byte("null")) && !bytes.Equal(bytes.TrimSpace(defaultValue), []byte("null"))) {
			current[key] = defaultValue
			changed = true
			continue
		}
		var currentObject, defaultObject map[string]json.RawMessage
		if err := json.Unmarshal(currentValue, &currentObject); err != nil || currentObject == nil {
			continue
		}
		if err := json.Unmarshal(defaultValue, &defaultObject); err != nil || defaultObject == nil {
			continue
		}
		nestedChanged, err := mergeMissingCodexModelFields(currentObject, defaultObject)
		if err != nil {
			return false, err
		}
		if !nestedChanged {
			continue
		}
		mergedValue, err := json.Marshal(currentObject)
		if err != nil {
			return false, fmt.Errorf("encode field %q: %w", key, err)
		}
		current[key] = mergedValue
		changed = true
	}
	return changed, nil
}

func decodeCodexManifestModels(body []byte) (map[string]json.RawMessage, []json.RawMessage, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, nil, fmt.Errorf("decode JSON object: %w", err)
	}
	var models []json.RawMessage
	if err := json.Unmarshal(envelope["models"], &models); err != nil {
		return nil, nil, fmt.Errorf("decode top-level models array: %w", err)
	}
	return envelope, models, nil
}

func encodeCodexManifestModels(envelope map[string]json.RawMessage, models []json.RawMessage) ([]byte, error) {
	encodedModels, err := json.Marshal(models)
	if err != nil {
		return nil, fmt.Errorf("encode top-level models array: %w", err)
	}
	envelope["models"] = encodedModels
	body, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("encode JSON object: %w", err)
	}
	return body, nil
}
