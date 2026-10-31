package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const moderationSavedProxyID = "115f4ef7-8343-4ce4-b4cb-a451cc098725"

type moderationHandlerSettings struct {
	service.SettingRepository
	values map[string]string
}

func (s *moderationHandlerSettings) GetValue(_ context.Context, key string) (string, error) {
	return s.values[key], nil
}

func (s *moderationHandlerSettings) Set(_ context.Context, key, value string) error {
	s.values[key] = value
	return nil
}

func moderationSettings(t *testing.T, config any) *moderationHandlerSettings {
	t.Helper()
	raw, err := json.Marshal(config)
	require.NoError(t, err)
	return &moderationHandlerSettings{values: map[string]string{
		service.SettingKeyContentModerationConfig: string(raw),
	}}
}

func callModerationHandler(t *testing.T, handler gin.HandlerFunc, method string, payload any) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", bytes.NewReader(raw))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)
	return w
}

func TestContentModerationHandlerUpdateProxySelection(t *testing.T) {
	for _, engine := range []string{service.ContentModerationEngineOpenAI, service.ContentModerationEngineTypeSafe} {
		for _, nested := range []bool{false, true} {
			for _, tc := range []struct {
				name    string
				include bool
				value   any
				want    string
			}{
				{name: "omitted preserves proxy", want: moderationSavedProxyID},
				{name: "null clears proxy", include: true},
				{name: "empty string clears proxy", include: true, value: ""},
				{name: "UUID selects proxy", include: true, value: "ab9083b9-b9e0-465e-a5b3-4b92feea1800", want: "ab9083b9-b9e0-465e-a5b3-4b92feea1800"},
			} {
				t.Run(fmt.Sprintf("%s/nested=%t/%s", engine, nested, tc.name), func(t *testing.T) {
					settings := moderationSettings(t, map[string]any{
						"engine": engine, "proxy_id": moderationSavedProxyID,
						"typesafe": map[string]any{"proxy_id": moderationSavedProxyID},
					})
					svc := service.NewContentModerationService(settings, nil, nil, nil, nil, nil, nil, nil)
					profile := map[string]any{"model": "updated-audit-model", "retry_count": 1}
					if tc.include {
						profile["proxy_id"] = tc.value
					}
					payload := profile
					if nested {
						payload = map[string]any{"engine_configs": map[string]any{engine: profile}}
					}
					w := callModerationHandler(t, NewContentModerationHandler(svc).UpdateConfig, http.MethodPut, payload)
					require.Equal(t, http.StatusOK, w.Code, w.Body.String())
					var result struct {
						Data service.ContentModerationConfigView `json:"data"`
					}
					require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
					if tc.want == "" {
						require.Nil(t, result.Data.ProxyID)
					} else {
						require.NotNil(t, result.Data.ProxyID)
						require.Equal(t, tc.want, *result.Data.ProxyID)
					}
					require.Equal(t, "updated-audit-model", result.Data.Model)
					require.Equal(t, 1, result.Data.RetryCount)
					other := service.ContentModerationEngineOpenAI
					if engine == other {
						other = service.ContentModerationEngineTypeSafe
					}
					require.NotNil(t, result.Data.EngineConfigs[other].ProxyID)
					require.Equal(t, moderationSavedProxyID, *result.Data.EngineConfigs[other].ProxyID)
					// Reload through a fresh service to verify persistence, not only the response.
					reloaded := service.NewContentModerationService(settings, nil, nil, nil, nil, nil, nil, nil)
					view, err := reloaded.GetConfig(context.Background())
					require.NoError(t, err)
					require.Equal(t, result.Data.ProxyID, view.ProxyID)
					require.Equal(t, result.Data.EngineConfigs[engine].ProxyID, view.EngineConfigs[engine].ProxyID)
				})
			}
		}
	}
}

type moderationHandlerProxyRepo struct {
	service.ProxyRepository
	proxy *service.Proxy
}

func (r *moderationHandlerProxyRepo) GetByID(_ context.Context, id string) (*service.Proxy, error) {
	if id != r.proxy.ID {
		return nil, fmt.Errorf("unexpected proxy ID: %s", id)
	}
	return r.proxy, nil
}

func TestContentModerationHandlerTestAPIKeysProxySelection(t *testing.T) {
	var directCalls, proxyCalls atomic.Int64
	respond := func(counter *atomic.Int64) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			counter.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"results":[{"flagged":false}]}`))
		}
	}
	direct := httptest.NewServer(respond(&directCalls))
	defer direct.Close()
	proxy := httptest.NewServer(respond(&proxyCalls))
	defer proxy.Close()
	proxyURL, err := url.Parse(proxy.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(proxyURL.Port())
	require.NoError(t, err)
	proxyRepo := &moderationHandlerProxyRepo{proxy: &service.Proxy{
		ID: moderationSavedProxyID, Protocol: "http", Host: proxyURL.Hostname(), Port: port, Status: service.StatusActive,
	}}
	for _, tc := range []struct {
		name    string
		include bool
		value   any
		direct  bool
	}{
		{name: "omitted uses saved proxy"},
		{name: "null forces direct", include: true, direct: true},
		{name: "empty string forces direct", include: true, value: "", direct: true},
		{name: "UUID uses selected proxy", include: true, value: moderationSavedProxyID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			settings := moderationSettings(t, map[string]any{
				"base_url": direct.URL, "proxy_id": moderationSavedProxyID, "api_keys": []string{"sk-local-test"},
			})
			svc := service.NewContentModerationService(settings, nil, nil, nil, nil, proxyRepo, nil, nil)
			payload := map[string]any{}
			if tc.include {
				payload["proxy_id"] = tc.value
			}
			beforeDirect, beforeProxy := directCalls.Load(), proxyCalls.Load()
			w := callModerationHandler(t, NewContentModerationHandler(svc).TestAPIKeys, http.MethodPost, payload)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			var result struct {
				Data service.TestContentModerationAPIKeysResult `json:"data"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
			require.Len(t, result.Data.Items, 1)
			require.Equal(t, "ok", result.Data.Items[0].Status)
			if tc.direct {
				require.Equal(t, beforeDirect+1, directCalls.Load())
				require.Equal(t, beforeProxy, proxyCalls.Load())
			} else {
				require.Equal(t, beforeDirect, directCalls.Load())
				require.Equal(t, beforeProxy+1, proxyCalls.Load())
			}
			view, err := svc.GetConfig(context.Background())
			require.NoError(t, err)
			require.NotNil(t, view.ProxyID)
			require.Equal(t, moderationSavedProxyID, *view.ProxyID)
		})
	}
}
