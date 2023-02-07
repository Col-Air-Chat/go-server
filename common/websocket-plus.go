package common

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ConnectionClient struct {
	Conn               *websocket.Conn
	Uid                int64
	AuthorizationToken string
	ResourceToken      string
}

type ConnectionGroup struct {
	Conn map[string][]*websocket.Conn
}

type PrivateGroup struct {
	Conn_1 *websocket.Conn
	Conn_2 *websocket.Conn
}

var (
	client     = make([]ConnectionClient, 0)
	groups     = make(map[string]*ConnectionGroup)
	private    = make(map[string]*PrivateGroup)
	clientLock = sync.RWMutex{}
)

func AddClient(conn *websocket.Conn, uid int64, authorizationToken string, resourceToken string) {
	clientLock.Lock()
	client = append(client, ConnectionClient{
		Conn:               conn,
		Uid:                uid,
		AuthorizationToken: authorizationToken,
		ResourceToken:      resourceToken,
	})
	clientLock.Unlock()
}

func RemoveClient(conn *websocket.Conn) {
	clientLock.Lock()
	for k, v := range client {
		if v.Conn == conn {
			client = append(client[:k], client[k+1:]...)
			break
		}
	}
	clientLock.Unlock()
}

func SendToAllClient(message interface{}) {
	for _, v := range client {
		if err := v.Conn.WriteJSON(message); err != nil {
			RemoveClient(v.Conn)
		}
	}
}

func SendToClient(uid int64, message interface{}) {
	for _, v := range client {
		if v.Uid == uid {
			if err := v.Conn.WriteJSON(message); err != nil {
				RemoveClient(v.Conn)
			}
		}
	}
}

func AddGroup(group string, conn *websocket.Conn) {
	if _, ok := groups[group]; !ok {
		groups[group] = &ConnectionGroup{
			Conn: make(map[string][]*websocket.Conn),
		}
	}
	groups[group].Conn[conn.RemoteAddr().String()] = append(groups[group].Conn[conn.RemoteAddr().String()], conn)
}

func RemoveGroup(group string, conn *websocket.Conn) {
	if _, ok := groups[group]; ok {
		delete(groups[group].Conn, conn.RemoteAddr().String())
	}
}

func SendToGroup(group string, message interface{}) {
	if _, ok := groups[group]; ok {
		for _, v := range groups[group].Conn {
			for _, c := range v {
				if err := c.WriteJSON(message); err != nil {
					RemoveGroup(group, c)
				}
			}
		}
	}
}

func AddPrivateGroup(group string, conn_1 *websocket.Conn, conn_2 *websocket.Conn) {
	private[group] = &PrivateGroup{
		Conn_1: conn_1,
		Conn_2: conn_2,
	}
}

func RemovePrivateGroup(group string) {
	delete(private, group)
}

func SendToPrivateGroup(group string, message interface{}) {
	if _, ok := private[group]; ok {
		if err := private[group].Conn_1.WriteJSON(message); err != nil {
			RemovePrivateGroup(group)
		}
		if err := private[group].Conn_2.WriteJSON(message); err != nil {
			RemovePrivateGroup(group)
		}
	}
}
