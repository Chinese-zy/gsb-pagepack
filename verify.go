package pagepack

import (
	"bytes"
	"fmt"
)

// Verify checks that compaction only dropped half pages and kept
// holes, page order, slot indexes, and record bytes of the rest.
// Failures report the page number and slot index.
func Verify(before, after []Page) error {
	want := make([]Page, 0, len(before))
	for _, p := range before {
		if len(p.Slots) == 0 || halfPage(p) {
			continue
		}
		want = append(want, p)
	}
	if len(after) != len(want) {
		return fmt.Errorf("page count: got %d, want %d", len(after), len(want))
	}
	for i, w := range want {
		a := after[i]
		if a.No != w.No {
			return fmt.Errorf("page %d: out of order, got page %d", w.No, a.No)
		}
		if len(a.Slots) != len(w.Slots) {
			return fmt.Errorf("page %d: slot count got %d, want %d", w.No, len(a.Slots), len(w.Slots))
		}
		for j := range w.Slots {
			switch {
			case w.Slots[j] == nil && a.Slots[j] != nil:
				return fmt.Errorf("page %d slot %d: hole filled", w.No, j)
			case w.Slots[j] != nil && a.Slots[j] == nil:
				return fmt.Errorf("page %d slot %d: record lost", w.No, j)
			case !bytes.Equal(a.Slots[j], w.Slots[j]):
				return fmt.Errorf("page %d slot %d: bytes changed", w.No, j)
			}
		}
	}
	return nil
}
