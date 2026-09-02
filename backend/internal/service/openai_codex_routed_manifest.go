package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/pkg/claude"
	"github.com/AsukaCC/EasySub2api/internal/pkg/openai"
	"github.com/AsukaCC/EasySub2api/internal/pkg/xai"
)

const (
	codexAutoModelPrefix               = "codex-auto-"
	configuredCodexModelPriority       = 50
	configuredCodexCustomDescription   = "Custom model routed through Sub2API."
	configuredCodexFallbackContext     = 272_000
	configuredCodexDeepSeekV4Context   = 1_000_000
	configuredCodexGrokContext         = 500_000
	configuredCodexGrokBuildContext    = 256_000
	configuredCodexGPT56MaxContext     = 872_000
	configuredCodexToolOutputMaxTokens = 10_000
)

// FilterCodexModelIDsForGroup removes routing patterns and dedicated media
// models. Automatic modes require an explicit group picker opt-in.
func FilterCodexModelIDsForGroup(modelIDs []string, group *Group) []string {
	explicitlyEnabled := make(map[string]struct{})
	if group != nil && group.CustomModelsListEnabled() {
		for _, modelID := range group.ModelsListConfig.Models {
			modelID = strings.TrimSpace(modelID)
			if strings.HasPrefix(modelID, codexAutoModelPrefix) {
				explicitlyEnabled[modelID] = struct{}{}
			}
		}
	}
	filtered := make([]string, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		modelID = strings.TrimSpace(modelID)
		if modelID == "" || strings.Contains(modelID, "*") || isCodexDedicatedMediaModel(modelID) {
			continue
		}
		if strings.HasPrefix(modelID, codexAutoModelPrefix) {
			if _, ok := explicitlyEnabled[modelID]; !ok {
				continue
			}
		}
		filtered = append(filtered, modelID)
	}
	return filtered
}

func isCodexDedicatedMediaModel(modelID string) bool {
	canonical := codexProviderQualifiedModelID(modelID)
	return IsGPTImageGenerationModel(canonical) || isImageGenerationModel(canonical) || xai.IsGrokImagineModel(modelID)
}

func codexProviderQualifiedModelID(modelID string) string {
	modelID = strings.TrimSpace(modelID)
	if slash := strings.LastIndexByte(modelID, '/'); slash >= 0 {
		modelID = strings.TrimSpace(modelID[slash+1:])
	}
	return strings.TrimPrefix(modelID, "models/")
}

// BuildGroupConfiguredCodexModelsManifest prefers explicit OpenAI account
// mappings over live discovery. The catalog's capabilities are intersected
// across persistently enabled group accounts.
func (s *OpenAIGatewayService) BuildGroupConfiguredCodexModelsManifest(ctx context.Context, group *Group, ifNoneMatch string) (*CodexModelsManifest, bool, error) {
	if s == nil || s.accountRepo == nil || group == nil || group.Platform != PlatformOpenAI {
		return nil, false, nil
	}
	visible, catalog, err := loadCodexGroupCatalogAccounts(ctx, s.accountRepo, group.ID)
	if err != nil {
		return nil, false, fmt.Errorf("load group configured Codex models: %w", err)
	}
	configuredModels := openAIConfiguredCodexModelIDsForGroup(visible, group)
	if len(configuredModels) == 0 {
		return nil, false, nil
	}
	body, err := buildCodexModelsManifestForAccounts(PlatformOpenAI, configuredModels, catalog, nil, true)
	if err != nil {
		return nil, false, fmt.Errorf("initialize group configured Codex models: %w", err)
	}
	body, _, err = mergeConfiguredCodexModelsManifest(body, nil, group.ModelsListConfig.Models, group.CustomModelsListEnabled())
	if err != nil {
		return nil, false, fmt.Errorf("build group configured Codex models: %w", err)
	}
	manifest := &CodexModelsManifest{Body: body, ETag: codexModelsManifestBodyETag(body)}
	if codexModelsManifestETagMatches(ifNoneMatch, manifest.ETag) {
		manifest.Body = nil
		manifest.NotModified = true
	}
	return manifest, true, nil
}

// MergeGroupConfiguredCodexModels applies aliases and group picker filtering
// after upstream discovery. Its ETag is always based on the final response.
func (s *OpenAIGatewayService) MergeGroupConfiguredCodexModels(ctx context.Context, group *Group, manifest *CodexModelsManifest, ifNoneMatch string) error {
	if s == nil || s.accountRepo == nil || group == nil || manifest == nil || manifest.NotModified || len(manifest.Body) == 0 {
		return nil
	}
	if group.Platform != PlatformOpenAI {
		return nil
	}
	configuredModels, err := s.groupConfiguredCodexModelIDs(ctx, group)
	if err != nil {
		return fmt.Errorf("load group configured Codex models: %w", err)
	}
	body, changed, err := mergeConfiguredCodexModelsManifest(manifest.Body, configuredModels, group.ModelsListConfig.Models, group.CustomModelsListEnabled())
	if err != nil {
		return fmt.Errorf("merge group configured Codex models: %w", err)
	}
	if changed {
		manifest.Body = body
	}
	manifest.ETag = codexModelsManifestBodyETag(manifest.Body)
	if codexModelsManifestETagMatches(ifNoneMatch, manifest.ETag) {
		manifest.Body = nil
		manifest.NotModified = true
	}
	return nil
}

func (s *OpenAIGatewayService) groupConfiguredCodexModelIDs(ctx context.Context, group *Group) ([]string, error) {
	if group == nil {
		return nil, nil
	}
	accounts, err := s.accountRepo.ListSchedulableByGroupID(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	return openAIConfiguredCodexModelIDsForGroup(accounts, group), nil
}

// Visible aliases use the current scheduler set. Capability intersection uses
// persistently enabled accounts and intentionally ignores temporary 429,
// overload, and temporary-unschedulable windows.
func loadCodexGroupCatalogAccounts(ctx context.Context, repo AccountRepository, groupID string) (visible []Account, catalog []Account, err error) {
	if repo == nil {
		return nil, nil, nil
	}
	visible, err = repo.ListSchedulableByGroupID(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}
	catalog = visible
	groupAccounts, listErr := repo.ListModelAvailabilityCandidates(ctx, &groupID, []string{
		PlatformAnthropic, PlatformOpenAI, PlatformGemini, PlatformAntigravity,
		PlatformGrok, PlatformKimi, PlatformZhipu, PlatformDeepseek,
	}, false)
	if listErr != nil {
		return visible, catalog, nil
	}
	return visible, groupAccounts, nil
}

func openAIConfiguredCodexModelIDs(accounts []Account) []string {
	seen := make(map[string]struct{})
	models := make([]string, 0)
	for i := range accounts {
		account := &accounts[i]
		if account.Platform != PlatformOpenAI {
			continue
		}
		for modelID := range account.GetModelMapping() {
			modelID = strings.TrimSpace(modelID)
			if modelID == "" || strings.Contains(modelID, "*") {
				continue
			}
			if _, exists := seen[modelID]; exists {
				continue
			}
			seen[modelID] = struct{}{}
			models = append(models, modelID)
		}
	}
	sort.Strings(models)
	return models
}

func openAIConfiguredCodexModelIDsForGroup(accounts []Account, group *Group) []string {
	models := openAIConfiguredCodexModelIDs(accounts)
	if group == nil || !group.CustomModelsListEnabled() {
		return models
	}
	seen := make(map[string]struct{}, len(models)+len(group.ModelsListConfig.Models))
	for _, modelID := range models {
		seen[modelID] = struct{}{}
	}
	for _, selectedModel := range group.ModelsListConfig.Models {
		selectedModel = strings.TrimSpace(selectedModel)
		if selectedModel == "" || strings.Contains(selectedModel, "*") {
			continue
		}
		for i := range accounts {
			account := &accounts[i]
			if account.Platform != PlatformOpenAI {
				continue
			}
			mapped, matched := account.ResolveMappedModel(selectedModel)
			if !matched || strings.TrimSpace(mapped) == "" {
				continue
			}
			if _, exists := seen[selectedModel]; !exists {
				seen[selectedModel] = struct{}{}
				models = append(models, selectedModel)
			}
			break
		}
	}
	sort.Strings(models)
	return models
}

type configuredCodexReasoningLevel struct {
	Effort      string `json:"effort"`
	Description string `json:"description"`
}

type configuredCodexTruncationPolicy struct {
	Mode  string `json:"mode"`
	Limit int64  `json:"limit"`
}

type configuredCodexServiceTier struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type configuredCodexModelMessages struct {
	InstructionsTemplate  string `json:"instructions_template"`
	InstructionsVariables any    `json:"instructions_variables"`
	Approvals             any    `json:"approvals"`
	CollaborationModes    any    `json:"collaboration_modes"`
	AutoReview            any    `json:"auto_review"`
	Permissions           any    `json:"permissions"`
	MultiAgent            any    `json:"multi_agent"`
	TokenBudget           any    `json:"token_budget"`
	GuardianV2            any    `json:"guardian_v2"`
}

type configuredCodexModelDescriptor struct {
	Slug                              string                          `json:"slug"`
	DisplayName                       string                          `json:"display_name"`
	Description                       string                          `json:"description"`
	DefaultReasoningLevel             *string                         `json:"default_reasoning_level,omitempty"`
	SupportedReasoningLevels          []configuredCodexReasoningLevel `json:"supported_reasoning_levels"`
	ShellType                         string                          `json:"shell_type"`
	Visibility                        string                          `json:"visibility"`
	SupportedInAPI                    bool                            `json:"supported_in_api"`
	Priority                          int                             `json:"priority"`
	AdditionalSpeedTiers              []string                        `json:"additional_speed_tiers"`
	ServiceTiers                      []configuredCodexServiceTier    `json:"service_tiers"`
	DefaultServiceTier                any                             `json:"default_service_tier"`
	AvailabilityNUX                   any                             `json:"availability_nux"`
	Upgrade                           any                             `json:"upgrade"`
	ModelMessages                     configuredCodexModelMessages    `json:"model_messages"`
	IncludeSkillsUsageInstructions    bool                            `json:"include_skills_usage_instructions"`
	IncludePluginUsageInstructions    bool                            `json:"include_plugin_usage_instructions"`
	IncludeAppsUsageInstructions      bool                            `json:"include_apps_usage_instructions"`
	SupportsReasoningSummaryParameter bool                            `json:"supports_reasoning_summary_parameter"`
	DefaultReasoningSummary           string                          `json:"default_reasoning_summary"`
	SupportVerbosity                  bool                            `json:"support_verbosity"`
	DefaultVerbosity                  *string                         `json:"default_verbosity"`
	ApplyPatchToolType                *string                         `json:"apply_patch_tool_type"`
	WebSearchToolType                 string                          `json:"web_search_tool_type"`
	TruncationPolicy                  configuredCodexTruncationPolicy `json:"truncation_policy"`
	SupportsImageDetailOriginal       bool                            `json:"supports_image_detail_original"`
	SupportsParallelToolCalls         bool                            `json:"supports_parallel_tool_calls"`
	ContextWindow                     int64                           `json:"context_window"`
	MaxContextWindow                  int64                           `json:"max_context_window"`
	AutoCompactTokenLimit             any                             `json:"auto_compact_token_limit"`
	CompHash                          any                             `json:"comp_hash"`
	EffectiveContextWindowPercent     int64                           `json:"effective_context_window_percent"`
	ExperimentalSupportedTools        []string                        `json:"experimental_supported_tools"`
	InputModalities                   []string                        `json:"input_modalities"`
	SupportsSearchTool                bool                            `json:"supports_search_tool"`
	UseResponsesLite                  bool                            `json:"use_responses_lite"`
	NodeREPLAutoReviewRequired        bool                            `json:"node_repl_auto_review_required"`
	NodeREPLDisabled                  bool                            `json:"node_repl_disabled"`
	AutoReviewModelOverride           any                             `json:"auto_review_model_override"`
	ModelSpecialty                    any                             `json:"model_specialty"`
	ToolMode                          any                             `json:"tool_mode"`
	MultiAgentVersion                 any                             `json:"multi_agent_version"`
}

type codexModelMetadataOverride struct {
	UpstreamModelMetadata
	reasoningConflict       bool
	inputModalitiesConflict bool
}

func newConfiguredCodexModelDescriptor(modelID string) configuredCodexModelDescriptor {
	modelID = strings.TrimSpace(modelID)
	none := "none"
	d := configuredCodexModelDescriptor{
		Slug: modelID, DisplayName: modelID, Description: configuredCodexCustomDescription,
		DefaultReasoningLevel:    &none,
		SupportedReasoningLevels: []configuredCodexReasoningLevel{{Effort: "none", Description: configuredCodexReasoningLevelDescription("none")}},
		ShellType:                "unified_exec", Visibility: "list", SupportedInAPI: true, Priority: configuredCodexModelPriority,
		AdditionalSpeedTiers: []string{}, ServiceTiers: []configuredCodexServiceTier{},
		ModelMessages:                     configuredCodexModelMessages{InstructionsTemplate: openai.CodexBaseInstructionsForModel(modelID)},
		SupportsReasoningSummaryParameter: true, DefaultReasoningSummary: "auto", WebSearchToolType: "text",
		TruncationPolicy: configuredCodexTruncationPolicy{Mode: "bytes", Limit: configuredCodexToolOutputMaxTokens},
		ContextWindow:    configuredCodexFallbackContext, MaxContextWindow: configuredCodexFallbackContext,
		EffectiveContextWindowPercent: 95, ExperimentalSupportedTools: []string{}, InputModalities: []string{"text"},
	}
	if isDeepSeekCodexModel(modelID) {
		high := "high"
		d.DisplayName = deepSeekCodexDisplayName(modelID)
		d.Description = "DeepSeek coding and reasoning model routed through EasySub2api."
		d.DefaultReasoningLevel = &high
		d.SupportedReasoningLevels = reasoningLevels("low", "high", "max")
		d.SupportsParallelToolCalls = true
		d.ContextWindow, d.MaxContextWindow = configuredCodexDeepSeekV4Context, configuredCodexDeepSeekV4Context
	}
	if isGrokCodexModel(modelID) {
		d.DisplayName = grokCodexDisplayName(modelID)
		d.Description = "Grok coding and reasoning model routed through EasySub2api."
		d.SupportsParallelToolCalls = true
		d.ContextWindow = grokCodexContextWindow(modelID)
		d.MaxContextWindow = d.ContextWindow
		if grokCodexSupportsReasoningEffort(modelID) {
			high := "high"
			d.DefaultReasoningLevel = &high
			d.SupportedReasoningLevels = configuredCodexGrokReasoningLevels(modelID)
		}
	}
	if isClaudeCodexModel(modelID) {
		d.DisplayName = claudeCodexDisplayName(modelID)
		d.Description = "Claude coding and reasoning model routed through EasySub2api."
		d.SupportsParallelToolCalls = true
		if levels := configuredCodexClaudeReasoningLevels(modelID); len(levels) > 0 {
			level := claudeCodexDefaultReasoningLevel(levels)
			d.DefaultReasoningLevel, d.SupportedReasoningLevels = &level, levels
		}
	}
	if isOpenAICodexGPTModel(modelID) {
		d.DisplayName = openaiCodexDisplayName(modelID)
		d.Description = "OpenAI GPT coding model routed through EasySub2api."
		d.SupportsParallelToolCalls = true
		if configuredCodexSupportsPriorityServiceTier(modelID) {
			d.ServiceTiers = []configuredCodexServiceTier{{ID: "priority", Name: "Fast", Description: "Priority processing for lower latency."}}
		}
		if isOpenAICodexReasoningGPTModel(modelID) {
			level := "medium"
			if getNormalizedCodexModel(modelID) == "gpt-5.6-sol" {
				level = "low"
			}
			d.DefaultReasoningLevel = &level
			d.SupportedReasoningLevels = configuredCodexGPTReasoningLevels(modelID)
			d.DefaultReasoningSummary = "none"
			d.TruncationPolicy = configuredCodexTruncationPolicy{Mode: "tokens", Limit: configuredCodexToolOutputMaxTokens}
			if isOpenAIGPT56Model(modelID) {
				d.MaxContextWindow = configuredCodexGPT56MaxContext
			}
		}
		if SupportsVerbosity(modelID) {
			verbosity := "low"
			d.SupportVerbosity, d.DefaultVerbosity = true, &verbosity
		}
	}
	return d
}

func reasoningLevels(values ...string) []configuredCodexReasoningLevel {
	levels := make([]configuredCodexReasoningLevel, 0, len(values))
	for _, value := range values {
		levels = append(levels, configuredCodexReasoningLevel{Effort: value, Description: configuredCodexReasoningLevelDescription(value)})
	}
	return levels
}

func configuredCodexSupportsPriorityServiceTier(modelID string) bool {
	normalized := canonicalizeOpenAIModelAliasSpelling(modelID)
	for _, family := range []string{"gpt-5.4", "gpt-5.5", "gpt-5.6"} {
		if normalized == family || strings.HasPrefix(normalized, family+"-") {
			return true
		}
	}
	return false
}

func configuredCodexGrokReasoningLevels(modelID string) []configuredCodexReasoningLevel {
	levels := reasoningLevels("low", "medium", "high")
	if GrokSupportsXHighReasoningEffort(modelID) {
		levels = append(levels, configuredCodexReasoningLevel{Effort: "xhigh", Description: configuredCodexReasoningLevelDescription("xhigh")})
	}
	return levels
}

func configuredCodexClaudeReasoningLevels(modelID string) []configuredCodexReasoningLevel {
	return reasoningLevels(claude.EffortLevelsForModel(modelID)...)
}

func claudeCodexDefaultReasoningLevel(levels []configuredCodexReasoningLevel) string {
	for _, preferred := range []string{"medium", "high", "low"} {
		for _, level := range levels {
			if level.Effort == preferred {
				return preferred
			}
		}
	}
	if len(levels) > 0 {
		return levels[0].Effort
	}
	return ""
}

func configuredCodexGPTReasoningLevels(modelID string) []configuredCodexReasoningLevel {
	values := []string{"low", "medium", "high", "xhigh"}
	if isOpenAIGPT56Model(modelID) {
		values = append(values, "max")
	}
	normalized := getNormalizedCodexModel(modelID)
	if normalized == "gpt-5.6-sol" || normalized == "gpt-5.6-terra" {
		values = append(values, "ultra")
	}
	return reasoningLevels(values...)
}

func GrokSupportsXHighReasoningEffort(model string) bool {
	model = strings.ToLower(xai.StripGrokProviderPrefix(strings.TrimSpace(model)))
	return model == "grok-4.6" || model == "grok-4.6-latest"
}

func isOpenAICodexGPTModel(modelID string) bool {
	normalized := canonicalizeOpenAIModelAliasSpelling(modelID)
	return normalized != "" && !strings.HasPrefix(normalized, "gpt-image") && strings.HasPrefix(normalized, "gpt-")
}

func isOpenAICodexReasoningGPTModel(modelID string) bool {
	return strings.HasPrefix(canonicalizeOpenAIModelAliasSpelling(modelID), "gpt-5")
}

func isOpenAICodexImageInputModel(modelID string) bool {
	normalized := canonicalizeOpenAIModelAliasSpelling(modelID)
	return strings.HasPrefix(normalized, "gpt-5") || strings.HasPrefix(normalized, "gpt-4o") ||
		strings.HasPrefix(normalized, "gpt-4.1") || strings.HasPrefix(normalized, "gpt-4.5") ||
		strings.HasPrefix(normalized, "gpt-4-turbo") || strings.HasPrefix(normalized, "gpt-4-vision")
}

func isOfficialOpenAICodexCatalogModel(modelID string) bool {
	normalized := strings.ToLower(codexProviderQualifiedModelID(modelID))
	if normalized == "" || isCodexDedicatedMediaModel(normalized) {
		return false
	}
	if strings.HasPrefix(normalized, "codex-") || strings.HasPrefix(normalized, "o1") || strings.HasPrefix(normalized, "o3") || strings.HasPrefix(normalized, "o4") {
		return true
	}
	if !strings.HasPrefix(normalized, "gpt-") {
		return false
	}
	for _, incompatibleFamily := range []string{"audio", "realtime", "transcribe", "tts"} {
		if strings.Contains(normalized, incompatibleFamily) {
			return false
		}
	}
	return true
}

func openaiCodexDisplayName(modelID string) string {
	normalized := canonicalizeOpenAIModelAliasSpelling(modelID)
	for _, model := range openai.DefaultModels {
		if strings.EqualFold(model.ID, normalized) && strings.TrimSpace(model.DisplayName) != "" {
			return model.DisplayName
		}
	}
	return modelID
}

func deepSeekCodexDisplayName(modelID string) string {
	switch strings.ToLower(strings.TrimSpace(modelID)) {
	case "deepseek-v4-pro", "deepseek-4-pro":
		return "DeepSeek V4 Pro"
	case "deepseek-v4-flash", "deepseek-4-flash":
		return "DeepSeek V4 Flash"
	default:
		return modelID
	}
}

func isDeepSeekCodexModel(modelID string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(modelID)), "deepseek-")
}
func isGrokCodexModel(modelID string) bool { return xai.IsGrokModelID(modelID) }

func grokCodexSupportsReasoningEffort(modelID string) bool {
	canonical := strings.ToLower(xai.ResolveGrokTextResponsesModelID(modelID))
	switch canonical {
	case "grok-4.3", "grok-4.5", "grok-4.6", "grok-3-mini", "grok-3-mini-fast", "grok-4.20-0309-reasoning", "grok-4.20-multi-agent-0309":
		return true
	default:
		return false
	}
}

func grokCodexDisplayName(modelID string) string {
	normalized := strings.ToLower(xai.StripGrokProviderPrefix(strings.TrimSpace(modelID)))
	canonical := strings.ToLower(xai.ResolveGrokTextResponsesModelID(normalized))
	for _, model := range xai.DefaultModels() {
		if model.ID == normalized || model.ID == canonical {
			return model.DisplayName
		}
	}
	return modelID
}

func grokCodexContextWindow(modelID string) int64 {
	if strings.HasPrefix(strings.ToLower(xai.StripGrokProviderPrefix(modelID)), "grok-build") {
		return configuredCodexGrokBuildContext
	}
	return configuredCodexGrokContext
}

func isClaudeCodexModel(modelID string) bool {
	platform, detected := DetectModelPlatform(modelID)
	return detected && platform == PlatformAnthropic
}

func claudeCodexDisplayName(modelID string) string {
	normalized := strings.ToLower(codexProviderQualifiedModelID(modelID))
	normalized = strings.TrimPrefix(normalized, "anthropic.")
	for _, model := range claude.DefaultModels {
		if strings.EqualFold(model.ID, normalized) && strings.TrimSpace(model.DisplayName) != "" {
			return model.DisplayName
		}
	}
	if canonical, ok := claude.ModelIDOverrides[normalized]; ok {
		for _, model := range claude.DefaultModels {
			if model.ID == canonical && strings.TrimSpace(model.DisplayName) != "" {
				return model.DisplayName
			}
		}
	}
	return modelID
}

// BuildCodexModelsManifest builds a standalone catalog for a routed provider.
func BuildCodexModelsManifest(modelIDs []string) ([]byte, error) {
	return buildCodexModelsManifest(modelIDs, nil, nil, nil)
}

// BuildCodexModelsManifestForGroup derives capabilities from the concrete
// Responses route. Unknown or mixed capabilities fail closed.
func (s *GatewayService) BuildCodexModelsManifestForGroup(ctx context.Context, group *Group, platformOverride string, modelIDs []string) ([]byte, error) {
	if s == nil || s.accountRepo == nil || group == nil {
		return BuildCodexModelsManifest(modelIDs)
	}
	effectivePlatform := strings.TrimSpace(platformOverride)
	if effectivePlatform == "" {
		effectivePlatform = group.Platform
	}
	if effectivePlatform != PlatformComposite && !isConcreteRequestPlatform(effectivePlatform) {
		return BuildCodexModelsManifest(modelIDs)
	}
	_, catalog, err := loadCodexGroupCatalogAccounts(ctx, s.accountRepo, group.ID)
	if err != nil {
		return BuildCodexModelsManifest(modelIDs)
	}
	var routes []CompositeModelRoute
	routesAvailable := true
	if effectivePlatform == PlatformComposite && s.compositeResolver != nil && s.compositeResolver.repo != nil {
		routes, err = s.compositeResolver.repo.ListByGroup(ctx, group.ID, false)
		if err != nil {
			routesAvailable = false
		}
	}
	return buildCodexModelsManifestForAccounts(effectivePlatform, modelIDs, catalog, routes, routesAvailable)
}

func buildCodexModelsManifestForAccounts(platform string, modelIDs []string, accounts []Account, routes []CompositeModelRoute, routesAvailable bool) ([]byte, error) {
	imageInput := make(map[string]bool, len(modelIDs))
	metadataModels := codexCatalogMetadataModels(platform, modelIDs, accounts, routes, routesAvailable)
	metadata := make(map[string]codexModelMetadataOverride, len(modelIDs))
	for _, modelID := range modelIDs {
		modelID = strings.TrimSpace(modelID)
		if groupCodexModelSupportsImageInput(platform, modelID, accounts, routes, routesAvailable) {
			imageInput[modelID] = true
		}
		if value, ok := groupCodexModelMetadata(platform, modelID, accounts, routes, routesAvailable); ok {
			metadata[modelID] = value
		}
	}
	return buildCodexModelsManifest(modelIDs, imageInput, metadataModels, metadata)
}

func buildCodexModelsManifest(modelIDs []string, imageInput map[string]bool, metadataModels map[string]string, metadata map[string]codexModelMetadataOverride) ([]byte, error) {
	seen := make(map[string]struct{}, len(modelIDs))
	models := make([]configuredCodexModelDescriptor, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}
		if _, exists := seen[modelID]; exists {
			continue
		}
		metadataModelID := strings.TrimSpace(metadataModels[modelID])
		if metadataModelID == "" {
			metadataModelID = modelID
		}
		if isCodexDedicatedMediaModel(modelID) || isCodexDedicatedMediaModel(metadataModelID) {
			continue
		}
		seen[modelID] = struct{}{}
		descriptor := newConfiguredCodexModelDescriptor(metadataModelID)
		descriptor.Slug = modelID
		if imageInput[modelID] {
			descriptor.InputModalities = []string{"text", "image"}
		}
		if value, ok := metadata[modelID]; ok {
			applyUpstreamModelMetadataToCodexDescriptor(&descriptor, value)
		}
		if metadataModelID != modelID {
			descriptor.DisplayName = modelID
			descriptor.Description = configuredCodexCustomDescription
		}
		models = append(models, descriptor)
	}
	return json.Marshal(struct {
		Models []configuredCodexModelDescriptor `json:"models"`
	}{Models: models})
}

func codexCatalogMetadataModels(platform string, modelIDs []string, accounts []Account, routes []CompositeModelRoute, routesAvailable bool) map[string]string {
	result := make(map[string]string, len(modelIDs))
	for _, modelID := range modelIDs {
		modelID = strings.TrimSpace(modelID)
		mapped := resolveCodexCatalogMetadataModel(platform, modelID, accounts, routes, routesAvailable)
		if mapped != "" && mapped != modelID {
			result[modelID] = mapped
		}
	}
	return result
}

func resolveCodexCatalogMetadataModel(platform, modelID string, accounts []Account, routes []CompositeModelRoute, routesAvailable bool) string {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return ""
	}
	if platform == PlatformComposite {
		if !routesAvailable {
			return modelID
		}
		if route, matched := matchCompositeRoute(routes, modelID, CompositeRouteEndpointResponses); matched {
			if upstream := strings.TrimSpace(route.UpstreamModel); upstream != "" {
				return upstream
			}
			return modelID
		}
		if codexCompositeRouteMatchesModel(routes, modelID) {
			return modelID
		}
		claimed := make(map[string]struct{})
		for _, account := range accounts {
			if isConcreteRequestPlatform(account.Platform) && codexExplicitModelMappingClaims(account, modelID) {
				claimed[account.Platform] = struct{}{}
			}
		}
		if len(claimed) > 1 {
			return modelID
		}
		for accountPlatform := range claimed {
			return uniqueCodexMappedModel(accounts, accountPlatform, modelID)
		}
		detected, ok := DetectModelPlatform(modelID)
		if !ok {
			return modelID
		}
		platform = detected
	}
	return uniqueCodexMappedModel(accounts, platform, modelID)
}

func uniqueCodexMappedModel(accounts []Account, platform, modelID string) string {
	targets := make(map[string]struct{})
	for i := range accounts {
		account := &accounts[i]
		if account.Platform != platform {
			continue
		}
		mapped, matched := account.ResolveMappedModel(modelID)
		mapped = strings.TrimSpace(mapped)
		if matched && mapped != "" {
			targets[mapped] = struct{}{}
		}
	}
	if len(targets) != 1 {
		return modelID
	}
	for target := range targets {
		return target
	}
	return modelID
}

func groupCodexModelSupportsImageInput(platform, modelID string, accounts []Account, routes []CompositeModelRoute, routesAvailable bool) bool {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return false
	}
	upstreamModel := modelID
	if platform == PlatformComposite {
		var resolved bool
		platform, upstreamModel, resolved = resolveCodexCompositeModelTarget(modelID, accounts, routes, routesAvailable)
		if !resolved {
			return false
		}
	}
	if platform != PlatformOpenAI && platform != PlatformGrok {
		return false
	}
	candidates := 0
	for i := range accounts {
		account := &accounts[i]
		if account.Platform != platform || !account.IsModelSupported(upstreamModel) {
			continue
		}
		candidates++
		if !accountCodexModelSupportsImageInput(account, account.GetMappedModel(upstreamModel)) {
			return false
		}
	}
	return candidates > 0
}

func resolveCodexCompositeModelTarget(modelID string, accounts []Account, routes []CompositeModelRoute, routesAvailable bool) (string, string, bool) {
	if !routesAvailable {
		return "", "", false
	}
	if route, matched := matchCompositeRoute(routes, modelID, CompositeRouteEndpointResponses); matched {
		upstream := strings.TrimSpace(route.UpstreamModel)
		if upstream == "" {
			upstream = modelID
		}
		return route.TargetPlatform, upstream, true
	}
	if codexCompositeRouteMatchesModel(routes, modelID) {
		return "", "", false
	}
	claimed := make(map[string]struct{})
	for _, account := range accounts {
		if isConcreteRequestPlatform(account.Platform) && codexExplicitModelMappingClaims(account, modelID) {
			claimed[account.Platform] = struct{}{}
		}
	}
	if len(claimed) > 1 {
		return "", "", false
	}
	for platform := range claimed {
		return platform, modelID, true
	}
	platform, detected := DetectModelPlatform(modelID)
	if !detected {
		return "", "", false
	}
	return platform, modelID, true
}

func codexCompositeRouteMatchesModel(routes []CompositeModelRoute, modelID string) bool {
	for _, route := range routes {
		publicModel := strings.TrimSpace(route.PublicModel)
		if publicModel == "" {
			continue
		}
		if normalizeCompositeRouteMatchType(route.MatchType) == CompositeRouteMatchPrefix {
			if strings.HasPrefix(modelID, publicModel) {
				return true
			}
		} else if modelID == publicModel {
			return true
		}
	}
	return false
}

func codexExplicitModelMappingClaims(account Account, modelID string) bool {
	if account.Credentials == nil || strings.TrimSpace(modelID) == "" {
		return false
	}
	return strings.TrimSpace(account.GetModelMapping()[modelID]) != ""
}

func accountCodexModelSupportsImageInput(account *Account, upstreamModel string) bool {
	if account == nil {
		return false
	}
	if metadata, ok := account.GetUpstreamModelMetadata(upstreamModel); ok {
		if modalities := normalizeCodexInputModalities(metadata.InputModalities); len(modalities) > 0 {
			return stringSliceContains(modalities, "image")
		}
	}
	switch account.Platform {
	case PlatformOpenAI:
		return isOpenAICodexImageInputModel(upstreamModel) && (account.IsOpenAIOAuth() || account.IsOpenAIApiKey())
	case PlatformGrok:
		if !isOfficialGrokCodexBaseURL(account.GetGrokBaseURL()) {
			return false
		}
		return isGrokCodexImageInputModel(xai.ResolveGrokTextResponsesModelID(upstreamModel))
	default:
		return false
	}
}

func isGrokCodexImageInputModel(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "grok-4.3", "grok-4.5", "grok-4.6", "grok-build-0.1",
		"grok-4.20-0309-reasoning", "grok-4.20-0309-non-reasoning", "grok-4.20-multi-agent-0309":
		return true
	default:
		return false
	}
}

func isOfficialGrokCodexBaseURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && parsed.Host != "" && xai.IsOfficialBaseURLHost(strings.TrimSuffix(parsed.Hostname(), "."))
}

func BuildDeepSeekCodexModelsManifest(modelIDs []string) ([]byte, error) {
	return BuildCodexModelsManifest(modelIDs)
}

func mergeConfiguredCodexModelsManifest(body []byte, configuredModels, selectedModels []string, filterBySelection bool) ([]byte, bool, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, false, err
	}
	var upstreamModels []json.RawMessage
	if err := json.Unmarshal(envelope["models"], &upstreamModels); err != nil {
		return nil, false, err
	}
	selected := make(map[string]struct{}, len(selectedModels))
	for _, modelID := range selectedModels {
		if modelID = strings.TrimSpace(modelID); modelID != "" {
			selected[modelID] = struct{}{}
		}
	}
	seen := make(map[string]struct{}, len(upstreamModels)+len(configuredModels))
	merged := make([]json.RawMessage, 0, len(upstreamModels)+len(configuredModels))
	changed := false
	for _, rawModel := range upstreamModels {
		var descriptor struct {
			Slug string `json:"slug"`
		}
		if err := json.Unmarshal(rawModel, &descriptor); err != nil || strings.TrimSpace(descriptor.Slug) == "" {
			if filterBySelection {
				changed = true
				continue
			}
			merged = append(merged, rawModel)
			continue
		}
		descriptor.Slug = strings.TrimSpace(descriptor.Slug)
		if isCodexDedicatedMediaModel(descriptor.Slug) {
			changed = true
			continue
		}
		if filterBySelection {
			if _, allowed := selected[descriptor.Slug]; !allowed {
				changed = true
				continue
			}
		}
		if strings.HasPrefix(descriptor.Slug, codexAutoModelPrefix) {
			_, enabled := selected[descriptor.Slug]
			if !filterBySelection || !enabled {
				changed = true
				continue
			}
			updatedRawModel, visibilityChanged, visibilityErr := codexModelWithVisibility(rawModel, "list")
			if visibilityErr != nil {
				return nil, false, visibilityErr
			}
			rawModel = updatedRawModel
			changed = changed || visibilityChanged
		}
		seen[descriptor.Slug] = struct{}{}
		merged = append(merged, rawModel)
	}
	for _, modelID := range configuredModels {
		modelID = strings.TrimSpace(modelID)
		if modelID == "" || isCodexDedicatedMediaModel(modelID) {
			continue
		}
		if filterBySelection {
			if _, allowed := selected[modelID]; !allowed {
				continue
			}
		}
		if strings.HasPrefix(modelID, codexAutoModelPrefix) {
			if _, enabled := selected[modelID]; !filterBySelection || !enabled {
				continue
			}
		}
		if _, exists := seen[modelID]; exists {
			continue
		}
		rawModel, err := json.Marshal(newConfiguredCodexModelDescriptor(modelID))
		if err != nil {
			return nil, false, err
		}
		merged = append(merged, rawModel)
		seen[modelID] = struct{}{}
		changed = true
	}
	if !changed {
		return body, false, nil
	}
	rawModels, err := json.Marshal(merged)
	if err != nil {
		return nil, false, err
	}
	envelope["models"] = rawModels
	mergedBody, err := json.Marshal(envelope)
	return mergedBody, true, err
}

func codexModelWithVisibility(rawModel json.RawMessage, visibility string) (json.RawMessage, bool, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rawModel, &fields); err != nil {
		return nil, false, err
	}
	var current string
	if raw, ok := fields["visibility"]; ok && json.Unmarshal(raw, &current) == nil && current == visibility {
		return rawModel, false, nil
	}
	raw, err := json.Marshal(visibility)
	if err != nil {
		return nil, false, err
	}
	fields["visibility"] = raw
	updated, err := json.Marshal(fields)
	return updated, true, err
}

// CodexModelsManifestETag computes the validator for the final client body.
func CodexModelsManifestETag(body []byte) string { return codexModelsManifestBodyETag(body) }

// CodexModelsManifestETagMatches applies weak, wildcard, and comma-separated
// If-None-Match semantics to the final catalog validator.
func CodexModelsManifestETagMatches(ifNoneMatch, etag string) bool {
	return codexModelsManifestETagMatches(ifNoneMatch, etag)
}
