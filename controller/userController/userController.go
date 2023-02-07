package userController

import (
	"col-air-go/common"
	"col-air-go/jwt"
	"col-air-go/model"
	"col-air-go/util"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Register(c *gin.Context) {
	var (
		dbClient = common.GetMongoDBClient()
	)
	var user model.User
	var userRegister model.UserRegister
	if err := c.ShouldBindJSON(&userRegister); err != nil {
		util.Fail(c, "invalid request body.")
		return
	}
	user.Username = "NewUser"
	user.Email = userRegister.Email
	if util.IsEmail(userRegister.Email) {
		userCollection := dbClient.Database("colair").Collection("users")
		ctx, cancel := common.GetContextWithTimeout(5 * time.Second)
		defer cancel()
		var result model.User
		err := userCollection.FindOne(ctx, bson.M{"email": userRegister.Email}).Decode(&result)
		if err == nil {
			util.Response(c, http.StatusOK, 400, nil, "email address has been registered.")
			return
		}
	} else {
		util.Response(c, http.StatusOK, 400, nil, "invalid email address.")
		return
	}
	userCollection := dbClient.Database("colair").Collection("users")
	ctx, cancel := common.GetContextWithTimeout(5 * time.Second)
	defer cancel()
	var result model.User
	err := userCollection.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.M{"uid": -1})).Decode(&result)
	println(result.Uid)
	if err == nil {
		user.Uid = result.Uid + 1
	} else {
		user.Uid = 10001
	}
	user.Password = util.Sha256(userRegister.Password + "colair" + util.Int64ToString(user.Uid))
	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		util.Response(c, http.StatusInternalServerError, 500, nil, err.Error())
		return
	}
	util.Success(c, nil, "registered successfully.")
}

func Login(c *gin.Context) {
	var (
		dbClient = common.GetMongoDBClient()
		ch       = common.GetRedisChannel()
	)
	var user model.UserLogin
	if err := c.ShouldBindJSON(&user); err != nil {
		util.Fail(c, "invalid request body.")
		return
	}
	user.Password = util.Sha256(user.Password + "colair" + util.Int64ToString(user.Uid))
	userCollection := dbClient.Database("colair").Collection("users")
	ctx, cancel := common.GetContextWithTimeout(5 * time.Second)
	defer cancel()
	var result model.User
	err := userCollection.FindOne(ctx, user).Decode(&result)
	if err != nil {
		util.Response(c, http.StatusOK, 402, nil, "account or password is incorrect.")
		return
	}
	if result.Disabled {
		util.Response(c, http.StatusOK, 403, nil, "account has been disabled.")
		return
	}
	userId := util.Int64ToString(result.Uid)
	token, err := jwt.GenerateToken(userId)
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
			"key":   userId + ":info",
			"value": string(userJson),
		},
		{
			"type":   "set",
			"key":    util.Int64ToString(util.MurmurHash64([]byte(token + "|#ColAir"))),
			"value":  userId,
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
	common.GetCache().Set(util.Int64ToString(util.MurmurHash64([]byte(token+"|#ColAir"))), userId, common.CacheExpire(7*24*time.Hour))
	common.GetCache().Set(userId+":info", string(userJson), common.CacheExpire(1*24*time.Hour))
	util.Success(c, gin.H{"authorizationToken": token, "userInfo": result}, "login successfully.")
}

func GetUserInfo(c *gin.Context) {
	userId := c.GetString("userId")
	key := userId + ":info"
	var result interface{}
	var user model.User
	var resultChan = make(chan interface{})
	var resultErrChan = make(chan error)
	go common.GetDataFromRedis(resultChan, resultErrChan, key)
	resultErr := <-resultErrChan
	var userJson string
	if resultErr != nil {
		var resultChan = make(chan interface{})
		var resultBoolChan = make(chan bool)
		go common.GetDataFromCache(resultChan, resultBoolChan, key)
		found := <-resultBoolChan
		if found {
			result = <-resultChan
			userJson = result.(string)
			ch := common.GetRedisChannel()
			msg := amqp.Publishing{
				Headers: amqp.Table{
					"type":  "set",
					"key":   key,
					"value": userJson,
				},
			}
			ch.Publish("", "redis", false, false, msg)
		} else {
			dbClient := common.GetMongoDBClient()
			ctx, cancel := common.GetContextWithTimeout(5 * time.Second)
			defer cancel()
			userCollection := dbClient.Database("colair").Collection("users")
			userCollection.FindOne(ctx, bson.M{"uid": util.StringToInt64(userId)}).Decode(&user)
			marshal, _ := json.Marshal(user)
			userJson = string(marshal)
			common.GetCache().Set(key, userJson, common.CacheExpire(1*24*time.Hour))
			ch := common.GetRedisChannel()
			msg := amqp.Publishing{
				Headers: amqp.Table{
					"type":  "set",
					"key":   key,
					"value": userJson,
				},
			}
			ch.Publish("", "redis", false, false, msg)
		}
	} else {
		result = <-resultChan
		userJson = result.(string)
	}
	err := json.Unmarshal([]byte(userJson), &user)
	if err != nil {
		util.Response(c, http.StatusInternalServerError, 500, nil, err.Error())
		return
	}
	util.Success(c, gin.H{"user": user}, "get user info successfully.")
}

func Logout(c *gin.Context) {
	var (
		ch = common.GetRedisChannel()
	)
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
