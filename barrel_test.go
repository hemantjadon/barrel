package barrel

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/adamluzsi/testcase"
)

func TestRollingWriter(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Describe(`Write`, SpecRollingWriterWrite)
	s.Describe(`Close`, SpecRollingWriterClose)
}

func SpecRollingWriterWrite(s *testcase.Spec) {
	subject := func(t *testcase.T, p []byte) (int, error) {
		writer := t.I(`writer`).(*RollingWriter)
		return writer.Write(p)
	}

	s.Let(`writer`, func(t *testcase.T) interface{} {
		underlyingWriter, _ := t.I(`underlyingWriter`).(io.Writer)
		trigger, _ := t.I(`trigger`).(Trigger)
		rotator, _ := t.I(`rotator`).(Rotator)
		return &RollingWriter{
			Writer:  underlyingWriter,
			Trigger: trigger,
			Rotator: rotator,
		}
	})

	s.When(`closed`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			writer := t.I(`writer`).(*RollingWriter)
			_ = writer.Close()
		})

		s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return io.Writer(nil) })
		s.Let(`trigger`, func(t *testcase.T) interface{} { return Trigger(nil) })
		s.Let(`rotator`, func(t *testcase.T) interface{} { return Rotator(nil) })

		s.Then(`fails with ErrClosed`, func(t *testcase.T) {
			n, err := subject(t, nil)

			if !errors.Is(err, ErrClosed) {
				t.Errorf("got err = %v, want err = %v", err, ErrClosed)
			}

			if n != 0 {
				t.Errorf("got n = %v, want n = %v", n, 0)
			}
		})
	})

	s.When(`trigger errors`, func(s *testcase.Spec) {
		s.Let(`trigger`, func(t *testcase.T) interface{} { return faultyTrigger{} })

		s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return io.Writer(nil) })
		s.Let(`rotator`, func(t *testcase.T) interface{} { return Rotator(nil) })

		s.Then(`fails with error wrapping underlying error`, func(t *testcase.T) {
			n, err := subject(t, nil)

			if !errors.Is(err, errTrigger) {
				t.Errorf("got err = %v, want err = %v", err, errTrigger)
			}

			if n != 0 {
				t.Errorf("got n = %v, want n = %v", n, 0)
			}
		})
	})

	s.When(`trigger returns false`, func(s *testcase.Spec) {
		s.Let(`trigger`, func(t *testcase.T) interface{} { return falseTrigger{} })

		s.And(`underlying writer errors`, func(s *testcase.Spec) {
			s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return faultyWriter{} })

			s.Let(`rotator`, func(t *testcase.T) interface{} { return Rotator(nil) })

			s.Then(`fails with error wrapping underlying error`, func(t *testcase.T) {
				n, err := subject(t, nil)

				if !errors.Is(err, errWriter) {
					t.Errorf("got err = %v, want err = %v", err, errWriter)
				}

				if n != 0 {
					t.Errorf("got n = %v, want n = %v", n, 0)
				}
			})
		})

		s.And(`underlying writer buffers`, func(s *testcase.Spec) {
			s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return &bytes.Buffer{} })

			s.Let(`rotator`, func(t *testcase.T) interface{} { return Rotator(nil) })

			s.Then(`succeeds with correct number of bytes`, func(t *testcase.T) {
				bs := []byte(`hello`)
				n, err := subject(t, bs)

				if err != nil {
					t.Errorf("got err = %v, want err = %v", err, nil)
				}

				if n != len(bs) {
					t.Errorf("got n = %v, want n = %v", n, len(bs))
				}

				underlyingWriter := t.I(`underlyingWriter`).(*bytes.Buffer)

				if !bytes.Equal(underlyingWriter.Bytes(), bs) {
					t.Errorf("got bytes = %v, want bytes = %v", underlyingWriter.Bytes(), bs)
				}
			})
		})
	})

	s.When(`trigger returns true`, func(s *testcase.Spec) {
		s.Let(`trigger`, func(t *testcase.T) interface{} { return trueTrigger{} })

		s.And(`rotator errors`, func(s *testcase.Spec) {
			s.Let(`rotator`, func(t *testcase.T) interface{} { return faultyRotator{} })

			s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return io.Writer(nil) })

			s.Then(`fails with error wrapping underlying error`, func(t *testcase.T) {
				n, err := subject(t, nil)

				if !errors.Is(err, errRotate) {
					t.Errorf("got err = %v, want err = %v", err, errRotate)
				}

				if n != 0 {
					t.Errorf("got n = %v, want n = %v", n, 0)
				}
			})
		})

		s.And(`rotator successfully rotates`, func(s *testcase.Spec) {
			s.Let(`rotator`, func(t *testcase.T) interface{} { return noopRotator{} })

			s.And(`underlying writer errors`, func(s *testcase.Spec) {
				s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return faultyWriter{} })

				s.Then(`fails with error wrapping underlying error`, func(t *testcase.T) {
					n, err := subject(t, nil)

					if !errors.Is(err, errWriter) {
						t.Errorf("got err = %v, want err = %v", err, errWriter)
					}

					if n != 0 {
						t.Errorf("got n = %v, want n = %v", n, 0)
					}
				})
			})

			s.And(`underlying writer buffers`, func(s *testcase.Spec) {
				s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return &bytes.Buffer{} })

				s.Then(`succeeds with correct number of bytes`, func(t *testcase.T) {
					bs := []byte(`hello`)
					n, err := subject(t, bs)

					if err != nil {
						t.Errorf("got err = %v, want err = %v", err, nil)
					}

					if n != len(bs) {
						t.Errorf("got n = %v, want n = %v", n, len(bs))
					}

					underlyingWriter := t.I(`underlyingWriter`).(*bytes.Buffer)

					if !bytes.Equal(underlyingWriter.Bytes(), bs) {
						t.Errorf("got bytes = %v, want bytes = %v", underlyingWriter.Bytes(), bs)
					}
				})
			})
		})
	})
}

func SpecRollingWriterClose(s *testcase.Spec) {
	subject := func(t *testcase.T) error {
		writer := t.I(`writer`).(*RollingWriter)
		return writer.Close()
	}

	s.Let(`writer`, func(t *testcase.T) interface{} {
		underlyingWriter, _ := t.I(`underlyingWriter`).(io.Writer)
		trigger, _ := t.I(`trigger`).(Trigger)
		rotator, _ := t.I(`rotator`).(Rotator)
		return &RollingWriter{
			Writer:  underlyingWriter,
			Trigger: trigger,
			Rotator: rotator,
		}
	})

	s.When(`closed`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			writer := t.I(`writer`).(*RollingWriter)
			_ = writer.Close()
		})

		s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return io.Writer(nil) })
		s.Let(`trigger`, func(t *testcase.T) interface{} { return Trigger(nil) })
		s.Let(`rotator`, func(t *testcase.T) interface{} { return Rotator(nil) })

		s.Then(`fails with ErrClosed`, func(t *testcase.T) {
			err := subject(t)

			if !errors.Is(err, ErrClosed) {
				t.Errorf("got err = %v, want err = %v", err, ErrClosed)
			}
		})
	})

	s.When(`open`, func(s *testcase.Spec) {
		s.Let(`trigger`, func(t *testcase.T) interface{} { return Trigger(nil) })
		s.Let(`rotator`, func(t *testcase.T) interface{} { return Rotator(nil) })

		s.And(`underlying writer is not io.Closer`, func(s *testcase.Spec) {
			s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return noopWriter{} })

			s.Then(`succeeds with no error`, func(t *testcase.T) {
				err := subject(t)

				if err != nil {
					t.Errorf("got err = %v, want err = %v", err, nil)
				}
			})
		})

		s.And(`underlying writer is faulty io.Closer`, func(s *testcase.Spec) {
			s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return writeCloser{Writer: noopWriter{}, Closer: faultyCloser{}} })

			s.Then(`fails with underlying error`, func(t *testcase.T) {
				err := subject(t)

				if !errors.Is(err, errClose) {
					t.Errorf("got err = %v, want err = %v", err, errClose)
				}
			})
		})

		s.And(`underlying writer is correct io.Closer`, func(s *testcase.Spec) {
			s.Let(`underlyingWriter`, func(t *testcase.T) interface{} { return writeCloser{Writer: noopWriter{}, Closer: noopCloser{}} })

			s.Then(`succeeds with no error`, func(t *testcase.T) {
				err := subject(t)

				if err != nil {
					t.Errorf("got err = %v, want err = %v", err, nil)
				}
			})
		})
	})
}

type falseTrigger struct {
}

func (t falseTrigger) Trigger(_ io.Writer, _ []byte) (bool, error) {
	return false, nil
}

type trueTrigger struct {
}

func (t trueTrigger) Trigger(_ io.Writer, _ []byte) (bool, error) {
	return true, nil
}

const errTrigger = testErr("err trigger")

type faultyTrigger struct {
}

func (t faultyTrigger) Trigger(_ io.Writer, _ []byte) (bool, error) {
	return false, errTrigger
}

type noopRotator struct {
}

func (t noopRotator) Rotate(w io.Writer) (io.Writer, error) {
	return w, nil
}

const errRotate = testErr("err rotate")

type faultyRotator struct {
}

func (t faultyRotator) Rotate(_ io.Writer) (io.Writer, error) {
	return nil, errRotate
}

type noopWriter struct {
}

func (t noopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

const errWriter = testErr("err writer")

type faultyWriter struct {
}

func (t faultyWriter) Write(_ []byte) (int, error) {
	return 0, errWriter
}

type noopCloser struct {
}

func (c noopCloser) Close() error {
	return nil
}

const errClose = testErr("err close")

type faultyCloser struct {
	io.Writer
}

func (c faultyCloser) Close() error {
	return errClose
}

type writeCloser struct {
	io.Writer
	io.Closer
}

type testErr string

func (e testErr) Error() string {
	return string(e)
}
