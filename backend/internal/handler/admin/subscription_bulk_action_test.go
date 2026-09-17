//go:build unit

package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func bulkActionHandlerRequest(router http.Handler, path, body, key string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestSubscriptionBulkAction_RejectsInvalidRequestsBeforeExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service.SetDefaultIdempotencyCoordinator(service.NewIdempotencyCoordinator(storeUnavailableRepoStub{}, service.DefaultIdempotencyConfig()))
	t.Cleanup(func() { service.SetDefaultIdempotencyCoordinator(nil) })
	router := gin.New()
	path := "/api/v1/admin/subscriptions/bulk-action"
	router.POST(path, NewSubscriptionHandler(nil).BulkAction)
	tooManyIDs := make([]string, 101)
	for i := range tooManyIDs {
		tooManyIDs[i] = "sub-1"
	}
	tooManyBody, err := json.Marshal(service.BulkSubscriptionActionInput{SubscriptionIDs: tooManyIDs, Action: "revoke"})
	require.NoError(t, err)
	for _, body := range []string{
		`{"subscription_ids":["sub-1"],"action":"delete"}`,
		`{"subscription_ids":["sub-1",""],"action":"revoke"}`,
		`{"subscription_ids":[],"action":"restore"}`,
		`{"subscription_ids":["sub-1"],"action":"extend","days":0}`,
		`{"subscription_ids":["sub-1"],"action":"extend","days":36501}`,
		`{"subscription_ids":["sub-1"],"action":"reset_quota"}`,
		`{"subscription_ids":["sub-1"],"action":"reset_quota","daily":"yes"}`,
		`{"subscription_ids":["sub-1"],"action":"revoke"`,
		string(tooManyBody),
	} {
		t.Run(body, func(t *testing.T) {
			response := bulkActionHandlerRequest(router, path, body, "bulk-invalid")
			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
		})
	}

	response := bulkActionHandlerRequest(router, path, `{"subscription_ids":["sub-1"],"action":"revoke"}`, "bulk-valid")
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestSubscriptionBulkAssign_RejectsInvalidUserIDsBeforeExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	path := "/api/v1/admin/subscriptions/bulk-assign"
	router.POST(path, NewSubscriptionHandler(nil).BulkAssign)
	for _, ids := range [][]string{{}, {"user-1", ""}, make([]string, 101)} {
		if len(ids) == 101 {
			for i := range ids {
				ids[i] = "user-1"
			}
		}
		t.Run(fmt.Sprint(len(ids), ids), func(t *testing.T) {
			body, err := json.Marshal(BulkAssignSubscriptionRequest{UserIDs: ids, GroupID: "group-1", ValidityDays: 30})
			require.NoError(t, err)
			response := bulkActionHandlerRequest(router, path, string(body), "")
			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
		})
	}
}
