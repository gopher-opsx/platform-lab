#!/bin/sh
set -eu

export PGPASSWORD="${POSTGRES_PASSWORD:-platform}"
PGUSER="${POSTGRES_USER:-platform}"
PGHOST="${POSTGRES_HOST:-postgres}"
PGPORT="${POSTGRES_PORT:-5432}"

psql_base() {
  psql -v ON_ERROR_STOP=1 -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" "$@"
}

wait_for_postgres() {
  echo "Waiting for PostgreSQL..."

  attempts=0
  max_attempts=60

  until psql_base -d postgres -tAc "SELECT 1" >/dev/null 2>&1; do
    attempts=$((attempts + 1))

    if [ "$attempts" -ge "$max_attempts" ]; then
      echo "PostgreSQL did not become ready after ${max_attempts} attempts."
      exit 1
    fi

    sleep 1
  done

  echo "PostgreSQL is ready."
}

ensure_database() {
  db="$1"
  exists="$(psql_base -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='${db}'")"

  if [ "$exists" != "1" ]; then
    echo "Creating database ${db}"
    psql_base -d postgres -c "CREATE DATABASE ${db}"
  fi
}

apply_dir() {
  db="$1"
  dir="$2"

  for file in "$dir"/*.sql; do
    [ -f "$file" ] || continue
    echo "Applying $(basename "$file") -> ${db}"
    psql_base -d "$db" -f "$file"
  done
}

wait_for_postgres

for db in catalog_db orders_db inventory_db payments_db notifications_db; do
  ensure_database "$db"
done

apply_dir catalog_db /migrations/catalog
apply_dir orders_db /migrations/orders
apply_dir inventory_db /migrations/inventory
apply_dir payments_db /migrations/payments
apply_dir notifications_db /migrations/notifications

echo "Applying seed data..."
psql_base -d catalog_db -f /seed/001_catalog_products.sql
psql_base -d inventory_db -f /seed/002_inventory_stock.sql

echo "PostgreSQL databases, migrations, and seed data are ready."
