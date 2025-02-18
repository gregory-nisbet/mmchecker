package mmchecker

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

// Function makeDiff switches the argument order to put
// got-before-want, in the Go style.
//
// Also, use generics.
func makeDiff[T any](got T, want T) error {
	if diff := cmp.Diff(want, got); diff != "" {
		return fmt.Errorf("unexpected diff (-want +got): %s", diff)
	}
	return nil
}

func errContains(e error, msg string) error {
	if e == nil {
		if msg == "" {
			return nil
		}
		return fmt.Errorf("expected error containing %q", msg)
	}
	content := e.Error()
	if content == "" {
		panic("non-nil error stringifies to empty string!")
	}
	if msg == "" {
		return fmt.Errorf("unexpected error: %w", e)
	}
	if strings.Contains(content, msg) {
		return nil
	}
	return fmt.Errorf("error %q does not contain %q", e, msg)
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
