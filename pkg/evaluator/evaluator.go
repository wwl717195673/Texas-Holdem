package evaluator

import (
	"sort"

	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
)

// HandRank 表示手牌的牌型等级（数字越小牌型越大）
type HandRank int

const (
	RankHighCard     HandRank = iota + 1 // 高牌
	RankOnePair                        // 一对
	RankTwoPair                        // 两对
	RankThreeOfAKind                   // 三条
	RankStraight                       // 顺子
	RankFlush                          // 同花
	RankFullHouse                      // 葫芦
	RankFourOfAKind                    // 四条
	RankStraightFlush                  // 同花顺
	RankRoyalFlush                     // 皇家同花顺
)

// 牌型名称
var rankNames = []string{
	"高牌", "一对", "两对", "三条", "顺子",
	"同花", "葫芦", "四条", "同花顺", "皇家同花顺",
}

// 牌型符号
var rankSymbols = []string{
	"HC", "1P", "2P", "3K", "ST",
	"FL", "FH", "4K", "SF", "RF",
}

// String 返回牌型名称
func (r HandRank) String() string {
	if r >= 1 && int(r) <= len(rankNames) {
		return rankNames[r-1]
	}
	return "未知"
}

// Symbol 返回牌型符号
func (r HandRank) Symbol() string {
	if r >= 1 && int(r) <= len(rankSymbols) {
		return rankSymbols[r-1]
	}
	return "?"
}

// HandEvaluation 表示手牌评估结果
type HandEvaluation struct {
	Rank      HandRank      // 牌型等级
	MainValue int           // 主要比较值（用于同牌型比较）
	Kickers   []int         // 踢脚牌（用于进一步比较）
	RawCards  []card.Card   // 参与评估的5张牌
}

// Evaluator 是扑克手牌评估器
type Evaluator struct{}

// NewEvaluator 创建一个新的评估器
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate 评估一手牌（2张底牌 + 5张公共牌）
func (e *Evaluator) Evaluate(holeCards [2]card.Card, communityCards [5]card.Card) HandEvaluation {
	allCards := make([]card.Card, 0, 7)
	allCards = append(allCards, holeCards[:]...)
	allCards = append(allCards, communityCards[:]...)
	return e.evaluate7Cards(allCards)
}

// evaluate7Cards 从7张牌中找到最佳的5张牌组合
func (e *Evaluator) evaluate7Cards(cards []card.Card) HandEvaluation {
	// 按花色分组
	suitGroups := make(map[card.Suit][]card.Card)
	// 统计每个点数出现的次数
	rankCounts := make(map[card.Rank]int)

	for _, c := range cards {
		suitGroups[c.Suit] = append(suitGroups[c.Suit], c)
		rankCounts[c.Rank]++
	}

	// 按点数从大到小排序
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank > cards[j].Rank
	})

	// 从高到低检查各种牌型
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
	if eval := e.checkStraight(cards); eval.Rank > 0 {
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

// checkRoyalFlush 检查皇家同花顺（A、K、Q、J、10 同花）
func (e *Evaluator) checkRoyalFlush(suitGroups map[card.Suit][]card.Card) HandEvaluation {
	for _, cards := range suitGroups {
		if len(cards) >= 5 {
			ranks := make(map[card.Rank]bool)
			for _, c := range cards {
				ranks[c.Rank] = true
			}
			if ranks[card.Ten] && ranks[card.Jack] && ranks[card.Queen] &&
				ranks[card.King] && ranks[card.Ace] {
				return HandEvaluation{
					Rank:      RankRoyalFlush,
					MainValue: 14, // A
				}
			}
		}
	}
	return HandEvaluation{}
}

// checkStraightFlush 检查同花顺（5张同花色的顺子）
func (e *Evaluator) checkStraightFlush(cards []card.Card, suitGroups map[card.Suit][]card.Card) HandEvaluation {
	for _, flushCards := range suitGroups {
		if len(flushCards) >= 5 {
			// 提取并排序点数
			flushRanks := make([]card.Rank, len(flushCards))
			for i, c := range flushCards {
				flushRanks[i] = c.Rank
			}
			sort.Slice(flushRanks, func(i, j int) bool {
				return flushRanks[i] > flushRanks[j]
			})
			if eval := e.evaluateStraight(flushRanks); eval.Rank > 0 {
				// 找到实际的同花顺牌
				var straightCards []card.Card
				for _, c := range cards {
					if c.Suit == flushCards[0].Suit && containsRank(flushRanks, c.Rank) {
						straightCards = append(straightCards, c)
						if len(straightCards) == 5 {
							break
						}
					}
				}
				eval.RawCards = straightCards
				eval.Rank = RankStraightFlush
				return eval
			}
		}
	}
	return HandEvaluation{}
}

// checkFourOfAKind 检查四条（四张相同点数）
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

// checkFullHouse 检查葫芦（三条加一对）
func (e *Evaluator) checkFullHouse(rankCounts map[card.Rank]int) HandEvaluation {
	var threeRanks []card.Rank // 三条的点数列表
	var pairRanks []card.Rank  // 对子的点数列表

	for rank, count := range rankCounts {
		if count == 3 {
			threeRanks = append(threeRanks, rank)
		} else if count == 2 {
			pairRanks = append(pairRanks, rank)
		}
	}

	// 有两组三条时，取最大的作为三条，第二大的作为对子
	if len(threeRanks) >= 2 {
		sort.Slice(threeRanks, func(i, j int) bool {
			return threeRanks[i] > threeRanks[j]
		})
		three := threeRanks[0]
		threeAsPair := threeRanks[1]

		return HandEvaluation{
			Rank:      RankFullHouse,
			MainValue: int(three),
			Kickers:   []int{int(threeAsPair)},
		}
	}

	// 一组三条加上一组对子
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

// checkFlush 检查同花（5张同花色）
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

// checkStraight 检查顺子（5张连续点数）
func (e *Evaluator) checkStraight(sorted []card.Card) HandEvaluation {
	// 去重
	unique := make([]card.Rank, 0)
	seen := make(map[card.Rank]bool)
	for _, r := range sorted {
		if !seen[r.Rank] {
			seen[r.Rank] = true
			unique = append(unique, r.Rank)
		}
	}

	// 检查 A2345 顺子（Ace 作为 1）
	hasAce := false
	for _, r := range unique {
		if r == card.Ace {
			hasAce = true
			break
		}
	}

	candidates := unique
	if hasAce {
		candidates = append(candidates, 1) // 将 Ace 视为 1
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i] > candidates[j]
		})
	}

	return e.evaluateStraight(candidates)
}

// evaluateStraight 判断点数序列是否为顺子
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
			return HandEvaluation{
				Rank:      RankStraight,
				MainValue: int(sortedRanks[i]),
			}
		}
	}
	return HandEvaluation{}
}

// checkThreeOfAKind 检查三条（三张相同点数）
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

// checkTwoPair 检查两对
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

// checkOnePair 检查一对
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

// checkHighCard 检查高牌（无任何组合时）
func (e *Evaluator) checkHighCard(sorted []card.Card) HandEvaluation {
	best5 := sorted
	if len(best5) > 5 {
		best5 = best5[:5]
	}
	kickers := make([]int, 0, 5)
	for _, c := range best5 {
		kickers = append(kickers, int(c.Rank))
	}
	return HandEvaluation{
		Rank:      RankHighCard,
		MainValue: kickers[0],
		Kickers:   kickers[1:],
		RawCards:  best5,
	}
}

// Compare 比较两手牌
// 返回 1 表示 h1 赢，-1 表示 h2 赢，0 表示平局
func (e *Evaluator) Compare(h1, h2 HandEvaluation) int {
	if h1.Rank != h2.Rank {
		if h1.Rank > h2.Rank {
			return -1 // 数字越小牌型越大
		}
		return 1
	}
	if h1.MainValue != h2.MainValue {
		if h1.MainValue > h2.MainValue {
			return 1
		}
		return -1
	}
	// 比较踢脚牌
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

// IsBetter 判断 h1 是否比 h2 强
func (e *Evaluator) IsBetter(h1, h2 HandEvaluation) bool {
	return e.Compare(h1, h2) > 0
}

// IsTie 判断 h1 是否与 h2 平局
func (e *Evaluator) IsTie(h1, h2 HandEvaluation) bool {
	return e.Compare(h1, h2) == 0
}

// containsRank 辅助函数：检查点数列表是否包含指定点数
func containsRank(ranks []card.Rank, target card.Rank) bool {
	for _, r := range ranks {
		if r == target {
			return true
		}
	}
	return false
}
