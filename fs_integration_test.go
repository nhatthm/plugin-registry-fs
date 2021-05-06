package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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
		file     string
	}{
		{
			scenario: "folder",
			file:     "resources/fixtures/fs/folder",
		},
		{
			scenario: "file",
			file:     "resources/fixtures/fs/file",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			dest := t.TempDir()

			fs := afero.NewOsFs()
			i := NewFsInstaller(fs)

			result, err := i.Install(context.Background(), dest, tc.file)
			require.NoError(t, err)

			assert.Equal(t, expectedResult, result)

			file := filepath.Join(dest, result.Name, result.Name)

			info, err := fs.Stat(file)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0755), info.Mode())

			data, err := afero.ReadFile(fs, file)
			require.NoError(t, err)

			expected := "#!/bin/bash\n"

			assert.Equal(t, expected, string(data))
		})
	}
}
