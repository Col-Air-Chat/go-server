package middleware

import (
	"net/http"
	"strings"

	"col-air-go/jwt"
	"col-air-go/redis"
	"col-air-go/util"
	"github.com/gin-gonic/gin"
)

var (
	rdClient = redis.GetRedisClient()
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			util.Response(c, http.StatusUnauthorized, 401, nil, "missing or incorrect credentials.")
			c.Abort()
			return
		}
		token = token[7:]
		claims, _ := jwt.ParseToken(token, "col-air-go-token-main")
		msg := util.Int64ToString(util.MurmurHash64([]byte(token + "|#ColAir")))
		resultUid, _ := rdClient.Get(msg).Result()
		if resultUid != claims.UserId {
			util.Response(c, http.StatusUnauthorized, 401, nil, "missing or incorrect credentials.")
			c.Abort()
			return
		}
		userId := claims.UserId
		c.Set("userId", userId)
		c.Set("token", token)
		c.Next()
	}
}
