package bot

import (
	"fmt"
	"slices"
	"testing"
)

// These tests are for bots that search ahead. They only check moves any correct
// search must find (or avoid) at a given depth, not which of several reasonable
// moves it picks, so they stay valid as the search and evaluation change.

// searchBots lists the bots these tests run against. Add new searching bots here.
var searchBots = []BotType{NegaMax, AlphaBeta}

func forEachSearchBot(t *testing.T, test func(t *testing.T, bot Bot)) {
	t.Helper()
	for _, bt := range searchBots {
		t.Run(fmt.Sprintf("bot type %d", bt), func(t *testing.T) {
			test(t, New(bt))
		})
	}
}

func TestSearchFindsMateInOne(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want string
	}{
		// The king is boxed in by its own pawns: the rook mates on the back rank.
		{"white mates on the back rank", "6k1/5ppp/8/8/8/8/8/R5K1 w - - 0 1", "a1a8"},
		{"black mates on the back rank", "r5k1/8/8/8/8/8/5PPP/6K1 b - - 0 1", "a8a1"},
	}
	forEachSearchBot(t, func(t *testing.T, bot Bot) {
		for _, tt := range tests {
			// Depth 2 is the first depth that sees the opponent has no reply.
			for depth := 2; depth <= 3; depth++ {
				b := mustBoard(t, tt.fen)
				if got := bot.BestMove(*b, Limits{Depth: depth}).String(); got != tt.want {
					t.Errorf("%s, depth %d: BestMove = %s, want the mate %s", tt.name, depth, got, tt.want)
				}
			}
		}
	})
}

func TestSearchWinsFreeMaterial(t *testing.T) {
	// The black queen on d5 is undefended; Rxd5 wins it.
	fen := "3k4/8/8/3q4/8/8/8/3RK3 w - - 0 1"
	forEachSearchBot(t, func(t *testing.T, bot Bot) {
		for depth := 1; depth <= 3; depth++ {
			b := mustBoard(t, fen)
			if got := bot.BestMove(*b, Limits{Depth: depth}).String(); got != "d1d5" {
				t.Errorf("depth %d: BestMove = %s, want d1d5 (wins the queen)", depth, got)
			}
		}
	})
}

// Stalemate is a draw. A side a queen up must not prefer it over keeping its advantage.
func TestSearchDoesNotStalemateWhenWinning(t *testing.T) {
	// Qf7 stalemates the black king on h8 (g8, g7 and h7 are all covered).
	fen := "7k/8/6K1/8/8/8/8/5Q2 w - - 0 1"
	forEachSearchBot(t, func(t *testing.T, bot Bot) {
		for depth := 2; depth <= 3; depth++ {
			b := mustBoard(t, fen)
			if got := bot.BestMove(*b, Limits{Depth: depth}).String(); got == "f1f7" {
				t.Errorf("depth %d: BestMove = f1f7, which stalemates while a queen up", depth)
			}
		}
	})
}

// Even when every move loses, the bot must still return one of them: returning
// no move would forfeit the game.
func TestSearchReturnsALegalMoveWhenLost(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		// Kb8 is the only move, and Rh8 mates after it.
		{"only move walks into mate", "k7/8/1K6/8/8/8/8/7R b - - 0 1"},
		// The same for White: Kb1 is the only move, and Rh1 mates after it.
		{"white's only move walks into mate", "7r/8/8/8/8/1k6/8/K7 w - - 0 1"},
	}
	forEachSearchBot(t, func(t *testing.T, bot Bot) {
		for _, tt := range tests {
			for depth := 1; depth <= 4; depth++ {
				b := mustBoard(t, tt.fen)
				legal := b.GenerateLegalMoves(nil)
				if len(legal) == 0 {
					t.Fatalf("%s: test position has no legal moves", tt.name)
				}
				if got := bot.BestMove(*b, Limits{Depth: depth}); !slices.Contains(legal, got) {
					t.Errorf("%s, depth %d: BestMove = %s, which is not a legal move", tt.name, depth, got)
				}
			}
		}
	})
}

// Searching must leave the caller's board untouched at every depth.
func TestSearchDoesNotChangeBoard(t *testing.T) {
	forEachSearchBot(t, func(t *testing.T, bot Bot) {
		for _, fen := range contractFENs {
			b := mustBoard(t, fen)
			before := b.FEN()
			bot.BestMove(*b, Limits{Depth: 3})
			if after := b.FEN(); after != before {
				t.Errorf("BestMove changed the board: %s → %s", before, after)
			}
		}
	})
}
