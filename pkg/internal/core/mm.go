package core

import (
	"errors"
	"fmt"
)

type MM struct {
	BeginLabel   *Label
	EndLabel     *Label
	Constants    map[string]struct{}
	FS           *FrameStack
	Labels       map[Label]*FullStmt
	VerifyProofs bool
}

func NewMM(beginLabel *Label) *MM {
	return &MM{
		BeginLabel:   beginLabel,
		EndLabel:     nil,
		Constants:    map[string]TUnit{},
		Labels:       map[Label]*FullStmt{},
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

func (self *MM) AddV(tok string) error {
	if self.FS.LookupV(tok) {
		return MMError{fmt.Errorf("variable %q already declared and active", tok)}
	}
	frame := self.FS.LastFrame()
	if frame == nil {
		panic("impossible: frame stack is empty")
	}
	frame.V[tok] = struct{}{}
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

func (self *MM) ReadNonPStatement(stmttype string, toks *Toks) (Stmt, error) {
	return self.ReadStmtAux(stmttype, toks, "$.")
}

func (self *MM) ReadPStatement(toks *Toks) (Stmt, Stmt, error) {
	stmt, err := self.ReadStmtAux("$p", toks, "$=")
	if err != nil {
		return nil, nil, fmt.Errorf("read $= aux statment in p statement: %w", err)
	}
	proof, err := self.ReadStmtAux("$=", toks, "$.")
	if err != nil {
		return nil, nil, fmt.Errorf("read $. aux statement in p statement: %w", err)
	}
	return stmt, proof, nil
}

func (self *MM) Read(toks *Toks) error {
	self.FS.Push()
	var label *Label
	tok, err := toks.Readc()
	if err != nil {
		return fmt.Errorf("readc: %w", err)
	}
	for tok != "" && tok != "$}" {
		switch tok {
		case "$c":
			stmt, err := self.ReadNonPStatement(tok, toks)
			if err != nil {
				return fmt.Errorf("read non-p statement: %w", err)
			}
			for _, w := range stmt {
				if err := self.AddC(w); err != nil {
					return fmt.Errorf("addc: %w", err)
				}
			}
		case "$v":
			stmt, err := self.ReadNonPStatement(tok, toks)
			if err != nil {
				return fmt.Errorf("read non-p statement in $v: %w", err)
			}
			for _, w := range stmt {
				if err := self.AddV(w); err != nil {
					return fmt.Errorf("add variable $v: %w", err)
				}
			}
		case "$f":
			stmt, err := self.ReadNonPStatement(tok, toks)
			if err != nil {
				return MMError{fmt.Errorf("read statement in $f: %w", err)}
			}
			if label == nil {
				return MMError{fmt.Errorf("$f must have label (statement: %s)", stmt.String())}
			}
			if len(stmt) != 2 {
				return MMError{fmt.Errorf("$f must have length 2 but is %v", stmt.String())}
			}
			if err := self.AddF(stmt[0], stmt[1], *label); err != nil {
				return MMError{fmt.Errorf("$f: %w", err)}
			}
			self.Labels[*label] = (&FullStmt{
				SType: "$f",
				MStmt: &stmt,
			}).Check()
			label = nil
		case "$e":
			if label == nil {
				return MMError{errors.New("$e must have label")}
			}
			stmt, err := self.ReadNonPStatement(tok, toks)
			if err != nil {
				return MMError{fmt.Errorf("$e failed to read: %w", err)}
			}
			self.FS.AddE(stmt, *label)
			self.Labels[*label] = (&FullStmt{
				SType: "$e",
				MStmt: &stmt,
			}).Check()
			label = nil
		case "$a":
			if label == nil {
				return MMError{errors.New("$a must have label")}
			}
			stmt, err := self.ReadNonPStatement(tok, toks)
			if err != nil {
				return fmt.Errorf("reading statement in $a: %w", err)
			}
			assertion := self.FS.MakeAssertion(stmt)
			self.Labels[*label] = (&FullStmt{
				SType:      "$a",
				MAssertion: &assertion,
			})
			label = nil
		case "$p":
			if label == nil {
				return MMError{errors.New("label cannot be new in $p statement")}
			}
			_, _, err := self.ReadPStatement(toks)
			if err != nil {
				return fmt.Errorf("$p failed to read statement: %w", err)
			}
			if self.VerifyProofs {
				Vprint(2, "Verify:", string(*label))
				panic("not yet implemented")
				// self.Verify
			}
			panic("not yet implemented")
			label = nil
		}
	}
	panic("not yet implemented")
}
