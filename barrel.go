package barrel

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// Trigger tells whether rotation should be done or not.
// If it returns true then rotation is preformed, otherwise not.
// If there are any errors while determining the trigger then a non-nil error
// is returned.
type Trigger interface {
	Trigger(w io.Writer, p []byte) (bool, error)
}

// Rotator changes the writer. If there is an error while rotating then a
// non-nil error is returned.
type Rotator interface {
	Rotate(w io.Writer) (io.Writer, error)
}

// RollingWriter wraps an io.Writer, providing mechanisms to check and perform
// rotation on each write.
type RollingWriter struct {
	// Writer which is wrapped.
	io.Writer

	// Trigger used to check whether to rotate or not.
	Trigger

	// Rotator used to change the Writer.
	Rotator

	// Serializes access to underlying Writer, Trigger and Rotator but does not
	// serialize the actual Write calls.
	mu sync.Mutex

	// Tells whether closed or not.
	closed int32
}

// Write writes the given bytes to the underlying Writer.
//
// Before writing, it calls Trigger.Trigger, to check whether to rotate or not,
// if it returns true then, Rotator.Rotate is called and the underlying writer
// is changed to the new writer. And bytes are written to the new Writer.
//
// If Trigger.Trigger returns false the bytes are immediately written to current
// Writer.
func (w *RollingWriter) Write(p []byte) (int, error) {
	if w.isClosed() {
		return 0, ErrClosed
	}
	w.mu.Lock()
	trigger, err := w.Trigger.Trigger(w.Writer, p)
	if err != nil {
		w.mu.Unlock()
		return 0, fmt.Errorf("trigger: %w", err)
	}
	if !trigger {
		w.mu.Unlock()
		return w.Writer.Write(p)
	}
	newWriter, err := w.Rotator.Rotate(w.Writer)
	if err != nil {
		w.mu.Unlock()
		return 0, fmt.Errorf("rotator rotate: %w", err)
	}
	w.Writer = newWriter
	w.mu.Unlock()
	return w.Writer.Write(p)
}

// Close closes the RollingWriter.
func (w *RollingWriter) Close() error {
	if w.isClosed() {
		return ErrClosed
	}
	return w.close()
}

func (w *RollingWriter) close() error {
	defer atomic.StoreInt32(&w.closed, 1)
	w.mu.Lock()
	defer w.mu.Unlock()
	if wc, ok := w.Writer.(io.Closer); ok {
		return wc.Close()
	}
	return nil
}

func (w *RollingWriter) isClosed() bool {
	return atomic.LoadInt32(&w.closed) != 0
}
