package model

type Msg struct {
	MsgId         string `json:"mid" bson:"mid" binding:"required"`
	MsgSender     int64  `json:"msender" bson:"msender" binding:"required"`
	MsgValue      string `json:"mvalue" bson:"mvalue" binding:"required"`
	MsgType       string `json:"mtype" bson:"mtype" binding:"required"`
	MsgTime       int64  `json:"mtime" bson:"mtime" binding:"required"`
	MsgUrl        string `json:"murl" bson:"murl"`
	MsgStatus     int64  `json:"mstatus" bson:"mstatus" binding:"required"`
	MsgReceiver   int64  `json:"mreceiver" bson:"mreceiver" binding:"required"`
	MsgReply      string `json:"mreply" bson:"mreply"`
	MsgEdited     bool   `json:"medited" bson:"medited" default:"false"`
	MsgEditedTime int64  `json:"meditedtime" bson:"meditedtime"`
}

type WebsocketMsg struct {
	Header map[string]interface{} `json:"header"`
	Data   interface{}            `json:"data"`
	Msg    string                 `json:"msg"`
}
