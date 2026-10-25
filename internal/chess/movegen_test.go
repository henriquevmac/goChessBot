package chess

import (
	"slices"
	"testing"
)

const startFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func mustBoard(t *testing.T, fen string) *Board {
	t.Helper()
	b, err := CreateBoard(fen)
	if err != nil {
		t.Fatalf("CreateBoard(%q): %v", fen, err)
	}
	return b
}

func uciList(moves []Move) []string {
	out := make([]string, len(moves))
	for i, m := range moves {
		out[i] = m.String()
	}
	slices.Sort(out)
	return out
}

func findMove(moves []Move, uci string) (Move, bool) {
	for _, m := range moves {
		if m.String() == uci {
			return m, true
		}
	}
	return 0, false
}

// checkMoveInvariants verifies properties every generated move must have,
// whatever the position: right side moving, flags matching what is on the target square.
func checkMoveInvariants(t *testing.T, b *Board, moves []Move) {
	t.Helper()
	seen := map[Move]bool{}
	for _, m := range moves {
		if seen[m] {
			t.Errorf("duplicate move %s", m)
		}
		seen[m] = true

		piece := b.Square[m.From()]
		target := b.Square[m.To()]
		if piece.IsEmpty() || piece.Color() != b.ToMove {
			t.Errorf("%s: moves from a square without a piece of the side to move (piece %d)", m, piece)
			continue
		}
		if m.From() == m.To() {
			t.Errorf("%s: from and to are the same square", m)
		}

		switch flag := m.Flag(); {
		case flag == EnPassant:
			if piece.Type() != Pawn || m.To() != b.EnPassantSquare || !target.IsEmpty() {
				t.Errorf("%s: invalid en passant", m)
			}
		case m.IsCapture():
			if !target.IsOpponent(piece) {
				t.Errorf("%s: capture flag but no enemy piece on the target square", m)
			}
		default:
			if !target.IsEmpty() {
				t.Errorf("%s: non-capture flag but the target square is occupied", m)
			}
		}

		if m.Flag() == DoublePawn && piece.Type() != Pawn {
			t.Errorf("%s: DoublePawn flag on a non-pawn", m)
		}
		if m.IsPromotion() {
			if piece.Type() != Pawn {
				t.Errorf("%s: promotion flag on a non-pawn", m)
			}
			if row := m.To() / 8; row != 0 && row != 7 {
				t.Errorf("%s: promotion not on the last rank", m)
			}
		} else if piece.Type() == Pawn {
			if row := m.To() / 8; row == 0 || row == 7 {
				t.Errorf("%s: pawn reaches the last rank without promoting", m)
			}
		}
	}
}

func TestGenerateMoves(t *testing.T) {
	tests := []struct {
		name  string
		fen   string
		want  []string            // every expected move in UCI, any order
		flags map[string]MoveFlag // optional: exact flags for some of the moves
	}{
		// Start positions
		{
			name: "start position, white",
			fen:  startFEN,
			want: []string{
				"a2a3", "a2a4", "b2b3", "b2b4", "c2c3", "c2c4", "d2d3", "d2d4",
				"e2e3", "e2e4", "f2f3", "f2f4", "g2g3", "g2g4", "h2h3", "h2h4",
				"b1a3", "b1c3", "g1f3", "g1h3",
			},
			flags: map[string]MoveFlag{"e2e3": Normal, "e2e4": DoublePawn, "g1f3": Normal},
		},
		{
			name: "start position, black",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
			want: []string{
				"a7a6", "a7a5", "b7b6", "b7b5", "c7c6", "c7c5", "d7d6", "d7d5",
				"e7e6", "e7e5", "f7f6", "f7f5", "g7g6", "g7g5", "h7h6", "h7h5",
				"b8a6", "b8c6", "g8f6", "g8h6",
			},
			flags: map[string]MoveFlag{"e7e5": DoublePawn},
		},

		// Knights
		{
			name: "knight in the center",
			fen:  "8/8/8/8/3N4/8/8/8 w - - 0 1",
			want: []string{"d4b3", "d4b5", "d4c2", "d4c6", "d4e2", "d4e6", "d4f3", "d4f5"},
		},
		{
			name: "knight in the corner",
			fen:  "8/8/8/8/8/8/8/N7 w - - 0 1",
			want: []string{"a1b3", "a1c2"},
		},
		{
			name:  "knight captures enemy, not own piece",
			fen:   "8/8/8/8/8/1p6/2P5/N7 w - - 0 1",
			want:  []string{"a1b3", "c2b3", "c2c3", "c2c4"},
			flags: map[string]MoveFlag{"a1b3": Capture, "c2b3": Capture},
		},

		// Kings (no castling yet)
		{
			name: "king in the center",
			fen:  "8/8/8/8/3K4/8/8/8 w - - 0 1",
			want: []string{"d4c3", "d4c4", "d4c5", "d4d3", "d4d5", "d4e3", "d4e4", "d4e5"},
		},
		{
			name:  "king in the corner with own and enemy pieces",
			fen:   "8/8/8/8/8/8/Pp6/K7 w - - 0 1",
			want:  []string{"a1b1", "a1b2", "a2a3", "a2a4"},
			flags: map[string]MoveFlag{"a1b2": Capture},
		},

		// Sliding pieces on an empty board
		{
			name: "rook in the corner",
			fen:  "8/8/8/8/8/8/8/R7 w - - 0 1",
			want: []string{
				"a1a2", "a1a3", "a1a4", "a1a5", "a1a6", "a1a7", "a1a8",
				"a1b1", "a1c1", "a1d1", "a1e1", "a1f1", "a1g1", "a1h1",
			},
		},
		{
			name: "bishop in the center",
			fen:  "8/8/8/8/3B4/8/8/8 w - - 0 1",
			want: []string{
				"d4e5", "d4f6", "d4g7", "d4h8",
				"d4c5", "d4b6", "d4a7",
				"d4e3", "d4f2", "d4g1",
				"d4c3", "d4b2", "d4a1",
			},
		},
		{
			name: "bishop in the corner",
			fen:  "8/8/8/8/8/8/8/B7 w - - 0 1",
			want: []string{"a1b2", "a1c3", "a1d4", "a1e5", "a1f6", "a1g7", "a1h8"},
		},
		{
			name: "queen in the center",
			fen:  "8/8/8/8/3Q4/8/8/8 w - - 0 1",
			want: []string{
				"d4d5", "d4d6", "d4d7", "d4d8",
				"d4d3", "d4d2", "d4d1",
				"d4e4", "d4f4", "d4g4", "d4h4",
				"d4c4", "d4b4", "d4a4",
				"d4e5", "d4f6", "d4g7", "d4h8",
				"d4c5", "d4b6", "d4a7",
				"d4e3", "d4f2", "d4g1",
				"d4c3", "d4b2", "d4a1",
			},
		},

		// Sliding pieces blocked
		{
			// Rook a4: up to the enemy pawn on a6 (captured, a7 not reached),
			// down to a3 (own pawn on a2 blocks), and right along the 4th rank.
			name: "rook stops at own piece and after capturing",
			fen:  "8/8/p7/8/R7/8/P7/8 w - - 0 1",
			want: []string{
				"a4a5", "a4a6", "a4a3",
				"a4b4", "a4c4", "a4d4", "a4e4", "a4f4", "a4g4", "a4h4",
				"a2a3",
			},
			flags: map[string]MoveFlag{"a4a6": Capture, "a4a5": Normal},
		},
		{
			// Bishop c1 blocked by own pawns on b2 and d2.
			name: "bishop fully blocked",
			fen:  "8/8/8/8/8/8/1P1P4/2B5 w - - 0 1",
			want: []string{"b2b3", "b2b4", "d2d3", "d2d4"},
		},

		// White pawns
		{
			name: "pawn blocked by enemy piece",
			fen:  "8/8/8/8/4p3/4P3/8/8 w - - 0 1",
			want: []string{},
		},
		{
			name: "double push blocked on the second square",
			fen:  "8/8/8/8/4p3/8/4P3/8 w - - 0 1",
			want: []string{"e2e3"},
		},
		{
			name: "double push blocked on the first square",
			fen:  "8/8/8/8/8/4n3/4P3/8 w - - 0 1",
			want: []string{},
		},
		{
			name: "no double push off the starting rank",
			fen:  "8/8/8/8/8/4P3/8/8 w - - 0 1",
			want: []string{"e3e4"},
		},
		{
			name:  "pawn captures both ways",
			fen:   "8/8/8/3p1p2/4P3/8/8/8 w - - 0 1",
			want:  []string{"e4d5", "e4e5", "e4f5"},
			flags: map[string]MoveFlag{"e4d5": Capture, "e4f5": Capture, "e4e5": Normal},
		},
		{
			name: "pawn does not capture own piece",
			fen:  "8/8/8/3N4/4P3/8/8/8 w - - 0 1",
			want: []string{
				"e4e5",
				"d5b4", "d5b6", "d5c3", "d5c7", "d5e3", "d5e7", "d5f4", "d5f6",
			},
		},
		{
			name:  "promotion",
			fen:   "8/4P3/8/8/8/8/8/8 w - - 0 1",
			want:  []string{"e7e8n", "e7e8b", "e7e8r", "e7e8q"},
			flags: map[string]MoveFlag{"e7e8n": PromoKn, "e7e8b": PromoB, "e7e8r": PromoR, "e7e8q": PromoQ},
		},
		{
			name: "promotion with captures",
			fen:  "3r1r2/4P3/8/8/8/8/8/8 w - - 0 1",
			want: []string{
				"e7e8n", "e7e8b", "e7e8r", "e7e8q",
				"e7d8n", "e7d8b", "e7d8r", "e7d8q",
				"e7f8n", "e7f8b", "e7f8r", "e7f8q",
			},
			flags: map[string]MoveFlag{
				"e7e8q": PromoQ, "e7d8q": PromoQCap, "e7f8n": PromoKnCap,
				"e7d8b": PromoBCap, "e7f8r": PromoRCap,
			},
		},
		{
			name:  "promotion only by capture when blocked",
			fen:   "3rr3/4P3/8/8/8/8/8/8 w - - 0 1",
			want:  []string{"e7d8n", "e7d8b", "e7d8r", "e7d8q"},
			flags: map[string]MoveFlag{"e7d8q": PromoQCap},
		},
		{
			name:  "en passant",
			fen:   "8/8/8/3pP3/8/8/8/8 w - d6 0 1",
			want:  []string{"e5d6", "e5e6"},
			flags: map[string]MoveFlag{"e5d6": EnPassant, "e5e6": Normal},
		},
		{
			name: "no en passant without the en passant square",
			fen:  "8/8/8/3pP3/8/8/8/8 w - - 0 1",
			want: []string{"e5e6"},
		},
		{
			// h3 attacks only g4; wrapping around would add h3xa5.
			name: "edge pawn does not wrap around the board",
			fen:  "8/8/8/p7/8/7P/8/8 w - - 0 1",
			want: []string{"h3h4"},
		},

		// Black pawns
		{
			name:  "black pawn pushes",
			fen:   "8/3p4/8/8/8/8/8/8 b - - 0 1",
			want:  []string{"d7d6", "d7d5"},
			flags: map[string]MoveFlag{"d7d5": DoublePawn},
		},
		{
			name:  "black pawn captures",
			fen:   "8/8/8/3p4/2P1P3/8/8/8 b - - 0 1",
			want:  []string{"d5d4", "d5c4", "d5e4"},
			flags: map[string]MoveFlag{"d5c4": Capture, "d5e4": Capture},
		},
		{
			name:  "black promotion",
			fen:   "8/8/8/8/8/8/3p4/8 b - - 0 1",
			want:  []string{"d2d1n", "d2d1b", "d2d1r", "d2d1q"},
			flags: map[string]MoveFlag{"d2d1q": PromoQ},
		},
		{
			name: "black promotion with capture",
			fen:  "8/8/8/8/8/8/3p4/2R1R3 b - - 0 1",
			want: []string{
				"d2d1n", "d2d1b", "d2d1r", "d2d1q",
				"d2c1n", "d2c1b", "d2c1r", "d2c1q",
				"d2e1n", "d2e1b", "d2e1r", "d2e1q",
			},
			flags: map[string]MoveFlag{"d2c1q": PromoQCap, "d2e1n": PromoKnCap},
		},
		{
			name:  "black en passant",
			fen:   "8/8/8/8/3Pp3/8/8/8 b - d3 0 1",
			want:  []string{"e4d3", "e4e3"},
			flags: map[string]MoveFlag{"e4d3": EnPassant},
		},
		{
			// a6 attacks only b5; wrapping around would add a6xh4.
			name: "black edge pawn does not wrap around the board",
			fen:  "8/8/p7/8/7P/8/8/8 b - - 0 1",
			want: []string{"a6a5"},
		},

		// Side to move
		{
			name: "only the side to move generates moves",
			fen:  "8/8/8/8/3n4/8/8/R7 b - - 0 1",
			want: []string{"d4b3", "d4b5", "d4c2", "d4c6", "d4e2", "d4e6", "d4f3", "d4f5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			moves := b.GenerateMoves(nil, false)

			want := slices.Clone(tt.want)
			slices.Sort(want)
			if got := uciList(moves); !slices.Equal(got, want) {
				t.Errorf("%s\n got  %d moves: %v\n want %d moves: %v", tt.fen, len(got), got, len(want), want)
			}

			for uci, flag := range tt.flags {
				m, ok := findMove(moves, uci)
				if !ok {
					continue // already reported as a missing move above
				}
				if m.Flag() != flag {
					t.Errorf("%s: flag = %d, want %d", uci, m.Flag(), flag)
				}
			}

			checkMoveInvariants(t, b, moves)
		})
	}
}

func TestGenerateMovesAppends(t *testing.T) {
	b := mustBoard(t, startFEN)
	prefix := CreateMove(0, 1, Normal)

	moves := b.GenerateMoves([]Move{prefix}, false)
	if len(moves) != 21 {
		t.Fatalf("got %d moves, want 21 (1 existing + 20 generated)", len(moves))
	}
	if moves[0] != prefix {
		t.Errorf("existing move was overwritten: moves[0] = %s, want %s", moves[0], prefix)
	}
}

func TestGenerateMovesReusesBuffer(t *testing.T) {
	b := mustBoard(t, startFEN)
	buf := make([]Move, 0, 256)

	first := b.GenerateMoves(buf[:0], false)
	firstList := uciList(first)
	second := b.GenerateMoves(buf[:0], false)

	if !slices.Equal(firstList, uciList(second)) {
		t.Errorf("second call with the same buffer gave different moves")
	}
	if cap(second) != 256 || &second[0] != &buf[:1][0] {
		t.Errorf("GenerateMoves allocated a new slice instead of using the buffer's capacity")
	}
}
