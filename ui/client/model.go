package client

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/protocol"
	"github.com/wilenwang/just_play/Texas-Holdem/server/client"
	"github.com/wilenwang/just_play/Texas-Holdem/ui/components"
)

// ==================== 样式定义 ====================

var (
	// 标题样式
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("228")). // 亮黄色
			Align(lipgloss.Center)

	// 副标题样式
	styleSubtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")). // 灰色
			Faint(true)

	// 边框样式
	styleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1)

	// 激活状态样式
	styleActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("50")). // 亮绿色
			Bold(true)

	// 非激活状态样式
	styleInactive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")). // 暗灰色
			Faint(true)

	// 警告样式
	styleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // 红色
			Bold(true)

	// 错误样式
	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("160")). // 深红色
			Bold(true)

	// 底池样式
	stylePot = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226")) // 黄色

	// 按钮样式
	styleButton = lipgloss.NewStyle().
			Background(lipgloss.Color("237")). // 深灰色
			Foreground(lipgloss.Color("15")).  // 白色
			Padding(0, 2)

	// 激活按钮样式
	styleButtonActive = lipgloss.NewStyle().
				Background(lipgloss.Color("57")).  // 紫色
				Foreground(lipgloss.Color("15")). // 白色
				Padding(0, 2)

	// 输入框样式
	styleInput = lipgloss.NewStyle().
			Background(lipgloss.Color("235")). // 暗灰色
			Foreground(lipgloss.Color("15")).  // 白色
			Padding(0, 1)

	// 动作提示样式
	styleAction = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")). // 青色
			Bold(true)

	// 当前玩家高亮样式
	styleCurrentPlayer = lipgloss.NewStyle().
				Foreground(lipgloss.Color("228")). // 亮黄色
				Bold(true)

	// 高亮样式
	styleHighlight = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // 橙色
			Bold(true)

	// 操作按钮样式 - 弃牌（红色）
	styleBtnFold = lipgloss.NewStyle().
			Background(lipgloss.Color("52")).  // 深红背景
			Foreground(lipgloss.Color("196")). // 红色文字
			Bold(true).
			Padding(0, 1)

	// 操作按钮样式 - 过牌/跟注（绿色）
	styleBtnCall = lipgloss.NewStyle().
			Background(lipgloss.Color("22")).  // 深绿背景
			Foreground(lipgloss.Color("46")).  // 绿色文字
			Bold(true).
			Padding(0, 1)

	// 操作按钮样式 - 加注（黄色）
	styleBtnRaise = lipgloss.NewStyle().
			Background(lipgloss.Color("58")).  // 深黄背景
			Foreground(lipgloss.Color("226")). // 黄色文字
			Bold(true).
			Padding(0, 1)

	// 操作按钮样式 - 全下（紫色）
	styleBtnAllIn = lipgloss.NewStyle().
			Background(lipgloss.Color("53")).  // 深紫背景
			Foreground(lipgloss.Color("213")). // 亮紫文字
			Bold(true).
			Padding(0, 1)

	// 操作按钮样式 - 功能键（青色）
	styleBtnFunc = lipgloss.NewStyle().
			Background(lipgloss.Color("236")). // 暗灰背景
			Foreground(lipgloss.Color("81")).   // 青色文字
			Padding(0, 1)

	// 操作按钮样式 - 禁用状态
	styleBtnDisabled = lipgloss.NewStyle().
				Background(lipgloss.Color("235")). // 暗灰背景
				Foreground(lipgloss.Color("240")). // 灰色文字
				Faint(true).
				Padding(0, 1)

	// 操作栏容器样式
	styleActionBar = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false). // 仅顶部边框
			BorderForeground(lipgloss.Color("238")).
			PaddingTop(1)
)

// ==================== 屏幕类型定义 ====================

// ScreenType 表示当前屏幕类型
type ScreenType int

const (
	ScreenConnect  ScreenType = iota // 连接屏幕
	ScreenLobby                     // 大厅屏幕
	ScreenGame                      // 游戏屏幕
	ScreenAction                    // 动作输入屏幕
	ScreenShowdown                  // 摊牌结果屏幕
	ScreenResult                    // 结算屏幕
	ScreenChat                      // 聊天屏幕
)

// String 返回屏幕类型的字符串表示
func (s ScreenType) String() string {
	names := []string{"连接", "大厅", "游戏", "动作", "摊牌", "结算", "聊天"}
	if int(s) < len(names) {
		return names[s]
	}
	return "未知"
}

// ==================== TUI 模型 ====================

// Model TUI 主模型
type Model struct {
	// 客户端连接
	client *client.Client

	// 当前屏幕
	screen ScreenType

	// 连接屏幕输入
	serverInput  string // 服务器地址输入
	playerInput  string // 玩家名称输入
	connectField int    // 当前聚焦的输入框 (0=服务器, 1=玩家)
	connecting   bool   // 是否正在连接

	// 游戏状态
	gameState  *protocol.GameState // 游戏状态
	playerID   string              // 玩家 ID
	playerName string              // 玩家名称
	serverURL  string              // 服务器地址

	// 动作输入
	actionInput   string // 加注金额输入
	minRaise      int    // 最小加注金额
	maxRaise      int    // 最大加注金额
	currentBet    int    // 当前最高下注
	isYourTurn    bool   // 是否轮到玩家
	timeLeft      int    // 剩余时间

	// 摊牌结果
	showdown *protocol.Showdown // 摊牌结果

	// 游戏结算
	gameResult      *protocol.Showdown // 游戏结算结果
	finalChips      int                // 最终筹码
	initialChips    int                // 初始筹码
	chipsWon        int                // 赢得筹码
	gameWon         bool               // 是否获胜

	// 结算后准备状态
	readyPlayers    []string // 已准备好的玩家名称列表
	totalPlayers    int      // 总玩家数
	selfReady       bool     // 自己是否已准备
	resultChoice    int      // 结算屏幕选择：0=下一局，1=退出

	// 聊天
	chatModel *components.ChatModel // 聊天组件

	// 通知消息
	notifications []string // 通知消息列表

	// 错误
	err error // 错误信息

	// 连接状态
	connected bool // 是否已连接

	// 终端窗口尺寸
	winWidth  int // 终端宽度
	winHeight int // 终端高度

	// 外部消息通道（用于从 WebSocket 回调接收消息）
	extMsgChan chan tea.Msg // 外部消息通道
}

// NewModel 创建新的 TUI 模型
func NewModel() *Model {
	// 创建聊天组件
	chat := components.NewChatModel()

	return &Model{
		screen:         ScreenConnect,
		serverInput:    "localhost:8080",
		playerInput:    "Player",
		connectField:   0,
		connecting:     false,
		actionInput:    "",
		notifications:  make([]string, 0),
		chatModel:      chat,
		connected:      false,
		extMsgChan:     make(chan tea.Msg, 100), // 缓冲通道
	}
}

// ==================== Bubble Tea 接口实现 ====================

// Init 初始化模型
func (m *Model) Init() tea.Cmd {
	// 启动监听外部消息的循环，但不阻塞键盘输入
	// 使用 tick 来定期检查通道
	return m.tick()
}

// tickMsg 自定义的 tick 消息，用于持续检查外部通道
type tickMsg time.Time

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

// waitForExtMsg 等待外部消息
func (m *Model) waitForExtMsg() tea.Cmd {
	return func() tea.Msg {
		return <-m.extMsgChan
	}
}

// Update 更新模型状态
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 先处理聊天组件更新（如果聊天可见）
	if m.screen == ScreenChat {
		var cmd tea.Cmd
		m.chatModel, cmd = m.chatModel.Update(msg)
		if cmd != nil {
			return m, tea.Batch(cmd, m.tick())
		}
	}

	// 处理不同类型的消息
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tickMsg:
		// tick 消息，继续检查通道
		return m, m.tick()

	case tea.WindowSizeMsg:
		// 记录窗口尺寸，用于 View 填充保证每次输出行数一致
		m.winWidth = msg.Width
		m.winHeight = msg.Height
		return m, m.tick()

	// 服务器消息
	case ConnectedMsg:
		m.connected = true
		m.connecting = false
		return m, m.tick()

	case DisconnectedMsg:
		m.connected = false
		m.connecting = false
		m.err = fmt.Errorf("与服务器断开连接")
		m.screen = ScreenConnect
		return m, m.tick()

	case JoinAckResultMsg:
		if msg.Success {
			m.playerID = msg.PlayerID
			m.addNotification(fmt.Sprintf("加入成功! 座位: %d", msg.Seat+1))
			m.screen = ScreenLobby
		} else {
			m.connecting = false
			m.err = fmt.Errorf("加入游戏失败")
		}
		return m, m.tick()

	case GameStateMsg:
		m.gameState = msg.State
		// 如果在大厅屏幕，收到游戏状态则切换到游戏屏幕
		if m.screen == ScreenLobby {
			m.screen = ScreenGame
		}
		// 如果在摊牌或结算屏幕，收到新游戏状态（说明新局开始了）则返回游戏屏幕
		if m.screen == ScreenShowdown || m.screen == ScreenResult {
			m.screen = ScreenGame
			// 重置准备状态
			m.selfReady = false
			m.readyPlayers = nil
		}
		return m, m.tick()

	case ActionSentMsg:
		// 动作已成功发送到服务器，重置行动标志
		// 如果服务器拒绝动作，会重新发送 YourTurn 恢复状态
		m.isYourTurn = false
		return m, m.tick()

	case YourTurnMsg:
		m.isYourTurn = true
		m.currentBet = msg.Turn.CurrentBet
		m.minRaise = msg.Turn.MinAction
		m.maxRaise = msg.Turn.MaxAction
		m.timeLeft = msg.Turn.TimeLeft
		m.addNotification("轮到你行动了!")
		return m, m.tick()

	case PlayerJoinedMsg:
		m.addNotification(fmt.Sprintf("玩家 %s 加入了游戏 (座位 %d)",
			msg.Player.Name, msg.Player.Seat+1))
		return m, m.tick()

	case PlayerLeftMsg:
		m.addNotification(fmt.Sprintf("玩家 %s 离开了游戏", msg.PlayerName))
		return m, m.tick()

	case PlayerActedMsg:
		actionText := getActionText(msg.Action)
		if msg.Amount > 0 {
			m.addNotification(fmt.Sprintf("%s %s %d", msg.PlayerName, actionText, msg.Amount))
		} else {
			m.addNotification(fmt.Sprintf("%s %s", msg.PlayerName, actionText))
		}
		return m, m.tick()

	case ShowdownMsg:
		m.showdown = msg.Showdown
		m.gameResult = msg.Showdown
		m.isYourTurn = false

		// 计算玩家的筹码变化
		m.gameWon = false
		m.chipsWon = 0
		m.finalChips = 0

		for _, p := range msg.Showdown.AllPlayers {
			// 通过玩家名称匹配（因为 PlayerDetail 中没有 PlayerID）
			if p.PlayerName == m.playerName {
				m.chipsWon = p.WonAmount
				m.finalChips = p.ChipsAfter
				m.gameWon = p.IsWinner
				break
			}
		}

		// 重置准备状态
		m.selfReady = false
		m.readyPlayers = nil
		m.totalPlayers = len(msg.Showdown.AllPlayers)
		m.resultChoice = 0

		// 先显示摊牌屏幕
		m.screen = ScreenShowdown
		return m, m.tick()

	case PlayerReadyMsg:
		// 更新准备状态
		m.readyPlayers = msg.Notify.ReadyPlayers
		m.totalPlayers = msg.Notify.TotalPlayers

		if msg.Notify.AllReady {
			m.addNotification("所有玩家已准备，开始下一局!")
		} else {
			m.addNotification(fmt.Sprintf("玩家 %s 已准备 (%d/%d)",
				msg.Notify.PlayerName, len(msg.Notify.ReadyPlayers), msg.Notify.TotalPlayers))
		}
		return m, m.tick()

	case ChatMsg:
		// 添加聊天消息
		if msg.Message.IsSystem {
			m.chatModel.AddSystemMessage(msg.Message.Content)
		} else {
			m.chatModel.AddMessage(msg.Message.PlayerID, msg.Message.PlayerName, msg.Message.Content)
		}
		return m, m.tick()

	case ErrorMsg:
		m.err = msg.Err
		return m, m.tick()
	}

	// 默认继续等待外部消息
	return m, m.tick()
}

// View 渲染视图
func (m *Model) View() string {
	var content string
	switch m.screen {
	case ScreenConnect:
		content = m.viewConnect()
	case ScreenLobby:
		content = m.viewLobby()
	case ScreenGame:
		content = m.viewGame()
	case ScreenAction:
		content = m.viewAction()
	case ScreenShowdown:
		content = m.viewShowdown()
	case ScreenResult:
		content = m.viewResult()
	case ScreenChat:
		content = m.viewChat()
	default:
		content = "未知屏幕"
	}

	// 用空行填充到终端高度，防止旧内容残留（Bubble Tea 渲染行数减少时不会清除多余行）
	return m.padToWindowHeight(content)
}

// padToWindowHeight 将内容填充到终端窗口高度，避免渲染残留
func (m *Model) padToWindowHeight(content string) string {
	if m.winHeight <= 0 {
		return content
	}
	lines := strings.Count(content, "\n") + 1
	if lines < m.winHeight {
		content += strings.Repeat("\n", m.winHeight-lines)
	}
	return content
}

// ==================== 键盘消息处理 ====================

// handleKeyMsg 处理键盘消息
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 全局快捷键
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// 根据当前屏幕分发处理
	switch m.screen {
	case ScreenConnect:
		return m.updateConnect(msg)
	case ScreenLobby:
		return m.updateLobby(msg)
	case ScreenGame:
		return m.updateGame(msg)
	case ScreenAction:
		return m.updateAction(msg)
	case ScreenShowdown:
		return m.updateShowdown(msg)
	case ScreenResult:
		return m.updateResult(msg)
	case ScreenChat:
		return m.updateChat(msg)
	}

	return m, m.tick()
}

// ==================== 连接屏幕 ====================

// updateConnect 更新连接屏幕
func (m *Model) updateConnect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// 尝试连接
		if !m.connecting {
			m.connecting = true
			m.err = nil
			cmd := m.doConnect()
			return m, tea.Batch(cmd, m.tick())
		}
		return m, m.tick()

	case "tab":
		// 切换输入框
		m.connectField = (m.connectField + 1) % 2
		return m, m.tick()

	case "up", "shift+tab":
		// 上一个输入框
		m.connectField = (m.connectField - 1 + 2) % 2
		return m, m.tick()

	case "backspace":
		// 删除字符
		if m.connectField == 0 {
			if len(m.serverInput) > 0 {
				m.serverInput = m.serverInput[:len(m.serverInput)-1]
			}
		} else {
			if len(m.playerInput) > 0 {
				m.playerInput = m.playerInput[:len(m.playerInput)-1]
			}
		}
		return m, m.tick()

	default:
		// 输入字符
		if len(msg.Runes) > 0 {
			ch := msg.Runes[0]
			if m.connectField == 0 {
				// 服务器地址输入
				m.serverInput += string(ch)
			} else {
				// 玩家名称输入
				if len(m.playerInput) < 20 { // 限制长度
					m.playerInput += string(ch)
				}
			}
		}
		return m, m.tick()
	}
}

// doConnect 执行连接
func (m *Model) doConnect() tea.Cmd {
	// 创建客户端配置
	config := &client.Config{
		ServerURL:   "ws://" + m.serverInput,
		PlayerName:  m.playerInput,
		OnJoinAck: func(success bool, playerID string, seat int) {
			// 加入确认回调
			m.extMsgChan <- JoinAckResultMsg{
				Success:  success,
				PlayerID: playerID,
				Seat:     seat,
			}
		},
		OnStateChange: func(state *protocol.GameState) {
			// 通过通道发送消息到 Bubble Tea
			m.extMsgChan <- GameStateMsg{State: state}
		},
		OnTurn: func(turn *protocol.YourTurn) {
			m.extMsgChan <- YourTurnMsg{Turn: turn}
		},
		OnShowdown: func(showdown *protocol.Showdown) {
			m.extMsgChan <- ShowdownMsg{Showdown: showdown}
		},
		OnPlayerReady: func(notify *protocol.PlayerReadyNotify) {
			m.extMsgChan <- PlayerReadyMsg{Notify: notify}
		},
		OnChat: func(chatMsg *protocol.ChatMessage) {
			m.extMsgChan <- ChatMsg{Message: chatMsg}
		},
		OnError: func(err error) {
			m.extMsgChan <- ErrorMsg{Err: err}
		},
		OnConnect: func() {
			m.extMsgChan <- ConnectedMsg{}
		},
		OnDisconnect: func() {
			m.extMsgChan <- DisconnectedMsg{}
		},
	}

	// 创建客户端
	m.client = client.NewClient(config)
	m.serverURL = config.ServerURL
	m.playerName = config.PlayerName

	// 返回一个 Cmd 执行连接
	return func() tea.Msg {
		if err := m.client.Connect(); err != nil {
			m.connecting = false
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

// viewConnect 渲染连接屏幕
func (m *Model) viewConnect() string {
	var content strings.Builder

	// 标题
	content.WriteString(styleTitle.Render("Texas Hold'em Poker"))
	content.WriteString("\n\n")

	// 服务器地址输入框
	serverLabel := "服务器地址:"
	if m.connectField == 0 {
		serverLabel = styleActive.Render(serverLabel)
	} else {
		serverLabel = styleInactive.Render(serverLabel)
	}
	serverInput := m.serverInput
	if m.connectField == 0 {
		serverInput = styleInput.Render(serverInput + " ")
	} else {
		serverInput = styleInput.Render(serverInput)
	}
	content.WriteString(fmt.Sprintf("%s %s\n\n", serverLabel, serverInput))

	// 玩家名称输入框
	playerLabel := "玩家名称:"
	if m.connectField == 1 {
		playerLabel = styleActive.Render(playerLabel)
	} else {
		playerLabel = styleInactive.Render(playerLabel)
	}
	playerInput := m.playerInput
	if m.connectField == 1 {
		playerInput = styleInput.Render(playerInput + " ")
	} else {
		playerInput = styleInput.Render(playerInput)
	}
	content.WriteString(fmt.Sprintf("%s %s\n\n", playerLabel, playerInput))

	// 连接状态
	if m.connecting {
		content.WriteString(styleSubtitle.Render("正在连接..."))
	} else if m.err != nil {
		content.WriteString(styleError.Render(fmt.Sprintf("错误: %v", m.err)))
	} else {
		content.WriteString(styleSubtitle.Render("按 Enter 连接，Tab 切换输入框"))
	}
	content.WriteString("\n\n")

	// 底部提示
	content.WriteString(styleInactive.Render("Ctrl+C: 退出"))

	return lipgloss.Place(
		60, 20,
		lipgloss.Center, lipgloss.Center,
		styleBox.Render(content.String()),
	)
}

// ==================== 大厅屏幕 ====================

// updateLobby 更新大厅屏幕
func (m *Model) updateLobby(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// 进入游戏（实际上游戏由服务器控制）
		m.addNotification("等待游戏开始...")
		return m, m.tick()

	case "h":
		// 打开聊天
		m.chatModel.SetVisible(true)
		m.screen = ScreenChat
		return m, m.tick()
	}

	return m, m.tick()
}

// viewLobby 渲染大厅屏幕
func (m *Model) viewLobby() string {
	var content strings.Builder

	// 标题
	content.WriteString(styleTitle.Render("Texas Hold'em Poker"))
	content.WriteString("\n\n")

	// 玩家信息
	content.WriteString(styleSubtitle.Render("玩家信息"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("  ID: %s\n", m.playerID))
	content.WriteString(fmt.Sprintf("  名称: %s\n", m.playerName))
	content.WriteString("\n")

	// 游戏状态
	if m.gameState != nil && len(m.gameState.Players) > 0 {
		content.WriteString(styleSubtitle.Render("已连接玩家:"))
		content.WriteString("\n")
		for _, p := range m.gameState.Players {
			status := "游戏中"
			if p.Status == models.PlayerStatusInactive {
				status = "未入座"
			}
			marker := " "
			if p.ID == m.playerID {
				marker = "★"
			}
			content.WriteString(fmt.Sprintf("  %s [%d] %s - %s\n",
				marker, p.Seat+1, p.Name, status))
		}
	} else {
		content.WriteString(styleInactive.Render("等待其他玩家加入..."))
	}
	content.WriteString("\n")

	// 提示
	content.WriteString(styleSubtitle.Render("等待游戏开始..."))
	content.WriteString("\n\n")

	// 快捷键提示
	content.WriteString(styleInactive.Render("[H] 聊天  [Ctrl+C] 退出"))

	return lipgloss.Place(
		60, 25,
		lipgloss.Center, lipgloss.Center,
		styleBox.Render(content.String()),
	)
}

// ==================== 游戏屏幕 ====================

// updateGame 更新游戏屏幕
func (m *Model) updateGame(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		// 弃牌
		return m, tea.Batch(m.sendAction(models.ActionFold, 0), m.tick())

	case "c", "k":
		// 跟注/过牌 - 需要计算实际发送的动作类型
		toCall := m.calculateToCall()
		if toCall > 0 {
			// 需要跟注
			return m, tea.Batch(m.sendAction(models.ActionCall, 0), m.tick())
		} else {
			// 可以过牌
			return m, tea.Batch(m.sendAction(models.ActionCheck, 0), m.tick())
		}

	case "r":
		// 加注
		if m.isYourTurn {
			m.actionInput = ""
			m.screen = ScreenAction
		} else {
			m.addNotification("还没轮到你")
		}
		return m, m.tick()

	case "a":
		// 全下
		return m, tea.Batch(m.sendAction(models.ActionAllIn, 0), m.tick())

	case "h":
		// 聊天
		m.chatModel.SetVisible(true)
		m.screen = ScreenChat
		return m, m.tick()

	case "q":
		// 退出
		return m, tea.Quit
	}

	return m, m.tick()
}

// sendAction 发送玩家动作
// 发送成功返回 ActionSentMsg，由 Update 中安全地重置 isYourTurn
// 如果服务器拒绝动作，会重新发送 YourTurn 恢复行动状态
func (m *Model) sendAction(action models.ActionType, amount int) tea.Cmd {
	return func() tea.Msg {
		if err := m.client.SendPlayerAction(action, amount); err != nil {
			return ErrorMsg{Err: err}
		}
		return ActionSentMsg{}
	}
}

// viewGame 渲染游戏屏幕
func (m *Model) viewGame() string {
	var content strings.Builder

	// 标题栏
	title := styleTitle.Render("Texas Hold'em Poker")
	playerInfo := styleSubtitle.Render(fmt.Sprintf("玩家: %s", m.playerName))
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, title, "    ", playerInfo))
	content.WriteString("\n")

	// 游戏状态行
	if m.gameState != nil {
		stage := m.gameState.Stage.String()
		pot := stylePot.Render(fmt.Sprintf("底池: %d", m.gameState.Pot))
		dealer := fmt.Sprintf("庄家: [%d]", m.gameState.DealerButton+1)

		content.WriteString(styleBox.Render(
			lipgloss.JoinHorizontal(lipgloss.Top,
				fmt.Sprintf(" 阶段: %s ", stage),
				"    ",
				pot,
				"    ",
				dealer,
			),
		))
		content.WriteString("\n\n")
	}

	// 公共牌
	content.WriteString(m.renderCommunityCards())
	content.WriteString("\n\n")

	// 玩家列表
	content.WriteString(m.renderPlayers())
	content.WriteString("\n")

	// 通知消息
	if len(m.notifications) > 0 {
		content.WriteString(styleSubtitle.Render("────────────────────────────────────"))
		content.WriteString("\n")
		for _, msg := range m.notifications {
			content.WriteString(styleHighlight.Render("  " + msg))
			content.WriteString("\n")
		}
		// 清空通知
		m.notifications = make([]string, 0)
	}

	// 动作提示
	content.WriteString("\n")
	content.WriteString(m.renderActionPrompt())

	return content.String()
}

// renderCommunityCards 渲染公共牌
func (m *Model) renderCommunityCards() string {
	if m.gameState == nil {
		return styleSubtitle.Render("公共牌: 等待发牌")
	}

	var cards []card.Card
	for _, c := range m.gameState.CommunityCards {
		if c.Rank != 0 {
			cards = append(cards, c)
		}
	}

	if len(cards) == 0 {
		return styleSubtitle.Render("公共牌: [  ?  ] [  ?  ] [  ?  ] [  ?  ] [  ?  ]")
	}

	return styleSubtitle.Render("公共牌: ") + components.RenderCardsCompact(cards, true)
}

// renderPlayers 渲染玩家列表
func (m *Model) renderPlayers() string {
	if m.gameState == nil || len(m.gameState.Players) == 0 {
		return styleSubtitle.Render("玩家: 等待加入...")
	}

	var parts []string
	parts = append(parts, styleSubtitle.Render("玩家:"))
	parts = append(parts, "\n")

	for _, p := range m.gameState.Players {
		// 玩家状态
		status := ""
		switch p.Status {
		case models.PlayerStatusActive:
			if p.IsSelf {
				status = styleActive.Render("【你】")
			} else {
				status = styleInactive.Render("    ")
			}
		case models.PlayerStatusFolded:
			status = styleInactive.Render("已弃牌")
		case models.PlayerStatusAllIn:
			status = styleHighlight.Render("全下  ")
		default:
			status = styleInactive.Render("    ")
		}

		// 是否当前玩家
		current := ""
		if m.gameState != nil && m.gameState.CurrentPlayer == p.Seat {
			current = styleCurrentPlayer.Render("◀")
		}

		// 是否庄家
		dealer := ""
		if p.IsDealer {
			dealer = "[D]"
		}

		// 玩家信息行
		line := fmt.Sprintf("  [%d] %s %-10s 筹码:%-4d 下注:%-3d %s %s",
			p.Seat+1, dealer, p.Name, p.Chips, p.CurrentBet, status, current)
		parts = append(parts, line)

		// 底牌行
		if p.IsSelf && p.HoleCards[0].Rank != 0 {
			holeCards := components.RenderCardsCompact(p.HoleCards[:], true)
			parts = append(parts, fmt.Sprintf("      底牌: %s", holeCards))
		} else if p.HoleCards[0].Rank != 0 {
			// 其他玩家的底牌（只在摊牌时显示）
			holeCards := components.RenderCardsCompact(p.HoleCards[:], true)
			parts = append(parts, fmt.Sprintf("      底牌: %s", holeCards))
		} else {
			parts = append(parts, "      底牌: [??] [??]")
		}
		parts = append(parts, "\n")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

// renderActionPrompt 渲染动作提示栏
func (m *Model) renderActionPrompt() string {
	var content strings.Builder

	// 计算玩家需要跟注的金额
	toCall := 0
	if m.gameState != nil {
		for _, p := range m.gameState.Players {
			if p.IsSelf {
				toCall = m.gameState.CurrentBet - p.CurrentBet
				break
			}
		}
	}

	// 判断是否可以过牌（无需跟注）
	canCheck := toCall == 0
	sep := "  " // 按钮间距

	if m.isYourTurn {
		// 轮到自己，显示状态提示
		content.WriteString(styleCurrentPlayer.Render("▶ 轮到你行动"))
		content.WriteString("\n\n")

		// 游戏操作按钮（带颜色区分）
		var gameActions []string
		gameActions = append(gameActions, styleBtnFold.Render(" F 弃牌 "))
		if canCheck {
			gameActions = append(gameActions, styleBtnCall.Render(" K 过牌 "))
		} else {
			gameActions = append(gameActions, styleBtnCall.Render(fmt.Sprintf(" C 跟注 %d ", toCall)))
		}
		gameActions = append(gameActions, styleBtnRaise.Render(" R 加注 "))
		gameActions = append(gameActions, styleBtnAllIn.Render(" A 全下 "))
		content.WriteString(strings.Join(gameActions, sep))
	} else {
		// 未轮到自己，灰色显示
		content.WriteString(styleInactive.Render("  等待对手行动..."))
		content.WriteString("\n\n")

		var gameActions []string
		gameActions = append(gameActions, styleBtnDisabled.Render(" F 弃牌 "))
		if canCheck {
			gameActions = append(gameActions, styleBtnDisabled.Render(" K 过牌 "))
		} else {
			gameActions = append(gameActions, styleBtnDisabled.Render(fmt.Sprintf(" C 跟注 %d ", toCall)))
		}
		gameActions = append(gameActions, styleBtnDisabled.Render(" R 加注 "))
		gameActions = append(gameActions, styleBtnDisabled.Render(" A 全下 "))
		content.WriteString(strings.Join(gameActions, sep))
	}

	// 功能按钮单独一行
	content.WriteString("\n\n")
	funcActions := []string{
		styleBtnFunc.Render(" H 聊天 "),
		styleBtnFunc.Render(" Q 退出 "),
	}
	content.WriteString(strings.Join(funcActions, sep))

	// 快捷键提示
	content.WriteString("\n")
	content.WriteString(styleInactive.Render("[↑/↓] 选择  [Enter] 确认  [Q] 退出"))

	return styleActionBar.Render(content.String())
}

// ==================== 动作屏幕 ====================

// updateAction 更新动作屏幕
func (m *Model) updateAction(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// 确认加注
		var amount int
		fmt.Sscanf(m.actionInput, "%d", &amount)
		if amount > 0 {
			m.screen = ScreenGame
			return m, tea.Batch(m.sendAction(models.ActionRaise, amount), m.tick())
		}
		return m, m.tick()

	case "esc":
		// 取消
		m.screen = ScreenGame
		return m, m.tick()

	case "backspace":
		// 删除字符
		if len(m.actionInput) > 0 {
			m.actionInput = m.actionInput[:len(m.actionInput)-1]
		}
		return m, m.tick()

	default:
		// 输入数字
		if len(msg.Runes) > 0 {
			ch := msg.Runes[0]
			if ch >= '0' && ch <= '9' {
				m.actionInput += string(ch)
			}
		}
	}

	return m, m.tick()
}

// viewAction 渲染动作屏幕
func (m *Model) viewAction() string {
	var content strings.Builder

	// 标题
	content.WriteString(styleTitle.Render("加注"))
	content.WriteString("\n\n")

	// 提示
	content.WriteString(styleSubtitle.Render("请输入加注金额:"))
	content.WriteString("\n\n")

	// 输入框
	content.WriteString(styleButtonActive.Render(" " + m.actionInput + " "))
	content.WriteString("\n\n")

	// 限制提示
	if m.minRaise > 0 || m.maxRaise > 0 {
		content.WriteString(styleSubtitle.Render(fmt.Sprintf("最小: %d  最大: %d", m.minRaise, m.maxRaise)))
		content.WriteString("\n\n")
	}

	// 确认提示
	content.WriteString(styleInactive.Render("[Enter] 确认  [Esc] 取消"))

	return lipgloss.Place(
		40, 15,
		lipgloss.Center, lipgloss.Center,
		styleBox.Render(content.String()),
	)
}

// ==================== 摊牌屏幕 ====================

// updateShowdown 更新摊牌屏幕
func (m *Model) updateShowdown(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc", "q", " ":
		// 进入结算屏幕
		m.screen = ScreenResult
		return m, m.tick()
	}

	return m, m.tick()
}

// viewShowdown 渲染摊牌屏幕
func (m *Model) viewShowdown() string {
	var content strings.Builder

	// 标题
	content.WriteString(styleTitle.Render("摊牌结果"))
	content.WriteString("\n\n")

	if m.showdown == nil {
		content.WriteString(styleInactive.Render("等待摊牌..."))
		return lipgloss.Place(
			60, 15,
			lipgloss.Center, lipgloss.Center,
			styleBox.Render(content.String()),
		)
	}

	// 底池
	content.WriteString(stylePot.Render(fmt.Sprintf("总底池: %d", m.showdown.Pot)))
	content.WriteString("\n\n")

	// 公共牌
	if len(m.showdown.CommunityCards) > 0 {
		var cards []card.Card
		for _, c := range m.showdown.CommunityCards {
			if c.Rank != 0 {
				cards = append(cards, c)
			}
		}
		if len(cards) > 0 {
			content.WriteString(styleSubtitle.Render("公共牌: "))
			content.WriteString(components.RenderCardsCompact(cards, true))
			content.WriteString("\n\n")
		}
	}

	// 获胜者列表
	if len(m.showdown.Winners) > 0 {
		content.WriteString(styleHighlight.Render("获胜者:"))
		content.WriteString("\n")
		for _, w := range m.showdown.Winners {
			handName := w.HandName
			if handName == "" {
				handName = "高牌"
			}
			content.WriteString(styleActive.Render(fmt.Sprintf("  ★ %s 以 [%s] 赢得 %d\n",
				w.PlayerName, handName, w.WonChips)))
		}
		content.WriteString("\n")
	}

	// 所有玩家详情
	content.WriteString(styleSubtitle.Render("玩家详情:"))
	content.WriteString("\n")
	content.WriteString(m.renderPlayerDetails(m.showdown.AllPlayers, ""))
	content.WriteString("\n")
	content.WriteString(styleInactive.Render("[Enter] 查看结算"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(70).
		Render(content.String())
}

// ==================== 结算屏幕 ====================

// updateResult 更新结算屏幕
func (m *Model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// 切换到上一个选项
		if m.resultChoice > 0 {
			m.resultChoice--
		}
		return m, m.tick()

	case "down", "j":
		// 切换到下一个选项
		if m.resultChoice < 1 {
			m.resultChoice++
		}
		return m, m.tick()

	case "enter", " ":
		if m.resultChoice == 0 {
			// 选择"下一局" - 发送准备请求
			if !m.selfReady {
				m.selfReady = true
				m.addNotification("已准备，等待其他玩家...")
				return m, tea.Batch(m.sendReadyForNext(), m.tick())
			}
			return m, m.tick()
		}
		// 选择"退出"
		return m, tea.Quit

	case "q":
		// 退出游戏
		return m, tea.Quit
	}

	return m, m.tick()
}

// sendReadyForNext 发送准备下一局请求
func (m *Model) sendReadyForNext() tea.Cmd {
	return func() tea.Msg {
		if err := m.client.SendReadyForNext(); err != nil {
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

// viewResult 渲染结算屏幕
func (m *Model) viewResult() string {
	var content strings.Builder

	// 标题
	if m.gameWon {
		content.WriteString(styleHighlight.Render("★ 本局获胜! ★"))
	} else {
		content.WriteString(styleSubtitle.Render("本局结束"))
	}
	content.WriteString("\n\n")

	// 筹码变化
	content.WriteString(stylePot.Render(fmt.Sprintf("筹码变化: %+d", m.chipsWon)))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("最终筹码: %d", m.finalChips))
	content.WriteString("\n\n")

	// 获胜者列表
	if m.gameResult != nil && len(m.gameResult.Winners) > 0 {
		content.WriteString(styleHighlight.Render("获胜者:"))
		content.WriteString("\n")
		for _, w := range m.gameResult.Winners {
			handName := w.HandName
			if handName == "" {
				handName = "高牌"
			}
			isYou := ""
			if w.PlayerName == m.playerName {
				isYou = " (你)"
			}
			content.WriteString(styleActive.Render(fmt.Sprintf("  ★ %s%s 以 [%s] 赢得 %d\n",
				w.PlayerName, isYou, handName, w.WonChips)))
		}
		content.WriteString("\n")
	}

	// 所有玩家详情
	if m.gameResult != nil {
		content.WriteString(styleSubtitle.Render("玩家详情:"))
		content.WriteString("\n")
		content.WriteString(m.renderPlayerDetails(m.gameResult.AllPlayers, m.playerName))
	}

	content.WriteString("\n")
	content.WriteString(styleSubtitle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	content.WriteString("\n")

	// 准备状态显示
	content.WriteString(m.renderReadyStatus())
	content.WriteString("\n")

	// 选择菜单
	content.WriteString(m.renderResultMenu())

	return lipgloss.Place(
		70, 35,
		lipgloss.Center, lipgloss.Center,
		styleBox.Render(content.String()),
	)
}

// renderReadyStatus 渲染玩家准备状态
func (m *Model) renderReadyStatus() string {
	var content strings.Builder

	if len(m.readyPlayers) > 0 {
		content.WriteString(styleSubtitle.Render(fmt.Sprintf("  准备状态 (%d/%d):",
			len(m.readyPlayers), m.totalPlayers)))
		content.WriteString("\n")

		// 显示已准备的玩家
		for _, name := range m.readyPlayers {
			isYou := ""
			if name == m.playerName {
				isYou = " (你)"
			}
			content.WriteString(styleActive.Render(fmt.Sprintf("    ✓ %s%s", name, isYou)))
			content.WriteString("\n")
		}
	}

	return content.String()
}

// renderResultMenu 渲染结算屏幕选择菜单
func (m *Model) renderResultMenu() string {
	var content strings.Builder

	// 选项 0: 下一局
	nextLabel := "下一局"
	if m.selfReady {
		nextLabel = "已准备，等待其他玩家..."
	}
	if m.resultChoice == 0 {
		content.WriteString(styleButtonActive.Render(fmt.Sprintf(" ▸ %s ", nextLabel)))
	} else {
		content.WriteString(styleButton.Render(fmt.Sprintf("   %s ", nextLabel)))
	}
	content.WriteString("\n\n")

	// 选项 1: 退出
	if m.resultChoice == 1 {
		content.WriteString(styleButtonActive.Render(" ▸ 退出游戏 "))
	} else {
		content.WriteString(styleButton.Render("   退出游戏 "))
	}
	content.WriteString("\n\n")

	// 快捷键提示
	content.WriteString(styleInactive.Render("[↑/↓] 选择  [Enter] 确认  [Q] 退出"))

	return content.String()
}

// ==================== 聊天屏幕 ====================

// updateChat 更新聊天屏幕
func (m *Model) updateChat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// 关闭聊天
		m.chatModel.SetVisible(false)
		m.screen = ScreenGame
		return m, m.tick()

	case "enter":
		// 发送消息
		input := m.chatModel.GetInputValue()
		if input != "" {
			// 发送聊天消息
			_ = m.client.SendChat(input)
			m.chatModel.ClearInput()
		}
		return m, m.tick()
	}

	return m, m.tick()
}

// viewChat 渲染聊天屏幕
func (m *Model) viewChat() string {
	return m.chatModel.View()
}

// ==================== 辅助方法 ====================

// calculateToCall 计算需要跟注的金额
func (m *Model) calculateToCall() int {
	if m.gameState == nil {
		return 0
	}
	for _, p := range m.gameState.Players {
		if p.IsSelf {
			return m.gameState.CurrentBet - p.CurrentBet
		}
	}
	return 0
}

// addNotification 添加通知消息
func (m *Model) addNotification(msg string) {
	m.notifications = append(m.notifications, msg)
	// 限制通知数量
	if len(m.notifications) > 5 {
		m.notifications = m.notifications[len(m.notifications)-5:]
	}
}

// SetClient 设置客户端（用于从 main.go 传入）
func (m *Model) SetClient(client *client.Client) {
	m.client = client
}

// GetClient 获取客户端
func (m *Model) GetClient() *client.Client {
	return m.client
}

// GetExtMsgChan 获取外部消息通道（用于 WebSocket 回调）
func (m *Model) GetExtMsgChan() chan tea.Msg {
	return m.extMsgChan
}

// SetGameState 设置游戏状态
func (m *Model) SetGameState(state *protocol.GameState) {
	m.gameState = state
}

// GetGameState 获取游戏状态
func (m *Model) GetGameState() *protocol.GameState {
	return m.gameState
}

// renderPlayerDetails 渲染玩家结算详情（摊牌屏幕和结算屏幕共用）
// selfName 为空时不标记"你"，非空时用于匹配当前玩家并添加标记
func (m *Model) renderPlayerDetails(players []protocol.ShowdownPlayerDetail, selfName string) string {
	var result strings.Builder
	indent := "    " // 统一缩进（4空格，与标记对齐）

	for _, p := range players {
		// 第一行：标记 + 玩家名 + 状态/底牌
		marker := "  "
		if p.IsWinner {
			marker = styleHighlight.Render("★ ")
		}

		selfTag := ""
		if selfName != "" && p.PlayerName == selfName {
			selfTag = styleCurrentPlayer.Render(" ◀")
		}

		if p.IsFolded {
			// 弃牌玩家：标记 + 名字 + [已弃牌]
			result.WriteString(fmt.Sprintf("  %s%-10s %s%s\n",
				marker,
				p.PlayerName,
				styleInactive.Render("[已弃牌]"),
				selfTag))
		} else {
			// 未弃牌玩家：标记 + 名字 + 底牌 + 牌型
			holeCards := components.RenderCardsCompact(p.HoleCards[:], true)
			handName := p.HandName
			if handName == "" {
				handName = "-"
			}
			result.WriteString(fmt.Sprintf("  %s%-10s 底牌: %s  牌型: %s%s\n",
				marker,
				p.PlayerName,
				holeCards,
				styleAction.Render(handName),
				selfTag))
		}

		// 第二行：筹码变化（统一缩进对齐）
		if p.WonAmount > 0 {
			result.WriteString(styleActive.Render(fmt.Sprintf("%s赢得 +%d (剩余: %d)", indent, p.WonAmount, p.ChipsAfter)))
		} else if p.WonAmount < 0 {
			result.WriteString(styleWarning.Render(fmt.Sprintf("%s输掉 %d", indent, -p.WonAmount)) +
				styleInactive.Render(fmt.Sprintf(" (剩余: %d)", p.ChipsAfter)))
		} else {
			result.WriteString(styleInactive.Render(fmt.Sprintf("%s筹码不变 (剩余: %d)", indent, p.ChipsAfter)))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// getActionText 获取动作描述文本
func getActionText(action models.ActionType) string {
	switch action {
	case models.ActionFold:
		return "弃牌"
	case models.ActionCheck:
		return "看牌"
	case models.ActionCall:
		return "跟注"
	case models.ActionRaise:
		return "加注"
	case models.ActionAllIn:
		return "全下"
	default:
		return "未知"
	}
}
