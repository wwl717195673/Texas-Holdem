package models

import (
	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
)

// PlayerStatus 表示玩家的状态
type PlayerStatus int

const (
	PlayerStatusInactive PlayerStatus = iota // 未入座
	PlayerStatusActive                       // 游戏中
	PlayerStatusFolded                      // 已弃牌
	PlayerStatusAllIn                       // 全下
)

func (s PlayerStatus) String() string {
	names := []string{"未入座", "游戏中", "已弃牌", "全下"}
	if int(s) < len(names) {
		return names[s]
	}
	return "未知"
}

// ActionType 表示玩家可以执行的动作类型
type ActionType int

const (
	ActionFold ActionType = iota // 弃牌
	ActionCheck                  // 看牌
	ActionCall                   // 跟注
	ActionRaise                  // 加注
	ActionAllIn                  // 全下
)

func (a ActionType) String() string {
	names := []string{"弃牌", "看牌", "跟注", "加注", "全下"}
	if int(a) < len(names) {
		return names[a]
	}
	return "未知"
}

func (a ActionType) ShortName() string {
	names := []string{"F", "X", "C", "R", "A"}
	if int(a) < len(names) {
		return names[a]
	}
	return "?"
}

// PlayerAction 表示玩家在某一轮的动作记录
type PlayerAction struct {
	PlayerID string    // 玩家ID
	Action   ActionType // 执行的动作
	Amount   int       // 下注金额
}

// Player 表示一名扑克玩家
type Player struct {
	ID         string         // 玩家唯一标识
	Name       string         // 玩家名称
	Chips      int           // 剩余筹码
	Seat       int           // 座位号
	Status     PlayerStatus   // 玩家状态
	HoleCards  [2]card.Card  // 底牌
	CurrentBet int           // 当前下注金额
	IsDealer   bool          // 是否为庄家
	HasActed   bool          // 是否已完成本轮动作

	// 统计信息
	HandsPlayed int // 参与的手牌数
	HandsWon    int // 获胜的手牌数
}

// NewPlayer 创建一个新玩家
func NewPlayer(id, name string, chips int, seat int) *Player {
	return &Player{
		ID:    id,
		Name:  name,
		Chips: chips,
		Seat:  seat,
		Status: PlayerStatusActive,
	}
}

// IsActive 判断玩家是否仍在当前手牌中（未弃牌）
func (p *Player) IsActive() bool {
	return p.Status == PlayerStatusActive
}

// CanAct 判断玩家是否可以执行动作
func (p *Player) CanAct() bool {
	return p.IsActive() && !p.HasActed
}

// HasHoleCards 判断玩家是否已发到底牌
func (p *Player) HasHoleCards() bool {
	return p.HoleCards[0].Rank != 0
}

// GetHoleCardsDisplay 返回底牌的显示字符串
func (p *Player) GetHoleCardsDisplay() string {
	if !p.HasHoleCards() {
		return "[  ?  ][  ?  ]"
	}
	return p.HoleCards[0].String() + " " + p.HoleCards[1].String()
}

// GetHiddenHoleCardsDisplay 返回隐藏底牌的显示字符串（用于其他玩家）
func (p *Player) GetHiddenHoleCardsDisplay() string {
	return "[  ?  ][  ?  ]"
}

// NewPlayerWithID 使用自动生成的ID创建新玩家
func NewPlayerWithID(name string, chips int, seat int) *Player {
	return NewPlayer(generateID(), name, chips, seat)
}

// 生成随机ID
func generateID() string {
	return randomID(8)
}

func randomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
