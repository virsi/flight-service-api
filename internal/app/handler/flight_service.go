package handler

import (
	"errors"
	"fmt"
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

	if errors.Is(err, repository.ErrUserExists) {
		return http.StatusConflict
	}
	if errors.Is(err, repository.ErrDraftExists) {
		return http.StatusConflict
	}
	if errors.Is(err, repository.ErrInvalidMedia) {
		return http.StatusBadRequest
	}
	if errors.Is(err, repository.ErrForbidden) {
		return http.StatusForbidden
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

// resolveFeedService — услуга ленты по id (path :id или ?id=, + ?next=true) и число лайков; при ошибке возвращает HTTP-статус
func (h *Handler) resolveFeedService(ctx *gin.Context) (ds.FlightService, int64, int, error) {
	var id uint

	idStr := ctx.Param("id")
	if idStr == "" {
		idStr = ctx.Query("id")
	}
	if idStr != "" {
		parsed, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return ds.FlightService{}, 0, http.StatusBadRequest, err
		}
		id = uint(parsed)
	}

	if id == 0 || ctx.Query("next") == "true" {
		next, err := h.Repository.GetNextFlightServiceID(id)
		if err != nil {
			return ds.FlightService{}, 0, statusForError(err), err
		}
		id = next
	}

	service, err := h.Repository.GetFlightService(id)
	if err != nil {
		return ds.FlightService{}, 0, statusForError(err), err
	}

	likesCount, err := h.Repository.GetLikesCount(service.ID)
	if err != nil {
		return ds.FlightService{}, 0, http.StatusInternalServerError, err
	}

	return service, likesCount, http.StatusOK, nil
}

// GetFlightFeed — лента: /flight-feed (+ ?id=N, ?next=true)
func (h *Handler) GetFlightFeed(ctx *gin.Context) {
	service, likesCount, status, err := h.resolveFeedService(ctx)
	if errors.Is(err, repository.ErrFlightServiceNotFound) && ctx.Query("id") != "" {
		// удалённой или несуществующей услуги нет — просто возвращаем в ленту
		ctx.Redirect(http.StatusFound, "/flight-feed")
		return
	}
	if err != nil {
		h.errorHandler(ctx, status, err)
		return
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"flightService": service,
		"likesCount":    likesCount,
	})
}

// GetFlightDraft — страница добавления
func (h *Handler) GetFlightDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftFlightService(CurrentUserID())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"flightService": draft,
	})
}

// CreateFlightDraft — кнопка «Далее»: создание черновика
func (h *Handler) CreateFlightDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftFlightService(CurrentUserID())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if draft != nil {
		h.errorHandler(ctx, http.StatusConflict, fmt.Errorf("у пользователя уже есть черновик"))
		return
	}

	err = h.Repository.CreateDraftFlightService(&ds.FlightService{Name: ctx.PostForm("name")}, CurrentUserID())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/flight-draft")
}

// PublishFlightService — кнопка «Опубликовать»: смена статуса черновика
func (h *Handler) PublishFlightService(ctx *gin.Context) {
	price, err := strconv.ParseFloat(ctx.PostForm("price"), 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.PublishFlightService(
		CurrentUserID(),
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

// DeleteFlightService — логическое удаление услуги с плитки
func (h *Handler) DeleteFlightService(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.PostForm("flight_service_id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err = h.Repository.DeleteFlightService(uint(id), CurrentUserID()); err != nil {
		h.errorHandler(ctx, statusForError(err), err)
		return
	}

	ctx.Redirect(http.StatusFound, "/flight-resources")
}
