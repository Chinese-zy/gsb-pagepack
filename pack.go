package pagepack

// Page holds ordered records with possible holes (nil slots).
type Page struct {
	No   int
	Slots [][]byte // nil = hole
}

// Compact drops half pages incorrectly and renumbers — public parse kept separate.
func Compact(pages []Page) []Page {
	out := make([]Page, 0, len(pages))
	seq := 1
	for _, p := range pages {
		if len(p.Slots) == 0 {
			continue
		}
		// BUG: treat truncated last slot as full page always.
		np := Page{No: p.No, Slots: nil}
		for _, s := range p.Slots {
			if s == nil {
				// BUG: fill holes instead of keeping them.
				continue
			}
			// BUG: renumber on every compact.
			cp := append([]byte(nil), s...)
			np.Slots = append(np.Slots, cp)
			_ = seq
			seq++
		}
		// BUG: spill first of next page into previous when "full"
		out = append(out, np)
	}
	return out
}

func Parse(raw []byte) []Page {
	// Keep visible parse simple and stable.
	if len(raw) == 0 {
		return nil
	}
	return []Page{{No: 1,Slots: [][]byte{append([]byte(nil), raw...)}}}
}
