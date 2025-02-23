package mmchecker

import (
	"strings"
	"testing"
)

// TestReadComment tests reading a comment and skipping past it.
func TestReadComment(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		k      *kernel
		tokens [][]string
		output [][]string
		errPat string
	}{
		{
			name: "multiline file",
			tokens: parseString(`
				$(
					multi-
					line
					comment
				$)
			`),
			output: nil,
			errPat: "",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := readComment(tt.k, skipBlankLines(tt.tokens))

			if e := makeDiff(got, tt.output); e != nil {
				t.Error(e)
			}

			if e := errContains(err, tt.errPat); e != nil {
				t.Error(e)
			}
		})
	}
}

// TestSkipBlankLines tests skipping blank lines.
func TestSkipBlankLines(t *testing.T) {
	t.Parallel()

	got := skipBlankLines([][]string{nil, strings.Fields("a b c")})

	if e := makeDiff(got, [][]string{strings.Fields("a b c")}); e != nil {
		t.Error(e)
	}
}

// TestSkipBlankLines tests skipping blank lines when nothing should be done.
func TestSkipBlankLines_Noop(t *testing.T) {
	t.Parallel()

	got := skipBlankLines([][]string{strings.Fields("a b c")})

	if e := makeDiff(got, [][]string{strings.Fields("a b c")}); e != nil {
		t.Error(e)
	}
}

// TestReadComment_SingleTest tests the pipeline for reading a comment and stripping stuff out.
func TestReadComment_SingleTest(t *testing.T) {
	t.Parallel()

	input := `
	$(
		multi-
		line
		comment
	$)`

	parsed := parseString(input)

	if e := makeDiff(parsed, [][]string{
		{},
		{"$("},
		{"multi-"},
		{"line"},
		{"comment"},
		{"$)"},
	}); e != nil {
		t.Error(e)
	}

	trimmed := skipBlankLines(parsed)

	if e := makeDiff(trimmed, [][]string{
		{"$("},
		{"multi-"},
		{"line"},
		{"comment"},
		{"$)"},
	}); e != nil {
		t.Error(e)
	}

	advanced, err := readComment(nil, trimmed)
	if e := errContains(err, ""); e != nil {
		t.Error(e)
	}

	if e := makeDiff(combinedLength(advanced), 0); e != nil {
		t.Error(e)
	}
}

// TestProcessConstant_HappyPath tests reading a constant statement and updating the kernel.
func TestProcessConstant_HappyPath(t *testing.T) {
	t.Parallel()

	k := newKernel()

	err := processConstant(k, 1, "foo")
	if err != nil {
		t.Error(err)
	}
}

// TestProcessVariable_HappyPath tests reading a variable statement and updating the kernel.
func TestProcessVariable_HappyPath(t *testing.T) {
	t.Parallel()

	k := newKernel()

	err := processVariable(k, 1, "foo")
	if err != nil {
		t.Error(err)
	}
}

// TestProcessAxiom_HappyPath tests adding an axiom to the kernel state.
func TestProcessAxiom_HappyPath(t *testing.T) {
	t.Parallel()

	k := newKernel()

	must(processConstant(k, 1, "a"))
	must(processVariable(k, 1, "b"))
	must(processConstant(k, 1, "c"))

	err := processAxiom(k, 1, "e", strings.Fields("a b c"))
	if err != nil {
		t.Error(err)
	}
}
