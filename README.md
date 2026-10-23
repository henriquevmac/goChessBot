# goChessBot

A chess game written in Go, with a graphical board built on [Ebitengine](https://ebitengine.org/), a rules-correct chess engine, and the tooling to build and measure chess bots: play against a bot in the window, or run bot-vs-bot matches from the terminal.

## Status

Work in progress.

- [x] Board rendering with piece sprites
- [x] Loading positions from FEN (piece placement, side to move, castling rights, en passant square, move clocks)
- [x] Click-to-move input with highlighted destination squares
- [x] Compact 16-bit move encoding with UCI output (`e2e4`, `e7e8q`)
- [x] Precomputed move tables (distance to edge, knight, king and pawn attack targets)
- [x] Move generation
- [x] Attack and check detection
- [x] Make / unmake move
- [x] Legal move filtering, verified with perft
- [x] Game end detection (checkmate, stalemate, fifty-move rule, insufficient material, threefold repetition)
- [x] Zobrist hashing
- [x] Rule enforcement in the UI
- [x] Restart button; Escape or right click cancels a selected piece
- [x] Bot plumbing: a `Bot` interface, a random bot, and the bot playing Black in the UI
- [x] Headless matches and a parallel arena command for bot-vs-bot testing
- [x] Opening positions sampled from Lichess game databases
- [ ] A bot that plays well

## Requirements

- Go 1.27 or newer
- Ebitengine's system dependencies. On Linux this means a C compiler and the X11/OpenGL development libraries. See the [Ebitengine install guide](https://ebitengine.org/en/documents/install.html) for your platform.

## Running

From the project root:

```sh
go run .
```

The piece images are loaded from `./assets/chessPieces/` relative to the current directory, so run the game from the project root.

You play White against the bot. Click a piece to select it (its legal destinations are highlighted), then click the target square. Escape or a right click cancels the selection, and the Restart button below the board starts a new game.

## Bot matches

`cmd/arena` plays bots against each other without opening a window:

```sh
go run ./cmd/arena -a random -b random -games 100
go run ./cmd/arena -openings openings.txt -q
```

| Flag | Meaning |
|---|---|
| `-a`, `-b` | the two bots, by name (default `random`) |
| `-games` | number of games (default 10; with `-openings`, every opening twice) |
| `-openings` | file with one FEN per line; each position is played twice, once with each bot as White |
| `-fen` | play every game from this one position instead |
| `-maxplies` | stop unfinished games after this many half-moves (default 300) |
| `-depth` | maximum search depth for both bots (default 4; 0 = no depth limit) |
| `-movetime` | maximum thinking time per move for both bots, e.g. `100ms` (default 0 = no time limit) |
| `-adjudicate` | decide games that hit the move limit by material: a lead of at least this many centipawns wins, less is a draw (default 0 = off) |
| `-workers` | games played at the same time (default: one per CPU core) |
| `-q` / `-v` | only the summary / also print every game's moves in UCI notation |

Colors alternate between games. The summary shows each bot's wins and score (1 point per win, half per draw), the estimated Elo difference between them with a 95% margin, and how the games ended, including forfeits if a bot ever returns an illegal move.

### Comparing alpha-beta versions

Parts of the alpha-beta search can be switched off by adding them to its name, so a change can be measured against the full bot:

```sh
go run ./cmd/arena -openings openings.txt -a alphaBeta -b alphaBeta-noTT -depth 0 -movetime 100ms -adjudicate 300 -q
```

The switches are `noTT` (transposition table), `freshTT` (a new, empty table for every move instead of one kept for the whole game), `noKillers`, `noHistory` and `noRepetition`, and can be combined, e.g. `alphaBeta-noKillers-noHistory`. Use a time limit rather than a depth limit, since most of these make the search faster rather than different. Between such close versions a few hundred games give a margin of about ±40 Elo, so small differences need more games.

### Opening positions

Playing every game from the start position makes a deterministic bot repeat the same game, so matches are best run from varied positions. `scripts/pgn_to_fens.py` samples them from a PGN file, such as a [Lichess database](https://database.lichess.org/) download (`.pgn` or `.pgn.zst`):

```sh
python3 -m venv .venv
.venv/bin/pip install chess zstandard
.venv/bin/python scripts/pgn_to_fens.py lichess_db.pgn -o openings.txt -n 500 --min-elo 1800
```

It picks one position per game between half-moves 8 and 20 (`--min-ply`, `--max-ply`), skips variants and games that are over, and keeps an evenly spread random sample. Run it with `-h` for all options.

### Adding a bot

1. Add a `BotType` constant and return the new bot from `bot.New` in `internal/bot/bot.go`.
2. Add the type to `allBotTypes` in `internal/bot/bot_test.go`, so the contract tests (legal moves only, no move when the game is over, safe to run in parallel) check it.
3. Add its name to `botTypes` in `cmd/arena/main.go`, then measure it: `go run ./cmd/arena -a newbot -b random -openings openings.txt`.

## Testing

```sh
go test ./...
go test -race ./internal/bot/ ./internal/match/   # also check bots for data races
```

The rules, bot and match packages have no dependency on Ebitengine, so their tests run without opening a window.

Move generation is verified with [perft](https://www.chessprogramming.org/Perft_Results): the number of possible move sequences from seven standard positions is compared against the known results, up to about 4.8 million positions per test. Add `-short` to skip the deepest runs.

## Benchmarks

`internal/chess/bench_test.go` measures the engine's hot paths: move generation, legal move filtering, make/unmake, attack detection, perft and game result detection, each on several positions.

```sh
go test ./internal/chess/ -run '^$' -bench . -benchmem            # all benchmarks (about 30 seconds)
go test ./internal/chess/ -run '^$' -bench Perft                  # only the perft benchmarks
go test ./internal/chess/ -run '^$' -bench 'MakeUnmake/start'     # one benchmark, one position
```

`-run '^$'` skips the normal tests, `-bench` takes a pattern matching benchmark names, and `-benchmem` adds memory allocations per operation. In the output, `ns/op` is the time for one run (lower is better), `allocs/op` the number of memory allocations, and the perft benchmarks also report `nodes/s`, positions searched per second.

### Comparing before and after a change

Single runs vary a little, so measure several times and compare with [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```sh
go install golang.org/x/perf/cmd/benchstat@latest
mkdir -p benchmarks

go test ./internal/chess/ -run '^$' -bench . -benchmem -count 6 | tee benchmarks/before.txt
# ...change the engine...
go test ./internal/chess/ -run '^$' -bench . -benchmem -count 6 | tee benchmarks/after.txt
benchstat benchmarks/before.txt benchmarks/after.txt
```

The `benchmarks/` folder is in `.gitignore`, so results stay local.

`-count 6` repeats every benchmark six times (about 3 minutes in total; add `-benchtime 300ms` for a quicker, slightly noisier run). `benchstat` shows the change for each benchmark as a percentage and marks differences too small to be meaningful with `~`.

## Project structure

```
.
├── main.go                 Entry point: loads assets, creates the board, starts the game
├── assets/chessPieces/     Piece sprites (w/b + K, Q, R, B, Kn, P)
├── cmd/arena/              Command-line bot-vs-bot matches
├── scripts/pgn_to_fens.py  Samples opening positions (FENs) from a PGN file
└── internal/
    ├── chess/              Chess rules, with no graphics dependencies; tests in *_test.go
    │   ├── piece.go        Piece and color encoding
    │   ├── board.go        Board state and FEN parsing
    │   ├── move.go         16-bit move encoding and UCI notation
    │   ├── directions.go   Precomputed direction and target tables
    │   ├── movegen.go      Pseudo-legal move generation, including castling
    │   ├── attack.go       Attack detection
    │   ├── makemove.go     Make / unmake move with an undo record
    │   ├── legal.go        Legal move filtering, perft, finding a move from two squares
    │   ├── gameend.go      Check, checkmate, stalemate and draw detection
    │   └── zobrist.go      Zobrist hashing and threefold repetition
    ├── bot/                The Bot interface and the bots; contract tests for every bot
    ├── match/              Plays one game between two bots, without any UI
    └── ui/                 Ebitengine front end
        ├── game.go         Game loop, bot moves via a goroutine, restart, result display
        ├── board.go        Drawing pieces and move highlights
        └── pieces.go       Sprite loading and click-to-move input
```

## Design notes

**Pieces** are stored as a single integer: the lowest 3 bits hold the piece type (King = 1, Pawn = 2, Knight = 3, Bishop = 4, Rook = 5, Queen = 6) and bits 3–4 hold the color (White = 8, Black = 16). For example, a white queen is `8 | 6 = 14`.

**Squares** are indexed 0–63, with a1 = 0, h1 = 7, a2 = 8 and h8 = 63. Index 0 is White's bottom-left corner; the UI flips the ranks when drawing so White appears at the bottom.

**Moves** are packed into a `uint16`:

```
 15  14  13  12 | 11  10   9   8   7   6 |  5   4   3   2   1   0
 └─── flag ────┘ └──────── to ──────────┘ └──────── from ────────┘
```

The 4-bit flag marks special moves (double pawn push, castling, captures, en passant and promotions). Bit 2 of the flag means capture, and bit 3 means promotion. `Move(0)` is used as "no move".

**Legal moves** are found by generating pseudo-legal moves, making each one, and keeping it only if the mover's king is not attacked afterwards. `MakeMove` returns an undo record that `UnmakeMove` uses to restore the position exactly.

**Positions** are identified by a 64-bit Zobrist hash, updated incrementally in `MakeMove`. It covers the pieces, side to move, castling rights and en passant file, but not the move clocks. Threefold repetition compares the current hash with earlier positions back to the last capture or pawn move.

**Bots** implement `BestMove(b chess.Board) chess.Move`. The board is passed by value, so a bot searches its own copy and can run in a goroutine while the UI keeps drawing, or many games can run in parallel in the arena.
