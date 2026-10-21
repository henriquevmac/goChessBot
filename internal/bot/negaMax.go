package bot

import (
	"goChessBot/internal/chess"
	"math"
)

type negaMaxBot struct{}

func (nM negaMaxBot) negaMax(b *chess.Board, depth int) int {
	if depth == 0 {
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
		score := -nM.negaMax(b, depth-1)
		b.UnmakeMove(move, u)
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}

func (nM negaMaxBot) BestMove(b chess.Board, depth int) chess.Move {
	maxScore := -math.MaxInt
	legalMoves := b.GenerateLegalMoves(nil)
	if len(legalMoves) == 0 {
		return chess.Move(0)
	}
	bestMove := legalMoves[0]
	for _, move := range legalMoves {
		u := b.MakeMove(move)
		score := -nM.negaMax(&b, depth-1)
		b.UnmakeMove(move, u)
		if score > maxScore {
			maxScore = score
			bestMove = move
		}
	}
	return bestMove
}
