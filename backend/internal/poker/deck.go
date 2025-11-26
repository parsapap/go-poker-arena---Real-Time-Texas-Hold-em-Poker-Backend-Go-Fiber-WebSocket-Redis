package poker

import (
	"crypto/rand"
	"math/big"
)

type Suit int
type Rank int

const (
	Hearts Suit = iota
	Diamonds
	Clubs
	Spades
)

const (
	Two Rank = iota
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

var suitStrings = []string{"hearts", "diamonds", "clubs", "spades"}
var rankStrings = []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

type Card struct {
	Suit Suit `json:"suit"`
	Rank Rank `json:"rank"`
}

func (c Card) String() string {
	return rankStrings[c.Rank] + " of " + suitStrings[c.Suit]
}

func (c Card) SuitString() string {
	return suitStrings[c.Suit]
}

func (c Card) RankString() string {
	return rankStrings[c.Rank]
}

type Deck struct {
	Cards []Card
}

func NewDeck() *Deck {
	deck := &Deck{Cards: make([]Card, 0, 52)}
	for suit := Hearts; suit <= Spades; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			deck.Cards = append(deck.Cards, Card{Suit: suit, Rank: rank})
		}
	}
	return deck
}

func (d *Deck) Shuffle() {
	for i := len(d.Cards) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		d.Cards[i], d.Cards[j.Int64()] = d.Cards[j.Int64()], d.Cards[i]
	}
}

func (d *Deck) Draw() Card {
	if len(d.Cards) == 0 {
		panic("deck is empty")
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card
}

func (d *Deck) DrawN(n int) []Card {
	cards := make([]Card, n)
	for i := 0; i < n; i++ {
		cards[i] = d.Draw()
	}
	return cards
}
