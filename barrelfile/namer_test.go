package barrelfile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/matryer/is"
)

func TestTimestampSequenceNamer_Name(t *testing.T) {
	t.Parallel()

	t.Run("no file at current path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		namer := TimestampSequenceNamer{}

		// Hopefully this is random
		randomPath := filepath.Join(dir, "g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn")

		name, err := namer.Name(randomPath)
		r.True(err != nil)         // should be non nil
		r.True(name == randomPath) // path returned should be same as given path
	})

	t.Run("directory at current path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		namer := TimestampSequenceNamer{}

		name, err := namer.Name(dir)
		r.True(err != nil)  // should be non nil
		r.True(name == dir) // path returned should be same as given path
	})

	t.Run("directory with no other files with current timestamp", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file, err := os.Create(filepath.Join(dir, "application.log"))
		r.NoErr(err) // should not be any error
		t.Cleanup(func() {
			err := os.Remove(file.Name())
			r.NoErr(err) // should not be any error
		})

		err = file.Close()
		r.NoErr(err) // should not be any error

		err = os.Chtimes(file.Name(), testTime, testTime)
		r.NoErr(err) // should not be any error

		namer := TimestampSequenceNamer{}

		wantName := filepath.Join(dir, fmt.Sprintf("application_%s.0%s", testTime.Format("2006-01-02"), ".log"))

		newName, err := namer.Name(file.Name())
		r.NoErr(err) // should not be any error
		r.True(newName == wantName)
	})

	t.Run("directory with other files with current timestamp", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file1, err := os.Create(filepath.Join(dir, fmt.Sprintf("application_%s.0%s", testTime.Format("2006-01-02"), ".log")))
		r.NoErr(err) // should not be any error
		t.Cleanup(func() {
			err := os.Remove(file1.Name())
			r.NoErr(err) // should not be any error
		})

		file2, err := os.Create(filepath.Join(dir, fmt.Sprintf("application_%s.1%s", testTime.Format("2006-01-02"), ".log")))
		r.NoErr(err) // should not be any error
		t.Cleanup(func() {
			err := os.Remove(file2.Name())
			r.NoErr(err) // should not be any error
		})

		file3, err := os.Create(filepath.Join(dir, fmt.Sprintf("application_%s.2%s", testTime.Format("2006-01-02"), ".log")))
		r.NoErr(err) // should not be any error
		t.Cleanup(func() {
			err := os.Remove(file3.Name())
			r.NoErr(err) // should not be any error
		})

		file4, err := os.Create(filepath.Join(dir, fmt.Sprintf("application_%s.3%s", testTime.Format("2006-01-02"), ".log")))
		r.NoErr(err) // should not be any error
		t.Cleanup(func() {
			err := os.Remove(file4.Name())
			r.NoErr(err) // should not be any error
		})

		file, err := os.Create(filepath.Join(dir, "application.log"))
		r.NoErr(err) // should not be any error
		t.Cleanup(func() {
			err := os.Remove(file.Name())
			r.NoErr(err) // should not be any error
		})

		err = file.Close()
		r.NoErr(err) // should not be any error

		err = os.Chtimes(file.Name(), testTime, testTime)
		r.NoErr(err) // should not be any error

		namer := TimestampSequenceNamer{}

		wantName := filepath.Join(dir, fmt.Sprintf("application_%s.4%s", testTime.Format("2006-01-02"), ".log"))

		newName, err := namer.Name(file.Name())
		r.NoErr(err) // should not be any error
		r.True(newName == wantName)
	})
}
