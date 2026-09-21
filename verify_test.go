package pagepack

import "testing"

func TestHolesRemainAndSeqStable(t *testing.T) {
	pages := []Page{{No: 1, Slots: [][]byte{[]byte("a"), nil, []byte("c")}}}
	once := Compact(pages)
	twice := Compact(once)
	if err := Verify(pages, once); err != nil {
		t.Fatal(err)
	}
	// Desired: holes remain, sequence unchanged across re-compact.
	if len(once[0].Slots) != 3 || once[0].Slots[1] != nil {
		t.Fatalf("hole lost: %#v", once[0].Slots)
	}
	if len(twice[0].Slots) != len(once[0].Slots) {
		t.Fatalf("recompact changed slots")
	}
}
