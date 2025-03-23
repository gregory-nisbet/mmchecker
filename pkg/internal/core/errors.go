package core

import (
	"errors"
)

type IOError struct {
	err error
}

func (i IOError) Error() string {
	msg := i.Error()
	if msg == "" {
		panic(`IOError wrapped error is stringifies to ""`)
	}
	return msg
}

func AsIOError(e error) *IOError {
	var i IOError
	if errors.As(e, &i) {
		return &i
	}
	return nil
}

type MMError struct {
	err error
}

func (i MMError) Error() string {
	msg := i.Error()
	if msg == "" {
		panic(`MMError stringifies to ""`)
	}
	return msg
}

func AsMMError(e error) *MMError {
	var m MMError
	if errors.As(e, &m) {
		return &m
	}
	return nil
}
