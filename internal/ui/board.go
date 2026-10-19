package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"goChessBot/internal/chess"
)

const SquareSize = 80;

func DrawBoardPieces(screen *ebiten.Image, g *Game) {
	b := g.Board
	imgs := g.Images
	for idx, square := range b.Square {
		if square & chess.PieceMask == chess.Empty {
			continue
		}
		if g.IsDragged && g.DraggedPieceOrigin == idx {
			continue
		}
		row := idx / 8
		col := idx % 8
		boardRow := 7 - row
		posX := SquareSize * col
		posY := SquareSize * boardRow
		DrawPiece(screen, imgs[square], float64(posX), float64(posY))
	}
}
