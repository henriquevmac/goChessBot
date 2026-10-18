package chess

import (
	"slices"
	"testing"
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// sameSquares reports whether a and b contain the same squares, in any order.
func sameSquares(a, b []int) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

// File and rank change for each direction, in the same order as the direction constants.
var directionDeltas = [8]offset{
	north:     {0, 1},
	south:     {0, -1},
	east:      {1, 0},
	west:      {-1, 0},
	northEast: {1, 1},
	northWest: {-1, 1},
	southEast: {1, -1},
	southWest: {-1, -1},
}

func TestDirectionOffsets(t *testing.T) {
	for dir, d := range directionDeltas {
		if want := d.dr*8 + d.dc; directionOffsets[dir] != want {
			t.Errorf("directionOffsets[%d] = %d, want %d", dir, directionOffsets[dir], want)
		}
	}
}

// Walk each direction one step at a time and count how many steps stay on the board.
func TestDistanceToEdgeAllSquares(t *testing.T) {
	for sq := 0; sq < 64; sq++ {
		for dir, d := range directionDeltas {
			want := 0
			col, row := sq%8+d.dc, sq/8+d.dr
			for col >= 0 && col < 8 && row >= 0 && row < 8 {
				want++
				col += d.dc
				row += d.dr
			}
			if got := distanceToEdge[sq][dir]; got != want {
				t.Errorf("distanceToEdge[%s][%d] = %d, want %d", squareName(sq), dir, got, want)
			}
		}
	}
}

func TestDistanceToEdgeExamples(t *testing.T) {
	tests := []struct {
		sq   int
		want [8]int // n, s, e, w, ne, nw, se, sw
	}{
		{0, [8]int{7, 0, 7, 0, 7, 0, 0, 0}},  // a1
		{7, [8]int{7, 0, 0, 7, 0, 7, 0, 0}},  // h1
		{9, [8]int{6, 1, 6, 1, 6, 1, 1, 1}},  // b2
		{63, [8]int{0, 7, 0, 7, 0, 0, 0, 7}}, // h8
		{28, [8]int{4, 3, 3, 4, 3, 4, 3, 3}}, // e4
	}
	for _, tt := range tests {
		if got := distanceToEdge[tt.sq]; got != tt.want {
			t.Errorf("distanceToEdge[%s] = %v, want %v", squareName(tt.sq), got, tt.want)
		}
	}
}

// Compare a target table against every square that is a valid jump by the given rule.
func checkTargets(t *testing.T, name string, table *[64][]int, isTarget func(df, dr int) bool) {
	t.Helper()
	for from := 0; from < 64; from++ {
		var want []int
		for to := 0; to < 64; to++ {
			if isTarget(to%8-from%8, to/8-from/8) {
				want = append(want, to)
			}
		}
		if !sameSquares(table[from], want) {
			t.Errorf("%s[%s] = %v, want %v", name, squareName(from), table[from], want)
		}
	}
}

func TestKnightTargetsAllSquares(t *testing.T) {
	checkTargets(t, "knightTargets", &knightTargets, func(df, dr int) bool {
		return (abs(df) == 1 && abs(dr) == 2) || (abs(df) == 2 && abs(dr) == 1)
	})
}

func TestKingTargetsAllSquares(t *testing.T) {
	checkTargets(t, "kingTargets", &kingTargets, func(df, dr int) bool {
		return max(abs(df), abs(dr)) == 1
	})
}

func TestJumpTargetExamples(t *testing.T) {
	tests := []struct {
		name  string
		table *[64][]int
		sq    int
		want  []int
	}{
		{"knightTargets", &knightTargets, 0, []int{10, 17}},                         // a1 → c2, b3
		{"knightTargets", &knightTargets, 1, []int{11, 16, 18}},                     // b1 → d2, a3, c3
		{"knightTargets", &knightTargets, 7, []int{13, 22}},                         // h1 → f2, g3
		{"knightTargets", &knightTargets, 28, []int{11, 13, 18, 22, 34, 38, 43, 45}}, // e4: all 8
		{"kingTargets", &kingTargets, 0, []int{1, 8, 9}},                            // a1: corner, 3
		{"kingTargets", &kingTargets, 4, []int{3, 5, 11, 12, 13}},                   // e1: edge, 5
		{"kingTargets", &kingTargets, 28, []int{19, 20, 21, 27, 29, 35, 36, 37}},    // e4: all 8
	}
	for _, tt := range tests {
		if got := tt.table[tt.sq]; !sameSquares(got, tt.want) {
			t.Errorf("%s[%s] = %v, want %v", tt.name, squareName(tt.sq), got, tt.want)
		}
	}
}

func TestPawnTargetsAllSquares(t *testing.T) {
	white, black := colorConversion(White), colorConversion(Black)
	for from := 0; from < 64; from++ {
		var wantWhite, wantBlack []int
		for to := 0; to < 64; to++ {
			df, dr := to%8-from%8, to/8-from/8
			if abs(df) == 1 && dr == 1 {
				wantWhite = append(wantWhite, to)
			}
			if abs(df) == 1 && dr == -1 {
				wantBlack = append(wantBlack, to)
			}
		}
		if got := pawnTargets[from][white]; !sameSquares(got, wantWhite) {
			t.Errorf("white pawnTargets[%s] = %v, want %v", squareName(from), got, wantWhite)
		}
		if got := pawnTargets[from][black]; !sameSquares(got, wantBlack) {
			t.Errorf("black pawnTargets[%s] = %v, want %v", squareName(from), got, wantBlack)
		}
	}
}

func TestPawnTargetExamples(t *testing.T) {
	tests := []struct {
		color PieceId
		sq    int
		want  []int
	}{
		{White, 28, []int{35, 37}}, // e4 → d5, f5
		{Black, 36, []int{27, 29}}, // e5 → d4, f4
		{White, 15, []int{22}},     // h2 → g3 only, no wrap to the a-file
		{White, 8, []int{17}},      // a2 → b3 only, no wrap to the h-file
		{Black, 55, []int{46}},     // h7 → g6 only
		{Black, 48, []int{41}},     // a7 → b6 only
	}
	for _, tt := range tests {
		if got := pawnTargets[tt.sq][colorConversion(tt.color)]; !sameSquares(got, tt.want) {
			t.Errorf("pawnTargets[%s][color %d] = %v, want %v", squareName(tt.sq), tt.color, got, tt.want)
		}
	}
}

func TestColorConversion(t *testing.T) {
	tests := []struct {
		piece PieceId
		want  int
	}{
		{White, 0},
		{Black, 1},
		{White | Knight, 0},
		{Black | Queen, 1},
	}
	for _, tt := range tests {
		if got := colorConversion(tt.piece); got != tt.want {
			t.Errorf("colorConversion(%d) = %d, want %d", tt.piece, got, tt.want)
		}
	}
}
