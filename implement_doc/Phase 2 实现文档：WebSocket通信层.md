# Phase 2 实现文档：WebSocket 通信层

## 概述

本文档描述了德州扑克游戏 Phase 2（WebSocket 通信层）的实现细节，包括消息协议定义、WebSocket 服务器/客户端实现以及 TUI 界面集成。

## 项目结构

```
Texas-Holdem/
├── internal/
│   ├── card/
│   │   ├── card.go          # 扑克牌基础类型
│   │   ├── deck.go          # 牌组管理
│   │   └── *_test.go        # 单元测试
│   ├── common/
│   │   └── models/
│   │       └── player.go     # 玩家数据结构
│   └── protocol/
│       ├── message.go        # 消息协议定义
│       └── message_test.go   # 协议单元测试
├── pkg/
│   ├── evaluator/
│   │   ├── evaluator.go      # 牌型评估器
│   │   └── *_test.go        # 单元测试
│   └── game/
│       └── engine.go        # 游戏引擎
├── server/
│   ├── host/
│   │   ├── server.go        # WebSocket 服务器
│   │   └── handler.go       # 消息处理器
│   └── client/
│       └── client.go         # WebSocket 客户端
├── ui/
│   ├── host/
│   │   └── model.go         # HOST TUI 界面
│   └── client/
│       └── model.go         # CLIENT TUI 界面
├── go.mod
└── implement_doc/
    ├── Phase 1 实现文档：核心游戏引擎.md
    └── Phase 2 实现文档：WebSocket通信层.md
```

---

## 1. 消息协议 (internal/protocol)

### 1.1 消息类型

**客户端 → 服务器消息类型：**

| 类型 | 说明 |
|-----|------|
| `join` | 玩家加入游戏 |
| `leave` | 玩家离开游戏 |
| `player_action` | 玩家执行动作 |
| `chat` | 发送聊天消息 |
| `ping` | 心跳检测 |

**服务器 → 客户端消息类型：**

| 类型 | 说明 |
|-----|------|
| `join_ack` | 加入游戏确认 |
| `game_state` | 游戏状态更新 |
| `your_turn` | 通知玩家回合 |
| `player_joined` | 新玩家加入通知 |
| `player_left` | 玩家离开通知 |
| `player_acted` | 玩家动作通知 |
| `showdown` | 摊牌结果 |
| `chat` | 聊天消息 |
| `pong` | 心跳响应 |
| `error` | 错误消息 |

### 1.2 基础消息结构

```go
type BaseMessage struct {
    Type      MessageType `json:"type"`       // 消息类型
    Timestamp int64       `json:"timestamp"`  // 时间戳（毫秒）
}
```

### 1.3 客户端请求消息

**JoinRequest - 玩家加入请求：**

```go
type JoinRequest struct {
    BaseMessage
    PlayerName string `json:"player_name"` // 玩家名称
    Seat       int    `json:"seat"`        // 请求座位号（-1表示随机）
}
```

**PlayerActionRequest - 玩家动作请求：**

```go
type PlayerActionRequest struct {
    BaseMessage
    PlayerID string          `json:"player_id"` // 玩家ID
    Action   models.ActionType `json:"action"`   // 动作类型
    Amount   int             `json:"amount"`    // 下注金额
}
```

**ChatRequest - 聊天消息请求：**

```go
type ChatRequest struct {
    BaseMessage
    PlayerID string `json:"player_id"` // 玩家ID
    Content  string `json:"content"`   // 消息内容
}
```

### 1.4 服务器响应消息

**JoinAck - 加入确认：**

```go
type JoinAck struct {
    BaseMessage
    Success   bool   `json:"success"`    // 是否成功
    PlayerID  string `json:"player_id"`  // 分配的玩家ID
    Seat      int    `json:"seat"`       // 座位号
    Message   string `json:"message"`    // 附加消息
    GameState *GameState `json:"game_state,omitempty"` // 当前游戏状态
}
```

**GameState - 游戏状态：**

```go
type GameState struct {
    BaseMessage
    GameID         string            `json:"game_id"`          // 游戏ID
    Stage          game.Stage        `json:"stage"`            // 当前阶段
    DealerButton  int               `json:"dealer_button"`    // 庄家按钮位置
    CurrentPlayer int               `json:"current_player"`   // 当前行动玩家索引
    CurrentBet    int               `json:"current_bet"`      // 当前最高下注
    Pot           int               `json:"pot"`              // 底池金额
    CommunityCards [5]card.Card     `json:"community_cards"`  // 公共牌
    Players       []PlayerInfo      `json:"players"`          // 所有玩家信息
    MinRaise      int               `json:"min_raise"`        // 最小加注金额
    MaxRaise      int               `json:"max_raise"`        // 最大加注金额
}
```

**PlayerInfo - 玩家信息：**

```go
type PlayerInfo struct {
    ID         string              `json:"id"`           // 玩家ID
    Name       string              `json:"name"`         // 玩家名称
    Seat       int                 `json:"seat"`         // 座位号
    Chips      int                 `json:"chips"`        // 剩余筹码
    CurrentBet int                 `json:"current_bet"`  // 当前下注金额
    Status     models.PlayerStatus `json:"status"`       // 玩家状态
    IsDealer   bool                `json:"is_dealer"`    // 是否为庄家
    HoleCards  [2]card.Card        `json:"hole_cards"`   // 底牌
    IsSelf     bool                `json:"is_self"`      // 是否是请求者自己
}
```

**YourTurn - 轮到玩家：**

```go
type YourTurn struct {
    BaseMessage
    PlayerID    string `json:"player_id"`    // 玩家ID
    MinAction   int    `json:"min_action"`  // 最小下注金额
    MaxAction   int    `json:"max_action"`  // 最大下注金额
    CurrentBet  int    `json:"current_bet"` // 当前最高下注
    TimeLeft    int    `json:"time_left"`   // 剩余时间（秒）
}
```

**Showdown - 摊牌结果：**

```go
type Showdown struct {
    BaseMessage
    Winners   []WinnerInfo `json:"winners"`    // 获胜者列表
    Pot       int          `json:"pot"`        // 底池金额
    GameState *GameState  `json:"game_state"` // 最终游戏状态
}

type WinnerInfo struct {
    PlayerID   string                 `json:"player_id"`   // 玩家ID
    PlayerName string                 `json:"player_name"` // 玩家名称
    HandRank   evaluator.HandRank    `json:"hand_rank"`   // 手牌等级
    HandName   string                `json:"hand_name"`   // 手牌名称
    WonChips   int                   `json:"won_chips"`   // 赢得筹码
    RawCards   []card.Card           `json:"raw_cards"`   // 构成最佳手牌的牌
}
```

### 1.5 辅助函数

```go
// 创建带时间戳的基本消息
func NewBaseMessage(msgType MessageType) BaseMessage

// 创建各种请求消息
func NewJoinRequest(playerName string, seat int) *JoinRequest
func NewLeaveRequest(playerID string) *LeaveRequest
func NewPlayerActionRequest(playerID string, action models.ActionType, amount int) *PlayerActionRequest
func NewChatRequest(playerID, content string) *ChatRequest
func NewPingRequest() *PingRequest
```

---

## 2. WebSocket 服务器 (server/host)

### 2.1 Server 结构

```go
type Server struct {
    gameEngine   *game.GameEngine        // 游戏引擎实例
    gameID       string                  // 游戏ID
    upgrader     websocket.Upgrader      // WebSocket 升级器
    clients      map[string]*Client      // 所有客户端
    register     chan *Client            // 客户端注册通道
    unregister   chan *Client           // 客户端注销通道
    broadcast    chan []byte            // 广播消息通道
    handleMsg    chan *ClientMessage    // 消息处理通道
    gameStarted  bool                   // 游戏是否已开始
}
```

### 2.2 Client 结构

```go
type Client struct {
    ID       string          // 客户端唯一标识
    Conn     *websocket.Conn // WebSocket 连接
    GameID   string         // 所属游戏ID
    Send     chan []byte    // 发送消息通道
    IsHost   bool          // 是否为庄家
    Seat     int           // 座位号
    Name     string        // 玩家名称
    JoinedAt time.Time     // 加入时间
}
```

### 2.3 主要方法

| 方法 | 说明 |
|-----|------|
| `NewServer(config)` | 创建新的游戏服务器 |
| `ServeHTTP(w, r)` | 处理 WebSocket 连接请求 |
| `Run()` | 服务器主循环 |
| `sendToClient(clientID, msg)` | 发送消息给指定客户端 |
| `broadcastMessage(data)` | 广播消息给所有客户端 |
| `broadcastToOthers(excludeID, data)` | 广播给除指定客户端外的所有人 |
| `sendError(clientID, message, code)` | 发送错误消息 |

### 2.4 消息处理器

| 方法 | 说明 |
|-----|------|
| `handleJoin(client, data)` | 处理玩家加入 |
| `handleLeave(client)` | 处理玩家离开 |
| `handlePlayerAction(client, data)` | 处理玩家动作 |
| `handleChat(client, data)` | 处理聊天消息 |
| `handlePing(client)` | 处理心跳检测 |
| `findAvailableSeat()` | 查找可用座位 |

### 2.5 心跳机制

- 服务器每 30 秒发送 Ping 消息
- 客户端需要在 60 秒内响应
- 超时未响应则断开连接

---

## 3. WebSocket 客户端 (server/client)

### 3.1 Client 结构

```go
type Client struct {
    serverURL   string           // 服务器地址
    playerID    string           // 玩家ID
    playerName  string           // 玩家名称
    conn        *websocket.Conn  // WebSocket 连接
    connected   bool             // 是否已连接
    reconnecting bool            // 是否正在重连
    send        chan []byte      // 发送消息通道
    receive     chan []byte      // 接收消息通道
    done        chan struct{}    // 关闭信号

    // 回调函数
    onStateChange func(*protocol.GameState)
    onTurn       func(*protocol.YourTurn)
    onChat       func(*protocol.ChatMessage)
    onError      func(error)
    onConnect    func()
    onDisconnect func()
}
```

### 3.2 主要方法

| 方法 | 说明 |
|-----|------|
| `NewClient(config)` | 创建新的客户端 |
| `Connect()` | 连接到服务器 |
| `Disconnect()` | 断开连接 |
| `Reconnect()` | 重新连接（最多 5 次尝试） |
| `Send(msg)` | 发送消息 |
| `SendPlayerAction(action, amount)` | 发送玩家动作 |
| `SendChat(content)` | 发送聊天消息 |
| `SendPing()` | 发送心跳 |
| `IsConnected()` | 检查是否已连接 |

### 3.3 回调处理

```go
// 连接成功
OnConnect: func()

// 断开连接
OnDisconnect: func()

// 收到错误
OnError: func(error)

// 游戏状态更新
OnStateChange: func(*protocol.GameState)

// 轮到玩家行动
OnTurn: func(*protocol.YourTurn)

// 收到聊天消息
OnChat: func(*protocol.ChatMessage)
```

### 3.4 重连机制

- 自动检测连接断开
- 最多尝试 5 次重连
- 重连间隔递增（1s, 2s, 3s, 4s, 5s）
- 重连成功后自动重新加入游戏

---

## 4. HOST TUI 界面 (ui/host)

### 4.1 Model 结构

```go
type Model struct {
    server       *host.Server
    gameState    *protocol.GameState
    selectedMenu int
    menuItems    []string
    width        int
    height       int
    err          error
}
```

### 4.2 界面布局

```
┌─────────────────────────────────────┐
│ Texas Hold'em Poker - 游戏控制台      │
│ 当前时间: 15:04:05                  │
├─────────────────────────────────────┤
│ 当前阶段: 翻牌圈                     │
├─────────────────────────────────────┤
│ 公共牌: [A♠] [K♠] [Q♠]             │
├─────────────────────────────────────┤
│ 1. Alice (座位0) | 筹码: 1000      │
│    [A♥] [K♥]                       │
│ 2. Bob (座位1) | 筹码: 800         │
│    [  ?  ] [  ?  ]                  │
│ 3. Charlie (座位2) | 筹码: 1200     │
│    [  ?  ] [  ?  ]                  │
├─────────────────────────────────────┤
│ 底池: 300                            │
├─────────────────────────────────────┤
│     [ 开始游戏 ] [ 暂停游戏 ]         │
│     [ 查看日志 ] [ 退出 ]             │
└─────────────────────────────────────┘
```

### 4.3 操作说明

| 按键 | 功能 |
|-----|------|
| ↑/k | 上移菜单 |
| ↓/j | 下移菜单 |
| Enter | 确认选择 |
| q/Ctrl+C | 退出 |

### 4.4 样式定义

| 样式 | 颜色 | 用途 |
|-----|------|------|
| styleTitle | #FF79C6 (粉色) | 标题 |
| styleActive | #50FA7B (绿色) | 激活状态 |
| styleInactive | #6272A4 (灰色) | 非激活状态 |
| stylePot | #F1FA8C (黄色) | 底池金额 |
| styleAction | #FFB86C (橙色) | 动作提示 |

---

## 5. CLIENT TUI 界面 (ui/client)

### 5.1 屏幕类型

```go
type ScreenType int

const (
    ScreenConnect ScreenType = iota  // 连接屏幕
    ScreenLobby                       // 大厅屏幕
    ScreenGame                        // 游戏屏幕
    ScreenAction                      // 下注屏幕
    ScreenChat                        // 聊天屏幕
)
```

### 5.2 连接屏幕

```
┌─────────────────────────────────────┐
│ Texas Hold'em Poker                 │
│                                     │
│ 请输入您的名称:                      │
│ [ Alice                            ] │
│                                     │
│ 按 Enter 连接，按 Ctrl+C 退出        │
└─────────────────────────────────────┘
```

### 5.3 游戏屏幕

```
┌─────────────────────────────────────┐
│ Texas Hold'em Poker                 │
│ 当前阶段: 翻牌圈                     │
│ 底池: 300                            │
├─────────────────────────────────────┤
│ 公共牌: [A♠] [K♠] [Q♠]             │
├─────────────────────────────────────┤
│ Alice | 筹码: 1000 | 下注: 50 | 游戏中│
│    [A♥] [K♥]                       │
│ Bob | 筹码: 800 | 下注: 50 | 游戏中  │
│    [  ?  ] [  ?  ]                  │
├─────────────────────────────────────┤
│ [F]old [C]all [R]aise [A]ll-in [H]elp │
└─────────────────────────────────────┘
```

### 5.4 操作说明

**游戏屏幕：**

| 按键 | 功能 |
|-----|------|
| f/F | 弃牌 (Fold) |
| c/C | 跟注 (Call) |
| r/R | 加注 (Raise) |
| a/A | 全下 (All-in) |
| h/H | 帮助 |

**聊天屏幕：**

| 按键 | 功能 |
|-----|------|
| 输入文字 | 聊天内容 |
| Enter | 发送消息 |
| Esc | 返回游戏 |

---

## 6. 测试覆盖

### 6.1 协议测试 (internal/protocol/message_test.go)

| 测试项 | 说明 |
|-------|------|
| TestNewBaseMessage | 创建基本消息 |
| TestNewJoinRequest | 创建加入请求 |
| TestNewPlayerActionRequest | 创建动作请求 |
| TestNewChatRequest | 创建聊天请求 |
| TestNewPingRequest | 创建心跳请求 |
| TestJoinAck_JSON | 加入确认序列化/反序列化 |
| TestGameState_JSON | 游戏状态序列化/反序列化 |
| TestYourTurn | 回合通知结构 |
| TestShowdown | 摊牌结果结构 |
| TestChatMessage | 聊天消息结构 |
| TestError | 错误消息结构 |
| TestPong | 心跳响应结构 |

### 6.2 运行测试

```bash
# 运行所有测试
go test ./...

# 运行协议测试（带详细输出）
go test ./internal/protocol/ -v

# 构建所有包
go build ./...
```

### 6.3 测试结果

```
ok  github.com/wilenwang/just_play/Texas-Holdem/internal/card      (cached)
ok  github.com/wilenwang/just_play/Texas-Holdem/internal/protocol  0.279s
ok  github.com/wilenwang/just_play/Texas-Holdem/pkg/evaluator       (cached)
```

---

## 7. 依赖库

| 库 | 版本 | 用途 |
|---|------|------|
| gorilla/websocket | v1.5.3 | WebSocket 实现 |
| charmbracelet/bubbletea | v1.3.10 | TUI 框架 |
| charmbracelet/lipgloss | v1.1.0 | 终端样式渲染 |

---

## 8. API 使用示例

### 8.1 启动 HOST 服务器

```go
package main

import (
    "github.com/wilenwang/just_play/Texas-Holdem/pkg/game"
    "github.com/wilenwang/just_play/Texas-Holdem/server/host"
    "github.com/wilenwang/just_play/Texas-Holdem/ui/host"
)

func main() {
    // 创建游戏配置
    config := &game.Config{
        MaxPlayers:    9,
        SmallBlind:    10,
        BigBlind:      20,
        StartingChips: 1000,
    }

    // 创建游戏服务器
    server := host.NewServer(config)

    // 启动 TUI
    host.Start(server)
}
```

### 8.2 启动 CLIENT 客户端

```go
package main

import (
    "github.com/wilenwang/just_play/Texas-Holdem/ui/client"
)

func main() {
    // 启动客户端 TUI
    client.Start()
}
```

### 8.3 直接使用客户端库

```go
import "github.com/wilenwang/just_play/Texas-Holdem/server/client"

func main() {
    // 创建客户端
    cli := client.NewClient(&client.Config{
        ServerURL:  "ws://localhost:8080/ws",
        PlayerName: "Alice",
        OnConnect: func() {
            println("Connected!")
        },
        OnGameState: func(state *protocol.GameState) {
            // 处理游戏状态更新
        },
        OnTurn: func(turn *protocol.YourTurn) {
            // 轮到玩家行动
            cli.SendPlayerAction(models.ActionCall, 0)
        },
    })

    // 连接服务器
    cli.Connect()

    // 发送聊天
    cli.SendChat("Hello everyone!")
}
```

---

## 9. 注意事项

1. **线程安全**：服务器使用读写锁保护客户端列表
2. **消息序列化**：所有消息使用 JSON 格式
3. **心跳检测**：防止长时间空闲连接被断开
4. **自动重连**：客户端支持自动重连机制
5. **座位分配**：自动分配可用座位
6. **底牌隐私**：玩家只能看到自己的底牌

---

## 10. 下一步（Phase 3）

- [ ] AI 玩家实现
- [ ] 完整聊天系统
- [ ] 游戏历史记录
- [ ] 房间管理系统
- [ ] 多房间支持
