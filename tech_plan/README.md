# 德州扑克游戏 - 技术方案

## 文档信息

| 项目 | 内容 |
|-----|------|
| 项目名称 | Texas Hold'em Online |
| 框架 | Bubble Tea (charmbracelet/bubbletea) |
| 协议 | WebSocket (支持本地/远程连接) |
| 语言 | Go 1.21+ |
| 文档版本 | v1.0 |

---

# 架构概览

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         德州扑克游戏架构                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌──────────────┐                              ┌──────────────┐          │
│   │   HOST       │                              │   CLIENT      │          │
│   │  (服务端)    │◄────── WebSocket ────────►   │  (玩家端)    │          │
│   │              │                              │              │          │
│   │ • 游戏引擎    │                              │ • TUI界面    │          │
│   │ • 状态管理    │                              │ • 输入处理    │          │
│   │ • 牌型评估    │                              │ • 状态显示    │          │
│   │ • WebSocket  │                              │ • WebSocket  │          │
│   │ • CLI界面    │                              │              │          │
│   └──────────────┘                              └──────────────┘          │
│                                                                              │
│   通信协议: JSON over WebSocket                                              │
│   连接方式: localhost (本地) / TCP (远程)                                    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

# Phase 1: 核心游戏引擎

## 1.1 目录结构

```
cmd/
├── host/main.go         # HOST入口
└── client/main.go       # CLIENT入口

internal/
├── common/
│   ├── protocol/        # 通信协议定义
│   │   ├── message.go      # 消息结构
│   │   └── error.go       # 错误定义
│   ├── card/            # 扑克牌
│   │   ├── card.go        # 牌定义
│   │   └── deck.go        # 牌组管理
│   └── models/          # 共享模型
│       └── player.go     # 玩家模型

pkg/
├── game/               # 游戏核心 (Host专用)
│   ├── engine.go           # 游戏引擎
│   ├── state.go            # 游戏状态
│   ├── betting.go          # 下注逻辑
│   ├── evaluator/          # 牌型评估
│   │   ├── evaluator.go    # 评估器
│   │   └── hand_rank.go    # 牌型定义
│   ├── dealer/             # 发牌逻辑
│   │   ├── dealer.go       # 发牌员
│   │   └── shuffle.go      # 洗牌算法
│   └── pot/                # 底池管理
│       ├── pot.go         # 底池
│       └── side_pot.go    # 边池

server/
├── host/
│   ├── server.go           # WebSocket服务器
│   ├── handler.go          # 消息处理
│   └── game_manager.go     # 游戏管理器
└── client/
    ├── client.go           # WebSocket客户端
    └── connector.go       # 连接管理

ui/
├── components/         # 共享UI组件
│   ├── card/               # 扑克牌渲染
│   │   └── render.go
│   ├── table/              # 牌桌渲染
│   │   └── table.go
│   └── player/             # 玩家信息渲染
│       └── player.go
    ├── host/            # HOST TUI
    │   ├── model.go          # 主模型
    │   ├── screen/           # 各屏幕
    │   │   ├── lobby.go      # 大厅
    │   │   ├── game.go       # 游戏界面
    │   │   └── result.go     # 结果
    │   └── views/           # 子视图
    │       ├── pot_view.go
    │       ├── action_view.go
    │       └── log_view.go
    └── client/          # CLIENT TUI
        ├── model.go          # 主模型
        ├── screen/           # 各屏幕
        │   ├── login.go       # 登录/连接
        │   ├── lobby.go       # 大厅
        │   ├── game.go        # 游戏界面
        │   ├── hand.go        # 手牌视图
        │   └── action.go      # 行动选择
        └── views/            # 子视图
            ├── chat.go
            └── help.go
```

## 1.2 核心数据模型

### 1.2.1 扑克牌

```go
// card/card.go
package card

// 花色定义
type Suit int

const (
    Clubs Suit = iota
    Diamonds
    Hearts
    Spades
)

func (s Suit) String() string {
    return []string{"♣", "♦", "♥", "♠"}[s]
}

// 点数定义
type Rank int

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

func (r Rank) String() string {
    names := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
    return names[r-2]
}

// 牌
type Card struct {
    Suit Suit
    Rank Rank
}

func (c Card) String() string {
    return c.Rank.String() + c.Suit.String()
}

// 快速比较
func (c Card) Value() int {
    return int(c.Rank)
}
```

### 1.2.2 牌组

```go
// card/deck.go
package card

import "math/rand"

type Deck struct {
    cards []Card
}

func NewDeck() *Deck {
    d := &Deck{cards: make([]Card, 0, 52)}
    for suit := Clubs; suit <= Spades; suit++ {
        for rank := Two; rank <= Ace; rank++ {
            d.cards = append(d.cards, Card{Suit: suit, Rank: rank})
        }
    }
    return d
}

func (d *Deck) Shuffle(r *rand.Rand) {
    for i := len(d.cards) - 1; i > 0; i-- {
        j := r.Intn(i + 1)
        d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
    }
}

func (d *Deck) Deal(count int) []Card {
    if len(d.cards) < count {
        return nil
    }
    cards := d.cards[:count]
    d.cards = d.cards[count:]
    return cards
}

func (d *Deck) Burn(count int) {
    if len(d.cards) >= count {
        d.cards = d.cards[count:]
    }
}
```

### 1.2.3 玩家

```go
// common/models/player.go
package models

// 玩家状态
type PlayerStatus int

const (
    PlayerStatusInactive PlayerStatus = iota // 未入座
    PlayerStatusActive                      // 活跃
    PlayerStatusFolded                      // 已弃牌
    PlayerStatusAllIn                       // 全下
    PlayerStatusDisconnected                // 断线
)

// 玩家行动
type ActionType int

const (
    ActionFold ActionType = iota
    ActionCheck
    ActionCall
    ActionRaise
    ActionAllIn
)

// 玩家
type Player struct {
    ID         string       `json:"id"`
    Name       string       `json:"name"`
    Chips      int          `json:"chips"`
    Seat       int          `json:"seat"` // 座位号 0-8
    Status     PlayerStatus `json:"status"`
    HoleCards  [2]Card      `json:"hole_cards,omitempty"`
    CurrentBet int          `json:"current_bet"`
    IsDealer   bool         `json:"is_dealer"`

    // 统计
    HandsPlayed  int `json:"hands_played"`
    HandsWon     int `json:"hands_won"`
    TotalWinnings int `json:"total_winnings"`
}

// 行动记录
type PlayerAction struct {
    PlayerID   string     `json:"player_id"`
    Action     ActionType `json:"action"`
    Amount     int        `json:"amount,omitempty"`
    Timestamp  int64      `json:"timestamp"`
}
```

### 1.2.4 游戏状态

```go
// game/state.go
package game

// 游戏阶段
type GameStage int

const (
    StageWaiting GameStage = iota       // 等待开始
    StagePreFlop                       // 翻牌前
    StageFlop                          // 翻牌圈
    StageTurn                          // 转牌圈
    StageRiver                         // 河牌圈
    StageShowdown                      // 摊牌
    StageEnd                           // 局结束
)

// 游戏状态
type GameState struct {
    ID             string       `json:"id"`
    Stage          GameStage    `json:"stage"`
    DealerButton   int          `json:"dealer_button"`   // 庄家座位
    CurrentPlayer  int          `json:"current_player"`  // 当前行动玩家
    CurrentBet     int          `json:"current_bet"`     // 当前最高下注
    Pot            int          `json:"pot"`             // 主池金额
    SidePots       []SidePot    `json:"side_pots"`       // 边池
    CommunityCards [5]Card      `json:"community_cards"` // 公共牌
    Players        []Player     `json:"players"`         // 玩家列表
    Deck           *Deck        `json:"-"`               // 牌组(服务端)
    Actions        []PlayerAction `json:"actions"`        // 行动历史
}

// 边池
type SidePot struct {
    Amount           int    `json:"amount"`
    EligiblePlayers  []int  `json:"eligible_players"`
    Contributors     map[int]int `json:"contributors"`
}
```

## 1.3 通信协议

### 1.3.1 消息类型定义

```go
// protocol/message.go
package protocol

// 消息类型
type MessageType string

const (
    // 客户端 -> 服务器
    MsgJoin      MessageType = "join"       // 加入游戏
    MsgLeave     MessageType = "leave"       // 离开游戏
    MsgAction    MessageType = "action"      // 玩家行动
    MsgChat      MessageType = "chat"       // 聊天
    MsgSitDown   MessageType = "sit_down"   // 入座
    MsgStandUp   MessageType = "stand_up"   // 离座

    // 服务器 -> 客户端
    MsgGameState MessageType = "game_state" // 游戏状态更新
    MsgYourTurn  MessageType = "your_turn"  // 轮到行动
    MsgHandStart MessageType = "hand_start" // 新局开始
    MsgShowdown  MessageType = "showdown"   // 摊牌
    MsgGameEnd   MessageType = "game_end"   // 游戏结束
    MsgError     MessageType = "error"      // 错误
    MsgChat      MessageType = "chat"       // 聊天
    MsgJoinAck   MessageType = "join_ack"   // 加入确认
)

// 消息基类
type Message struct {
    Type    MessageType `json:"type"`
    Payload interface{} `json:"payload,omitempty"`
    PlayerID string    `json:"player_id,omitempty"`
}

// 加入游戏
type JoinPayload struct {
    PlayerName string `json:"player_name"`
    Seat      int    `json:"seat,omitempty"` // 可选座位
    BuyIn     int    `json:"buy_in"`         // 买入金额
}

// 玩家行动
type ActionPayload struct {
    Action ActionType `json:"action"`
    Amount int        `json:"amount,omitempty"` // 加注金额
}

// 游戏状态
type GameStatePayload struct {
    GameID          string       `json:"game_id"`
    Stage           GameStage    `json:"stage"`
    DealerButton    int          `json:"dealer_button"`
    CurrentPlayer   int          `json:"current_player"`
    CurrentBet      int          `json:"current_bet"`
    Pot             int          `json:"pot"`
    CommunityCards  []Card       `json:"community_cards"`
    Players         []PlayerInfo `json:"players"`
    MinRaise        int          `json:"min_raise"`
    MaxRaise        int          `json:"max_raise,omitempty"` // -1表示无限
    ActionTimeout   int          `json:"action_timeout"`      // 秒
}

// 玩家信息(客户端可见)
type PlayerInfo struct {
    ID          string       `json:"id"`
    Name        string       `json:"name"`
    Chips       int          `json:"chips"`
    Seat        int          `json:"seat"`
    Status      PlayerStatus `json:"status"`
    CurrentBet  int          `json:"current_bet"`
    IsDealer    bool         `json:"is_dealer"`
    HoleCards   []Card       `json:"hole_cards,omitempty"` // 仅自己可见
    HasActed    bool         `json:"has_acted"`
}

// 行动通知
type YourTurnPayload struct {
    PlayerID     string   `json:"player_id"`
    MinBet       int      `json:"min_bet"`
    CallAmount   int      `json:"call_amount"`
    MinRaise     int      `json:"min_raise"`
    MaxRaise     int      `json:"max_raise"`
    AllowedActions []ActionType `json:"allowed_actions"`
    ActionTimeout int     `json:"action_timeout"`
    LastActions  []PlayerAction `json:"last_actions"`
}

// 摊牌结果
type ShowdownPayload struct {
    CommunityCards []Card          `json:"community_cards"`
    Players        []ShowdownPlayer `json:"players"`
    PotWinners     []PotWinner     `json:"pot_winners"`
}

type ShowdownPlayer struct {
    PlayerID   string `json:"player_id"`
    Name       string `json:"name"`
    HoleCards  [2]Card `json:"hole_cards"`
    HandRank   HandRank `json:"hand_rank"`
    HandName   string `json:"hand_name"`
    BestCards  []Card `json:"best_cards"`
}

type PotWinner struct {
    PlayerID string `json:"player_id"`
    Amount   int    `json:"amount"`
    IsSidePot bool  `json:"is_side_pot"`
}

// 错误
type ErrorPayload struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

### 1.3.2 HandRank 牌型枚举

```go
// evaluator/hand_rank.go
package evaluator

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

type HandEvaluation struct {
    Rank      HandRank
    MainValue int       // 主要比较值 (如对子点数, 顺子最大点数)
    Kickers   []int     // 踢脚牌(从大到小)
    RawCards  []Card    // 参与比较的5张牌
}

func (h HandRank) String() string {
    names := []string{
        "高牌",
        "一对",
        "两对",
        "三条",
        "顺子",
        "同花",
        "葫芦",
        "四条",
        "同花顺",
        "皇家同花顺",
    }
    if h >= 1 && h <= 10 {
        return names[h-1]
    }
    return "未知"
}
```

## 1.4 牌型评估器

```go
// evaluator/evaluator.go
package evaluator

import (
    "sort"
    "yourproject/card"
)

type Evaluator struct{}

func NewEvaluator() *Evaluator {
    return &Evaluator{}
}

func (e *Evaluator) Evaluate(holeCards [2]card.Card, communityCards [5]card.Card) HandEvaluation {
    // 合并所有7张牌
    allCards := make([]card.Card, 0, 7)
    allCards = append(allCards, holeCards[:]...)
    allCards = append(allCards, communityCards[:]...)

    return e.evaluate7Cards(allCards)
}

func (e *Evaluator) evaluate7Cards(cards []card.Card) HandEvaluation {
    // 排序
    sorted := make([]card.Card, len(cards))
    copy(sorted, cards)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].Rank > sorted[j].Rank
    })

    // 统计花色和点数
    suitGroups := make(map[card.Suit][]card.Card)
    rankCounts := make(map[card.Rank]int)
    ranks := make([]card.Rank, 0, len(cards))

    for _, c := range cards {
        suitGroups[c.Suit] = append(suitGroups[c.Suit], c)
        rankCounts[c.Rank]++
        ranks = append(ranks, c.Rank)
    }

    // 检查各种牌型 (从高到低)
    if sf := e.checkStraightFlush(sorted, suitGroups); sf.Rank > 0 {
        return sf
    }
    if four := e.checkFourOfAKind(rankCounts, sorted); four.Rank > 0 {
        return four
    }
    if full := e.checkFullHouse(rankCounts); full.Rank > 0 {
        return full
    }
    if flush := e.checkFlush(suitGroups); flush.Rank > 0 {
        return flush
    }
    if straight := e.checkStraight(ranks); straight.Rank > 0 {
        return straight
    }
    if three := e.checkThreeOfAKind(rankCounts, sorted); three.Rank > 0 {
        return three
    }
    if two := e.checkTwoPair(rankCounts, sorted); two.Rank > 0 {
        return two
    }
    if one := e.checkOnePair(rankCounts, sorted); one.Rank > 0 {
        return one
    }

    return e.checkHighCard(sorted)
}

func (e *Evaluator) checkStraightFlush(cards []card.Card, suitGroups map[card.Suit][]card.Card) HandEvaluation {
    // 检查每个花色的顺子
    for suit, suitCards := range suitGroups {
        if len(suitCards) >= 5 {
            ranks := make([]card.Rank, len(suitCards))
            for i, c := range suitCards {
                ranks[i] = c.Rank
            }
            if straight := e.checkStraight(ranks); straight.Rank > 0 {
                // 找出同花顺的5张牌
                var sfCards []card.Card
                for _, c := range cards {
                    if c.Suit == suit && e.containsRank(ranks, c.Rank) {
                        sfCards = append(sfCards, c)
                        if len(sfCards) == 5 {
                            break
                        }
                    }
                }
                straight.RawCards = sfCards
                if straight.MainValue == 14 { // A高顺子
                    straight.Rank = RankRoyalFlush
                } else {
                    straight.Rank = RankStraightFlush
                }
                return straight
            }
        }
    }
    return HandEvaluation{}
}

func (e *Evaluator) checkFourOfAKind(rankCounts map[card.Rank]int, sorted []card.Card) HandEvaluation {
    for rank, count := range rankCounts {
        if count == 4 {
            var kickers []int
            for _, c := range sorted {
                if c.Rank != rank {
                    kickers = append(kickers, int(c.Rank))
                }
            }
            return HandEvaluation{
                Rank:      RankFourOfAKind,
                MainValue: int(rank),
                Kickers:   kickers[:1],
            }
        }
    }
    return HandEvaluation{}
}

func (e *Evaluator) checkFullHouse(rankCounts map[card.Rank]int) HandEvaluation {
    var threeOfKind card.Rank
    var pairs []card.Rank

    for rank, count := range rankCounts {
        if count == 3 {
            if threeOfKind == 0 || rank > threeOfKind {
                threeOfKind = rank
            }
        } else if count == 2 {
            pairs = append(pairs, rank)
        }
    }

    // 有两组三条
    if len(pairs) >= 1 || threeOfKind != 0 {
        var pair card.Rank
        for _, r := range pairs {
            if r != threeOfKind {
                if pair == 0 || r > pair {
                    pair = r
                }
            }
        }
        // 如果没有对子,取第二大的三条
        if pair == 0 && len(pairs) == 0 {
            for rank := range rankCounts {
                if rank != threeOfKind && rankCounts[rank] == 3 {
                    if pair == 0 || rank > pair {
                        pair = rank
                    }
                }
            }
        }

        return HandEvaluation{
            Rank:      RankFullHouse,
            MainValue: int(threeOfKind),
            Kickers:   []int{int(pair)},
        }
    }
    return HandEvaluation{}
}

func (e *Evaluator) checkFlush(suitGroups map[card.Suit][]card.Card) HandEvaluation {
    for _, cards := range suitGroups {
        if len(cards) >= 5 {
            sort.Slice(cards, func(i, j int) bool {
                return cards[i].Rank > cards[j].Rank
            })
            best := cards[:5]
            values := make([]int, 5)
            for i, c := range best {
                values[i] = int(c.Rank)
            }
            return HandEvaluation{
                Rank:      RankFlush,
                MainValue: values[0],
                Kickers:   values[1:],
                RawCards:  best,
            }
        }
    }
    return HandEvaluation{}
}

func (e *Evaluator) checkStraight(ranks []card.Rank) HandEvaluation {
    unique := make([]card.Rank, 0)
    seen := make(map[card.Rank]bool)
    for _, r := range ranks {
        if !seen[r] {
            seen[r] = true
            unique = append(unique, r)
        }
    }

    // 检查A2345顺子 (A作1)
    hasAce := false
    for _, r := range unique {
        if r == card.Ace {
            hasAce = true
            break
        }
    }

    candidates := unique
    if hasAce {
        candidates = append(candidates, 1) // A作1
        sort.Slice(candidates, func(i, j int) bool {
            return candidates[i] > candidates[j]
        })
    }

    for i := 0; i <= len(candidates)-5; i++ {
        straight := true
        for j := 0; j < 4; j++ {
            if candidates[i+j]-candidates[i+j+1] != 1 {
                straight = false
                break
            }
        }
        if straight {
            return HandEvaluation{
                Rank:      RankStraight,
                MainValue: int(candidates[i]),
            }
        }
    }
    return HandEvaluation{}
}

func (e *Evaluator) checkThreeOfAKind(rankCounts map[card.Rank]int, sorted []card.Card) HandEvaluation {
    var threeRank card.Rank
    for rank, count := range rankCounts {
        if count == 3 && (threeRank == 0 || rank > threeRank) {
            threeRank = rank
        }
    }

    if threeRank > 0 {
        var kickers []int
        for _, c := range sorted {
            if c.Rank != threeRank {
                kickers = append(kickers, int(c.Rank))
            }
        }
        return HandEvaluation{
            Rank:      RankThreeOfAKind,
            MainValue: int(threeRank),
            Kickers:   kickers[:2],
        }
    }
    return HandEvaluation{}
}

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

        var kicker int
        for _, c := range sorted {
            if c.Rank != pairs[0] && c.Rank != pairs[1] {
                kicker = int(c.Rank)
                break
            }
        }

        return HandEvaluation{
            Rank:      RankTwoPair,
            MainValue: int(pairs[0]),
            Kickers:   []int{int(pairs[1]), kicker},
        }
    }
    return HandEvaluation{}
}

func (e *Evaluator) checkOnePair(rankCounts map[card.Rank]int, sorted []card.Card) HandEvaluation {
    var pairRank card.Rank
    for rank, count := range rankCounts {
        if count == 2 && (pairRank == 0 || rank > pairRank) {
            pairRank = rank
        }
    }

    if pairRank > 0 {
        var kickers []int
        for _, c := range sorted {
            if c.Rank != pairRank {
                kickers = append(kickers, int(c.Rank))
            }
        }
        return HandEvaluation{
            Rank:      RankOnePair,
            MainValue: int(pairRank),
            Kickers:   kickers[:3],
        }
    }
    return HandEvaluation{}
}

func (e *Evaluator) checkHighCard(sorted []card.Card) HandEvaluation {
    values := make([]int, 5)
    cards := sorted[:5]
    for i, c := range cards {
        values[i] = int(c.Rank)
    }
    return HandEvaluation{
        Rank:      RankHighCard,
        MainValue: values[0],
        Kickers:   values[1:],
        RawCards:  cards,
    }
}

func (e *Evaluator) containsRank(ranks []card.Rank, target card.Rank) bool {
    for _, r := range ranks {
        if r == target {
            return true
        }
    }
    return false
}

// Compare 比较两手牌, 返回 1=hand1赢, -1=hand2赢, 0=平手
func (e *Evaluator) Compare(h1, h2 HandEvaluation) int {
    if h1.Rank != h2.Rank {
        if h1.Rank > h2.Rank {
            return 1
        }
        return -1
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
```

## 1.5 游戏引擎

```go
// game/engine.go
package game

import (
    "math/rand"
    "sync"
    "time"
    "yourproject/card"
    "yourproject/evaluator"
    "yourproject/protocol"
)

type GameEngine struct {
    state      *GameState
    config     *Config
    evaluator  *evaluator.Evaluator
    deck       *card.Deck
    rand       *rand.Rand
    mutex      sync.Mutex
    onStateChange func(state *GameState)
}

type Config struct {
    MinPlayers      int
    MaxPlayers      int
    SmallBlind      int
    BigBlind        int
    StartingChips   int
    ActionTimeout   int
    MaxRaises       int // 每轮最大加注次数, 0表示无限
}

func NewEngine(config *Config) *GameEngine {
    return &GameEngine{
        state: &GameState{
            Players: make([]Player, 0),
            Actions: make([]PlayerAction, 0),
        },
        config:    config,
        evaluator: evaluator.NewEvaluator(),
        rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

func (e *GameEngine) SetStateChangeCallback(fn func(state *GameState)) {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    e.onStateChange = fn
}

func (e *GameEngine) GetState() *GameState {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    return e.copyState()
}

func (e *GameEngine) Join(player Player) error {
    e.mutex.Lock()
    defer e.mutex.Unlock()

    if len(e.state.Players) >= e.config.MaxPlayers {
        return ErrGameFull
    }

    // 检查座位
    if player.Seat < 0 || player.Seat >= e.config.MaxPlayers {
        return ErrInvalidSeat
    }

    // 检查座位是否被占用
    for _, p := range e.state.Players {
        if p.Seat == player.Seat {
            return ErrSeatOccupied
        }
    }

    player.Chips = e.config.StartingChips
    player.Status = PlayerStatusActive
    e.state.Players = append(e.state.Players, player)

    e.notifyStateChange()
    return nil
}

func (e *GameEngine) Leave(playerID string) error {
    e.mutex.Lock()
    defer e.mutex.Unlock()

    for i, p := range e.state.Players {
        if p.ID == playerID {
            if p.Status == PlayerStatusActive && e.state.Stage != StageWaiting {
                // 游戏中离开视为弃牌
                e.state.Players[i].Status = PlayerStatusFolded
            } else {
                e.state.Players = append(e.state.Players[:i], e.state.Players[i+1:]...)
            }
            break
        }
    }

    e.notifyStateChange()
    return nil
}

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
    e.state.Actions = make([]PlayerAction, 0)

    // 洗牌
    e.deck = card.NewDeck()
    e.deck.Shuffle(e.rand)

    // 更新庄家按钮
    e.rotateDealerButton()

    // 重置玩家状态
    for i := range e.state.Players {
        e.state.Players[i].HoleCards = [2]card.Card{}
        e.state.Players[i].CurrentBet = 0
        e.state.Players[i].HasActed = false
    }

    // 扣除盲注
    e.collectBlinds()

    // 发底牌
    e.dealHoleCards()

    e.notifyStateChange()
    return nil
}

func (e *GameEngine) PlayerAction(playerID string, action ActionType, amount int) error {
    e.mutex.Lock()
    defer e.mutex.Unlock()

    // 验证是否是当前玩家
    if e.state.CurrentPlayer >= len(e.state.Players) {
        return ErrNotYourTurn
    }

    currentPlayer := &e.state.Players[e.state.CurrentPlayer]
    if currentPlayer.ID != playerID {
        return ErrNotYourTurn
    }

    // 验证行动合法性
    if err := e.validateAction(currentPlayer, action, amount); err != nil {
        return err
    }

    // 执行行动
    switch action {
    case ActionFold:
        currentPlayer.Status = PlayerStatusFolded
    case ActionCheck:
        // 无操作
    case ActionCall:
        callAmount := e.state.CurrentBet - currentPlayer.CurrentBet
        currentPlayer.Chips -= callAmount
        currentPlayer.CurrentBet += callAmount
    case ActionRaise:
        raiseAmount := amount - currentPlayer.CurrentBet
        currentPlayer.Chips -= raiseAmount
        currentPlayer.CurrentBet += raiseAmount
        e.state.CurrentBet = amount
    case ActionAllIn:
        allIn := currentPlayer.Chips
        currentPlayer.Chips = 0
        currentPlayer.CurrentBet += allIn
        currentPlayer.Status = PlayerStatusAllIn
        if currentPlayer.CurrentBet > e.state.CurrentBet {
            e.state.CurrentBet = currentPlayer.CurrentBet
        }
    }

    // 记录行动
    e.state.Actions = append(e.state.Actions, PlayerAction{
        PlayerID:  playerID,
        Action:    action,
        Amount:    currentPlayer.CurrentBet,
        Timestamp: time.Now().Unix(),
    })

    currentPlayer.HasActed = true

    // 检查下注轮是否结束
    if e.isBettingRoundComplete() {
        e.advanceBettingRound()
    } else {
        e.nextPlayer()
    }

    e.notifyStateChange()
    return nil
}

func (e *GameEngine) rotateDealerButton() {
    // 找到当前庄家位置
    currentBtn := -1
    for i, p := range e.state.Players {
        if p.IsDealer {
            currentBtn = i
            break
        }
    }

    // 顺时针移动到下一位活跃玩家
    for i := 1; i <= len(e.state.Players); i++ {
        nextIdx := (currentBtn + i) % len(e.state.Players)
        if e.state.Players[nextIdx].Status == PlayerStatusActive {
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
    // 找到SB和BB位置
    dealerIdx := e.state.DealerButton
    sbIdx := -1
    bbIdx := -1

    for i := 1; i <= len(e.state.Players); i++ {
        idx := (dealerIdx + i) % len(e.state.Players)
        if e.state.Players[idx].Status == PlayerStatusActive {
            if sbIdx < 0 {
                sbIdx = idx
            } else if e.state.Players[idx].Status == PlayerStatusActive {
                bbIdx = idx
                break
            }
        }
    }

    // 扣除SB
    if sbIdx >= 0 {
        sb := e.state.Players[sbIdx]
        sbAmount := min(sb.Chips, e.config.SmallBlind)
        sb.Chips -= sbAmount
        sb.CurrentBet = sbAmount
        e.state.Pot += sbAmount
    }

    // 扣除BB
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
    // 弃牌一张
    e.deck.Burn(1)

    // 发底牌
    for i := range e.state.Players {
        if e.state.Players[i].Status == PlayerStatusActive {
            cards := e.deck.Deal(2)
            e.state.Players[i].HoleCards = [2]card.Card{cards[0], cards[1]}
        }
    }
}

func (e *GameEngine) validateAction(p *Player, action ActionType, amount int) error {
    switch action {
    case ActionFold:
        return nil
    case ActionCheck:
        if e.state.CurrentBet > p.CurrentBet {
            return ErrCannotCheck
        }
        return nil
    case ActionCall:
        callAmount := e.state.CurrentBet - p.CurrentBet
        if callAmount > p.Chips {
            return ErrNotEnoughChips
        }
        return nil
    case ActionRaise:
        minRaise := e.state.CurrentBet * 2
        if amount < minRaise {
            return ErrRaiseTooSmall
        }
        if amount > p.Chips+p.CurrentBet {
            return ErrNotEnoughChips
        }
        return nil
    case ActionAllIn:
        return nil
    }
    return ErrInvalidAction
}

func (e *GameEngine) isBettingRoundComplete() bool {
    activePlayers := 0
    actedPlayers := 0
    allInPlayers := 0

    for _, p := range e.state.Players {
        if p.Status == PlayerStatusActive {
            activePlayers++
            if p.HasActed || p.Status == PlayerStatusAllIn {
                actedPlayers++
            }
            if p.Status == PlayerStatusAllIn {
                allInPlayers++
            }
        }
    }

    // 所有人已行动
    if actedPlayers >= activePlayers {
        return true
    }

    // 所有人已全下
    if allInPlayers >= activePlayers && activePlayers > 1 {
        return true
    }

    return false
}

func (e *GameEngine) nextPlayer() {
    for i := 1; i <= len(e.state.Players); i++ {
        nextIdx := (e.state.CurrentPlayer + i) % len(e.state.Players)
        if e.state.Players[nextIdx].Status == PlayerStatusActive {
            e.state.CurrentPlayer = nextIdx
            return
        }
    }
}

func (e *GameEngine) advanceBettingRound() {
    // 重置玩家行动状态
    for i := range e.state.Players {
        e.state.Players[i].HasActed = false
    }
    e.state.CurrentBet = 0

    switch e.state.Stage {
    case StagePreFlop:
        e.state.Stage = StageFlop
        // 发翻牌
        e.deck.Burn(1)
        flop := e.deck.Deal(3)
        e.state.CommunityCards = [5]card.Card{flop[0], flop[1], flop[2]}
        e.state.CurrentPlayer = e.findFirstToAct()

    case StageFlop:
        e.state.Stage = StageTurn
        // 发转牌
        e.deck.Burn(1)
        turn := e.deck.Deal(1)
        e.state.CommunityCards[3] = turn[0]
        e.state.CurrentPlayer = e.findFirstToAct()

    case StageTurn:
        e.state.Stage = StageRiver
        // 发河牌
        e.deck.Burn(1)
        river := e.deck.Deal(1)
        e.state.CommunityCards[4] = river[0]
        e.state.CurrentPlayer = e.findFirstToAct()

    case StageRiver:
        e.state.Stage = StageShowdown
        e.determineWinners()

    case StageShowdown:
        e.state.Stage = StageEnd
    }
}

func (e *GameEngine) findFirstToAct() int {
    // 从庄家左側开始
    for i := 1; i <= len(e.state.Players); i++ {
        idx := (e.state.DealerButton + i) % len(e.state.Players)
        if e.state.Players[idx].Status == PlayerStatusActive {
            return idx
        }
    }
    return 0
}

func (e *GameEngine) determineWinners() {
    // 评估所有活跃玩家的手牌
    var bestEval evaluator.HandEvaluation
    var bestPlayerIdx int = -1
    ties := []int{}

    for i, p := range e.state.Players {
        if p.Status == PlayerStatusActive {
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

func (e *GameEngine) copyState() *GameState {
    copy := *e.state
    copy.Players = make([]Player, len(e.state.Players))
    copy(copy.Players, e.state.Players)
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
    ErrGameFull         = &GameError{"game_full", "游戏已满"}
    ErrInvalidSeat      = &GameError{"invalid_seat", "无效座位"}
    ErrSeatOccupied     = &GameError{"seat_occupied", "座位已被占用"}
    ErrHandInProgress   = &GameError{"hand_in_progress", "手牌进行中"}
    ErrNotEnoughPlayers = &GameError{"not_enough_players", "玩家不足"}
    ErrNotYourTurn      = &GameError{"not_your_turn", "还未轮到您"}
    ErrCannotCheck      = &GameError{"cannot_check", "无法看牌,需要跟注"}
    ErrRaiseTooSmall    = &GameError{"raise_too_small", "加注额太小"}
    ErrNotEnoughChips   = &GameError{"not_enough_chips", "筹码不足"}
    ErrInvalidAction    = &GameError{"invalid_action", "无效行动"}
)

type GameError struct {
    Code    string
    Message string
}

func (e *GameError) Error() string {
    return e.Message
}
```

## 1.6 HOST TUI 实现

```go
// ui/host/model.go
package host

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/list"
    "yourproject/server/host"
    "yourproject/game"
)

type Model struct {
    server     *host.GameServer
    gameEngine *game.GameEngine

    screen     ScreenType
    err        error

    // 子模型
    lobby      *LobbyModel
    gameView   *GameViewModel
}

type ScreenType int

const (
    ScreenWelcome ScreenType = iota
    ScreenLobby
    ScreenGame
    ScreenResult
)

func NewModel() *Model {
    return &Model{
        screen: ScreenWelcome,
    }
}

func (m *Model) Init() tea.Cmd {
    return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            if m.server != nil {
                m.server.Close()
            }
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m *Model) View() string {
    switch m.screen {
    case ScreenWelcome:
        return m.renderWelcome()
    case ScreenLobby:
        return m.lobby.View()
    case ScreenGame:
        return m.gameView.View()
    case ScreenResult:
        return m.gameView.RenderResult()
    }
    return ""
}
```

```go
// ui/host/screen/lobby.go
package screen

type LobbyModel struct {
    gameEngine *game.GameEngine
    playerList list.Model
}

func NewLobbyModel(engine *game.GameEngine) *LobbyModel {
    items := []list.Item{}
    for _, p := range engine.GetState().Players {
        items = append(items, PlayerItem{
            name:   p.Name,
            chips:  p.Chips,
            seat:   p.Seat,
        })
    }

    l := list.NewModel(items, list.NewDefaultDelegate(), 0, 0)
    l.Title = "玩家列表"

    return &LobbyModel{
        gameEngine: engine,
        playerList: l,
    }
}

func (m *LobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    newList, cmd := m.playerList.Update(msg)
    m.playerList = newList
    return m, cmd
}

func (m *LobbyModel) View() string {
    return m.playerList.View()
}
```

```go
// ui/host/screen/game.go
package screen

type GameViewModel struct {
    engine    *game.GameEngine
    playerIdx int // 当前选中的玩家
}

func NewGameViewModel(engine *game.GameEngine) *GameViewModel {
    return &GameViewModel{
        engine: engine,
        playerIdx: 0,
    }
}

func (m *GameViewModel) View() string {
    state := m.engine.GetState()
    return m.renderGameTable(state)
}

func (m *GameViewModel) renderGameTable(state *game.GameState) string {
    var s string

    // 标题
    s += m.renderHeader(state)

    // 公共牌
    s += m.renderCommunityCards(state.CommunityCards)

    // 底池
    s += m.renderPot(state)

    // 玩家
    s += m.renderPlayers(state)

    // 行动日志
    s += m.renderActionLog(state.Actions)

    return s
}

func (m *GameViewModel) renderHeader(state *game.GameState) string {
    stageNames := []string{
        "等待开始", "翻牌前", "翻牌圈", "转牌圈", "河牌圈", "摊牌", "局结束",
    }
    return fmt.Sprintf("牌局 #%d | 阶段: %s | 庄家: [P%d]\n\n",
        1, stageNames[state.Stage], state.DealerButton+1)
}

func (m *GameViewModel) renderCommunityCards(cards [5]card.Card) string {
    s := "公共牌: "
    for i, c := range cards {
        if c.Rank != 0 {
            s += fmt.Sprintf("[%s] ", c.String())
        } else {
            s += "[  ?  ] "
        }
    }
    s += "\n"
    return s
}

func (m *GameViewModel) renderPot(state *game.GameState) string {
    return fmt.Sprintf("底池: %d\n", state.Pot)
}

func (m *GameViewModel) renderPlayers(state *game.GameState) string {
    s := "\n玩家:\n"
    for i, p := range state.Players {
        status := "等待"
        switch p.Status {
        case protocol.PlayerStatusActive:
            status = "活跃"
            if i == state.CurrentPlayer {
                status = "行动中"
            }
        case protocol.PlayerStatusFolded:
            status = "已弃牌"
        case protocol.PlayerStatusAllIn:
            status = "全下"
        }

        cards := ""
        if p.HoleCards[0].Rank != 0 {
            cards = fmt.Sprintf(" 底牌: %s %s", p.HoleCards[0].String(), p.HoleCards[1].String())
        }

        s += fmt.Sprintf("[P%d] %s  筹码: %d  下注: %d  状态: %s%s\n",
            i+1, p.Name, p.Chips, p.CurrentBet, status, cards)
    }
    return s
}

func (m *GameViewModel) renderActionLog(actions []game.PlayerAction) string {
    if len(actions) == 0 {
        return ""
    }

    s := "\n最近行动:\n"
    start := len(actions) - 5
    if start < 0 {
        start = 0
    }
    for _, a := range actions[start:] {
        s += fmt.Sprintf("  - P%d: %s (%d)\n", a.PlayerID+1, a.Action.String(), a.Amount)
    }
    return s
}
```

---

# Phase 2: WebSocket 通信层

## 2.1 HOST 服务器

```go
// server/host/server.go
package host

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
    "yourproject/game"
    "yourproject/protocol"
)

type GameServer struct {
    engine       *game.GameEngine
    upgrader     websocket.Upgrader
    clients      map[string]*Client
    clientsMutex sync.RWMutex
    broadcast    chan []byte
}

type Client struct {
    conn     *websocket.Conn
    playerID string
    send     chan []byte
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境应该检查Origin
    },
}

func NewGameServer(engine *game.GameEngine) *GameServer {
    server := &GameServer{
        engine:    engine,
        clients:   make(map[string]*Client),
        broadcast: make(chan []byte, 256),
    }

    // 设置状态变化回调
    engine.SetStateChangeCallback(func(state *game.GameState) {
        server.broadcastGameState(state)
    })

    return server
}

func (s *GameServer) Run(addr string) error {
    http.HandleFunc("/ws", s.handleWebSocket)
    http.HandleFunc("/game/state", s.handleGetState))
    http.HandleFunc("/", s.handleIndex)

    log.Printf("服务器启动: %s", addr)
    return http.ListenAndServe(addr, nil)
}

func (s *GameServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("升级连接失败: %v", err)
        return
    }

    client := &Client{
        conn: conn,
        send: make(chan []byte, 256),
    }

    // 读取玩家ID
    _, msg, _ := conn.ReadMessage()
    var joinMsg protocol.Message
    json.Unmarshal(msg, &joinMsg)

    if joinMsg.Type == protocol.MsgJoin {
        payload := joinMsg.Payload.(map[string]interface{})
        playerID := fmt.Sprintf("player_%d", len(s.clients))
        client.playerID = playerID

        // 创建玩家并加入游戏
        player := game.Player{
            ID:   playerID,
            Name: payload["player_name"].(string),
            Seat: len(s.clients),
        }

        if err := s.engine.Join(player); err != nil {
            s.sendError(client, err)
            return
        }

        s.clientsMutex.Lock()
        s.clients[playerID] = client
        s.clientsMutex.Unlock()

        // 发送加入确认
        s.sendMessage(client, protocol.Message{
            Type: protocol.MsgJoinAck,
            Payload: map[string]interface{}{
                "player_id": playerID,
                "seat":      player.Seat,
            },
        })
    }

    // 启动读写协程
    go s.writePump(client)
    s.readPump(client)
}

func (s *GameServer) readPump(client *Client) {
    defer func() {
        s.disconnectClient(client)
    }()

    for {
        _, message, err := client.conn.ReadMessage()
        if err != nil {
            break
        }

        var msg protocol.Message
        if err := json.Unmarshal(message, &msg); err != nil {
            continue
        }

        s.handleMessage(client, &msg)
    }
}

func (s *GameServer) writePump(client *Client) {
    ticker := time.NewTicker(time.Second * 30)
    defer func() {
        ticker.Stop()
        client.conn.Close()
    }()

    for {
        select {
        case message, ok := <-client.send:
            client.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
            if !ok {
                client.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, _ := client.conn.NextWriter(websocket.TextMessage)
            w.Write(message)
            w.Close()

        case <-ticker.C:
            client.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
            if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func (s *GameServer) handleMessage(client *Client, msg *protocol.Message) {
    switch msg.Type {
    case protocol.MsgAction:
        payload := msg.Payload.(map[string]interface{})
        action := protocol.ActionType(int(payload["action"].(float64)))
        amount := int(payload["amount"].(float64))

        if err := s.engine.PlayerAction(client.playerID, action, amount); err != nil {
            s.sendError(client, err)
        }

    case protocol.MsgChat:
        payload := msg.Payload.(map[string]interface{})
        s.broadcastChat(client.playerID, payload["message"].(string))
    }
}

func (s *GameServer) broadcastGameState(state *game.GameState) {
    // 转换为客户端可见的状态
    payload := s.convertToClientState(state)

    s.clientsMutex.RLock()
    defer s.clientsMutex.RUnlock()

    msg := protocol.Message{
        Type:    protocol.MsgGameState,
        Payload: payload,
    }

    data, _ := json.Marshal(msg)
    for _, client := range s.clients {
        select {
        case client.send <- data:
        default:
            // 队列满,跳过
        }
    }
}

func (s *GameServer) convertToClientState(state *game.GameState) *protocol.GameStatePayload {
    players := make([]protocol.PlayerInfo, len(state.Players))

    for i, p := range state.Players {
        players[i] = protocol.PlayerInfo{
            ID:         p.ID,
            Name:       p.Name,
            Chips:      p.Chips,
            Seat:       p.Seat,
            Status:     p.Status,
            CurrentBet: p.CurrentBet,
            IsDealer:   p.IsDealer,
            HasActed:   p.HasActed,
        }

        // 只发送自己的底牌
        // 注意: 这里需要知道当前客户端对应的玩家
        // 实际实现中需要在Client结构中保存seat信息
    }

    communityCards := make([]card.Card, 0, 5)
    for _, c := range state.CommunityCards {
        if c.Rank != 0 {
            communityCards = append(communityCards, c)
        }
    }

    return &protocol.GameStatePayload{
        GameID:         "default",
        Stage:          state.Stage,
        DealerButton:   state.DealerButton,
        CurrentPlayer:  state.CurrentPlayer,
        CurrentBet:     state.CurrentBet,
        Pot:            state.Pot,
        CommunityCards: communityCards,
        Players:        players,
    }
}

func (s *GameServer) disconnectClient(client *Client) {
    s.clientsMutex.Lock()
    delete(s.clients, client.playerID)
    s.clientsMutex.Unlock()

    s.engine.Leave(client.playerID)
    client.conn.Close()
}

func (s *GameServer) sendMessage(client *Client, msg interface{}) {
    data, _ := json.Marshal(msg)
    select {
    case client.send <- data:
    default:
        // 队列满
    }
}

func (s *GameServer) sendError(client *Client, err error) {
    s.sendMessage(client, protocol.Message{
        Type: protocol.MsgError,
        Payload: protocol.ErrorPayload{
            Code:    "error",
            Message: err.Error(),
        },
    })
}

func (s *GameServer) broadcastChat(playerID, message string) {
    // 实现聊天广播
}
```

## 2.2 CLIENT 客户端

```go
// server/client/client.go
package client

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/gorilla/websocket"
    "yourproject/protocol"
)

type Client struct {
    conn       *websocket.Conn
    gameState  *protocol.GameStatePayload
    playerID   string
    recv       chan *protocol.Message
    connected  bool
}

func NewClient() *Client {
    return &Client{
        recv: make(chan *protocol.Message, 256),
    }
}

func (c *Client) Connect(addr string) error {
    url := fmt.Sprintf("ws://%s/ws", addr)
    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        return fmt.Errorf("连接失败: %v", err)
    }

    c.conn = conn
    c.connected = true

    // 发送加入消息
    joinMsg := protocol.Message{
        Type: protocol.MsgJoin,
        Payload: protocol.JoinPayload{
            PlayerName: "Player",
            BuyIn:      1000,
        },
    }
    if err := c.conn.WriteJSON(joinMsg); err != nil {
        return err
    }

    // 启动读取协程
    go c.readLoop()

    return nil
}

func (c *Client) readLoop() {
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            c.connected = false
            close(c.recv)
            return
        }

        var msg protocol.Message
        if err := json.Unmarshal(message, &msg); err != nil {
            continue
        }

        c.recv <- &msg
    }
}

func (c *Client) Receive() <-chan *protocol.Message {
    return c.recv
}

func (c *Client) SendAction(action protocol.ActionType, amount int) error {
    msg := protocol.Message{
        Type: protocol.MsgAction,
        Payload: protocol.ActionPayload{
            Action: action,
            Amount: amount,
        },
    }
    return c.conn.WriteJSON(msg)
}

func (c *Client) SendChat(message string) error {
    msg := protocol.Message{
        Type: protocol.MsgChat,
        Payload: map[string]string{
            "message": message,
        },
    }
    return c.conn.WriteJSON(msg)
}

func (c *Client) Close() {
    if c.conn != nil {
        c.conn.Close()
    }
}

func (c *Client) IsConnected() bool {
    return c.connected
}

func (c *Client) GetGameState() *protocol.GameStatePayload {
    return c.gameState
}

func (c *Client) GetPlayerID() string {
    return c.playerID
}
```

## 2.3 CLIENT TUI 实现

```go
// ui/client/model.go
package client

import (
    "github.com/charmbracelet/bubbletea"
    "yourproject/server/client"
    "yourproject/protocol"
)

type Model struct {
    client   *client.Client
    screen   ScreenType
    err      error

    // 子模型
    login    *LoginModel
    lobby    *LobbyModel
    game     *GameModel
}

type ScreenType int

const (
    ScreenLogin ScreenType = iota
    ScreenLobby
    ScreenGame
)

func NewModel() *Model {
    return &Model{
        screen: ScreenLogin,
    }
}

func (m *Model) Init() tea.Cmd {
    return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            if m.client != nil {
                m.client.Close()
            }
            return m, tea.Quit
        }

    case *protocol.Message:
        return m.handleServerMessage(msg)
    }

    return m, nil
}

func (m *Model) handleServerMessage(msg *protocol.Message) (tea.Model, tea.Cmd) {
    switch msg.Type {
    case protocol.MsgJoinAck:
        payload := msg.Payload.(map[string]interface{})
        m.client.SetPlayerID(payload["player_id"].(string))
        m.screen = ScreenLobby
        m.lobby = NewLobbyModel(m.client)

    case protocol.MsgGameState:
        state := msg.Payload.(*protocol.GameStatePayload)
        m.client.SetGameState(state)
        m.game.UpdateState(state)

    case protocol.MsgYourTurn:
        m.screen = ScreenGame
        m.game.ShowActionPrompt(msg.Payload.(*protocol.YourTurnPayload))

    case protocol.MsgError:
        m.err = fmt.Errorf("%v", msg.Payload)
    }

    return m, nil
}

func (m *Model) View() string {
    switch m.screen {
    case ScreenLogin:
        return m.login.View()
    case ScreenLobby:
        return m.lobby.View()
    case ScreenGame:
        return m.game.View()
    }
    return ""
}
```

---

# Phase 3: 高级功能

## 3.1 AI 玩家

```go
// player/ai.go
package player

import (
    "math/rand"
    "yourproject/protocol"
)

type AIPlayer struct {
    ID       string
    Name     string
    Chips    int
    Seat     int
    Strategy Strategy
}

type Strategy int

const (
    StrategyTight Strategy = iota  // 只玩强牌
    StrategyLoose                  // 宽松策略
    StrategyAggressive             // 激进策略
    StrategyPassive                // 被动策略
)

type AIController struct {
    player  *AIPlayer
    rand    *rand.Rand
}

func NewAIController(name string, chips int, seat int) *AIController {
    return &AIController{
        player: &AIPlayer{
            ID:    fmt.Sprintf("ai_%d", seat),
            Name:  name,
            Chips: chips,
            Seat:  seat,
        },
        rand: rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

func (ai *AIController) Decide(state *protocol.GameStatePayload) (protocol.ActionType, int) {
    // 简化版AI决策
    // 实际实现应该考虑:
    // - 手牌强度
    // - 位置
    // - 底池赔率
    // - 对手风格
    // - 下注模式

    // 检查是否轮到自己
    myPlayer := ai.getMyPlayer(state)
    if myPlayer == nil || state.CurrentPlayer != myPlayer.Seat {
        return protocol.ActionCheck, 0
    }

    // 随机决策 (简化版)
    roll := ai.rand.Float64()

    if roll < 0.1 {
        return protocol.ActionFold, 0
    } else if roll < 0.5 {
        return protocol.ActionCheck, 0
    } else if roll < 0.8 {
        return protocol.ActionCall, 0
    } else if roll < 0.95 {
        minRaise := state.MinRaise
        return protocol.ActionRaise, minRaise * (1 + ai.rand.Intn(3))
    } else {
        return protocol.ActionAllIn, 0
    }
}

func (ai *AIController) getMyPlayer(state *protocol.GameStatePayload) *protocol.PlayerInfo {
    for _, p := range state.Players {
        if p.ID == ai.player.ID {
            return &p
        }
    }
    return nil
}
```

## 3.2 历史记录

```go
// game/history.go
package game

type HandHistory struct {
    HandID         int
    Timestamp      time.Time
    Players        []HandHistoryPlayer
    CommunityCards [5]card.Card
    Actions        []HandHistoryAction
    Showdown       []ShowdownInfo
    Pot            int
}

type HandHistoryPlayer struct {
    PlayerID string
    Name     string
    HoleCards [2]card.Card
    FinalChips int
}

type HandHistoryAction struct {
    PlayerID  string
    Round     GameStage
    Action    ActionType
    Amount    int
    Timestamp time.Time
}

type ShowdownInfo struct {
    PlayerID  string
    HandRank  evaluator.HandRank
    HandName  string
    BestCards [5]card.Card
}

type HistoryManager struct {
    hands []HandHistory
}

func (m *HistoryManager) SaveHand(state *GameState) {
    // 保存当前局到历史
}

func (m *HistoryManager) GetHandHistory(handID int) *HandHistory {
    // 获取历史记录
}

func (m *HistoryManager) ExportToJSON() ([]byte, error) {
    // 导出为JSON
}
```

## 3.3 聊天功能

```go
// ui/components/chat.go
package components

import (
    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbletea"
)

type ChatModel struct {
    messages []ChatMessage
    input    textarea.Model
    visible  bool
}

type ChatMessage struct {
    PlayerID string
    Name     string
    Text     string
    Time     time.Time
}

func NewChatModel() *ChatModel {
    ti := textarea.New()
    ti.Placeholder = "输入消息..."
    ti.Focus()

    return &ChatModel{
        messages: make([]ChatMessage, 0),
        input:    ti,
        visible:  false,
    }
}

func (m *ChatModel) Toggle() {
    m.visible = !m.visible
}

func (m *ChatModel) AddMessage(playerID, name, text string) {
    m.messages = append(m.messages, ChatMessage{
        PlayerID: playerID,
        Name:     name,
        Text:     text,
        Time:     time.Now(),
    })
}

func (m *ChatModel) View() string {
    if !m.visible {
        return ""
    }

    var s string
    for _, msg := range m.messages {
        s += fmt.Sprintf("[%s] %s: %s\n", msg.Time.Format("15:04"), msg.Name, msg.Text)
    }
    s += "\n" + m.input.View()
    return s
}
```

---

# Phase 4: 优化与完善

## 4.1 性能优化

- 使用对象池减少GC压力
- 批量发送游戏状态更新
- 压缩WebSocket消息

## 4.2 界面美化

- 使用Bubble Tea的样式系统
- 添加ASCII艺术牌面
- 实现更好的布局

## 4.3 测试

- 单元测试覆盖所有核心逻辑
- 集成测试验证完整游戏流程
- 压力测试验证并发处理

---

# 文件清单

| 文件 | Phase | 描述 |
|-----|-------|------|
| cmd/host/main.go | 1 | HOST入口 |
| cmd/client/main.go | 1 | CLIENT入口 |
| internal/card/card.go | 1 | 扑克牌定义 |
| internal/card/deck.go | 1 | 牌组管理 |
| internal/common/models/player.go | 1 | 玩家模型 |
| internal/protocol/message.go | 1 | 通信协议 |
| pkg/evaluator/evaluator.go | 1 | 牌型评估器 |
| pkg/evaluator/hand_rank.go | 1 | 牌型枚举 |
| pkg/game/engine.go | 1 | 游戏引擎 |
| pkg/game/state.go | 1 | 游戏状态 |
| pkg/game/betting.go | 1 | 下注逻辑 |
| pkg/dealer/dealer.go | 1 | 发牌逻辑 |
| pkg/dealer/shuffle.go | 1 | 洗牌算法 |
| server/host/server.go | 2 | HOST服务器 |
| server/host/handler.go | 2 | 消息处理 |
| server/client/client.go | 2 | CLIENT客户端 |
| ui/host/model.go | 1 | HOST主模型 |
| ui/host/screen/game.go | 1 | 游戏界面 |
| ui/client/model.go | 2 | CLIENT主模型 |
| player/ai.go | 3 | AI玩家 |
| game/history.go | 3 | 历史记录 |
| ui/components/chat.go | 3 | 聊天组件 |

---

*文档版本: v1.0*
*创建日期: 2026-02-04*
