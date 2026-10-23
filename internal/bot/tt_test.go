package bot

import (
	"testing"
	"unsafe"

	"goChessBot/internal/chess"
)

// testSearcher returns a searcher with a small, fresh transposition table, as
// BestMove would create one.
func testSearcher(b *chess.Board) *searcher {
	return &searcher{board: b, tt: newTranspositionTable(16)}
}

// Keep entries at 16 bytes: a new field, or the fields in another order, can
// add padding without anyone noticing.
func TestTTEntrySize(t *testing.T) {
	if size := unsafe.Sizeof(ttEntry{}); size != 16 {
		t.Errorf("ttEntry is %d bytes, want 16", size)
	}
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
		tt.store(ttEntry{key: i<<32 | i, depth: int8(i % 10), score: int32(i)})
	}
	for i := uint64(0); i < 1<<12; i++ {
		got, ok := tt.probe(i<<32 | i)
		if !ok || got.score != int32(i) {
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

// Two new bots searching the same position give the same move.
func TestBestMoveIsRepeatable(t *testing.T) {
	for _, bt := range searchBots {
		for _, fen := range orderingFENs {
			b := mustBoard(t, fen)
			first := New(bt).BestMove(*b, Limits{Depth: 3})
			second := New(bt).BestMove(*b, Limits{Depth: 3})
			if first != second {
				t.Errorf("bot type %d, %s: two identical searches gave %s and %s", bt, fen, first, second)
			}
		}
	}
}

func filledEntries(tt *transpositionTable) int {
	n := 0
	for _, e := range tt.entries {
		if e != (ttEntry{}) {
			n++
		}
	}
	return n
}

// The bot keeps its table from one move to the next.
func TestTableIsKeptBetweenMoves(t *testing.T) {
	bot := New(AlphaBeta).(*alphaBetaBot)
	b := mustBoard(t, orderingFENs[0])
	bot.BestMove(*b, Limits{Depth: 3})
	table := bot.tt
	if table == nil || filledEntries(table) == 0 {
		t.Fatalf("no filled table after the first move")
	}
	after := filledEntries(table)

	b = mustBoard(t, orderingFENs[4])
	bot.BestMove(*b, Limits{Depth: 3})
	if bot.tt != table {
		t.Errorf("the second move got a new table")
	}
	if filledEntries(table) <= after {
		t.Errorf("the table did not grow on the second move (%d entries before, %d after)", after, filledEntries(table))
	}
}

// Separate bots don't share a table.
func TestBotsHaveTheirOwnTable(t *testing.T) {
	b := mustBoard(t, orderingFENs[0])
	first, second := New(AlphaBeta).(*alphaBetaBot), New(AlphaBeta).(*alphaBetaBot)
	first.BestMove(*b, Limits{Depth: 2})
	second.BestMove(*b, Limits{Depth: 2})
	if first.tt == second.tt {
		t.Errorf("two bots share one table")
	}
}

// A table filled by earlier moves must not change what the bot finds: after
// searching other positions, it still plays the mate in one, and its score for
// the mate in two is the same as with a fresh table.
func TestReusedTableGivesSameResults(t *testing.T) {
	bot := New(AlphaBeta).(*alphaBetaBot)
	for _, fen := range orderingFENs {
		bot.BestMove(*mustBoard(t, fen), Limits{Depth: 3})
	}
	if got := bot.BestMove(*mustBoard(t, mateIn1FEN), Limits{Depth: 3}).String(); got != "a1a8" {
		t.Errorf("with a used table: BestMove = %s, want the mate a1a8", got)
	}

	b := mustBoard(t, mateIn2FEN)
	used := &searcher{board: b, tt: bot.tt}
	fresh := testSearcher(b)
	if u, f := used.alphaBeta(-infinite, infinite, 4), fresh.alphaBeta(-infinite, infinite, 4); u != f {
		t.Errorf("mate in two scores %d with a used table, %d with a fresh one", u, f)
	}
}

// With NoTT the bot never creates a table.
func TestNoTTBotHasNoTable(t *testing.T) {
	bot := NewAlphaBeta(AlphaBetaOptions{NoTT: true}).(*alphaBetaBot)
	bot.BestMove(*mustBoard(t, orderingFENs[0]), Limits{Depth: 3})
	if bot.tt != nil {
		t.Errorf("a NoTT bot created a table")
	}
}

// With FreshTT the bot starts every move with a new, empty table.
func TestFreshTTGivesANewTableEveryMove(t *testing.T) {
	bot := NewAlphaBeta(AlphaBetaOptions{FreshTT: true}).(*alphaBetaBot)
	b := mustBoard(t, orderingFENs[0])
	bot.BestMove(*b, Limits{Depth: 3})
	first := bot.tt
	bot.BestMove(*b, Limits{Depth: 3})
	if bot.tt == first {
		t.Errorf("the second move reused the first move's table")
	}
	if filledEntries(first) == 0 || filledEntries(bot.tt) == 0 {
		t.Errorf("a FreshTT bot did not fill its tables")
	}
}
