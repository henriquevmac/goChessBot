package bot

import (
	"goChessBot/internal/chess"
	"time"
)

const mateValue = 1_000_000
const maxPly = 128

type searcher struct {
	board         *chess.Board
	deadline      time.Time
	nodes         int
	stopped       bool
	tt            *transpositionTable
	history       []uint64
	ply           int
	killers       [maxPly][2]chess.Move
	historyScores [64][64]int
	options       AlphaBetaOptions
	moveStack     []chess.Move
}

func matedScore(ply int) int {
	return -mateValue + ply
}

func startNextDepth(elapsed, limit time.Duration) bool {
	if limit == 0 {
		return true
	} else {
		return elapsed <= limit/2
	}
}

func (s *searcher) recordCutoff(m chess.Move, depth int) {
	if m.IsCapture() || m.IsPromotion() {
		return
	}
	if !s.options.NoHistory {
		s.historyScores[m.From()][m.To()] += depth * depth
	}
	if s.options.NoKillers || s.ply >= maxPly {
		return
	}
	k := &s.killers[s.ply]
	if k[0] != m {
		k[1] = k[0]
		k[0] = m
	}
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

	start := len(s.moveStack)
	s.moveStack = s.board.GenerateCaptureMoves(s.moveStack)
	captures := s.moveStack[start:]
	defer func() { s.moveStack = s.moveStack[:start] }()
	s.orderMoves(captures, chess.Move(0))
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
