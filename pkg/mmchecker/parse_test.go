package mmchecker

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  string
		output [][]string
		errPat string
	}{
		{
			name:   "empty",
			input:  "",
			output: [][]string{{}},
			errPat: "",
		},
		{
			name:  "singleton",
			input: "a",
			output: [][]string{
				strings.Fields("a"),
			},
			errPat: "",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parse(strings.NewReader(tt.input))

			if e := errContains(err, tt.errPat); e != nil {
				t.Error(e)
			}

			if e := makeDiff(got, tt.output); e != nil {
				t.Error(e)
			}
		})
	}
}

func TestParseString(t *testing.T) {
	t.Parallel()

	if e := makeDiff(parseString("a b c d e f g")[0], strings.Fields("a b c d e f g")); e != nil {
		t.Error(e)
	}
}

func TestParseString_MultiLineString(t *testing.T) {
	t.Parallel()

	got := parseString(`
		$(
			multi-
			line
			comment
		$)
	`)

	if e := makeDiff(got, [][]string{
		strings.Fields(""),
		strings.Fields("$("),
		strings.Fields("multi-"),
		strings.Fields("line"),
		strings.Fields("comment"),
		strings.Fields("$)"),
		strings.Fields(""),
	}); e != nil {
		t.Error(e)
	}
}
