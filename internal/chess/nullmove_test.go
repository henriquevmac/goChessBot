package chess

import "testing"

var nullMovePositions = []string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R b KQkq - 3 1",
	"rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3", // en passant possible
	"rnbqkbnr/pppp1ppp/8/8/3Pp3/8/PPP1PPPP/RNBQKBNR b KQkq d3 0 2",  // en passant possible, black to move
}

// A null move only passes the turn: the other side moves, en passant is gone,
// the halfmove clock starts again, and the hash matches one computed from scratch.
func TestMakeNullMove(t *testing.T) {
	for _, fen := range nullMovePositions {
		b := mustBoard(t, fen)
		before := *b
		b.MakeNullMove()
		want := before
		want.ToMove = before.ToMove.OppositeColor()
		want.EnPassantSquare = -1
		want.HalfmoveClock = 0
		if diff := boardDiff(b, &want); diff != "" {
			t.Errorf("%s: after the null move:\n  %s", fen, diff)
		}
		if b.Hash != b.computeHash() {
			t.Errorf("%s: hash after the null move is %x, computed from scratch %x", fen, b.Hash, b.computeHash())
		}
	}
}

func TestMakeUnmakeNullMoveRestoresPosition(t *testing.T) {
	for _, fen := range nullMovePositions {
		b := mustBoard(t, fen)
		before := *b
		u := b.MakeNullMove()
		b.UnmakeNullMove(u)
		if diff := boardDiff(b, &before); diff != "" {
			t.Errorf("%s: make and unmake null move changed the board:\n  %s", fen, diff)
		}
		if b.Hash != before.Hash {
			t.Errorf("%s: hash after unmake is %x, want %x", fen, b.Hash, before.Hash)
		}
	}
}

// After a null move, real moves must keep the hash right: the en passant key
// must not be left behind for the next MakeMove to toggle back in.
func TestMovesAfterNullMoveKeepHash(t *testing.T) {
	for _, fen := range nullMovePositions {
		b := mustBoard(t, fen)
		b.MakeNullMove()
		for _, m := range b.GenerateLegalMoves(nil) {
			u := b.MakeMove(m)
			if b.Hash != b.computeHash() {
				t.Errorf("%s: after the null move and %s the hash is %x, computed from scratch %x", fen, m, b.Hash, b.computeHash())
			}
			b.UnmakeMove(m, u)
		}
	}
}

// The side that did not pass cannot capture en passant: the double pawn push
// was not the last move any more.
func TestNullMoveRemovesEnPassant(t *testing.T) {
	b := mustBoard(t, "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3")
	b.MakeNullMove() // black to move
	b.MakeNullMove() // white to move again, but black passed in between
	for _, m := range b.GenerateLegalMoves(nil) {
		if m.Flag() == EnPassant {
			t.Errorf("en passant %s is still possible after two null moves", m)
		}
	}
}
