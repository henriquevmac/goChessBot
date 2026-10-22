package bot

import (
	"goChessBot/internal/chess"
	"time"
)

const mateValue = 1_000_000

type searcher struct {
	board    *chess.Board
	deadline time.Time
	nodes    int
	stopped  bool
	tt       *transpositionTable
	history  []uint64
	ply      int
}

func matedScore(ply int) int {
	return -mateValue + ply
}

func (s *searcher) quiesce(alpha, beta int) int {
	if s.check() {
		return 0
	}
	staticEval := s.board.Evaluate()

	bestValue := staticEval
	if bestValue >= beta {
		return bestValue
	}
	if bestValue > alpha {
		alpha = bestValue
	}

	captures := s.board.GenerateCaptureMoves(nil)
	orderMoves(s.board, captures, chess.Move(0))
	for _, move := range captures {
		u := s.board.MakeMove(move)
		score := -s.quiesce(-beta, -alpha)
		s.board.UnmakeMove(move, u)

		if score >= beta {
			return score
		}
		if score > bestValue {
			bestValue = score
		}
		if score > alpha {
			alpha = score
		}
	}
	return bestValue
}

func (s *searcher) hasDeadline() bool {
	return !s.deadline.IsZero()
}

func (s *searcher) check() bool {
	s.nodes++
	if s.nodes&1023 == 0 && s.hasDeadline() && time.Now().After(s.deadline) {
		s.stopped = true
	}
	return s.stopped
}
