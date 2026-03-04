package mattermost

import (
	"net/http"
	"xiaozhi-server-go/src/core/utils"

	"github.com/gin-gonic/gin"
)

// Handler HTTP处理器
type Handler struct {
	client *Client
	logger *utils.Logger
}

// NewHandler 创建处理器
func NewHandler(client *Client, logger *utils.Logger) *Handler {
	return &Handler{
		client: client,
		logger: logger,
	}
}

// SendRequest 发送消息请求结构
type SendRequest struct {
	Channel string `json:"channel" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// Send 发送消息到Mattermost
func (h *Handler) Send(c *gin.Context) {
	var req SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Mattermost发送请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.client.SendMessage(c.Request.Context(), req.Channel, req.Message)
	if err != nil {
		h.logger.Error("Mattermost发送消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "消息发送成功"})
}

// Health 健康检查
func (h *Handler) Health(c *gin.Context) {
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "Mattermost客户端未初始化",
		})
		return
	}

	// 简单的健康检查 - 验证配置
	if h.client.config.BaseURL == "" || h.client.config.Token == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "Mattermost配置不完整",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "mattermost",
		"config": map[string]interface{}{
			"base_url": h.client.config.BaseURL,
			"timeout":  h.client.config.Timeout,
		},
	})
}
