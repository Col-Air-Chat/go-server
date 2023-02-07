package middleware

import (
	"net/http"
	"strings"

	"col-air-go/common"
	"col-air-go/jwt"
	"col-air-go/util"
	"github.com/gin-gonic/gin"
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
		claims, err := jwt.ParseToken(token, "col-air-go-token-main")
		if err != nil {
			util.Response(c, http.StatusUnauthorized, 401, nil, "missing or incorrect credentials.")
			c.Abort()
			return
		}
		msg := util.Int64ToString(util.MurmurHash64([]byte(token + "|#ColAir")))
		var resultChan = make(chan interface{})
		var resultErrChan = make(chan error)
		go common.GetDataFromRedis(resultChan, resultErrChan, msg)
		resultErr := <-resultErrChan
		if resultErr != nil {
			util.Response(c, http.StatusUnauthorized, 401, nil, "missing or incorrect credentials.")
			c.Abort()
			return
		}
		result := <-resultChan
		resultUid := result
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
