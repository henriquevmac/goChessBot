#!/usr/bin/env python3
"""Pick random positions from a PGN file and write them as FENs, one per line.

For every game, one position is chosen at a random ply (half-move) within a range,
by default from move 4 to move 10 of the opening. Positions are sampled evenly
across the whole file (reservoir sampling), so even a multi-gigabyte Lichess
database can be streamed without loading it into memory.

Lichess databases come compressed as .pgn.zst. Either decompress them first
(`zstd -d file.pgn.zst`) or install the `zstandard` package and pass the .zst file.

Requires: pip install chess   (and optionally: pip install zstandard)

Examples:
    python3 scripts/pgn_to_fens.py games.pgn -o openings.txt -n 500
    python3 scripts/pgn_to_fens.py lichess_db.pgn.zst -o openings.txt -n 1000 --min-elo 2000 --max-games 200000
"""

import argparse
import io
import random
import sys

import chess
import chess.pgn


def open_pgn(path):
    """Open a .pgn or .pgn.zst file as a text stream."""
    if path == "-":
        return sys.stdin
    if path.endswith(".zst"):
        try:
            import zstandard
        except ImportError:
            sys.exit("reading .zst files needs the zstandard package (pip install zstandard), "
                     "or decompress first with: zstd -d " + path)
        raw = open(path, "rb")
        reader = zstandard.ZstdDecompressor().stream_reader(raw)
        return io.TextIOWrapper(reader, encoding="utf-8", errors="replace")
    return open(path, encoding="utf-8", errors="replace")


def rating(headers, key):
    try:
        return int(headers.get(key, ""))
    except ValueError:
        return None


def accept_game(headers, args):
    """Skip games that would not give useful standard-chess positions."""
    if headers.get("Variant", "Standard") not in ("Standard", "From Position") or "FEN" in headers:
        return False  # only games that start from the normal starting position
    if args.min_elo:
        white, black = rating(headers, "WhiteElo"), rating(headers, "BlackElo")
        if white is None or black is None or min(white, black) < args.min_elo:
            return False
    return True


def pick_position(game, args, rng):
    """Play the game to a random ply in [min_ply, max_ply] and return that board, or None."""
    moves = list(game.mainline_moves())
    last = min(args.max_ply, len(moves))
    if last < args.min_ply:
        return None  # game too short
    target = rng.randint(args.min_ply, last)

    board = game.board()
    for move in moves[:target]:
        board.push(move)
    if board.is_game_over(claim_draw=False):
        return None  # no point starting a game from a finished position
    return board


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("pgn", help="input .pgn or .pgn.zst file ('-' for stdin)")
    parser.add_argument("-o", "--output", default="-", help="output file (default: stdout)")
    parser.add_argument("-n", "--count", type=int, default=200, help="number of positions to keep (default: 200)")
    parser.add_argument("--min-ply", type=int, default=8, help="earliest half-move to sample (default: 8)")
    parser.add_argument("--max-ply", type=int, default=20, help="latest half-move to sample (default: 20)")
    parser.add_argument("--min-elo", type=int, default=0, help="only games where both players are rated at least this")
    parser.add_argument("--max-games", type=int, default=0, help="stop after reading this many games (0 = all)")
    parser.add_argument("--seed", type=int, default=None, help="random seed, for reproducible output")
    args = parser.parse_args()

    if args.min_ply < 0 or args.max_ply < args.min_ply:
        parser.error("need 0 <= --min-ply <= --max-ply")

    rng = random.Random(args.seed)
    sample = []      # reservoir of FENs
    seen = set()     # avoid duplicate positions
    candidates = 0   # distinct positions offered to the reservoir
    games_read = 0

    with open_pgn(args.pgn) as pgn:
        while True:
            game = chess.pgn.read_game(pgn)
            if game is None:
                break
            games_read += 1
            if games_read % 10000 == 0:
                print(f"read {games_read} games, {candidates} candidate positions", file=sys.stderr)

            if accept_game(game.headers, args):
                board = pick_position(game, args, rng)
                if board is not None:
                    # Reset the clocks: the position becomes the start of a new game.
                    board.halfmove_clock = 0
                    fen = board.fen(en_passant="fen")
                    if fen not in seen:
                        seen.add(fen)
                        candidates += 1
                        # Reservoir sampling: every candidate ends up kept with equal probability.
                        if len(sample) < args.count:
                            sample.append(fen)
                        else:
                            j = rng.randrange(candidates)
                            if j < args.count:
                                sample[j] = fen

            if args.max_games and games_read >= args.max_games:
                break

    rng.shuffle(sample)
    out = sys.stdout if args.output == "-" else open(args.output, "w", encoding="utf-8")
    with out:
        out.write(f"# {len(sample)} positions from {args.pgn}, plies {args.min_ply}-{args.max_ply}\n")
        for fen in sample:
            out.write(fen + "\n")

    print(f"read {games_read} games, kept {len(sample)} of {candidates} distinct positions", file=sys.stderr)
    if len(sample) < args.count:
        print(f"warning: wanted {args.count} positions but only found {len(sample)}", file=sys.stderr)


if __name__ == "__main__":
    main()
