# ЛР3: запуск целиком — `make start` (нужны docker, go, curl).
# Шаги идемпотентны: повторный запуск ничего не ломает.

MINIO_CONTAINER := minio_storage

.PHONY: help start env up wait media db run stop reset-db

help:
	@echo "make start     — всё сразу: .env, docker, бакет и медиа, БД, сервер на :8080"
	@echo "make env       — создать .env из .env.example (если его нет)"
	@echo "make up        — поднять PostgreSQL, MinIO, Adminer"
	@echo "make media     — бакет flight-media (публичный) и файлы услуг из сида"
	@echo "make db        — миграция и сид (чистая БД) или lab3-migration.sql (БД от ЛР2)"
	@echo "make run       — запустить сервер"
	@echo "make stop      — остановить контейнеры (данные сохраняются)"
	@echo "make reset-db  — пересоздать БД с нуля: make reset-db CONFIRM=yes"

start: env up wait media db run

env:
	@test -f .env || { cp .env.example .env; echo ".env создан из .env.example (пароли: change-me)"; }

up: env
	docker compose up -d

wait:
	@until docker exec flight_postgres pg_isready -q; do sleep 1; done
	@until curl -sf http://localhost:9100/minio/health/live >/dev/null; do sleep 1; done

media: wait
	@docker run --rm -i --network container:$(MINIO_CONTAINER) --env-file .env \
		-v "$(CURDIR)/resources/media:/media:ro" --entrypoint sh minio/mc -s < scripts/minio-init.sh

db: wait
	@./scripts/db-init.sh

run:
	go run ./cmd/app

stop:
	docker compose down

reset-db: wait
	@test "$(CONFIRM)" = "yes" || { echo "Удалит все данные БД. Повторите: make reset-db CONFIRM=yes"; exit 1; }
	@set -a; . ./.env; docker exec flight_postgres psql -q -U $${DB_USER:-postgres} -d $${DB_NAME:-flight_service_db} \
		-c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'
	@./scripts/db-init.sh
