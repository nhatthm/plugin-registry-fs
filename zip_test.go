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

func TestIsZipPlugin(t *testing.T) {
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
			scenario: "file is not a zip",
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
				fs.On("Stat", "/tmp/random.zip").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.zip")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(nil, os.ErrNotExist)
			}),
			path: "/tmp/random.zip",
		},
		{
			scenario: "success",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random.zip").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.zip")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(aferomock.NewFileInfo(), nil)
			}),
			path:     "/tmp/random.zip",
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			ctx := fsCtx.WithFs(context.Background(), tc.mockFs(t))

			assert.Equal(t, tc.expected, isZipPlugin(ctx, tc.path))
		})
	}
}

func TestParseZipPath(t *testing.T) {
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
			scenario: "file is not a zip",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random")
					}), nil)
			}),
			path:          "/tmp/random",
			expectedError: "plugin is not a zip",
		},
		{
			scenario: "metadata does not exist",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random.zip").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.zip")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(nil, os.ErrNotExist)
			}),
			path:          "/tmp/random.zip",
			expectedError: "plugin has no metadata: file does not exist",
		},
		{
			scenario: "success",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/random.zip").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.zip")
					}), nil)

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(aferomock.NewFileInfo(), nil)
			}),
			path:                 "/tmp/random.zip",
			expectedPath:         "/tmp/random.zip",
			expectedMetadataPath: "/tmp",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			path, metadataPath, err := parseZipPath(tc.mockFs(t), tc.path)

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

func TestZipInstaller_Install_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockFs        aferomock.FsMocker
		expectedError string
	}{
		{
			scenario: "could not parse zip path",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(nil, os.ErrNotExist)
			}),
			expectedError: `could not parse plugin path: file does not exist`,
		},
		{
			scenario: "could not load metadata",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.zip")
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
				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
						i.On("Name").Return("random.zip")
					}), nil)

				f := newShadowedFile(".plugin.registry.yaml", "resources/fixtures/zip/.plugin.registry.yaml")

				fs.On("Stat", "/tmp/.plugin.registry.yaml").
					Return(f.Info(), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(f, nil)

				fs.On("Open", "/tmp/my-plugin.zip").
					Return(nil, errors.New("could not open zip file"))
			}),
			expectedError: `could not install plugin: could not open zip file`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			i := NewZipInstaller(tc.mockFs(t))
			result, err := i.Install(context.Background(), "/app/plugins", "/tmp/my-plugin.zip")

			assert.Nil(t, result)
			assert.EqualError(t, err, tc.expectedError)
		})
	}
}

func TestInstallZip_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockFs        aferomock.FsMocker
		path          string
		expectedError string
	}{
		{
			scenario: "could not stat zip file",
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
				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(aferomock.NewFileInfo(), nil)

				fs.On("Open", "/tmp/my-plugin.zip").
					Return(nil, errors.New("open error"))
			}),
			path:          "/tmp/my-plugin.zip",
			expectedError: `open error`,
		},
		{
			scenario: "could not open zip",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("Size").Return(-1)
					}), nil)

				fs.On("Open", "/tmp/my-plugin.zip").
					Return(newEmptyFile("my-plugin.zip"), nil)
			}),
			path:          "/tmp/my-plugin.zip",
			expectedError: `zip: size cannot be negative`,
		},
		{
			scenario: "not a valid zip file",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("Size").Return(10)
					}), nil)

				fs.On("Open", "/tmp/my-plugin.zip").
					Return(newEmptyFile("my-plugin.zip"), nil)
			}),
			path:          "/tmp/my-plugin.zip",
			expectedError: "zip: not a valid zip file",
		},
		{
			scenario: "fail to remove dest",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				zipFile := newShadowedFile("my-plugin.zip", "resources/fixtures/zip/my-plugin.zip")

				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(zipFile.Stat())

				fs.On("Open", "/tmp/my-plugin.zip").
					Return(zipFile, nil)

				fs.On("RemoveAll", mock.Anything).
					Return(errors.New("remove error"))
			}),
			path:          "/tmp/my-plugin.zip",
			expectedError: "remove error",
		},
		{
			scenario: "fail to create dest",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				zipFile := newShadowedFile("my-plugin.zip", "resources/fixtures/zip/my-plugin.zip")

				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(zipFile.Stat())

				fs.On("Open", "/tmp/my-plugin.zip").
					Return(zipFile, nil)

				fs.On("RemoveAll", mock.Anything).
					Return(nil)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0o755)).
					Return(errors.New("mkdir error"))
			}),
			path:          "/tmp/my-plugin.zip",
			expectedError: "mkdir error",
		},
		{
			scenario: "fail to create path",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				zipFile := newShadowedFile("my-plugin.zip", "resources/fixtures/zip/my-plugin-empty.zip")

				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(zipFile.Stat())

				fs.On("Open", "/tmp/my-plugin.zip").
					Return(zipFile, nil)

				fs.On("RemoveAll", mock.Anything).
					Return(nil)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0o755)).Once().
					Return(nil)

				fs.On("Stat", mock.Anything).
					Return(nil, os.ErrNotExist)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0o755)).
					Return(errors.New("could not create path"))
			}),
			path:          "/tmp/my-plugin.zip",
			expectedError: "could not create path",
		},
		{
			scenario: "file to open file",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				zipFile := newShadowedFile("my-plugin.zip", "resources/fixtures/zip/my-plugin.zip")

				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(zipFile.Stat())

				fs.On("Open", "/tmp/my-plugin.zip").Once().
					Return(zipFile, nil)

				fs.On("RemoveAll", mock.Anything).
					Return(nil)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0o755)).Once().
					Return(nil)

				fs.On("OpenFile", mock.Anything, mock.Anything, mock.Anything).Once().
					Return(nil, errors.New("could not open file"))
			}),
			path:          "/tmp/my-plugin.zip",
			expectedError: "could not open file",
		},
		{
			scenario: "zipslip",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				zipFile := newShadowedFile("my-plugin.zip", "resources/fixtures/zip/my-plugin-zipslip.zip")

				fs.On("Stat", "/tmp/my-plugin.zip").
					Return(zipFile.Stat())

				fs.On("Open", "/tmp/my-plugin.zip").Once().
					Return(zipFile, nil)

				fs.On("RemoveAll", mock.Anything).
					Return(nil)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0o755)).Once().
					Return(nil)
			}),
			path:          "/tmp/my-plugin.zip",
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
			err := installZip(fs, dest, p, tc.path)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
