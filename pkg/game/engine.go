package game

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/pkg/evaluator"
)

// GameEngine 管理扑克游戏的状态和逻辑
type GameEngine struct {
	state     *GameState       // 游戏状态
	config    *Config         // 游戏配置
	evaluator *evaluator.Evaluator // 手牌评估器
	deck      *card.Deck      // 牌组
	rand      *rand.Rand      // 随机数生成器
	mutex     sync.RWMutex    // 读写锁

	// 状态变化回调
	onStateChange func(state *GameState)
}

// Config 保存游戏配置
type Config struct {
	MinPlayers     int // 最少玩家数
	MaxPlayers     int // 最多玩家数
	SmallBlind     int // 小盲注金额
	BigBlind       int // 大盲注金额
	StartingChips  int // 初始筹码
	ActionTimeout  int // 动作超时时间
}

// GameState 表示当前的游戏状态
type GameState struct {
	ID             string               // 游戏ID
	Stage          Stage               // 当前阶段
	DealerButton   int                 // 庄家按钮位置
	CurrentPlayer  int                 // 当前行动玩家索引
	CurrentBet     int                 // 当前最高下注
	Pot            int                 // 底池金额
	SidePots       []SidePot           // 边池
	CommunityCards [5]card.Card       // 公共牌
	Players        []*models.Player    // 所有玩家
	Actions        []models.PlayerAction // 动作记录
}

// Stage 表示当前的下注阶段
type Stage int

const (
	StageWaiting Stage = iota // 等待开始
	StagePreFlop              // 翻牌前
	StageFlop                 // 翻牌圈
	StageTurn                 // 转牌圈
	StageRiver                // 河牌圈
	StageShowdown             // 摊牌
	StageEnd                  // 局结束
)

// 阶段名称
var stageNames = []string{
	"等待开始", "翻牌前", "翻牌圈", "转牌圈", "河牌圈", "摊牌", "局结束",
}

// String 返回阶段名称
func (s Stage) String() string {
	if s >= 0 && int(s) < len(stageNames) {
		return stageNames[s]
	}
	return "未知"
}

// SidePot 表示边池
type SidePot struct {
	Amount          int   // 边池金额
	EligiblePlayers []int // 有资格获得边池的玩家索引列表
}

// NewEngine 创建新的游戏引擎
func NewEngine(config *Config) *GameEngine {
	if config.MinPlayers < 2 {
		config.MinPlayers = 2
	}
	if config.MaxPlayers > 9 {
		config.MaxPlayers = 9
	}

	engine := &GameEngine{
		state: &GameState{
			ID:       fmt.Sprintf("game_%d", time.Now().Unix()),
			Stage:    StageWaiting,
			Players:  make([]*models.Player, 0),
			Pot:      0,
			SidePots: make([]SidePot, 0),
		},
		config:    config,
		evaluator: evaluator.NewEvaluator(),
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return engine
}

// SetOnStateChange 设置状态变化回调函数
func (e *GameEngine) SetOnStateChange(fn func(state *GameState)) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.onStateChange = fn
}

// GetState 获取当前游戏状态（线程安全）
func (e *GameEngine) GetState() *GameState {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.copyState()
}

// AddPlayer 添加玩家到游戏
func (e *GameEngine) AddPlayer(id, name string, seat int) (*models.Player, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if len(e.state.Players) >= e.config.MaxPlayers {
		return nil, ErrGameFull
	}

	if seat < 0 || seat >= e.config.MaxPlayers {
		return nil, ErrInvalidSeat
	}

	// 检查座位是否已被占用
	for _, p := range e.state.Players {
		if p.Seat == seat {
			return nil, ErrSeatOccupied
		}
	}

	player := &models.Player{
		ID:    id,
		Name:  name,
		Chips: e.config.StartingChips,
		Seat:  seat,
		Status: models.PlayerStatusActive,
	}

	e.state.Players = append(e.state.Players, player)
	e.notifyStateChange()

	return player, nil
}

// RemovePlayer 从游戏中移除玩家
func (e *GameEngine) RemovePlayer(id string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for i, p := range e.state.Players {
		if p.ID == id {
			if e.state.Stage != StageWaiting && p.Status == models.PlayerStatusActive {
				p.Status = models.PlayerStatusFolded
			} else {
				e.state.Players = append(e.state.Players[:i], e.state.Players[i+1:]...)
			}
			e.notifyStateChange()
			return nil
		}
	}
	return ErrPlayerNotFound
}

// StartHand 开始新的一局
func (e *GameEngine) StartHand() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.state.Stage != StageWaiting {
		return ErrHandInProgress
	}

	activePlayers := e.getActivePlayers()
	if len(activePlayers) < 2 {
		return ErrNotEnoughPlayers
	}

	// 准备新局
	e.state.Stage = StagePreFlop
	e.state.CommunityCards = [5]card.Card{}
	e.state.Actions = make([]models.PlayerAction, 0)
	e.state.Pot = 0
	e.state.SidePots = make([]SidePot, 0)

	// 洗牌
	e.deck = card.NewDeck()
	e.deck.Shuffle()

	// 轮转庄家按钮
	e.rotateDealerButton()

	// 重置玩家状态
	for _, p := range e.state.Players {
		p.HoleCards = [2]card.Card{}
		p.CurrentBet = 0
		p.HasActed = false
		if p.Chips <= 0 {
			p.Status = models.PlayerStatusFolded
		}
	}

	// 扣除盲注
	e.collectBlinds()

	// 发底牌
	e.dealHoleCards()

	e.notifyStateChange()
	return nil
}

// PlayerAction 处理玩家的动作
func (e *GameEngine) PlayerAction(playerID string, action models.ActionType, amount int) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	player := e.getPlayerByID(playerID)
	if player == nil {
		return ErrPlayerNotFound
	}

	if e.state.CurrentPlayer >= len(e.state.Players) {
		return ErrNotYourTurn
	}

	currentPlayer := e.state.Players[e.state.CurrentPlayer]
	if currentPlayer.ID != playerID {
		return ErrNotYourTurn
	}

	// 验证动作是否合法
	if err := e.validateAction(player, action, amount); err != nil {
		return err
	}

	// 执行动作
	switch action {
	case models.ActionFold:
		player.Status = models.PlayerStatusFolded

	case models.ActionCheck:
		// 看牌无需操作

	case models.ActionCall:
		callAmount := e.state.CurrentBet - player.CurrentBet
		player.Chips -= callAmount
		player.CurrentBet += callAmount
		e.state.Pot += callAmount

	case models.ActionRaise:
		raiseAmount := amount - player.CurrentBet
		player.Chips -= raiseAmount
		player.CurrentBet += raiseAmount
		e.state.Pot += raiseAmount
		e.state.CurrentBet = amount

	case models.ActionAllIn:
		allIn := player.Chips
		player.Chips = 0
		player.CurrentBet += allIn
		e.state.Pot += allIn
		player.Status = models.PlayerStatusAllIn
		if player.CurrentBet > e.state.CurrentBet {
			e.state.CurrentBet = player.CurrentBet
		}
	}

	player.HasActed = true

	// 记录动作
	e.state.Actions = append(e.state.Actions, models.PlayerAction{
		PlayerID: playerID,
		Action:   action,
		Amount:   player.CurrentBet,
	})

	// 检查下注轮是否结束
	if e.isBettingRoundComplete() {
		e.advanceBettingRound()
	} else {
		e.nextPlayer()
	}

	e.notifyStateChange()
	return nil
}

// ==================== 私有方法 ====================

// rotateDealerButton 轮转庄家按钮
func (e *GameEngine) rotateDealerButton() {
	currentBtn := -1
	for i, p := range e.state.Players {
		if p.IsDealer {
			currentBtn = i
			break
		}
	}

	// 找到下一位活跃玩家
	for i := 1; i <= len(e.state.Players); i++ {
		nextIdx := (currentBtn + i) % len(e.state.Players)
		if e.state.Players[nextIdx].Status == models.PlayerStatusActive {
			e.state.Players[nextIdx].IsDealer = true
			e.state.DealerButton = nextIdx
			if currentBtn >= 0 {
				e.state.Players[currentBtn].IsDealer = false
			}
			break
		}
	}
}

// collectBlinds 扣除盲注
func (e *GameEngine) collectBlinds() {
	dealerIdx := e.state.DealerButton
	sbIdx, bbIdx := -1, -1

	// 找到小盲和大盲位置
	for i := 1; i <= len(e.state.Players); i++ {
		idx := (dealerIdx + i) % len(e.state.Players)
		if e.state.Players[idx].Status == models.PlayerStatusActive {
			if sbIdx < 0 {
				sbIdx = idx
			} else {
				bbIdx = idx
				break
			}
		}
	}

	// 扣除小盲
	if sbIdx >= 0 {
		sb := e.state.Players[sbIdx]
		sbAmount := min(sb.Chips, e.config.SmallBlind)
		sb.Chips -= sbAmount
		sb.CurrentBet = sbAmount
		e.state.Pot += sbAmount
	}

	// 扣除大盲
	if bbIdx >= 0 {
		bb := e.state.Players[bbIdx]
		bbAmount := min(bb.Chips, e.config.BigBlind)
		bb.Chips -= bbAmount
		bb.CurrentBet = bbAmount
		e.state.Pot += bbAmount
		e.state.CurrentBet = bbAmount
	}
}

// dealHoleCards 发底牌
func (e *GameEngine) dealHoleCards() {
	e.deck.Burn(1) // 弃掉一张牌

	for _, p := range e.state.Players {
		if p.Status == models.PlayerStatusActive {
			cards, _ := e.deck.DealN(2)
			p.HoleCards = [2]card.Card{cards[0], cards[1]}
		}
	}
}

// validateAction 验证玩家动作是否合法
func (e *GameEngine) validateAction(p *models.Player, action models.ActionType, amount int) error {
	switch action {
	case models.ActionFold:
		return nil
	case models.ActionCheck:
		if e.state.CurrentBet > p.CurrentBet {
			return ErrCannotCheck
		}
		return nil
	case models.ActionCall:
		callAmount := e.state.CurrentBet - p.CurrentBet
		if callAmount > p.Chips {
			return ErrNotEnoughChips
		}
		return nil
	case models.ActionRaise:
		minRaise := e.state.CurrentBet * 2
		if amount < minRaise {
			return fmt.Errorf("minimum raise is %d", minRaise)
		}
		if amount > p.Chips+p.CurrentBet {
			return ErrNotEnoughChips
		}
		return nil
	case models.ActionAllIn:
		return nil
	}
	return ErrInvalidAction
}

// isBettingRoundComplete 检查下注轮是否结束
func (e *GameEngine) isBettingRoundComplete() bool {
	activePlayers := 0
	actedPlayers := 0

	for _, p := range e.state.Players {
		if p.Status == models.PlayerStatusActive {
			activePlayers++
			if p.HasActed {
				actedPlayers++
			}
		}
	}

	return actedPlayers >= activePlayers
}

// nextPlayer 轮到下一位玩家
func (e *GameEngine) nextPlayer() {
	for i := 1; i <= len(e.state.Players); i++ {
		nextIdx := (e.state.CurrentPlayer + i) % len(e.state.Players)
		if e.state.Players[nextIdx].Status == models.PlayerStatusActive {
			e.state.CurrentPlayer = nextIdx
			return
		}
	}
}

// advanceBettingRound 进入下一轮下注
func (e *GameEngine) advanceBettingRound() {
	// 重置所有玩家的行动状态
	for _, p := range e.state.Players {
		p.HasActed = false
	}
	e.state.CurrentBet = 0

	switch e.state.Stage {
	case StagePreFlop:
		// 翻牌：发3张公共牌
		e.state.Stage = StageFlop
		e.deck.Burn(1)
		flop, _ := e.deck.DealN(3)
		e.state.CommunityCards = [5]card.Card{flop[0], flop[1], flop[2]}
		e.state.CurrentPlayer = e.findFirstToAct()

	case StageFlop:
		// 转牌：发第4张公共牌
		e.state.Stage = StageTurn
		e.deck.Burn(1)
		turn, _ := e.deck.DealN(1)
		e.state.CommunityCards[3] = turn[0]
		e.state.CurrentPlayer = e.findFirstToAct()

	case StageTurn:
		// 河牌：发第5张公共牌
		e.state.Stage = StageRiver
		e.deck.Burn(1)
		river, _ := e.deck.DealN(1)
		e.state.CommunityCards[4] = river[0]
		e.state.CurrentPlayer = e.findFirstToAct()

	case StageRiver:
		// 河牌结束，进入摊牌
		e.state.Stage = StageShowdown
		e.determineWinners()
	}
}

// findFirstToAct 找到庄家后第一位需要行动的玩家
func (e *GameEngine) findFirstToAct() int {
	for i := 1; i <= len(e.state.Players); i++ {
		idx := (e.state.DealerButton + i) % len(e.state.Players)
		if e.state.Players[idx].Status == models.PlayerStatusActive {
			return idx
		}
	}
	return 0
}

// determineWinners 判定获胜者并分配底池
func (e *GameEngine) determineWinners() {
	var bestEval evaluator.HandEvaluation
	var bestPlayerIdx int = -1
	ties := []int{} // 平局玩家列表

	for i, p := range e.state.Players {
		if p.Status == models.PlayerStatusActive {
			eval := e.evaluator.Evaluate(p.HoleCards, e.state.CommunityCards)
			if bestPlayerIdx < 0 {
				bestEval = eval
				bestPlayerIdx = i
			} else {
				cmp := e.evaluator.Compare(eval, bestEval)
				if cmp > 0 {
					bestEval = eval
					bestPlayerIdx = i
					ties = []int{i}
				} else if cmp == 0 {
					ties = append(ties, i)
				}
			}
		}
	}

	// 分配底池
	if len(ties) > 0 {
		// 平分底池
		share := e.state.Pot / len(ties)
		remainder := e.state.Pot % len(ties)
		for _, idx := range ties {
			e.state.Players[idx].Chips += share
			if remainder > 0 {
				e.state.Players[idx].Chips++
				remainder--
			}
		}
	} else if bestPlayerIdx >= 0 {
		// 单人获胜
		e.state.Players[bestPlayerIdx].Chips += e.state.Pot
	}

	e.state.Pot = 0
}

// getActivePlayers 获取所有活跃玩家
func (e *GameEngine) getActivePlayers() []*models.Player {
	active := make([]*models.Player, 0)
	for _, p := range e.state.Players {
		if p.Status == models.PlayerStatusActive {
			active = append(active, p)
		}
	}
	return active
}

// getPlayerByID 根据ID获取玩家
func (e *GameEngine) getPlayerByID(id string) *models.Player {
	for _, p := range e.state.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// copyState 复制游戏状态（用于返回给外部）
func (e *GameEngine) copyState() *GameState {
	copy := *e.state
	copy.Players = make([]*models.Player, len(e.state.Players))
	for i, p := range e.state.Players {
		playerCopy := *p
		copy.Players[i] = &playerCopy
	}
	return &copy
}

// notifyStateChange 通知状态变化
func (e *GameEngine) notifyStateChange() {
	if e.onStateChange != nil {
		go e.onStateChange(e.copyState())
	}
}

// min 取两个数的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ==================== 错误定义 ====================
var (
	ErrGameFull         = errors.New("游戏已满")
	ErrInvalidSeat      = errors.New("无效座位")
	ErrSeatOccupied     = errors.New("座位已被占用")
	ErrHandInProgress   = errors.New("手牌进行中")
	ErrNotEnoughPlayers = errors.New("玩家不足")
	ErrNotYourTurn      = errors.New("还未轮到您")
	ErrCannotCheck      = errors.New("无法看牌")
	ErrNotEnoughChips   = errors.New("筹码不足")
	ErrInvalidAction    = errors.New("无效动作")
	ErrPlayerNotFound   = errors.New("玩家不存在")
)
