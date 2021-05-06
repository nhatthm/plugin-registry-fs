package fs

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	fsCtx "github.com/nhatthm/plugin-registry/context"
	"github.com/nhatthm/plugin-registry/installer"
	"github.com/nhatthm/plugin-registry/plugin"
	"github.com/spf13/afero"
)

// ErrPluginNotGzip indicates that the plugin is not a zip.
var ErrPluginNotGzip = errors.New("plugin is not a gzip")

func init() { // nolint: gochecknoinits
	installer.Register("gzip", isGzipPlugin, func(fs afero.Fs) installer.Installer {
		return NewGzipInstaller(fs)
	})
}

// NewGzipInstaller creates a new filesystem installer.
func NewGzipInstaller(fs afero.Fs) *ArchiveInstaller {
	i := &ArchiveInstaller{
		fs: fs,

		parseURL: parseGzipPath,
		install:  installGzip,
	}

	return i
}

func isGzipPlugin(ctx context.Context, path string) bool {
	_, _, err := parseGzipPath(fsCtx.Fs(ctx), path)

	return err == nil
}

func parseGzipPath(fs afero.Fs, path string) (string, string, error) {
	fi, err := statPlugin(fs, path)
	if err != nil {
		return "", "", err
	}

	if filepath.Ext(fi.Name()) != ".gz" {
		return "", "", ErrPluginNotGzip
	}

	metadataPath := filepath.Dir(path)
	metadataFile := filepath.Join(metadataPath, plugin.MetadataFile)

	if _, err := fs.Stat(metadataFile); err != nil {
		return "", "", metadataError(err, metadataFile)
	}

	return path, metadataPath, nil
}

func installGzip(fs afero.Fs, dst string, p plugin.Plugin, tarFile string) error {
	fi, r, err := openPluginFile(fs, tarFile)
	if err != nil {
		return err
	}
	defer r.Close() // nolint: errcheck

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close() // nolint: errcheck

	pluginDir := fmt.Sprintf("%s%c", p.Name, os.PathSeparator)
	dst = filepath.Join(dst, p.Name)

	if err := recreatePath(fs, dst); err != nil {
		return err
	}

	if strings.HasSuffix(tarFile, ".tar.gz") {
		return extractTar(fs, dst, pluginDir, tar.NewReader(gzr))
	}

	return installStream(fs, filepath.Join(dst, p.Name), gzr, fi.Mode())
}

func extractTar(fs afero.Fs, dst, pluginDir string, tr *tar.Reader) error {
	for {
		header, err := tr.Next()

		switch {
		case errors.Is(err, io.EOF):
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		path := filepath.Join(dst, strings.TrimPrefix(header.Name, pluginDir))

		if !strings.HasPrefix(path, dst) {
			return fmt.Errorf("%s: %w", path, ErrIllegalFilePath)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := createPathIfNotExists(fs, path); err != nil {
				return err
			}

		case tar.TypeReg:
			if err = installStream(fs, path, tr, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}
}
