package barrelfile

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

// Namer generates the new name of the file given the current name.
type Namer interface {
	Name(current string) (string, error)
}

// RenameTransformer renames/moves the file.
type RenameTransformer struct {
	// Namer generates the new path of the file.
	Namer Namer

	// ForceMove forces the physical file move, even if inode rename is
	// possible. Use carefully.
	ForceMove bool
}

// Transform renames/moves the file at the given path, to the destination as
// given by the provided Namer, for the given path. It tries to first perform
// less expensive inode rename, to just rename the file's inode, but if inode
// rename is unsuccessful (if file is being moved to a different disk), then
// file contents are physically moved by copying the data from original path to
// destination path. Note that this physical move of file is expensive
// operation. If ForceMove is set true, then inode rename is not even tried,
// and more expensive operation of physically moving the file is performed.
//
// If there are any error while generating the new path or actually performing
// the rename/move, then non-nil error is returned.
func (t RenameTransformer) Transform(path string) (string, error) {
	newPath, err := t.Namer.Name(path)
	if err != nil {
		return path, fmt.Errorf("namer new name: %w", err)
	}
	if err := fileMove(path, newPath, t.ForceMove); err != nil {
		return path, fmt.Errorf("rename: %w", err)
	}
	return newPath, nil
}

// GzipTransformer compresses and converts file to gzip format.
type GzipTransformer struct {
	// GzipLevel defines the gzip level used for compression. If unset (ie. 0),
	// it will correspond to gzip.NoCompression.
	GzipLevel int
}

// Transform compresses the file at the given path using gzip compression at
// the provided level. The resulting compressed file is created in the same
// directory as the original file, but ".gz" extension is added to the file
// name.
//
// If there are any error while compressing the file at given path then non-nil
// error is returned.
func (t GzipTransformer) Transform(path string) (string, error) {
	gzPath := fmt.Sprintf("%s.gz", path)
	if err := fileGzip(path, gzPath, t.GzipLevel); err != nil {
		return path, fmt.Errorf("")
	}
	return gzPath, nil
}

func fileGzip(src, dst string, level int) error {
	stat, err := os.Stat(src)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("no such file: %w", err)
	}
	if err != nil {
		return fmt.Errorf("os stat src file: %w", err)
	}
	if stat.IsDir() {
		return fmt.Errorf("src is path of a directory not a file")
	}
	srcFile, err := os.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("os open src file: %w", err)
	}

	gzFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stat.Mode())
	if err != nil {
		_ = srcFile.Close()
		return fmt.Errorf("os open gzip file: %w", err)
	}
	gzw, err := gzip.NewWriterLevel(gzFile, level)
	if err != nil {
		_ = gzFile.Close()
		_ = os.Remove(dst)
		_ = srcFile.Close()
		return fmt.Errorf("gzip new writer level: %w", err)
	}
	if _, err := io.Copy(gzw, srcFile); err != nil {
		_ = gzw.Close()
		_ = gzFile.Close()
		_ = os.Remove(dst)
		_ = srcFile.Close()
		return fmt.Errorf("copy and gzip compress: %w", err)
	}
	if err := gzw.Close(); err != nil {
		_ = gzFile.Close()
		_ = os.Remove(dst)
		_ = srcFile.Close()
		return fmt.Errorf("close gzip writer: %w", err)
	}
	if err := gzFile.Sync(); err != nil {
		_ = gzFile.Close()
		_ = srcFile.Close()
		return fmt.Errorf("sync gzip file: %w", err)
	}
	if err := gzFile.Close(); err != nil {
		_ = srcFile.Close()
		return fmt.Errorf("close gzip file: %w", err)
	}
	if err := srcFile.Close(); err != nil {
		return fmt.Errorf("close src file: %w", err)
	}
	if err := os.Chtimes(dst, stat.ModTime(), stat.ModTime()); err != nil {
		return fmt.Errorf("os change times gzip file: %w", err)
	}
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("os remove src file: %w", err)
	}
	return nil
}

func fileMove(src, dst string, force bool) error {
	stat, err := os.Stat(src)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("no such file: %w", err)
	}
	if err != nil {
		return fmt.Errorf("os stat src file: %w", err)
	}
	if stat.IsDir() {
		return fmt.Errorf("src is path of a directory not a file")
	}
	if !force {
		if err := inodeRename(src, dst); err == nil {
			return nil
		}
	}
	if err := physicalMove(src, dst); err != nil {
		return fmt.Errorf("move: %w", err)
	}
	return nil
}

func inodeRename(src, dst string) error {
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("os rename inode: %w", err)
	}
	return nil
}

func physicalMove(src, dst string) error {
	stat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("os stat src file: %w", err)
	}
	srcFile, err := os.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("os open src file: %w", err)
	}
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, stat.Mode())
	if err != nil {
		_ = srcFile.Close()
		return fmt.Errorf("os open dst file: %w", err)
	}
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		_ = srcFile.Close()
		_ = dstFile.Close()
		return fmt.Errorf("io copy src to dst file: %w", err)
	}
	if err := dstFile.Sync(); err != nil {
		_ = srcFile.Close()
		_ = dstFile.Close()
		return fmt.Errorf("sync dst file: %w", err)
	}
	if err := dstFile.Close(); err != nil {
		_ = srcFile.Close()
		return fmt.Errorf("close dst file: %w", err)
	}
	if err := srcFile.Close(); err != nil {
		return fmt.Errorf("close src file: %w", err)
	}
	if err := os.Chtimes(dst, stat.ModTime(), stat.ModTime()); err != nil {
		return fmt.Errorf("os change times dst file: %w", err)
	}
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("os remove src file: %w", err)
	}
	return nil
}
