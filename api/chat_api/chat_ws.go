package chat_api

import (
	"fmt"
	"myblogx/global"
	"myblogx/service/chat_service"
	"myblogx/utils/jwts"
	"net/http"
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
func (ChatApi) ChatWsView(c *gin.Context) {
	claims := jwts.MustGetClaimsByGin(c)

	// 升级 ws 连接
	rawConn, err := chatWSUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.Logger.Errorf("升级聊天 ws 连接失败 user_id=%d err=%v", claims.UserID, err)
		return
	}

	// 将聊天 ws 连接注册到在线用户中
	conn := chat_service.NewChatConn(claims.UserID, rawConn)
	store := chat_service.DefaultOnlineUserStore
	store.Register(conn)
	defer func() {
		if err := conn.Close(); err != nil {
			global.Logger.Warnf("关闭聊天 ws 连接失败 user_id=%d err=%v", claims.UserID, err)
		}
		store.Unregister(conn)
		global.Logger.Infof("聊天 ws 已清理 user_id=%d online_conn_count=%d", claims.UserID, store.Count(claims.UserID))
	}()

	// 配置 ws 连接
	configureChatWSConn(conn.Conn)
	global.Logger.Infof("聊天 ws 已连接 user_id=%d online_conn_count=%d", claims.UserID, store.Count(claims.UserID))

	done := make(chan struct{})
	defer close(done)

	go func() {
		if err := conn.RunPingLoop(done, chatWSPingPeriod, chatWSWriteWait); err != nil {
			global.Logger.Warnf("聊天 ws ping 失败 user_id=%d err=%v", conn.UserID, err)
		}
	}()

	for {
		msgType, msgContent, err := conn.Conn.ReadMessage()
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
		if err := writeChatWSEcho(conn, msgType, msgContent); err != nil {
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

// writeChatWSEcho 发送一条 echo 消息。
// ChatConn 内部已经封装了写锁，这里只负责设置超时和写业务消息。
func writeChatWSEcho(conn *chat_service.ChatConn, msgType int, msgContent []byte) error {
	_ = conn.SetWriteDeadline(time.Now().Add(chatWSWriteWait))
	return conn.WriteMessage(msgType, []byte(fmt.Sprintf("你说的是：%s", string(msgContent))))
}
