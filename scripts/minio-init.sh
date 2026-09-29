#!/bin/sh
# Выполняется внутри контейнера minio/mc (см. `make media`).
# Создаёт публичный бакет и загружает файлы услуг из сида, если их там ещё нет.
# Берёт resources/media/<имя>.jpg|mp4, а при отсутствии — default.jpg|mp4.
set -e

mc alias set local http://localhost:9000 "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY" >/dev/null
mc mb --ignore-existing "local/$MINIO_BUCKET"
mc anonymous set download "local/$MINIO_BUCKET" >/dev/null

for name in tug jet-fuel catering gpu deicing baggage-tug; do
  for ext in jpg mp4; do
    mc stat "local/$MINIO_BUCKET/$name.$ext" >/dev/null 2>&1 && continue
    src="/media/$name.$ext"
    [ -f "$src" ] || src="/media/default.$ext"
    mc cp "$src" "local/$MINIO_BUCKET/$name.$ext" >/dev/null
    echo "загружен $name.$ext"
  done
done
echo "MinIO: бакет $MINIO_BUCKET готов"
