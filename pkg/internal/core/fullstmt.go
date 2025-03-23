package core

type FullStmt struct {
	SType string
	// Only one of these can be non-nil
	MStmt      *Stmt
	MAssertion *Assertion
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
