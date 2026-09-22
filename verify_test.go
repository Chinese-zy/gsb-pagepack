package pagepack

import (
	"reflect"
	"strings"
	"testing"
)

// fixedRec builds a complete fixed-size record from one byte.
func fixedRec(b byte) []byte {
	return []byte{b, b, b, b}
}

func TestHolesRemainAndSeqStable(t *testing.T) {
	pages := []Page{{No: 1, Slots: [][]byte{fixedRec('a'), nil, fixedRec('c')}}}
	once := Compact(pages)
	if err := Verify(pages, once); err != nil {
		t.Fatal(err)
	}
	if len(once[0].Slots) != 3 || once[0].Slots[1] != nil {
		t.Fatalf("hole lost: %#v", once[0].Slots)
	}
	twice := Compact(once)
	if err := Verify(once, twice); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(once, twice) {
		t.Fatalf("recompact changed slots or holes: %#v -> %#v", once, twice)
	}
}

func TestHalfPageDroppedAndOrderKept(t *testing.T) {
	pages := []Page{
		{No: 1, Slots: [][]byte{fixedRec('a'), fixedRec('b')}},
		{No: 2, Slots: [][]byte{fixedRec('c'), []byte{'x'}}}, // truncated tail
		{No: 3, Slots: [][]byte{fixedRec('d')}},
	}
	got := Compact(pages)
	if err := Verify(pages, got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].No != 1 || got[1].No != 3 {
		t.Fatalf("half page kept or order changed: %#v", got)
	}
}

func TestCrossPageRecordStaysOnOwnPage(t *testing.T) {
	p1 := Page{No: 1, Slots: [][]byte{fixedRec('a'), fixedRec('b')}}
	p2 := Page{No: 2, Slots: [][]byte{fixedRec('c'), fixedRec('d')}}
	got := Compact([]Page{p1, p2})
	if len(got) != 2 {
		t.Fatalf("pages merged or split: %#v", got)
	}
	if !reflect.DeepEqual(got[0].Slots, p1.Slots) {
		t.Fatalf("page 1 tail eaten: %#v", got[0].Slots)
	}
	if !reflect.DeepEqual(got[1].Slots, p2.Slots) {
		t.Fatalf("page 2 first record moved: %#v", got[1].Slots)
	}
}

func TestVerifyReportsPageAndSlot(t *testing.T) {
	before := []Page{{No: 7, Slots: [][]byte{fixedRec('a'), nil, fixedRec('c')}}}

	filled := []Page{{No: 7, Slots: [][]byte{fixedRec('a'), fixedRec('z'), fixedRec('c')}}}
	if err := Verify(before, filled); err == nil ||
		!strings.Contains(err.Error(), "page 7 slot 1") {
		t.Fatalf("filled hole not reported with page/slot: %v", err)
	}

	renumbered := []Page{{No: 7, Slots: [][]byte{fixedRec('b'), nil, fixedRec('c')}}}
	if err := Verify(before, renumbered); err == nil ||
		!strings.Contains(err.Error(), "page 7 slot 0") {
		t.Fatalf("renumber not reported with page/slot: %v", err)
	}

	half := []Page{
		{No: 1, Slots: [][]byte{fixedRec('a')}},
		{No: 2, Slots: [][]byte{[]byte{'x'}}},
	}
	if err := Verify(half, half); err == nil {
		t.Fatal("kept half page passed verify")
	}

	split := []Page{
		{No: 1, Slots: [][]byte{fixedRec('a')}},
		{No: 2, Slots: [][]byte{fixedRec('c')}},
	}
	moved := []Page{{No: 1, Slots: [][]byte{fixedRec('a'), fixedRec('c')}}}
	if err := Verify(split, moved); err == nil {
		t.Fatal("cross-page move passed verify")
	}
}
