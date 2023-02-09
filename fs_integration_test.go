//go:build integration
// +build integration

package fs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	fs "github.com/nhatthm/plugin-registry-fs"
	fsCtx "github.com/nhatthm/plugin-registry/context"
	"github.com/nhatthm/plugin-registry/installer"
	"github.com/nhatthm/plugin-registry/plugin"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFsInstaller_Install_Success(t *testing.T) {
	t.Parallel()

	expectedResult := &plugin.Plugin{
		Name:    "my-plugin",
		Enabled: true,
		Hidden:  true,
		Artifacts: plugin.Artifacts{
			plugin.RuntimeArtifactIdentifier(): {
				File: "${name}-${version}-${os}-${arch}.tar.gz",
			},
		},
	}

	testCases := []struct {
		scenario string
		path     string
	}{
		{
			scenario: "folder",
			path:     "resources/fixtures/fs/folder",
		},
		{
			scenario: "file",
			path:     "resources/fixtures/fs/file",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			dest := t.TempDir()

			osFs := afero.NewOsFs()
			ctx := fsCtx.WithFs(context.Background(), osFs)
			i, err := installer.Find(ctx, tc.path)
			require.NoError(t, err)
			assert.IsType(t, &fs.Installer{}, i)

			result, err := i.Install(context.Background(), dest, tc.path)
			require.NoError(t, err)

			assert.Equal(t, expectedResult, result)

			file := filepath.Join(dest, result.Name, result.Name)

			info, err := osFs.Stat(file)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o755), info.Mode())

			data, err := afero.ReadFile(osFs, file)
			require.NoError(t, err)

			expected := "#!/bin/bash\n"

			assert.Equal(t, expected, string(data))
		})
	}
}
