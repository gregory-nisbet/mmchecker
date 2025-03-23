package core

import "fmt"

type Frame struct {
	V       map[string]TUnit
	D       map[Dv]TUnit
	F       []Fhyp
	FLabels map[string]Label
	E       []Ehyp
	ELabels map[Symbols]Label
	// Only for testing
	Name string
}

// NewFrame initializes the maps.
func NewFrame() *Frame {
	return &Frame{
		V:       map[string]TUnit{},
		D:       map[Dv]TUnit{},
		F:       nil,
		FLabels: map[string]Label{},
		E:       nil,
		ELabels: map[Symbols]Label{},
	}
}

// This is a non-capture-avoiding substitution because that's what
// metamath is based on.
func (self *Frame) ApplySubst(stmt Stmt, subst map[string]Stmt) Stmt {
	var result Stmt
	for _, tok := range stmt {
		newThing, ok := mapHasVar(subst, tok)
		if ok {
			result = append(result, newThing...)
		} else {
			result = append(result, tok)
		}
	}
	Vprint(20, "Applying subst", fmt.Sprintf("%v", subst), "to stmt", fmt.Sprintf("%v", stmt), ":", fmt.Sprintf("%v", result))
	return result
}
