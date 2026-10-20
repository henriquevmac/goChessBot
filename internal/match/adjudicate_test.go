package match

import (
	"testing"

	"goChessBot/internal/chess"
)

func mustBoard(t *testing.T, fen string) *chess.Board {
	t.Helper()
	b, err := chess.CreateBoard(fen)
	if err != nil {
		t.Fatalf("CreateBoard(%q): %v", fen, err)
	}
	return b
}

func TestMaterial(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want int
	}{
		{"start position is level", "", 0},
		{"kings only", "4k3/8/8/8/8/8/8/4K3 w - - 0 1", 0},
		{"white extra queen", "4k3/8/8/8/8/8/8/3QK3 w - - 0 1", 900},
		{"black extra rook", "r3k3/8/8/8/8/8/8/4K3 w - - 0 1", -500},
		{"knight 320, bishop 330", "4k3/8/8/8/8/8/8/1NB1K3 b - - 0 1", 650},
		{"pawns are worth 100", "4k3/pppp4/8/8/8/8/PP6/4K3 w - - 0 1", -200},
		{"side to move does not matter", "4k3/8/8/8/8/8/8/3QK3 b - - 0 1", 900},
		{"mixed material", "r1b1k3/8/8/8/8/8/PPP5/RN2K3 w - - 0 1", 300 + 500 + 320 - 500 - 330},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			if got := Material(b); got != tt.want {
				t.Errorf("Material(%q) = %d, want %d", tt.fen, got, tt.want)
			}
		})
	}
}

// Play must record the position the game ended in.
func TestPlayRecordsFinalPosition(t *testing.T) {
	tests := []struct {
		name     string
		white    *scriptedBot
		black    *scriptedBot
		fen      string
		maxPlies int
		wantFEN  string
	}{
		{
			name:  "checkmate",
			white: script("f2f3", "g2g4"), black: script("e7e5", "d8h4"),
			maxPlies: 100,
			wantFEN:  "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3",
		},
		{
			name:  "move limit",
			white: script("g1f3", "f3g1"), black: script("g8f6"),
			maxPlies: 3,
			wantFEN:  "rnbqkb1r/pppppppp/5n2/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 3 2",
		},
		{
			name:  "forfeit: the position before the illegal move",
			white: script("e2e4"), black: script(),
			maxPlies: 100,
			wantFEN:  "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		},
		{
			name:  "already over: the starting position",
			white: script(), black: script(),
			fen:      "R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1",
			maxPlies: 100,
			wantFEN:  "R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := mustPlay(t, tt.white, tt.black, tt.fen, tt.maxPlies)
			if got := g.Final.FEN(); got != tt.wantFEN {
				t.Errorf("Final.FEN() =\n  %q\nwant\n  %q", got, tt.wantFEN)
			}
		})
	}
}

func TestAdjudicate(t *testing.T) {
	// White is a queen up; both bots shuffle their kings until the move limit.
	queenUp := "4k3/8/8/8/8/8/8/3QK3 w - - 0 1"
	// Black is a rook up.
	rookDown := "r3k3/8/8/8/8/8/8/4K3 w - - 0 1"
	// Level material: a rook each (kings only would already be a draw by insufficient material).
	level := "r3k3/8/8/8/8/8/8/R3K3 w - - 0 1"

	unfinished := func(fen string) Game {
		t.Helper()
		g := mustPlay(t, script("e1f1"), script("e8f8"), fen, 2)
		if g.Result != chess.Ongoing || g.Forfeit {
			t.Fatalf("setup: expected an unfinished game, got Result %v, Forfeit %v", g.Result, g.Forfeit)
		}
		return g
	}

	t.Run("lead above threshold wins for white", func(t *testing.T) {
		g := Adjudicate(unfinished(queenUp), 300)
		if !g.Adjudicated || g.Winner != chess.White || g.Result != chess.Ongoing {
			t.Errorf("got Adjudicated %v, Winner %d, Result %v; want true, White, Ongoing", g.Adjudicated, g.Winner, g.Result)
		}
	})
	t.Run("lead above threshold wins for black", func(t *testing.T) {
		g := Adjudicate(unfinished(rookDown), 300)
		if !g.Adjudicated || g.Winner != chess.Black {
			t.Errorf("got Adjudicated %v, Winner %d; want true, Black", g.Adjudicated, g.Winner)
		}
	})
	t.Run("lead exactly at the threshold wins", func(t *testing.T) {
		g := Adjudicate(unfinished(rookDown), 500)
		if !g.Adjudicated || g.Winner != chess.Black {
			t.Errorf("got Adjudicated %v, Winner %d; want true, Black", g.Adjudicated, g.Winner)
		}
	})
	t.Run("lead below threshold is a draw", func(t *testing.T) {
		g := Adjudicate(unfinished(rookDown), 600)
		if !g.Adjudicated || g.Winner != 0 {
			t.Errorf("got Adjudicated %v, Winner %d; want true, 0 (draw)", g.Adjudicated, g.Winner)
		}
	})
	t.Run("level material is a draw", func(t *testing.T) {
		g := Adjudicate(unfinished(level), 300)
		if !g.Adjudicated || g.Winner != 0 {
			t.Errorf("got Adjudicated %v, Winner %d; want true, 0 (draw)", g.Adjudicated, g.Winner)
		}
	})
	t.Run("threshold 0 means off", func(t *testing.T) {
		before := unfinished(queenUp)
		g := Adjudicate(before, 0)
		if g.Adjudicated || g.Winner != 0 {
			t.Errorf("got Adjudicated %v, Winner %d; want the game unchanged", g.Adjudicated, g.Winner)
		}
	})
	t.Run("the moves are kept", func(t *testing.T) {
		before := unfinished(queenUp)
		g := Adjudicate(before, 300)
		if len(g.Moves) != len(before.Moves) {
			t.Errorf("Moves changed: %d moves, want %d", len(g.Moves), len(before.Moves))
		}
	})

	t.Run("finished games are not touched", func(t *testing.T) {
		// Fool's mate: Black wins by checkmate, even though material is level.
		mate := mustPlay(t, script("f2f3", "g2g4"), script("e7e5", "d8h4"), "", 100)
		g := Adjudicate(mate, 100)
		if g.Adjudicated || g.Winner != chess.Black || g.Result != chess.Checkmate {
			t.Errorf("checkmate changed: Adjudicated %v, Winner %d, Result %v", g.Adjudicated, g.Winner, g.Result)
		}
		// A draw by insufficient material stays a draw.
		draw := mustPlay(t, script("e1e2"), script(), "4k3/8/8/8/8/8/4q3/4K3 w - - 0 1", 100)
		g = Adjudicate(draw, 100)
		if g.Adjudicated || g.Winner != 0 || g.Result != chess.InsufficientMaterial {
			t.Errorf("draw changed: Adjudicated %v, Winner %d, Result %v", g.Adjudicated, g.Winner, g.Result)
		}
	})

	t.Run("forfeits are not touched", func(t *testing.T) {
		// White forfeits while a queen up: the forfeit stands.
		forfeit := mustPlay(t, illegalBot{}, script(), queenUp, 100)
		g := Adjudicate(forfeit, 300)
		if g.Adjudicated || g.Winner != chess.Black || !g.Forfeit {
			t.Errorf("forfeit changed: Adjudicated %v, Winner %d, Forfeit %v", g.Adjudicated, g.Winner, g.Forfeit)
		}
	})
}
