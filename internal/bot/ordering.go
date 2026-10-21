package bot

import (
	"goChessBot/internal/chess"
)

const (
	kingOrderValue = 1_000
	captureBase    = 10_000
	promotionBase  = 20_000
	ttMoveScore    = 1_000_000
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

func orderMoves(b *chess.Board, moves []chess.Move, ttMove chess.Move) {
	var scores [256]int
	for idx, move := range moves {
		if move == ttMove {
			scores[idx] = ttMoveScore
		} else {
			scores[idx] = scoreMove(b, move)
		}
	}
	for i := 1; i < len(moves); i++ {
		m, s := moves[i], scores[i]
		j := i - 1
		for j >= 0 && scores[j] < s {
			moves[j+1] = moves[j]
			scores[j+1] = scores[j]
			j--
		}
		moves[j+1] = m
		scores[j+1] = s
	}
}
