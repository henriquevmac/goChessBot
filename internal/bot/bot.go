package bot

import (
	"goChessBot/internal/chess"
	"math/rand/v2"
	"time"
)

type BotType uint8

const (
	Random BotType = iota
	Greedy
	NegaMax
	AlphaBeta
)

type Limits struct {
	Depth   int
	Time    time.Duration
	History []uint64
}

type Bot interface {
	BestMove(b chess.Board, l Limits) chess.Move
}

type randomBot struct{}

func (r randomBot) BestMove(b chess.Board, l Limits) chess.Move {
	legalMoves := b.GenerateLegalMoves(nil)
	if len(legalMoves) == 0 {
		return chess.Move(0)
	}
	return legalMoves[rand.IntN(len(legalMoves))]
}

type greedyBot struct{}

func (g greedyBot) BestMove(b chess.Board, l Limits) chess.Move {
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
				value = chess.PieceValues[chess.Pawn]
				if materialWon < value {
					materialWon = value
					bestMove = move
				}
			} else {
				value = chess.PieceValues[b.Square[move.To()].Type()]
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
	case NegaMax:
		return negaMaxBot{}
	case AlphaBeta:
		return alphaBetaBot{}
	default:
		return nil
	}
}
