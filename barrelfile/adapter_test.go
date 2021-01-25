package barrelfile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestTriggerAdapter_Trigger(t *testing.T) {
	t.Parallel()

	t.Run("trigger with non file writer", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		triggerAdapter := TriggerAdapter{FileTrigger: fixedTrigger(true)}

		v, err := triggerAdapter.Trigger(&bytes.Buffer{}, nil)
		r.True(err != nil) // error should be non-nil
		r.True(v == false)
	})

	t.Run("trigger with file writer and faulty file trigger", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)
		filePath := NewFile(t, dir, "trigger-adapter-*")
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0644)
		r.NoErr(err) // should not be any error
		t.Cleanup(func() {
			err := file.Close()
			r.NoErr(err) // should not be any error
		})

		triggerAdapter := TriggerAdapter{FileTrigger: faultyTrigger(errTrigger)}

		v, err := triggerAdapter.Trigger(file, nil)
		r.True(errors.Is(err, errTrigger)) // error should wrap underlying FileTrigger error
		r.True(v == false)
	})

	t.Run("trigger with file writer and correct file trigger", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)
		filePath := NewFile(t, dir, "trigger-adapter-*")
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0644)
		r.NoErr(err) // should not be any error
		t.Cleanup(func() {
			err := file.Close()
			r.NoErr(err) // should not be any error
		})

		triggerAdapter := TriggerAdapter{FileTrigger: fixedTrigger(true)}

		v, err := triggerAdapter.Trigger(file, nil)
		r.NoErr(err) // should not be any error
		r.True(v == true)

		triggerAdapter = TriggerAdapter{FileTrigger: fixedTrigger(false)}

		v, err = triggerAdapter.Trigger(file, nil)
		r.NoErr(err) // should not be any error
		r.True(v == false)
	})
}

func TestRotatorAdapter_Rotate(t *testing.T) {
	t.Parallel()

	t.Run("rotator with non file writer", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		rotatorAdapter := RotatorAdapter{FileRotator: fixedRotator("")}

		buf := &bytes.Buffer{}
		newWriter, err := rotatorAdapter.Rotate(buf)
		r.True(err != nil)       // error should be non-nil
		r.True(newWriter == buf) // writer returned should be same as the original one
	})

	t.Run("rotator with closed file", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)
		filePath := NewFile(t, dir, "rotator-adapter-*")
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0644)
		r.NoErr(err) // should not be any error

		err = file.Close()
		r.NoErr(err) // should not be any error

		rotatorAdapter := RotatorAdapter{FileRotator: faultyRotator(errRotator)}

		newWriter, err := rotatorAdapter.Rotate(file)
		r.True(err != nil)        // should be non-nil
		r.True(newWriter == file) // writer returned should be same as the original one
	})

	t.Run("rotator with file writer with faulty rotator", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)
		filePath := NewFile(t, dir, "rotator-adapter-*")
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0644)
		r.NoErr(err) // should not be any error

		rotatorAdapter := RotatorAdapter{FileRotator: faultyRotator(errRotator)}

		newWriter, err := rotatorAdapter.Rotate(file)
		r.True(errors.Is(err, errRotator)) // error should wrap underlying FileRotator error
		r.True(newWriter == file)          // writer returned should be same as the original one

		err = file.Close()
		r.True(err != nil) // file should already be closed
	})

	t.Run("rotator with file writer with correct rotator", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)
		filePath := NewFile(t, dir, "rotator-adapter-*")
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0644)
		r.NoErr(err) // should not be any error

		newPath := fmt.Sprintf("%s-next", filePath)
		rotatorAdapter := RotatorAdapter{FileRotator: fixedRotator(newPath)}

		newWriter, err := rotatorAdapter.Rotate(file)
		r.NoErr(err) // should not be any error

		err = file.Close()
		r.True(err != nil) // file should already be closed

		newFile, ok := newWriter.(*os.File)
		r.True(ok) // new writer should be reference to os.File
		t.Cleanup(func() {
			err := newFile.Close()
			r.NoErr(err) // should not be any error

			err = os.Remove(newFile.Name())
			r.NoErr(err) // should not be any error
		})

		r.True(newFile.Name() == newPath) // new file name should be same as given by FileRotator.
	})
}
