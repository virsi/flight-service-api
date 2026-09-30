package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"flight-service-api/internal/app/ds"
	"flight-service-api/internal/app/repository"
)

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

// resolveFeedService — услуга ленты по ?id= (+ ?next=true) и число лайков; при ошибке возвращает HTTP-статус
func (h *Handler) resolveFeedService(ctx *gin.Context) (ds.FlightService, int64, int, error) {
	var id uint

	if idStr := ctx.Query("id"); idStr != "" {
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
