package match

import (
	"slices"
	"testing"

	"goChessBot/internal/bot"
	"goChessBot/internal/chess"
)

// scriptedBot plays a fixed list of moves (in UCI), one per call. When the script
// runs out, or a scripted move is not legal, it returns chess.Move(0).
type scriptedBot struct {
	moves []string
	next  int
}

func (s *scriptedBot) BestMove(b chess.Board, l bot.Limits) chess.Move {
	if s.next >= len(s.moves) {
		return chess.Move(0)
	}
	uci := s.moves[s.next]
	s.next++
	for _, m := range b.GenerateLegalMoves(nil) {
		if m.String() == uci {
			return m
		}
	}
	return chess.Move(0)
}

func script(moves ...string) *scriptedBot {
	return &scriptedBot{moves: moves}
}

// illegalBot always returns a move that is never legal.
type illegalBot struct{}

func (illegalBot) BestMove(b chess.Board, l bot.Limits) chess.Move { return chess.Move(0) }

func uciList(moves []chess.Move) []string {
	out := make([]string, len(moves))
	for i, m := range moves {
		out[i] = m.String()
	}
	return out
}

func mustPlay(t *testing.T, white, black bot.Bot, fen string, maxPlies int) Game {
	t.Helper()
	g, err := Play(white, black, fen, maxPlies, bot.Limits{Depth: 1})
	if err != nil {
		t.Fatalf("Play(%q): unexpected error %v", fen, err)
	}
	return g
}

func checkGame(t *testing.T, g Game, wantMoves []string, wantResult chess.GameResult, wantWinner chess.PieceId, wantForfeit bool) {
	t.Helper()
	if got := uciList(g.Moves); !slices.Equal(got, wantMoves) {
		t.Errorf("Moves = %v, want %v", got, wantMoves)
	}
	if g.Result != wantResult {
		t.Errorf("Result = %v, want %v", g.Result, wantResult)
	}
	if g.Winner != wantWinner {
		t.Errorf("Winner = %d, want %d", g.Winner, wantWinner)
	}
	if g.Forfeit != wantForfeit {
		t.Errorf("Forfeit = %v, want %v", g.Forfeit, wantForfeit)
	}
}

func TestPlayCheckmate(t *testing.T) {
	t.Run("black mates (fool's mate)", func(t *testing.T) {
		g := mustPlay(t, script("f2f3", "g2g4"), script("e7e5", "d8h4"), "", 100)
		checkGame(t, g, []string{"f2f3", "e7e5", "g2g4", "d8h4"}, chess.Checkmate, chess.Black, false)
	})
	t.Run("white mates (scholar's mate)", func(t *testing.T) {
		g := mustPlay(t,
			script("e2e4", "f1c4", "d1h5", "h5f7"),
			script("e7e5", "b8c6", "g8f6"),
			"", 100)
		checkGame(t, g,
			[]string{"e2e4", "e7e5", "f1c4", "b8c6", "d1h5", "g8f6", "h5f7"},
			chess.Checkmate, chess.White, false)
	})
}

func TestPlayDraws(t *testing.T) {
	t.Run("threefold repetition", func(t *testing.T) {
		g := mustPlay(t,
			script("g1f3", "f3g1", "g1f3", "f3g1"),
			script("g8f6", "f6g8", "g8f6", "f6g8"),
			"", 100)
		checkGame(t, g,
			[]string{"g1f3", "g8f6", "f3g1", "f6g8", "g1f3", "g8f6", "f3g1", "f6g8"},
			chess.ThreeFoldRepetition, 0, false)
	})
	t.Run("insufficient material after a capture", func(t *testing.T) {
		g := mustPlay(t, script("e1e2"), script(), "4k3/8/8/8/8/8/4q3/4K3 w - - 0 1", 100)
		checkGame(t, g, []string{"e1e2"}, chess.InsufficientMaterial, 0, false)
	})
	t.Run("fifty-move rule", func(t *testing.T) {
		g := mustPlay(t, script("a1a2"), script(), "4k3/8/8/8/8/8/8/R3K3 w - - 99 80", 100)
		checkGame(t, g, []string{"a1a2"}, chess.FiftyMoveRule, 0, false)
	})
	t.Run("stalemate", func(t *testing.T) {
		// Qf7 leaves the black king on h8 with no moves and not in check.
		g := mustPlay(t, script("e7f7"), script(), "7k/4Q3/6K1/8/8/8/8/8 w - - 0 1", 100)
		checkGame(t, g, []string{"e7f7"}, chess.Stalemate, 0, false)
	})
}

func TestPlayForfeit(t *testing.T) {
	t.Run("white plays an illegal move", func(t *testing.T) {
		g := mustPlay(t, illegalBot{}, script("e7e5"), "", 100)
		checkGame(t, g, []string{}, chess.Ongoing, chess.Black, true)
	})
	t.Run("black plays an illegal move after one move", func(t *testing.T) {
		g := mustPlay(t, script("e2e4"), illegalBot{}, "", 100)
		checkGame(t, g, []string{"e2e4"}, chess.Ongoing, chess.White, true)
	})
	t.Run("a bot runs out of moves", func(t *testing.T) {
		// The scripted bot returns chess.Move(0) once its script is used up.
		g := mustPlay(t, script("e2e4", "d2d4"), script("e7e5"), "", 100)
		checkGame(t, g, []string{"e2e4", "e7e5", "d2d4"}, chess.Ongoing, chess.White, true)
	})
}

func TestPlayMaxPlies(t *testing.T) {
	g := mustPlay(t,
		script("g1f3", "f3g1", "g1f3"),
		script("g8f6", "f6g8", "g8f6"),
		"", 3)
	checkGame(t, g, []string{"g1f3", "g8f6", "f3g1"}, chess.Ongoing, 0, false)
}

func TestPlayStartingPositions(t *testing.T) {
	t.Run("already checkmate: no moves are played", func(t *testing.T) {
		g := mustPlay(t, illegalBot{}, illegalBot{}, "R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1", 100)
		checkGame(t, g, []string{}, chess.Checkmate, chess.White, false)
	})
	t.Run("black to move: the black bot moves first", func(t *testing.T) {
		g := mustPlay(t, illegalBot{}, script("e7e5"), "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1", 1)
		checkGame(t, g, []string{"e7e5"}, chess.Ongoing, 0, false)
	})
	t.Run("empty FEN means the start position", func(t *testing.T) {
		g := mustPlay(t, script("e2e4"), illegalBot{}, "", 1)
		checkGame(t, g, []string{"e2e4"}, chess.Ongoing, 0, false)
	})
	t.Run("invalid FEN returns an error", func(t *testing.T) {
		if _, err := Play(illegalBot{}, illegalBot{}, "8/8/8/8/8/8/8/8 w - - x 1", 100, bot.Limits{Depth: 1}); err == nil {
			t.Errorf("Play with an unparsable FEN returned no error")
		}
	})
}

// Random bots: replaying each recorded game must give only legal moves and end in the
// recorded result, and the winner must match the result.
func TestPlayRandomGamesAreConsistent(t *testing.T) {
	random := bot.New(bot.Random)
	for i := 0; i < 20; i++ {
		g := mustPlay(t, random, random, "", 300)

		b, _ := chess.CreateBoard("")
		var history []uint64
		for ply, m := range g.Moves {
			if !slices.Contains(b.GenerateLegalMoves(nil), m) {
				t.Fatalf("game %d, ply %d: recorded move %s is not legal", i, ply, m)
			}
			history = append(history, b.Hash)
			b.MakeMove(m)
		}

		want := b.ResultWithMoves(b.GenerateLegalMoves(nil))
		if want == chess.Ongoing && b.IsThreeFoldRepetition(history) {
			want = chess.ThreeFoldRepetition
		}
		if g.Forfeit {
			t.Errorf("game %d: random bot forfeited", i)
		}
		if g.Result != want {
			t.Errorf("game %d: Result = %v, but replaying the moves gives %v", i, g.Result, want)
		}
		if g.Result == chess.Ongoing && len(g.Moves) != 300 {
			t.Errorf("game %d: unfinished after %d plies, want exactly maxPlies (300)", i, len(g.Moves))
		}

		var wantWinner chess.PieceId
		if g.Result == chess.Checkmate {
			wantWinner = b.ToMove.OppositeColor()
		}
		if g.Winner != wantWinner {
			t.Errorf("game %d: Winner = %d, want %d for result %v", i, g.Winner, wantWinner, g.Result)
		}
	}
}
