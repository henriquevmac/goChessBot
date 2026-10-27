package main

import (
	"math/rand/v2"
	"strings"
	"testing"

	"goChessBot/internal/chess"
)

func TestParseEndgame(t *testing.T) {
	for name, pieces := range map[string]string{"KQK": "Q", "KRK": "R", "KBNK": "BN", "KRRK": "RR"} {
		eg, err := parseEndgame(name)
		if err != nil || eg.pieces != pieces {
			t.Errorf("parseEndgame(%q) = %+v, %v; want pieces %q", name, eg, err, pieces)
		}
	}
	for _, name := range []string{"", "K", "KK", "QK", "KQ", "KPK", "KqK", "KQKK"} {
		if _, err := parseEndgame(name); err == nil {
			t.Errorf("parseEndgame(%q) gave no error", name)
		}
	}
}

func TestFenPlacement(t *testing.T) {
	var board [64]byte
	board[4] = 'K'  // e1
	board[7] = 'R'  // h1
	board[60] = 'k' // e8
	if got, want := fenPlacement(board), "4k3/8/8/8/8/8/8/4K2R"; got != want {
		t.Errorf("fenPlacement = %q, want %q", got, want)
	}
}

// Every generated position must be legal, have the right pieces, and have the
// winning side to move.
func TestRandomPositionsAreLegal(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 0))
	for _, name := range []string{"KQK", "KRK", "KBNK"} {
		eg, _ := parseEndgame(name)
		for i := range 200 {
			whiteWins := i%2 == 0
			fen := randomPosition(r, eg, whiteWins)
			b, err := chess.CreateBoard(fen)
			if err != nil {
				t.Fatalf("%s: %v", fen, err)
			}
			wantToMove := chess.White
			if !whiteWins {
				wantToMove = chess.Black
			}
			if b.ToMove != wantToMove {
				t.Errorf("%s: the winning side is not to move", fen)
			}

			placement := strings.Fields(fen)[0]
			winner, loser := strings.ToUpper, strings.ToLower
			if !whiteWins {
				winner, loser = strings.ToLower, strings.ToUpper
			}
			for _, p := range "KQRBN" {
				want := strings.Count(winner("K"+eg.pieces), winner(string(p))) + strings.Count(loser("K"), loser(string(p)))
				got := strings.Count(placement, winner(string(p))) + strings.Count(placement, loser(string(p)))
				if got != want {
					t.Errorf("%s: %d of %c, want %d", fen, got, p, want)
				}
			}

			if kingsTouch(b.KingPos[0], b.KingPos[1]) {
				t.Errorf("%s: the kings touch", fen)
			}
			other := "b"
			if !whiteWins {
				other = "w"
			}
			loserToMove, _ := chess.CreateBoard(strings.Replace(fen, " "+strings.Fields(fen)[1]+" ", " "+other+" ", 1))
			if loserToMove.InCheck() {
				t.Errorf("%s: the side not to move is in check", fen)
			}
		}
	}
}

func TestRandomPositionsRepeatWithSeed(t *testing.T) {
	eg, _ := parseEndgame("KRK")
	a, b := rand.New(rand.NewPCG(7, 0)), rand.New(rand.NewPCG(7, 0))
	for i := range 20 {
		if fa, fb := randomPosition(a, eg, i%2 == 0), randomPosition(b, eg, i%2 == 0); fa != fb {
			t.Fatalf("position %d differs with the same seed: %q and %q", i, fa, fb)
		}
	}
}
