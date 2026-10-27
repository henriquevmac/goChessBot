package chess

var PieceValues = [7]int{
	Pawn:   100,
	Knight: 320,
	Bishop: 330,
	Rook:   500,
	Queen:  900,
}

const (
	maxPhase = 24
)

var phaseValues = [7]int{
	Pawn:   0,
	Knight: 1,
	Bishop: 1,
	Rook:   2,
	Queen:  4,
}

func (b *Board) Evaluate() int {
	matCount := 0
	phase := 0
	for pieceId, val := range PieceValues {
		matCount += val*int(b.Counts[PieceId(pieceId)|White]) - val*int(b.Counts[PieceId(pieceId)|Black])
		phase += phaseValues[pieceId]*int(b.Counts[PieceId(pieceId)|White]) + phaseValues[pieceId]*int(b.Counts[PieceId(pieceId)|Black])
	}
	phase = min(phase, maxPhase)
	for idx, piece := range b.Square {
		if piece.Color() == White {
			if piece.Type() == King {
				matCount += (kingMidSquareTable[idx^56]*phase + kingEndSquareTable[idx^56]*(maxPhase-phase)) / maxPhase
			} else {
				matCount += pieceSquareTables[piece.Type()][idx^56]
			}
		} else {
			if piece.Type() == King {
				matCount -= (kingMidSquareTable[idx]*phase + kingEndSquareTable[idx]*(maxPhase-phase)) / maxPhase
			} else {
				matCount -= pieceSquareTables[piece.Type()][idx]
			}
		}
	}
	if b.ToMove == Black {
		return -matCount
	}
	return matCount
}
