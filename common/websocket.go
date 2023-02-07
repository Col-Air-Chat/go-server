package common

import (
	"col-air-go/jwt"
	"col-air-go/model"
	"col-air-go/util"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	heartbeat    = time.Second * 30
	responseTime = time.Second * 40
)

func InitWebsocket(context *gin.Context) {
	uppgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		Subprotocols:     []string{context.GetHeader("Sec-WebSocket-Protocol")},
		HandshakeTimeout: time.Second * 5,
	}
	wsConn, err := uppgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer wsConn.Close()
	wsConn.SetPongHandler(func(string) error {
		wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
		return nil
	})
	wsConn.SetReadDeadline(time.Now().Add(responseTime))
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()
	go KeepAlive(ticker, wsConn)
	ListenMessage(wsConn)
}

func KeepAlive(ticker *time.Ticker, connection *websocket.Conn) {
	for range ticker.C {
		newMsg := &model.WebsocketMsg{
			Header: map[string]interface{}{
				"code": websocket.PingMessage,
			},
			Data: nil,
			Msg:  "ping",
		}
		if err := connection.WriteJSON(newMsg); err != nil {
			log.Println(err)
			connection.WriteMessage(websocket.CloseMessage, []byte{})
			connection.Close()
			return
		}
	}
}

func ListenMessage(connection *websocket.Conn) {
	var count int = 0
	for {
		var result map[string]interface{}
		if err := connection.ReadJSON(&result); err != nil {
			connection.WriteMessage(websocket.CloseMessage, []byte{})
			connection.Close()
			RemoveClient(connection)
			log.Println(err)
			return
		}
		log.Println(result)
		connection.SetReadDeadline(time.Now().Add(responseTime))
		code := result["header"].(map[string]interface{})["code"].(float64)
		println(code)
		switch code {
		case websocket.PongMessage:
			count++
			log.Println("pong")
			var err error
			var newToken string
			var userId string
			heartbeatToken := result["data"].(map[string]interface{})["resourceToken"].(string)
			mainToken := result["data"].(map[string]interface{})["authorizationToken"].(string)
			uid := result["data"].(map[string]interface{})["uid"].(float64)
			if count%120 == 0 {
				mainToken, userId, _ = jwt.RenewToken(mainToken)
				newToken, err = jwt.GenerateHeartbeatToken(util.Float64ToString(uid), mainToken)
			} else {
				newToken, userId, err = jwt.RenewHeartbeatToken(heartbeatToken, mainToken)
			}
			if err != nil || int64(uid) != util.StringToInt64(userId) {
				log.Println(err)
				connection.WriteMessage(websocket.CloseMessage, []byte{})
				connection.Close()
				return
			}
			newMsg := &model.WebsocketMsg{
				Header: map[string]interface{}{
					"code": websocket.PongMessage,
				},
				Data: map[string]interface{}{
					"resourceToken":      newToken,
					"authorizationToken": mainToken,
				},
				Msg: "pong",
			}
			if err := connection.WriteJSON(newMsg); err != nil {
				log.Println(err)
				connection.WriteMessage(websocket.CloseMessage, []byte{})
				connection.Close()
				return
			}
		case websocket.CloseMessage:
			log.Println("close")
			RemoveClient(connection)
		case 1145114:
			mainToken := result["data"].(map[string]interface{})["authorizationToken"].(string)
			uid := result["data"].(map[string]interface{})["uid"].(float64)
			resourceToken, _ := jwt.GenerateHeartbeatToken(util.Float64ToString(uid), mainToken)
			newMsg := &model.WebsocketMsg{
				Header: map[string]interface{}{
					"code": 200,
				},
				Data: resourceToken,
				Msg:  "connect success",
			}
			if err := connection.WriteJSON(newMsg); err != nil {
				log.Println(err)
				connection.WriteMessage(websocket.CloseMessage, []byte{})
				connection.Close()
				return
			}
			AddClient(connection, int64(uid), mainToken, resourceToken)
		}
	}
}
