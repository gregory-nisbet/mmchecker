package core

import "strings"

func ToTokens(message string) [][]string {
	var out [][]string
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		out = append(out, strings.Fields(line))
	}
	return out
}
