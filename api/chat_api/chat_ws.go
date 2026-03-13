package chat_api

import (
	"encoding/json"
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum/chat_msg_enum"
	"myblogx/models/enum/relationship_enum"
	"myblogx/service/chat_service"
	"myblogx/service/follow_service"
	"myblogx/utils/jwts"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
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

	// 升级成 ws 连接
	rawConn, err := chatWSUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.Logger.Errorf("升级聊天 ws 连接失败 user_id=%d err=%v", claims.UserID, err)
		return
	}

	// 注册聊天 ws 连接
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

	// 配置聊天 ws 连接
	configureChatWSConn(conn)
	global.Logger.Infof("聊天 ws 已连接 user_id=%d online_conn_count=%d", claims.UserID, store.Count(claims.UserID))

	// 心跳检测
	done := make(chan struct{})
	defer close(done)
	go func() {
		if err := conn.RunPingLoop(done, chatWSPingPeriod, chatWSWriteWait); err != nil {
			global.Logger.Warnf("聊天 ws ping 失败 user_id=%d err=%v", conn.UserID, err)
		}
	}()

	for {
		// 读取消息
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

		// 解析消息
		var req ChatRequest
		if err := json.Unmarshal(msgContent, &req); err != nil {
			if err := res.SendConnFailWithMsg("消息格式错误", conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
			continue
		}

		// 检测接收人
		var revUser models.UserModel
		if err := global.DB.Preload("UserConfModel").First(&revUser, req.ReceiverID).Error; err != nil {
			if err := res.SendConnFailWithMsg("接收人不存在", conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
			continue
		}

		// 检测发送权限
		if err := validateChatSendPermission(claims.UserID, &revUser); err != nil {
			if err := res.SendConnFailWithMsg(err.Error(), conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
			continue
		}

		// 处理消息进入数据库
		var msgModel *models.ChatMsgModel
		var msgErr error
		switch req.MsgType {
		case chat_msg_enum.MsgTypeText:
			msgModel, msgErr = chat_service.ToTextChat(chat_service.ToTextChatRequest{
				SenderID:   claims.UserID,
				ReceiverID: req.ReceiverID,
				Text:       req.Content,
			})
		case chat_msg_enum.MsgTypeImage:
			msgModel, msgErr = chat_service.ToImageChat(chat_service.ToImageChatRequest{
				SenderID:   claims.UserID,
				ReceiverID: req.ReceiverID,
				ImageURL:   req.Content,
			})
		case chat_msg_enum.MsgTypeMarkdown:
			msgModel, msgErr = chat_service.ToMarkdownChat(chat_service.ToMarkdownChatRequest{
				SenderID:   claims.UserID,
				ReceiverID: req.ReceiverID,
				Markdown:   req.Content,
			})
		default:
			msgErr = errors.New("不支持的消息类型")
		}
		if msgErr != nil {
			if err := res.SendConnFailWithMsg("消息发送失败", conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
			global.Logger.Warnf("聊天消息落库失败 user_id=%d err=%v", claims.UserID, msgErr)
			continue
		}

		// 检测接收人是否在线
		if !store.IsOnline(req.ReceiverID) {
			if err := res.SendConnFailWithMsg("接收人不在线", conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
			continue
		}

		// 发送消息
		item := ChatMsgResponse{
			Content:    msgModel.Content,
			MsgType:    msgModel.MsgType,
			ID:         msgModel.ID,
			SendTime:   msgModel.SendTime,
			SenderID:   msgModel.SenderID,
			ReceiverID: msgModel.ReceiverID,
			SessionID:  msgModel.SessionID,
			IsSelf:     msgModel.SenderID == claims.UserID,
			IsRead:     false,
			MsgStatus:  msgModel.MsgStatus,
		}
		if req.ReceiverID == claims.UserID {
			if successCount := res.SendWsMsg(item, store, req.ReceiverID); successCount == 0 {
				if err := res.SendConnOkWithMsg("给自己发送消息", conn, chatWSWriteWait); err != nil {
					global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
					return
				}
			}
			continue
		}
		if successCount := res.SendWsMsg(item, store, req.ReceiverID); successCount == 0 {
			if err := res.SendConnFailWithMsg("消息发送失败", conn, chatWSWriteWait); err != nil {
				global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
				return
			}
		}

		// DEBUG：给自己也发一份
		if err := res.SendConnOkWithData(item, conn, chatWSWriteWait); err != nil {
			global.Logger.Warnf("聊天 ws 写入失败 user_id=%d err=%v", claims.UserID, err)
			return
		}
	}
}

// 初始化连接的读限制和 pong 续期逻辑。
func configureChatWSConn(conn *chat_service.ChatConn) {
	conn.SetReadLimit(chatWSReadLimit)
	_ = conn.SetReadDeadline(time.Now().Add(chatWSPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(chatWSPongWait))
	})
}

// 检测好友关系发送权限
func validateChatSendPermission(senderID uint, receiver *models.UserModel) error {
	if senderID == receiver.ID {
		return nil
	}

	// 陌生人：如果用户设置接收陌生人消息才允许发送
	// 好友：好友之间可以互发消息
	// 粉丝：若关注者未回复，粉丝每周可以向关注者发送4条消息
	// 关注者：若粉丝未回复，关注者每周可以向粉丝发送4条消息
	relation := follow_service.CalUserRelationship(senderID, receiver.ID)
	switch relation {
	case relationship_enum.RelationFriend:
		return nil
	case relationship_enum.RelationStranger:
		if receiver.UserConfModel != nil && receiver.UserConfModel.StrangerChatEnabled {
			return nil
		}
		return errors.New("对方未开启陌生人私信")
	case relationship_enum.RelationFans, relationship_enum.RelationFollowed:
		cutoff := time.Now().AddDate(0, 0, -7)
		startTime := cutoff

		var lastReply models.ChatMsgModel
		err := global.DB.Select("send_time").
			Where("sender_id = ? AND receiver_id = ? AND send_time >= ?", receiver.ID, senderID, cutoff).
			Order("send_time desc").
			Take(&lastReply).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil && lastReply.SendTime.After(startTime) {
			startTime = lastReply.SendTime
		}

		var count int64
		if err := global.DB.Model(&models.ChatMsgModel{}).
			Where("sender_id = ? AND receiver_id = ? AND send_time >= ?", senderID, receiver.ID, startTime).
			Count(&count).Error; err != nil {
			return err
		}
		if count < 4 {
			return nil
		}
		return errors.New("本周可发送消息次数已达上限，请等待对方回复")
	default:
		return errors.New("当前关系不支持发送消息")
	}
}
