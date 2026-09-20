//go:build unit

package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/handler/dto"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUpdateSettingsAffiliateTransferValidity(t *testing.T) {
	for _, days := range []int{1, 30, 3650} {
		t.Run(strconv.Itoa(days), func(t *testing.T) {
			h, repo := newStepUpSwitchTestHandler(t, map[string]string{
				service.SettingKeyBonusBalanceDefaultValidityDays: "180",
			})
			rec := doUpdateSettings(t, h, map[string]any{"affiliate_transfer_validity_days": days}, nil)
			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, strconv.Itoa(days), repo.values[service.SettingKeyAffiliateTransferValidityDays])
			require.Equal(t, "180", repo.values[service.SettingKeyBonusBalanceDefaultValidityDays])
			var result struct {
				Data dto.SystemSettings `json:"data"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
			require.Equal(t, days, result.Data.AffiliateTransferValidityDays)
		})
	}
}

func TestUpdateSettingsAffiliateTransferValidityRejectsInvalid(t *testing.T) {
	for _, value := range []any{0, -1, 3651, 1.5, "30"} {
		h, repo := newStepUpSwitchTestHandler(t, map[string]string{
			service.SettingKeyAffiliateTransferValidityDays: "45",
		})
		rec := doUpdateSettings(t, h, map[string]any{"affiliate_transfer_validity_days": value}, nil)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Equal(t, "45", repo.values[service.SettingKeyAffiliateTransferValidityDays])
		require.Empty(t, repo.lastUpdates)
	}
}

func TestUpdateSettingsAffiliateTransferValidityPreservesOmitted(t *testing.T) {
	for _, stored := range []string{"", "45"} {
		h, repo := newStepUpSwitchTestHandler(t, map[string]string{
			service.SettingKeyAffiliateTransferValidityDays:   stored,
			service.SettingKeyBonusBalanceDefaultValidityDays: "180",
		})
		rec := doUpdateSettings(t, h, map[string]any{"affiliate_rebate_rate": 10}, nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, stored, repo.values[service.SettingKeyAffiliateTransferValidityDays])
		require.NotContains(t, repo.lastUpdates, service.SettingKeyAffiliateTransferValidityDays)
	}
}

func TestGetSettingsAffiliateTransferValidity(t *testing.T) {
	h, _ := newStepUpSwitchTestHandler(t, map[string]string{
		service.SettingKeyAffiliateTransferValidityDays:   "45",
		service.SettingKeyBonusBalanceDefaultValidityDays: "180",
	})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	h.GetSettings(c)
	require.Equal(t, http.StatusOK, rec.Code)
	var result struct {
		Data dto.SystemSettings `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	require.Equal(t, 45, result.Data.AffiliateTransferValidityDays)
}

func TestAffiliateTransferValidityAudit(t *testing.T) {
	changed := diffSettings(&service.SystemSettings{AffiliateTransferValidityDays: 90},
		&service.SystemSettings{AffiliateTransferValidityDays: 30}, nil, nil, UpdateSettingsRequest{})
	require.Contains(t, changed, service.SettingKeyAffiliateTransferValidityDays)
}
