package client

import (
	"encoding/json"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/protocol"
)

// Client WebSocket 客户端
type Client struct {
	serverURL   string           // 服务器地址
	playerID    string           // 玩家ID
	playerName  string           // 玩家名称
	conn        *websocket.Conn  // WebSocket 连接
	connected   bool             // 是否已连接
	connecting  bool             // 是否正在连接
	reconnecting bool            // 是否正在重连
	send        chan []byte      // 发送消息通道
	receive     chan []byte      // 接收消息通道
	onStateChange  func(*protocol.GameState)       // 状态变化回调
	onJoinAck      func(bool, string, int)        // 加入确认回调(success, playerID, seat)
	onTurn         func(*protocol.YourTurn)       // 轮到玩家回合回调
	onShowdown     func(*protocol.Showdown)       // 摊牌结果回调
	onPlayerReady  func(*protocol.PlayerReadyNotify) // 玩家准备状态回调
	onChat         func(*protocol.ChatMessage)    // 收到聊天消息回调
	onError        func(error)                    // 错误回调
	onConnect      func()                         // 连接成功回调
	onDisconnect   func()                         // 断开连接回调
	mu          sync.RWMutex  // 读写锁
	done        chan struct{}  // 关闭信号
}

// Config 客户端配置
type Config struct {
	ServerURL   string               // 服务器地址
	PlayerName  string               // 玩家名称
	Seat        int                  // 请求座位号（-1表示随机）
	OnStateChange  func(*protocol.GameState)       // 状态变化回调
	OnJoinAck      func(bool, string, int)        // 加入确认回调(success, playerID, seat)
	OnTurn         func(*protocol.YourTurn)       // 轮到玩家回合回调
	OnShowdown     func(*protocol.Showdown)       // 摊牌结果回调
	OnPlayerReady  func(*protocol.PlayerReadyNotify) // 玩家准备状态回调
	OnChat         func(*protocol.ChatMessage)    // 收到聊天消息回调
	OnError        func(error)                    // 错误回调
	OnConnect      func()                         // 连接成功回调
	OnDisconnect   func()                         // 断开连接回调
}

// NewClient 创建新的客户端
func NewClient(config *Config) *Client {
	return &Client{
		serverURL:   config.ServerURL,
		playerName:  config.PlayerName,
		send:        make(chan []byte, 256),
		receive:     make(chan []byte, 256),
		onStateChange:  config.OnStateChange,
		onJoinAck:      config.OnJoinAck,
		onTurn:         config.OnTurn,
		onShowdown:     config.OnShowdown,
		onPlayerReady:  config.OnPlayerReady,
		onChat:         config.OnChat,
		onError:        config.OnError,
		onConnect:      config.OnConnect,
		onDisconnect:   config.OnDisconnect,
		done:         make(chan struct{}),
	}
}

// Connect 连接到服务器
func (c *Client) Connect() error {
	c.mu.Lock()
	if c.connected || c.connecting {
		c.mu.Unlock()
		return nil
	}
	c.connecting = true
	c.mu.Unlock()

	// 解析 URL
	u, err := url.Parse(c.serverURL)
	if err != nil {
		c.mu.Lock()
		c.connecting = false
		c.mu.Unlock()
		return err
	}

	// 构建 WebSocket URL
	wsURL := "ws://" + u.Host + "/ws"
	if u.Path != "" {
		wsURL = "ws://" + u.Host + u.Path + "/ws"
	}

	log.Printf("Connecting to %s...", wsURL)

	// 建立连接
	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		c.mu.Lock()
		c.connecting = false
		c.mu.Unlock()
		c.notifyError(err)
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.connecting = false
	c.mu.Unlock()

	log.Printf("Connected to %s", wsURL)

	// 启动读写协程
	go c.readPump()
	go c.writePump()

	// 发送加入游戏请求
	c.sendJoin()

	// 通知连接成功
	if c.onConnect != nil {
		c.onConnect()
	}

	return nil
}

// Disconnect 断开连接
func (c *Client) Disconnect() {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return
	}
	c.connected = false
	c.mu.Unlock()

	close(c.done)

	if c.conn != nil {
		c.conn.Close()
	}

	if c.onDisconnect != nil {
		c.onDisconnect()
	}
}

// Reconnect 重新连接
func (c *Client) Reconnect() {
	c.mu.Lock()
	if c.reconnecting {
		c.mu.Unlock()
		return
	}
	c.reconnecting = true
	c.mu.Unlock()

	log.Println("Reconnecting...")

	// 尝试重新连接
	for i := 0; i < 5; i++ {
		if err := c.Connect(); err == nil {
			c.reconnecting = false
			return
		}
		// 等待后重试
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	c.reconnecting = false
	log.Println("Failed to reconnect after 5 attempts")
}

// Send 发送消息
func (c *Client) Send(msg interface{}) error {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return websocket.ErrCloseSent
	}
	c.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		return websocket.ErrCloseSent
	}
}

// SendPlayerAction 发送玩家动作
func (c *Client) SendPlayerAction(action models.ActionType, amount int) error {
	req := &protocol.PlayerActionRequest{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypePlayerAction),
		PlayerID:   c.playerID,
		Action:     action,
		Amount:     amount,
	}
	return c.Send(req)
}

// SendChat 发送聊天消息
func (c *Client) SendChat(content string) error {
	req := &protocol.ChatRequest{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeChat),
		PlayerID:   c.playerID,
		Content:    content,
	}
	return c.Send(req)
}

// SendPing 发送心跳
func (c *Client) SendPing() error {
	req := &protocol.PingRequest{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypePing),
	}
	return c.Send(req)
}

// SendReadyForNext 发送准备下一局请求
func (c *Client) SendReadyForNext() error {
	req := protocol.NewReadyForNextRequest(c.playerID)
	return c.Send(req)
}

// PlayerID 获取玩家ID
func (c *Client) PlayerID() string {
	return c.playerID
}

// IsConnected 检查是否已连接
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// SetPlayerID 设置玩家ID
func (c *Client) SetPlayerID(id string) {
	c.playerID = id
}

// readPump 处理读取消息
func (c *Client) readPump() {
	defer func() {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		c.conn.Close()
		close(c.receive)

		if c.onDisconnect != nil {
			c.onDisconnect()
		}
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		select {
		case c.receive <- message:
		default:
			// 接收队列满，跳过
		}

		// 处理消息
		c.handleMessage(message)
	}
}

// writePump 处理发送消息
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(data []byte) {
	var baseMsg protocol.BaseMessage
	if err := json.Unmarshal(data, &baseMsg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	switch baseMsg.Type {
	case protocol.MsgTypeJoinAck:
		c.handleJoinAck(data)

	case protocol.MsgTypeGameState:
		c.handleGameState(data)

	case protocol.MsgTypeYourTurn:
		c.handleYourTurn(data)

	case protocol.MsgTypeChat:
		c.handleChat(data)

	case protocol.MsgTypeShowdown:
		c.handleShowdown(data)

	case protocol.MsgTypePlayerReady:
		c.handlePlayerReady(data)

	case protocol.MsgTypePlayerJoined:
		c.handlePlayerJoined(data)

	case protocol.MsgTypePlayerLeft:
		c.handlePlayerLeft(data)

	case protocol.MsgTypePlayerActed:
		c.handlePlayerActed(data)

	case protocol.MsgTypePong:
		// 心跳响应，忽略

	case protocol.MsgTypeError:
		c.handleError(data)

	default:
		log.Printf("Unknown message type: %s", baseMsg.Type)
	}
}

// handleJoinAck 处理加入确认
func (c *Client) handleJoinAck(data []byte) {
	var msg protocol.JoinAck
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal JoinAck: %v", err)
		return
	}

	if msg.Success {
		c.playerID = msg.PlayerID
		log.Printf("Joined game as %s at seat %d", c.playerID, msg.Seat)
		// 调用 JoinAck 回调
		if c.onJoinAck != nil {
			c.onJoinAck(true, msg.PlayerID, msg.Seat)
		}
	} else {
		log.Printf("Failed to join game: %s", msg.Message)
		c.notifyError(&GameError{Message: msg.Message})
		// 调用 JoinAck 回调（失败）
		if c.onJoinAck != nil {
			c.onJoinAck(false, "", -1)
		}
	}
}

// handleGameState 处理游戏状态更新
// 服务器直接发送 protocol.GameState 结构，不是嵌套的
func (c *Client) handleGameState(data []byte) {
	var gameState protocol.GameState
	if err := json.Unmarshal(data, &gameState); err != nil {
		log.Printf("Failed to unmarshal GameState: %v", err)
		return
	}

	if c.onStateChange != nil {
		c.onStateChange(&gameState)
	}
}

// handleYourTurn 处理轮到玩家回合
func (c *Client) handleYourTurn(data []byte) {
	var msg protocol.YourTurn
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal YourTurn: %v", err)
		return
	}

	log.Printf("It's your turn! Min: %d, Max: %d", msg.MinAction, msg.MaxAction)

	if c.onTurn != nil {
		c.onTurn(&msg)
	}
}

// handleChat 处理聊天消息
func (c *Client) handleChat(data []byte) {
	var msg protocol.ChatMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal ChatMessage: %v", err)
		return
	}

	if c.onChat != nil {
		c.onChat(&msg)
	}
}

// handleShowdown 处理摊牌结果消息
func (c *Client) handleShowdown(data []byte) {
	var msg protocol.Showdown
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal Showdown: %v", err)
		return
	}

	if c.onShowdown != nil {
		c.onShowdown(&msg)
	}
}

// handlePlayerReady 处理玩家准备状态通知
func (c *Client) handlePlayerReady(data []byte) {
	var msg protocol.PlayerReadyNotify
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal PlayerReady: %v", err)
		return
	}

	if c.onPlayerReady != nil {
		c.onPlayerReady(&msg)
	}
}

// handlePlayerJoined 处理玩家加入通知
func (c *Client) handlePlayerJoined(data []byte) {
	var msg protocol.PlayerJoined
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal PlayerJoined: %v", err)
		return
	}
	// 通过回调传递（如果有）
}

// handlePlayerLeft 处理玩家离开通知
func (c *Client) handlePlayerLeft(data []byte) {
	var msg protocol.PlayerLeft
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal PlayerLeft: %v", err)
		return
	}
	// 通过回调传递（如果有）
}

// handlePlayerActed 处理玩家动作通知
func (c *Client) handlePlayerActed(data []byte) {
	var msg protocol.PlayerActed
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal PlayerActed: %v", err)
		return
	}
	// 通过回调传递（如果有）
}

// handleError 处理错误消息
func (c *Client) handleError(data []byte) {
	var msg protocol.Error
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal Error: %v", err)
		return
	}

	log.Printf("Error: %s (code: %d)", msg.Message, msg.Code)
	c.notifyError(&GameError{Message: msg.Message, Code: msg.Code})
}

// sendJoin 发送加入游戏请求
func (c *Client) sendJoin() {
	req := &protocol.JoinRequest{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeJoin),
		PlayerName:  c.playerName,
		Seat:        -1, // 随机座位
	}
	c.Send(req)
}

// notifyError 通知错误
func (c *Client) notifyError(err error) {
	if c.onError != nil {
		c.onError(err)
	}
}

// GameError 游戏错误
type GameError struct {
	Message string
	Code    int
}

func (e *GameError) Error() string {
	return e.Message
}
