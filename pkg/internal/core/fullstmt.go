package core

type FullStmt struct {
	SType string
	// Only one of these can be non-nil
	MStmt      *Stmt
	MAssertion *Assertion
}

func (fullStmt *FullStmt) Check() *FullStmt {
	if fullStmt == nil {
		panic("Check invalid on nil pointer")
	}
	if fullStmt.MStmt == nil && fullStmt.MAssertion == nil {
		panic("MStmt and MAssertion cannot both be nil")
	}
	if fullStmt.MStmt != nil && fullStmt.MAssertion != nil {
		panic("MStmt and MAssertion cannot both be provided")
	}
	return fullStmt
}

func IsHypothesis(stmt FullStmt) bool {
	switch stmt.SType {
	case "$e", "$f":
		return true
	}
	return false
}

func IsAssertion(stmt FullStmt) bool {
	switch stmt.SType {
	case "$a", "$p":
		return true
	}
	return false
}
