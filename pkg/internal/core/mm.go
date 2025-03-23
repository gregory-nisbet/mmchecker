package core

import "fmt"

type MM struct {
	BeginLabel   *Label
	EndLabel     *Label
	Constants    map[string]struct{}
	FS           *FrameStack
	Labels       map[Label]FullStmt
	VerifyProofs bool
}

func NewMM(beginLabel *Label) *MM {
	return &MM{
		BeginLabel:   beginLabel,
		EndLabel:     nil,
		Constants:    map[string]TUnit{},
		Labels:       map[Label]FullStmt{},
		VerifyProofs: beginLabel == nil,
		FS:           NewFrameStack(),
	}
}

func (self *MM) AddC(tok string) error {
	_, ok := self.Constants[tok]
	if ok {
		return MMError{fmt.Errorf("constant %q already declared", tok)}
	}
	self.Constants[tok] = struct{}{}
	return nil
}

func (self *MM) AddF(typecode string, va string, label Label) error {
	if self.FS.LookupV(va) {
		// Good. We need the variable to already exist.
	} else {
		return MMError{fmt.Errorf("var in $f not declared: %q", va)}
	}
	if _, ok := self.Constants[typecode]; ok {
		// Good. The constant must exist already.
	} else {
		return MMError{fmt.Errorf("typecode in $f not declared: %q", typecode)}
	}

	alreadyTyped := false
	self.FS.Foreach(func(frame *Frame) int8 {
		_, ok := frame.FLabels[va]
		if ok {
			alreadyTyped = true
			return STOP
		}
		return GO
	})
	if alreadyTyped {
		return MMError{fmt.Errorf("var in $f already typed by an active $f-statement: %q", va)}
	}
	frame := self.FS.LastFrame()
	if frame == nil {
		panic("impossible")
	}
	frame.F = append(frame.F, Fhyp{
		Typecode: typecode,
		V:        va,
	})
	frame.FLabels[va] = label
	return nil
}

// *Symbol, *Var, *Const
func (self *MM) LookupSymbolByName(tok string) (*string, *string, *string) {
	isActiveVar := self.FS.LookupV(tok)
	_, isConstant := self.Constants[tok]
	switch {
	case isActiveVar && isConstant:
		panic(fmt.Sprintf("string %q is both var and const", tok))
	case isActiveVar:
		return &tok, &tok, nil
	case isConstant:
		return &tok, nil, &tok
	default:
		return nil, nil, nil
	}
}

// endToken is "$=" or "$.".
// endToken shouldn't be a string this function is too general.
func (self *MM) ReadStmtAux(stmttype string, toks *Toks, endToken string) (Stmt, error) {
	Assert(endToken == "$=" || endToken == "$.", `endToken is $. or $=`)
	var stmt Stmt
	tok, err := toks.Readc()
	if err != nil {
		return nil, fmt.Errorf("failed to readc: %w", err)
	}
	for tok != "" && tok != endToken {
		// What do we do if the symbol doesn't exist?
		sym, va, constant := self.LookupSymbolByName(tok)
		// Validate active symbol.
		switch stmttype {
		case "$d", "$e", "$a", "$p":
			if va == nil && constant == nil {
				return nil, MMError{fmt.Errorf("Token %q is not an active symbol", tok)}
			}
		}
		// Validate symbol typed by hypothesis.
		switch stmttype {
		case "$e", "$a", "$p":
			if va != nil && self.FS.LookupF(*va) == nil {
				return nil, MMError{fmt.Errorf("Variable %q in %s-statement is not typed by an active $f-statement", tok, stmttype)}
			}
		}
		stmt = append(stmt, *sym)
		tok, err = toks.Readc()
		if err != nil {
			return nil, fmt.Errorf("failed to readc in processing loop: %w", err)
		}
	}
	if tok == "" {
		return nil, MMError{fmt.Errorf("Unclosed %q-statement at the end of file", stmttype)}
	}
	if tok != endToken {
		panic("tok must equal endToken")
	}
	Vprint(20, "Statement:", stmt.String())
	return stmt, nil
}
