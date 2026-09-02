package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type compositeRouteDecisionResolver interface {
	Resolve(context.Context, string, string, string) (service.CompositeRouteDecision, error)
}

func ensureCompositeTargetPlatformWithResolver(
	c *gin.Context,
	apiKey *service.APIKey,
	model string,
	endpoint string,
	resolver compositeRouteDecisionResolver,
) (bool, error) {
	if c == nil || c.Request == nil || apiKey == nil || apiKey.Group == nil || apiKey.Group.Platform != service.PlatformComposite {
		return true, nil
	}

	ctx := c.Request.Context()
	if _, sourceOK := service.CompositeRouteSourceFromContext(ctx); sourceOK {
		_, platformOK := service.ResolvedTargetPlatformFromContext(ctx)
		return platformOK, nil
	}
	if resolver == nil {
		ensureCompositeTargetPlatform(c, apiKey, model)
		_, ok := service.ResolvedTargetPlatformFromContext(c.Request.Context())
		return ok, nil
	}

	decision, err := resolver.Resolve(ctx, apiKey.Group.ID, model, endpoint)
	if err != nil {
		return false, err
	}
	if !decision.Matched {
		return false, nil
	}
	c.Request = c.Request.WithContext(service.WithCompositeRouteDecision(ctx, decision))
	return true, nil
}

func ensureCompositeTargetPlatform(c *gin.Context, apiKey *service.APIKey, model string) {
	if c == nil || c.Request == nil || apiKey == nil || apiKey.Group == nil || apiKey.Group.Platform != service.PlatformComposite {
		return
	}
	if _, ok := service.ResolvedTargetPlatformFromContext(c.Request.Context()); ok {
		return
	}
	if platform, ok := service.DetectModelPlatform(model); ok {
		c.Request = c.Request.WithContext(service.WithResolvedTargetPlatform(c.Request.Context(), platform))
	}
}

func compositeTargetPlatformAllowed(c *gin.Context, apiKey *service.APIKey, model string, allowed ...string) bool {
	if c == nil || c.Request == nil || apiKey == nil || apiKey.Group == nil || apiKey.Group.Platform != service.PlatformComposite {
		return true
	}
	ensureCompositeTargetPlatform(c, apiKey, model)
	platform, ok := service.ResolvedTargetPlatformFromContext(c.Request.Context())
	if !ok {
		return false
	}
	for _, allowedPlatform := range allowed {
		if platform == allowedPlatform {
			return true
		}
	}
	return false
}

func compositeTargetPlatformResolved(c *gin.Context, apiKey *service.APIKey, model string) bool {
	if c == nil || c.Request == nil || apiKey == nil || apiKey.Group == nil || apiKey.Group.Platform != service.PlatformComposite {
		return true
	}
	ensureCompositeTargetPlatform(c, apiKey, model)
	_, ok := service.ResolvedTargetPlatformFromContext(c.Request.Context())
	return ok
}

func effectiveAPIKeyPlatform(c *gin.Context, apiKey *service.APIKey) string {
	if c != nil && c.Request != nil {
		if platform, ok := service.ResolvedTargetPlatformFromContext(c.Request.Context()); ok {
			return platform
		}
	}
	if apiKey == nil || apiKey.Group == nil {
		return ""
	}
	return apiKey.Group.Platform
}

func openAIReasoningEffortPolicyForRequest(c *gin.Context, apiKey *service.APIKey) (string, []service.ReasoningEffortMapping, string, bool) {
	return openAIReasoningEffortPolicyForSelectedAccount(c, apiKey, nil)
}

func bindRequestedReasoningEffort(c *gin.Context, body []byte, model string) {
	if c == nil || c.Request == nil {
		return
	}
	effort := service.CanonicalRequestedReasoningEffort(body, model)
	if effort == nil {
		return
	}
	c.Request = c.Request.WithContext(service.WithRequestedReasoningEffort(c.Request.Context(), *effort))
}

func stampOpenAIRequestedReasoningEffort(result *service.OpenAIForwardResult, c *gin.Context) {
	if result == nil || result.RequestedReasoningEffort != nil || c == nil || c.Request == nil {
		return
	}
	result.RequestedReasoningEffort = service.RequestedReasoningEffortFromContext(c.Request.Context())
}

func stampForwardRequestedReasoningEffort(result *service.ForwardResult, requested *string) {
	if result == nil || result.RequestedReasoningEffort != nil {
		return
	}
	result.RequestedReasoningEffort = requested
}

func openAIReasoningEffortPolicyForSelectedAccount(c *gin.Context, apiKey *service.APIKey, account *service.Account) (string, []service.ReasoningEffortMapping, string, bool) {
	if apiKey == nil || apiKey.Group == nil {
		return "", nil, "", false
	}
	if apiKey.Group.Platform != service.PlatformOpenAI && apiKey.Group.Platform != service.PlatformComposite {
		return "", nil, "", false
	}
	if account != nil {
		if !account.IsOpenAI() {
			return "", nil, "", false
		}
	} else if effectiveAPIKeyPlatform(c, apiKey) != service.PlatformOpenAI {
		return "", nil, "", false
	}
	return apiKey.Group.MaxReasoningEffort, apiKey.Group.ReasoningEffortMappings, apiKey.Group.MaxReasoningEffortOverLimit, true
}

func applyOpenAIReasoningEffortPolicyForRequest(c *gin.Context, apiKey *service.APIKey, body []byte) ([]byte, bool, error) {
	return applyOpenAIReasoningEffortPolicyForSelectedAccount(c, apiKey, nil, body)
}

func applyOpenAIReasoningEffortPolicyForSelectedAccount(c *gin.Context, apiKey *service.APIKey, account *service.Account, body []byte) ([]byte, bool, error) {
	maxEffort, mappings, overLimit, ok := openAIReasoningEffortPolicyForSelectedAccount(c, apiKey, account)
	if !ok {
		return body, false, nil
	}
	return service.ApplyOpenAIReasoningEffortPolicy(body, maxEffort, mappings, overLimit)
}

func respondOpenAIReasoningEffortPolicyError(c *gin.Context, err error, write func(*gin.Context, int, string, string)) {
	if c == nil || err == nil || write == nil {
		return
	}
	service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalPolicyDenied)
	write(c, http.StatusForbidden, "permission_error", err.Error())
}

func bindOpenAIReasoningEffortPolicyForMessagesRequest(c *gin.Context, apiKey *service.APIKey, body []byte) {
	if c == nil || c.Request == nil {
		return
	}
	c.Request = c.Request.WithContext(withOpenAIReasoningEffortPolicyForMessagesRequest(c.Request.Context(), c, apiKey, nil, body))
}

func withOpenAIReasoningEffortPolicyForMessagesRequest(ctx context.Context, c *gin.Context, apiKey *service.APIKey, account *service.Account, body []byte) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	// The Messages bridge synthesizes a default OpenAI effort when
	// output_config.effort is omitted. Bind the group policy only for an
	// explicit client value so the ceiling does not alter that default.
	effort := gjson.GetBytes(body, "output_config.effort")
	if !effort.Exists() || effort.Type != gjson.String || strings.TrimSpace(effort.String()) == "" {
		return ctx
	}
	maxEffort, mappings, overLimit, ok := openAIReasoningEffortPolicyForSelectedAccount(c, apiKey, account)
	if !ok {
		return ctx
	}
	requestModel := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	return service.WithOpenAIReasoningEffortPolicyForModel(ctx, maxEffort, mappings, overLimit, requestModel)
}
