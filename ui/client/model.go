package client

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/protocol"
	"github.com/wilenwang/just_play/Texas-Holdem/server/client"
)

// 样式定义
var (
	styleTitle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6")).MarginBottom(1)
	styleSubtitle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	styleBox        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)
	styleActive     = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	styleInactive   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	styleWarning    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	stylePot        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1FA8C"))
	styleButton     = lipgloss.NewStyle().Background(lipgloss.Color("#44475A")).Foreground(lipgloss.Color("#F8F8F2")).Padding(0, 2)
	styleButtonActive = lipgloss.NewStyle().Background(lipgloss.Color("#FF79C6")).Foreground(lipgloss.Color("#F8F8F2")).Padding(0, 2)
	styleAction     = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
	styleChat       = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9"))
)

// ScreenType 表示当前屏幕
type ScreenType int

const (
	ScreenConnect ScreenType = iota
	ScreenLobby
	ScreenGame
	ScreenAction
	ScreenChat
)

// Model TUI 模型
type Model struct {
	client       *client.Client
	screen       ScreenType
	gameState    *protocol.GameState
	playerID     string
	playerName   string
	serverURL    string
	actionAmount int
	chatInput    string
	messages     []string
	selectedMenu int
	err          error
	connected    bool
}

// NewModel 创建新的 TUI 模型
func NewModel() *Model {
	return &Model{
		screen:     ScreenConnect,
		messages:   make([]string, 0),
		actionAmount: 0,
	}
}

// Init 初始化
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenConnect:
		return m.updateConnect(msg)

	case ScreenLobby:
		return m.updateLobby(msg)

	case ScreenGame:
		return m.updateGame(msg)

	case ScreenAction:
		return m.updateAction(msg)

	case ScreenChat:
		return m.updateChat(msg)
	}

	return m, nil
}

// updateConnect 更新连接屏幕
func (m *Model) updateConnect(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			// 尝试连接
			m.client = client.NewClient(&client.Config{
				ServerURL:  m.serverURL,
				PlayerName: m.playerName,
				OnConnect: func() {
					m.connected = true
				},
				OnDisconnect: func() {
					m.connected = false
				},
				OnError: func(err error) {
					m.err = err
				},
			})
			if err := m.client.Connect(); err != nil {
				m.err = err
			} else {
				m.screen = ScreenLobby
			}

		case "backspace":
			if len(m.playerName) > 0 {
				m.playerName = m.playerName[:len(m.playerName)-1]
			}
		default:
			if len(msg.Runes) > 0 {
				m.playerName += string(msg.Runes[0])
			}
		}
	}

	return m, nil
}

// updateLobby 更新大厅屏幕
func (m *Model) updateLobby(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			m.screen = ScreenGame

		case "c":
			m.screen = ScreenChat
		}
	}

	return m, nil
}

// updateGame 更新游戏屏幕
func (m *Model) updateGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "f": // Fold
			m.fold()

		case "c": // Check/Call
			m.call()

		case "r": // Raise
			m.screen = ScreenAction

		case "a": // All-in
			m.allIn()

		case "h": // Chat
			m.screen = ScreenChat
		}
	}

	return m, nil
}

// updateAction 更新动作屏幕
func (m *Model) updateAction(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			m.raise()
			m.screen = ScreenGame

		case "esc":
			m.screen = ScreenGame

		case "backspace":
			if len(m.chatInput) > 0 {
				m.chatInput = m.chatInput[:len(m.chatInput)-1]
			}

		default:
			if len(msg.Runes) > 0 {
				m.chatInput += string(msg.Runes[0])
			}
		}
	}

	return m, nil
}

// updateChat 更新聊天屏幕
func (m *Model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			m.sendChat()
			m.screen = ScreenGame

		case "esc":
			m.screen = ScreenGame

		case "backspace":
			if len(m.chatInput) > 0 {
				m.chatInput = m.chatInput[:len(m.chatInput)-1]
			}

		default:
			if len(msg.Runes) > 0 {
				m.chatInput += string(msg.Runes[0])
			}
		}
	}

	return m, nil
}

// View 渲染视图
func (m *Model) View() string {
	switch m.screen {
	case ScreenConnect:
		return m.viewConnect()

	case ScreenLobby:
		return m.viewLobby()

	case ScreenGame:
		return m.viewGame()

	case ScreenAction:
		return m.viewAction()

	case ScreenChat:
		return m.viewChat()
	}

	return ""
}

// viewConnect 渲染连接屏幕
func (m *Model) viewConnect() string {
	content := styleTitle.Render("Texas Hold'em Poker") + "\n\n"
	content += styleSubtitle.Render("请输入您的名称:") + "\n\n"
	content += styleButtonActive.Render(" "+m.playerName+" ")

	if m.err != nil {
		content += "\n\n" + styleWarning.Render(fmt.Sprintf("错误: %v", m.err))
	}

	content += "\n\n按 Enter 连接，按 Ctrl+C 退出"

	return content
}

// viewLobby 渲染大厅屏幕
func (m *Model) viewLobby() string {
	content := styleTitle.Render("等待游戏开始...") + "\n\n"

	if m.client != nil && m.client.PlayerID() != "" {
		content += styleSubtitle.Render(fmt.Sprintf("您的 ID: %s", m.client.PlayerID())) + "\n"
	}

	if m.gameState != nil {
		content += fmt.Sprintf("\n已连接玩家: %d/%d", len(m.gameState.Players), 9)
	}

	content += "\n\n按 Enter 进入游戏，按 C 进入聊天"

	return content
}

// viewGame 渲染游戏屏幕
func (m *Model) viewGame() string {
	var content string

	// 标题
	content += styleTitle.Render("Texas Hold'em Poker") + "\n"

	// 游戏状态
	stage := "等待开始"
	if m.gameState != nil {
		stage = m.gameState.Stage.String()
	}
	content += styleSubtitle.Render("当前阶段: ") + styleActive.Render(stage) + "\n"

	// 底池
	pot := 0
	if m.gameState != nil {
		pot = m.gameState.Pot
	}
	content += stylePot.Render(fmt.Sprintf("底池: %d", pot)) + "\n\n"

	// 公共牌
	content += m.renderCommunityCards() + "\n\n"

	// 玩家列表
	content += m.renderPlayers() + "\n"

	// 动作提示
	content += m.renderActionPrompt()

	// 聊天消息
	if len(m.messages) > 0 {
		content += "\n" + styleChat.Render("最近消息:")
		for _, msg := range m.messages {
			content += "\n" + styleInactive.Render(msg)
		}
	}

	return content
}

// viewAction 渲染动作屏幕
func (m *Model) viewAction() string {
	content := styleTitle.Render("下注金额") + "\n\n"
	content += styleSubtitle.Render("输入下注金额:") + "\n\n"
	content += styleButtonActive.Render(" "+m.chatInput+" ") + "\n\n"
	content += styleSubtitle.Render("按 Enter 确认，按 Esc 返回")

	return content
}

// viewChat 渲染聊天屏幕
func (m *Model) viewChat() string {
	content := styleTitle.Render("聊天") + "\n\n"
	content += styleSubtitle.Render("输入消息:") + "\n\n"
	content += styleButtonActive.Render(" "+m.chatInput+" ") + "\n\n"
	content += styleSubtitle.Render("按 Enter 发送，按 Esc 返回")

	return content
}

// renderCommunityCards 渲染公共牌
func (m *Model) renderCommunityCards() string {
	cards := make([]string, 0)

	if m.gameState != nil {
		for _, c := range m.gameState.CommunityCards {
			if c.Rank != 0 {
				cards = append(cards, renderCard(c))
			}
		}
	}

	if len(cards) == 0 {
		cards = []string{"[  ?  ]", "[  ?  ]", "[  ?  ]"}
	}

	return styleSubtitle.Render("公共牌: ") + lipgloss.JoinHorizontal(lipgloss.Center, cards...)
}

// renderPlayers 渲染玩家列表
func (m *Model) renderPlayers() string {
	if m.gameState == nil || len(m.gameState.Players) == 0 {
		return styleSubtitle.Render("等待玩家加入...") + "\n"
	}

	var players []string
	for _, p := range m.gameState.Players {
		isSelf := p.ID == m.playerID
		nameStyle := styleInactive
		if isSelf {
			nameStyle = styleActive
		}

		var holeCards string
		if isSelf && p.HoleCards[0].Rank != 0 {
			holeCards = renderCard(p.HoleCards[0]) + " " + renderCard(p.HoleCards[1])
		} else {
			holeCards = "[  ?  ] [  ?  ]"
		}

		playerStr := fmt.Sprintf("%s | 筹码: %d | 下注: %d | %s",
			nameStyle.Render(p.Name), p.Chips, p.CurrentBet, p.Status.String())
		players = append(players, playerStr)
		players = append(players, "    "+holeCards)
	}

	return styleSubtitle.Render("玩家:") + "\n" + lipgloss.JoinVertical(lipgloss.Left, players...)
}

// renderActionPrompt 渲染动作提示
func (m *Model) renderActionPrompt() string {
	var actions []string
	actions = append(actions, styleAction.Render("[F]old 弃牌"))
	actions = append(actions, styleAction.Render("[C]all 跟注"))
	actions = append(actions, styleAction.Render("[R]aise 加注"))
	actions = append(actions, styleAction.Render("[A]ll-in 全下"))
	actions = append(actions, styleAction.Render("[H]elp 帮助"))

	return "\n" + lipgloss.JoinHorizontal(lipgloss.Center, actions...)
}

// renderCard 渲染单张牌
func renderCard(card interface{ String() string }) string {
	return card.String()
}

// fold 弃牌
func (m *Model) fold() {
	if m.client != nil {
		m.client.SendPlayerAction(models.ActionFold, 0)
	}
}

// call 跟注
func (m *Model) call() {
	if m.client != nil {
		m.client.SendPlayerAction(models.ActionCall, 0)
	}
}

// raise 加注
func (m *Model) raise() {
	if m.client != nil && m.actionAmount > 0 {
		m.client.SendPlayerAction(models.ActionRaise, m.actionAmount)
	}
}

// allIn 全下
func (m *Model) allIn() {
	if m.client != nil {
		m.client.SendPlayerAction(models.ActionAllIn, 0)
	}
}

// sendChat 发送聊天消息
func (m *Model) sendChat() {
	if m.client != nil && m.chatInput != "" {
		m.client.SendChat(m.chatInput)
		m.messages = append(m.messages, "你说: "+m.chatInput)
		m.chatInput = ""
	}
}

// SetGameState 设置游戏状态
func (m *Model) SetGameState(state *protocol.GameState) {
	m.gameState = state
}

// SetPlayerID 设置玩家ID
func (m *Model) SetPlayerID(id string) {
	m.playerID = id
}

// AddMessage 添加消息
func (m *Model) AddMessage(msg string) {
	m.messages = append(m.messages, msg)
	if len(m.messages) > 10 {
		m.messages = m.messages[len(m.messages)-10:]
	}
}

// SetServerURL 设置服务器URL
func (m *Model) SetServerURL(url string) {
	m.serverURL = url
}

// Start 启动 TUI
func Start() error {
	model := NewModel()

	// 创建 TUI 程序
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	// 运行 TUI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI 运行错误: %w", err)
	}

	return nil
}
