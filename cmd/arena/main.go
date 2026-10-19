package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"goChessBot/internal/bot"
	"goChessBot/internal/chess"
	"goChessBot/internal/match"
)

type tally struct {
	winsA, winsB, draws, unfinished, forfeitsA, forfeitsB int
	plies                                                 int
	results                                               map[chess.GameResult]int
}

var botTypes = map[string]bot.BotType{
	"random": bot.Random,
}

func newBot(name string) (bot.Bot, error) {
	bt, ok := botTypes[name]
	if !ok {
		return nil, fmt.Errorf("unknown bot %q", name)
	}
	return bot.New(bt), nil
}

type finishedGame struct {
	index    int
	aIsWhite bool
	game     match.Game
	elapsed  time.Duration
}

func playGame(i int, a, b bot.Bot, fen string, maxPlies int) finishedGame {
	aIsWhite := i%2 == 0
	white, black := a, b
	if !aIsWhite {
		white, black = b, a
	}
	start := time.Now()
	g, err := match.Play(white, black, fen, maxPlies)
	if err != nil {
		log.Fatal(err)
	}
	return finishedGame{index: i, aIsWhite: aIsWhite, game: g, elapsed: time.Since(start)}
}

func (t *tally) add(fg finishedGame) {
	g := fg.game
	t.plies += len(g.Moves)
	if !g.Forfeit {
		t.results[g.Result]++
	}
	if g.Winner == 0 {
		if g.Result == chess.Ongoing {
			t.unfinished++
		} else {
			t.draws++
		}
		return
	}
	aWon := (g.Winner == chess.White) == fg.aIsWhite
	if aWon {
		t.winsA++
		if g.Forfeit {
			t.forfeitsB++
		}
	} else {
		t.winsB++
		if g.Forfeit {
			t.forfeitsA++
		}
	}
}

func outcome(g match.Game, whiteName, blackName string) string {
	winner, loser := whiteName, blackName
	if g.Winner == chess.Black {
		winner, loser = blackName, whiteName
	}
	switch {
	case g.Forfeit:
		return fmt.Sprintf("%s wins (illegal move by %s)", winner, loser)
	case g.Result == chess.Checkmate:
		return fmt.Sprintf("%s wins by checkmate", winner)
	case g.Result == chess.Ongoing:
		return "unfinished (move limit)"
	default:
		return fmt.Sprintf("draw (%v)", g.Result)
	}
}

func percent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return 100 * float64(part) / float64(total)
}

func loadOpenings(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var fens []string
	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		if _, err := chess.CreateBoard(text); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, line, err)
		}
		fens = append(fens, text)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(fens) == 0 {
		return nil, fmt.Errorf("%s: no positions found", path)
	}
	return fens, nil
}

func main() {
	botA := flag.String("a", "random", "first bot")
	botB := flag.String("b", "random", "second bot")
	games := flag.Int("games", 10, "number of games")
	maxPlies := flag.Int("maxplies", 300, "stop a game after this many moves")
	fen := flag.String("fen", "", "starting position (default: the normal start)")
	verbose := flag.Bool("v", false, "also print the moves of every game")
	quiet := flag.Bool("q", false, "only print the summary")
	workers := flag.Int("workers", runtime.NumCPU(), "number of games played at the same time")
	openingsFile := flag.String("openings", "", "file with one FEN per line; each opening is played twice, colors swapped")
	flag.Parse()

	set := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		set[f.Name] = true
	})

	if *fen != "" && *openingsFile != "" {
		log.Fatal("use either -fen or -openings, not both")
	}

	if _, err := chess.CreateBoard(*fen); err != nil {
		log.Fatalf("invalid -fen: %v", err)
	}

	starts := []string{*fen}
	if *openingsFile != "" {
		startsTemp, err := loadOpenings(*openingsFile)
		if err != nil {
			log.Fatal(err)
		}
		starts = startsTemp
		if !set["games"] {
			*games = len(starts) * 2
		}
	}

	*workers = max(1, min(*workers, *games))

	a, err := newBot(*botA)
	if err != nil {
		log.Fatal(err)
	}
	b, err := newBot(*botB)
	if err != nil {
		log.Fatal(err)
	}
	nameA := *botA + " (A)"
	nameB := *botB + " (B)"

	fmt.Printf("%s vs %s, %d games, colors alternate, move limit %d, %d workers, openings: %s (%d positions)\n", nameA, nameB, *games, *maxPlies, *workers, *openingsFile, len(starts))
	if *fen != "" {
		fmt.Printf("start position: %s\n", *fen)
	}
	fmt.Println()

	// Workers take game numbers from jobs, play them, and send the results back.
	// Only this goroutine touches the tally and prints, so no locking is needed.
	jobs := make(chan int)
	done := make(chan finishedGame)
	start := time.Now()
	for w := 0; w < *workers; w++ {
		go func() {
			for i := range jobs {
				done <- playGame(i, a, b, starts[(i/2)%len(starts)], *maxPlies)
			}
		}()
	}
	go func() {
		for i := 0; i < *games; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Games finish out of order; hold them until every earlier game is printed.
	t := tally{results: map[chess.GameResult]int{}}
	numWidth := len(fmt.Sprint(*games))
	pending := map[int]finishedGame{}
	next := 0
	var gameTime time.Duration
	for range *games {
		fg := <-done
		pending[fg.index] = fg
		for {
			fg, ok := pending[next]
			if !ok {
				break
			}
			delete(pending, next)
			next++

			gameTime += fg.elapsed
			t.add(fg)
			whiteName, blackName := nameA, nameB
			if !fg.aIsWhite {
				whiteName, blackName = nameB, nameA
			}
			if !*quiet {
				fmt.Printf("game %*d  white: %-14s black: %-14s %-40s %4d plies  %v\n",
					numWidth, fg.index+1, whiteName, blackName, outcome(fg.game, whiteName, blackName),
					len(fg.game.Moves), fg.elapsed.Round(time.Microsecond))
			}
			if *verbose {
				uci := make([]string, len(fg.game.Moves))
				for j, m := range fg.game.Moves {
					uci[j] = m.String()
				}
				fmt.Printf("  moves: %s\n\n", strings.Join(uci, " "))
			}
		}
	}
	total := time.Since(start)

	scoreA := float64(t.winsA) + float64(t.draws)/2
	scoreB := float64(t.winsB) + float64(t.draws)/2
	finished := *games - t.unfinished

	fmt.Printf("\nSummary: %d games in %v (%v of game time, %.1f plies per game on average)\n",
		*games, total.Round(time.Millisecond), gameTime.Round(time.Millisecond), float64(t.plies)/float64(max(*games, 1)))
	// Score: 1 point per win, half a point per draw, out of the finished games.
	fmt.Printf("  %-20s %4d wins  (%5.1f%%)  score %.1f / %d finished  (%5.1f%%)\n",
		nameA, t.winsA, percent(t.winsA, *games), scoreA, finished, percent(int(2*scoreA), 2*finished))
	fmt.Printf("  %-20s %4d wins  (%5.1f%%)  score %.1f / %d finished  (%5.1f%%)\n",
		nameB, t.winsB, percent(t.winsB, *games), scoreB, finished, percent(int(2*scoreB), 2*finished))
	fmt.Printf("  %-20s %4d       (%5.1f%%)\n", "draws", t.draws, percent(t.draws, *games))
	fmt.Printf("  %-20s %4d       (%5.1f%%)\n", "unfinished", t.unfinished, percent(t.unfinished, *games))

	fmt.Println("\n  How games ended:")
	for _, r := range []chess.GameResult{
		chess.Checkmate, chess.Stalemate, chess.FiftyMoveRule,
		chess.InsufficientMaterial, chess.ThreeFoldRepetition, chess.Ongoing,
	} {
		if n := t.results[r]; n > 0 {
			label := r.String()
			if r == chess.Ongoing {
				label = "move limit"
			}
			fmt.Printf("    %-22s %4d\n", label, n)
		}
	}
	if t.forfeitsA+t.forfeitsB > 0 {
		fmt.Printf("    %-22s %4d  (%s: %d, %s: %d)\n", "illegal move (forfeit)",
			t.forfeitsA+t.forfeitsB, nameA, t.forfeitsA, nameB, t.forfeitsB)
	}
}
