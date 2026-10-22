package bot

import "goChessBot/internal/chess"

type ttFlag uint8

const (
	ttExact ttFlag = iota
	ttLower
	ttUpper
)

type ttEntry struct {
	key      uint64
	bestMove chess.Move
	depth    int
	score    int
	flag     ttFlag
}

type transpositionTable struct {
	entries []ttEntry
	mask    uint64
}

func newTranspositionTable(sizeLog2 int) *transpositionTable {
	entries := make([]ttEntry, 1<<sizeLog2)
	mask := uint64(1)<<sizeLog2 - 1
	return &transpositionTable{entries: entries, mask: mask}
}

func scoreToTT(recScore int, ply int) int {
	score := recScore
	if score > mateValue-1000 {
		score += ply
	} else if score < -(mateValue - 1000) {
		score -= ply
	}
	return score
}

func scoreFromTT(recScore int, ply int) int {
	score := recScore
	if score > mateValue-1000 {
		score -= ply
	} else if score < -(mateValue - 1000) {
		score += ply
	}
	return score
}

func (tt *transpositionTable) probe(key uint64) (ttEntry, bool) {
	entry := tt.entries[key&tt.mask]
	if entry.key == key {
		return entry, true
	}
	return ttEntry{}, false
}

func (tt *transpositionTable) store(e ttEntry) {
	tt.entries[e.key&tt.mask] = e
}
