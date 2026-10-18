package chess

type pawnRules struct {
	forward  int
	startRow int
	promoRow int
}

var pawnInfo = [2]pawnRules{
	{forward: 8, startRow: 1, promoRow: 7},
	{forward: -8, startRow: 6, promoRow: 0},
}

type castlingRule struct {
	right    CastlingRights
	kingFrom int
	kingTo   int
	rookFrom int
	rookTo   int
	empty    []int
	safe     [3]int
	flag     MoveFlag
}

var castlingInfo = [2][2]castlingRule{
	{{right: WhiteKing, kingFrom: 4, kingTo: 6, rookFrom: 7, rookTo: 5, empty: []int{5, 6}, safe: [3]int{4, 5, 6}, flag: KingCastle},
		{right: WhiteQueen, kingFrom: 4, kingTo: 2, rookFrom: 0, rookTo: 3, empty: []int{1, 2, 3}, safe: [3]int{2, 3, 4}, flag: QueenCastle}},
	{{right: BlackKing, kingFrom: 60, kingTo: 62, rookFrom: 63, rookTo: 61, empty: []int{61, 62}, safe: [3]int{60, 61, 62}, flag: KingCastle},
		{right: BlackQueen, kingFrom: 60, kingTo: 58, rookFrom: 56, rookTo: 59, empty: []int{57, 58, 59}, safe: [3]int{58, 59, 60}, flag: QueenCastle}},
}

func (b *Board) generateKnightOrKingMoves(moves []Move, sq int, targets []int) []Move {
	for _, target := range targets {
		if b.Square[target].IsEmpty() {
			move := CreateMove(sq, target, Normal)
			moves = append(moves, move)
		} else if b.Square[target].IsOpponent(b.Square[sq]) {
			move := CreateMove(sq, target, Capture)
			moves = append(moves, move)
		}
	}
	return moves
}

func (b *Board) generateSlidingMoves(moves []Move, sq int, start, end int) []Move {
	for dir := start; dir <= end; dir++ {
		for step := 1; step <= distanceToEdge[sq][dir]; step++ {
			target := sq + directionOffsets[dir]*step
			if b.Square[target].IsEmpty() {
				move := CreateMove(sq, target, Normal)
				moves = append(moves, move)
			} else if b.Square[target].IsOpponent(b.Square[sq]) {
				move := CreateMove(sq, target, Capture)
				moves = append(moves, move)
				break
			} else {
				break
			}
		}
	}
	return moves
}

func addPawnPromotions(moves []Move, from, to int, isCapture bool) []Move {
	if isCapture {
		for mFlag := PromoKnCap; mFlag <= PromoQCap; mFlag++ {
			move := CreateMove(from, to, mFlag)
			moves = append(moves, move)
		}
	} else {
		for mFlag := PromoKn; mFlag <= PromoQ; mFlag++ {
			move := CreateMove(from, to, mFlag)
			moves = append(moves, move)
		}
	}
	return moves
}

func (b *Board) generatePawnMoves(moves []Move, sq int) []Move {
	row := sq / 8
	piece := b.Square[sq]
	c := colorConversion(piece)
	rules := pawnInfo[c]
	oneAhead := sq + rules.forward
	twoAhead := oneAhead + rules.forward
	if b.Square[oneAhead].IsEmpty() {
		if oneAhead/8 == rules.promoRow {
			moves = addPawnPromotions(moves, sq, oneAhead, false)
		} else {
			move := CreateMove(sq, oneAhead, Normal)
			moves = append(moves, move)
		}
		if row == rules.startRow && b.Square[twoAhead].IsEmpty() {
			move := CreateMove(sq, twoAhead, DoublePawn)
			moves = append(moves, move)
		}
	}
	for _, target := range pawnTargets[sq][c] {
		if b.Square[target].IsOpponent(piece) {
			if target/8 == rules.promoRow {
				moves = addPawnPromotions(moves, sq, target, true)
			} else {
				move := CreateMove(sq, target, Capture)
				moves = append(moves, move)
			}
		} else if target == b.EnPassantSquare {
			move := CreateMove(sq, target, EnPassant)
			moves = append(moves, move)
		}
	}
	return moves
}

func (b *Board) generateCastleMoves(moves []Move) []Move {
	rules := castlingInfo[colorConversion(b.ToMove)]
	opponent := b.ToMove.OppositeColor()

ruleLoop:
	for _, rule := range rules {
		if b.CastlingRights&rule.right == 0 {
			continue
		}
		for _, sqr := range rule.empty {
			if !b.Square[sqr].IsEmpty() {
				continue ruleLoop
			}
		}
		for _, sqr := range rule.safe {
			if b.IsSquareAttacked(sqr, opponent) {
				continue ruleLoop
			}
		}
		move := CreateMove(rule.kingFrom, rule.kingTo, rule.flag)
		moves = append(moves, move)
	}

	return moves
}

func (b *Board) GenerateMoves(moves []Move) []Move {
	for idx, piece := range b.Square {
		if piece.IsOpponent(b.ToMove) {
			continue
		}
		switch piece.Type() {
		case King:
			moves = b.generateKnightOrKingMoves(moves, idx, kingTargets[idx])
		case Pawn:
			moves = b.generatePawnMoves(moves, idx)
		case Knight:
			moves = b.generateKnightOrKingMoves(moves, idx, knightTargets[idx])
		case Bishop:
			moves = b.generateSlidingMoves(moves, idx, northEast, southWest)
		case Rook:
			moves = b.generateSlidingMoves(moves, idx, north, west)
		case Queen:
			moves = b.generateSlidingMoves(moves, idx, north, southWest)
		}
	}
	moves = b.generateCastleMoves(moves)

	return moves
}
