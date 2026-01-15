package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Responder interface {
	Success(c *gin.Context, message string, data interface{})
	Created(c *gin.Context, message string, data interface{})
	Error(c *gin.Context, statusCode int, message string, err error)
	BadRequest(c *gin.Context, message string, err error)
	Unauthorized(c *gin.Context, message string, err error)
	Forbidden(c *gin.Context, message string, err error)
	NotFound(c *gin.Context, message string, err error)
	ServerError(c *gin.Context, err error)
}

type responseHandler struct {
	debug bool
}

func NewResponder(debug bool) *responseHandler {
	return &responseHandler{
		debug: debug,
	}
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (rh *responseHandler) Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func (rh *responseHandler) Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func (rh *responseHandler) Error(c *gin.Context, statusCode int, message string, err error) {
	errorDetails := ""
	if rh.debug && err != nil {
		errorDetails = err.Error()
	}
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   errorDetails,
	})
}

func (rh *responseHandler) BadRequest(c *gin.Context, message string, err error) {
	rh.Error(c, http.StatusBadRequest, message, err)
}

func (rh *responseHandler) Unauthorized(c *gin.Context, message string, err error) {
	rh.Error(c, http.StatusUnauthorized, message, err)
}

func (rh *responseHandler) Forbidden(c *gin.Context, message string, err error) {
	rh.Error(c, http.StatusForbidden, message, err)
}

func (rh *responseHandler) NotFound(c *gin.Context, message string, err error) {
	rh.Error(c, http.StatusNotFound, message, err)
}

func (rh *responseHandler) ServerError(c *gin.Context, err error) {
	rh.Error(c, http.StatusInternalServerError, "Internal Server Error", err)
}
