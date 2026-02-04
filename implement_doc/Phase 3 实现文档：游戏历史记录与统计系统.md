# Phase 3 实现文档：游戏历史记录与统计系统

## 概述

本文档描述了德州扑克游戏 Phase 3（游戏历史记录与统计系统）的实现细节，包括手牌历史记录、玩家统计数据、TUI 聊天组件和扑克牌渲染组件。

## 项目结构

```
Texas-Holdem/
├── pkg/
│   └── game/
│       ├── history.go     # 手牌历史记录管理
│       └── stats.go      # 玩家统计数据管理
├── ui/
│   └── components/
│       ├── chat.go       # 聊天组件
│       └── render.go     # 扑克牌渲染组件
├── implement_doc/
│   ├── Phase 1 实现文档：核心游戏引擎.md
│   ├── Phase 2 实现文档：WebSocket通信层.md
│   └── Phase 3 实现文档：游戏历史记录与统计系统.md
```

---

## 1. 手牌历史记录 (pkg/game/history.go)

### 1.1 核心数据结构

**HandHistory - 一手牌的历史记录：**

```go
type HandHistory struct {
    HandID         int              // 手牌编号
    Timestamp      time.Time        // 时间戳
    GameID         string           // 游戏ID
    Players        []HistoryPlayer  // 参与的玩家
    CommunityCards [5]card.Card     // 公共牌
    Actions        []HistoryAction  // 行动记录
    Showdown       []ShowdownInfo   // 摊牌信息
    Pot            int              // 底池金额
    Winners        []WinnerInfo     // 获胜者信息
}
```

**HistoryPlayer - 历史记录中的玩家信息：**

```go
type HistoryPlayer struct {
    ID         string       // 玩家ID
    Name       string       // 玩家名称
    Seat       int          // 座位号
    HoleCards  [2]card.Card // 底牌
    FinalChips int          // 最终筹码
    WonChips   int          // 赢得筹码
    IsWinner   bool         // 是否获胜
}
```

**HistoryAction - 历史记录中的行动：**

```go
type HistoryAction struct {
    PlayerID   string           // 玩家ID
    PlayerName string           // 玩家名称
    Round      Stage           // 所在阶段
    Action     models.ActionType // 执行的行动
    Amount     int              // 下注金额
    Timestamp  time.Time        // 时间戳
}
```

### 1.2 HistoryManager 管理器

```go
type HistoryManager struct {
    hands       []HandHistory   // 所有手牌历史
    currentHand *HandHistory     // 当前正在进行的牌局
    file       *os.File          // 文件句柄（用于保存）
    filename   string             // 保存文件名
    mu         chan struct{}     // 互斥锁
}
```

### 1.3 主要方法

| 方法 | 说明 |
|-----|------|
| `NewHistoryManager(filename)` | 创建历史记录管理器，可选加载历史文件 |
| `StartHand(handID, gameID)` | 开始记录新的一手牌 |
| `AddPlayer(id, name, seat, chips)` | 添加玩家到当前手牌记录 |
| `RecordAction(playerID, playerName, round, action, amount)` | 记录玩家行动 |
| `RecordShowdown(...)` | 记录摊牌信息 |
| `EndHand(community, pot, winners)` | 结束当前手牌记录 |
| `GetRecentHands(n)` | 获取最近 N 手牌历史 |
| `GetAllHands()` | 获取所有手牌历史 |
| `GetHandCount()` | 获取历史手牌总数 |
| `ExportToText()` | 将手牌历史导出为文本格式 |

### 1.4 文本导出示例

```
=== 手牌 #5 ===
时间: 2025-01-15 14:30:25
底池: 450

玩家:
 [1] Alice: 筹码 1200 底牌: A♠ K♠ [胜]
 [2] Bob: 筹码 800

公共牌: A♥ K♥ Q♠ J♦ 10♠

行动:
 Alice [翻牌前]: 加注 (100)
 Bob [翻牌前]: 跟注 (100)
 Alice [翻牌]: 过牌
 Bob [翻牌]: 过牌
 Alice [转牌]: 下注 (200)
 Bob [转牌]: 弃牌

获胜者:
 Alice: 赢得 450 筹码
```

---

## 2. 玩家统计系统 (pkg/game/stats.go)

### 2.1 PlayerStats 玩家统计信息

```go
type PlayerStats struct {
    PlayerID      string    // 玩家ID
    Name          string    // 玩家名称
    HandsPlayed   int       // 参与的手牌数
    HandsWon      int       // 获胜的手牌数
    TotalWinnings int      // 总盈利
    TotalLosses   int      // 总亏损
    BiggestPot    int       // 最大底池
    TotalBets     int      // 总下注金额
    TotalCalls    int      // 跟注次数
    TotalRaises   int      // 加注次数
    TotalFolds    int      // 弃牌次数
    TotalChecks   int      // 看牌次数
    WinRate       float64  // 胜率
    ProfitPerHand float64  // 每手平均盈利
    FirstActions  int      // 首位行动次数
    LastActions   int      // 最后行动次数
    CreatedAt     time.Time // 统计开始时间
    UpdatedAt     time.Time // 最后更新时间
}
```

### 2.2 StatsManager 统计管理器

```go
type StatsManager struct {
    mu          sync.RWMutex            // 读写锁
    playerStats map[string]*PlayerStats  // 玩家ID到统计信息的映射
}
```

### 2.3 主要方法

| 方法 | 说明 |
|-----|------|
| `NewStatsManager()` | 创建统计管理器 |
| `GetOrCreateStats(playerID, name)` | 获取或创建玩家统计信息 |
| `UpdateHandPlayed(playerID, name)` | 更新玩家参与的手牌数 |
| `UpdateHandWon(playerID, name, wonChips)` | 更新玩家获胜信息 |
| `UpdateHandLost(playerID, name, lostChips)` | 更新玩家亏损信息 |
| `UpdateAction(playerID, name, action, amount)` | 记录玩家动作 |
| `GetStats(playerID)` | 获取玩家统计信息（线程安全） |
| `GetAllStats()` | 获取所有玩家的统计信息 |
| `GetLeaderboard(limit)` | 获取排行榜（按盈利排序） |
| `GetWinRateLeaderboard(limit)` | 获取胜率排行榜（至少10手牌） |
| `GetTopWinner()` | 获取最大赢家 |
| `GetTopLoser()` | 获取最大输家 |
| `RemoveStats(playerID)` | 删除玩家统计信息 |
| `Clear()` | 清空所有统计信息 |

### 2.4 统计报告示例

```text
=== Alice 统计 ===
参与手牌: 50
获胜手牌: 15 (30.0%)
总盈利: +500
每手平均: +10.00
最大底池: 1200
总下注: 2500
跟注次数: 80
加注次数: 25
弃牌次数: 45
看牌次数: 30
统计时间: 2025-01-01 - 2025-01-15 14:30
```

### 2.5 排行榜查询

```go
// 获取盈利排行榜前10名
leaderboard := statsManager.GetLeaderboard(10)

// 获取胜率排行榜（至少10手牌）
winRateBoard := statsManager.GetWinRateLeaderboard(10)
```

---

## 3. TUI 聊天组件 (ui/components/chat.go)

### 3.1 ChatMessage 聊天消息

```go
type ChatMessage struct {
    PlayerID   string    // 玩家ID
    PlayerName string    // 玩家名称
    Content    string    // 消息内容
    Time       time.Time // 发送时间
    IsSystem   bool      // 是否为系统消息
}
```

### 3.2 ChatModel 聊天组件模型

```go
type ChatModel struct {
    messages []ChatMessage   // 消息列表
    input    textarea.Model  // 输入框（Bubble Tea）
    visible  bool            // 是否显示
    maxLines int             // 最大显示行数
    width    int             // 组件宽度
    height   int             // 组件高度
    focused  bool            // 是否获得焦点
}
```

### 3.3 主要方法

| 方法 | 说明 |
|-----|------|
| `NewChatModel()` | 创建新的聊天组件 |
| `Toggle()` | 切换显示/隐藏状态 |
| `SetVisible(visible)` | 设置是否显示 |
| `IsVisible()` | 返回是否显示 |
| `SetSize(width, height)` | 设置组件尺寸 |
| `Focus()` / `Blur()` | 聚焦/取消聚焦输入框 |
| `AddMessage(playerID, playerName, content)` | 添加玩家消息 |
| `AddSystemMessage(content)` | 添加系统消息 |
| `ClearMessages()` | 清空所有消息 |
| `GetMessages()` | 获取所有消息 |
| `GetInputValue()` | 获取输入框内容 |
| `ClearInput()` | 清空输入框 |
| `SetInputValue(value)` | 设置输入框内容 |

### 3.4 界面布局

```
┌────────────────────────────────────┐
│           聊天                       │
│                                    │
│ [14:30] Alice: 大家好！              │
│ [14:31] Bob: 准备开始了吗？           │
│ [14:32] 系统: Charlie 已加入游戏      │
│                                    │
│ » 输入消息...                        │
│                                    │
│  Enter: 发送  |  Esc: 关闭            │
└────────────────────────────────────┘
```

### 3.5 使用示例

```go
// 创建聊天组件
chat := components.NewChatModel()

// 添加消息
chat.AddMessage("player1", "Alice", "轮到我了")
chat.AddSystemMessage("游戏开始")

// 在 TUI 中更新
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ...
    m.chat, cmd = m.chat.Update(msg)
    return m, cmd
}

// 在 TUI 中渲染
func (m *Model) View() string {
    // ...
    return lipgloss.JoinHorizontal(lipgloss.Top,
        mainContent,
        m.chat.View())
}
```

---

## 4. 扑克牌渲染组件 (ui/components/render.go)

### 4.1 颜色定义

| 颜色变量 | 值 | 用途 |
|---------|-----|------|
| `suitRed` | 196 | 红桃、方块 - 亮红色 |
| `suitBlack` | 15 | 黑桃、梅花 - 亮白色 |
| `backColor` | 239 | 牌背 - 深灰色 |
| `bgColor` | 236 | 背景色 |

### 4.2 渲染函数

| 函数 | 说明 |
|-----|------|
| `RenderCard(c, faceUp)` | 渲染单张扑克牌（带颜色） |
| `RenderCardBack()` | 渲染牌背 |
| `RenderCards(cards, faceUp)` | 渲染多张牌（水平排列） |
| `RenderCardsCompact(cards, faceUp)` | 紧凑模式渲染多张牌 |
| `RenderCardCompact(c, faceUp)` | 紧凑模式渲染单张牌 |
| `RenderCardASCII(c, faceUp)` | ASCII 版本渲染（无颜色） |
| `RenderCardsASCII(cards, faceUp)` | ASCII 版本渲染多张牌 |
| `RenderCardLarge(c, faceUp)` | 大尺寸渲染单张牌 |
| `RenderCommunityCards(cards, faceUp)` | 渲染公共牌区域 |
| `RenderPot(pot)` | 渲染底池信息 |
| `RenderPlayerInfo(name, chips, status)` | 渲染玩家信息行 |

### 4.3 辅助函数

```go
// 将点数转换为字符串
RankToString(r card.Rank) string

// 将花色转换为单个字符
SuitToChar(s card.Suit) string

// 将花色转换为完整名称
SuitToName(s card.Suit) string

// 获取牌面的颜色
GetCardColor(c card.Card) lipgloss.Color
```

### 4.4 渲染效果

**标准模式：**
```
[A♠] [K♥] [Q♦] [J♣] [10♠]
```

**紧凑模式：**
```
[AS] [KH] [QD] [JC] [TS]
```

**ASCII 模式（无颜色）：**
```
[AS] [KH] [QD] [JC] [TS]
```

**牌背：**
```
[??]
```

### 4.5 使用示例

```go
import "github.com/wilenwang/just_play/Texas-Holdem/ui/components"

// 渲染手牌
hand := [2]card.Card{{Rank: card.Ace, Suit: card.Spades}, {Rank: card.King, Suit: card.Spades}}
handStr := components.RenderCardsCompact(hand[:], true)

// 渲染公共牌
community := [5]card.Card{...}
communityStr := components.RenderCommunityCards(community, true)

// 渲染底池
potStr := components.RenderPot(500)

// 渲染玩家信息
playerStr := components.RenderPlayerInfo("Alice", 1000, "已下注")
```

---

## 5. 集成示例

### 5.1 在游戏引擎中集成历史记录

```go
engine := game.NewGameEngine(config)

// 创建历史记录管理器
historyManager, _ := game.NewHistoryManager("history.json")

// 创建统计管理器
statsManager := game.NewStatsManager()

// 游戏过程中记录
historyManager.StartHand(handID, gameID)
historyManager.AddPlayer(p.ID, p.Name, p.Seat, p.Chips)

// 记录行动
historyManager.RecordAction(p.ID, p.Name, stage, action, amount)

// 更新统计
statsManager.UpdateAction(p.ID, p.Name, action, amount)

// 摊牌时记录
historyManager.RecordShowdown(p.ID, p.Name, handRank, handName, bestCards)
historyManager.EndHand(community, pot, winners)

// 更新玩家盈利统计
if wonChips > 0 {
    statsManager.UpdateHandWon(p.ID, p.Name, wonChips)
} else {
    statsManager.UpdateHandLost(p.ID, p.Name, lostChips)
}
```

### 5.2 在 TUI 中集成聊天和渲染

```go
type GameModel struct {
    chat   *components.ChatModel
    width  int
    height int
}

func NewGameModel() *GameModel {
    return &GameModel{
        chat: components.NewChatModel(),
    }
}

func (m *GameModel) View() string {
    // 顶部：公共牌
    community := components.RenderCommunityCards(state.CommunityCards, true)

    // 玩家区域
    var players strings.Builder
    for _, p := range state.Players {
        holeCards := components.RenderCardsCompact(p.HoleCards[:], p.IsSelf)
        info := components.RenderPlayerInfo(p.Name, p.Chips, p.Status.String())
        players.WriteString(info + "\n" + holeCards + "\n\n")
    }

    // 底池
    pot := components.RenderPot(state.Pot)

    // 右侧：聊天组件
    chatView := m.chat.View()

    return lipgloss.JoinHorizontal(lipgloss.Top,
        lipgloss.JoinVertical(lipgloss.Left,
            community,
            pot,
            players.String()),
        chatView)
}
```

---

## 6. 依赖库

| 库 | 版本 | 用途 |
|---|------|------|
| charmbracelet/bubbletea | v1.3.10 | TUI 框架（textarea） |
| charmbracelet/lipgloss | v1.1.0 | 终端样式渲染 |

---

## 7. 注意事项

1. **线程安全**：StatsManager 使用读写锁保护数据
2. **历史持久化**：HistoryManager 支持从文件加载和保存历史记录
3. **消息限制**：聊天消息限制 200 字符，最多显示 50 条
4. **颜色兼容性**：RenderCardASCII 适用于不支持颜色的终端
5. **自动保存**：一手牌结束后自动保存历史记录到文件

---

## 8. 下一步（Phase 4）

- [ ] AI 玩家实现
- [ ] 房间管理系统
- [ ] 多房间支持
- [ ] 游戏回放功能
- [ ] 统计图表展示
