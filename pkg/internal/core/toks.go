package core

import (
	"errors"
	"fmt"
)

type Toks struct {
	FilesBuf      []*ScanCloser
	TokBuf        []string
	ImportedFiles map[string]TUnit
}

func NewToks(path string, tokens [][]string) (*Toks, error) {
	scanCloser, err := NewScanCloser(path, tokens)
	if err != nil {
		return nil, fmt.Errorf("NewToks: %w", err)
	}
	return &Toks{
		FilesBuf: []*ScanCloser{scanCloser},
		TokBuf:   nil,
		ImportedFiles: map[string]TUnit{
			path: Unit,
		},
	}, nil
}

func (self *Toks) getLastFile() *ScanCloser {
	if len(self.FilesBuf) == 0 {
		return nil
	}
	return self.FilesBuf[-1+len(self.FilesBuf)]
}

func (self *Toks) popFile() error {
	if len(self.FilesBuf) == 0 {
		return MMError{errors.New("out of files")}
	}
	self.FilesBuf[-1+len(self.FilesBuf)].MustClose()
	self.FilesBuf = self.FilesBuf[:-1+len(self.FilesBuf)]
	return nil
}
