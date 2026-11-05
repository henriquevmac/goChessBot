package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"goChessBot/internal/chess"
)

// run sends the commands to a new engine, one per line, and returns its output.
func run(t *testing.T, commands ...string) string {
	t.Helper()
	var out bytes.Buffer
	e := newEngine(&out)
	for _, c := range commands {
		if e.handle(c) {
			break
		}
	}
	return out.String()
}

// bestMove returns the move of the last "bestmove" line in the output.
func bestMove(t *testing.T, out string) string {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	last := lines[len(lines)-1]
	move, ok := strings.CutPrefix(last, "bestmove ")
	if !ok {
		t.Fatalf("the last line is %q, want a bestmove; output:\n%s", last, out)
	}
	return move
}

func TestHandshake(t *testing.T) {
	out := run(t, "uci", "isready")
	for _, want := range []string{"id name goChessBot", "option name Move Overhead type spin", "uciok", "readyok"} {
		if !strings.Contains(out, want) {
			t.Errorf("output has no %q:\n%s", want, out)
		}
	}
}

func TestQuitStopsReading(t *testing.T) {
	if out := run(t, "quit", "isready"); out != "" {
		t.Errorf("commands after quit were answered: %q", out)
	}
}

func TestSetPosition(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
		wantFEN string
		plies   int
	}{
		{"start position", "position startpos", startFEN, 0},
		{"moves from the start", "position startpos moves e2e4 e7e5 g1f3", "rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2", 3},
		{"castling", "position fen r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1 moves e1g1 e8c8", "2kr3r/8/8/8/8/8/8/R4RK1 w - - 2 2", 2},
		{"promotion", "position fen 8/P6k/8/8/8/8/8/K7 w - - 0 1 moves a7a8n", "N7/7k/8/8/8/8/8/K7 b - - 0 1", 1},
		{"en passant", "position fen 4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 1 moves e5d6", "4k3/8/3P4/8/8/8/8/4K3 b - - 0 1", 1},
		{"fen without moves", "position fen 4k3/8/8/8/8/8/8/4K2R b K - 3 40", "4k3/8/8/8/8/8/8/4K2R b K - 3 40", 0},
	} {
		e := newEngine(&bytes.Buffer{})
		e.handle(tc.command)
		if got := e.board.FEN(); got != tc.wantFEN {
			t.Errorf("%s: board is %s, want %s", tc.name, got, tc.wantFEN)
		}
		if len(e.history) != tc.plies {
			t.Errorf("%s: history has %d positions, want %d", tc.name, len(e.history), tc.plies)
		}
	}
}

// The history holds the position before every move, so the bot can see
// repetitions that happened before the current position.
func TestHistoryHasEveryEarlierPosition(t *testing.T) {
	e := newEngine(&bytes.Buffer{})
	e.handle("position startpos moves g1f3 g8f6 f3g1 f6g8")
	start, _ := chess.CreateBoard(startFEN)
	if len(e.history) != 4 || e.history[0] != start.Hash || e.board.Hash != start.Hash {
		t.Errorf("after the knights went out and back, history = %x and hash = %x; want 4 entries starting with %x and the start hash again", e.history, e.board.Hash, start.Hash)
	}
}

func TestIllegalMoveKeepsOldPosition(t *testing.T) {
	var out bytes.Buffer
	e := newEngine(&out)
	e.handle("position startpos moves e2e4")
	before := e.board.FEN()
	e.handle("position startpos moves e2e5")
	if e.board.FEN() != before {
		t.Errorf("an illegal move changed the board to %s", e.board.FEN())
	}
	if !strings.Contains(out.String(), "info string") {
		t.Errorf("no error reported for an illegal move: %q", out.String())
	}
}

func TestGoReturnsLegalMove(t *testing.T) {
	for _, command := range []string{"go movetime 50", "go depth 2", "go wtime 1000 btime 1000 winc 10 binc 10"} {
		out := run(t, "position startpos moves e2e4", command)
		b, _ := chess.CreateBoard("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1")
		if _, ok := findMove(b, bestMove(t, out)); !ok {
			t.Errorf("%s: bestmove %s is not legal for black", command, bestMove(t, out))
		}
	}
}

func TestGoFindsMate(t *testing.T) {
	// Scholar's mate: Qxf7#.
	out := run(t, "position fen r1bqkb1r/pppp1ppp/2n2n2/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4", "go depth 3")
	if got := bestMove(t, out); got != "h5f7" {
		t.Errorf("bestmove %s, want h5f7", got)
	}
}

func TestGoWithNoLegalMove(t *testing.T) {
	// Black is checkmated.
	out := run(t, "position fen r1bqkb1r/pppp1Qpp/2n2n2/4p3/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 0 4", "go movetime 50")
	if got := bestMove(t, out); got != "0000" {
		t.Errorf("bestmove %s, want 0000", got)
	}
}

func TestParseGo(t *testing.T) {
	got := parseGo(strings.Fields("wtime 60000 btime 30000 winc 1000 binc 500 movestogo 12 depth 6 movetime 250 infinite"))
	want := goParams{wtime: time.Minute, btime: 30 * time.Second, winc: time.Second, binc: 500 * time.Millisecond, movesToGo: 12, moveTime: 250 * time.Millisecond, depth: 6}
	if got != want {
		t.Errorf("parseGo = %+v, want %+v", got, want)
	}
}

func TestLimitsUseTheSideToMovesClock(t *testing.T) {
	p := goParams{wtime: time.Minute, btime: 10 * time.Second}
	white, black := p.limits(chess.White, defaultMoveOverhead), p.limits(chess.Black, defaultMoveOverhead)
	if white.Time != clockMoveTime(time.Minute, 0, 0, defaultMoveOverhead) || black.Time != clockMoveTime(10*time.Second, 0, 0, defaultMoveOverhead) {
		t.Errorf("white thinks %v and black %v; want them from their own clocks", white.Time, black.Time)
	}
}

func TestLimitsWithoutTime(t *testing.T) {
	if l := (goParams{depth: 5}).limits(chess.White, defaultMoveOverhead); l.Depth != 5 || l.Time != 0 {
		t.Errorf("go depth 5 gives %+v, want depth 5 without a time limit", l)
	}
	if l := (goParams{}).limits(chess.White, defaultMoveOverhead); l.Time != defaultMoveTime {
		t.Errorf("go without limits gives %+v, want %v", l, defaultMoveTime)
	}
}

func TestClockMoveTime(t *testing.T) {
	for _, tc := range []struct {
		remaining, inc time.Duration
		movesToGo      int
		want           time.Duration
	}{
		{time.Minute, 0, 0, 3*time.Second - defaultMoveOverhead},                       // a 20th of the clock
		{time.Minute, time.Second, 0, 4*time.Second - defaultMoveOverhead},             // plus the increment
		{time.Minute, 0, 10, 6*time.Second - defaultMoveOverhead},                      // moves to go given
		{time.Second, 10 * time.Second, 0, 500*time.Millisecond - defaultMoveOverhead}, // never more than half the clock
		{20 * time.Millisecond, 0, 0, minMoveTime},                                     // almost out of time
	} {
		if got := clockMoveTime(tc.remaining, tc.inc, tc.movesToGo, defaultMoveOverhead); got != tc.want {
			t.Errorf("clockMoveTime(%v, %v, %d) = %v, want %v", tc.remaining, tc.inc, tc.movesToGo, got, tc.want)
		}
	}
}

func TestSetMoveOverhead(t *testing.T) {
	var out bytes.Buffer
	e := newEngine(&out)
	e.handle("setoption name Move Overhead value 100")
	if e.moveOverhead != 100*time.Millisecond {
		t.Errorf("Move Overhead is %v, want 100ms", e.moveOverhead)
	}
	if l := parseGo([]string{"movetime", "1000"}).limits(chess.White, e.moveOverhead); l.Time != 900*time.Millisecond {
		t.Errorf("movetime 1000 with 100 ms overhead thinks %v, want 900ms", l.Time)
	}
	for _, bad := range []string{"setoption name Hash value 512", "setoption name Move Overhead value -5", "setoption name Move Overhead value lots", "setoption Move Overhead"} {
		out.Reset()
		e.handle(bad)
		if e.moveOverhead != 100*time.Millisecond || !strings.Contains(out.String(), "info string") {
			t.Errorf("%q: Move Overhead is %v and output %q; want it unchanged and an error reported", bad, e.moveOverhead, out.String())
		}
	}
}
