# GPT-6 Sol / Luna 调用支持

OpenAI 平台新增 `gpt-6-sol` 和 `gpt-6-luna`，可通过现有 Responses、Chat Completions 兼容入口和 Codex 模型列表选择。已有账号的显式模型白名单或映射仍然生效；需要在相应配置中允许新模型，上游账号也必须具有访问权限。

两个模型支持文本、图片输入，文本输出，1,050,000 token 上下文和最多 128,000 token 输出。推理强度为 `none`、`low`、`medium`（默认）、`high`、`xhigh`、`max`。

使用工具调用时推荐 Responses API；直连 OpenAI Chat Completions 的函数调用要求 `reasoning_effort: "none"`。

```json
{
  "model": "gpt-6-sol",
  "input": "Hello",
  "reasoning": { "effort": "medium" }
}
```

内置标准价格（美元 / 百万 token）：

| 模型 | 输入 | 缓存读取 | 缓存写入 | 输出 |
| --- | ---: | ---: | ---: | ---: |
| gpt-6-sol | 2 | 0.20 | 2.50 | 10 |
| gpt-6-luna | 0.10 | 0.01 | 0.125 | 0.50 |

Fast 价格为标准价格的 2 倍。输入超过 272,000 token 时，整次请求的输入和缓存价格为 2 倍、输出价格为 1.5 倍；沿用项目现有长上下文计费开关和分组、渠道价格配置。

资料核对日期：2026-09-23。

- [发布说明](https://openai.com/index/introducing-gpt-6-sol-and-luna/)
- [Sol API 文档](https://developers.openai.com/api/docs/models/gpt-6-sol)
- [Luna API 文档](https://developers.openai.com/api/docs/models/gpt-6-luna)
- [官方价格](https://developers.openai.com/api/docs/pricing)
