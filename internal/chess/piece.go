package chess

const (
	PieceMask = 7
	PieceColorMask = 24
)

type PieceId int;

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
