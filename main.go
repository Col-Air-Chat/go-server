package main

import (
	"col-air-go/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	client = util.GetMongoDBClient()
)

func main() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	router.Run(":8080")
}


