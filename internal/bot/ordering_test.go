package bot

import (
	"fmt"
	"slices"
	"testing"

	"goChessBot/internal/chess"
)

// orderingFENs mix captures of every kind, promotions and quiet moves.
var orderingFENs = []string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
	"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
	"4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2",
	"3r1r2/4P3/8/8/8/8/8/k3K3 w - - 0 1",
}

func uciList(moves []chess.Move) []string {
	out := make([]string, len(moves))
	for i, m := range moves {
		out[i] = m.String()
	}
	return out
}

func ordered(t *testing.T, fen string) (*chess.Board, []chess.Move) {
	t.Helper()
	b := mustBoard(t, fen)
	moves := b.GenerateLegalMoves(nil)
	testSearcher(b).orderMoves(moves, chess.Move(0))
	return b, moves
}

func indexOf(t *testing.T, moves []chess.Move, uci string) int {
	t.Helper()
	for i, m := range moves {
		if m.String() == uci {
			return i
		}
	}
	t.Fatalf("move %s not in the list", uci)
	return -1
}

// Ordering must only rearrange the moves: same moves, none lost or duplicated.
func TestOrderMovesIsAPermutation(t *testing.T) {
	for _, fen := range orderingFENs {
		b := mustBoard(t, fen)
		original := b.GenerateLegalMoves(nil)
		moves := slices.Clone(original)
		before := *b
		testSearcher(b).orderMoves(moves, chess.Move(0))

		want, got := slices.Clone(original), slices.Clone(moves)
		slices.Sort(want)
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Errorf("%s: orderMoves changed the set of moves\n got  %v\n want %v", fen, uciList(got), uciList(want))
		}
		if b.FEN() != before.FEN() {
			t.Errorf("%s: orderMoves changed the board", fen)
		}
	}
}

// Orders a capture-only list too (as quiescence does), without needing the full move list.
func TestOrderMovesWorksOnASubset(t *testing.T) {
	b := mustBoard(t, orderingFENs[1])
	captures := b.GenerateCaptureMoves(nil)
	moves := slices.Clone(captures)
	testSearcher(b).orderMoves(moves, chess.Move(0))
	want, got := slices.Clone(captures), slices.Clone(moves)
	slices.Sort(want)
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("ordering the captures changed the list\n got  %v\n want %v", uciList(got), uciList(want))
	}
}

func TestCapturesBeforeQuietMoves(t *testing.T) {
	for _, fen := range orderingFENs {
		_, moves := ordered(t, fen)
		seenQuiet := ""
		for _, m := range moves {
			switch {
			case !m.IsCapture() && !m.IsPromotion():
				if seenQuiet == "" {
					seenQuiet = m.String()
				}
			case m.IsCapture() && seenQuiet != "":
				t.Errorf("%s: capture %s comes after the quiet move %s", fen, m, seenQuiet)
			}
		}
	}
}

// Most valuable victim first; for the same victim, least valuable attacker first.
func TestMVVLVA(t *testing.T) {
	tests := []struct {
		name          string
		fen           string
		before, after string // before must be ordered ahead of after
	}{
		// The queen on d5 can be taken by the pawn (e4) or by the queen (d1).
		{"pawn takes queen before queen takes queen",
			"4k3/8/8/3q4/4P3/8/8/3QK3 w - - 0 1", "e4d5", "d1d5"},
		// The knight on c3 can take the queen on d5 or the rook on a4.
		{"capturing a queen before capturing a rook",
			"4k3/8/8/3q4/r7/2N5/8/4K3 w - - 0 1", "c3d5", "c3a4"},
		// Knight takes pawn vs. queen takes rook: the rook is the bigger victim.
		{"bigger victim wins even with a bigger attacker",
			"4k3/8/8/2p5/8/1N6/8/r2QK3 w - - 0 1", "d1a1", "b3c5"},
		// The rook on b3 can be taken by the king or by the knight.
		{"king is the last attacker for the same victim",
			"4k3/8/8/8/8/1r6/2K5/N7 w - - 0 1", "a1b3", "c2b3"},
		{"en passant counts as capturing a pawn",
			"4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2", "e5d6", "e5e6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, moves := ordered(t, tt.fen)
			if i, j := indexOf(t, moves, tt.before), indexOf(t, moves, tt.after); i > j {
				t.Errorf("%s is ordered after %s; order: %v", tt.before, tt.after, uciList(moves))
			}
		})
	}
}

func TestPromotionsBeforeQuietMoves(t *testing.T) {
	_, moves := ordered(t, "4k3/P7/8/8/8/8/8/4K3 w - - 0 1")
	queen := indexOf(t, moves, "a7a8q")
	for i, m := range moves {
		if !m.IsCapture() && !m.IsPromotion() && i < queen {
			t.Errorf("quiet move %s is ordered before the queen promotion; order: %v", m, uciList(moves))
		}
	}
	if knight := indexOf(t, moves, "a7a8n"); knight < queen {
		t.Errorf("knight promotion is ordered before the queen promotion")
	}
}

// referenceScore is a plain negamax (no pruning, no ordering) whose leaves call
// quiesce with a full window, the same leaf alphaBeta uses. Checkmate scores
// matedScore(ply), stalemate 0, like in alphaBeta.
func referenceScore(s *searcher, depth int) int {
	return referenceScoreAt(s, depth, 0)
}

func referenceScoreAt(s *searcher, depth, ply int) int {
	if depth == 0 {
		return s.quiesce(-infinite, infinite)
	}
	moves := s.board.GenerateLegalMoves(nil)
	if len(moves) == 0 {
		if s.board.InCheck() {
			return matedScore(ply)
		}
		return 0
	}
	best := -infinite
	for _, m := range moves {
		u := s.board.MakeMove(m)
		best = max(best, -referenceScoreAt(s, depth-1, ply+1))
		s.board.UnmakeMove(m, u)
	}
	return best
}

// Alpha-beta only skips work, so at a fixed depth it must return exactly the
// score of a full negamax with the same leaves, however the moves are ordered.
// The reference has no pruning and runs quiescence at every leaf, so in busy
// positions it gets slow very quickly; maxDepth keeps the whole test well under
// a second.
func TestAlphaBetaScoreMatchesNegamax(t *testing.T) {
	cases := []struct {
		fen      string
		maxDepth int
	}{
		{orderingFENs[0], 3}, // start position
		{orderingFENs[2], 1}, // perft position 4: many captures
		{orderingFENs[3], 2}, // perft position 5
		{orderingFENs[4], 1}, // middlegame
		{orderingFENs[5], 3}, // en passant
		{orderingFENs[6], 3}, // promotions with captures
		// Kiwipete (orderingFENs[1]) is left out: even depth 1 takes about 2 seconds.
	}
	for _, c := range cases {
		fen := c.fen
		for depth := 1; depth <= c.maxDepth; depth++ {
			b := mustBoard(t, fen)
			want := referenceScore(&searcher{board: b}, depth)
			got := testSearcher(b).alphaBeta(-infinite, infinite, depth)
			if got != want {
				t.Errorf("%s, depth %d: alphaBeta = %d, full negamax = %d", fen, depth, got, want)
			}
		}
	}
}

// BenchmarkAlphaBeta reports time and positions per search, the numbers move
// ordering is meant to bring down. The depths are chosen so a run without move
// ordering still finishes in about 20 seconds (unordered quiescence is very slow
// in the busy positions).
func BenchmarkAlphaBeta(b *testing.B) {
	positions := []struct {
		name               string
		fen                string
		minDepth, maxDepth int
	}{
		{"start", orderingFENs[0], 3, 5},
		{"middlegame", orderingFENs[4], 1, 3},
		{"kiwipete", orderingFENs[1], 1, 2},
	}
	for _, pos := range positions {
		for depth := pos.minDepth; depth <= pos.maxDepth; depth++ {
			b.Run(fmt.Sprintf("%s/depth%d", pos.name, depth), func(b *testing.B) {
				board, err := chess.CreateBoard(pos.fen)
				if err != nil {
					b.Fatal(err)
				}
				// A fresh table for every search, as BestMove creates one per move;
				// allocating it is not part of the measured time.
				nodes := 0
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					s := testSearcher(board)
					b.StartTimer()
					s.alphaBeta(-infinite, infinite, depth)
					nodes = s.nodes
				}
				b.ReportMetric(float64(nodes), "positions/op")
			})
		}
	}
}
