package browser

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeZip(t *testing.T, entries map[string]string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "a.zip")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	w := zip.NewWriter(f)
	for name, body := range entries {
		e, wErr := w.Create(name)
		require.NoError(t, wErr)
		_, wErr = e.Write([]byte(body))
		require.NoError(t, wErr)
	}
	require.NoError(t, w.Close())
	return path
}

func TestUnzip_ExtractsRegularEntries(t *testing.T) {
	src := writeZip(t, map[string]string{"chrome-linux64/chrome": "binary"})
	dest := t.TempDir()

	require.NoError(t, unzip(src, dest))

	got, err := os.ReadFile(filepath.Join(dest, "chrome-linux64", "chrome"))
	require.NoError(t, err)
	require.Equal(t, "binary", string(got))
}

func TestUnzip_RejectsPathTraversal(t *testing.T) {
	src := writeZip(t, map[string]string{"../escaped": "pwned"})
	dest := t.TempDir()

	err := unzip(src, dest)

	require.Error(t, err)
	require.Contains(t, err.Error(), "за пределами каталога")
	require.NoFileExists(t, filepath.Join(filepath.Dir(dest), "escaped"))
}
