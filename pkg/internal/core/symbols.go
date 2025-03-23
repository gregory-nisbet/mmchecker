package core

import (
	"fmt"
	"strings"
)

type Symbols string

func ToSymbols(symbols []string) string {
	for _, msg := range symbols {
		for _, ch := range msg {
			switch ch {
			case ' ', '\n', '\t':
				panic(fmt.Sprintf("symbol %q contains forbidden char %v", msg, ch))
			}
		}
	}
	return strings.Join(symbols, "\n")
}

func FromSymbols(symbols string) []string {
	return strings.Split(symbols, "\n")
}
