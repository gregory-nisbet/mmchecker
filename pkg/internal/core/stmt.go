package core

import "strings"

type Stmt []string

func (self Stmt) String() string {
	return strings.Join([]string(self), " ")
}

func (self Stmt) Equals(other Stmt) bool {
	if len(self) != len(other) {
		return false
	}
	for i, _ := range []string(self) {
		if self[i] != other[i] {
			return false
		}
	}
	return true
}
