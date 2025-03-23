package core

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

// TestScanCloser tests opening a file and scanning it.
func TestScanCloser(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()
	tempfile := filepath.Join(tempdir, "a.txt")
	if err := ioutil.WriteFile(tempfile, []byte("eee"), 0x700); err != nil {
		t.Error(err)
	}

	scanCloser, err := NewScanCloser(tempfile, nil)
	if err != nil {
		t.Error(err)
	}

	readText := scanCloser.Text()
	if strings.Join(readText.Data, ",") != "eee" {
		t.Errorf("unexpected result of Text(): %v", readText)
	}

	readText = scanCloser.Text()
	if strings.Join(readText.Data, ",") != "" {
		t.Errorf("unexpected result of Text(): %v", readText)
	}

	scanCloser.MustClose()
}
