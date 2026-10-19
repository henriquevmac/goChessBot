package chess

import "testing"

func TestHashFromFEN(t *testing.T) {
	start := mustBoard(t, startFEN)
	if start.Hash == 0 {
		t.Fatalf("start position hash is 0; is it computed in CreateBoard?")
	}
	if start.Hash != start.computeHash() {
		t.Errorf("CreateBoard hash %x does not match computeHash %x", start.Hash, start.computeHash())
	}

	// Positions that differ only in one hashed detail must get different hashes.
	differ := []struct {
		name string
		a, b string
	}{
		{"side to move", "4k3/8/8/8/8/8/8/R3K3 w - - 0 1", "4k3/8/8/8/8/8/8/R3K3 b - - 0 1"},
		{"castling rights", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", "r3k2r/8/8/8/8/8/8/R3K2R w Kkq - 0 1"},
		{"no castling vs some", "r3k2r/8/8/8/8/8/8/R3K2R w - - 0 1", "r3k2r/8/8/8/8/8/8/R3K2R w q - 0 1"},
		{"en passant square", "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2", "4k3/8/8/3pP3/8/8/8/4K3 w - - 0 2"},
		{"en passant file", "4k3/8/8/2ppP3/8/8/8/4K3 w - d6 0 2", "4k3/8/8/2ppP3/8/8/8/4K3 w - c6 0 2"},
		{"piece on another square", "4k3/8/8/8/8/8/8/R3K3 w - - 0 1", "4k3/8/8/8/8/8/8/1R2K3 w - - 0 1"},
		{"different piece type", "4k3/8/8/8/8/8/8/R3K3 w - - 0 1", "4k3/8/8/8/8/8/8/N3K3 w - - 0 1"},
		{"different piece color", "4k3/8/8/8/8/8/8/R3K3 w - - 0 1", "4k3/8/8/8/8/8/8/r3K3 w - - 0 1"},
		{"extra piece", "4k3/8/8/8/8/8/8/R3K3 w - - 0 1", "4k3/8/8/8/8/8/P7/R3K3 w - - 0 1"},
	}
	for _, tt := range differ {
		a, b := mustBoard(t, tt.a), mustBoard(t, tt.b)
		if a.Hash == b.Hash {
			t.Errorf("%s: %q and %q have the same hash %x", tt.name, tt.a, tt.b, a.Hash)
		}
	}

	// The move clocks are not part of the position, so they must not change the hash.
	a := mustBoard(t, "4k3/8/8/8/8/8/8/R3K3 w - - 0 1")
	b := mustBoard(t, "4k3/8/8/8/8/8/8/R3K3 w - - 37 80")
	if a.Hash != b.Hash {
		t.Errorf("move clocks changed the hash: %x vs %x", a.Hash, b.Hash)
	}
}

// After every move (and two moves deep), the incrementally updated hash must equal
// a hash computed from scratch, and unmaking must restore it exactly.
func TestHashIncrementalUpdate(t *testing.T) {
	for _, fen := range makeUnmakeFENs {
		b := mustBoard(t, fen)
		original := b.Hash
		for _, m1 := range b.GenerateLegalMoves(nil) {
			u1 := b.MakeMove(m1)
			if b.Hash != b.computeHash() {
				t.Errorf("%s: after %s, Hash = %x, computeHash = %x", fen, m1, b.Hash, b.computeHash())
			}
			afterFirst := b.Hash
			for _, m2 := range b.GenerateLegalMoves(nil) {
				u2 := b.MakeMove(m2)
				if b.Hash != b.computeHash() {
					t.Errorf("%s: after %s %s, Hash = %x, computeHash = %x", fen, m1, m2, b.Hash, b.computeHash())
				}
				b.UnmakeMove(m2, u2)
				if b.Hash != afterFirst {
					t.Errorf("%s: unmaking %s after %s did not restore the hash", fen, m2, m1)
				}
			}
			b.UnmakeMove(m1, u1)
			if b.Hash != original {
				t.Errorf("%s: unmaking %s did not restore the hash", fen, m1)
			}
		}
	}
}

// Every kind of special move must update the hash correctly.
func TestHashSpecialMoves(t *testing.T) {
	tests := []struct {
		name  string
		fen   string
		moves []string
	}{
		{"double push sets en passant", startFEN, []string{"e2e4"}},
		{"en passant square is cleared", startFEN, []string{"e2e4", "g8f6"}},
		{"en passant capture", "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2", []string{"e5d6"}},
		{"kingside castling", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"e1g1"}},
		{"queenside castling", "r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1", []string{"e8c8"}},
		{"rook capture removes castling right", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"h1h8"}},
		{"promotion", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", []string{"e7e8n"}},
		{"capture promotion", "3r1r2/4P3/8/8/8/8/8/k3K3 w - - 0 1", []string{"e7d8q"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			for _, uci := range tt.moves {
				b.MakeMove(moveByUCI(t, b, uci))
				if b.Hash != b.computeHash() {
					t.Errorf("after %s: Hash = %x, computeHash = %x", uci, b.Hash, b.computeHash())
				}
			}
		})
	}
}

// The same position reached by different move orders must have the same hash,
// and the same hash as loading that position from a FEN.
func TestHashTransposition(t *testing.T) {
	a := mustBoard(t, startFEN)
	for _, uci := range []string{"g1f3", "g8f6", "b1c3"} {
		a.MakeMove(moveByUCI(t, a, uci))
	}
	b := mustBoard(t, startFEN)
	for _, uci := range []string{"b1c3", "g8f6", "g1f3"} {
		b.MakeMove(moveByUCI(t, b, uci))
	}
	if a.Hash != b.Hash {
		t.Errorf("transposed move orders give different hashes: %x vs %x", a.Hash, b.Hash)
	}
	fromFEN := mustBoard(t, "rnbqkb1r/pppppppp/5n2/8/8/2N2N2/PPPPPPPP/R1BQKB1R b KQkq - 3 2")
	if a.Hash != fromFEN.Hash {
		t.Errorf("played position hash %x differs from the FEN hash %x", a.Hash, fromFEN.Hash)
	}
}

// playWithHistory plays moves from fen, recording each position's hash before the move,
// and reports after every move whether IsThreeFoldRepetition is true.
func playWithHistory(t *testing.T, fen string, moves []string) []bool {
	t.Helper()
	b := mustBoard(t, fen)
	var history []uint64
	results := make([]bool, len(moves))
	for i, uci := range moves {
		history = append(history, b.Hash)
		b.MakeMove(moveByUCI(t, b, uci))
		results[i] = b.IsThreeFoldRepetition(history)
	}
	return results
}

func TestThreefoldRepetition(t *testing.T) {
	t.Run("knights out and back twice", func(t *testing.T) {
		// The start position occurs after 0, 4 and 8 moves: the third time is move 8.
		moves := []string{
			"g1f3", "g8f6", "f3g1", "f6g8",
			"g1f3", "g8f6", "f3g1", "f6g8",
		}
		got := playWithHistory(t, startFEN, moves)
		for i, rep := range got {
			want := i == len(moves)-1
			if rep != want {
				t.Errorf("after move %d (%s): IsThreeFoldRepetition = %v, want %v", i+1, moves[i], rep, want)
			}
		}
	})

	t.Run("second occurrence is not yet threefold", func(t *testing.T) {
		moves := []string{"g1f3", "g8f6", "f3g1", "f6g8"}
		got := playWithHistory(t, startFEN, moves)
		if got[len(got)-1] {
			t.Errorf("start position seen twice reported as threefold")
		}
	})

	t.Run("positions in between also count", func(t *testing.T) {
		// After move 1 (Nf3) the position with the knight on f3 and black to move
		// occurs again after moves 5 and 9.
		moves := []string{
			"g1f3", "g8f6", "f3g1", "f6g8",
			"g1f3", "g8f6", "f3g1", "f6g8",
			"g1f3",
		}
		got := playWithHistory(t, startFEN, moves)
		if !got[7] {
			t.Errorf("after move 8: start position (third time) not detected")
		}
		if !got[8] {
			t.Errorf("after move 9: knight-on-f3 position (third time) not detected")
		}
	})

	t.Run("lost castling rights make it a different position", func(t *testing.T) {
		// The kings step out and back. After move 4 the pieces stand exactly as at the
		// start, but without castling rights, so it is a different position: after
		// move 8 the start setup has been on the board 3 times, but only twice
		// without castling rights, which is not a repetition.
		moves := []string{
			"e1f1", "e8f8", "f1e1", "f8e8",
			"e1f1", "e8f8", "f1e1", "f8e8",
			"e1f1", "e8f8",
		}
		got := playWithHistory(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", moves)
		for i := 0; i < 9; i++ {
			if got[i] {
				t.Errorf("after move %d (%s): IsThreeFoldRepetition = true, want false", i+1, moves[i])
			}
		}
		// Kings on f1 and f8 without castling rights first occurred after move 2,
		// again after move 6, and a third time after move 10.
		if !got[9] {
			t.Errorf("after move 10: third occurrence of the kings on f1/f8 not detected")
		}
	})

	t.Run("a second occurrence is never threefold", func(t *testing.T) {
		// Rooks shuffle back and forth; every position occurs at most twice.
		moves := []string{"a1a2", "h8h7", "a2a1", "h7h8", "a1a2"}
		got := playWithHistory(t, "4k2r/8/8/8/8/8/8/R3K3 w - - 0 1", moves)
		for i, rep := range got {
			if rep {
				t.Errorf("after move %d (%s): reported threefold too early", i+1, moves[i])
			}
		}
	})

	t.Run("empty history", func(t *testing.T) {
		b := mustBoard(t, startFEN)
		if b.IsThreeFoldRepetition(nil) {
			t.Errorf("IsThreeFoldRepetition(nil) = true, want false")
		}
	})
}
