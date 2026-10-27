package chess

import (
	"math/rand/v2"
	"strings"
	"sync"
	"testing"
	"unicode"
)

// These tests check properties every evaluation must have, not specific scores,
// so they stay valid as Evaluate grows. Evaluate returns the score from the point
// of view of the side to move.

// mirrorFEN flips the board vertically and swaps the colors: White's pieces become
// Black's on the mirrored squares, and the other side is to move. It is the same
// position seen from the other side, so the side to move must get the same score.
func mirrorFEN(t *testing.T, fen string) string {
	t.Helper()
	parts := strings.Fields(fen)
	if len(parts) != 6 {
		t.Fatalf("mirrorFEN needs a full FEN, got %q", fen)
	}

	swapCase := func(s string) string {
		return strings.Map(func(r rune) rune {
			if unicode.IsUpper(r) {
				return unicode.ToLower(r)
			}
			return unicode.ToUpper(r)
		}, s)
	}

	ranks := strings.Split(parts[0], "/")
	for i, j := 0, len(ranks)-1; i < j; i, j = i+1, j-1 {
		ranks[i], ranks[j] = ranks[j], ranks[i]
	}
	placement := swapCase(strings.Join(ranks, "/"))

	side := "w"
	if parts[1] == "w" {
		side = "b"
	}

	castling := "-"
	if parts[2] != "-" {
		swapped := swapCase(parts[2])
		castling = ""
		for _, r := range "KQkq" {
			if strings.ContainsRune(swapped, r) {
				castling += string(r)
			}
		}
	}

	ep := parts[3]
	if ep != "-" {
		rank := map[byte]byte{'3': '6', '6': '3'}[ep[1]]
		ep = string([]byte{ep[0], rank})
	}

	return strings.Join([]string{placement, side, castling, ep, parts[4], parts[5]}, " ")
}

// evalPositions returns a varied set of positions: the standard test positions,
// plus positions reached by random play from them (fixed seed, so it is repeatable).
func evalPositions(t *testing.T) []string {
	t.Helper()
	fens := []string{
		startFEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		"4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2",
		"8/4P3/8/8/8/8/8/k3K3 w - - 0 1",
	}
	rng := rand.New(rand.NewPCG(1, 2))
	var out []string
	for _, fen := range fens {
		out = append(out, fen)
		b := mustBoard(t, fen)
		for ply := 0; ply < 40; ply++ {
			moves := b.GenerateLegalMoves(nil)
			if len(moves) == 0 {
				break
			}
			b.MakeMove(moves[rng.IntN(len(moves))])
			out = append(out, b.FEN())
		}
	}
	return out
}

func TestMirrorFEN(t *testing.T) {
	// Sanity check of the test helper itself.
	got := mirrorFEN(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b Kq e3 0 1")
	want := "rnbqkbnr/pppp1ppp/8/4p3/8/8/PPPPPPPP/RNBQKBNR w Qk e6 0 1"
	if got != want {
		t.Errorf("mirrorFEN =\n  %q\nwant\n  %q", got, want)
	}
}

// The same position seen from the other side must get the same score for the side to move.
func TestEvaluateColorSymmetry(t *testing.T) {
	for _, fen := range evalPositions(t) {
		mirrored := mirrorFEN(t, fen)
		a, b := mustBoard(t, fen), mustBoard(t, mirrored)
		if ea, eb := a.Evaluate(), b.Evaluate(); ea != eb {
			t.Errorf("Evaluate differs between a position and its mirror image:\n  %s  → %d\n  %s  → %d",
				fen, ea, mirrored, eb)
		}
	}
}

// A side that is far ahead must see a positive score, and its opponent a negative one.
func TestEvaluateSideToMove(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want int // +1: the side to move is far ahead, -1: far behind
	}{
		{"white far ahead, white to move", "4k3/8/8/8/8/8/PPPPPPPP/RN1QK1NR w - - 0 1", +1},
		{"white far ahead, black to move", "4k3/8/8/8/8/8/PPPPPPPP/RN1QK1NR b - - 0 1", -1},
		{"black far ahead, black to move", "rn1qk1nr/pppppppp/8/8/8/8/8/4K3 b - - 0 1", +1},
		{"black far ahead, white to move", "rn1qk1nr/pppppppp/8/8/8/8/8/4K3 w - - 0 1", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustBoard(t, tt.fen).Evaluate()
			if (tt.want > 0 && got <= 0) || (tt.want < 0 && got >= 0) {
				sign := "positive"
				if tt.want < 0 {
					sign = "negative"
				}
				t.Errorf("Evaluate(%q) = %d, want %s", tt.fen, got, sign)
			}
		})
	}
}

// The king hides in the middlegame and comes to the centre in the endgame. The
// same two king squares must swap order once the pieces are off the board.
func TestEvaluateKingDependsOnPhase(t *testing.T) {
	tests := []struct {
		name          string
		better, worse string // the same position with the white king on two squares
	}{
		{
			"all pieces on: castled beats centre",
			"rnbq1rk1/pppppppp/8/8/8/8/PPPPPPPP/RNBQ1RK1 w - - 0 1",
			"rnbq1rk1/pppppppp/8/8/4K3/8/PPPPPPPP/RNBQ1R2 w - - 0 1",
		},
		{
			"only pawns left: centre beats castled",
			"6k1/pppp4/8/8/4K3/8/PPPP4/8 w - - 0 1",
			"6k1/pppp4/8/8/8/8/PPPP4/6K1 w - - 0 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			better, worse := mustBoard(t, tt.better).Evaluate(), mustBoard(t, tt.worse).Evaluate()
			if better <= worse {
				t.Errorf("%s scores %d, not above %s at %d", tt.better, better, tt.worse, worse)
			}
		})
	}
}

// Promotions can push the phase past its maximum. The king must then be scored
// with the middlegame table alone, not extrapolated beyond it.
func TestEvaluateKingPhaseIsCapped(t *testing.T) {
	const (
		g1 = 6
		e4 = 28
	)
	castled := mustBoard(t, "rnbq1rk1/pppppppp/8/8/8/QQQQQQQQ/PPPPPPPP/QQQQ1RK1 w - - 0 1")
	centre := mustBoard(t, "rnbq1rk1/pppppppp/8/8/4K3/QQQQQQQQ/PPPPPPPP/QQQQ1R2 w - - 0 1")
	got := castled.Evaluate() - centre.Evaluate()
	want := kingMidSquareTable[g1^56] - kingMidSquareTable[e4^56]
	if got != want {
		t.Errorf("moving the king from e4 to g1 with twelve queens on the board gains %d, want %d (middlegame table only)", got, want)
	}
}

// kingTableScore is what the king tables alone give a king on sq, from its own
// side's point of view.
func kingTableScore(sq int, white bool, phase int) int {
	if white {
		sq ^= 56
	}
	return (kingMidSquareTable[sq]*phase + kingEndSquareTable[sq]*(maxPhase-phase)) / maxPhase
}

// Endgames where one side is winning, and where nobody is, must score the same
// from either side, so the mop-up bonus goes to whoever is ahead, white or black.
func TestEvaluateMopUpColorSymmetry(t *testing.T) {
	for _, fen := range []string{
		"k7/8/8/8/8/2K5/8/7R w - - 0 1",      // rook against bare king
		"k7/8/8/8/8/2K5/8/7R b - - 0 1",      // the same, the loser to move
		"7k/8/8/3QK3/8/8/8/8 w - - 0 1",      // queen against bare king
		"4k3/pp6/8/8/8/8/PP6/4K3 w - - 0 1",  // equal pawn endgame: no bonus for anyone
		"4k3/ppp5/8/8/8/8/PP6/4K3 w - - 0 1", // one pawn up: not winning enough
	} {
		mirrored := mirrorFEN(t, fen)
		if a, b := mustBoard(t, fen).Evaluate(), mustBoard(t, mirrored).Evaluate(); a != b {
			t.Errorf("Evaluate differs between a position and its mirror image:\n  %s  → %d\n  %s  → %d", fen, a, mirrored, b)
		}
	}
}

// With a rook against a bare king, the winner wants the losing king in a corner,
// and that must count for more than the king tables alone say.
func TestEvaluateMopUpPushesLosingKingToEdge(t *testing.T) {
	const (
		a8 = 56
		e5 = 36
	)
	phase := phaseValues[Rook]
	corner := mustBoard(t, "k7/8/8/8/8/2K5/8/7R w - - 0 1")
	centre := mustBoard(t, "8/8/8/4k3/8/2K5/8/7R w - - 0 1")
	got := corner.Evaluate() - centre.Evaluate()
	tables := kingTableScore(e5, false, phase) - kingTableScore(a8, false, phase)
	if got <= tables {
		t.Errorf("the losing king in the corner instead of the centre gains white %d, want more than the king tables' %d", got, tables)
	}
}

// With a rook against a bare king, the winning king walking up to the losing
// king must count for more than the king tables alone say.
func TestEvaluateMopUpBringsKingsTogether(t *testing.T) {
	const (
		c6 = 42
		h1 = 7
	)
	phase := phaseValues[Rook]
	near := mustBoard(t, "k7/8/2K5/8/8/8/8/1R6 w - - 0 1")
	far := mustBoard(t, "k7/8/8/8/8/8/8/1R5K w - - 0 1")
	got := near.Evaluate() - far.Evaluate()
	tables := kingTableScore(c6, true, phase) - kingTableScore(h1, true, phase)
	if got <= tables {
		t.Errorf("the winning king next to the losing one instead of far away gains white %d, want more than the king tables' %d", got, tables)
	}
}

// A one-pawn lead is not winning enough for the mop-up bonus: in a pawn ending,
// where the bonus would count in full, moving the kings together changes only
// what the king tables say.
func TestEvaluateNoMopUpForSmallLead(t *testing.T) {
	const (
		e1 = 4
		e2 = 12
	)
	// Black is a pawn up.
	a := mustBoard(t, "4k3/ppp5/8/8/8/8/PP2K3/8 w - - 0 1")
	b := mustBoard(t, "4k3/ppp5/8/8/8/8/PP6/4K3 w - - 0 1")
	got := a.Evaluate() - b.Evaluate()
	want := kingTableScore(e2, true, 0) - kingTableScore(e1, true, 0)
	if got != want {
		t.Errorf("Ke2 instead of Ke1 changes the score by %d, want %d from the king table alone", got, want)
	}
}

func TestEvaluateDoesNotChangeBoard(t *testing.T) {
	for _, fen := range evalPositions(t) {
		b := mustBoard(t, fen)
		before := *b
		b.Evaluate()
		if diff := boardDiff(b, &before); diff != "" {
			t.Errorf("%s: Evaluate changed the board:\n  %s", fen, diff)
		}
	}
}

// Bots evaluate positions in several goroutines at once (parallel games, a search
// split over cores). Run with -race to check this.
func TestEvaluateConcurrent(t *testing.T) {
	positions := evalPositions(t)
	want := make([]int, len(positions))
	for i, fen := range positions {
		want[i] = mustBoard(t, fen).Evaluate()
	}

	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i, fen := range positions {
				b, err := CreateBoard(fen)
				if err != nil {
					t.Errorf("CreateBoard(%q): %v", fen, err)
					return
				}
				if got := b.Evaluate(); got != want[i] {
					t.Errorf("%s: concurrent Evaluate = %d, want %d", fen, got, want[i])
				}
			}
		}()
	}
	wg.Wait()
}
