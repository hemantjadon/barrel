// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package barrel

import (
	"io"
	"sync"
)

// Ensure, that TriggerMock does implement Trigger.
// If this is not the case, regenerate this file with moq.
var _ Trigger = &TriggerMock{}

// TriggerMock is a mock implementation of Trigger.
//
//     func TestSomethingThatUsesTrigger(t *testing.T) {
//
//         // make and configure a mocked Trigger
//         mockedTrigger := &TriggerMock{
//             TriggerFunc: func(w io.Writer, p []byte) (bool, error) {
// 	               panic("mock out the Trigger method")
//             },
//         }
//
//         // use mockedTrigger in code that requires Trigger
//         // and then make assertions.
//
//     }
type TriggerMock struct {
	// TriggerFunc mocks the Trigger method.
	TriggerFunc func(w io.Writer, p []byte) (bool, error)

	// calls tracks calls to the methods.
	calls struct {
		// Trigger holds details about calls to the Trigger method.
		Trigger []struct {
			// W is the w argument value.
			W io.Writer
			// P is the p argument value.
			P []byte
		}
	}
	lockTrigger sync.RWMutex
}

// Trigger calls TriggerFunc.
func (mock *TriggerMock) Trigger(w io.Writer, p []byte) (bool, error) {
	if mock.TriggerFunc == nil {
		panic("TriggerMock.TriggerFunc: method is nil but Trigger.Trigger was just called")
	}
	callInfo := struct {
		W io.Writer
		P []byte
	}{
		W: w,
		P: p,
	}
	mock.lockTrigger.Lock()
	mock.calls.Trigger = append(mock.calls.Trigger, callInfo)
	mock.lockTrigger.Unlock()
	return mock.TriggerFunc(w, p)
}

// TriggerCalls gets all the calls that were made to Trigger.
// Check the length with:
//     len(mockedTrigger.TriggerCalls())
func (mock *TriggerMock) TriggerCalls() []struct {
	W io.Writer
	P []byte
} {
	var calls []struct {
		W io.Writer
		P []byte
	}
	mock.lockTrigger.RLock()
	calls = mock.calls.Trigger
	mock.lockTrigger.RUnlock()
	return calls
}

// Ensure, that RotatorMock does implement Rotator.
// If this is not the case, regenerate this file with moq.
var _ Rotator = &RotatorMock{}

// RotatorMock is a mock implementation of Rotator.
//
//     func TestSomethingThatUsesRotator(t *testing.T) {
//
//         // make and configure a mocked Rotator
//         mockedRotator := &RotatorMock{
//             RotateFunc: func(w io.Writer) (io.Writer, error) {
// 	               panic("mock out the Rotate method")
//             },
//         }
//
//         // use mockedRotator in code that requires Rotator
//         // and then make assertions.
//
//     }
type RotatorMock struct {
	// RotateFunc mocks the Rotate method.
	RotateFunc func(w io.Writer) (io.Writer, error)

	// calls tracks calls to the methods.
	calls struct {
		// Rotate holds details about calls to the Rotate method.
		Rotate []struct {
			// W is the w argument value.
			W io.Writer
		}
	}
	lockRotate sync.RWMutex
}

// Rotate calls RotateFunc.
func (mock *RotatorMock) Rotate(w io.Writer) (io.Writer, error) {
	if mock.RotateFunc == nil {
		panic("RotatorMock.RotateFunc: method is nil but Rotator.Rotate was just called")
	}
	callInfo := struct {
		W io.Writer
	}{
		W: w,
	}
	mock.lockRotate.Lock()
	mock.calls.Rotate = append(mock.calls.Rotate, callInfo)
	mock.lockRotate.Unlock()
	return mock.RotateFunc(w)
}

// RotateCalls gets all the calls that were made to Rotate.
// Check the length with:
//     len(mockedRotator.RotateCalls())
func (mock *RotatorMock) RotateCalls() []struct {
	W io.Writer
} {
	var calls []struct {
		W io.Writer
	}
	mock.lockRotate.RLock()
	calls = mock.calls.Rotate
	mock.lockRotate.RUnlock()
	return calls
}
