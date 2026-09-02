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

type resourceView struct {
	Resource   repository.Resource
	LikesCount int
}

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

func (h *Handler) GetResources(ctx *gin.Context) {
	var resources []repository.Resource
	var err error

	priceQuery := ctx.Query("price")
	if priceQuery == "" {
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
		"price":     priceQuery,
	})
}

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

func (h *Handler) GetFeed(ctx *gin.Context) {
	id, err := h.Repository.GetNextResourceID(0)
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

func (h *Handler) GetDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetDraft()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"resource": draft,
	})
}
