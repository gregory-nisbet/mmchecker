package core

import (
	"errors"
	"fmt"
	"testing"
)

func TestEOF(t *testing.T) {
	t.Parallel()

	wrappedError := fmt.Errorf("wrapped error: %w", EOF)
	if !IsEOF(wrappedError) {
		t.Error("IsEOF failed")
	}
}

func TestIOError(t *testing.T) {
	t.Parallel()

	e := IOError{errors.New("hi")}
	if AsIOError(e) == nil {
		t.Error("AsIOError failed")
	}
}

func TestMMError(t *testing.T) {
	t.Parallel()

	e := MMError{errors.New("hi")}
	if AsMMError(e) == nil {
		t.Error("AsMMError failed")
	}
}
