package barrel

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// FileTrigger tells whether the given file should be rotated or not.
// If it returns true then rotation is preformed, otherwise not.
// If there are any errors while determining the trigger then a non-nil error
// is returned.
type FileTrigger interface {
	Trigger(file *os.File, p []byte) (bool, error)
}

// AdaptFileTrigger converts a FileTrigger into a generic Trigger, which can be
// used with RollingWriter.
//
// The Trigger returned will return a non-nil error if the writer provided is
// not a reference of os.File.
func AdaptFileTrigger(ft FileTrigger) Trigger {
	return fileTriggerAdapter(ft.Trigger)
}

// Rotator changes the file. If there is an error while rotating then a non-nil
// error is returned.
type FileRotator interface {
	Rotate(file *os.File) (*os.File, error)
}

// AdaptFileRotator converts a FileRotator into a generic Rotator, which can be
// used with RollingWriter.
//
// The Rotator returned will return a non-nil error if the writer provided is
// not a reference of os.File.
func AdaptFileRotator(fr FileRotator) Rotator {
	return fileRotatorAdapter(fr.Rotate)
}

// Clock wraps stdlib time functions, for alternate timing functions.
type Clock interface {
	Now() time.Time
}

// FileSizeBasedTrigger describes a trigger which works on size of the file.
type FileSizeBasedTrigger struct {
	// Max size of file.
	Size int64
}

// Trigger returns true if size of file plus size of bytes to be written exceeds
// the max size provided, otherwise it returns false.
//
// If size of bytes to be written is greater than max size provided then a
// non-nil error is returned.
func (t FileSizeBasedTrigger) Trigger(file *os.File, p []byte) (bool, error) {
	stat, err := os.Stat(file.Name())
	if err != nil {
		return false, fmt.Errorf("os stat: %w", err)
	}
	writeSize := int64(len(p))
	if writeSize > t.Size {
		return false, fmt.Errorf("write size greater than max file size")
	}
	fileSize := stat.Size()
	if fileSize+writeSize > t.Size {
		return true, nil
	}
	return false, nil
}

// FileTimeBasedTrigger describes a trigger which works on mod time of the file.
type FileTimeBasedTrigger struct {
	// Describes the rotation schedule.
	CronExpression string

	// Clock to wrap stdblib timing functions for testing.
	Clock Clock

	schedule cron.Schedule
	rotateAt time.Time
	once     sync.Once
}

// Trigger returns true when current time (as given by the Clock) exceeds the
// time of next schedule, as described by the provided CronExpression, otherwise
// it returns false.
func (t *FileTimeBasedTrigger) Trigger(file *os.File, _ []byte) (bool, error) {
	var nowTime time.Time
	if t.Clock != nil {
		nowTime = t.Clock.Now()
	} else {
		nowTime = time.Now()
	}

	var err error
	t.once.Do(func() {
		stat, ierr := os.Stat(file.Name())
		if ierr != nil {
			err = fmt.Errorf("os stat: %w", ierr)
			return
		}
		schedule, ierr := cron.ParseStandard(t.CronExpression)
		if ierr != nil {
			err = fmt.Errorf("cron parse standard: %v", ierr)
			return
		}
		t.schedule = schedule
		if t.rotateAt.IsZero() {
			t.rotateAt = t.schedule.Next(stat.ModTime())
		}
	})
	if err != nil {
		return false, err
	}

	var rotate bool
	for nowTime.After(t.rotateAt) {
		rotate = true
		t.rotateAt = t.schedule.Next(t.rotateAt)
	}
	return rotate, nil
}

// FileTransformer takes a file path, transforms the file and returns the path
// of resulting file. If any error occurs non-nil error is returned.
type FileTransformer interface {
	Transform(path string) (string, error)
}

// FileTransformingRotator describes a Rotator, which works by rotating files.
type FileTransformingRotator struct {
	// Flag to be used to Open the the new file, os.O_CREATE is added
	// automatically.
	OpenFlag int

	// Transformers to to transform the file.
	Transformers []FileTransformer
}

// Rotate closes the current file, runs all the provided transformers in order,
// and returns a reference to a new file, which is created with same name and
// permissions (os.FileMode) as the original file.
//
// It uses OpenFlag while opening the file.
//
// If there is any error in any of the transformers, then no further
// transformers are applied and non-nil error is returned.
func (r FileTransformingRotator) Rotate(file *os.File) (*os.File, error) {
	name := file.Name()
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("orignal file stat: %w", err)
	}
	path := name
	for idx, transformer := range r.Transformers {
		newPath, err := transformer.Transform(path)
		if err != nil {
			return nil, fmt.Errorf("transformer[%d]: %w", idx, err)
		}
		path = newPath
	}
	if err := file.Sync(); err != nil {
		return nil, fmt.Errorf("file sync: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("file close: %w", err)
	}
	newFile, err := os.OpenFile(name, os.O_CREATE|r.OpenFlag, stat.Mode())
	if err != nil {
		return nil, fmt.Errorf("os open file: %w", err)
	}
	return newFile, nil
}

// FileGzipTransformer converts file to gzip format.
type FileGzipTransformer struct {
	// Describes gzip compression level.
	GzipLevel int
}

// Transform transforms the file at the provided path to a gzip file in the same
// directory.
// It add .gz extension to the file. The original file is not retained.
// The mod time and access time of the new gzip file is same as mod time of the
//original file.
func (t FileGzipTransformer) Transform(path string) (string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("gzip: os stat src file: %w", err)
	}
	rdFile, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("gzip: os open src file: %w", err)
	}
	defer func() { _ = rdFile.Close() }()

	gzFileName := fmt.Sprintf("%s.gz", path)
	gzFile, err := os.OpenFile(gzFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, stat.Mode())
	if err != nil {
		return "", fmt.Errorf("gzip: os open gz file: %w", err)
	}
	gzw, err := gzip.NewWriterLevel(gzFile, t.GzipLevel)
	if err != nil {
		_ = gzFile.Close()
		_ = os.Remove(gzFileName)
		return "", fmt.Errorf("gzip: create new writer level: %w", err)
	}
	if _, err := io.Copy(gzw, rdFile); err != nil {
		_ = gzFile.Close()
		_ = os.Remove(gzFileName)
		return "", fmt.Errorf("gzip: io copy src to gz file: %w", err)
	}
	if err := gzw.Close(); err != nil {
		_ = gzFile.Close()
		_ = os.Remove(gzFileName)
		return "", fmt.Errorf("gzip: close gz writer: %w", err)
	}
	if err := gzFile.Sync(); err != nil {
		_ = gzFile.Close()
		return "", fmt.Errorf("gzip: gz file sync: %w", err)
	}
	if err := gzFile.Close(); err != nil {
		return "", fmt.Errorf("gzip: gz file close: %w", err)
	}
	if err := os.Remove(path); err != nil {
		return "", fmt.Errorf("gzip: os remove src file: %w", err)
	}
	if err := os.Chtimes(gzFileName, stat.ModTime(), stat.ModTime()); err != nil {
		return "", fmt.Errorf("gzip: change gz file mtime: %w", err)
	}
	return gzFileName, nil
}

// NameGenerator generates a new name with the provided current name.
type NameGenerator interface {
	Name(current string) (string, error)
}

// FileRenameTransformer renames a file.
type FileRenameTransformer struct {
	// Generator to be used for generating new name of the file.
	NameGenerator NameGenerator

	// Force physical move of the file.
	ForceMove bool
}

// Transform transforms the file at provided path, to a path generated by the
// NameGenerator. It returns the generated path.
func (t FileRenameTransformer) Transform(path string) (string, error) {
	newPath, err := t.NameGenerator.Name(path)
	if err != nil {
		return "", fmt.Errorf("rename: get new name: %w", err)
	}
	if err := fileMove(path, newPath, t.ForceMove); err != nil {
		return "", fmt.Errorf("rename: %w", err)
	}
	return newPath, nil
}

// FileTimestampNameGenerator is a NameGenerator which generates timestamp
// suffixed name for the file with current name.
type FileTimestampNameGenerator struct {
	// Format of timestamp to use.
	Format string
}

// Name generates a new timestamp suffixed name for file at current path.
//
//     current = application.log, out = application_2020-01-02.log
//
// If the directory of the file already has file with same name and suffixed
// timestamp with indexes, then it returns the file with next available index.
//
//     dir
//      |- application_2020-01-02.0.log
//      |- application_2020-01-02.1.log
//      |- application_2020-01-02.2.log
//
//     current = application.log, out = application_2020-01-02.3.log
//
// If the directory of file contains file with same name and and suffixed
// timestamp without indexes, then it moves the file to correct index, and
// returns the file with next index.
//
//     dir
//      |- application_2020-01-02.log
//
//     current = application.log, out application_2020-01-02.1.log
//
//     Afterwards,
//
//     dir
//      |- application_2020-01-02.0.log
//
// If there are any, errors while reading the directories, then non-nil error
// is returned.
func (g FileTimestampNameGenerator) Name(current string) (string, error) {
	stat, err := os.Stat(current)
	if err != nil {
		return "", fmt.Errorf("name gen: os stat current file: %w", err)
	}
	dir := filepath.Dir(current)
	base := filepath.Base(current)
	baseName := filepathStripFullExt(base)
	baseExt := filepathFullExt(base)
	format := g.Format
	if len(format) == 0 {
		format = "2006-01-02"
	}
	ts := stat.ModTime().Format(format)
	newBase := fmt.Sprintf("%s_%s%s", baseName, ts, baseExt)
	newPath := filepath.Join(dir, newBase)
	maxIdx, err := g.maxIndexInDir(newPath)
	if err != nil {
		return "", err
	}
	if !g.isPresent(newPath) {
		if maxIdx != -1 {
			return g.indexPath(newPath, maxIdx+1), nil
		}
		return newPath, nil
	}
	if err := fileMove(newPath, g.indexPath(newPath, maxIdx+1), false); err != nil {
		return "", fmt.Errorf("name gen: %w", err)
	}
	return g.indexPath(newPath, maxIdx+2), nil
}

func (g FileTimestampNameGenerator) isPresent(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (g FileTimestampNameGenerator) indexPath(path string, index int) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	baseName := filepathStripFullExt(base)
	baseExt := filepathFullExt(base)
	newBase := fmt.Sprintf("%s.%d%s", baseName, index, baseExt)
	return filepath.Join(dir, newBase)
}

func (g FileTimestampNameGenerator) maxIndexInDir(path string) (int, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	baseName := filepathStripFullExt(base)
	baseExt := filepathFullExt(base)
	re, err := regexp.Compile(fmt.Sprintf(`%s\.(?P<index>\d+)%s`, baseName, baseExt))
	if err != nil {
		return 0, fmt.Errorf("name gen: regexp compile: %w", err)
	}
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("name gen: ioutil read dir: %w", err)
	}
	maxIdx := -1
	for _, fileInfo := range fileInfos {
		match := re.FindStringSubmatch(fileInfo.Name())
		if match == nil {
			continue
		}
		idxStr := match[1]
		idx, _ := strconv.Atoi(idxStr)
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	return maxIdx, nil
}

func filepathFullExt(path string) string {
	var ext []rune
	var foundDot bool
	for idx, ch := range path {
		if (idx != 0 && ch == '.') || foundDot {
			ext = append(ext, ch)
			foundDot = true
		}
	}
	return string(ext)
}

func filepathStripFullExt(path string) string {
	var name []rune
	for idx, ch := range path {
		if idx != 0 && ch == '.' {
			break
		}
		name = append(name, ch)
	}
	return string(name)
}

func fileMove(src, dst string, force bool) error {
	stat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("move: src file stat: %w", err)
	}
	if force {
		return physicalMove(src, dst, stat)
	}
	if err := inodeRename(src, dst); err != nil {
		return physicalMove(src, dst, stat)
	}
	return nil
}

func inodeRename(old, new string) error {
	if err := os.Rename(old, new); err != nil {
		return fmt.Errorf("move: os rename: %w", err)
	}
	return nil
}

func physicalMove(src, dst string, stat os.FileInfo) error {
	rdFile, err := os.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("move: os open src file: %w", err)
	}
	defer func() { _ = rdFile.Close() }()
	wrFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, stat.Mode())
	if err != nil {
		return fmt.Errorf("move: os open dst file: %w", err)
	}
	if _, err := io.Copy(wrFile, rdFile); err != nil {
		_ = wrFile.Close()
		return fmt.Errorf("move: io copy: %w", err)
	}
	if err := wrFile.Sync(); err != nil {
		_ = wrFile.Close()
		return fmt.Errorf("move: dst file sync: %w", err)
	}
	if err := wrFile.Close(); err != nil {
		return fmt.Errorf("move: dst file close: %w", err)
	}
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("move: src file remove: %w", err)
	}
	if err := os.Chtimes(dst, stat.ModTime(), stat.ModTime()); err != nil {
		return fmt.Errorf("move: change dst file mtime: %w", err)
	}
	return nil
}

type fileTriggerAdapter func(file *os.File, p []byte) (bool, error)

func (fta fileTriggerAdapter) Trigger(writer io.Writer, p []byte) (bool, error) {
	file, ok := writer.(*os.File)
	if !ok {
		return false, fmt.Errorf("writer is not reference to os.File")
	}
	return fta(file, p)
}

type fileRotatorAdapter func(file *os.File) (*os.File, error)

func (fra fileRotatorAdapter) Rotate(writer io.Writer) (io.Writer, error) {
	file, ok := writer.(*os.File)
	if !ok {
		return nil, fmt.Errorf("writer is not reference to os.File")
	}
	return fra(file)
}
