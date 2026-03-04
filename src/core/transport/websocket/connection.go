package websocket

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketConnection WebSocket连接适配器
type WebSocketConnection struct {
	id         string
	conn       *websocket.Conn
	closed     int32
	lastActive int64
	mu         sync.Mutex
	// 新增：心跳保活相关
	heartbeatTicker *time.Ticker // 心跳定时器
	stopHeartbeat   chan struct{}// 停止心跳通道
}

// NewWebSocketConnection 创建新的WebSocket连接适配器
func NewWebSocketConnection(id string, conn *websocket.Conn) *WebSocketConnection {
	wsConn := &WebSocketConnection{
		id:         id,
		conn:       conn,
		closed:     0,
		lastActive: time.Now().Unix(),
		stopHeartbeat: make(chan struct{}), // 初始化停止通道
	}
	// 启动心跳保活（30秒一次Ping，防止连接断开）
	wsConn.startHeartbeat()
	return wsConn
}

// 新增：启动心跳保活协程
func (c *WebSocketConnection) startHeartbeat() {
	// 设置30秒发送一次Ping包
	c.heartbeatTicker = time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-c.stopHeartbeat:
				c.heartbeatTicker.Stop()
				return
			case <-c.heartbeatTicker.C:
				// 发送Ping包，更新活跃时间，防止连接断开
				if !c.IsClosed() {
					c.mu.Lock()
					// 发送Ping消息（空内容），触发Pong响应
					err := c.conn.WriteMessage(websocket.PingMessage, nil)
					if err == nil {
						atomic.StoreInt64(&c.lastActive, time.Now().Unix())
					}
					c.mu.Unlock()
				}
			}
		}
	}()
}

// WriteMessage 发送消息
func (c *WebSocketConnection) WriteMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.closed) == 1 {
		return fmt.Errorf("连接已关闭")
	}

	atomic.StoreInt64(&c.lastActive, time.Now().Unix())
	return c.conn.WriteMessage(messageType, data)
}

// ReadMessage 读取消息
func (c *WebSocketConnection) ReadMessage(stopChan <-chan struct{}) (int, []byte, error) {
	// 新增：读取消息时增加超时处理，避免阻塞导致连接假死
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 60秒读取超时
	messageType, data, err := c.conn.ReadMessage()
	
	if err == nil {
		atomic.StoreInt64(&c.lastActive, time.Now().Unix())
		// 重置读取超时（收到数据后延长超时）
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	} else {
		// 仅打印错误，不主动关闭连接（兼容ESP32断连重连）
		fmt.Printf("读取WS消息: %v（非致命，保留连接）\n", err)
	}
	return messageType, data, err
}

// Close 关闭连接
func (c *WebSocketConnection) Close() error {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		// 新增：停止心跳保活
		close(c.stopHeartbeat)
		if c.heartbeatTicker != nil {
			c.heartbeatTicker.Stop()
		}
		return c.conn.Close()
	}
	return nil
}

// GetID 获取连接ID
func (c *WebSocketConnection) GetID() string {
	return c.id
}

// GetType 获取连接类型
func (c *WebSocketConnection) GetType() string {
	return "websocket"
}

// IsClosed 检查连接是否已关闭
func (c *WebSocketConnection) IsClosed() bool {
	return atomic.LoadInt32(&c.closed) == 1
}

// GetLastActiveTime 获取最后活跃时间
func (c *WebSocketConnection) GetLastActiveTime() time.Time {
	return time.Unix(atomic.LoadInt64(&c.lastActive), 0)
}

// IsStale 检查连接是否过期
func (c *WebSocketConnection) IsStale(timeout time.Duration) bool {
	return time.Since(c.GetLastActiveTime()) > timeout
}
