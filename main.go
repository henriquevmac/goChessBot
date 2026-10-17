package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"goChessBot/internal/chess"
	"goChessBot/internal/ui"
)

func main() {
	images, err := ui.LoadPieces()
	if err != nil {
		log.Fatal(err)
	}
	board, err := chess.CreateBoard("")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Board: ", board)
	ebiten.SetWindowSize(ui.BoardSize, ui.BoardSize)
	ebiten.SetWindowTitle("goChessBot")
	if err := ebiten.RunGame(&ui.Game{Images: images, Board: board, IsDragged: false, DraggedPiece: chess.Empty, DraggedPieceOrigin: -1}); err != nil {
		log.Fatal(err)
	}
}
