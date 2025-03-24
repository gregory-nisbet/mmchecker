package core

import "fmt"

type FrameStack struct {
	Frames []*Frame
}

func NewFrameStack() *FrameStack {
	return &FrameStack{
		Frames: nil,
	}
}

func (self *FrameStack) Push() {
	self.Frames = append(self.Frames, NewFrame())
}

func (self *FrameStack) Pop() {
	self.Frames = self.Frames[:-1+len(self.Frames)]
}

const GO = 1
const STOP = 2

func (self *FrameStack) Foreach(callback func(f *Frame) int8) {
	if self == nil {
		panic("framestack cannot be nil")
	}
	for i := -1 + len(self.Frames); i >= 0; i-- {
		frame := self.Frames[i]
		status := callback(frame)
		switch status {
		case GO:
			// do nothing
		case STOP:
			return
		default:
			panic(fmt.Sprintf("bad status %v", status))
		}
	}
}

func (self *FrameStack) LastFrame() *Frame {
	if len(self.Frames) == 0 {
		self.Push()
	}
	out := self.Frames[-1+len(self.Frames)]
	return out
}

// Can this fail?
func (self *FrameStack) AddE(stmt Stmt, label Label) {
	frame := self.LastFrame()
	frame.E = append(frame.E, Ehyp(stmt))
	// Go doesn't have tuples (or another hashable connection)
	// So we convert to a string, painfully.
	// I wonder how slow this will be in benchmarks.
	frame.ELabels[ToSymbols(stmt)] = label
}

// Add all distinct pairs
//
// This is quadratic time for, like, no reason.
//
// TODO: represent a collection of distinct variables as a map
//
//	going to the LEAST variable lexicographically.
func (self *FrameStack) AddD(varlist []string) {
	frame := self.LastFrame()
	for _, x := range varlist {
		for _, y := range varlist {
			min := x
			max := y
			if string(y) < string(x) {
				min = y
				max = x
			}
			newRecord := Dv{
				First:  min,
				Second: max,
			}
			frame.D[newRecord] = struct{}{}
		}
	}
}

func (self *FrameStack) LookupV(tok string) bool {
	out := false
	self.Foreach(func(frame *Frame) int8 {
		_, ok := frame.V[tok]
		if ok {
			out = true
			return STOP
		}
		return GO
	})
	return out
}

func (self *FrameStack) LookupD(x string, y string) bool {
	min := x
	max := y
	if string(y) < string(x) {
		min = y
		max = x
	}
	newRecord := Dv{
		First:  min,
		Second: max,
	}
	out := false
	self.Foreach(func(frame *Frame) int8 {
		_, ok := frame.D[newRecord]
		if ok {
			out = true
			return STOP
		}
		return GO
	})
	return out
}

// return pointer to label of active floating hypothesis or nil
// if none exists.
//
// TODO: OptionalLabel?
func (self *FrameStack) LookupF(va string) *Label {
	var out *Label
	self.Foreach(func(frame *Frame) int8 {
		label, ok := frame.FLabels[va]
		if ok {
			out = &label
			return STOP
		}
		return GO
	})
	return out
}

func (self *FrameStack) LookupE(stmt Stmt) (*Label, error) {
	var out *Label
	var err error
	encoded := ToSymbols(stmt)
	self.Foreach(func(frame *Frame) int8 {
		label, ok := frame.ELabels[encoded]
		if ok {
			out = &label
			return STOP
		}
		return GO
	})
	if out != nil {
		err = fmt.Errorf("lookup e failed: %v", stmt)
	}
	return out, err
}

func (self *FrameStack) FindVars(stmt Stmt) map[string]TUnit {
	out := map[string]TUnit{}
	for _, x := range stmt {
		if self.LookupV(x) {
			out[x] = Unit
		}
	}
	return out
}

func (self *FrameStack) MakeAssertion(stmt Stmt) Assertion {
	var eHyps []Ehyp
	mandVars := map[string]TUnit{}
	dvs := map[Dv]TUnit{}
	var fHyps []Fhyp

	self.Foreach(func(frame *Frame) int8 {
		for _, eh := range frame.E {
			eHyps = append(eHyps, eh)
		}
		return GO
	})

	// Do the weird thing for "efficiency".
	// Add our statement to eHyps and then remove it.
	// replaces itertools.chain(e_hyps, [stmt])
	eHyps = append(eHyps, Ehyp(stmt))
	for _, hyp := range eHyps {
		for _, tok := range hyp {
			if self.LookupV(tok) {
				mandVars[tok] = Unit
			}
		}
	}
	eHyps = eHyps[:-1+len(eHyps)]

	self.Foreach(func(frame *Frame) int8 {
		for _, p := range frame.F {
			typecode := p.Typecode
			va := p.V
			if _, ok := mandVars[va]; ok {
				fHyps = append(fHyps, Fhyp{
					Typecode: typecode,
					V:        va,
				})
				delete(mandVars, va)
			}
		}
		return GO
	})

	out := Assertion{
		Dvs: dvs,
		F:   fHyps,
		E:   eHyps,
		S:   stmt,
	}
	Vprint(18, "Make assertion:", out.String())
	return out
}
