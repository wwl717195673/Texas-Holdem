# TUI 客户端 Bug 修复文档

## 概述

本文档记录了 TUI 客户端在实现过程中发现的问题及其修复方案。通过创建自测程序，我们诊断并修复了两个关键问题。

## 问题发现过程

### 创建自测程序

为了诊断 TUI 客户端的问题，创建了 `cmd/client/test/main.go` 自测程序，该程序能够：

1. 自动连接到 WebSocket 服务器
2. 发送加入游戏请求
3. 接收并解析 JoinAck 消息
4. 接收并解析 GameState 消息
5. 等待轮到玩家回合
6. 发送玩家动作

**文件位置**: `cmd/client/test/main.go`

**运行方式**:
```bash
go run cmd/client/test/main.go -server localhost:8080 -name TestPlayer
```

## 发现的问题

### 问题 1: GameState 消息解析错误

**症状**: 测试程序在阶段4失败，错误信息 "GameState 为空"

**根因**: 在 `server/client/client.go` 的 `handleGameState` 函数中，代码期望服务器发送的消息格式为：
```json
{
  "type": "game_state",
  "game_state": { ... }
}
```

但实际服务器直接发送 GameState 结构：
```json
{
  "type": "game_state",
  "game_id": "...",
  "stage": 1,
  ...
}
```

**修复位置**: `server/client/client.go:375-387`

**修复前**:
```go
func (c *Client) handleGameState(data []byte) {
	var msg struct {
		protocol.BaseMessage
		GameState *protocol.GameState `json:"game_state"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal GameState: %v", err)
		return
	}

	if msg.GameState != nil && c.onStateChange != nil {
		c.onStateChange(msg.GameState)
	}
}
```

**修复后**:
```go
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
```

### 问题 2: TUI 客户端消息接收循环中断

**症状**: TUI 客户端连接成功后变得无响应，键盘输入也无法工作

**根因**: Bubble Tea 的 `tea.Tick` 机制在返回 `nil` 时不会被继续调用。原始实现中：
1. `tick()` 函数在没有消息时返回 `nil`
2. `Update()` 方法使用 `waitForExtMsg()` 阻塞等待
3. 这导致消息循环中断，无法处理后续消息

**修复位置**: `ui/client/model.go`

**修复方案**:

1. **引入 `tickMsg` 类型** (line 196-197):
```go
// tickMsg 自定义的 tick 消息，用于持续检查外部通道
type tickMsg time.Time
```

2. **修改 `tick()` 函数** (line 199-213):
```go
// tick 定期检查外部消息通道
// 每次都返回一个新的 tick 命令，确保持续检查
func (m *Model) tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		select {
		case msg := <-m.extMsgChan:
			// 收到外部消息，返回该消息，并继续 tick
			// 在 Update 中处理消息后会再次调用 tick
			return msg
		default:
			// 没有消息，返回 tickMsg 保持循环
			return tickMsg(t)
		}
	})
}
```

3. **在 `Update()` 中处理 `tickMsg`** (line 238-240):
```go
case tickMsg:
	// tick 消息，继续检查通道
	return m, m.tick()
```

4. **所有消息处理返回 `m.tick()` 而非 `m.waitForExtMsg()`**:
   - ConnectedMsg
   - DisconnectedMsg
   - JoinAckResultMsg
   - GameStateMsg
   - YourTurnMsg
   - PlayerJoinedMsg
   - PlayerLeftMsg
   - PlayerActedMsg
   - ShowdownMsg
   - ChatMsg
   - ErrorMsg
   - tickMsg
   - tea.WindowSizeMsg

5. **所有屏幕的 `update` 方法返回 `m.tick()` 而非 `nil`**:
   - updateConnect
   - updateLobby
   - updateGame
   - updateAction
   - updateShowdown
   - updateChat

6. **使用 `tea.Batch` 组合命令**:
```go
// 在 updateConnect 中
case "enter":
	if !m.connecting {
		m.connecting = true
		m.err = nil
		cmd := m.doConnect()
		return m, tea.Batch(cmd, m.tick())
	}
	return m, m.tick()

// 在 updateGame 中
case "f":
	// 弃牌
	return m, tea.Batch(m.sendAction(models.ActionFold, 0), m.tick())
```

## 修复效果

修复后的 TUI 客户端能够：
1. ✅ 正确接收并解析服务器发送的 GameState 消息
2. ✅ 持续监听 WebSocket 回调，不会中断消息循环
3. ✅ 正常响应键盘输入
4. ✅ 实时更新游戏状态显示
5. ✅ 处理玩家动作（弃牌、跟注、加注、全下）

## 验证方法

### 方法 1: 使用自测程序
```bash
# 启动服务器
go run cmd/server/main.go

# 运行测试程序
go run cmd/client/test/main.go -server localhost:8080 -name TestPlayer
```

### 方法 2: 使用 TUI 客户端
```bash
# 启动服务器
go run cmd/server/main.go

# 运行 TUI 客户端
go run cmd/client/main.go
```

操作步骤：
1. 在连接屏幕中，输入服务器地址（默认 localhost:8080）
2. 按 Tab 切换到玩家名称，输入名称
3. 按 Enter 连接
4. 观察是否能正常显示游戏状态
5. 尝试使用 F/C/R/A 键进行游戏操作

## 修改的文件

| 文件 | 修改内容 | 行数 |
|------|----------|------|
| `server/client/client.go` | 修复 GameState 消息解析 | ~12 行 |
| `ui/client/model.go` | 修复消息循环机制 | ~50 行 |
| `cmd/client/test/main.go` | 新增自测程序 | ~370 行 |

## 技术要点

### Bubble Tea 消息循环机制

**关键理解**:
- `tea.Tick` 返回 `nil` 时不会被继续调用
- 需要返回一个自定义消息类型来维持循环
- 每次处理消息后都必须返回继续检查的命令

**正确模式**:
```go
type tickMsg time.Time

func (m *Model) tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		select {
		case msg := <-m.extMsgChan:
			return msg
		default:
			return tickMsg(t)  // 返回自定义消息而非 nil
		}
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, m.tick()  // 继续循环
	case SomeOtherMsg:
		// 处理消息
		return m, m.tick()  // 处理后继续循环
	}
	return m, m.tick()
}
```

### WebSocket 消息桥接

使用缓冲通道将 WebSocket 回调桥接到 Bubble Tea 消息系统：
```go
extMsgChan: make(chan tea.Msg, 100)

// 在回调中发送消息
OnStateChange: func(state *protocol.GameState) {
	m.extMsgChan <- GameStateMsg{State: state}
}
```

## 注意事项

1. **消息类型匹配**: 确保客户端和服务器的消息格式完全一致
2. **持续循环**: 所有 `Update` 返回路径都必须包含 `m.tick()`
3. **命令组合**: 使用 `tea.Batch()` 组合多个命令时，别忘了加上 `m.tick()`
4. **缓冲通道**: `extMsgChan` 使用缓冲通道 (size=100) 防止阻塞 WebSocket 回调

## 后续优化建议

1. 添加连接超时处理
2. 添加断线重连机制
3. 优化消息通道大小
4. 添加更详细的错误日志
5. 完善测试程序，覆盖更多场景
