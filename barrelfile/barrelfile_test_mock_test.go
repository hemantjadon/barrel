// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package barrelfile

import (
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
//             TriggerFunc: func(path string, p []byte) (bool, error) {
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
	TriggerFunc func(path string, p []byte) (bool, error)

	// calls tracks calls to the methods.
	calls struct {
		// Trigger holds details about calls to the Trigger method.
		Trigger []struct {
			// Path is the path argument value.
			Path string
			// P is the p argument value.
			P []byte
		}
	}
	lockTrigger sync.RWMutex
}

// Trigger calls TriggerFunc.
func (mock *TriggerMock) Trigger(path string, p []byte) (bool, error) {
	if mock.TriggerFunc == nil {
		panic("TriggerMock.TriggerFunc: method is nil but Trigger.Trigger was just called")
	}
	callInfo := struct {
		Path string
		P    []byte
	}{
		Path: path,
		P:    p,
	}
	mock.lockTrigger.Lock()
	mock.calls.Trigger = append(mock.calls.Trigger, callInfo)
	mock.lockTrigger.Unlock()
	return mock.TriggerFunc(path, p)
}

// TriggerCalls gets all the calls that were made to Trigger.
// Check the length with:
//     len(mockedTrigger.TriggerCalls())
func (mock *TriggerMock) TriggerCalls() []struct {
	Path string
	P    []byte
} {
	var calls []struct {
		Path string
		P    []byte
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
//             RotateFunc: func(path string) (string, error) {
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
	RotateFunc func(path string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Rotate holds details about calls to the Rotate method.
		Rotate []struct {
			// Path is the path argument value.
			Path string
		}
	}
	lockRotate sync.RWMutex
}

// Rotate calls RotateFunc.
func (mock *RotatorMock) Rotate(path string) (string, error) {
	if mock.RotateFunc == nil {
		panic("RotatorMock.RotateFunc: method is nil but Rotator.Rotate was just called")
	}
	callInfo := struct {
		Path string
	}{
		Path: path,
	}
	mock.lockRotate.Lock()
	mock.calls.Rotate = append(mock.calls.Rotate, callInfo)
	mock.lockRotate.Unlock()
	return mock.RotateFunc(path)
}

// RotateCalls gets all the calls that were made to Rotate.
// Check the length with:
//     len(mockedRotator.RotateCalls())
func (mock *RotatorMock) RotateCalls() []struct {
	Path string
} {
	var calls []struct {
		Path string
	}
	mock.lockRotate.RLock()
	calls = mock.calls.Rotate
	mock.lockRotate.RUnlock()
	return calls
}

// Ensure, that TransformerMock does implement Transformer.
// If this is not the case, regenerate this file with moq.
var _ Transformer = &TransformerMock{}

// TransformerMock is a mock implementation of Transformer.
//
//     func TestSomethingThatUsesTransformer(t *testing.T) {
//
//         // make and configure a mocked Transformer
//         mockedTransformer := &TransformerMock{
//             TransformFunc: func(path string) (string, error) {
// 	               panic("mock out the Transform method")
//             },
//         }
//
//         // use mockedTransformer in code that requires Transformer
//         // and then make assertions.
//
//     }
type TransformerMock struct {
	// TransformFunc mocks the Transform method.
	TransformFunc func(path string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Transform holds details about calls to the Transform method.
		Transform []struct {
			// Path is the path argument value.
			Path string
		}
	}
	lockTransform sync.RWMutex
}

// Transform calls TransformFunc.
func (mock *TransformerMock) Transform(path string) (string, error) {
	if mock.TransformFunc == nil {
		panic("TransformerMock.TransformFunc: method is nil but Transformer.Transform was just called")
	}
	callInfo := struct {
		Path string
	}{
		Path: path,
	}
	mock.lockTransform.Lock()
	mock.calls.Transform = append(mock.calls.Transform, callInfo)
	mock.lockTransform.Unlock()
	return mock.TransformFunc(path)
}

// TransformCalls gets all the calls that were made to Transform.
// Check the length with:
//     len(mockedTransformer.TransformCalls())
func (mock *TransformerMock) TransformCalls() []struct {
	Path string
} {
	var calls []struct {
		Path string
	}
	mock.lockTransform.RLock()
	calls = mock.calls.Transform
	mock.lockTransform.RUnlock()
	return calls
}

// Ensure, that NamerMock does implement Namer.
// If this is not the case, regenerate this file with moq.
var _ Namer = &NamerMock{}

// NamerMock is a mock implementation of Namer.
//
//     func TestSomethingThatUsesNamer(t *testing.T) {
//
//         // make and configure a mocked Namer
//         mockedNamer := &NamerMock{
//             NameFunc: func(current string) (string, error) {
// 	               panic("mock out the Name method")
//             },
//         }
//
//         // use mockedNamer in code that requires Namer
//         // and then make assertions.
//
//     }
type NamerMock struct {
	// NameFunc mocks the Name method.
	NameFunc func(current string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Name holds details about calls to the Name method.
		Name []struct {
			// Current is the current argument value.
			Current string
		}
	}
	lockName sync.RWMutex
}

// Name calls NameFunc.
func (mock *NamerMock) Name(current string) (string, error) {
	if mock.NameFunc == nil {
		panic("NamerMock.NameFunc: method is nil but Namer.Name was just called")
	}
	callInfo := struct {
		Current string
	}{
		Current: current,
	}
	mock.lockName.Lock()
	mock.calls.Name = append(mock.calls.Name, callInfo)
	mock.lockName.Unlock()
	return mock.NameFunc(current)
}

// NameCalls gets all the calls that were made to Name.
// Check the length with:
//     len(mockedNamer.NameCalls())
func (mock *NamerMock) NameCalls() []struct {
	Current string
} {
	var calls []struct {
		Current string
	}
	mock.lockName.RLock()
	calls = mock.calls.Name
	mock.lockName.RUnlock()
	return calls
}
