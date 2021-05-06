package fs

import (
	"context"

	"github.com/bool64/ctxd"
	"github.com/nhatthm/plugin-registry/plugin"
	"github.com/spf13/afero"
)

// ArchiveInstaller is an installer for archive file.
type ArchiveInstaller struct {
	fs afero.Fs

	parseURL func(fs afero.Fs, pluginURL string) (path string, metadataPath string, err error)
	install  func(fs afero.Fs, dst string, p plugin.Plugin, archiveFile string) error
}

// Install installs the plugin.
func (i *ArchiveInstaller) Install(ctx context.Context, dest, pluginURL string) (*plugin.Plugin, error) {
	path, metadataPath, err := i.parseURL(i.fs, pluginURL)
	if err != nil {
		return nil, ctxd.WrapError(ctx, err, "could not parse plugin path", "path", pluginURL)
	}

	p, err := plugin.Load(i.fs, metadataPath)
	if err != nil {
		return nil, err
	}

	if err := i.install(i.fs, dest, *p, path); err != nil {
		return nil, ctxd.WrapError(ctx, err, "could not install plugin", "path", path)
	}

	return p, nil
}
