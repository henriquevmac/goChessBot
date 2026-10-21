package bot

import (
	"fmt"
	"slices"
	"sync"
	"testing"

	"goChessBot/internal/chess"
)

// These tests only check the contract every bot must keep, not how good its moves are,
// so they stay valid as bots are added or improved.

// allBotTypes lists every bot the contract tests run against. Add new bot types here.
var allBotTypes = []BotType{Random, Greedy, NegaMax, AlphaBeta}

// testLimits is what the contract tests pass to BestMove. The depth is kept low
// so the full self-play game stays fast; the contract does not depend on it.
var testLimits = Limits{Depth: 2}

// forEachBot runs test once per bot type, as a subtest named after the type.
func forEachBot(t *testing.T, test func(t *testing.T, bot Bot)) {
	t.Helper()
	for _, bt := range allBotTypes {
		t.Run(fmt.Sprintf("bot type %d", bt), func(t *testing.T) {
			bot := New(bt)
			if bot == nil {
				t.Fatalf("New(%d) returned nil", bt)
			}
			test(t, bot)
		})
	}
}

var contractFENs = []string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", // Kiwipete
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
	"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	"8/4P3/8/8/8/8/8/k3K3 w - - 0 1",    // promotion
	"4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2", // en passant available
	"4k3/8/8/8/8/8/4r3/4K3 w - - 0 1",   // in check
}

func mustBoard(t *testing.T, fen string) *chess.Board {
	t.Helper()
	b, err := chess.CreateBoard(fen)
	if err != nil {
		t.Fatalf("CreateBoard(%q): %v", fen, err)
	}
	return b
}

func TestBestMoveIsLegal(t *testing.T) {
	forEachBot(t, func(t *testing.T, bot Bot) {
		for _, fen := range contractFENs {
			b := mustBoard(t, fen)
			legal := b.GenerateLegalMoves(nil)
			before := *b

			m := bot.BestMove(*b, testLimits)

			if !slices.Contains(legal, m) {
				t.Errorf("%s: BestMove returned %s, which is not a legal move", fen, m)
			}
			if b.Hash != before.Hash || b.Square != before.Square || b.ToMove != before.ToMove {
				t.Errorf("%s: BestMove changed the caller's board", fen)
			}
		}
	})
}

func TestBestMoveForcedMove(t *testing.T) {
	// Black king on a8, white king on b6: a7 and b7 are covered, so Kb8 is the only move.
	fen := "k7/8/1K6/8/8/8/8/7R b - - 0 1"
	forEachBot(t, func(t *testing.T, bot Bot) {
		b := mustBoard(t, fen)
		legal := b.GenerateLegalMoves(nil)
		if len(legal) != 1 {
			t.Fatalf("test position should have exactly 1 legal move, has %d", len(legal))
		}
		if m := bot.BestMove(*b, testLimits); m != legal[0] {
			t.Errorf("BestMove = %s, want the only legal move %s", m, legal[0])
		}
	})
}

func TestBestMoveNoLegalMoves(t *testing.T) {
	positions := []struct {
		name string
		fen  string
	}{
		{"checkmate", "R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1"},
		{"stalemate", "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1"},
	}
	forEachBot(t, func(t *testing.T, bot Bot) {
		for _, pos := range positions {
			b := mustBoard(t, pos.fen)
			if m := bot.BestMove(*b, testLimits); m != chess.Move(0) {
				t.Errorf("%s: BestMove = %s, want chess.Move(0) when there are no legal moves", pos.name, m)
			}
		}
	})
}

// Several searches on different board copies at the same time must not interfere.
// Run with `go test -race ./internal/bot/` to let the race detector check this.
func TestBestMoveConcurrent(t *testing.T) {
	forEachBot(t, func(t *testing.T, bot Bot) {
		var wg sync.WaitGroup
		for _, fen := range contractFENs {
			b := mustBoard(t, fen)
			legal := b.GenerateLegalMoves(nil)
			wg.Add(1)
			go func(board chess.Board) {
				defer wg.Done()
				m := bot.BestMove(board, testLimits)
				if !slices.Contains(legal, m) {
					t.Errorf("%s: concurrent BestMove returned illegal move %s", fen, m)
				}
			}(*b)
		}
		wg.Wait()
	})
}

// The bot plays both sides of a whole game: every move must be legal, and the game
// must end in a proper result or reach the move limit without problems.
func TestBotPlaysAGame(t *testing.T) {
	if testing.Short() {
		t.Skip("full self-play game skipped in -short mode")
	}
	forEachBot(t, func(t *testing.T, bot Bot) {
		b := mustBoard(t, contractFENs[0])
		var history []uint64

		for ply := 0; ply < 200; ply++ {
			legal := b.GenerateLegalMoves(nil)
			result := b.ResultWithMoves(legal)
			if result == chess.Ongoing && b.IsThreeFoldRepetition(history) {
				result = chess.ThreeFoldRepetition
			}
			if result != chess.Ongoing {
				t.Logf("game ended after %d plies: %v", ply, result)
				return
			}

			m := bot.BestMove(*b, testLimits)
			if !slices.Contains(legal, m) {
				t.Fatalf("ply %d: BestMove returned illegal move %s", ply, m)
			}
			history = append(history, b.Hash)
			b.MakeMove(m)
		}
		t.Logf("no result after 200 plies (fine for a weak bot)")
	})
}

// gameWithHistory plays moves from the start and returns the board plus the history
// of earlier position hashes, exactly as match.Play and the UI build it.
func gameWithHistory(t *testing.T, moves ...string) (*chess.Board, []uint64) {
	t.Helper()
	b := mustBoard(t, contractFENs[0])
	var history []uint64
	for _, uci := range moves {
		var found chess.Move
		for _, m := range b.GenerateLegalMoves(nil) {
			if m.String() == uci {
				found = m
			}
		}
		if found == 0 {
			t.Fatalf("setup: %s is not legal", uci)
		}
		history = append(history, b.Hash)
		b.MakeMove(found)
	}
	return b, history
}

// With a game history, every bot must still return a legal move, including when
// the current position has already occurred before.
func TestBestMoveWithHistory(t *testing.T) {
	forEachBot(t, func(t *testing.T, bot Bot) {
		// The knights go out and back: the start position is on the board for the second time.
		b, history := gameWithHistory(t, "g1f3", "g8f6", "f3g1", "f6g8")
		legal := b.GenerateLegalMoves(nil)
		l := testLimits
		l.History = history
		if m := bot.BestMove(*b, l); !slices.Contains(legal, m) {
			t.Errorf("BestMove with a history returned %s, which is not a legal move", m)
		}
	})
}

// The history belongs to the caller. A bot must not change it, not even the spare
// capacity behind it, where an append from the bot would silently write.
func TestBestMoveDoesNotModifyHistory(t *testing.T) {
	forEachBot(t, func(t *testing.T, bot Bot) {
		b, history := gameWithHistory(t, "e2e4", "e7e5", "g1f3", "b8c6")
		const sentinel = 0xDEADBEEFDEADBEEF
		backing := make([]uint64, len(history), len(history)+64)
		copy(backing, history)
		for i := len(history); i < cap(backing); i++ {
			backing[:cap(backing)][i] = sentinel
		}
		before := slices.Clone(backing)

		l := testLimits
		l.History = backing
		bot.BestMove(*b, l)

		if !slices.Equal(backing, before) {
			t.Errorf("BestMove changed the history: got %x, want %x", backing, before)
		}
		for i, v := range backing[len(backing):cap(backing)] {
			if v != sentinel {
				t.Errorf("BestMove wrote into the history's spare capacity at index %d", len(backing)+i)
				break
			}
		}
	})
}
