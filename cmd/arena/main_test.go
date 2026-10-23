package main

import (
	"math"
	"testing"
)

func TestNewBotNames(t *testing.T) {
	for _, name := range []string{"random", "alphaBeta", "alphaBeta-noTT", "alphaBeta-noKillers-noHistory", "alphaBeta-noTT-noKillers-noHistory-noRepetition", "alphaBeta-freshTT"} {
		if _, err := newBot(name); err != nil {
			t.Errorf("newBot(%q): %v", name, err)
		}
	}
	for _, name := range []string{"", "alphabeta", "alphaBeta-", "alphaBeta-noFoo", "negaMax-noTT", "alphaBeta-noTT-"} {
		if _, err := newBot(name); err == nil {
			t.Errorf("newBot(%q) gave no error", name)
		}
	}
}

func TestEloDifference(t *testing.T) {
	tests := []struct {
		wins, losses, draws int
		elo                 float64
	}{
		{50, 50, 0, 0},
		{0, 0, 100, 0},
		{75, 25, 0, 191},  // 75% score
		{25, 75, 0, -191}, // and the other way round
		{50, 0, 50, 191},  // draws count half
	}
	for _, tt := range tests {
		elo, margin := eloDifference(tt.wins, tt.losses, tt.draws)
		if math.Abs(elo-tt.elo) > 1 {
			t.Errorf("eloDifference(%d, %d, %d) = %.1f, want about %.0f", tt.wins, tt.losses, tt.draws, elo, tt.elo)
		}
		if tt.draws < 100 && (math.IsInf(margin, 0) || margin <= 0) {
			t.Errorf("eloDifference(%d, %d, %d): margin %v, want a finite positive margin", tt.wins, tt.losses, tt.draws, margin)
		}
	}

	// More games, same proportions: a smaller margin.
	_, small := eloDifference(600, 400, 0)
	_, large := eloDifference(60, 40, 0)
	if small >= large {
		t.Errorf("margin with 1000 games %.1f, with 100 games %.1f: want more games to give a smaller margin", small, large)
	}

	for _, c := range [][3]int{{10, 0, 0}, {0, 10, 0}, {0, 0, 0}, {0, 0, 10}} {
		if _, margin := eloDifference(c[0], c[1], c[2]); !math.IsInf(margin, 1) {
			t.Errorf("eloDifference(%d, %d, %d): margin %v, want +Inf (not enough information)", c[0], c[1], c[2], margin)
		}
	}
}
