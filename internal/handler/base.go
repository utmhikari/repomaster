package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const SuccessCode int = 0
const DefaultErrorCode int = -1

// Success responds json with success = True and StatusOK
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"code":    SuccessCode,
	})
}

// Error responds json with success = False yet StatusOK
func Error(c *gin.Context, err error, code int) {
	if code == SuccessCode {
		code = DefaultErrorCode
	}
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"err":     err.Error(),
		"code":    code,
	})
}

// RequestError responds json with success = False and http status code of error
func RequestError(c *gin.Context, statusCode int, err error, code int) {
	if code == SuccessCode {
		code = DefaultErrorCode
	}
	if statusCode < http.StatusBadRequest {
		statusCode = http.StatusBadRequest
	}
	c.JSON(statusCode, gin.H{
		"success": false,
		"err":     err.Error(),
		"code":    code,
	})
}

type base struct{}

// Base is the instance of base handler
var Base base

// HealthCheck
func (*base) HealthCheck(c *gin.Context) {
	Success(c, "ok")
}
