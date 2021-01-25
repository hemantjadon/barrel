package barrelfile

import (
	"io/ioutil"
	"os"
	"testing"
)

//go:generate moq -fmt goimports -out ./barrelfile_test_mock_test.go . Trigger Rotator Transformer Namer

type testError string

func (e testError) Error() string {
	return string(e)
}

func SetupDir(tb testing.TB) (path string) {
	tb.Helper()

	dir, err := ioutil.TempDir("/tmp", "barrel-temp-*")
	if err != nil {
		tb.Fatalf("create temp dir: %v", err)
	}
	tb.Cleanup(func() {
		if err := os.Remove(dir); err != nil {
			tb.Fatalf("os remove temp dir: %v", err)
		}
	})
	return dir
}

func NewFile(tb testing.TB, dir, pattern string) (path string) {
	tb.Helper()

	file, err := ioutil.TempFile(dir, pattern)
	if err != nil {
		tb.Fatalf("create temp file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			tb.Fatalf("close temp file: %v", err)
		}
	}()
	tb.Cleanup(func() {
		if err := os.Remove(file.Name()); err != nil {
			tb.Fatalf("remove temp file: %v", err)
		}
	})
	return file.Name()
}

func fixedTrigger(value bool) Trigger {
	return &TriggerMock{
		TriggerFunc: func(_ string, _ []byte) (bool, error) {
			return value, nil
		},
	}
}

func faultyTrigger(err error) Trigger {
	return &TriggerMock{
		TriggerFunc: func(_ string, _ []byte) (bool, error) {
			return false, err
		},
	}
}

func fixedRotator(path string) Rotator {
	return &RotatorMock{
		RotateFunc: func(_ string) (string, error) {
			return path, nil
		},
	}
}

func faultyRotator(err error) Rotator {
	return &RotatorMock{
		RotateFunc: func(path string) (string, error) {
			return path, err
		},
	}
}

func noopTransformer() Transformer {
	return &TransformerMock{
		TransformFunc: func(path string) (string, error) {
			return path, nil
		},
	}
}

func faultyTransformer(err error) Transformer {
	return &TransformerMock{
		TransformFunc: func(path string) (string, error) {
			return path, err
		},
	}
}

func fixedNamer(name string) Namer {
	return &NamerMock{
		NameFunc: func(_ string) (string, error) {
			return name, nil
		},
	}
}

func faultyNamer(err error) Namer {
	return &NamerMock{
		NameFunc: func(current string) (string, error) {
			return current, err
		},
	}
}

const (
	errTrigger     testError = "err trigger"
	errRotator     testError = "err rotator"
	errTransformer testError = "err transformer"
	errNamer       testError = "err namer"
)
