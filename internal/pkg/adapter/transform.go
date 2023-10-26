package adapter

import (
	"io"
)

// Workout represents a fitness workout to be used within the program.
type Workout struct {
}

type Transformer interface {
	Transform(io.Reader) (Workout, error)
}

type StrongApp struct{}

func (sa *StrongApp) Transform(r io.Reader) (Workout, error) {
	// transform logic specific to how the strong app produces workout data
	return Workout{}, nil
}

// TransformData takes a concreate implementation of Transformer to perform a transform on io.Reader
func TransformData(t Transformer, r io.Reader) (Workout, error) {
	return t.Transform(r)
}
