package core

type Toks struct {
	FilesBuf      []*ScanCloser
	TokBuf        []string
	ImportedFiles map[string]TUnit
}
