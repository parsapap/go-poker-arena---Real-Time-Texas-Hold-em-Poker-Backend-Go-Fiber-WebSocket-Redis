package poker

import "testing"

func newTestPlayers(n int, chips int64) []*Player {
	players := make([]*Player, n)
	for i := 0; i < n; i++ {
		players[i] = &Player{
			ID:       uint(i + 1),
			Username: string(rune('A' + i)),
			Chips:    chips,
			Position: i,
			Active:   true,
		}
	}
	return players
}

// TestReRaiseReopensRound verifies that a re-raise resets ActionsThisRound so
// that every other active player gets another chance to act before the
// betting round is considered complete.
func TestReRaiseReopensRound(t *testing.T) {
	// 3 players, deep stacks so nobody goes all-in.
	g := NewGame(1, newTestPlayers(3, 1000), 10, 20)
	if err := g.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Preflop with dealer=0: SB=pos1, BB=pos2, first to act = pos0.
	if g.Phase != PhasePreFlop {
		t.Fatalf("expected preflop, got %s", g.Phase)
	}

	// Player at pos0 raises by 20 (to 40 total).
	first := g.Players[g.CurrentPosition]
	if err := g.ProcessAction(first.ID, ActionRaise, 20); err != nil {
		t.Fatalf("raise: %v", err)
	}
	// A raise must reopen the round: still preflop, not advanced to flop.
	if g.Phase != PhasePreFlop {
		t.Fatalf("round ended prematurely after a raise; phase=%s", g.Phase)
	}

	// SB re-raises. This must keep the round open for the BB to respond.
	reRaiser := g.Players[g.CurrentPosition]
	if err := g.ProcessAction(reRaiser.ID, ActionRaise, 20); err != nil {
		t.Fatalf("re-raise: %v", err)
	}
	if g.Phase != PhasePreFlop {
		t.Fatalf("round ended after re-raise before all players acted; phase=%s", g.Phase)
	}
}

// TestSidePotNoDoubleCount verifies pot accounting in a multi-bet all-in
// scenario: the total of all pots must equal the total chips committed, with
// no double-counting.
func TestSidePotNoDoubleCount(t *testing.T) {
	g := NewGame(1, newTestPlayers(3, 1000), 10, 20)
	g.Phase = PhasePreFlop

	// Simulate committed bets directly via TotalBet (the source of truth for
	// pot construction): A=100, B=300, C=300.
	g.Players[0].TotalBet = 100
	g.Players[0].Bet = 100
	g.Players[1].TotalBet = 300
	g.Players[1].Bet = 300
	g.Players[2].TotalBet = 300
	g.Players[2].Bet = 300

	g.collectBets()

	total := int64(0)
	for _, p := range g.Pots {
		total += p.Amount
	}
	if total != 700 {
		t.Fatalf("pot total = %d, want 700 (no double-count)", total)
	}

	// Expect a main pot of 300 (3 x 100) eligible to all three, and a side
	// pot of 400 (2 x 200) eligible to B and C only.
	if len(g.Pots) != 2 {
		t.Fatalf("expected 2 pots, got %d", len(g.Pots))
	}
	if g.Pots[0].Amount != 300 {
		t.Errorf("main pot = %d, want 300", g.Pots[0].Amount)
	}
	if len(g.Pots[0].Players) != 3 {
		t.Errorf("main pot eligible = %d, want 3", len(g.Pots[0].Players))
	}
	if g.Pots[1].Amount != 400 {
		t.Errorf("side pot = %d, want 400", g.Pots[1].Amount)
	}
	if len(g.Pots[1].Players) != 2 {
		t.Errorf("side pot eligible = %d, want 2", len(g.Pots[1].Players))
	}
}

// TestFoldedPlayerChipsStayInPot verifies dead money from a folded player
// remains in the pot but the folded player is not eligible to win it.
func TestFoldedPlayerChipsStayInPot(t *testing.T) {
	g := NewGame(1, newTestPlayers(3, 1000), 10, 20)
	g.Phase = PhasePreFlop

	// A bets 50 then folds; B and C each commit 50.
	g.Players[0].TotalBet = 50
	g.Players[0].Bet = 50
	g.Players[0].Folded = true
	g.Players[1].TotalBet = 50
	g.Players[1].Bet = 50
	g.Players[2].TotalBet = 50
	g.Players[2].Bet = 50

	g.collectBets()

	total := int64(0)
	for _, p := range g.Pots {
		total += p.Amount
	}
	if total != 150 {
		t.Fatalf("pot total = %d, want 150 (folded chips stay in)", total)
	}
	// The single pot should only list non-folded players as eligible.
	for _, p := range g.Pots {
		for _, id := range p.Players {
			if id == g.Players[0].ID {
				t.Errorf("folded player %d should not be eligible to win", id)
			}
		}
	}
}

// TestMoveToNextPlayerNoActionable verifies moveToNextPlayer reports false and
// leaves position unchanged when no player can act.
func TestMoveToNextPlayerNoActionable(t *testing.T) {
	g := NewGame(1, newTestPlayers(3, 1000), 10, 20)
	for _, p := range g.Players {
		p.AllIn = true
	}
	g.CurrentPosition = 1

	if g.moveToNextPlayer() {
		t.Fatal("expected no actionable player, got true")
	}
	if g.CurrentPosition != 1 {
		t.Fatalf("position changed to %d, want it unchanged at 1", g.CurrentPosition)
	}
}
