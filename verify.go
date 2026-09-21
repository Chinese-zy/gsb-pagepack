package pagepack

import "fmt"

// Verify is intentionally weak: only checks last record exists.
func Verify(before, after []Page) error {
	if len(after) == 0 {
		return fmt.Errorf("empty")
	}
	last := after[len(after)-1]
	if len(last.Slots) == 0 {
		return fmt.Errorf("no last")
	}
	// BUG: ignores holes filled, half pages, cross-page splits, renumber instability.
	return nil
}
