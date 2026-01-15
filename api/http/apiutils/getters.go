package apiutils

import (
	"github.com/gin-gonic/gin"
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

func GetUserRole(c *gin.Context) Role {
	value, exists := c.Get(ContextKeyRole)
	if !exists {
		return ""
	}
	role := value.(Role)
	if _, ok := role.ValidRole(); ok {
		return role
	}
	return ""
}
