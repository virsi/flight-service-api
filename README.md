# flight-service-api

Бэкенд заявочной системы по курсу «Разработка Интернет Приложений» (РИП).
Тема 26 — **«Обслуживание рейса в аэропорте»** (лабораторная работа №1).

## Предметная область

- **Ресурс обслуживания** (услуга) — виды ресурсов, персонала и техники (топливо, аэродромные тягачи, представители авиакомпании, бортовое питание и т. д.) с единицей измерения и ценой.
- **Заявка на обслуживание рейса** — заявка с указанием количества ресурсов и расчётом стоимости.

## Стек

- Go + [Gin](https://github.com/gin-gonic/gin)
- Серверные шаблоны `html/template`
- [logrus](https://github.com/sirupsen/logrus)
- MinIO (объектное хранилище изображений/видео)
- Палитра Aeroflot: `#041839` / `#10349E` / `#EF8A06`

## Запуск

```bash
go run ./cmd/app
```

Страницы:

- `http://localhost:8080/resources` — плитка (список ресурсов)
- `http://localhost:8080/feed` — лента
- `http://localhost:8080/draft` — добавление (черновик)

## Маршруты

| Метод | URL | Описание |
|---|---|---|
| GET | `/resources` | список опубликованных ресурсов, фильтр по цене `?price=` |
| GET | `/resource/:id` | лента по ID, `?next=true` — следующий ресурс |
| GET | `/feed` | лента без ID (первый опубликованный) |
| GET | `/draft` | ресурс в статусе «черновик» |

## Структура

```
cmd/app/main.go
internal/api/server.go
internal/app/repository/repository.go
internal/app/handler/handler.go
templates/*.html
resources/styles/style.css
docker-compose.yml
```

## MinIO

`docker compose up -d`, затем создать публичный бакет `flight-media` и загрузить в него изображения/видео ресурсов (ключи на латинице). Пароль задаётся в `.env` (см. `.env.example`).
