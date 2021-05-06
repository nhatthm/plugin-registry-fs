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

func TestZipInstaller_Install_Success(t *testing.T) {
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
			scenario: "with my-plugin/",
			file:     "resources/fixtures/zip/my-plugin.zip",
		},
		{
			scenario: "without my-plugin/",
			file:     "resources/fixtures/zip/my-plugin-no-parent.zip",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			dest := t.TempDir()

			fs := afero.NewOsFs()
			i := NewZipInstaller(fs)

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

func TestInstallZip_Success(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		file     string
	}{
		{
			scenario: "with my-plugin/",
			file:     "resources/fixtures/zip/my-plugin.zip",
		},
		{
			scenario: "without my-plugin/",
			file:     "resources/fixtures/zip/my-plugin-no-parent.zip",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			dest := t.TempDir()

			fs := afero.NewOsFs()
			p := plugin.Plugin{Name: "my-plugin"}

			err := installZip(fs, dest, p, tc.file)
			require.NoError(t, err)

			file := filepath.Join(dest, p.Name, p.Name)

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
