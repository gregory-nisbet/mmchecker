package core

import "testing"

func TestToks(t *testing.T) {
	t.Parallel()

	toks, err := NewToks("", ToTokens("a e\na e"))
	if err != nil {
		t.Error("NewToks failed")
	}
	if toks == nil {
		t.Error("NewToks failed")
	}
	if toks.getLastFile() == nil {
		t.Error("getLastFile failed")
	}
	if err := toks.popFile(); err != nil {
		t.Errorf("popFile failed: %v", err)
	}
}
