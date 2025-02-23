package mmchecker

import (
	"strings"
	"testing"
)

func TestFindFirstInstanceOf_SimpleTest(t *testing.T) {
	t.Parallel()

	rowIndex, tokenIndex, isOnePastEnd, err := findFirstInstanceAfter([][]string{
		strings.Fields("$( a b c $) d e f"),
	}, "$)", 1, nil)
	if isOnePastEnd {
		t.Errorf("isOnePastEnd should not be true")
	}

	if e := makeDiff(rowIndex, 0); e != nil {
		t.Error(e)
	}

	if e := makeDiff(tokenIndex, 5); e != nil {
		t.Error(e)
	}

	if e := errContains(err, ""); e != nil {
		t.Error(e)
	}
}

func TestFindFirstInstanceOf(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		tokens     [][]string
		item       string
		walk       int
		rowIndex   int
		tokenIndex int
		onePastEnd bool
		errPat     string
	}{
		{
			name:   "empty",
			tokens: parseString(""),
			item:   "$)",
			errPat: "item not found",
		},
		{
			name:       "simple comment",
			tokens:     parseString("$( a b c $)"),
			item:       "$)",
			rowIndex:   0,
			tokenIndex: 4,
		},
		{
			name:       "walk after comment",
			tokens:     parseString("$( a b c $) d e f"),
			item:       "$)",
			walk:       2,
			rowIndex:   0,
			tokenIndex: 6,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rowIndex, tokenIndex, onePastEnd, err := findFirstInstanceAfter(tt.tokens, tt.item, tt.walk, nil)
			if e := makeDiff(rowIndex, tt.rowIndex); e != nil {
				t.Error(e)
			}

			if e := makeDiff(tokenIndex, tt.tokenIndex); e != nil {
				t.Error(e)
			}

			if e := makeDiff(onePastEnd, tt.onePastEnd); e != nil {
				t.Error(e)
			}

			if e := errContains(err, tt.errPat); e != nil {
				t.Error(e)
			}
		})
	}
}
