package core

import "testing"

func TestAddC(t *testing.T) {
	t.Parallel()

	mm := NewMM(nil)
	mm.AddC("a")

	_, ok := mm.Constants["a"]
	if !ok {
		t.Error("AddC failed")
	}

	sym, va, constant := mm.LookupSymbolByName("a")
	if sym == nil {
		t.Error("LookupSymbolByName failed")
	}
	if va != nil {
		t.Error("LookupSymbolByName failed")
	}
	if constant == nil {
		t.Error("LookupSymbolByName failed")
	}
}
