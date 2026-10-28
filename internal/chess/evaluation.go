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

var midPassedPawnVal = [7]int{
	1: 5,
	2: 10,
	3: 15,
	4: 25,
	5: 40,
	6: 60,
}

var endPassedPawnVal = [7]int{
	1: 20,
	2: 25,
	3: 45,
	4: 75,
	5: 120,
	6: 180,
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

func (b *Board) mopUp(phase, matCount int) int {
	if matCount >= 200 {
		return (47*b.centreDistance(Black) + 16*(14-b.kingDistance())) * (maxPhase - phase) / (10 * maxPhase)
	} else if matCount <= -200 {
		return -((47*b.centreDistance(White) + 16*(14-b.kingDistance())) * (maxPhase - phase) / (10 * maxPhase))
	}
	return 0
}

func passedPawn(row, col int, blackPawns, whitePawns [8]int, color PieceId) bool {
	lookups := []int{-1, 0, 1}
	if col == 0 {
		lookups = []int{0, 1}
	} else if col == 7 {
		lookups = []int{-1, 0}
	}
	for _, offset := range lookups {
		if color == White {
			if blackPawns[col+offset] > row {
				return false
			}
		} else {
			if whitePawns[col+offset] < row {
				return false
			}
		}
	}
	return true
}

func (b *Board) Evaluate() int {
	blackPawns := [8]int{-1, -1, -1, -1, -1, -1, -1, -1}
	whitePawns := [8]int{10, 10, 10, 10, 10, 10, 10, 10}
	matCount, phase := b.materialAndPhaseCounts()
	count := matCount
	count += b.mopUp(phase, matCount)
	for idx, piece := range b.Square {
		if piece.Type() == Pawn {
			row, col := ConvertIndexToRowCol(idx)
			if piece.Color() == White {
				if row < whitePawns[col] {
					whitePawns[col] = row
				}
			} else {
				if row > blackPawns[col] {
					blackPawns[col] = row
				}
			}
		}
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
	for idx, piece := range b.Square {
		if piece.Type() != Pawn {
			continue
		}
		row, col := ConvertIndexToRowCol(idx)
		if passedPawn(row, col, blackPawns, whitePawns, piece.Color()) {
			if piece.Color() == White {
				count += (midPassedPawnVal[row]*phase + endPassedPawnVal[row]*(maxPhase-phase)) / maxPhase
			} else {
				count -= (midPassedPawnVal[7-row]*phase + endPassedPawnVal[7-row]*(maxPhase-phase)) / maxPhase
			}
		}
	}
	if b.ToMove == Black {
		return -count
	}
	return count
}
