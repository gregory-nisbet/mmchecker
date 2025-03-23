package core

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

type ScanCloser struct {
	isMemoryCloser bool
	tokens         [][]string
	path           string
	fh             *os.File
	scanner        *bufio.Scanner
}

func NewScanCloser(path string, tokens [][]string) (*ScanCloser, error) {
	if path != "" && len(tokens) != 0 {
		return nil, &MMError{errors.New("ScanCloser can either be in memory or from a path on disk, not both")}
	}
	if len(tokens) != 0 {
		return &ScanCloser{
			isMemoryCloser: true,
			tokens:         tokens,
		}, nil
	}
	fh, err := os.Open(path)
	if err != nil {
		return nil, &IOError{err}
	}
	return &ScanCloser{
		path:    path,
		fh:      fh,
		scanner: bufio.NewScanner(fh),
	}, nil
}

func (scanCloser *ScanCloser) Text() StringListOption {
	// Control does not leave this block if we enter it.
	if scanCloser.isMemoryCloser {
		if len(scanCloser.tokens) == 0 {
			return StringListOption{}
		}
		out := scanCloser.tokens[-1+len(scanCloser.tokens)]
		scanCloser.tokens = scanCloser.tokens[:-1+len(scanCloser.tokens)]
		return StringListOption{Just: true, Data: out}
	}

	ok := scanCloser.scanner.Scan()
	if !ok {
		return StringListOption{}
	}
	return StringListOption{
		Just: true,
		Data: strings.Fields(scanCloser.scanner.Text()),
	}
}

func (scanCloser *ScanCloser) MustClose() {
	if scanCloser.isMemoryCloser {
		return
	}
	if err := scanCloser.fh.Close(); err != nil {
		panic(err)
	}
	scanCloser.fh = nil
	scanCloser.scanner = nil
}
