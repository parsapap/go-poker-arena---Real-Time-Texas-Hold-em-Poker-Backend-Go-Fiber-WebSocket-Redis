package poker

import (
	"errors"
	"fmt"
	"sort"
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

// Winner records the authoritative result of a showdown (or a win by fold)
// for a single player. It is populated by Showdown/awardToLastPlayer so that
// downstream consumers (history, broadcasting) use the same numbers that were
// actually credited to player stacks, rather than recomputing them.
type Winner struct {
	PlayerID uint     `json:"player_id"`
	Username string   `json:"username"`
	Amount   int64    `json:"amount"`     // chips actually awarded to this player
	HandRank string   `json:"hand_rank"`  // empty when the player won by fold
	BestFive []Card   `json:"best_five"`  // best 5-card hand, empty on win by fold
}

type Game struct {
	ID                    uint        `json:"id"`
	RoomID                uint        `json:"room_id"`
	Phase                 GamePhase   `json:"phase"`
	Deck                  *Deck       `json:"-"`
	CommunityCards        []Card      `json:"community_cards"`
	Players               []*Player   `json:"players"`
	Pots                  []Pot       `json:"pots"`
	CurrentBet            int64       `json:"current_bet"`
	MinRaise              int64       `json:"min_raise"`
	SmallBlind            int64       `json:"small_blind"`
	BigBlind              int64       `json:"big_blind"`
	DealerPosition        int         `json:"dealer_position"`
	CurrentPosition       int         `json:"current_position"`
	LastRaiseAmount       int64       `json:"last_raise_amount"`
	LastAggressorPosition int         `json:"last_aggressor_position"` // Position of last raiser/bettor
	ActionsThisRound      int         `json:"actions_this_round"`      // Count of actions in current betting round
	// Winners holds the authoritative payout result, set by Showdown() or
	// AwardToLastPlayer(). EndRound/broadcasting should read this instead of
	// recomputing winners independently.
	Winners               []Winner    `json:"winners"`
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

	// Preflop: action starts after BB, BB is the last aggressor (posted blind)
	g.CurrentPosition = (bbPos + 1) % len(g.Players)
	g.LastAggressorPosition = bbPos // BB is considered the aggressor preflop
	g.ActionsThisRound = 0
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
	g.ActionsThisRound = 0

	// Reset player bets for new round
	for _, player := range g.Players {
		player.Bet = 0
	}

	// Set position to first active player after dealer
	// Post-flop, action starts with first active player left of dealer
	g.CurrentPosition = (g.DealerPosition + 1) % len(g.Players)
	
	// Find first active (non-folded, non-all-in) player
	for i := 0; i < len(g.Players); i++ {
		player := g.Players[g.CurrentPosition]
		if !player.Folded && !player.AllIn {
			break
		}
		g.CurrentPosition = (g.CurrentPosition + 1) % len(g.Players)
	}
	
	// No aggressor yet in new betting round (no one has bet/raised)
	g.LastAggressorPosition = -1

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
		g.LastAggressorPosition = g.CurrentPosition // Raiser becomes aggressor
		// A raise reopens the betting: every other active player must get a
		// chance to respond. Reset the action counter to 0 here; the
		// g.ActionsThisRound++ after the switch then counts the raiser as the
		// first action of this new sub-round.
		g.ActionsThisRound = 0

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
			g.LastAggressorPosition = g.CurrentPosition // All-in raise becomes aggressor
			// An all-in that exceeds the current bet is a raise, so it also
			// reopens the betting round for the remaining players.
			g.ActionsThisRound = 0
		}
	}

	g.ActionsThisRound++
	g.moveToNextPlayer()

	// Check if betting round is complete
	if g.isBettingRoundComplete() {
		return g.advance()
	}

	return nil
}

// advance moves the hand forward after a betting round completes. Normally it
// just opens the next street. But when no further betting is possible — every
// remaining contender is all-in (or only one can act) while two or more are
// still contesting the pot — it deals out all remaining streets and runs the
// showdown. Without this, an all-in confrontation stalls because NextPhase
// advances only one street per call and no further player actions will arrive.
func (g *Game) advance() error {
	if err := g.NextPhase(); err != nil {
		return err
	}
	for g.Phase != PhaseShowdown && g.Phase != PhaseFinished &&
		g.playersCanAct() < 2 && g.playersInHand() >= 2 {
		if err := g.NextPhase(); err != nil {
			return err
		}
	}
	return nil
}

// playersCanAct counts players who can still take a betting action.
func (g *Game) playersCanAct() int {
	n := 0
	for _, p := range g.Players {
		if !p.Folded && !p.AllIn {
			n++
		}
	}
	return n
}

// playersInHand counts players still contesting the pot (not folded).
func (g *Game) playersInHand() int {
	n := 0
	for _, p := range g.Players {
		if !p.Folded {
			n++
		}
	}
	return n
}

func (g *Game) collectBets() {
	// Per-round bets have been folded into each player's cumulative TotalBet
	// already (see ProcessAction). Reset the per-round Bet field and rebuild
	// the entire pot structure from TotalBet. Rebuilding from scratch (rather
	// than incrementally adding to Pots[0]) is what prevents the previous
	// double-counting bug, where chips were added to the main pot AND counted
	// again when side pots were derived from TotalBet.
	for _, player := range g.Players {
		player.Bet = 0
	}
	g.rebuildPots()
}

// rebuildPots reconstructs the main pot and any side pots purely from each
// player's cumulative TotalBet. It correctly handles:
//   - multi-round all-in scenarios (TotalBet spans every betting round)
//   - "dead money" from folded players (their chips stay in the pot but they
//     are not eligible to win it)
//   - partial contributions from players who went all-in below a bet level
func (g *Game) rebuildPots() {
	// Gather the distinct positive contribution levels across ALL players,
	// including folded ones, since their chips remain in the pot.
	levelSet := make(map[int64]struct{})
	for _, p := range g.Players {
		if p.TotalBet > 0 {
			levelSet[p.TotalBet] = struct{}{}
		}
	}

	if len(levelSet) == 0 {
		g.Pots = []Pot{{Amount: 0, Players: make([]uint, 0)}}
		return
	}

	levels := make([]int64, 0, len(levelSet))
	for level := range levelSet {
		levels = append(levels, level)
	}
	sort.Slice(levels, func(i, j int) bool { return levels[i] < levels[j] })

	// Each consecutive bet level defines one pot "layer". A layer spans from the
	// previous level to the current one; every player contributes the portion
	// of that layer their TotalBet covers.
	pots := make([]Pot, 0, len(levels))
	previousLevel := int64(0)
	for _, level := range levels {
		layer := level - previousLevel
		amount := int64(0)
		eligible := make([]uint, 0)

		for _, p := range g.Players {
			switch {
			case p.TotalBet >= level:
				// Player covers this whole layer.
				amount += layer
				if !p.Folded {
					eligible = append(eligible, p.ID)
				}
			case p.TotalBet > previousLevel:
				// Player went all-in partway through this layer; they contribute
				// only the slice between previousLevel and their TotalBet.
				amount += p.TotalBet - previousLevel
			}
		}

		if amount > 0 {
			pots = append(pots, Pot{Amount: amount, Players: eligible})
		}
		previousLevel = level
	}

	if len(pots) == 0 {
		pots = []Pot{{Amount: 0, Players: make([]uint, 0)}}
	}
	g.Pots = pots
}

func (g *Game) Showdown() error {
	g.Winners = make([]Winner, 0)

	activePlayers := make([]*Player, 0)
	for _, player := range g.Players {
		if !player.Folded {
			activePlayers = append(activePlayers, player)
		}
	}

	if len(activePlayers) == 1 {
		// Only one player left, they win all pots.
		winner := activePlayers[0]
		total := int64(0)
		for _, pot := range g.Pots {
			total += pot.Amount
		}
		winner.Chips += total
		g.Winners = append(g.Winners, Winner{
			PlayerID: winner.ID,
			Username: winner.Username,
			Amount:   total,
		})
		return nil
	}

	// Evaluate hands once per player.
	playerHands := make(map[uint]*Hand)
	for _, player := range activePlayers {
		allCards := append(player.HoleCards, g.CommunityCards...)
		playerHands[player.ID] = EvaluateHand(allCards)
	}

	// Accumulate awards per player so a player winning multiple (side) pots is
	// reported as a single Winner entry with the summed amount.
	awarded := make(map[uint]int64)

	// Distribute each pot to its eligible winner(s).
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

		// Find winner(s) of this pot.
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

		// Split pot among winners. The remainder from an uneven split is given
		// to the first winner so no chips are lost.
		share := pot.Amount / int64(len(winners))
		remainder := pot.Amount % int64(len(winners))
		for i, winner := range winners {
			amount := share
			if i == 0 {
				amount += remainder
			}
			winner.Chips += amount
			awarded[winner.ID] += amount
		}
	}

	// Build the authoritative Winners list from the amounts actually awarded.
	for _, player := range activePlayers {
		if amount, ok := awarded[player.ID]; ok && amount > 0 {
			hand := playerHands[player.ID]
			g.Winners = append(g.Winners, Winner{
				PlayerID: player.ID,
				Username: player.Username,
				Amount:   amount,
				HandRank: hand.Rank.String(),
				BestFive: hand.BestFive,
			})
		}
	}

	return nil
}

// AwardToLastPlayer awards the entire pot to the single remaining player when
// everyone else has folded, and records the result in Winners. This is the
// authoritative win-by-fold path; callers should use it instead of crediting
// chips manually so that Winners stays consistent with player stacks.
func (g *Game) AwardToLastPlayer() error {
	g.Winners = make([]Winner, 0)

	// Fold any outstanding per-round bets into the pot structure first.
	g.collectBets()

	var last *Player
	count := 0
	for _, p := range g.Players {
		if !p.Folded {
			count++
			last = p
		}
	}

	if count != 1 || last == nil {
		return errors.New("AwardToLastPlayer requires exactly one remaining player")
	}

	total := int64(0)
	for _, pot := range g.Pots {
		total += pot.Amount
	}
	last.Chips += total

	g.Phase = PhaseFinished
	g.Winners = append(g.Winners, Winner{
		PlayerID: last.ID,
		Username: last.Username,
		Amount:   total,
	})
	return nil
}

// moveToNextPlayer advances CurrentPosition to the next player who can still
// act (not folded, not all-in). It returns true if such a player was found.
// If no player can act (everyone remaining is folded or all-in), it leaves
// CurrentPosition unchanged and returns false, so the caller can detect the
// terminal state instead of spinning on an inconsistent position.
func (g *Game) moveToNextPlayer() bool {
	for i := 0; i < len(g.Players); i++ {
		next := (g.CurrentPosition + 1 + i) % len(g.Players)
		player := g.Players[next]
		if !player.Folded && !player.AllIn {
			g.CurrentPosition = next
			return true
		}
	}
	// No actionable player remains; keep CurrentPosition as-is. The betting
	// round is necessarily complete in this case.
	return false
}

func (g *Game) isBettingRoundComplete() bool {
	// Count active players (not folded, not all-in)
	activePlayers := 0
	for _, player := range g.Players {
		if !player.Folded && !player.AllIn {
			activePlayers++
		}
	}

	// If only one or zero active players, round is complete
	if activePlayers <= 1 {
		return true
	}

	// Check if all active players have matched the current bet
	allMatched := true
	for _, player := range g.Players {
		if !player.Folded && !player.AllIn {
			if player.Bet != g.CurrentBet {
				allMatched = false
				break
			}
		}
	}

	if !allMatched {
		return false
	}

	// All players have matched the current bet.
	// Round is complete when everyone has had a chance to act.
	return g.ActionsThisRound >= activePlayers
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
