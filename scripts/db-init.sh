#!/usr/bin/env bash
# Готовит БД к ЛР3 (см. `make db`):
#   таблиц нет            -> go run ./cmd/migrate + docs/lab2-seed.sql
#   есть image_url (ЛР2)  -> docs/lab3-migration.sql
#   иначе                 -> ничего не делает
set -euo pipefail
cd "$(dirname "$0")/.."
set -a; . ./.env; set +a

psql() { docker exec -i flight_postgres psql -q -U "${DB_USER:-postgres}" -d "${DB_NAME:-flight_service_db}" -v ON_ERROR_STOP=1 "$@"; }
q() { psql -Atc "$1"; }

if [ "$(q "select to_regclass('public.flight_services') is not null")" != "t" ]; then
  echo "БД пустая: миграция и сид"
  go run ./cmd/migrate
  psql < docs/lab2-seed.sql >/dev/null
elif [ "$(q "select count(*) from information_schema.columns where table_name='flight_services' and column_name='image_url'")" != "0" ]; then
  echo "БД от ЛР2: docs/lab3-migration.sql"
  psql < docs/lab3-migration.sql
else
  echo "БД уже готова"
fi
