package bot

import (
	"testing"

	"goChessBot/internal/chess"
)

// A side that is checkmated when quiescence reaches it scores as mated at the
// ply it is reached, not with its material.
func TestQuiesceSeesCheckmate(t *testing.T) {
	// The rook on a8 mates the king boxed in by its own pawns.
	b := mustBoard(t, "R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1")
	s := testSearcher(b)
	s.ply = 5
	if got, want := s.quiesce(-infinite, infinite), matedScore(5); got != want {
		t.Errorf("quiesce = %d, want %d (mated at ply 5)", got, want)
	}
}

// A capture that mates is found inside quiescence, and the mate is scored one
// ply deeper than the node that captures. The ply is restored afterwards.
func TestQuiesceMateAfterCaptureCountsPly(t *testing.T) {
	// Rxa8 takes the rook and mates on the back rank.
	b := mustBoard(t, "r5k1/5ppp/8/8/8/8/8/R5K1 w - - 0 1")
	s := testSearcher(b)
	s.ply = 5
	if got, want := s.quiesce(-infinite, infinite), -matedScore(6); got != want {
		t.Errorf("quiesce = %d, want %d (mate at ply 6)", got, want)
	}
	if s.ply != 5 {
		t.Errorf("ply after quiesce = %d, want 5", s.ply)
	}
}

// In check there is no standing pat: the side to move has to get out of check
// even when that loses material.
func TestQuiesceDoesNotStandPatInCheck(t *testing.T) {
	// The knight on f3 checks the king and forks the queen on d2. Nothing can
	// take the knight, so the king steps away and the queen is lost.
	b := mustBoard(t, "7k/8/8/8/8/5n2/3Q4/R5K1 w - - 0 1")
	s := testSearcher(b)
	standPat := b.Evaluate()
	// Losing the queen costs far more than any piece-square bonus can make up.
	if got := s.quiesce(-infinite, infinite); got > standPat-chess.PieceValues[chess.Queen]/2 {
		t.Errorf("quiesce = %d, static eval %d: the queen should be lost", got, standPat)
	}
}

// A side that is stalemated when quiescence reaches it scores a draw, not its
// material.
func TestQuiesceSeesStalemate(t *testing.T) {
	// The queen on c7 takes every square from the king on a8 without checking it.
	b := mustBoard(t, "k7/2Q5/8/8/8/8/8/7K b - - 0 1")
	s := testSearcher(b)
	if got := s.quiesce(-infinite, infinite); got != 0 {
		t.Errorf("quiesce = %d, want 0 (stalemate)", got)
	}
}

// A capture that stalemates is worth a draw, so a side that is ahead keeps its
// standing pat score instead.
func TestQuiesceCaptureThatStalematesIsADraw(t *testing.T) {
	// Qxc7, the only capture, leaves the king on a8 without a move.
	b := mustBoard(t, "k7/2n5/8/8/8/8/8/2Q4K w - - 0 1")
	s := testSearcher(b)
	standPat := b.Evaluate()
	if got := s.quiesce(-infinite, infinite); got != standPat {
		t.Errorf("quiesce = %d, want the standing pat score %d (Qxc7 stalemates)", got, standPat)
	}
}
