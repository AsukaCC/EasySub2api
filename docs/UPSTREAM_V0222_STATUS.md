# Sub2API v0.2.22 适配记录

本地基线：EasySub2api v0.0.84，`ea9695d8f`。
参考：[上游 v0.2.22](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.22)，比较范围为上游 v0.2.21 至 v0.2.22。

本次按当前项目的实际调用路径选择性适配，并非整版合并。保留本地品牌、页面结构、SCSS 样式和部署版本号。

## 已适配

| 范围 | 本地处理 |
| --- | --- |
| Codex 上下文 | 处理 page context、browser_info、goal_objective、untrusted_context 独立输出，保留真实配对的工具调用。 |
| 工具声明 | 保留 additional_tools 声明 ID；仅规范 UUID 形式的 at_ ID，同步对应引用。 |
| 工具 Schema | 按平台修复根 type、required:null 和明确的 object union；OpenAI 路径清理不支持的正则，处理嵌套补丁及 patternProperties。 |
| 协议转换 | 保留 encrypted function call 内容；会话中间 system 转 developer；JSON schema strict 默认 false；保留合法 metadata.user_id。 |
| Grok | 服务端搜索调用不再输出为客户端工具调用；增加搜索统计；在 Grok 路径识别扁平错误，保持共享错误解析语义。 |
| Gemini | 增加 Carbon 思考配置及目录，修正 array/items；Chat Completions 传递结构化 response_format；分离同名并行工具和分片参数。 |
| Vertex | 支持 us/eu 多区域主机、模型不支持时降级 global、已有 cachedContent 的区域路由；实际多区域请求默认 token 定价加价 10%，保留渠道/分组/账号自定义价优先级。 |
| 配额 | Kimi 新月度套餐及无效响应保护、缺席窗口清理；5h 重置保留有效周观测；读取 Fable scoped limits，恢复时按原因条件清除限流。 |
| 配额界面 | 跟随账号快照刷新，丢弃切换账号后的旧探测；未知用量不显示为 0%；主动刷新失败保留观测时间且不判定渠道故障。 |
| 分组加速 | Ultra Fast 支持 Astra / Sol，与 Fast 并存，优先级 Ultra Fast > Fast > 客户端；沿用全局过滤规则，补齐鉴权缓存和复制分组字段。 |
| 定价 | DeepSeek 显式覆盖价格（包括零价）保留；补 Gemini Carbon/Hydrogen、Qwen3 embedding/reranker 回退价；本地价格已加载时后台校验远程 hash。 |
| 批量倍率 | 分组费率/RPM 同步改为事务和去重，使用本项目 UUID。RPM override 只用于匹配分组。 |
| 网关可靠性 | 已发送 SSE keepalive 后用 SSE 返回错误；恢复组合模型公开名称；Bedrock 按错误语义处理并保留部分 usage，限流使用调度模型键。 |
| 连接及缓存 | TCP keepalive、Linux TCP_USER_TIMEOUT 设置失败不阻断连接；等待 ping 可配置为 0；首输出临时文件创建失败使用有界内存；动态 HTML 禁止缓存。 |
| 性能 | 选择性减少 Anthropic SSE、工具名和 cache_control 路径的解析/复制；转义 JSON 保留解析回退；调度投影保留 openai_capabilities。 |

## 保留的本地约束

- UUID、积分钱包、请求预占、退款和幂等结算继续使用本地实现；充值倍率仍为 1。
- 分组/渠道/账号显式定价继续优先，显式零价有效。
- 30 分钟滑动会话亲和、Key 亲和、空闲公平性和并发释放继续使用本地调度。
- 下游显式携带 image_generation 时透传，不自动添加；原生生图工具不走 Responses Lite。
- 保留本轮已有的 Codex/CC Switch 导出配置：默认主模型 gpt-6.2-sol、子代理默认跟随主模型、可编辑模型及 WebSocket 开关。

## 未直接迁移

- 上游按用户通道合并扣费：与本地积分钱包、预占和幂等边界不同，未引入。
- 上游分组汇总 SQL 和历史迁移：与本地数据库结构不同，未覆盖。
- 上游调度快照请求合并、整池成员与冷却恢复重构：本地已有独立生命周期、运行时复核和 Key 调度，未整体替换；本次仅补能力字段投影。
- Vertex 显式缓存管理接口和 batch API：本地没有对应功能，本次仅适配已有转发路径。
- 上游其他账号观测、直通白名单、原生会话哈希和完整热路径重构未宣称全部合入。已有本地实现和尚未迁移条目需要按各自行为单独核对。
- 上游版本号、赞助内容、页面布局和业务默认值不覆盖本地。

## 迁移与生成

新增迁移 `291_group_force_openai_ultrafast.sql`，默认 false，保持现有分组行为。
Ent 使用本地 UUID 自定义模板重新生成；生成命令仅加载 `templates/create.tmpl`，避免把许可证文件当作模板。

## 验证

- 前端生产构建和 ESLint 通过；相关组件、密钥导出、账号编辑等 145 项测试通过。
- Playwright 检查 1440px / 390px 分组弹窗，Ultra Fast 可切换，弹窗无横向溢出、无页面异常。测试使用模拟 API，不连接生产数据。
- 后端生产构建通过；Schema、Codex 上下文、原生生图透传、Gemini 工具/格式、Vertex、配额、SSE 错误、临时缓冲、钱包、预占和 Key 调度定向回归通过。
- 仓储事务回滚/去重、调度投影及连接设置定向回归通过。
- apicompat、antigravity、gemini、geminicli、web、config 完整包测试通过。
- 临时 PostgreSQL/Redis 执行全部迁移，验证 Ultra Fast 字段读写、按原因清除模型限流、UUID 费率事务回滚和显式零价；积分钱包退款及并发退款/消费测试通过。
- 可复跑数据库验证：`python3 backend/scripts/test-release0222-integration.py`。测试容器独立于现有业务数据库。

扩大范围的历史测试并非全部通过：旧模型目录/生图用例仍期望已退役模型，若干图片计费测试仍使用旧倍率预期，Alpha Search 测试仍期望旧 beta header，`TestCodexModelsDoesNotFailOverFromUpstreamConfigurationError` 的旧夹具出现 nil panic。这些失败已在未修改的 `ea9695d8f` 源码快照复现，未通过改变当前业务规则来迁就旧预期。

未使用真实上游凭证进行付费请求或实测生图，也未在本次适配中发布 GitHub Release。
