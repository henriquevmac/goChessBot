package bot

import (
	"goChessBot/internal/chess"
)

const (
	kingOrderValue = 1_000
	captureBase    = 10_000
	promotionBase  = 20_000
	ttMoveScore    = 1_000_000
	killer1Score   = 9_000
	killer2Score   = 8_000
	historyMax     = 7_999
)

func scoreMove(b *chess.Board, m chess.Move) int {
	score := 0
	if m.IsCapture() {
		victimValue := chess.PieceValues[b.Square[m.To()].Type()]
		if m.Flag() == chess.EnPassant {
			victimValue = chess.PieceValues[chess.Pawn]
		}
		attacker := b.Square[m.From()].Type()
		attackerValue := chess.PieceValues[attacker]
		if attacker == chess.King {
			attackerValue = kingOrderValue
		}
		score += captureBase + 10*victimValue - attackerValue
	}
	if m.IsPromotion() {
		score += promotionBase + chess.PieceValues[m.GetPromotion()]
	}
	return score
}

func (s *searcher) orderMoves(moves []chess.Move, ttMove chess.Move) {
	var scores [256]int
	for idx, move := range moves {
		if move == ttMove {
			scores[idx] = ttMoveScore
		} else if move.IsCapture() || move.IsPromotion() {
			scores[idx] = scoreMove(s.board, move)
		} else {
			scores[idx] = min(s.historyScores[move.From()][move.To()], historyMax)
			if s.ply < maxPly {
				if move == s.killers[s.ply][0] {
					scores[idx] = killer1Score
				} else if move == s.killers[s.ply][1] {
					scores[idx] = killer2Score
				}
			}
		}
	}
	for i := 1; i < len(moves); i++ {
		move, score := moves[i], scores[i]
		j := i - 1
		for j >= 0 && scores[j] < score {
			moves[j+1] = moves[j]
			scores[j+1] = scores[j]
			j--
		}
		moves[j+1] = move
		scores[j+1] = score
	}
}
