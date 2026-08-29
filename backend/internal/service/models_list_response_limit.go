package service

import "github.com/AsukaCC/EasySub2api/internal/config"

func resolveModelsListReadLimit(cfg *config.Config) int64 {
	if cfg != nil && cfg.Gateway.ModelsListReadMaxBytes > 0 {
		return cfg.Gateway.ModelsListReadMaxBytes
	}
	return config.DefaultModelsListReadMaxBytes
}
