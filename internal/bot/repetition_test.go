package bot

import (
	"fmt"
	"testing"

	"goChessBot/internal/chess"
)

// repetitionBots lists the bots that use Limits.History to recognize repetitions.
var repetitionBots = []BotType{AlphaBeta}

func forEachRepetitionBot(t *testing.T, test func(t *testing.T, bot Bot)) {
	t.Helper()
	for _, bt := range repetitionBots {
		t.Run(fmt.Sprintf("bot type %d", bt), func(t *testing.T) {
			test(t, New(bt))
		})
	}
}

// playFrom plays moves from fen and returns the board and the history of earlier
// positions, built the way match.Play and the UI build it.
func playFrom(t *testing.T, fen string, moves ...string) (*chess.Board, []uint64) {
	t.Helper()
	b := mustBoard(t, fen)
	var history []uint64
	for _, uci := range moves {
		var found chess.Move
		for _, m := range b.GenerateLegalMoves(nil) {
			if m.String() == uci {
				found = m
			}
		}
		if found == 0 {
			t.Fatalf("setup: %s is not legal in %s", uci, b.FEN())
		}
		history = append(history, b.Hash)
		b.MakeMove(found)
	}
	return b, history
}

// makesThreefold reports whether playing m from b would be a threefold repetition.
func makesThreefold(b *chess.Board, history []uint64, m chess.Move) bool {
	h := append(append([]uint64{}, history...), b.Hash)
	u := b.MakeMove(m)
	defer b.UnmakeMove(m, u)
	return b.IsThreeFoldRepetition(h)
}

// Both kings walk out and back twice. In the position reached, the side to move
// has exactly one move (the king stepping back) that repeats a position for the
// third time, which the referee would call a draw.

// Black is a queen up and to move: b8b7 would throw the win away as a draw. The
// black queen does the shuffling so that the repeating move is the first one
// generated: a bot that ignores the history, seeing every move as equally good,
// would play it.
const winningShuffleFEN = "4k3/1q6/8/8/8/8/8/4K3 w - - 0 1"

var winningShuffle = []string{"e1e2", "b7b8", "e2e1", "b8b7", "e1e2", "b7b8", "e2e1"}

// White has only its king against king and queen, and is to move: e2e1 saves
// the game with a draw by repetition.
const losingShuffleFEN = "q3k3/8/8/8/8/8/8/4K3 b - - 0 1"

var losingShuffle = []string{"e8f8", "e1e2", "f8e8", "e2e1", "e8f8", "e1e2", "f8e8"}

func TestRepetitionSetup(t *testing.T) {
	// Sanity checks of the positions the tests below rely on.
	for _, c := range []struct {
		name, fen string
		moves     []string
		repeat    string
	}{
		{"winning side", winningShuffleFEN, winningShuffle, "b8b7"},
		{"losing side", losingShuffleFEN, losingShuffle, "e2e1"},
	} {
		b, history := playFrom(t, c.fen, c.moves...)
		if b.IsThreeFoldRepetition(history) {
			t.Fatalf("%s: the game is already drawn by repetition", c.name)
		}
		repeats := 0
		for _, m := range b.GenerateLegalMoves(nil) {
			if makesThreefold(b, history, m) {
				repeats++
				if m.String() != c.repeat {
					t.Errorf("%s: unexpected repeating move %s", c.name, m)
				}
			}
		}
		if repeats != 1 {
			t.Errorf("%s: %d moves repeat the position, want exactly 1 (%s)", c.name, repeats, c.repeat)
		}
	}
}

// A side that is clearly winning must not agree to a draw by repetition.
func TestWinningSideAvoidsRepetition(t *testing.T) {
	forEachRepetitionBot(t, func(t *testing.T, bot Bot) {
		for depth := 1; depth <= 4; depth++ {
			b, history := playFrom(t, winningShuffleFEN, winningShuffle...)
			m := bot.BestMove(*b, Limits{Depth: depth, History: history})
			if makesThreefold(b, history, m) {
				t.Errorf("depth %d: played %s, a draw by repetition while a queen up", depth, m)
			}
		}
	})
}

// A side that is clearly losing should take the draw by repetition.
func TestLosingSideTakesRepetition(t *testing.T) {
	forEachRepetitionBot(t, func(t *testing.T, bot Bot) {
		for depth := 1; depth <= 4; depth++ {
			b, history := playFrom(t, losingShuffleFEN, losingShuffle...)
			m := bot.BestMove(*b, Limits{Depth: depth, History: history})
			if !makesThreefold(b, history, m) {
				t.Errorf("depth %d: played %s instead of e2e1, the draw by repetition, while a queen down", depth, m)
			}
		}
	})
}
