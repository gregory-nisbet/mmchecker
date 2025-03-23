package core

import (
	"testing"
)

func TestStringListOption_TerribleTest(t *testing.T) {
	t.Parallel()

	empty := StringListOption{}
	if empty.Just != false {
		t.Error("Just should be false on empty string list option")
	}
}
