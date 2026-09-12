# ЛР2: PostgreSQL + GORM в flight-service-api — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Перевести приложение с in-memory слайса на PostgreSQL: три таблицы по предметной области, шесть HTTP-методов (пять через GORM, логическое удаление — сырым SQL UPDATE через курсор), три существующие страницы работают на данных из БД.

**Architecture:** Слоистая структура по методичке `lab2-go`: `internal/app/ds` — модели предметной области; `internal/app/repository` — единственный слой, знающий про GORM и SQL; `internal/app/handler` — HTTP-слой на Gin; `internal/pkg` — сборка и запуск; `internal/app/config` (viper + TOML) и `internal/app/dsn` (godotenv + переменные окружения) — конфигурация. `internal/api/server.go` удаляется, его роль переходит к `internal/pkg/app.go`.

**Tech Stack:** Go 1.25, Gin, GORM + `gorm.io/driver/postgres`, PostgreSQL (docker), Adminer (docker), viper, godotenv, logrus, MinIO (с ЛР1, не меняется).

## Global Constraints

Требования из `iu5git/Web/README.md` (ЛР2). Каждая задача обязана их соблюдать.

- **Ровно 3 таблицы**: услуги, лайки (м-м пользователь-услуги), пользователи. Ни одной лишней.
- **Каскадное удаление запрещено** — все внешние ключи с `OnDelete:RESTRICT`.
- **Ровно 6 HTTP-методов**: три GET, POST добавления (ORM), POST публикации (ORM), POST удаления (SQL UPDATE, без ORM).
- **Получение и поиск услуг, создание и публикация — через ORM.**
- **Логическое удаление — SQL-запросом `UPDATE`, без ORM, через `SQL курсор`.**
- **Названия таблиц и полей — по предметной области** (вариант 26: обслуживание рейса в аэропорте).
- **Обязательно 3 услуги** в статусах `черновик`, `удален`, `опубликован`.
- **У каждого пользователя не более одной услуги в статусе `черновик`.**
- **Удалённые услуги просматривать нельзя.**
- **Лайки ставить нельзя** — только отображать из БД.
- **Без JavaScript** (ограничение ЛР1 сохраняется): все действия — HTML-формы + redirect.
- **Ровно два поля по предметной области** в таблице услуг: `unit` (единица измерения) и `price` (цена). Никаких других доменных полей в этой таблице быть не должно.
- **НЕ делать** (материал ЛР3, в методичке `lab2-go` присутствует, но заданием ЛР2 не требуется): таблицу заявок `Message`, м-м `MessageChat`, `GetCartCount`, иконку корзины, страницу заявки, авторизацию, роли модератора, префикс `/api`, загрузку файлов в MinIO из формы.

## Отклонение от TDD — прочитать до начала

В проекте нет тестового харнесса, и задание ЛР2 тестов не требует; заводить его — прямое нарушение «не делать ничего лишнего». Поэтому цикл «красный → зелёный» заменён на **исполняемую верификацию**: каждая задача заканчивается конкретной командой (`go build`, `go run`, `curl`, SQL-запрос) с явно указанным ожидаемым выводом. Шаг верификации так же обязателен, как тест: не переходить к следующей задаче, пока вывод не совпал с ожидаемым.

## File Structure

**Создаются:**

| Файл | Ответственность |
|---|---|
| `config/config.toml` | Хост и порт сервиса (файл конфигурации) |
| `internal/app/config/config.go` | Чтение `config.toml` через viper |
| `internal/app/dsn/dsn.go` | Сборка DSN-строки PostgreSQL из переменных окружения |
| `internal/app/ds/FlightService.go` | Модель услуги обслуживания рейса |
| `internal/app/ds/User.go` | Модель пользователя |
| `internal/app/ds/FlightServiceLike.go` | Модель м-м «пользователь — услуга» (лайки) |
| `internal/app/repository/flight_service.go` | Все запросы к БД: ORM + сырой SQL с курсором |
| `internal/app/handler/flight_service.go` | Шесть обработчиков |
| `internal/pkg/app.go` | Сборка и запуск приложения |
| `cmd/migrate/main.go` | Миграция схемы через `AutoMigrate` |

**Изменяются:**

| Файл | Что меняется |
|---|---|
| `docker-compose.yml` | Добавляются сервисы `postgres` и `adminer`, том `postgres-data` |
| `.env` / `.env.example` | Добавляются `DB_*` |
| `internal/app/repository/repository.go` | Полностью переписывается: слайс-заглушка → подключение GORM |
| `internal/app/handler/handler.go` | Полностью переписывается: регистрация 6 роутов, статики, `errorHandler` |
| `cmd/app/main.go` | Полностью переписывается: config + dsn + repository + handler + pkg |
| `templates/index.html` | Новая ссылка карточки, кнопка логического удаления |
| `templates/feed.html` | Новая ссылка кнопки «следующий» |
| `templates/add.html` | Два состояния: «черновика нет» и «черновик есть» |
| `resources/styles/style.css` | Стили обёртки карточки и кнопки удаления |

**Удаляется:** `internal/api/server.go` (и пустой каталог `internal/api`).

---

### Task 1: Инфраструктура — PostgreSQL и Adminer

**Files:**
- Modify: `docker-compose.yml`
- Modify: `.env`
- Modify: `.env.example`

**Interfaces:**
- Consumes: ничего.
- Produces: БД `flight_service_db` на `localhost:5432`, пользователь `postgres`. Переменные окружения `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME` — их читает `dsn.FromEnv()` в Task 2.

- [ ] **Step 1: Создать ветку**

```bash
git checkout -b lab2-flight-service-db
```

- [ ] **Step 2: Добавить сервисы в `docker-compose.yml`**

Добавить в блок `services:` (после сервиса `minio`) и заменить существующий блок `volumes:` в конце файла:

```yaml
  postgres:
    container_name: flight_postgres
    image: postgres:latest
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASS}
      POSTGRES_DB: flight_service_db
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql

  adminer:
    container_name: flight_adminer
    image: adminer:latest
    ports:
      - "8081:8080"
    depends_on:
      - postgres

volumes:
  minio-data:
  postgres-data:
```

Том монтируется в `/var/lib/postgresql`, а НЕ в `/var/lib/postgresql/data`: в `postgres:latest` (18.x) каталог данных переехал в `/var/lib/postgresql/18/docker`, и монтирование в `.../data` не сохраняло бы БД между пересозданиями контейнера.

Проброс порта Adminer — именно `"8081:8080"`. В методичке `lab2-go` указано `"8081:8081"`, но внутри контейнера Adminer слушает 8080, и с тем пробросом интерфейс не открывается.

- [ ] **Step 3: Добавить переменные в `.env`**

Дописать в конец существующего `.env` (файл уже в `.gitignore`):

```dotenv
DB_HOST=localhost
DB_NAME=flight_service_db
DB_PORT=5432
DB_USER=postgres
DB_PASS=password
```

- [ ] **Step 4: Добавить те же ключи в `.env.example`**

Дописать в конец `.env.example` (этот файл коммитится):

```dotenv
# Параметры подключения к PostgreSQL (ЛР2)
DB_HOST=localhost
DB_NAME=flight_service_db
DB_PORT=5432
DB_USER=postgres
DB_PASS=change-me
```

- [ ] **Step 5: Поднять контейнеры**

Run: `docker compose up -d && docker compose ps`
Expected: три сервиса `minio_storage`, `flight_postgres`, `flight_adminer` в состоянии `Up`.

- [ ] **Step 6: Проверить, что БД принимает подключения**

Run: `docker exec flight_postgres psql -U postgres -d flight_service_db -c "SELECT 1;"`
Expected:
```
 ?column?
----------
        1
(1 row)
```

- [ ] **Step 7: Проверить Adminer**

Run: `curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8081`
Expected: `200`

Открыть `http://localhost:8081` в браузере и войти: система **PostgreSQL**, сервер **postgres**, пользователь **postgres**, пароль **password**, база **flight_service_db**.

- [ ] **Step 8: Подключиться к БД через IDE**

Требование методички — показать подключение из IDE (DBeaver, DataGrip или pgAdmin). Новое соединение → PostgreSQL → хост `localhost`, порт `5432`, база `flight_service_db`, пользователь `postgres`, пароль `password`.
Expected: соединение установлено, в дереве видна база `flight_service_db`. Сделать скриншот — он понадобится на защите.

- [ ] **Step 9: Commit**

```bash
git add docker-compose.yml .env.example
git commit -m "ЛР2: развертывание PostgreSQL и Adminer в docker compose"
```

---

### Task 2: Модели, DSN и миграция

**Files:**
- Create: `internal/app/dsn/dsn.go`
- Create: `internal/app/ds/User.go`
- Create: `internal/app/ds/FlightService.go`
- Create: `internal/app/ds/FlightServiceLike.go`
- Create: `cmd/migrate/main.go`
- Modify: `go.mod`, `go.sum`

**Interfaces:**
- Consumes: переменные окружения `DB_*` из Task 1.
- Produces:
  - `dsn.FromEnv() string`
  - `ds.User{ID uint; Login string; Password string; IsModerator bool}` → таблица `users`
  - `ds.FlightService{ID uint; Name string; Description string; Status string; ImageURL string; VideoURL string; Unit string; Price float64; CreatedAt time.Time; CreatorID uint; FormedAt sql.NullTime; Creator User}` → таблица `flight_services`
  - `ds.FlightServiceLike{ID uint; UserID uint; FlightServiceID uint; User User; FlightService FlightService}` → таблица `flight_service_likes`

- [ ] **Step 1: Установить зависимости**

```bash
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/spf13/viper
go get github.com/joho/godotenv
go mod tidy
```

- [ ] **Step 2: Создать `internal/app/dsn/dsn.go`**

```go
package dsn

import (
	"fmt"
	"os"
)

func FromEnv() string {
	host := os.Getenv("DB_HOST")
	if host == "" {
		return ""
	}
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)
}
```

- [ ] **Step 3: Создать `internal/app/ds/User.go`**

```go
package ds

type User struct {
	ID          uint   `gorm:"primaryKey"`
	Login       string `gorm:"type:varchar(25);unique;not null"`
	Password    string `gorm:"type:varchar(100);not null"`
	IsModerator bool   `gorm:"type:boolean;default:false"`
}
```

- [ ] **Step 4: Создать `internal/app/ds/FlightService.go`**

```go
package ds

import (
	"database/sql"
	"time"
)

// FlightService — услуга обслуживания рейса (ресурс, персонал, техника)
type FlightService struct {
	ID          uint         `gorm:"primaryKey"`
	Name        string       `gorm:"type:varchar(100);not null"`
	Description string       `gorm:"type:varchar(255)"`
	Status      string       `gorm:"type:varchar(15);not null;default:'черновик'"`
	ImageURL    string       `gorm:"type:varchar(200)"`
	VideoURL    string       `gorm:"type:varchar(200)"`
	Unit        string       `gorm:"type:varchar(20)"`
	Price       float64      `gorm:"type:numeric(10,2)"`
	CreatedAt   time.Time    `gorm:"not null"`
	CreatorID   uint         `gorm:"not null"`
	FormedAt    sql.NullTime `gorm:"default:null"`

	Creator User `gorm:"foreignKey:CreatorID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}
```

- [ ] **Step 5: Создать `internal/app/ds/FlightServiceLike.go`**

```go
package ds

// FlightServiceLike — м-м «пользователь — услуга» (лайки):
// первичный ключ ID и два внешних ключа
type FlightServiceLike struct {
	ID              uint `gorm:"primaryKey"`
	UserID          uint `gorm:"not null;uniqueIndex:idx_user_flight_service"`
	FlightServiceID uint `gorm:"not null;uniqueIndex:idx_user_flight_service"`

	User          User          `gorm:"foreignKey:UserID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	FlightService FlightService `gorm:"foreignKey:FlightServiceID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}
```

- [ ] **Step 6: Создать `cmd/migrate/main.go`**

Порядок моделей важен: `User` создаётся первой, на неё ссылаются внешние ключи.

```go
package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"flight-service-api/internal/app/ds"
	"flight-service-api/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.User{},
		&ds.FlightService{},
		&ds.FlightServiceLike{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
```

- [ ] **Step 7: Запустить миграцию**

Run: `go run cmd/migrate/main.go`
Expected: команда завершается без вывода и без паники (код возврата 0).

- [ ] **Step 8: Проверить, что созданы ровно три таблицы**

Run:
```bash
docker exec flight_postgres psql -U postgres -d flight_service_db -c "\dt"
```
Expected: в списке `flight_service_likes`, `flight_services`, `users` — и больше ничего.

- [ ] **Step 9: Проверить типы и длины столбцов услуг**

Run:
```bash
docker exec flight_postgres psql -U postgres -d flight_service_db \
  -c "\d flight_services"
```
Expected: `name` — `character varying(100)`, `status` — `character varying(15)` с `default 'черновик'`, `price` — `numeric(10,2)`, `created_at` — `timestamp with time zone not null`, `formed_at` — nullable `timestamp`, есть внешний ключ на `users`.

- [ ] **Step 10: Проверить, что каскадного удаления нет**

Run:
```bash
docker exec flight_postgres psql -U postgres -d flight_service_db -c \
"SELECT tc.constraint_name, rc.delete_rule FROM information_schema.table_constraints tc
 JOIN information_schema.referential_constraints rc
   ON tc.constraint_name = rc.constraint_name
 WHERE tc.constraint_type = 'FOREIGN KEY';"
```
Expected: три строки, у всех `delete_rule` = `RESTRICT`. Если хоть где-то `CASCADE` — требование нарушено, вернуться к тегам моделей.

- [ ] **Step 11: Commit**

```bash
git add go.mod go.sum internal/app/dsn internal/app/ds cmd/migrate
git commit -m "ЛР2: модели предметной области, DSN и миграция схемы"
```

---

### Task 3: Ограничение на черновик и наполнение БД через Adminer

**Files:** изменений в коде нет — только SQL, выполняемый в Adminer.

**Interfaces:**
- Consumes: таблицы из Task 2.
- Produces: данные, на которых работают все обработчики: услуги с `id` 1–4 (`опубликован`), `id` 5 (`черновик`, `creator_id = 1`), `id` 6 (`удален`); пользователи `id` 1–3; девять строк лайков.

- [ ] **Step 1: Создать частичный уникальный индекс «не более одного черновика»**

Выполнить в Adminer (`http://localhost:8081` → SQL-запрос):

```sql
CREATE UNIQUE INDEX idx_one_draft_per_creator
  ON flight_services (creator_id)
  WHERE status = 'черновик';
```

- [ ] **Step 2: Наполнить таблицы данными**

Выполнить в Adminer одним запросом. Услуги повторяют коллекцию из ЛР1, url ведут в бакет MinIO `flight-media`.

```sql
INSERT INTO users (id, login, password, is_moderator) VALUES
  (1, 'dispatcher', 'dispatcher', false),
  (2, 'moderator',  'moderator',  true),
  (3, 'engineer',   'engineer',   false);

INSERT INTO flight_services
  (id, name, description, status, image_url, video_url, unit, price, created_at, creator_id, formed_at)
VALUES
  (1, 'Аэродромный тягач Goldhofer AST-2X', 'Буксировка и постановка воздушного судна на стоянку.',
   'опубликован', 'http://localhost:9100/flight-media/tug.jpg',
   'http://localhost:9100/flight-media/tug.mp4', 'час', 18500,
   '2026-08-25 10:00:00', 1, '2026-08-25 10:05:00'),
  (2, 'Авиатопливо ТС-1', 'Заправка воздушного судна авиационным керосином.',
   'опубликован', 'http://localhost:9100/flight-media/jet-fuel.jpg',
   'http://localhost:9100/flight-media/jet-fuel.mp4', 'литр', 82,
   '2026-08-25 10:00:00', 1, '2026-08-25 10:05:00'),
  (3, 'Бортовое питание (эконом)', 'Комплект бортового питания на одного пассажира.',
   'опубликован', 'http://localhost:9100/flight-media/catering.jpg',
   'http://localhost:9100/flight-media/catering.mp4', 'порция', 460,
   '2026-08-25 10:00:00', 1, '2026-08-25 10:05:00'),
  (4, 'Наземный источник питания (GPU)', 'Обеспечение самолёта электропитанием на стоянке.',
   'опубликован', 'http://localhost:9100/flight-media/gpu.jpg',
   'http://localhost:9100/flight-media/gpu.mp4', 'час', 9400,
   '2026-08-25 10:00:00', 1, '2026-08-25 10:05:00'),
  (5, 'Багажный тягач', 'Транспортировка багажа между терминалом и самолётом.',
   'черновик', 'http://localhost:9100/flight-media/baggage-tug.jpg',
   'http://localhost:9100/flight-media/baggage-tug.mp4', 'час', 7600,
   '2026-08-25 10:00:00', 1, NULL),
  (6, 'Противообледенительная обработка', 'Обработка воздушного судна противообледенительной жидкостью.',
   'удален', 'http://localhost:9100/flight-media/deicing.jpg',
   'http://localhost:9100/flight-media/deicing.mp4', 'рейс', 15000,
   '2026-08-25 10:00:00', 1, NULL);

INSERT INTO flight_service_likes (user_id, flight_service_id) VALUES
  (1, 1), (2, 1), (3, 1),
  (1, 2), (2, 2),
  (1, 3), (2, 3), (3, 3),
  (1, 4);
```

- [ ] **Step 3: Сдвинуть последовательности первичных ключей**

Строки вставлены с явными `id`. Без этого шага первый же `INSERT` из приложения (кнопка «Далее» в Task 5) упадёт на дублировании первичного ключа.

```sql
SELECT setval('users_id_seq', (SELECT max(id) FROM users));
SELECT setval('flight_services_id_seq', (SELECT max(id) FROM flight_services));
```

- [ ] **Step 4: Проверить наличие трёх обязательных статусов**

Run:
```bash
docker exec flight_postgres psql -U postgres -d flight_service_db -c \
"SELECT status, count(*) FROM flight_services GROUP BY status ORDER BY status;"
```
Expected:
```
   status    | count
-------------+-------
 удален      |     1
 опубликован |     4
 черновик    |     1
```

- [ ] **Step 5: Проверить, что второй черновик создать нельзя**

Run:
```bash
docker exec flight_postgres psql -U postgres -d flight_service_db -c \
"INSERT INTO flight_services (name, status, created_at, creator_id)
 VALUES ('проверка', 'черновик', now(), 1);"
```
Expected: ошибка `duplicate key value violates unique constraint "idx_one_draft_per_creator"`. Строка не вставлена — это и есть подтверждение требования «не более одной услуги в статусе черновик у пользователя».

- [ ] **Step 6: Проверить, что каскадного удаления нет на практике**

Run:
```bash
docker exec flight_postgres psql -U postgres -d flight_service_db -c \
"DELETE FROM users WHERE id = 1;"
```
Expected: ошибка вида `update or delete on table "users" violates RESTRICT setting of foreign key constraint` (на PostgreSQL 18 формулировка именно такая; на более старых версиях — `violates foreign key constraint`). Пользователь не удалён, услуги и лайки на месте.

- [ ] **Step 7: Сохранить SQL наполнения в репозиторий**

Записать SQL из шагов 1–3 в файл `docs/lab2-seed.sql` — он понадобится, чтобы воспроизвести данные на защите.

```bash
git add docs/lab2-seed.sql
git commit -m "ЛР2: индекс единственного черновика и SQL наполнения БД"
```

---

### Task 4: Переход на GORM — три GET-метода на данных из БД

Самая крупная задача: приложение перестаёт зависеть от слайса-заглушки. Здесь же **два GET-обработчика ленты объединяются в один** — сейчас лента живёт на `/flight-feed` (без ID) и `/flight-resource/:id` (по ID), это два GET-метода из четырёх, а задание требует ровно три GET. Один обработчик `GetFlightFeed` регистрируется на двух путях: `/flight-feed` и `/flight-feed/:id`, что соответствует формулировке ЛР1 «для панели вкладок `ID` не указывается».

**Files:**
- Create: `config/config.toml`
- Create: `internal/app/config/config.go`
- Create: `internal/app/repository/flight_service.go`
- Create: `internal/app/handler/flight_service.go`
- Create: `internal/pkg/app.go`
- Modify: `internal/app/repository/repository.go` (переписать полностью)
- Modify: `internal/app/handler/handler.go` (переписать полностью)
- Modify: `cmd/app/main.go` (переписать полностью)
- Modify: `templates/index.html:28` (ссылка карточки)
- Modify: `templates/feed.html:25` (ссылка кнопки «следующий»)
- Delete: `internal/api/server.go`

**Interfaces:**
- Consumes: `dsn.FromEnv()`, модели `ds.*` из Task 2; данные из Task 3.
- Produces:
  - `config.NewConfig() (*config.Config, error)`; `config.Config{ServiceHost string; ServicePort int}`
  - `repository.New(dsn string) (*repository.Repository, error)`
  - Константы `repository.CreatorID = 1`, `repository.StatusDraft`, `repository.StatusPublished`, `repository.StatusDeleted`
  - `(*Repository) GetPublishedFlightServices() ([]ds.FlightService, error)`
  - `(*Repository) GetFlightServicesByPrice(maxPrice float64) ([]ds.FlightService, error)`
  - `(*Repository) GetFlightService(id uint) (ds.FlightService, error)`
  - `(*Repository) GetNextFlightServiceID(id uint) (uint, error)`
  - `(*Repository) GetDraftFlightService() (*ds.FlightService, error)`
  - `(*Repository) GetLikesCounts() (map[uint]int64, error)`
  - `(*Repository) GetLikesCount(flightServiceID uint) (int64, error)`
  - `(*Handler) GetFlightServices`, `(*Handler) GetFlightFeed`, `(*Handler) GetFlightDraft` — `gin.HandlerFunc`
  - `(*Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error)` — используется задачами 5–7
  - `handler.NewHandler(r *repository.Repository) *Handler`, `(*Handler) RegisterHandler(router *gin.Engine)`, `(*Handler) RegisterStatic(router *gin.Engine)`
  - `pkg.NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *pkg.Application`, `(*Application) RunApp()`

- [ ] **Step 1: Создать `config/config.toml`**

```toml
ServiceHost = "0.0.0.0"
ServicePort = 8080
```

- [ ] **Step 2: Создать `internal/app/config/config.go`**

```go
package config

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
}

func NewConfig() (*Config, error) {
	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	log.Info("config parsed")

	return cfg, nil
}
```

- [ ] **Step 3: Переписать `internal/app/repository/repository.go`**

Файл заменяется целиком — слайс-заглушка и структура `FlightService` из ЛР1 удаляются, модель теперь живёт в `ds`.

```go
package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// CreatorID — пользователь пока зафиксирован, авторизация появится в ЛР4
const CreatorID = 1

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}
```

- [ ] **Step 4: Создать `internal/app/repository/flight_service.go`**

Здесь только методы чтения. Методы записи добавляются в задачах 5–7.

```go
package repository

import (
	"errors"

	"gorm.io/gorm"

	"flight-service-api/internal/app/ds"
)

const (
	StatusDraft     = "черновик"
	StatusPublished = "опубликован"
	StatusDeleted   = "удален"
)

// GetPublishedFlightServices — список опубликованных услуг (ORM)
func (r *Repository) GetPublishedFlightServices() ([]ds.FlightService, error) {
	var services []ds.FlightService
	err := r.db.Where("status = ?", StatusPublished).Order("id").Find(&services).Error
	if err != nil {
		return nil, err
	}

	return services, nil
}

// GetFlightServicesByPrice — поиск-фильтрация по цене (ORM)
func (r *Repository) GetFlightServicesByPrice(maxPrice float64) ([]ds.FlightService, error) {
	var services []ds.FlightService
	err := r.db.Where("status = ? AND price <= ?", StatusPublished, maxPrice).
		Order("id").Find(&services).Error
	if err != nil {
		return nil, err
	}

	return services, nil
}

// GetFlightService — одна услуга; удалённые просматривать нельзя (ORM)
func (r *Repository) GetFlightService(id uint) (ds.FlightService, error) {
	var service ds.FlightService
	err := r.db.Where("id = ? AND status <> ?", id, StatusDeleted).First(&service).Error
	if err != nil {
		return ds.FlightService{}, err
	}

	return service, nil
}

// GetNextFlightServiceID — следующая опубликованная услуга по кругу (ORM)
func (r *Repository) GetNextFlightServiceID(id uint) (uint, error) {
	var service ds.FlightService

	err := r.db.Where("status = ? AND id > ?", StatusPublished, id).
		Order("id").First(&service).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.Where("status = ?", StatusPublished).Order("id").First(&service).Error
	}
	if err != nil {
		return 0, err
	}

	return service.ID, nil
}

// GetDraftFlightService — черновик текущего пользователя; nil, если его нет (ORM)
func (r *Repository) GetDraftFlightService() (*ds.FlightService, error) {
	var service ds.FlightService
	err := r.db.Where("creator_id = ? AND status = ?", CreatorID, StatusDraft).First(&service).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &service, nil
}

// GetLikesCounts — количество лайков по всем услугам (для плитки)
func (r *Repository) GetLikesCounts() (map[uint]int64, error) {
	type likeRow struct {
		FlightServiceID uint
		Count           int64
	}

	var rows []likeRow
	err := r.db.Model(&ds.FlightServiceLike{}).
		Select("flight_service_id, count(*) as count").
		Group("flight_service_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uint]int64, len(rows))
	for _, row := range rows {
		counts[row.FlightServiceID] = row.Count
	}

	return counts, nil
}

// GetLikesCount — количество лайков одной услуги (для ленты)
func (r *Repository) GetLikesCount(flightServiceID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.FlightServiceLike{}).
		Where("flight_service_id = ?", flightServiceID).Count(&count).Error

	return count, err
}
```

- [ ] **Step 5: Переписать `internal/app/handler/handler.go`**

Роуты POST закомментированы не будут — они добавляются в задачах 5–7. Здесь регистрируются только GET.

```go
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"flight-service-api/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// RegisterHandler регистрирует маршруты приложения
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/flight-resources", h.GetFlightServices)
	router.GET("/flight-feed", h.GetFlightFeed)
	router.GET("/flight-feed/:id", h.GetFlightFeed)
	router.GET("/flight-draft", h.GetFlightDraft)
}

// RegisterStatic регистрирует шаблоны и статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

// errorHandler для более удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
```

- [ ] **Step 6: Создать `internal/app/handler/flight_service.go`**

`GetFlightFeed` — один обработчик на оба пути: если `:id` не передан, он равен нулю, и берётся первая опубликованная услуга.

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"flight-service-api/internal/app/ds"
)

type flightServiceView struct {
	FlightService ds.FlightService
	LikesCount    int64
}

// GetFlightServices — плитка карточек с фильтрацией по цене
func (h *Handler) GetFlightServices(ctx *gin.Context) {
	var services []ds.FlightService
	var err error

	priceQuery := ctx.Query("price")
	if priceQuery == "" {
		services, err = h.Repository.GetPublishedFlightServices()
	} else {
		maxPrice, parseErr := strconv.ParseFloat(priceQuery, 64)
		if parseErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, parseErr)
			return
		}
		services, err = h.Repository.GetFlightServicesByPrice(maxPrice)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	counts, err := h.Repository.GetLikesCounts()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	views := make([]flightServiceView, 0, len(services))
	for _, service := range services {
		views = append(views, flightServiceView{
			FlightService: service,
			LikesCount:    counts[service.ID],
		})
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"flightServices": views,
		"price":          priceQuery,
	})
}

// GetFlightFeed — лента: /flight-feed и /flight-feed/:id (+ ?next=true)
func (h *Handler) GetFlightFeed(ctx *gin.Context) {
	var id uint

	if idStr := ctx.Param("id"); idStr != "" {
		parsed, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		id = uint(parsed)
	}

	if id == 0 || ctx.Query("next") == "true" {
		next, err := h.Repository.GetNextFlightServiceID(id)
		if err != nil {
			h.errorHandler(ctx, http.StatusNotFound, err)
			return
		}
		id = next
	}

	service, err := h.Repository.GetFlightService(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	likesCount, err := h.Repository.GetLikesCount(service.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"flightService": service,
		"likesCount":    likesCount,
	})
}

// GetFlightDraft — страница добавления
func (h *Handler) GetFlightDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftFlightService()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"flightService": draft,
	})
}
```

- [ ] **Step 7: Создать `internal/pkg/app.go`**

```go
package pkg

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"flight-service-api/internal/app/config"
	"flight-service-api/internal/app/handler"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Server down")
}
```

- [ ] **Step 8: Переписать `cmd/app/main.go`**

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"flight-service-api/internal/app/config"
	"flight-service-api/internal/app/dsn"
	"flight-service-api/internal/app/handler"
	"flight-service-api/internal/app/repository"
	"flight-service-api/internal/pkg"
)

func main() {
	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	rep, err := repository.New(dsn.FromEnv())
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	hand := handler.NewHandler(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
```

- [ ] **Step 9: Удалить старый запуск сервера**

```bash
rm internal/api/server.go
rmdir internal/api
```

- [ ] **Step 10: Обновить ссылку карточки в `templates/index.html`**

Строка 28. Было:

```html
      <a class="card" href="/flight-resource/{{ .FlightService.ID }}">
```

Стало:

```html
      <a class="card" href="/flight-feed/{{ .FlightService.ID }}">
```

- [ ] **Step 11: Обновить ссылку кнопки «следующий» в `templates/feed.html`**

Строка 25. Было:

```html
          <a class="btn-next" href="/flight-resource/{{ .flightService.ID }}?next=true">→</a>
```

Стало:

```html
          <a class="btn-next" href="/flight-feed/{{ .flightService.ID }}?next=true">→</a>
```

- [ ] **Step 12: Собрать проект**

Run: `go build ./... && go vet ./...`
Expected: пустой вывод, код возврата 0. Если есть ссылки на удалённый `internal/api` — исправить импорты.

- [ ] **Step 13: Запустить приложение в фоне**

Проверочные шаги 14–20 обращаются к работающему серверу, поэтому запускать нужно фоном, а лог писать в файл.

```bash
go run cmd/app/main.go > /tmp/flight-app.log 2>&1 &
sleep 3 && grep -E "config parsed|Server start up|GIN-debug.*flight" /tmp/flight-app.log
```
Expected: в логе `config parsed`, `Server start up` и четыре строки маршрутов `GET /flight-resources`, `GET /flight-feed`, `GET /flight-feed/:id`, `GET /flight-draft`.

Останавливать процесс между задачами: `pkill -f "cmd/app/main.go"`.

- [ ] **Step 14: Проверить плитку — данные приходят из БД**

Run: `curl -s http://localhost:8080/flight-resources | grep -c 'class="card"'`
Expected: `4` — ровно четыре опубликованные услуги. Черновик и удалённая не показываются.

- [ ] **Step 15: Проверить фильтрацию по цене**

Run: `curl -s "http://localhost:8080/flight-resources?price=1000" | grep -c 'class="card"'`
Expected: `2` — услуги с ценой 82 и 460.

- [ ] **Step 16: Проверить, что счётчик лайков берётся из БД**

Run: `curl -s http://localhost:8080/flight-resources | grep -o '♥ [0-9]*'`
Expected: `♥ 3`, `♥ 2`, `♥ 3`, `♥ 1` — совпадает с наполнением таблицы `flight_service_likes`.

- [ ] **Step 17: Проверить ленту без ID и по ID**

Run:
```bash
curl -s -o /dev/null -w "feed:%{http_code}\n" http://localhost:8080/flight-feed
curl -s -o /dev/null -w "feed/2:%{http_code}\n" http://localhost:8080/flight-feed/2
```
Expected: `feed:200`, `feed/2:200`

- [ ] **Step 18: Проверить переход к следующей услуге**

Run: `curl -s "http://localhost:8080/flight-feed/2?next=true" | grep 'feed-name'`
Expected: строка с названием `Бортовое питание (эконом)` — услуга с `id = 3`.

- [ ] **Step 19: Проверить, что удалённую услугу открыть нельзя**

Run: `curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/flight-feed/6`
Expected: `404`

- [ ] **Step 20: Проверить страницу черновика**

Run: `curl -s http://localhost:8080/flight-draft | grep -c 'Багажный тягач'`
Expected: `1` — на странице добавления открыт черновик из БД.

- [ ] **Step 21: Commit**

`git add -A` нужен, чтобы в коммит попало удаление `internal/api/server.go`.

```bash
git add -A
git commit -m "ЛР2: подключение БД через GORM, три GET-метода, объединение обработчиков ленты"
```

---

### Task 5: POST создания черновика через ORM

**Files:**
- Modify: `internal/app/repository/flight_service.go` (добавить метод)
- Modify: `internal/app/handler/flight_service.go` (добавить обработчик)
- Modify: `internal/app/handler/handler.go` (добавить роут)
- Modify: `templates/add.html` (переписать блок `.form` — два состояния)

**Interfaces:**
- Consumes: `repository.CreatorID`, `repository.StatusDraft`, `(*Repository) GetDraftFlightService()`, `(*Handler) errorHandler` из Task 4.
- Produces:
  - `(*Repository) CreateDraftFlightService(name, imageURL, videoURL string) error`
  - `(*Handler) CreateFlightDraft(ctx *gin.Context)` на `POST /flight-draft`
  - Форма публикации в `add.html`, отправляющая `description`, `unit`, `price` на `POST /flight-publish` (роут появится в Task 6).

- [ ] **Step 1: Добавить метод в `internal/app/repository/flight_service.go`**

Дописать в конец файла. Также добавить `"time"` в блок импортов.

```go
// CreateDraftFlightService — создание черновика по кнопке «Далее» (ORM)
func (r *Repository) CreateDraftFlightService(name, imageURL, videoURL string) error {
	service := ds.FlightService{
		Name:      name,
		ImageURL:  imageURL,
		VideoURL:  videoURL,
		Status:    StatusDraft,
		CreatorID: CreatorID,
		CreatedAt: time.Now(),
	}

	return r.db.Create(&service).Error
}
```

- [ ] **Step 2: Добавить обработчик в `internal/app/handler/flight_service.go`**

Дописать в конец файла. Также добавить `"fmt"` в блок импортов.

Проверка существующего черновика дублирует частичный уникальный индекс из Task 3 — индекс защищает данные, а эта проверка даёт понятную ошибку вместо 500 от БД.

```go
// CreateFlightDraft — кнопка «Далее»: создание черновика
func (h *Handler) CreateFlightDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftFlightService()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if draft != nil {
		h.errorHandler(ctx, http.StatusConflict, fmt.Errorf("у пользователя уже есть черновик"))
		return
	}

	err = h.Repository.CreateDraftFlightService(
		ctx.PostForm("name"),
		ctx.PostForm("image_url"),
		ctx.PostForm("video_url"),
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/flight-draft")
}
```

- [ ] **Step 3: Зарегистрировать роут в `internal/app/handler/handler.go`**

В `RegisterHandler` после блока GET-маршрутов добавить:

```go
	router.POST("/flight-draft", h.CreateFlightDraft)
```

- [ ] **Step 4: Переписать блок `.form` в `templates/add.html`**

Заменить содержимое `<div class="form"> ... </div>` (строки 16–48) целиком. Обработчик передаёт `nil`, когда черновика нет, поэтому нужен `{{ if .flightService }}`.

```html
    <div class="form">
      {{ if .flightService }}
      <form action="/flight-publish" method="POST">
        <p class="section-title">Изображение</p>
        <div class="upload">
          <img src="{{ .flightService.ImageURL }}" alt="{{ .flightService.Name }}">
        </div>

        <p class="section-title">Видео</p>
        <div class="upload">
          <video src="{{ .flightService.VideoURL }}" controls></video>
        </div>

        <p class="section-title">Текст</p>
        <div class="field">
          <label>Название</label>
          <input type="text" value="{{ .flightService.Name }}" readonly>
        </div>
        <div class="field">
          <label>Краткое описание</label>
          <input type="text" name="description" value="{{ .flightService.Description }}" required>
        </div>

        <p class="section-title">Параметры</p>
        <div class="field-row">
          <div class="field">
            <label>Единица измерения</label>
            <input type="text" name="unit" value="{{ .flightService.Unit }}" required>
          </div>
          <div class="field">
            <label>Цена</label>
            <input type="number" step="0.01" name="price" value="{{ .flightService.Price }}" required>
          </div>
        </div>

        <button type="submit" class="btn-primary">Опубликовать</button>
      </form>
      {{ else }}
      <form action="/flight-draft" method="POST">
        <p class="section-title">Изображение</p>
        <div class="field">
          <label>URL изображения</label>
          <input type="text" name="image_url" required>
        </div>

        <p class="section-title">Видео</p>
        <div class="field">
          <label>URL видео</label>
          <input type="text" name="video_url" required>
        </div>

        <p class="section-title">Текст</p>
        <div class="field">
          <label>Название</label>
          <input type="text" name="name" required>
        </div>

        <button type="submit" class="btn-primary">Далее</button>
      </form>
      {{ end }}
    </div>
```

Кнопка «Опубликовать» до Task 6 вернёт 404 — это ожидаемо, роут `/flight-publish` появится в следующей задаче.

- [ ] **Step 5: Добавить стиль кнопки в `resources/styles/style.css`**

Дописать в конец файла:

```css
.btn-primary {
  display: block;
  width: 100%;
  margin: 16px 0 0;
  padding: 12px;
  border: none;
  border-radius: 10px;
  background: var(--blue);
  color: var(--white);
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
}
```

- [ ] **Step 6: Пересобрать и перезапустить**

```bash
pkill -f "cmd/app/main.go"
go build ./... && go run cmd/app/main.go > /tmp/flight-app.log 2>&1 &
sleep 3 && grep "POST.*flight-draft" /tmp/flight-app.log
```
Expected: сборка без ошибок, в логе Gin строка `POST /flight-draft`.

- [ ] **Step 7: Проверить, что второй черновик создать нельзя**

Черновик `id = 5` уже есть в БД с Task 3.

```bash
curl -s -X POST http://localhost:8080/flight-draft \
  -d "name=Проверка" -d "image_url=http://x/i.jpg" -d "video_url=http://x/v.mp4"
```
Expected: `{"description":"у пользователя уже есть черновик","status":"error"}` — HTTP 409.

- [ ] **Step 8: Удалить черновик и проверить создание**

```bash
docker exec flight_postgres psql -U postgres -d flight_service_db -c \
  "DELETE FROM flight_services WHERE id = 5;"
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/flight-draft \
  -d "name=Багажный тягач" \
  -d "image_url=http://localhost:9100/flight-media/baggage-tug.jpg" \
  -d "video_url=http://localhost:9100/flight-media/baggage-tug.mp4"
```
Expected: `302` — редирект на `/flight-draft`.

- [ ] **Step 9: Проверить строку в БД**

Run:
```bash
docker exec flight_postgres psql -U postgres -d flight_service_db -c \
"SELECT id, name, status, creator_id, created_at IS NOT NULL AS has_created,
        formed_at IS NULL AS formed_is_null
 FROM flight_services WHERE status = 'черновик';"
```
Expected: одна строка, `status` = `черновик`, `creator_id` = 1, `has_created` = `t`, `formed_is_null` = `t`.

- [ ] **Step 10: Проверить, что страница открылась с заполненными полями**

Run: `curl -s http://localhost:8080/flight-draft | grep -c 'Опубликовать'`
Expected: `1` — страница показывает заполненный черновик и кнопку «Опубликовать».

- [ ] **Step 11: Commit**

```bash
git add internal templates resources
git commit -m "ЛР2: POST создания черновика через ORM и два состояния страницы добавления"
```

---

### Task 6: POST публикации через ORM

**Files:**
- Modify: `internal/app/repository/flight_service.go` (добавить метод)
- Modify: `internal/app/handler/flight_service.go` (добавить обработчик)
- Modify: `internal/app/handler/handler.go` (добавить роут)

**Interfaces:**
- Consumes: `repository.CreatorID`, `repository.StatusDraft`, `repository.StatusPublished`, форма публикации из Task 5.
- Produces:
  - `(*Repository) PublishFlightService(description, unit string, price float64) error`
  - `(*Handler) PublishFlightService(ctx *gin.Context)` на `POST /flight-publish`

- [ ] **Step 1: Добавить метод в `internal/app/repository/flight_service.go`**

Дописать в конец файла. Также добавить `"fmt"` в блок импортов.

```go
// PublishFlightService — публикация черновика по кнопке «Опубликовать» (ORM)
func (r *Repository) PublishFlightService(description, unit string, price float64) error {
	res := r.db.Model(&ds.FlightService{}).
		Where("creator_id = ? AND status = ?", CreatorID, StatusDraft).
		Updates(map[string]interface{}{
			"description": description,
			"unit":        unit,
			"price":       price,
			"status":      StatusPublished,
			"formed_at":   time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("черновик не найден")
	}

	return nil
}
```

- [ ] **Step 2: Добавить обработчик в `internal/app/handler/flight_service.go`**

Дописать в конец файла.

```go
// PublishFlightService — кнопка «Опубликовать»: смена статуса черновика
func (h *Handler) PublishFlightService(ctx *gin.Context) {
	price, err := strconv.ParseFloat(ctx.PostForm("price"), 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.PublishFlightService(
		ctx.PostForm("description"),
		ctx.PostForm("unit"),
		price,
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/flight-resources")
}
```

- [ ] **Step 3: Зарегистрировать роут в `internal/app/handler/handler.go`**

В `RegisterHandler` после `router.POST("/flight-draft", ...)` добавить:

```go
	router.POST("/flight-publish", h.PublishFlightService)
```

- [ ] **Step 4: Пересобрать и перезапустить**

```bash
pkill -f "cmd/app/main.go"
go build ./... && go run cmd/app/main.go > /tmp/flight-app.log 2>&1 &
sleep 3 && grep "POST.*flight-publish" /tmp/flight-app.log
```
Expected: сборка без ошибок, в логе Gin строка `POST /flight-publish`.

- [ ] **Step 5: Опубликовать черновик**

```bash
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/flight-publish \
  -d "description=Транспортировка багажа между терминалом и самолётом." \
  -d "unit=час" -d "price=7600"
```
Expected: `302` — редирект на `/flight-resources`.

- [ ] **Step 6: Проверить смену статуса и дату формирования**

Run:
```bash
docker exec flight_postgres psql -U postgres -d flight_service_db -c \
"SELECT id, name, status, unit, price, formed_at IS NOT NULL AS has_formed
 FROM flight_services WHERE name = 'Багажный тягач';"
```
Expected: `status` = `опубликован`, `unit` = `час`, `price` = `7600.00`, `has_formed` = `t`.

- [ ] **Step 7: Проверить, что черновиков не осталось**

Run: `curl -s http://localhost:8080/flight-draft | grep -c 'Далее'`
Expected: `1` — страница переключилась в состояние «черновика нет», показана форма с кнопкой «Далее».

- [ ] **Step 8: Проверить, что услуга появилась в плитке**

Run: `curl -s http://localhost:8080/flight-resources | grep -c 'class="card"'`
Expected: `5` — было четыре опубликованные, добавилась пятая.

- [ ] **Step 9: Проверить, что повторная публикация даёт ошибку**

```bash
curl -s -X POST http://localhost:8080/flight-publish \
  -d "description=x" -d "unit=час" -d "price=1"
```
Expected: `{"description":"черновик не найден","status":"error"}`

- [ ] **Step 10: Commit**

```bash
git add internal
git commit -m "ЛР2: POST публикации услуги через ORM"
```

---

### Task 7: POST логического удаления — SQL UPDATE через курсор

Единственный метод, где ORM не используется: запрос пишется вручную, GORM выступает только драйвером соединения. `RETURNING id` нужен, чтобы у `UPDATE` была возвращаемая строка, которую можно прочитать курсором `Raw().Row()`.

**Files:**
- Modify: `internal/app/repository/flight_service.go` (добавить метод)
- Modify: `internal/app/handler/flight_service.go` (добавить обработчик)
- Modify: `internal/app/handler/handler.go` (добавить роут)
- Modify: `templates/index.html` (обёртка карточки и форма удаления)
- Modify: `resources/styles/style.css` (стили обёртки и кнопки)

**Interfaces:**
- Consumes: `repository.StatusDeleted`, `(*Handler) errorHandler`.
- Produces:
  - `(*Repository) DeleteFlightService(id uint) error`
  - `(*Handler) DeleteFlightService(ctx *gin.Context)` на `POST /flight-delete`

- [ ] **Step 1: Добавить метод в `internal/app/repository/flight_service.go`**

Дописать в конец файла. Также добавить `"database/sql"` в блок импортов.

```go
// DeleteFlightService — логическое удаление: SQL UPDATE через курсор, без ORM
func (r *Repository) DeleteFlightService(id uint) error {
	query := `UPDATE flight_services SET status = $1 WHERE id = $2 AND status <> $1 RETURNING id`

	// Создание курсора (строковый указатель)
	row := r.db.Raw(query, StatusDeleted, id).Row()

	var deletedID uint
	if err := row.Scan(&deletedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("услуга %d не найдена или уже удалена", id)
		}
		return err
	}

	return nil
}
```

- [ ] **Step 2: Добавить обработчик в `internal/app/handler/flight_service.go`**

Дописать в конец файла.

```go
// DeleteFlightService — логическое удаление услуги с плитки
func (h *Handler) DeleteFlightService(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.PostForm("flight_service_id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err = h.Repository.DeleteFlightService(uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/flight-resources")
}
```

- [ ] **Step 3: Зарегистрировать роут в `internal/app/handler/handler.go`**

В `RegisterHandler` после `router.POST("/flight-publish", ...)` добавить:

```go
	router.POST("/flight-delete", h.DeleteFlightService)
```

- [ ] **Step 4: Добавить кнопку удаления в `templates/index.html`**

Заменить блок `<div class="grid"> ... </div>` целиком. Форма — **сосед** ссылки, а не вложена в неё: вложенная форма внутри `<a>` невалидна, и клик по кнопке перехватывался бы переходом по ссылке.

```html
    <div class="grid">
      {{ range .flightServices }}
      <div class="card-wrap">
        <a class="card" href="/flight-feed/{{ .FlightService.ID }}">
          <img src="{{ .FlightService.ImageURL }}" alt="{{ .FlightService.Name }}">
          <div class="card-body">
            <p class="name">{{ .FlightService.Name }}</p>
            <p class="price">{{ .FlightService.Price }} ₽/{{ .FlightService.Unit }}</p>
            <span class="likes">♥ {{ .LikesCount }}</span>
          </div>
        </a>
        <form class="card-delete" action="/flight-delete" method="POST">
          <input type="hidden" name="flight_service_id" value="{{ .FlightService.ID }}">
          <button type="submit" class="btn-delete">✕</button>
        </form>
      </div>
      {{ end }}
    </div>
```

- [ ] **Step 5: Добавить стили в `resources/styles/style.css`**

Дописать в конец файла. `.card` — это `<a>`, внутри обёртки он перестаёт быть элементом grid и нуждается в `display: block`, иначе карточка схлопнется.

```css
.card-wrap { position: relative; }
.card-wrap .card { display: block; height: 100%; }
.card-delete { position: absolute; top: 8px; right: 8px; margin: 0; }
.btn-delete {
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 50%;
  background: rgba(4, 24, 57, 0.55);
  color: var(--white);
  font-size: 13px;
  line-height: 1;
  cursor: pointer;
}
```

- [ ] **Step 6: Пересобрать и перезапустить**

```bash
pkill -f "cmd/app/main.go"
go build ./... && go run cmd/app/main.go > /tmp/flight-app.log 2>&1 &
sleep 3 && grep "POST.*flight-delete" /tmp/flight-app.log
```
Expected: сборка без ошибок, в логе Gin строка `POST /flight-delete`.

- [ ] **Step 7: Удалить услугу**

```bash
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/flight-delete \
  -d "flight_service_id=4"
```
Expected: `302`

- [ ] **Step 8: Проверить смену статуса в БД**

Run:
```bash
docker exec flight_postgres psql -U postgres -d flight_service_db -c \
"SELECT id, name, status FROM flight_services WHERE id = 4;"
```
Expected: `status` = `удален`. Строка на месте — удаление логическое, физического `DELETE` не произошло.

- [ ] **Step 9: Проверить, что услуга пропала из плитки**

Run: `curl -s http://localhost:8080/flight-resources | grep -c 'class="card"'`
Expected: `4` — было пять, стало четыре.

- [ ] **Step 10: Проверить, что удалённую услугу нельзя открыть**

Run: `curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/flight-feed/4`
Expected: `404`

- [ ] **Step 11: Проверить повторное удаление**

```bash
curl -s -X POST http://localhost:8080/flight-delete -d "flight_service_id=4"
```
Expected: `{"description":"услуга 4 не найдена или уже удалена","status":"error"}` — сработала ветка `sql.ErrNoRows`, то есть курсор действительно прочитан.

- [ ] **Step 12: Проверить кнопку в браузере**

Открыть `http://localhost:8080/flight-resources`, нажать ✕ на карточке. Ожидается: страница перезагрузилась, карточка исчезла. Клик по остальной площади карточки по-прежнему открывает ленту.

- [ ] **Step 13: Commit**

```bash
git add internal templates resources
git commit -m "ЛР2: логическое удаление услуги через SQL UPDATE и курсор"
```

---

### Task 8: Финальная проверка и материалы к защите

**Files:**
- Create: `docs/lab2-control-questions.md`
- Modify: файл диаграмм StarUML (общий на весь курс)

**Interfaces:**
- Consumes: работающее приложение из задач 1–7.
- Produces: комплект материалов к показу.

- [ ] **Step 1: Проверить, что HTTP-методов ровно шесть**

Run: `grep -n 'router\.\(GET\|POST\)' internal/app/handler/handler.go`
Expected: семь строк регистрации — это шесть логических методов, где лента зарегистрирована на двух путях (`/flight-feed` и `/flight-feed/:id`) одним и тем же обработчиком `GetFlightFeed`. Список должен быть ровно таким:

```
GET    /flight-resources
GET    /flight-feed
GET    /flight-feed/:id
GET    /flight-draft
POST   /flight-draft
POST   /flight-publish
POST   /flight-delete
```

- [ ] **Step 2: Проверить, что ORM используется в пяти методах, а сырой SQL — в одном**

Run: `grep -n "r.db.Raw" internal/app/repository/flight_service.go`
Expected: ровно одно вхождение — в `DeleteFlightService`. Все остальные методы используют `r.db.Where/Find/First/Create/Model/Updates/Count`.

- [ ] **Step 3: Проверить, что заглушек из ЛР1 не осталось**

Run: `grep -rn "имитируем работу с БД\|массив пустой" internal/`
Expected: пустой вывод — данные берутся только из БД.

- [ ] **Step 4: Восстановить состояние БД для показа**

Требуется наличие трёх услуг в статусах `черновик`, `опубликован`, `удален`. После проверок задач 5–7 черновика нет. Выполнить в Adminer:

```sql
INSERT INTO flight_services
  (name, description, status, image_url, video_url, unit, price, created_at, creator_id, formed_at)
VALUES
  ('Уборка салона', 'Полная уборка пассажирского салона перед рейсом.',
   'черновик', 'http://localhost:9100/flight-media/cleaning.jpg',
   'http://localhost:9100/flight-media/cleaning.mp4', 'рейс', 5200,
   now(), 1, NULL);
```

Затем проверить:

```bash
docker exec flight_postgres psql -U postgres -d flight_service_db -c \
"SELECT status, count(*) FROM flight_services GROUP BY status ORDER BY status;"
```
Expected: присутствуют все три статуса, `черновик` — ровно один.

- [ ] **Step 5: Построить ER-диаграмму в StarUML**

В общем файле диаграмм курса: Model → Add Diagram → Entity-Relationship Diagram.

Три сущности с полным перечнем столбцов, типов и длин:

| Таблица | Столбцы |
|---|---|
| `users` | `id serial` **PK**; `login varchar(25)` NOT NULL UNIQUE; `password varchar(100)` NOT NULL; `is_moderator boolean` DEFAULT false |
| `flight_services` | `id serial` **PK**; `name varchar(100)` NOT NULL; `description varchar(255)`; `status varchar(15)` NOT NULL DEFAULT 'черновик'; `image_url varchar(200)`; `video_url varchar(200)`; `unit varchar(20)`; `price numeric(10,2)`; `created_at timestamp` NOT NULL; `creator_id integer` **FK** → `users.id`; `formed_at timestamp` NULL |
| `flight_service_likes` | `id serial` **PK**; `user_id integer` **FK** → `users.id`; `flight_service_id integer` **FK** → `flight_services.id` |

Связи и кардинальность:
- `users` 1 → ∞ `flight_services` по `creator_id` («создатель услуги»)
- `users` 1 → ∞ `flight_service_likes` по `user_id`
- `flight_services` 1 → ∞ `flight_service_likes` по `flight_service_id`

Две последние связи вместе образуют м-м «пользователь ↔ услуга». На всех связях отметить ограничивающее удаление (RESTRICT) — каскадное удаление запрещено.

- [ ] **Step 6: Записать письменные ответы на контрольные вопросы**

Создать `docs/lab2-control-questions.md` с ответами по шести темам: виды БД; SQL-запросы; курсоры; ORM; модель и миграции; чистая архитектура. Опорные ответы — в разделе 14 гайда `~/Desktop/ЛР2 — гайд (flight-service-api).pdf`. Ответы про курсор и чистую архитектуру привязать к своему коду: `DeleteFlightService` с `Raw().Row()` и разделение `ds` / `repository` / `handler` / `pkg`.

- [ ] **Step 7: Подготовить скриншоты 1–21**

Порядок показа — раздел 13 гайда. Кратко:
- 1–2: Adminer, логическое удаление сменой статуса + `SELECT`
- 3–10: три страницы — поиск по цене, удаление с плитки, переход по url удалённой (404), добавление, `SELECT`, публикация, `SELECT`, лента с кнопкой «следующий»
- 11–13: правка `unit`/`price` и строк `flight_service_likes` в Adminer → изменения в приложении
- 14–21: код — три модели, `AutoMigrate`, пять контроллеров через ORM, шестой через сырой `UPDATE` с курсором

- [ ] **Step 8: Подготовить .doc с исправлениями по ЛР1**

Каждое замечание из таблицы по ЛР1 прокомментировать текстом и приложить скриншот, подтверждающий исправление.

- [ ] **Step 9: Финальная сборка и пуш ветки**

```bash
go build ./... && go vet ./...
git add docs
git commit -m "ЛР2: ответы на контрольные вопросы"
git push -u origin lab2-flight-service-db
```

---

## Проверка соответствия заданию

Пройти перед защитой.

| Требование | Где выполнено | Задача |
|---|---|---|
| 3 таблицы, названия по предметной области | `flight_services`, `flight_service_likes`, `users` | 2 |
| Каскадное удаление запрещено | `constraint:OnDelete:RESTRICT` у всех FK; проверено SQL | 2, 3 |
| м-м лайки: первичный и два внешних ключа | `ds.FlightServiceLike` | 2 |
| Все обязательные поля услуги | name, description, status, image_url, video_url, unit, price, created_at, creator_id, formed_at | 2 |
| Два поля по предметной области | `unit`, `price` — заполняются при публикации | 2, 6 |
| 3 услуги: черновик / удалён / опубликован | наполнение + восстановление состояния | 3, 8 |
| Не более одного черновика у пользователя | `idx_one_draft_per_creator` + проверка в `CreateFlightDraft` | 3, 5 |
| Наполнение через Adminer | `docs/lab2-seed.sql` | 3 |
| Ровно 6 HTTP-методов, три из них GET | `RegisterHandler`; лента — один обработчик на двух путях | 4–7 |
| Получение и поиск через ORM | `GetPublishedFlightServices`, `GetFlightServicesByPrice`, `GetFlightService`, `GetNextFlightServiceID`, `GetDraftFlightService` | 4 |
| Создание через ORM, кнопка «Далее» | `CreateDraftFlightService` — `db.Create` | 5 |
| Публикация через ORM, кнопка «Опубликовать» | `PublishFlightService` — `db.Model().Updates()` | 6 |
| Удаление: SQL UPDATE + курсор, без ORM | `DeleteFlightService` — `Raw().Row()` + `Scan` | 7 |
| Кнопка удаления на странице плитки | форма в `index.html` | 7 |
| Удалённые услуги просматривать нельзя | `status <> 'удален'`, ответ 404 | 4, 7 |
| Лайки только отображаются | `GetLikesCount(s)`, кнопка лайка неактивна | 4 |
| Данные трёх шаблонов — из БД | заглушек ЛР1 не осталось | 4, 8 |
| Без JavaScript | обе кнопки — HTML-формы + redirect | 5, 7 |
| Отдельная ветка | `lab2-flight-service-db` | 1, 8 |
| ER-диаграмма в StarUML | общий файл диаграмм | 8 |
| Письменные ответы на контрольные вопросы | `docs/lab2-control-questions.md` | 8 |
| Скриншоты по порядку показа | 21 штука | 8 |
| .doc с исправлениями по ЛР1 | отдельный документ | 8 |
