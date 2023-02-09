package fs

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	fsCtx "github.com/nhatthm/plugin-registry/context"
	"github.com/nhatthm/plugin-registry/installer"
	"github.com/nhatthm/plugin-registry/plugin"
	"github.com/spf13/afero"
)

// ErrPluginNotZip indicates that the plugin is not a zip.
var ErrPluginNotZip = errors.New("plugin is not a zip")

func init() { //nolint: gochecknoinits
	installer.Register("zip", isZipPlugin, func(fs afero.Fs) installer.Installer {
		return NewZipInstaller(fs)
	})
}

// NewZipInstaller creates a new filesystem installer.
func NewZipInstaller(fs afero.Fs) *ArchiveInstaller {
	i := &ArchiveInstaller{
		fs: fs,

		parseURL: parseZipPath,
		install:  installZip,
	}

	return i
}

func isZipPlugin(ctx context.Context, path string) bool { //nolint: contextcheck,nolintlint
	_, _, err := parseZipPath(fsCtx.Fs(ctx), path) //nolint: contextcheck,nolintlint

	return err == nil
}

func parseZipPath(fs afero.Fs, path string) (string, string, error) { //nolint: contextcheck,nolintlint
	fi, err := statPlugin(fs, path)
	if err != nil {
		return "", "", err
	}

	if filepath.Ext(fi.Name()) != ".zip" {
		return "", "", ErrPluginNotZip
	}

	metadataPath := filepath.Dir(path)
	metadataFile := filepath.Join(metadataPath, plugin.MetadataFile)

	if _, err := fs.Stat(metadataFile); err != nil {
		return "", "", metadataError(err, metadataFile)
	}

	return path, metadataPath, nil
}

func installZip(fs afero.Fs, dst string, p plugin.Plugin, zipFile string) error {
	fi, r, err := openPluginFile(fs, zipFile)
	if err != nil {
		return err
	}
	defer r.Close() //nolint: errcheck

	zr, err := zip.NewReader(r, fi.Size())
	if err != nil {
		return err
	}

	pluginDir := fmt.Sprintf("%s%c", p.Name, os.PathSeparator)
	dst = filepath.Join(dst, p.Name)

	if err := recreatePath(fs, dst); err != nil {
		return err
	}

	return extractZip(fs, dst, pluginDir, zr)
}

func extractZip(fs afero.Fs, dst, pluginDir string, zr *zip.Reader) error {
	for _, f := range zr.File {
		path := filepath.Join(dst, strings.TrimPrefix(f.Name, pluginDir))

		if !strings.HasPrefix(path, dst) {
			return fmt.Errorf("%s: %w", path, ErrIllegalFilePath)
		}

		if f.FileInfo().IsDir() {
			if err := createPathIfNotExists(fs, path); err != nil {
				return err
			}

			continue
		}

		src, err := f.Open()
		if err != nil {
			return err
		}

		err = installStream(fs, path, src, f.FileInfo().Mode())
		_ = src.Close() //nolint: errcheck

		if err != nil {
			return err
		}
	}

	return nil
}
