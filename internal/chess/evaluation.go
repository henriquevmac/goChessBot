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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (b *Board) materialAndPhaseCounts() (int, int) {
	matCount := 0
	phase := 0
	for pieceId, val := range PieceValues {
		matCount += val*int(b.Counts[PieceId(pieceId)|White]) - val*int(b.Counts[PieceId(pieceId)|Black])
		phase += phaseValues[pieceId]*int(b.Counts[PieceId(pieceId)|White]) + phaseValues[pieceId]*int(b.Counts[PieceId(pieceId)|Black])
	}
	phase = min(phase, maxPhase)
	return matCount, phase
}

func (b *Board) centreDistance(color PieceId) int {
	row, col := ConvertIndexToRowCol(b.KingPos[colorConversion(color)])
	fileDis := max(3-col, col-4)
	rankDis := max(3-row, row-4)
	return fileDis + rankDis
}

func (b *Board) kingDistance() int {
	rowB, colB := ConvertIndexToRowCol(b.KingPos[colorConversion(Black)])
	rowW, colW := ConvertIndexToRowCol(b.KingPos[colorConversion(White)])
	return abs(rowB-rowW) + abs(colB-colW)
}

func (b *Board) Evaluate() int {
	matCount, phase := b.materialAndPhaseCounts()
	count := matCount
	if matCount >= 200 {
		count += (47*b.centreDistance(Black) + 16*(14-b.kingDistance())) * (maxPhase - phase) / (10 * maxPhase)
	} else if matCount <= -200 {
		count -= (47*b.centreDistance(White) + 16*(14-b.kingDistance())) * (maxPhase - phase) / (10 * maxPhase)
	}
	for idx, piece := range b.Square {
		if piece.Color() == White {
			if piece.Type() == King {
				count += (kingMidSquareTable[idx^56]*phase + kingEndSquareTable[idx^56]*(maxPhase-phase)) / maxPhase
			} else {
				count += pieceSquareTables[piece.Type()][idx^56]
			}
		} else {
			if piece.Type() == King {
				count -= (kingMidSquareTable[idx]*phase + kingEndSquareTable[idx]*(maxPhase-phase)) / maxPhase
			} else {
				count -= pieceSquareTables[piece.Type()][idx]
			}
		}
	}
	if b.ToMove == Black {
		return -count
	}
	return count
}
