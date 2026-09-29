package handler

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"flight-service-api/internal/app/ds"
	"flight-service-api/internal/app/repository"
	"flight-service-api/internal/app/serializer"
)

// priceLimit — верхняя граница колонки price numeric(10,2)
const priceLimit = 99999999.99

// GetFlightServicesAPI — список опубликованных услуг с фильтром ?price=
func (h *Handler) GetFlightServicesAPI(ctx *gin.Context) {
	var services []ds.FlightService
	var err error

	if priceQuery := ctx.Query("price"); priceQuery == "" {
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

	userID := CurrentUserID()
	data := make([]serializer.FlightService, 0, len(services))
	for _, service := range services {
		data = append(data, serializer.NewFlightService(service, counts[service.ID], userID, h.Repository.MediaURL))
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": data})
}

// GetFlightFeedAPI — лента: /feed и /:id (+ ?next=true)
func (h *Handler) GetFlightFeedAPI(ctx *gin.Context) {
	service, likesCount, status, err := h.resolveFeedService(ctx)
	if err != nil {
		h.errorHandler(ctx, status, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   serializer.NewFlightService(service, likesCount, CurrentUserID(), h.Repository.MediaURL),
	})
}

// GetFlightDraftAPI — черновик текущего пользователя
func (h *Handler) GetFlightDraftAPI(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftFlightService(CurrentUserID())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if draft == nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("черновик не найден"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   serializer.NewFlightService(*draft, 0, CurrentUserID(), h.Repository.MediaURL),
	})
}

// CreateFlightServiceAPI — создание черновика услуги: multipart с файлами image и video
func (h *Handler) CreateFlightServiceAPI(ctx *gin.Context) {
	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	service := ds.FlightService{
		Name:        ctx.PostForm("name"),
		Description: ctx.PostForm("description"),
		Unit:        ctx.PostForm("unit"),
	}
	if service.Name == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("поле name обязательно"))
		return
	}
	if priceForm := ctx.PostForm("price"); priceForm != "" {
		price, err := strconv.ParseFloat(priceForm, 64)
		if err != nil || math.IsNaN(price) || math.IsInf(price, 0) || price < 0 || price > priceLimit {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("price должна быть числом от 0 до 99999999.99"))
			return
		}
		service.Price = price
	}

	image, err := ctx.FormFile("image")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	video, err := ctx.FormFile("video")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// черновик проверяем до загрузки файлов, чтобы не мусорить в MinIO
	draft, err := h.Repository.GetDraftFlightService(CurrentUserID())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if draft != nil {
		h.errorHandler(ctx, http.StatusConflict, repository.ErrDraftExists)
		return
	}

	if image != nil {
		if service.Image, err = h.Repository.UploadMedia(image, "image"); err != nil {
			h.errorHandler(ctx, statusForError(err), err)
			return
		}
	}
	if video != nil {
		if service.Video, err = h.Repository.UploadMedia(video, "video"); err != nil {
			h.Repository.RemoveMedia(service.Image)
			h.errorHandler(ctx, statusForError(err), err)
			return
		}
	}

	if err = h.Repository.CreateDraftFlightService(&service, CurrentUserID()); err != nil {
		h.Repository.RemoveMedia(service.Image)
		h.Repository.RemoveMedia(service.Video)
		h.errorHandler(ctx, statusForError(err), err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"data":    serializer.NewFlightService(service, 0, CurrentUserID(), h.Repository.MediaURL),
		"message": "черновик услуги создан",
	})
}

// PublishFlightServiceAPI — публикация черновика текущего пользователя
func (h *Handler) PublishFlightServiceAPI(ctx *gin.Context) {
	service, err := h.Repository.PublishDraftFlightService(CurrentUserID())
	if err != nil {
		h.errorHandler(ctx, statusForError(err), err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   serializer.NewFlightService(service, 0, CurrentUserID(), h.Repository.MediaURL),
	})
}

// DeleteFlightServiceAPI — логическое удаление своей услуги
func (h *Handler) DeleteFlightServiceAPI(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err = h.Repository.DeleteFlightService(uint(id), CurrentUserID()); err != nil {
		h.errorHandler(ctx, statusForError(err), err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "услуга удалена"})
}

type likeRequest struct {
	Like *int `json:"like" binding:"required,oneof=0 1"`
}

// LikeFlightServiceAPI — поставить (1) или снять (0) лайк текущего пользователя
func (h *Handler) LikeFlightServiceAPI(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req likeRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	likesCount, err := h.Repository.SetLike(uint(id), CurrentUserID(), *req.Like == 1)
	if err != nil {
		h.errorHandler(ctx, statusForError(err), err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   gin.H{"like": *req.Like, "likes_count": likesCount},
	})
}
