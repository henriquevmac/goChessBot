package bot

import (
	"goChessBot/internal/chess"
	"math/rand/v2"
)

type BotType uint8

var pieceValues = [7]int{
	chess.Pawn:   100,
	chess.Knight: 320,
	chess.Bishop: 330,
	chess.Rook:   500,
	chess.Queen:  900,
}

const (
	Random BotType = iota
	Greedy
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

type greedyBot struct{}

func (g greedyBot) BestMove(b chess.Board) chess.Move {
	legalMoves := b.GenerateLegalMoves(nil)
	if len(legalMoves) == 0 {
		return chess.Move(0)
	}
	materialWon := 0
	bestMove := chess.Move(0)
	for _, move := range legalMoves {
		if move.IsCapture() {
			value := 0
			if move.Flag() == chess.EnPassant {
				value = pieceValues[chess.Pawn]
				if materialWon < value {
					materialWon = value
					bestMove = move
				}
			} else {
				value = pieceValues[b.Square[move.To()].Type()]
				if materialWon < value {
					materialWon = value
					bestMove = move
				}
			}
		}
	}
	if materialWon == 0 {
		bestMove = legalMoves[rand.IntN(len(legalMoves))]
	}
	return bestMove
}

func New(bt BotType) Bot {
	switch bt {
	case Random:
		return randomBot{}
	case Greedy:
		return greedyBot{}
	default:
		return nil
	}
}
