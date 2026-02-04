package host

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/protocol"
	"github.com/wilenwang/just_play/Texas-Holdem/pkg/game"
)

// WebSocket upgrader 配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有跨域请求（开发环境）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client 连接一个客户端
type Client struct {
	ID       string          // 客户端唯一标识
	Conn     *websocket.Conn // WebSocket 连接
	GameID   string         // 所属游戏ID
	Send     chan []byte    // 发送消息通道
	IsHost   bool          // 是否为庄家（HOST）
	Seat     int           // 座位号
	Name     string        // 玩家名称
	JoinedAt time.Time     // 加入时间
	mu       sync.Mutex    // 连接锁
}

// Server WebSocket 服务器
type Server struct {
	gameEngine   *game.GameEngine        // 游戏引擎实例
	gameID       string                  // 游戏ID
	upgrader     websocket.Upgrader      // WebSocket 升级器
	clients      map[string]*Client      // 所有客户端
	clientsMu    sync.RWMutex           // 客户端管理锁
	register     chan *Client            // 客户端注册通道
	unregister   chan *Client           // 客户端注销通道
	broadcast    chan []byte            // 广播消息通道
	handleMsg    chan *ClientMessage    // 消息处理通道
	gameStarted  bool                   // 游戏是否已开始
	gameEngineMu sync.RWMutex          // 游戏引擎锁
}

// ClientMessage 客户端消息
type ClientMessage struct {
	Client *Client
	Data   []byte
}

// NewServer 创建新的游戏服务器
func NewServer(config *game.Config) *Server {
	s := &Server{
		gameEngine:  game.NewEngine(config),
		gameID:      "",
		upgrader:    upgrader,
		clients:     make(map[string]*Client),
		register:    make(chan *Client, 10),
		unregister:  make(chan *Client, 10),
		broadcast:   make(chan []byte, 100),
		handleMsg:   make(chan *ClientMessage, 100),
		gameStarted: false,
	}

	// 设置状态变化回调
	s.gameEngine.SetOnStateChange(s.onGameStateChange)

	return s
}

// ServeHTTP 处理 WebSocket 连接请求
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取游戏ID（从 URL 路径或查询参数）
	gameID := r.URL.Query().Get("game_id")
	if gameID == "" {
		gameID = "default"
	}

	// 升级 HTTP 连接为 WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 创建客户端
	client := &Client{
		ID:       generateClientID(),
		Conn:     conn,
		GameID:   gameID,
		Send:     make(chan []byte, 256),
		JoinedAt: time.Now(),
	}

	// 注册客户端
	s.register <- client

	// 启动读写协程
	go client.writePump(s)
	go client.readPump(s)
}

// Run 服务器主循环
func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.handleRegister(client)

		case client := <-s.unregister:
			s.handleUnregister(client)

		case msg := <-s.handleMsg:
			s.handleMessage(msg)

		case data := <-s.broadcast:
			s.broadcastMessage(data)
		}
	}
}

// handleRegister 处理客户端注册
func (s *Server) handleRegister(client *Client) {
	s.clientsMu.Lock()
	s.clients[client.ID] = client
	s.clientsMu.Unlock()

	log.Printf("Client registered: %s (Total: %d)", client.ID, len(s.clients))

	// 发送欢迎消息
	welcome := &protocol.JoinAck{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeJoinAck),
		Success:     true,
		PlayerID:    client.ID,
		Message:     "Welcome to Texas Hold'em Poker!",
		GameState:   s.getGameStateInfo(""),
	}

	s.sendToClient(client.ID, welcome)
}

// handleUnregister 处理客户端注销
func (s *Server) handleUnregister(client *Client) {
	s.clientsMu.Lock()
	if _, ok := s.clients[client.ID]; ok {
		delete(s.clients, client.ID)
		close(client.Send)
	}
	s.clientsMu.Unlock()

	log.Printf("Client unregistered: %s (Total: %d)", client.ID, len(s.clients))

	// 通知其他玩家该玩家离开
	s.broadcastPlayerLeft(client)
}

// handleMessage 处理客户端消息
func (s *Server) handleMessage(msg *ClientMessage) {
	client := msg.Client
	var baseMsg protocol.BaseMessage

	if err := json.Unmarshal(msg.Data, &baseMsg); err != nil {
		s.sendError(client.ID, "Invalid message format", 1001)
		return
	}

	switch baseMsg.Type {
	case protocol.MsgTypeJoin:
		s.handleJoin(client, msg.Data)

	case protocol.MsgTypeLeave:
		s.handleLeave(client)

	case protocol.MsgTypePlayerAction:
		s.handlePlayerAction(client, msg.Data)

	case protocol.MsgTypeChat:
		s.handleChat(client, msg.Data)

	case protocol.MsgTypePing:
		s.handlePing(client)

	default:
		s.sendError(client.ID, "Unknown message type", 1002)
	}
}

// broadcastMessage 广播消息给所有客户端
func (s *Server) broadcastMessage(data []byte) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		select {
		case client.Send <- data:
		default:
			// 发送队列满，跳过
			log.Printf("Client %s send queue full, skipping", client.ID)
		}
	}
}

// broadcastToOthers 广播消息给除指定客户端外的所有客户端
func (s *Server) broadcastToOthers(excludeID string, data []byte) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		if client.ID == excludeID {
			continue
		}
		select {
		case client.Send <- data:
		default:
		}
	}
}

// sendToClient 发送消息给指定客户端
func (s *Server) sendToClient(clientID string, msg interface{}) {
	s.clientsMu.RLock()
	client, ok := s.clients[clientID]
	s.clientsMu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case client.Send <- data:
	default:
		log.Printf("Client %s send queue full", clientID)
	}
}

// sendError 发送错误消息
func (s *Server) sendError(clientID string, message string, code int) {
	errMsg := &protocol.Error{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeError),
		Code:        code,
		Message:     message,
	}
	s.sendToClient(clientID, errMsg)
}

// broadcastPlayerLeft 广播玩家离开消息
func (s *Server) broadcastPlayerLeft(client *Client) {
	msg := &protocol.PlayerLeft{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypePlayerLeft),
		PlayerID:    client.ID,
		PlayerName:  client.Name,
	}

	data, _ := json.Marshal(msg)
	s.broadcastToOthers(client.ID, data)
}

// onGameStateChange 游戏状态变化回调
func (s *Server) onGameStateChange(state *game.GameState) {
	// 转换为客户端可用的游戏状态
	stateInfo := s.getGameStateInfo("")

	// 广播游戏状态更新
	stateMsg := &protocol.GameState{
		GameID:         stateInfo.GameID,
		Stage:          stateInfo.Stage,
		DealerButton:  stateInfo.DealerButton,
		CurrentPlayer: stateInfo.CurrentPlayer,
		CurrentBet:    stateInfo.CurrentBet,
		Pot:           stateInfo.Pot,
		CommunityCards: stateInfo.CommunityCards,
		Players:       stateInfo.Players,
		MinRaise:      stateInfo.MinRaise,
		MaxRaise:      stateInfo.MaxRaise,
	}

	data, _ := json.Marshal(stateMsg)
	s.broadcast <- data
}

// getGameStateInfo 获取游戏状态信息
func (s *Server) getGameStateInfo(requestorID string) *protocol.GameState {
	s.gameEngineMu.RLock()
	defer s.gameEngineMu.RUnlock()

	state := s.gameEngine.GetState()

	players := make([]protocol.PlayerInfo, 0, len(state.Players))
	for _, p := range state.Players {
		playerInfo := protocol.PlayerInfo{
			ID:         p.ID,
			Name:       p.Name,
			Seat:       p.Seat,
			Chips:      p.Chips,
			CurrentBet: p.CurrentBet,
			Status:     p.Status,
			IsDealer:   p.IsDealer,
			IsSelf:     p.ID == requestorID,
		}

		// 如果是玩家自己，显示底牌
		if p.ID == requestorID {
			playerInfo.HoleCards = p.HoleCards
		}

		players = append(players, playerInfo)
	}

	return &protocol.GameState{
		GameID:         state.ID,
		Stage:          state.Stage,
		DealerButton:  state.DealerButton,
		CurrentPlayer: state.CurrentPlayer,
		CurrentBet:    state.CurrentBet,
		Pot:           state.Pot,
		CommunityCards: state.CommunityCards,
		Players:       players,
		MinRaise:      state.CurrentBet * 2,
		MaxRaise:      state.CurrentBet + s.getPlayerChips(requestorID),
	}
}

// getPlayerChips 获取玩家筹码
func (s *Server) getPlayerChips(playerID string) int {
	state := s.gameEngine.GetState()
	for _, p := range state.Players {
		if p.ID == playerID {
			return p.Chips
		}
	}
	return 0
}

// writePump 处理向客户端写入消息
func (c *Client) writePump(s *Server) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		s.unregister <- c
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// 服务器关闭了发送通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump 处理从客户端读取消息
func (c *Client) readPump(s *Server) {
	defer func() {
		s.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// 发送到消息处理队列
		s.handleMsg <- &ClientMessage{
			Client: c,
			Data:   message,
		}
	}
}

// generateClientID 生成客户端唯一标识
func generateClientID() string {
	return randomID(8)
}

// randomID 生成随机ID
func randomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
