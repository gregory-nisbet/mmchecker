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
}
