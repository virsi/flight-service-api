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
	router.POST("/flight-draft", h.CreateFlightDraft)
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
