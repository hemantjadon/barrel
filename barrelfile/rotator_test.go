package barrelfile

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/matryer/is"
)

func TestIdentityRotator_Rotate(t *testing.T) {
	t.Parallel()

	t.Run("no file at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		// Hopefully this is random
		randomPath := filepath.Join(dir, "g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn")

		rotator := IdentityRotator{}

		path, err := rotator.Rotate(randomPath)
		r.NoErr(err)               // should not be any error
		r.True(path == randomPath) // path should be same as original path
	})

	t.Run("a file at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file := NewFile(t, dir, "transform-rotator-*")

		rotator := IdentityRotator{}

		path, err := rotator.Rotate(file)
		r.NoErr(err)         // should not be any error
		r.True(path == file) // path should be same as original path
	})
}

func TestTransformRotator_Rotate(t *testing.T) {
	t.Parallel()

	t.Run("no transformers", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file := NewFile(t, dir, "transform-rotator-*")

		newPath := filepath.Join(dir, "new")

		rotator := TransformRotator{Rotator: fixedRotator(newPath)}

		path, err := rotator.Rotate(file)
		r.NoErr(err)            // should not be any error
		r.True(path == newPath) // path should be same as given by the rotator
	})

	t.Run("some faulty transformers", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file := NewFile(t, dir, "transform-rotator-*")

		rotator := TransformRotator{
			Transformers: []Transformer{
				noopTransformer(),
				noopTransformer(),
				faultyTransformer(errTransformer),
				noopTransformer(),
			},
		}

		path, err := rotator.Rotate(file)
		r.True(errors.Is(err, errTransformer)) // error should wrap the underling transformer error
		r.True(path == path)                   // path should be same as original path
	})

	t.Run("all correct transformers", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file := NewFile(t, dir, "transform-rotator-*")

		newPath := filepath.Join(dir, "new")

		rotator := TransformRotator{
			Transformers: []Transformer{
				noopTransformer(),
				noopTransformer(),
				noopTransformer(),
			},
			Rotator: fixedRotator(newPath),
		}

		path, err := rotator.Rotate(file)
		r.NoErr(err)            // should not be any error
		r.True(path == newPath) // path should be same as given by the rotator
	})
}
