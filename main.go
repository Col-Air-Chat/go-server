package main

import (
	"net/http"
	"os"

	"col-air-go/common"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

func HttpServer() http.Handler {
	router := gin.Default()
	router = CollectRoutes(router)
	return router
}

func WebSocketServer() http.Handler {
	router := gin.Default()
	router.GET(viper.GetString("websocket.uri"), common.InitWebsocket)
	return router
}

func main() {
	initConfig()
	HttpServerConfig := &http.Server{
		Addr:    viper.GetString("server.hsot") + ":" + viper.GetString("server.port"),
		Handler: HttpServer(),
	}
	WebSocketServerConfig := &http.Server{
		Addr:    viper.GetString("websocket.host") + ":" + viper.GetString("websocket.port"),
		Handler: WebSocketServer(),
	}
	g.Go(func() error {
		return HttpServerConfig.ListenAndServe()
	})
	g.Go(func() error {
		return WebSocketServerConfig.ListenAndServe()
	})
	if err := g.Wait(); err != nil {
		panic(err)
	}
}

func initConfig() {
	workDir, _ := os.Getwd()
	viper.SetConfigName("application")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	common.SetRedisConfig(
		viper.GetString("redis.host"),
		viper.GetString("redis.password"),
		viper.GetInt("redis.db"),
		viper.GetString("redis.amqp.url"),
	)
	common.InitRedis()
	common.SetMongoDBConfig(
		viper.GetString("mongodb.url"),
	)
	common.InitMongodb()
	common.InitCache()
}
