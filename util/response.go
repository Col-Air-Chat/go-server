package util

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Response(c *gin.Context, httpCode int, code int, data gin.H, msg string) {
	c.JSON(httpCode, gin.H{
		"code":      code,
		"data":      data,
		"msg":       msg,
		"timestamp": time.Now().Unix(),
	})
}

func Success(c *gin.Context, data gin.H, msg string) {
	Response(c, http.StatusOK, 200, data, msg)
}

func Fail(c *gin.Context, msg string) {
	Response(c, http.StatusOK, 400, nil, msg)
}
