package barrelfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TriggerAdapter wraps the given barrelfile.Trigger in a barrel.Trigger.
type TriggerAdapter struct {
	FileTrigger Trigger
}

// Trigger triggers the underlying FileTrigger, if the provided io.Writer is a
// reference to os.File. Errors and values form underlying FileTrigger are
// returned directly.
//
// If the provided writer is not an io.Writer, then a non-nil error is returned.
func (t TriggerAdapter) Trigger(w io.Writer, p []byte) (bool, error) {
	file, ok := w.(*os.File)
	if !ok {
		return false, fmt.Errorf("writer not reference to os.File")
	}
	return t.FileTrigger.Trigger(file.Name(), p)
}

// RotatorAdapter wraps the given barrelfile.Trigger in a barrel.Trigger.
type RotatorAdapter struct {
	// FileRotator used to rotate the file.
	FileRotator Rotator

	// Flag to be used to Open the the new file, os.O_CREATE is added
	// automatically.
	OpenFlag int
}

// Rotate rotates the given writer using the underlying FileRotator, if the
// provided writer is a reference to os.File. Errors and values from
// underlying FileRotator are returned directly. The provided os.File is closed
// before rotating it via FileRotator, and the file at path returned by the
// FileRotator is opened and returned as an io.Writer.
//
// If there are any errors while closing the original os.File before rotation,
// then rotation is not performed.
//
// While opening the file at path returned by FileRotator, given OpenFlag are
// used, while os.O_CREATE is added in addition to given flags.
//
// If the provided writer is not an io.Writer, then a non-nil error is returned.
func (r RotatorAdapter) Rotate(w io.Writer) (io.Writer, error) {
	file, ok := w.(*os.File)
	if !ok {
		return w, fmt.Errorf("writer not reference to os.File")
	}
	stat, err := file.Stat()
	if err != nil {
		return w, fmt.Errorf("stat current file: %w", err)
	}
	if err := file.Sync(); err != nil {
		return w, fmt.Errorf("sync current file: %w", err)
	}
	if err := file.Close(); err != nil {
		return w, fmt.Errorf("close current file: %w", err)
	}
	absFilePath, err := filepath.Abs(file.Name())
	if err != nil {
		return w, fmt.Errorf("determine current file absolute path: %w", err)
	}
	newFilePath, err := r.FileRotator.Rotate(absFilePath)
	if err != nil {
		return w, fmt.Errorf("rotate file: %w", err)
	}
	absNewFilePath, err := filepath.Abs(newFilePath)
	if err != nil {
		return w, fmt.Errorf("determine new file absolute path: %w", err)
	}
	newFile, err := os.OpenFile(absNewFilePath, os.O_CREATE|r.OpenFlag, stat.Mode())
	if err != nil {
		return nil, fmt.Errorf("open new file: %w", err)
	}
	return newFile, nil
}
