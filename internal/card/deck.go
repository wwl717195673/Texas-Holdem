package card

import (
	"errors"
	"math/rand"
	"time"
)

// Deck 表示一副扑克牌
type Deck struct {
	cards []Card // 牌组中的所有牌
	index int    // 当前发牌位置
}

// ErrNoCardsLeft 表示牌组已空，无法继续发牌
var ErrNoCardsLeft = errors.New("no cards left in deck")

// NewDeck 创建一副新的标准52张牌
func NewDeck() *Deck {
	d := &Deck{
		cards: make([]Card, 0, 52),
	}
	// 按花色和点数创建52张牌
	for suit := Clubs; suit <= Spades; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			d.cards = append(d.cards, Card{Suit: suit, Rank: rank})
		}
	}
	return d
}

// NewDeckWithSeed 创建一副新牌并使用指定种子洗牌
func NewDeckWithSeed(seed int64) *Deck {
	d := NewDeck()
	d.ShuffleWithSeed(seed)
	return d
}

// Shuffle 使用当前时间作为种子洗牌
func (d *Deck) Shuffle() {
	d.ShuffleWithSeed(time.Now().UnixNano())
}

// ShuffleWithSeed 使用指定种子洗牌
func (d *Deck) ShuffleWithSeed(seed int64) {
	r := rand.New(rand.NewSource(seed))
	// Fisher-Yates 洗牌算法
	for i := len(d.cards) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	}
	d.index = 0
}

// Deal 从牌组顶部发一张牌
func (d *Deck) Deal() (Card, error) {
	if d.index >= len(d.cards) {
		return Card{}, ErrNoCardsLeft
	}
	card := d.cards[d.index]
	d.index++
	return card, nil
}

// DealN 从牌组顶部发 n 张牌
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

// Burn 弃掉牌组顶部的 n 张牌（不发）
func (d *Deck) Burn(n int) error {
	if d.index+n > len(d.cards) {
		return ErrNoCardsLeft
	}
	d.index += n
	return nil
}

// Remaining 返回牌组中剩余的牌数
func (d *Deck) Remaining() int {
	return len(d.cards) - d.index
}

// Reset 重置牌组为完整的52张牌
func (d *Deck) Reset() {
	d.index = 0
	d.cards = make([]Card, 0, 52)
	for suit := Clubs; suit <= Spades; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			d.cards = append(d.cards, Card{Suit: suit, Rank: rank})
		}
	}
}

// Cards 返回牌组中剩余的所有牌
func (d *Deck) Cards() []Card {
	return d.cards[d.index:]
}

// Peek 查看牌组顶部的牌但不发牌
func (d *Deck) Peek() (Card, error) {
	if d.index >= len(d.cards) {
		return Card{}, ErrNoCardsLeft
	}
	return d.cards[d.index], nil
}

// PeekN 查看牌组顶部 n 张牌但不发牌
func (d *Deck) PeekN(n int) ([]Card, error) {
	if d.index+n > len(d.cards) {
		return nil, ErrNoCardsLeft
	}
	return d.cards[d.index : d.index+n], nil
}
