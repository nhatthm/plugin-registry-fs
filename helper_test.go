package fs

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/nhatthm/aferomock"
	"github.com/spf13/afero"
	"github.com/spf13/afero/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEmptyFile(name string) *mem.File {
	return mem.NewFileHandle(mem.CreateFile(name))
}

func newShadowedFile(name string, source string) *mem.File {
	f := newEmptyFile(name)

	data, err := afero.ReadFile(afero.NewOsFs(), source)
	if err != nil {
		panic(err)
	}

	_, _ = f.Write(data)           // nolint: errcheck
	_, _ = f.Seek(0, io.SeekStart) // nolint: errcheck

	return f
}

func TestMetadataError(t *testing.T) {
	t.Parallel()

	err := metadataError(errors.New("error"), "/tmp")
	expected := `plugin has no metadata: error`

	assert.EqualError(t, err, expected)
}

func TestCreatePathIfNotExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockFs        aferomock.FsMocker
		expectedError string
	}{
		{
			scenario: "directory exists",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(aferomock.NewFileInfo(), nil)
			}),
		},
		{
			scenario: "could not mkdir",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(nil, os.ErrNotExist)

				fs.On("MkdirAll", "/tmp", os.FileMode(0o755)).
					Return(errors.New("mkdir error"))
			}),
			expectedError: "mkdir error",
		},
		{
			scenario: "success",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("Stat", "/tmp").
					Return(nil, os.ErrNotExist)

				fs.On("MkdirAll", "/tmp", os.FileMode(0o755)).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := createPathIfNotExists(tc.mockFs(t), "/tmp")

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRecreatePath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockFs        aferomock.FsMocker
		expectedError string
	}{
		{
			scenario: "could not remove directory",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("RemoveAll", "/tmp").
					Return(errors.New("remove error"))
			}),
			expectedError: "remove error",
		},
		{
			scenario: "success to remove",
			mockFs: aferomock.MockFs(func(fs *aferomock.Fs) {
				fs.On("RemoveAll", "/tmp").
					Return(nil)

				fs.On("MkdirAll", "/tmp", os.FileMode(0o755)).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := recreatePath(tc.mockFs(t), "/tmp")

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestInstallFile_OpenFail(t *testing.T) {
	t.Parallel()

	fs := aferomock.MockFs(func(fs *aferomock.Fs) {
		fs.On("OpenFile", "/tmp/temp.txt", os.O_CREATE|os.O_RDWR, os.FileMode(0o755)).
			Return(nil, errors.New("open error"))
	})(t)

	err := installStream(fs, "/tmp/temp.txt", nil, os.FileMode(0o755))
	expected := `open error`

	assert.EqualError(t, err, expected)
}

func TestInstallFile_Success(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	err := installStream(fs, "/tmp/temp.txt", strings.NewReader("hello world"), os.FileMode(0o755))
	require.NoError(t, err)

	fi, err := fs.Stat("/tmp/temp.txt")
	require.NoError(t, err)

	assert.Equal(t, os.FileMode(0o755), fi.Mode())

	content, err := afero.ReadFile(fs, "/tmp/temp.txt")
	require.NoError(t, err)

	expected := []byte(`hello world`)

	assert.Equal(t, expected, content)
}
