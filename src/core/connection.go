package mattermost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"xiaozhi-server-go/src/configs"
	"xiaozhi-server-go/src/core/utils"
)

// Client Mattermost客户端
type Client struct {
	config   *configs.MattermostConfig
	logger   *utils.Logger
	baseURL  string
	token    string
	timeout  time.Duration
}

// SendOptions 发送消息选项
type SendOptions struct {
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
	ThreadID  string `json:"thread_id,omitempty"`
}

// NewClient 创建Mattermost客户端
func NewClient(config *configs.MattermostConfig, logger *utils.Logger) *Client {
	return &Client{
		config:  config,
		logger:  logger,
		baseURL: config.BaseURL,
		token:   config.Token,
		timeout: time.Duration(config.Timeout) * time.Second,
	}
}

// SendMessage 发送消息到Mattermost
func (c *Client) SendMessage(ctx context.Context, channel, message string) error {
	return c.SendMessageWithOptions(ctx, &SendOptions{
		ChannelID: channel,
		Message:   message,
	})
}

// SendMessageWithOptions 支持更多选项的发送消息方法
func (c *Client) SendMessageWithOptions(ctx context.Context, opts *SendOptions) error {
	if c.token == "" {
		return fmt.Errorf("Mattermost token未配置")
	}

	// 构建API请求URL
	url := fmt.Sprintf("%s/posts", strings.TrimSuffix(c.baseURL, "/"))

	// 构建请求体
	requestBody := map[string]interface{}{
		"channel_id": opts.ChannelID,
		"message":    opts.Message,
	}

	// 如果有ThreadID，添加到请求体中
	if opts.ThreadID != "" {
		requestBody["root_id"] = opts.ThreadID
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	// 发送请求
	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Error("读取Mattermost API响应体失败: %v", err)
			return fmt.Errorf("读取响应失败: %v", err)
		}
		return fmt.Errorf("API返回错误: %d, %s", resp.StatusCode, string(body))
	}

	c.logger.Info("Mattermost消息发送成功 - 频道: %s", opts.ChannelID)
	return nil
}

// Initialize 初始化客户端
func (c *Client) Initialize() error {
	c.logger.Info("Mattermost客户端初始化成功")
	return nil
}

// Cleanup 清理资源
func (c *Client) Cleanup() error {
	// 补充：关闭HTTP客户端连接池（若有）
	c.logger.Info("Mattermost客户端清理完成")
	return nil
}
