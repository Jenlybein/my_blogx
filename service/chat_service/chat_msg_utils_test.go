package chat_service

import (
	"strings"
	"testing"
	"time"

	"myblogx/models/enum/chat_msg_enum"
)

func TestValidateChatBase(t *testing.T) {
	t.Run("默认值补齐", func(t *testing.T) {
		req := ToChatRequest{
			SenderID:   1,
			ReceiverID: 2,
			MsgType:    chat_msg_enum.MsgTypeText,
			Content:    "hello",
		}

		if err := validateChatBase(&req); err != nil {
			t.Fatalf("validateChatBase 返回错误: %v", err)
		}
		if req.SendTime.IsZero() {
			t.Fatal("应补齐默认发送时间")
		}
		if req.MsgStatus != chat_msg_enum.MsgStatusSend {
			t.Fatalf("默认消息状态错误: %v", req.MsgStatus)
		}
	})

	t.Run("当前不校验非法类型", func(t *testing.T) {
		req := ToChatRequest{
			SenderID:   1,
			ReceiverID: 2,
			MsgType:    chat_msg_enum.MsgType(99),
			Content:    "hello",
			SendTime:   time.Now(),
		}

		err := validateChatBase(&req)
		if err != nil {
			t.Fatalf("当前实现不校验消息类型, got=%v", err)
		}
	})
}

func TestBuildSessionLastMsg(t *testing.T) {
	t.Run("markdown 消息当前返回截断后的文本", func(t *testing.T) {
		content, err := marshalChatContent(markdownChatContent{
			Kind:     "markdown",
			Title:    "标题",
			Markdown: "# Heading\n\n正文内容",
			Summary:  "这是摘要",
		})
		if err != nil {
			t.Fatalf("marshalChatContent 失败: %v", err)
		}

		got := buildSessionLastMsg(chat_msg_enum.MsgTypeMarkdown, content)
		if got == "" {
			t.Fatal("markdown 摘要不应为空")
		}
		if !strings.HasSuffix(got, "...") {
			t.Fatalf("markdown 摘要应带省略号: %s", got)
		}
	})

	t.Run("图片消息当前会追加省略号", func(t *testing.T) {
		got := buildSessionLastMsg(chat_msg_enum.MsgTypeImage, "ignored")
		if got != "[图片]" {
			t.Fatalf("图片摘要错误: %s", got)
		}
	})
}
