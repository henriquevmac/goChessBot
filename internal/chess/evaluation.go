package chess

var PieceValues = [7]int{
	Pawn:   100,
	Knight: 320,
	Bishop: 330,
	Rook:   500,
	Queen:  900,
}

func (b *Board) Evaluate() int {
	matCount := 0
	for pieceId, val := range PieceValues {
		matCount += val*int(b.Counts[PieceId(pieceId)|White]) - val*int(b.Counts[PieceId(pieceId)|Black])
	}
	if b.ToMove == Black {
		return -matCount
	}
	return matCount
}

func (b *Board) Quiesce(alpha, beta int) int {
	staticEval := b.Evaluate()

	bestValue := staticEval
	if bestValue >= beta {
		return bestValue
	}
	if bestValue > alpha {
		alpha = bestValue
	}

	captures := b.GenerateCaptureMoves(nil)
	for _, move := range captures {
		u := b.MakeMove(move)
		score := -b.Quiesce(-beta, -alpha)
		b.UnmakeMove(move, u)

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
