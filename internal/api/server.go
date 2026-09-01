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

	r.GET("/resources", handlers.GetResources)
	r.GET("/resource/:id", handlers.GetResource)
	r.GET("/feed", handlers.GetFeed)
	r.GET("/draft", handlers.GetDraft)

	r.Run()

	log.Println("Server down")
}
