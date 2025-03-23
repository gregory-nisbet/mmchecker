package core

import "testing"

func TestIsHypothesis(t *testing.T) {
	t.Parallel()

	if !IsHypothesis(FullStmt{SType: "$e"}) {
		t.Error("IsHypothesis failed")
	}
}
