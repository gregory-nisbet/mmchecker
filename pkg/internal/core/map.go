package core

// We could just return []Symbol{tok} on failure, but nah let's not do that.
func mapHasVar(subst map[string]Stmt, tok string) ([]string, bool) {
	newThing, ok := subst[tok]
	if ok {
		return newThing, true
	}
	return nil, false
}
