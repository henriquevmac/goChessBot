package chess

import "math/rand/v2"

var (
	zobristPiece    [32][64]uint64
	zobristBlack    uint64
	zobristCastling [16]uint64
	zobristEP       [8]uint64
)

func init() {
	r := rand.New(rand.NewPCG(1, 2))
	for p := range zobristPiece {
		for sq := range zobristPiece[p] {
			zobristPiece[p][sq] = r.Uint64()
		}
	}
	zobristBlack = r.Uint64()
	for cst := range zobristCastling {
		zobristCastling[cst] = r.Uint64()
	}
	for ep := range zobristEP {
		zobristEP[ep] = r.Uint64()
	}
}

func (b *Board) computeHash() uint64 {
	h := uint64(0)
	for idx, sq := range b.Square {
		if !sq.IsEmpty() {
			h ^= zobristPiece[sq][idx]
		}
	}

	if b.ToMove == Black {
		h ^= zobristBlack
	}

	h ^= zobristCastling[b.CastlingRights]

	if b.EnPassantSquare != -1 {
		h ^= zobristEP[b.EnPassantSquare%8]
	}

	return h
}

func (b *Board) IsThreeFoldRepetition(history []uint64) bool {
	count := 0
	stop := max(len(history)-b.HalfmoveClock, 0)
	for i := len(history) - 2; i >= stop; i -= 2 {
		if history[i] == b.Hash {
			count++
			if count == 2 {
				return true
			}
		}
	}
	return false
}
