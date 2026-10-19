package match

import (
	"goChessBot/internal/bot"
	"goChessBot/internal/chess"
	"slices"
)

type Game struct {
	Moves   []chess.Move
	Result  chess.GameResult
	Winner  chess.PieceId
	Forfeit bool
}

func getCurrentGameState(board *chess.Board, legalMoves, moves []chess.Move, history []uint64) (Game, bool) {
	result := board.ResultWithMoves(legalMoves)
	if result == chess.Ongoing && board.IsThreeFoldRepetition(history) {
		result = chess.ThreeFoldRepetition
	}
	if result == chess.Ongoing {
		return Game{}, false
	}

	g := Game{Moves: moves, Result: result}
	if result == chess.Checkmate {
		g.Winner = board.ToMove.OppositeColor()
	}
	return g, true
}

func Play(white, black bot.Bot, fen string, maxPlies int) (Game, error) {
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
			move = white.BestMove(*board)
		case chess.Black:
			move = black.BestMove(*board)
		default:
			return Game{}, nil
		}
		if !slices.Contains(legalMoves, move) {
			return Game{
				Moves:   moves,
				Result:  chess.Ongoing,
				Winner:  board.ToMove.OppositeColor(),
				Forfeit: true,
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
		Moves:   moves,
		Result:  chess.Ongoing,
		Winner:  chess.Empty,
		Forfeit: false,
	}, nil
}
