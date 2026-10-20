package chess

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const startPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

type CastlingRights uint8

const (
	WhiteKing  CastlingRights = 8
	WhiteQueen CastlingRights = 4
	BlackKing  CastlingRights = 2
	BlackQueen CastlingRights = 1
)

type Board struct {
	Square          [64]PieceId
	ToMove          PieceId
	CastlingRights  CastlingRights
	EnPassantSquare int
	HalfmoveClock   int
	FullmoveNumber  int
	KingPos         [2]int
	Hash            uint64
}

func ConvertRowColToIdx(row int, col int) int {
	return col + row*8
}

func ConvertIndexToRowCol(idx int) (int, int) {
	return idx / 8, idx % 8
}

func CreateBoard(fen string) (*Board, error) {
	Square := [64]PieceId{0}
	ToMove := White
	castlingRights := CastlingRights(0)
	EnPassantSquare := -1
	HalfmoveClock := 0
	FullmoveNumber := 1
	KingPos := [2]int{-1, -1}
	if fen == "" {
		fen = startPosition
	}
	pieceTypeFromSymbol := map[rune]PieceId{
		'k': King,
		'p': Pawn,
		'n': Knight,
		'b': Bishop,
		'r': Rook,
		'q': Queen,
	}

	fenParts := strings.Fields(fen)
	if len(fenParts) == 0 {
		return nil, fmt.Errorf("empty fen string")
	}
	file, rank := 0, 7
	for _, symbol := range fenParts[0] {
		if symbol == '/' {
			if file != 8 {
				return nil, fmt.Errorf("shorter rank: %d", rank)
			}
			if rank == 0 {
				return nil, fmt.Errorf("too many ranks")
			}
			file = 0
			rank--
		} else {
			if unicode.IsDigit(symbol) {
				emptySpaces := int(symbol - '0')
				if emptySpaces < 1 || file+emptySpaces > 8 {
					return nil, fmt.Errorf("impossible number of empty spaces")
				}
				file += emptySpaces
			} else {
				pieceColor := Black
				if unicode.IsUpper(symbol) {
					pieceColor = White
				}
				pieceId, ok := pieceTypeFromSymbol[unicode.ToLower(symbol)]
				if !ok {
					return nil, fmt.Errorf("unknown symbol: %c", symbol)
				}
				if file >= 8 {
					return nil, fmt.Errorf("too many symbols on rank: %d", rank)
				}
				index := ConvertRowColToIdx(rank, file)
				if pieceId == King {
					if KingPos[colorConversion(pieceColor)] != -1 {
						return nil, fmt.Errorf("more than one king for a color")
					}
					KingPos[colorConversion(pieceColor)] = index
				}
				Square[index] = pieceId | pieceColor
				file++
			}
		}
	}
	if rank != 0 || file != 8 {
		return nil, fmt.Errorf("wrong number of rows and cols: %d %d", 8-rank, file)
	}
	for idx, part := range fenParts[1:] {
		switch idx {
		case 0:
			switch part {
			case "b":
				ToMove = Black
			case "w":
				ToMove = White
			default:
				return nil, fmt.Errorf("wrong symbol for starting side")
			}
		case 1:
			if strings.Contains(part, "K") {
				castlingRights |= WhiteKing
			}
			if strings.Contains(part, "Q") {
				castlingRights |= WhiteQueen
			}
			if strings.Contains(part, "k") {
				castlingRights |= BlackKing
			}
			if strings.Contains(part, "q") {
				castlingRights |= BlackQueen
			}
			if strings.Contains(part, "-") && len(part) != 1 {
				return nil, fmt.Errorf("wrong castling rights")
			}
			for _, c := range part {
				if c != '-' && unicode.ToLower(c) != 'k' && unicode.ToLower(c) != 'q' {
					return nil, fmt.Errorf("wrong symbol for castling rights: %c", c)
				}
			}
		case 2:
			if part == "-" {
				continue
			}
			if len(part) != 2 {
				return nil, fmt.Errorf("en passant square with wrong length: %s", part)
			}
			col := int(part[0] - 'a')
			row := int(part[1] - '1')
			if col < 0 || col > 7 || (row != 2 && row != 5) {
				return nil, fmt.Errorf("not an en passant square: %s", part)
			}
			EnPassantSquare = ConvertRowColToIdx(row, col)
		case 3:
			parseHalf, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			} else if parseHalf < 0 {
				return nil, fmt.Errorf("halfmove clock < 0")
			}
			HalfmoveClock = parseHalf
		case 4:
			parseFull, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			} else if parseFull < 1 {
				return nil, fmt.Errorf("fullmove number < 1")
			}
			FullmoveNumber = parseFull
		}
	}
	board := &Board{
		Square,
		ToMove,
		castlingRights,
		EnPassantSquare,
		HalfmoveClock,
		FullmoveNumber,
		KingPos,
		0,
	}
	board.Hash = board.computeHash()
	return board, nil
}

func (b *Board) FEN() string {
	var sb strings.Builder
	pieceSymbolFromType := map[PieceId]rune{
		King:   'k',
		Pawn:   'p',
		Knight: 'n',
		Bishop: 'b',
		Rook:   'r',
		Queen:  'q',
	}
	for row := 7; row >= 0; row-- {
		empty := 0
		for col := range 8 {
			piece := b.Square[ConvertRowColToIdx(row, col)]
			if piece.IsEmpty() {
				empty++
				continue
			} else if empty > 0 {
				sb.WriteByte(byte('0' + empty))
				empty = 0
			}
			symbol := pieceSymbolFromType[piece.Type()]
			if piece.Color() == White {
				symbol = unicode.ToUpper(symbol)
			}
			sb.WriteRune(symbol)
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
			empty = 0
		}
		if row > 0 {
			sb.WriteRune('/')
		}
	}

	sb.WriteRune(' ')
	if b.ToMove == White {
		sb.WriteRune('w')
	} else {
		sb.WriteRune('b')
	}
	sb.WriteRune(' ')

	hasCastle := false
	if b.CastlingRights&WhiteKing != 0 {
		sb.WriteRune('K')
		hasCastle = true
	}
	if b.CastlingRights&WhiteQueen != 0 {
		sb.WriteRune('Q')
		hasCastle = true
	}
	if b.CastlingRights&BlackKing != 0 {
		sb.WriteRune('k')
		hasCastle = true
	}
	if b.CastlingRights&BlackQueen != 0 {
		sb.WriteRune('q')
		hasCastle = true
	}
	if !hasCastle {
		sb.WriteRune('-')
	}
	sb.WriteRune(' ')

	if b.EnPassantSquare == -1 {
		sb.WriteRune('-')
	} else {
		sb.WriteString(squareName(b.EnPassantSquare))
	}
	sb.WriteRune(' ')

	sb.WriteString(strconv.Itoa(b.HalfmoveClock))
	sb.WriteRune(' ')
	sb.WriteString(strconv.Itoa(b.FullmoveNumber))

	return sb.String()
}
