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
	if b.Counts[Pawn|White]+b.Counts[Pawn|Black] > 0 ||
		b.Counts[Rook|White]+b.Counts[Rook|Black] > 0 ||
		b.Counts[Queen|White]+b.Counts[Queen|Black] > 0 {
		return false
	}

	knights := b.Counts[Knight|White] + b.Counts[Knight|Black]
	bishops := b.Counts[Bishop|White] + b.Counts[Bishop|Black]

	if bishops == 0 {
		return knights <= 1
	}

	if knights > 0 {
		return false
	}

	var onColor [2]int
	for idx, piece := range b.Square {
		if piece.Type() == Bishop {
			row, col := ConvertIndexToRowCol(idx)
			onColor[(row+col)%2]++
		}
	}
	return onColor[0] == 0 || onColor[1] == 0
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
