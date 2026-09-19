package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type singleLevelHandlerRepo struct {
	service.UserLevelRepository
	service.UserLevelRulesRepository
	rule   service.UserLevelRule
	search string
}

func (r *singleLevelHandlerRepo) GetLevelRule(context.Context, string) (*service.UserLevelRule, error) {
	return &r.rule, nil
}
func (r *singleLevelHandlerRepo) ListLevelRuleMembers(_ context.Context, _ string, _, _ int, search ...string) ([]service.User, int64, error) {
	if len(search) > 0 {
		r.search = search[0]
	}
	return []service.User{}, 0, nil
}

func TestSingleUserLevelAdminContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const id = "11111111-1111-4111-8111-111111111111"
	repo := &singleLevelHandlerRepo{rule: service.UserLevelRule{ID: id, Name: "Default", WindowDays: 7, Enabled: true, IsDefault: true}}
	handler := &UserHandler{userLevelService: service.NewUserLevelService(repo, nil, nil, nil, nil)}
	engine := gin.New()
	engine.PUT("/users/:id/level-rules", handler.ReplaceUserLevelRules)
	engine.PUT("/rules/:id", handler.UpdateLevelRule)
	engine.DELETE("/rules/:id", handler.DeleteLevelRule)
	engine.GET("/rules/:id/members", handler.ListLevelRuleMembers)
	for _, test := range []struct {
		method, path, body string
		status             int
	}{
		{"PUT", "/users/" + id + "/level-rules", `{"rule_ids":["` + id + `","22222222-2222-4222-8222-222222222222"]}`, http.StatusBadRequest},
		{"PUT", "/rules/" + id, `{"enabled":false}`, http.StatusConflict},
		{"DELETE", "/rules/" + id, "", http.StatusConflict},
		{"GET", "/rules/" + id + "/members?search=alice", "", http.StatusOK},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)
		require.Equal(t, test.status, w.Code, w.Body.String())
	}
	require.Equal(t, "alice", repo.search)
}
