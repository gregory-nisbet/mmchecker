package core

import "strings"

type Stmt []string

func (self *Stmt) String() string {
	if self == nil {
		return "<nil>"
	}
	return strings.Join([]string(*self), " ")
}
