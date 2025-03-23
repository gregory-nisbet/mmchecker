package core

import (
	"testing"
)

func TestSymbols(t *testing.T) {
	t.Parallel()

	if ToSymbols(FromSymbols("a\ne")) != "a\ne" {
		t.Error("ToSymbols failed")
	}
}
