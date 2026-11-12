# Sub2API v0.2.20 / v0.2.21 适配记录

基线：EasySub2api 0.0.81，`1f91b636c`。发布目标：0.0.82。
参考：[v0.2.20](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.20)、[v0.2.21](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.21)。

## 业务约束

继续使用 UUID、积分钱包、请求预占/退款/幂等结算，以及分组/渠道/账号显式价格优先规则。显式零价保留；充值倍率固定 1，不引入上游汇率倍率。

会话最近一次成功调度后续期 30 分钟，窗口内粘性优先；其余情况保留 Key 亲和、现有调度因子和空闲公平性。请求结束释放占用，同 Key 并发不去重。生图尺寸约束与此共用，不覆盖原有溢出绑定。

保留归因测试队列、取消/超时、空闲/可用/额度条件、每请求计费、7 天记录及账号 Codex 积分快照。账号凭证/2FA 沿用权限隔离。默认模型仍为 gpt-5.6-sol，停用模型规则继续生效；models_list_config 仅控制展示。

## 适配条目

| 条目 | 本地处理 |
| --- | --- |
| sealed-calls | 保留 namespace、encrypted_function_args 和工具调用生命周期字段。 |
| nonstream-responses | 识别缺少 Content-Type 的 SSE，重建 incomplete/cancelled 及完整正文。 |
| metadata-failover | Codex 元数据不提前提交响应，保留后续过载切换能力。 |
| selection-slot-release | 账号加载失败即释放并发槽位，沿用本地 Key 占用跟踪。 |
| refresh-token-consumption | 刷新令牌原子消费，MCP 刷新与撤销同样防止并发重放。 |
| request-errors | 分类请求读取失败，非法 UTF-8 返回 400，修正 Key/Grok Voice 错误。 |
| body-memory-budget | 自适应进程内存预算，修正单次补丁和异步生图预算生命周期。 |
| bounded-io | 限制停滞读写、Bedrock 帧、Gemini SSE 和 Antigravity 非流式读取。 |
| bounded-background | 限制 drain、后台任务、SSE 队列及会话缓存；保留本地结算。 |
| hotpath | 减少重复 JSON 解析，使用已有鉴权投影处理 RPM 与隐私门控。 |
| security-dependencies | 更新 brotli/compress，新增 MCP SDK 依赖；Go 1.27.4 修复 GO-2026-6859。 |
| websocket | 后续轮次失败关闭、搜索激活、HTTP 桥接指纹、逐轮 TTFT 配置。 |
| stable-session-seeds | 工具回合及追加 system/developer 消息不改变无显式 ID 的会话种子。 |
| reasoning-policy | 合成 effort 应用映射和上限，保留 GLM 降级并处理外部 reasoning 封装。 |
| account-error-policies | 显式模型冷却覆盖 overload/slow_down/流内 503，匹配实际错误字段。 |
| openai-request-compat | 空 instructions 透传和原值回显、Desktop 上下文、强制路由探测及 413 策略。 |
| openai-images | 更新 GPT Image 原生参数、测试参数和真实 HTTP 错误状态。 |
| protocol-content | 保留缓冲 reasoning/消息字段，规范化 Grok 委派消息及 Gemini 工具媒体。 |
| gemini-antigravity-compat | 思考和正文分离、直连流式、JSON schema、并行工具图片和身份前导。 |
| claude-models | Fable/Mythos 5.2 目录、Bedrock、effort 和兜底定价；CLI 基线更新。 |
| openai-cache-ttl | 30m/12h 贯穿用量、数据库、价卡、结算、统计及页面；显式零价有效。 |
| pricing-fixes | 修正 Terra 订阅附加费、Sonnet 缓存价、免费输入 Fast 倍率及目录哈希重试。 |
| image-size-routing | 按 1K/2K/4K 分流；30 分钟粘性可排队，溢出不覆盖原绑定。 |
| transport-cooldown-switch | 新增关闭 OpenAI 传输错误冷却开关，默认保持启用。 |
| antigravity-quota | 真实 5h/周窗口、独立 Argon 冷却及强制刷新。 |
| antigravity-500 | 图像模型单独冷却，支持配置惩罚档位。 |
| connection-test-prompts | Antigravity 使用管理员自定义提示词并返回实际回复。 |
| setup-token-usage | 无 5h 窗口仍保留采样周用量，不主动调用 profile API。 |
| kimi-cooldown | 并发 403 短冷却，保留余额和配额保护。 |
| cloudflare-beta | Cloudflare REST messages 不携带 beta 查询。 |
| admin-permissions | 补全统计、用量、OAuth、邀请等权限，保留 UUID 资源和凭证/2FA 隔离。 |
| admin-full-access | 显式确认后才能启用；保留旧密钥受限、撤销、审计与幂等隔离。 |
| email-blacklist | 上限 20000，按邮箱身份规范化，批量禁用需明确选择。 |
| management-mcp | 默认关闭，显式 HTTPS URL（允许 localhost HTTP），接入本地权限和 OAuth 撤销。 |
| keys-list-performance | 先渲染列表再加载统计，新增并发覆盖索引。 |
| account-refresh-ui | 正常刷新保留行、配额和积分快照。 |
| usage-export | 1000 行分页、Retry-After 重试及本地积分列/CSV 保护。 |
| audit-wakeup | 入队唤醒本进程 worker，保留数据库认领。 |
| client-export | 保留搜索、生图配置；默认模型仍为 gpt-5.6-sol。 |
| settings-save-ui | 保留本地已有吸附保存栏。 |
| easypay-refund-code | 每支付实例选择 0/1 成功码，沿用钱包退款。 |
| translations | 补齐新增页面、权限、模型、配额及错误的中英文。 |
| migrations-and-docker | 本地迁移编号 288–290，隔离 PostgreSQL/Redis 集成及应用验证。 |

## 差异处理

- 上游 cyber transcript 长文本误拦截修复不直接移植：本地仅使用明确会话 ID 封禁，没有内容哈希/长 transcript 识别路径。
- 上游默认 GPT-6.1 Sol 和充值倍率精度变更不覆盖本地业务设定。
- 本地已有浮动保存栏和 WebSocket TTFT predicate，复用实际调用路径，避免添加重复配置。
- 修复空数据库管理员初始化仍写入旧 balance 列的问题，改为 recharge_balance；四个钱包桶按现有默认值初始化。
- PostgreSQL UUID upsert 的 RETURNING 兼容驱动返回 []byte；通过 Ent 自定义生成模板固化，后续重新生成不会丢失。
- MCP 默认关闭，开启需显式 MCP_PUBLIC_URL。OAuth 授权不能绕过既有凭证权限或管理写操作二次验证。

## 迁移及配置

迁移 288（MCP OAuth）、289（缓存 TTL 字段）、290（用量覆盖索引）接在本地历史迁移之后。290 使用并发建索引，大表首次启动可能较慢。测试只使用临时 Docker 数据库；没有修改现有业务数据库。

请求体预算键为 gateway.request_body_memory_budget_bytes；未配置时依据进程内存上限自动计算。读写超时及预算可按部署资源配置，详见 deploy/config.example.yaml 与 deploy/EDGE_SECURITY.md。

## 验证

- 前端 lint、vue-tsc、仓库 CI 的鉴权/API、保护/计费、钱包/归因测试和管理员组件回归通过；新增 14 个翻译/页面测试文件共 64 项通过。
- 后端生产构建、所有 unit 测试编译、网关/管理员、保护/停用模型、0.2.19/0.2.20/0.2.21 兼容及 Key/会话调度回归通过；钱包/归因/计费与历史业务兼容回归通过。
- httputil、Antigravity、apicompat、config、managementmcp、WebSocket relay、setup 完整包测试通过。
- golangci-lint v2.14.0（与 CI 相同，tests=false）通过，0 issues。
- Docker 最终镜像构建通过；临时空库自动执行全部迁移并创建管理员成功。嵌入前端、登录、邮箱黑名单、UUID 分组及显式免费 12h 价格、全权密钥确认/撤销、MCP 默认关闭/开启/PKCE 授权/工具列表/即时撤销均通过。
- Docker PostgreSQL/Redis 仓储集成通过：钱包、旧版适配、刷新竞争、30m/12h 用量单条/批量/渠道持久化、MCP 并发 code/refresh/replay/账号撤销与多实例密钥收敛。
- 部署 shell 生命周期、Compose 安全、运行资源及 Caddy 缓存/流式检查通过。

可复跑：backend/scripts/test-release0221-integration.py；构建镜像后运行 backend/scripts/test-release0221-smoke.py。GitHub CI/Security Scan 将在提交后再次执行，成功后才创建发布 tag。

仓库全量历史测试仍包含旧数字 ID 和旧余额模型夹具；本次不将它们当作通过，也不关闭现有 CI。数据库回归选择当前 UUID/积分钱包夹具，正常应用迁移另由 Docker smoke 验证。
