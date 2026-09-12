package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"flight-service-api/internal/app/ds"
	"flight-service-api/internal/app/repository"
)

type flightServiceView struct {
	FlightService ds.FlightService
	LikesCount    int64
}

// statusForError различает «услуги нет» (404) и настоящую ошибку БД (500)
func statusForError(err error) int {
	if errors.Is(err, repository.ErrFlightServiceNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
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
			h.errorHandler(ctx, statusForError(err), err)
			return
		}
		id = next
	}

	service, err := h.Repository.GetFlightService(id)
	if err != nil {
		h.errorHandler(ctx, statusForError(err), err)
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
