package match

import (
	"goChessBot/internal/bot"
	"goChessBot/internal/chess"
	"slices"
)

var pieceValues = [7]int{
	chess.Pawn:   100,
	chess.Knight: 320,
	chess.Bishop: 330,
	chess.Rook:   500,
	chess.Queen:  900,
}

type Game struct {
	Moves       []chess.Move
	Result      chess.GameResult
	Winner      chess.PieceId
	Forfeit     bool
	Final       chess.Board
	Adjudicated bool
}

func getCurrentGameState(board *chess.Board, legalMoves, moves []chess.Move, history []uint64) (Game, bool) {
	result := board.ResultWithMoves(legalMoves)
	if result == chess.Ongoing && board.IsThreeFoldRepetition(history) {
		result = chess.ThreeFoldRepetition
	}
	if result == chess.Ongoing {
		return Game{}, false
	}

	g := Game{Moves: moves, Result: result, Final: *board, Adjudicated: false}
	if result == chess.Checkmate {
		g.Winner = board.ToMove.OppositeColor()
	}
	return g, true
}

// Play runs one game between two bots, starting from fen ("" means the start position).
// limits is passed to both bots' BestMove.
func Play(white, black bot.Bot, fen string, maxPlies int, limits bot.Limits) (Game, error) {
	board, err := chess.CreateBoard(fen)
	if err != nil {
		return Game{}, err
	}
	history := []uint64{}
	moves := []chess.Move{}
	for range maxPlies {
		legalMoves := board.GenerateLegalMoves(nil)
		if g, over := getCurrentGameState(board, legalMoves, moves, history); over {
			return g, nil
		}
		move := chess.Move(0)
		switch board.ToMove {
		case chess.White:
			move = white.BestMove(*board, limits)
		case chess.Black:
			move = black.BestMove(*board, limits)
		default:
			return Game{}, nil
		}
		if !slices.Contains(legalMoves, move) {
			return Game{
				Moves:       moves,
				Result:      chess.Ongoing,
				Winner:      board.ToMove.OppositeColor(),
				Forfeit:     true,
				Final:       *board,
				Adjudicated: false,
			}, nil
		}
		history = append(history, board.Hash)
		moves = append(moves, move)
		board.MakeMove(move)
	}

	if g, ok := getCurrentGameState(board, board.GenerateLegalMoves(nil), moves, history); ok {
		return g, nil
	}
	return Game{
		Moves:       moves,
		Result:      chess.Ongoing,
		Winner:      chess.Empty,
		Forfeit:     false,
		Final:       *board,
		Adjudicated: false,
	}, nil
}

func Material(b *chess.Board) int {
	matCount := 0
	for pieceId, val := range pieceValues {
		matCount += val*int(b.Counts[chess.PieceId(pieceId)|chess.White]) - val*int(b.Counts[chess.PieceId(pieceId)|chess.Black])
	}
	return matCount
}

func Adjudicate(g Game, threshold int) Game {
	if threshold <= 0 || g.Result != chess.Ongoing || g.Forfeit {
		return g
	}
	material := Material(&g.Final)
	g.Adjudicated = true
	if material >= threshold {
		g.Winner = chess.White
	} else if material <= -threshold {
		g.Winner = chess.Black
	} else {
		g.Winner = 0
	}
	return g
}
