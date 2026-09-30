# flight-service-api

Бэкенд заявочной системы по курсу «Разработка Интернет Приложений» (РИП).
Тема 26 — **«Обслуживание рейса в аэропорте»** (лабораторные работы №1–3).

## Предметная область

- **Ресурс обслуживания** (услуга) — виды ресурсов, персонала и техники (топливо, аэродромные тягачи, представители авиакомпании, бортовое питание и т. д.) с единицей измерения и ценой.
- **Заявка на обслуживание рейса** — заявка с указанием количества ресурсов и расчётом стоимости.

## Стек

- Go + [Gin](https://github.com/gin-gonic/gin)
- PostgreSQL + [GORM](https://gorm.io) (всё взаимодействие с БД — только через ORM)
- MinIO (объектное хранилище изображений и видео)
- [logrus](https://github.com/sirupsen/logrus)

## Запуск

1. Поднять PostgreSQL, MinIO и Adminer:

   ```bash
   cp .env.example .env   # задать пароли (MINIO_ROOT_PASSWORD, DB_PASS, MINIO_SECRET_KEY)
   docker compose up -d
   ```

2. Создать публичный бакет `flight-media` в MinIO (консоль: `http://localhost:9101`) и загрузить в него файлы услуг из сида (`tug.jpg`, `tug.mp4` и т. д.; имена латиницей). Adminer для просмотра БД: `http://localhost:8081`.
3. Подготовить БД одним из двух способов.

   **а) Чистая БД:**

   ```bash
   go run ./cmd/migrate
   docker exec -i flight_postgres psql -U postgres -d flight_service_db < docs/lab2-seed.sql
   ```

   `go run ./cmd/migrate` создаёт таблицы по моделям (столбцы `image`/`video`), `lab2-seed.sql` наполняет их; в медиа-полях лежат только имена файлов, часть услуг принадлежит второму создателю (`creator_id = 3`), чтобы был виден `is_mine = 0`.

   **б) Обновление существующей БД из ЛР2** (со столбцами `image_url`/`video_url`): `go run ./cmd/migrate` перед этим не нужен, достаточно один раз выполнить

   ```bash
   docker exec -i flight_postgres psql -U postgres -d flight_service_db < docs/lab3-migration.sql
   ```

   Скрипт переименовывает `image_url`/`video_url` в `image`/`video`, оставляет в них только имя файла и назначает часть услуг второму создателю. На чистой БД его запускать нельзя.

4. Запустить сервер (порт 8080, см. `config/config.toml`):

   ```bash
   go run ./cmd/app
   ```

Переменные окружения (`.env`): параметры PostgreSQL (`DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASS`) и MinIO (`MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`, `MINIO_PUBLIC_URL`).

## REST API

Все методы начинаются с `/api`, ответы — JSON. Пользователь зафиксирован (см. ниже), авторизации в ЛР3 нет. Коллекция запросов в порядке показа: [`docs/lab3.postman_collection.json`](docs/lab3.postman_collection.json) (Postman v2.1, импортируется и в Insomnia; переменная `baseUrl = http://localhost:8080`).

Формат успеха: `{"status":"success","data":...}` (для операций без данных — `"message"`). Формат ошибки: `{"status":"error","description":"..."}`. Услуги в статусе «удален» ни одним методом не отдаются.

### Услуги `/api/flight-services`

| Метод | URL | Параметры / тело | Успех | Ошибки |
|---|---|---|---|---|
| GET | `/api/flight-services` | query `price` — максимальная цена (необязательно) | 200, массив опубликованных услуг с `is_mine` и `likes_count` | 400 (`price` не число), 500 |
| GET | `/api/flight-services/feed` | — | 200, первая опубликованная услуга | 404 (опубликованных нет) |
| GET | `/api/flight-services/:id` | query `next=true` — следующая опубликованная по кругу | 200, услуга | 400 (id не число), 404 (нет, не опубликована или удалена) |
| GET | `/api/flight-services/draft` | — | 200, черновик текущего пользователя | 404 (черновика нет) |
| POST | `/api/flight-services` | multipart/form-data: `name` (обязательно), `description`, `unit`, `price` (0…99999999.99), файлы `image`, `video` | 201, созданный черновик | 400 (нет `name`, некорректная `price`, недопустимый файл), 409 (черновик уже есть), 500 |
| PUT | `/api/flight-services/draft/publish` | — | 200, опубликованная услуга | 404 (черновика нет), 500 |
| DELETE | `/api/flight-services/:id` | — | 200, `{"status":"success","message":"услуга удалена"}` | 400 (id не число), 403 (чужая услуга), 404 (нет или уже удалена) |
| POST | `/api/flight-services/:id/like` | JSON `{"like": 1}` — поставить, `{"like": 0}` — снять | 200, `{"like":1,"likes_count":N}` | 400 (id не число, `like` не 0/1), 404 (услуги нет или она не опубликована) |

Особенности:

- `POST` создаёт **черновик** сразу со всеми полями и файлами; `PUT .../draft/publish` только меняет статус и ставит `formed_at`. Метода редактирования полей нет.
- Файлы `image` и `video` приходят как файлы, а не URL. Тип определяется по содержимому: изображения jpeg/png/webp/gif до 5 МБ, видео mp4/webm до 20 МБ. Файлы необязательны.
- С клиента не принимаются `id`, `status`, `creator_id`, `is_moderator`, `formed_at`.
- Список и лента содержат только опубликованные услуги; `is_mine` = 1, если создатель совпадает с текущим пользователем, иначе 0.
- Повторный лайк не создаёт дубль, снятие отсутствующего лайка безопасно.

Пример ответа `GET /api/flight-services?price=10000`:

```json
{
  "status": "success",
  "data": [
    {
      "id": 3,
      "name": "Бортовое питание (эконом)",
      "description": "Комплект бортового питания на одного пассажира.",
      "status": "опубликован",
      "unit": "порция",
      "price": 460,
      "image_url": "http://localhost:9100/flight-media/catering.jpg",
      "video_url": "http://localhost:9100/flight-media/catering.mp4",
      "formed_at": "2026-08-25T10:05:00Z",
      "likes_count": 3,
      "is_mine": 0
    }
  ]
}
```

Пример ошибки (`DELETE /api/flight-services/3`, услуга чужая):

```json
{"status": "error", "description": "можно удалять только свои услуги"}
```

### Пользователи `/api/users`

| Метод | URL | Тело | Успех | Ошибки |
|---|---|---|---|---|
| POST | `/api/users/register` | JSON `{"login":"...","password":"..."}` (login до 25, password до 100 символов) | 201, `{"id","login","is_moderator"}` | 400 (нет полей или слишком длинные), 409 (логин занят), 500 |
| POST | `/api/users/login` | — | 200, заглушка (реализация в ЛР4) | — |
| POST | `/api/users/logout` | — | 200, заглушка (реализация в ЛР4) | — |

Регистрация настоящая: создаётся обычный пользователь (`is_moderator = false`), пароль хранится как есть и в ответах не отдаётся. Хэширование, аутентификация и сессии — ЛР4.

### Статусы и переходы

```
черновик ──publish──> опубликован
   │                      │
   └──────delete──────────┴──> удален
```

- `черновик -> опубликован` — `PUT /api/flight-services/draft/publish`;
- `черновик | опубликован -> удален` — `DELETE /api/flight-services/:id` (логическое удаление, строка остаётся в БД);
- возврата в `черновик` нет — такого метода не существует;
- у пользователя не более одного черновика (частичный уникальный индекс `idx_one_draft_per_creator`).

### Текущий пользователь

Создатель зафиксирован константой `defaultUserID = 1` (`internal/app/handler/current_user.go`). Доступ к нему — только через функцию-singleton `CurrentUserID()`; методы репозитория константу не читают, id приходит им параметром из обработчика. В ЛР4 здесь появится пользователь из сессии.

## База данных

PostgreSQL, схема создаётся `go run ./cmd/migrate` (GORM AutoMigrate), индекс `idx_one_draft_per_creator` — скриптом `docs/lab2-seed.sql`.

### `users`

| Поле | Тип | Ограничения | Описание |
|---|---|---|---|
| `id` | bigint | PK, автоинкремент | идентификатор |
| `login` | varchar(25) | NOT NULL, UNIQUE | логин |
| `password` | varchar(100) | NOT NULL | пароль (до ЛР4 хранится как есть) |
| `is_moderator` | boolean | DEFAULT false | признак модератора |

### `flight_services`

| Поле | Тип | Ограничения | Описание |
|---|---|---|---|
| `id` | bigint | PK, автоинкремент | идентификатор |
| `name` | varchar(100) | NOT NULL | название услуги |
| `description` | varchar(255) | — | описание |
| `status` | varchar(15) | NOT NULL, DEFAULT `'черновик'` | `черновик` / `опубликован` / `удален` |
| `image` | varchar(200) | — | имя файла изображения в MinIO (`image-<hex>.jpg`) |
| `video` | varchar(200) | — | имя файла видео в MinIO (`video-<hex>.mp4`) |
| `unit` | varchar(20) | — | единица измерения (час, литр, порция) |
| `price` | numeric(10,2) | — | цена за единицу |
| `creator_id` | bigint | NOT NULL, FK -> `users(id)` ON UPDATE/DELETE RESTRICT | создатель |
| `formed_at` | timestamptz | — | дата публикации (формирования), ставится методом публикации |

Индексы: `flight_services_pkey` (`id`); `idx_one_draft_per_creator` — UNIQUE по `creator_id` `WHERE status = 'черновик'` (не более одного черновика на пользователя).

### `flight_service_likes`

Связь многие-ко-многим «пользователь — услуга» (лайки).

| Поле | Тип | Ограничения | Описание |
|---|---|---|---|
| `id` | bigint | PK, автоинкремент | идентификатор |
| `user_id` | bigint | NOT NULL, FK -> `users(id)` ON UPDATE/DELETE RESTRICT | кто поставил лайк |
| `flight_service_id` | bigint | NOT NULL, FK -> `flight_services(id)` ON UPDATE/DELETE RESTRICT | какой услуге |

Индексы: `flight_service_likes_pkey` (`id`); `idx_user_flight_service` — UNIQUE (`user_id`, `flight_service_id`): один лайк пользователя на услугу.

## MinIO

- Бакет `flight-media` создаётся вручную (публичный на чтение); приложение при старте проверяет, что он существует.
- В MinIO лежат файлы, в БД (`flight_services.image`, `flight_services.video`) — только их имена. Полный URL собирается при отдаче как `MINIO_PUBLIC_URL + "/" + имя` (`http://localhost:9100/flight-media/<имя>`).
- Имена загруженных через API файлов генерируются на латинице: `image-<hex>.jpg`, `video-<hex>.mp4` (`<hex>` — 8 случайных байт в hex, расширение по типу файла).
- Если создание записи в БД не удалось, загруженные файлы удаляются из бакета.

## Структура

```
cmd/app/main.go                          запуск сервера
cmd/migrate/main.go                      миграция таблиц (GORM AutoMigrate)
config/config.toml                       хост и порт сервера
internal/pkg/app.go                      сборка приложения
internal/app/config/                     чтение конфигурации
internal/app/dsn/                        DSN PostgreSQL из окружения
internal/app/ds/                         модели: FlightService, User, FlightServiceLike
internal/app/serializer/                 JSON-представления для клиента (FlightService, User)
internal/app/repository/                 работа с БД (GORM) и MinIO
internal/app/handler/                    REST API (api_*.go), CurrentUserID()
resources/media/                         медиа по умолчанию (загружается в MinIO)
docs/lab2-seed.sql                       наполнение БД (ЛР2)
docs/lab3-migration.sql                  переход на хранение имён файлов (ЛР3)
docs/lab3.postman_collection.json        коллекция запросов
docs/lab3-class-diagram.puml             диаграмма классов
docker-compose.yml                       PostgreSQL, MinIO, Adminer
```
