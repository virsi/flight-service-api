-- Ограничение "не более одного черновика на пользователя" и наполнение БД.
-- Для чистой БД после `go run ./cmd/migrate` (схема ЛР3: столбцы image/video с именами файлов).
-- Выполнять одним запуском.

CREATE UNIQUE INDEX idx_one_draft_per_creator
  ON flight_services (creator_id)
  WHERE status = 'черновик';

INSERT INTO users (id, login, password, is_moderator) VALUES
  (1, 'dispatcher', 'dispatcher', false),
  (2, 'moderator',  'moderator',  true),
  (3, 'engineer',   'engineer',   false);

INSERT INTO flight_services
  (id, name, description, status, image, video, unit, price, creator_id, formed_at)
VALUES
  (1, 'Аэродромный тягач Goldhofer AST-2X', 'Буксировка и постановка воздушного судна на стоянку.',
   'опубликован', 'tug.jpg',
   'tug.mp4', 'час', 18500,
   1, '2026-08-25 10:05:00'),
  (2, 'Авиатопливо ТС-1', 'Заправка воздушного судна авиационным керосином.',
   'опубликован', 'jet-fuel.jpg',
   'jet-fuel.mp4', 'литр', 82,
   1, '2026-08-25 10:05:00'),
  (3, 'Бортовое питание (эконом)', 'Комплект бортового питания на одного пассажира.',
   'опубликован', 'catering.jpg',
   'catering.mp4', 'порция', 460,
   3, '2026-08-25 10:05:00'),
  (4, 'Наземный источник питания (GPU)', 'Обеспечение самолёта электропитанием на стоянке.',
   'опубликован', 'gpu.jpg',
   'gpu.mp4', 'час', 9400,
   3, '2026-08-25 10:05:00'),
  (5, 'Багажный тягач', 'Транспортировка багажа между терминалом и самолётом.',
   'черновик', 'baggage-tug.jpg',
   'baggage-tug.mp4', 'час', 7600,
   1, NULL),
  (6, 'Противообледенительная обработка', 'Обработка воздушного судна противообледенительной жидкостью.',
   'удален', 'deicing.jpg',
   'deicing.mp4', 'рейс', 15000,
   1, NULL);

INSERT INTO flight_service_likes (user_id, flight_service_id) VALUES
  (1, 1), (2, 1), (3, 1),
  (1, 2), (2, 2),
  (1, 3), (2, 3), (3, 3),
  (1, 4);

SELECT setval('users_id_seq', (SELECT max(id) FROM users));
SELECT setval('flight_services_id_seq', (SELECT max(id) FROM flight_services));
