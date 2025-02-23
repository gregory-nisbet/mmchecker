package mmchecker

import "errors"

var notFoundError = errors.New("item not found")

var pastEndError = errors.New("past end error")

func assert(condition bool, message string) {
	if condition {
		// do nothing
	} else {
		panic(message)
	}
}

// rowIndex, tokenIndex, isOnePastEnd, err.
func findFirstInstanceAfter(tokens [][]string, item string, walk int, before *[]string) (int, int, bool, error) {
	found := false
	count := 0

	for rowIndex, row := range tokens {
		for tokenIndex, token := range row {
			if before != nil {
				*before = append(*before, token)
			}

			if !found && token == item {
				found = true
				count = walk
			}

			if found {
				switch {
				case count == 0:
					return rowIndex, tokenIndex, false, nil
				default:
					count--
				}
			}
		}
	}

	switch {
	case !found:
		return 0, 0, false, notFoundError
	case count == 0:
		return 0, 0, true, nil
	default:
		return 0, 0, false, pastEndError
	}
}

func combinedLength(array [][]string) int {
	count := 0
	for _, line := range array {
		count += len(line)
	}

	return count
}
