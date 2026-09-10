package service

import (
	"context"
	"errors"
	"time"
)

const defaultUsageGuideContent = `# 使用说明

## 1. 创建 API Key

进入“API 密钥”页面，选择一个可用分组并创建密钥。每个 API Key 绑定一个分组，实际可用模型取决于该分组、账号权限和订阅状态。

## 2. 调用接口

使用创建的 API Key 作为 Bearer Token 调用兼容 OpenAI 的接口，并在请求体中填写模型名称。具体接口地址和示例请以站点提供的 API 地址为准。

## 3. 计费说明

请求会按照实际模型用量、渠道价格和分组倍率扣除积分。余额不足、订阅失效或分组不可用时，请求可能会被拒绝。

## 4. 常见问题

- 修改 API Key 的分组后，新的请求会按新分组进行鉴权和计费。
- 模型是否显示在模型广场，不代表当前账号一定具备调用权限。
- 如遇到请求失败，请先检查 API Key、分组状态、订阅和账户余额。
`

const MaxUsageGuideContentBytes = 1 << 20

type UsageGuide struct {
	ContentMD string
	UpdatedAt time.Time
}

func (s *SettingService) IsUsageGuideEnabled(ctx context.Context) bool {
	if s == nil || s.settingRepo == nil {
		return false
	}
	value, err := s.settingRepo.GetValue(ctx, SettingKeyUsageGuideEnabled)
	return err == nil && value == "true"
}

func (s *SettingService) GetUsageGuide(ctx context.Context) (*UsageGuide, error) {
	if s == nil || s.settingRepo == nil {
		return nil, ErrSettingNotFound
	}
	item, err := s.settingRepo.Get(ctx, SettingKeyUsageGuideContent)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return &UsageGuide{ContentMD: defaultUsageGuideContent}, nil
		}
		return nil, err
	}
	return &UsageGuide{ContentMD: item.Value, UpdatedAt: item.UpdatedAt}, nil
}
