package fs

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bool64/ctxd"
	"github.com/nhatthm/aferocopy"
	fsCtx "github.com/nhatthm/plugin-registry/context"
	"github.com/nhatthm/plugin-registry/installer"
	"github.com/nhatthm/plugin-registry/plugin"
	"github.com/spf13/afero"
)

var (
	// ErrPluginNotDir indicates that the plugin is not a directory.
	ErrPluginNotDir = errors.New("plugin is not a directory")
	// ErrPluginIsDir indicates that the plugin is a directory.
	ErrPluginIsDir = errors.New("plugin is a directory")
	// ErrIllegalFilePath indicates that the file path is illegal.
	ErrIllegalFilePath = errors.New("illegal file path")
)

func init() { // nolint: gochecknoinits
	installer.Register("fs", isFsPlugin, func(fs afero.Fs) installer.Installer {
		return NewFsInstaller(fs)
	})
}

// Installer is a file system installer.
type Installer struct {
	fs afero.Fs
}

// Install installs the plugin.
func (i *Installer) Install(ctx context.Context, dest, path string) (*plugin.Plugin, error) {
	path, p, err := parseFsPlugin(i.fs, path)
	if err != nil {
		return nil, ctxd.WrapError(ctx, err, "could not parse plugin path", "path", path)
	}

	if err := installFs(i.fs, dest, path, p); err != nil {
		return nil, ctxd.WrapError(ctx, err, "could not install plugin", "path", path)
	}

	return p, nil
}

// NewFsInstaller creates a new filesystem installer.
func NewFsInstaller(fs afero.Fs) *Installer {
	i := &Installer{
		fs: fs,
	}

	return i
}

func isFsPlugin(ctx context.Context, path string) bool {
	_, _, err := parseFsPlugin(fsCtx.Fs(ctx), path)

	return err == nil
}

func parseFsPlugin(fs afero.Fs, path string) (string, *plugin.Plugin, error) {
	path = filepath.Clean(strings.TrimPrefix(path, "file://"))

	isDir, err := afero.IsDir(fs, path)
	if err != nil {
		return "", nil, err
	}

	if !isDir {
		return "", nil, ErrPluginNotDir
	}

	p, err := plugin.Load(fs, path)
	if err != nil {
		return "", nil, err
	}

	pluginPath := filepath.Join(path, p.Name)

	if _, err := fs.Stat(pluginPath); err != nil {
		return "", nil, fmt.Errorf("%s: %w", pluginPath, err)
	}

	return path, p, nil
}

func installFs(fs afero.Fs, dest, src string, p *plugin.Plugin) error {
	src = filepath.Join(src, p.Name)
	dest = filepath.Join(dest, p.Name)

	if err := recreatePath(fs, dest); err != nil {
		return err
	}

	if isDir, _ := afero.IsDir(fs, src); !isDir { // nolint: errcheck
		dest = filepath.Join(dest, p.Name)
	}

	return aferocopy.Copy(src, dest, aferocopy.Options{
		SrcFs:         fs,
		PreserveTimes: true,
	})
}
