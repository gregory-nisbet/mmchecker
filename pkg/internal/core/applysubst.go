package core

import "fmt"

// This is a non-capture-avoiding substitution because that's what
// metamath is based on.
func ApplySubst(stmt Stmt, subst map[string]Stmt) Stmt {
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
