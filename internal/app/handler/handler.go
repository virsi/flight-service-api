package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"flight-service-api/internal/app/repository"
)

// Handler обрабатывает входящие запросы, валидирует параметры
// и вызывает методы уровня репозитория.
type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// resourceView — ресурс + количество лайков, вычисленное в обработчике
// по коллекции (для страницы «Плитка»).
type resourceView struct {
	Resource   repository.Resource
	LikesCount int
}

// toViews превращает список ресурсов в список представлений с количеством лайков.
func toViews(resources []repository.Resource) []resourceView {
	views := make([]resourceView, 0, len(resources))
	for _, res := range resources {
		views = append(views, resourceView{
			Resource:   res,
			LikesCount: len(res.Likes),
		})
	}
	return views
}

// GetResources — страница «Плитка»: список опубликованных ресурсов
// с фильтрацией на сервере по цене.
func (h *Handler) GetResources(ctx *gin.Context) {
	var resources []repository.Resource
	var err error

	priceQuery := ctx.Query("price") // значение из поля фильтра по цене
	if priceQuery == "" {            // если поле пусто, отдаём все опубликованные
		resources, err = h.Repository.GetPublishedResources()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		maxPrice, parseErr := strconv.ParseFloat(priceQuery, 64)
		if parseErr != nil {
			logrus.Error(parseErr)
		}
		resources, err = h.Repository.GetResourcesByPrice(maxPrice)
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"resources": toViews(resources),
		"price":     priceQuery, // возвращаем введённый фильтр обратно на страницу
	})
}

// GetResource — страница «Лента»: одна карточка по ID.
// Если передан ?next=true, открывается следующий ресурс после этого ID.
func (h *Handler) GetResource(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	if ctx.Query("next") == "true" {
		id, err = h.Repository.GetNextResourceID(id)
		if err != nil {
			logrus.Error(err)
		}
	}

	resource, err := h.Repository.GetResource(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"resource":   resource,
		"likesCount": len(resource.Likes),
	})
}

// GetFeed — страница «Лента» без указания ID (переход из панели вкладок).
// Открывает первый опубликованный ресурс.
func (h *Handler) GetFeed(ctx *gin.Context) {
	id, err := h.Repository.GetNextResourceID(0) // 0 — берём первый опубликованный
	if err != nil {
		logrus.Error(err)
	}

	resource, err := h.Repository.GetResource(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"resource":   resource,
		"likesCount": len(resource.Likes),
	})
}

// GetDraft — страница «Добавление»: показывает ресурс в статусе «черновик».
func (h *Handler) GetDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetDraft()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"resource": draft,
	})
}
