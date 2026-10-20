package chess

import "testing"

// Every FEN CreateBoard accepts must come back unchanged from FEN().
var roundTripFENs = []string{
	startFEN,
	"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
	"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2",
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
	"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
	"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
	"r3k2r/8/8/8/8/8/8/R3K2R w Kq - 12 40",
	"r3k2r/8/8/8/8/8/8/R3K2R b Qk - 3 7",
	"4k3/8/8/8/8/8/8/4K3 w - - 99 120",
	"8/8/8/8/8/8/8/8 w - - 0 1",
	"7k/5Q2/6K1/8/8/8/8/8 b - - 0 1",
}

func TestFENRoundTrip(t *testing.T) {
	for _, fen := range roundTripFENs {
		b := mustBoard(t, fen)
		if got := b.FEN(); got != fen {
			t.Errorf("CreateBoard(%q).FEN()\n got  %q\n want %q", fen, got, fen)
		}
	}
}

// FEN() must describe the position after moves, including the en passant square,
// lost castling rights, the clocks and the side to move.
func TestFENAfterMoves(t *testing.T) {
	tests := []struct {
		name  string
		fen   string
		moves []string
		want  string
	}{
		{"double push", startFEN, []string{"e2e4"},
			"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"},
		{"black reply, fullmove increments", startFEN, []string{"e2e4", "c7c5"},
			"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2"},
		{"knight move, halfmove clock", startFEN, []string{"g1f3"},
			"rnbqkbnr/pppppppp/8/8/8/5N2/PPPPPPPP/RNBQKB1R b KQkq - 1 1"},
		{"castling removes both rights", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"e1g1"},
			"r3k2r/8/8/8/8/8/8/R4RK1 b kq - 1 1"},
		{"all castling rights gone", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", []string{"e1g1", "e8c8"},
			"2kr3r/8/8/8/8/8/8/R4RK1 w - - 2 2"},
		{"promotion", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", []string{"e7e8n"},
			"4N3/8/8/8/8/8/8/k3K3 b - - 0 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			for _, uci := range tt.moves {
				b.MakeMove(moveByUCI(t, b, uci))
			}
			if got := b.FEN(); got != tt.want {
				t.Errorf("after %v:\n got  %q\n want %q", tt.moves, got, tt.want)
			}
		})
	}
}

// FENs without the last fields get the defaults, and FEN() always writes all six fields.
func TestFENFillsInMissingFields(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", startFEN},
		{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1"},
		{"4k3/8/8/8/8/8/8/4K3 b", "4k3/8/8/8/8/8/8/4K3 b - - 0 1"},
		{"r3k2r/8/8/8/8/8/8/R3K2R w KQkq -", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"},
	}
	for _, tt := range tests {
		b := mustBoard(t, tt.in)
		if got := b.FEN(); got != tt.want {
			t.Errorf("CreateBoard(%q).FEN()\n got  %q\n want %q", tt.in, got, tt.want)
		}
	}
}

// Every FEN used by the engine's other tests must keep loading, including
// positions without kings, which many move generation tests rely on.
func TestCreateBoardAcceptsValidFENs(t *testing.T) {
	valid := append([]string{
		"",                                 // empty means the start position
		"8/8/8/8/3N4/8/8/8 w - - 0 1",      // no kings at all
		"4k3/8/8/8/8/8/8/8 b - - 0 1",      // only a black king
		"  " + startFEN + "  ",             // surrounding spaces
		"r3k2r/8/8/8/8/8/8/R3K2R w KQkq -", // missing clocks
	}, roundTripFENs...)
	for _, fen := range valid {
		if _, err := CreateBoard(fen); err != nil {
			t.Errorf("CreateBoard(%q) returned error %v, want success", fen, err)
		}
	}
}

func TestCreateBoardRejectsInvalidFENs(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		// Piece placement
		{"only spaces", "   "},
		{"unknown piece letter", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNX w KQkq - 0 1"},
		{"rank with too many squares", "rnbqkbnr/ppppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"rank with too many squares (digits)", "rnbqkbnr/pppppppp/8/8/44P/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"rank with too few squares", "rnbqkbnr/ppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"digit 9", "rnbqkbnr/pppppppp/9/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"digit 0", "rnbqkbnr/pppppppp/08/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"only 7 ranks", "rnbqkbnr/pppppppp/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"9 ranks", "rnbqkbnr/pppppppp/8/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"two white kings", "4k3/8/8/8/8/8/8/3KK3 w - - 0 1"},
		{"two black kings", "3kk3/8/8/8/8/8/8/4K3 w - - 0 1"},

		// Side to move
		{"side to move is not w or b", "4k3/8/8/8/8/8/8/4K3 x - - 0 1"},
		{"side to move in capitals", "4k3/8/8/8/8/8/8/4K3 W - - 0 1"},

		// Castling rights
		{"unknown castling letter", "r3k2r/8/8/8/8/8/8/R3K2R w KQkx - 0 1"},
		{"castling field mixing - and letters", "r3k2r/8/8/8/8/8/8/R3K2R w K- - 0 1"},

		// En passant square
		{"en passant file out of range", "4k3/8/8/8/4P3/8/8/4K3 b - i3 0 1"},
		{"en passant rank out of range", "4k3/8/8/8/4P3/8/8/4K3 b - e9 0 1"},
		{"en passant on an impossible rank", "4k3/8/8/8/4P3/8/8/4K3 b - e4 0 1"},
		{"en passant too long", "4k3/8/8/8/4P3/8/8/4K3 b - e33 0 1"},

		// Clocks
		{"halfmove clock not a number", "4k3/8/8/8/8/8/8/4K3 w - - x 1"},
		{"negative halfmove clock", "4k3/8/8/8/8/8/8/4K3 w - - -1 1"},
		{"fullmove number not a number", "4k3/8/8/8/8/8/8/4K3 w - - 0 y"},
		{"fullmove number 0", "4k3/8/8/8/8/8/8/4K3 w - - 0 0"},

		// The example that slipped through before
		{"nonsense", "bad fen x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if b, err := CreateBoard(tt.fen); err == nil {
				t.Errorf("CreateBoard(%q) accepted an invalid FEN (board %v)", tt.fen, b.Square)
			}
		})
	}
}
