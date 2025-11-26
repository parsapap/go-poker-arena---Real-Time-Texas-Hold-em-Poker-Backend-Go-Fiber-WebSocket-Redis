package poker

import (
	"testing"
)

func TestEvaluateRoyalFlush(t *testing.T) {
	cards := []Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Hearts, Rank: King},
		{Suit: Hearts, Rank: Queen},
		{Suit: Hearts, Rank: Jack},
		{Suit: Hearts, Rank: Ten},
		{Suit: Diamonds, Rank: Two},
		{Suit: Clubs, Rank: Three},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != RoyalFlush {
		t.Errorf("Expected Royal Flush, got %v", hand.Rank)
	}
}

func TestEvaluateStraightFlush(t *testing.T) {
	cards := []Card{
		{Suit: Spades, Rank: Nine},
		{Suit: Spades, Rank: Eight},
		{Suit: Spades, Rank: Seven},
		{Suit: Spades, Rank: Six},
		{Suit: Spades, Rank: Five},
		{Suit: Hearts, Rank: Ace},
		{Suit: Diamonds, Rank: King},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != StraightFlush {
		t.Errorf("Expected Straight Flush, got %v", hand.Rank)
	}
}

func TestEvaluateFourOfAKind(t *testing.T) {
	cards := []Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Diamonds, Rank: Ace},
		{Suit: Clubs, Rank: Ace},
		{Suit: Spades, Rank: Ace},
		{Suit: Hearts, Rank: King},
		{Suit: Diamonds, Rank: Queen},
		{Suit: Clubs, Rank: Jack},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != FourOfAKind {
		t.Errorf("Expected Four of a Kind, got %v", hand.Rank)
	}
}

func TestEvaluateFullHouse(t *testing.T) {
	cards := []Card{
		{Suit: Hearts, Rank: King},
		{Suit: Diamonds, Rank: King},
		{Suit: Clubs, Rank: King},
		{Suit: Spades, Rank: Queen},
		{Suit: Hearts, Rank: Queen},
		{Suit: Diamonds, Rank: Two},
		{Suit: Clubs, Rank: Three},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != FullHouse {
		t.Errorf("Expected Full House, got %v", hand.Rank)
	}
}

func TestEvaluateFlush(t *testing.T) {
	// Use A, K, J, 9, 7 all hearts (no consecutive 5 ranks)
	cards := []Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Hearts, Rank: King},
		{Suit: Hearts, Rank: Jack},
		{Suit: Hearts, Rank: Nine},
		{Suit: Hearts, Rank: Seven},
		{Suit: Diamonds, Rank: Two},
		{Suit: Clubs, Rank: Three},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != Flush {
		t.Errorf("Expected Flush, got %v", hand.Rank)
	}
}

func TestEvaluateStraight(t *testing.T) {
	cards := []Card{
		{Suit: Hearts, Rank: Nine},
		{Suit: Diamonds, Rank: Eight},
		{Suit: Clubs, Rank: Seven},
		{Suit: Spades, Rank: Six},
		{Suit: Hearts, Rank: Five},
		{Suit: Diamonds, Rank: Ace},
		{Suit: Clubs, Rank: King},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != Straight {
		t.Errorf("Expected Straight, got %v", hand.Rank)
	}
}

func TestEvaluateWheel(t *testing.T) {
	// A-2-3-4-5 straight (wheel)
	cards := []Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Diamonds, Rank: Two},
		{Suit: Clubs, Rank: Three},
		{Suit: Spades, Rank: Four},
		{Suit: Hearts, Rank: Five},
		{Suit: Diamonds, Rank: King},
		{Suit: Clubs, Rank: Queen},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != Straight {
		t.Errorf("Expected Straight (wheel), got %v", hand.Rank)
	}
}

func TestEvaluateThreeOfAKind(t *testing.T) {
	// Use 6, 6, 6, A, K, J, 2 (ranks 4, 4, 4, 12, 11, 9, 0)
	cards := []Card{
		{Suit: Hearts, Rank: Six},
		{Suit: Diamonds, Rank: Six},
		{Suit: Clubs, Rank: Six},
		{Suit: Spades, Rank: Ace},
		{Suit: Hearts, Rank: King},
		{Suit: Diamonds, Rank: Jack},
		{Suit: Clubs, Rank: Two},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != ThreeOfAKind {
		t.Errorf("Expected Three of a Kind, got %v", hand.Rank)
	}
}

func TestEvaluateTwoPair(t *testing.T) {
	// Use A, A, 6, 6, K, J, 2 (ranks 12, 12, 4, 4, 11, 9, 0)
	cards := []Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Diamonds, Rank: Ace},
		{Suit: Clubs, Rank: Six},
		{Suit: Spades, Rank: Six},
		{Suit: Hearts, Rank: King},
		{Suit: Diamonds, Rank: Jack},
		{Suit: Clubs, Rank: Two},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != TwoPair {
		t.Errorf("Expected Two Pair, got %v", hand.Rank)
	}
}

func TestEvaluateOnePair(t *testing.T) {
	// Use 6, 6, A, K, J, 9, 2 (ranks 4, 4, 12, 11, 9, 7, 0)
	cards := []Card{
		{Suit: Hearts, Rank: Six},
		{Suit: Diamonds, Rank: Six},
		{Suit: Clubs, Rank: Ace},
		{Suit: Spades, Rank: King},
		{Suit: Hearts, Rank: Jack},
		{Suit: Diamonds, Rank: Nine},
		{Suit: Clubs, Rank: Two},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != OnePair {
		t.Errorf("Expected One Pair, got %v", hand.Rank)
	}
}

func TestEvaluateHighCard(t *testing.T) {
	// Use A, K, J, 9, 7, 6, 2 with different suits (ranks 12, 11, 9, 7, 5, 4, 0)
	cards := []Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Diamonds, Rank: King},
		{Suit: Clubs, Rank: Jack},
		{Suit: Spades, Rank: Nine},
		{Suit: Hearts, Rank: Seven},
		{Suit: Diamonds, Rank: Six},
		{Suit: Clubs, Rank: Two},
	}

	hand := EvaluateHand(cards)
	if hand.Rank != HighCard {
		t.Errorf("Expected High Card, got %v", hand.Rank)
	}
}

func TestCompareHands(t *testing.T) {
	// Royal Flush vs Straight Flush
	royalFlush := []Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Hearts, Rank: King},
		{Suit: Hearts, Rank: Queen},
		{Suit: Hearts, Rank: Jack},
		{Suit: Hearts, Rank: Ten},
		{Suit: Diamonds, Rank: Two},
		{Suit: Clubs, Rank: Three},
	}

	straightFlush := []Card{
		{Suit: Spades, Rank: Nine},
		{Suit: Spades, Rank: Eight},
		{Suit: Spades, Rank: Seven},
		{Suit: Spades, Rank: Six},
		{Suit: Spades, Rank: Five},
		{Suit: Hearts, Rank: Ace},
		{Suit: Diamonds, Rank: King},
	}

	hand1 := EvaluateHand(royalFlush)
	hand2 := EvaluateHand(straightFlush)

	result := CompareHands(hand1, hand2)
	if result != 1 {
		t.Errorf("Expected Royal Flush to beat Straight Flush")
	}
}

func TestCompareHandsSameRank(t *testing.T) {
	// Two pairs with different high cards
	pair1 := []Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Diamonds, Rank: Ace},
		{Suit: Clubs, Rank: King},
		{Suit: Spades, Rank: Jack},
		{Suit: Hearts, Rank: Nine},
		{Suit: Diamonds, Rank: Six},
		{Suit: Clubs, Rank: Two},
	}

	pair2 := []Card{
		{Suit: Hearts, Rank: King},
		{Suit: Diamonds, Rank: King},
		{Suit: Clubs, Rank: Ace},
		{Suit: Spades, Rank: Jack},
		{Suit: Hearts, Rank: Nine},
		{Suit: Diamonds, Rank: Six},
		{Suit: Clubs, Rank: Two},
	}

	hand1 := EvaluateHand(pair1)
	hand2 := EvaluateHand(pair2)

	result := CompareHands(hand1, hand2)
	if result != 1 {
		t.Errorf("Expected Ace pair to beat King pair")
	}
}

func TestCompareTie(t *testing.T) {
	cards1 := []Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Diamonds, Rank: King},
		{Suit: Clubs, Rank: Queen},
		{Suit: Spades, Rank: Jack},
		{Suit: Hearts, Rank: Ten},
		{Suit: Diamonds, Rank: Two},
		{Suit: Clubs, Rank: Three},
	}

	cards2 := []Card{
		{Suit: Spades, Rank: Ace},
		{Suit: Hearts, Rank: King},
		{Suit: Diamonds, Rank: Queen},
		{Suit: Clubs, Rank: Jack},
		{Suit: Spades, Rank: Ten},
		{Suit: Hearts, Rank: Four},
		{Suit: Diamonds, Rank: Five},
	}

	hand1 := EvaluateHand(cards1)
	hand2 := EvaluateHand(cards2)

	result := CompareHands(hand1, hand2)
	if result != 0 {
		t.Errorf("Expected tie, got %d", result)
	}
}

func TestDeckShuffle(t *testing.T) {
	deck1 := NewDeck()
	deck2 := NewDeck()

	deck1.Shuffle()
	deck2.Shuffle()

	// Check that shuffled decks are different
	same := true
	for i := 0; i < 52; i++ {
		if deck1.Cards[i] != deck2.Cards[i] {
			same = false
			break
		}
	}

	if same {
		t.Error("Shuffled decks should be different")
	}
}

func TestDeckDraw(t *testing.T) {
	deck := NewDeck()
	initialLen := len(deck.Cards)

	card := deck.Draw()
	if len(deck.Cards) != initialLen-1 {
		t.Errorf("Expected deck length %d, got %d", initialLen-1, len(deck.Cards))
	}

	if card.Suit < Hearts || card.Suit > Spades {
		t.Error("Invalid card suit")
	}

	if card.Rank < Two || card.Rank > Ace {
		t.Error("Invalid card rank")
	}
}

func TestDeckDrawN(t *testing.T) {
	deck := NewDeck()
	cards := deck.DrawN(5)

	if len(cards) != 5 {
		t.Errorf("Expected 5 cards, got %d", len(cards))
	}

	if len(deck.Cards) != 47 {
		t.Errorf("Expected 47 cards remaining, got %d", len(deck.Cards))
	}
}

func BenchmarkEvaluateHand(b *testing.B) {
	cards := []Card{
		{Suit: Hearts, Rank: Ace},
		{Suit: Diamonds, Rank: King},
		{Suit: Clubs, Rank: Queen},
		{Suit: Spades, Rank: Jack},
		{Suit: Hearts, Rank: Ten},
		{Suit: Diamonds, Rank: Nine},
		{Suit: Clubs, Rank: Eight},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluateHand(cards)
	}
}
