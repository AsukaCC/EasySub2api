# EasySub2api

[简体中文](README_CN.md) · [日本語](README_JA.md)

EasySub2api is a self-hosted AI API gateway for managing provider accounts, API keys, routing, quotas, usage, payments, and operational visibility from one web interface.

## Highlights

- Unified Anthropic-, OpenAI-, Gemini-, Grok-, and compatible upstream routing.
- Account pools, model mapping, rate and quota controls, sticky sessions, and failover.
- User, group, key, subscription, wallet, payment, affiliate, and usage management.
- PostgreSQL and Redis persistence with Docker/Compose deployment.
- Responsive multilingual administration and user interfaces.

## Quick start for local development

```bash
git clone https://github.com/AsukaCC/EasySub2api.git
cd EasySub2api
cp deploy/.env.example deploy/.env
# Set POSTGRES_PASSWORD, JWT_SECRET and TOTP_ENCRYPTION_KEY in deploy/.env.
docker compose -f deploy/docker-compose.dev.yml up -d --build easysub2api
docker compose -f deploy/docker-compose.dev.yml ps
```

Open `http://127.0.0.1:8080` unless `SERVER_PORT` was changed.

## Documentation

- [Documentation index](docs/README.md)
- [Deployment](deploy/README.md)
- [Docker and Compose](deploy/DOCKER.md)
- [Development](docs/DEVELOPMENT.md)
- [Maintainer release process](docs/RELEASE.md)
- [EasySub2api runtime-name migration](docs/MIGRATION_EASYSUB2API.md)

Git pushes and tags never publish a release. Maintainers explicitly start releases from a local, authenticated command; GitHub then builds the GHCR multi-architecture image.

## Repository layout

```text
backend/   Go API, services, persistence and migrations
frontend/  Vue application and shared visual system
deploy/    Compose, systemd and migration assets
docs/      Development, release, operations and feature documentation
scripts/   Maintainer-only local commands
```

## License

See [LICENSE](LICENSE).
