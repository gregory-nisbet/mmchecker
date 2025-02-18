package mmchecker

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// parseString parses a string, it doesn't return an error to make it easier to use in tests.
func parseString(content string) [][]string {
	out, err := parse(strings.NewReader(content))
	if err != nil {
		panic(err)
	}
	return out
}

// parseFile reads a file into tokens.
func parseFile(path string) ([][]string, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("parseFile: %w", err)
	}
	out, err := parse(fh)
	if err != nil {
		return nil, fmt.Errorf("parseFile: %w", err)
	}
	return out, nil
}

// parse parses content into an array of lines of tokens.
func parse(content io.Reader) ([][]string, error) {
	bytes, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	lines := strings.Split(string(bytes), "\n")
	out := make([][]string, len(lines), len(lines))
	for i, line := range lines {
		parsedLine := strings.Fields(line)
		out[i] = parsedLine
	}
	return out, nil
}
