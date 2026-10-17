package chess

type Move uint16;
type MoveFlag uint8;

const (
	Normal MoveFlag = iota
	DoublePawn
	KingCastle
	QueenCastle
	Capture
	EnPassant
	PromoKn MoveFlag = 8
	PromoB MoveFlag = 9
	PromoR MoveFlag = 10
	PromoQ MoveFlag = 11
	PromoKnCap MoveFlag = 12
	PromoBCap MoveFlag = 13
	PromoRCap MoveFlag = 14
	PromoQCap MoveFlag = 15
)

const (
	fromShift = 0
	toShift = 6
	flagShift = 12

	squareMask Move = 0x3F
	flagMask Move = 0xF
)

func squareName(sq int) string {
	file := sq % 8
	rank := sq / 8
	return string([]byte{'a' + byte(file), '1' + byte(rank)})
}

func CreateMove(from, to int, flag MoveFlag) Move {
	return (Move(from)&squareMask) << fromShift | (Move(to)&squareMask) << toShift | (Move(flag)&flagMask) << flagShift
}

func (m Move) From() int {
	return int(m >> fromShift & squareMask)
}

func (m Move) To() int {
	return int(m >> toShift & squareMask)
}

func (m Move) Flag() MoveFlag {
	return MoveFlag(m >> flagShift & flagMask)
}

func (m Move) IsCapture() bool {
	return m.Flag() & 4 != 0
}

func (m Move) IsPromotion() bool {
	return m.Flag() & 8 != 0
}

func (m Move) String() string {
	str := squareName(m.From()) + squareName(m.To())
	if m.IsPromotion() {
		promoLetters := "nbrq"
		return str + string(promoLetters[m.Flag() & 3])
	}
	return str
}
