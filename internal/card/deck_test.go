package card

import (
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if deck == nil {
		t.Fatal("NewDeck returned nil")
	}

	if len(deck.cards) != 52 {
		t.Errorf("expected 52 cards, got %d", len(deck.cards))
	}

	if deck.Remaining() != 52 {
		t.Errorf("expected 52 remaining cards, got %d", deck.Remaining())
	}
}

func TestDeck_Deal(t *testing.T) {
	deck := NewDeck()
	initialRemaining := deck.Remaining()

	card, err := deck.Deal()
	if err != nil {
		t.Fatalf("Deal failed: %v", err)
	}

	if deck.Remaining() != initialRemaining-1 {
		t.Errorf("expected %d remaining, got %d", initialRemaining-1, deck.Remaining())
	}

	if card.Suit < Clubs || card.Suit > Spades {
		t.Errorf("invalid suit: %v", card.Suit)
	}
	if card.Rank < Two || card.Rank > Ace {
		t.Errorf("invalid rank: %v", card.Rank)
	}
}

func TestDeck_DealN(t *testing.T) {
	deck := NewDeck()
	initialRemaining := deck.Remaining()

	cards, err := deck.DealN(5)
	if err != nil {
		t.Fatalf("DealN failed: %v", err)
	}

	if len(cards) != 5 {
		t.Errorf("expected 5 cards, got %d", len(cards))
	}

	if deck.Remaining() != initialRemaining-5 {
		t.Errorf("expected %d remaining, got %d", initialRemaining-5, deck.Remaining())
	}
}

func TestDeck_Burn(t *testing.T) {
	deck := NewDeck()
	initialRemaining := deck.Remaining()

	err := deck.Burn(1)
	if err != nil {
		t.Fatalf("Burn failed: %v", err)
	}

	if deck.Remaining() != initialRemaining-1 {
		t.Errorf("expected %d remaining, got %d", initialRemaining-1, deck.Remaining())
	}
}

func TestDeck_Reset(t *testing.T) {
	deck := NewDeck()
	deck.DealN(10)
	deck.Burn(2)

	deck.Reset()

	if deck.Remaining() != 52 {
		t.Errorf("expected 52 remaining after reset, got %d", deck.Remaining())
	}
}

func TestDeck_Shuffle(t *testing.T) {
	deck1 := NewDeck()
	deck2 := NewDeck()

	deck2.Shuffle()

	if deck1.Remaining() != deck2.Remaining() {
		t.Error("decks have different number of cards after shuffle")
	}
}

func TestDeck_ShuffleWithSeed(t *testing.T) {
	deck1 := NewDeckWithSeed(12345)
	deck2 := NewDeckWithSeed(12345)

	deck1.DealN(10)
	deck2.DealN(10)

	card1, _ := deck1.Deal()
	card2, _ := deck2.Deal()

	if card1 != card2 {
		t.Error("same seed should produce same shuffle")
	}
}

func TestDeck_Empty(t *testing.T) {
	deck := NewDeck()
	for i := 0; i < 52; i++ {
		_, err := deck.Deal()
		if err != nil {
			t.Fatalf("Deal %d failed: %v", i+1, err)
		}
	}

	_, err := deck.Deal()
	if err != ErrNoCardsLeft {
		t.Errorf("expected ErrNoCardsLeft, got %v", err)
	}
}

func TestDeck_Peek(t *testing.T) {
	deck := NewDeck()

	peeked, err := deck.Peek()
	if err != nil {
		t.Fatalf("Peek failed: %v", err)
	}

	dealt, err := deck.Deal()
	if err != nil {
		t.Fatalf("Deal failed: %v", err)
	}

	if peeked != dealt {
		t.Error("Peek should return the same card as Deal")
	}
}

func TestDeck_PeekN(t *testing.T) {
	deck := NewDeck()

	peeked, err := deck.PeekN(3)
	if err != nil {
		t.Fatalf("PeekN failed: %v", err)
	}

	dealt, err := deck.DealN(3)
	if err != nil {
		t.Fatalf("DealN failed: %v", err)
	}

	for i := range peeked {
		if peeked[i] != dealt[i] {
			t.Errorf("PeekN[%d] != DealN[%d]", i, i)
		}
	}
}

func TestCard_Display(t *testing.T) {
	tests := []struct {
		card   Card
		expect string
	}{
		{NewCard(Clubs, Two), "2♣"},
		{NewCard(Diamonds, Ten), "10♦"},
		{NewCard(Hearts, Ace), "A♥"},
		{NewCard(Spades, King), "K♠"},
	}

	for _, tt := range tests {
		if got := tt.card.String(); got != tt.expect {
			t.Errorf("Card.String() = %q, want %q", got, tt.expect)
		}
	}
}

func TestCard_IsBlack(t *testing.T) {
	if NewCard(Clubs, Ace).IsBlack() != true {
		t.Error("Clubs should be black")
	}
	if NewCard(Spades, King).IsBlack() != true {
		t.Error("Spades should be black")
	}
	if NewCard(Hearts, Queen).IsBlack() != false {
		t.Error("Hearts should be red")
	}
	if NewCard(Diamonds, Jack).IsBlack() != false {
		t.Error("Diamonds should be red")
	}
}

func TestCard_Compare(t *testing.T) {
	aceOfSpades := NewCard(Spades, Ace)
	kingOfSpades := NewCard(Spades, King)

	if aceOfSpades.Compare(kingOfSpades) != 1 {
		t.Error("Ace should be greater than King")
	}
	if kingOfSpades.Compare(aceOfSpades) != -1 {
		t.Error("King should be less than Ace")
	}
	if aceOfSpades.Compare(aceOfSpades) != 0 {
		t.Error("Same cards should be equal")
	}
}
