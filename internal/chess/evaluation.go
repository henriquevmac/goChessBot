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
