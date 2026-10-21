package bot

import (
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"goChessBot/internal/chess"
)

// timedBots lists the bots that honor Limits.Time (iterative deepening with a clock).
var timedBots = []BotType{AlphaBeta}

func forEachTimedBot(t *testing.T, test func(t *testing.T, bot Bot)) {
	t.Helper()
	for _, bt := range timedBots {
		t.Run(fmt.Sprintf("bot type %d", bt), func(t *testing.T) {
			test(t, New(bt))
		})
	}
}

// bestMoveWithin runs BestMove, failing the test instead of hanging if it takes
// longer than max.
func bestMoveWithin(t *testing.T, bot Bot, b chess.Board, l Limits, max time.Duration) (chess.Move, time.Duration) {
	t.Helper()
	done := make(chan chess.Move, 1)
	start := time.Now()
	go func() { done <- bot.BestMove(b, l) }()
	select {
	case m := <-done:
		return m, time.Since(start)
	case <-time.After(max):
		t.Fatalf("BestMove(%+v) did not return within %v", l, max)
		return 0, 0
	}
}

// Busy positions, where a search has plenty to do.
var busyFENs = []string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
}

// With only a time limit, the bot must return a legal move close to that time.
func TestTimeLimitIsRespected(t *testing.T) {
	const limit = 100 * time.Millisecond
	const slack = 100 * time.Millisecond // time to notice the deadline and unwind
	forEachTimedBot(t, func(t *testing.T, bot Bot) {
		for _, fen := range busyFENs {
			b := mustBoard(t, fen)
			legal := b.GenerateLegalMoves(nil)
			m, took := bestMoveWithin(t, bot, *b, Limits{Time: limit}, 10*time.Second)
			if !slices.Contains(legal, m) {
				t.Errorf("%s: BestMove returned %s, which is not a legal move", fen, m)
			}
			if took > limit+slack {
				t.Errorf("%s: BestMove took %v with a %v limit", fen, took.Round(time.Millisecond), limit)
			}
		}
	})
}

// Even a time limit too short to finish depth 1 must give a legal move.
func TestTinyTimeLimitStillGivesALegalMove(t *testing.T) {
	forEachTimedBot(t, func(t *testing.T, bot Bot) {
		for _, limit := range []time.Duration{time.Nanosecond, time.Microsecond, time.Millisecond} {
			for _, fen := range busyFENs {
				b := mustBoard(t, fen)
				legal := b.GenerateLegalMoves(nil)
				m, _ := bestMoveWithin(t, bot, *b, Limits{Time: limit}, 5*time.Second)
				if !slices.Contains(legal, m) {
					t.Errorf("%s, limit %v: BestMove returned %s, which is not a legal move", fen, limit, m)
				}
			}
		}
	})
}

// Limits{} (no depth and no time limit) must not search forever, for any bot.
func TestNoLimitsReturns(t *testing.T) {
	forEachBot(t, func(t *testing.T, bot Bot) {
		for _, fen := range busyFENs {
			b := mustBoard(t, fen)
			legal := b.GenerateLegalMoves(nil)
			m, _ := bestMoveWithin(t, bot, *b, Limits{}, 10*time.Second)
			if !slices.Contains(legal, m) {
				t.Errorf("%s: BestMove(Limits{}) returned %s, which is not a legal move", fen, m)
			}
		}
	})
}

// A depth limit without a time limit must really search that deep. This position
// has a mate in two (1.Rb7 or 1.Ra7, then mate on the back rank), which takes a
// depth-4 search to see and visits far more than a few thousand positions.
func TestDepthLimitSearchesFullDepth(t *testing.T) {
	fen := "7k/8/8/8/8/8/1R6/R5K1 w - - 0 1"
	forEachSearchBot(t, func(t *testing.T, bot Bot) {
		b := mustBoard(t, fen)
		first, _ := bestMoveWithin(t, bot, *b, Limits{Depth: 4}, 60*time.Second)
		u := b.MakeMove(first)
		defer b.UnmakeMove(first, u)

		// Whatever Black replies, White must have a mate next move.
		for _, reply := range b.GenerateLegalMoves(nil) {
			ur := b.MakeMove(reply)
			mate, _ := bestMoveWithin(t, bot, *b, Limits{Depth: 2}, 30*time.Second)
			um := b.MakeMove(mate)
			if r := b.Result(); r != chess.Checkmate {
				t.Errorf("after %s %s %s: result %v, want Checkmate (the first move %s does not force mate)",
					first, reply, mate, r, first)
			}
			b.UnmakeMove(mate, um)
			b.UnmakeMove(reply, ur)
		}
	})
}

// With both limits, the bot stops at whichever comes first: a small depth with
// a long time must return quickly.
func TestDepthLimitWithTimeLimit(t *testing.T) {
	forEachTimedBot(t, func(t *testing.T, bot Bot) {
		b := mustBoard(t, busyFENs[0])
		_, took := bestMoveWithin(t, bot, *b, Limits{Depth: 1, Time: time.Hour}, 10*time.Second)
		if took > time.Second {
			t.Errorf("Limits{Depth: 1, Time: 1h} took %v, want it to stop after depth 1", took.Round(time.Millisecond))
		}
	})
}

// Several timed searches at once must not share state. Run with -race.
func TestTimedSearchConcurrent(t *testing.T) {
	forEachTimedBot(t, func(t *testing.T, bot Bot) {
		var wg sync.WaitGroup
		for _, fen := range busyFENs {
			b := mustBoard(t, fen)
			legal := b.GenerateLegalMoves(nil)
			wg.Add(1)
			go func(board chess.Board) {
				defer wg.Done()
				if m := bot.BestMove(board, Limits{Time: 50 * time.Millisecond}); !slices.Contains(legal, m) {
					t.Errorf("%s: concurrent timed BestMove returned illegal move %s", fen, m)
				}
			}(*b)
		}
		wg.Wait()
	})
}
