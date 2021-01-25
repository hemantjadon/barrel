package barrelfile

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Trigger tells whether the file at the given path should be rotated before
// writing the given bytes to it. If it returns true rotation is performed
// otherwise not.
//
// If there is any error while determining the trigger then non-nil error is
// returned.
type Trigger interface {
	Trigger(path string, p []byte) (bool, error)
}

// SizeBasedTrigger describes a trigger which works on size of the file.
type SizeBasedTrigger struct {
	// Max size of file.
	Size int64
}

var _ Trigger = (*SizeBasedTrigger)(nil)

// Trigger stats the file at the given path, and returns true if size of
// file plus size of bytes to be written exceeds the max size provided,
// otherwise it returns false.
//
// If there is any error while checking file stat, or if the given path is not
// a path to a file then non-nil error is returned.
func (t SizeBasedTrigger) Trigger(path string, p []byte) (bool, error) {
	writeSize := int64(len(p))
	if writeSize > t.Size {
		return false, fmt.Errorf("write size greater than max file size")
	}
	stat, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false, fmt.Errorf("no such file: %w", err)
	}
	if err != nil {
		return false, fmt.Errorf("os stat: %w", err)
	}
	if stat.IsDir() {
		return false, fmt.Errorf("path is of a directory not a file")
	}

	fileSize := stat.Size()
	if fileSize+writeSize < t.Size {
		return false, nil
	}
	return true, nil
}

// CronBasedTrigger describes a trigger which works on mod time of the file.
type CronBasedTrigger struct {
	// CronExpression describes the rotation schedule.
	CronExpression string

	// NowFunc to wrap stdlib time.Now for testing.
	NowFunc func() time.Time

	schedule cron.Schedule
	rotateAt time.Time
	once     sync.Once
}

var _ Trigger = (*CronBasedTrigger)(nil)

// Trigger  stats the file at the given path, returns true when current time
// (as given by the NowFunc) exceeds the time of next schedule, as described by
// the provided CronExpression, otherwise it returns false.
//
// If there is any error while checking file stat, or if the given path is not
// a path to a file, or the given cron expression is invalid then non-nil error
// is returned.
func (t *CronBasedTrigger) Trigger(path string, _ []byte) (bool, error) {
	var nowTime time.Time
	if t.NowFunc != nil {
		nowTime = t.NowFunc()
	} else {
		nowTime = time.Now()
	}

	var err error
	t.once.Do(func() {
		stat, ierr := os.Stat(path)
		if ierr != nil && os.IsNotExist(ierr) {
			err = fmt.Errorf("no such file: %w", ierr)
			return
		}
		if ierr != nil {
			err = fmt.Errorf("os stat: %w", ierr)
			return
		}
		if stat.IsDir() {
			err = fmt.Errorf("path is of a directory not a file")
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
