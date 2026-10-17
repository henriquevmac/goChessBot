package chess

import "testing"

func TestMoveRoundTrip(t *testing.T) {
	for from := 0; from < 64; from++ {
		for to := 0; to < 64; to++ {
			for flag := MoveFlag(0); flag < 16; flag++ {
				m := CreateMove(from, to, flag)
				if m.From() != from || m.To() != to || m.Flag() != flag {
					t.Fatalf("CreateMove(%d, %d, %d) unpacked to (%d, %d, %d)",
						from, to, flag, m.From(), m.To(), m.Flag())
				}
			}
		}
	}
}

func TestCreateMoveEncoding(t *testing.T) {
	if got := CreateMove(52, 60, PromoQ); got != 0xBF34 {
		t.Errorf("CreateMove(52, 60, PromoQ) = %#04x, want 0xbf34", uint16(got))
	}
	if got := CreateMove(0, 0, Normal); got != 0 {
		t.Errorf("CreateMove(0, 0, Normal) = %#04x, want 0 (the \"no move\" value)", uint16(got))
	}
}

func TestMoveFlags(t *testing.T) {
	tests := []struct {
		flag        MoveFlag
		isCapture   bool
		isPromotion bool
	}{
		{Normal, false, false},
		{DoublePawn, false, false},
		{KingCastle, false, false},
		{QueenCastle, false, false},
		{Capture, true, false},
		{EnPassant, true, false},
		{PromoKn, false, true},
		{PromoB, false, true},
		{PromoR, false, true},
		{PromoQ, false, true},
		{PromoKnCap, true, true},
		{PromoBCap, true, true},
		{PromoRCap, true, true},
		{PromoQCap, true, true},
	}
	for _, from := range []int{0, 15, 63} {
		for _, tt := range tests {
			m := CreateMove(from, 0, tt.flag)
			if got := m.IsCapture(); got != tt.isCapture {
				t.Errorf("from %d, flag %d: IsCapture() = %v, want %v", from, tt.flag, got, tt.isCapture)
			}
			if got := m.IsPromotion(); got != tt.isPromotion {
				t.Errorf("from %d, flag %d: IsPromotion() = %v, want %v", from, tt.flag, got, tt.isPromotion)
			}
		}
	}
}

func TestSquareName(t *testing.T) {
	files := "abcdefgh"
	ranks := "12345678"
	for sq := 0; sq < 64; sq++ {
		want := string(files[sq%8]) + string(ranks[sq/8])
		if got := squareName(sq); got != want {
			t.Errorf("squareName(%d) = %q, want %q", sq, got, want)
		}
	}
}

func TestMoveString(t *testing.T) {
	tests := []struct {
		from, to int
		flag     MoveFlag
		want     string
	}{
		{12, 28, DoublePawn, "e2e4"},
		{6, 21, Normal, "g1f3"},
		{4, 6, KingCastle, "e1g1"},
		{4, 2, QueenCastle, "e1c1"},
		{60, 62, KingCastle, "e8g8"},
		{60, 58, QueenCastle, "e8c8"},
		{36, 43, EnPassant, "e5d6"},
		{28, 35, Capture, "e4d5"},
		{0, 63, Normal, "a1h8"},
		{52, 60, PromoQ, "e7e8q"},
		{52, 60, PromoR, "e7e8r"},
		{52, 60, PromoB, "e7e8b"},
		{52, 60, PromoKn, "e7e8n"},
		{52, 61, PromoQCap, "e7f8q"},
		{9, 0, PromoKnCap, "b2a1n"},
		{0, 0, Normal, "a1a1"},
	}
	for _, tt := range tests {
		m := CreateMove(tt.from, tt.to, tt.flag)
		if got := m.String(); got != tt.want {
			t.Errorf("CreateMove(%d, %d, %d).String() = %q, want %q", tt.from, tt.to, tt.flag, got, tt.want)
		}
	}
}
