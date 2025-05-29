package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Error represents an API error response
// @Description API error information
type Error struct {
	Message string `json:"message"`
}

// StatusResponse represents a simple status response
// @Description Basic status response
type statusResponse struct {
	Status string `json:"status"`
}

// StatusFloat represents a numeric status response
// @Description Numeric status response

// NewErrorResponse logs error and returns error response
func NewErrorResponse(c *gin.Context, statusCode int, message string) {
	logrus.Error()
	c.AbortWithStatusJSON(statusCode, Error{message})
}
