package game

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/pkg/evaluator"
)

// HandHistory 表示一手牌的历史记录
type HandHistory struct {
	HandID         int              `json:"hand_id"`          // 手牌编号
	Timestamp      time.Time        `json:"timestamp"`        // 时间戳
	GameID         string           `json:"game_id"`         // 游戏ID
	Players        []HistoryPlayer  `json:"players"`         // 参与的玩家
	CommunityCards [5]card.Card    `json:"community_cards"` // 公共牌
	Actions        []HistoryAction  `json:"actions"`         // 行动记录
	Showdown       []ShowdownInfo  `json:"showdown"`        // 摊牌信息
	Pot            int              `json:"pot"`             // 底池金额
	Winners        []WinnerInfo    `json:"winners"`         // 获胜者信息
}

// HistoryPlayer 表示历史记录中的玩家信息
type HistoryPlayer struct {
	ID         string       `json:"id"`          // 玩家ID
	Name       string       `json:"name"`        // 玩家名称
	Seat       int          `json:"seat"`        // 座位号
	HoleCards  [2]card.Card `json:"hole_cards"`  // 底牌
	FinalChips int          `json:"final_chips"` // 最终筹码
	WonChips   int          `json:"won_chips"`   // 赢得筹码
	IsWinner   bool         `json:"is_winner"`   // 是否获胜
}

// HistoryAction 表示历史记录中的行动
type HistoryAction struct {
	PlayerID   string           `json:"player_id"`   // 玩家ID
	PlayerName string           `json:"player_name"` // 玩家名称
	Round      Stage           `json:"round"`       // 所在阶段
	Action     models.ActionType `json:"action"`     // 执行的行动
	Amount     int              `json:"amount"`      // 下注金额
	Timestamp  time.Time        `json:"timestamp"`   // 时间戳
}

// ShowdownInfo 表示摊牌信息
type ShowdownInfo struct {
	PlayerID   string                 `json:"player_id"`   // 玩家ID
	PlayerName string                 `json:"player_name"` // 玩家名称
	HandRank   evaluator.HandRank    `json:"hand_rank"`   // 手牌等级
	HandName   string                `json:"hand_name"`   // 手牌名称
	BestCards  []card.Card           `json:"best_cards"`  // 最佳手牌的五张牌
}

// WinnerInfo 表示获胜者信息
type WinnerInfo struct {
	PlayerID   string `json:"player_id"`   // 玩家ID
	PlayerName string `json:"player_name"` // 玩家名称
	Amount     int    `json:"amount"`      // 赢得筹码
}

// HistoryManager 管理手牌历史记录
type HistoryManager struct {
	hands      []HandHistory      // 所有手牌历史
	currentHand *HandHistory      // 当前正在进行的牌局
	file       *os.File           // 文件句柄（用于保存）
	filename   string             // 保存文件名
	mu         chan struct{}      // 互斥锁（使用chan模拟）
}

// NewHistoryManager 创建历史记录管理器
func NewHistoryManager(filename string) (*HistoryManager, error) {
	h := &HistoryManager{
		hands:    make([]HandHistory, 0),
		filename: filename,
		mu:       make(chan struct{}, 1),
	}

	// 如果指定了文件名，尝试加载历史记录
	if filename != "" {
		if file, err := os.Open(filename); err == nil {
			defer file.Close()
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&h.hands); err == nil {
				// 加载成功
			}
		}
	}

	return h, nil
}

// StartHand 开始记录新的一手牌
func (h *HistoryManager) StartHand(handID int, gameID string) {
	h.mu <- struct{}{}
	defer func() { <-h.mu }()

	h.currentHand = &HandHistory{
		HandID:    handID,
		Timestamp: time.Now(),
		GameID:    gameID,
		Players:   make([]HistoryPlayer, 0),
		Actions:   make([]HistoryAction, 0),
		Showdown:  make([]ShowdownInfo, 0),
	}
}

// AddPlayer 添加玩家到当前手牌记录
func (h *HistoryManager) AddPlayer(id, name string, seat int, chips int) {
	if h.currentHand == nil {
		return
	}

	h.currentHand.Players = append(h.currentHand.Players, HistoryPlayer{
		ID:         id,
		Name:       name,
		Seat:       seat,
		FinalChips: chips,
	})
}

// RecordAction 记录玩家行动
func (h *HistoryManager) RecordAction(playerID, playerName string, round Stage, action models.ActionType, amount int) {
	if h.currentHand == nil {
		return
	}

	h.currentHand.Actions = append(h.currentHand.Actions, HistoryAction{
		PlayerID:   playerID,
		PlayerName: playerName,
		Round:      round,
		Action:     action,
		Amount:     amount,
		Timestamp:  time.Now(),
	})
}

// RecordShowdown 记录摊牌信息
func (h *HistoryManager) RecordShowdown(playerID, playerName string, handRank evaluator.HandRank, handName string, bestCards []card.Card) {
	if h.currentHand == nil {
		return
	}

	h.currentHand.Showdown = append(h.currentHand.Showdown, ShowdownInfo{
		PlayerID:   playerID,
		PlayerName: playerName,
		HandRank:   handRank,
		HandName:   handName,
		BestCards:  bestCards,
	})
}

// EndHand 结束当前手牌记录
func (h *HistoryManager) EndHand(community [5]card.Card, pot int, winners []WinnerInfo) {
	if h.currentHand == nil {
		return
	}

	h.currentHand.CommunityCards = community
	h.currentHand.Pot = pot
	h.currentHand.Winners = winners

	// 标记获胜者
	winnerSet := make(map[string]bool)
	for _, w := range winners {
		winnerSet[w.PlayerID] = true
	}

	for i := range h.currentHand.Players {
		if winnerSet[h.currentHand.Players[i].ID] {
			h.currentHand.Players[i].IsWinner = true
		}
	}

	h.hands = append(h.hands, *h.currentHand)
	h.currentHand = nil

	// 保存到文件
	if h.filename != "" {
		h.saveToFile()
	}
}

// saveToFile 保存历史记录到文件
func (h *HistoryManager) saveToFile() {
	if h.filename == "" {
		return
	}

	file, err := os.OpenFile(h.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(h.hands)
}

// GetRecentHands 获取最近N手牌历史
func (h *HistoryManager) GetRecentHands(n int) []HandHistory {
	h.mu <- struct{}{}
	defer func() { <-h.mu }()

	if n <= 0 {
		n = 10
	}
	if n > len(h.hands) {
		n = len(h.hands)
	}

	result := make([]HandHistory, n)
	copy(result, h.hands[len(h.hands)-n:])
	return result
}

// GetAllHands 获取所有手牌历史
func (h *HistoryManager) GetAllHands() []HandHistory {
	h.mu <- struct{}{}
	defer func() { <-h.mu }()

	result := make([]HandHistory, len(h.hands))
	copy(result, h.hands)
	return result
}

// GetHandCount 获取历史手牌总数
func (h *HistoryManager) GetHandCount() int {
	h.mu <- struct{}{}
	defer func() { <-h.mu }()

	return len(h.hands)
}

// ExportToText 将手牌历史导出为文本格式
func (h *HandHistory) ExportToText() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("=== 手牌 #%d ===\n", h.HandID))
	b.WriteString(fmt.Sprintf("时间: %s\n", h.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("底池: %d\n\n", h.Pot))

	// 玩家信息
	b.WriteString("玩家:\n")
	for _, p := range h.Players {
		cards := ""
		if p.HoleCards[0].Rank != 0 {
			cards = fmt.Sprintf(" 底牌: %s %s", p.HoleCards[0].String(), p.HoleCards[1].String())
		}
		winnerMark := ""
		if p.IsWinner {
			winnerMark = " [胜]"
		}
		b.WriteString(fmt.Sprintf("  [%d] %s: 筹码%d%s%s\n", p.Seat+1, p.Name, p.FinalChips, cards, winnerMark))
	}

	// 公共牌
	b.WriteString("\n公共牌: ")
	for _, c := range h.CommunityCards {
		if c.Rank != 0 {
			b.WriteString(c.String() + " ")
		}
	}
	b.WriteString("\n")

	// 行动记录
	b.WriteString("\n行动:\n")
	for _, a := range h.Actions {
		amountStr := ""
		if a.Amount > 0 {
			amountStr = fmt.Sprintf(" (%d)", a.Amount)
		}
		b.WriteString(fmt.Sprintf("  %s [%s]: %s%s\n", a.PlayerName, a.Round.ShortString(), a.Action.ShortName(), amountStr))
	}

	// 摊牌信息
	b.WriteString("\n摊牌:\n")
	for _, s := range h.Showdown {
		cards := ""
		for _, c := range s.BestCards {
			cards += c.String() + " "
		}
		b.WriteString(fmt.Sprintf("  %s: %s (%s)\n", s.PlayerName, s.HandName, cards))
	}

	// 获胜者
	b.WriteString("\n获胜者:\n")
	for _, w := range h.Winners {
		b.WriteString(fmt.Sprintf("  %s: 赢得 %d 筹码\n", w.PlayerName, w.Amount))
	}

	return b.String()
}

// ShortString 返回阶段的简短名称
func (s Stage) ShortString() string {
	names := []string{"等待", "翻牌前", "翻牌", "转牌", "河牌", "摊牌", "结束"}
	if s >= 0 && int(s) < len(names) {
		return names[s]
	}
	return "?"
}
