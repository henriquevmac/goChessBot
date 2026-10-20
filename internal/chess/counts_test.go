package chess

import (
	"fmt"
	"testing"
)

// describeCounts lists only the piece entries that differ, using piece names.
// Counts[Empty] is not kept up to date, so it is ignored.
func describeCounts(got, want [32]uint8) string {
	names := map[PieceId]string{King: "K", Pawn: "P", Knight: "N", Bishop: "B", Rook: "R", Queen: "Q"}
	out := ""
	for p := range got {
		if PieceId(p) == Empty {
			continue
		}
		if got[p] != want[p] {
			id := PieceId(p)
			name := names[id.Type()]
			if !id.IsEmpty() {
				if id.Color() == White {
					name = "white " + name
				} else {
					name = "black " + name
				}
			}
			out += fmt.Sprintf(" %s: got %d, want %d;", name, got[p], want[p])
		}
	}
	return out
}

func TestCountsFromFEN(t *testing.T) {
	b := mustBoard(t, startFEN)
	want := map[PieceId]int{
		Empty:        32,
		White | Pawn: 8, Black | Pawn: 8,
		White | Knight: 2, Black | Knight: 2,
		White | Bishop: 2, Black | Bishop: 2,
		White | Rook: 2, Black | Rook: 2,
		White | Queen: 1, Black | Queen: 1,
		White | King: 1, Black | King: 1,
	}
	for p, n := range want {
		if int(b.Counts[p]) != n {
			t.Errorf("start position: Counts[%d] = %d, want %d", p, b.Counts[p], n)
		}
	}
}

// Counts are updated incrementally in MakeMove, like the hash: after every move,
// two moves deep, they must match a fresh count of the board, and unmaking must
// restore them exactly.
func TestCountsIncrementalUpdate(t *testing.T) {
	for _, fen := range makeUnmakeFENs {
		b := mustBoard(t, fen)
		original := b.Counts
		for _, m1 := range b.GenerateLegalMoves(nil) {
			u1 := b.MakeMove(m1)
			if diff := describeCounts(b.Counts, b.getCounts()); diff != "" {
				t.Errorf("%s: after %s:%s", fen, m1, diff)
			}
			afterFirst := b.Counts
			for _, m2 := range b.GenerateLegalMoves(nil) {
				u2 := b.MakeMove(m2)
				if diff := describeCounts(b.Counts, b.getCounts()); diff != "" {
					t.Errorf("%s: after %s %s:%s", fen, m1, m2, diff)
				}
				b.UnmakeMove(m2, u2)
				if b.Counts != afterFirst {
					t.Errorf("%s: unmaking %s after %s did not restore Counts", fen, m2, m1)
				}
			}
			b.UnmakeMove(m1, u1)
			if b.Counts != original {
				t.Errorf("%s: unmaking %s did not restore Counts", fen, m1)
			}
		}
	}
}

// Each kind of special move on its own, so a failure names the case.
func TestCountsSpecialMoves(t *testing.T) {
	tests := []struct {
		name  string
		fen   string
		moves []string
	}{
		{"quiet move", startFEN, []string{"g1f3"}},
		{"capture", "4k3/8/8/3p4/8/8/8/3RK3 w - - 0 1", []string{"d1d5"}},
		{"en passant", "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2", []string{"e5d6"}},
		{"castling", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"e1g1"}},
		{"promotion", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", []string{"e7e8q"}},
		{"underpromotion", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", []string{"e7e8n"}},
		{"capture promotion", "3r1r2/4P3/8/8/8/8/8/k3K3 w - - 0 1", []string{"e7d8q"}},
		{"black capture promotion", "4k3/8/8/8/8/8/3p4/2R1K3 b - - 0 1", []string{"d2c1n"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			for _, uci := range tt.moves {
				b.MakeMove(moveByUCI(t, b, uci))
				if diff := describeCounts(b.Counts, b.getCounts()); diff != "" {
					t.Errorf("after %s:%s", uci, diff)
				}
			}
		})
	}
}
