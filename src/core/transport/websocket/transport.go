package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
	"xiaozhi-server-go/src/configs"
	"xiaozhi-server-go/src/core/transport"
	"xiaozhi-server-go/src/core/utils"

	"github.com/gorilla/websocket"
)

// WebSocketTransport WebSocket传输层实现
type WebSocketTransport struct {
	config            *configs.Config
	server            *http.Server
	logger            *utils.Logger
	connHandler       transport.ConnectionHandlerFactory
	activeConnections sync.Map
	upgrader          *websocket.Upgrader
}

// NewWebSocketTransport 创建新的WebSocket传输层
func NewWebSocketTransport(config *configs.Config, logger *utils.Logger) *WebSocketTransport {
	return &WebSocketTransport{
		config: config,
		logger: logger,
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源的连接（兼容ESP32无Origin的情况）
			},
			// 新增：增大读写缓冲区，兼容ESP32大体积语音数据传输
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			// 新增：握手超时时间延长到10秒，给ESP32足够的握手时间
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// Start 启动WebSocket传输层
func (t *WebSocketTransport) Start(ctx context.Context) error {
	// 兼容配置：若IP配置为空，默认监听0.0.0.0（确保局域网可访问）
	listenIP := t.config.Transport.WebSocket.IP
	if listenIP == "" || listenIP == "127.0.0.1" {
		listenIP = "0.0.0.0"
		t.logger.Warn("WebSocket监听IP配置为本地回环，强制改为0.0.0.0（局域网可访问）")
	}
	addr := fmt.Sprintf("%s:%d", listenIP, t.config.Transport.WebSocket.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", t.handleWebSocket)

	t.server = &http.Server{
		Addr:    addr,
		Handler: mux,
		// 新增：设置服务器读写超时，避免ESP32连接假死
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	t.logger.Info("启动WebSocket传输层 ws://%s", addr)

	// 监听关闭信号
	go func() {
		<-ctx.Done()
		t.Stop()
	}()

	if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("WebSocket传输层启动失败: %v", err)
	}

	return nil
}

// Stop 停止WebSocket传输层
func (t *WebSocketTransport) Stop() error {
	if t.server != nil {
		t.logger.Info("停止WebSocket传输层...")

		// 关闭所有活动连接
		t.activeConnections.Range(func(key, value interface{}) bool {
			if handler, ok := value.(transport.ConnectionHandler); ok {
				handler.Close()
			}
			t.activeConnections.Delete(key)
			return true
		})

		return t.server.Close()
	}
	return nil
}

// SetConnectionHandler 设置连接处理器工厂
func (t *WebSocketTransport) SetConnectionHandler(handler transport.ConnectionHandlerFactory) {
	t.connHandler = handler
}

// GetActiveConnectionCount 获取活跃连接数
func (t *WebSocketTransport) GetActiveConnectionCount() (int, int) {
	count := 0
	t.activeConnections.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count, count
}

// GetType 获取传输类型
func (t *WebSocketTransport) GetType() string {
	return "websocket"
}

// handleWebSocket 处理WebSocket连接（优化ESP32兼容）
func (t *WebSocketTransport) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 新增：打印ESP32的完整请求信息，方便排查
	t.logger.Info("[ESP32兼容] 收到连接请求 - URL: %s, Header: %v", r.URL.String(), r.Header)

	// 升级WebSocket连接（兼容ESP32握手）
	conn, err := t.upgrader.Upgrade(w, r, nil)
	if err != nil {
		t.logger.Error("[ESP32兼容] WebSocket升级失败: %v", err)
		return
	}

	// 兼容ESP32无Device-Id/Client-Id的情况
	deviceID := r.Header.Get("Device-Id")
	if deviceID == "" {
		deviceID = r.URL.Query().Get("device-id")
		if deviceID == "" {
			deviceID = "ESP32-" + fmt.Sprintf("%d", time.Now().UnixNano()%10000) // 自动生成Device-Id
		}
		r.Header.Set("Device-Id", deviceID)
		t.logger.Info("[ESP32兼容] 自动生成Device-Id: %s", deviceID)
	}

	clientID := r.Header.Get("Client-Id")
	if clientID == "" {
		clientID = r.URL.Query().Get("client-id")
		if clientID == "" {
			clientID = fmt.Sprintf("ESP32-CLIENT-%p", conn) // 兼容空Client-Id
		}
		r.Header.Set("Client-Id", clientID)
	}

	t.logger.Info("[WebSocket] [ESP32连接请求] DeviceID: %s, ClientID: %s", deviceID, clientID)
	wsConn := NewWebSocketConnection(clientID, conn)

	// 兼容：无连接处理器时，不立即关闭连接（给ESP32响应时间）
	if t.connHandler == nil {
		t.logger.Warn("[ESP32兼容] 连接处理器工厂未设置，保留连接30秒")
		// 临时保留连接30秒，避免ESP32立即收到关闭信号
		go func() {
			time.Sleep(30 * time.Second)
			conn.Close()
		}()
		return
	}

	handler := t.connHandler.CreateHandler(wsConn, r)
	if handler == nil {
		t.logger.Error("[ESP32兼容] 创建连接处理器失败，保留连接30秒")
		// 临时保留连接30秒
		go func() {
			time.Sleep(30 * time.Second)
			conn.Close()
		}()
		return
	}

	t.activeConnections.Store(clientID, handler)
	t.logger.Info("[WebSocket] [ESP32连接建立] DeviceID: %s, ClientID: %s 资源已分配", deviceID, clientID)

	// 启动连接处理，并在结束时清理资源
	go func() {
		defer func() {
			// 连接结束时清理
			t.activeConnections.Delete(clientID)
			handler.Close()
			t.logger.Info("[WebSocket] [ESP32连接断开] ClientID: %s", clientID)
		}()

		handler.Handle()
	}()
}
