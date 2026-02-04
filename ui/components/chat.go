package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChatMessage 表示一条聊天消息
type ChatMessage struct {
	PlayerID   string    // 玩家ID
	PlayerName string    // 玩家名称
	Content    string    // 消息内容
	Time       time.Time // 发送时间
	IsSystem   bool      // 是否为系统消息
}

// ChatModel 聊天组件模型
type ChatModel struct {
	messages []ChatMessage  // 消息列表
	input    textarea.Model // 输入框
	visible  bool          // 是否显示
	maxLines int           // 最大显示行数
	width    int           // 组件宽度
	height   int           // 组件高度
	focused  bool          // 是否获得焦点
}

// NewChatModel 创建新的聊天组件
func NewChatModel() *ChatModel {
	ti := textarea.New()
	ti.Placeholder = "输入消息... (Enter发送, Esc取消)"
	ti.Focus()
	ti.Prompt = "» "
	ti.CharLimit = 200
	ti.SetWidth(40)
	ti.SetHeight(3)

	return &ChatModel{
		messages: make([]ChatMessage, 0),
		input:    ti,
		visible:  false,
		maxLines: 50,
		width:    40,
		height:   15,
		focused:  false,
	}
}

// Toggle 切换聊天组件的显示/隐藏状态
func (m *ChatModel) Toggle() {
	m.visible = !m.visible
}

// SetVisible 设置聊天组件是否显示
func (m *ChatModel) SetVisible(visible bool) {
	m.visible = visible
}

// IsVisible 返回聊天组件是否显示
func (m *ChatModel) IsVisible() bool {
	return m.visible
}

// SetSize 设置聊天组件的尺寸
func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.SetWidth(width)
}

// Focus 聚焦聊天输入框
func (m *ChatModel) Focus() {
	m.focused = true
	m.input.Focus()
}

// Blur 取消聚焦聊天输入框
func (m *ChatModel) Blur() {
	m.focused = false
	m.input.Blur()
}

// AddMessage 添加一条玩家消息
func (m *ChatModel) AddMessage(playerID, playerName, content string) {
	m.messages = append(m.messages, ChatMessage{
		PlayerID:   playerID,
		PlayerName: playerName,
		Content:    content,
		Time:       time.Now(),
		IsSystem:   false,
	})

	// 保持消息数量在限制内
	if len(m.messages) > m.maxLines {
		m.messages = m.messages[len(m.messages)-m.maxLines:]
	}
}

// AddSystemMessage 添加一条系统消息
func (m *ChatModel) AddSystemMessage(content string) {
	m.messages = append(m.messages, ChatMessage{
		PlayerID:   "system",
		PlayerName: "系统",
		Content:    content,
		Time:       time.Now(),
		IsSystem:   true,
	})

	// 保持消息数量在限制内
	if len(m.messages) > m.maxLines {
		m.messages = m.messages[len(m.messages)-m.maxLines:]
	}
}

// ClearMessages 清空所有消息
func (m *ChatModel) ClearMessages() {
	m.messages = make([]ChatMessage, 0)
}

// GetMessages 返回所有消息
func (m *ChatModel) GetMessages() []ChatMessage {
	return m.messages
}

// Update 处理消息
func (m *ChatModel) Update(msg tea.Msg) (*ChatModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View 返回聊天组件的渲染字符串
func (m *ChatModel) View() string {
	if !m.visible {
		return ""
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(m.width).
		Height(m.height)

	var content strings.Builder

	// 标题
	content.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Render(" 聊天 ") + "\n")

	// 消息区域
	msgHeight := m.height - 5 // 留出标题和输入框的空间
	if msgHeight < 1 {
		msgHeight = 1
	}

	// 计算消息显示的起始位置
	msgCount := len(m.messages)
	startIdx := 0
	if msgCount > msgHeight {
		startIdx = msgCount - msgHeight
	}

	// 显示消息
	for i := startIdx; i < msgCount && i < startIdx+msgHeight; i++ {
		msg := m.messages[i]
		timeStr := msg.Time.Format("15:04")

		if msg.IsSystem {
			// 系统消息使用不同颜色
			content.WriteString(fmt.Sprintf("[%s] %s\n",
				timeStr,
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("241")).
					Render(msg.Content)))
		} else {
			// 玩家消息
			content.WriteString(fmt.Sprintf("[%s] %s: %s\n",
				timeStr,
				lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("214")).
					Render(msg.PlayerName),
				msg.Content))
		}
	}

	// 填充空白行
	linesShown := msgCount - startIdx
	if linesShown < msgHeight {
		for i := 0; i < msgHeight-linesShown; i++ {
			content.WriteString("\n")
		}
	}

	// 输入框
	content.WriteString("\n" + m.input.View())

	// 底部提示
	content.WriteString("\n" + lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(" Enter: 发送  |  Esc: 关闭"))

	return style.Render(content.String())
}

// GetInputValue 获取输入框的内容
func (m *ChatModel) GetInputValue() string {
	return m.input.Value()
}

// ClearInput 清空输入框
func (m *ChatModel) ClearInput() {
	m.input.Reset()
}

// SetInputValue 设置输入框的内容
func (m *ChatModel) SetInputValue(value string) {
	m.input.SetValue(value)
}

// IsFocused 返回输入框是否获得焦点
func (m *ChatModel) IsFocused() bool {
	return m.focused
}
