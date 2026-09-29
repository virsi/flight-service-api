-- ТОЛЬКО для обновления БД, созданной в ЛР2 (столбцы image_url/video_url).
-- Не выполнять на чистой БД после `go run ./cmd/migrate`: там уже есть image/video, скрипт упадёт.
-- ЛР3: в полях медиа храним имя файла, а не URL; часть услуг — от другого
-- пользователя, чтобы в списке был виден признак is_mine = 0.
ALTER TABLE flight_services RENAME COLUMN image_url TO image;
ALTER TABLE flight_services RENAME COLUMN video_url TO video;
UPDATE flight_services SET image = regexp_replace(image, '^.*/', ''),
                           video = regexp_replace(video, '^.*/', '');
UPDATE flight_services SET creator_id = 3 WHERE id IN (3, 4);
