package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"flight-service-api/internal/app/handler"
	"flight-service-api/internal/app/repository"
)

func StartServer() {
	log.Println("Server start up")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handlers := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/flight-resources", handlers.GetFlightServices)
	r.GET("/flight-resource/:id", handlers.GetFlightService)
	r.GET("/flight-feed", handlers.GetFlightFeed)
	r.GET("/flight-draft", handlers.GetFlightDraft)

	r.Run()

	log.Println("Server down")
}
