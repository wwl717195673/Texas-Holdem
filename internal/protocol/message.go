package protocol

import (
	"time"

	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/pkg/evaluator"
	"github.com/wilenwang/just_play/Texas-Holdem/pkg/game"
)

// MessageType 表示消息类型
type MessageType string

const (
	// 客户端 -> 服务器消息类型
	MsgTypeJoin          MessageType = "join"           // 玩家加入游戏
	MsgTypeLeave        MessageType = "leave"           // 玩家离开游戏
	MsgTypePlayerAction MessageType = "player_action"   // 玩家执行动作
	MsgTypeChat         MessageType = "chat"            // 发送聊天消息
	MsgTypePing         MessageType = "ping"           // 心跳检测

	// 服务器 -> 客户端消息类型
	MsgTypeJoinAck      MessageType = "join_ack"       // 加入游戏确认
	MsgTypeGameState    MessageType = "game_state"     // 游戏状态更新
	MsgTypeYourTurn     MessageType = "your_turn"     // 通知玩家回合
	MsgTypePlayerJoined MessageType = "player_joined"  // 玩家加入通知
	MsgTypePlayerLeft   MessageType = "player_left"    // 玩家离开通知
	MsgTypePlayerActed  MessageType = "player_acted"   // 玩家动作通知
	MsgTypeShowdown     MessageType = "showdown"      // 摊牌结果
	MsgTypePong         MessageType = "pong"           // 心跳响应
	MsgTypeError        MessageType = "error"         // 错误消息
)

// BaseMessage 消息基类
type BaseMessage struct {
	Type      MessageType `json:"type"`       // 消息类型
	Timestamp int64       `json:"timestamp"`  // 时间戳
}

// ==================== 客户端 -> 服务器消息 ====================

// JoinRequest 玩家加入游戏请求
type JoinRequest struct {
	BaseMessage
	PlayerName string `json:"player_name"` // 玩家名称
	Seat       int    `json:"seat"`        // 请求座位号（-1表示随机）
}

// LeaveRequest 玩家离开游戏请求
type LeaveRequest struct {
	BaseMessage
	PlayerID string `json:"player_id"` // 玩家ID
}

// PlayerActionRequest 玩家动作请求
type PlayerActionRequest struct {
	BaseMessage
	PlayerID string          `json:"player_id"` // 玩家ID
	Action   models.ActionType `json:"action"`   // 动作类型
	Amount   int             `json:"amount"`    // 下注金额（加注时使用）
}

// ChatRequest 聊天消息请求
type ChatRequest struct {
	BaseMessage
	PlayerID string `json:"player_id"` // 玩家ID
	Content  string `json:"content"`   // 消息内容
}

// PingRequest 心跳检测请求
type PingRequest struct {
	BaseMessage
}

// ==================== 服务器 -> 客户端消息 ====================

// JoinAck 加入游戏确认响应
type JoinAck struct {
	BaseMessage
	Success   bool   `json:"success"`    // 是否成功
	PlayerID  string `json:"player_id"`  // 分配的玩家ID
	Seat      int    `json:"seat"`       // 座位号
	Message   string `json:"message"`     // 附加消息
	GameState *GameState `json:"game_state,omitempty"` // 当前游戏状态
}

// GameState 游戏状态信息（用于同步给客户端）
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
	MaxRaise      int               `json:"max_raise"`        // 最大加注金额（当前最高下注+玩家筹码）
}

// PlayerInfo 玩家公开信息
type PlayerInfo struct {
	ID         string              `json:"id"`           // 玩家ID
	Name       string              `json:"name"`         // 玩家名称
	Seat       int                 `json:"seat"`         // 座位号
	Chips      int                 `json:"chips"`        // 剩余筹码
	CurrentBet int                 `json:"current_bet"`  // 当前下注金额
	Status     models.PlayerStatus `json:"status"`       // 玩家状态
	IsDealer   bool                `json:"is_dealer"`    // 是否为庄家
	HoleCards  [2]card.Card        `json:"hole_cards"`   // 底牌（仅在摊牌或自己可见时发送）
	IsSelf     bool                `json:"is_self"`      // 是否是请求者自己
}

// YourTurn 通知玩家轮到其行动
type YourTurn struct {
	BaseMessage
	PlayerID    string `json:"player_id"`    // 玩家ID
	MinAction   int    `json:"min_action"`  // 最小下注金额
	MaxAction   int    `json:"max_action"`  // 最大下注金额
	CurrentBet  int    `json:"current_bet"` // 当前最高下注
	TimeLeft    int    `json:"time_left"`   // 剩余时间（秒）
}

// PlayerJoined 通知有新玩家加入
type PlayerJoined struct {
	BaseMessage
	Player PlayerInfo `json:"player"` // 新加入的玩家信息
}

// PlayerLeft 通知有玩家离开
type PlayerLeft struct {
	BaseMessage
	PlayerID string `json:"player_id"` // 离开的玩家ID
	PlayerName string `json:"player_name"` // 离开的玩家名称
}

// PlayerActed 通知有玩家执行了动作
type PlayerActed struct {
	BaseMessage
	PlayerID string          `json:"player_id"` // 玩家ID
	PlayerName string        `json:"player_name"` // 玩家名称
	Action   models.ActionType `json:"action"`   // 执行的動作
	Amount   int             `json:"amount"`    // 下注金额
	TotalBet int             `json:"total_bet"` // 总下注金额
}

// Showdown 摊牌结果
type Showdown struct {
	BaseMessage
	Winners   []WinnerInfo `json:"winners"`    // 获胜者列表
	Pot       int          `json:"pot"`        // 底池金额
	GameState *GameState  `json:"game_state"` // 最终游戏状态
}

// WinnerInfo 获胜者信息
type WinnerInfo struct {
	PlayerID   string                 `json:"player_id"`   // 玩家ID
	PlayerName string                 `json:"player_name"` // 玩家名称
	HandRank   evaluator.HandRank    `json:"hand_rank"`   // 手牌等级
	HandName   string                `json:"hand_name"`   // 手牌名称
	WonChips   int                   `json:"won_chips"`   // 赢得筹码
	RawCards   []card.Card           `json:"raw_cards"`   // 构成最佳手牌的牌
}

// ChatMessage 聊天消息
type ChatMessage struct {
	BaseMessage
	PlayerID   string `json:"player_id"`   // 玩家ID
	PlayerName string `json:"player_name"` // 玩家名称
	Content    string `json:"content"`     // 消息内容
	IsSystem   bool   `json:"is_system"`  // 是否为系统消息
}

// Pong 心跳响应
type Pong struct {
	BaseMessage
	ServerTime int64 `json:"server_time"` // 服务器时间
}

// Error 错误消息
type Error struct {
	BaseMessage
	Code    int    `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误描述
}

// ==================== 辅助函数 ====================

// NewBaseMessage 创建带时间戳的基本消息
func NewBaseMessage(msgType MessageType) BaseMessage {
	return BaseMessage{
		Type:      msgType,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewJoinRequest 创建加入游戏请求
func NewJoinRequest(playerName string, seat int) *JoinRequest {
	return &JoinRequest{
		BaseMessage: NewBaseMessage(MsgTypeJoin),
		PlayerName:  playerName,
		Seat:        seat,
	}
}

// NewLeaveRequest 创建离开游戏请求
func NewLeaveRequest(playerID string) *LeaveRequest {
	return &LeaveRequest{
		BaseMessage: NewBaseMessage(MsgTypeLeave),
		PlayerID:    playerID,
	}
}

// NewPlayerActionRequest 创建玩家动作请求
func NewPlayerActionRequest(playerID string, action models.ActionType, amount int) *PlayerActionRequest {
	return &PlayerActionRequest{
		BaseMessage: NewBaseMessage(MsgTypePlayerAction),
		PlayerID:    playerID,
		Action:      action,
		Amount:      amount,
	}
}

// NewChatRequest 创建聊天消息请求
func NewChatRequest(playerID, content string) *ChatRequest {
	return &ChatRequest{
		BaseMessage: NewBaseMessage(MsgTypeChat),
		PlayerID:    playerID,
		Content:     content,
	}
}

// NewPingRequest 创建心跳检测请求
func NewPingRequest() *PingRequest {
	return &PingRequest{
		BaseMessage: NewBaseMessage(MsgTypePing),
	}
}
