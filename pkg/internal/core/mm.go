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
