package chess

import (
	"fmt"
	"strings"
	"testing"
)

// boardDiff describes how two boards differ, or returns "" if they are identical.
func boardDiff(got, want *Board) string {
	var diffs []string
	for sq := 0; sq < 64; sq++ {
		if got.Square[sq] != want.Square[sq] {
			diffs = append(diffs, fmt.Sprintf("%s: got piece %d, want %d", squareName(sq), got.Square[sq], want.Square[sq]))
		}
	}
	if got.ToMove != want.ToMove {
		diffs = append(diffs, fmt.Sprintf("ToMove: got %d, want %d", got.ToMove, want.ToMove))
	}
	if got.CastlingRights != want.CastlingRights {
		diffs = append(diffs, fmt.Sprintf("CastlingRights: got %04b, want %04b", got.CastlingRights, want.CastlingRights))
	}
	if got.EnPassantSquare != want.EnPassantSquare {
		diffs = append(diffs, fmt.Sprintf("EnPassantSquare: got %d, want %d", got.EnPassantSquare, want.EnPassantSquare))
	}
	if got.HalfmoveClock != want.HalfmoveClock {
		diffs = append(diffs, fmt.Sprintf("HalfmoveClock: got %d, want %d", got.HalfmoveClock, want.HalfmoveClock))
	}
	if got.FullmoveNumber != want.FullmoveNumber {
		diffs = append(diffs, fmt.Sprintf("FullmoveNumber: got %d, want %d", got.FullmoveNumber, want.FullmoveNumber))
	}
	if got.KingPos != want.KingPos {
		diffs = append(diffs, fmt.Sprintf("KingPos [white, black]: got %v, want %v", got.KingPos, want.KingPos))
	}
	if len(diffs) == 0 && *got != *want {
		diffs = append(diffs, "boards differ in a field boardDiff does not know about")
	}
	return strings.Join(diffs, "\n  ")
}

// moveByUCI finds the generated move with the given UCI string.
func moveByUCI(t *testing.T, b *Board, uci string) Move {
	t.Helper()
	m, ok := findMove(b.GenerateMoves(nil), uci)
	if !ok {
		t.Fatalf("move %s is not generated in this position", uci)
	}
	return m
}

// Each case plays a sequence of moves and compares the result with a FEN.
func TestMakeMove(t *testing.T) {
	tests := []struct {
		name  string
		fen   string
		moves []string
		want  string
	}{
		// Basic moves, side to move, clocks, en passant square
		{"double push sets en passant square", startFEN, []string{"e2e4"},
			"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"},
		{"black reply, fullmove increments", startFEN, []string{"e2e4", "e7e5"},
			"rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2"},
		{"en passant square is cleared, halfmove counts", startFEN, []string{"e2e4", "e7e5", "g1f3"},
			"rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2"},
		{"knight move increments halfmove clock", startFEN, []string{"g1f3"},
			"rnbqkbnr/pppppppp/8/8/8/5N2/PPPPPPPP/RNBQKB1R b KQkq - 1 1"},
		{"single push, no en passant square", startFEN, []string{"e2e3"},
			"rnbqkbnr/pppppppp/8/8/8/4P3/PPPP1PPP/RNBQKBNR b KQkq - 0 1"},
		{"black double push", startFEN, []string{"e2e4", "d7d5"},
			"rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2"},
		{"pawn move resets halfmove clock", "4k3/8/8/8/8/8/4P3/4K3 w - - 7 20", []string{"e2e3"},
			"4k3/8/8/8/8/4P3/8/4K3 b - - 0 20"},
		{"black move increments fullmove", "4k3/8/8/8/8/8/8/4K3 b - - 3 20", []string{"e8d8"},
			"3k4/8/8/8/8/8/8/4K3 w - - 4 21"},

		// Captures
		{"capture replaces the piece and resets halfmove clock", "4k3/8/8/3p4/8/8/8/3RK3 w - - 5 10", []string{"d1d5"},
			"4k3/8/8/3R4/8/8/8/4K3 b - - 0 10"},
		{"pawn capture", startFEN, []string{"e2e4", "d7d5", "e4d5"},
			"rnbqkbnr/ppp1pppp/8/3P4/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 2"},

		// En passant
		{"white en passant removes the pawn behind", "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2", []string{"e5d6"},
			"4k3/8/3P4/8/8/8/8/4K3 b - - 0 2"},
		{"black en passant removes the pawn behind", "4k3/8/8/8/3Pp3/8/8/4K3 b - d3 0 2", []string{"e4d3"},
			"4k3/8/8/8/8/3p4/8/4K3 w - - 0 3"},

		// Promotion
		{"promotion to queen", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", []string{"e7e8q"},
			"4Q3/8/8/8/8/8/8/k3K3 b - - 0 1"},
		{"promotion to knight", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", []string{"e7e8n"},
			"4N3/8/8/8/8/8/8/k3K3 b - - 0 1"},
		{"promotion to rook", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", []string{"e7e8r"},
			"4R3/8/8/8/8/8/8/k3K3 b - - 0 1"},
		{"promotion to bishop", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", []string{"e7e8b"},
			"4B3/8/8/8/8/8/8/k3K3 b - - 0 1"},
		{"black promotion with capture", "4k3/8/8/8/8/8/3p4/2R1K3 b - - 0 1", []string{"d2c1q"},
			"4k3/8/8/8/8/8/8/2q1K3 w - - 0 2"},

		// Castling moves the rook too
		{"white kingside castling", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"e1g1"},
			"r3k2r/8/8/8/8/8/8/R4RK1 b kq - 1 1"},
		{"white queenside castling", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"e1c1"},
			"r3k2r/8/8/8/8/8/8/2KR3R b kq - 1 1"},
		{"black kingside castling", "r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1", []string{"e8g8"},
			"r4rk1/8/8/8/8/8/8/R3K2R w KQ - 1 2"},
		{"black queenside castling", "r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1", []string{"e8c8"},
			"2kr3r/8/8/8/8/8/8/R3K2R w KQ - 1 2"},

		// Castling rights
		{"king move loses both rights", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"e1f1"},
			"r3k2r/8/8/8/8/8/8/R4K1R b kq - 1 1"},
		{"a1 rook move loses queenside right", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"a1b1"},
			"r3k2r/8/8/8/8/8/8/1R2K2R b Kkq - 1 1"},
		{"h1 rook move loses kingside right", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"h1g1"},
			"r3k2r/8/8/8/8/8/8/R3K1R1 b Qkq - 1 1"},
		{"black king move loses both rights", "r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1", []string{"e8d8"},
			"r2k3r/8/8/8/8/8/8/R3K2R w KQ - 1 2"},
		{"capturing a rook removes the opponent's right", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"h1h8"},
			"r3k2R/8/8/8/8/8/8/R3K3 b Qq - 0 1"},
		{"black captures a white rook", "r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1", []string{"a8a1"},
			"4k2r/8/8/8/8/8/8/r3K2R w Kk - 0 2"},
		{"rook moving away and back keeps the right lost", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			[]string{"h1g1", "a8b8", "g1h1"},
			"1r2k2r/8/8/8/8/8/8/R3K2R b Qk - 3 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			for _, uci := range tt.moves {
				b.MakeMove(moveByUCI(t, b, uci))
			}
			want := mustBoard(t, tt.want)
			if diff := boardDiff(b, want); diff != "" {
				t.Errorf("after %v from %s:\n  %s", tt.moves, tt.fen, diff)
			}
		})
	}
}

var makeUnmakeFENs = []string{
	startFEN,
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", // Kiwipete
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
	"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
	"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
	"r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
	"r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1",
	"4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2",
	"4k3/8/8/8/3Pp3/8/8/4K3 b - d3 0 2",
	"3r1r2/4P3/8/8/8/8/8/k3K3 w - - 0 1",
	"4k3/8/8/8/8/8/3p4/2R1K3 b - - 0 1",
}

// Making and then unmaking every move must give back exactly the same position.
func TestMakeUnmakeRestoresPosition(t *testing.T) {
	for _, fen := range makeUnmakeFENs {
		b := mustBoard(t, fen)
		before := *b // Board only holds arrays and numbers, so this is a full copy
		for _, m := range b.GenerateMoves(nil) {
			u := b.MakeMove(m)
			b.UnmakeMove(m, u)
			if diff := boardDiff(b, &before); diff != "" {
				t.Errorf("%s: make/unmake %s changed the position:\n  %s", fen, m, diff)
				*b = before // continue the remaining moves from a clean position
			}
		}
	}
}

// Two levels deep: undo records must not interfere with each other.
func TestMakeUnmakeTwoPly(t *testing.T) {
	for _, fen := range makeUnmakeFENs {
		b := mustBoard(t, fen)
		before := *b
		for _, m1 := range b.GenerateMoves(nil) {
			u1 := b.MakeMove(m1)
			afterFirst := *b
			for _, m2 := range b.GenerateMoves(nil) {
				u2 := b.MakeMove(m2)
				b.UnmakeMove(m2, u2)
				if diff := boardDiff(b, &afterFirst); diff != "" {
					t.Errorf("%s: after %s, make/unmake %s changed the position:\n  %s", fen, m1, m2, diff)
					*b = afterFirst
				}
			}
			b.UnmakeMove(m1, u1)
			if diff := boardDiff(b, &before); diff != "" {
				t.Errorf("%s: make/unmake %s (with replies) changed the position:\n  %s", fen, m1, diff)
				*b = before
			}
		}
	}
}

// After any move, the side to move flips and the piece count only drops on captures.
func TestMakeMoveInvariants(t *testing.T) {
	count := func(b *Board) int {
		n := 0
		for _, p := range b.Square {
			if !p.IsEmpty() {
				n++
			}
		}
		return n
	}
	for _, fen := range makeUnmakeFENs {
		b := mustBoard(t, fen)
		for _, m := range b.GenerateMoves(nil) {
			mover := b.ToMove
			piecesBefore := count(b)
			u := b.MakeMove(m)

			if b.ToMove != mover.OppositeColor() {
				t.Errorf("%s: %s: side to move did not switch", fen, m)
			}
			if b.Square[m.From()] != Empty {
				t.Errorf("%s: %s: from square is not empty after the move", fen, m)
			}
			if p := b.Square[m.To()]; p.Color() != mover {
				t.Errorf("%s: %s: to square does not hold the mover's piece (%d)", fen, m, p)
			}
			want := piecesBefore
			if m.IsCapture() {
				want--
			}
			if got := count(b); got != want {
				t.Errorf("%s: %s: %d pieces on the board, want %d", fen, m, got, want)
			}
			b.UnmakeMove(m, u)
		}
	}
}
