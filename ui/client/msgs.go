package client

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/protocol"
)

// ==================== Bubble Tea 自定义消息类型 ====================

// ServerMsg 服务器消息（原始字节）
type ServerMsg []byte

// JoinAckMsg 加入确认消息
type JoinAckMsg struct {
	Success  bool
	PlayerID string
	Seat     int
	Message  string
}

// JoinAckResultMsg 加入确认结果消息（来自客户端库回调）
type JoinAckResultMsg struct {
	Success  bool
	PlayerID string
	Seat     int
}

// GameStateMsg 游戏状态更新消息
type GameStateMsg struct {
	State *protocol.GameState
}

// YourTurnMsg 轮到玩家回合消息
type YourTurnMsg struct {
	Turn *protocol.YourTurn
}

// PlayerJoinedMsg 玩家加入通知消息
type PlayerJoinedMsg struct {
	Player protocol.PlayerInfo
}

// PlayerLeftMsg 玩家离开通知消息
type PlayerLeftMsg struct {
	PlayerName string
}

// PlayerActedMsg 玩家动作通知消息
type PlayerActedMsg struct {
	PlayerName string
	Action     models.ActionType
	Amount     int
}

// ShowdownMsg 摊牌结果消息
type ShowdownMsg struct {
	Showdown *protocol.Showdown
}

// ChatMsg 聊天消息
type ChatMsg struct {
	Message *protocol.ChatMessage
}

// ErrorMsg 错误消息
type ErrorMsg struct {
	Err error
}

// ConnectedMsg 连接成功消息
type ConnectedMsg struct{}

// DisconnectedMsg 断开连接消息
type DisconnectedMsg struct{}

// ==================== 辅助函数 ====================

// SendMsg 向 Bubble Tea 程序发送消息
func SendMsg(p tea.Program, msg tea.Msg) {
	p.Send(msg)
}
