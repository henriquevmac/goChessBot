package chess

func (b *Board) IsSquareAttacked(sq int, by PieceId) bool {
	for _, t := range knightTargets[sq] {
		piece := b.Square[t]
		if piece.Type() == Knight && piece.Color() == by {
			return true
		}
	}
	for _, t := range kingTargets[sq] {
		piece := b.Square[t]
		if piece.Type() == King && piece.Color() == by {
			return true
		}
	}
	for _, t := range pawnTargets[sq][colorConversion(by.OppositeColor())] {
		piece := b.Square[t]
		if piece.Type() == Pawn && piece.Color() == by {
			return true
		}
	}
	for dir := north; dir <= southWest; dir++ {
		for step := 1; step <= distanceToEdge[sq][dir]; step++ {
			target := sq + directionOffsets[dir]*step
			piece := b.Square[target]
			if !piece.IsEmpty() {
				if (piece.Type() == Queen || (piece.Type() == Rook && dir <= west) || (piece.Type() == Bishop && dir > west)) && piece.Color() == by {
					return true
				}
				break
			}
		}
	}
	return false
}
