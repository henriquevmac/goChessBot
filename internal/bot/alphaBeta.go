package bot

import (
	"goChessBot/internal/chess"
	"math"
	"slices"
	"time"
)

const (
	infinite = math.MaxInt
)

type alphaBetaBot struct{}

func (s *searcher) alphaBeta(alpha, beta, depth int) int {
	if s.check() {
		return 0
	}
	stop := max(len(s.history)-s.board.HalfmoveClock, 0)
	for i := len(s.history) - 2; i >= stop; i -= 2 {
		if s.history[i] == s.board.Hash {
			return 0
		}
	}
	if depth == 0 {
		return s.quiesce(alpha, beta)
	}
	alphaOrig := alpha
	entry, found := s.tt.probe(s.board.Hash)
	ttBestMove := chess.Move(0)
	if found {
		if entry.depth >= depth {
			switch entry.flag {
			case ttExact:
				return entry.score
			case ttLower:
				if entry.score >= beta {
					return entry.score
				}
			case ttUpper:
				if entry.score <= alpha {
					return entry.score
				}
			}
		}
		ttBestMove = entry.bestMove
	}
	bestValue := -infinite
	bestMove := chess.Move(0)
	legalMoves := s.board.GenerateLegalMoves(nil)
	if len(legalMoves) == 0 {
		if s.board.InCheck() {
			return bestValue
		} else {
			return 0
		}
	}
	orderMoves(s.board, legalMoves, ttBestMove)
	for _, move := range legalMoves {
		s.history = append(s.history, s.board.Hash)
		u := s.board.MakeMove(move)
		score := -s.alphaBeta(-beta, -alpha, depth-1)
		s.history = s.history[:len(s.history)-1]
		s.board.UnmakeMove(move, u)
		if score > bestValue {
			bestValue = score
			bestMove = move
			if score > alpha {
				alpha = score
			}
		}
		if score >= beta {
			break
		}
	}
	if !s.stopped {
		flag := ttExact
		if bestValue <= alphaOrig {
			flag = ttUpper
		} else if bestValue >= beta {
			flag = ttLower
		}
		s.tt.store(ttEntry{key: s.board.Hash, bestMove: bestMove, depth: depth, score: bestValue, flag: flag})
	}
	return bestValue
}

func (aB alphaBetaBot) BestMove(b chess.Board, limits Limits) chess.Move {
	legalMoves := b.GenerateLegalMoves(nil)
	if len(legalMoves) == 0 {
		return chess.Move(0)
	} else if len(legalMoves) == 1 {
		return legalMoves[0]
	}
	bestMove := legalMoves[0]
	maxDepth := limits.Depth
	if maxDepth == 0 {
		if limits.Time == 0 {
			maxDepth = 1
		} else {
			maxDepth = infinite
		}
	}
	deadline := time.Time{}
	if limits.Time > 0 {
		deadline = time.Now().Add(limits.Time)
	}
	s := searcher{
		board:    &b,
		deadline: deadline,
		nodes:    0,
		stopped:  false,
		tt:       newTranspositionTable(18),
		history:  slices.Clone(limits.History),
	}
	orderMoves(s.board, legalMoves, chess.Move(0))
	for depth := 1; depth <= maxDepth; depth++ {
		maxScore := -infinite
		depthBest := bestMove
		for _, move := range legalMoves {
			s.history = append(s.history, s.board.Hash)
			u := b.MakeMove(move)
			score := -s.alphaBeta(-infinite, -maxScore, depth-1)
			s.history = s.history[:len(s.history)-1]
			b.UnmakeMove(move, u)
			if score > maxScore {
				maxScore = score
				depthBest = move
			}
			if s.stopped {
				break
			}
		}
		if !s.stopped {
			bestMove = depthBest
			if i := slices.Index(legalMoves, bestMove); i > 0 {
				copy(legalMoves[1:i+1], legalMoves[:i])
				legalMoves[0] = bestMove
			}
		} else {
			break
		}
	}
	return bestMove
}
