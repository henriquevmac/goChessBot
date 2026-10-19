package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"goChessBot/internal/chess"
)

const (
	BoardSize  = SquareSize * 8
)

var (
	lightSquare = color.RGBA{240, 217, 181, 255}
	darkSquare  = color.RGBA{181, 136, 99, 255}
	higlightSquare = color.NRGBA{20, 85, 30, 90}
)

type Game struct {
	Images map[chess.PieceId]*ebiten.Image
	Board *chess.Board
	IsDragged bool
	DraggedPiece chess.PieceId
	DraggedPieceOrigin int
	LegalMoves []chess.Move
}

func NewGame(images map[chess.PieceId]*ebiten.Image, board *chess.Board) *Game {
	legalMoves := board.GenerateLegalMoves(nil)
	return &Game{
		Images: images,
		Board: board,
		IsDragged: false,
		DraggedPiece: chess.Empty,
		DraggedPieceOrigin: -1,
		LegalMoves: legalMoves,
	}
}

func (g *Game) Update() error {
	LetGoPiece(g)
	HoldPiece(g)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			c := lightSquare
			if (rank+file)%2 == 1 {
				c = darkSquare
			}
			x := float32(file * SquareSize)
			y := float32(rank * SquareSize)
			vector.FillRect(screen, x, y, SquareSize, SquareSize, c, false)
		}
	}
	if g.IsDragged {
		DrawPossibleMoves(screen, g)
	}
	DrawBoardPieces(screen, g)
	if g.IsDragged {
		x, y := ebiten.CursorPosition()
		DrawPieceCentered(screen, g.Images[g.DraggedPiece], float64(x), float64(y))
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return BoardSize, BoardSize
}
