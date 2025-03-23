package core

type IOError struct {
	err error
}

func (i *IOError) Error() string {
	if i == nil {
		return ""
	}
	msg := i.Error()
	if msg == "" {
		panic(`IOError wrapped error is non-nil but stringifies to ""`)
	}
	return msg
}

type MMError struct {
	err error
}

func (i *MMError) Error() string {
	if i == nil {
		return ""
	}
	msg := i.Error()
	if msg == "" {
		panic(`MMError wrapped error is non-nil but stringifies to ""`)
	}
	return msg
}
