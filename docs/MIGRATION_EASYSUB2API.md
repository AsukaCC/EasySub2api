# EasySub2api 运行命名迁移

新部署统一使用 `easysub2api`。迁移版本仍接受旧环境变量、API 路径和 WebSocket 协议，但这些兼容入口将在下一个发布版本删除。

## 主要映射

| 旧资源 | 新资源 |
|---|---|
| `/opt/sub2api` | `/opt/easysub2api` |
| `/etc/sub2api` | `/etc/easysub2api` |
| `/var/lib/sub2api` | `/var/lib/easysub2api` |
| `sub2api.service` | `easysub2api.service` |
| `/app/sub2api` | `/app/easysub2api` |
| `sub2api-*` 容器/卷/网络 | `easysub2api-*` |
| PostgreSQL 用户和库 `sub2api` | `easysub2api` |
| `SUB2API_*` | `EASYSUB2API_*` |
| `/v1/sub2api/billing` | `/v1/easysub2api/billing` |
| `sub2api-admin` | `easysub2api-admin` |

## 安全迁移顺序

迁移工具默认只显示操作，不修改系统：

```bash
./deploy/migrate-easysub2api-names.sh --native
./deploy/migrate-easysub2api-names.sh --docker --project deploy
./deploy/migrate-easysub2api-names.sh --database
```

确认目标不存在、备份目录正确后，逐项加 `--apply`。不要一次运行全部作用域；每一步完成后先检查备份和服务状态。

```bash
sudo ./deploy/migrate-easysub2api-names.sh --native --apply \
  --backup-dir /var/backups/easysub2api-name-migration

./deploy/migrate-easysub2api-names.sh --docker --project deploy --apply \
  --backup-dir "$PWD/backups/easysub2api-name-migration"

./deploy/migrate-easysub2api-names.sh --database --apply \
  --backup-dir "$PWD/backups/easysub2api-name-migration"
```

数据库步骤会先执行 `pg_dumpall`，备份为空时中止。脚本不会覆盖已存在的新目录或新卷。

## 更新配置并重建

将 `.env` 中的 PostgreSQL 用户、数据库、日志服务名和 Apple container 镜像变量更新为 `easysub2api`，然后检查 Compose：

```bash
docker compose -f deploy/docker-compose.dev.yml config --quiet
docker compose -f deploy/docker-compose.dev.yml up -d --build easysub2api
docker compose -f deploy/docker-compose.dev.yml ps
```

## 兼容行为

- 新环境变量优先；只设置旧变量时仍可读取并记录弃用警告。
- 旧计费路径返回旧对象标识和弃用响应头，新路径返回 `easysub2api.key_billing`。
- 上游探测先请求新路径，仅在 404/405 时回退旧路径。
- 浏览器 Storage 会复制到新前缀并删除旧键。
- Redis 配置失效通知在兼容期向新旧通道同时发布。

## 回滚

停止新服务，恢复迁移脚本生成的文件/数据库备份，再恢复旧 `.env` 和旧 Compose 版本。不要只回滚镜像：数据库角色或目录已经迁移时，镜像回滚无法恢复资源名。
