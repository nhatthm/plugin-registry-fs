package fs

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/nhatthm/aferomock"
	fsCtx "github.com/nhatthm/plugin-registry/context"
	"github.com/nhatthm/plugin-registry/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIsGzipPlugin(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		mockFs   aferomock.FsMocker
		path     string
		expected bool
	}{
		{
			scenario: "path does not exist",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(nil, os.ErrNotExist)
			}),
			path: "/tmp",
		},
		{
			scenario: "path is a directory",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
					}), nil)
			}),
			path: "/tmp",
		},
		{
			scenario: "file is not a gzip",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random")
					}), nil)
			}),
			path: "/tmp/random",
		},
		{
			scenario: "metadata does not exist",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random.gz").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.gz")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(nil, os.ErrNotExist)
			}),
			path: "/tmp/random.gz",
		},
		{
			scenario: "success with .gz",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random.gz").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.gz")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(aferomock.NewFileInfo(), nil)
			}),
			path:     "/tmp/random.gz",
			expected: true,
		},
		{
			scenario: "success with .tar.gz",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random.tar.gz").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.tar.gz")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(aferomock.NewFileInfo(), nil)
			}),
			path:     "/tmp/random.tar.gz",
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			ctx := fsCtx.WithFs(context.Background(), tc.mockFs(t))

			assert.Equal(t, tc.expected, isGzipPlugin(ctx, tc.path))
		})
	}
}

func TestParseGzipPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario             string
		mockFs               aferomock.FsMocker
		path                 string
		expectedPath         string
		expectedMetadataPath string
		expectedError        string
	}{
		{
			scenario: "path does not exist",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(nil, os.ErrNotExist)
			}),
			path:          "/tmp",
			expectedError: "file does not exist",
		},
		{
			scenario: "path is a directory",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
					}), nil)
			}),
			path:          "/tmp",
			expectedError: "plugin is a directory",
		},
		{
			scenario: "file is not a gzip",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random")
					}), nil)
			}),
			path:          "/tmp/random",
			expectedError: "plugin is not a gzip",
		},
		{
			scenario: "metadata does not exist",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random.gz").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.gz")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(nil, os.ErrNotExist)
			}),
			path:          "/tmp/random.gz",
			expectedError: "plugin has no metadata: file does not exist",
		},
		{
			scenario: "success with .gz",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random.gz").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.gz")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(aferomock.NewFileInfo(), nil)
			}),
			path:                 "/tmp/random.gz",
			expectedPath:         "/tmp/random.gz",
			expectedMetadataPath: "/tmp",
		},
		{
			scenario: "success with .tar.gz",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random.tar.gz").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.tar.gz")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(aferomock.NewFileInfo(), nil)
			}),
			path:                 "/tmp/random.tar.gz",
			expectedPath:         "/tmp/random.tar.gz",
			expectedMetadataPath: "/tmp",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			path, metadataPath, err := parseGzipPath(tc.mockFs(t), tc.path)

			assert.Equal(t, tc.expectedPath, path)
			assert.Equal(t, tc.expectedMetadataPath, metadataPath)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestGZipInstaller_Install_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockFs        aferomock.FsMocker
		expectedError string
	}{
		{
			scenario: "could not parse gzip path",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/my-plugin.tar.gz").
					Return(nil, os.ErrNotExist)
			}),
			expectedError: `could not parse plugin path: file does not exist`,
		},
		{
			scenario: "could not load metadata",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/my-plugin.tar.gz").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.tar.gz")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(aferomock.NewFileInfo(), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(nil, errors.New("could not open file"))
			}),
			expectedError: `could not read metadata: could not open file`,
		},
		{
			scenario: "could not install",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/my-plugin.tar.gz").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.tar.gz")
					}), nil)

				f := newShadowedFile(".plugin.registry.yaml", "resources/fixtures/gzip/.plugin.registry.yaml")

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(f.Info(), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(f, nil)

				fs.On("Open", "/tmp/my-plugin.tar.gz").
					Return(nil, errors.New("could not open gzip file"))
			}),
			expectedError: `could not install plugin: could not open gzip file`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			i := NewGzipInstaller(tc.mockFs(t))
			result, err := i.Install(context.Background(), "/app/plugins", "/tmp/my-plugin.tar.gz")

			assert.Nil(t, result)
			assert.EqualError(t, err, tc.expectedError)
		})
	}
}

func TestInstallGzip_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockFs        aferomock.FsMocker
		path          string
		expectedError string
	}{
		{
			scenario: "could not stat gzip file",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(nil, os.ErrNotExist)
			}),
			path:          "/tmp",
			expectedError: `file does not exist`,
		},
		{
			scenario: "could not open file",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/my-plugin.tar.gz").
					Return(aferomock.NewFileInfo(), nil)

				fs.On("Open", "/tmp/my-plugin.tar.gz").
					Return(nil, errors.New("open error"))
			}),
			path:          "/tmp/my-plugin.tar.gz",
			expectedError: `open error`,
		},
		{
			scenario: "not a valid gzip file",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/my-plugin.tar.gz").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("Size").Return(10)
					}), nil)

				fs.On("Open", "/tmp/my-plugin.tar.gz").
					Return(newEmptyFile("my-plugin.tar.gz"), nil)
			}),
			path:          "/tmp/my-plugin.tar.gz",
			expectedError: "EOF",
		},
		{
			scenario: "fail to remove dest",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				gzipFile := newShadowedFile("my-plugin.tar.gz", "resources/fixtures/gzip/my-plugin.tar.gz")

				fs.On("Stat", "/tmp/my-plugin.tar.gz").
					Return(gzipFile.Stat())

				fs.On("Open", "/tmp/my-plugin.tar.gz").
					Return(gzipFile, nil)

				fs.On("RemoveAll", mock.Anything).
					Return(errors.New("remove error"))
			}),
			path:          "/tmp/my-plugin.tar.gz",
			expectedError: "remove error",
		},
		{
			scenario: "fail to create dest",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				gzipFile := newShadowedFile("my-plugin.tar.gz", "resources/fixtures/gzip/my-plugin.tar.gz")

				fs.On("Stat", "/tmp/my-plugin.tar.gz").
					Return(gzipFile.Stat())

				fs.On("Open", "/tmp/my-plugin.tar.gz").
					Return(gzipFile, nil)

				fs.On("RemoveAll", mock.Anything).
					Return(nil)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0755)).
					Return(errors.New("mkdir error"))
			}),
			path:          "/tmp/my-plugin.tar.gz",
			expectedError: "mkdir error",
		},
		{
			scenario: "fail to create path",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				gzipFile := newShadowedFile("my-plugin.tar.gz", "resources/fixtures/gzip/my-plugin-empty.tar.gz")

				fs.On("Stat", "/tmp/my-plugin.tar.gz").
					Return(gzipFile.Stat())

				fs.On("Open", "/tmp/my-plugin.tar.gz").
					Return(gzipFile, nil)

				fs.On("RemoveAll", mock.Anything).
					Return(nil)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0755)).Once().
					Return(nil)

				fs.On("Stat", mock.Anything).
					Return(nil, os.ErrNotExist)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0755)).
					Return(errors.New("could not create path"))
			}),
			path:          "/tmp/my-plugin.tar.gz",
			expectedError: "could not create path",
		},
		{
			scenario: "file to open file",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				gzipFile := newShadowedFile("my-plugin.tar.gz", "resources/fixtures/gzip/my-plugin.tar.gz")

				fs.On("Stat", "/tmp/my-plugin.tar.gz").
					Return(gzipFile.Stat())

				fs.On("Open", "/tmp/my-plugin.tar.gz").Once().
					Return(gzipFile, nil)

				fs.On("RemoveAll", mock.Anything).
					Return(nil)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0755)).Once().
					Return(nil)

				fs.On("Stat", mock.Anything).
					Return(nil, os.ErrNotExist)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0755)).
					Return(nil)

				fs.On("OpenFile", mock.Anything, mock.Anything, mock.Anything).Once().
					Return(nil, errors.New("could not open file"))
			}),
			path:          "/tmp/my-plugin.tar.gz",
			expectedError: "could not open file",
		},
		{
			scenario: "zipslip",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				gzipFile := newShadowedFile("my-plugin.tar.gz", "resources/fixtures/gzip/my-plugin-zipslip.tar.gz")

				fs.On("Stat", "/tmp/my-plugin.tar.gz").
					Return(gzipFile.Stat())

				fs.On("Open", "/tmp/my-plugin.tar.gz").Once().
					Return(gzipFile, nil)

				fs.On("RemoveAll", mock.Anything).
					Return(nil)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0755)).Once().
					Return(nil)
			}),
			path:          "/tmp/my-plugin.tar.gz",
			expectedError: "/tmp/evil.sh: illegal file path",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			dest := t.TempDir()

			fs := tc.mockFs(t)
			p := plugin.Plugin{Name: "my-plugin"}
			err := installGzip(fs, dest, p, tc.path)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
