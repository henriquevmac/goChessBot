package ui

import (
	"image/color"
	"log"
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"goChessBot/internal/bot"
	"goChessBot/internal/chess"
)

const (
	BoardSize = SquareSize * 8
	PanelSize = 40
)

var (
	lightSquare     = color.RGBA{240, 217, 181, 255}
	darkSquare      = color.RGBA{181, 136, 99, 255}
	highlightSquare = color.NRGBA{20, 85, 30, 90}
)

type Game struct {
	Images             map[chess.PieceId]*ebiten.Image
	Board              *chess.Board
	IsDragged          bool
	DraggedPiece       chess.PieceId
	DraggedPieceOrigin int
	LegalMoves         []chess.Move
	Result             chess.GameResult
	History            []uint64
	Bot                bot.Bot
	BotColor           chess.PieceId
	BotLimits          bot.Limits
	thinking           bool
	botMoves           chan chess.Move
}

func NewGame(images map[chess.PieceId]*ebiten.Image, board *chess.Board) *Game {
	legalMoves := board.GenerateLegalMoves(nil)
	return &Game{
		Images:             images,
		Board:              board,
		IsDragged:          false,
		DraggedPiece:       chess.Empty,
		DraggedPieceOrigin: -1,
		LegalMoves:         legalMoves,
		Result:             board.ResultWithMoves(legalMoves),
		History:            []uint64{},
		Bot:                bot.New(bot.AlphaBeta),
		BotColor:           chess.Black,
		BotLimits:          bot.Limits{Depth: 6, Time: time.Second * 2},
		thinking:           false,
		botMoves:           make(chan chess.Move, 1),
	}
}

func (g *Game) applyMove(m chess.Move) {
	g.History = append(g.History, g.Board.Hash)
	g.Board.MakeMove(m)
	g.LegalMoves = g.Board.GenerateLegalMoves(nil)
	g.Result = g.Board.ResultWithMoves(g.LegalMoves)
	if g.Result == chess.Ongoing && g.Board.IsThreeFoldRepetition(g.History) {
		g.Result = chess.ThreeFoldRepetition
	}
}

func inRect(x, y, rx, ry, rw, rh int) bool {
	return x >= rx && x < rx+rw && y >= ry && y < ry+rh
}

func (g *Game) restart() {
	board, err := chess.CreateBoard("")
	if err != nil {
		log.Println("Error Creating new game")
		return
	}
	*g = *NewGame(g.Images, board)
}

func (g *Game) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if inRect(x, y, 5, BoardSize+5, len("Restart")*7, 30) {
			g.restart()
			return nil
		}
	}
	if g.Result == chess.Ongoing {
		if g.Board.ToMove != g.BotColor {
			LetGoPiece(g)
			HoldPiece(g)
		} else if !g.thinking {
			g.thinking = true
			boardCopy := *g.Board
			ch := g.botMoves
			b := g.Bot
			limits := g.BotLimits
			limits.History = slices.Clone(g.History) // the UI keeps appending to its own
			go func() { ch <- b.BestMove(boardCopy, limits) }()
		} else {
			select {
			case m := <-g.botMoves:
				g.thinking = false
				if slices.Contains(g.LegalMoves, m) {
					g.applyMove(m)
				} else {
					log.Fatal("Bot reported buggy move: " + m.String())
				}
			default:
			}
		}
	}
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
	restartText := "Restart"
	vector.FillRect(screen, 5, BoardSize+5, float32(len(restartText)*7), 30, darkSquare, false)
	ebitenutil.DebugPrintAt(screen, restartText, 7, BoardSize+10)
	if g.IsDragged {
		DrawPossibleMoves(screen, g)
	}
	DrawBoardPieces(screen, g)
	if g.IsDragged {
		x, y := ebiten.CursorPosition()
		DrawPieceCentered(screen, g.Images[g.DraggedPiece], float64(x), float64(y))
	}
	if g.Result != chess.Ongoing {
		msg := "Game Ended Via: " + g.Result.String()
		if g.Result == chess.Checkmate {
			if g.Board.ToMove.OppositeColor() == chess.White {
				msg += " White Wins!"
			} else {
				msg += " Black Wins!"
			}
		}
		vector.FillRect(screen, 70, BoardSize+5, float32(len(msg)*7), 30, color.NRGBA{0, 0, 0, 160}, false)
		ebitenutil.DebugPrintAt(screen, msg, 75, BoardSize+10)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return BoardSize, BoardSize + PanelSize
}
