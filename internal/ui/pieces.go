package ui

import (
	"fmt"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"goChessBot/internal/chess"
)

const (
	piecesPath        = "./assets/chessPieces/"
	piecesSpriteScale = 0.55
)

var names = map[chess.PieceId]string{
	chess.Empty:  "",
	chess.King:   "K",
	chess.Pawn:   "P",
	chess.Knight: "Kn",
	chess.Bishop: "B",
	chess.Rook:   "R",
	chess.Queen:  "Q",
	chess.White:  "w",
	chess.Black:  "b",
}

func LoadPieces() (map[chess.PieceId]*ebiten.Image, error) {
	images := make(map[chess.PieceId]*ebiten.Image)
	for i := 1; i <= 6; i++ {
		for c := 8; c <= 16; c += 8 {
			colorString := names[chess.PieceId(c)]
			pieceName := colorString + names[chess.PieceId(i)]
			filePath := piecesPath + pieceName + ".png"
			img, _, err := ebitenutil.NewImageFromFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("loading %s: %w", names[chess.PieceId(i)], err)
			}
			pieceId := c | i
			images[chess.PieceId(pieceId)] = img
		}
	}
	return images, nil
}

func DrawPiece(screen *ebiten.Image, img *ebiten.Image, posX float64, posY float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(piecesSpriteScale, piecesSpriteScale)
	offsetX := (SquareSize - (float64(img.Bounds().Dx()) * piecesSpriteScale)) / 2
	offsetY := (SquareSize - (float64(img.Bounds().Dy()) * piecesSpriteScale)) / 2
	op.GeoM.Translate(posX+offsetX, posY+offsetY)
	screen.DrawImage(img, op)
}

func DrawPieceCentered(screen *ebiten.Image, img *ebiten.Image, posX float64, posY float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(piecesSpriteScale, piecesSpriteScale)
	offsetX := float64(img.Bounds().Dx()) * piecesSpriteScale / 2
	offsetY := float64(img.Bounds().Dy()) * piecesSpriteScale / 2
	op.GeoM.Translate(posX-offsetX, posY-offsetY)
	screen.DrawImage(img, op)
}

func squareAt(x, y int) (int, bool) {
	if outsideOfBorder(x, y) {
		return -1, false
	}
	col, row := x/SquareSize, 7-y/SquareSize
	index := chess.ConvertRowColToIdx(row, col)
	return index, true
}

func HoldPiece(g *Game) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !g.IsDragged {
		x, y := ebiten.CursorPosition()
		index, inside := squareAt(x, y)
		if !inside {
			return
		}
		if g.Board.Square[index] != 0 && g.Board.Square[index].Color() == g.Board.ToMove {
			g.DraggedPiece = g.Board.Square[index]
			g.DraggedPieceOrigin = index
			g.IsDragged = true
		}
	}
}

func outsideOfBorder(x int, y int) bool {
	return x < 0 || x >= BoardSize || y < 0 || y >= BoardSize
}

func LetGoPiece(g *Game) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.IsDragged {
		cleanup := func() {
			g.IsDragged = false
			g.DraggedPiece = 0
			g.DraggedPieceOrigin = -1
		}
		defer cleanup()
		x, y := ebiten.CursorPosition()
		index, inside := squareAt(x, y)
		if !inside {
			return
		}
		move, legal := g.Board.FindLegalMove(g.DraggedPieceOrigin, index, chess.Queen)
		if legal {
			g.Board.MakeMove(move)
			g.LegalMoves = g.Board.GenerateLegalMoves(nil)
			g.Result = g.Board.Result()
		}
	}
}
