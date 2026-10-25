package chess

import (
	"slices"
	"testing"
)

func TestGenerateLegalMoves(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want []string // every legal move in UCI, any order
	}{
		{
			// King on e1 in check from a rook on e2: it may capture the rook or step aside,
			// but not to d2 or f2, which the rook attacks.
			name: "king in check from an adjacent rook",
			fen:  "4k3/8/8/8/8/8/4r3/4K3 w - - 0 1",
			want: []string{"e1d1", "e1f1", "e1e2"},
		},
		{
			// Bishop e2 is pinned to the king by the rook on e7.
			name: "pinned piece cannot move",
			fen:  "4k3/4r3/8/8/8/8/4B3/4K3 w - - 0 1",
			want: []string{"e1d1", "e1d2", "e1f1", "e1f2"},
		},
		{
			// Rook h1 checks along the first rank. d1 is still attacked once the king
			// has left e1, and f1 is attacked directly. Only Rxh1 or stepping off the rank helps.
			name: "check along a rank: capture or step off the line",
			fen:  "4k3/8/8/8/8/8/7R/4K2r w - - 0 1",
			want: []string{"h2h1", "e1d2", "e1e2", "e1f2"},
		},
		{
			// bxc6 en passant would remove both pawns from the 5th rank and expose
			// the king on a5 to the rook on h5.
			name: "en passant that exposes the king is illegal",
			fen:  "8/8/8/KPp4r/8/8/8/7k w - c6 0 1",
			want: []string{"a5a4", "a5a6", "a5b6", "b5b6"},
		},
		{
			name: "back rank checkmate: no legal moves",
			fen:  "R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1",
			want: []string{},
		},
		{
			name: "stalemate: no legal moves",
			fen:  "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1",
			want: []string{},
		},
		{
			// The bishop on b5 checks along b5-c6-d7-e8. No black piece can block or
			// capture it, so only king moves remain; d7 is on the bishop's line and
			// castling out of check is not allowed.
			name: "black in check: only king moves, no castling",
			fen:  "r3k2r/8/8/1B6/8/8/8/4K3 b kq - 0 1",
			want: []string{"e8d8", "e8f8", "e8e7", "e8f7"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			before := *b
			moves := b.GenerateLegalMoves(nil)

			want := slices.Clone(tt.want)
			slices.Sort(want)
			if got := uciList(moves); !slices.Equal(got, want) {
				t.Errorf("%s\n got  %d moves: %v\n want %d moves: %v", tt.fen, len(got), got, len(want), want)
			}
			if diff := boardDiff(b, &before); diff != "" {
				t.Errorf("GenerateLegalMoves changed the board:\n  %s", diff)
			}
		})
	}
}

// Like GenerateMoves, GenerateLegalMoves must append and leave moves already in the
// slice untouched. The existing move here (e1d2, into check) would be dropped if the
// filter wrongly ran over the caller's moves too.
func TestGenerateLegalMovesAppends(t *testing.T) {
	b := mustBoard(t, "4k3/8/8/8/8/8/4r3/4K3 w - - 0 1")
	prefix := CreateMove(sq("e1"), sq("d2"), Normal)

	moves := b.GenerateLegalMoves([]Move{prefix})
	if len(moves) != 4 {
		t.Fatalf("got %d moves %v, want 4 (1 existing + 3 legal)", len(moves), uciList(moves))
	}
	if moves[0] != prefix {
		t.Errorf("existing move was changed: moves[0] = %s, want %s", moves[0], prefix)
	}
}

// Every legal move must be one of the pseudo-legal moves, and must leave the mover's king safe.
func TestLegalMovesAreSafePseudoLegalMoves(t *testing.T) {
	for _, fen := range makeUnmakeFENs {
		b := mustBoard(t, fen)
		pseudo := b.GenerateMoves(nil, false)
		for _, m := range b.GenerateLegalMoves(nil) {
			if !slices.Contains(pseudo, m) {
				t.Errorf("%s: legal move %s is not a pseudo-legal move", fen, m)
				continue
			}
			mover := b.ToMove
			u := b.MakeMove(m)
			for sq, p := range b.Square {
				if p == King|mover && b.IsSquareAttacked(sq, b.ToMove) {
					t.Errorf("%s: legal move %s leaves the king on %s attacked", fen, m, squareName(sq))
				}
			}
			b.UnmakeMove(m, u)
		}
	}
}

// GenerateCaptureMoves must return exactly the captures among the legal moves,
// in every position a couple of plies into the perft positions.
func TestCaptureMovesAreLegalCaptures(t *testing.T) {
	var walk func(t *testing.T, b *Board, depth int)
	walk = func(t *testing.T, b *Board, depth int) {
		var want []Move
		for _, m := range b.GenerateLegalMoves(nil) {
			if m.IsCapture() {
				want = append(want, m)
			}
		}
		got := b.GenerateCaptureMoves(nil)
		slices.Sort(want)
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Fatalf("%s: GenerateCaptureMoves = %v, want %v", b.FEN(), got, want)
		}
		if depth == 0 {
			return
		}
		for _, m := range b.GenerateLegalMoves(nil) {
			u := b.MakeMove(m)
			walk(t, b, depth-1)
			b.UnmakeMove(m, u)
		}
	}
	for _, pos := range perftPositions {
		t.Run(pos.name, func(t *testing.T) {
			b, err := CreateBoard(pos.fen)
			if err != nil {
				t.Fatal(err)
			}
			walk(t, b, 2)
		})
	}
}
