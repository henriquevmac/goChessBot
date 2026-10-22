package bot

import (
	"testing"

	"goChessBot/internal/chess"
)

// testSearcher returns a searcher with a small, fresh transposition table, as
// BestMove would create one.
func testSearcher(b *chess.Board) *searcher {
	return &searcher{board: b, tt: newTranspositionTable(16)}
}

func TestTTStoreAndProbe(t *testing.T) {
	tt := newTranspositionTable(10)
	e := ttEntry{key: 0xDEADBEEF12345678, bestMove: chess.CreateMove(12, 28, chess.DoublePawn), depth: 4, score: 37, flag: ttExact}
	tt.store(e)

	got, ok := tt.probe(e.key)
	if !ok {
		t.Fatalf("probe did not find an entry that was just stored")
	}
	if got != e {
		t.Errorf("probe returned %+v, want %+v", got, e)
	}
}

func TestTTEmptyTableFindsNothing(t *testing.T) {
	tt := newTranspositionTable(10)
	for _, key := range []uint64{1, 0x9E3779B97F4A7C15, 0xFFFFFFFFFFFFFFFF} {
		if _, ok := tt.probe(key); ok {
			t.Errorf("probe(%#x) found an entry in an empty table", key)
		}
	}
}

// Two different positions can share a slot (same low bits). The table must
// never return one position's entry for the other.
func TestTTDifferentKeySameSlotIsNotAMatch(t *testing.T) {
	const sizeLog2 = 10
	tt := newTranspositionTable(sizeLog2)
	a := uint64(0x0000_0000_0000_0123)
	b := a | 1<<40 // same lowest sizeLog2 bits as a, so the same slot
	tt.store(ttEntry{key: a, depth: 3, score: 100, flag: ttExact})

	if _, ok := tt.probe(b); ok {
		t.Errorf("probe returned the entry of a different position sharing its slot")
	}
	if _, ok := tt.probe(a); !ok {
		t.Errorf("probe did not find the stored position")
	}
}

func TestTTStoreReplaces(t *testing.T) {
	tt := newTranspositionTable(10)
	key := uint64(0xABCDEF)
	tt.store(ttEntry{key: key, depth: 2, score: 10, flag: ttLower})
	newer := ttEntry{key: key, depth: 5, score: -20, flag: ttUpper}
	tt.store(newer)
	if got, _ := tt.probe(key); got != newer {
		t.Errorf("after storing twice, probe returned %+v, want the newer %+v", got, newer)
	}
}

// Different keys in different slots don't disturb each other.
func TestTTManyKeys(t *testing.T) {
	tt := newTranspositionTable(12)
	for i := uint64(0); i < 1<<12; i++ {
		tt.store(ttEntry{key: i<<32 | i, depth: int(i % 10), score: int(i)})
	}
	for i := uint64(0); i < 1<<12; i++ {
		got, ok := tt.probe(i<<32 | i)
		if !ok || got.score != int(i) {
			t.Fatalf("key %d: probe = %+v, %v; want score %d", i, got, ok, i)
		}
	}
}

// The stored best move is tried first, before even the best capture.
func TestOrderMovesTTMoveFirst(t *testing.T) {
	for _, fen := range orderingFENs {
		b := mustBoard(t, fen)
		moves := b.GenerateLegalMoves(nil)
		// Use the last quiet move as the "stored" move, the least likely to come first otherwise.
		ttMove := chess.Move(0)
		for _, m := range moves {
			if !m.IsCapture() && !m.IsPromotion() {
				ttMove = m
			}
		}
		if ttMove == 0 {
			continue
		}
		testSearcher(b).orderMoves(moves, ttMove)
		if moves[0] != ttMove {
			t.Errorf("%s: stored move %s is at position %d, want 0; order %v",
				fen, ttMove, indexOf(t, moves, ttMove.String()), uciList(moves))
		}
	}
}

// With no stored move, the order must be the same as before (captures first, MVV-LVA).
func TestOrderMovesNoTTMove(t *testing.T) {
	b := mustBoard(t, "4k3/8/8/3q4/4P3/8/8/3QK3 w - - 0 1")
	moves := b.GenerateLegalMoves(nil)
	testSearcher(b).orderMoves(moves, chess.Move(0))
	if moves[0].String() != "e4d5" {
		t.Errorf("without a stored move, first move = %s, want e4d5 (pawn takes queen)", moves[0])
	}
}

// Searching the same position twice with a fresh table gives the same move.
func TestBestMoveIsRepeatable(t *testing.T) {
	forEachSearchBot(t, func(t *testing.T, bot Bot) {
		for _, fen := range orderingFENs {
			b := mustBoard(t, fen)
			first := bot.BestMove(*b, Limits{Depth: 3})
			second := bot.BestMove(*b, Limits{Depth: 3})
			if first != second {
				t.Errorf("%s: two identical searches gave %s and %s", fen, first, second)
			}
		}
	})
}
