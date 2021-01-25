package barrel

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/matryer/is"
)

//go:generate moq -fmt goimports -out ./barrel_test_mock_test.go . Trigger Rotator

func TestRollingWriter_Write(t *testing.T) {
	t.Parallel()

	t.Run("closed writer", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		writer := RollingWriter{}
		err := writer.Close()
		r.NoErr(err) // should not be any error

		_, err = writer.Write([]byte("hello"))
		r.True(errors.Is(err, ErrClosed)) // error should wrap ErrClosed
	})

	t.Run("trigger errors", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		writer := RollingWriter{Trigger: faultyTrigger(errTrigger)}

		_, err := writer.Write([]byte("hello"))
		r.True(errors.Is(err, errTrigger)) // error should wrap underlying Trigger error
	})

	t.Run("trigger returns false", func(t *testing.T) {
		t.Parallel()

		t.Run("underlying writer errors", func(t *testing.T) {
			t.Parallel()
			r := is.New(t)

			writer := RollingWriter{Writer: faultyWriter{Err: errWrite}, Trigger: fixedTrigger(false)}

			data := []byte("hello")
			_, err := writer.Write(data)
			r.True(errors.Is(err, errWrite)) // error should wrap underlying io.Writer error
		})

		t.Run("underlying writer succeeds", func(t *testing.T) {
			t.Parallel()
			r := is.New(t)

			var buf bytes.Buffer

			writer := RollingWriter{Writer: &buf, Trigger: fixedTrigger(false)}

			data := []byte("hello")
			n, err := writer.Write(data)
			r.NoErr(err) //should not be any error

			r.True(n == len(data))                 // all bytes should be written
			r.True(n == len(buf.Bytes()))          // all bytes should be written to underlying io.Writer
			r.True(bytes.Equal(buf.Bytes(), data)) // bytes written io.Writer should be same as given bytes
		})
	})

	t.Run("trigger returns true", func(t *testing.T) {
		t.Parallel()

		t.Run("rotator errors", func(t *testing.T) {
			t.Parallel()
			r := is.New(t)

			var buf1 bytes.Buffer

			writer := RollingWriter{Writer: &buf1, Trigger: fixedTrigger(true), Rotator: faultyRotator(errRotate)}

			data := []byte("hello")
			n, err := writer.Write(data)
			r.True(errors.Is(err, errRotate)) // error should wrap underlying Rotator error

			r.True(n == 0)                 // no bytes should be written
			r.True(len(buf1.Bytes()) == 0) // original io.Writer should not receive any data
		})

		t.Run("rotator succeeds", func(t *testing.T) {
			t.Parallel()

			t.Run("rotated writer errors", func(t *testing.T) {
				t.Parallel()
				r := is.New(t)

				var buf1 bytes.Buffer

				writer := RollingWriter{Writer: &buf1, Trigger: fixedTrigger(true), Rotator: fixedRotator(faultyWriter{Err: errWrite})}

				data := []byte("hello")
				n, err := writer.Write(data)
				r.True(errors.Is(err, errWrite)) // error should wrap underlying Rotator error

				r.True(n == 0)                 // no bytes should be written
				r.True(len(buf1.Bytes()) == 0) // original io.Writer should not receive any data
			})

			t.Run("rotated writer succeeds", func(t *testing.T) {
				t.Parallel()
				r := is.New(t)

				var buf1, buf2 bytes.Buffer

				writer := RollingWriter{Writer: &buf1, Trigger: fixedTrigger(true), Rotator: fixedRotator(&buf2)}

				data := []byte("hello")
				n, err := writer.Write(data)
				r.NoErr(err) // should not be any error

				r.True(len(buf1.Bytes()) == 0) // original io.Writer should not receive any data

				r.True(n == len(data))                  // all bytes should be written
				r.True(n == len(buf2.Bytes()))          // bytes should be written to rotated io.Writer
				r.True(bytes.Equal(buf2.Bytes(), data)) // bytes written to rotated io.Writer should be same as given bytes
			})
		})
	})
}

func TestRollingWriter_Close(t *testing.T) {
	t.Parallel()

	t.Run("closed writer", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		writer := RollingWriter{}
		err := writer.Close()
		r.NoErr(err) // should not be any error

		err = writer.Close()
		r.True(errors.Is(err, ErrClosed)) // error should wrap ErrClosed
	})

	t.Run("writer not io closer", func(t *testing.T) {
		t.Parallel()
		r := is.New(t)

		var buf bytes.Buffer

		writer := RollingWriter{Writer: &buf}

		err := writer.Close()
		r.NoErr(err) // should not be any error
	})

	t.Run("writer is io closer", func(t *testing.T) {
		t.Parallel()

		t.Run("writer close errors", func(t *testing.T) {
			t.Parallel()
			r := is.New(t)

			var buf bytes.Buffer

			writer := RollingWriter{Writer: writeCloser{Writer: &buf, Closer: faultyCloser{Err: errClose}}}

			err := writer.Close()
			r.True(errors.Is(err, errClose)) // error should wrap underlying io.WriteCloser error
		})

		t.Run("writer close succeeds", func(t *testing.T) {
			t.Parallel()
			r := is.New(t)

			var buf bytes.Buffer

			writer := RollingWriter{Writer: writeCloser{Writer: &buf, Closer: noopCloser{}}}

			err := writer.Close()
			r.NoErr(err) // should not be any error
		})
	})
}

func fixedTrigger(value bool) Trigger {
	return &TriggerMock{
		TriggerFunc: func(_ io.Writer, _ []byte) (bool, error) {
			return value, nil
		},
	}
}

func faultyTrigger(err error) Trigger {
	return &TriggerMock{
		TriggerFunc: func(_ io.Writer, _ []byte) (bool, error) {
			return false, err
		},
	}
}

func fixedRotator(value io.Writer) Rotator {
	return &RotatorMock{
		RotateFunc: func(_ io.Writer) (io.Writer, error) {
			return value, nil
		},
	}
}

func faultyRotator(err error) Rotator {
	return &RotatorMock{
		RotateFunc: func(w io.Writer) (io.Writer, error) {
			return w, err
		},
	}
}

type writeCloser struct {
	io.Writer
	io.Closer
}

type faultyWriter struct {
	Err error
}

func (w faultyWriter) Write(_ []byte) (int, error) {
	return 0, w.Err
}

type testErr string

func (e testErr) Error() string {
	return string(e)
}

type faultyCloser struct {
	Err error
}

func (c faultyCloser) Close() error {
	return c.Err
}

type noopCloser struct {
}

func (c noopCloser) Close() error {
	return nil
}

const (
	errTrigger testErr = "err trigger"
	errRotate  testErr = "err rotate"
	errWrite   testErr = "err write"
	errClose   testErr = "err close"
)
