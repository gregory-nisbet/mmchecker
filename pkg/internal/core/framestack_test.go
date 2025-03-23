package core

import "testing"

func TestFrameStack(t *testing.T) {
	t.Parallel()

	framestack := NewFrameStack()

	framestack.Push()

	if len(framestack.Frames) != 1 {
		t.Error("framestack push failed")
	}
}
