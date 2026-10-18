package chess

import (
	"strconv"
	"strings"
	"unicode"
)

const startPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

type CastlingRights uint8;

const (
	WhiteKing CastlingRights = 8
	WhiteQueen CastlingRights = 4
	BlackKing CastlingRights = 2
	BlackQueen CastlingRights = 1
)

type Board struct {
	Square [64]PieceId
	ToMove PieceId
	CastlingRights CastlingRights
	EnPassantSquare int
	HalfmoveClock int
	FullmoveNumber int
	KingPos [2]int
}

func ConvertRowColToIdx(row int, col int) int {
	return col + row * 8
}

func CreateBoard(fen string) (*Board, error) {
	Square := [64]PieceId{0}
	ToMove := White
	castlingRights := CastlingRights(0)
	EnPassantSquare := -1
	HalfmoveClock := 0
	FullmoveNumber := 1
	KingPos := [2]int{}
	if fen == "" {
		fen = startPosition
	}
	pieceTypeFromSymbol := map[rune]PieceId {
		'k': King,
		'p': Pawn,
		'n': Knight,
		'b': Bishop,
		'r': Rook,
		'q': Queen,
	}

	fenParts := strings.Fields(fen)
	file, rank := 0, 7
	for _, symbol := range fenParts[0] {
		if symbol == '/' {
			file = 0
			rank--
		} else {
			if unicode.IsDigit(symbol) {
				file += int(symbol - '0')
			} else {
				pieceColor := Black
				if unicode.IsUpper(symbol) {
					pieceColor = White
				}
				pieceId := pieceTypeFromSymbol[unicode.ToLower(symbol)]
				index := ConvertRowColToIdx(rank, file)
				if pieceId == King {
					KingPos[colorConversion(pieceColor)] = index
				}
				Square[index] = pieceId | pieceColor
				file++
			}
		}
	}
	for idx, part := range fenParts[1:] {
		switch idx {
		case 0:
			if part == "b" {
				ToMove = Black
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
		case 2:
			if len(part) < 2 {
				continue
			}
			col := int(part[0] - 'a')
			row := int(part[1] - '1')
			EnPassantSquare = ConvertRowColToIdx(row, col)
		case 3:
			parseHalf, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			HalfmoveClock = parseHalf
		case 4:
			parseFull, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			FullmoveNumber = parseFull
		}
	}
	return &Board{Square, ToMove, castlingRights, EnPassantSquare, HalfmoveClock, FullmoveNumber, KingPos}, nil
}
