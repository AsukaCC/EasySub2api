# Sub2API v0.2.19 适配记录

参考：[上游 v0.2.19 发布说明](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.19)。

基线：EasySub2api v0.0.75，提交 `0d6c94b07f8854750a178326a2ad27d7d9d95788`。
适配分支：`codex/sync-upstream-v0.2.19`；核验日期：2026-10-29。

本次按 v0.2.19 发布说明核对新增功能、优化和修复，结合上游 v0.2.18 → v0.2.19 的实现进行适配。
这份记录不代表补齐 [v0.2.13–v0.2.17 记录](UPSTREAM_V0217_STATUS.md) 中仍标记 pending/partial 的历史功能。
发布目标：EasySub2api v0.0.76。此记录不代表已在现有业务数据库执行迁移。

## 现有业务约束

- 实体 ID 和关联继续使用 UUID/string；管理密钥的创建人外键引用 `users.id UUID`，不引入上游数字 ID。
- 用户充值积分、赠送积分、退款冻结/扣回、请求预占及幂等结算继续使用本地钱包体系。
  受限管理密钥的余额写入也调用现有 `UpdateUserWalletBalance`。
- 保留分组模型定价、多分组选择、渠道定价和分组媒体价格的本地优先级；新增图片开关不改写历史账单。
- 分组 `models_list_config` 继续只控制模型列表展示，不转为执行白名单。
  管理密钥的分组模型权限对应这一展示配置，账户模型映射仍是独立权限。
- 保留 GPT-5.4/5.5 的全局停用规则，不重新开放这些模型；默认模型仍为本地 `gpt-5.6-sol`。
- 保留 Grok 未设置显式映射时的直通行为、raw Chat 的上游原始 `model`、本地 Chat 桥接适用范围。
  模型目录诊断仅提供信息，不自动修改映射或模型可用性。

## 新增功能（5 项）

| 上游条目 | 本地结果 | 主要验证 |
| --- | --- | --- |
| 按操作与资源授权的管理员 API Key | 已适配。独立摘要存储；操作目录明确登记，未登记路由默认拒绝；支持分组、账户类型、平台、过期和 IP/CIDR。创建、编辑、轮换、撤销与隐藏记录接入设置页面和管理员 JWT 二次验证。 | `custom_admin_keys_test.go`、`adminaccess`、中间件测试、`release0219_scoped_admin_test.go`、设置组件测试 |
| OpenAI / Composite 按实际图片张数计费 | 已适配。默认关闭，仅在允许生图的 OpenAI / Composite 分组及专用图片端点生效。优先实际出图数，启用后无法识别时回退 1 张；关闭保留原回退规则。 | `image_billing_count_test.go`、`image_count_billing_switch_test.go`、`gateway_usage_billing` 回归 |
| GPT-6.1 Terra | 已适配模型识别、别名、Codex manifest、Responses 约束、协议转换、价格和前端目录。保留本地默认模型与停用规则。 | `openai_gpt61_terra_test.go`、现有 GPT61 / retired-model 回归 |
| Gemini 同账号重试配置 | 已适配 `gateway.max_same_account_retries_gemini`，默认 4，范围 0–10；Native / Messages / Chat 共用预算。签名兼容降级不消耗普通重试次数。 | `gemini_same_account_retries_test.go`、配置校验和单元测试编译 |
| Grok 账号模型目录诊断 | 已适配来源、时效、可信状态、映射差异和手动同步；后台与手动同步共用解析和 scope 规则。前端保护未保存草稿及过期异步结果。 | `grok_model_catalog*`、`grok_model_diagnostics_test.go`、`GrokModelDiagnosticsPanel.spec.ts` |

管理员密钥的补充保护：

- 每次认证从数据库读取，创建人必须仍是有效管理员；轮换、撤销、策略变更会使旧认证状态失效。
- 账户与关联查询在同一事务中执行，提交后才发布缓存副作用；受限写入拒绝会级联影响 shadow 账户的操作。
- 敏感凭证、用户财务、分组、订阅、用量、代理等响应按权限投影，幂等重放也执行投影。
- 幂等存储按密钥及策略隔离；各密钥独立限流，写入请求不会重复消耗严格档额度。
- 删除列表记录只隐藏已撤销密钥，保留原记录和历史审计。

## 优化改进（6 项）

| 上游条目 | 本地结果 | 主要验证 |
| --- | --- | --- |
| Gemini 默认目录移除 2.0/2.5 与 3-pro-preview，测试模型改为 3.5-flash | 已更新后端、前端默认目录及连接测试模型；账户显式映射仍可声明旧模型。 | Gemini / GeminiCLI 模型测试、`useModelWhitelist` 回归 |
| Gemini 组内账号均有映射时，原生列表与详情按声明返回 | 已适配 `/v1beta/models` 列表和详情；保留本地展示配置语义。 | `TestRelease0219GeminiDeclaredModels` |
| 绑定显式缓存的池模式请求在原账号重试 | 当前不适用。本地没有显式缓存资源、绑定账户和预占结算子系统，无法仅移植该重试分支。本次未新建该历史功能；普通 Gemini 三条转发路径已接入新重试配置。 | 对照本地路由/服务与 v0.2.17 的显式缓存 pending 记录 |
| 国产 Coding Plan 频率型 429 改为秒级冷却 | 已适配频率/额度分类；明确额度耗尽或快照接近耗尽仍保留额度窗口保护。 | `ratelimit_cn_frequency_429_test.go` |
| 并发冷启动不重复获取上游模型列表 | 已补 singleflight 内的缓存二次检查，包含 Codex API Key manifest；保留本地已有的目录缓存复查。 | `TestRelease0219ModelsRefreshReusesFreshResult` |
| Grok 各协议记录上游原始响应模型 | 已覆盖原生 Responses、raw Chat、Chat→Responses、Messages、流式及聚合路径；失败重试不遗留前一次观测。 | `grok_response_model_audit_test.go` |

## Bug 修复（12 项）

| 上游条目 | 本地结果 | 主要验证 |
| --- | --- | --- |
| Gemini `fileData` 图片未计费 | 已修复流式与非流式图片识别，累计响应不会重复计数。 | `gemini_image_output_accounting_test.go` |
| 渠道按次图片档位丢失 token 用量 | 已将 token 用量传入按次/图片计费输入，保留上下文分档。 | `TestCalculateRecordUsageCost_ChannelPerRequestContextTierUsesTokens` |
| DeepSeek V4 Pro 9 月 14 日后价格失效 | 本地已有覆盖该日期之后的价格实现，已保留并核对；无需覆盖本地价格策略。 | 本地 DeepSeek 定价实现与上游差异核对 |
| DeepSeek `prompt_cache_hit_tokens` 未计缓存读取 | 已补用量解析回退。 | `openai_usage_deepseek_cache_test.go` |
| OpenCode / DeepSeek 的 `developer` 角色触发 422 | 已在适用目标转换为 `system`，保留其他平台请求语义。 | `openai_chat_roles_test.go` |
| Antigravity 非流式畸形工具调用返回空成功 | 已拒绝无法转换的畸形工具调用。 | `response_transformer_malformed_test.go` |
| 超大 multipart 图片被静默截断 | OpenAI 超过 50 MiB 的单个 part 直接拒绝；有效大图片完整转发，Grok 也移除静默截断。 | `openai_images_upload_limit_test.go`，包含 21 MiB 有效上传及超限回归 |
| 分页溢出导致 500 | 已限制安全页码并处理非法/超大数字，`offset` 与下一页边界不溢出。 | pagination / response 边界测试 |
| 管理员保护规则校验返回 500 | 保护规则输入错误映射为 400，保留本地保护规则。 | 现有 `TestProtection*` 回归 |
| 保存渠道覆盖其他平台的 `web_search_emulation` | 已按平台维度合并；前端多平台编辑保留未编辑维度。 | `channel_websearch_merge_test.go`、`ChannelsView.webSearchEmulation.spec.ts` |
| 模型广场分组模型价格被图片兜底覆盖 | 已优先分组模型定价，再使用图片兜底价格。 | `TestRelease0219PlazaGroupModelPricingWinsOverImageFallback` |
| Grok 目录被失败、乱序或旧表单覆盖 | 已使用条件更新和单一解析规则，失败保留有效成功快照；旧表单不能写入受管目录字段。本地额外的目录/能力缓存也按观测快照条件写入。 | Grok catalog/store/sync、repository CAS、managed-extra 回归 |

另外同步 Claude Code 指纹的 JavaScript UTF-16 采样语义，并用 Node 参考值验证；
`golang.org/x/net` 更新至 v0.62.0。Ent 与 Wire 已重新生成。

## 数据库和配置

上游迁移编号 247 / 248 已按本地序列调整为：

| 本地迁移 | 内容 | 默认行为 |
| --- | --- | --- |
| `285_scoped_admin_api_keys.sql` | 新建受限管理密钥表；UUID 创建人外键；摘要、权限、期限、IP、撤销及隐藏记录 | 不会创建或分发新密钥 |
| `286_group_image_count_billing.sql` | 增加 `groups.image_count_billing NOT NULL` | `false`，旧分组保持原计费行为 |

沿用应用启动时自动执行迁移的机制。没有在现有业务数据库执行这些迁移。
图片开关可在分组编辑页启用；切换到不支持的平台或关闭生图时自动清除。

Gemini 配置示例已写入 `deploy/config.example.yaml`：

```yaml
gateway:
  max_same_account_retries_gemini: 4
```

如需继续显示已从 Gemini 默认目录移除的模型，使用现有账户模型白名单或映射进行声明。

## 本地账号调度顺序调整

按现有业务调度规则先确定候选池，再在同等调度条件下应用对话粘性：

- 通用网关和模型路由不再提前返回粘性账号。先比较账号优先级、适用的重置窗口与负载，粘性只在相同调度层内优先于 LRU / 随机选择。
- OpenAI 普通调度先比较 compact 能力、启用的低倍率规则、账号优先级和负载。批量负载读取失败、槽位竞争和排队回退也遵守对应候选顺序。
- OpenAI 高级调度先执行订阅池分层、基础分和 Top K，再执行既有加权选择。粘性权值不再加入基础分，只能在基础分、优先级、负载和排队数相同的候选之间决定亲和顺序。
- 可迁移的 `previous_response_id` 同样后置；必须依赖原账号上游状态的续接请求继续以绑定账号作为协议约束。
- Gemini Messages 兼容入口采用相同的先调度、后粘性规则。
- 保留窗口 / RPM 粘性豁免、会话数量限制、利润终检、失效绑定清理以及临时溢出不覆盖长期绑定的行为。
- 管理页面的粘性分改为显示独立权值，不再显示无限优先级；中英文说明同步更新。

CI 增加七组调度专项，覆盖优先级、负载、重置窗口、成本与运行质量、订阅池、同条件粘性、排队候选范围及 Gemini 兼容入口。

## 验证结果

已通过：

- Go 1.27.2 `go build ./...`，以及所有 `unit` 测试文件的编译检查。
- 本次精确选取的 165 个后端专项测试；新增 v0.2.19 CI 测试步骤也已本地执行通过。
- 调度顺序专项、粘性逃逸、续接协议和订阅优先回归：83 个通过事件（含子测试），无失败；新增七组调度测试已纳入 CI。
- 现有 CI 的网关/管理、账户保护/停用模型、钱包/退款/支付/指纹/上游兼容三组回归。
- golangci-lint v2.14.0：0 issues。
- 前端 lint、typecheck、生产构建；常规大 chunk 提示仍存在。
- 前端 `test:ci`：9 个文件、125 个测试通过。
- 前端业务保护、钱包/计费与新功能合并回归：34 个文件、339 个测试通过。
- `git diff --check`。

数据库集成验证使用独立 PostgreSQL / Redis 容器，入口：

```bash
cd backend
python3 scripts/test-release0219-integration.py
```

此入口通过 Go source-file 模式选择当前生产代码、现有容器与迁移 harness、UUID 夹具、钱包退款测试和本次新增数据库测试。
它避开旧集成测试中尚未适配 UUID 的数字 ID 夹具；Docker 不可用会失败，不会静默报告成功。
PostgreSQL 18.1 / Redis 8.4 上，完整迁移链和前 5 项测试已通过；
在项目使用的 PostgreSQL 16 / Redis 8.4 上，最终全部 6 项测试通过：

- `TestPermanentPointsRefundAndLedger`
- `TestPermanentPointsConcurrentRefundAndConsumption`
- `TestRelease0219AdminKeyLifecycleWithUUIDCreator`
- `TestRelease0219ImageCountBillingMigrationAndRepository`
- `TestRelease0219GrokCatalogConditionalWritesAndRollback`
- `TestRelease0219ScopedAccountQueriesKeepRulesSeparate`

数据库回归额外验证受限账户查询的规则不会交叉拼接、凭证字段投影，以及 Codex 导入候选的 UUID 范围查询。
Grok 回滚测试同时验证原生条件 SQL 和普通 ORM 账户编辑均参与调用者事务。

CI 已增加该数据库回归入口，以及后端 v0.2.19 专项和前端三个新增管理功能测试文件。

全量测试的已知限制：

- v0.2.19 适配阶段的前端全量 Vitest：原始基线有 181 个失败，适配后为 180 个失败、1972 个通过；对照失败用例无新增失败。
  主要遗留项涉及旧数字 ID 和 Pinia/i18n 测试夹具。这不等于全量测试通过。
- 调度扩大回归有 337 个通过事件、11 项失败；与调整前调度基线及原始代码成本测试对照，无新增失败。
  其中两项上游倍率/稀疏信号断言已在原始代码单独复现，不为满足旧断言而修改当前业务倍率规则。
  本次前端账号分数页面与 i18n 专项为 5 项通过、1 项旧失败，旧失败同样已在原始全量基线确认。
- 后端没有声称全量 unit 行为测试通过。宽泛筛选会触发原有的 Grok 取消测试挂起；
  两个 OpenAI long-context malformed-input 的错误 reason 断言已在原始基线复现失败。
- 配置包完整测试中 `TestConfigKeysAreEnvReachable` 失败，原始基线同样缺少 8 个旧配置键的默认注册；
  本次新增的 Gemini 重试配置已注册默认值，没有增加该测试的失败项。
- 全量 `integration` 包仍有旧数字 ID / scheduler mock 签名导致的编译错误；本次没有为适配这些旧夹具而改动业务 ID。
- 没有向真实上游付费接口发起验收请求，未执行线上迁移、部署或浏览器完整端到端验收。
