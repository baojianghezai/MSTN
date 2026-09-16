package hrc

import (
	"sync"

	"github.com/gorilla/websocket"
)

// ChatClient 单个 WebSocket 连接（每个连接有独立写锁，避免多发送方并发写 panic）
type ChatClient struct {
	UID  uint64
	Conn *websocket.Conn
	mu   sync.Mutex
}

// NewChatClient 构造带写锁的连接包装
func NewChatClient(uid uint64, conn *websocket.Conn) *ChatClient {
	return &ChatClient{UID: uid, Conn: conn}
}

// WriteJSON 串行写 JSON（gorilla 连接不支持并发写）
func (c *ChatClient) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}

// ChatHub 在线连接注册表（uid → 连接集合，允许同账号多标签页）
type ChatHub struct {
	mu    sync.RWMutex
	conns map[uint64]map[*ChatClient]struct{}
}

// NewChatHub 构造连接注册表
func NewChatHub() *ChatHub {
	return &ChatHub{conns: make(map[uint64]map[*ChatClient]struct{})}
}

// Register 注册连接
func (h *ChatHub) Register(client *ChatClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[client.UID] == nil {
		h.conns[client.UID] = make(map[*ChatClient]struct{})
	}
	h.conns[client.UID][client] = struct{}{}
}

// Unregister 注销连接
func (h *ChatHub) Unregister(client *ChatClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.conns[client.UID]; ok {
		delete(set, client)
		if len(set) == 0 {
			delete(h.conns, client.UID)
		}
	}
}

// SendTo 向指定 uid 的所有在线连接推送消息；离线则静默丢弃（落库兜底）
func (h *ChatHub) SendTo(uid uint64, payload interface{}) {
	h.mu.RLock()
	clients := make([]*ChatClient, 0, len(h.conns[uid]))
	for c := range h.conns[uid] {
		clients = append(clients, c)
	}
	h.mu.RUnlock()
	for _, c := range clients {
		_ = c.WriteJSON(payload)
	}
}

// ChatHubInstance 全局单例
var ChatHubInstance = NewChatHub()
