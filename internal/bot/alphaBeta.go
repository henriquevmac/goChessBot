package bot

import (
	"goChessBot/internal/chess"
	"math"
)

const (
	infinite = math.MaxInt
)

type alphaBetaBot struct{}

func (aB alphaBetaBot) alphaBeta(alpha, beta, depth int, b *chess.Board) int {
	if depth == 0 {
		return b.Evaluate()
	}
	bestValue := -infinite
	legalMoves := b.GenerateLegalMoves(nil)
	if len(legalMoves) == 0 {
		if b.InCheck() {
			return bestValue
		} else {
			return 0
		}
	}
	for _, move := range legalMoves {
		u := b.MakeMove(move)
		score := -aB.alphaBeta(-beta, -alpha, depth-1, b)
		b.UnmakeMove(move, u)
		if score > bestValue {
			bestValue = score
			if score > alpha {
				alpha = score
			}
		}
		if score >= beta {
			return score
		}
	}
	return bestValue
}

func (aB alphaBetaBot) BestMove(b chess.Board, depth int) chess.Move {
	maxScore := -infinite
	legalMoves := b.GenerateLegalMoves(nil)
	if len(legalMoves) == 0 {
		return chess.Move(0)
	}
	bestMove := legalMoves[0]
	for _, move := range legalMoves {
		u := b.MakeMove(move)
		score := -aB.alphaBeta(-infinite, -maxScore, depth-1, &b)
		b.UnmakeMove(move, u)
		if score > maxScore {
			maxScore = score
			bestMove = move
		}
	}
	return bestMove
}
