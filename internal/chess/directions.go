package chess

const (
	north = iota
	south
	east
	west
	northEast
	northWest
	southEast
	southWest
)

var (
	directionOffsets = [8]int{8, -8, 1, -1, 9, 7, -7, -9}
	distanceToEdge   [64][8]int
	knightTargets    [64][]int
	kingTargets      [64][]int
	pawnTargets      [64][2][]int
)

type offset struct {
	dc, dr int
}

var knightPattern = []offset{
	{1, 2}, {2, 1}, {2, -1}, {1, -2},
	{-1, -2}, {-2, -1}, {-2, 1}, {-1, 2},
}

var kingPattern = []offset{
	{-1, -1}, {-1, 0}, {-1, 1}, {0, -1},
	{0, 1}, {1, -1}, {1, 0}, {1, 1},
}

var pawnOffset = [2][2]offset{
	{{-1, 1}, {1, 1}},
	{{-1, -1}, {1, -1}},
}

func colorConversion(color PieceId) int {
	return int(color/8 - 1)
}

func outsideOfBorder(row, col int) bool {
	return row < 0 || row > 7 || col < 0 || col > 7
}

func init() {
	for idx := range 64 {
		row := idx / 8
		col := idx % 8

		n := 7 - row
		s := row
		e := 7 - col
		w := col
		nw := min(n, w)
		ne := min(n, e)
		sw := min(s, w)
		se := min(s, e)
		distanceToEdge[idx] = [8]int{n, s, e, w, ne, nw, se, sw}

		for _, move := range knightPattern {
			if outsideOfBorder(row+move.dr, col+move.dc) {
				continue
			}
			knightTargets[idx] = append(knightTargets[idx], ConvertRowColToIdx(row+move.dr, col+move.dc))
		}

		for _, move := range kingPattern {
			if outsideOfBorder(row+move.dr, col+move.dc) {
				continue
			}
			kingTargets[idx] = append(kingTargets[idx], ConvertRowColToIdx(row+move.dr, col+move.dc))
		}

		for i := range 2 {
			pawnTarget := []int{}
			for _, move := range pawnOffset[i] {
				if outsideOfBorder(row+move.dr, col+move.dc) {
					continue
				}
				pawnTarget = append(pawnTarget, ConvertRowColToIdx(row+move.dr, col+move.dc))
			}
			pawnTargets[idx][i] = pawnTarget
		}
	}
}
