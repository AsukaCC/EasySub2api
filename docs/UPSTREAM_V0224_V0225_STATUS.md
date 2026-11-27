# 上游 v0.2.24 / v0.2.25：网关选择性适配

本地基线：EasySub2api v0.0.89（`385a7400a`）。本轮只参考 [v0.2.24](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.24) 和 [v0.2.25](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.25) 的增量，并非完整合并。

| 路径 | 变更及边界 |
| --- | --- |
| Embeddings、Alpha Search、System One | 六个文本路由（含兼容别名）在组合路由读取之前使用已有的文本上限。压缩请求同时限制传输及解压后的字节数；复用已读缓冲区时校验大小而不拷贝。媒体路由不启用文本上限。 |
| Alpha Search | OpenAI OAuth 文献检索明确返回 `Literature search is unavailable` 的 403 时，不记账号故障或惩罚；继续按原预算换号。PAT 的 Responses 搜索回退使用同一判断。其他 403 不一概豁免。 |
| OpenAI HTTP 透传 | 密文失效的 400 和识别到的上下文超限响应保留可供客户端采取行动的错误码。使用本地文案，仅放行已知协议字段路径作为 param。不开放原始错误体或响应头。 |
| Compact 流式桥接 | 已提交的流仍以 response.failed 结束；容量降载使用本地已有的可重试客户端码，不新增网关重试。运维记录保留错误分类和原本应返回的 HTTP 状态。 |

没有数据库迁移，不追溯历史账单，不改钱包、预占退款、分组模型映射、价格或会话亲和。本轮不包含重复字段统一校验、OAuth 凭据并发轮换修正、身份请求头模板或上游分组自动同步；不能将本记录视为这两版全部功能或安全修复已上线。

定向验证（在 backend 目录）：

```sh
go test -tags unit ./internal/pkg/httputil
go test -tags unit ./internal/server/routes -run 'TestGatewayRoutes'
go test -tags unit ./internal/service -run 'Test(AlphaSearchLiterature|ForwardAlphaSearchUnauthorized|ForwardAlphaSearchPATResponsesFallbackUnauthorized|ForwardAlphaSearchReturnsFailover|ShouldApplyOpenAIAlphaSearch|OpenAIPassthroughRecovery|CompactFailureReports|OpenAIGatewayService_APIKeyPassthrough_(ContextWindow502|CompactError)|NativeAnthropicCacheBilling|OpenAICacheWriteTTL)'
```

扩大路由回归时，`TestGatewayRoutesKeyBillingInfoEndToEnd/standard_mode` 返回 500（期望 200）；在 `385a7400a` 的未修改 Go 源码上单独复跑，结果相同。该计费信息接口不在本轮文本路由变更范围内，没有通过修改预期来消除失败。

服务扩大回归另有一项失败：`TestForwardAlphaSearchPATResponsesFallbackUnauthorizedDoesNotMarkAccountError` 仍期望旧 `OpenAI-Beta: responses=experimental`，实际为空。历史适配记录已标记此类头部断言，本轮未改该请求的构建器或测试预期。新增的 OAuth / PAT 文献检索 403 回归与此用例独立。
