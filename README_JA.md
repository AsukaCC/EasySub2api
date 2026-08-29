# EasySub2api

[English](README.md) · [简体中文](README_CN.md)

EasySub2api は、プロバイダーアカウント、API キー、モデルルーティング、クォータ、利用量、決済、運用状態を一元管理するセルフホスト型 AI API ゲートウェイです。

## 主な機能

- Anthropic、OpenAI、Gemini、Grok、および互換アップストリームの統合。
- アカウントプール、モデルマッピング、倍率・クォータ、スティッキーセッション、フェイルオーバー。
- ユーザー、グループ、キー、サブスクリプション、ウォレット、決済、アフィリエイト、利用量管理。
- PostgreSQL と Redis、Docker/Compose デプロイ。
- レスポンシブで多言語対応の管理・ユーザー画面。

## ローカル開発のクイックスタート

```bash
git clone https://github.com/AsukaCC/EasySub2api.git
cd EasySub2api
cp deploy/.env.example deploy/.env
# deploy/.env で POSTGRES_PASSWORD、JWT_SECRET、TOTP_ENCRYPTION_KEY を設定します。
docker compose -f deploy/docker-compose.dev.yml up -d --build easysub2api
docker compose -f deploy/docker-compose.dev.yml ps
```

`SERVER_PORT` を変更していなければ `http://127.0.0.1:8080` を開きます。

## ドキュメント

- [ドキュメント一覧](docs/README.md)
- [デプロイ](deploy/README.md)
- [Docker と Compose](deploy/DOCKER.md)
- [開発ガイド](docs/DEVELOPMENT.md)
- [メンテナー向け手動リリース](docs/RELEASE.md)
- [実行時名称の移行](docs/MIGRATION_EASYSUB2API.md)

ブランチや tag の push だけではリリースされません。メンテナーがローカルの認証済みコマンドを明示的に実行した後、GitHub が GHCR マルチアーキテクチャイメージとスタンドアロン配布アーカイブを構築します。

## リポジトリ構成

```text
backend/   Go API、サービス、永続化、マイグレーション
frontend/  Vue アプリケーションと共通ビジュアルシステム
deploy/    Compose、systemd、移行ツール
docs/      開発、リリース、運用、機能ドキュメント
scripts/   メンテナー専用ローカルコマンド
```

## ライセンス

[LICENSE](LICENSE) を参照してください。
