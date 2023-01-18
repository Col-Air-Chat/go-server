package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router = CollectRoutes(router)
	router.Run(":8080")
}