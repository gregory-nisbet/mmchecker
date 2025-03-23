package core

import "fmt"

func TreatNormalProof(mm *MM, proof []string) (*ProofStack, error) {
	stack := NewProofStack()
	activeHypotheses := map[Label]TUnit{}

	mm.FS.Foreach(func(frame *Frame) int8 {
		for _, label := range frame.FLabels {
			activeHypotheses[label] = Unit
		}
		for _, label := range frame.ELabels {
			activeHypotheses[label] = Unit
		}
		return GO
	})

	for _, label := range proof {
		label := Label(label)
		stmtInfo, ok := mm.Labels[label]
		if !ok {
			return nil, MMError{fmt.Errorf("no statement information found for label %q", label)}
		}
		labelType := stmtInfo.SType
		if labelType == "$e" || labelType == "$f" {
			if _, ok := activeHypotheses[label]; ok {
				if err := stack.TreatStep(mm, stmtInfo); err != nil {
					return nil, fmt.Errorf("treating %q step: %w", labelType, err)
				}
			} else {
				return nil, MMError{fmt.Errorf("the label %q is the label of a nonactive hypothesis", label)}
			}
		} else {
			if err := stack.TreatStep(mm, stmtInfo); err != nil {
				return nil, fmt.Errorf("treating non-{$e,$f} %q step: %w", labelType, err)
			}
		}
	}
	return stack, nil
}
