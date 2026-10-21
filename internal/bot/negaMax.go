package bot

import (
	"goChessBot/internal/chess"
	"math"
)

type negaMaxBot struct{}

func (nM negaMaxBot) negaMax(b *chess.Board, limits Limits) int {
	if limits.Depth == 0 {
		return b.Evaluate()
	}
	maxScore := -math.MaxInt
	legalMoves := b.GenerateLegalMoves(nil)
	if len(legalMoves) == 0 {
		if b.InCheck() {
			return maxScore
		} else {
			return 0
		}
	}
	for _, move := range legalMoves {
		u := b.MakeMove(move)
		score := -nM.negaMax(b, Limits{limits.Depth - 1, limits.Time})
		b.UnmakeMove(move, u)
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}

func (nM negaMaxBot) BestMove(b chess.Board, limits Limits) chess.Move {
	maxScore := -math.MaxInt
	legalMoves := b.GenerateLegalMoves(nil)
	if len(legalMoves) == 0 {
		return chess.Move(0)
	}
	bestMove := legalMoves[0]
	depth := limits.Depth
	if depth == 0 {
		depth = 1
	}
	for _, move := range legalMoves {
		u := b.MakeMove(move)
		score := -nM.negaMax(&b, Limits{depth - 1, limits.Time})
		b.UnmakeMove(move, u)
		if score > maxScore {
			maxScore = score
			bestMove = move
		}
	}
	return bestMove
}
