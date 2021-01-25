package barrelfile

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matryer/is"
)

var (
	testTime = time.Date(2021, 1, 1, 6, 15, 0, 0, time.UTC)
	testText = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur volutpat eleifend velit, quis elementum metus maximus non. Aliquam rutrum placerat sapien, vitae egestas ante egestas sit amet. Morbi finibus ex eget leo cursus, a porta lorem aliquam. Sed sollicitudin eu purus sit amet lobortis. Nullam nec nulla hendrerit, aliquet magna dictum, facilisis nibh. Duis lacinia nec velit ut consequat. Maecenas vel lobortis erat. Nunc sollicitudin consectetur nulla id hendrerit. Nunc rhoncus eros non efficitur ullamcorper. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Aliquam pulvinar accumsan cursus. Curabitur id orci commodo, viverra elit ut, posuere leo. Cras at dolor arcu. Maecenas efficitur enim commodo turpis efficitur dapibus.`
)

func TestRenameTransformer_Transform(t *testing.T) {
	t.Parallel()

	t.Run("faulty namer", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		transformer := RenameTransformer{Namer: faultyNamer(errNamer)}

		// Hopefully this is random
		randomPath := filepath.Join(dir, "g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn")

		path, err := transformer.Transform(randomPath)
		r.True(errors.Is(err, errNamer)) // error should wrap underlying Namer error
		r.True(path == randomPath)       // path returned should be same as given path
	})

	t.Run("no file at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		newPath := filepath.Join(dir, "random")
		transformer := RenameTransformer{Namer: fixedNamer(newPath)}

		// Hopefully this is random
		randomPath := filepath.Join(dir, "g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn")

		path, err := transformer.Transform(randomPath)
		r.True(err != nil)         // should be non nil
		r.True(path == randomPath) // path returned should be same as given path
	})

	t.Run("directory at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		newPath := filepath.Join(dir, "random")
		transformer := RenameTransformer{Namer: fixedNamer(newPath)}

		path, err := transformer.Transform(dir)
		r.True(err != nil)  // should be non nil
		r.True(path == dir) // path returned should be same as given path
	})

	t.Run("transform file at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file, err := ioutil.TempFile(dir, "rename-transformer-*")
		r.NoErr(err) // should not be any error

		// file has some content
		content := []byte(testText)
		_, err = file.Write(content)
		r.NoErr(err) // should not be any error

		err = file.Close()
		r.NoErr(err) // should not be any error

		originalModTime := testTime
		err = os.Chtimes(file.Name(), originalModTime, originalModTime)
		r.NoErr(err) // should not be any error

		newPath := fmt.Sprintf("%s.out", file.Name())
		t.Cleanup(func() {
			err := os.Remove(newPath)
			r.NoErr(err) // should not be any error
		})

		transformer := RenameTransformer{Namer: fixedNamer(newPath)}

		path, err := transformer.Transform(file.Name())
		r.NoErr(err)            // should not be any error
		r.True(path == newPath) // path should be as returned by given Namer

		_, err = os.Open(file.Name())
		r.True(os.IsNotExist(err)) // original file should not exist after transform

		newFileContent, err := ioutil.ReadFile(newPath)
		r.NoErr(err)                                 // should not be any error
		r.True(bytes.Equal(newFileContent, content)) // content of new file should be same as the original file

		stat, err := os.Stat(newPath)
		r.NoErr(err)                                  // should not be any error
		r.True(stat.ModTime().Equal(originalModTime)) // new file's mod time should be same as original one's
	})

	t.Run("transform file at path with force move", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file, err := ioutil.TempFile(dir, "rename-transformer-*")
		r.NoErr(err) // should not be any error

		// file has some content
		content := []byte(testText)
		_, err = file.Write(content)
		r.NoErr(err) // should not be any error

		err = file.Close()
		r.NoErr(err) // should not be any error

		originalModTime := testTime
		err = os.Chtimes(file.Name(), originalModTime, originalModTime)
		r.NoErr(err) // should not be any error

		newPath := fmt.Sprintf("%s.out", file.Name())
		t.Cleanup(func() {
			err := os.Remove(newPath)
			r.NoErr(err) // should not be any error
		})

		transformer := RenameTransformer{Namer: fixedNamer(newPath), ForceMove: true}

		path, err := transformer.Transform(file.Name())
		r.NoErr(err)            // should not be any error
		r.True(path == newPath) // path should be as returned by given Namer

		_, err = os.Open(file.Name())
		r.True(os.IsNotExist(err)) // original file should not exist after transform

		newFileContent, err := ioutil.ReadFile(newPath)
		r.NoErr(err)                                 // should not be any error
		r.True(bytes.Equal(newFileContent, content)) // content of new file should be same as the original file

		stat, err := os.Stat(newPath)
		r.NoErr(err)                                  // should not be any error
		r.True(stat.ModTime().Equal(originalModTime)) // new file's mod time should be same as original one's
	})
}

func TestGzipTransformer_Transform(t *testing.T) {
	t.Parallel()

	t.Run("no file at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		transformer := GzipTransformer{GzipLevel: gzip.DefaultCompression}

		// Hopefully this is random
		randomPath := filepath.Join(dir, "g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn")

		path, err := transformer.Transform(randomPath)
		r.True(err != nil)         // should be non nil
		r.True(path == randomPath) // path returned should be same as given path
	})

	t.Run("directory at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		transformer := GzipTransformer{GzipLevel: gzip.DefaultCompression}

		path, err := transformer.Transform(dir)
		r.True(err != nil)  // should be non nil
		r.True(path == dir) // path returned should be same as given path
	})

	t.Run("invalid gzip level", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		transformer := GzipTransformer{GzipLevel: -100}

		file, err := ioutil.TempFile(dir, "rename-transformer-*")
		r.NoErr(err) // should not be any error
		t.Cleanup(func() {
			err := os.Remove(file.Name())
			r.NoErr(err) // should not be any error
		})

		// file has some content
		content := []byte(testText)
		_, err = file.Write(content)
		r.NoErr(err) // should not be any error

		err = file.Close()
		r.NoErr(err) // should not be any error

		originalModTime := testTime
		err = os.Chtimes(file.Name(), originalModTime, originalModTime)
		r.NoErr(err) // should not be any error

		path, err := transformer.Transform(file.Name())
		r.True(err != nil)          // should be non nil
		r.True(path == file.Name()) // path returned should be same as given path
	})

	t.Run("transform file at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file, err := ioutil.TempFile(dir, "rename-transformer-*")
		r.NoErr(err) // should not be any error

		// file has some content
		content := []byte(testText)
		_, err = file.Write(content)
		r.NoErr(err) // should not be any error

		err = file.Close()
		r.NoErr(err) // should not be any error

		originalModTime := testTime
		err = os.Chtimes(file.Name(), originalModTime, originalModTime)
		r.NoErr(err) // should not be any error

		newPath := fmt.Sprintf("%s.gz", file.Name())
		t.Cleanup(func() {
			err := os.Remove(newPath)
			r.NoErr(err) // should not be any error
		})

		gzipLevel := gzip.DefaultCompression
		transformer := GzipTransformer{GzipLevel: gzipLevel}

		path, err := transformer.Transform(file.Name())
		r.NoErr(err)            // should not be any error
		r.True(path == newPath) // path should be as returned by given Namer

		_, err = os.Open(file.Name())
		r.True(os.IsNotExist(err)) // original file should not exist after transform

		newFileBytes, err := ioutil.ReadFile(newPath)
		r.NoErr(err) // should not be any error

		gzr, err := gzip.NewReader(bytes.NewBuffer(newFileBytes))
		r.NoErr(err) // should not be any error

		var newFileContentBuf bytes.Buffer
		_, err = io.Copy(&newFileContentBuf, gzr)
		r.NoErr(err)

		newFileContent := newFileContentBuf.Bytes()

		r.True(bytes.Equal(newFileContent, content)) // decompressed content of new file should be same as the original file

		stat, err := os.Stat(newPath)
		r.NoErr(err)                                  // should not be any error
		r.True(stat.ModTime().Equal(originalModTime)) // new file's mod time should be same as original one's
	})
}
