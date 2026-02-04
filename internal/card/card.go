package card

import "fmt"

// Suit 表示扑克牌的花色
type Suit int

const (
	Clubs    Suit = iota // 梅花
	Diamonds             // 方块
	Hearts               // 红心
	Spades               // 黑桃
)

// 花色符号（用于显示）
var suitNames = []string{"♣", "♦", "♥", "♠"}
var suitFullNames = []string{"梅花", "方块", "红心", "黑桃"}

// String 返回花色的符号表示
func (s Suit) String() string {
	if s >= 0 && int(s) < len(suitNames) {
		return suitNames[s]
	}
	return "?"
}

// FullName 返回花色的中文全称
func (s Suit) FullName() string {
	if s >= 0 && int(s) < len(suitFullNames) {
		return suitFullNames[s]
	}
	return "未知"
}

// Rank 表示扑克牌的点数
type Rank int

const (
	Two   Rank = iota + 2 // 2
	Three                 // 3
	Four                  // 4
	Five                  // 5
	Six                   // 6
	Seven                 // 7
	Eight                 // 8
	Nine                  // 9
	Ten                   // 10
	Jack                  // J
	Queen                 // Q
	King                  // K
	Ace                   // A
)

// 点数名称（用于显示）
var rankNames = []string{
	"", "", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A",
}

var rankSymbols = []string{
	"", "", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A",
}

// String 返回点数的符号表示
func (r Rank) String() string {
	if r >= 2 && int(r) < len(rankSymbols) {
		return rankSymbols[r]
	}
	return "?"
}

// FullName 返回点数的全称
func (r Rank) FullName() string {
	if r >= 2 && int(r) < len(rankNames) {
		return rankNames[r]
	}
	return "未知"
}

// Value 返回点数的数值（用于比较）
func (r Rank) Value() int {
	return int(r)
}

// Card 表示一张扑克牌
type Card struct {
	Suit Suit // 花色
	Rank Rank // 点数
}

// NewCard 创建一张新扑克牌
func NewCard(suit Suit, rank Rank) Card {
	return Card{Suit: suit, Rank: rank}
}

// String 返回扑克牌的字符串表示（如 "A♠"）
func (c Card) String() string {
	return c.Rank.String() + c.Suit.String()
}

// Compare 比较两张牌的大小
// 返回 1 表示 c > other, -1 表示 c < other, 0 表示相等
func (c Card) Compare(other Card) int {
	if c.Rank != other.Rank {
		if c.Rank > other.Rank {
			return 1
		}
		return -1
	}
	return 0
}

// IsBlack 判断是否为黑牌（梅花或黑桃）
func (c Card) IsBlack() bool {
	return c.Suit == Spades || c.Suit == Clubs
}

// IsRed 判断是否为红牌（方块或红心）
func (c Card) IsRed() bool {
	return c.Suit == Diamonds || c.Suit == Hearts
}

// FormatCard 返回格式化的牌字符串（用于显示）
func (c Card) FormatCard() string {
	return fmt.Sprintf("[%s]", c.String())
}
