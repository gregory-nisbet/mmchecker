package core

import (
	"errors"
	"testing"
)

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
