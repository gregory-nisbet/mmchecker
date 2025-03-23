package core

import "testing"

func TestIsHypothesis(t *testing.T) {
	t.Parallel()

	if !IsHypothesis(FullStmt{SType: "$e"}) {
		t.Error("IsHypothesis failed")
	}
}

func TestIsAssertion(t *testing.T) {
	t.Parallel()

	if !IsAssertion(FullStmt{SType: "$a"}) {
		t.Error("IsHypothesis failed")
	}
}
