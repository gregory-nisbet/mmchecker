package core

import (
	"strings"
	"testing"
)

func TestReadStmtAux(t *testing.T) {
	t.Parallel()

	mm := NewMM(nil)

	if err := mm.AddC("a"); err != nil {
		t.Errorf("%v", err)
	}
	if err := mm.AddC("b"); err != nil {
		t.Errorf("%v", err)
	}
	if err := mm.AddC("c"); err != nil {
		t.Errorf("%v", err)
	}

	toks, err := NewToks("", [][]string{strings.Fields(`a b c $.`)})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = mm.ReadStmtAux("$e", toks, "$.")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTinyMetamathDatabase(t *testing.T) {
	t.Parallel()
	must := func(e error) {
		if e == nil {
			return
		}
		panic(e.Error())
	}
	mm := NewMM(nil)
	must(mm.AddC("|-"))
	if _, _, k := mm.LookupSymbolByName("|-"); k == nil {
		t.Error("failed to add constant")
	}
	must(mm.AddC("wff"))
	if _, _, k := mm.LookupSymbolByName("wff"); k == nil {
		t.Error("failed to add constant")
	}
	must(mm.AddV("ph"))
	if _, v, _ := mm.LookupSymbolByName("ph"); v == nil {
		t.Error("failed to add variable")
	}
	must(mm.AddF("wff", "ph", "wph"))
	if label := mm.FS.LookupF("ph"); string(*label) != "wph" {
		t.Errorf("failed to add hypothesis")
	}
	mm.FS.AddE(Stmt{"|-", "ph"}, "idi.1")
	if v, ok := mm.FS.LastFrame().ELabels[ToSymbols(Stmt{"|-", "ph"})]; !ok || v != "idi.1" {
		t.Error("adding essential hypothesis failed")
	}
	label, err := mm.FS.LookupE(Stmt{"|-", "ph"})
	if err != nil {
		t.Error(err)
	}
	if *label != "idi.1" {
		t.Errorf("unexpected label %v", label)
	}
	if err := mm.CheckString("idi $p |- ph $= (  ) B $."); err != nil {
		t.Error(err)
	}
}

func TestCheckString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
		err     string
	}{
		{
			name:    "comment",
			content: "$( hi hi hi $)",
		},
		{
			name:    "define parenthesis",
			content: "$c ( $.",
		},
		{
			name: "tiny metamath database",
			content: `
$c |- $.
$c wff $.
$v ph $.
wph $f wff ph $.
idi.1 $e |- ph $.
idi $p |- ph $= (  ) B $.
`,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := CheckString(tt.content)

			if tt.err == "" {
				if e != nil {
					t.Errorf("unexpected error: %s", e)
				}
			} else {
				switch {
				case e == nil:
					t.Errorf("expected error containing %q but got nil", e)
				case strings.Contains(e.Error(), tt.err):
					t.Errorf("expected error containing %q but got inl", e)
				}
			}
		})
	}
}
