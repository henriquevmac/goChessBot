package bot

import (
	"testing"

	"goChessBot/internal/chess"
)

// quietMoves returns the moves that are neither captures nor promotions, in generation order.
func quietMoves(moves []chess.Move) []chess.Move {
	var out []chess.Move
	for _, m := range moves {
		if !m.IsCapture() && !m.IsPromotion() {
			out = append(out, m)
		}
	}
	return out
}

func moveByName(t *testing.T, b *chess.Board, uci string) chess.Move {
	t.Helper()
	for _, m := range b.GenerateLegalMoves(nil) {
		if m.String() == uci {
			return m
		}
	}
	t.Fatalf("%s is not legal in %s", uci, b.FEN())
	return 0
}

func TestRecordCutoffKillers(t *testing.T) {
	b := mustBoard(t, orderingFENs[0])
	s := testSearcher(b)
	s.ply = 3
	nf3, nc3 := moveByName(t, b, "g1f3"), moveByName(t, b, "b1c3")

	s.recordCutoff(nf3, 2)
	if got := s.killers[3]; got[0] != nf3 {
		t.Fatalf("after one cutoff: killers = %v, want %s first", got, nf3)
	}

	s.recordCutoff(nc3, 2)
	if got := s.killers[3]; got[0] != nc3 || got[1] != nf3 {
		t.Errorf("after a second cutoff: killers = [%s %s], want [%s %s]", got[0], got[1], nc3, nf3)
	}

	// The same move again must not fill both slots.
	s.recordCutoff(nc3, 2)
	if got := s.killers[3]; got[0] != nc3 || got[1] != nf3 {
		t.Errorf("after repeating the first killer: killers = [%s %s], want [%s %s] (no duplicate)", got[0], got[1], nc3, nf3)
	}

	// The second killer causing a cutoff again moves to the front.
	s.recordCutoff(nf3, 2)
	if got := s.killers[3]; got[0] != nf3 || got[1] != nc3 {
		t.Errorf("after the second killer cuts off again: killers = [%s %s], want [%s %s]", got[0], got[1], nf3, nc3)
	}

	// Other plies are untouched.
	for ply := range s.killers {
		if ply != 3 && s.killers[ply] != [2]chess.Move{} {
			t.Errorf("killers at ply %d changed: %v", ply, s.killers[ply])
		}
	}
}

func TestRecordCutoffHistory(t *testing.T) {
	b := mustBoard(t, orderingFENs[0])
	s := testSearcher(b)
	nf3 := moveByName(t, b, "g1f3")

	s.recordCutoff(nf3, 3)
	if got := s.historyScores[nf3.From()][nf3.To()]; got != 9 {
		t.Errorf("after a cutoff at depth 3: history = %d, want 9 (depth × depth)", got)
	}
	s.recordCutoff(nf3, 2)
	if got := s.historyScores[nf3.From()][nf3.To()]; got != 13 {
		t.Errorf("after another cutoff at depth 2: history = %d, want 13", got)
	}
}

// Captures and promotions are already ordered well, so they are not recorded.
func TestRecordCutoffIgnoresCapturesAndPromotions(t *testing.T) {
	tests := []struct{ name, fen, move string }{
		{"capture", "4k3/8/8/3q4/4P3/8/8/3QK3 w - - 0 1", "e4d5"},
		{"promotion", "4k3/P7/8/8/8/8/8/4K3 w - - 0 1", "a7a8q"},
		{"en passant", "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2", "e5d6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			s := testSearcher(b)
			m := moveByName(t, b, tt.move)
			s.recordCutoff(m, 4)
			if s.killers[0] != [2]chess.Move{} {
				t.Errorf("%s was stored as a killer: %v", m, s.killers[0])
			}
			if got := s.historyScores[m.From()][m.To()]; got != 0 {
				t.Errorf("%s got a history score of %d", m, got)
			}
		})
	}
}

// Extremely deep lines must not crash the killer table.
func TestRecordCutoffBeyondMaxPly(t *testing.T) {
	b := mustBoard(t, orderingFENs[0])
	s := testSearcher(b)
	s.ply = maxPly + 5
	s.recordCutoff(moveByName(t, b, "g1f3"), 1) // must not panic
	moves := b.GenerateLegalMoves(nil)
	s.orderMoves(moves, chess.Move(0)) // must not panic either
}

// Killers come right after the captures and promotions, the first before the second,
// and before every other quiet move.
func TestOrderMovesKillers(t *testing.T) {
	b := mustBoard(t, orderingFENs[4]) // middlegame with several captures
	s := testSearcher(b)
	quiet := quietMoves(b.GenerateLegalMoves(nil))
	// The two quiet moves generated last, so they would come last without killers.
	k1, k2 := quiet[len(quiet)-1], quiet[len(quiet)-2]
	s.killers[0] = [2]chess.Move{k1, k2}

	moves := b.GenerateLegalMoves(nil)
	s.orderMoves(moves, chess.Move(0))

	i1, i2 := indexOf(t, moves, k1.String()), indexOf(t, moves, k2.String())
	if i1 > i2 {
		t.Errorf("first killer %s at %d, second killer %s at %d: want the first killer ahead", k1, i1, k2, i2)
	}
	for i, m := range moves {
		noisy := m.IsCapture() || m.IsPromotion()
		switch {
		case noisy && i > i1:
			t.Errorf("capture/promotion %s at %d comes after the first killer at %d", m, i, i1)
		case !noisy && m != k1 && m != k2 && i < i2:
			t.Errorf("quiet move %s at %d comes before the second killer at %d", m, i, i2)
		}
	}
}

// The stored best move still comes first, even when it is also a killer.
func TestOrderMovesTTMoveBeforeKillers(t *testing.T) {
	b := mustBoard(t, orderingFENs[4])
	s := testSearcher(b)
	quiet := quietMoves(b.GenerateLegalMoves(nil))
	ttMove := quiet[0]
	s.killers[0] = [2]chess.Move{quiet[len(quiet)-1], ttMove}

	moves := b.GenerateLegalMoves(nil)
	s.orderMoves(moves, ttMove)
	if moves[0] != ttMove {
		t.Errorf("first move %s, want the stored move %s", moves[0], ttMove)
	}
}

// Killers are per ply: a killer stored for another ply has no effect here.
func TestOrderMovesKillersArePerPly(t *testing.T) {
	b := mustBoard(t, orderingFENs[4])
	s := testSearcher(b)
	quiet := quietMoves(b.GenerateLegalMoves(nil))
	last := quiet[len(quiet)-1]
	s.killers[5] = [2]chess.Move{last, 0}
	s.ply = 2

	moves := b.GenerateLegalMoves(nil)
	s.orderMoves(moves, chess.Move(0))
	if moves[len(moves)-1] != last {
		t.Errorf("a killer from ply 5 moved %s at ply 2; order %v", last, uciList(moves))
	}
}

// Among the other quiet moves, a higher history score comes first, but never
// ahead of a killer or a capture, however high it gets.
func TestOrderMovesHistory(t *testing.T) {
	b := mustBoard(t, orderingFENs[4])
	s := testSearcher(b)
	quiet := quietMoves(b.GenerateLegalMoves(nil))
	killer := quiet[0]
	high, low, huge := quiet[len(quiet)-1], quiet[len(quiet)-2], quiet[len(quiet)-3]
	s.killers[0] = [2]chess.Move{killer, 0}
	s.historyScores[high.From()][high.To()] = 500
	s.historyScores[low.From()][low.To()] = 100
	s.historyScores[huge.From()][huge.To()] = 1_000_000_000

	moves := b.GenerateLegalMoves(nil)
	s.orderMoves(moves, chess.Move(0))

	iKiller := indexOf(t, moves, killer.String())
	iHigh, iLow, iHuge := indexOf(t, moves, high.String()), indexOf(t, moves, low.String()), indexOf(t, moves, huge.String())
	if iHigh > iLow {
		t.Errorf("history 500 move %s at %d after history 100 move %s at %d", high, iHigh, low, iLow)
	}
	if iHuge < iKiller {
		t.Errorf("a huge history score put %s (at %d) ahead of the killer %s (at %d)", huge, iHuge, killer, iKiller)
	}
	for i, m := range moves {
		if (m.IsCapture() || m.IsPromotion()) && i > iHuge {
			t.Errorf("capture %s at %d comes after the huge-history quiet move %s at %d", m, i, huge, iHuge)
		}
	}
	// Quiet moves without history come after those with history.
	for i, m := range moves {
		if !m.IsCapture() && !m.IsPromotion() && m != killer && s.historyScores[m.From()][m.To()] == 0 && i < iLow {
			t.Errorf("quiet move %s without history at %d comes before %s with history at %d", m, i, low, iLow)
		}
	}
}

// After a real search, some quiet cutoff moves have been recorded.
func TestSearchRecordsCutoffs(t *testing.T) {
	b := mustBoard(t, orderingFENs[0])
	s := testSearcher(b)
	s.alphaBeta(-infinite, infinite, 4)

	killers, history := 0, 0
	for _, k := range s.killers {
		for _, m := range k {
			if m != 0 {
				killers++
			}
		}
	}
	for from := range s.historyScores {
		for to := range s.historyScores[from] {
			if s.historyScores[from][to] > 0 {
				history++
			}
		}
	}
	if killers == 0 || history == 0 {
		t.Errorf("after a depth-4 search: %d killer moves and %d history entries, want some of each", killers, history)
	}
}
