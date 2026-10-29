package admin

import (
	"encoding/json"
	"slices"

	"github.com/AsukaCC/EasySub2api/internal/handler/dto"
	"github.com/AsukaCC/EasySub2api/internal/pkg/adminaccess"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func scopedAdminWriteOnly(c *gin.Context, readPermission string) bool {
	p := adminaccess.FromContext(c.Request.Context())
	return p != nil && !slices.Contains(p.Policy.GlobalPermissions, readPermission)
}

func adminSubscriptionGrantResponse(c *gin.Context, grant *service.SubscriptionGrantResult) any {
	if adminaccess.FromContext(c.Request.Context()) == nil || grant == nil {
		return grant
	}
	out := gin.H{"activation_status": grant.ActivationStatus}
	if grant.Subscription != nil {
		out["subscription"] = adminSubscriptionResponse(c, dto.UserSubscriptionFromServiceAdmin(grant.Subscription))
	}
	if pending := grant.PendingSubscription; pending != nil {
		record := gin.H{"id": pending.ID}
		if !scopedAdminWriteOnly(c, "subscriptions.read") {
			record["user_id"], record["group_id"], record["platform"] = pending.UserID, pending.GroupID, pending.Platform
			record["status"], record["validity_days"] = pending.Status, pending.ValidityDays
			record["expected_activation_at"] = pending.ExpectedActivationAt
		}
		out["pending_subscription"] = record
	}
	return out
}

// adminSubscriptionResponse 在缓存重放后也应用字段白名单，不嵌入用户、分配人或分组配置。
// 写权限只确认订阅 ID；订阅用量等字段必须另有读取权限。
func adminSubscriptionResponse(c *gin.Context, data any) any {
	if adminaccess.FromContext(c.Request.Context()) == nil {
		return data
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return projectSubscriptionJSON(raw, !scopedAdminWriteOnly(c, "subscriptions.read"))
}

func projectSubscriptionJSON(raw json.RawMessage, readable bool) any {
	var list []json.RawMessage
	if len(raw) > 0 && raw[0] == '[' && json.Unmarshal(raw, &list) == nil {
		out := make([]any, 0, len(list))
		for _, item := range list {
			out = append(out, projectSubscriptionJSON(item, readable))
		}
		return out
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil
	}
	out := make(map[string]any)
	if subscriptions, ok := fields["subscriptions"]; ok {
		for _, key := range []string{"success_count", "created_count", "reused_count", "failed_count", "errors", "statuses"} {
			if value, exists := fields[key]; exists {
				out[key] = value
			}
		}
		out["subscriptions"] = projectSubscriptionJSON(subscriptions, readable)
		return out
	}
	keys := []string{"id"}
	if readable {
		keys = append(keys, "user_id", "group_id", "starts_at", "expires_at", "status",
			"daily_window_start", "weekly_window_start", "monthly_window_start",
			"daily_usage_usd", "weekly_usage_usd", "monthly_usage_usd",
			"created_at", "updated_at", "revoked_at", "assigned_by", "assigned_at", "notes")
	}
	for _, key := range keys {
		if value, exists := fields[key]; exists {
			out[key] = value
		}
	}
	return out
}
