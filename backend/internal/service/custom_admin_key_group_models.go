package service

import (
	"context"
	"strings"

	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/pkg/pagination"
)

// ScopedAdminUpdateGroupModels 仅更新模型展示配置；仓储在锁内重新验证分组和平台。
func (s *adminServiceImpl) ScopedAdminUpdateGroupModels(ctx context.Context, id string, cfg GroupModelsListConfig) (*Group, error) {
	if len(cfg.Models) > 1024 {
		return nil, infraerrors.BadRequest("SCOPED_GROUP_MODELS_INVALID", "模型数量超过限制")
	}
	for _, model := range cfg.Models {
		if strings.TrimSpace(model) == "" || len(model) > 512 {
			return nil, infraerrors.BadRequest("SCOPED_GROUP_MODELS_INVALID", "模型 ID 格式无效")
		}
	}
	repo, ok := s.groupRepo.(interface {
		ScopedAdminUpdateGroupModels(context.Context, string, GroupModelsListConfig) (*Group, error)
	})
	if !ok {
		return nil, infraerrors.ServiceUnavailable("SCOPED_GROUP_MODELS_UNAVAILABLE", "受限分组模型服务不可用")
	}
	normalized := normalizeGroupModelsListConfig(cfg)
	updated, err := repo.ScopedAdminUpdateGroupModels(ctx, id, normalized)
	if err != nil {
		return nil, err
	}
	// 仓储返回时事务已提交，避免回滚后缓存已提前失效。
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByGroupID(ctx, id)
	}
	return updated, nil
}

func (s *adminServiceImpl) ScopedAdminListGroupModels(ctx context.Context, params pagination.PaginationParams, all bool, platform, status, search string) ([]Group, int64, error) {
	repo, ok := s.groupRepo.(interface {
		ScopedAdminListGroupModels(context.Context, pagination.PaginationParams, bool, string, string, string) ([]Group, int64, error)
	})
	if !ok {
		return nil, 0, infraerrors.ServiceUnavailable("SCOPED_GROUP_MODELS_UNAVAILABLE", "受限分组模型服务不可用")
	}
	return repo.ScopedAdminListGroupModels(ctx, params, all, platform, status, search)
}
