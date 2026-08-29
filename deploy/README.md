# EasySub2api 部署入口

新版本以容器镜像作为唯一正式发布产物。GitHub Release 不再提供二进制附件，旧二进制安装器仅在 `legacy/` 中保留一个兼容版本用于迁移或卸载。

## 选择部署方式

| 场景 | 文件/文档 |
|---|---|
| 本地源码开发 | `docker-compose.dev.yml` |
| 生产 Compose + 命名卷 | `docker-compose.yml` |
| 生产 Compose + 本地目录 | `docker-compose.local.yml` |
| 只运行应用、外接 PostgreSQL/Redis | `docker-compose.standalone.yml` |
| macOS Apple container | [APPLE_CONTAINER.md](APPLE_CONTAINER.md) |
| 容器配置说明 | [DOCKER.md](DOCKER.md) |
| 边缘代理与安全 | [EDGE_SECURITY.md](EDGE_SECURITY.md) |
| 旧名称迁移 | [../docs/MIGRATION_EASYSUB2API.md](../docs/MIGRATION_EASYSUB2API.md) |

## 本地开发

```bash
cp deploy/.env.example deploy/.env
# 编辑 deploy/.env，至少设置 POSTGRES_PASSWORD、JWT_SECRET、TOTP_ENCRYPTION_KEY。
docker compose -f deploy/docker-compose.dev.yml config --quiet
docker compose -f deploy/docker-compose.dev.yml up -d --build easysub2api
docker compose -f deploy/docker-compose.dev.yml ps
```

## 生产部署

固定具体版本，不要依赖 `latest` 进行长期运行：

```bash
cp deploy/.env.example deploy/.env
# deploy/.env: EASYSUB2API_IMAGE=ghcr.io/asukacc/easysub2api:0.0.2
docker compose -f deploy/docker-compose.yml config --quiet
docker compose -f deploy/docker-compose.yml up -d
docker compose -f deploy/docker-compose.yml ps
```

升级前应备份 PostgreSQL、应用数据和 `.env`。数据库迁移只向前执行，回滚镜像不能撤销已经执行的数据库结构迁移。

## 资源名称

- Compose 服务：`easysub2api`
- 应用容器：`EasySub2api`
- PostgreSQL/Redis 容器与网络：`easysub2api-*`
- 默认 PostgreSQL 用户/数据库：`easysub2api`
- 应用镜像：`ghcr.io/asukacc/easysub2api`

如果现有部署仍使用旧资源名，先阅读迁移文档；Compose 不会自动移动或覆盖持久化卷。
