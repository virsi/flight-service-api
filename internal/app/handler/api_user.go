package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"flight-service-api/internal/app/serializer"
)

// registerRequest — с клиента принимаются только логин и пароль,
// роль (is_moderator) и id вычисляются на бэкенде
type registerRequest struct {
	Login    string `json:"login" binding:"required,max=25"`
	Password string `json:"password" binding:"required,max=100"`
}

// RegisterAPI — регистрация пользователя
func (h *Handler) RegisterAPI(ctx *gin.Context) {
	var req registerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.CreateUser(req.Login, req.Password)
	if err != nil {
		h.errorHandler(ctx, statusForError(err), err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": serializer.NewUser(user)})
}

// LoginAPI — заглушка аутентификации (будет в ЛР4)
func (h *Handler) LoginAPI(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "заглушка: аутентификация будет в ЛР4"})
}

// LogoutAPI — заглушка деавторизации (будет в ЛР4)
func (h *Handler) LogoutAPI(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "заглушка: деавторизация будет в ЛР4"})
}
