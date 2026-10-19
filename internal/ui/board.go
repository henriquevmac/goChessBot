package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"goChessBot/internal/chess"
)

const SquareSize = 80

func squareToScreen(sq int) (float32, float32) {
	col, row := sq%8, 7-sq/8
	x := float32(col * SquareSize)
	y := float32(row * SquareSize)
	return x, y
}

func DrawBoardPieces(screen *ebiten.Image, g *Game) {
	b := g.Board
	imgs := g.Images
	for idx, square := range b.Square {
		if square&chess.PieceMask == chess.Empty {
			continue
		}
		if g.IsDragged && g.DraggedPieceOrigin == idx {
			continue
		}
		posX, posY := squareToScreen(idx)
		DrawPiece(screen, imgs[square], float64(posX), float64(posY))
	}
}

func DrawPossibleMoves(screen *ebiten.Image, g *Game) {
	for _, move := range g.LegalMoves {
		if move.From() == g.DraggedPieceOrigin {
			if !move.IsPromotion() || move.GetPromotion() == chess.Queen {
				x, y := squareToScreen(move.To())
				vector.FillRect(screen, x, y, SquareSize, SquareSize, highlightSquare, false)
			}
		}
	}
}
