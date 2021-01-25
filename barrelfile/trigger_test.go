package barrelfile

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestSizeBasedTrigger_Trigger(t *testing.T) {
	t.Parallel()

	t.Run("no file at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		trigger := SizeBasedTrigger{Size: 100 * 1e3 * 1e3}

		// Hopefully this is random
		randomPath := filepath.Join(dir, "g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn")

		v, err := trigger.Trigger(randomPath, nil)
		r.True(err != nil)
		r.True(v == false)
	})

	t.Run("directory at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		trigger := SizeBasedTrigger{Size: 100 * 1e3 * 1e3}

		v, err := trigger.Trigger(dir, nil)
		r.True(err != nil)
		r.True(v == false)
	})

	t.Run("write size greater than max size", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file := NewFile(t, dir, "size-based-trigger-*")

		trigger := SizeBasedTrigger{Size: 5}

		v, err := trigger.Trigger(file, []byte("hello world"))
		r.True(err != nil)
		r.True(v == false)
	})

	t.Run("write size plus file size less max size", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file := NewFile(t, dir, "size-based-trigger-*")

		trigger := SizeBasedTrigger{Size: 20}

		v, err := trigger.Trigger(file, []byte("hello world"))
		r.True(err == nil) // should not be any error
		r.True(v == false) // trigger should return false
	})

	t.Run("write size plus file size greater max size", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file := NewFile(t, dir, "size-based-trigger-*")

		fp, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
		r.NoErr(err) // should not be any error

		_, err = fp.Write([]byte("hello world"))
		r.NoErr(err) // should not be any error

		err = fp.Close()
		r.NoErr(err) // should not be any error

		trigger := SizeBasedTrigger{Size: 20}

		v, err := trigger.Trigger(file, []byte("hello world"))
		r.NoErr(err)      // should not be any error
		r.True(v == true) // trigger should return true
	})
}

func TestCronBasedTrigger_Trigger(t *testing.T) {
	t.Parallel()

	t.Run("no file at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		trigger := CronBasedTrigger{CronExpression: "* * * * *"}

		// Hopefully this is random
		randomPath := filepath.Join(dir, "g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn.g3t4c5x15nx3k0fs3125nl400000gn")

		v, err := trigger.Trigger(randomPath, nil)
		r.True(err != nil)
		r.True(v == false)
	})

	t.Run("directory at path", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		trigger := CronBasedTrigger{CronExpression: "* * * * *"}

		v, err := trigger.Trigger(dir, nil)
		r.True(err != nil)
		r.True(v == false)
	})

	t.Run("invalid cron expression", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file := NewFile(t, dir, "cron-based-trigger-*")

		trigger := CronBasedTrigger{CronExpression: "abcd"}

		v, err := trigger.Trigger(file, nil)
		r.True(err != nil)
		r.True(v == false)
	})

	t.Run("hourly cron and file has recent mtime", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file := NewFile(t, dir, "cron-based-trigger-*")
		err := os.Chtimes(file, time.Now(), time.Now())
		r.NoErr(err) // should not be any error

		trigger := CronBasedTrigger{CronExpression: "0 * * * *"}

		v, err := trigger.Trigger(file, nil)
		r.NoErr(err)       // should not be any error
		r.True(v == false) // trigger should return false
	})

	t.Run("hourly cron and file has old mtime", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		file := NewFile(t, dir, "cron-based-trigger-*")
		err := os.Chtimes(file, time.Now().Add(-5*time.Hour), time.Now().Add(-5*time.Hour))
		r.NoErr(err) // should not be any error

		trigger := CronBasedTrigger{CronExpression: "0 * * * *"}

		v, err := trigger.Trigger(file, nil)
		r.NoErr(err)      // should not be any error
		r.True(v == true) // trigger should return true
	})

	t.Run("hourly cron and time advances forward", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		dir := SetupDir(t)

		clk := clock{}
		clk.Set(testTime)

		file := NewFile(t, dir, "cron-based-trigger-*")
		err := os.Chtimes(file, clk.Now(), clk.Now())
		r.NoErr(err) // should not be any error

		trigger := CronBasedTrigger{CronExpression: "0 * * * *", NowFunc: clk.Now}

		v, err := trigger.Trigger(file, nil)
		r.NoErr(err)       // should not be any error
		r.True(v == false) // trigger should return false

		// time moves forward
		clk.Set(testTime.Add(5 * time.Hour))

		v, err = trigger.Trigger(file, nil)
		r.NoErr(err)      // should not be any error
		r.True(v == true) // trigger should return true
	})
}

type clock struct {
	time time.Time

	mu sync.Mutex
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.time
}

func (c *clock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.time = t
}
