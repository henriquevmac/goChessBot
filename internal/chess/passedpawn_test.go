package chess

import (
	"math/rand/v2"
	"testing"
)

// pawnCols records, for every col, the row of White's lowest pawn (10 when the
// col has none) and of Black's highest pawn (-1 when it has none): the arrays
// passedPawn takes.
func pawnCols(b *Board) (blackPawns, whitePawns [8]int) {
	for col := range 8 {
		blackPawns[col], whitePawns[col] = -1, 10
	}
	for idx, piece := range b.Square {
		if piece.Type() != Pawn {
			continue
		}
		row, col := ConvertIndexToRowCol(idx)
		if piece.Color() == White {
			whitePawns[col] = min(whitePawns[col], row)
		} else {
			blackPawns[col] = max(blackPawns[col], row)
		}
	}
	return blackPawns, whitePawns
}

// passedByScan answers the same question the slow way: it looks at every square
// in front of the pawn, in its own col and the cols next to it, for an enemy pawn.
func passedByScan(b *Board, row, col int, color PieceId) bool {
	step, enemy := 1, Black|Pawn
	if color == Black {
		step, enemy = -1, White|Pawn
	}
	for r := row + step; r >= 0 && r < 8; r += step {
		for c := max(0, col-1); c <= min(7, col+1); c++ {
			if b.Square[r*8+c] == enemy {
				return false
			}
		}
	}
	return true
}

func TestPassedPawn(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fen      string
		row, col int
		color    PieceId
		want     bool
	}{
		{"no enemy pawns", "4k3/8/8/3P4/8/8/8/4K3 w - - 0 1", 4, 3, White, true},
		{"enemy pawn in front, same col", "4k3/3p4/8/3P4/8/8/8/4K3 w - - 0 1", 4, 3, White, false},
		{"enemy pawn in front, col to the left", "4k3/2p5/8/3P4/8/8/8/4K3 w - - 0 1", 4, 3, White, false},
		{"enemy pawn in front, col to the right", "4k3/4p3/8/3P4/8/8/8/4K3 w - - 0 1", 4, 3, White, false},
		{"enemy pawn attacking it now", "4k3/8/4p3/3P4/8/8/8/4K3 w - - 0 1", 4, 3, White, false},
		{"enemy pawn two cols away", "4k3/5p2/8/3P4/8/8/8/4K3 w - - 0 1", 4, 3, White, true},
		{"enemy pawn on the same row", "4k3/8/8/3Pp3/8/8/8/4K3 w - - 0 1", 4, 3, White, true},
		{"enemy pawn behind", "4k3/8/8/3P4/4p3/8/8/4K3 w - - 0 1", 4, 3, White, true},
		{"own pawn in front", "4k3/8/3P4/3P4/8/8/8/4K3 w - - 0 1", 4, 3, White, true},
		{"col 0, enemy pawn in col 1", "4k3/1p6/8/P7/8/8/8/4K3 w - - 0 1", 4, 0, White, false},
		{"col 0, enemy pawn in col 7", "4k3/7p/8/P7/8/8/8/4K3 w - - 0 1", 4, 0, White, true},
		{"col 7, enemy pawn in col 6", "4k3/6p1/8/7P/8/8/8/4K3 w - - 0 1", 4, 7, White, false},
		{"col 7, enemy pawn in col 0", "4k3/p7/8/7P/8/8/8/4K3 w - - 0 1", 4, 7, White, true},
		{"black: no enemy pawns", "4k3/8/8/8/3p4/8/8/4K3 w - - 0 1", 3, 3, Black, true},
		{"black: enemy pawn in front, col to the right", "4k3/8/8/8/3p4/8/4P3/4K3 w - - 0 1", 3, 3, Black, false},
		{"black: enemy pawn on the same row", "4k3/8/8/8/3pP3/8/8/4K3 w - - 0 1", 3, 3, Black, true},
		{"black: enemy pawn behind", "4k3/8/8/4P3/3p4/8/8/4K3 w - - 0 1", 3, 3, Black, true},
		{"black: col 0, enemy pawn in col 1", "4k3/8/8/8/p7/8/1P6/4K3 w - - 0 1", 3, 0, Black, false},
		{"black: col 7, enemy pawn in col 6", "4k3/8/8/8/7p/8/6P1/4K3 w - - 0 1", 3, 7, Black, false},
	} {
		b := mustBoard(t, tc.fen)
		if b.Square[tc.row*8+tc.col] != tc.color|Pawn {
			t.Fatalf("%s: no pawn of that color at row %d, col %d in %s", tc.name, tc.row, tc.col, tc.fen)
		}
		blackPawns, whitePawns := pawnCols(b)
		if got := passedPawn(tc.row, tc.col, blackPawns, whitePawns, tc.color); got != tc.want {
			t.Errorf("%s: passedPawn(row %d, col %d) = %v, want %v\n  %s", tc.name, tc.row, tc.col, got, tc.want, tc.fen)
		}
	}
}

// On random pawn structures, passedPawn must agree with a square-by-square scan
// for every pawn of both colors.
func TestPassedPawnMatchesScan(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 0))
	for range 2000 {
		var b Board
		for range r.IntN(16) {
			sq := 8 + r.IntN(48) // pawns only on rows 1 to 6
			if r.IntN(2) == 0 {
				b.Square[sq] = White | Pawn
			} else {
				b.Square[sq] = Black | Pawn
			}
		}
		blackPawns, whitePawns := pawnCols(&b)
		for sq, piece := range b.Square {
			if piece.Type() != Pawn {
				continue
			}
			row, col := ConvertIndexToRowCol(sq)
			color := piece.Color()
			if got, want := passedPawn(row, col, blackPawns, whitePawns, color), passedByScan(&b, row, col, color); got != want {
				t.Fatalf("pawn at row %d, col %d (white: %v): passedPawn = %v, the scan says %v\n  board: %v", row, col, color == White, got, want, b.Square)
			}
		}
	}
}

// Positions with passed pawns for both sides must score the same from either side.
func TestEvaluatePassedPawnColorSymmetry(t *testing.T) {
	for _, fen := range []string{
		"4k3/8/8/3P4/8/8/8/4K3 w - - 0 1",             // one white passed pawn
		"4k3/1P6/8/8/8/8/6p1/4K3 b - - 0 1",           // passed pawns about to promote, both sides
		"4k3/pp3p2/8/2P5/8/8/5PPP/4K3 w - - 0 1",      // passed and blocked pawns mixed
		"r3k3/p4ppp/8/3P4/1p6/8/P4PPP/R3K3 w - - 0 1", // with rooks: the bonus counts in every phase
	} {
		mirrored := mirrorFEN(t, fen)
		if a, b := mustBoard(t, fen).Evaluate(), mustBoard(t, mirrored).Evaluate(); a != b {
			t.Errorf("Evaluate differs between a position and its mirror image:\n  %s  → %d\n  %s  → %d", fen, a, mirrored, b)
		}
	}
}

// passedPawnScore is the passed pawn bonus for a pawn that has advanced the given
// number of rows, blended by phase.
func passedPawnScore(advance, phase int) int {
	return (midPassedPawnVal[advance]*phase + endPassedPawnVal[advance]*(maxPhase-phase)) / maxPhase
}

// Moving the black pawn from e7, where it stops the white pawn on d5, to h7,
// where it does not, gives White the passed pawn bonus for row 4. The h7 pawn
// is passed itself, and the pawn table scores the two squares differently.
// Without pieces only the endgame values count; with a rook each (phase 4) the
// bonus is blended.
func TestEvaluatePassedPawnBonus(t *testing.T) {
	const (
		e7 = 52
		h7 = 55
	)
	for _, tc := range []struct {
		name            string
		blocked, passed string
		phase           int
	}{
		{"pawn ending", "4k3/4p3/8/3P4/8/8/8/4K3 w - - 0 1", "4k3/7p/8/3P4/8/8/8/4K3 w - - 0 1", 0},
		{"a rook each", "r3k3/4p3/8/3P4/8/8/8/R3K3 w - - 0 1", "r3k3/7p/8/3P4/8/8/8/R3K3 w - - 0 1", 2 * phaseValues[Rook]},
	} {
		got := mustBoard(t, tc.passed).Evaluate() - mustBoard(t, tc.blocked).Evaluate()
		want := passedPawnScore(4, tc.phase) - passedPawnScore(1, tc.phase) - pieceSquareTables[Pawn][h7] + pieceSquareTables[Pawn][e7]
		if got != want {
			t.Errorf("%s: moving the blocking pawn from e7 to h7 changes the score by %d, want %d", tc.name, got, want)
		}
	}
}

// A passed pawn is worth more in the endgame than with all the pieces on, and
// more the further it has advanced.
func TestPassedPawnTables(t *testing.T) {
	for advance := 1; advance <= 6; advance++ {
		if midPassedPawnVal[advance] > endPassedPawnVal[advance] {
			t.Errorf("advance %d: middlegame bonus %d is above the endgame bonus %d", advance, midPassedPawnVal[advance], endPassedPawnVal[advance])
		}
		if advance > 1 && (midPassedPawnVal[advance] < midPassedPawnVal[advance-1] || endPassedPawnVal[advance] < endPassedPawnVal[advance-1]) {
			t.Errorf("advance %d: the bonus is smaller than for advance %d", advance, advance-1)
		}
	}
}
