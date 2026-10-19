package bot

import (
	"goChessBot/internal/chess"
	"math/rand/v2"
)

type BotType uint8

const (
	Random BotType = iota
)

type Bot interface {
	BestMove(b chess.Board) chess.Move
}

type randomBot struct{}

func (r randomBot) BestMove(b chess.Board) chess.Move {
	legalMoves := b.GenerateLegalMoves(nil)
	if len(legalMoves) == 0 {
		return chess.Move(0)
	}
	return legalMoves[rand.IntN(len(legalMoves))]
}

func New(bt BotType) Bot {
	switch bt {
	case Random:
		return randomBot{}
	default:
		return nil
	}
}
