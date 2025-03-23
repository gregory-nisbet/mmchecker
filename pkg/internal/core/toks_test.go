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

func TestReadc(t *testing.T) {
	t.Parallel()

	toks, err := NewToks("", ToTokens("$( hi hi hi $)\na e"))
	if err != nil {
		t.Error("NewToks failed")
	}
	if toks == nil {
		t.Error("NewToks failed")
	}

	tok, err := toks.Readc()
	if err != nil {
		t.Error("readc failed")
	}

	if tok != "a" {
		t.Errorf("bad value of tok: %q", tok)
	}
}
