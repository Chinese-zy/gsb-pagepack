package pagepack

// Page holds ordered records with possible holes (nil slots).
type Page struct {
	No    int
	Slots [][]byte // nil = hole
}

// RecSize is the fixed byte length of a complete record.
const RecSize = 4

// halfPage reports whether p ends in a truncated (short) record.
func halfPage(p Page) bool {
	for i := len(p.Slots) - 1; i >= 0; i-- {
		if p.Slots[i] != nil {
			return len(p.Slots[i]) != RecSize
		}
	}
	return false
}

// Compact drops half pages and keeps holes, page order, and slot
// indexes stable, so re-compacting a page changes nothing.
func Compact(pages []Page) []Page {
	out := make([]Page, 0, len(pages))
	for _, p := range pages {
		if len(p.Slots) == 0 {
			continue
		}
		if halfPage(p) {
			continue
		}
		np := Page{No: p.No, Slots: make([][]byte, len(p.Slots))}
		for i, s := range p.Slots {
			if s == nil {
				continue
			}
			np.Slots[i] = append([]byte(nil), s...)
		}
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
