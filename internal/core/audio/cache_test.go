package audio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isolateCacheDir points the cache path at a fresh temp dir and returns the expected cache file location.
func isolateCacheDir(t *testing.T) string {
	t.Helper()

	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	return filepath.Join(configRoot, "wire_d", "cache.json")
}

func TestGetCachePath(t *testing.T) {
	cacheFile := isolateCacheDir(t)

	got, err := getCachePath()
	require.NoError(t, err)
	assert.Equal(t, cacheFile, got)
}

func TestLoadCacheMissingFile(t *testing.T) {
	isolateCacheDir(t)

	cache, err := LoadCache()
	require.NoError(t, err)
	assert.Nil(t, cache, "missing cache file should read as nil, not an error")
}

func TestCacheRoundTrip(t *testing.T) {
	cacheFile := isolateCacheDir(t)
	require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o755))

	want := map[string]*AudioFile{
		"/music/boa/fool.flac": {
			Path:      "/music/boa/fool.flac",
			Title:     "Fool",
			Artist:    "bôa",
			Album:     "Twilight",
			Length:    245,
			SizeBytes: 1024,
		},
		"/music/other.mp3": {Path: "/music/other.mp3", Title: "Other", SizeBytes: 8},
	}

	require.NoError(t, WriteCache(want))

	got, err := LoadCache()
	require.NoError(t, err)
	require.Len(t, got, len(want))

	for path, wantFile := range want {
		gotFile, ok := got[path]
		require.True(t, ok, "cache is missing key %q", path)
		assert.Equal(t, wantFile, gotFile, "round-tripped AudioFile mismatch for %q", path)
	}
}

func TestWriteCacheReplacesPreviousContent(t *testing.T) {
	cacheFile := isolateCacheDir(t)
	require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o755))

	require.NoError(t, WriteCache(map[string]*AudioFile{
		"/old.flac": {Path: "/old.flac", Title: "Old"},
	}))

	require.NoError(t, WriteCache(map[string]*AudioFile{
		"/new.flac": {Path: "/new.flac", Title: "New"},
	}))

	got, err := LoadCache()
	require.NoError(t, err)

	require.Len(t, got, 1)
	assert.Contains(t, got, "/new.flac")
	assert.NotContains(t, got, "/old.flac")
}

func TestWriteCacheDoesNotCreateDirs(t *testing.T) {
	isolateCacheDir(t)
	require.Error(t, WriteCache(map[string]*AudioFile{}))
}

func TestLoadCacheCorruptFile(t *testing.T) {
	cacheFile := isolateCacheDir(t)

	require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o755))
	require.NoError(t, os.WriteFile(cacheFile, []byte("this is not json"), 0o644))

	cache, err := LoadCache()
	require.Error(t, err)
	assert.Nil(t, cache)
	assert.True(
		t,
		strings.Contains(err.Error(), "[audio:LoadCache]"),
		"error should carry the [audio:LoadCache] prefix, got %q",
		err.Error(),
	)
}

func TestLoadCacheEmptyFile(t *testing.T) {
	cacheFile := isolateCacheDir(t)

	require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o755))
	require.NoError(t, os.WriteFile(cacheFile, []byte{}, 0o644))

	cache, err := LoadCache()
	require.Error(t, err, "an empty cache file is invalid JSON and must surface as an error, not a silent miss")
	assert.Nil(t, cache)
}

func TestLoadCacheWrongShapeJSON(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{name: "json array top level", payload: `[1,2,3]`},
		{name: "string top level", payload: `"hello"`},
		{name: "number values instead of objects", payload: `{"/a.flac": 42}`},
		{name: "wrong field types", payload: `{"/a.flac": {"Path": 7}}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cacheFile := isolateCacheDir(t)

			require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o755))
			require.NoError(t, os.WriteFile(cacheFile, []byte(test.payload), 0o644))

			cache, err := LoadCache()
			require.Error(t, err, "structurally wrong cache content must surface as an error")
			assert.Nil(t, cache)
		})
	}
}

func TestLoadCacheNullEntriesAreDropped(t *testing.T) {
	cacheFile := isolateCacheDir(t)

	require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o755))
	require.NoError(
		t,
		os.WriteFile(cacheFile, []byte(`{"/gone.flac": null, "/kept.flac": {"Path": "/kept.flac"}}`), 0o644),
	)

	cache, err := LoadCache()
	require.NoError(t, err)

	require.Len(t, cache, 1)
	assert.NotContains(t, cache, "/gone.flac", "a JSON null entry should be dropped at the read boundary")
	require.NotNil(t, cache["/kept.flac"])
}

func TestLoadCachePermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permission bits")
	}

	cacheFile := isolateCacheDir(t)

	require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o755))
	require.NoError(t, os.WriteFile(cacheFile, []byte(`{}`), 0o000))

	cache, err := LoadCache()
	require.Error(t, err)
	assert.Nil(t, cache)
	assert.True(
		t,
		strings.Contains(err.Error(), "[audio:LoadCache]"),
		"error should carry the [audio:LoadCache] prefix, got %q",
		err.Error(),
	)
}

func TestWriteCachePermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permission bits")
	}

	cacheFile := isolateCacheDir(t)

	// The cache dir exists but is read-only, so writing the cache file fails at the OS level.
	require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o500))
	require.Error(t, WriteCache(map[string]*AudioFile{"/a.flac": {Path: "/a.flac"}}))
}

func TestWriteCacheExistingFileIsDirectory(t *testing.T) {
	cacheFile := isolateCacheDir(t)

	require.NoError(t, os.MkdirAll(cacheFile, 0o755))
	require.Error(t, WriteCache(map[string]*AudioFile{"/a.flac": {Path: "/a.flac"}}))
}
