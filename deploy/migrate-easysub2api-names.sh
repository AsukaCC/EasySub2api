#!/usr/bin/env bash
# Migrate legacy runtime resource names to EasySub2api names.
# The script is a dry run unless --apply is explicitly provided.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APPLY=false
MIGRATE_NATIVE=false
MIGRATE_DOCKER=false
MIGRATE_DATABASE=false
COMPOSE_PROJECT="deploy"
BACKUP_ROOT="${EASYSUB2API_MIGRATION_BACKUP_DIR:-${PWD}/easysub2api-migration-backup-$(date +%Y%m%d%H%M%S)}"

usage() {
    cat <<'EOF'
Usage: migrate-easysub2api-names.sh [options]

Options:
  --native              Migrate /opt, /etc, /var/lib and systemd resources.
  --docker              Copy the legacy application named volume to the new name.
  --database            Rename the legacy PostgreSQL role and database.
  --project NAME        Compose project prefix (default: deploy).
  --backup-dir PATH     Required backup destination override.
  --apply               Perform changes. Without this flag the script is read-only.
  -h, --help            Show this help.

Run each scope separately and inspect its backup before continuing. Existing
new-name targets are never overwritten.
EOF
}

while [ "$#" -gt 0 ]; do
    case "$1" in
        --native) MIGRATE_NATIVE=true ;;
        --docker) MIGRATE_DOCKER=true ;;
        --database) MIGRATE_DATABASE=true ;;
        --project)
            shift
            [ "$#" -gt 0 ] || { echo "--project requires a value" >&2; exit 2; }
            COMPOSE_PROJECT="$1"
            ;;
        --backup-dir)
            shift
            [ "$#" -gt 0 ] || { echo "--backup-dir requires a value" >&2; exit 2; }
            BACKUP_ROOT="$1"
            ;;
        --apply) APPLY=true ;;
        -h|--help) usage; exit 0 ;;
        *) echo "Unknown option: $1" >&2; usage >&2; exit 2 ;;
    esac
    shift
done

if ! $MIGRATE_NATIVE && ! $MIGRATE_DOCKER && ! $MIGRATE_DATABASE; then
    echo "Choose at least one scope: --native, --docker, or --database." >&2
    exit 2
fi

run() {
    if $APPLY; then
        "$@"
    else
        printf '[dry-run]'
        printf ' %q' "$@"
        printf '\n'
    fi
}

require_absent() {
    if [ -e "$1" ]; then
        echo "Refusing to overwrite existing target: $1" >&2
        exit 1
    fi
}

backup_path() {
    source_path="$1"
    if [ -e "$source_path" ]; then
        run mkdir -p "$BACKUP_ROOT"
        run cp -a "$source_path" "$BACKUP_ROOT/"
    fi
}

migrate_native() {
    if [ "$(id -u)" -ne 0 ]; then
        echo "--native requires root (use sudo)." >&2
        exit 1
    fi

    backup_path /opt/sub2api
    backup_path /etc/sub2api
    backup_path /var/lib/sub2api
    backup_path /etc/systemd/system/sub2api.service

    if command -v systemctl >/dev/null 2>&1; then
        run systemctl stop sub2api.service
        run systemctl disable sub2api.service
    fi

    for pair in "/opt/sub2api:/opt/easysub2api" "/etc/sub2api:/etc/easysub2api" "/var/lib/sub2api:/var/lib/easysub2api"; do
        old_path="${pair%%:*}"
        new_path="${pair#*:}"
        if [ -e "$old_path" ]; then
            require_absent "$new_path"
            run mv "$old_path" "$new_path"
        fi
    done

    if ! id easysub2api >/dev/null 2>&1; then
        run useradd --system --no-create-home --shell /usr/sbin/nologin easysub2api
    fi
    [ ! -e /opt/easysub2api ] || run chown -R easysub2api:easysub2api /opt/easysub2api
    [ ! -e /var/lib/easysub2api ] || run chown -R easysub2api:easysub2api /var/lib/easysub2api

    if [ -f /opt/easysub2api/sub2api ] && [ ! -e /opt/easysub2api/easysub2api ]; then
        run mv /opt/easysub2api/sub2api /opt/easysub2api/easysub2api
    fi
    run install -m 0644 "$SCRIPT_DIR/easysub2api.service" /etc/systemd/system/easysub2api.service
    run rm -f /etc/systemd/system/sub2api.service
    if command -v systemctl >/dev/null 2>&1; then
        run systemctl daemon-reload
        run systemctl enable --now easysub2api.service
    fi
}

migrate_docker_volume() {
    command -v docker >/dev/null 2>&1 || { echo "docker is required for --docker." >&2; exit 1; }
    old_volume="${COMPOSE_PROJECT}_sub2api_data"
    new_volume="${COMPOSE_PROJECT}_easysub2api_data"
    docker volume inspect "$old_volume" >/dev/null 2>&1 || {
        echo "Legacy volume not found: $old_volume" >&2
        exit 1
    }
    if docker volume inspect "$new_volume" >/dev/null 2>&1; then
        echo "Refusing to overwrite existing volume: $new_volume" >&2
        exit 1
    fi

    run mkdir -p "$BACKUP_ROOT"
    run docker run --rm -v "$old_volume:/source:ro" -v "$BACKUP_ROOT:/backup" alpine:3.21 \
        sh -c 'tar -C /source -czf /backup/application-data.tar.gz .'
    run docker volume create "$new_volume"
    run docker run --rm -v "$old_volume:/source:ro" -v "$new_volume:/target" alpine:3.21 \
        sh -c 'tar -C /source -cf - . | tar -C /target -xf -'
}

migrate_database() {
    command -v docker >/dev/null 2>&1 || { echo "docker is required for --database." >&2; exit 1; }
    postgres_container=""
    for candidate in sub2api-postgres easysub2api-postgres sub2api-postgres-dev easysub2api-postgres-dev; do
        if docker container inspect "$candidate" >/dev/null 2>&1; then
            postgres_container="$candidate"
            break
        fi
    done
    [ -n "$postgres_container" ] || { echo "No compatible PostgreSQL container was found." >&2; exit 1; }

    # PostgreSQL cannot rename a database while the application still owns
    # active sessions. Leave the app stopped so the operator can update .env
    # before recreating it with the canonical credentials.
    if docker container inspect EasySub2api >/dev/null 2>&1; then
        run docker stop EasySub2api
    fi

    run mkdir -p "$BACKUP_ROOT"
    if $APPLY; then
        docker exec "$postgres_container" pg_dumpall -U sub2api > "$BACKUP_ROOT/postgresql.sql"
        [ -s "$BACKUP_ROOT/postgresql.sql" ] || { echo "Database backup is empty; aborting." >&2; exit 1; }
    else
        echo "[dry-run] docker exec $postgres_container pg_dumpall -U sub2api > $BACKUP_ROOT/postgresql.sql"
    fi

    run docker exec "$postgres_container" psql -v ON_ERROR_STOP=1 -U sub2api -d postgres -c \
        'ALTER DATABASE sub2api RENAME TO easysub2api;'
    run docker exec "$postgres_container" psql -v ON_ERROR_STOP=1 -U sub2api -d postgres -c \
        'ALTER ROLE sub2api RENAME TO easysub2api;'
}

echo "Backup directory: $BACKUP_ROOT"
$APPLY || echo "Dry run only. Re-run with --apply after reviewing every command."
$MIGRATE_NATIVE && migrate_native
$MIGRATE_DOCKER && migrate_docker_volume
$MIGRATE_DATABASE && migrate_database

echo "Migration scope completed. Update deploy/.env to use easysub2api values, then recreate the stack and verify health."
