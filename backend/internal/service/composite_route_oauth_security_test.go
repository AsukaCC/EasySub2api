package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCompositeRouteOAuthAmbiguityRejectsPublicAliases(t *testing.T) {
	r := NewCompositeRouteResolver(nil)
	r.SetModelOwnershipResolver(func(context.Context, string, string) (CompositeModelOwnership, error) {
		return CompositeModelOwnership{Ambiguous: true, OpenAIGatewayOAuthClaimed: true}, nil
	})
	for _, model := range []string{"gpt-6.1-sol", "claude-haiku-5-5", "private-alias"} {
		for _, endpoint := range []string{CompositeRouteEndpointResponses, CompositeRouteEndpointMessages, CompositeRouteEndpointCountTokens, CompositeRouteEndpointChatCompletions} {
			decision, err := r.Resolve(context.Background(), "group", model, endpoint)
			require.NoError(t, err)
			require.True(t, decision.OAuthAmbiguous())
			require.Contains(t, decision.AmbiguousMessage(), model)
		}
	}
}
