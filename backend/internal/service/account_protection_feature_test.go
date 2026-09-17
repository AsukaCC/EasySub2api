package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/pkg/claude"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func protectionTestAccount(id string) *Account {
	return &Account{ID: id, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": id}, Extra: map[string]any{}}
}

func TestProtectionNewAndExistingDefaults(t *testing.T) {
	old := protectionTestAccount("old")
	require.False(t, old.AntiDegradationEnabled())
	require.Equal(t, "off", old.RequestIntegrityMode())
	a := protectionTestAccount("new")
	PrepareNewAccountProtection(a)
	require.Equal(t, "legacy", a.ProtectionMode())
	require.Equal(t, 16, a.Mode1EffectiveConcurrency())
	require.Equal(t, "observe", a.RequestIntegrityMode())
	require.NoError(t, ValidateAccountProtectionConfiguration(a))
	a.Concurrency = 32
	BoundAccountProtectionConcurrency(a)
	require.Equal(t, 32, a.Mode1EffectiveConcurrency())
	copy := *a
	PrepareNewAccountProtection(&copy)
	require.NotEqual(t, a.Extra[codexFingerprintSeedExtraKey], copy.Extra[codexFingerprintSeedExtraKey])
	generic := &Account{Platform: PlatformGemini, Type: AccountTypeAPIKey}
	PrepareNewAccountProtection(generic)
	require.Equal(t, "generic_v1", generic.ProtectionScope())
	require.False(t, generic.IsTLSFingerprintEnabled())
}

type protectionFeatureStore struct {
	a      *Account
	writes int
}

func (s *protectionFeatureStore) GetAccount(context.Context, string) (*Account, error) {
	a := *s.a
	return &a, nil
}
func (s *protectionFeatureStore) UpdateAccount(ctx context.Context, _ string, in *UpdateAccountInput) (*Account, error) {
	_, ok := GetProtectionWriteExpectation(ctx)
	if !ok {
		panic("missing write expectation")
	}
	a := *s.a
	a.Extra = prepareCodexFingerprintExtraForUpdate(&a, in.Extra)
	if in.Concurrency != nil {
		a.Concurrency = *in.Concurrency
	}
	s.a = &a
	s.writes++
	return s.GetAccount(ctx, a.ID)
}

func TestProtectionTransitionsAndPreservation(t *testing.T) {
	a := protectionTestAccount("a")
	a.Concurrency = 7
	store := &protectionFeatureStore{a: a}
	svc := NewAntiDegradeService(store)
	ctx := context.Background()
	_, err := svc.ApplyAntiDegradeMode(ctx, a.ID, AntiDegradeModeLegacy)
	require.NoError(t, err)
	require.Equal(t, 1, store.writes)
	_, err = svc.ApplyAntiDegradeMode(ctx, a.ID, AntiDegradeModeLegacy)
	require.NoError(t, err)
	require.Equal(t, 1, store.writes)
	_, err = svc.ApplyAntiDegradeMode(ctx, a.ID, AntiDegradeMode1)
	require.NoError(t, err)
	require.Equal(t, 2, store.writes)
	require.Equal(t, "enforce", store.a.RequestIntegrityMode())
	preserved := PreserveAccountProtection(ctx, store.a, map[string]any{"anti_degradation": false, "enable_tls_fingerprint": true})
	require.Equal(t, true, preserved["anti_degradation"])
	require.Equal(t, false, preserved["enable_tls_fingerprint"])
	_, err = svc.SetProtection(ctx, a.ID, false, false)
	require.Error(t, err)
	store.a.Concurrency = 12
	_, err = svc.SetProtection(ctx, a.ID, false, true)
	require.NoError(t, err)
	require.Equal(t, 12, store.a.Concurrency)
	require.False(t, store.a.AntiDegradationEnabled())
}

func TestProtectionNamespaceIndependentOfAPIKey(t *testing.T) {
	a, b := protectionTestAccount("a"), protectionTestAccount("b")
	first := isolateOpenAIUpstreamSessionID("key-a", a, "session")
	require.Regexp(t, uuidV4Pattern, first)
	require.Equal(t, first, isolateOpenAIUpstreamSessionID("key-b", a, "session"))
	require.NotEqual(t, first, isolateOpenAIUpstreamSessionID("key-a", b, "session"))
	a.Credentials["access_token"] = "rotated-test-value"
	require.Equal(t, first, isolateOpenAIUpstreamSessionID("", a, "session"))
	h := http.Header{"Session_id": []string{"session"}, "X-Codex-Installation-Id": []string{"device"}}
	body := []byte(`{"client_metadata":{"session_id":"session","x-codex-installation-id":"device"}}`)
	projected, _, err := applyCodexAccountIdentityClientMetadataRaw(body, a, "key-b")
	require.NoError(t, err)
	applyCodexAccountIdentityHeaders(h, a, "key-a")
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(projected, &decoded))
	cm := decoded["client_metadata"].(map[string]any)
	require.Equal(t, first, h.Get("session_id"))
	require.Equal(t, cm["session_id"], h.Get("session_id"))
	require.Equal(t, cm["x-codex-installation-id"], h.Get("x-codex-installation-id"))
}

func TestProtectionIntegrityModesAndLosslessConversion(t *testing.T) {
	a := protectionTestAccount("a")
	before := []byte(`{"model":"gpt-5.6-sol","input":"hello","instructions":"keep","reasoning":{"effort":"high"},"tools":[{"type":"function","name":"test","parameters":{"type":"object"}}]}`)
	after := []byte(`{"model":"gpt-5.6-sol","input":[{"role":"user","content":[{"type":"input_text","text":"hello"}]}],"instructions":"keep","reasoning":{"effort":"high"},"tools":[{"type":"function","name":"test","parameters":{"type":"object"}}]}`)
	require.NoError(t, validateMode1RequestIntegrityForAccount(a, before, after))
	broken := []byte(strings.Replace(string(after), `"effort":"high"`, `"effort":"low"`, 1))
	a.Extra["request_integrity_mode"] = "observe"
	require.NoError(t, checkAccountRequestIntegrity(nil, a, before, broken))
	a.Extra["request_integrity_mode"] = "enforce"
	require.True(t, IsAccountProtectionError(checkAccountRequestIntegrity(nil, a, before, broken)))
	a.Type = AccountTypeAPIKey
	require.Error(t, validateMode1RequestIntegrityForAccount(a, []byte(`{"max_output_tokens":100}`), []byte(`{}`)))
}

func TestProtectionHTTPAndWSIdentityHeaders(t *testing.T) {
	a := protectionTestAccount("upstream")
	s := &OpenAIGatewayService{}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Request.Header.Set("session_id", "session")
	c.Request.Header.Set("x-codex-installation-id", "device")
	body := []byte(`{"model":"gpt-5.6-sol","input":[],"client_metadata":{"x-codex-installation-id":"device"}}`)
	req, err := s.buildUpstreamRequest(context.Background(), c, a, body, "test", true, "session", true)
	require.NoError(t, err)
	require.Equal(t, "codex-tui", req.Header.Get("originator"))
	require.Equal(t, scopeCodexAccountIdentityValue(a, "", "installation", "device"), req.Header.Get("x-codex-installation-id"))
	ws, _, err := s.buildOpenAIWSHeaders(context.Background(), c, a, "test", OpenAIWSProtocolDecision{}, true, "", "", "session", "gpt-5.6-sol", "")
	require.NoError(t, err)
	require.Equal(t, req.Header.Get("session_id"), ws.Get("session_id"))
	require.NotEqual(t, "responses=experimental", ws.Get("OpenAI-Beta"))
}

func TestProtectionBridgeIdentityWithoutOriginator(t *testing.T) {
	previous := codexIdentityEnforcement.Load()
	t.Cleanup(func() { SetCodexIdentityEnforcementEnabled(previous) })
	for _, tc := range []struct {
		name      string
		enforce   bool
		accountUA string
		clientUA  string
		product   string
	}{
		{name: "default", enforce: true, product: "codex-tui"},
		{name: "account override", enforce: true, accountUA: "codex_cli_rs/9.9.9 (Linux; arm64) xterm", product: "codex_cli_rs"},
		{name: "client identity with enforcement disabled", clientUA: "codex_vscode/9.9.9 (Linux; arm64) vscode", product: "codex_vscode"},
		{name: "default with enforcement disabled", product: "codex-tui"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			SetCodexIdentityEnforcementEnabled(tc.enforce)
			a := protectionTestAccount("bridge")
			a.Credentials["user_agent"] = tc.accountUA
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
			c.Request.Header.Set("originator", "opencode")
			c.Request.Header.Set("User-Agent", tc.clientUA)
			setOpenAICompatMessagesBridgeContext(c, true)
			s := &OpenAIGatewayService{}
			req, err := s.buildUpstreamRequest(context.Background(), c, a,
				[]byte(`{"model":"gpt-5.6-sol","input":[],"instructions":"test"}`), "test", true, "", false)
			require.NoError(t, err)
			require.Equal(t, tc.product+"/"+codexCLIVersion, strings.Fields(req.UserAgent())[0])
			require.Equal(t, codexCLIVersion, req.Header.Get("version"))
			require.Empty(t, req.Header.Get("originator"))
			require.Empty(t, req.Header.Get("OpenAI-Beta"))
		})
	}
}

func TestProtectionTLSSelectionAndClaudeFloor(t *testing.T) {
	a := protectionTestAccount("a")
	PrepareNewAccountProtection(a)
	p, err := protectedTLSProfile(a, true)
	require.NoError(t, err)
	require.NotNil(t, p)
	p, err = protectedTLSProfile(a, false)
	require.NoError(t, err)
	require.Nil(t, p)
	cfg := &config.Config{}
	require.Equal(t, "standard", ResolveProtectionRuntime(a, cfg).EffectiveTLS)
	floor, changed := floorClaudeCLIUserAgentVersion("claude-cli/1.0.0 (custom, sdk)")
	require.True(t, changed)
	require.Equal(t, "claude-cli/"+claude.CLIVersion()+" (custom, sdk)", floor)
	floor, changed = floorClaudeCLIUserAgentVersion("other/1.0.0")
	require.False(t, changed)
	require.Equal(t, "other/1.0.0", floor)
}
