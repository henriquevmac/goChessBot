// Command uci runs the alpha-beta bot as a UCI engine, so chess GUIs and
// lichess-bot can play against it. It reads commands on stdin and answers on
// stdout.
//
// The search runs in the same goroutine that reads commands, so "stop" cannot
// cut a search short: every "go" ends after its own time or depth limit.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"goChessBot/internal/bot"
	"goChessBot/internal/chess"
)

const startFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

const (
	// When the GUI does not say how many moves are left until the next time
	// control, the remaining time is split as if this many were.
	defaultMovesToGo = 20
	// Kept back from every move for the time it takes to send the move, unless
	// the GUI sets the "Move Overhead" option.
	defaultMoveOverhead = 50 * time.Millisecond
	maxMoveOverhead     = 5 * time.Second
	// The least time a move gets, however little is left on the clock.
	minMoveTime = 10 * time.Millisecond
	// The time "go infinite", or a "go" without any limit, searches for.
	defaultMoveTime = 5 * time.Second
)

type engine struct {
	out          io.Writer
	bot          bot.Bot
	board        *chess.Board
	history      []uint64 // hashes of the positions before each move, as match.Play keeps them
	moveOverhead time.Duration
}

func newEngine(out io.Writer) *engine {
	e := &engine{out: out, moveOverhead: defaultMoveOverhead}
	e.newGame()
	return e
}

// newGame starts with a new bot, so the transposition table of one game does
// not carry over into the next.
func (e *engine) newGame() {
	e.bot = bot.NewAlphaBeta(bot.AlphaBetaOptions{})
	e.board, _ = chess.CreateBoard(startFEN)
	e.history = nil
}

// handle runs one command and reports whether the engine should quit.
func (e *engine) handle(line string) (quit bool) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	switch fields[0] {
	case "uci":
		fmt.Fprintln(e.out, "id name goChessBot")
		fmt.Fprintln(e.out, "id author Henrique Machado")
		fmt.Fprintf(e.out, "option name Move Overhead type spin default %d min 0 max %d\n", defaultMoveOverhead.Milliseconds(), maxMoveOverhead.Milliseconds())
		fmt.Fprintln(e.out, "uciok")
	case "isready":
		fmt.Fprintln(e.out, "readyok")
	case "setoption":
		if err := e.setOption(fields[1:]); err != nil {
			fmt.Fprintf(e.out, "info string %v\n", err)
		}
	case "ucinewgame":
		e.newGame()
	case "position":
		if err := e.setPosition(fields[1:]); err != nil {
			fmt.Fprintf(e.out, "info string %v\n", err)
		}
	case "go":
		e.goSearch(fields[1:])
	case "quit":
		return true
	}
	return false
}

// setOption handles "setoption name <name> value <value>". The only option is
// Move Overhead, in milliseconds.
func (e *engine) setOption(args []string) error {
	line := strings.Join(args, " ")
	rest, ok := strings.CutPrefix(line, "name ")
	if !ok {
		return fmt.Errorf("setoption: want name ... value ..., got %q", line)
	}
	name, value, _ := strings.Cut(rest, " value ")
	if !strings.EqualFold(strings.TrimSpace(name), "Move Overhead") {
		return fmt.Errorf("setoption: unknown option %q", name)
	}
	ms, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || ms < 0 || time.Duration(ms)*time.Millisecond > maxMoveOverhead {
		return fmt.Errorf("setoption: Move Overhead must be 0 to %d milliseconds, got %q", maxMoveOverhead.Milliseconds(), value)
	}
	e.moveOverhead = time.Duration(ms) * time.Millisecond
	return nil
}

// setPosition handles "position startpos [moves ...]" and
// "position fen <fen> [moves ...]".
func (e *engine) setPosition(args []string) error {
	fen := startFEN
	var moves []string
	switch {
	case len(args) > 0 && args[0] == "startpos":
		args = args[1:]
	case len(args) > 0 && args[0] == "fen":
		end := len(args)
		for i, a := range args {
			if a == "moves" {
				end = i
				break
			}
		}
		fen = strings.Join(args[1:end], " ")
		args = args[end:]
	default:
		return fmt.Errorf("position: want startpos or fen, got %q", strings.Join(args, " "))
	}
	if len(args) > 0 && args[0] == "moves" {
		moves = args[1:]
	}

	board, err := chess.CreateBoard(fen)
	if err != nil {
		return fmt.Errorf("position: %v", err)
	}
	var history []uint64
	for _, uci := range moves {
		m, ok := findMove(board, uci)
		if !ok {
			return fmt.Errorf("position: %s is not a legal move in %s", uci, board.FEN())
		}
		history = append(history, board.Hash)
		board.MakeMove(m)
	}
	e.board, e.history = board, history
	return nil
}

// findMove finds the legal move written in UCI notation, e.g. e2e4 or e7e8q.
func findMove(b *chess.Board, uci string) (chess.Move, bool) {
	for _, m := range b.GenerateLegalMoves(nil) {
		if m.String() == uci {
			return m, true
		}
	}
	return chess.Move(0), false
}

// goParams are the parts of a "go" command the engine uses.
type goParams struct {
	wtime, btime, winc, binc time.Duration
	movesToGo                int
	moveTime                 time.Duration
	depth                    int
}

func parseGo(args []string) goParams {
	var p goParams
	for i := 0; i+1 < len(args); i++ {
		n, err := strconv.Atoi(args[i+1])
		if err != nil {
			continue
		}
		ms := time.Duration(n) * time.Millisecond
		switch args[i] {
		case "wtime":
			p.wtime = ms
		case "btime":
			p.btime = ms
		case "winc":
			p.winc = ms
		case "binc":
			p.binc = ms
		case "movestogo":
			p.movesToGo = n
		case "movetime":
			p.moveTime = ms
		case "depth":
			p.depth = n
		default:
			continue
		}
		i++
	}
	return p
}

// limits turns a "go" command into search limits for the side to move,
// keeping overhead back from every move for sending it.
func (p goParams) limits(toMove chess.PieceId, overhead time.Duration) bot.Limits {
	remaining, inc := p.wtime, p.winc
	if toMove == chess.Black {
		remaining, inc = p.btime, p.binc
	}
	switch {
	case p.moveTime > 0:
		return bot.Limits{Depth: p.depth, Time: max(p.moveTime-overhead, minMoveTime)}
	case remaining > 0:
		return bot.Limits{Depth: p.depth, Time: clockMoveTime(remaining, inc, p.movesToGo, overhead)}
	case p.depth > 0:
		return bot.Limits{Depth: p.depth}
	default:
		return bot.Limits{Time: defaultMoveTime}
	}
}

// clockMoveTime decides how long to think when playing on a clock: an even
// share of the remaining time plus the increment, but never more than
// half of what is left.
func clockMoveTime(remaining, inc time.Duration, movesToGo int, overhead time.Duration) time.Duration {
	if movesToGo <= 0 {
		movesToGo = defaultMovesToGo
	}
	t := remaining/time.Duration(movesToGo) + inc
	t = min(t, remaining/2) - overhead
	return max(t, minMoveTime)
}

func (e *engine) goSearch(args []string) {
	limits := parseGo(args).limits(e.board.ToMove, e.moveOverhead)
	limits.History = e.history
	m := e.bot.BestMove(*e.board, limits)
	if m == chess.Move(0) {
		// No legal move: the game is over. UCI asks for a null move here.
		fmt.Fprintln(e.out, "bestmove 0000")
		return
	}
	fmt.Fprintf(e.out, "bestmove %s\n", m)
}

func main() {
	e := newEngine(os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if e.handle(scanner.Text()) {
			return
		}
	}
}
