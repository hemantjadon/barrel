package barrel_test

import (
	"os"

	"github.com/hemantjadon/barrel"
)

func ExampleRollingWriter() {
	file, _ := os.OpenFile("/path/to/file", os.O_CREATE|os.O_WRONLY|os.O_CREATE, 0644)
	writer := barrel.RollingWriter{
		Writer:  file,
		Trigger: barrel.AdaptFileTrigger(barrel.FileSizeBasedTrigger{Size: 100 * 1000 * 1000}),
		Rotator: barrel.AdaptFileRotator(barrel.FileTransformingRotator{
			OpenFlag: os.O_CREATE | os.O_WRONLY | os.O_CREATE,
			Transformers: []barrel.FileTransformer{
				barrel.FileRenameTransformer{
					NameGenerator: barrel.FileTimestampNameGenerator{},
				},
			},
		}),
	}
	_, _ = writer.Write([]byte(`hello world`))
}
