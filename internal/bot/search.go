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

	bestValue := -infinite
	var moves []chess.Move
	start := len(s.moveStack)
	if !s.board.InCheck() {
		bestValue = s.board.Evaluate()

		if bestValue >= beta {
			return bestValue
		}
		if bestValue > alpha {
			alpha = bestValue
		}
		s.moveStack = s.board.GenerateCaptureMoves(s.moveStack)
		moves = s.moveStack[start:]
	} else {
		s.moveStack = s.board.GenerateLegalMoves(s.moveStack)
		moves = s.moveStack[start:]
		if len(moves) == 0 {
			return matedScore(s.ply)
		}
	}
	defer func() { s.moveStack = s.moveStack[:start] }()
	s.orderMoves(moves, chess.Move(0))

	for _, move := range moves {
		u := s.board.MakeMove(move)
		s.ply++
		score := -s.quiesce(-beta, -alpha)
		s.ply--
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
