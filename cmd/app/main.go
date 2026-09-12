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
