package pagepack

// Page holds ordered records with possible holes (nil slots).
type Page struct {
	No    int
	Slots [][]byte // nil = hole
}

// fullSlotCount reports the fixed page width: a complete page carries this
// many slots, and a page with fewer slots is a truncated (half) page. The
// width is taken from the widest page present so callers drive it with
// fixed-size fixtures instead of a hidden constant.
func fullSlotCount(pages []Page) int {
	width := 0
	for _, p := range pages {
		if len(p.Slots) > width {
			width = len(p.Slots)
		}
	}
	return width
}

// Compact packs pages while preserving record identity:
//   - holes (nil slots) stay at their original slot number;
//   - surviving records keep their original slot number (sequence);
//   - truncated (half) pages, including empty ones, are dropped entirely;
//   - records never cross a page boundary: a full page never eats the tail
//     of another page, and the first record of a fresh page stays put;
//   - compacting the output again changes neither page order, slot numbers
//     nor holes.
func Compact(pages []Page) []Page {
	width := fullSlotCount(pages)
	out := make([]Page, 0, len(pages))
	for _, p := range pages {
		if len(p.Slots) < width {
			// Truncated (half) page: drop it without spilling anything.
			continue
		}
		np := Page{No: p.No, Slots: make([][]byte, len(p.Slots))}
		for i, s := range p.Slots {
			if s == nil {
				// Hole stays at its original slot number.
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
	return []Page{{No: 1, Slots: [][]byte{append([]byte(nil), raw...)}}}
}
