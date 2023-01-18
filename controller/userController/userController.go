package userController

import (
	"col-air-go/jwt"
	"col-air-go/model"
	"col-air-go/mongodb"
	"col-air-go/redis"
	"col-air-go/util"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

var (
	dbClient = mongodb.GetMongoDBClient()
	rdClient = redis.GetRedisClient()
	ch       = redis.GetRedisChannel()
)

func Register(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		util.Fail(c, "invalid request body.")
		return
	}
	if user.Saying == "" {
		user.Saying = "Hello, I'm " + user.Username
	}
	userCollection := dbClient.Database("colair").Collection("test")
	ctx, cancel := mongodb.GetContextWithTimeout(5 * time.Second)
	defer cancel()
	_, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		util.Response(c, http.StatusInternalServerError, 500, nil, err.Error())
		return
	}
	util.Success(c, nil, "registered successfully.")
}

func Login(c *gin.Context) {
	var user model.UserLogin
	if err := c.ShouldBindJSON(&user); err != nil {
		util.Fail(c, "invalid request body.")
		return
	}
	userCollection := dbClient.Database("colair").Collection("test")
	ctx, cancel := mongodb.GetContextWithTimeout(5 * time.Second)
	defer cancel()
	var result model.User
	err := userCollection.FindOne(ctx, user).Decode(&result)
	if err != nil {
		util.Response(c, http.StatusUnauthorized, 401, nil, "missing or incorrect credentials.")
		return
	}
	token, err := jwt.GenerateToken(result.Username)
	if err != nil {
		util.Response(c, http.StatusInternalServerError, 500, nil, err.Error())
		return
	}
	userJson, err := json.Marshal(result)
	if err != nil {
		util.Response(c, http.StatusInternalServerError, 500, nil, err.Error())
		return
	}
	cmd := []amqp.Table{
		{
			"type":  "set",
			"key":   result.Username + ":info",
			"value": string(userJson),
		},
		{
			"type":   "set",
			"key":    util.Int64ToString(util.MurmurHash64([]byte(token + "|#ColAir"))),
			"value":  result.Username,
			"expire": 7 * 24 * time.Hour,
		},
	}
	cmdJson, _ := json.Marshal(cmd)
	msg := amqp.Publishing{
		Headers: amqp.Table{
			"type": "pipe",
		},
		Body: cmdJson,
	}
	ch.Publish("", "redis", false, false, msg)
	util.Success(c, gin.H{"token": token}, "login successfully.")
}

func GetUserInfo(c *gin.Context) {
	username := c.GetString("userId")
	userJson, err := rdClient.Get(username + ":info").Result()
	if err != nil {
		util.Response(c, http.StatusInternalServerError, 500, nil, err.Error())
		return
	}
	var user model.User
	err = json.Unmarshal([]byte(userJson), &user)
	if err != nil {
		util.Response(c, http.StatusInternalServerError, 500, nil, err.Error())
		return
	}
	util.Success(c, gin.H{"user": user}, "get user info successfully.")
}

func Logout(c *gin.Context) {
	token := c.GetString("token")
	msg := amqp.Publishing{
		Headers: amqp.Table{
			"type": "del",
			"key":  util.Int64ToString(util.MurmurHash64([]byte(token + "|#ColAir"))),
		},
	}
	ch.Publish("", "redis", false, false, msg)
	util.Success(c, nil, "logout successfully.")
}
