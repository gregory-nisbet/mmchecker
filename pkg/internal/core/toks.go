package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
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

func (self *Toks) Read() (string, error) {
	// Fill the token buffer if it is not already full.
	for len(self.TokBuf) == 0 {
		lastFile := self.getLastFile()
		if lastFile == nil {
			return "", MMError{errors.New("Unclosed ${ ... $} block at end of file")}
		}

		line := lastFile.Text()
		if line.Just {
			self.TokBuf = line.Data
			reverse(self.TokBuf)
		} else {
			err := self.popFile()
			if err != nil {
				return "", fmt.Errorf("popping file: %w", err)
			}
			if len(self.FilesBuf) == 0 {
				return "", EOF
			}
		}
	}

	tok := self.TokBuf[-1+len(self.TokBuf)]
	self.TokBuf = self.TokBuf[:-1+len(self.TokBuf)]
	Vprint(90, "Token:", tok)
	return tok, nil
}

func (self *Toks) Readf() (string, error) {
	tok, err := self.Read()
	if err != nil {
		return "", fmt.Errorf("readf: %w", err)
	}
	for tok == "$[" {
		filename, err := self.Read()
		if err != nil {
			return "", fmt.Errorf("reading from file: %w", err)
		}

		endbracket, err := self.Read()
		if endbracket != "$]" {
			return "", fmt.Errorf("reading endbracket: %w", err)
		}

		filename, aErr := filepath.Abs(filename)
		if aErr != nil {
			return "", fmt.Errorf("resolving file: %w", err)
		}

		_, alreadySeen := self.ImportedFiles[filename]
		if alreadySeen {
			// do nothing
		} else {
			// Put the current line back on the stack of files
			// as a fake file.
			reversedTokBufs := self.TokBuf[:]
			reverse(reversedTokBufs)
			scanCloser, err := NewScanCloser("", [][]string{reversedTokBufs})
			if err != nil {
				return "", fmt.Errorf("making scancloser from tokbufs: %w", err)
			}
			self.FilesBuf = append(
				self.FilesBuf,
				scanCloser,
			)
			self.TokBuf = nil
			// Add the new file
			// TODO: I need a method for this.
			newFile, err := NewScanCloser(filename, nil)
			if err != nil {
				return "", fmt.Errorf("making scancloser from %q: %w", filename, err)
			}
			self.FilesBuf = append(self.FilesBuf, newFile)
			self.ImportedFiles[filename] = Unit
			// Change from original. Print the absolute path to the thing we imported.
			Vprint(5, "Importing file:", filename)
		}
		tok, err = self.Read()
		if err != nil {
			return "", fmt.Errorf("reading: %w", err)
		}
	}
	Vprint(80, "Token once included files expanded:", tok)
	return tok, nil
}

func (self *Toks) Readc() (string, error) {
	tok, err := self.Readf()
	if err != nil {
		return "", fmt.Errorf("reading: %w", err)
	}
	for tok == "$(" {
		tok, err = self.Read()
		if err != nil {
			return "", fmt.Errorf("reading token: %w", err)
		}
		for tok != "" && tok != "$)" {
			// This errors are worse than the original.
			if strings.Contains(tok, "$(") {
				return "", MMError{errors.New("token cannot contain $(")}
			}
			if strings.Contains(tok, "$)") {
				return "", MMError{errors.New("token cannot contain $)")}
			}
			tok, err = self.Read()
			if err != nil {
				return "", fmt.Errorf("reading token: %w", err)
			}
		}
		if tok != "$)" {
			panic("internal error: comment not closed")
		}
		tok, err = self.Readf()
		if err != nil {
			// Is this comment correct?
			return "", fmt.Errorf("reading token at end of skipping comment: %w", err)
		}
	}
	Vprint(70, "Token once comment skipped:", tok)
	return tok, nil
}
