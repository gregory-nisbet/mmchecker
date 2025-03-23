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

func reverse(slice []string) {
	last := -1 + len(slice)
	halfLen := len(slice) / 2
	for i := 0; i < halfLen; i++ {
		slice[i], slice[last-i] = slice[last-i], slice[i]
	}
}
