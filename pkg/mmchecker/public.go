package mmchecker

import (
	"context"
	"errors"
)

func Validate(ctx context.Context, path string, content string) error {
	params := 0
	if path != "" {
		params++
	}
	if content != "" {
		params++
	}
	switch params {
	case 0:
		return errors.New("no parameters given")
	case 1:
		// continue
	default:
		return errors.New("too many parameters given")
	}
	var parsed [][]string
	var err error
	switch {
	case path != "":
		parsed, err = parseFile(path)
		if err != nil {
			return err
		}
	case content != "":
		parsed = parseString(content)
	}

	k := newKernel()
}
