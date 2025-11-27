package poker

import (
	"errors"
	"fmt"
)

type GamePhase string

const (
	PhaseWaiting   GamePhase = "waiting"
	PhasePreFlop   GamePhase = "preflop"
	PhaseFlop      GamePhase = "flop"
	PhaseTurn      GamePhase = "turn"
	PhaseRiver     GamePhase = "river"
	PhaseShowdown  GamePhase = "showdown"
	PhaseFinished  GamePhase = "finished"
)

type Action string

const (
	ActionFold  Action = "fold"
	ActionCheck Action = "check"
	ActionCall  Action = "call"
	ActionRaise Action = "raise"
	ActionAllIn Action = "allin"
)

type Player struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Chips       int64  `json:"chips"`
	Bet         int64  `json:"bet"`
	TotalBet    int64  `json:"total_bet"`
	Position    int    `json:"position"`
	HoleCards   []Card `json:"hole_cards,omitempty"`
	Folded      bool   `json:"folded"`
	AllIn       bool   `json:"all_in"`
	Active      bool   `json:"active"`
}

type Pot struct {
	Amount  int64    `json:"amount"`
	Players []uint   `json:"players"`
}

type Game struct {
	ID              uint        `json:"id"`
	RoomID          uint        `json:"room_id"`
	Phase           GamePhase   `json:"phase"`
	Deck            *Deck       `json:"-"`
	CommunityCards  []Card      `json:"community_cards"`
	Players         []*Player   `json:"players"`
	Pots            []Pot       `json:"pots"`
	CurrentBet      int64       `json:"current_bet"`
	MinRaise        int64       `json:"min_raise"`
	SmallBlind      int64       `json:"small_blind"`
	BigBlind        int64       `json:"big_blind"`
	DealerPosition  int         `json:"dealer_position"`
	CurrentPosition int         `json:"current_position"`
	LastRaiseAmount int64       `json:"last_raise_amount"`
}

func NewGame(roomID uint, players []*Player, smallBlind, bigBlind int64) *Game {
	return &Game{
		RoomID:         roomID,
		Phase:          PhaseWaiting,
		Players:        players,
		Pots:           []Pot{{Amount: 0, Players: make([]uint, 0)}},
		SmallBlind:     smallBlind,
		BigBlind:       bigBlind,
		DealerPosition: 0,
		CommunityCards: make([]Card, 0),
	}
}

func (g *Game) Start() error {
	if len(g.Players) < 2 {
		return errors.New("need at least 2 players to start")
	}

	g.Deck = NewDeck()
	g.Deck.Shuffle()
	g.Phase = PhasePreFlop
	g.CurrentBet = g.BigBlind
	g.MinRaise = g.BigBlind

	// Post blinds
	sbPos := (g.DealerPosition + 1) % len(g.Players)
	bbPos := (g.DealerPosition + 2) % len(g.Players)

	g.Players[sbPos].Bet = g.SmallBlind
	g.Players[sbPos].TotalBet = g.SmallBlind
	g.Players[sbPos].Chips -= g.SmallBlind

	g.Players[bbPos].Bet = g.BigBlind
	g.Players[bbPos].TotalBet = g.BigBlind
	g.Players[bbPos].Chips -= g.BigBlind

	g.Pots[0].Amount = g.SmallBlind + g.BigBlind

	// Deal hole cards
	for _, player := range g.Players {
		if !player.Folded {
			player.HoleCards = g.Deck.DrawN(2)
			player.Active = true
		}
	}

	g.CurrentPosition = (bbPos + 1) % len(g.Players)
	return nil
}

func (g *Game) NextPhase() error {
	// Collect bets to pot
	g.collectBets()

	switch g.Phase {
	case PhasePreFlop:
		g.Phase = PhaseFlop
		g.CommunityCards = append(g.CommunityCards, g.Deck.DrawN(3)...)
	case PhaseFlop:
		g.Phase = PhaseTurn
		g.CommunityCards = append(g.CommunityCards, g.Deck.Draw())
	case PhaseTurn:
		g.Phase = PhaseRiver
		g.CommunityCards = append(g.CommunityCards, g.Deck.Draw())
	case PhaseRiver:
		g.Phase = PhaseShowdown
		return g.Showdown()
	case PhaseShowdown:
		g.Phase = PhaseFinished
		return nil
	default:
		return errors.New("invalid phase transition")
	}

	g.CurrentBet = 0
	g.MinRaise = g.BigBlind
	g.CurrentPosition = (g.DealerPosition + 1) % len(g.Players)

	// Reset player bets for new round
	for _, player := range g.Players {
		player.Bet = 0
	}

	return nil
}

func (g *Game) ProcessAction(playerID uint, action Action, amount int64) error {
	player := g.getPlayer(playerID)
	if player == nil {
		return errors.New("player not found")
	}

	if g.Players[g.CurrentPosition].ID != playerID {
		return errors.New("not player's turn")
	}

	if player.Folded || player.AllIn {
		return errors.New("player cannot act")
	}

	switch action {
	case ActionFold:
		player.Folded = true
		player.Active = false

	case ActionCheck:
		if player.Bet < g.CurrentBet {
			return errors.New("cannot check, must call or raise")
		}

	case ActionCall:
		callAmount := g.CurrentBet - player.Bet
		if callAmount > player.Chips {
			callAmount = player.Chips
			player.AllIn = true
		}
		player.Chips -= callAmount
		player.Bet += callAmount
		player.TotalBet += callAmount

	case ActionRaise:
		if amount < g.MinRaise {
			return fmt.Errorf("raise must be at least %d", g.MinRaise)
		}
		totalAmount := g.CurrentBet - player.Bet + amount
		if totalAmount > player.Chips {
			return errors.New("insufficient chips")
		}
		player.Chips -= totalAmount
		player.Bet += totalAmount
		player.TotalBet += totalAmount
		g.LastRaiseAmount = amount
		g.CurrentBet = player.Bet
		g.MinRaise = amount

	case ActionAllIn:
		allInAmount := player.Chips
		player.Chips = 0
		player.Bet += allInAmount
		player.TotalBet += allInAmount
		player.AllIn = true
		if player.Bet > g.CurrentBet {
			g.LastRaiseAmount = player.Bet - g.CurrentBet
			g.CurrentBet = player.Bet
			g.MinRaise = g.LastRaiseAmount
		}
	}

	g.moveToNextPlayer()

	// Check if betting round is complete
	if g.isBettingRoundComplete() {
		return g.NextPhase()
	}

	return nil
}

func (g *Game) collectBets() {
	for _, player := range g.Players {
		if player.Bet > 0 {
			g.Pots[0].Amount += player.Bet
			player.Bet = 0
		}
	}
	g.createSidePots()
}

func (g *Game) createSidePots() {
	// Collect all unique bet amounts from all-in players
	type betLevel struct {
		amount  int64
		players []uint
	}
	
	betLevels := make(map[int64][]uint)
	
	// Group players by their total bet amounts
	for _, player := range g.Players {
		if !player.Folded && player.TotalBet > 0 {
			betLevels[player.TotalBet] = append(betLevels[player.TotalBet], player.ID)
		}
	}
	
	// Sort bet levels
	levels := make([]int64, 0, len(betLevels))
	for level := range betLevels {
		levels = append(levels, level)
	}
	
	// Simple bubble sort for small arrays
	for i := 0; i < len(levels); i++ {
		for j := i + 1; j < len(levels); j++ {
			if levels[i] > levels[j] {
				levels[i], levels[j] = levels[j], levels[i]
			}
		}
	}
	
	// Create side pots based on bet levels
	previousLevel := int64(0)
	remainingPlayers := make([]uint, 0)
	
	for _, player := range g.Players {
		if !player.Folded {
			remainingPlayers = append(remainingPlayers, player.ID)
		}
	}
	
	for _, level := range levels {
		if len(remainingPlayers) == 0 {
			break
		}
		
		potAmount := int64(0)
		eligiblePlayers := make([]uint, 0)
		
		for _, playerID := range remainingPlayers {
			player := g.getPlayer(playerID)
			if player != nil && player.TotalBet >= level {
				contribution := level - previousLevel
				potAmount += contribution
				eligiblePlayers = append(eligiblePlayers, playerID)
			}
		}
		
		if potAmount > 0 && len(eligiblePlayers) > 0 {
			g.Pots = append(g.Pots, Pot{
				Amount:  potAmount,
				Players: eligiblePlayers,
			})
		}
		
		// Remove players who are all-in at this level
		newRemaining := make([]uint, 0)
		for _, playerID := range remainingPlayers {
			player := g.getPlayer(playerID)
			if player != nil && player.TotalBet > level {
				newRemaining = append(newRemaining, playerID)
			}
		}
		remainingPlayers = newRemaining
		previousLevel = level
	}
}

func (g *Game) Showdown() error {
	activePlayers := make([]*Player, 0)
	for _, player := range g.Players {
		if !player.Folded {
			activePlayers = append(activePlayers, player)
		}
	}

	if len(activePlayers) == 1 {
		// Only one player left, they win all pots
		winner := activePlayers[0]
		for _, pot := range g.Pots {
			winner.Chips += pot.Amount
		}
		return nil
	}

	// Evaluate hands
	playerHands := make(map[uint]*Hand)
	for _, player := range activePlayers {
		allCards := append(player.HoleCards, g.CommunityCards...)
		playerHands[player.ID] = EvaluateHand(allCards)
	}

	// Distribute each pot
	for _, pot := range g.Pots {
		eligiblePlayers := make([]*Player, 0)
		for _, playerID := range pot.Players {
			player := g.getPlayer(playerID)
			if player != nil && !player.Folded {
				eligiblePlayers = append(eligiblePlayers, player)
			}
		}

		if len(eligiblePlayers) == 0 {
			continue
		}

		// Find winner(s)
		winners := []*Player{eligiblePlayers[0]}
		bestHand := playerHands[eligiblePlayers[0].ID]

		for i := 1; i < len(eligiblePlayers); i++ {
			player := eligiblePlayers[i]
			hand := playerHands[player.ID]
			cmp := CompareHands(hand, bestHand)
			if cmp > 0 {
				winners = []*Player{player}
				bestHand = hand
			} else if cmp == 0 {
				winners = append(winners, player)
			}
		}

		// Split pot among winners
		share := pot.Amount / int64(len(winners))
		for _, winner := range winners {
			winner.Chips += share
		}
	}

	return nil
}

func (g *Game) moveToNextPlayer() {
	for i := 0; i < len(g.Players); i++ {
		g.CurrentPosition = (g.CurrentPosition + 1) % len(g.Players)
		player := g.Players[g.CurrentPosition]
		if !player.Folded && !player.AllIn {
			return
		}
	}
}

func (g *Game) isBettingRoundComplete() bool {
	activePlayers := 0
	playersActed := 0

	for _, player := range g.Players {
		if !player.Folded {
			activePlayers++
			if player.Bet == g.CurrentBet || player.AllIn {
				playersActed++
			}
		}
	}

	return activePlayers <= 1 || playersActed == activePlayers
}

func (g *Game) GetPlayer(playerID uint) *Player {
	for _, player := range g.Players {
		if player.ID == playerID {
			return player
		}
	}
	return nil
}

func (g *Game) getPlayer(playerID uint) *Player {
	return g.GetPlayer(playerID)
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
