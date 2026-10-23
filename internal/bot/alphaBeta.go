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

type alphaBetaBot struct {
	options AlphaBetaOptions
}

func (s *searcher) alphaBeta(alpha, beta, depth int) int {
	if s.check() {
		return 0
	}
	if !s.options.NoRepetition {
		stop := max(len(s.history)-s.board.HalfmoveClock, 0)
		for i := len(s.history) - 2; i >= stop; i -= 2 {
			if s.history[i] == s.board.Hash {
				return 0
			}
		}
	}
	if depth == 0 {
		return s.quiesce(alpha, beta)
	}
	alphaOrig := alpha
	ttBestMove := chess.Move(0)
	entry, found := ttEntry{}, false
	if !s.options.NoTT {
		entry, found = s.tt.probe(s.board.Hash)
	}
	if found {
		score := scoreFromTT(int(entry.score), s.ply)
		if int(entry.depth) >= depth {
			switch entry.flag {
			case ttExact:
				return score
			case ttLower:
				if score >= beta {
					return score
				}
			case ttUpper:
				if score <= alpha {
					return score
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
			return matedScore(s.ply)
		} else {
			return 0
		}
	}
	s.orderMoves(legalMoves, ttBestMove)
	for _, move := range legalMoves {
		s.history = append(s.history, s.board.Hash)
		u := s.board.MakeMove(move)
		s.ply++
		score := -s.alphaBeta(-beta, -alpha, depth-1)
		s.ply--
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
			s.recordCutoff(move, depth)
			break
		}
	}
	if !s.stopped && !s.options.NoTT {
		flag := ttExact
		if bestValue <= alphaOrig {
			flag = ttUpper
		} else if bestValue >= beta {
			flag = ttLower
		}
		score := scoreToTT(bestValue, s.ply)
		s.tt.store(ttEntry{key: s.board.Hash, bestMove: bestMove, depth: int8(depth), score: int32(score), flag: flag})
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
	start := time.Now()
	if limits.Time > 0 {
		deadline = time.Now().Add(limits.Time)
	}
	s := searcher{
		board:    &b,
		deadline: deadline,
		nodes:    0,
		stopped:  false,
		history:  slices.Clone(limits.History),
		ply:      1,
		options:  aB.options,
	}
	if !aB.options.NoTT {
		s.tt = newTranspositionTable(18)
	}
	s.orderMoves(legalMoves, chess.Move(0))
	for depth := 1; depth <= maxDepth; depth++ {
		maxScore := -infinite
		depthBest := bestMove
		for _, move := range legalMoves {
			s.history = append(s.history, s.board.Hash)
			u := b.MakeMove(move)
			score := -s.alphaBeta(-infinite, -maxScore, depth-1)
			s.history = s.history[:len(s.history)-1]
			b.UnmakeMove(move, u)
			if s.stopped {
				break
			}
			if score > maxScore {
				maxScore = score
				depthBest = move
			}
		}
		bestMove = depthBest
		if i := slices.Index(legalMoves, bestMove); i > 0 {
			copy(legalMoves[1:i+1], legalMoves[:i])
			legalMoves[0] = bestMove
		}
		if s.stopped || !startNextDepth(time.Since(start), limits.Time) {
			break
		}
	}
	return bestMove
}
