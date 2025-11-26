package poker

type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

var handRankStrings = []string{
	"High Card", "One Pair", "Two Pair", "Three of a Kind",
	"Straight", "Flush", "Full House", "Four of a Kind",
	"Straight Flush", "Royal Flush",
}

func (hr HandRank) String() string {
	return handRankStrings[hr]
}

type Hand struct {
	Cards    []Card
	Rank     HandRank
	Value    uint32
	BestFive []Card
}

// EvaluateHand evaluates the best 5-card poker hand from 7 cards
func EvaluateHand(cards []Card) *Hand {
	if len(cards) < 5 {
		return nil
	}

	bestHand := &Hand{Cards: cards, Rank: HighCard, Value: 0}

	// Generate all 5-card combinations from 7 cards
	combinations := generateCombinations(cards, 5)
	
	for _, combo := range combinations {
		hand := evaluateFiveCards(combo)
		if hand.Value > bestHand.Value || (hand.Value == bestHand.Value && hand.Rank > bestHand.Rank) {
			bestHand = hand
			bestHand.BestFive = combo
		}
	}

	return bestHand
}

func evaluateFiveCards(cards []Card) *Hand {
	hand := &Hand{Cards: cards}

	// Count ranks and suits
	rankCounts := make([]int, 13)
	suitCounts := make([]int, 4)
	rankMask := uint32(0)

	for _, card := range cards {
		rankCounts[card.Rank]++
		suitCounts[card.Suit]++
		rankMask |= (1 << uint(card.Rank))
	}

	isFlush := false
	for _, count := range suitCounts {
		if count == 5 {
			isFlush = true
			break
		}
	}

	isStraight, straightHigh := checkStraight(rankMask)

	// Check for Royal Flush
	if isFlush && isStraight && straightHigh == Ace {
		hand.Rank = RoyalFlush
		hand.Value = (uint32(RoyalFlush) << 20) | uint32(Ace)
		return hand
	}

	// Check for Straight Flush
	if isFlush && isStraight {
		hand.Rank = StraightFlush
		hand.Value = (uint32(StraightFlush) << 20) | uint32(straightHigh)
		return hand
	}

	// Count pairs, trips, quads
	quads, trips, pairs := countRanks(rankCounts)

	// Four of a Kind
	if len(quads) > 0 {
		hand.Rank = FourOfAKind
		kicker := getKicker(rankCounts, quads[0])
		hand.Value = (uint32(FourOfAKind) << 20) | (uint32(quads[0]) << 8) | uint32(kicker)
		return hand
	}

	// Full House
	if len(trips) > 0 && len(pairs) > 0 {
		hand.Rank = FullHouse
		hand.Value = (uint32(FullHouse) << 20) | (uint32(trips[0]) << 8) | uint32(pairs[0])
		return hand
	}

	// Flush
	if isFlush {
		hand.Rank = Flush
		hand.Value = (uint32(Flush) << 20) | getHighCards(rankCounts, 5)
		return hand
	}

	// Straight
	if isStraight {
		hand.Rank = Straight
		hand.Value = (uint32(Straight) << 20) | uint32(straightHigh)
		return hand
	}

	// Three of a Kind
	if len(trips) > 0 {
		hand.Rank = ThreeOfAKind
		kickers := getKickers(rankCounts, trips[0], 2)
		hand.Value = (uint32(ThreeOfAKind) << 20) | (uint32(trips[0]) << 8) | kickers
		return hand
	}

	// Two Pair
	if len(pairs) >= 2 {
		hand.Rank = TwoPair
		kicker := getKicker(rankCounts, pairs[0], pairs[1])
		hand.Value = (uint32(TwoPair) << 20) | (uint32(pairs[0]) << 12) | (uint32(pairs[1]) << 8) | uint32(kicker)
		return hand
	}

	// One Pair
	if len(pairs) == 1 {
		hand.Rank = OnePair
		kickers := getKickers(rankCounts, pairs[0], 3)
		hand.Value = (uint32(OnePair) << 20) | (uint32(pairs[0]) << 8) | kickers
		return hand
	}

	// High Card
	hand.Rank = HighCard
	hand.Value = (uint32(HighCard) << 20) | getHighCards(rankCounts, 5)
	return hand
}

func checkStraight(rankMask uint32) (bool, Rank) {
	// Check for wheel (A-2-3-4-5)
	if rankMask&0x100F == 0x100F {
		return true, Five
	}

	// Check for regular straights
	for i := Ace; i >= Five; i-- {
		mask := uint32(0x1F << uint(i-4))
		if rankMask&mask == mask {
			return true, i
		}
	}

	return false, 0
}

func countRanks(rankCounts []int) (quads, trips, pairs []Rank) {
	for rank := Ace; rank >= Two; rank-- {
		switch rankCounts[rank] {
		case 4:
			quads = append(quads, rank)
		case 3:
			trips = append(trips, rank)
		case 2:
			pairs = append(pairs, rank)
		}
	}
	return
}

func getKicker(rankCounts []int, exclude ...Rank) Rank {
	excludeMap := make(map[Rank]bool)
	for _, r := range exclude {
		excludeMap[r] = true
	}

	for rank := Ace; rank >= Two; rank-- {
		if !excludeMap[rank] && rankCounts[rank] > 0 {
			return rank
		}
	}
	return Two
}

func getKickers(rankCounts []int, exclude Rank, count int) uint32 {
	kickers := uint32(0)
	found := 0
	for rank := Ace; rank >= Two && found < count; rank-- {
		if rank != exclude && rankCounts[rank] > 0 {
			kickers |= uint32(rank) << uint(found*4)
			found++
		}
	}
	return kickers
}

func getHighCards(rankCounts []int, count int) uint32 {
	cards := uint32(0)
	found := 0
	for rank := Ace; rank >= Two && found < count; rank-- {
		if rankCounts[rank] > 0 {
			cards |= uint32(rank) << uint(found*4)
			found++
		}
	}
	return cards
}

func generateCombinations(cards []Card, k int) [][]Card {
	var result [][]Card
	var combination []Card
	
	var generate func(start, k int)
	generate = func(start, k int) {
		if k == 0 {
			combo := make([]Card, len(combination))
			copy(combo, combination)
			result = append(result, combo)
			return
		}
		
		for i := start; i <= len(cards)-k; i++ {
			combination = append(combination, cards[i])
			generate(i+1, k-1)
			combination = combination[:len(combination)-1]
		}
	}
	
	generate(0, k)
	return result
}

// CompareHands returns 1 if hand1 wins, -1 if hand2 wins, 0 if tie
func CompareHands(hand1, hand2 *Hand) int {
	if hand1.Rank > hand2.Rank {
		return 1
	}
	if hand1.Rank < hand2.Rank {
		return -1
	}
	
	if hand1.Value > hand2.Value {
		return 1
	}
	if hand1.Value < hand2.Value {
		return -1
	}
	
	return 0
}
