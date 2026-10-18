package chess

import (
	"slices"
	"testing"
)

// castlingMoves returns the UCI strings of all moves flagged as castling, sorted.
func castlingMoves(moves []Move) []string {
	out := []string{}
	for _, m := range moves {
		if f := m.Flag(); f == KingCastle || f == QueenCastle {
			out = append(out, m.String())
		}
	}
	slices.Sort(out)
	return out
}

func TestCastling(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want []string // castling moves only, any order
	}{
		// Rights
		{"white, both sides", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"e1g1", "e1c1"}},
		{"black, both sides", "r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1", []string{"e8g8", "e8c8"}},
		{"white, kingside right only", "r3k2r/8/8/8/8/8/8/R3K2R w K - 0 1", []string{"e1g1"}},
		{"white, queenside right only", "r3k2r/8/8/8/8/8/8/R3K2R w Q - 0 1", []string{"e1c1"}},
		{"black, kingside right only", "r3k2r/8/8/8/8/8/8/R3K2R b k - 0 1", []string{"e8g8"}},
		{"black, queenside right only", "r3k2r/8/8/8/8/8/8/R3K2R b q - 0 1", []string{"e8c8"}},
		{"no rights", "r3k2r/8/8/8/8/8/8/R3K2R w - - 0 1", []string{}},
		{"black's rights do not let white castle", "r3k2r/8/8/8/8/8/8/R3K2R w kq - 0 1", []string{}},
		{"white's rights do not let black castle", "r3k2r/8/8/8/8/8/8/R3K2R b KQ - 0 1", []string{}},
		{"start position: blocked", startFEN, []string{}},

		// Squares between king and rook must be empty
		{"kingside blocked on g1 by own piece", "4k3/8/8/8/8/8/8/R3K1NR w KQ - 0 1", []string{"e1c1"}},
		{"kingside blocked on f1 by enemy piece", "4k3/8/8/8/8/8/8/R3Kb1R w KQ - 0 1", []string{"e1c1"}},
		{"queenside blocked on b1", "4k3/8/8/8/8/8/8/RN2K2R w KQ - 0 1", []string{"e1g1"}},
		{"queenside blocked on c1", "4k3/8/8/8/8/8/8/R1B1K2R w KQ - 0 1", []string{"e1g1"}},
		{"queenside blocked on d1", "4k3/8/8/8/8/8/8/R2QK2R w KQ - 0 1", []string{"e1g1"}},

		// King must not be in check, pass through or land on an attacked square
		{"in check: no castling", "4r1k1/8/8/8/8/8/8/R3K2R w KQ - 0 1", []string{}},
		{"f1 attacked: no kingside", "5rk1/8/8/8/8/8/8/R3K2R w KQ - 0 1", []string{"e1c1"}},
		{"g1 attacked: no kingside", "6rk/8/8/8/8/8/8/R3K2R w KQ - 0 1", []string{"e1c1"}},
		{"d1 attacked: no queenside", "3rk3/8/8/8/8/8/8/R3K2R w KQ - 0 1", []string{"e1g1"}},
		{"c1 attacked: no queenside", "2r1k3/8/8/8/8/8/8/R3K2R w KQ - 0 1", []string{"e1g1"}},
		{"b1 attacked is fine (king does not cross it)", "1r2k3/8/8/8/8/8/8/R3K2R w KQ - 0 1", []string{"e1g1", "e1c1"}},
		{"attacked rook is fine", "4k2r/8/8/8/8/8/8/R3K2R w KQ - 0 1", []string{"e1g1", "e1c1"}},
		{"pawn attacks d1 and f1", "4k3/8/8/8/8/8/4p3/R3K2R w KQ - 0 1", []string{}},
		{"knight attacks d1 and f1", "4k3/8/8/8/8/4n3/8/R3K2R w KQ - 0 1", []string{}},
		{"knight attacks only g1", "4k3/8/8/8/8/7n/8/R3K2R w KQ - 0 1", []string{"e1c1"}},
		{"own piece attacking does not matter", "4k3/8/8/8/8/8/8/R3K2R w KQ - 0 1", []string{"e1g1", "e1c1"}},

		// Black
		{"black: f8 attacked", "r3k2r/8/8/8/8/8/8/4KR2 b kq - 0 1", []string{"e8c8"}},
		{"black: d8 attacked", "r3k2r/8/8/8/8/8/8/3RK3 b kq - 0 1", []string{"e8g8"}},
		{"black: in check from a bishop", "r3k2r/8/8/1B6/8/8/8/4K3 b kq - 0 1", []string{}},
		{"black: b8 attacked is fine", "r3k2r/8/8/8/8/8/8/1R2K3 b kq - 0 1", []string{"e8g8", "e8c8"}},
		{"black: kingside blocked", "r3kb1r/8/8/8/8/8/8/4K3 b kq - 0 1", []string{"e8c8"}},
		{"black: white pawn attacks f8", "r3k2r/6P1/8/8/8/8/8/4K3 b kq - 0 1", []string{"e8c8"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			moves := b.GenerateMoves(nil)

			want := slices.Clone(tt.want)
			slices.Sort(want)
			if got := castlingMoves(moves); !slices.Equal(got, want) {
				t.Errorf("%s\n got  castling moves %v\n want castling moves %v", tt.fen, got, want)
			}

			for _, m := range moves {
				piece := b.Square[m.From()]
				twoSquares := abs(m.To()%8-m.From()%8) == 2

				// Castling flags must match the direction the king moves.
				switch m.Flag() {
				case KingCastle:
					if piece.Type() != King || m.To() != m.From()+2 {
						t.Errorf("%s: KingCastle flag, but not a king moving two squares east", m)
					}
				case QueenCastle:
					if piece.Type() != King || m.To() != m.From()-2 {
						t.Errorf("%s: QueenCastle flag, but not a king moving two squares west", m)
					}
				default:
					// A king moving two files is only ever castling.
					if piece.Type() == King && twoSquares {
						t.Errorf("%s: king moves two squares without a castling flag", m)
					}
				}
			}

			checkMoveInvariants(t, b, moves)
		})
	}
}
