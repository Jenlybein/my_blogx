package chat_api

import (
	"fmt"
	"myblogx/global"
	"myblogx/utils/jwts"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	chatWSReadBufferSize  = 1024
	chatWSWriteBufferSize = 1024
	chatWSReadLimit       = 4 * 1024
	chatWSPongWait        = 60 * time.Second
	chatWSWriteWait       = 10 * time.Second
	chatWSPingPeriod      = chatWSPongWait * 9 / 10
)

var chatWSUpgrader = websocket.Upgrader{
	ReadBufferSize:  chatWSReadBufferSize,
	WriteBufferSize: chatWSWriteBufferSize,
	// 当前聊天 ws 主要给浏览器端使用，开发阶段先放开同源校验，
	// 避免本地前后端不同端口时升级握手被浏览器跨域策略拦截。
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ChatWsView 处理聊天 WebSocket 长连接。
// 当前先保留成一个稳定的 echo 通道，用来验证握手、收发和连接保活链路。
func (ChatApi) ChatWsView(c *gin.Context) {
	claims := jwts.MustGetClaimsByGin(c)

	// 升级聊天 ws 连接
	conn, err := chatWSUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.Logger.Errorf("升级聊天 ws 连接失败 user_id=%d err=%v", claims.UserID, err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			global.Logger.Warnf("关闭聊天 ws 连接失败 user_id=%d err=%v", claims.UserID, err)
		}
	}()

	var writeMu sync.Mutex
	configureChatWSConn(conn)
	global.Logger.Infof("聊天 ws 已连接 user_id=%d", claims.UserID)

	done := make(chan struct{})
	defer close(done)

	go startChatWSPingLoop(conn, claims.UserID, done, &writeMu)

	// 读取聊天 ws 消息
	for {
		msgType, msgContent, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				global.Logger.Warnf("聊天 ws 读取异常关闭 user_id=%d err=%v", claims.UserID, err)
			} else {
				global.Logger.Infof("聊天 ws 已断开 user_id=%d err=%v", claims.UserID, err)
			}
			return
		}

		if msgType != websocket.TextMessage && msgType != websocket.BinaryMessage {
			continue
		}
		if err := writeChatWSEcho(conn, msgType, msgContent, &writeMu); err != nil {
			global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
			return
		}
	}
}

// configureChatWSConn 初始化连接的读限制和 pong 续期逻辑。
func configureChatWSConn(conn *websocket.Conn) {
	conn.SetReadLimit(chatWSReadLimit)
	conn.SetReadDeadline(time.Now().Add(chatWSPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(chatWSPongWait))
	})
}

// startChatWSPingLoop 定时发送 ping，浏览器会自动回 pong。
// 只要 pong 按时回来，读超时就会持续向后续期。
func startChatWSPingLoop(conn *websocket.Conn, userID uint, done <-chan struct{}, writeMu *sync.Mutex) {
	ticker := time.NewTicker(chatWSPingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			writeMu.Lock()
			conn.SetWriteDeadline(time.Now().Add(chatWSWriteWait))
			err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(chatWSWriteWait))
			writeMu.Unlock()
			if err != nil {
				global.Logger.Warnf("聊天 ws ping 失败 user_id=%d err=%v", userID, err)
				return
			}
		}
	}
}

// writeChatWSEcho 发送一条 echo 消息。
// ping 和业务回包都会写同一条连接，这里用互斥锁串行化写操作。
func writeChatWSEcho(conn *websocket.Conn, msgType int, msgContent []byte, writeMu *sync.Mutex) error {
	writeMu.Lock()
	defer writeMu.Unlock()

	conn.SetWriteDeadline(time.Now().Add(chatWSWriteWait))
	return conn.WriteMessage(msgType, []byte(fmt.Sprintf("你说的是：%s", string(msgContent))))
}
