package chess

import "testing"

func TestInCheck(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want bool
	}{
		{"start position", startFEN, false},
		{"white king attacked by an adjacent rook", "4k3/8/8/8/8/8/4r3/4K3 w - - 0 1", true},
		{"white king attacked along the rank", "4k3/8/8/8/8/8/8/4K2r w - - 0 1", true},
		{"black king attacked by a bishop", "r3k2r/8/8/1B6/8/8/8/4K3 b kq - 0 1", true},
		{"black king attacked by a knight", "6rk/5Npp/8/8/8/8/8/6K1 b - - 0 1", true},
		{"checkmated side is in check", "R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1", true},
		{"stalemated side is not in check", "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1", false},
		{"rook attack along the rank blocked by own bishop", "4k3/8/8/8/8/8/8/4KB1r w - - 0 1", false},
		{"rook attack along the file blocked by own pawn", "4k3/4r3/8/8/8/8/4P3/4K3 w - - 0 1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			before := *b
			if got := b.InCheck(); got != tt.want {
				t.Errorf("%s: InCheck() = %v, want %v", tt.fen, got, tt.want)
			}
			if diff := boardDiff(b, &before); diff != "" {
				t.Errorf("InCheck changed the board:\n  %s", diff)
			}
		})
	}
}

func TestResult(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want GameResult
	}{
		// Ongoing
		{"start position", startFEN, Ongoing},
		{"in check but can escape", "4k3/8/8/8/8/8/4r3/4K3 w - - 0 1", Ongoing},
		{"in check, only a block or capture helps", "4k3/8/8/8/8/8/7R/4K2r w - - 0 1", Ongoing},

		// Checkmate
		{"back rank mate", "R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1", Checkmate},
		{"fool's mate (white is mated)", "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3", Checkmate},
		{"smothered mate", "6rk/5Npp/8/8/8/8/8/6K1 b - - 0 1", Checkmate},
		{"queen and king mate", "k7/1Q6/2K5/8/8/8/8/8 b - - 0 1", Checkmate},

		// Stalemate
		{"stalemate with queen", "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1", Stalemate},
		{"stalemate with pawn", "k7/P7/K7/8/8/8/8/8 b - - 0 1", Stalemate},
		{"stalemate in the corner", "k7/8/1Q6/8/8/8/8/7K b - - 0 1", Stalemate},

		// Fifty-move rule
		{"halfmove clock 100 is a draw", "4k3/8/8/8/8/8/8/R3K3 w - - 100 80", FiftyMoveRule},
		{"halfmove clock 120 is a draw", "4k3/8/8/8/8/8/8/R3K3 w - - 120 80", FiftyMoveRule},
		{"halfmove clock 99 is not yet", "4k3/8/8/8/8/8/8/R3K3 w - - 99 80", Ongoing},
		{"checkmate takes precedence over the fifty-move rule", "R5k1/5ppp/8/8/8/8/8/6K1 b - - 100 80", Checkmate},
		{"stalemate is reported before the fifty-move rule", "7k/5Q2/6K1/8/8/8/8/8 b - - 100 80", Stalemate},

		// Insufficient material
		{"king vs king", "4k3/8/8/8/8/8/8/4K3 w - - 0 1", InsufficientMaterial},
		{"king and bishop vs king", "4k3/8/8/8/8/8/8/2B1K3 w - - 0 1", InsufficientMaterial},
		{"king and knight vs king", "4k3/8/8/8/8/8/8/1N2K3 w - - 0 1", InsufficientMaterial},
		{"king vs king and black knight", "4kn2/8/8/8/8/8/8/4K3 w - - 0 1", InsufficientMaterial},
		{"bishops on the same color (both dark)", "4kb2/8/8/8/8/8/8/2B1K3 w - - 0 1", InsufficientMaterial},
		{"two bishops on the same color", "4k3/8/8/8/8/4B3/8/2B1K3 w - - 0 1", InsufficientMaterial},
		{"bishops on opposite colors", "2b1k3/8/8/8/8/8/8/2B1K3 w - - 0 1", Ongoing},
		{"two knights vs king", "4k3/8/8/8/8/8/8/1NN1K3 w - - 0 1", Ongoing},
		{"knight vs knight", "1n2k3/8/8/8/8/8/8/1N2K3 w - - 0 1", Ongoing},
		{"bishop vs knight", "4kn2/8/8/8/8/8/8/2B1K3 w - - 0 1", Ongoing},
		{"bishop and knight vs king", "4k3/8/8/8/8/8/8/1NB1K3 w - - 0 1", Ongoing},
		{"same-color bishops plus a knight", "4k3/8/8/8/8/4B3/8/1NB1K3 w - - 0 1", Ongoing},
		{"three bishops on the same color", "4kb2/8/8/8/8/4B3/8/2B1K3 w - - 0 1", InsufficientMaterial},
		{"king and pawn vs king", "4k3/8/8/8/8/8/4P3/4K3 w - - 0 1", Ongoing},
		{"king and rook vs king", "4k3/8/8/8/8/8/8/R3K3 w - - 0 1", Ongoing},
		{"king and queen vs king", "4k3/8/8/8/8/8/8/3QK3 w - - 0 1", Ongoing},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			before := *b
			if got := b.Result(); got != tt.want {
				t.Errorf("%s: Result() = %v, want %v", tt.fen, got, tt.want)
			}
			if diff := boardDiff(b, &before); diff != "" {
				t.Errorf("Result changed the board:\n  %s", diff)
			}
		})
	}
}

// Playing fool's mate from the start: ongoing after each move until the last one.
func TestResultAfterMoves(t *testing.T) {
	b := mustBoard(t, startFEN)
	moves := []string{"f2f3", "e7e5", "g2g4", "d8h4"}
	for i, uci := range moves {
		if got := b.Result(); got != Ongoing {
			t.Fatalf("before move %d (%s): Result() = %v, want Ongoing", i+1, uci, got)
		}
		b.MakeMove(moveByUCI(t, b, uci))
	}
	if got := b.Result(); got != Checkmate {
		t.Errorf("after fool's mate: Result() = %v, want Checkmate", got)
	}
	if !b.InCheck() {
		t.Errorf("after fool's mate: InCheck() = false, want true")
	}
	if b.ToMove != White {
		t.Errorf("after fool's mate: the mated side should be White (to move), got %d", b.ToMove)
	}
}
