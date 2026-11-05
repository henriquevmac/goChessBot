package bot

import (
	"goChessBot/internal/chess"
	"math"
	"slices"
	"sync"
	"time"
)

const (
	infinite = math.MaxInt
)

type alphaBetaBot struct {
	options AlphaBetaOptions
	mu      sync.Mutex
	tt      *transpositionTable
}

func (s *searcher) isRepetition() bool {
	if s.options.NoRepetition {
		return false
	}
	stop := max(len(s.history)-s.board.HalfmoveClock, 0)
	for i := len(s.history) - 2; i >= stop; i -= 2 {
		if s.history[i] == s.board.Hash {
			return true
		}
	}
	return false
}

func (s *searcher) probeTT(alpha, beta, depth int) (int, bool, chess.Move) {
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
				return score, true, entry.bestMove
			case ttLower:
				if score >= beta {
					return score, true, entry.bestMove
				}
			case ttUpper:
				if score <= alpha {
					return score, true, entry.bestMove
				}
			}
		}
		ttBestMove = entry.bestMove
	}
	return 0, false, ttBestMove
}

func (s *searcher) nullMoveCutoff(beta, depth int) bool {
	if s.options.NoNullMove || s.board.InCheck() || valuePieces(s.board) == 0 || depth < 3 || s.nullMove {
		return false
	}
	R := 2
	s.history = append(s.history, s.board.Hash)
	u := s.board.MakeNullMove()
	s.ply++
	s.nullMove = true
	score := -s.alphaBeta(-beta, -beta+1, depth-1-R)
	s.nullMove = false
	s.ply--
	s.history = s.history[:len(s.history)-1]
	s.board.UnmakeNullMove(u)
	return score >= beta
}

func (s *searcher) noMoveScore() int {
	if s.board.InCheck() {
		return matedScore(s.ply)
	}
	return 0
}

func (s *searcher) storeTT(bestValue, alphaOrig, beta, depth int, bestMove chess.Move) {
	if s.stopped || s.options.NoTT {
		return
	}
	flag := ttExact
	if bestValue <= alphaOrig {
		flag = ttUpper
	} else if bestValue >= beta {
		flag = ttLower
	}
	score := scoreToTT(bestValue, s.ply)
	s.tt.store(ttEntry{key: s.board.Hash, bestMove: bestMove, depth: int8(depth), score: int32(score), flag: flag})
}

func (s *searcher) alphaBeta(alpha, beta, depth int) int {
	if s.check() {
		return 0
	}

	if s.isRepetition() {
		return 0
	}

	if depth == 0 {
		return s.quiesce(alpha, beta)
	}

	alphaOrig := alpha
	score, cutoff, ttBestMove := s.probeTT(alpha, beta, depth)
	if cutoff {
		return score
	}

	if s.nullMoveCutoff(beta, depth) {
		return beta
	}

	bestValue := -infinite
	bestMove := chess.Move(0)
	start := len(s.moveStack)
	s.moveStack = s.board.GenerateLegalMoves(s.moveStack)
	legalMoves := s.moveStack[start:]
	defer func() { s.moveStack = s.moveStack[:start] }()

	if len(legalMoves) == 0 {
		return s.noMoveScore()
	}

	s.orderMoves(legalMoves, ttBestMove)
	for _, move := range legalMoves {
		s.history = append(s.history, s.board.Hash)
		u := s.board.MakeMove(move)
		s.ply++
		nullMove := s.nullMove
		s.nullMove = false
		score := -s.alphaBeta(-beta, -alpha, depth-1)
		s.nullMove = nullMove
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

	s.storeTT(bestValue, alphaOrig, beta, depth, bestMove)
	return bestValue
}

func (aB *alphaBetaBot) BestMove(b chess.Board, limits Limits) chess.Move {
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
	aB.mu.Lock()
	defer aB.mu.Unlock()
	deadline := time.Time{}
	start := time.Now()
	if limits.Time > 0 {
		deadline = time.Now().Add(limits.Time)
	}
	s := searcher{
		board:     &b,
		deadline:  deadline,
		nodes:     0,
		stopped:   false,
		history:   slices.Clone(limits.History),
		ply:       1,
		options:   aB.options,
		moveStack: make([]chess.Move, 0, 4096),
		nullMove:  false,
	}
	if !aB.options.NoTT {
		if aB.tt == nil || aB.options.FreshTT {
			aB.tt = newTranspositionTable(18)
		}
		s.tt = aB.tt
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
