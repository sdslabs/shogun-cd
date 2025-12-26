package middleware

import "github.com/gin-gonic/gin"

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// [TODO] check if user is logged in.
	}
}

func VerifyAdmin(c *gin.Context) {
	//[TODO] verify if user is admin
}
