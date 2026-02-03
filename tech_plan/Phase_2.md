# Phase 2: WebSocket 通信层实现

## 阶段目标

实现HOST和CLIENT之间的WebSocket通信，使玩家可以连接到游戏主机进行在线对战。

## 目录结构

```
server/
├── host/
│   ├── server.go           # WebSocket服务器
│   ├── handler.go          # 消息处理
│   └── game_manager.go     # 游戏管理器
└── client/
    ├── client.go           # WebSocket客户端
    └── connector.go        # 连接管理

ui/
├── host/
│   ├── model.go            # 主模型
│   └── views/              # 视图组件
└── client/
    ├── model.go            # 主模型
    ├── screens/            # 各屏幕
    └── views/              # 视图组件
```

## 详细任务

### 任务 2.1: WebSocket 服务器

**目标**: 实现基于WebSocket的游戏服务器

```go
// server/host/server.go

package host

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "yourproject/game"
    "yourproject/protocol"
)

type GameServer struct {
    engine       *game.GameEngine
    upgrader    websocket.Upgrader
    clients     map[string]*Client
    clientsMux  sync.RWMutex
    broadcast   chan []byte

    // 配置
    addr        string
    pingInterval time.Duration
}

type Client struct {
    conn       *websocket.Conn
    playerID   string
    seat       int
    send       chan []byte
    recv       chan []byte
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境应该检查Origin
    },
}

func NewGameServer(engine *game.GameEngine, addr string) *GameServer {
    return &GameServer{
        engine:       engine,
        clients:      make(map[string]*Client),
        broadcast:    make(chan []byte, 256),
        addr:         addr,
        pingInterval: time.Second * 30,
    }
}

func (s *GameServer) Run() error {
    http.HandleFunc("/ws", s.handleWebSocket)
    http.HandleFunc("/game/state", s.handleGetState)
    http.HandleFunc("/game/stats", s.handleGetStats)
    http.HandleFunc("/", s.handleIndex)

    log.Printf("游戏服务器启动: ws://%s/ws", s.addr)
    return http.ListenAndServe(s.addr, nil)
}

func (s *GameServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket升级失败: %v", err)
        return
    }

    client := &Client{
        conn: conn,
        send: make(chan []byte, 256),
        recv: make(chan []byte, 256),
    }

    // 读取加入消息
    _, message, err := conn.ReadMessage()
    if err != nil {
        log.Printf("读取加入消息失败: %v", err)
        conn.Close()
        return
    }

    var msg protocol.Message
    if err := json.Unmarshal(message, &msg); err != nil {
        log.Printf("解析加入消息失败: %v", err)
        conn.Close()
        return
    }

    if msg.Type != protocol.MsgJoin {
        s.sendError(conn, &protocol.GameError{
            Code:    "invalid_message",
            Message: "第一条消息必须是加入游戏",
        })
        conn.Close()
        return
    }

    // 处理加入
    payload := msg.Payload.(map[string]interface{})
    playerName := payload["player_name"].(string)

    // 获取可用座位
    seat := -1
    if seatArg, ok := payload["seat"].(float64); ok {
        seat = int(seatArg)
    }

    // 添加玩家
    playerID := fmt.Sprintf("player_%d", time.Now().UnixNano())
    player, err := s.engine.AddPlayer(playerID, playerName, seat)
    if err != nil {
        s.sendError(conn, &protocol.GameError{
            Code:    "join_failed",
            Message: err.Error(),
        })
        conn.Close()
        return
    }

    client.playerID = playerID
    client.seat = player.Seat

    // 注册客户端
    s.clientsMux.Lock()
    s.clients[playerID] = client
    s.clientsMux.Unlock()

    log.Printf("玩家加入: %s (座位: %d)", playerName, player.Seat)

    // 发送加入确认
    joinAck := protocol.Message{
        Type: protocol.MsgJoinAck,
        Payload: protocol.JoinAckPayload{
            PlayerID:  playerID,
            Seat:      player.Seat,
            PlayerName: playerName,
            GameState: s.convertToClientState(nil),
        },
    }
    s.sendToClient(client, joinAck)

    // 启动读写协程
    go s.writePump(client)
    s.readPump(client, playerID)
}

func (s *GameServer) readPump(client *Client, playerID string) {
    defer func() {
        s.disconnectClient(playerID)
    }()

    client.conn.SetReadLimit(512 * 1024) // 512KB
    client.conn.SetReadDeadline(time.Now().Add(time.Hour))
    client.conn.SetPongHandler(func(string) error {
        client.conn.SetReadDeadline(time.Now().Add(time.Hour))
        return nil
    })

    for {
        _, message, err := client.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err,
                websocket.CloseGoingAway,
                websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket错误: %v", err)
            }
            break
        }

        var msg protocol.Message
        if err := json.Unmarshal(message, &msg); err != nil {
            log.Printf("解析消息失败: %v", err)
            continue
        }

        s.handleMessage(playerID, &msg)
    }
}

func (s *GameServer) writePump(client *Client) {
    ticker := time.NewTicker(s.pingInterval)
    defer func() {
        ticker.Stop()
        client.conn.Close()
    }()

    for {
        select {
        case message, ok := <-client.send:
            client.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
            if !ok {
                client.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := client.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)
            w.Close()

        case <-ticker.C:
            client.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
            if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func (s *GameServer) handleMessage(playerID string, msg *protocol.Message) {
    switch msg.Type {
    case protocol.MsgAction:
        payload := msg.Payload.(map[string]interface{})
        action := protocol.ActionType(int(payload["action"].(float64)))
        amount := 0
        if a, ok := payload["amount"].(float64); ok {
            amount = int(a)
        }

        if err := s.engine.PlayerAction(playerID, action, amount); err != nil {
            s.sendErrorToPlayer(playerID, &protocol.GameError{
                Code:    "action_failed",
                Message: err.Error(),
            })
        }

    case protocol.MsgChat:
        payload := msg.Payload.(map[string]interface{})
        message := payload["message"].(string)
        s.broadcastChat(playerID, message)

    case protocol.MsgStandUp:
        s.engine.RemovePlayer(playerID)

    case protocol.MsgSitDown:
        payload := msg.Payload.(map[string]interface{})
        if seat, ok := payload["seat"].(float64); ok {
            s.engine.RemovePlayer(playerID)
            s.engine.AddPlayer(playerID, "", int(seat))
        }
    }
}

func (s *GameServer) broadcastGameState(state *game.GameState) {
    payload := s.convertToClientState(state)

    s.clientsMux.RLock()
    defer s.clientsMux.RUnlock()

    msg := protocol.Message{
        Type:    protocol.MsgGameState,
        Payload: payload,
    }

    data, _ := json.Marshal(msg)
    for _, client := range s.clients {
        select {
        case client.send <- data:
        default:
            // 队列满,跳过
        }
    }
}

func (s *GameServer) broadcastYourTurn(playerID string, payload *protocol.YourTurnPayload) {
    s.clientsMux.RLock()
    client, ok := s.clients[playerID]
    s.clientsMux.RUnlock()

    if !ok {
        return
    }

    msg := protocol.Message{
        Type:    protocol.MsgYourTurn,
        Payload: payload,
    }

    data, _ := json.Marshal(msg)
    select {
    case client.send <- data:
    default:
    }
}

func (s *GameServer) broadcastShowdown(payload *protocol.ShowdownPayload) {
    s.clientsMux.RLock()
    defer s.clientsMux.RUnlock()

    msg := protocol.Message{
        Type:    protocol.MsgShowdown,
        Payload: payload,
    }

    data, _ := json.Marshal(msg)
    for _, client := range s.clients {
        select {
        case client.send <- data:
        default:
        }
    }
}

func (s *GameServer) broadcastChat(senderID, message string) {
    s.clientsMux.RLock()
    defer s.clientsMux.RUnlock()

    sender := s.clients[senderID]

    msg := protocol.Message{
        Type: protocol.MsgChat,
        Payload: protocol.ChatPayload{
            PlayerID:  senderID,
            PlayerName: s.engine.GetPlayerName(senderID),
            Message:   message,
            Timestamp: time.Now().Unix(),
        },
    }

    data, _ := json.Marshal(msg)
    for _, client := range s.clients {
        select {
        case client.send <- data:
        default:
        }
    }
}

func (s *GameServer) convertToClientState(state *game.GameState) *protocol.GameStatePayload {
    if state == nil {
        state = s.engine.GetState()
    }

    players := make([]protocol.PlayerInfo, len(state.Players))
    for i, p := range state.Players {
        players[i] = protocol.PlayerInfo{
            ID:         p.ID,
            Name:       p.Name,
            Chips:      p.Chips,
            Seat:       p.Seat,
            Status:     protocol.PlayerStatus(p.Status),
            CurrentBet: p.CurrentBet,
            IsDealer:   p.IsDealer,
            HasActed:   p.HasActed,
        }

        // 只发送给玩家自己的底牌
        players[i].HoleCards = [2]card.Card{}
    }

    community := make([]card.Card, 0, 5)
    for _, c := range state.CommunityCards {
        if c.Rank != 0 {
            community = append(community, c)
        }
    }

    return &protocol.GameStatePayload{
        GameID:        state.ID,
        Stage:         protocol.GameStage(state.Stage),
        DealerButton:  state.DealerButton,
        CurrentPlayer: state.CurrentPlayer,
        CurrentBet:    state.CurrentBet,
        Pot:           state.Pot,
        CommunityCards: community,
        Players:       players,
    }
}

func (s *GameServer) sendToClient(client *Client, msg interface{}) {
    data, _ := json.Marshal(msg)
    select {
    case client.send <- data:
    default:
    }
}

func (s *GameServer) sendError(conn *websocket.Conn, err *protocol.GameError) {
    msg := protocol.Message{
        Type: protocol.MsgError,
        Payload: protocol.ErrorPayload{
            Code:    err.Code,
            Message: err.Message,
        },
    }
    conn.WriteJSON(msg)
}

func (s *GameServer) sendErrorToPlayer(playerID string, err *protocol.GameError) {
    s.clientsMux.RLock()
    client, ok := s.clients[playerID]
    s.clientsMux.RUnlock()

    if !ok {
        return
    }

    s.sendError(client.conn, err)
}

func (s *GameServer) disconnectClient(playerID string) {
    s.clientsMux.Lock()
    client, ok := s.clients[playerID]
    if ok {
        delete(s.clients, playerID)
        close(client.send)
    }
    s.clientsMux.Unlock()

    s.engine.RemovePlayer(playerID)
    log.Printf("玩家断开连接: %s", playerID)
}

func (s *GameServer) handleGetState(w http.ResponseWriter, r *http.Request) {
    state := s.engine.GetState()
    json.NewEncoder(w).Encode(s.convertToClientState(state))
}

func (s *GameServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
    stats := map[string]interface{}{
        "player_count":  len(s.clients),
        "game_running":  s.engine.IsHandInProgress(),
    }
    json.NewEncoder(w).Encode(stats)
}

func (s *GameServer) handleIndex(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`
        <html>
        <head><title>Texas Hold'em Server</title></head>
        <body>
            <h1>Texas Hold'em Game Server</h1>
            <p>WebSocket端点: /ws</p>
        </body>
        </html>
    `))
}
```

### 任务 2.2: WebSocket 客户端

**目标**: 实现客户端连接和消息处理

```go
// server/client/client.go

package client

import (
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "yourproject/protocol"
)

type Client struct {
    conn      *websocket.Conn
    playerID  string
    playerName string

    // 消息通道
    incoming  chan *protocol.Message
    outgoing chan *protocol.Message

    // 状态
    connected bool
    reconnect bool
    reconnectAttempts int
    maxReconnectAttempts int

    // 回调
    onStateChange func(state *protocol.GameStatePayload)
    onYourTurn   func(info *protocol.YourTurnPayload)
    onShowdown   func(result *protocol.ShowdownPayload)
    onError      func(err error)
    onChat       func(msg *protocol.ChatPayload)
    onDisconnect  func()

    // 同步
    mu sync.Mutex
}

type Config struct {
    URL                    string
    PlayerName             string
    Reconnect              bool
    MaxReconnectAttempts   int
    ReconnectDelay         time.Duration
}

func NewClient(config *Config) *Client {
    return &Client{
        playerID:             "",
        playerName:           config.PlayerName,
        incoming:             make(chan *protocol.Message, 256),
        outgoing:             make(chan *protocol.Message, 256),
        connected:            false,
        reconnect:            config.Reconnect,
        maxReconnectAttempts: config.MaxReconnectAttempts,
    }
}

func (c *Client) Connect() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    conn, _, err := websocket.DefaultDialer.Dial(c.getURL(), nil)
    if err != nil {
        return fmt.Errorf("连接失败: %v", err)
    }

    c.conn = conn
    c.connected = true

    // 发送加入消息
    joinMsg := &protocol.Message{
        Type: protocol.MsgJoin,
        Payload: protocol.JoinPayload{
            PlayerName: c.playerName,
            BuyIn:      1000,
        },
    }

    if err := conn.WriteJSON(joinMsg); err != nil {
        c.disconnect()
        return fmt.Errorf("发送加入消息失败: %v", err)
    }

    // 启动读写协程
    go c.readLoop()
    go c.writeLoop()

    return nil
}

func (c *Client) getURL() string {
    // 支持环境变量配置
    // WS_URL=localhost:8080/ws ./client
    if url := os.Getenv("WS_URL"); url != "" {
        return url
    }
    return "ws://localhost:8080/ws"
}

func (c *Client) readLoop() {
    defer func() {
        c.disconnect()
    }()

    c.conn.SetReadLimit(512 * 1024)

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err,
                websocket.CloseGoingAway,
                websocket.CloseAbnormalClosure) {
                log.Printf("读取消息错误: %v", err)
            }
            break
        }

        var msg protocol.Message
        if err := json.Unmarshal(message, &msg); err != nil {
            log.Printf("解析消息失败: %v", err)
            continue
        }

        c.handleMessage(&msg)
    }
}

func (c *Client) writeLoop() {
    for {
        msg, ok := <-c.outgoing
        if !ok {
            c.conn.WriteMessage(websocket.CloseMessage, []byte{})
            return
        }

        data, _ := json.Marshal(msg)
        if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
            return
        }
    }
}

func (c *Client) handleMessage(msg *protocol.Message) {
    switch msg.Type {
    case protocol.MsgJoinAck:
        payload := msg.Payload.(map[string]interface{})
        c.playerID = payload["player_id"].(string)
        log.Printf("加入成功,玩家ID: %s", c.playerID)

    case protocol.MsgGameState:
        payload := msg.Payload.(*protocol.GameStatePayload)
        c.incoming <- msg
        if c.onStateChange != nil {
            c.onStateChange(payload)
        }

    case protocol.MsgYourTurn:
        payload := msg.Payload.(*protocol.YourTurnPayload)
        c.incoming <- msg
        if c.onYourTurn != nil {
            c.onYourTurn(payload)
        }

    case protocol.MsgShowdown:
        payload := msg.Payload.(*protocol.ShowdownPayload)
        c.incoming <- msg
        if c.onShowdown != nil {
            c.onShowdown(payload)
        }

    case protocol.MsgChat:
        payload := msg.Payload.(*protocol.ChatPayload)
        if c.onChat != nil {
            c.onChat(payload)
        }

    case protocol.MsgError:
        payload := msg.Payload.(map[string]interface{})
        err := fmt.Errorf("%s: %s", payload["code"], payload["message"])
        if c.onError != nil {
            c.onError(err)
        }
    }
}

// 发送行动
func (c *Client) Action(action protocol.ActionType, amount int) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if !c.connected {
        return fmt.Errorf("未连接")
    }

    msg := &protocol.Message{
        Type: protocol.MsgAction,
        Payload: protocol.ActionPayload{
            Action: action,
            Amount: amount,
        },
    }

    select {
    case c.outgoing <- msg:
        return nil
    default:
        return fmt.Errorf("发送队列满")
    }
}

// 发送聊天
func (c *Client) Chat(message string) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if !c.connected {
        return fmt.Errorf("未连接")
    }

    msg := &protocol.Message{
        Type: protocol.MsgChat,
        Payload: map[string]string{
            "message": message,
        },
    }

    select {
    case c.outgoing <- msg:
        return nil
    default:
        return fmt.Errorf("发送队列满")
    }
}

// 发送聊天
func (c *Client) SendChat(message string) error {
    msg := &protocol.Message{
        Type: protocol.MsgChat,
        Payload: protocol.ChatPayload{
            Message: message,
        },
    }

    select {
    case c.outgoing <- msg:
        return nil
    default:
        return fmt.Errorf("发送队列满")
    }
}

// 断线重连
func (c *Client) reconnectLoop() {
    delay := time.Second

    for c.reconnect && c.reconnectAttempts < c.maxReconnectAttempts {
        log.Printf("尝试重连 (%d/%d)...", c.reconnectAttempts+1, c.maxReconnectAttempts)

        if err := c.Connect(); err != nil {
            log.Printf("重连失败: %v", err)
            time.Sleep(delay)
            delay *= 2
            if delay > time.Minute {
                delay = time.Minute
            }
            c.reconnectAttempts++
            continue
        }

        log.Println("重连成功")
        c.reconnectAttempts = 0
        return
    }

    if c.onDisconnect != nil {
        c.onDisconnect()
    }
}

func (c *Client) disconnect() {
    c.mu.Lock()
    if c.connected {
        c.connected = false
        c.conn.Close()
    }
    c.mu.Unlock()

    if c.reconnect && c.reconnectAttempts < c.maxReconnectAttempts {
        go c.reconnectLoop()
    } else if c.onDisconnect != nil {
        c.onDisconnect()
    }
}

// Setter callbacks
func (c *Client) OnStateChange(fn func(*protocol.GameStatePayload)) {
    c.onStateChange = fn
}

func (c *Client) OnYourTurn(fn func(*protocol.YourTurnPayload)) {
    c.onYourTurn = fn
}

func (c *Client) OnShowdown(fn func(*protocol.ShowdownPayload)) {
    c.onShowdown = fn
}

func (c *Client) OnError(fn func(error)) {
    c.onError = fn
}

func (c *Client) OnChat(fn func(*protocol.ChatPayload)) {
    c.onChat = fn
}

func (c *Client) OnDisconnect(fn func()) {
    c.onDisconnect = fn
}

// Getters
func (c *Client) IsConnected() bool {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.connected
}

func (c *Client) GetPlayerID() string {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.playerID
}
```

### 任务 2.3: CLIENT TUI 主模型

```go
// ui/client/model.go

package client

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"
    "yourproject/server/client"
    "yourproject/protocol"
)

type Model struct {
    client   *client.Client
    screen   ScreenType
    err      error

    // 连接配置
    serverURL string
    playerName string

    // 子模型
    login      *LoginScreen
    lobby      *LobbyScreen
    game       *GameScreen

    // 通用
    viewport   viewport.Model
}

type ScreenType int

const (
    ScreenConnecting ScreenType = iota
    ScreenLogin
    ScreenLobby
    ScreenGame
)

func NewModel() *Model {
    return &Model{
        screen: ScreenLogin,
        login:  NewLoginScreen(),
    }
}

func (m *Model) Init() tea.Cmd {
    return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            if m.client != nil && m.client.IsConnected() {
                m.client.Close()
            }
            return m, tea.Quit
        }

    case tea.WindowSizeMsg:
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height
        return m, nil

    case *protocol.Message:
        return m.handleServerMessage(msg)
    }

    // 子模型更新
    switch m.screen {
    case ScreenLogin:
        newLogin, cmd := m.login.Update(msg)
        m.login = newLogin.(*LoginScreen)
        if m.login.Done {
            return m.connect(m.login.GetServerURL(), m.login.GetPlayerName())
        }
        return m, cmd

    case ScreenGame:
        if m.game != nil {
            newGame, cmd := m.game.Update(msg)
            m.game = newGame.(*GameScreen)
            return m, cmd
        }
    }

    return m, nil
}

func (m *Model) handleServerMessage(msg *protocol.Message) (tea.Model, tea.Cmd) {
    switch msg.Type {
    case protocol.MsgJoinAck:
        m.screen = ScreenLobby
        m.lobby = NewLobbyScreen()
        return m, nil

    case protocol.MsgGameState:
        state := msg.Payload.(*protocol.GameStatePayload)
        if m.game == nil {
            m.game = NewGameScreen()
        }
        m.game.UpdateState(state)
        m.screen = ScreenGame
        return m, nil

    case protocol.MsgYourTurn:
        payload := msg.Payload.(*protocol.YourTurnPayload)
        m.game.ShowActionPrompt(payload)
        return m, nil

    case protocol.MsgShowdown:
        payload := msg.Payload.(*protocol.ShowdownPayload)
        m.game.ShowResult(payload)
        return m, nil

    case protocol.MsgChat:
        payload := msg.Payload.(*protocol.ChatPayload)
        m.game.AddChatMessage(payload)
        return m, nil

    case protocol.MsgError:
        payload := msg.Payload.(map[string]interface{})
        m.err = fmt.Errorf("%s: %s", payload["code"], payload["message"])
    }

    return m, nil
}

func (m *Model) connect(url, name string) (tea.Model, tea.Cmd) {
    m.serverURL = url
    m.playerName = name
    m.screen = ScreenConnecting

    c := client.NewClient(&client.Config{
        URL:                   url,
        PlayerName:           name,
        Reconnect:            true,
        MaxReconnectAttempts: 5,
    })

    c.OnStateChange(func(state *protocol.GameStatePayload) {
        if m.game == nil {
            m.game = NewGameScreen()
        }
        m.game.UpdateState(state)
        m.screen = ScreenGame
    })

    c.OnYourTurn(func(info *protocol.YourTurnPayload) {
        m.game.ShowActionPrompt(info)
    })

    c.OnError(func(err error) {
        m.err = err
    })

    m.client = c

    if err := c.Connect(); err != nil {
        m.err = err
        m.screen = ScreenLogin
        return m, nil
    }

    return m, tea.Batch(
        func() tea.Msg { return &protocol.Message{Type: protocol.MsgJoinAck} },
        textinput.Blink,
    )
}

func (m *Model) View() string {
    switch m.screen {
    case ScreenConnecting:
        return "正在连接到游戏服务器...\n\n" + m.serverURL

    case ScreenLogin:
        return m.login.View()

    case ScreenLobby:
        return m.lobby.View()

    case ScreenGame:
        if m.game != nil {
            return m.game.View()
        }
        return "等待游戏状态..."

    default:
        return "未知状态"
    }
}
```

### 任务 2.4: 游戏界面组件

```go
// ui/client/screens/game.go

package screens

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/viewport"
    "yourproject/protocol"
)

type GameScreen struct {
    state   *protocol.GameStatePayload
    mySeat  int

    // 行动提示
    actionPrompt *ActionPrompt
    canAct      bool

    // 聊天
    chatMessages []string
    chatInput   textinput.Model
    showChat    bool

    // 帮助
    help help.Model
    keys keyMap

    // 视图
    viewport viewport.Model
}

type keyMap struct {
    fold   key.Binding
    check  key.Binding
    call   key.Binding
    raise  key.Binding
    allin  key.Binding
    chat   key.Binding
    help   key.Binding
    quit   key.Binding
}

func newKeyMap() keyMap {
    return keyMap{
        fold: key.NewBinding(
            key.WithKeys("f", "F"),
            key.WithHelp("f", "弃牌"),
        ),
        check: key.NewBinding(
            key.WithKeys("x", "X", "c", "C"),
            key.WithHelp("x", "看牌"),
        ),
        call: key.NewBinding(
            key.WithKeys("l", "L", "enter"),
            key.WithHelp("l", "跟注"),
        ),
        raise: key.NewBinding(
            key.WithKeys("r", "R"),
            key.WithHelp("r", "加注"),
        ),
        allin: key.NewBinding(
            key.WithKeys("a", "A"),
            key.WithHelp("a", "全下"),
        ),
        chat: key.NewBinding(
            key.WithKeys("t", "T"),
            key.WithHelp("t", "聊天"),
        ),
        help: key.NewBinding(
            key.WithKeys("?", "h", "H"),
            key.WithHelp("?", "帮助"),
        ),
        quit: key.NewBinding(
            key.WithKeys("q", "ctrl+c"),
            key.WithHelp("q", "退出"),
        ),
    }
}

func NewGameScreen() *GameScreen {
    ti := textinput.New()
    ti.Placeholder = "输入消息..."
    ti.Focus()

    h := help.New()
    h.ShowAll = true

    return &GameScreen{
        chatMessages: make([]string, 0),
        chatInput:    ti,
        help:         h,
        keys:         newKeyMap(),
        viewport:     viewport.New(80, 20),
    }
}

func (g *GameScreen) UpdateState(state *protocol.GameStatePayload) {
    g.state = state

    // 找到自己的座位
    for _, p := range state.Players {
        if p.ID != "" {
            g.mySeat = p.Seat
            break
        }
    }
}

func (g *GameScreen) ShowActionPrompt(info *protocol.YourTurnPayload) {
    g.canAct = true
    g.actionPrompt = &ActionPrompt{
        MinBet:     info.MinBet,
        CallAmount: info.CallAmount,
        MinRaise:   info.MinRaise,
        MaxRaise:   info.MaxRaise,
        Timeout:    info.ActionTimeout,
        Allowed:    info.AllowedActions,
    }
}

func (g *GameScreen) ShowResult(result *protocol.ShowdownPayload) {
    g.canAct = false
    g.actionPrompt = nil
}

func (g *GameScreen) AddChatMessage(msg *protocol.ChatPayload) {
    timestamp := time.Now().Format("15:04")
    line := fmt.Sprintf("[%s] %s: %s", timestamp, msg.PlayerName, msg.Message)
    g.chatMessages = append(g.chatMessages, line)
    if len(g.chatMessages) > 100 {
        g.chatMessages = g.chatMessages[len(g.chatMessages)-100:]
    }
}

func (g *GameScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, g.keys.fold):
            // TODO: 发送弃牌

        case key.Matches(msg, g.keys.check):
            // TODO: 发送看牌

        case key.Matches(msg, g.keys.call):
            // TODO: 发送跟注

        case key.Matches(msg, g.keys.raise):
            // TODO: 进入加注输入模式

        case key.Matches(msg, g.keys.allin):
            // TODO: 发送全下

        case key.Matches(msg, g.keys.chat):
            g.showChat = !g.showChat
            if g.showChat {
                return g, textinput.Blink
            }

        case key.Matches(msg, g.keys.help):
            // 显示帮助

        case key.Matches(msg, g.keys.quit):
            return g, tea.Quit
        }
    }

    if g.showChat {
        var cmd tea.Cmd
        g.chatInput, cmd = g.chatInput.Update(msg)
        if g.chatInput.Value() != "" && g.chatInput.Value() != g.chatInput.Placeholder {
            // 发送聊天消息
            g.chatInput.Reset()
        }
        return g, cmd
    }

    return g, nil
}

func (g *GameScreen) View() string {
    if g.state == nil {
        return "等待游戏状态..."
    }

    var b strings.Builder

    // 标题
    b.WriteString(g.renderHeader())

    // 公共牌
    b.WriteString(g.renderCommunityCards())

    // 底池
    b.WriteString(g.renderPot())

    // 玩家
    b.WriteString(g.renderPlayers())

    // 行动提示
    if g.canAct && g.actionPrompt != nil {
        b.WriteString(g.actionPrompt.View())
    }

    // 聊天
    if g.showChat {
        b.WriteString(g.renderChat())
    }

    // 帮助
    b.WriteString("\n" + g.help.View(g.keys))

    return b.String()
}

func (g *GameScreen) renderHeader() string {
    stageNames := []string{
        "等待开始", "翻牌前", "翻牌圈", "转牌圈", "河牌圈", "摊牌",
    }
    stage := stageNames[g.state.Stage]
    dealer := fmt.Sprintf("庄家: [P%d]", g.state.DealerButton+1)

    return fmt.Sprintf("\n=== 阶段: %s | %s ===\n\n", stage, dealer)
}

func (g *GameScreen) renderCommunityCards() string {
    s := "公共牌: "
    for _, c := range g.state.CommunityCards {
        s += fmt.Sprintf("[%s] ", cardToString(c))
    }
    s += "\n\n"
    return s
}

func (g *GameScreen) renderPot() string {
    return fmt.Sprintf("底池: %d\n\n", g.state.Pot)
}

func (g *GameScreen) renderPlayers() string {
    s := "玩家:\n"
    for i, p := range g.state.Players {
        status := g.getPlayerStatus(p)
        isMe := p.Seat == g.mySeat

        cards := ""
        if len(p.HoleCards) == 2 && p.HoleCards[0].Rank != 0 {
            cards = fmt.Sprintf(" 底牌: %s %s",
                cardToString(p.HoleCards[0]),
                cardToString(p.HoleCards[1]))
        }

        prefix := "  "
        if isMe {
            prefix = "*"
        }
        if i == g.state.CurrentPlayer {
            prefix = "▶"
        }

        s += fmt.Sprintf("  %s[P%d] %-10s 筹码: %-5d 下注: %-5d %s%s\n",
            prefix, p.Seat+1, p.Name, p.Chips, p.CurrentBet, status, cards)
    }
    return s + "\n"
}

func (g *GameScreen) getPlayerStatus(p *protocol.PlayerInfo) string {
    switch p.Status {
    case protocol.PlayerStatusActive:
        if p.HasActed {
            return "已行动"
        }
        return "等待行动"
    case protocol.PlayerStatusFolded:
        return "[已弃牌]"
    case protocol.PlayerStatusAllIn:
        return "[全下]"
    }
    return ""
}

func (g *GameScreen) renderChat() string {
    var b strings.Builder
    b.WriteString("\n--- 聊天 ---\n")
    for _, msg := range g.chatMessages {
        b.WriteString(msg + "\n")
    }
    b.WriteString(g.chatInput.View())
    b.WriteString("\n(按Esc退出聊天)\n")
    return b.String()
}

func cardToString(c protocol.Card) string {
    ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
    suits := []string{"♣", "♦", "♥", "♠"}

    if c.Rank == 0 {
        return "  ?  "
    }
    return fmt.Sprintf("%s%s", ranks[c.Rank-2], suits[c.Suit])
}
```

## 通信协议

### 消息格式

```json
{
  "type": "message_type",
  "payload": {}
}
```

### 消息类型

| 方向 | 类型 | 说明 |
|-----|------|------|
| C→S | join | 加入游戏 |
| S→C | join_ack | 加入确认 |
| C→S | action | 玩家行动 |
| S→C | game_state | 游戏状态更新 |
| S→C | your_turn | 轮到行动 |
| S→C | showdown | 摊牌结果 |
| C→S/ S→C | chat | 聊天消息 |
| S→C | error | 错误 |

## 阶段验收清单

- [ ] WebSocket服务器正常启动
- [ ] 客户端可以连接服务器
- [ ] 玩家加入/离开功能正常
- [ ] 游戏状态实时同步
- [ ] 玩家行动正确发送
- [ ] 聊天功能可用
- [ ] 断线重连机制工作
