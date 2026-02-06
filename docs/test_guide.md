# Texas Hold'em 测试文档

## 文档信息

| 项目 | 内容 |
|-----|------|
| 项目名称 | Texas Hold'em CLI |
| 文档版本 | v1.0 |
| 创建日期 | 2026-02-04 |

---

## 1. 测试概览

### 1.1 测试目录结构

```
Texas-Holdem/
├── internal/
│   ├── card/
│   │   ├── deck.go          # 扑克牌模块
│   │   └── deck_test.go     # 牌组测试（已实现）
│   ├── common/models/
│   │   └── player.go        # 玩家模型
│   └── protocol/
│       └── message_test.go  # 消息协议测试（已实现）
├── pkg/
│   ├── evaluator/
│   │   ├── evaluator.go     # 牌型评估器
│   │   └── evaluator_test.go # 牌型评估测试（已实现）
│   └── game/
│       ├── engine.go        # 游戏引擎
│       ├── history.go       # 游戏历史
│       └── stats.go         # 玩家统计
└── docs/
    └── test_guide.md        # 本测试文档
```

### 1.2 运行测试

```bash
# 运行所有测试
go test ./...

# 运行指定模块测试
go test ./internal/card/...
go test ./pkg/evaluator/...
go test ./pkg/game/...

# 运行测试并显示详细输出
go test -v ./...

# 运行测试并显示覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

---

## 2. 牌组模块测试 (internal/card)

### 2.1 已实现测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestNewDeck` | 创建新牌组 | 52张牌 |
| `TestDeck_Deal` | 发单张牌 | 牌数正确减少 |
| `TestDeck_DealN` | 发多张牌 | 正确返回N张牌 |
| `TestDeck_Burn` | 弃牌 | 牌数正确减少 |
| `TestDeck_Reset` | 重置牌组 | 恢复52张牌 |
| `TestDeck_Shuffle` | 洗牌 | 牌数不变 |
| `TestDeck_ShuffleWithSeed` | 相同种子洗牌 | 结果一致 |
| `TestDeck_Empty` | 空牌组发牌 | 返回错误 |
| `TestDeck_Peek` | 查看顶牌 | 不减少牌数 |
| `TestDeck_PeekN` | 查看多张顶牌 | 不减少牌数 |
| `TestCard_Display` | 牌面显示 | 正确格式化 |
| `TestCard_IsBlack` | 黑牌判断 | 正确识别 |
| `TestCard_Compare` | 牌面比较 | 正确排序 |

### 2.2 待补充测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestDeck_ShuffleUniformity` | 洗牌均匀性 | 统计分布均匀 |
| `TestCard_AllCombinations` | 所有52张牌组合 | 无重复/遗漏 |

---

## 3. 牌型评估器测试 (pkg/evaluator)

### 3.1 已实现测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestEvaluator_RoyalFlush` | 皇家同花顺 | 正确识别 |
| `TestEvaluator_StraightFlush` | 同花顺 | 正确识别+牌点 |
| `TestEvaluator_FourOfAKind` | 四条 | 正确识别+牌点 |
| `TestEvaluator_FullHouse` | 葫芦 | 正确识别 |
| `TestEvaluator_Flush` | 同花 | 正确识别(非顺子) |
| `TestEvaluator_Straight` | 顺子(A高) | 正确识别+牌点 |
| `TestEvaluator_A2345Straight` | 顺子(5高轮子) | 正确识别为5高 |
| `TestEvaluator_ThreeOfAKind` | 三条 | 正确识别 |
| `TestEvaluator_TwoPair` | 两对 | 正确识别 |
| `TestEvaluator_OnePair` | 一对 | 正确识别 |
| `TestEvaluator_HighCard` | 高牌 | 正确识别 |
| `TestEvaluator_Compare` | 手牌比较 | 高对子胜低对子 |
| `TestEvaluator_Tie` | 平局判定 | 返回0 |

### 3.2 待补充测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestEvaluator_KickerComparison` | 踢脚比较 | 高踢脚获胜 |
| `TestEvaluator_SameRankDifferentSuits` | 同级不同花色 | 花色不影响 |
| `TestEvaluator_Badugi` | 特殊顺子边界 | A-2-3-4-5 |

---

## 4. 游戏引擎测试 (pkg/game)

### 4.1 单元测试

#### 4.1.1 游戏配置测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestConfig_DefaultValues` | 默认配置 | Min=2, Max=9 |
| `TestConfig_CustomValues` | 自定义配置 | 正确保存 |

#### 4.1.2 玩家管理测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestAddPlayer_Basic` | 添加玩家 | 成功返回玩家 |
| `TestAddPlayer_FullGame` | 满员时添加 | 返回错误 |
| `TestAddPlayer_InvalidSeat` | 无效座位 | 返回错误 |
| `TestAddPlayer_DuplicateSeat` | 重复座位 | 返回错误 |
| `TestRemovePlayer_Basic` | 移除玩家 | 成功移除 |
| `TestRemovePlayer_InGame` | 游戏中移除 | 标记弃牌 |

#### 4.1.3 游戏流程测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestStartHand_NotEnoughPlayers` | 玩家不足 | 返回错误 |
| `TestStartHand_Success` | 正常开始 | 状态=PreFlop |
| `TestStartHand_DealHoleCards` | 发底牌 | 每人2张牌 |
| `TestStartHand_CollectBlinds` | 扣除盲注 | 正确扣减 |

#### 4.1.4 下注系统测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestPlayerAction_Fold` | 弃牌 | 状态=Folded |
| `TestPlayerAction_Check` | 看牌 | 成功(无人下注) |
| `TestPlayerAction_Check_Invalid` | 违规看牌 | 返回错误 |
| `TestPlayerAction_Call` | 跟注 | 筹码正确扣减 |
| `TestPlayerAction_Raise` | 加注 | 注额翻倍 |
| `TestPlayerAction_AllIn` | 全下 | 状态=AllIn |
| `TestPlayerAction_NotEnoughChips` | 筹码不足 | 返回错误 |
| `TestPlayerAction_NotYourTurn` | 顺序错误 | 返回错误 |

#### 4.1.5 前注功能测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestAnte_Enabled` | 启用前注 | 自动扣除 |
| `TestAnte_Disabled` | 禁用前注(Ante=0) | 不扣除 |
| `TestAnte_InsufficientChips` | 筹码不足 | 扣完为止 |

#### 4.1.6 边池测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestSidePot_TwoPlayersAllIn` | 两人全下相同金额 | 单池平分 |
| `TestSidePot_ThreePlayersAllIn` | 三人全下不同金额 | 多池分配 |
| `TestSidePot_SingleWinner` | 单一获胜者 | 获得所有池 |
| `TestSidePot_Tie` | 边池平局 | 池金额平分 |
| `TestSidePot_EligiblePlayers` | 有资格玩家 | 仅他们参与分配 |

**边池测试场景示例**：

```go
// 场景：A全下50，B全下100，C再加注150
// 预期结果：
// 主池：50×3=150，A、B、C都有资格
// 边池1：50×2=100，B、C有资格
// 边池2：100×1=100，只有C有资格
```

#### 4.1.7 提前结束测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestEarlyFinish_TwoPlayers` | 两人时一人弃牌 | 剩余者获胜 |
| `TestEarlyFinish_MultiPlayers` | 多人时只剩一人 | 直接摊牌 |

#### 4.1.8 阶段流转测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestStageTransition_PreFlopToFlop` | 翻牌前→翻牌圈 | 发3张公共牌 |
| `TestStageTransition_FlopToTurn` | 翻牌圈→转牌圈 | 发1张公共牌 |
| `TestStageTransition_TurnToRiver` | 转牌圈→河牌圈 | 发1张公共牌 |
| `TestStageTransition_RiverToShowdown` | 河牌圈→摊牌 | 判定胜者 |

### 4.2 边界测试

| 测试名称 | 测试场景 | 预期结果 |
|---------|---------|---------|
| `TestEdgeCase_AllInDifferentAmounts` | 全下金额各不相同 | 正确创建多个边池 |
| `TestEdgeCase_AllPlayersAllIn` | 所有人全下 | 正确结算 |
| `TestEdgeCase_SinglePlayerAllIn` | 仅一人未全下 | 其他人不能加注 |
| `TestEdgeCase_ZeroChips` | 玩家筹码为0 | 跳过下注 |
| `TestEdgeCase_ManyPlayers` | 9人游戏 | 正常运行 |
| `TestEdgeCase_MinPlayers` | 2人游戏 | 正常运行 |
| `TestEdgeCase_PotOverflow` | 底池金额溢出 | 正确处理大数 |

### 4.3 集成测试

| 测试名称 | 测试内容 | 预期结果 |
|---------|---------|---------|
| `TestIntegration_FullHand` | 完整一局流程 | 正常完成所有阶段 |
| `TestIntegration_MultipleHands` | 多局连续游戏 | 庄家按钮轮转 |
| `TestIntegration_PlayerJoinLeave` | 玩家加入离开 | 状态正确更新 |
| `TestIntegration_StateCallback` | 状态回调 | 正确触发 |

---

## 5. 协议模块测试 (internal/protocol)

### 5.1 已实现测试

- 消息序列化/反序列化测试
- 消息类型验证测试

### 5.2 待补充测试

| 测试名称 | 测试内容 |
|---------|---------|
| `TestMessage_JSON` | JSON格式验证 |
| `TestMessage_WebSocket` | WebSocket消息格式 |

---

## 6. 手动测试用例

### 6.1 基础流程测试

```
测试步骤：
1. 创建游戏配置 (2-9玩家, SB/BB/Ante)
2. 添加玩家
3. 开始游戏
4. 完整经历：PreFlop → Flop → Turn → River → Showdown
5. 检查筹码变化
6. 重复多局
```

### 6.2 下注流程测试

```
测试步骤：
1. 翻牌前阶段
   - 小盲下注
   - 大盲下注
   - 测试跟注/加注/弃牌
2. 翻牌圈
   - 测试看牌/下注
   - 测试最小加注限制
3. 转牌圈
   - 继续下注流程
4. 河牌圈
   - 最后下注机会
```

### 6.3 边池场景测试

```
场景1：两人全下
- A全下50，B跟注全下100
- 预期：主池150平分

场景2：三人全下不同金额
- A全下50，B全下100，C全下150
- 预期：主池150，边池100，边池50

场景3：混合全下
- A全下50，B弃牌，C全下150，D全下100
- 预期：主池150(A,C,D)，边池50(C,D)，边池100(C)
```

### 6.4 提前结束测试

```
测试步骤：
1. 开始3人游戏
2. 翻牌前：玩家1弃牌
3. 预期：游戏应继续（还剩2人）
4. 翻牌后：玩家2弃牌
5. 预期：玩家3直接获胜
```

---

## 7. 性能测试

### 7.1 牌型评估性能

```bash
go test -bench=. -benchmem ./pkg/evaluator/
```

### 7.2 游戏引擎性能

```bash
go test -bench=. -benchmem ./pkg/game/
```

### 7.3 内存使用

```bash
go test -bench=. -benchmem -memprofile=mem.out ./...
go tool pprof -pdf mem.out
```

---

## 8. 测试覆盖率

### 8.1 当前覆盖率

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### 8.2 目标覆盖率

| 模块 | 目标覆盖率 |
|-----|----------|
| internal/card | 100% |
| pkg/evaluator | 100% |
| pkg/game | 90% |
| internal/protocol | 90% |

---

## 9. 持续集成

### 9.1 GitHub Actions 示例

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run tests
        run: go test ./...
      - name: Run coverage
        run: go test -coverprofile=coverage.out ./...
```

---

## 10. 已知问题与待修复

| 问题编号 | 描述 | 优先级 | 状态 |
|---------|------|-------|------|
| TBD | 边池结算平分余数处理 | 中 | 待验证 |
| TBD | 多轮下注边池累积 | 中 | 待验证 |
| TBD | 全下后加注边界情况 | 低 | 待测试 |

---

## 11. 附录

### 11.1 测试命令速查

```bash
# 运行所有测试
go test ./...

# 运行指定包测试
go test ./pkg/game/...

# 详细输出
go test -v ./...

# 覆盖率
go test -cover ./...

# 性能测试
go test -bench=. ./...

# 竞赛测试
go test -race ./...
```

### 11.2 测试数据生成

```go
// 示例：生成随机手牌测试
func generateRandomHand() ([2]card.Card, [5]card.Card) {
    deck := card.NewDeck()
    deck.Shuffle()
    hole, _ := deck.DealN(2)
    deck.Burn(1)
    community, _ := deck.DealN(5)
    return hole, community
}
```
