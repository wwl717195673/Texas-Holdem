package host

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/protocol"
	"github.com/wilenwang/just_play/Texas-Holdem/server/host"
)

// 样式定义
var (
	styleTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6")).MarginBottom(1)
	styleSubtitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	styleBox      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)
	styleActive   = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	styleInactive = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	stylePot      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1FA8C"))
	styleButton   = lipgloss.NewStyle().Background(lipgloss.Color("#44475A")).Foreground(lipgloss.Color("#F8F8F2")).Padding(0, 2)
	styleSelected = lipgloss.NewStyle().Background(lipgloss.Color("#FF79C6")).Foreground(lipgloss.Color("#F8F8F2")).Padding(0, 2)
)

// Model TUI 模型
type Model struct {
	server       *host.Server
	gameState    *protocol.GameState
	selectedMenu int
	menuItems    []string
	width        int
	height       int
	err          error
}

// NewModel 创建新的 TUI 模型
func NewModel(server *host.Server) *Model {
	return &Model{
		server:    server,
		menuItems: []string{"开始游戏", "暂停游戏", "查看日志", "退出"},
	}
}

// Init 初始化
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// handleKeyMsg 处理键盘消息
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if m.selectedMenu > 0 {
			m.selectedMenu--
		}

	case "down", "j":
		if m.selectedMenu < len(m.menuItems)-1 {
			m.selectedMenu++
		}

	case "enter":
		return m.handleMenuSelect()
	}

	return m, nil
}

// handleMenuSelect 处理菜单选择
func (m *Model) handleMenuSelect() (tea.Model, tea.Cmd) {
	switch m.selectedMenu {
	case 0: // 开始游戏
		// TODO: 开始游戏逻辑

	case 1: // 暂停游戏
		// TODO: 暂停游戏逻辑

	case 2: // 查看日志
		// TODO: 查看日志逻辑

	case 3: // 退出
		return m, tea.Quit
	}

	return m, nil
}

// View 渲染视图
func (m *Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("错误: %v\n", m.err)
	}

	var content string

	// 标题
	content += styleTitle.Render("Texas Hold'em Poker - 游戏控制台") + "\n\n"

	// 游戏状态
	content += m.renderGameState() + "\n"

	// 公共牌
	content += m.renderCommunityCards() + "\n"

	// 玩家列表
	content += m.renderPlayers() + "\n"

	// 底池
	content += m.renderPot() + "\n"

	// 菜单
	content += m.renderMenu()

	return content
}

// renderGameState 渲染游戏状态
func (m *Model) renderGameState() string {
	stage := "等待开始"
	if m.gameState != nil {
		stage = m.gameState.Stage.String()
	}

	return styleBox.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			styleSubtitle.Render("当前阶段: ")+styleActive.Render(stage),
		),
	)
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
	for i, p := range m.gameState.Players {
		status := p.Status.String()

		var holeCards string
		if p.HoleCards[0].Rank != 0 {
			holeCards = renderCard(p.HoleCards[0]) + " " + renderCard(p.HoleCards[1])
		} else {
			holeCards = "[  ?  ] [  ?  ]"
		}

		playerStr := fmt.Sprintf("%d. %s (座位%d) | 筹码: %d | 下注: %d | %s",
			i+1, p.Name, p.Seat, p.Chips, p.CurrentBet, status)
		players = append(players, playerStr)
		players = append(players, "    "+holeCards)
	}

	return styleSubtitle.Render("玩家:") + "\n" + lipgloss.JoinVertical(lipgloss.Left, players...)
}

// renderPot 渲染底池
func (m *Model) renderPot() string {
	pot := 0
	if m.gameState != nil {
		pot = m.gameState.Pot
	}
	return stylePot.Render(fmt.Sprintf("底池: %d", pot))
}

// renderMenu 渲染菜单
func (m *Model) renderMenu() string {
	var items []string
	for i, item := range m.menuItems {
		if i == m.selectedMenu {
			items = append(items, styleSelected.Render("> "+item+" <"))
		} else {
			items = append(items, styleButton.Render("  "+item+"  "))
		}
	}

	return "\n" + lipgloss.JoinHorizontal(lipgloss.Center, items...)
}

// SetGameState 设置游戏状态
func (m *Model) SetGameState(state *protocol.GameState) {
	m.gameState = state
}

// renderCard 渲染单张牌
func renderCard(card interface{ String() string }) string {
	return card.String()
}

// Start 启动 TUI
func Start(server *host.Server) error {
	model := NewModel(server)

	// 创建 TUI 程序
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// 处理退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		p.Kill()
	}()

	// 启动 HTTP 服务器（用于 WebSocket 连接）
	go func() {
		http.HandleFunc("/ws", server.ServeHTTP)
		addr := ":8080"
		fmt.Printf("游戏服务器启动在 http://%s\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP 服务器错误: %v\n", err)
		}
	}()

	// 运行 TUI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI 运行错误: %w", err)
	}

	return nil
}
