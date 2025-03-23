package core

import "fmt"

type Assertion struct {
	Dvs map[Dv]struct{}
	F   []Fhyp
	E   []Ehyp
	S   Stmt
}

func (assertion *Assertion) String() string {
	if assertion == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Assertion %v %v %v %v", assertion.Dvs, assertion.F, assertion.E, assertion.S)
}
