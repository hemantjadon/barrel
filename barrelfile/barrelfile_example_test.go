package barrelfile_test

import (
	"compress/gzip"
	"os"

	"github.com/hemantjadon/barrel"
	"github.com/hemantjadon/barrel/barrelfile"
)

func Example() {
	file, _ := os.OpenFile("/path/to/file", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	rollingWriter := barrel.RollingWriter{
		Writer: file,
		Trigger: barrelfile.TriggerAdapter{
			FileTrigger: barrelfile.SizeBasedTrigger{Size: 1 * 1e3 * 1e3}, // 1 MB
		},
		Rotator: barrelfile.RotatorAdapter{
			FileRotator: barrelfile.TransformRotator{
				Transformers: []barrelfile.Transformer{
					barrelfile.RenameTransformer{
						Namer: barrelfile.TimestampSequenceNamer{},
					},
					barrelfile.GzipTransformer{
						GzipLevel: gzip.DefaultCompression,
					},
				},
				Rotator: barrelfile.IdentityRotator{},
			},
			OpenFlag: os.O_WRONLY | os.O_CREATE | os.O_APPEND,
		},
	}

	_, _ = rollingWriter.Write([]byte("hello world"))
	defer func() { _ = rollingWriter.Close() }()
}
