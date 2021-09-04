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

func TestIsFsPlugin(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		mockFs   aferomock.FsMocker
		expected bool
	}{
		{
			scenario: "path does not exist",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(nil, os.ErrNotExist)
			}),
		},
		{
			scenario: "path is a not directory",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
					}), nil)
			}),
		},
		{
			scenario: "metadata does not exist",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("random")
					}), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(nil, os.ErrNotExist)
			}),
		},
		{
			scenario: "plugin not found",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				file := newShadowedFile(".plugin.registry.yaml", "resources/fixtures/fs/file/.plugin.registry.yaml")

				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("random")
					}), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(file, nil)

				fs.On("Stat", "/tmp/my-plugin").
					Return(nil, os.ErrNotExist)
			}),
		},
		{
			scenario: "success",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				file := newShadowedFile(".plugin.registry.yaml", "resources/fixtures/fs/file/.plugin.registry.yaml")

				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("random")
					}), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(file, nil)

				fs.On("Stat", "/tmp/my-plugin").
					Return(aferomock.NewFileInfo(), nil)
			}),
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			ctx := fsCtx.WithFs(context.Background(), tc.mockFs(t))

			assert.Equal(t, tc.expected, isFsPlugin(ctx, "/tmp"))
		})
	}
}

func TestParseFsPlugin(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		mockFs         aferomock.FsMocker
		path           string
		expectedPath   string
		expectedPlugin *plugin.Plugin
		expectedError  string
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
			scenario: "path is a not directory",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(false)
					}), nil)
			}),
			path:          "/tmp",
			expectedError: "plugin is not a directory",
		},
		{
			scenario: "metadata does not exist",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("random")
					}), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(nil, os.ErrNotExist)
			}),
			path:          "/tmp",
			expectedError: "could not read metadata: file does not exist",
		},
		{
			scenario: "plugin not found",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				file := newShadowedFile(".plugin.registry.yaml", "resources/fixtures/fs/file/.plugin.registry.yaml")

				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("random")
					}), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(file, nil)

				fs.On("Stat", "/tmp/my-plugin").
					Return(nil, os.ErrNotExist)
			}),
			path:          "/tmp",
			expectedError: "/tmp/my-plugin: file does not exist",
		},
		{
			scenario: "success",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				file := newShadowedFile(".plugin.registry.yaml", "resources/fixtures/fs/file/.plugin.registry.yaml")

				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("random")
					}), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(file, nil)

				fs.On("Stat", "/tmp/my-plugin").
					Return(aferomock.NewFileInfo(), nil)
			}),
			path:         "/tmp",
			expectedPath: "/tmp",
			expectedPlugin: &plugin.Plugin{
				Name:    "my-plugin",
				Enabled: true,
				Hidden:  true,
				Artifacts: plugin.Artifacts{
					plugin.RuntimeArtifactIdentifier(): {
						File: "${name}-${version}-${os}-${arch}.tar.gz",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			path, p, err := parseFsPlugin(tc.mockFs(t), tc.path)

			assert.Equal(t, tc.expectedPath, path)
			assert.Equal(t, tc.expectedPlugin, p)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestFsInstaller_Install_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockFs        aferomock.FsMocker
		expectedError string
	}{
		{
			scenario: "could not parse path",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(nil, os.ErrNotExist)
			}),
			expectedError: `could not parse plugin path: file does not exist`,
		},
		{
			scenario: "could not load metadata",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("tmp")
					}), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(nil, errors.New("could not open file"))
			}),
			expectedError: `could not parse plugin path: could not read metadata: could not open file`,
		},
		{
			scenario: "fail to remove dest",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				file := newShadowedFile(".plugin.registry.yaml", "resources/fixtures/fs/file/.plugin.registry.yaml")

				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("tmp")
					}), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(file, nil)

				fs.On("Stat", "/tmp/my-plugin").
					Return(aferomock.NewFileInfo(), nil)

				fs.On("RemoveAll", mock.Anything).
					Return(errors.New("remove error"))
			}),
			expectedError: "could not install plugin: remove error",
		},
		{
			scenario: "fail to create dest",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				file := newShadowedFile(".plugin.registry.yaml", "resources/fixtures/fs/file/.plugin.registry.yaml")

				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("tmp")
					}), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(file, nil)

				fs.On("Stat", "/tmp/my-plugin").
					Return(aferomock.NewFileInfo(), nil)

				fs.On("RemoveAll", mock.Anything).
					Return(nil)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0o755)).
					Return(errors.New("mkdir error"))
			}),
			expectedError: "could not install plugin: mkdir error",
		},
		{
			scenario: "could not install",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				file := newShadowedFile(".plugin.registry.yaml", "resources/fixtures/fs/file/.plugin.registry.yaml")

				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("tmp")
					}), nil)

				fs.On("Open", "/tmp/.plugin.registry.yaml").
					Return(file, nil)

				fs.On("Stat", "/tmp/my-plugin").Once().
					Return(aferomock.NewFileInfo(func(i *aferomock.FileInfo) {
						i.On("IsDir").Return(true)
						i.On("Name").Return("my-plugin")
					}), nil)

				fs.On("RemoveAll", mock.Anything).
					Return(nil)

				fs.On("MkdirAll", mock.Anything, os.FileMode(0o755)).
					Return(nil)

				fs.On("Stat", "/tmp/my-plugin").
					Return(nil, os.ErrNotExist)
			}),
			expectedError: `could not install plugin: file does not exist`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			i := NewFsInstaller(tc.mockFs(t))
			result, err := i.Install(context.Background(), "/app/plugins", "/tmp")

			assert.Nil(t, result)
			assert.EqualError(t, err, tc.expectedError)
		})
	}
}
