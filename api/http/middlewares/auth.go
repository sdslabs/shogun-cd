package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/apiutils"
)

func (m *Manager) AuthRequired(c *gin.Context) {

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		m.response.Unauthorized(c, "Authorization header required", nil)
		c.Abort()
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		m.response.Unauthorized(c, "Invalid authorization format", nil)
		c.Abort()
		return
	}

	tokenString := parts[1]
	claims, err := apiutils.ParseToken(tokenString, m.config.ApiConfig.JWT.Secret)
	if err != nil {
		m.response.Unauthorized(c, "Invalid or expired session", err)
		c.Abort()
		return
	}

	c.Set(apiutils.ContextKeyEmail, claims.Email)
	c.Set(apiutils.ContextKeyRole, claims.Role)

	c.Next()

}

func (m *Manager) VerifyAdmin(c *gin.Context) {
	role := apiutils.GetUserRole(c)

	if role != apiutils.RoleAdmin {
		m.response.Forbidden(c, "Access denied: Admin privileges required", nil)
		c.Abort()
		return
	}

	c.Next()
}
