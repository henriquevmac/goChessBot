package bot

import "testing"

// These positions all have a forced mate, found within depth 4.
const (
	// White mates at once with Ra8 (the black king is boxed in by its own pawns).
	mateIn1FEN = "6k1/5ppp/8/8/8/8/8/R5K1 w - - 0 1"
	// White mates in two: 1.Rb7 (or 1.Ra7), then mate on the back rank.
	mateIn2FEN = "7k/8/8/8/8/8/1R6/R5K1 w - - 0 1"
	// Black to move is mated next move whatever it plays (Kb8, then Rh8 mates).
	matedIn1FEN = "k7/8/1K6/8/8/8/8/7R b - - 0 1"
)

func searchScore(t *testing.T, fen string, depth int) int {
	t.Helper()
	b := mustBoard(t, fen)
	return testSearcher(b).alphaBeta(-infinite, infinite, depth)
}

func TestMateScoresPreferFasterMates(t *testing.T) {
	const depth = 4
	in1 := searchScore(t, mateIn1FEN, depth)
	in2 := searchScore(t, mateIn2FEN, depth)

	// Any mate must beat any material advantage (all material together is under 10,000).
	for name, score := range map[string]int{"mate in 1": in1, "mate in 2": in2} {
		if score < 10_000 {
			t.Errorf("%s scores %d, want a mate score above any material score", name, score)
		}
	}
	if in1 <= in2 {
		t.Errorf("mate in 1 scores %d, mate in 2 scores %d: want the faster mate to score higher", in1, in2)
	}
}

func TestBeingMatedLaterIsBetter(t *testing.T) {
	// From the losing side's view: already mated (ply 0) < mated in 1 (after its only move).
	mated := searchScore(t, "R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1", 3) // checkmated right now
	matedIn1 := searchScore(t, matedIn1FEN, 3)

	for name, score := range map[string]int{"checkmated": mated, "mated in 1": matedIn1} {
		if score > -10_000 {
			t.Errorf("%s scores %d, want a mate score below any material score", name, score)
		}
	}
	if matedIn1 <= mated {
		t.Errorf("being mated in 1 scores %d, being checkmated now scores %d: want later mates to score higher", matedIn1, mated)
	}
}

// Mate scores stored in the transposition table must come back meaning the same
// thing: searching again on the same (now filled) table must give the same score.
func TestMateScoreSurvivesTheTable(t *testing.T) {
	for _, fen := range []string{mateIn1FEN, mateIn2FEN, matedIn1FEN} {
		b := mustBoard(t, fen)
		s := testSearcher(b)
		var scores []int
		for depth := 1; depth <= 5; depth++ {
			scores = append(scores, s.alphaBeta(-infinite, infinite, depth))
		}
		fresh := testSearcher(b).alphaBeta(-infinite, infinite, 5)
		if last := scores[len(scores)-1]; last != fresh {
			t.Errorf("%s: depth 5 scores %d after reusing the table, %d with a fresh table (scores by depth %v)",
				fen, last, fresh, scores)
		}
	}
}

// The bot itself still plays a mate in one when there is one, at any depth.
func TestBestMovePlaysMateInOne(t *testing.T) {
	forEachSearchBot(t, func(t *testing.T, bot Bot) {
		for depth := 2; depth <= 5; depth++ {
			b := mustBoard(t, mateIn1FEN)
			if got := bot.BestMove(*b, Limits{Depth: depth}).String(); got != "a1a8" {
				t.Errorf("depth %d: BestMove = %s, want the mate a1a8", depth, got)
			}
		}
	})
}

// scoreToTT and scoreFromTT move mate scores between "plies from the root" and
// "plies from this position". Normal scores must pass through unchanged.
func TestTTMateScoreConversion(t *testing.T) {
	for _, score := range []int{0, 37, -250, 9_000, -9_000} {
		for _, ply := range []int{0, 1, 7} {
			if got := scoreToTT(score, ply); got != score {
				t.Errorf("scoreToTT(%d, %d) = %d, want a normal score unchanged", score, ply, got)
			}
			if got := scoreFromTT(score, ply); got != score {
				t.Errorf("scoreFromTT(%d, %d) = %d, want a normal score unchanged", score, ply, got)
			}
		}
	}

	// Mated 3 plies from the root, seen at ply 2: one ply away from this position.
	losing := matedScore(3)
	winning := -matedScore(3)
	for _, score := range []int{losing, winning} {
		stored := scoreToTT(score, 2)
		if back := scoreFromTT(stored, 2); back != score {
			t.Errorf("round trip at the same ply: %d → %d → %d", score, stored, back)
		}
		// The same position reached at ply 4: the mate is now 5 plies from the root.
		want := matedScore(5)
		if score > 0 {
			want = -want
		}
		if got := scoreFromTT(stored, 4); got != want {
			t.Errorf("mate score %d stored at ply 2, read at ply 4: got %d, want %d", score, got, want)
		}
	}
}

// After a search, the table entry for the root of a mate in 1 must read back as
// exactly that mate.
func TestTTStoresMateRelativeToPosition(t *testing.T) {
	b := mustBoard(t, mateIn1FEN)
	s := testSearcher(b)
	s.alphaBeta(-infinite, infinite, 3)

	entry, ok := s.tt.probe(b.Hash)
	if !ok {
		t.Fatalf("no table entry for the root after searching it")
	}
	if got, want := scoreFromTT(entry.score, 0), -matedScore(1); got != want {
		t.Errorf("root entry reads back as %d, want %d (mate in 1)", got, want)
	}
}

// A stored mate score must be converted back when the search uses it. The entry
// is planted as if found at ply 2, then the same position is searched at ply 4:
// the score alphaBeta returns must say "mate 5 plies from the root", not 3.
func TestTTMateScoreIsConvertedWhenRead(t *testing.T) {
	b := mustBoard(t, mateIn2FEN)
	s := testSearcher(b)
	s.ply = 4

	winning := -matedScore(3) // mate 3 plies from the root, seen from ply 2
	s.tt.store(ttEntry{key: b.Hash, depth: 99, score: scoreToTT(winning, 2), flag: ttExact})

	got := s.alphaBeta(-infinite, infinite, 1)
	if want := -matedScore(5); got != want {
		t.Errorf("alphaBeta at ply 4 returned %d from the table, want %d (mate 5 plies from the root)", got, want)
	}
}
