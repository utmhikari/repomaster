package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const SuccessCode int = 0
const DefaultErrorCode int = -1

type Response struct {
	Success bool
	Message string
	Data    interface{}
	Code    int
}

// Success responds json with success = True and StatusOK
func Success(c *gin.Context, resp Response) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": resp.Message,
		"data":    resp.Data,
		"code":    SuccessCode,
	})
}

// SuccessResponse fast way to make success response
func SuccessResponse(c *gin.Context, data interface{}) {
	Success(c, Response{Data: data})
}

// Error responds json with success = False yet StatusOK
func Error(c *gin.Context, resp Response) {
	if resp.Code == SuccessCode {
		resp.Code = DefaultErrorCode
	}
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": resp.Message,
		"data":    resp.Data,
		"code":    resp.Code,
	})
}

// ErrorResponse fast way to make error response
func ErrorResponse(c *gin.Context, err error) {
	Error(c, Response{Message: err.Error()})
}

// RequestError responds json with success = False and http status code of error
func RequestError(c *gin.Context, statusCode int, resp Response) {
	if resp.Code == SuccessCode {
		resp.Code = DefaultErrorCode
	}
	if statusCode < http.StatusBadRequest {
		statusCode = http.StatusBadRequest
	}
	c.JSON(statusCode, gin.H{
		"success": false,
		"message": resp.Message,
		"data":    resp.Data,
		"code":    resp.Code,
	})
}

// HealthCheck api for health check
func HealthCheck(c *gin.Context) {
	Success(c, Response{Message: "ok"})
}
