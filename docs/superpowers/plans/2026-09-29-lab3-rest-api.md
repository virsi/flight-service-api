# ЛР3: REST API веб-сервиса для SPA в flight-service-api — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Задачи выполнять строго по порядку: каждая следующая опирается на интерфейсы предыдущей.

**Goal:** Добавить в бэкенд JSON веб-сервис под префиксом `/api` со всей итоговой бизнес-логикой (кроме авторизации): 8 методов домена услуги + 3 метода домена пользователя, файлы изображения и видео в MinIO, работа с БД только через ORM, пользователь-создатель — константа через функцию-singleton. Плюс README с описанием методов и таблиц, коллекция Postman/Insomnia и диаграмма классов.

**Architecture:** Слои остаются как в ЛР2: `ds` (модели) → `repository` (единственный слой, знающий про GORM и MinIO) → `handler` (Gin). Новое: `internal/app/serializer` (JSON-представления ответов), `internal/app/handler/current_user.go` (singleton), `internal/app/handler/api_*.go` (API-обработчики). HTML-страницы ЛР1–2 продолжают работать рядом с API.

**Tech Stack:** Go 1.25, Gin 1.12, GORM + postgres, `github.com/minio/minio-go/v7`, PostgreSQL / MinIO / Adminer в docker compose (уже подняты: `flight_postgres`, `minio_storage`, `flight_adminer`).

---

## Global Constraints (требования задания ЛР3 — соблюдать в каждой задаче)

- **Все методы API начинаются с `/api`.** Методы и URL — по REST (ресурс в URL, действие — HTTP-методом).
- **Взаимодействие с БД — только через ORM** (GORM). Никакого `Raw`/`Exec` в новом и изменённом коде.
- **Записи в статусе `удален` на клиент не передаются** ни одним методом.
- **Статусы:** `черновик` → `опубликован` (метод публикации) и `черновик|опубликован` → `удален` (метод удаления). Это **два разных метода**. Вернуть в `черновик` нельзя — такого метода нет.
- **Пользователь-создатель зафиксирован** константой, доступ к нему — только через функцию-singleton `CurrentUserID()`. Ни один метод репозитория не читает константу сам: id пользователя приходит параметром из обработчика.
- **Системные поля с клиента не принимаются:** `id`, `status`, `creator_id`, модератор (`is_moderator`), `formed_at` (дата формирования). Запросы разбираются только в структуры/поля, где этих полей нет. Даты создания и завершения в модели отсутствуют (убраны в ЛР2) — не добавлять.
- **Список услуг — только опубликованные**, фильтрация на бэкенде `?price=` (максимальная цена, как в ЛР1). У каждого элемента признак `is_mine` = `0|1` (создатель совпадает с текущим пользователем).
- **Лента — только опубликованные.**
- **Черновик — не более 1 на пользователя**, получается без id.
- **Добавление:** multipart-форма, файлы `image` и `video` приходят **как файлы**, а не URL. В MinIO лежат файлы, в БД — **только имена файлов**. Имена генерируются на латинице.
- **Удаление — только soft delete** (смена статуса) и **только своих** услуг.
- **Лайк:** поле `like` = `0|1` (0 снимает, 1 ставит), от текущего пользователя.
- **Пользователь:** регистрация — настоящая; аутентификация и деавторизация — заглушки для ЛР4.
- **Ответы — JSON.** Успех: `{"status":"success","data":...}`. Ошибка — существующий `errorHandler`: `{"status":"error","description":"..."}`. Коды: 200, 201, 400, 403, 404, 409, 500. Ответ «OK» только если БД реально изменилась.
- **НЕ делать:** авторизацию/JWT/сессии/хэширование паролей (это ЛР4), заявки/корзину/модератора, Swagger, тесты-харнесс, CORS, изменение внешнего вида HTML-страниц, редактирование полей услуги (такого метода в задании нет).
- **Коммиты НЕ делать.** Работа остаётся незакоммиченной до ревью пользователем (правило пользователя). Ветка уже создана: `lab3-flight-service-spa-backend`.
- Стиль кода — как в существующих файлах: короткие комментарии на русском над функциями, `h.errorHandler(ctx, code, err)`, `ctx.JSON(...)`.

## Принятые решения (зафиксированы, не пересматривать без пользователя)

| # | Решение | Почему |
|---|---|---|
| D1 | HTML-страницы ЛР1–2 остаются, API добавляется рядом под `/api` | задание ЛР3 их не отменяет; SPA появится в ЛР5+ |
| D2 | `POST /api/flight-services` создаёт **черновик** сразу со всеми полями и файлами; `PUT .../draft/publish` только меняет статус и ставит `formed_at` | в задании PUT публикации = «смена статуса»; метода редактирования полей нет |
| D3 | Soft delete переводится на ORM (`Updates`), один метод репозитория для HTML и API | ЛР3: «взаимодействие с БД через ORM»; сырой SQL-курсор был требованием только ЛР2 (он остался в ветке `lab2-flight-service-db`) |
| D4 | Колонки `image_url`/`video_url` переименовываются в `image`/`video` и хранят имя файла; полный URL собирается при отдаче (сериализатор, шаблоны) | «название файлов сохраняются в полях БД» |
| D5 | Пароль при регистрации хранится как есть | хэширование и логин — ЛР4; сид-пользователи тоже с открытыми паролями |
| D6 | Удаление чужой услуги → 403, несуществующей/удалённой → 404 | понятнее на защите, чем общий 404 |

## Итоговые маршруты API

| Метод | URL | Тело | Успех | Ошибки |
|---|---|---|---|---|
| GET | `/api/flight-services?price=` | — | 200 | 400 (price не число), 500 |
| GET | `/api/flight-services/feed` | — | 200 (первая опубликованная) | 404 (нет опубликованных) |
| GET | `/api/flight-services/:id?next=true` | — | 200 | 400, 404 |
| GET | `/api/flight-services/draft` | — | 200 | 404 (черновика нет) |
| POST | `/api/flight-services` | multipart: `name`*, `description`, `unit`, `price`, файлы `image`, `video` | 201 | 400, 409 (черновик уже есть), 500 |
| PUT | `/api/flight-services/draft/publish` | — | 200 | 404 (черновика нет) |
| DELETE | `/api/flight-services/:id` | — | 200 | 400, 403, 404 |
| POST | `/api/flight-services/:id/like` | JSON `{"like":1}` / `{"like":0}` | 200 | 400, 404 |
| POST | `/api/users/register` | JSON `{"login","password"}` | 201 | 400, 409 |
| POST | `/api/users/login` | — | 200 (заглушка) | — |
| POST | `/api/users/logout` | — | 200 (заглушка) | — |

Gin 1.12 допускает статический сегмент рядом с параметром (`/draft`, `/feed` и `/:id` на одном уровне) — это проверяется запуском в Task 5.

## File Structure

**Создаются:**

| Файл | Ответственность |
|---|---|
| `internal/app/handler/current_user.go` | Функция-singleton `CurrentUserID()` |
| `internal/app/repository/media.go` | Загрузка/удаление файлов в MinIO, генерация имён, `MediaURL` |
| `internal/app/repository/user.go` | Регистрация пользователя (ORM) |
| `internal/app/serializer/flight_service.go` | JSON-представление услуги |
| `internal/app/serializer/user.go` | JSON-представление пользователя (без пароля) |
| `internal/app/handler/api_flight_service.go` | 8 API-методов домена услуги |
| `internal/app/handler/api_user.go` | 3 API-метода домена пользователя |
| `docs/lab3-migration.sql` | Переименование колонок медиа + сид для демонстрации `is_mine` |
| `docs/lab3.postman_collection.json` | Коллекция запросов (импорт в Postman и Insomnia) |
| `docs/lab3-class-diagram.puml` | Диаграмма классов (PlantUML) |

**Изменяются:** `internal/app/ds/*.go` (json-теги, поля медиа), `internal/app/repository/repository.go` (MinIO-клиент, убрать `CreatorID`), `internal/app/repository/flight_service.go` (userID параметром, ORM-удаление, новые методы), `internal/app/handler/handler.go` (роуты `/api`, FuncMap), `internal/app/handler/flight_service.go` (userID через singleton), `cmd/app/main.go`, `templates/*.html` (только выражения src медиа), `.env`, `.env.example`, `go.mod`, `go.sum`, `README.md`.

---

### Task 1: Функция-singleton текущего пользователя

**Files:** Create `internal/app/handler/current_user.go`; Modify `internal/app/repository/repository.go`, `internal/app/repository/flight_service.go`, `internal/app/handler/flight_service.go`.

**Produces:** `handler.CurrentUserID() uint`; сигнатуры репозитория с `userID uint`:
- `GetDraftFlightService(userID uint) (*ds.FlightService, error)`
- `CreateDraftFlightService(service *ds.FlightService, userID uint) error` — сама проставляет `Status = StatusDraft`, `CreatorID = userID` (системные поля вычисляются на бэкенде)
- `PublishFlightService(userID uint, description, unit string, price float64) error` — для HTML-страницы, логика та же

- [ ] **Step 1: Создать singleton**

```go
package handler

import "sync"

// defaultUserID — создатель зафиксирован до ЛР4 (авторизация)
const defaultUserID uint = 1

var (
	currentUserOnce sync.Once
	currentUserID   uint
)

// CurrentUserID — функция-singleton: текущий пользователь определяется один раз
// и дальше возвращается тот же; в ЛР4 здесь появится пользователь из сессии
func CurrentUserID() uint {
	currentUserOnce.Do(func() {
		currentUserID = defaultUserID
	})

	return currentUserID
}
```

- [ ] **Step 2:** Удалить `const CreatorID` и комментарий к нему из `repository.go`. Во всех методах `flight_service.go`, где был `CreatorID`, принять `userID uint` параметром (сигнатуры выше). `CreateDraftFlightService` принимает `*ds.FlightService` и перезаписывает `Status`/`CreatorID` перед `Create`.
- [ ] **Step 3:** В HTML-обработчиках (`handler/flight_service.go`) передавать `CurrentUserID()`; `CreateFlightDraft` вызывает `CreateDraftFlightService(&ds.FlightService{Name: ctx.PostForm("name")}, CurrentUserID())`.
- [ ] **Step 4: Verify**

Run: `go build ./... && grep -rn "CreatorID\b" internal/app/repository internal/app/handler | grep -v "CreatorID:" | grep -v "creator_id"`
Expected: сборка без ошибок; grep не находит использований константы (только поле модели `CreatorID`).

Run: `go run ./cmd/app &` затем `curl -s -o /dev/null -w "%{http_code}\n" localhost:8080/flight-resources` и `.../flight-draft`
Expected: `200` и `200`. Остановить сервер.

---

### Task 2: Имена файлов вместо URL в БД

**Files:** Modify `internal/app/ds/FlightService.go`, `templates/index.html`, `templates/feed.html`, `templates/add.html`, `internal/app/handler/handler.go`; Create `docs/lab3-migration.sql`.

**Produces:** поля `ds.FlightService.Image string` (колонка `image`) и `Video string` (колонка `video`), хранят имя объекта в бакете (`tug.jpg`). Шаблонная функция `media` (имя → полный URL) — реализуется в Task 3, в этой задаче временно `func(name string) string { return name }` не делать: Task 2 и 3 выполнять одним агентом подряд, проверка — в конце Task 3.

- [ ] **Step 1: Миграция данных** — создать `docs/lab3-migration.sql`:

```sql
-- ЛР3: в полях медиа храним имя файла, а не URL; часть услуг — от другого
-- пользователя, чтобы в списке был виден признак is_mine = 0.
ALTER TABLE flight_services RENAME COLUMN image_url TO image;
ALTER TABLE flight_services RENAME COLUMN video_url TO video;
UPDATE flight_services SET image = regexp_replace(image, '^.*/', ''),
                           video = regexp_replace(video, '^.*/', '');
UPDATE flight_services SET creator_id = 3 WHERE id IN (3, 4);
```

Выполнить: `docker exec -i flight_postgres psql -U postgres -d flight_service_db < docs/lab3-migration.sql`
Expected: `ALTER TABLE` ×2, `UPDATE 6`, `UPDATE 2`.

- [ ] **Step 2: Модель** — заменить поля `ImageURL`/`VideoURL`:

```go
	Image string `gorm:"type:varchar(100)"` // имя файла изображения в MinIO
	Video string `gorm:"type:varchar(100)"` // имя файла видео в MinIO
```

- [ ] **Step 3: Шаблоны** — в трёх шаблонах заменить `.ImageURL` → `.Image`, `.VideoURL` → `.Video` в условиях, а в `src` писать `{{ media .FlightService.Image }}` / `{{ media .flightService.Video }}` (регистр переменной — как в конкретном шаблоне). Ветки с `default.jpg`/`default.mp4` не трогать.
- [ ] **Step 4:** В `RegisterStatic` перед `LoadHTMLGlob` добавить `router.SetFuncMap(template.FuncMap{"media": h.Repository.MediaURL})` (импорт `html/template`).

---

### Task 3: MinIO-клиент в репозитории

**Files:** Modify `internal/app/repository/repository.go`, `cmd/app/main.go`, `.env`, `.env.example`, `go.mod`, `go.sum`; Create `internal/app/repository/media.go`.

**Produces:**
- `repository.New(dsn string, minioCfg MinioSettings) (*Repository, error)`
- `type MinioSettings struct{ Endpoint, AccessKey, SecretKey, Bucket, PublicURL string }`
- `(*Repository) MediaURL(name string) string` — `""` для пустого имени, иначе `PublicURL + "/" + name`
- `(*Repository) UploadMedia(header *multipart.FileHeader, kind string) (string, error)` — `kind` ∈ `"image"|"video"`; валидирует тип по содержимому, грузит, возвращает сгенерированное имя
- `(*Repository) RemoveMedia(name string)` — удаление объекта, ошибки только логируются
- `var ErrInvalidMedia` — неверный тип/размер файла (обработчик отдаёт 400)

- [ ] **Step 1:** `go get github.com/minio/minio-go/v7`
- [ ] **Step 2: env** — дописать в `.env` (значение секрета = текущий `MINIO_ROOT_PASSWORD`, скопировать его, не печатать в логи) и в `.env.example` (с `change-me`):

```dotenv
# MinIO (ЛР3): загрузка файлов услуг
MINIO_ENDPOINT=localhost:9100
MINIO_ACCESS_KEY=root
MINIO_SECRET_KEY=change-me
MINIO_BUCKET=flight-media
MINIO_PUBLIC_URL=http://localhost:9100/flight-media
```

- [ ] **Step 3: `media.go`** — ключевая логика:

```go
// допустимые типы: определяем по первым 512 байтам, а не по заголовку клиента
var mediaTypes = map[string]map[string]string{
	"image": {"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "image/gif": ".gif"},
	"video": {"video/mp4": ".mp4", "video/webm": ".webm"},
}

var mediaMaxSize = map[string]int64{"image": 5 << 20, "video": 20 << 20} // короткое видео — до 20 МБ

var ErrInvalidMedia = errors.New("недопустимый файл")

// UploadMedia — проверка файла и загрузка в MinIO под сгенерированным латинским именем
func (r *Repository) UploadMedia(header *multipart.FileHeader, kind string) (string, error)
```

Алгоритм `UploadMedia`: размер > лимита → `fmt.Errorf("%w: ...", ErrInvalidMedia)`; `header.Open()`, `defer Close`; прочитать 512 байт, `http.DetectContentType`, `Seek(0,0)`; тип не из `mediaTypes[kind]` → `ErrInvalidMedia`; имя = `kind + "-" + hex.EncodeToString(8 байт из crypto/rand) + ext` (пример `image-3f9a1c2b7d4e5f60.jpg`); `r.minio.PutObject(ctx, r.bucket, name, file, header.Size, minio.PutObjectOptions{ContentType: contentType})`.

- [ ] **Step 4:** В `Repository` добавить поля `minio *minio.Client`, `bucket string`, `publicURL string`; в `New` создать клиент `minio.New(cfg.Endpoint, &minio.Options{Creds: credentials.NewStaticV4(...), Secure: false})` и проверить `BucketExists` — если бакета нет, вернуть ошибку (бакет создан в ЛР1, не создавать автоматически).
- [ ] **Step 5:** `cmd/app/main.go`: собрать `MinioSettings` из `os.Getenv(...)` (env уже загружен `config.NewConfig` через godotenv — проверить порядок: `config.NewConfig()` вызывается раньше `repository.New`, так и оставить).
- [ ] **Step 6: Verify (Task 2 + 3)**

Run: `go build ./... && go run ./cmd/app &`, затем
`curl -s localhost:8080/flight-resources | grep -o 'src="[^"]*"' | head -3`
Expected: `src="http://localhost:9100/flight-media/tug.jpg"` и т. п. — полные URL, собранные из имён. Картинки открываются: `curl -s -o /dev/null -w "%{http_code}\n" http://localhost:9100/flight-media/tug.jpg` → `200` (если в бакете нет этого файла — `404` от MinIO допустим, фолбэк на default отрабатывает только при пустом имени; сообщить пользователю, какие файлы отсутствуют).

Run: `docker exec flight_postgres psql -U postgres -d flight_service_db -c "SELECT id, image, video, creator_id FROM flight_services ORDER BY id;"`
Expected: в `image`/`video` только имена (`tug.jpg`), у id 3 и 4 `creator_id = 3`.

---

### Task 4: Модели и сериализаторы

**Files:** Modify `internal/app/ds/*.go`; Create `internal/app/serializer/flight_service.go`, `internal/app/serializer/user.go`.

**Produces:**
- json-теги у моделей (`json:"id"`, `json:"name"` …), у `User.Password` — `json:"-"`, связи (`Creator`, `User`, `FlightService`) — `json:"-"`.
- `serializer.FlightService` и конструктор:

```go
package serializer

// FlightService — то, что клиент получает об услуге (без служебных связей)
type FlightService struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Unit        string     `json:"unit"`
	Price       float64    `json:"price"`
	ImageURL    string     `json:"image_url"`
	VideoURL    string     `json:"video_url"`
	FormedAt    *time.Time `json:"formed_at"`
	LikesCount  int64      `json:"likes_count"`
	IsMine      int        `json:"is_mine"` // 1 — создатель совпадает с текущим пользователем
}

func NewFlightService(s ds.FlightService, likes int64, currentUserID uint, mediaURL func(string) string) FlightService
```

`FormedAt` — `nil`, если `s.FormedAt.Valid == false`. `IsMine` — `1`, если `s.CreatorID == currentUserID`, иначе `0`.

- `serializer.User{ID uint "json:id"; Login string "json:login"; IsModerator bool "json:is_moderator"}` + `NewUser(u ds.User) User`.

Запросы (входящие данные) — это **отдельные** структуры в обработчиках (Task 6, 8) без системных полей; сериализаторы — только для ответов.

- [ ] **Verify:** `go build ./... && go vet ./...` — без ошибок.

---

### Task 5: API — GET-методы (список, лента, черновик)

**Files:** Create `internal/app/handler/api_flight_service.go`; Modify `internal/app/handler/handler.go`, `internal/app/repository/flight_service.go`.

**Consumes:** `CurrentUserID()`, `serializer.NewFlightService`, `h.Repository.MediaURL`.
**Produces (repository):** `GetFlightService(id)` — теперь **только опубликованные** (`status = StatusPublished`) — лента HTML и API согласно заданию. Остальные методы чтения уже есть.

- [ ] **Step 1:** Обработчики:
  - `GetFlightServicesAPI` — как `GetFlightServices`, но `ctx.JSON(200, gin.H{"status":"success","data":[]serializer.FlightService})`; пустой список → `[]`, не `null` (`make(..., 0, len)`).
  - `GetFlightFeedAPI` — логика `GetFlightFeed` (id из `:id`, `?next=true`, без id — первая опубликованная), но: невалидный id → 400, не найдено → 404 JSON (без редиректа). Для `/feed` параметра `:id` нет → id = 0 → первая опубликованная.
  - `GetFlightDraftAPI` — `GetDraftFlightService(CurrentUserID())`; `nil` → 404 `"черновик не найден"`; иначе 200 с сериализатором (лайков у черновика нет — 0).
- [ ] **Step 2:** В `RegisterHandler` добавить группу:

```go
	api := router.Group("/api")
	api.GET("/flight-services", h.GetFlightServicesAPI)
	api.GET("/flight-services/feed", h.GetFlightFeedAPI)
	api.GET("/flight-services/draft", h.GetFlightDraftAPI)
	api.GET("/flight-services/:id", h.GetFlightFeedAPI)
```

- [ ] **Step 3: Verify** (сервер запущен заново)

```bash
curl -s "localhost:8080/api/flight-services?price=10000" | jq '.data[] | {id, price, is_mine}'
```
Expected (демо-состояние БД: опубликованы 1, 2, 3, 8; удалены 4, 6; черновик 9): ids 2, 3, 8; у id 3 `is_mine: 0`, у остальных `1`.

```bash
curl -s localhost:8080/api/flight-services/feed | jq '.data.id'                 # 1
curl -s "localhost:8080/api/flight-services/1?next=true" | jq '.data.id'        # 2
curl -s localhost:8080/api/flight-services/6 -w "\n%{http_code}\n"              # 404 (удалена)
curl -s localhost:8080/api/flight-services/9 -w "\n%{http_code}\n"              # 404 (черновик в ленте не отдаётся)
curl -s localhost:8080/api/flight-services/-1 -w "\n%{http_code}\n"             # 400
curl -s localhost:8080/api/flight-services/999999999999 -w "\n%{http_code}\n"   # 404
curl -s localhost:8080/api/flight-services/draft | jq '.data | {id, status}'    # {id:9, status:"черновик"}
```

---

### Task 6: API — добавление услуги с файлами

**Files:** Modify `internal/app/handler/api_flight_service.go`, `handler.go`, `internal/app/repository/flight_service.go`.

**Produces (repository):** `var ErrDraftExists = errors.New("у пользователя уже есть черновик")`; `CreateDraftFlightService` возвращает `ErrDraftExists`, если черновик уже есть (проверка через `GetDraftFlightService` внутри). В `statusForError` добавить: `ErrDraftExists` → 409, `ErrInvalidMedia` → 400.

- [ ] **Step 1: Обработчик `CreateFlightServiceAPI`** (`POST /api/flight-services`, multipart):
  1. `ctx.Request.ParseMultipartForm(32 << 20)`, ошибка → 400.
  2. Поля **только** `name` (обязательно, иначе 400), `description`, `unit`, `price` (если непустое — `ParseFloat`, ошибка или `< 0` → 400). Поля `id`, `status`, `creator_id`, `formed_at` из формы **не читаются**.
  3. Файлы: `ctx.FormFile("image")`, `ctx.FormFile("video")` — `http.ErrMissingFile` допустим (будет медиа по умолчанию), другая ошибка → 400.
  4. Проверить отсутствие черновика **до** загрузки файлов (чтобы не мусорить в MinIO): есть → 409.
  5. `UploadMedia` для каждого присланного файла; при ошибке второго — `RemoveMedia` первого; ошибка `ErrInvalidMedia` → 400, иначе 500.
  6. `CreateDraftFlightService(&service, CurrentUserID())`; при ошибке — `RemoveMedia` загруженных, код по `statusForError`.
  7. Ответ **201** `{"status":"success","data": serializer, "message":"черновик услуги создан"}`.
- [ ] **Step 2:** роут `api.POST("/flight-services", h.CreateFlightServiceAPI)`.
- [ ] **Step 3: Verify** — сначала убрать существующий черновик сида, чтобы можно было создать новый: `PUT /api/flight-services/draft/publish` появится только в Task 7, поэтому здесь проверить 409:

```bash
curl -s -F name=Тест -F image=@resources/media/default.jpg localhost:8080/api/flight-services -w "\n%{http_code}\n"   # 409
curl -s -F name=Тест -F image=@go.mod localhost:8080/api/flight-services -w "\n%{http_code}\n"                       # 409 (проверка черновика раньше файлов — ок)
```
Полная проверка создания — в Task 7, Step 4.

---

### Task 7: API — публикация, удаление, лайк

**Files:** Modify `api_flight_service.go`, `handler.go`, `internal/app/repository/flight_service.go`, `internal/app/handler/flight_service.go` (HTML-удаление).

**Produces (repository):**
- `PublishDraftFlightService(userID uint) (ds.FlightService, error)` — ORM `Updates{"status": StatusPublished, "formed_at": time.Now()}` где `creator_id = userID AND status = черновик`; `RowsAffected == 0` → `ErrFlightServiceNotFound`; вернуть обновлённую запись.
- `DeleteFlightService(id, userID uint) error` — **переписать на ORM** (D3): `First` по `id AND status <> удален` → нет → `ErrFlightServiceNotFound`; `CreatorID != userID` → `ErrForbidden` (новая `var ErrForbidden = errors.New("можно удалять только свои услуги")`, в `statusForError` → 403); затем `Model(&service).Update("status", StatusDeleted)`. Удалить импорт `database/sql`, если стал не нужен.
- `SetLike(serviceID, userID uint, like bool) (int64, error)` — услуга должна быть опубликована (`GetFlightService`), иначе `ErrFlightServiceNotFound`; `like` → `r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&ds.FlightServiceLike{UserID: userID, FlightServiceID: serviceID})`; `!like` → `r.db.Where("user_id = ? AND flight_service_id = ?", ...).Delete(&ds.FlightServiceLike{})`; вернуть `GetLikesCount(serviceID)`. Повторный лайк/снятие — идемпотентны (200).

- [ ] **Step 1: Обработчики**
  - `PublishFlightServiceAPI` — `PUT /api/flight-services/draft/publish`, тело не читается; 200 + сериализатор.
  - `DeleteFlightServiceAPI` — `DELETE /api/flight-services/:id`; id `ParseUint` → 400; `DeleteFlightService(id, CurrentUserID())`; 200 `{"status":"success","message":"услуга удалена"}`.
  - `LikeFlightServiceAPI` — `POST /api/flight-services/:id/like`, тело:

```go
type likeRequest struct {
	Like *int `json:"like" binding:"required,oneof=0 1"`
}
```
  ответ 200 `{"status":"success","data":{"like": 0|1, "likes_count": N}}`.
  - HTML-обработчик `DeleteFlightService` — передать `CurrentUserID()`, ошибки через `statusForError`.
- [ ] **Step 2:** роуты:

```go
	api.PUT("/flight-services/draft/publish", h.PublishFlightServiceAPI)
	api.DELETE("/flight-services/:id", h.DeleteFlightServiceAPI)
	api.POST("/flight-services/:id/like", h.LikeFlightServiceAPI)
```

- [ ] **Step 3: Verify — полный жизненный цикл**

```bash
curl -s -X PUT localhost:8080/api/flight-services/draft/publish | jq '.data | {id, status, formed_at}'  # id 9, опубликован, дата
curl -s -X PUT localhost:8080/api/flight-services/draft/publish -w "\n%{http_code}\n"                    # 404
curl -s -F name=Тест -F image=@go.mod localhost:8080/api/flight-services -w "\n%{http_code}\n"           # 400 (не картинка)
curl -s -F name="Трап самоходный" -F description="Посадка пассажиров" -F unit=час -F price=5200 \
     -F status=удален -F creator_id=3 \
     -F image=@resources/media/default.jpg -F video=@resources/media/default.mp4 \
     localhost:8080/api/flight-services | jq '.data | {id, status, image_url, video_url, is_mine}'
```
Expected последнего: `status: "черновик"`, `is_mine: 1` (подсунутые `status`/`creator_id` проигнорированы), URL вида `.../flight-media/image-<16 hex>.jpg` и `video-<16 hex>.mp4`; `curl -s -o /dev/null -w "%{http_code}" <image_url>` → `200`. Повторный POST → 409.

```bash
curl -s -X POST localhost:8080/api/flight-services/1/like -H 'Content-Type: application/json' -d '{"like":1}' | jq .data
curl -s -X POST localhost:8080/api/flight-services/1/like -H 'Content-Type: application/json' -d '{"like":0}' | jq .data   # likes_count на 1 меньше (сид: user 1 уже лайкнул id 1)
curl -s -X POST localhost:8080/api/flight-services/1/like -H 'Content-Type: application/json' -d '{"like":5}' -w "\n%{http_code}\n"  # 400
curl -s -X DELETE localhost:8080/api/flight-services/3 -w "\n%{http_code}\n"   # 403 (создатель — user 3)
curl -s -X DELETE localhost:8080/api/flight-services/2 -w "\n%{http_code}\n"   # 200
curl -s -X DELETE localhost:8080/api/flight-services/2 -w "\n%{http_code}\n"   # 404
curl -s localhost:8080/api/flight-services | jq '[.data[].id]'                   # без 2 и 6
```

После проверки **вернуть БД в демонстрационное состояние** (чтобы пользователь сам сделал скриншоты): `.superpowers/sdd/restore-demo.sh` (TRUNCATE трёх таблиц + данные из снимка `demo-baseline.sql`, включая значения последовательностей). Это разовая чистка данных, не код приложения. Никакого ручного `DELETE ... WHERE id > N` — в БД есть услуги 8 и 9.
Файлы тестовой услуги удалить из бакета через `docker exec minio_storage` или веб-консоль (`http://localhost:9101`) — сообщить пользователю имена.

---

### Task 8: API — домен пользователя

**Files:** Create `internal/app/handler/api_user.go`, `internal/app/repository/user.go`; Modify `handler.go`, `handler/flight_service.go` (`statusForError`).

**Produces (repository):** `var ErrUserExists = errors.New("пользователь с таким логином уже есть")`; `CreateUser(login, password string) (ds.User, error)` — проверка `Count` по login → `ErrUserExists`; `Create(&ds.User{Login, Password, IsModerator: false})`. В `statusForError`: `ErrUserExists` → 409.

- [ ] **Step 1: Обработчики**

```go
// registerRequest — с клиента принимаются только логин и пароль,
// роль (is_moderator) и id вычисляются на бэкенде
type registerRequest struct {
	Login    string `json:"login" binding:"required,max=25"`
	Password string `json:"password" binding:"required,max=100"`
}
```
  - `RegisterAPI` — `ShouldBindJSON` → 400; `CreateUser` → 201 `{"status":"success","data": serializer.NewUser(user)}` (без пароля).
  - `LoginAPI`, `LogoutAPI` — заглушки: 200 `{"status":"success","message":"заглушка: аутентификация будет в ЛР4"}` / `"...деавторизация будет в ЛР4"`. Тело не читают.
- [ ] **Step 2:** роуты `api.POST("/users/register", ...)`, `api.POST("/users/login", ...)`, `api.POST("/users/logout", ...)`.
- [ ] **Step 3: Verify**

```bash
curl -s -X POST localhost:8080/api/users/register -H 'Content-Type: application/json' \
     -d '{"login":"agent_check","password":"123","is_moderator":true}' | jq .    # 201, is_moderator:false, нет password
curl -s -X POST localhost:8080/api/users/register -H 'Content-Type: application/json' \
     -d '{"login":"agent_check","password":"123"}' -w "\n%{http_code}\n"         # 409
curl -s -X POST localhost:8080/api/users/register -H 'Content-Type: application/json' -d '{}' -w "\n%{http_code}\n"   # 400
curl -s -X POST localhost:8080/api/users/login | jq .message
```
Потом: `.superpowers/sdd/restore-demo.sh`.

---

### Task 9: Коллекция запросов

**Files:** Create `docs/lab3.postman_collection.json`.

- [ ] **Step 1:** Коллекция Postman v2.1 с переменной `baseUrl = http://localhost:8080`. Папки и запросы **в порядке показа на защите**:
  1. «1. Список услуг (фильтр)» — `GET {{baseUrl}}/api/flight-services?price=10000`
  2. «2. Добавить услугу с картинкой и видео» — `POST {{baseUrl}}/api/flight-services`, body `formdata`: `name`, `description`, `unit`, `price` (text), `image`, `video` (type `file`, `src` пустой — файл выбирает пользователь)
  3. «3. Черновик» — `GET .../api/flight-services/draft`
  4. «4. Опубликовать черновик» — `PUT .../api/flight-services/draft/publish`
  5. «5. Лента без id» — `GET .../api/flight-services/feed`
  6. «6. Лента по id (next)» — `GET .../api/flight-services/1?next=true`
  7. «7. Лайк» — `POST .../api/flight-services/1/like`, raw JSON `{"like": 1}`
  8. «8. Удалить услугу» — `DELETE .../api/flight-services/10` (id услуги, созданной запросом 2 из демо-состояния; в описании запроса указать, что id берётся из ответа запроса 2)
  9. «9. Регистрация» — `POST .../api/users/register`, raw JSON `{"login":"new_dispatcher","password":"secret"}`
  10. «10. Аутентификация (заглушка)» — `POST .../api/users/login`
  11. «11. Деавторизация (заглушка)» — `POST .../api/users/logout`
  Папка «Ошибки»: `GET /api/flight-services/-1` (400), `GET /api/flight-services/999999999999` (404), `GET /api/flight-services/6` (404, удалена), `DELETE /api/flight-services/3` (403), повторный POST добавления (409), лайк `{"like":5}` (400), `GET ?price=abc` (400).
- [ ] **Verify:** `jq '.item | length' docs/lab3.postman_collection.json` и `jq empty` без ошибок. Файл импортируется и в Insomnia (формат Postman v2.1 она понимает).

---

### Task 10: README

**Files:** Modify `README.md` (переписать устаревшие разделы — там маршруты ЛР1 `/resources`, `/feed`, которых уже нет).

- [ ] Разделы: описание темы (сохранить), стек (+PostgreSQL, GORM, MinIO), запуск (`docker compose up -d`, `.env` из `.env.example`, `go run ./cmd/migrate`, `docs/lab2-seed.sql`, `docs/lab3-migration.sql`, `go run ./cmd/app`), HTML-страницы (актуальные `/flight-resources`, `/flight-feed`, `/flight-draft`), **API: таблица всех 11 методов** (метод, URL, параметры/тело, ответ, коды ошибок — из раздела «Итоговые маршруты API» выше, с примером JSON-ответа списка), **статусы и переходы** (черновик → опубликован → удален, возврата нет), **пользователь зафиксирован** (`CurrentUserID()`, id = 1), **таблицы БД** — по каждой из `users`, `flight_services`, `flight_service_likes` таблица «поле | тип | ограничения | описание» (типы взять из gorm-тегов и `\d flight_services` в psql, включая частичный уникальный индекс `idx_one_draft_per_creator` и уникальный `idx_user_flight_service`), MinIO (бакет `flight-media`, в БД — имена файлов, формат имени `image-<hex>.jpg`), актуальная структура каталогов.
- [ ] **Verify:** каждый URL из README есть в `handler.go` (`grep`), каждое поле таблиц есть в `\d <table>`.

---

### Task 11: Диаграмма классов

**Files:** Create `docs/lab3-class-diagram.puml` (+ `docs/lab3-class-diagram.png`, если получится отрендерить).

- [ ] **Step 1:** PlantUML-диаграмма из пакетов:
  - **Frontend (4 страницы)** — классы-страницы: `Список услуг`, `Лента`, `Добавление услуги`, `Регистрация / вход`.
  - **Домен услуг `/api/flight-services`** — интерфейс `FlightServiceAPI` со всеми 8 методами в виде `+ GET /api/flight-services?price : FlightService[]` и т. д.
  - **Домен пользователя `/api/users`** — интерфейс `UserAPI` с 3 методами.
  - **Handler** — класс `Handler` с Go-методами обработчиков, реализует оба интерфейса; `CurrentUserID()`.
  - **Repository** — класс `Repository` с методами (из итогового кода, выписать все публичные).
  - **Модели `ds`** — `FlightService`, `User`, `FlightServiceLike`; **сериализаторы** — `serializer.FlightService`, `serializer.User`.
  - **БД** — таблицы `flight_services`, `users`, `flight_service_likes` с полями; **MinIO** — бакет `flight-media`.
  - Зависимости (пунктир `..>`): страницы → домены (`Список` → GET список, лайк, удаление; `Лента` → лента, лайк; `Добавление` → черновик, добавление, публикация; `Регистрация` → register/login/logout); домены → модели/сериализаторы; модели → таблицы; Repository → MinIO.
- [ ] **Step 2:** Рендер: `plantuml` не установлен. Попробовать `brew install plantuml` **только с согласия пользователя**; иначе — оставить `.puml` и сообщить, что его можно отрендерить на https://www.plantuml.com/plantuml или перерисовать в StarUML (как ER-диаграмму в ЛР2).

---

### Task 12: Финальная сверка

- [ ] `go build ./... && go vet ./... && gofmt -l .` — пусто.
- [ ] `grep -rn "Raw(\|Exec(" internal/` — пусто (всё через ORM).
- [ ] Пройти все запросы коллекции curl-ом по порядку от демо-состояния, затем `.superpowers/sdd/restore-demo.sh` и удалить из бакета файлы, загруженные проверками.
- [ ] HTML-страницы `/flight-resources`, `/flight-feed`, `/flight-draft` отдают 200, картинки грузятся.
- [ ] Сверить каждый пункт «Global Constraints» с кодом, отчитаться списком «требование → где выполнено».
- [ ] `git status` — изменения не закоммичены, лишних файлов нет (`.env` не в индексе).
