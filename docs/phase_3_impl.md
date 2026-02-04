# Phase_3 实现文档

## 本次实现内容

实现了 Phase_3.md 中除 AI 玩家外的所有高级功能。

## 实现文件

| 文件 | 功能 |
|-----|------|
| `pkg/game/history.go` | 手牌历史记录管理 |
| `ui/components/chat.go` | 聊天系统组件 |
| `pkg/game/stats.go` | 游戏统计管理 |
| `ui/components/render.go` | 扑克牌渲染组件 |

## 各模块实现说明

### 1. 手牌历史记录 (history.go)

**功能：**
- 记录每手牌的完整信息（玩家、行动、公共牌、摊牌结果）
- 支持保存到 JSON 文件
- 支持导出为可读文本格式
- 按时间戳记录所有玩家行动

**核心结构：**
- `HandHistory` - 一手牌的历史记录
- `HistoryManager` - 历史记录管理器
- `HistoryPlayer` - 参与玩家信息
- `HistoryAction` - 玩家行动记录
- `ShowdownInfo` - 摊牌详细信息

**主要方法：**
- `StartHand(handID, gameID)` - 开始记录新牌局
- `RecordAction(...)` - 记录玩家行动
- `RecordShowdown(...)` - 记录摊牌信息
- `EndHand(...)` - 结束并保存牌局
- `GetRecentHands(n)` - 获取最近N手记录
- `ExportToText()` - 导出为文本格式

### 2. 聊天系统 (chat.go)

**功能：**
- 支持玩家发送和接收聊天消息
- 系统公告/通知消息
- 消息历史记录（最大50条）
- 可切换显示/隐藏的聊天面板
- 支持 Bubble Tea 的 textarea 输入组件

**核心结构：**
- `ChatMessage` - 单条聊天消息
- `ChatModel` - 聊天组件模型

**主要方法：**
- `Toggle()` / `SetVisible(bool)` - 控制显示
- `AddMessage(playerID, playerName, content)` - 添加玩家消息
- `AddSystemMessage(content)` - 添加系统消息
- `Update(msg)` - 处理 Bubble Tea 消息
- `View()` - 渲染聊天界面

### 3. 游戏统计 (stats.go)

**功能：**
- 跟踪玩家参与手牌数、获胜数
- 计算胜率、每手平均盈利
- 统计各种动作次数（跟注、加注、弃牌等）
- 排行榜功能（按盈利、胜率）
- 最大赢家/输家统计

**核心结构：**
- `PlayerStats` - 玩家统计数据
- `StatsManager` - 统计管理器

**主要方法：**
- `UpdateHandPlayed()` - 记录参与手牌
- `UpdateHandWon(wonChips)` - 记录获胜
- `UpdateHandLost(lostChips)` - 记录亏损
- `UpdateAction(action, amount)` - 记录动作
- `GetLeaderboard(limit)` - 获取盈利排行榜
- `GetWinRateLeaderboard(limit)` - 获取胜率排行榜
- `Report()` / `ShortReport()` - 生成统计报告

### 4. 扑克牌渲染 (render.go)

**功能：**
- 支持彩色渲染（红桃/方块红色，黑桃/梅花白色）
- 牌背渲染
- 多牌水平/紧凑排列
- 大尺寸详细渲染
- ASCII 纯文本版本（无颜色终端）

**核心结构：**
- `CardStyle` - 自定义渲染样式配置

**主要方法：**
- `RenderCard(card, faceUp)` - 渲染单张牌
- `RenderCards(cards, faceUp)` - 渲染多张牌（水平）
- `RenderCardCompact(card, faceUp)` - 紧凑模式
- `RenderCardASCII(card, faceUp)` - ASCII 纯文本
- `RenderCommunityCards(cards, faceUp)` - 公共牌区域
- `RenderPot(pot)` - 渲染底池
- `RenderPlayerInfo(...)` - 渲染玩家信息

## 使用方式

```go
// 历史记录
hm := game.NewHistoryManager("history.json")
hm.StartHand(1, "game_123")
hm.RecordAction("player1", "Alice", game.StagePreFlop, models.ActionCall, 100)
hm.EndHand(communityCards, 500, winners)

// 聊天
chat := components.NewChatModel()
chat.AddMessage("player1", "Alice", "Hello!")
chat.Toggle()
model.View() // 在 TUI 中渲染

// 统计
stats := game.NewStatsManager()
stats.UpdateHandPlayed("player1", "Alice")
stats.UpdateHandWon("player1", "Alice", 250)
report := stats.GetStats("player1").Report()

// 渲染
cards := []card.Card{{Rank: card.Ace, Suit: card.Spades}}
components.RenderCards(cards, true)
```

## 待实现

- [ ] AI 玩家功能（Phase_3.md 中标记为 P0，但用户要求暂不实现）
- [ ] 与游戏引擎集成（状态同步）
- [ ] 单元测试
