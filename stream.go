package fs

import (
	"io"
	"os"

	"github.com/spf13/afero"
)

func installStream(fs afero.Fs, dest string, src io.Reader, mode os.FileMode) error {
	out, err := fs.OpenFile(dest, os.O_CREATE|os.O_RDWR, mode)
	if err != nil {
		return err
	}
	defer out.Close() // nolint: errcheck

	_, err = io.Copy(out, src)

	return err
}
