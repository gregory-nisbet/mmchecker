package core

func Assert(cond bool, message string) {
	if cond {
		return
	}
	panic(message)
}
