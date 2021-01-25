# barrel

![Go](https://github.com/hemantjadon/barrel/workflows/barrel/badge.svg?branch=master&event=push)

`barrel` provides a rolling writer which can be used as an `io.Writer` when
the underlying writer may change.

## Usage

`barrel` can be used at any place where some changing `io.Writer` is needed.

### Files

`barrel` can be used with package `barrel/barrelfile` to write rolling files, 
where the underlying file vary according to some trigger.

```go
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
```
