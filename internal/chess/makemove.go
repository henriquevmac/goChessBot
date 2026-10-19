package chess

type Undo struct {
	captured        PieceId
	castlingRights  CastlingRights
	enPassantSquare int
	halfmoveClock   int
	hash            uint64
}

func (b *Board) MakeMove(m Move) Undo {
	undoRecord := Undo{
		captured:        Empty,
		castlingRights:  b.CastlingRights,
		enPassantSquare: b.EnPassantSquare,
		halfmoveClock:   b.HalfmoveClock,
		hash:            b.Hash,
	}

	if m.IsCapture() {
		undoRecord.captured = b.Square[m.To()]
		if !(m.Flag() == EnPassant) {
			b.Hash ^= zobristPiece[undoRecord.captured][m.To()]
		}
	}

	piece := b.Square[m.From()]
	color := b.ToMove

	if m.IsPromotion() {
		piece = m.GetPromotion() | color
	}

	if m.Flag() == EnPassant {
		idx := m.To() - pawnInfo[colorConversion(color)].forward
		undoRecord.captured = b.Square[idx]
		b.Hash ^= zobristPiece[b.Square[idx]][idx]
		b.Square[idx] = Empty
	}

	if m.IsCastle() {
		side := 0
		if m.Flag() == QueenCastle {
			side = 1
		}
		rookFrom := castlingInfo[colorConversion(color)][side].rookFrom
		rookTo := castlingInfo[colorConversion(color)][side].rookTo
		b.Hash ^= zobristPiece[Rook|color][rookFrom]
		b.Square[rookFrom] = Empty
		b.Square[rookTo] = Rook | color
		b.Hash ^= zobristPiece[Rook|color][rookTo]
	}

	if b.EnPassantSquare != -1 {
		b.Hash ^= zobristEP[b.EnPassantSquare%8]
	}
	if m.Flag() == DoublePawn {
		b.EnPassantSquare = m.To() - pawnInfo[colorConversion(color)].forward
	} else {
		b.EnPassantSquare = -1
	}
	if b.EnPassantSquare != -1 {
		b.Hash ^= zobristEP[b.EnPassantSquare%8]
	}

	b.Hash ^= zobristCastling[b.CastlingRights]
	b.CastlingRights &= castlingKeep[m.From()] & castlingKeep[m.To()]
	b.Hash ^= zobristCastling[b.CastlingRights]

	if m.IsCapture() || b.Square[m.From()] == Pawn|b.ToMove {
		b.HalfmoveClock = 0
	} else {
		b.HalfmoveClock++
	}

	if b.ToMove == Black {
		b.FullmoveNumber++
	}

	b.ToMove = color.OppositeColor()
	b.Hash ^= zobristBlack

	if b.Square[m.From()].Type() == King {
		b.KingPos[colorConversion(color)] = m.To()
	}

	b.Hash ^= zobristPiece[b.Square[m.From()]][m.From()]
	b.Square[m.From()] = Empty
	b.Square[m.To()] = piece
	b.Hash ^= zobristPiece[piece][m.To()]

	return undoRecord
}

func (b *Board) UnmakeMove(m Move, u Undo) {
	b.ToMove = b.ToMove.OppositeColor()
	if b.ToMove == Black {
		b.FullmoveNumber--
	}

	piece := b.Square[m.To()]
	capturedPiece := u.captured

	if m.IsPromotion() {
		piece = Pawn | b.ToMove
	}

	b.HalfmoveClock = u.halfmoveClock
	b.EnPassantSquare = u.enPassantSquare
	b.CastlingRights = u.castlingRights
	b.Hash = u.hash

	if m.IsCastle() {
		side := 0
		if m.Flag() == QueenCastle {
			side = 1
		}
		rookFrom := castlingInfo[colorConversion(b.ToMove)][side].rookFrom
		rookTo := castlingInfo[colorConversion(b.ToMove)][side].rookTo
		b.Square[rookTo] = Empty
		b.Square[rookFrom] = Rook | b.ToMove
	}

	if m.Flag() == EnPassant {
		capturedPiece = Empty
		targetSquare := m.To() - pawnInfo[colorConversion(b.ToMove)].forward
		b.Square[targetSquare] = u.captured
	}

	if b.Square[m.To()].Type() == King {
		b.KingPos[colorConversion(b.ToMove)] = m.From()
	}

	b.Square[m.From()] = piece
	b.Square[m.To()] = capturedPiece
}
