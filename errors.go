package barrel

const (
	// ErrClosed is returned when the operations are performed on an already
	// closed item.
	ErrClosed = barrelError("closed writer")
)

type barrelError string

func (e barrelError) Error() string {
	return string(e)
}
