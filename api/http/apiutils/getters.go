package apiutils

import (
	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/internal/models"
)

const (
	ContextKeyEmail = "user_email"
	ContextKeyRole  = "user_role"
)

func GetUserEmail(c *gin.Context) string {
	value, exists := c.Get(ContextKeyEmail)
	if !exists {
		return ""
	}
	return value.(string)
}

func GetUserRole(c *gin.Context) string {
	value, exists := c.Get(ContextKeyRole)
	if !exists {
		return ""
	}
	role := value.(string)
	if role == models.RoleAdmin || role == models.RoleUser {
		return role
	}
	return ""
}
