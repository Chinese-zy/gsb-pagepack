package pagepack

import (
	"bytes"
	"fmt"
)

// Verify checks a compaction result page by page and record by record
// instead of only asserting that the last record exists. It models what
// Compact must do: truncated (half) pages disappear, complete pages survive
// in their original order, and on every surviving page the slot numbers
// (sequence), holes and record bytes stay exactly where they were.
//
// Every failure message names the page number (Page.No) and the 1-based
// slot number (sequence) involved.
func Verify(before, after []Page) error {
	width := fullSlotCount(before)

	want := make([]Page, 0, len(before))
	for _, p := range before {
		if len(p.Slots) >= width {
			want = append(want, p)
		}
	}

	if len(after) != len(want) {
		detail := ""
		if len(after) > len(want) {
			detail = fmt.Sprintf(" (unexpected page %d seq 1)", after[len(want)].No)
		} else {
			detail = fmt.Sprintf(" (missing page %d seq 1)", want[len(after)].No)
		}
		return fmt.Errorf("page count mismatch: want %d pages, got %d%s", len(want), len(after), detail)
	}

	for pi := range want {
		w, g := want[pi], after[pi]
		if g.No != w.No {
			return fmt.Errorf("page %d seq 1: page number changed to %d", w.No, g.No)
		}
		if len(g.Slots) != len(w.Slots) {
			firstDiff := len(g.Slots) + 1
			if len(w.Slots)+1 < firstDiff {
				firstDiff = len(w.Slots) + 1
			}
			return fmt.Errorf("page %d seq %d: slot count changed from %d to %d", w.No, firstDiff, len(w.Slots), len(g.Slots))
		}
		for si := range w.Slots {
			seq := si + 1
			switch {
			case w.Slots[si] == nil && g.Slots[si] != nil:
				return fmt.Errorf("page %d seq %d: hole was filled", w.No, seq)
			case w.Slots[si] != nil && g.Slots[si] == nil:
				return fmt.Errorf("page %d seq %d: record lost", w.No, seq)
			case w.Slots[si] != nil && !bytes.Equal(w.Slots[si], g.Slots[si]):
				return fmt.Errorf("page %d seq %d: record changed from %x to %x", w.No, seq, w.Slots[si], g.Slots[si])
			}
		}
	}
	return nil
}
