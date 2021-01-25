package barrelfile

import (
	"fmt"
)

// Rotator rotates the file at given path. If there is any error while rotating
// the file then non-nil error is returned.
//
// Implementations should ensure that rotations performed are safe and do not
// corrupt or destroy the data.
type Rotator interface {
	Rotate(path string) (string, error)
}

// Transformer takes a file path, transforms the file and returns the path
// of resulting file. If any error occurs non-nil error is returned.
//
// Implementations should ensure that transformations being performed are safe
// and do not corrupt or destroy the data.
type Transformer interface {
	Transform(path string) (string, error)
}

// IdentityRotator rotates the file without any modifications.
type IdentityRotator struct {
}

// Rotate rotates the file by returning the same path without any modifications
// to the original path.
func (i IdentityRotator) Rotate(path string) (string, error) {
	return path, nil
}

// TransformRotator transforms the file with the provided transformers before
// rotation.
type TransformRotator struct {
	// Transformers transform the file. These are executed sequentially.
	Transformers []Transformer

	// Rotator is used to rotate the file after transforms are complete.
	Rotator Rotator
}

// Rotate executes the given transformers, and then rotates the file at the
// given path using the Rotator provided.
func (r TransformRotator) Rotate(path string) (string, error) {
	p := path
	for i, transformer := range r.Transformers {
		np, err := transformer.Transform(p)
		if err != nil {
			return path, fmt.Errorf("transformer[%d]: %w", i, err)
		}
		p = np
	}
	return r.Rotator.Rotate(path)
}
