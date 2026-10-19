package chess

const (
	PieceMask      = 7
	PieceColorMask = 24
)

type PieceId int

const (
	Empty PieceId = iota
	King
	Pawn
	Knight
	Bishop
	Rook
	Queen
)

const (
	White PieceId = 8
	Black PieceId = 16
)

func (p PieceId) Color() PieceId {
	return p & PieceColorMask
}

func (p PieceId) OppositeColor() PieceId {
	if p.Color() == White {
		return Black
	} else {
		return White
	}
}

func (p PieceId) Type() PieceId {
	return p & PieceMask
}

func (p PieceId) IsOpponent(opp PieceId) bool {
	return !p.IsEmpty() && !opp.IsEmpty() && p.Color() != opp.Color()
}

func (p PieceId) IsEmpty() bool {
	return p == Empty
}
