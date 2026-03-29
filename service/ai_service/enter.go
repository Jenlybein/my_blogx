package ai_service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"myblogx/global"
	"net/http"
	"strings"
)

// AI对话消息结构体
type Message struct {
	Role    string `json:"role"`    // 角色：system/user/assistant
	Content string `json:"content"` // 消息内容
}

// AI请求结构体
type Request struct {
	Model    string    `json:"model"`    // 模型名称
	Messages []Message `json:"messages"` // 对话消息列表
	Stream   bool      `json:"stream"`   // 是否开启流式响应
}

// AI非流式响应结构体
type ChatCompletion struct {
	ID      string `json:"id"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Object  string `json:"object"`
	Usage   struct {
		CompletionTokens        int `json:"completion_tokens"`
		PromptTokens            int `json:"prompt_tokens"`
		TotalTokens             int `json:"total_tokens"`
		CompletionTokensDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
		PromptTokensDetails struct {
			AudioTokens  int `json:"audio_tokens"`
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
	} `json:"usage"`
	SystemFingerprint interface{} `json:"system_fingerprint"`
}

// AI流式响应结构体
type StreamData struct {
	ID      string `json:"id"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Logprobs     any `json:"logprobs"`
		FinishReason any `json:"finish_reason"`
	} `json:"choices"`
	Created           int    `json:"created"`
	Model             string `json:"model"`
	Object            string `json:"object"`
	SystemFingerprint any    `json:"system_fingerprint"`
}

// 基础请求方法，封装通用的AI请求逻辑
func BaseRequest(req Request) (*http.Response, error) {
	// 1. 基础配置校验
	if global.Config == nil {
		return nil, errors.New("系统配置未初始化")
	}
	if !global.Config.AI.Enable {
		return nil, errors.New("AI服务未开启")
	}

	key := global.Config.AI.SecretKey
	url := global.Config.AI.BaseURL

	if key == "" {
		return nil, errors.New("AI SecretKey 未配置")
	}
	if url == "" {
		return nil, errors.New("AI BaseURL 未配置")
	}

	// 2. 序列化请求体
	byteData, err := json.Marshal(req)
	if err != nil {
		global.Logger.Errorf("请求体序列化失败: 错误=%v", err)
		return nil, fmt.Errorf("请求体序列化失败: %w", err)
	}

	// 3. 创建HTTP请求
	reqHTTP, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(byteData))
	if err != nil {
		global.Logger.Errorf("创建 HTTP 请求失败: 错误=%v", err)
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 4. 设置请求头
	reqHTTP.Header.Add("Authorization", fmt.Sprintf("Bearer %s", key))
	reqHTTP.Header.Add("Content-Type", "application/json; charset=utf-8")

	// 5. 发送请求
	client := &http.Client{}
	res, err := client.Do(reqHTTP)
	if err != nil {
		global.Logger.Errorf("发送 AI 请求失败: 错误=%v", err)
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	return res, nil
}

// Chat 非流式AI对话接口
func Chat(msgList []Message) (string, error) {
	model := global.Config.AI.ChatModel
	return chatWithModel(msgList, model)
}

// ChatStream 流式AI对话接口（返回流式输出的内容通道）
func ChatStream(msgList []Message) (chan string, chan error) {
	model := global.Config.AI.ChatModel
	return chatStreamWithModel(msgList, model)
}

func chatWithModel(msgList []Message, model string) (string, error) {
	if len(msgList) == 0 {
		return "", errors.New("消息列表不能为空")
	}

	res, err := BaseRequest(Request{
		Model:    model,
		Messages: msgList,
		Stream:   false,
	})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		msg := strings.TrimSpace(string(bodyBytes))
		if msg == "" {
			return "", fmt.Errorf("请求失败，响应状态码: %d", res.StatusCode)
		}
		return "", fmt.Errorf("请求失败，响应状态码: %d，响应: %s", res.StatusCode, msg)
	}

	var response ChatCompletion
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		if global.Logger != nil {
			global.Logger.Errorf("解析 AI 响应失败: 错误=%v", err)
		}
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", errors.New("AI响应无有效Choices数据")
	}
	replyContent := response.Choices[0].Message.Content
	if replyContent == "" {
		return "", errors.New("AI回复内容为空")
	}

	return replyContent, nil
}

func chatStreamWithModel(msgList []Message, model string) (chan string, chan error) {
	// 初始化通道
	contentChan := make(chan string)
	errChan := make(chan error, 1) // 带缓冲，防止goroutine阻塞

	go func() {
		defer close(contentChan)
		defer close(errChan)

		if len(msgList) == 0 {
			errChan <- errors.New("消息列表不能为空")
			return
		}

		// 复用BaseRequest发送请求
		res, err := BaseRequest(Request{
			Model:    model,
			Messages: msgList,
			Stream:   true,
		})
		if err != nil {
			errChan <- err
			return
		}
		defer res.Body.Close()

		// 检查响应状态码
		if res.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("流式请求失败，响应状态码: %d", res.StatusCode)
			return
		}

		// 处理流式响应
		scanner := bufio.NewScanner(res.Body)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			text := scanner.Text()
			if text == "" {
				continue
			}

			// 处理SSE格式（data: 前缀）
			if !strings.HasPrefix(text, "data: ") {
				continue
			}
			data := strings.TrimPrefix(text, "data: ")

			// 结束标记
			if data == "[DONE]" {
				break
			}

			// 解析流式数据
			var item StreamData
			if err := json.Unmarshal([]byte(data), &item); err != nil {
				if global.Logger != nil {
					global.Logger.Warnf("流式数据解析错误: 错误=%v 原始数据=%s", err, text)
				}
				continue // 单个数据解析失败，继续处理下一条
			}

			// 发送内容到通道
			if len(item.Choices) > 0 {
				contentChan <- item.Choices[0].Delta.Content
			}
		}

		// 检查scanner错误
		if err := scanner.Err(); err != nil {
			if global.Logger != nil {
				global.Logger.Errorf("读取流式响应失败: 错误=%v", err)
			}
			errChan <- fmt.Errorf("读取响应流失败: %w", err)
		}
	}()

	return contentChan, errChan
}
