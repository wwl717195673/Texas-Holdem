# Phase 1: 核心游戏引擎实现

## 阶段目标

完成德州扑克的核心游戏逻辑，包括牌组管理、发牌、牌型评估、下注逻辑。不包含网络通信，专注于单机游戏功能验证。

## 交付物

| 组件 | 文件 | 优先级 |
|-----|------|-------|
| 扑克牌 | `internal/card/card.go` | P0 |
| 牌组 | `internal/card/deck.go` | P0 |
| 玩家模型 | `internal/common/models/player.go` | P0 |
| 牌型评估器 | `pkg/evaluator/evaluator.go` | P0 |
| 游戏引擎 | `pkg/game/engine.go` | P0 |
| 单元测试 | `*_test.go` | P1 |

## 详细任务

### 任务 1.1: 扑克牌基础实现

**目标**: 实现牌的表示、比较和渲染

```go
// internal/card/card.go

package card

import "fmt"

// 花色常量
const (
    Clubs Suit = iota
    Diamonds
    Hearts
    Spades
)

// 花色名称
var suitNames = []string{"♣", "♦", "♥", "♠"}
var suitFullNames = []string{"梅花", "方块", "红心", "黑桃"}

func (s Suit) String() string {
    if s >= 0 && int(s) < len(suitNames) {
        return suitNames[s]
    }
    return "?"
}

func (s Suit) FullName() string {
    if s >= 0 && int(s) < len(suitFullNames) {
        return suitFullNames[s]
    }
    return "未知"
}

// 点数常量
const (
    Two Rank = iota + 2
    Three
    Four
    Five
    Six
    Seven
    Eight
    Nine
    Ten
    Jack
    Queen
    King
    Ace
)

// 点数名称
var rankNames = []string{
    "", "", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A",
}

var rankSymbols = []string{
    "", "", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A",
}

func (r Rank) String() string {
    if r >= 2 && int(r) < len(rankSymbols) {
        return rankSymbols[r]
    }
    return "?"
}

func (r Rank) FullName() string {
    if r >= 2 && int(r) < len(rankNames) {
        return rankNames[r]
    }
    return "未知"
}

func (r Rank) Value() int {
    return int(r)
}

// 牌结构
type Card struct {
    Suit Suit
    Rank Rank
}

// 创建一个牌
func NewCard(suit Suit, rank Rank) *Card {
    return &Card{Suit: suit, Rank: rank}
}

// 牌的字符串表示
func (c Card) String() string {
    return c.Rank.String() + c.Suit.String()
}

// 用于排序的比较
func (c Card) Compare(other Card) int {
    if c.Rank != other.Rank {
        if c.Rank > other.Rank {
            return 1
        }
        return -1
    }
    return 0
}

// 是否为黑桃 (用于渲染颜色)
func (c Card) IsBlack() bool {
    return c.Suit == Spades || c.Suit == Clubs
}
```

**验收标准**:
- [ ] 创建52张牌正确
- [ ] 花色和点数显示正确
- [ ] 牌的比较功能正常

### 任务 1.2: 牌组管理

**目标**: 实现牌组创建、洗牌、发牌功能

```go
// internal/card/deck.go

package card

import (
    "math/rand"
    "time"
)

type Deck struct {
    cards []Card
    index int
}

// 创建新牌组
func NewDeck() *Deck {
    d := &Deck{
        cards: make([]Card, 0, 52),
    }
    // 按花色分组创建
    for suit := Clubs; suit <= Spades; suit++ {
        for rank := Two; rank <= Ace; rank++ {
            d.cards = append(d.cards, Card{Suit: suit, Rank: rank})
        }
    }
    return d
}

// 使用指定种子创建牌组 (用于测试)
func NewDeckWithSeed(seed int64) *Deck {
    d := NewDeck()
    d.ShuffleWithSeed(seed)
    return d
}

// 洗牌 (使用随机种子)
func (d *Deck) Shuffle() {
    d.ShuffleWithSeed(time.Now().UnixNano())
}

// 使用指定种子洗牌
func (d *Deck) ShuffleWithSeed(seed int64) {
    r := rand.New(rand.NewSource(seed))
    for i := len(d.cards) - 1; i > 0; i-- {
        j := r.Intn(i + 1)
        d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
    }
    d.index = 0
}

// 发一张牌
func (d *Deck) Deal() (Card, error) {
    if d.index >= len(d.cards) {
        return Card{}, ErrNoCardsLeft
    }
    card := d.cards[d.index]
    d.index++
    return card, nil
}

// 发多张牌
func (d *Deck) DealN(n int) ([]Card, error) {
    if d.index+n > len(d.cards) {
        return nil, ErrNoCardsLeft
    }
    cards := make([]Card, n)
    for i := 0; i < n; i++ {
        cards[i] = d.cards[d.index]
        d.index++
    }
    return cards, nil
}

// 弃牌 (不发,跳过)
func (d *Deck) Burn(n int) error {
    if d.index+n > len(d.cards) {
        return ErrNoCardsLeft
    }
    d.index += n
    return nil
}

// 剩余牌数
func (d *Deck) Remaining() int {
    return len(d.cards) - d.index
}

// 重置牌组
func (d *Deck) Reset() {
    d.index = 0
    // 恢复所有牌
    d.cards = make([]Card, 0, 52)
    for suit := Clubs; suit <= Spades; suit++ {
        for rank := Two; rank <= Ace; rank++ {
            d.cards = append(d.cards, Card{Suit: suit, Rank: rank})
        }
    }
}

// 错误
var ErrNoCardsLeft = &DeckError{"no cards left in deck"}

type DeckError struct {
    msg string
}

func (e *DeckError) Error() string {
    return e.msg
}
```

**验收标准**:
- [ ] 牌组包含52张牌
- [ ] 洗牌后牌顺序随机
- [ ] 发牌后正确更新索引
- [ ] 弃牌功能正常

### 任务 1.3: 玩家模型

```go
// internal/common/models/player.go

package models

import (
    "yourproject/card"
)

// 玩家状态
type PlayerStatus int

const (
    PlayerStatusInactive PlayerStatus = iota  // 未入座
    PlayerStatusActive                       // 游戏中
    PlayerStatusFolded                      // 已弃牌
    PlayerStatusAllIn                       // 全下
)

// 玩家行动类型
type ActionType int

const (
    ActionFold ActionType = iota
    ActionCheck
    ActionCall
    ActionRaise
    ActionAllIn
)

func (a ActionType) String() string {
    names := []string{"弃牌", "看牌", "跟注", "加注", "全下"}
    if int(a) < len(names) {
        return names[a]
    }
    return "未知"
}

// 玩家
type Player struct {
    ID          string
    Name        string
    Chips       int
    Seat        int
    Status      PlayerStatus
    HoleCards   [2]card.Card  // 底牌
    CurrentBet  int
    IsDealer    bool
    HasActed    bool

    // 统计
    HandsPlayed  int
    HandsWon     int
}

// 创建新玩家
func NewPlayer(id, name string, chips int, seat int) *Player {
    return &Player{
        ID:    id,
        Name:  name,
        Chips: chips,
        Seat:  seat,
        Status: PlayerStatusActive,
    }
}

// 玩家是否活跃 (未弃牌且未全下)
func (p *Player) IsActive() bool {
    return p.Status == PlayerStatusActive
}

// 玩家是否可以行动
func (p *Player) CanAct() bool {
    return p.IsActive() && !p.HasActed
}

// 获取底牌显示
func (p *Player) GetHoleCardsDisplay() string {
    if p.HoleCards[0].Rank == 0 {
        return "[  ?  ][  ?  ]"
    }
    return p.HoleCards[0].String() + " " + p.HoleCards[1].String()
}
```

### 任务 1.4: 牌型评估器

这是最核心的组件,实现10种牌型的识别和比较。

**实现策略**:使用位运算和查表法优化评估性能

```go
// pkg/evaluator/evaluator.go

package evaluator

import (
    "sort"
    "yourproject/card"
)

// 牌型等级 (1=最高,10=最低)
type HandRank int

const (
    RankHighCard HandRank = iota + 1
    RankOnePair
    RankTwoPair
    RankThreeOfAKind
    RankStraight
    RankFlush
    RankFullHouse
    RankFourOfAKind
    RankStraightFlush
    RankRoyalFlush
)

// 牌型名称
var rankNames = []string{
    "高牌", "一对", "两对", "三条", "顺子",
    "同花", "葫芦", "四条", "同花顺", "皇家同花顺",
}

var rankSymbols = []string{
    "HC", "1P", "2P", "3K", "ST",
    "FL", "FH", "4K", "SF", "RF",
}

func (r HandRank) String() string {
    if r >= 1 && int(r) <= len(rankNames) {
        return rankNames[r-1]
    }
    return "未知"
}

func (r HandRank) Symbol() string {
    if r >= 1 && int(r) <= len(rankSymbols) {
        return rankSymbols[r-1]
    }
    return "?"
}

// 评估结果
type HandEvaluation struct {
    Rank      HandRank
    MainValue int      // 主要比较值
    Kickers   []int   // 踢脚牌
    RawCards  []card.Card // 参与比较的5张牌
}

// 评估一手牌
func (e *Evaluator) Evaluate(holeCards [2]card.Card, communityCards [5]card.Card) HandEvaluation {
    // 收集所有7张牌
    allCards := make([]card.Card, 0, 7)
    allCards = append(allCards, holeCards[:]...)
    allCards = append(allCards, communityCards[:]...)

    return e.evaluate7Cards(allCards)
}

// 评估7张牌的最佳5张组合
func (e *Evaluator) evaluate7Cards(cards []card.Card) HandEvaluation {
    // 统计信息
    suitGroups := make(map[card.Suit][]card.Card)
    rankCounts := make(map[card.Rank]int)
    ranks := make([]card.Rank, 0, 7)

    for _, c := range cards {
        suitGroups[c.Suit] = append(suitGroups[c.Suit], c)
        rankCounts[c.Rank]++
        ranks = append(ranks, c.Rank)
    }

    // 排序 (从大到小)
    sort.Slice(cards, func(i, j int) bool {
        return cards[i].Rank > cards[j].Rank
    })

    // 从高到低检查牌型
    if eval := e.checkRoyalFlush(suitGroups); eval.Rank > 0 {
        return eval
    }
    if eval := e.checkStraightFlush(cards, suitGroups); eval.Rank > 0 {
        return eval
    }
    if eval := e.checkFourOfAKind(rankCounts, cards); eval.Rank > 0 {
        return eval
    }
    if eval := e.checkFullHouse(rankCounts); eval.Rank > 0 {
        return eval
    }
    if eval := e.checkFlush(suitGroups); eval.Rank > 0 {
        return eval
    }
    if eval := e.checkStraight(ranks); eval.Rank > 0 {
        return eval
    }
    if eval := e.checkThreeOfAKind(rankCounts, cards); eval.Rank > 0 {
        return eval
    }
    if eval := e.checkTwoPair(rankCounts, cards); eval.Rank > 0 {
        return eval
    }
    if eval := e.checkOnePair(rankCounts, cards); eval.Rank > 0 {
        return eval
    }

    return e.checkHighCard(cards)
}

// 检查皇家同花顺
func (e *Evaluator) checkRoyalFlush(suitGroups map[card.Suit][]card.Card) HandEvaluation {
    for _, cards := range suitGroups {
        if len(cards) >= 5 {
            ranks := make([]card.Rank, 0)
            for _, c := range cards {
                ranks = append(ranks, c.Rank)
            }
            // 检查是否包含10,J,Q,K,A
            has := make(map[card.Rank]bool)
            for _, r := range ranks {
                has[r] = true
            }
            if has[card.Ten] && has[card.Jack] && has[card.Queen] &&
               has[card.King] && has[card.Ace] {
                return HandEvaluation{
                    Rank:     RankRoyalFlush,
                    MainValue: 14,
                }
            }
        }
    }
    return HandEvaluation{}
}

// 检查同花顺 (不含皇家同花顺)
func (e *Evaluator) checkStraightFlush(cards []card.Card, suitGroups map[card.Suit][]card.Card) HandEvaluation {
    for _, flushCards := range suitGroups {
        if len(flushCards) >= 5 {
            ranks := make([]card.Rank, 0)
            for _, c := range flushCards {
                ranks = append(ranks, c.Rank)
            }
            if eval := e.evaluateStraight(ranks); eval.Rank > 0 {
                // 找出5张顺子牌
                var straightCards []card.Card
                for _, c := range cards {
                    if containsRank(ranks, c.Rank) && c.Suit == flushCards[0].Suit {
                        straightCards = append(straightCards, c)
                        if len(straightCards) == 5 {
                            break
                        }
                    }
                }
                eval.RawCards = straightCards
                return eval
            }
        }
    }
    return HandEvaluation{}
}

// 检查四条
func (e *Evaluator) checkFourOfAKind(rankCounts map[card.Rank]int, sorted []card.Card) HandEvaluation {
    for rank, count := range rankCounts {
        if count == 4 {
            kickers := make([]int, 0)
            for _, c := range sorted {
                if c.Rank != rank {
                    kickers = append(kickers, int(c.Rank))
                    if len(kickers) == 1 {
                        break
                    }
                }
            }
            return HandEvaluation{
                Rank:      RankFourOfAKind,
                MainValue: int(rank),
                Kickers:   kickers,
            }
        }
    }
    return HandEvaluation{}
}

// 检查葫芦
func (e *Evaluator) checkFullHouse(rankCounts map[card.Rank]int) HandEvaluation {
    var threeRanks []card.Rank
    var pairRanks []card.Rank

    for rank, count := range rankCounts {
        if count == 3 {
            threeRanks = append(threeRanks, rank)
        } else if count == 2 {
            pairRanks = append(pairRanks, rank)
        }
    }

    if len(threeRanks) >= 2 {
        // 有两组三条,取较大的作为三条
        sort.Slice(threeRanks, func(i, j int) bool {
            return threeRanks[i] > threeRanks[j]
        })
        three := threeRanks[0]
        // 取另一组三条作为对子
        threeAsPair := threeRanks[1]

        return HandEvaluation{
            Rank:      RankFullHouse,
            MainValue: int(three),
            Kickers:   []int{int(threeAsPair)},
        }
    }

    if len(threeRanks) >= 1 && len(pairRanks) >= 1 {
        three := threeRanks[0]
        sort.Slice(pairRanks, func(i, j int) bool {
            return pairRanks[i] > pairRanks[j]
        })
        pair := pairRanks[0]

        return HandEvaluation{
            Rank:      RankFullHouse,
            MainValue: int(three),
            Kickers:   []int{int(pair)},
        }
    }

    return HandEvaluation{}
}

// 检查同花
func (e *Evaluator) checkFlush(suitGroups map[card.Suit][]card.Card) HandEvaluation {
    for _, cards := range suitGroups {
        if len(cards) >= 5 {
            sort.Slice(cards, func(i, j int) bool {
                return cards[i].Rank > cards[j].Rank
            })
            best5 := cards[:5]
            kickers := make([]int, 5)
            for i, c := range best5 {
                kickers[i] = int(c.Rank)
            }
            return HandEvaluation{
                Rank:      RankFlush,
                MainValue: kickers[0],
                Kickers:   kickers[1:],
                RawCards:  best5,
            }
        }
    }
    return HandEvaluation{}
}

// 检查顺子
func (e *Evaluator) checkStraight(ranks []card.Rank) HandEvaluation {
    unique := make([]card.Rank, 0)
    seen := make(map[card.Rank]bool)
    for _, r := range ranks {
        if !seen[r] {
            seen[r] = true
            unique = append(unique, r)
        }
    }

    // A2345顺子 (A作1)
    hasAce := false
    for _, r := range unique {
        if r == card.Ace {
            hasAce = true
            break
        }
    }

    candidates := unique
    if hasAce {
        candidates = append(candidates, 1)
        sort.Slice(candidates, func(i, j int) bool {
            return candidates[i] > candidates[j]
        })
    }

    return e.evaluateStraight(candidates)
}

func (e *Evaluator) evaluateStraight(sortedRanks []card.Rank) HandEvaluation {
    for i := 0; i <= len(sortedRanks)-5; i++ {
        straight := true
        for j := 0; j < 4; j++ {
            if sortedRanks[i+j]-sortedRanks[i+j+1] != 1 {
                straight = false
                break
            }
        }
        if straight {
            rank := HandRank(RankStraight)
            if sortedRanks[i] == card.Ace {
                // 10JQKA皇家同花顺已在前面处理
            }
            return HandEvaluation{
                Rank:      rank,
                MainValue: int(sortedRanks[i]),
            }
        }
    }
    return HandEvaluation{}
}

// 检查三条
func (e *Evaluator) checkThreeOfAKind(rankCounts map[card.Rank]int, sorted []card.Card) HandEvaluation {
    var threeRank card.Rank
    for rank, count := range rankCounts {
        if count == 3 && (threeRank == 0 || rank > threeRank) {
            threeRank = rank
        }
    }

    if threeRank > 0 {
        kickers := make([]int, 0)
        for _, c := range sorted {
            if c.Rank != threeRank {
                kickers = append(kickers, int(c.Rank))
                if len(kickers) == 2 {
                    break
                }
            }
        }
        return HandEvaluation{
            Rank:      RankThreeOfAKind,
            MainValue: int(threeRank),
            Kickers:   kickers,
        }
    }
    return HandEvaluation{}
}

// 检查两对
func (e *Evaluator) checkTwoPair(rankCounts map[card.Rank]int, sorted []card.Card) HandEvaluation {
    var pairs []card.Rank
    for rank, count := range rankCounts {
        if count == 2 {
            pairs = append(pairs, rank)
        }
    }

    if len(pairs) >= 2 {
        sort.Slice(pairs, func(i, j int) bool {
            return pairs[i] > pairs[j]
        })

        kicker := card.Two
        for _, c := range sorted {
            if c.Rank != pairs[0] && c.Rank != pairs[1] {
                kicker = c.Rank
                break
            }
        }

        return HandEvaluation{
            Rank:      RankTwoPair,
            MainValue: int(pairs[0]),
            Kickers:   []int{int(pairs[1]), int(kicker)},
        }
    }
    return HandEvaluation{}
}

// 检查一对
func (e *Evaluator) checkOnePair(rankCounts map[card.Rank]int, sorted []card.Card) HandEvaluation {
    var pairRank card.Rank
    for rank, count := range rankCounts {
        if count == 2 && (pairRank == 0 || rank > pairRank) {
            pairRank = rank
        }
    }

    if pairRank > 0 {
        kickers := make([]int, 0)
        for _, c := range sorted {
            if c.Rank != pairRank {
                kickers = append(kickers, int(c.Rank))
                if len(kickers) == 3 {
                    break
                }
            }
        }
        return HandEvaluation{
            Rank:      RankOnePair,
            MainValue: int(pairRank),
            Kickers:   kickers,
        }
    }
    return HandEvaluation{}
}

// 检查高牌
func (e *Evaluator) checkHighCard(sorted []card.Card) HandEvaluation {
    if len(sorted) < 5 {
        sorted = append(sorted, card.Card{})
    }
    best5 := sorted[:5]
    kickers := make([]int, 5)
    for i, c := range best5 {
        kickers[i] = int(c.Rank)
    }
    return HandEvaluation{
        Rank:      RankHighCard,
        MainValue: kickers[0],
        Kickers:   kickers[1:],
        RawCards:  best5,
    }
}

// 比较两手牌
func (e *Evaluator) Compare(h1, h2 HandEvaluation) int {
    if h1.Rank != h2.Rank {
        if h1.Rank > h2.Rank {
            return -1 // 牌型等级数字越小越大
        }
        return 1
    }
    if h1.MainValue != h2.MainValue {
        if h1.MainValue > h2.MainValue {
            return 1
        }
        return -1
    }
    for i := 0; i < len(h1.Kickers) && i < len(h2.Kickers); i++ {
        if h1.Kickers[i] != h2.Kickers[i] {
            if h1.Kickers[i] > h2.Kickers[i] {
                return 1
            }
            return -1
        }
    }
    return 0
}

// 辅助函数
func containsRank(ranks []card.Rank, target card.Rank) bool {
    for _, r := range ranks {
        if r == target {
            return true
        }
    }
    return false
}
```

### 任务 1.5: 游戏引擎

**目标**: 实现完整的游戏流程控制

```go
// pkg/game/engine.go

package game

import (
    "errors"
    "fmt"
    "math/rand"
    "sync"
    "time"

    "yourproject/card"
    "yourproject/common/models"
    "yourproject/evaluator"
)

type GameEngine struct {
    state      *GameState
    config     *Config
    evaluator  *evaluator.Evaluator
    deck       *card.Deck
    rand       *rand.Rand
    mutex      sync.RWMutex

    // 回调
    onStateChange func(state *GameState)
}

type Config struct {
    MinPlayers     int
    MaxPlayers    int
    SmallBlind    int
    BigBlind      int
    StartingChips int
    ActionTimeout int
}

type GameState struct {
    ID              string
    Stage           Stage
    DealerButton    int
    CurrentPlayer   int
    CurrentBet      int
    Pot             int
    SidePots        []SidePot
    CommunityCards  [5]card.Card
    Players         []*models.Player
    Actions         []models.PlayerAction
}

type Stage int

const (
    StageWaiting Stage = iota
    StagePreFlop
    StageFlop
    StageTurn
    StageRiver
    StageShowdown
    StageEnd
)

var stageNames = []string{
    "等待开始", "翻牌前", "翻牌圈", "转牌圈", "河牌圈", "摊牌", "局结束",
}

func (s Stage) String() string {
    if s >= 0 && int(s) < len(stageNames) {
        return stageNames[s]
    }
    return "未知"
}

type SidePot struct {
    Amount          int
    EligiblePlayers []int
}

// 创建游戏引擎
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
        },
        config:    config,
        evaluator: evaluator.NewEvaluator(),
        rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
    }

    return engine
}

// 设置状态变化回调
func (e *GameEngine) SetOnStateChange(fn func(state *GameState)) {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    e.onStateChange = fn
}

// 获取当前状态
func (e *GameEngine) GetState() *GameState {
    e.mutex.RLock()
    defer e.mutex.RUnlock()
    return e.copyState()
}

// 添加玩家
func (e *GameEngine) AddPlayer(id, name string, seat int) (*models.Player, error) {
    e.mutex.Lock()
    defer e.mutex.Unlock()

    if len(e.state.Players) >= e.config.MaxPlayers {
        return nil, ErrGameFull
    }

    // 检查座位
    if seat < 0 || seat >= e.config.MaxPlayers {
        return nil, ErrInvalidSeat
    }

    // 检查座位是否被占用
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

// 移除玩家
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

// 开始一局
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

    // 更新庄家按钮
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

// 玩家行动
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

    // 验证行动
    if err := e.validateAction(player, action, amount); err != nil {
        return err
    }

    // 执行行动
    switch action {
    case models.ActionFold:
        player.Status = models.PlayerStatusFolded

    case models.ActionCheck:
        // 无操作

    case models.ActionCall:
        callAmount := e.state.CurrentBet - player.CurrentBet
        player.Chips -= callAmount
        player.CurrentBet += callAmount

    case models.ActionRaise:
        raiseAmount := amount - player.CurrentBet
        player.Chips -= raiseAmount
        player.CurrentBet += raiseAmount
        e.state.CurrentBet = amount

    case models.ActionAllIn:
        allIn := player.Chips
        player.Chips = 0
        player.CurrentBet += allIn
        player.Status = models.PlayerStatusAllIn
        if player.CurrentBet > e.state.CurrentBet {
            e.state.CurrentBet = player.CurrentBet
        }
    }

    player.HasActed = true

    // 记录行动
    e.state.Actions = append(e.state.Actions, models.PlayerAction{
        PlayerID: playerID,
        Action:   action,
        Amount:   player.CurrentBet,
        Time:     time.Now(),
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

// 内部方法
func (e *GameEngine) rotateDealerButton() {
    currentBtn := -1
    for i, p := range e.state.Players {
        if p.IsDealer {
            currentBtn = i
            break
        }
    }

    // 找下一位活跃玩家
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

func (e *GameEngine) collectBlinds() {
    dealerIdx := e.state.DealerButton
    sbIdx, bbIdx := -1, -1

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

    if sbIdx >= 0 {
        sb := e.state.Players[sbIdx]
        sbAmount := min(sb.Chips, e.config.SmallBlind)
        sb.Chips -= sbAmount
        sb.CurrentBet = sbAmount
        e.state.Pot += sbAmount
    }

    if bbIdx >= 0 {
        bb := e.state.Players[bbIdx]
        bbAmount := min(bb.Chips, e.config.BigBlind)
        bb.Chips -= bbAmount
        bb.CurrentBet = bbAmount
        e.state.Pot += bbAmount
        e.state.CurrentBet = bbAmount
    }
}

func (e *GameEngine) dealHoleCards() {
    e.deck.Burn(1)

    for _, p := range e.state.Players {
        if p.Status == models.PlayerStatusActive {
            cards := e.deck.DealN(2)
            p.HoleCards = [2]card.Card{cards[0], cards[1]}
        }
    }
}

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
            return fmt.Errorf("最小加注额为 %d", minRaise)
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

func (e *GameEngine) nextPlayer() {
    for i := 1; i <= len(e.state.Players); i++ {
        nextIdx := (e.state.CurrentPlayer + i) % len(e.state.Players)
        if e.state.Players[nextIdx].Status == models.PlayerStatusActive {
            e.state.CurrentPlayer = nextIdx
            return
        }
    }
}

func (e *GameEngine) advanceBettingRound() {
    // 重置行动状态
    for _, p := range e.state.Players {
        p.HasActed = false
    }
    e.state.CurrentBet = 0

    switch e.state.Stage {
    case StagePreFlop:
        e.state.Stage = StageFlop
        e.deck.Burn(1)
        flop := e.deck.DealN(3)
        e.state.CommunityCards = [5]card.Card{flop[0], flop[1], flop[2]}
        e.state.CurrentPlayer = e.findFirstToAct()

    case StageFlop:
        e.state.Stage = StageTurn
        e.deck.Burn(1)
        turn := e.deck.DealN(1)
        e.state.CommunityCards[3] = turn[0]
        e.state.CurrentPlayer = e.findFirstToAct()

    case StageTurn:
        e.state.Stage = StageRiver
        e.deck.Burn(1)
        river := e.deck.DealN(1)
        e.state.CommunityCards[4] = river[0]
        e.state.CurrentPlayer = e.findFirstToAct()

    case StageRiver:
        e.state.Stage = StageShowdown
        e.determineWinners()
    }
}

func (e *GameEngine) findFirstToAct() int {
    for i := 1; i <= len(e.state.Players); i++ {
        idx := (e.state.DealerButton + i) % len(e.state.Players)
        if e.state.Players[idx].Status == models.PlayerStatusActive {
            return idx
        }
    }
    return 0
}

func (e *GameEngine) determineWinners() {
    var bestEval evaluator.HandEvaluation
    var bestPlayerIdx int = -1
    ties := []int{}

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
        e.state.Players[bestPlayerIdx].Chips += e.state.Pot
    }

    e.state.Pot = 0
}

func (e *GameEngine) getActivePlayers() []*models.Player {
    active := make([]*models.Player, 0)
    for _, p := range e.state.Players {
        if p.Status == models.PlayerStatusActive {
            active = append(active, p)
        }
    }
    return active
}

func (e *GameEngine) getPlayerByID(id string) *models.Player {
    for _, p := range e.state.Players {
        if p.ID == id {
            return p
        }
    }
    return nil
}

func (e *GameEngine) copyState() *GameState {
    copy := *e.state
    copy.Players = make([]*models.Player, len(e.state.Players))
    for i, p := range e.state.Players {
        playerCopy := *p
        copy.Players[i] = &playerCopy
    }
    return &copy
}

func (e *GameEngine) notifyStateChange() {
    if e.onStateChange != nil {
        go e.onStateChange(e.copyState())
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// 错误定义
var (
    ErrGameFull        = errors.New("游戏已满")
    ErrInvalidSeat     = errors.New("无效座位")
    ErrSeatOccupied    = errors.New("座位已被占用")
    ErrHandInProgress  = errors.New("手牌进行中")
    ErrNotEnoughPlayers = errors.New("玩家不足")
    ErrNotYourTurn     = errors.New("还未轮到您")
    ErrCannotCheck     = errors.New("无法看牌")
    ErrRaiseTooSmall   = errors.New("加注额太小")
    ErrNotEnoughChips = errors.New("筹码不足")
    ErrInvalidAction   = errors.New("无效行动")
    ErrPlayerNotFound  = errors.New("玩家不存在")
)
```

## 测试用例

```go
// pkg/evaluator/evaluator_test.go

package evaluator

import (
    "testing"
    "yourproject/card"
)

func TestEvaluator_RoyalFlush(t *testing.T) {
    e := NewEvaluator()

    // 皇家同花顺
    hole := [2]card.Card{
        {card.Spades, card.Ace},
        {card.Spades, card.King},
    }
    community := [5]card.Card{
        {card.Spades, card.Queen},
        {card.Spades, card.Jack},
        {card.Spades, card.Ten},
        {card.Hearts, card.Two},
        {card.Hearts, card.Three},
    }

    eval := e.Evaluate(hole, community)
    if eval.Rank != RankRoyalFlush {
        t.Errorf("期望皇家同花顺,得到 %v", eval.Rank)
    }
}

func TestEvaluator_FullHouse(t *testing.T) {
    e := NewEvaluator()

    // 葫芦
    hole := [2]card.Card{
        {card.Hearts, card.Queen},
        {card.Diamonds, card.Queen},
    }
    community := [5]card.Card{
        {card.Clubs, card.Queen},
        {card.Spades, card.King},
        {card.Hearts, card.King},
        {card.Diamonds, card.Ten},
        {card.Hearts, card.Five},
    }

    eval := e.Evaluate(hole, community)
    if eval.Rank != RankFullHouse {
        t.Errorf("期望葫芦,得到 %v", eval.Rank)
    }
}

func TestEvaluator_Compare(t *testing.T) {
    e := NewEvaluator()

    // 玩家1: 一对A
    hole1 := [2]card.Card{{card.Hearts, card.Ace}, {card.Diamonds, card.Ace}}
    community := [5]card.Card{
        {card.Clubs, card.King}, {card.Spades, card.Queen},
        {card.Hearts, card.Jack}, {card.Diamonds, card.Ten},
        {card.Clubs, card.Two},
    }

    // 玩家2: 一对K
    hole2 := [2]card.Card{{card.Hearts, card.King}, {card.Diamonds, card.King}}

    eval1 := e.Evaluate(hole1, community)
    eval2 := e.Evaluate(hole2, community)

    cmp := e.Compare(eval1, eval2)
    if cmp != 1 {
        t.Error("玩家1应该赢")
    }
}

func TestEvaluator_Straight(t *testing.T) {
    e := NewEvaluator()

    // 顺子 10JQKA
    hole := [2]card.Card{{card.Hearts, card.Ten}, {card.Diamonds, card.Jack}}
    community := [5]card.Card{
        {card.Clubs, card.Queen}, {card.Spades, card.King},
        {card.Hearts, card.Ace}, {card.Diamonds, card.Two},
        {card.Clubs, card.Three},
    }

    eval := e.Evaluate(hole, community)
    if eval.Rank != RankStraight {
        t.Errorf("期望顺子,得到 %v", eval.Rank)
    }
    if eval.MainValue != int(card.Ace) {
        t.Errorf("期望A高顺子,得到 %d", eval.MainValue)
    }
}

func TestEvaluator_A2345Straight(t *testing.T) {
    e := NewEvaluator()

    // A2345顺子
    hole := [2]card.Card{{card.Hearts, card.Ace}, {card.Diamonds, card.Two}}
    community := [5]card.Card{
        {card.Clubs, card.Three}, {card.Spades, card.Four},
        {card.Hearts, card.Five}, {card.Diamonds, card.King},
        {card.Clubs, card.Queen},
    }

    eval := e.Evaluate(hole, community)
    if eval.Rank != RankStraight {
        t.Errorf("期望顺子,得到 %v", eval.Rank)
    }
    if eval.MainValue != 5 { // 5高顺子
        t.Errorf("期望5高顺子,得到 %d", eval.MainValue)
    }
}
```

## 阶段验收清单

- [ ] 52张牌正确创建和管理
- [ ] 洗牌算法均匀分布
- [ ] 10种牌型正确识别
- [ ] 牌型比较逻辑正确
- [ ] 完整游戏流程运行正常
- [ ] 单元测试覆盖 >80%
