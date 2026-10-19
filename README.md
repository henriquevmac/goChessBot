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
| `-workers` | games played at the same time (default: one per CPU core) |
| `-q` / `-v` | only the summary / also print every game's moves in UCI notation |

Colors alternate between games. The summary shows each bot's wins and score (1 point per win, half per draw) and how the games ended, including forfeits if a bot ever returns an illegal move.

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
