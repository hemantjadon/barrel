package barrel

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/adamluzsi/testcase"
)

var (
	testTime = time.Date(2020, 1, 1, 6, 15, 0, 0, time.UTC)
	testText = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur volutpat eleifend velit, 
quis elementum metus maximus non. Aliquam rutrum placerat sapien, vitae egestas ante egestas 
sit amet. Morbi finibus ex eget leo cursus, a porta lorem aliquam. Sed sollicitudin eu purus 
sit amet lobortis. Nullam nec nulla hendrerit, aliquet magna dictum, facilisis nibh. Duis 
lacinia nec velit ut consequat. Maecenas vel lobortis erat. Nunc sollicitudin consectetur 
nulla id hendrerit. Nunc rhoncus eros non efficitur ullamcorper. Pellentesque habitant morbi 
tristique senectus et netus et malesuada fames ac turpis egestas. Aliquam pulvinar accumsan cursus. 
Curabitur id orci commodo, viverra elit ut, posuere leo. Cras at dolor arcu. Maecenas efficitur enim 
commodo turpis efficitur dapibus.`
)

func TestAdaptFileTrigger(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`Trigger`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (bool, error) {
			writer := t.I(`writer`).(io.Writer)
			bs := t.I(`bytes`).([]byte)
			trigger := t.I(`trigger`).(Trigger)
			return trigger.Trigger(writer, bs)
		}

		s.Let(`trigger`, func(t *testcase.T) interface{} {
			fileTrigger := t.I(`fileTrigger`).(FileTrigger)
			return AdaptFileTrigger(fileTrigger)
		})

		s.Let(`fileTrigger`, func(t *testcase.T) interface{} {
			return &noopFileTrigger{}
		})

		s.When(`os.File is passed to trigger`, func(s *testcase.Spec) {
			s.Let(`writer`, func(t *testcase.T) interface{} { return &os.File{} })

			s.Let(`bytes`, func(t *testcase.T) interface{} { return []byte(nil) })

			s.Then(`no error is returned`, func(t *testcase.T) {
				_, err := subject(t)

				if err != nil {
					t.Fatalf("got err = %v, want err = %v", err, nil)
				}
			})
		})

		s.When(`writer other than os.File is passed to trigger`, func(s *testcase.Spec) {
			s.Let(`writer`, func(t *testcase.T) interface{} { return &bytes.Buffer{} })

			s.Let(`bytes`, func(t *testcase.T) interface{} { return []byte(nil) })

			s.Then(`non nil error is returned`, func(t *testcase.T) {
				_, err := subject(t)

				if err == nil {
					t.Fatalf("got err = %v, want err = <non-nil>", err)
				}
			})
		})
	})
}

func TestAdaptFileRotator(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`Rotator`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (io.Writer, error) {
			writer := t.I(`writer`).(io.Writer)
			rotator := t.I(`rotator`).(Rotator)
			return rotator.Rotate(writer)
		}

		s.Let(`rotator`, func(t *testcase.T) interface{} {
			fileRotator := t.I(`fileRotator`).(FileRotator)
			return AdaptFileRotator(fileRotator)
		})

		s.Let(`fileRotator`, func(t *testcase.T) interface{} {
			return &noopFileRotator{}
		})

		s.When(`os.File is passed to rotator`, func(s *testcase.Spec) {
			s.Let(`writer`, func(t *testcase.T) interface{} { return &os.File{} })

			s.Then(`no error is returned`, func(t *testcase.T) {
				_, err := subject(t)

				if err != nil {
					t.Fatalf("got err = %v, want err = %v", err, nil)
				}
			})
		})

		s.When(`writer other than os.File is passed to rotator`, func(s *testcase.Spec) {
			s.Let(`writer`, func(t *testcase.T) interface{} { return &bytes.Buffer{} })

			s.Then(`non nil error is returned`, func(t *testcase.T) {
				_, err := subject(t)

				if err == nil {
					t.Fatalf("got err = %v, want err = <non-nil>", err)
				}
			})
		})
	})
}

func TestFileSizeBasedTrigger(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`Trigger`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (bool, error) {
			trigger := t.I(`trigger`).(*FileSizeBasedTrigger)
			file := t.I(`file`).(*os.File)
			p := t.I(`bytes`).([]byte)
			return trigger.Trigger(file, p)
		}

		s.Let(`trigger`, func(t *testcase.T) interface{} {
			size := t.I(`size`).(int64)
			return &FileSizeBasedTrigger{Size: size}
		})

		s.Let(`testDir`, func(t *testcase.T) interface{} {
			dir, err := ioutil.TempDir("/tmp", "barrel-temp-*")
			if err != nil {
				t.Fatalf("create temp dir: %v", err)
			}
			t.Defer(func() {
				if err := os.Remove(dir); err != nil {
					t.Fatalf("remove temp dir: %v", err)
				}
			})
			return dir
		})

		s.Let(`file`, func(t *testcase.T) interface{} {
			testDir := t.I(`testDir`).(string)
			file, err := ioutil.TempFile(testDir, "file-size-based-trigger-*")
			if err != nil {
				t.Fatalf("os open test file: %v", err)
			}
			t.Defer(func() {
				if err := os.Remove(file.Name()); err != nil {
					t.Fatalf("os remove test file: %v", err)
				}
			})
			t.Defer(file.Close)
			return file
		})

		s.When(`file size plus bytes are smaller than max size`, func(s *testcase.Spec) {
			s.Let(`size`, func(t *testcase.T) interface{} { return int64(500) })

			s.Before(func(t *testcase.T) {
				file := t.I(`file`).(*os.File)
				_, err := file.Write([]byte("hello\n"))
				if err != nil {
					t.Fatalf("file write: %v", err)
				}
			})

			s.Let(`bytes`, func(t *testcase.T) interface{} { return []byte(`world`) })

			s.Then(`it returns false with no error`, func(t *testcase.T) {
				trigger, err := subject(t)

				if err != nil {
					t.Errorf("got err = %v, want err = %v", err, nil)
				}

				if trigger != false {
					t.Errorf("got trigger = %v, want trigger = %v", trigger, false)
				}
			})
		})

		s.When(`file size plus bytes are greater than max size`, func(s *testcase.Spec) {
			s.Let(`size`, func(t *testcase.T) interface{} { return int64(7) })

			s.Before(func(t *testcase.T) {
				file := t.I(`file`).(*os.File)
				_, err := file.Write([]byte("hello\n"))
				if err != nil {
					t.Fatalf("file write: %v", err)
				}
			})

			s.Let(`bytes`, func(t *testcase.T) interface{} { return []byte(`world`) })

			s.Then(`it returns true with no error`, func(t *testcase.T) {
				trigger, err := subject(t)

				if err != nil {
					t.Errorf("got err = %v, want err = %v", err, nil)
				}

				if trigger != true {
					t.Errorf("got trigger = %v, want trigger = %v", trigger, true)
				}
			})
		})

		s.When(`bytes are greater than max size`, func(s *testcase.Spec) {
			s.Let(`size`, func(t *testcase.T) interface{} { return int64(10) })

			s.Before(func(t *testcase.T) {
				file := t.I(`file`).(*os.File)
				_, err := file.Write([]byte("hello\n"))
				if err != nil {
					t.Fatalf("file write: %v", err)
				}
			})

			s.Let(`bytes`, func(t *testcase.T) interface{} { return []byte(`beautiful world`) })

			s.Then(`it returns non nil error`, func(t *testcase.T) {
				_, err := subject(t)

				if err == nil {
					t.Errorf("got err = %v, want err = <non-nil>", err)
				}
			})
		})
	})
}

func TestFileTimeBasedTrigger(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`Trigger`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (bool, error) {
			trigger := t.I(`trigger`).(*FileTimeBasedTrigger)
			file := t.I(`file`).(*os.File)
			p := t.I(`bytes`).([]byte)
			return trigger.Trigger(file, p)
		}

		s.Let(`trigger`, func(t *testcase.T) interface{} {
			cronExp := t.I(`cron`).(string)
			clk := t.I(`clock`).(Clock)
			return &FileTimeBasedTrigger{CronExpression: cronExp, Clock: clk}
		})

		s.Let(`clock`, func(t *testcase.T) interface{} {
			clk := clock.NewMock()
			clk.Set(testTime)
			return clk
		})

		s.Let(`testDir`, func(t *testcase.T) interface{} {
			dir, err := ioutil.TempDir("/tmp", "barrel-temp-*")
			if err != nil {
				t.Fatalf("create temp dir: %v", err)
			}
			t.Defer(func() {
				if err := os.Remove(dir); err != nil {
					t.Fatalf("remove temp dir: %v", err)
				}
			})
			return dir
		})

		s.Let(`file`, func(t *testcase.T) interface{} {
			testDir := t.I(`testDir`).(string)
			file, err := ioutil.TempFile(testDir, "file-time-based-trigger-*")
			if err != nil {
				t.Fatalf("os open test file: %v", err)
			}
			t.Defer(func() {
				if err := os.Remove(file.Name()); err != nil {
					t.Fatalf("os remove test file: %v", err)
				}
			})
			t.Defer(file.Close)
			return file
		})

		s.When(`invalid cron expression is used`, func(s *testcase.Spec) {
			s.Let(`cron`, func(t *testcase.T) interface{} { return "abcd" })
			s.Let(`bytes`, func(t *testcase.T) interface{} { return []byte(nil) })

			s.Then(`it returns non nil error`, func(t *testcase.T) {
				_, err := subject(t)

				if err == nil {
					t.Errorf("got err = %v, want err = <non-nil>", err)
				}
			})
		})

		s.Context(`using hourly cron expression`, func(s *testcase.Spec) {
			s.Let(`cron`, func(t *testcase.T) interface{} { return "0 * * * *" })

			s.When(`file was modified in last one hour`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					file := t.I(`file`).(*os.File)
					clk := t.I(`clock`).(Clock)
					ft := clk.Now().Add(-15 * time.Minute)
					if err := os.Chtimes(file.Name(), ft, ft); err != nil {
						t.Fatalf("change mod time of file: %v", err)
					}
				})

				s.Let(`bytes`, func(t *testcase.T) interface{} { return []byte(nil) })

				s.Then(`it returns false with no error`, func(t *testcase.T) {
					trigger, err := subject(t)

					if err != nil {
						t.Errorf("got err = %v, want err = %v", err, nil)
					}

					if trigger != false {
						t.Errorf("got trigger = %v, want trigger = %v", trigger, false)
					}
				})
			})

			s.When(`file was modified in little more than one hour ago`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					file := t.I(`file`).(*os.File)
					clk := t.I(`clock`).(Clock)
					ft := clk.Now().Add(-75 * time.Minute)
					if err := os.Chtimes(file.Name(), ft, ft); err != nil {
						t.Fatalf("change mod time of file: %v", err)
					}
				})

				s.Let(`bytes`, func(t *testcase.T) interface{} { return []byte(nil) })

				s.Then(`it returns true with no error`, func(t *testcase.T) {
					trigger, err := subject(t)

					if err != nil {
						t.Errorf("got err = %v, want err = %v", err, nil)
					}

					if trigger != true {
						t.Errorf("got trigger = %v, want trigger = %v", trigger, true)
					}
				})
			})

			s.When(`file was modified in long time ago`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					file := t.I(`file`).(*os.File)
					clk := t.I(`clock`).(Clock)
					ft := clk.Now().Add(-75 * 24 * time.Hour)
					if err := os.Chtimes(file.Name(), ft, ft); err != nil {
						t.Fatalf("change mod time of file: %v", err)
					}
				})

				s.Let(`bytes`, func(t *testcase.T) interface{} { return []byte(nil) })

				s.Then(`it returns true with no error`, func(t *testcase.T) {
					trigger, err := subject(t)

					if err != nil {
						t.Errorf("got err = %v, want err = %v", err, nil)
					}

					if trigger != true {
						t.Errorf("got trigger = %v, want trigger = %v", trigger, true)
					}
				})
			})
		})
	})
}

func TestFileTransformingRotator(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`Rotate`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (*os.File, error) {
			rotator := t.I(`rotator`).(*FileTransformingRotator)
			file := t.I(`file`).(*os.File)
			return rotator.Rotate(file)
		}

		s.Let(`rotator`, func(t *testcase.T) interface{} {
			transformers := t.I(`transformers`).([]FileTransformer)
			return &FileTransformingRotator{
				OpenFlag:     os.O_RDWR,
				Transformers: transformers,
			}
		})

		s.Let(`testDir`, func(t *testcase.T) interface{} {
			dir, err := ioutil.TempDir("/tmp", "barrel-temp-*")
			if err != nil {
				t.Fatalf("create temp dir: %v", err)
			}
			t.Defer(func() {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatalf("remove temp dir: %v", err)
				}
			})
			return dir
		})

		s.Let(`file`, func(t *testcase.T) interface{} {
			testDir := t.I(`testDir`).(string)
			file, err := ioutil.TempFile(testDir, "file-transfoming-rotator-*")
			if err != nil {
				t.Fatalf("os open test file: %v", err)
			}
			return file
		})

		s.When(`one of the transformers is faulty`, func(s *testcase.Spec) {
			s.Let(`transformers`, func(t *testcase.T) interface{} {
				return []FileTransformer{
					&spyTransformer{FileTransformer: noopTransformer{}},
					&spyTransformer{FileTransformer: noopTransformer{}},
					&spyTransformer{FileTransformer: faultyTransformer{}},
					&spyTransformer{FileTransformer: noopTransformer{}},
					&spyTransformer{FileTransformer: noopTransformer{}},
				}
			})

			s.Then(`all transformers upto faulty transformer is called and a wrapped error is returned`, func(t *testcase.T) {
				_, err := subject(t)

				if !errors.Is(err, errTransform) {
					t.Fatalf("got err = %v, want err = %v", err, errTransform)
				}

				transformers := t.I(`transformers`).([]FileTransformer)

				wantCalled := true
				for idx, transformer := range transformers {
					spy := transformer.(*spyTransformer)
					if spy.called != wantCalled {
						t.Errorf("transformer[%d].called: got = %v, want = %v", idx, spy.called, wantCalled)
					}
					if _, ok := spy.FileTransformer.(faultyTransformer); ok {
						wantCalled = false
					}
				}
			})
		})

		s.When(`all transformers are correct`, func(s *testcase.Spec) {
			s.Let(`transformers`, func(t *testcase.T) interface{} {
				return []FileTransformer{
					&spyTransformer{FileTransformer: noopTransformer{}},
					&spyTransformer{FileTransformer: noopTransformer{}},
					&spyTransformer{FileTransformer: noopTransformer{}},
					&spyTransformer{FileTransformer: noopTransformer{}},
					&spyTransformer{FileTransformer: noopTransformer{}},
				}
			})

			var originalFileMode os.FileMode
			var originalFileName string

			s.Before(func(t *testcase.T) {
				file := t.I(`file`).(*os.File)
				stat, err := file.Stat()
				if err != nil {
					t.Fatalf("file stat: %v", err)
				}
				originalFileMode = stat.Mode()
				originalFileName = file.Name()
			})

			s.Then(`all transformers are called and a new file with same name and mode is returned`, func(t *testcase.T) {
				got, err := subject(t)

				if err != nil {
					t.Fatalf("got err = %v, want err = %v", err, nil)
				}

				transformers := t.I(`transformers`).([]FileTransformer)

				wantCalled := true
				for idx, transformer := range transformers {
					spy := transformer.(*spyTransformer)
					if spy.called != wantCalled {
						t.Errorf("transformer[%d].called: got = %v, want = %v", idx, spy.called, wantCalled)
					}
					if _, ok := spy.FileTransformer.(faultyTransformer); ok {
						wantCalled = false
					}
				}

				gotFileName := got.Name()

				if gotFileName != originalFileName {
					t.Errorf("got file name = %v, want file name = %v", gotFileName, originalFileName)
				}

				gotFileStat, err := got.Stat()
				if err != nil {
					t.Fatalf("got file stat: %v", err)
				}

				if gotFileStat.Mode() != originalFileMode {
					t.Errorf("got file mode = %v, want file mode = %v", gotFileStat.Mode(), originalFileMode)
				}
			})
		})
	})
}

func TestFileGzipTransformer(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`Transform`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (string, error) {
			transformer := t.I(`transformer`).(*FileGzipTransformer)
			path := t.I(`path`).(string)
			return transformer.Transform(path)
		}

		s.Let(`transformer`, func(t *testcase.T) interface{} {
			gzipLevel := t.I(`gzipLevel`).(int)
			return &FileGzipTransformer{GzipLevel: gzipLevel}
		})

		s.When(`file path is invalid`, func(s *testcase.Spec) {
			s.Let(`path`, func(t *testcase.T) interface{} {
				return "/some/invalid/path/with/something/invalid.go"
			})

			s.Let(`gzipLevel`, func(t *testcase.T) interface{} { return gzip.DefaultCompression })

			s.And(`transformer is executed`, func(s *testcase.Spec) {
				s.Then(`then non-nil error is returned`, func(t *testcase.T) {
					_, err := subject(t)

					if err == nil {
						t.Fatalf("got err = %v, want err = <non-nil>", err)
					}
				})
			})
		})

		s.When(`file path is correct`, func(s *testcase.Spec) {
			s.Let(`testDir`, func(t *testcase.T) interface{} {
				dir, err := ioutil.TempDir("/tmp", "barrel-temp-*")
				if err != nil {
					t.Fatalf("create temp dir: %v", err)
				}
				t.Defer(func() {
					if err := os.Remove(dir); err != nil {
						t.Fatalf("remove temp dir: %v", err)
					}
				})
				return dir
			})

			s.Let(`file`, func(t *testcase.T) interface{} {
				testDir := t.I(`testDir`).(string)
				file, err := ioutil.TempFile(testDir, "file-gzip-transformer-*")
				if err != nil {
					t.Fatalf("os open test file: %v", err)
				}
				return file
			})

			s.Let(`path`, func(t *testcase.T) interface{} {
				file := t.I(`file`).(*os.File)
				return file.Name()
			})

			s.And(`file has some content`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					file := t.I(`file`).(*os.File)
					if _, err := file.Write([]byte(testText)); err != nil {
						t.Fatalf("file write test bytes: %v", err)
					}
					if err := os.Chtimes(file.Name(), testTime, testTime); err != nil {
						t.Fatalf("os chtimes test file: %v", err)
					}
					if err := file.Sync(); err != nil {
						t.Errorf("file sync: %v", err)
					}
					if err := file.Close(); err != nil {
						t.Errorf("file close: %v", err)
					}
				})

				s.And(`transformer has invalid gzip level`, func(s *testcase.Spec) {
					s.Let(`gzipLevel`, func(t *testcase.T) interface{} {
						return gzip.BestCompression + 1
					})

					s.After(func(t *testcase.T) {
						path := t.I(`path`).(string)
						if err := os.Remove(path); err != nil {
							t.Errorf("os remove: original file: %w", err)
						}
					})

					s.And(`transformer is executed`, func(s *testcase.Spec) {
						s.Then(`then non-nil error is returned`, func(t *testcase.T) {
							_, err := subject(t)

							if err == nil {
								t.Fatalf("got err = %v, want err = <non-nil>", err)
							}
						})

						s.Then(`original file is not removed`, func(t *testcase.T) {
							_, err := subject(t)

							path := t.I(`path`).(string)

							_, err = os.Stat(path)
							if err != nil {
								t.Errorf("original file is removed")
							}
						})
					})
				})

				s.And(`transformer has valid gzip level`, func(s *testcase.Spec) {
					s.Let(`gzipLevel`, func(t *testcase.T) interface{} {
						return gzip.DefaultCompression
					})

					s.And(`transformer is executed`, func(s *testcase.Spec) {
						s.Then(`new file is created with gzipped content and no error`, func(t *testcase.T) {
							gzFilePath, err := subject(t)
							t.Defer(func() {
								if err := os.Remove(gzFilePath); err != nil {
									t.Errorf("gz file remove: got err = %v, want err = %v", err, nil)
								}
							})

							if err != nil {
								t.Fatalf("got err = %v, want err = %v", err, nil)
							}

							gzFileBytes, err := ioutil.ReadFile(gzFilePath)
							if err != nil {
								t.Fatalf("gz file open: got err = %v, want err = %v", err, nil)
							}

							gzr, err := gzip.NewReader(bytes.NewBuffer(gzFileBytes))
							if err != nil {
								t.Fatalf("gz new reader: got err = %v, want err = %v", err, nil)
							}

							rawTextBuf := bytes.Buffer{}
							if _, err := io.Copy(&rawTextBuf, gzr); err != nil {
								t.Fatalf("io copy gz to buffer: got err %v, want err = %v", err, nil)
							}

							if rawTextBuf.String() != testText {
								t.Errorf("got text = %v, want text = %v", rawTextBuf.String(), testText)
							}
						})

						s.Then(`new file has same mtime as original file`, func(t *testcase.T) {
							gzFilePath, err := subject(t)
							t.Defer(func() {
								if err := os.Remove(gzFilePath); err != nil {
									t.Errorf("gz file remove: got err = %v, want err = %v", err, nil)
								}
							})

							stat, err := os.Stat(gzFilePath)
							if err != nil {
								t.Errorf("gz file stat: got err = %v, want err = %v", err, nil)
							}

							if !stat.ModTime().Equal(testTime) {
								t.Errorf("gz file mtime: got = %v, want = %v", stat.ModTime(), testTime)
							}
						})

						s.Then(`original file is removed`, func(t *testcase.T) {
							gzFilePath, err := subject(t)
							t.Defer(func() {
								if err := os.Remove(gzFilePath); err != nil {
									t.Errorf("gz file remove: got err = %v, want err = %v", err, nil)
								}
							})

							path := t.I(`path`).(string)

							_, err = os.Stat(path)
							if err == nil {
								t.Errorf("original file not removed")
							}
						})
					})
				})
			})
		})
	})
}

func TestFileRenameTransformer(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`Transform`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (string, error) {
			transformer := t.I(`transformer`).(*FileRenameTransformer)
			path := t.I(`path`).(string)
			return transformer.Transform(path)
		}

		s.Let(`transformer`, func(t *testcase.T) interface{} {
			nameGenerator := t.I(`nameGenerator`).(NameGenerator)
			forceMove := t.I(`forceMove`).(bool)
			return &FileRenameTransformer{NameGenerator: nameGenerator, ForceMove: forceMove}
		})

		s.When(`file path is invalid`, func(s *testcase.Spec) {
			s.Let(`path`, func(t *testcase.T) interface{} {
				return "/some/invalid/path/with/something/invalid.go"
			})

			s.Let(`nameGenerator`, func(t *testcase.T) interface{} { return &testNameGenerator{} })
			s.Let(`forceMove`, func(t *testcase.T) interface{} { return false })

			s.And(`transformer is executed`, func(s *testcase.Spec) {
				s.Then(`then non-nil error is returned`, func(t *testcase.T) {
					_, err := subject(t)

					if err == nil {
						t.Fatalf("got err = %v, want err = <non-nil>", err)
					}
				})
			})
		})

		s.When(`file path is correct`, func(s *testcase.Spec) {
			s.Let(`testDir`, func(t *testcase.T) interface{} {
				dir, err := ioutil.TempDir("/tmp", "barrel-temp-*")
				if err != nil {
					t.Fatalf("create temp dir: %v", err)
				}
				t.Defer(func() {
					if err := os.Remove(dir); err != nil {
						t.Fatalf("remove temp dir: %v", err)
					}
				})
				return dir
			})

			s.Let(`file`, func(t *testcase.T) interface{} {
				testDir := t.I(`testDir`).(string)
				file, err := ioutil.TempFile(testDir, "file-rename-transformer-*")
				if err != nil {
					t.Fatalf("os open test file: %v", err)
				}
				return file
			})

			s.Let(`path`, func(t *testcase.T) interface{} {
				file := t.I(`file`).(*os.File)
				return file.Name()
			})

			s.And(`file has some content`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					file := t.I(`file`).(*os.File)
					if _, err := file.Write([]byte(testText)); err != nil {
						t.Fatalf("file write test bytes: %v", err)
					}
					if err := os.Chtimes(file.Name(), testTime, testTime); err != nil {
						t.Fatalf("os chtimes test file: %v", err)
					}
					if err := file.Sync(); err != nil {
						t.Errorf("file sync: %v", err)
					}
					if err := file.Close(); err != nil {
						t.Errorf("file close: %v", err)
					}
				})

				s.And(`force move is false`, func(s *testcase.Spec) {
					s.Let(`forceMove`, func(t *testcase.T) interface{} { return false })

					s.Let(`nameGenerator`, func(t *testcase.T) interface{} { return &testNameGenerator{} })

					s.And(`transformer is executed`, func(s *testcase.Spec) {
						s.Then(`new file is created with same content and no error`, func(t *testcase.T) {
							renamedFilePath, err := subject(t)
							t.Defer(func() {
								if err := os.Remove(renamedFilePath); err != nil {
									t.Errorf("gz file remove: got err = %v, want err = %v", err, nil)
								}
							})

							if err != nil {
								t.Fatalf("got err = %v, want err = %v", err, nil)
							}

							renamedFileBytes, err := ioutil.ReadFile(renamedFilePath)
							if err != nil {
								t.Fatalf("gz file open: got err = %v, want err = %v", err, nil)
							}

							if string(renamedFileBytes) != testText {
								t.Errorf("got text = %v, want text = %v", renamedFileBytes, testText)
							}
						})

						s.Then(`new file has same mtime as original file`, func(t *testcase.T) {
							renamedFilePath, err := subject(t)
							t.Defer(func() {
								if err := os.Remove(renamedFilePath); err != nil {
									t.Errorf("gz file remove: got err = %v, want err = %v", err, nil)
								}
							})

							stat, err := os.Stat(renamedFilePath)
							if err != nil {
								t.Errorf("gz file stat: got err = %v, want err = %v", err, nil)
							}

							if !stat.ModTime().Equal(testTime) {
								t.Errorf("gz file mtime: got = %v, want = %v", stat.ModTime(), testTime)
							}
						})

						s.Then(`original file is removed`, func(t *testcase.T) {
							renamedFilePath, err := subject(t)
							t.Defer(func() {
								if err := os.Remove(renamedFilePath); err != nil {
									t.Errorf("gz file remove: got err = %v, want err = %v", err, nil)
								}
							})

							path := t.I(`path`).(string)

							_, err = os.Stat(path)
							if err == nil {
								t.Errorf("original file not removed")
							}
						})
					})
				})

				s.And(`force move is true`, func(s *testcase.Spec) {
					s.Let(`forceMove`, func(t *testcase.T) interface{} { return true })

					s.Let(`nameGenerator`, func(t *testcase.T) interface{} { return &testNameGenerator{} })

					s.And(`transformer is executed`, func(s *testcase.Spec) {
						s.Then(`new file is created with same content and no error`, func(t *testcase.T) {
							renamedFilePath, err := subject(t)
							t.Defer(func() {
								if err := os.Remove(renamedFilePath); err != nil {
									t.Errorf("gz file remove: got err = %v, want err = %v", err, nil)
								}
							})

							if err != nil {
								t.Fatalf("got err = %v, want err = %v", err, nil)
							}

							renamedFileBytes, err := ioutil.ReadFile(renamedFilePath)
							if err != nil {
								t.Fatalf("gz file open: got err = %v, want err = %v", err, nil)
							}

							if string(renamedFileBytes) != testText {
								t.Errorf("got text = %v, want text = %v", renamedFileBytes, testText)
							}
						})

						s.Then(`new file has same mtime as original file`, func(t *testcase.T) {
							renamedFilePath, err := subject(t)
							t.Defer(func() {
								if err := os.Remove(renamedFilePath); err != nil {
									t.Errorf("gz file remove: got err = %v, want err = %v", err, nil)
								}
							})

							stat, err := os.Stat(renamedFilePath)
							if err != nil {
								t.Errorf("gz file stat: got err = %v, want err = %v", err, nil)
							}

							if !stat.ModTime().Equal(testTime) {
								t.Errorf("gz file mtime: got = %v, want = %v", stat.ModTime(), testTime)
							}
						})

						s.Then(`original file is removed`, func(t *testcase.T) {
							renamedFilePath, err := subject(t)
							t.Defer(func() {
								if err := os.Remove(renamedFilePath); err != nil {
									t.Errorf("gz file remove: got err = %v, want err = %v", err, nil)
								}
							})

							path := t.I(`path`).(string)

							_, err = os.Stat(path)
							if err == nil {
								t.Errorf("original file not removed")
							}
						})
					})
				})
			})
		})
	})
}

func TestFileTimestampNameGenerator(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`Name`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (string, error) {
			generator := t.I(`generator`).(*FileTimestampNameGenerator)
			currentName := t.I(`currentName`).(string)
			return generator.Name(currentName)
		}

		s.Let(`generator`, func(t *testcase.T) interface{} {
			format := t.I(`format`).(string)
			return &FileTimestampNameGenerator{
				Format: format,
			}
		})

		s.Let(`testDir`, func(t *testcase.T) interface{} {
			dir, err := ioutil.TempDir("/tmp", "barrel-temp-*")
			if err != nil {
				t.Fatalf("create temp dir: %v", err)
			}
			t.Defer(func() {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatalf("remove temp dir: %v", err)
				}
			})
			return dir
		})

		s.When(`directory of file does not have any other file with same suffixed name`, func(s *testcase.Spec) {
			s.Let(`currentName`, func(t *testcase.T) interface{} {
				testDir := t.I(`testDir`).(string)
				filename := filepath.Join(testDir, "application.log")
				file, err := os.Create(filename)
				if err != nil {
					t.Fatalf("os create file: %v", err)
				}
				defer func() { _ = file.Close() }()
				t.Defer(func() {
					if err := os.Remove(filename); err != nil {
						t.Fatalf("remove temp file: %v", err)
					}
				})
				if err := os.Chtimes(filename, testTime, testTime); err != nil {
					t.Fatalf("os chtimes: %v", err)
				}
				return filename
			})

			s.And(`format is provided`, func(s *testcase.Spec) {
				s.Let(`format`, func(t *testcase.T) interface{} { return "01-02-2006" }) // DD-MM-YYYY

				s.Then(`it returns time suffixed name in provided format`, func(t *testcase.T) {
					got, err := subject(t)

					if err != nil {
						t.Errorf("got err = %v, want err = %v", err, nil)
					}

					filename := t.I(`currentName`).(string)
					fileNameExt := filepathFullExt(filename)
					fileNameBase := filepathStripFullExt(filename)

					format := t.I(`format`).(string)

					want := fmt.Sprintf("%s_%s%s", fileNameBase, testTime.Format(format), fileNameExt)

					if got != want {
						t.Errorf("got = %v, want = %v", got, want)
					}
				})
			})

			s.And(`format is not provided`, func(s *testcase.Spec) {
				s.Let(`format`, func(t *testcase.T) interface{} { return "" })

				s.Then(`it returns time suffixed name in default format`, func(t *testcase.T) {
					got, err := subject(t)

					if err != nil {
						t.Errorf("got err = %v, want err = %v", err, nil)
					}

					filename := t.I(`currentName`).(string)
					fileNameExt := filepathFullExt(filename)
					fileNameBase := filepathStripFullExt(filename)

					format := "2006-02-01" // Default format

					want := fmt.Sprintf("%s_%s%s", fileNameBase, testTime.Format(format), fileNameExt)

					if got != want {
						t.Errorf("got = %v, want = %v", got, want)
					}
				})
			})
		})

		s.When(`directory of file already contain a file with same time suffixed name`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				testDir := t.I(`testDir`).(string)
				filename := filepath.Join(testDir, fmt.Sprintf("application_2020-01-01.log"))
				file, err := os.Create(filename)
				if err != nil {
					t.Fatalf("os create file: %v", err)
				}
				defer func() { _ = file.Close() }()
			})

			s.Let(`currentName`, func(t *testcase.T) interface{} {
				testDir := t.I(`testDir`).(string)
				filename := filepath.Join(testDir, "application.log")
				file, err := os.Create(filename)
				if err != nil {
					t.Fatalf("os create file: %v", err)
				}
				defer func() { _ = file.Close() }()
				if err := os.Chtimes(filename, testTime, testTime); err != nil {
					t.Fatalf("os chtimes: %v", err)
				}
				return filename
			})

			s.Let(`format`, func(t *testcase.T) interface{} { return "" }) // Default format YYYY-MM-DD

			s.Then(`it returns max index plus 2 indexed time suffixed name`, func(t *testcase.T) {
				got, err := subject(t)

				if err != nil {
					t.Errorf("got err = %v, want err = %v", err, nil)
				}

				filename := t.I(`currentName`).(string)
				fileNameExt := filepathFullExt(filename)
				fileNameBase := filepathStripFullExt(filename)

				format := "2006-02-01" // Default format

				want := fmt.Sprintf("%s_%s.1%s", fileNameBase, testTime.Format(format), fileNameExt)

				if got != want {
					t.Errorf("got = %v, want = %v", got, want)
				}
			})

			s.Then(`it moves the existing file to max index plus 1 indexed time suffixed name`, func(t *testcase.T) {
				_, err := subject(t)

				if err != nil {
					t.Errorf("got err = %v, want err = %v", err, nil)
				}

				filename := t.I(`currentName`).(string)
				fileNameExt := filepathFullExt(filename)
				fileNameBase := filepathStripFullExt(filename)

				format := "2006-02-01" // Default format

				existingFileNewName := fmt.Sprintf("%s_%s.0%s", fileNameBase, testTime.Format(format), fileNameExt)

				if _, err := os.Stat(existingFileNewName); err != nil {
					t.Errorf("existing file not moved to correct name")
				}
			})
		})

		s.When(`directory of file already contain multiple indexed file with same time suffixed name`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				testDir := t.I(`testDir`).(string)

				// application_2020-01-01.0.log
				filename := filepath.Join(testDir, fmt.Sprintf("application_2020-01-01.0.log"))
				file1, err := os.Create(filename)
				if err != nil {
					t.Fatalf("os create file: %v", err)
				}
				defer func() { _ = file1.Close() }()

				// application_2020-01-01.1.log
				filename2 := filepath.Join(testDir, fmt.Sprintf("application_2020-01-01.1.log"))
				file2, err := os.Create(filename2)
				if err != nil {
					t.Fatalf("os create file: %v", err)
				}
				defer func() { _ = file2.Close() }()

				// application_2020-01-01.1.log
				filename3 := filepath.Join(testDir, fmt.Sprintf("application_2020-01-01.2.log"))
				file3, err := os.Create(filename3)
				if err != nil {
					t.Fatalf("os create file: %v", err)
				}
				defer func() { _ = file3.Close() }()
			})

			s.Let(`currentName`, func(t *testcase.T) interface{} {
				testDir := t.I(`testDir`).(string)
				filename := filepath.Join(testDir, "application.log")
				file, err := os.Create(filename)
				if err != nil {
					t.Fatalf("os create file: %v", err)
				}
				defer func() { _ = file.Close() }()
				if err := os.Chtimes(filename, testTime, testTime); err != nil {
					t.Fatalf("os chtimes: %v", err)
				}
				return filename
			})

			s.Let(`format`, func(t *testcase.T) interface{} { return "" }) // Default format YYYY-MM-DD

			s.Then(`it returns max index plus 2 indexed time suffixed name`, func(t *testcase.T) {
				got, err := subject(t)

				if err != nil {
					t.Errorf("got err = %v, want err = %v", err, nil)
				}

				filename := t.I(`currentName`).(string)
				fileNameExt := filepathFullExt(filename)
				fileNameBase := filepathStripFullExt(filename)

				format := "2006-02-01" // Default format

				want := fmt.Sprintf("%s_%s.3%s", fileNameBase, testTime.Format(format), fileNameExt)

				if got != want {
					t.Errorf("got = %v, want = %v", got, want)
				}
			})
		})
	})
}

type noopFileTrigger struct {
}

func (t noopFileTrigger) Trigger(_ *os.File, _ []byte) (bool, error) {
	return false, nil
}

type noopFileRotator struct {
}

func (r noopFileRotator) Rotate(file *os.File) (*os.File, error) {
	return file, nil
}

type testNameGenerator struct {
}

func (r testNameGenerator) Name(current string) (string, error) {
	return fmt.Sprintf("%s-new", current), nil
}

type spyTransformer struct {
	FileTransformer FileTransformer
	called          bool
}

func (t *spyTransformer) Transform(path string) (string, error) {
	t.called = true
	return t.FileTransformer.Transform(path)
}

type noopTransformer struct {
}

func (t noopTransformer) Transform(path string) (string, error) {
	return path, nil
}

const errTransform = testErr("transform err")

type faultyTransformer struct {
}

func (t faultyTransformer) Transform(_ string) (string, error) {
	return "", errTransform
}
