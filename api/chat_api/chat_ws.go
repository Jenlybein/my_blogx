package chat_api

import (
	"encoding/json"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
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
	store := chat_service.GetOnlineUserStore()
	store.Register(conn)
	defer func() {
		if err := conn.Close(); err != nil {
			global.Logger.Warnf("关闭聊天 ws 连接失败 user_id=%d err=%v", claims.UserID, err)
		}
		store.Unregister(conn)
		global.Logger.Infof("聊天 ws 已清理 user_id=%d online_conn_count=%d", claims.UserID, store.Count(claims.UserID))
	}()

	// 配置 ws 连接
	configureChatWSConn(conn)
	global.Logger.Infof("聊天 ws 已连接 user_id=%d online_conn_count=%d", claims.UserID, store.Count(claims.UserID))

	// 发送 ws ping 心跳
	done := make(chan struct{})
	defer close(done)
	go func() {
		if err := conn.RunPingLoop(done, chatWSPingPeriod, chatWSWriteWait); err != nil {
			global.Logger.Warnf("聊天 ws ping 失败 user_id=%d err=%v", conn.UserID, err)
		}
	}()

	for {
		// 读取 ws 消息
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

		// 消息格式校验
		var req ChatRequest
		if err := json.Unmarshal(msgContent, &req); err != nil {
			if err := res.SendConnFailWithMsg("消息格式错误", conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
		}

		// 接收人不存在
		var revUser models.UserModel
		if err := global.DB.First(&revUser, req.ReceiverID).Error; err != nil {
			if err := res.SendConnFailWithMsg("接收人不存在", conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
			continue
		}

		// 落库（TODO:后续处理MsgType)
		var msgModel *models.ChatMsgModel
		if msgModel, err = chat_service.ToTextChat(chat_service.ToTextChatRequest{
			SenderID:   claims.UserID,
			ReceiverID: req.ReceiverID,
			Text:       req.Content,
		}); err != nil {
			if err := res.SendConnFailWithMsg("消息发送失败", conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
			global.Logger.Warnf("聊天消息落库失败 user_id=%d err=%v", claims.UserID, err)
			continue
		}

		// 判断接收人在不在线
		if !store.IsOnline(req.ReceiverID) {
			if err := res.SendConnFailWithMsg("接收人不在线", conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
			continue
		}

		// 在线则发送消息
		item := ChatMsgResponse{
			Content:    msgModel.Content,
			MsgType:    msgModel.MsgType,
			ID:         msgModel.ID,
			SendTime:   msgModel.SendTime,
			SenderID:   msgModel.SenderID,
			ReceiverID: msgModel.ReceiverID,
			SessionID:  msgModel.SessionID,
			IsSelf:     msgModel.SenderID == claims.UserID,
			IsRead:     false, // TODO：READ逻辑
			MsgStatus:  msgModel.MsgStatus,
		}
		if successCount := res.SendWsMsg(item, store, req.ReceiverID); successCount == 0 {
			if err := res.SendConnFailWithMsg("消息发送失败", conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
		}
		// Debug：给自己发一份
		if err := res.SendConnOkWithData(item, conn, chatWSWriteWait); err != nil {
			global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
			return
		}
	}
}

// configureChatWSConn 初始化连接的读限制和 pong 续期逻辑。
func configureChatWSConn(conn *chat_service.ChatConn) {
	conn.SetReadLimit(chatWSReadLimit)
	_ = conn.SetReadDeadline(time.Now().Add(chatWSPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(chatWSPongWait))
	})
}
