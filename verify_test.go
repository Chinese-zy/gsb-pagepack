package pagepack

import (
	"strings"
	"testing"
)

// rec builds a record from fixed bytes so every fixture below uses
// deterministic bytes instead of string literals.
func rec(bs ...byte) []byte { return append([]byte(nil), bs...) }

func fullPage(no int, slots ...byte) Page {
	p := Page{No: no, Slots: make([][]byte, len(slots))}
	for i, b := range slots {
		p.Slots[i] = rec(b)
	}
	return p
}

// pageWithHoles treats a 0 slot as a hole; every other byte is one record.
func pageWithHoles(no int, slots ...byte) Page {
	p := Page{No: no, Slots: make([][]byte, len(slots))}
	for i, b := range slots {
		if b != 0 {
			p.Slots[i] = rec(b)
		}
	}
	return p
}

func TestHolesRemainAndSeqStable(t *testing.T) {
	pages := []Page{pageWithHoles(1, 0x01, 0, 0x03)}
	once := Compact(pages)
	twice := Compact(once)
	if err := Verify(pages, once); err != nil {
		t.Fatal(err)
	}
	if len(once[0].Slots) != 3 || once[0].Slots[1] != nil {
		t.Fatalf("hole lost: %#v", once[0].Slots)
	}
	if once[0].Slots[0][0] != 0x01 || once[0].Slots[2][0] != 0x03 {
		t.Fatalf("sequence changed: %#v", once[0].Slots)
	}
	if err := Verify(once, twice); err != nil {
		t.Fatalf("recompact not stable: %v", err)
	}
}

func TestHalfPageDropped(t *testing.T) {
	before := []Page{
		fullPage(1, 0x11, 0x12, 0x13, 0x14),
		fullPage(2, 0x21, 0x22), // half page: fewer slots than the page width
		fullPage(3, 0x31, 0x32, 0x33, 0x34),
	}
	after := Compact(before)
	if len(after) != 2 {
		t.Fatalf("half page not dropped: %d pages", len(after))
	}
	if after[0].No != 1 || after[1].No != 3 {
		t.Fatalf("page order wrong: %d, %d", after[0].No, after[1].No)
	}
	if err := Verify(before, after); err != nil {
		t.Fatal(err)
	}
}

func TestCrossPageRecordStaysOnItsPage(t *testing.T) {
	// Page 1 is full; page 2 starts with a record that must never be pulled
	// onto page 1 (the old "spill" bug), and page 1's tail must not be eaten
	// by page 2's first record.
	before := []Page{
		fullPage(1, 0x01, 0x02, 0x03, 0x04),
		fullPage(2, 0x05, 0x06, 0x07, 0x08),
	}
	after := Compact(before)
	if len(after) != 2 {
		t.Fatalf("pages merged or split: %d pages", len(after))
	}
	if got := after[0].Slots[3][0]; got != 0x04 {
		t.Fatalf("page 1 tail eaten: seq 4 = %#02x", got)
	}
	if got := after[1].Slots[0][0]; got != 0x05 {
		t.Fatalf("page 2 first record moved: seq 1 = %#02x", got)
	}
	if err := Verify(before, after); err != nil {
		t.Fatal(err)
	}
}

func TestRecompactLeavesSeqAndHolesUnchanged(t *testing.T) {
	before := []Page{
		pageWithHoles(1, 0x01, 0, 0x03, 0),
		pageWithHoles(2, 0, 0x06, 0, 0x08),
	}
	once := Compact(before)
	twice := Compact(once)
	if err := Verify(once, twice); err != nil {
		t.Fatal(err)
	}
	for pi := range once {
		if len(twice[pi].Slots) != len(once[pi].Slots) {
			t.Fatalf("page %d slot count changed", once[pi].No)
		}
		for si := range once[pi].Slots {
			a, b := once[pi].Slots[si], twice[pi].Slots[si]
			if (a == nil) != (b == nil) || (a != nil && a[0] != b[0]) {
				t.Fatalf("page %d seq %d changed on recompact", once[pi].No, si+1)
			}
		}
	}
}

func TestCompletePagesRemainInOrder(t *testing.T) {
	before := []Page{
		fullPage(7, 0x0a, 0x0b, 0x0c),
		fullPage(9, 0x1a, 0x1b, 0x1c),
	}
	after := Compact(before)
	if after[0].No != 7 || after[1].No != 9 {
		t.Fatalf("order changed: %d, %d", after[0].No, after[1].No)
	}
	for pi := range before {
		for si := range before[pi].Slots {
			if after[pi].Slots[si][0] != before[pi].Slots[si][0] {
				t.Fatalf("page %d seq %d reordered", before[pi].No, si+1)
			}
		}
	}
	if err := Verify(before, after); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyReportsPageAndSeq(t *testing.T) {
	before := []Page{pageWithHoles(4, 0x01, 0, 0x03, 0x04)}
	cases := []struct {
		name  string
		after []Page
	}{
		{"hole filled", []Page{fullPage(4, 0x01, 0x02, 0x03, 0x04)}},
		{"record lost", []Page{pageWithHoles(4, 0x01, 0, 0, 0x04)}},
		{"record changed", []Page{pageWithHoles(4, 0x09, 0, 0x03, 0x04)}},
		{"half page kept", []Page{pageWithHoles(4, 0x01, 0)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Verify(before, tc.after)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			msg := err.Error()
			if !strings.Contains(msg, "page 4") {
				t.Fatalf("error missing page number: %q", msg)
			}
			if !strings.Contains(msg, "seq") {
				t.Fatalf("error missing sequence number: %q", msg)
			}
		})
	}
}

func TestParseUnchanged(t *testing.T) {
	raw := []byte{0xde, 0xad, 0xbe, 0xef}
	got := Parse(raw)
	if len(got) != 1 || got[0].No != 1 || len(got[0].Slots) != 1 {
		t.Fatalf("parse shape changed: %#v", got)
	}
	if got[0].Slots[0][0] != 0xde || got[0].Slots[0][3] != 0xef {
		t.Fatalf("parse bytes changed: % x", got[0].Slots[0])
	}
	if Parse(nil) != nil {
		t.Fatal("Parse(nil) must stay nil")
	}
}
