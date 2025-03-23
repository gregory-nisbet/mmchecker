package core

type Frame struct {
	V       map[string]TUnit
	D       map[Dv]TUnit
	F       []Fhyp
	FLabels map[string]Label
	E       []Ehyp
	ELabels map[Symbols]Label
	// Only for testing
	Name string
}

// NewFrame initializes the maps.
func NewFrame() *Frame {
	return &Frame{
		V:       map[string]TUnit{},
		D:       map[Dv]TUnit{},
		F:       nil,
		FLabels: map[string]Label{},
		E:       nil,
		ELabels: map[Symbols]Label{},
	}
}
