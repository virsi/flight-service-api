package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"flight-service-api/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

type flightServiceView struct {
	FlightService repository.FlightService
	LikesCount    int
}

func toFlightServiceViews(services []repository.FlightService) []flightServiceView {
	views := make([]flightServiceView, 0, len(services))
	for _, res := range services {
		views = append(views, flightServiceView{
			FlightService: res,
			LikesCount:    len(res.Likes),
		})
	}
	return views
}

func (h *Handler) GetFlightServices(ctx *gin.Context) {
	var services []repository.FlightService
	var err error

	priceQuery := ctx.Query("price")
	if priceQuery == "" {
		services, err = h.Repository.GetPublishedFlightServices()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		maxPrice, parseErr := strconv.ParseFloat(priceQuery, 64)
		if parseErr != nil {
			logrus.Error(parseErr)
		}
		services, err = h.Repository.GetFlightServicesByPrice(maxPrice)
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"flightServices": toFlightServiceViews(services),
		"price":          priceQuery,
	})
}

func (h *Handler) GetFlightService(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	if ctx.Query("next") == "true" {
		id, err = h.Repository.GetNextFlightServiceID(id)
		if err != nil {
			logrus.Error(err)
		}
	}

	service, err := h.Repository.GetFlightService(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"flightService": service,
		"likesCount":    len(service.Likes),
	})
}

func (h *Handler) GetFlightFeed(ctx *gin.Context) {
	id, err := h.Repository.GetNextFlightServiceID(0)
	if err != nil {
		logrus.Error(err)
	}

	service, err := h.Repository.GetFlightService(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"flightService": service,
		"likesCount":    len(service.Likes),
	})
}

func (h *Handler) GetFlightDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftFlightService()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"flightService": draft,
	})
}
