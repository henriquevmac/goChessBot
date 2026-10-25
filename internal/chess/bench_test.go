package chess

import "testing"

// Benchmarks for the engine's hot paths. Run them with:
//
//	go test ./internal/chess/ -run '^$' -bench . -benchmem
//
// -run '^$' skips the normal tests, and -benchmem also reports memory allocations,
// which matter for speed: every allocation is work for the garbage collector.

var benchPositions = []struct {
	name string
	fen  string
}{
	{"start", startFEN},
	{"kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"},
	{"middlegame", "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10"},
	{"endgame", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"},
}

func benchBoard(b *testing.B, fen string) *Board {
	b.Helper()
	board, err := CreateBoard(fen)
	if err != nil {
		b.Fatalf("CreateBoard(%q): %v", fen, err)
	}
	return board
}

func BenchmarkGenerateMoves(b *testing.B) {
	for _, pos := range benchPositions {
		b.Run(pos.name, func(b *testing.B) {
			board := benchBoard(b, pos.fen)
			buf := make([]Move, 0, 256)
			for b.Loop() {
				buf = board.GenerateMoves(buf[:0], false)
			}
		})
	}
}

func BenchmarkGenerateLegalMoves(b *testing.B) {
	for _, pos := range benchPositions {
		b.Run(pos.name, func(b *testing.B) {
			board := benchBoard(b, pos.fen)
			buf := make([]Move, 0, 256)
			for b.Loop() {
				buf = board.GenerateLegalMoves(buf[:0])
			}
		})
	}
}

// One iteration makes and unmakes every legal move of the position once.
func BenchmarkMakeUnmake(b *testing.B) {
	for _, pos := range benchPositions {
		b.Run(pos.name, func(b *testing.B) {
			board := benchBoard(b, pos.fen)
			moves := board.GenerateLegalMoves(nil)
			for b.Loop() {
				for _, m := range moves {
					u := board.MakeMove(m)
					board.UnmakeMove(m, u)
				}
			}
			b.ReportMetric(float64(len(moves)), "moves/op")
		})
	}
}

func BenchmarkIsSquareAttacked(b *testing.B) {
	for _, pos := range benchPositions {
		b.Run(pos.name, func(b *testing.B) {
			board := benchBoard(b, pos.fen)
			for b.Loop() {
				for sq := 0; sq < 64; sq++ {
					board.IsSquareAttacked(sq, White)
					board.IsSquareAttacked(sq, Black)
				}
			}
		})
	}
}

// Perft exercises everything together (move generation, legality, make/unmake),
// the same work a search does. It also reports positions per second.
func BenchmarkPerft(b *testing.B) {
	cases := []struct {
		name  string
		fen   string
		depth int
	}{
		{"start/depth4", startFEN, 4},
		{"kiwipete/depth3", benchPositions[1].fen, 3},
		{"endgame/depth4", benchPositions[3].fen, 4},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			board := benchBoard(b, c.fen)
			nodes := 0
			for b.Loop() {
				nodes += board.Perft(c.depth)
			}
			b.ReportMetric(float64(nodes)/b.Elapsed().Seconds(), "nodes/s")
		})
	}
}

func BenchmarkResult(b *testing.B) {
	for _, pos := range benchPositions {
		b.Run(pos.name, func(b *testing.B) {
			board := benchBoard(b, pos.fen)
			for b.Loop() {
				board.Result()
			}
		})
	}
}

// Evaluate runs at every position a search visits, so its cost per call matters.
func BenchmarkEvaluate(b *testing.B) {
	for _, pos := range benchPositions {
		b.Run(pos.name, func(b *testing.B) {
			board := benchBoard(b, pos.fen)
			for b.Loop() {
				board.Evaluate()
			}
		})
	}
}
