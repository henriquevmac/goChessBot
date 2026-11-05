package bot

import (
	"testing"

	"goChessBot/internal/chess"
)

// searchNodes searches b to the given depth with a fresh searcher and returns
// the score and the number of nodes visited.
func searchNodes(t *testing.T, fen string, depth int, options AlphaBetaOptions) (score, nodes int) {
	t.Helper()
	s := testSearcher(mustBoard(t, fen))
	s.options = options
	score = s.alphaBeta(-infinite, infinite, depth)
	return score, s.nodes
}

// Null-move pruning must leave everything as it found it: the board, the ply,
// the history and the null move flag.
func TestNullMoveRestoresSearchState(t *testing.T) {
	b := mustBoard(t, "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")
	before := *b
	s := testSearcher(b)
	s.history = []uint64{1, 2, 3}
	s.alphaBeta(-infinite, infinite, 5)
	if *b != before {
		t.Errorf("the board changed during the search: %s, want %s", b.FEN(), before.FEN())
	}
	if s.ply != 0 || len(s.history) != 3 || s.nullMove {
		t.Errorf("after the search ply = %d, history length = %d, nullMove = %v; want 0, 3, false", s.ply, len(s.history), s.nullMove)
	}
}

// In a middlegame the null move cuts off part of the tree, so the same depth
// takes fewer nodes than without it.
func TestNullMoveSearchesFewerNodes(t *testing.T) {
	fen := "r1bq1rk1/pp2bppp/2n1pn2/3p4/2PP4/2N1PN2/PP2BPPP/R2QKB1R w KQ - 0 9"
	_, with := searchNodes(t, fen, 5, AlphaBetaOptions{})
	_, without := searchNodes(t, fen, 5, AlphaBetaOptions{NoNullMove: true})
	if with >= without {
		t.Errorf("depth 5 took %d nodes with null-move pruning and %d without, want fewer with it", with, without)
	}
}

// With only king and pawns, zugzwang is common and passing would mislead the
// search, so no null move is tried: score and node count are exactly the same
// as with null-move pruning switched off.
func TestNoNullMoveInPawnEndings(t *testing.T) {
	// Trébuchet-like position: whoever is to move loses a pawn.
	fen := "8/8/8/2k5/2Pp4/3K4/8/8 w - - 0 1"
	scoreWith, with := searchNodes(t, fen, 7, AlphaBetaOptions{})
	scoreWithout, without := searchNodes(t, fen, 7, AlphaBetaOptions{NoNullMove: true})
	if scoreWith != scoreWithout || with != without {
		t.Errorf("pawn ending: score %d in %d nodes with null-move pruning, %d in %d without; want them equal", scoreWith, with, scoreWithout, without)
	}
}

// Null-move pruning must not hide a forced mate.
func TestNullMoveStillFindsMate(t *testing.T) {
	// Scholar's mate: Qxf7#.
	b := mustBoard(t, "r1bqkb1r/pppp1ppp/2n2n2/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4")
	got := NewAlphaBeta(AlphaBetaOptions{}).BestMove(*b, Limits{Depth: 5})
	if want := chess.CreateMove(39, 53, chess.Capture); got != want {
		t.Errorf("BestMove = %s, want %s (Qxf7#)", got, want)
	}
}
