package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"
	"slices"
	"strings"
	"time"

	"goChessBot/internal/bot"
	"goChessBot/internal/chess"
	"goChessBot/internal/match"
)

// An endgame is written as the winning side's pieces against a bare king, e.g.
// KQK or KRK. The positions are random, with the winning side to move, and every
// other one has Black as the winner.
type endgame struct {
	name   string
	pieces string // the winning side's pieces besides the king, as FEN letters: "Q", "R", "BN"
}

func parseEndgame(name string) (endgame, error) {
	if len(name) < 3 || name[0] != 'K' || name[len(name)-1] != 'K' {
		return endgame{}, fmt.Errorf("endgame %q: want the form KQK, KRK, KBNK, ...", name)
	}
	pieces := name[1 : len(name)-1]
	for _, p := range pieces {
		if !strings.ContainsRune("QRBN", p) {
			return endgame{}, fmt.Errorf("endgame %q: %q is not one of Q, R, B, N", name, p)
		}
	}
	return endgame{name: name, pieces: pieces}, nil
}

// randomPosition places the two kings and the winner's pieces on random squares
// until the position is legal: the kings are not next to each other and the
// losing king, whose side is not to move, is not in check.
func randomPosition(r *rand.Rand, eg endgame, whiteWins bool) string {
	for {
		var board [64]byte
		squares := r.Perm(64)
		winnerKing, loserKing := squares[0], squares[1]
		if kingsTouch(winnerKing, loserKing) {
			continue
		}
		winner, loser := strings.ToUpper, strings.ToLower
		if !whiteWins {
			winner, loser = strings.ToLower, strings.ToUpper
		}
		board[winnerKing] = winner("K")[0]
		board[loserKing] = loser("K")[0]
		for i, p := range eg.pieces {
			board[squares[2+i]] = winner(string(p))[0]
		}

		placement := fenPlacement(board)
		toMove, other := "w", "b"
		if !whiteWins {
			toMove, other = "b", "w"
		}
		check, err := chess.CreateBoard(fmt.Sprintf("%s %s - - 0 1", placement, other))
		if err != nil || check.InCheck() {
			continue
		}
		return fmt.Sprintf("%s %s - - 0 1", placement, toMove)
	}
}

func kingsTouch(a, b int) bool {
	ra, ca := a/8, a%8
	rb, cb := b/8, b%8
	return max(ra-rb, rb-ra) <= 1 && max(ca-cb, cb-ca) <= 1
}

// fenPlacement writes the board, a1 = 0, as the first field of a FEN.
func fenPlacement(board [64]byte) string {
	var sb strings.Builder
	for row := 7; row >= 0; row-- {
		empty := 0
		for col := 0; col < 8; col++ {
			p := board[row*8+col]
			if p == 0 {
				empty++
				continue
			}
			if empty > 0 {
				sb.WriteByte(byte('0' + empty))
				empty = 0
			}
			sb.WriteByte(p)
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
		if row > 0 {
			sb.WriteByte('/')
		}
	}
	return sb.String()
}

type job struct {
	eg        endgame
	fen       string
	whiteWins bool
}

type outcome struct {
	job
	game match.Game
}

func (o outcome) mated() bool {
	winner := chess.White
	if !o.whiteWins {
		winner = chess.Black
	}
	return o.game.Result == chess.Checkmate && o.game.Winner == winner
}

// describe says how a game that was not won ended.
func (o outcome) describe() string {
	switch {
	case o.game.Forfeit:
		return "forfeit"
	case o.game.Result == chess.Ongoing:
		return "move limit"
	case o.game.Result == chess.Checkmate:
		return "lost"
	default:
		return o.game.Result.String()
	}
}

func main() {
	endgames := flag.String("endgames", "KQK,KRK", "comma-separated endgames, each the winning side's pieces against a bare king")
	positions := flag.Int("positions", 50, "random positions per endgame")
	seed := flag.Uint64("seed", 1, "seed for the random positions; the same seed gives the same positions")
	maxPlies := flag.Int("maxplies", 200, "stop a game after this many half-moves")
	depth := flag.Int("depth", 0, "maximum search depth (0 = no depth limit)")
	moveTime := flag.Duration("movetime", 100*time.Millisecond, "maximum thinking time per move (0 = no time limit)")
	workers := flag.Int("workers", runtime.NumCPU(), "number of games played at the same time")
	verbose := flag.Bool("v", false, "print the starting position of every game that was not won")
	flag.Parse()

	var egs []endgame
	for _, name := range strings.Split(*endgames, ",") {
		eg, err := parseEndgame(strings.TrimSpace(name))
		if err != nil {
			log.Fatal(err)
		}
		egs = append(egs, eg)
	}

	r := rand.New(rand.NewPCG(*seed, 0))
	var jobs []job
	for _, eg := range egs {
		for i := range *positions {
			whiteWins := i%2 == 0
			jobs = append(jobs, job{eg: eg, fen: randomPosition(r, eg, whiteWins), whiteWins: whiteWins})
		}
	}

	fmt.Printf("alphaBeta against itself, %d positions per endgame, move limit %d, depth %d, movetime %v, seed %d, %d workers\n\n",
		*positions, *maxPlies, *depth, *moveTime, *seed, *workers)

	// Every game gets new bots, since a bot keeps its transposition table between moves.
	limits := bot.Limits{Depth: *depth, Time: *moveTime}
	queue := make(chan job)
	done := make(chan outcome)
	for range max(1, *workers) {
		go func() {
			for j := range queue {
				white, black := bot.NewAlphaBeta(bot.AlphaBetaOptions{}), bot.NewAlphaBeta(bot.AlphaBetaOptions{})
				g, err := match.Play(white, black, j.fen, *maxPlies, limits)
				if err != nil {
					log.Fatalf("%s: %v", j.fen, err)
				}
				done <- outcome{job: j, game: g}
			}
		}()
	}
	go func() {
		for _, j := range jobs {
			queue <- j
		}
		close(queue)
	}()

	results := map[string][]outcome{}
	for range jobs {
		o := <-done
		results[o.eg.name] = append(results[o.eg.name], o)
	}

	for _, eg := range egs {
		outcomes := results[eg.name]
		slices.SortFunc(outcomes, func(a, b outcome) int { return strings.Compare(a.fen, b.fen) })
		mates, plies := 0, 0
		failures := map[string]int{}
		for _, o := range outcomes {
			if o.mated() {
				mates++
				plies += len(o.game.Moves)
			} else {
				failures[o.describe()]++
			}
		}
		fmt.Printf("%-6s mated %d / %d (%.0f%%)", eg.name, mates, len(outcomes), 100*float64(mates)/float64(len(outcomes)))
		if mates > 0 {
			fmt.Printf(", %.1f half-moves to mate on average", float64(plies)/float64(mates))
		}
		fmt.Println()
		reasons := make([]string, 0, len(failures))
		for reason := range failures {
			reasons = append(reasons, reason)
		}
		slices.Sort(reasons)
		for _, reason := range reasons {
			fmt.Printf("         not won: %-22s %d\n", reason, failures[reason])
		}
		if *verbose {
			for _, o := range outcomes {
				if !o.mated() {
					fmt.Printf("         %-22s %s\n", o.describe(), o.fen)
				}
			}
		}
	}
}
