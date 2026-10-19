package chess

type GameResult int

const (
	Ongoing GameResult = iota
	Checkmate
	Stalemate
	FiftyMoveRule
	InsufficientMaterial
	ThreeFoldRepetition
)

func (gR GameResult) String() string {
	switch gR {
	case Ongoing:
		return "Ongoing"
	case Checkmate:
		return "Checkmate"
	case Stalemate:
		return "Stalemate"
	case FiftyMoveRule:
		return "FiftyMoveRule"
	case InsufficientMaterial:
		return "InsufficientMaterial"
	case ThreeFoldRepetition:
		return "ThreeFoldRepetition"
	default:
		return ""
	}
}

func (b *Board) InCheck() bool {
	return b.IsSquareAttacked(b.KingPos[colorConversion(b.ToMove)], b.ToMove.OppositeColor())
}

func (b *Board) insufficientMaterial() bool {
	bishops := [2]int{0, 0}
	knights := 0
	for idx, piece := range b.Square {
		if piece.Type() == Pawn || piece.Type() == Rook || piece.Type() == Queen {
			return false
		}
		if piece.Type() == Bishop {
			row, col := ConvertIndexToRowCol(idx)
			sqCol := (row + col) % 2
			bishops[sqCol]++
		} else if piece.Type() == Knight {
			knights++
		}
	}

	return (knights == 0 && (bishops[0] == 0 || bishops[1] == 0)) || (knights == 1 && bishops[0]+bishops[1] == 0)
}

func (b *Board) ResultWithMoves(moves []Move) GameResult {
	if len(moves) == 0 {
		if b.InCheck() {
			return Checkmate
		} else {
			return Stalemate
		}
	}
	if b.HalfmoveClock >= 100 {
		return FiftyMoveRule
	}
	if b.insufficientMaterial() {
		return InsufficientMaterial
	}
	return Ongoing
}

func (b *Board) Result() GameResult {
	return b.ResultWithMoves(b.GenerateLegalMoves(nil))
}
