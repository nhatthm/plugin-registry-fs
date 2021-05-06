package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/bool64/ctxd"
	"github.com/spf13/afero"
)

func metadataError(err error, path string) error {
	return ctxd.WrapError(context.Background(), err, "plugin has no metadata", "path", path)
}

func statPlugin(fs afero.Fs, path string) (os.FileInfo, error) {
	path = filepath.Clean(strings.TrimPrefix(path, "file://"))

	fi, err := fs.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, ErrPluginIsDir
	}

	return fi, nil
}

func openPluginFile(fs afero.Fs, path string) (os.FileInfo, afero.File, error) {
	fi, err := fs.Stat(path)
	if err != nil {
		return nil, nil, err
	}

	f, err := fs.Open(path)
	if err != nil {
		return nil, nil, err
	}

	return fi, f, err
}

func createPathIfNotExists(fs afero.Fs, path string) error {
	if _, err := fs.Stat(path); err != nil {
		if err := fs.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

func recreatePath(fs afero.Fs, path string) error {
	if err := fs.RemoveAll(path); err != nil {
		return err
	}

	return fs.MkdirAll(path, 0755)
}
