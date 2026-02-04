# 阶段三实现：边池支持、提前结束判定与前注功能

## 实现日期
2026-02-04

## 功能概述
本次实现增加了三个重要的扑克游戏功能：
1. **边池 (Side Pot) 支持** - 处理多人全下时的底池分配
2. **提前结束判定** - 只剩一名活跃玩家时直接获胜
3. **前注 (Ante) 功能** - 每局开始前的强制投入

---

## 1. 边池支持

### 问题背景
在德州扑克中，当多名玩家全下但筹码不同时，需要创建多个边池来正确分配奖金。

**示例场景**：
- 玩家A：全下50筹码
- 玩家B：跟注50筹码（全下）
- 玩家C：再加注到150（全下）

**边池分配**：
- 主池（50×3=150）：A、B、C 都有资格
- 边池1（50×2=100）：B、C 有资格（A已出局）
- 边池2（100×1=100）：只有C 有资格

### 实现方案

#### 新增数据结构
```go
// Config 中添加前注配置
type Config struct {
    Ante int // 前注金额（可选）
}

// SidePot 结构保持不变
type SidePot struct {
    Amount          int     // 该边池总金额
    EligiblePlayers []int  // 有资格获得该边池的玩家索引列表
}
```

#### 边池收集算法
1. **collectSidePots()** - 主入口函数
   - 收集所有下注 > 0 的玩家
   - 按下注金额从小到大排序
   - 创建主池（最小下注金额 × 有资格玩家数）
   - 递归处理剩余下注

2. **collectRemainingSidePots()** - 递归辅助函数
   - 处理剩余未分配的下注
   - 创建额外的边池
   - 直到所有下注都分配完毕

#### 边池结算算法
**determineWinnersWithSidePots()** - 边池结算函数
1. 评估所有有资格玩家（活跃或全下）的手牌
2. **逆序处理边池**（从主池到最小边池）
3. 每个池单独比较手牌强度
4. 平分或单人获得该池
5. 获胜玩家可以参与多个池的结算

### 代码位置
- `pkg/game/engine.go`

---

## 2. 提前结束判定

### 功能说明
当某玩家弃牌后，如果场上只剩一名活跃玩家，该玩家直接获胜，无需继续进行剩余的下注轮次。

### 实现方案
```go
// checkEarlyFinish 检查是否只剩一名未弃牌玩家
func (e *GameEngine) checkEarlyFinish() bool {
    activePlayers := e.getActivePlayers()
    if len(activePlayers) == 1 {
        e.state.Stage = StageShowdown
        return true
    }
    return false
}
```

### 触发时机
在每次玩家执行动作后调用：
- 玩家弃牌 (Fold) 后检查
- 玩家全下 (All-in) 后检查

### 代码位置
- `pkg/game/engine.go` - `PlayerAction()` 方法中

---

## 3. 前注功能

### 功能说明
前注（Ante）是每局开始前所有玩家强制投入的小额注，用于增加底池的竞争性。

### 实现方案
```go
// collectAnte 扣除前注
func (e *GameEngine) collectAnte() {
    if e.config.Ante <= 0 {
        return // 未配置前注，跳过
    }

    for _, p := range e.state.Players {
        if p.Status == models.PlayerStatusActive && p.Chips > 0 {
            anteAmount := min(p.Chips, e.config.Ante)
            p.Chips -= anteAmount
            p.CurrentBet += anteAmount
            e.state.Pot += anteAmount
        }
    }
}
```

### 执行顺序
1. 庄家按钮轮转
2. **前注扣除**（新增）
3. 盲注扣除
4. 发底牌

### 配置示例
```go
config := &game.Config{
    MinPlayers:    2,
    MaxPlayers:    9,
    SmallBlind:    10,
    BigBlind:      20,
    Ante:          5,      // 前注5筹码（可选，设为0则不启用）
    StartingChips: 1000,
    ActionTimeout: 30,
}
```

---

## 测试验证

### 边池测试用例
| 场景 | 预期结果 |
|-----|---------|
| A全下50，B全下100 | 主池150（平分），边池50（B独得）|
| 三人全下不同金额 | 按比例创建多个边池 |
| 边池平局 | 边池金额平分 |

### 提前结束测试用例
| 场景 | 预期结果 |
|-----|---------|
| 3人游戏，2人弃牌 | 剩余玩家直接获胜 |
| 翻牌圈提前结束 | 立即进入摊牌 |

### 前注测试用例
| 场景 | 预期结果 |
|-----|---------|
| 配置前注5 | 所有玩家开局前投入5 |
| 前注超过玩家筹码 | 只扣除剩余全部筹码 |

---

## 修改文件
- `pkg/game/engine.go`

## 新增方法
1. `checkEarlyFinish()` - 提前结束判定
2. `collectAnte()` - 前注扣除
3. `collectSidePots()` - 边池收集
4. `collectRemainingSidePots()` - 递归边池收集
5. `determineWinnersWithSidePots()` - 边池结算

## 修改方法
1. `PlayerAction()` - 添加提前结束检查
2. `StartHand()` - 调用前注扣除
3. `advanceBettingRound()` - 调用边池收集
4. `determineWinners()` - 支持边池结算分支
