# goChessBot

A chess game written in Go, with a graphical board built on [Ebitengine](https://ebitengine.org/). The goal is a complete, rules-correct chess implementation that a chess bot can be built on top of.

## Status

Work in progress.

- [x] Board rendering with piece sprites
- [x] Loading positions from FEN (piece placement, side to move, castling rights, en passant square, move clocks)
- [x] Drag and drop for moving pieces (no rule checking yet)
- [x] Compact 16-bit move encoding with UCI output (`e2e4`, `e7e8q`)
- [x] Precomputed move tables (distance to edge, knight, king and pawn attack targets)
- [x] Move generation
- [x] Attack and check detection
- [ ] Make / unmake move
- [ ] Legal move filtering, verified with perft
- [ ] Game end detection (checkmate, stalemate, draws)
- [ ] Rule enforcement in the UI
- [ ] The bot

## Requirements

- Go 1.27 or newer
- Ebitengine's system dependencies. On Linux this means a C compiler and the X11/OpenGL development libraries. See the [Ebitengine install guide](https://ebitengine.org/en/documents/install.html) for your platform.

## Running

From the project root:

```sh
go run .
```

The piece images are loaded from `./assets/chessPieces/` relative to the current directory, so run the game from the project root.

## Testing

```sh
go test ./...
```

The rules package (`internal/chess`) has no dependency on Ebitengine, so its tests run without opening a window.

## Project structure

```
.
├── main.go                 Entry point: loads assets, creates the board, starts the game
├── assets/chessPieces/     Piece sprites (w/b + K, Q, R, B, Kn, P)
└── internal/
    ├── chess/              Chess rules, with no graphics dependencies
    │   ├── piece.go        Piece and color encoding
    │   ├── board.go        Board state and FEN parsing
    │   ├── move.go         16-bit move encoding and UCI notation
    │   ├── directions.go   Precomputed direction and target tables
    │   ├── movegen.go      Move generation
    │   └── attack.go       Attack detection
    └── ui/                 Ebitengine front end
        ├── game.go         Game loop: Update, Draw, Layout
        ├── board.go        Drawing pieces on the board
        └── pieces.go       Sprite loading and drag and drop
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
