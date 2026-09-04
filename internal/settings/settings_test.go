package settings_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/EgorTarasov/cu/internal/settings"
)

func TestLoad_MissingFileIsUnset(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got, err := settings.Load()

	require.NoError(t, err)
	require.Equal(t, settings.LoginMethodUnset, got.LoginMethod)
}

func TestSaveThenLoad_RoundTrips(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, settings.Save(settings.Settings{LoginMethod: settings.LoginMethodBrowser}))

	got, err := settings.Load()
	require.NoError(t, err)
	require.Equal(t, settings.LoginMethodBrowser, got.LoginMethod)
}

func TestSave_UsesOwnerOnlyPermissions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, settings.Save(settings.Settings{LoginMethod: settings.LoginMethodManual}))

	info, err := os.Stat(filepath.Join(home, ".cu-cli", "config.json"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	dir, err := os.Stat(filepath.Join(home, ".cu-cli"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o700), dir.Mode().Perm())
}

func TestLoad_CorruptFileReportsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".cu-cli"), 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(home, ".cu-cli", "config.json"), []byte("{not json"), 0o600))

	_, err := settings.Load()

	require.Error(t, err)
}
