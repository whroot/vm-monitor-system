package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	HeartbeatInterval time.Duration // 心跳间隔
	WriteTimeout      time.Duration // 写超时
	ReadTimeout       time.Duration // 读超时
	MaxMessageSize   int64        // 最大消息大小
}

// VMMetricsMessage VM指标消息
type VMMetricsMessage struct {
	Type      string                 `json:"type"`       // metrics, alert, heartbeat
	VMID      string                 `json:"vmId"`      // VM ID
	Timestamp time.Time             `json:"timestamp"`  // 时间戳
	Data      map[string]interface{} `json:"data"`      // 指标数据
	Alert     *AlertMessage         `json:"alert,omitempty"` // 告警信息
}

// AlertMessage 告警消息
type AlertMessage struct {
	AlertID   string    `json:"alertId"`
	RuleID    string    `json:"ruleId"`
	RuleName  string    `json:"ruleName"`
	Level     string    `json:"level"` // critical, warning, info
	Message   string    `json:"message"`
	VMID      string    `json:"vmId"`
	VMName    string    `json:"vmName"`
	Timestamp time.Time `json:"timestamp"`
}

// SubscribeMessage 订阅消息
type SubscribeMessage struct {
	Type   string   `json:"type"` // subscribe, unsubscribe
	VMIDs  []string `json:"vmIds"` // 订阅的VM列表
	Groups []string `json:"groups"` // 订阅的分组
}

// WebSocketHub WebSocket连接管理器
type WebSocketHub struct {
	config      *WebSocketConfig
	upgrader   websocket.Upgrader
	clients    map[string]*Client
	broadcast  chan *VMMetricsMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	stopChan   chan struct{}
}

// Client WebSocket客户端
type Client struct {
	ID        string
	UserID    string
	Conn      *websocket.Conn
	Send      chan []byte
	Hub       *WebSocketHub
	Subscribed map[string]bool // 订阅的VM ID
}

// NewWebSocketHub 创建WebSocket Hub
func NewWebSocketHub() *WebSocketHub {
	config := &WebSocketConfig{
		HeartbeatInterval: 30 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       60 * time.Second,
		MaxMessageSize:   512 * 1024, // 512KB
	}

	return &WebSocketHub{
		config:    config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:    make(map[string]*Client),
		broadcast:  make(chan *VMMetricsMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stopChan:   make(chan struct{}),
	}
}

// Start 启动WebSocket Hub
func (h *WebSocketHub) Start() {
	log.Println("WebSocket Hub 启动")
	go h.run()
	log.Println("WebSocket Hub 已启动")
}

// Stop 停止WebSocket Hub
func (h *WebSocketHub) Stop() {
	close(h.stopChan)
	h.mu.Lock()
	for _, client := range h.clients {
		close(client.Send)
		client.Conn.Close()
	}
	h.mu.Unlock()
	log.Println("WebSocket Hub 已停止")
}

// run 连接管理主循环
func (h *WebSocketHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("WebSocket 客户端已连接: %s", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket 客户端已断开: %s", client.ID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				client.Send <- h.encodeMessage(message)
			}
			h.mu.RUnlock()
		}
	}
}

// HandleConnection 处理新的WebSocket连接
func (h *WebSocketHub) HandleConnection(w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 连接升级失败: %v", err)
		return
	}

	client := &Client{
		ID:        generateWSClientID(),
		UserID:    userID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Hub:       h,
		Subscribed: make(map[string]bool),
	}

	h.register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// encodeMessage 编码消息
func (h *WebSocketHub) encodeMessage(msg *VMMetricsMessage) []byte {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("消息编码失败: %v", err)
		return nil
	}
	return data
}

// Broadcast 广播消息到所有客户端
func (h *WebSocketHub) Broadcast(message *VMMetricsMessage) {
	select {
	case h.broadcast <- message:
	default:
		log.Printf("WebSocket 消息队列已满，丢弃消息")
	}
}

// Subscribe 客户端订阅VM
func (c *Client) Subscribe(vmIDs []string) {
	for _, vmID := range vmIDs {
		c.Subscribed[vmID] = true
	}
}

// Unsubscribe 取消订阅VM
func (c *Client) Unsubscribe(vmIDs []string) {
	for _, vmID := range vmIDs {
		delete(c.Subscribed, vmID)
	}
}

// readPump 读消息泵
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(c.Hub.config.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.Hub.config.ReadTimeout))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.Hub.config.ReadTimeout))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 错误: %v", err)
			}
			break
		}

		var subMsg SubscribeMessage
		if err := json.Unmarshal(message, &subMsg); err != nil {
			log.Printf("订阅消息解析失败: %v", err)
			continue
		}

		switch subMsg.Type {
		case "subscribe":
			c.Subscribe(subMsg.VMIDs)
			log.Printf("客户端 %s 订阅 VM: %v", c.ID, subMsg.VMIDs)
		case "unsubscribe":
			c.Unsubscribe(subMsg.VMIDs)
			log.Printf("客户端 %s 取消订阅 VM: %v", c.ID, subMsg.VMIDs)
		}
	}
}

// writePump 写消息泵
func (c *Client) writePump() {
	ticker := time.NewTicker(c.Hub.config.HeartbeatInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Hub.config.WriteTimeout))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Hub.config.WriteTimeout))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateWSClientID 生成客户端ID
func generateWSClientID() string {
	return "ws_" + uuid.New().String()[:8]
}
