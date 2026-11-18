# Sub2API v0.2.23 选择性适配：缓存计费

本地基线：EasySub2api v0.0.88（`79388fb9b`）。参考[上游发布说明](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.23)，本轮集中修复已有转发路径的用量，不改版本号、数据库或历史账单。

- Anthropic 普通输入不含缓存；转为 OpenAI 内部用量时先加回缓存读写，再由结算入口拆分，避免普通输入被重复扣减。
- Messages、Responses 和 Chat Completions 走原生 Anthropic 上游时，5m/1h 写入明细进入计费和日志。流式合并保留上次观测；只给明细时从累计明细推算写入总量，明确报告的正总量优先。
- 若报告 100 个写入、其中 5 个为 5m、40 个为 1h，日志保留实报的 5 和 40；结算按 60 个 5m、40 个 1h 收费。账号成本统计也使用该规则。明细超过总量时仍沿用本地按比例封顶策略。
- 保留本地 OpenAI 30m/12h 独立计费、渠道定价、分组倍率和钱包结算。不涉及客户端所有协议间的 TTL 字段互转，也未移植上游渠道 Ultra Fast 倍率、账号重试或性能改动。

在 `backend` 目录运行：

```sh
go test -tags unit ./internal/pkg/apicompat
go test -tags unit ./internal/service -run 'Test(NativeAnthropic|OpenAICacheWriteTTL|CC.*FromNativeAnthropic|Responses.*FromNativeAnthropic|CalculateCost_LongContextAppliesMultiplierToCacheCreation5mAnd1h|GatewayService_ParseSSEUsagePassthrough|ParseClaudeUsageFromResponseBody)'
go build ./cmd/server
```

上述验证通过。新增回归使用模拟上游响应和仓储，不调用付费接口。另行运行的 `TestCalculateCost_OpenAIGPT54LongContextAppliesMultiplierToCacheCreation` 在本轮和未修改基线均失败（期望 `0.05`，实际 `0.025`）；未更改该模型价格或测试预期。
