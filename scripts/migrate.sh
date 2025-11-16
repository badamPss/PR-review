#!/bin/sh
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$REPO_ROOT/db/migration}"
: "${PGCONNECT_TIMEOUT:=5}"
: "${WAIT_TIMEOUT_SEC:=60}"

echo "Migration runner: using directory $MIGRATIONS_DIR"
echo "PG: host=$PGHOST port=$PGPORT db=$PGDATABASE user=$PGUSER"

if ! command -v psql >/dev/null 2>&1; then
  echo "psql not found. Please ensure PostgreSQL client is installed." >&2
  exit 1
fi
if ! command -v pg_isready >/dev/null 2>&1; then
  echo "pg_isready not found. Please ensure PostgreSQL client is installed." >&2
  exit 1
fi

MISSING=0
for v in PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE; do
  eval "val=\${$v:-}"
  if [ -z "$val" ]; then
    echo "Env var $v is required" >&2
    MISSING=1
  fi
done
if [ "$MISSING" -eq 1 ]; then
  exit 1
fi

echo "Waiting for Postgres at $PGHOST:$PGPORT (db=$PGDATABASE, user=$PGUSER)..."
elapsed=0
while ! pg_isready -h "$PGHOST" -p "$PGPORT" -d "$PGDATABASE" -U "$PGUSER" >/dev/null 2>&1 ; do
  if [ "$elapsed" -ge "$WAIT_TIMEOUT_SEC" ]; then
    echo "Postgres did not become ready within ${WAIT_TIMEOUT_SEC}s" >&2
    exit 1
  fi
  sleep 1
  elapsed=$((elapsed + 1))
done
echo "Postgres is ready."

if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo "Migrations dir not found: $MIGRATIONS_DIR" >&2
  exit 1
fi

files=$(ls -1 "$MIGRATIONS_DIR"/*_*.up.sql 2>/dev/null | sort || true)
if [ -z "${files}" ]; then
  echo "No migrations to apply in $MIGRATIONS_DIR"
  exit 0
fi

echo "Applying migrations:"
applied=0
for f in $files; do
  [ -z "$f" ] && continue
  echo "  -> $f"
  PGPASSWORD="$PGPASSWORD" psql "host=$PGHOST port=$PGPORT user=$PGUSER dbname=$PGDATABASE connect_timeout=$PGCONNECT_TIMEOUT" \
    -v ON_ERROR_STOP=1 -f "$f"
  applied=$((applied + 1))
done

echo "Migrations applied successfully. Applied $applied file(s)."


