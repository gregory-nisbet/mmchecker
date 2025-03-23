// mmverify.py -- Proof verifier for the Metamath language
// Copyright (C) 2002 Raph Levien raph (at) acm (dot) org
// Copyright (C) David A. Wheeler and mmverify.py contributors
//
// This program is free software distributed under the MIT license;
// see the file LICENSE for full license information.
// SPDX-License-Identifier: MIT
//
// To run the program, type
//   $ python3 mmverify.py set.mm --logfile set.log
// and set.log will have the verification results.  One can also use bash
// redirections and type '$ python3 mmverify.py < set.mm 2> set.log' but this
// would fail in case 'set.mm' contains (directly or not) a recursive inclusion
// statement $[ set.mm $] .
//
// To get help on the program usage, type
//   $ python3 mmverify.py -h
//
// (nm 27-Jun-2005) mmverify.py requires that a $f hypothesis must not occur
// after a $e hypothesis in the same scope, even though this is allowed by
// the Metamath spec.  This is not a serious limitation since it can be
// met by rearranging the hypothesis order.
// (rl 2-Oct-2006) removed extraneous line found by Jason Orendorff
// (sf 27-Jan-2013) ported to Python 3, added support for compressed proofs
// and file inclusion
// (bj 3-Apr-2022) streamlined code; obtained significant speedup (4x on set.mm)
// by verifying compressed proofs without converting them to normal proof format;
// added type hints
// (am 29-May-2023) added typeguards

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// import sys
// import itertools
// import pathlib
// import argparse
// import typing
// import io

type ScanCloser struct {
	isMemoryCloser bool
	tokens         [][]string
	path           string
	fh             *os.File
	scanner        *bufio.Scanner
}

func NewScanCloser(path string, tokens [][]string) (*ScanCloser, *Error) {
	if path != "" && len(tokens) != 0 {
		return nil, MMError("ScanCloser can either be in memory or from a path on disk, not both")
	}
	if len(tokens) != 0 {
		return &ScanCloser{
			isMemoryCloser: true,
			tokens:         tokens,
		}, nil
	}
	fh, err := os.Open(path)
	if err != nil {
		return nil, IOError(err).AddTag(1000)
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

// Take an idea from https://gitlab.com/esr/reposurgeon/blob/master/GoNotes.adoc
type Error struct {
	Class   string
	Message string
	Tally   int
	Tags    []int
}

func (e *Error) GetClass() string {
	if e == nil {
		return ""
	}
	return e.Class
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s(%v, %v): %s", e.Class, e.Tally, e.Tags, e.Message)
}

func (e *Error) AddTag(tag int) *Error {
	if e == nil {
		return nil
	}
	tags := e.Tags[:]
	tags = append(tags, tag)
	return &Error{
		Class:   e.Class,
		Message: e.Message,
		Tally:   e.Tally + 1,
		Tags:    tags,
	}
}

var _ error = &Error{}

type Label string
type Var string
type Const string
type StmtType string

// I use StringOption more than the original.
type StringOption struct {
	Just bool
	Data string
}

type StringListOption struct {
	Just bool
	Data []string
}

// Symbol is a hand-rolled option.
type Symbol struct {
	Inl *Var
	Inr *Const
}

// Warning, not a pointer receiver method
func (this Symbol) Equals(that Symbol) bool {
	sameNils := true
	sameNils = sameNils && (this.Inl == nil) == (that.Inl == nil)
	sameNils = sameNils && (this.Inr == nil) == (that.Inr == nil)

	if !sameNils {
		return false
	}
	if this.Inl != nil && *this.Inl != *that.Inl {
		return false
	}
	if this.Inr != nil && *this.Inr != *that.Inr {
		return false
	}
	return true
}

func MakeVar(name string) Symbol {
	x := Var(name)
	return Symbol{
		Inl: &x,
	}
}

func MakeConst(name string) Symbol {
	x := Const(name)
	return Symbol{
		Inr: &x,
	}
}

// Symbols is an encoding of a []Symbol that is known to Go
// to be immutable and thus usable as a key.
type Symbols string

// should we check for bytes instead? Whatever.
func verifyNoForbiddenCharacters(content string) {
	// This is probably wrong and we probably want bytes instead.
	for _, ch := range content {
		switch ch {
		case ' ', '\t', '\n':
			panic(fmt.Sprintf("identifier %q contains forbidden character", content))
		}
	}
}

func ToSymbols(symbols []Symbol) Symbols {
	var out []string
	for _, symbol := range symbols {
		switch {
		case symbol.Inl != nil:
			verifyNoForbiddenCharacters(string(*symbol.Inl))
			out = append(out, "v "+string(*symbol.Inl))
		case symbol.Inr != nil:
			verifyNoForbiddenCharacters(string(*symbol.Inr))
			out = append(out, "c "+string(*symbol.Inr))
		default:
			panic("issue in ToSymbols")
		}
	}
	return Symbols(strings.Join(out, "\n"))
}

func FromSymbols(symbols Symbols) []Symbol {
	var out []Symbol

	lines := strings.Split(string(symbols), "\n")

	for _, line := range lines {
		switch line[0] {
		case 'v':
			name := Var(line[2:])
			out = append(out, Symbol{Inl: &name})
		case 'c':
			name := Const(line[2:])
			out = append(out, Symbol{Inr: &name})
		}
	}

	return out
}

type Stmt []Symbol
type Ehyp []Symbol

type Fhyp struct {
	Typecode Const
	V        Var
}

// Dv is a pair
type Dv struct {
	First  Var
	Second Var
}

type Assertion struct {
	Dvs map[Dv]struct{}
	F   []Fhyp
	E   []Ehyp
	S   Stmt
}

type FullStmt struct {
	SType StmtType
	// Only one of these can be non-nil
	MStmt      *Stmt
	MAssertion *Assertion
}

// Label = str
// Var = str
// Const = str
// Stmttype = typing.Literal["$c", "$v", "$f", "$e", "$a", "$p", "$d", "$="]
// StringOption = typing.Optional[str]
// Symbol = typing.Union[Var, Const]
// Stmt = list[Symbol]
// Ehyp = Stmt
// Fhyp = tuple[Const, Var]
// Dv = tuple[Var, Var]
// Assertion = tuple[set[Dv], list[Fhyp], list[Ehyp], Stmt]
// FullStmt = tuple[Stmttype, typing.Union[Stmt, Assertion]]

// def is_hypothesis(stmt: FullStmt) -> typing.TypeGuard[tuple[Stmttype, Stmt]]:
//     """The second component of a FullStmt is a Stmt when its first
//     component is '$e' or '$f'."""
//     return stmt[0] in ('$e', '$f')

func IsHypothesis(stmt FullStmt) bool {
	return stmt.SType == "$e" || stmt.SType == "$f"
}

// def is_assertion(stmt: FullStmt) -> typing.TypeGuard[tuple[Stmttype, Assertion]]:
//     """The second component of a FullStmt is an Assertion if its first
//     component is '$a' or '$p'."""
//     return stmt[0] in ('$a', '$p')

func IsAssertion(stmt FullStmt) bool {
	return stmt.SType == "$a" || stmt.SType == "$p"
}

func MMError(message string) *Error {
	return &Error{
		Class:   "MMError",
		Message: message,
	}
}

func MMKeyError(message string) *Error {
	return &Error{
		Class:   "MMKeyError",
		Message: message,
	}
}

func IOError(err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Class:   "IOError",
		Message: err.Error(),
	}
}

func EOF() *Error {
	return &Error{
		Class:   "EOF",
		Message: "end of file",
	}
}

var verbosity int

// Vprint formats stuff wrong.
func Vprint(vlevel int, args ...any) {
	if verbosity >= vlevel {
		fmt.Fprintf(os.Stderr, "%v\n", args)
	}
}

type Toks struct {
	FilesBuf      []*ScanCloser
	TokBuf        []string
	ImportedFiles map[string]struct{}
}

func NewToks(path string, tokens [][]string) (*Toks, *Error) {
	scanCloser, err := NewScanCloser(path, tokens)
	if err != nil {
		return nil, IOError(err).AddTag(1010)
	}
	return &Toks{
		FilesBuf: []*ScanCloser{scanCloser},
		TokBuf:   nil,
		ImportedFiles: map[string]struct{}{
			path: struct{}{},
		},
	}, nil
}

func reverse(slice []string) {
	last := -1 + len(slice)
	halfLen := len(slice) / 2
	for i := 0; i < halfLen; i++ {
		slice[i], slice[last-i] = slice[last-i], slice[i]
	}
}

func (self *Toks) getLastFile() *ScanCloser {
	if len(self.FilesBuf) == 0 {
		return nil
	}
	return self.FilesBuf[-1+len(self.FilesBuf)]
}

func (self *Toks) popFile() *Error {
	if len(self.FilesBuf) == 0 {
		return MMError("out of files")
	}
	self.FilesBuf[-1+len(self.FilesBuf)].MustClose()
	self.FilesBuf = self.FilesBuf[:-1+len(self.FilesBuf)]
	return nil
}

func (self *Toks) Read() (string, *Error) {
	// Fill the token buffer if it is not already full.
	for len(self.TokBuf) == 0 {
		lastFile := self.getLastFile()
		if lastFile == nil {
			return "", MMError("Unclosed ${ ... $} block at end of file")
		}

		line := lastFile.Text()
		if line.Just {
			self.TokBuf = line.Data
			reverse(self.TokBuf)
		} else {
			err := self.popFile()
			if err != nil {
				return "", err.AddTag(1020)
			}
			if len(self.FilesBuf) == 0 {
				return "", EOF().AddTag(1030)
			}
		}
	}

	tok := self.TokBuf[-1+len(self.TokBuf)]
	self.TokBuf = self.TokBuf[:-1+len(self.TokBuf)]
	Vprint(90, "Token:", tok)
	return tok, nil
}

func (self *Toks) Readf() (string, *Error) {
	tok, err := self.Read()
	if err != nil {
		return "", err.AddTag(1040)
	}
	for tok == "$[" {
		filename, err := self.Read()
		if err != nil {
			return "", err.AddTag(1050)
		}

		endbracket, err := self.Read()
		if endbracket != "$]" {
			return "", err.AddTag(1060)
		}

		filename, aErr := filepath.Abs(filename)
		if aErr != nil {
			// not really an IOError
			return "", IOError(aErr).AddTag(1070)
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
				return "", err.AddTag(1080)
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
				return "", IOError(err).AddTag(1090)
			}
			self.FilesBuf = append(self.FilesBuf, newFile)
			self.ImportedFiles[filename] = struct{}{}
			// Change from original. Print the absolute path to the thing we imported.
			Vprint(5, "Importing file:", filename)
		}
		tok, err = self.Read()
		if err != nil {
			return "", err.AddTag(1100)
		}
	}
	Vprint(80, "Token once included files expanded:", tok)
	return tok, nil
}

func (self *Toks) Readc() (string, *Error) {
	tok, err := self.Readf()
	if err != nil {
		return "", err.AddTag(1110)
	}
	for tok == "$(" {
		tok, err = self.Read()
		if err != nil {
			return "", err.AddTag(1120)
		}
		for tok != "" && tok != "$)" {
			// This errors are worse than the original.
			if strings.Contains(tok, "$(") {
				return "", MMError("token cannot contain $(").AddTag(1130)
			}
			if strings.Contains(tok, "$)") {
				return "", MMError("token cannot contain $)").AddTag(1140)
			}
			tok, err = self.Read()
			if err != nil {
				return "", err.AddTag(1150)
			}
		}
		if tok != "$)" {
			panic("internal error: comment not closed")
		}
		tok, err = self.Readf()
		if err != nil {
			return "", err.AddTag(1160)
		}
	}
	Vprint(70, "Token once comment skipped:", tok)
	return tok, nil
}

// class Frame:
//     """Class of frames, keeping track of the environment."""
//
//     def __init__(self) -> None:
//         """Construct an empty frame."""
//         self.v: set[Var] = set()
//         self.d: set[Dv] = set()
//         self.f: list[Fhyp] = []
//         self.f_labels: dict[Var, Label] = {}
//         self.e: list[Ehyp] = []
//         self.e_labels: dict[tuple[Symbol, ...], Label] = {}
//         # Note: both self.e and self.e_labels are needed since the keys of
//         # self.e_labels form a set, but the order and repetitions of self.e
//         # are needed.

type Frame struct {
	V       map[Var]struct{}
	D       map[Dv]struct{}
	F       []Fhyp
	FLabels map[Var]Label
	E       []Ehyp
	ELabels map[Symbols]Label
	// Only for testing
	Name string
}

// NewFrame initializes the maps.
func NewFrame() *Frame {
	return &Frame{
		V:       map[Var]struct{}{},
		D:       map[Dv]struct{}{},
		F:       nil,
		FLabels: map[Var]Label{},
		E:       nil,
		ELabels: map[Symbols]Label{},
	}
}

// class FrameStack(list[Frame]):
//     """Class of frame stacks, which extends lists (considered and used as
//     stacks).
//     """
//
//     def push(self) -> None:
//         """Push an empty frame to the stack."""
//         self.append(Frame())
//
//     def add_e(self, stmt: Stmt, label: Label) -> None:
//         """Add an essential hypothesis (token tuple) to the frame stack
//         top.
//         """
//         frame = self[-1]
//         frame.e.append(stmt)
//         frame.e_labels[tuple(stmt)] = label
//         # conversion to tuple since dictionary keys must be hashable
//
//     def add_d(self, varlist: list[Var]) -> None:
//         """Add a disjoint variable condition (ordered pair of variables) to
//         the frame stack top.
//         """
//         self[-1].d.update((min(x, y), max(x, y))
//                           for x, y in itertools.product(varlist, varlist)
//                           if x != y)
//
//     def lookup_v(self, tok: Var) -> bool:
//         """Return whether the given token is an active variable."""
//         return any(tok in fr.v for fr in self)
//
//     def lookup_d(self, x: Var, y: Var) -> bool:
//         """Return whether the given ordered pair of tokens belongs to an
//         active disjoint variable statement.
//         """
//         return any((min(x, y), max(x, y)) in fr.d for fr in self)
//
//     def lookup_f(self, var: Var) -> typing.Optional[Label]:
//         """Return the label of the active floating hypothesis which types the
//         given variable.
//         """
//         for frame in self:
//             try:
//                 return frame.f_labels[var]
//             except KeyError:
//                 pass
//         return None  # Variable is not actively typed
//
//     def lookup_e(self, stmt: Stmt) -> Label:
//         """Return the label of the (earliest) active essential hypothesis with
//         the given statement.
//         """
//         stmt_t = tuple(stmt)
//         for frame in self:
//             try:
//                 return frame.e_labels[stmt_t]
//             except KeyError:
//                 pass
//         raise MMKeyError(stmt_t)
//
//     def find_vars(self, stmt: Stmt) -> set[Var]:
//         """Return the set of variables in the given statement."""
//         return {x for x in stmt if self.lookup_v(x)}
//
//     def make_assertion(self, stmt: Stmt) -> Assertion:
//         """Return a quadruple (disjoint variable conditions, floating
//         hypotheses, essential hypotheses, conclusion) describing the given
//         assertion.
//         """
//         e_hyps = [eh for fr in self for eh in fr.e]
//         mand_vars = {tok for hyp in itertools.chain(e_hyps, [stmt])
//                      for tok in hyp if self.lookup_v(tok)}
//         dvs = {(x, y) for fr in self for (x, y)
//                in fr.d if x in mand_vars and y in mand_vars}
//         f_hyps = []
//         for fr in self:
//             for typecode, var in fr.f:
//                 if var in mand_vars:
//                     f_hyps.append((typecode, var))
//                     mand_vars.remove(var)
//         assertion = dvs, f_hyps, e_hyps, stmt
//         vprint(18, 'Make assertion:', assertion)
//         return assertion

// def apply_subst(stmt: Stmt, subst: dict[Var, Stmt]) -> Stmt:
//     """Return the token list resulting from the given substitution
//     (dictionary) applied to the given statement (token list).
//     """
//     result = []
//     for tok in stmt:
//         if tok in subst:
//             result += subst[tok]
//         else:
//             result.append(tok)
//     vprint(20, 'Applying subst', subst, 'to stmt', stmt, ':', result)
//     return result

type FrameStack struct {
	Frames []*Frame
}

func NewFrameStack() *FrameStack {
	return &FrameStack{
		Frames: nil,
	}
}

func (self *FrameStack) Push() {
	self.Frames = append(self.Frames, NewFrame())
}

const GO = 1
const STOP = 2

func (self *FrameStack) Foreach(callback func(f *Frame) int) {
	for i := -1 + len(self.Frames); i >= 0; i-- {
		frame := self.Frames[i]
		status := callback(frame)
		switch status {
		case GO:
			// do nothing
		case STOP:
			return
		default:
			panic(fmt.Sprintf("bad status %v", status))
		}
	}
}

func (self *FrameStack) LastFrame() *Frame {
	out := self.Frames[-1+len(self.Frames)]
	return out
}

// Can this fail?
func (self *FrameStack) AddE(stmt Stmt, label Label) {
	frame := self.LastFrame()
	frame.E = append(frame.E, Ehyp(stmt))
	// Go doesn't have tuples (or another hashable connection)
	// So we convert to a string, painfully.
	// I wonder how slow this will be in benchmarks.
	frame.ELabels[ToSymbols(stmt)] = label
}

// Add all distinct pairs
//
// This is quadratic time for, like, no reason.
//
// TODO: represent a collection of distinct variables as a map
//
//	going to the LEAST variable lexicographically.
func (self *FrameStack) AddD(varlist []Var) {
	frame := self.LastFrame()
	for _, x := range varlist {
		for _, y := range varlist {
			min := x
			max := y
			if string(y) < string(x) {
				min = y
				max = x
			}
			newRecord := Dv{
				First:  min,
				Second: max,
			}
			frame.D[newRecord] = struct{}{}
		}
	}
}

func (self *FrameStack) LookupV(tok Var) bool {
	out := false
	self.Foreach(func(frame *Frame) int {
		_, ok := frame.V[tok]
		if ok {
			out = true
			return STOP
		}
		return GO
	})
	return out
}

func (self *FrameStack) LookupD(x Var, y Var) bool {
	min := x
	max := y
	if string(y) < string(x) {
		min = y
		max = x
	}
	newRecord := Dv{
		First:  min,
		Second: max,
	}
	out := false
	self.Foreach(func(frame *Frame) int {
		_, ok := frame.D[newRecord]
		if ok {
			out = true
			return STOP
		}
		return GO
	})
	return out
}

// return pointer to label of active floating hypothesis or nil
// if none exists.
//
// TODO: OptionalLabel?
func (self *FrameStack) LookupF(va Var) *Label {
	var out *Label
	self.Foreach(func(frame *Frame) int {
		label, ok := frame.FLabels[va]
		if ok {
			out = &label
			return STOP
		}
		return GO
	})
	return out
}

func (self *FrameStack) LookupE(stmt Stmt) (*Label, *Error) {
	var out *Label
	var err *Error
	encoded := ToSymbols(stmt)
	self.Foreach(func(frame *Frame) int {
		label, ok := frame.ELabels[encoded]
		if ok {
			out = &label
			return STOP
		}
		return GO
	})
	if out != nil {
		err = MMKeyError(fmt.Sprintf("%v", stmt))
	}
	return out, err.AddTag(1170)
}

func (self *FrameStack) FindVars(stmt Stmt) map[Var]struct{} {
	out := map[Var]struct{}{}
	for _, x := range stmt {
		v := x.Inl
		if v != nil {
			if self.LookupV(*v) {
				out[*v] = struct{}{}
			}
		}
	}
	return out
}

func (self *FrameStack) MakeAssertion(stmt Stmt) Assertion {
	var eHyps []Ehyp
	mandVars := map[Var]struct{}{}
	dvs := map[Dv]struct{}{}
	var fHyps []Fhyp

	self.Foreach(func(frame *Frame) int {
		for _, eh := range frame.E {
			eHyps = append(eHyps, eh)
		}
		return GO
	})

	// Do the weird thing for "efficiency".
	// Add our statement to eHyps and then remove it.
	// replaces itertools.chain(e_hyps, [stmt])
	eHyps = append(eHyps, Ehyp(stmt))
	for _, hyp := range eHyps {
		for _, tok := range hyp {
			v := tok.Inl
			if v != nil {
				if self.LookupV(*v) {
					mandVars[*v] = struct{}{}
				}
			}
		}
	}
	eHyps = eHyps[:-1+len(eHyps)]

	self.Foreach(func(frame *Frame) int {
		for _, p := range frame.F {
			typecode := p.Typecode
			va := p.V
			if _, ok := mandVars[va]; ok {
				fHyps = append(fHyps, Fhyp{
					Typecode: typecode,
					V:        va,
				})
				delete(mandVars, va)
			}
		}
		return GO
	})

	out := Assertion{
		Dvs: dvs,
		F:   fHyps,
		E:   eHyps,
		S:   stmt,
	}
	Vprint(18, "Make assertion:", out)
	return out
}

// We could just return []Symbol{tok} on failure, but nah let's not do that.
func mapHasVar(subst map[Var]Stmt, tok Symbol) ([]Symbol, bool) {
	v := tok.Inl
	if v == nil {
		return nil, false
	}
	newThing, ok := subst[*v]
	if ok {
		return newThing, true
	}
	return nil, false
}

// This is a non-capture-avoiding substitution because that's what
// metamath is based on.
func (self *Frame) ApplySubst(stmt Stmt, subst map[Var]Stmt) Stmt {
	var result Stmt
	for _, tok := range stmt {
		newThing, ok := mapHasVar(subst, tok)
		if ok {
			result = append(result, newThing...)
		} else {
			result = append(result, tok)
		}
	}
	Vprint(20, "Applying subst", subst, "to stmt", stmt, ":", result)
	return result
}

// class MM:
//     """Class of ("abstract syntax trees" describing) Metamath databases."""
//
//     def __init__(self, begin_label: Label, stop_label: Label) -> None:
//         """Construct an empty Metamath database."""
//         self.constants: set[Const] = set()
//         self.fs = FrameStack()
//         self.labels: dict[Label, FullStmt] = {}
//         self.begin_label = begin_label
//         self.stop_label = stop_label
//         self.verify_proofs = not self.begin_label
//
//     def add_c(self, tok: Const) -> None:
//         """Add a constant to the database."""
//         if tok in self.constants:
//             raise MMError(
//                 'Constant already declared: {}'.format(tok))
//         if self.fs.lookup_v(tok):
//             raise MMError(
//                 'Trying to declare as a constant an active variable: {}'.format(tok))
//         self.constants.add(tok)
//
//     def add_v(self, tok: Var) -> None:
//         """Add a variable to the frame stack top (that is, the current frame)
//         of the database.  Allow local variable declarations.
//         """
//         if self.fs.lookup_v(tok):
//             raise MMError('var already declared and active: {}'.format(tok))
//         if tok in self.constants:
//             raise MMError(
//                 'var already declared as constant: {}'.format(tok))
//         self.fs[-1].v.add(tok)
//
//     def add_f(self, typecode: Const, var: Var, label: Label) -> None:
//         """Add a floating hypothesis (ordered pair (variable, typecode)) to
//         the frame stack top (that is, the current frame) of the database.
//         """
//         if not self.fs.lookup_v(var):
//             raise MMError('var in $f not declared: {}'.format(var))
//         if typecode not in self.constants:
//             raise MMError('typecode in $f not declared: {}'.format(typecode))
//         if any(var in fr.f_labels for fr in self.fs):
//             raise MMError(
//                 ("var in $f already typed by an active " +
//                  "$f-statement: {}").format(var))
//         frame = self.fs[-1]
//         frame.f.append((typecode, var))
//         frame.f_labels[var] = label
//
//     def readstmt_aux(
//             self,
//             stmttype: Stmttype,
//             toks: Toks,
//             end_token: str) -> Stmt:
//         """Read tokens from the input (assumed to be at the beginning of a
//         statement) and return the list of tokens until the end_token
//         (typically "$=" or "$.").
//         """
//         stmt = []
//         tok = toks.readc()
//         while tok and tok != end_token:
//             is_active_var = self.fs.lookup_v(tok)
//             if stmttype in {'$d', '$e', '$a', '$p'} and not (
//                     tok in self.constants or is_active_var):
//                 raise MMError(
//                     "Token {} is not an active symbol".format(tok))
//             if stmttype in {
//                 '$e',
//                 '$a',
//                     '$p'} and is_active_var and not self.fs.lookup_f(tok):
//                 raise MMError(("Variable {} in {}-statement is not typed " +
//                                "by an active $f-statement).").format(tok, stmttype))
//             stmt.append(tok)
//             tok = toks.readc()
//         if not tok:
//             raise MMError(
//                 "Unclosed {}-statement at end of file.".format(stmttype))
//         assert tok == end_token
//         vprint(20, 'Statement:', stmt)
//         return stmt
//
//     def read_non_p_stmt(self, stmttype: Stmttype, toks: Toks) -> Stmt:
//         """Read tokens from the input (assumed to be at the beginning of a
//         non-$p-statement) and return the list of tokens until the next
//         end-statement token '$.'.
//         """
//         return self.readstmt_aux(stmttype, toks, end_token="$.")
//
//     def read_p_stmt(self, toks: Toks) -> tuple[Stmt, Stmt]:
//         """Read tokens from the input (assumed to be at the beginning of a
//         p-statement) and return the couple of lists of tokens (stmt, proof)
//         appearing in "$p stmt $= proof $.".
//         """
//         stmt = self.readstmt_aux("$p", toks, end_token="$=")
//         proof = self.readstmt_aux("$=", toks, end_token="$.")
//         return stmt, proof
//
//     def read(self, toks: Toks) -> None:
//         """Read the given token list to update the database and verify its
//         proofs.
//         """
//         self.fs.push()
//         label = None
//         tok = toks.readc()
//         while tok and tok != '$}':
//             if tok == '$c':
//                 for tok in self.read_non_p_stmt(tok, toks):
//                     self.add_c(tok)
//             elif tok == '$v':
//                 for tok in self.read_non_p_stmt(tok, toks):
//                     self.add_v(tok)
//             elif tok == '$f':
//                 stmt = self.read_non_p_stmt(tok, toks)
//                 if not label:
//                     raise MMError(
//                         '$f must have label (statement: {})'.format(stmt))
//                 if len(stmt) != 2:
//                     raise MMError(
//                         '$f must have length two but is {}'.format(stmt))
//                 self.add_f(stmt[0], stmt[1], label)
//                 self.labels[label] = ('$f', [stmt[0], stmt[1]])
//                 label = None
//             elif tok == '$e':
//                 if not label:
//                     raise MMError('$e must have label')
//                 stmt = self.read_non_p_stmt(tok, toks)
//                 self.fs.add_e(stmt, label)
//                 self.labels[label] = ('$e', stmt)
//                 label = None
//             elif tok == '$a':
//                 if not label:
//                     raise MMError('$a must have label')
//                 self.labels[label] = (
//                     '$a', self.fs.make_assertion(
//                         self.read_non_p_stmt(tok, toks)))
//                 label = None
//             elif tok == '$p':
//                 if not label:
//                     raise MMError('$p must have label')
//                 stmt, proof = self.read_p_stmt(toks)
//                 dvs, f_hyps, e_hyps, conclusion = self.fs.make_assertion(stmt)
//                 if self.verify_proofs:
//                     vprint(2, 'Verify:', label)
//                     self.verify(f_hyps, e_hyps, conclusion, proof)
//                 self.labels[label] = ('$p', (dvs, f_hyps, e_hyps, conclusion))
//                 label = None
//             elif tok == '$d':
//                 self.fs.add_d(self.read_non_p_stmt(tok, toks))
//             elif tok == '${':
//                 self.read(toks)
//             elif tok == '$)':
//                 raise MMError("Unexpected '$)' while not within a comment")
//             elif tok[0] != '$':
//                 if tok in self.labels:
//                     raise MMError("Label {} multiply defined.".format(tok))
//                 label = tok
//                 vprint(20, 'Label:', label)
//                 if label == self.stop_label:
//                     # TODO: exit gracefully the nested calls to self.read()
//                     sys.exit(0)
//                 if label == self.begin_label:
//                     self.verify_proofs = True
//             else:
//                 raise MMError("Unknown token: '{}'.".format(tok))
//             tok = toks.readc()
//         self.fs.pop()
//
//     def treat_step(self,
//                    step: FullStmt,
//                    stack: list[Stmt]) -> None:
//         """Carry out the given proof step (given the label to treat and the
//         current proof stack).  This modifies the given stack in place.
//         """
//         vprint(10, 'Proof step:', step)
//         if is_hypothesis(step):
//             _steptype, stmt = step
//             stack.append(stmt)
//         elif is_assertion(step):
//             _steptype, assertion = step
//             dvs0, f_hyps0, e_hyps0, conclusion0 = assertion
//             npop = len(f_hyps0) + len(e_hyps0)
//             sp = len(stack) - npop
//             if sp < 0:
//                 raise MMError(
//                     ("Stack underflow: proof step {} requires too many " +
//                      "({}) hypotheses.").format(
//                         step,
//                         npop))
//             subst: dict[Var, Stmt] = {}
//             for typecode, var in f_hyps0:
//                 entry = stack[sp]
//                 if entry[0] != typecode:
//                     raise MMError(
//                         ("Proof stack entry {} does not match floating " +
//                          "hypothesis ({}, {}).").format(entry, typecode, var))
//                 subst[var] = entry[1:]
//                 sp += 1
//             vprint(15, 'Substitution to apply:', subst)
//             for h in e_hyps0:
//                 entry = stack[sp]
//                 subst_h = apply_subst(h, subst)
//                 if entry != subst_h:
//                     raise MMError(("Proof stack entry {} does not match " +
//                                    "essential hypothesis {}.")
//                                   .format(entry, subst_h))
//                 sp += 1
//             for x, y in dvs0:
//                 vprint(16, 'dist', x, y, subst[x], subst[y])
//                 x_vars = self.fs.find_vars(subst[x])
//                 y_vars = self.fs.find_vars(subst[y])
//                 vprint(16, 'V(x) =', x_vars)
//                 vprint(16, 'V(y) =', y_vars)
//                 for x0, y0 in itertools.product(x_vars, y_vars):
//                     if x0 == y0 or not self.fs.lookup_d(x0, y0):
//                         raise MMError("Disjoint variable violation: " +
//                                       "{} , {}".format(x0, y0))
//             del stack[len(stack) - npop:]
//             stack.append(apply_subst(conclusion0, subst))
//         vprint(12, 'Proof stack:', stack)
//
//     def treat_normal_proof(self, proof: list[str]) -> list[Stmt]:
//         """Return the proof stack once the given normal proof has been
//         processed.
//         """
//         stack: list[Stmt] = []
//         active_hypotheses = {label for frame in self.fs for labels in (frame.f_labels, frame.e_labels) for label in labels.values()}
//         for label in proof:
//             stmt_info = self.labels.get(label)
//             if stmt_info:
//                 label_type = stmt_info[0]
//                 if label_type in {'$e', '$f'}:
//                     if label in active_hypotheses:
//                         self.treat_step(stmt_info, stack)
//                     else:
//                         raise MMError(f"The label {label} is the label of a nonactive hypothesis.")
//                 else:
//                     self.treat_step(stmt_info, stack)
//             else:
//                 raise MMError(f"No statement information found for label {label}")
//         return stack
//
//     def treat_compressed_proof(
//             self,
//             f_hyps: list[Fhyp],
//             e_hyps: list[Ehyp],
//             proof: list[str]) -> list[Stmt]:
//         """Return the proof stack once the given compressed proof for an
//         assertion with the given $f and $e-hypotheses has been processed.
//         """
//         # Preprocessing and building the lists of proof_ints and labels
//         flabels = [self.fs.lookup_f(v) for _, v in f_hyps]
//         elabels = [self.fs.lookup_e(s) for s in e_hyps]
//         plabels = flabels + elabels  # labels of implicit hypotheses
//         idx_bloc = proof.index(')')  # index of end of label bloc
//         plabels += proof[1:idx_bloc]  # labels which will be referenced later
//         compressed_proof = ''.join(proof[idx_bloc + 1:])
//         vprint(5, 'Referenced labels:', plabels)
//         label_end = len(plabels)
//         vprint(5, 'Number of referenced labels:', label_end)
//         vprint(5, 'Compressed proof steps:', compressed_proof)
//         vprint(5, 'Number of steps:', len(compressed_proof))
//         proof_ints = []  # integers referencing the labels in 'labels'
//         cur_int = 0  # counter for radix conversion
//         for ch in compressed_proof:
//             if ch == 'Z':
//                 proof_ints.append(-1)
//             elif 'A' <= ch <= 'T':
//                 proof_ints.append(20 * cur_int + ord(ch) - 65)  # ord('A') = 65
//                 cur_int = 0
//             else:  # 'U' <= ch <= 'Y'
//                 cur_int = 5 * cur_int + ord(ch) - 84  # ord('U') = 85
//         vprint(5, 'Integer-coded steps:', proof_ints)
//         # Processing of the proof
//         stack: list[Stmt] = []  # proof stack
//         # statements saved for later reuse (marked with a 'Z')
//         saved_stmts = []
//         # can be recovered as len(saved_stmts) but less efficient
//         n_saved_stmts = 0
//         for proof_int in proof_ints:
//             if proof_int == -1:  # save the current step for later reuse
//                 stmt = stack[-1]
//                 vprint(15, 'Saving step', stmt)
//                 saved_stmts.append(stmt)
//                 n_saved_stmts += 1
//             elif proof_int < label_end:
//                 # proof_int denotes an implicit hypothesis or a label in the
//                 # label bloc
//                 self.treat_step(self.labels[plabels[proof_int] or ''], stack)
//             elif proof_int >= label_end + n_saved_stmts:
//                 MMError(
//                     ("Not enough saved proof steps ({} saved but calling " +
//                     "the {}th).").format(
//                         n_saved_stmts,
//                         proof_int))
//             else:  # label_end <= proof_int < label_end + n_saved_stmts
//                 # proof_int denotes an earlier proof step marked with a 'Z'
//                 # A proof step that has already been proved can be treated as
//                 # a dv-free and hypothesis-free axiom.
//                 stmt = saved_stmts[proof_int - label_end]
//                 vprint(15, 'Reusing step', stmt)
//                 self.treat_step(
//                     ('$a',
//                      (set(), [], [], stmt)),
//                     stack)
//         return stack
//
//     def verify(
//             self,
//             f_hyps: list[Fhyp],
//             e_hyps: list[Ehyp],
//             conclusion: Stmt,
//             proof: list[str]) -> None:
//         """Verify that the given proof (in normal or compressed format) is a
//         correct proof of the given assertion.
//         """
//         # It would not be useful to also pass the list of dv conditions of the
//         # assertion as an argument since other dv conditions corresponding to
//         # dummy variables should be 'lookup_d'ed anyway.
//         if proof[0] == '(':  # compressed format
//             stack = self.treat_compressed_proof(f_hyps, e_hyps, proof)
//         else:  # normal format
//             stack = self.treat_normal_proof(proof)
//         vprint(10, 'Stack at end of proof:', stack)
//         if not stack:
//             raise MMError(
//                 "Empty stack at end of proof.")
//         if len(stack) > 1:
//             raise MMError(
//                 "Stack has more than one entry at end of proof (top " +
//                 "entry: {} ; proved assertion: {}).".format(
//                     stack[0],
//                     conclusion))
//         if stack[0] != conclusion:
//             raise MMError(("Stack entry {} does not match proved " +
//                           " assertion {}.").format(stack[0], conclusion))
//         vprint(3, 'Correct proof!')
//
//     def dump(self) -> None:
//         """Print the labels of the database."""
//         print(self.labels)

type MM struct {
	BeginLabel   *Label
	EndLabel     *Label
	Constants    map[Const]struct{}
	FS           *FrameStack
	Labels       map[Label]FullStmt
	VerifyProofs bool
}

func NewMM(beginLabel *Label, endLabel *Label) *MM {
	return &MM{
		BeginLabel:   beginLabel,
		EndLabel:     endLabel,
		Constants:    map[Const]struct{}{},
		FS:           NewFrameStack(),
		Labels:       map[Label]FullStmt{},
		VerifyProofs: beginLabel == nil,
	}
}

func (self *MM) AddC(tok Const) *Error {
	_, ok := self.Constants[tok]
	if ok {
		return MMError(fmt.Sprintf("constant %q already declared", tok))
	}
	self.Constants[tok] = struct{}{}
	return nil
}

func (self *MM) AddV(tok Var) *Error {
	if self.FS.LookupV(tok) {
		return MMError(fmt.Sprintf("variable %q already declared and active", tok))
	}
	frame := self.FS.LastFrame()
	if frame == nil {
		panic("impossible: frame stack is empty")
	}
	frame.V[tok] = struct{}{}
	return nil
}

func (self *MM) AddF(typecode Const, va Var, label Label) *Error {
	if self.FS.LookupV(va) {
		// Good. We need the variable to already exist.
	} else {
		return MMError(fmt.Sprintf("var in $f not declared: %q", va))
	}
	if _, ok := self.Constants[typecode]; ok {
		// Good. The constant must exist already.
	} else {
		return MMError(fmt.Sprintf("typecode in $f not declared: %q", typecode))
	}

	alreadyTyped := false
	self.FS.Foreach(func(frame *Frame) int {
		_, ok := frame.FLabels[va]
		if ok {
			alreadyTyped = true
			return STOP
		}
		return GO
	})
	if alreadyTyped {
		return MMError(fmt.Sprintf("var in $f already typed by an active $f-statement: %q", va))
	}
	frame := self.FS.LastFrame()
	if frame == nil {
		panic("impossible")
	}
	frame.F = append(frame.F, Fhyp{
		Typecode: typecode,
		V:        va,
	})
	frame.FLabels[va] = label
	return nil
}

func (self *MM) LookupSymbolByName(tok string) (*Symbol, *Var, *Const) {
	isActiveVar := self.FS.LookupV(Var(tok))
	_, isConstant := self.Constants[Const(tok)]
	switch {
	case isActiveVar && isConstant:
		panic(fmt.Sprintf("string %q is both var and const", tok))
	case isActiveVar:
		va := Var(tok)
		return &Symbol{Inl: &va}, &va, nil
	case isConstant:
		constant := Const(tok)
		return &Symbol{Inr: &constant}, nil, &constant
	default:
		return nil, nil, nil
	}
}

// endToken is "$=" or "$.".
// endToken shouldn't be a string this function is too general.
func (self *MM) ReadStmtAux(stmttype StmtType, toks *Toks, endToken string) (Stmt, *Error) {
	var stmt Stmt
	tok, err := toks.Readc()
	if err != nil {
		return nil, err.AddTag(1180)
	}
	for tok != "" && tok != endToken {
		// What do we do if the symbol doesn't exist?
		sym, va, constant := self.LookupSymbolByName(tok)
		// Validate active symbol.
		switch stmttype {
		case "$d", "$e", "$a", "$p":
			if va == nil && constant == nil {
				return nil, MMError(fmt.Sprintf("Token %q is not an active symbol", tok)).AddTag(1181)
			}
		}
		// Validate symbol typed by hypothesis.
		switch stmttype {
		case "$e", "$a", "$p":
			if va != nil && self.FS.LookupF(*va) == nil {
				return nil, MMError(fmt.Sprintf("Variable %q in %s-statement is not typed by an active $f-statement", tok, stmttype)).AddTag(1182)
			}
		}
		stmt = append(stmt, *sym)
		tok, err = toks.Readc()
		if err != nil {
			return nil, err.AddTag(1190)
		}
	}
	if tok == "" {
		return nil, MMError(fmt.Sprintf("Unclosed %q-statement at the end of file", stmttype))
	}
	if tok != endToken {
		panic("tok must equal endToken")
	}
	Vprint(20, "Statement:", stmt)
	return stmt, nil
}

func (self *MM) ReadNonPStatement(stmttype StmtType, toks *Toks) (Stmt, *Error) {
	return self.ReadStmtAux(stmttype, toks, "$.")
}

func (self *MM) ReadPStatement(toks *Toks) (Stmt, Stmt, *Error) {
	stmt, err := self.ReadStmtAux("$p", toks, "$=")
	if err != nil {
		return nil, nil, err.AddTag(1200)
	}
	proof, err := self.ReadStmtAux("$=", toks, "$.")
	if err != nil {
		return nil, nil, err.AddTag(1210)
	}
	return stmt, proof, nil
}

// func (self *MM) Read(toks *Toks) *Error {
// 	self.FS.Push()
// 	var label *Label
// 	tok, err := toks.Readc()
// 	if err != nil {
// 		return err.AddTag(1220)
// 	}
// 	for tok != "" && tok != "$}" {
// 		switch tok {
// 		case "$c":
// 			stmt, err := self.ReadNonPStatement(StmtType(tok), toks)
// 			if err != nil {
// 				return err.AddTag(1230)
// 			}
// 			for _, w := range stmt {
// 				if err := self.AddC(Const(w)); err != nil {
// 					return err.AddTag(1240)
// 				}
// 			}
// 		case "$v":
// 			stmt, err := self.ReadNonPStatement(tok, toks)
// 			if err != nil {
// 				return err.AddTag(1245)
// 			}
// 			for _, w := range stmt {
// 				if err := self.AddV(MakeVar(w)); err != nil {
// 					return err.AddTag(1250)
// 				}
// 			}
// 		case "$f":
// 			stmt, err := self.ReadNonPStatement(tok, toks)
// 			if err != nil {
// 				return err.AddTag(1260)
// 			}
// 			if label == nil {
// 				return MMError(fmt.Sprintf("$f must have label (statement: %v)", stmt)).AddTag(1270)
// 			}
// 			if len(stmt) != 2 {
// 				return MMError(fmt.Sprintf("$f must have length 2 but is %v", stmt)).AddTag(1280)
// 			}
// 			if err := self.AddF(stmt[0], stmt[1], label); err != nil {
// 				return err.AddTag(1290)
// 			}
// 
// 			panic("not yet implemented")
// 		}
// 	}
// }

func main() {
	fmt.Printf("hi\n")
}

// if __name__ == '__main__':
//     """Parse the arguments and verify the given Metamath database."""
//     parser = argparse.ArgumentParser(description="""Verify a Metamath database.
//       The grammar of the whole file is verified.  Proofs are verified between
//       the statements with labels BEGIN_LABEL (included) and STOP_LABEL (not
//       included).
//
//       One can also use bash redirections:
//          '$ python3 mmverify.py < file.mm 2> file.log'
//       in place of
//          '$ python3 mmverify.py file.mm --logfile file.log'
//       but this fails in case 'file.mm' contains (directly or not) a recursive
//       inclusion statement '$[ file.mm $]'.""")
//     parser.add_argument(
//         'database',
//         nargs='?',
//         type=argparse.FileType(
//             mode='r',
//             encoding='ascii'),
//         default=sys.stdin,
//         help="""database (Metamath file) to verify, expressed using relative
//           path (defaults to <stdin>)""")
//     parser.add_argument(
//         '-l',
//         '--logfile',
//         dest='logfile',
//         type=argparse.FileType(
//             mode='w',
//             encoding='ascii'),
//         default=sys.stderr,
//         help="""file to output logs, expressed using relative path (defaults to
//           <stderr>)""")
//     parser.add_argument(
//         '-v',
//         '--verbosity',
//         dest='verbosity',
//         default=0,
//         type=int,
//         help='verbosity level (default=0 is mute; higher is more verbose)')
//     parser.add_argument(
//         '-b',
//         '--begin-label',
//         dest='begin_label',
//         type=str,
//         help="""label where to begin verifying proofs (included, if it is a
//           provable statement)""")
//     parser.add_argument(
//         '-s',
//         '--stop-label',
//         dest='stop_label',
//         type=str,
//         help='label where to stop verifying proofs (not included)')
//     args = parser.parse_args()
//     verbosity = args.verbosity
//     db_file = args.database
//     logfile = args.logfile
//     vprint(1, 'mmverify.py -- Proof verifier for the Metamath language')
//     mm = MM(args.begin_label, args.stop_label)
//     vprint(1, 'Reading source file "{}"...'.format(db_file.name))
//     mm.read(Toks(db_file))
//     vprint(1, 'No errors were found.')
//     # mm.dump()
