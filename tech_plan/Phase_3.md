# Phase 3: 高级功能实现

## 阶段目标

实现AI玩家、历史记录、聊天增强、游戏统计等高级功能。

## 任务清单

| 功能 | 文件 | 优先级 |
|-----|------|-------|
| AI玩家 | `player/ai.go` | P0 |
| 历史记录 | `game/history.go` | P1 |
| 聊天系统 | `ui/components/chat.go` | P1 |
| 游戏统计 | `game/stats.go` | P2 |
| 界面美化 | `ui/components/render.go` | P2 |

---

## 任务 3.1: AI 玩家

### 3.1.1 AI 策略定义

```go
// player/ai.go

package player

import (
    "fmt"
    "math/rand"
    "time"

    "yourproject/card"
    "yourproject/evaluator"
    "yourproject/protocol"
)

// AI策略类型
type Strategy int

const (
    StrategyRandom Strategy = iota  // 随机策略
    StrategyConservative            // 保守策略
    StrategyAggressive              // 激进策略
    StrategyTight                   // 只玩强牌
)

// AIController AI控制器
type AIController struct {
    ID       string
    Name     string
    Chips    int
    Seat     int
    Strategy Strategy
    evaluator *evaluator.Evaluator
    rand     *rand.Rand

    // 状态
    lastAction  protocol.ActionType
    lastBet    int
    handCount  int
}

// NewAIPlayer 创建AI玩家
func NewAIPlayer(name string, chips int, seat int) *AIController {
    return &AIController{
        ID:       fmt.Sprintf("ai_%d_%d", seat, time.Now().UnixNano()%10000),
        Name:     name,
        Chips:    chips,
        Seat:     seat,
        Strategy: StrategyConservative,
        evaluator: evaluator.NewEvaluator(),
        rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

// SetStrategy 设置策略
func (ai *AIController) SetStrategy(s Strategy) {
    ai.Strategy = s
}

// Decide 决策
func (ai *AIController) Decide(state *protocol.GameStatePayload) (protocol.ActionType, int) {
    ai.handCount++

    // 获取当前玩家信息
    myPlayer := ai.getMyPlayer(state)
    if myPlayer == nil {
        return protocol.ActionFold, 0
    }

    // 检查是否轮到自己
    if state.CurrentPlayer != myPlayer.Seat {
        return protocol.ActionCheck, 0
    }

    // 评估手牌强度
    handStrength := ai.evaluateHand(myPlayer, state)

    // 根据策略决策
    switch ai.Strategy {
    case StrategyRandom:
        return ai.randomDecision(state, myPlayer, handStrength)
    case StrategyConservative:
        return ai.conservativeDecision(state, myPlayer, handStrength)
    case StrategyAggressive:
        return ai.aggressiveDecision(state, myPlayer, handStrength)
    case StrategyTight:
        return ai.tightDecision(state, myPlayer, handStrength)
    default:
        return ai.conservativeDecision(state, myPlayer, handStrength)
    }
}
```

### 3.1.2 手牌评估与决策

```go
// player/ai.go (续)

// evaluateHand 评估手牌强度 (0-1)
func (ai *AIController) evaluateHand(p *protocol.PlayerInfo, state *protocol.GameStatePayload) float64 {
    if len(p.HoleCards) != 2 || p.HoleCards[0].Rank == 0 {
        return 0.5
    }

    score := 0.0

    // 高牌加分
    for _, c := range p.HoleCards {
        score += float64(c.Rank) / 14.0
    }
    score /= 2.0

    // 同花加分
    if p.HoleCards[0].Suit == p.HoleCards[1].Suit {
        score += 0.1
    }

    // 对子加分
    if p.HoleCards[0].Rank == p.HoleCards[1].Rank {
        score += 0.3
    }

    // 间隔小加分
    diff := abs(int(p.HoleCards[0].Rank) - int(p.HoleCards[1].Rank))
    if diff <= 3 {
        score += 0.1
    }

    // 根据阶段调整
    switch state.Stage {
    case protocol.StagePreFlop:
        score *= 1.5
    case protocol.StageFlop:
        score *= 1.2
    case protocol.StageRiver:
        if score < 0.3 {
            score = 0.2
        }
    }

    return min(1.0, score)
}

// randomDecision 随机策略
func (ai *AIController) randomDecision(state *protocol.GameStatePayload,
    p *protocol.PlayerInfo, strength float64) (protocol.ActionType, int) {

    roll := ai.rand.Float64()

    if roll < 0.1 {
        return protocol.ActionFold, 0
    } else if roll < 0.4 {
        return ai.callOrCheck(state, p)
    } else if roll < 0.8 {
        return ai.raise(state, p)
    }
    return protocol.ActionAllIn, 0
}

// conservativeDecision 保守策略
func (ai *AIController) conservativeDecision(state *protocol.GameStatePayload,
    p *protocol.PlayerInfo, strength float64) (protocol.ActionType, int) {

    if strength < 0.3 {
        return protocol.ActionFold, 0
    }
    if strength < 0.6 {
        return ai.callOrCheck(state, p)
    }
    return ai.raise(state, p)
}

// aggressiveDecision 激进策略
func (ai *AIController) aggressiveDecision(state *protocol.GameStatePayload,
    p *protocol.PlayerInfo, strength float64) (protocol.ActionType, int) {

    if strength < 0.15 {
        roll := ai.rand.Float64()
        if roll < 0.7 {
            return ai.callOrCheck(state, p)
        }
        return protocol.ActionFold, 0
    }

    if strength > 0.4 {
        roll := ai.rand.Float64()
        if roll < 0.6 {
            return ai.raise(state, p)
        }
        return protocol.ActionAllIn, 0
    }

    return ai.callOrCheck(state, p)
}

// tightDecision 紧策略
func (ai *AIController) tightDecision(state *protocol.GameStatePayload,
    p *protocol.PlayerInfo, strength float64) (protocol.ActionType, int) {

    if strength < 0.5 {
        return protocol.ActionFold, 0
    }
    if strength > 0.7 {
        return ai.raise(state, p)
    }
    return ai.callOrCheck(state, p)
}

// callOrCheck 看牌或跟注
func (ai *AIController) callOrCheck(state *protocol.GameStatePayload, p *protocol.PlayerInfo) (protocol.ActionType, int) {
    callAmount := state.CurrentBet - p.CurrentBet
    if callAmount == 0 {
        return protocol.ActionCheck, 0
    }
    return protocol.ActionCall, callAmount
}

// raise 加注
func (ai *AIController) raise(state *protocol.GameStatePayload, p *protocol.PlayerInfo) (protocol.ActionType, int) {
    minRaise := state.MinRaise
    if minRaise <= 0 {
        return protocol.ActionCall, 0
    }

    availableChips := p.Chips
    raiseAmount := minRaise

    roll := ai.rand.Float64()
    if roll < 0.5 {
        raiseAmount = minRaise * 2
    } else if roll < 0.8 {
        raiseAmount = minRaise * 3
    } else {
        raiseAmount = minRaise * (2 + ai.rand.Intn(3))
    }

    if raiseAmount > availableChips {
        return protocol.ActionAllIn, 0
    }

    return protocol.ActionRaise, raiseAmount
}

func (ai *AIController) getMyPlayer(state *protocol.GameStatePayload) *protocol.PlayerInfo {
    for _, p := range state.Players {
        if p.ID == ai.ID {
            return &p
        }
    }
    return nil
}

func abs(n int) int {
    if n < 0 {
        return -n
    }
    return n
}

func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}
```

---

## 任务 3.2: 历史记录

```go
// game/history.go

package game

import (
    "encoding/json"
    "os"
    "strings"
    "time"

    "yourproject/card"
    "yourproject/evaluator"
    "yourproject/protocol"
)

// HandHistory 手牌历史
type HandHistory struct {
    HandID         int
    Timestamp      time.Time
    GameID         string
    Players        []HistoryPlayer
    CommunityCards [5]card.Card
    Actions        []HistoryAction
    Showdown       []ShowdownInfo
    Pot            int
    WinnerIDs      []string
}

// HistoryPlayer 玩家历史
type HistoryPlayer struct {
    ID         string
    Name       string
    Seat       int
    HoleCards  [2]card.Card
    FinalChips int
    WonChips   int
}

// HistoryAction 行动历史
type HistoryAction struct {
    PlayerID   string
    PlayerName string
    Round     protocol.GameStage
    Action    protocol.ActionType
    Amount    int
    Timestamp time.Time
}

// ShowdownInfo 摊牌信息
type ShowdownInfo struct {
    PlayerID   string
    PlayerName string
    HandRank   evaluator.HandRank
    HandName   string
    BestCards  []card.Card
}

// HistoryManager 历史管理器
type HistoryManager struct {
    hands      []HandHistory
    currentHand *HandHistory
    file       *os.File
    filename   string
}

// NewHistoryManager 创建历史管理器
func NewHistoryManager(filename string) (*HistoryManager, error) {
    h := &HistoryManager{
        hands:    make([]HandHistory, 0),
        filename: filename,
    }

    if filename != "" {
        if file, err := os.Open(filename); err == nil {
            defer file.Close()
            decoder := json.NewDecoder(file)
            decoder.Decode(&h.hands)
        }
    }

    return h, nil
}

// StartHand 开始记录新手牌
func (h *HistoryManager) StartHand(handID int, gameID string) {
    h.currentHand = &HandHistory{
        HandID:    handID,
        Timestamp: time.Now(),
        GameID:    gameID,
        Players:   make([]HistoryPlayer, 0),
        Actions:   make([]HistoryAction, 0),
        Showdown:  make([]ShowdownInfo, 0),
    }
}

// AddPlayer 添加玩家
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

// EndHand 结束手牌
func (h *HistoryManager) EndHand(community [5]card.Card, pot int, winners []string) {
    if h.currentHand == nil {
        return
    }

    h.currentHand.CommunityCards = community
    h.currentHand.Pot = pot
    h.currentHand.WinnerIDs = winners

    h.hands = append(h.hands, *h.currentHand)
    h.currentHand = nil

    if h.filename != "" {
        h.saveToFile()
    }
}

// saveToFile 保存到文件
func (h *HistoryManager) saveToFile() {
    if h.filename == "" {
        return
    }

    file, _ := os.OpenFile(h.filename, os.O_WRONLY|os.O_CREATE, 0644)
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    encoder.Encode(h.hands)
}

// GetRecentHands 获取最近N手牌
func (h *HistoryManager) GetRecentHands(n int) []HandHistory {
    if n <= 0 {
        n = 10
    }
    if n > len(h.hands) {
        n = len(h.hands)
    }
    return h.hands[len(h.hands)-n:]
}

// ExportToText 导出为文本
func (h *HandHistory) ExportToText() string {
    var b strings.Builder

    b.WriteString(fmt.Sprintf("=== 手牌 #%d ===\n", h.HandID))
    b.WriteString(fmt.Sprintf("时间: %s\n", h.Timestamp.Format("2006-01-02 15:04:05")))
    b.WriteString(fmt.Sprintf("底池: %d\n\n", h.Pot))

    b.WriteString("玩家:\n")
    for _, p := range h.Players {
        cards := ""
        if p.HoleCards[0].Rank != 0 {
            cards = fmt.Sprintf(" 底牌: %s %s", p.HoleCards[0].String(), p.HoleCards[1].String())
        }
        b.WriteString(fmt.Sprintf("  [%d] %s: %d%s\n", p.Seat+1, p.Name, p.FinalChips, cards))
    }

    b.WriteString("\n行动:\n")
    for _, a := range h.Actions {
        amountStr := ""
        if a.Amount > 0 {
            amountStr = fmt.Sprintf(" (%d)", a.Amount)
        }
        b.WriteString(fmt.Sprintf("  %s: %s%s\n", a.PlayerName, a.Action.String(), amountStr))
    }

    b.WriteString("\n摊牌:\n")
    for _, s := range h.Showdown {
        cards := ""
        for _, c := range s.BestCards {
            cards += c.String() + " "
        }
        b.WriteString(fmt.Sprintf("  %s: %s (%s)\n", s.PlayerName, s.HandName, cards))
    }

    return b.String()
}
```

---

## 任务 3.3: 扑克牌渲染组件

```go
// ui/components/card_render.go

package components

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "yourproject/card"
)

var (
    suitRed    = lipgloss.Color("196")
    suitBlack  = lipgloss.Color("15")
    backColor  = lipgloss.Color("239")
)

// RenderCard 渲染单张牌
func RenderCard(c card.Card, faceUp bool) string {
    if !faceUp || c.Rank == 0 {
        return renderCardBack()
    }

    suitColor := suitBlack
    if c.Suit == card.Hearts || c.Suit == card.Diamonds {
        suitColor = suitRed
    }

    rankStr := rankToString(c.Rank)
    suitStr := suitToString(c.Suit)

    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Width(4).Height(3).
        Render(
            lipgloss.NewStyle().
                Foreground(suitColor).
                Bold(true).
                Render(rankStr + "\n" + suitStr),
        )
}

// renderCardBack 渲染牌背
func renderCardBack() string {
    return lipgloss.NewStyle().
        Background(backColor).
        Width(4).Height(3).
        Render("??")
}

// RenderCards 渲染多张牌
func RenderCards(cards []card.Card, faceUp bool) string {
    var rendered []string
    for _, c := range cards {
        rendered = append(rendered, RenderCard(c, faceUp))
    }
    return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// 辅助函数
func rankToString(r card.Rank) string {
    names := map[card.Rank]string{
        card.Two:   "2",
        card.Three: "3",
        card.Four:  "4",
        card.Five:  "5",
        card.Six:   "6",
        card.Seven: "7",
        card.Eight: "8",
        card.Nine:  "9",
        card.Ten:   "10",
        card.Jack:  "J",
        card.Queen: "Q",
        card.King:  "K",
        card.Ace:   "A",
    }
    return names[r]
}

func suitToString(s card.Suit) string {
    names := map[card.Suit]string{
        card.Clubs:    "C",
        card.Diamonds: "D",
        card.Hearts:   "H",
        card.Spades:   "S",
    }
    return names[s]
}

// 简化ASCII版本 (无颜色终端)
func RenderCardASCII(c card.Card, faceUp bool) string {
    if !faceUp || c.Rank == 0 {
        return "[??]"
    }
    return fmt.Sprintf("[%s%s]", rankToString(c.Rank), suitToString(c.Suit))
}
```

---

## 任务 3.4: 游戏统计

```go
// game/stats.go

package game

import (
    "fmt"
    "strings"
    "sync"
    "time"
)

// PlayerStats 玩家统计
type PlayerStats struct {
    PlayerID      string
    Name          string
    HandsPlayed   int
    HandsWon      int
    TotalWinnings int
    WinRate       float64
    BiggestPot    int
}

// StatsManager 统计管理器
type StatsManager struct {
    mu           sync.RWMutex
    playerStats map[string]*PlayerStats
}

// NewStatsManager 创建统计管理器
func NewStatsManager() *StatsManager {
    return &StatsManager{
        playerStats: make(map[string]*PlayerStats),
    }
}

// Update 更新玩家统计
func (s *StatsManager) Update(playerID, name string, wonAmount int) {
    s.mu.Lock()
    defer s.mu.Unlock()

    stats, ok := s.playerStats[playerID]
    if !ok {
        stats = &PlayerStats{
            PlayerID: playerID,
            Name:     name,
        }
        s.playerStats[playerID] = stats
    }

    stats.HandsPlayed++
    if wonAmount > 0 {
        stats.HandsWon++
        stats.TotalWinnings += wonAmount
        if wonAmount > stats.BiggestPot {
            stats.BiggestPot = wonAmount
        }
    }

    if stats.HandsPlayed > 0 {
        stats.WinRate = float64(stats.TotalWinnings) / float64(stats.HandsPlayed)
    }
}

// GetStats 获取玩家统计
func (s *StatsManager) GetStats(playerID string) *PlayerStats {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.playerStats[playerID]
}

// Report 生成统计报告
func (p *PlayerStats) Report() string {
    return fmt.Sprintf(`
=== %s 统计 ===
手牌数: %d
获胜手牌: %d (%.1f%%)
总盈亏: %d
每百手盈亏: %.1f
最大底池: %d
`, p.Name, p.HandsPlayed, p.HandsWon,
        float64(p.HandsWon)/float64(p.HandsPlayed)*100,
        p.TotalWinnings, p.WinRate, p.BiggestPot)
}
```

---

## 阶段验收清单

- [ ] AI玩家可以参与游戏
- [ ] 历史记录正确保存和显示
- [ ] 牌面渲染美观
- [ ] 游戏统计可用
- [ ] 整体功能稳定
