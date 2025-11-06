package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) getProfile(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, "invalid user id param")
		return
	}

	user, err := h.services.User.GetUserByID(userID)
	if err != nil {
		// сервис вернул ошибку при получении пользователя
		NewErrorResponse(c, http.StatusInternalServerError, "failed to get user")
		return
	}
	// успешный ответ с данными пользователя
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
