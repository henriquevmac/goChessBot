package bot

import (
	"fmt"
	"slices"
	"testing"

	"goChessBot/internal/chess"
)

var alphaBetaVariants = []AlphaBetaOptions{
	{},
	{NoTT: true},
	{NoKillers: true},
	{NoHistory: true},
	{NoRepetition: true},
	{FreshTT: true},
	{NoTT: true, NoKillers: true, NoHistory: true, NoRepetition: true},
}

func forEachVariant(t *testing.T, test func(t *testing.T, bot Bot)) {
	t.Helper()
	for _, o := range alphaBetaVariants {
		t.Run(fmt.Sprintf("%+v", o), func(t *testing.T) {
			test(t, NewAlphaBeta(o))
		})
	}
}

// Every variant still plays legal moves and still finds a mate in one.
func TestVariantsPlayLegalMoves(t *testing.T) {
	forEachVariant(t, func(t *testing.T, bot Bot) {
		for _, fen := range busyFENs {
			b := mustBoard(t, fen)
			legal := b.GenerateLegalMoves(nil)
			if m := bot.BestMove(*b, Limits{Depth: 3}); !slices.Contains(legal, m) {
				t.Errorf("%s: BestMove returned %s, which is not a legal move", fen, m)
			}
		}
		b := mustBoard(t, mateIn1FEN)
		if got := bot.BestMove(*b, Limits{Depth: 3}).String(); got != "a1a8" {
			t.Errorf("BestMove = %s, want the mate a1a8", got)
		}
	})
}

// With every switch off, NewAlphaBeta is the same bot as New(AlphaBeta).
func TestVariantWithoutSwitchesMatchesAlphaBeta(t *testing.T) {
	plain, variant := New(AlphaBeta), NewAlphaBeta(AlphaBetaOptions{})
	for _, fen := range orderingFENs {
		b := mustBoard(t, fen)
		if p, v := plain.BestMove(*b, Limits{Depth: 3}), variant.BestMove(*b, Limits{Depth: 3}); p != v {
			t.Errorf("%s: New(AlphaBeta) plays %s, NewAlphaBeta(AlphaBetaOptions{}) plays %s", fen, p, v)
		}
	}
}

func TestNoTTLeavesTableUnused(t *testing.T) {
	b := mustBoard(t, orderingFENs[0])
	s := testSearcher(b)
	s.options.NoTT = true
	s.alphaBeta(-infinite, infinite, 3)
	for _, e := range s.tt.entries {
		if e != (ttEntry{}) {
			t.Fatalf("the table has an entry %+v after a search with NoTT", e)
		}
	}
}

func TestNoKillersAndNoHistoryRecordNothing(t *testing.T) {
	b := mustBoard(t, orderingFENs[0])
	s := testSearcher(b)
	s.options.NoKillers, s.options.NoHistory = true, true
	s.alphaBeta(-infinite, infinite, 4)
	if s.killers != [maxPly][2]chess.Move{} {
		t.Errorf("killers were recorded with NoKillers")
	}
	if s.historyScores != [64][64]int{} {
		t.Errorf("history scores were recorded with NoHistory")
	}
}

// Without repetition detection, a queen-up bot no longer sees b8b7 as a draw:
// all moves score the same, so it plays the first one generated, the repetition.
// This checks that the switch really turns the detection off.
func TestNoRepetitionIgnoresHistory(t *testing.T) {
	b, history := playFrom(t, winningShuffleFEN, winningShuffle...)
	m := NewAlphaBeta(AlphaBetaOptions{NoRepetition: true}).BestMove(*b, Limits{Depth: 1, History: history})
	if !makesThreefold(b, history, m) {
		t.Errorf("with NoRepetition, played %s; want the first generated move b8b7, which repeats", m)
	}
}
