package mmchecker

import (
	"errors"
	"fmt"
)

// available isn't a real type, it marks a symbol as available for use.
const (
	available = "available"
	constant  = "constant"
	variable  = "variable"
	floating  = "floating"
	essential = "essential"
	proof     = "proof"
	axiom     = "axiom"
	disjoint  = "disjoint"
)

type kernel struct {
	linum int

	// Note: we have an extra level of indirection here (*symbol instead of symbol) so that
	// symbols can contain pointers to other symbols.
	//
	// In Go, map[x] is not addressable and might move around.
	//
	// We only keep track of line numbers, not offsets within the line or byte numbers (this decision
	// is also questionable). This means that once the kernel finishes processing a definition we
	// can't reconstruct which of its dependencies (def, theTerm, theType) point to which previously
	// existing symbols from the name of the symbol alone.
	//
	// This is also necessary to prevent dependencies from being garbage collected. Once your scope ends,
	// the things that depend on you can keep you alive.
	//
	// I don't know whether keeping old definitions alive is actually necessary to check things in metamath-land.
	// I suspect that the answer is "no" but I don't know yet.
	//
	// Also, note that symbols do NOT contain backreferences to the scope that contains them, so one thing
	// being alive in a scope does not keep the scope alive.
	stack []scope
}

func newKernel() *kernel {
	var out kernel

	normal(&out)

	return &out
}

type symbol struct {
	name    string
	linum   int
	typ     string
	def     []*symbol
	theTerm *symbol
	theType *symbol
}

type scope struct {
	symbols map[string]*symbol

	// The distinctness entry must be a bimap to avoid scanning it.
	distinct map[*symbol][]*symbol
}

func lookup(k *kernel, name string) *symbol {
	for i := -1 + len(k.stack); i >= 0; i-- {
		if item, ok := k.stack[i].symbols[name]; ok {
			return item
		}
	}

	return nil
}

func lookupType(k *kernel, name string) string {
	if sym := lookup(k, name); k != nil {
		return sym.typ
	}

	return available
}

func last(k *kernel) int {
	return -1 + len(k.stack)
}

// normal normalizes a kernel prior to use.
func normal(k *kernel) {
	if k.linum == 0 {
		k.linum = 1
	}

	if len(k.stack) == 0 {
		k.stack = []scope{
			{
				symbols:  map[string]*symbol{},
				distinct: map[*symbol][]*symbol{},
			},
		}
	}
}

func skipBlankLines(tokens [][]string) [][]string {
	index := 0

	for _, line := range tokens {
		if len(line) == 0 {
			index++
		} else {
			break
		}
	}

	return tokens[index:]
}

// readComment reads a comment and returns the updated tokens. Like append, it takes ownership of its [][]string argument.
//
// Kernel: modified on success to reflect the new linum
// Tokens: argument gets modified. When using this function, assign back to tokens.
func readComment(k *kernel, tokens [][]string) ([][]string, error) {
	if len(tokens) == 0 {
		return nil, errors.New("tokens cannot be empty")
	}

	if len(tokens[0]) == 0 {
		return nil, errors.New("head position of tokens cannot be empty")
	}

	if tokens[0][0] != "$(" {
		return nil, fmt.Errorf(`comment begins with %q not "$("`, tokens[0][0])
	}

	rowIndex, tokenIndex, pastEnd, err := findFirstInstanceAfter(tokens, "$)", 1, nil)
	if pastEnd {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("readComment: %w", err)
	}
	// Truncate the line indexed by the old index first, and then clip
	// the result. Less index manipulation than doing it in the other order.
	tokens[rowIndex] = tokens[rowIndex][tokenIndex:]
	tokens = tokens[rowIndex:]
	k.linum += rowIndex

	return tokens, nil
}

// processConstant registers a new constant symbol.
func processConstant(k *kernel, linum int, name string) error {
	if sym := lookup(k, name); sym != nil {
		return fmt.Errorf("process constant: symbol %q already exists", name)
	}

	k.stack[last(k)].symbols[name] = &symbol{
		name:  name,
		linum: linum,
		typ:   constant,
	}

	return nil
}

// processVariable registers a new variable symbol.
func processVariable(k *kernel, linum int, name string) error {
	if sym := lookup(k, name); sym != nil {
		return fmt.Errorf("process variable: symbol %q already exists", name)
	}

	k.stack[last(k)].symbols[name] = &symbol{
		name:  name,
		linum: linum,
		typ:   constant,
	}

	return nil
}

// Also, the leftmost thing must be a constant.
func processAxiom(k *kernel, linum int, name string, sentence []string) error {
	if sym := lookup(k, name); sym != nil {
		return fmt.Errorf("axiom: symbol %q already exists", name)
	}

	var symbols []*symbol

	for _, item := range sentence {
		sym := lookup(k, item)

		switch {
		case sym == nil:
			return fmt.Errorf("axiom: symbol %q in definition of %q does not exist", item, name)
		case sym.typ == constant || sym.typ == variable:
			symbols = append(symbols, sym)
		default:
			return fmt.Errorf("axiom: symbol %q in definition of %q has bad type %q", item, name, sym.typ)
		}
	}

	k.stack[last(k)].symbols[name] = &symbol{
		name:  name,
		linum: linum,
		typ:   axiom,
		def:   symbols,
	}

	return nil
}

// processFloatingHypothesis processes a floating hypothesis.
func processFloatingHypothesis(k *kernel, linum int, name string, baseConstant string, baseVariable string) error {
	if sym := lookup(k, name); sym != nil {
		return fmt.Errorf("floating: symbol %q already exists", name)
	}

	if typ := lookupType(k, baseConstant); baseConstant != typ {
		return fmt.Errorf("floating: symbol %q is %q not constant", name, typ)
	}

	if typ := lookupType(k, baseVariable); baseVariable != typ {
		return fmt.Errorf("floating: symbol %q is %q not variable", name, typ)
	}

	k.stack[last(k)].symbols[name] = &symbol{
		name:  name,
		linum: linum,
		typ:   floating,
	}

	return nil
}

// processEssentialHypothesis processes an essential hypothesis.
func processEssentialHypothesis(k *kernel, linum int, name string, baseConstant string, baseVariable string) error {
	if sym := lookup(k, name); sym != nil {
		return fmt.Errorf("essential: symbol %q already exists", name)
	}

	if typ := lookupType(k, baseConstant); baseConstant != typ {
		return fmt.Errorf("essential: symbol %q is %q not constant", name, typ)
	}

	if typ := lookupType(k, baseVariable); baseVariable != typ {
		return fmt.Errorf("essential: symbol %q is %q not variable", name, typ)
	}

	k.stack[last(k)].symbols[name] = &symbol{
		name:  name,
		linum: linum,
		typ:   essential,
	}

	return nil
}

// processDisjointnessHypothesis processes a disjointness hypothesis. Wowzers.
func processDisjointnessHypothesis(k *kernel, linum int, item1 string, item2 string) error {
	if typ := lookupType(k, item1); typ != variable {
		return fmt.Errorf("disjoint: symbol %q has type %q not variable", item1, variable)
	}

	if typ := lookupType(k, item2); typ != variable {
		return fmt.Errorf("disjoint: symbol %q has type %q not variable", item2, variable)
	}

	k.stack[last(k)].distinct[lookup(k, item1)] = append(k.stack[last(k)].distinct[lookup(k, item1)], lookup(k, item2))
	k.stack[last(k)].distinct[lookup(k, item2)] = append(k.stack[last(k)].distinct[lookup(k, item2)], lookup(k, item1))

	return nil
}
