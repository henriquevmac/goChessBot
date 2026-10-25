package chess

func (b *Board) GenerateLegalMoves(moves []Move) []Move {
	pseudo := b.GenerateMoves(moves, false)
	start := len(moves)
	n := start
	for _, move := range pseudo[start:] {
		u := b.MakeMove(move)
		oppColor := b.ToMove.OppositeColor()
		if !b.IsSquareAttacked(b.KingPos[colorConversion(oppColor)], b.ToMove) {
			pseudo[n] = move
			n++
		}
		b.UnmakeMove(move, u)
	}
	return pseudo[:n]
}

func (b *Board) GenerateCaptureMoves(moves []Move) []Move {
	pseudo := b.GenerateMoves(moves, true)
	start := len(moves)
	n := start
	for _, move := range pseudo[start:] {
		u := b.MakeMove(move)
		oppColor := b.ToMove.OppositeColor()
		if !b.IsSquareAttacked(b.KingPos[colorConversion(oppColor)], b.ToMove) {
			pseudo[n] = move
			n++
		}
		b.UnmakeMove(move, u)
	}
	return pseudo[:n]
}

func (b *Board) Perft(depth int) int {
	if depth == 0 {
		return 1
	}
	total := 0
	moves := []Move{}
	moves = b.GenerateLegalMoves(moves)
	if depth == 1 {
		return len(moves)
	}
	for _, move := range moves {
		u := b.MakeMove(move)
		total += b.Perft(depth - 1)
		b.UnmakeMove(move, u)
	}

	return total
}

func (b *Board) FindLegalMove(from, to int, promotion PieceId) (Move, bool) {
	for _, move := range b.GenerateLegalMoves(nil) {
		if move.From() == from && move.To() == to {
			if move.IsPromotion() {
				if move.GetPromotion() == promotion {
					return move, true
				}
			} else {
				return move, true
			}
		}
	}
	return 0, false
}
