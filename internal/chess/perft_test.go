package chess

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

// Known perft results from the Chess Programming Wiki (https://www.chessprogramming.org/Perft_Results).
var perftPositions = []struct {
	name  string
	fen   string
	nodes []int // nodes[i] is the perft result at depth i+1
	quick int   // depths up to this one always run; deeper ones are skipped with -short
}{
	{
		name:  "start position",
		fen:   startFEN,
		nodes: []int{20, 400, 8902, 197281, 4865609},
		quick: 3,
	},
	{
		name:  "Kiwipete",
		fen:   "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		nodes: []int{48, 2039, 97862, 4085603},
		quick: 3,
	},
	{
		name:  "position 3",
		fen:   "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		nodes: []int{14, 191, 2812, 43238, 674624},
		quick: 4,
	},
	{
		name:  "position 4",
		fen:   "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		nodes: []int{6, 264, 9467, 422333},
		quick: 3,
	},
	{
		name:  "position 4, mirrored",
		fen:   "r2q1rk1/pP1p2pp/Q4n2/bbp1p3/Np6/1B3NBn/pPPP1PPP/R3K2R b KQ - 0 1",
		nodes: []int{6, 264, 9467, 422333},
		quick: 3,
	},
	{
		name:  "position 5",
		fen:   "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		nodes: []int{44, 1486, 62379, 2103487},
		quick: 3,
	},
	{
		name:  "position 6",
		fen:   "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		nodes: []int{46, 2079, 89890, 3894594},
		quick: 3,
	},
}

// divide lists the perft count below each legal move, sorted, in the same format
// Stockfish prints for "go perft N", so the two can be compared line by line.
func divide(b *Board, depth int) string {
	var lines []string
	for _, m := range b.GenerateLegalMoves(nil) {
		u := b.MakeMove(m)
		lines = append(lines, fmt.Sprintf("%s: %d", m, b.Perft(depth-1)))
		b.UnmakeMove(m, u)
	}
	slices.Sort(lines)
	return strings.Join(lines, "\n")
}

func TestPerft(t *testing.T) {
	for _, pos := range perftPositions {
		for i, want := range pos.nodes {
			depth := i + 1
			t.Run(fmt.Sprintf("%s/depth %d", pos.name, depth), func(t *testing.T) {
				if depth > pos.quick && testing.Short() {
					t.Skip("deep perft skipped in -short mode")
				}
				b := mustBoard(t, pos.fen)
				before := *b

				if got := b.Perft(depth); got != want {
					t.Errorf("Perft(%d) = %d, want %d (off by %+d)\nposition: %s\nper move (compare with Stockfish \"go perft %d\"):\n%s",
						depth, got, want, got-want, pos.fen, depth, divide(b, depth))
				}
				if diff := boardDiff(b, &before); diff != "" {
					t.Errorf("Perft changed the board:\n  %s", diff)
				}
			})
		}
	}
}

// Perft(1) is the number of legal moves, so the two functions must agree.
func TestPerftOneMatchesLegalMoves(t *testing.T) {
	for _, fen := range makeUnmakeFENs {
		b := mustBoard(t, fen)
		if got, want := b.Perft(1), len(b.GenerateLegalMoves(nil)); got != want {
			t.Errorf("%s: Perft(1) = %d, but there are %d legal moves", fen, got, want)
		}
	}
}

func TestPerftZero(t *testing.T) {
	b := mustBoard(t, startFEN)
	if got := b.Perft(0); got != 1 {
		t.Errorf("Perft(0) = %d, want 1", got)
	}
}
