package chess

import "testing"

func TestFindLegalMove(t *testing.T) {
	tests := []struct {
		name      string
		fen       string
		from, to  string
		promotion PieceId
		wantOK    bool
		wantUCI   string
		wantFlag  MoveFlag
	}{
		// Normal moves and their flags
		{"pawn single push", startFEN, "e2", "e3", Queen, true, "e2e3", Normal},
		{"pawn double push gets its flag", startFEN, "e2", "e4", Queen, true, "e2e4", DoublePawn},
		{"knight move", startFEN, "g1", "f3", Queen, true, "g1f3", Normal},
		{"capture gets its flag", "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
			"e4", "d5", Queen, true, "e4d5", Capture},
		{"en passant gets its flag", "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2", "e5", "d6", Queen, true, "e5d6", EnPassant},
		{"castling by moving the king two squares", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			"e1", "g1", Queen, true, "e1g1", KingCastle},
		{"queenside castling", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			"e1", "c1", Queen, true, "e1c1", QueenCastle},

		// Promotion: the requested piece is chosen
		{"promotion to queen", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", "e7", "e8", Queen, true, "e7e8q", PromoQ},
		{"promotion to knight", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", "e7", "e8", Knight, true, "e7e8n", PromoKn},
		{"promotion to rook", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", "e7", "e8", Rook, true, "e7e8r", PromoR},
		{"promotion to bishop", "8/4P3/8/8/8/8/8/k3K3 w - - 0 1", "e7", "e8", Bishop, true, "e7e8b", PromoB},
		{"capture promotion to queen", "3r4/4P3/8/8/8/8/8/k3K3 w - - 0 1", "e7", "d8", Queen, true, "e7d8q", PromoQCap},
		{"black promotion", "4k3/8/8/8/8/8/3p4/4K3 b - - 0 1", "d2", "d1", Queen, true, "d2d1q", PromoQ},

		// Not legal: no move found
		{"pawn cannot move three squares", startFEN, "e2", "e5", Queen, false, "", 0},
		{"cannot capture own piece", startFEN, "d1", "d2", Queen, false, "", 0},
		{"empty from square", startFEN, "e4", "e5", Queen, false, "", 0},
		{"not your turn", startFEN, "e7", "e5", Queen, false, "", 0},
		{"same square", startFEN, "e2", "e2", Queen, false, "", 0},
		{"pinned piece cannot move", "4k3/4r3/8/8/8/8/4B3/4K3 w - - 0 1", "e2", "d3", Queen, false, "", 0},
		{"king cannot move into check", "4k3/8/8/8/8/8/4r3/4K3 w - - 0 1", "e1", "d2", Queen, false, "", 0},
		{"castling through check", "5rk1/8/8/8/8/8/8/R3K2R w KQ - 0 1", "e1", "g1", Queen, false, "", 0},
		{"en passant that exposes the king", "8/8/8/KPp4r/8/8/8/7k w - c6 0 1", "b5", "c6", Queen, false, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustBoard(t, tt.fen)
			before := *b

			m, ok := b.FindLegalMove(sq(tt.from), sq(tt.to), tt.promotion)
			if ok != tt.wantOK {
				t.Fatalf("FindLegalMove(%s, %s) ok = %v, want %v (move %s)", tt.from, tt.to, ok, tt.wantOK, m)
			}
			if ok {
				if m.String() != tt.wantUCI {
					t.Errorf("got move %s, want %s", m, tt.wantUCI)
				}
				if m.Flag() != tt.wantFlag {
					t.Errorf("%s: flag = %d, want %d", m, m.Flag(), tt.wantFlag)
				}
			}
			if diff := boardDiff(b, &before); diff != "" {
				t.Errorf("FindLegalMove changed the board:\n  %s", diff)
			}
		})
	}
}

// The move found must be one of GenerateLegalMoves, and every legal move must be findable.
func TestFindLegalMoveMatchesLegalMoves(t *testing.T) {
	for _, fen := range makeUnmakeFENs {
		b := mustBoard(t, fen)
		for _, want := range b.GenerateLegalMoves(nil) {
			promotion := Queen
			if want.IsPromotion() {
				promotion = want.GetPromotion()
			}
			got, ok := b.FindLegalMove(want.From(), want.To(), promotion)
			if !ok || got != want {
				t.Errorf("%s: FindLegalMove(%s, %s, %d) = %s, %v; want %s, true",
					fen, squareName(want.From()), squareName(want.To()), promotion, got, ok, want)
			}
		}
	}
}
