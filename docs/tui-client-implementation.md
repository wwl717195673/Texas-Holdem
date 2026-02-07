# TUI 客户端改造实现文档

## 概述

将德州扑克游戏客户端从纯命令行界面 (CLI) 改造为使用 Bubble Tea + Lipgloss 的终端用户界面 (TUI)。

## 实现内容

### 1. 新增文件

#### ui/client/msgs.go
- **功能**：定义 Bubble Tea 自定义消息类型
- **主要消息类型**：
  - `ConnectedMsg` - 连接成功消息
  - `DisconnectedMsg` - 断开连接消息
  - `JoinAckResultMsg` - 加入确认结果消息
  - `GameStateMsg` - 游戏状态更新消息
  - `YourTurnMsg` - 轮到玩家回合消息
  - `PlayerJoinedMsg` - 玩家加入通知消息
  - `PlayerLeftMsg` - 玩家离开通知消息
  - `PlayerActedMsg` - 玩家动作通知消息
  - `ShowdownMsg` - 摊牌结果消息
  - `ChatMsg` - 聊天消息
  - `ErrorMsg` - 错误消息

### 2. 重写文件

#### ui/client/model.go
- **功能**：完整的 TUI 模型实现
- **屏幕类型**：
  - `ScreenConnect` - 连接屏幕（输入服务器地址和玩家名称）
  - `ScreenLobby` - 大厅屏幕（等待游戏开始）
  - `ScreenGame` - 游戏屏幕（显示游戏状态、玩家列表、动作按钮）
  - `ScreenAction` - 动作输入屏幕（输入加注金额）
  - `ScreenShowdown` - 摊牌结果屏幕（显示摊牌详情）
  - `ScreenChat` - 聊天屏幕（发送/接收聊天消息）

- **核心功能**：
  - 使用外部消息通道 (`extMsgChan`) 桥接 WebSocket 回调和 Bubble Tea 消息系统
  - 完整的游戏状态显示（阶段、底池、庄家位置、公共牌、玩家列表）
  - 键盘快捷键处理（F-弃牌、C-跟注、R-加注、A-全下、H-聊天、Q-退出）
  - 使用现有 `ui/components/render.go` 渲染扑克牌
  - 使用现有 `ui/components/chat.go` 处理聊天功能

#### cmd/client/main.go
- **功能**：TUI 程序入口
- **改动**：
  - 移除原有的 `bufio` 输入处理和 WebSocket 原生操作
  - 移除命令行参数解析（-server 和 -name）
  - 使用 Bubble Tea 的 `tea.NewProgram` 启动 TUI
  - 简化为约 30 行代码

### 3. 修改文件

#### server/client/client.go
- **改动**：添加 `OnJoinAck` 回调，用于处理加入游戏确认

## 关键技术点

### 消息桥接机制

由于 Bubble Tea 的架构限制（无法直接获取程序指针发送消息），采用**通道桥接**方案：

1. 在 Model 中创建 `extMsgChan chan tea.Msg` 通道
2. 在 `Init()` 方法中返回 `waitForExtMsg()` Cmd，持续监听通道
3. WebSocket 回调中向通道发送消息
4. 消息处理完成后继续返回 `waitForExtMsg()` 保持循环

```go
// 创建缓冲通道
extMsgChan: make(chan tea.Msg, 100)

// 等待外部消息
func (m *Model) waitForExtMsg() tea.Cmd {
    return func() tea.Msg {
        return <-m.extMsgChan
    }
}

// 回调中发送消息
OnStateChange: func(state *protocol.GameState) {
    m.extMsgChan <- GameStateMsg{State: state}
},
```

### 屏幕切换

通过 `screen` 字段控制当前显示的屏幕，`View()` 方法根据 `screen` 值调用对应的渲染方法：

```go
func (m *Model) View() string {
    switch m.screen {
    case ScreenConnect:
        return m.viewConnect()
    case ScreenLobby:
        return m.viewLobby()
    // ...
    }
}
```

### 样式定义

使用 Lipgloss 定义统一的 UI 样式：

```go
styleTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("228"))
styleActive   = lipgloss.NewStyle().Foreground(lipgloss.Color("50")).Bold(true)
styleError    = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Bold(true)
stylePot      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
```

## 使用方式

### 启动客户端

```bash
go run cmd/client/main.go
```

### 连接服务器

1. 在连接屏幕输入服务器地址（默认：localhost:8080）
2. 按 Tab 切换到玩家名称输入框
3. 输入玩家名称（默认：Player）
4. 按 Enter 连接

### 游戏操作

| 按键 | 功能 |
|------|------|
| F | 弃牌 |
| C | 跟注 |
| R | 加注 |
| A | 全下 |
| H | 聊天 |
| Q | 退出 |
| Ctrl+C | 强制退出 |

## 依赖关系

```
cmd/client/main.go
    └── ui/client (NewModel)
            ├── ui/client/model.go (TUI 模型)
            ├── ui/client/msgs.go (消息类型)
            ├── ui/components/render.go (扑克牌渲染)
            ├── ui/components/chat.go (聊天组件)
            └── server/client/client.go (WebSocket 客户端)
```

## 代码统计

| 文件 | 行数 | 说明 |
|------|------|------|
| ui/client/msgs.go | ~70 | 消息类型定义 |
| ui/client/model.go | ~1070 | TUI 模型实现 |
| cmd/client/main.go | ~30 | 程序入口 |
| server/client/client.go | +10 | 添加 JoinAck 回调 |

## 注意事项

1. 每个方法都添加了中文注释
2. 超过 30 行的方法在关键节点添加了中文注释
3. 复用了现有的 `ui/components/` 中的渲染组件
4. 使用通道机制解决了 Bubble Tea 与 WebSocket 回调的消息桥接问题
