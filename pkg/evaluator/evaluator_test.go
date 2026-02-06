package evaluator

import (
	"testing"

	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
)

func TestEvaluator_RoyalFlush(t *testing.T) {
	e := NewEvaluator()

	// Royal Flush
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
		t.Errorf("expected Royal Flush, got %v", eval.Rank)
	}
}

func TestEvaluator_StraightFlush(t *testing.T) {
	e := NewEvaluator()

	// Straight Flush 9-8-7-6-5 of hearts
	hole := [2]card.Card{
		{card.Hearts, card.Nine},
		{card.Hearts, card.Eight},
	}
	community := [5]card.Card{
		{card.Hearts, card.Seven},
		{card.Hearts, card.Six},
		{card.Hearts, card.Five},
		{card.Diamonds, card.Ten},
		{card.Clubs, card.King},
	}

	eval := e.Evaluate(hole, community)
	if eval.Rank != RankStraightFlush {
		t.Errorf("expected Straight Flush, got %v", eval.Rank)
	}
	if eval.MainValue != int(card.Nine) {
		t.Errorf("expected 9 high, got %d", eval.MainValue)
	}
}

func TestEvaluator_FourOfAKind(t *testing.T) {
	e := NewEvaluator()

	// Four of a kind Kings with Ace kicker
	hole := [2]card.Card{
		{card.Hearts, card.King},
		{card.Diamonds, card.King},
	}
	community := [5]card.Card{
		{card.Clubs, card.King},
		{card.Spades, card.King},
		{card.Hearts, card.Ace},
		{card.Diamonds, card.Ten},
		{card.Hearts, card.Five},
	}

	eval := e.Evaluate(hole, community)
	if eval.Rank != RankFourOfAKind {
		t.Errorf("expected Four of a Kind, got %v", eval.Rank)
	}
	if eval.MainValue != int(card.King) {
		t.Errorf("expected Kings, got %d", eval.MainValue)
	}
}

func TestEvaluator_FullHouse(t *testing.T) {
	e := NewEvaluator()

	// Full House: Queens over Kings
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
		t.Errorf("expected Full House, got %v", eval.Rank)
	}
}

func TestEvaluator_Flush(t *testing.T) {
	e := NewEvaluator()

	// Flush in hearts (NOT a straight flush)
	hole := [2]card.Card{
		{card.Hearts, card.Ace},
		{card.Hearts, card.King},
	}
	community := [5]card.Card{
		{card.Hearts, card.Jack},
		{card.Hearts, card.Seven},
		{card.Hearts, card.Five},
		{card.Diamonds, card.Two},
		{card.Clubs, card.Three},
	}

	eval := e.Evaluate(hole, community)
	if eval.Rank != RankFlush {
		t.Errorf("expected Flush, got %v", eval.Rank)
	}
}

func TestEvaluator_Straight(t *testing.T) {
	e := NewEvaluator()

	// Straight 10-J-Q-K-A
	hole := [2]card.Card{
		{card.Hearts, card.Ten},
		{card.Diamonds, card.Jack},
	}
	community := [5]card.Card{
		{card.Clubs, card.Queen},
		{card.Spades, card.King},
		{card.Hearts, card.Ace},
		{card.Diamonds, card.Two},
		{card.Clubs, card.Three},
	}

	eval := e.Evaluate(hole, community)
	if eval.Rank != RankStraight {
		t.Errorf("expected Straight, got %v", eval.Rank)
	}
	if eval.MainValue != int(card.Ace) {
		t.Errorf("expected Ace high, got %d", eval.MainValue)
	}
}

func TestEvaluator_A2345Straight(t *testing.T) {
	e := NewEvaluator()

	// A-2-3-4-5 straight (wheel)
	hole := [2]card.Card{
		{card.Hearts, card.Ace},
		{card.Diamonds, card.Two},
	}
	community := [5]card.Card{
		{card.Clubs, card.Three},
		{card.Spades, card.Four},
		{card.Hearts, card.Five},
		{card.Diamonds, card.King},
		{card.Clubs, card.Queen},
	}

	eval := e.Evaluate(hole, community)
	if eval.Rank != RankStraight {
		t.Errorf("expected Straight, got %v", eval.Rank)
	}
	if eval.MainValue != int(card.Five) {
		t.Errorf("expected 5 high straight, got %d", eval.MainValue)
	}
}

func TestEvaluator_ThreeOfAKind(t *testing.T) {
	e := NewEvaluator()

	// Three of a kind Sevens
	hole := [2]card.Card{
		{card.Hearts, card.Seven},
		{card.Diamonds, card.Seven},
	}
	community := [5]card.Card{
		{card.Clubs, card.Seven},
		{card.Spades, card.King},
		{card.Hearts, card.Queen},
		{card.Diamonds, card.Ten},
		{card.Hearts, card.Five},
	}

	eval := e.Evaluate(hole, community)
	if eval.Rank != RankThreeOfAKind {
		t.Errorf("expected Three of a Kind, got %v", eval.Rank)
	}
}

func TestEvaluator_TwoPair(t *testing.T) {
	e := NewEvaluator()

	// Two pair: Aces and Kings
	hole := [2]card.Card{
		{card.Hearts, card.Ace},
		{card.Diamonds, card.Ace},
	}
	community := [5]card.Card{
		{card.Clubs, card.King},
		{card.Spades, card.King},
		{card.Hearts, card.Queen},
		{card.Diamonds, card.Ten},
		{card.Hearts, card.Five},
	}

	eval := e.Evaluate(hole, community)
	if eval.Rank != RankTwoPair {
		t.Errorf("expected Two Pair, got %v", eval.Rank)
	}
}

func TestEvaluator_OnePair(t *testing.T) {
	e := NewEvaluator()

	// One pair: Jacks (no straight possible)
	hole := [2]card.Card{
		{card.Hearts, card.Jack},
		{card.Diamonds, card.Jack},
	}
	community := [5]card.Card{
		{card.Clubs, card.King},
		{card.Spades, card.Queen},
		{card.Hearts, card.Eight},
		{card.Diamonds, card.Five},
		{card.Hearts, card.Three},
	}

	eval := e.Evaluate(hole, community)
	if eval.Rank != RankOnePair {
		t.Errorf("expected One Pair, got %v", eval.Rank)
	}
}

func TestEvaluator_HighCard(t *testing.T) {
	e := NewEvaluator()

	// High card: Ace (no straight possible)
	hole := [2]card.Card{
		{card.Hearts, card.Ace},
		{card.Diamonds, card.King},
	}
	community := [5]card.Card{
		{card.Clubs, card.Queen},
		{card.Spades, card.Jack},
		{card.Hearts, card.Nine},
		{card.Diamonds, card.Five},
		{card.Hearts, card.Three},
	}

	eval := e.Evaluate(hole, community)
	if eval.Rank != RankHighCard {
		t.Errorf("expected High Card, got %v", eval.Rank)
	}
}

func TestEvaluator_Compare(t *testing.T) {
	e := NewEvaluator()

	// 公共牌不包含 A 和 K，确保玩家只有各自的一对
	community := [5]card.Card{
		{card.Clubs, card.Ten},
		{card.Spades, card.Queen},
		{card.Hearts, card.Jack},
		{card.Diamonds, card.Eight},
		{card.Clubs, card.Two},
	}

	// 玩家1: 一对 A（最大的一对）
	hole1 := [2]card.Card{
		{card.Hearts, card.Ace},
		{card.Diamonds, card.Ace},
	}

	// 玩家2: 一对 K
	hole2 := [2]card.Card{
		{card.Hearts, card.King},
		{card.Diamonds, card.King},
	}

	eval1 := e.Evaluate(hole1, community)
	eval2 := e.Evaluate(hole2, community)

	// 一对 A 应该赢一对 K
	cmp := e.Compare(eval1, eval2)
	if cmp != 1 {
		t.Errorf("一对A应该赢一对K, 实际结果: cmp=%d, eval1=%s(%d), eval2=%s(%d)",
			cmp, eval1.Rank, eval1.MainValue, eval2.Rank, eval2.MainValue)
	}

	// 额外测试：两对 应该赢 一对
	hole3 := [2]card.Card{
		{card.Hearts, card.Ten},   // 和公共牌的 10♣ 组成一对 10
		{card.Diamonds, card.Two}, // 和公共牌的 2♣ 组成一对 2
	}
	eval3 := e.Evaluate(hole3, community)

	cmp2 := e.Compare(eval3, eval2)
	if cmp2 != 1 {
		t.Errorf("两对应该赢一对, 实际结果: cmp=%d, eval3=%s(%d), eval2=%s(%d)",
			cmp2, eval3.Rank, eval3.MainValue, eval2.Rank, eval2.MainValue)
	}
}

func TestEvaluator_Tie(t *testing.T) {
	e := NewEvaluator()

	// Both players have identical hands
	hole1 := [2]card.Card{
		{card.Hearts, card.Ace},
		{card.Diamonds, card.King},
	}
	hole2 := [2]card.Card{
		{card.Clubs, card.Ace},
		{card.Spades, card.King},
	}
	community := [5]card.Card{
		{card.Clubs, card.Queen},
		{card.Spades, card.Jack},
		{card.Hearts, card.Ten},
		{card.Diamonds, card.Nine},
		{card.Hearts, card.Five},
	}

	eval1 := e.Evaluate(hole1, community)
	eval2 := e.Evaluate(hole2, community)

	cmp := e.Compare(eval1, eval2)
	if cmp != 0 {
		t.Error("Players should tie with identical high cards")
	}
}
