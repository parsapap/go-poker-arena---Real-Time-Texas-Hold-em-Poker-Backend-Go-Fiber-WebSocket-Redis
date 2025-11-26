package main

import (
	"fmt"
	"go-poker-arena/internal/poker"
)

func main() {
	// Test the problematic case
	cards := []poker.Card{
		{Suit: poker.Hearts, Rank: poker.Ace},
		{Suit: poker.Hearts, Rank: poker.Jack},
		{Suit: poker.Hearts, Rank: poker.Nine},
		{Suit: poker.Hearts, Rank: poker.Seven},
		{Suit: poker.Hearts, Rank: poker.Four},
		{Suit: poker.Diamonds, Rank: poker.Two},
		{Suit: poker.Clubs, Rank: poker.Three},
	}

	hand := poker.EvaluateHand(cards)
	fmt.Printf("Hand Rank: %v\n", hand.Rank)
	fmt.Printf("Best Five: %+v\n", hand.BestFive)
}
