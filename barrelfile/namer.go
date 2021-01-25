package barrelfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// TimestampSequenceNamer generates name on the basis of mtime timestamp of the
// file formatted using the provided format, and sequenced in the directory,
// with index starting from 0.
type TimestampSequenceNamer struct {
	// Format used to format the mtime of the file.
	TimestampFormat string
}

// Name generates a new timestamp and index suffixed name for file at current
// path.
//
//     current = application.log, out = application_2021-01-02.0.log
//
// If the directory of the file already has file with same name and suffixed
// timestamp with indexes, then it returns the name with next available index.
//
//     dir
//      |- application_2021-01-02.0.log
//      |- application_2021-01-02.1.log
//      |- application_2021-01-02.2.log
//
//     current = application.log, out = application_2021-01-02.3.log
//
// If there are any, errors while reading the directories, then non-nil error
// is returned.
func (n TimestampSequenceNamer) Name(current string) (string, error) {
	stat, err := os.Stat(current)
	if err != nil && os.IsNotExist(err) {
		return current, fmt.Errorf("no such file: %w", err)
	}
	if err != nil {
		return current, fmt.Errorf("os stat: %w", err)
	}
	if stat.IsDir() {
		return current, fmt.Errorf("path is of a directory not a file")
	}

	dir := filepath.Dir(current)
	base := filepath.Base(current)
	fullExt := filepathFullExt(base)
	baseName := strings.Replace(base, fullExt, "", 1)

	tsFormat := n.TimestampFormat
	if len(tsFormat) == 0 {
		tsFormat = "2006-01-02"
	}

	ts := stat.ModTime().Format(tsFormat)

	newName := fmt.Sprintf("%s_%s%s", baseName, ts, fullExt)
	newPath := filepath.Join(dir, newName)

	maxIdx, err := n.maxIdxInDir(newPath)
	if err != nil {
		return current, fmt.Errorf("get max sequence in directory: %w", err)
	}

	seqNewPath := n.indexPath(newPath, maxIdx+1)

	return seqNewPath, nil
}

func (n TimestampSequenceNamer) maxIdxInDir(path string) (int, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	fullExt := filepathFullExt(base)
	baseName := strings.Replace(base, fullExt, "", 1)

	re, err := regexp.Compile(fmt.Sprintf(`%s\.(?P<index>\d+)%s`, baseName, fullExt))
	if err != nil {
		return 0, fmt.Errorf("regexp compile: %w", err)
	}

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("ioutil read dir: %w", err)
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

func (n TimestampSequenceNamer) indexPath(path string, index int) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	fullExt := filepathFullExt(base)
	baseName := strings.Replace(base, fullExt, "", 1)
	newBase := fmt.Sprintf("%s.%d%s", baseName, index, fullExt)
	return filepath.Join(dir, newBase)
}

func filepathFullExt(path string) string {
	firstDotIdx := -1
	for i := len(path) - 1; i >= 0 && path[i] != filepath.Separator; i-- {
		if path[i] == '.' {
			firstDotIdx = i
		}
	}
	if firstDotIdx != -1 {
		return path[firstDotIdx:]
	}
	return ""
}
