package core

import (
	"fmt"
	"strings"
)

type Symbols string

func ToSymbols(symbols []string) Symbols {
	for _, msg := range symbols {
		for _, ch := range msg {
			switch ch {
			case ' ', '\n', '\t':
				panic(fmt.Sprintf("symbol %q contains forbidden char %v", msg, ch))
			}
		}
	}
	return Symbols(strings.Join(symbols, "\n"))
}

func FromSymbols(symbols Symbols) []string {
	return strings.Split(string(symbols), "\n")
}
