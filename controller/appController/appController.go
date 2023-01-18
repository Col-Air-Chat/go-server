package appController

import (
	"col-air-go/util"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	serverStartTime = time.Now()
)

func GetServerStatus(c *gin.Context) {
	util.Success(c, gin.H{
		"server_start_time": serverStartTime,
		"server_version":    util.GetVersion(),
		"running_time":      time.Since(serverStartTime).String(),
	}, "ok")
}
