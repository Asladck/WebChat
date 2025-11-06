package handler

import (
	"github.com/gin-gonic/gin"
	"websckt/internal/service"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}
func (h *Handler) InitRouter(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/refresh", h.refreshHandler)
		auth.POST("/sign-in", h.signIn)
	}
	api := r.Group("/api", h.userIdentity)
	user := api.Group("/user/:id")
	{
		user.GET("/", h.getProfile)
	}
}
