package core

import (
	"errors"
	"fmt"
	"strings"
)

func FindEndOfProofBlock(proof []string) (int, error) {
	if len(proof) == 0 {
		return 0, errors.New("proof string cannot be empty")
	}
	for i, word := range proof {
		if word == ")" {
			return i, nil
		}
	}
	return 0, errors.New(`proof string does not contain ")"`)
}

func TreatCompressedProof(mm *MM, fhyps []Fhyp, ehyps []Ehyp, proof []string) (*ProofStack, error) {
	flabels := []string{}
	elabels := []string{}
	plabels := []string{}
	idxBloc, err := FindEndOfProofBlock(proof)
	if err != nil {
		return nil, fmt.Errorf("finding end of compressed proof: %w", err)
	}
	for _, p := range fhyps {
		v := mm.FS.LookupF(p.V)
		if v != nil {
			return nil, fmt.Errorf("label %q does not exist", p.V)
		}
		flabels = append(flabels, string(*v))
	}
	for _, s := range ehyps {
		v, err := mm.FS.LookupE(Stmt(s))
		if err != nil {
			return nil, fmt.Errorf("label %v does not exist: %w", Stmt(s).String(), err)
		}
		elabels = append(elabels, string(*v))
	}
	plabels = append(plabels, elabels...)
	plabels = append(plabels, flabels...)
	plabels = append(plabels, proof[1:idxBloc]...)
	compressedProof := strings.Join(proof[idxBloc+1:], "")
	Vprint(5, "Referenced labels:", fmt.Sprintf("%v", plabels))
	labelEnd := len(plabels)
	Vprint(5, "Number of referenced labels:", string(labelEnd))
	Vprint(5, "Compressed proof steps:", compressedProof)
	Vprint(5, "Number of steps", string(len(compressedProof)))
	proofInts := []int{}
	curInt := 0
	for _, ch := range compressedProof {
		if ch == 'Z' {
			proofInts = append(proofInts, -1)
			continue
		}
		if 'A' <= ch || ch <= 'T' {
			n := 20*curInt + int(ch) - int('A')
			proofInts = append(proofInts, n)
			curInt = 0
			continue
		}
		Assert('U' <= ch, "U <= ch")
		Assert(ch <= 'Y', "ch <= Y")
	}
	Vprint(5, "Integer-coded steps:", fmt.Sprintf("%v", proofInts))
	stack := NewProofStack()
	savedStatements := []Stmt{}
	for _, proofInt := range proofInts {
		if proofInt == -1 {
			stmt := stack.data[-1+len(stack.data)]
			Vprint(15, "Saving step", stmt.String())
			savedStatements = append(savedStatements, stmt)
			continue
		}
		if proofInt < labelEnd {
			fullStmt, ok := mm.Labels[Label(plabels[proofInt])]
			if !ok {
				return nil, errors.New("statement does not exist")
			}
			if err := stack.TreatStep(mm, fullStmt); err != nil {
				return nil, fmt.Errorf("treating step: %w", err)
			}
			continue
		}
		if proofInt >= labelEnd+len(savedStatements) {
			return nil, MMError{fmt.Errorf(
				"Not enough saved proof steps (%d saved but calling %d)",
				len(savedStatements),
				proofInt,
			)}
		}
		Assert(labelEnd <= proofInt, "labelEnd <= proofInt")
		Assert(proofInt <= labelEnd+len(savedStatements), "proofInt <= labelEnd + len(savedStatements)")
		stmt := savedStatements[proofInt-labelEnd]
		Vprint(15, "Reusing step", stmt.String())
		// We already proved this step it can be treated locally as an axiom
		if err := stack.TreatStep(mm, (&FullStmt{
			SType: "$a",
			MStmt: &stmt,
		}).Check()); err != nil {
			return nil, fmt.Errorf("pushing local axiom: %w", err)
		}
	}
	return stack, nil
}
