package main

import (
	"col-air-go/controller/appController"
	"col-air-go/controller/userController"
	"col-air-go/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func CollectRoutes(r *gin.Engine) *gin.Engine {
	r.Use(middleware.TimeoutMiddleware(3 * time.Second))
	r.GET("/", appController.GetServerStatus)
	r.POST("/user/register", userController.Register)
	r.POST("/user/login", userController.Login)
	r.Use(middleware.AuthMiddleware())
	{
		r.POST("/auth/user/get", userController.GetUserInfo)
		r.POST("/auth/user/logout", userController.Logout)
	}
	return r
}
