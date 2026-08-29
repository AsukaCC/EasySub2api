# EasySub2api

[English](README.md) · [日本語](README_JA.md)

EasySub2api 是一个可自托管的 AI API 网关，用于统一管理上游账号、API Key、模型路由、配额、用量、支付与运行状态。

## 核心能力

- 统一接入 Anthropic、OpenAI、Gemini、Grok 及兼容上游。
- 账号池、模型映射、倍率与配额、粘性会话、故障切换。
- 用户、分组、密钥、订阅、钱包、支付、分销和用量管理。
- PostgreSQL 与 Redis 持久化，支持 Docker/Compose 部署。
- 响应式、多语言的管理端和用户端界面。

## 本地开发快速开始

```bash
git clone https://github.com/AsukaCC/EasySub2api.git
cd EasySub2api
cp deploy/.env.example deploy/.env
# 在 deploy/.env 中设置 POSTGRES_PASSWORD、JWT_SECRET、TOTP_ENCRYPTION_KEY。
docker compose -f deploy/docker-compose.dev.yml up -d --build easysub2api
docker compose -f deploy/docker-compose.dev.yml ps
```

未修改 `SERVER_PORT` 时访问 `http://127.0.0.1:8080`。

## 文档

- [完整文档索引](docs/README.md)
- [部署入口](deploy/README.md)
- [Docker 与 Compose](deploy/DOCKER.md)
- [开发指南](docs/DEVELOPMENT.md)
- [维护者手动发布](docs/RELEASE.md)
- [EasySub2api 运行命名迁移](docs/MIGRATION_EASYSUB2API.md)

推送分支或 tag 不会自动发布。维护者必须在本地显式执行经过认证的发布命令，随后由 GitHub 构建 GHCR 多架构镜像和独立程序包。

## 项目结构

```text
backend/   Go API、服务、数据访问与迁移
frontend/  Vue 应用与公共视觉系统
deploy/    Compose、systemd 和迁移资产
docs/      开发、发布、运维和功能文档
scripts/   仅供维护者使用的本地命令
```

## 许可证

请参阅 [LICENSE](LICENSE)。
