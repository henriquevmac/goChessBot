package chess

import "testing"

// sq converts a square name like "e4" to its index.
func sq(name string) int {
	return ConvertRowColToIdx(int(name[1]-'1'), int(name[0]-'a'))
}

func TestIsSquareAttacked(t *testing.T) {
	tests := []struct {
		name   string
		fen    string
		square string
		by     PieceId
		want   bool
	}{
		// Start position
		{"start: e3 by white pawns", startFEN, "e3", White, true},
		{"start: e3 not by black", startFEN, "e3", Black, false},
		{"start: e4 by nobody (white)", startFEN, "e4", White, false},
		{"start: e4 by nobody (black)", startFEN, "e4", Black, false},
		{"start: e6 by black pawns", startFEN, "e6", Black, true},
		{"start: f3 by knight and pawns", startFEN, "f3", White, true},
		{"start: own defended piece d1 counts", startFEN, "d1", White, true},

		// Knights
		{"knight a1 attacks b3", "8/8/8/8/8/8/8/N7 w - - 0 1", "b3", White, true},
		{"knight a1 attacks c2", "8/8/8/8/8/8/8/N7 w - - 0 1", "c2", White, true},
		{"knight a1 does not attack b2", "8/8/8/8/8/8/8/N7 w - - 0 1", "b2", White, false},
		{"knight h1 does not wrap to b3", "8/8/8/8/8/8/8/7N w - - 0 1", "b3", White, false},
		{"knight h1 does not wrap to a2", "8/8/8/8/8/8/8/7N w - - 0 1", "a2", White, false},
		{"knight attacks over pieces", "8/8/8/8/8/PPP5/PNP5/PPP5 w - - 0 1", "d3", White, true},

		// Kings
		{"king attacks neighbour", "8/8/8/8/8/8/8/4K3 w - - 0 1", "d2", White, true},
		{"king does not attack two away", "8/8/8/8/8/8/8/4K3 w - - 0 1", "e3", White, false},
		{"king h1 does not wrap to a2", "8/8/8/8/8/8/8/7K w - - 0 1", "a2", White, false},
		{"black king attacks", "8/8/8/8/3k4/8/8/8 w - - 0 1", "e5", Black, true},

		// Rooks
		{"rook attacks along the rank", "8/8/8/8/8/8/8/R7 w - - 0 1", "h1", White, true},
		{"rook attacks along the file", "8/8/8/8/8/8/8/R7 w - - 0 1", "a8", White, true},
		{"rook does not attack diagonally", "8/8/8/8/8/8/8/R7 w - - 0 1", "b2", White, false},
		{"rook attacks the blocking enemy piece", "8/8/8/8/8/8/8/R2p4 w - - 0 1", "d1", White, true},
		{"rook blocked by enemy piece", "8/8/8/8/8/8/8/R2p4 w - - 0 1", "e1", White, false},
		{"rook defends own piece", "8/8/8/8/8/P7/8/R7 w - - 0 1", "a3", White, true},
		{"rook blocked by own piece", "8/8/8/8/8/P7/8/R7 w - - 0 1", "a4", White, false},
		{"rook h1 does not wrap to a2", "8/8/8/8/8/8/8/7R w - - 0 1", "a2", White, false},
		{"rook does not x-ray through a bishop", "8/8/8/8/8/8/B7/R7 w - - 0 1", "a3", White, false},

		// Bishops
		{"bishop attacks diagonally", "8/8/8/8/8/8/8/2B5 w - - 0 1", "h6", White, true},
		{"bishop does not attack straight", "8/8/8/8/8/8/8/2B5 w - - 0 1", "c2", White, false},
		{"bishop attacks the blocking piece", "8/8/8/8/8/4p3/8/2B5 w - - 0 1", "e3", White, true},
		{"bishop blocked", "8/8/8/8/8/4p3/8/2B5 w - - 0 1", "f4", White, false},
		{"bishop h1 does not wrap", "8/8/8/8/8/8/8/7B w - - 0 1", "a3", White, false},
		{"bishop blocked by own pawn", "8/8/8/8/8/8/1P6/B7 w - - 0 1", "d4", White, false},

		// Queens
		{"queen attacks straight", "8/8/8/8/3Q4/8/8/8 w - - 0 1", "d8", White, true},
		{"queen attacks diagonally", "8/8/8/8/3Q4/8/8/8 w - - 0 1", "a7", White, true},
		{"queen does not attack like a knight", "8/8/8/8/3Q4/8/8/8 w - - 0 1", "e6", White, false},
		{"black queen attacks", "8/8/8/8/3q4/8/8/8 w - - 0 1", "h8", Black, true},

		// White pawns
		{"white pawn attacks d5", "8/8/8/8/4P3/8/8/8 w - - 0 1", "d5", White, true},
		{"white pawn attacks f5", "8/8/8/8/4P3/8/8/8 w - - 0 1", "f5", White, true},
		{"white pawn does not attack ahead", "8/8/8/8/4P3/8/8/8 w - - 0 1", "e5", White, false},
		{"white pawn does not attack backwards", "8/8/8/8/4P3/8/8/8 w - - 0 1", "d3", White, false},
		{"white pawn h2 does not wrap to a4", "8/8/8/8/8/8/7P/8 w - - 0 1", "a4", White, false},
		{"white pawn a2 does not wrap to h2", "8/8/8/8/8/8/P7/8 w - - 0 1", "h2", White, false},

		// Black pawns
		{"black pawn attacks d4", "8/8/8/4p3/8/8/8/8 w - - 0 1", "d4", Black, true},
		{"black pawn attacks f4", "8/8/8/4p3/8/8/8/8 w - - 0 1", "f4", Black, true},
		{"black pawn does not attack ahead", "8/8/8/4p3/8/8/8/8 w - - 0 1", "e4", Black, false},
		{"black pawn does not attack backwards", "8/8/8/4p3/8/8/8/8 w - - 0 1", "d6", Black, false},
		{"black pawn a7 does not wrap to h5", "8/p7/8/8/8/8/8/8 w - - 0 1", "h5", Black, false},

		// Colors
		{"white pieces do not attack for black", "8/8/8/8/3Q4/8/8/8 w - - 0 1", "d8", Black, false},
		{"black pawn attacks e4 for black only", "8/8/8/3p4/8/8/8/8 w - - 0 1", "e4", White, false},
		{"enemy king square is attacked", "k7/8/8/8/8/8/8/R7 w - - 0 1", "a8", White, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			if got := b.IsSquareAttacked(sq(tt.square), tt.by); got != tt.want {
				t.Errorf("%s: IsSquareAttacked(%s, %d) = %v, want %v", tt.fen, tt.square, tt.by, got, tt.want)
			}
		})
	}
}

// naiveAttacks is a deliberately simple, slow reference: it checks every piece of
// color `by` directly from chess geometry, without any of the precomputed tables.
func naiveAttacks(b *Board, target int, by PieceId) bool {
	tf, tr := target%8, target/8
	for from, piece := range b.Square {
		if piece.IsEmpty() || piece.Color() != by || from == target {
			continue
		}
		df, dr := tf-from%8, tr-from/8
		switch piece.Type() {
		case Pawn:
			forward := 1
			if by == Black {
				forward = -1
			}
			if dr == forward && abs(df) == 1 {
				return true
			}
		case Knight:
			if (abs(df) == 1 && abs(dr) == 2) || (abs(df) == 2 && abs(dr) == 1) {
				return true
			}
		case King:
			if max(abs(df), abs(dr)) == 1 {
				return true
			}
		case Rook, Bishop, Queen:
			straight := df == 0 || dr == 0
			diagonal := abs(df) == abs(dr)
			canStraight := piece.Type() != Bishop
			canDiagonal := piece.Type() != Rook
			if (straight && canStraight) || (diagonal && canDiagonal) {
				if pathClear(b, from, target) {
					return true
				}
			}
		}
	}
	return false
}

// pathClear reports whether every square strictly between from and to is empty.
func pathClear(b *Board, from, to int) bool {
	sf, sr := sign(to%8-from%8), sign(to/8-from/8)
	f, r := from%8+sf, from/8+sr
	for ConvertRowColToIdx(r, f) != to {
		if !b.Square[ConvertRowColToIdx(r, f)].IsEmpty() {
			return false
		}
		f += sf
		r += sr
	}
	return true
}

func sign(x int) int {
	switch {
	case x > 0:
		return 1
	case x < 0:
		return -1
	}
	return 0
}

// Every square, both colors, several busy positions: compare against the naive reference.
func TestIsSquareAttackedMatchesReference(t *testing.T) {
	fens := []string{
		startFEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", // Kiwipete
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		"7k/8/8/3q4/8/8/8/K7 w - - 0 1",
		"8/8/8/8/8/8/8/8 w - - 0 1",
	}
	for _, fen := range fens {
		b := mustBoard(t, fen)
		for target := 0; target < 64; target++ {
			for _, by := range []PieceId{White, Black} {
				want := naiveAttacks(b, target, by)
				if got := b.IsSquareAttacked(target, by); got != want {
					t.Errorf("%s: IsSquareAttacked(%s, %d) = %v, want %v",
						fen, squareName(target), by, got, want)
				}
			}
		}
	}
}
