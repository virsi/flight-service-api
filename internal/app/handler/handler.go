package handler

import (
	"html/template"

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
	router.GET("/flight-draft", h.GetFlightDraft)
	router.POST("/flight-draft", h.CreateFlightDraft)
	router.POST("/flight-publish", h.PublishFlightService)
	router.POST("/flight-delete", h.DeleteFlightService)

	api := router.Group("/api")
	api.GET("/flight-services", h.GetFlightServicesAPI)
	api.GET("/flight-services/feed", h.GetFlightFeedAPI)
	api.GET("/flight-services/draft", h.GetFlightDraftAPI)
	api.GET("/flight-services/:id", h.GetFlightFeedAPI)
	api.POST("/flight-services", h.CreateFlightServiceAPI)
	api.PUT("/flight-services/draft/publish", h.PublishFlightServiceAPI)
	api.DELETE("/flight-services/:id", h.DeleteFlightServiceAPI)
	api.POST("/flight-services/:id/like", h.LikeFlightServiceAPI)
	api.POST("/users/register", h.RegisterAPI)
	api.POST("/users/login", h.LoginAPI)
	api.POST("/users/logout", h.LogoutAPI)
}

// RegisterStatic регистрирует шаблоны и статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.SetFuncMap(template.FuncMap{"media": h.Repository.MediaURL})
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
