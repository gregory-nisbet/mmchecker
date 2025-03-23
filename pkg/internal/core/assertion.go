package core

type Assertion struct {
	Dvs map[Dv]struct{}
	F   []Fhyp
	E   []Ehyp
	S   Stmt
}
