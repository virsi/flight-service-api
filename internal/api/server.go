package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"flight-service-api/internal/app/handler"
	"flight-service-api/internal/app/repository"
)

// StartServer поднимает веб-сервис: создаёт репозиторий, обработчики,
// регистрирует роуты и раздаёт статику и шаблоны.
func StartServer() {
	log.Println("Server start up")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handlers := handler.NewHandler(repo)

	r := gin.Default()
	// подключаем html-шаблоны из папки templates
	r.LoadHTMLGlob("templates/*")
	// раздаём статику (css, изображения) из папки resources
	r.Static("/static", "./resources")

	// три GET-запроса: список карточек (плитка), лента, черновик.
	// Лента открывается по ID (/resource/:id), а из панели вкладок — без ID (/feed).
	r.GET("/resources", handlers.GetResources)
	r.GET("/resource/:id", handlers.GetResource)
	r.GET("/feed", handlers.GetFeed)
	r.GET("/draft", handlers.GetDraft)

	r.Run() // listen and serve on 0.0.0.0:8080

	log.Println("Server down")
}
