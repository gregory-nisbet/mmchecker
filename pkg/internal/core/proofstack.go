package core

import "fmt"

type ProofStack struct {
	data []Stmt
}

func NewProofStack() *ProofStack {
	return &ProofStack{}
}

func (stack *ProofStack) TreatStep(mm *MM, step *FullStmt) error {
	Vprint(10, "Proof step:", fmt.Sprintf("%v", step))
	if IsHypothesis(*step) {
		stmt := *step.MStmt
		stack.data = append(stack.data, stmt)
		return nil
	}
	if !IsAssertion(*step) {
		panic("TreatStep given argument that is neither hypothesis nor assertion")
	}
	assertion := *step.MAssertion
	dvs0 := assertion.Dvs
	fhyps0 := assertion.F
	ehyps0 := assertion.E
	conclusion0 := assertion.S
	npop := len(fhyps0) + len(ehyps0)
	sp := len(stack.data) - npop
	if sp < 0 {
		return MMError{fmt.Errorf("Stack underflow: proof step %v requires too many hypotehses %v", step, npop)}
	}
	subst := map[string]Stmt{}
	for _, p := range fhyps0 {
		typecode := p.Typecode
		va := p.V
		entry := stack.data[sp]
		if entry[0] != typecode {
			return MMError{fmt.Errorf("Proof stack entry %v does not match floating hypothesis %v %v", entry, typecode, va)}
		}
		subst[va] = entry[1:]
		sp += 1
	}
	Vprint(15, "Substitution to apply", fmt.Sprintf("%v", subst))
	for _, h := range ehyps0 {
		entry := stack.data[sp]
		substH := ApplySubst(Stmt(h), subst)
		if Stmt(entry).Equals(substH) {
			return MMError{fmt.Errorf("Proof stack entry %v does not match essential hypothesis %v", entry, substH)}
		}
		sp += 1
	}
	for p, _ := range dvs0 {
		x := p.First
		y := p.Second
		Vprint(16, "dist", x, y, subst[x].String(), subst[y].String())
		xVars := mm.FS.FindVars(subst[x])
		yVars := mm.FS.FindVars(subst[y])
		for x0, _ := range xVars {
			for y0, _ := range yVars {
				if x0 == y0 {
					return MMError{fmt.Errorf("new disjoint violation: %q", x0)}
				}
				if !mm.FS.LookupD(x0, y0) {
					return MMError{fmt.Errorf("variables %q and %q are not known to be disjoint", x0, y0)}
				}
			}
		}
	}
	stack.data = stack.data[:len(stack.data)-npop]
	newStmt := ApplySubst(conclusion0, subst)
	stack.data = append(stack.data, newStmt)
	return nil
}
