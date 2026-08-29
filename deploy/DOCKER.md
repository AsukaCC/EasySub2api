# EasySub2api Docker 与 Compose

## 镜像

正式镜像只发布到 GitHub Container Registry：

```bash
docker pull ghcr.io/asukacc/easysub2api:0.0.2
docker run --rm ghcr.io/asukacc/easysub2api:0.0.2 --version
```

支持 `linux/amd64` 和 `linux/arm64`。稳定版本同时维护 `X.Y.Z`、`X.Y`、`X` 与 `latest` 标签；生产环境应固定 `X.Y.Z` 或 digest。

## Compose

```bash
cp deploy/.env.example deploy/.env
docker compose -f deploy/docker-compose.yml config --quiet
docker compose -f deploy/docker-compose.yml up -d
docker compose -f deploy/docker-compose.yml ps
docker compose -f deploy/docker-compose.yml logs -f easysub2api
```

本地源码构建使用：

```bash
docker compose -f deploy/docker-compose.dev.yml up -d --build easysub2api
```

## 必需配置

至少在 `deploy/.env` 中设置：

```dotenv
POSTGRES_USER=easysub2api
POSTGRES_PASSWORD=replace-with-a-strong-password
POSTGRES_DB=easysub2api
JWT_SECRET=replace-with-a-random-secret
TOTP_ENCRYPTION_KEY=replace-with-a-random-secret
EASYSUB2API_IMAGE=ghcr.io/asukacc/easysub2api:0.0.2
```

默认只把应用端口暴露到宿主机；PostgreSQL 和 Redis 不应直接暴露到公网。生产环境还应配置可信代理、TLS、备份和资源限制，详见 [EDGE_SECURITY.md](EDGE_SECURITY.md)。

## 数据与升级

- `docker-compose.yml` 使用命名卷。
- `docker-compose.local.yml` 和开发配置使用 `deploy/` 下的本地数据目录。
- `.env`、`data/`、`postgres_data/`、`redis_data/` 均被 Git 忽略。
- 升级前备份数据库和应用数据；不要用 `docker compose down -v` 进行普通升级。

旧的 `sub2api` 资源不会被 Compose 自动接管。按 [命名迁移文档](../docs/MIGRATION_EASYSUB2API.md) 先预览并备份，再执行显式迁移。

## 发布说明

push/tag 不会自动构建镜像。维护者仅能通过 [本地手动发布流程](../docs/RELEASE.md) 触发 GitHub 的多架构构建。
