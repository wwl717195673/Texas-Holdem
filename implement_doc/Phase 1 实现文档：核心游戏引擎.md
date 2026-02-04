# Phase 1 实现文档：核心游戏引擎

## 概述

本文档描述了德州扑克游戏 Phase 1（核心游戏引擎）的实现细节，包括牌组管理、发牌逻辑、牌型评估和游戏引擎。

## 项目结构

```
Texas-Holdem/
├── internal/
│   ├── card/
│   │   ├── card.go          # 扑克牌基础类型
│   │   ├── deck.go          # 牌组管理
│   │   └── *_test.go        # 单元测试
│   └── common/
│       └── models/
│           └── player.go     # 玩家数据结构
├── pkg/
│   ├── evaluator/
│   │   ├── evaluator.go      # 牌型评估器
│   │   └── *_test.go        # 单元测试
│   └── game/
│       └── engine.go        # 游戏引擎
├── go.mod
└── implement_doc/
    └── README.md            # 本文档
```

---

## 1. 扑克牌模块 (internal/card)

### 1.1 花色 (Suit)

```go
type Suit int

const (
    Clubs    Suit = iota  // 梅花
    Diamonds             // 方块
    Hearts               // 红心
    Spades               // 黑桃
)
```

### 1.2 点数 (Rank)

```go
type Rank int

const (
    Two   Rank = iota + 2  // 2
    Three                   // 3
    Four                    // 4
    Five                    // 5
    Six                     // 6
    Seven                   // 7
    Eight                   // 8
    Nine                    // 9
    Ten                     // 10
    Jack                    // J
    Queen                   // Q
    King                    // K
    Ace                     // A
)
```

### 1.3 牌 (Card)

```go
type Card struct {
    Suit Suit  // 花色
    Rank Rank  // 点数
}
```

**主要方法**：
- `String()` - 返回牌的字符串表示（如 "A♠"）
- `Compare(other Card)` - 比较两张牌大小
- `IsBlack()` / `IsRed()` - 判断花色

### 1.4 牌组 (Deck)

```go
type Deck struct {
    cards []Card  // 牌组中的所有牌
    index int    // 当前发牌位置
}
```

**主要方法**：
- `NewDeck()` - 创建一副新牌（52张）
- `Shuffle()` - 洗牌
- `Deal()` / `DealN(n)` - 发牌
- `Burn(n)` - 弃牌
- `Remaining()` - 剩余牌数
- `Reset()` - 重置牌组

---

## 2. 玩家模型 (internal/common/models)

### 2.1 玩家状态

```go
type PlayerStatus int

const (
    PlayerStatusInactive  // 未入座
    PlayerStatusActive    // 游戏中
    PlayerStatusFolded    // 已弃牌
    PlayerStatusAllIn     // 全下
)
```

### 2.2 玩家动作

```go
type ActionType int

const (
    ActionFold   // 弃牌
    ActionCheck  // 看牌
    ActionCall   // 跟注
    ActionRaise  // 加注
    ActionAllIn  // 全下
)
```

### 2.3 玩家结构

```go
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
    HandsPlayed int          // 参与的手牌数
    HandsWon    int          // 获胜的手牌数
}
```

---

## 3. 牌型评估器 (pkg/evaluator)

### 3.1 牌型等级（从高到低）

| 等级 | 牌型 | 说明 |
|-----|------|------|
| 10 | Royal Flush | 皇家同花顺（A、K、Q、J、10 同花） |
| 9 | Straight Flush | 同花顺 |
| 8 | Four of a Kind | 四条 |
| 7 | Full House | 葫芦 |
| 6 | Flush | 同花 |
| 5 | Straight | 顺子 |
| 4 | Three of a Kind | 三条 |
| 3 | Two Pair | 两对 |
| 2 | One Pair | 一对 |
| 1 | High Card | 高牌 |

### 3.2 评估结果

```go
type HandEvaluation struct {
    Rank      HandRank      // 牌型等级
    MainValue int           // 主要比较值（同牌型比较用）
    Kickers   []int        // 踢脚牌
    RawCards  []card.Card   // 参与评估的5张牌
}
```

### 3.3 主要方法

- `Evaluate(holeCards, communityCards)` - 评估一手牌
- `Compare(h1, h2)` - 比较两手牌，返回 1/-1/0
- `IsBetter(h1, h2)` - 判断 h1 是否更强
- `IsTie(h1, h2)` - 判断是否平局

### 3.4 牌型比较规则

1. 首先比较牌型等级（数字越小牌型越大）
2. 同牌型比较主要比较值（如对子比较对子点数）
3. 仍相同则比较踢脚牌

---

## 4. 游戏引擎 (pkg/game)

### 4.1 游戏阶段

```go
type Stage int

const (
    StageWaiting   // 等待开始
    StagePreFlop  // 翻牌前
    StageFlop     // 翻牌圈（发3张公共牌）
    StageTurn     // 转牌圈（发第4张公共牌）
    StageRiver    // 河牌圈（发第5张公共牌）
    StageShowdown // 摊牌
    StageEnd      // 局结束
)
```

### 4.2 游戏状态

```go
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
```

### 4.3 游戏配置

```go
type Config struct {
    MinPlayers     int  // 最少玩家数（默认2）
    MaxPlayers     int  // 最多玩家数（默认9）
    SmallBlind     int  // 小盲注金额
    BigBlind       int  // 大盲注金额
    StartingChips  int  // 初始筹码
    ActionTimeout  int  // 动作超时时间
}
```

### 4.4 主要方法

| 方法 | 说明 |
|-----|------|
| `NewEngine(config)` | 创建游戏引擎 |
| `AddPlayer(id, name, seat)` | 添加玩家 |
| `RemovePlayer(id)` | 移除玩家 |
| `StartHand()` | 开始新的一局 |
| `PlayerAction(playerID, action, amount)` | 处理玩家动作 |
| `GetState()` | 获取当前游戏状态 |

### 4.5 游戏流程

```
┌─────────────────────────────────────────────────┐
│ 1. StartHand()                                 │
│    - 洗牌                                      │
│    - 轮转庄家按钮                              │
│    - 扣除盲注（SB/BB）                         │
│    - 发底牌（每人2张）                          │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 2. 玩家轮流行动                                 │
│    - Fold（弃牌）                              │
│    - Check（看牌）/ Call（跟注）               │
│    - Raise（加注）                             │
│    - All-in（全下）                            │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 3. 翻牌圈（Flop）                              │
│    - 弃1张牌                                   │
│    - 发3张公共牌                               │
│    - 玩家轮流行动                               │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 4. 转牌圈（Turn）                              │
│    - 弃1张牌                                   │
│    - 发1张公共牌（第4张）                        │
│    - 玩家轮流行动                               │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 5. 河牌圈（River）                             │
│    - 弃1张牌                                   │
│    - 发1张公共牌（第5张）                        │
│    - 玩家轮流行动                               │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 6. 摊牌（Showdown）                            │
│    - 使用evaluator评估所有玩家手牌               │
│    - 比较确定获胜者                             │
│    - 分配底池                                   │
└─────────────────────────────────────────────────┘
```

---

## 5. 测试覆盖

### 5.1 牌组测试 (internal/card/deck_test.go)

| 测试项 | 说明 |
|-------|------|
| TestNewDeck | 创建52张牌 |
| TestDeck_Deal/DealN | 发牌功能 |
| TestDeck_Burn | 弃牌功能 |
| TestDeck_Shuffle | 洗牌功能 |
| TestDeck_Reset | 重置功能 |
| TestCard_Display | 牌显示 |
| TestCard_Compare | 牌比较 |

### 5.2 牌型评估测试 (pkg/evaluator/evaluator_test.go)

| 测试项 | 说明 |
|-------|------|
| TestEvaluator_RoyalFlush | 皇家同花顺 |
| TestEvaluator_StraightFlush | 同花顺 |
| TestEvaluator_FourOfAKind | 四条 |
| TestEvaluator_FullHouse | 葫芦 |
| TestEvaluator_Flush | 同花 |
| TestEvaluator_Straight | 顺子（含A2345） |
| TestEvaluator_ThreeOfAKind | 三条 |
| TestEvaluator_TwoPair | 两对 |
| TestEvaluator_OnePair | 一对 |
| TestEvaluator_HighCard | 高牌 |
| TestEvaluator_Compare | 手牌比较 |
| TestEvaluator_Tie | 平局判定 |

---

## 6. 运行命令

```bash
# 运行所有测试
go test ./...

# 构建项目
go build ./...

# 运行测试（带详细输出）
go test ./... -v
```

---

## 7. 注意事项

1. **线程安全**：GameEngine 使用读写锁保护状态
2. **A2345 顺子**：Ace 可作为 1 使用形成最小的顺子
3. **牌型比较**：数字越小牌型越大（RoyalFlush=10, HighCard=1）
4. **边池支持**：已预留 SidePot 结构待后续实现

---

## 8. 下一步（Phase 2+）

- [ ] UI 界面（Bubble Tea）
- [ ] WebSocket 通信
- [ ] AI 玩家
- [ ] 聊天系统
- [ ] 历史记录
