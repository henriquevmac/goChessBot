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
	ebiten.SetWindowSize(ui.BoardSize, ui.BoardSize+ui.PanelSize)
	ebiten.SetWindowTitle("goChessBot")
	game := ui.NewGame(images, board)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
