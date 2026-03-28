package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestDir_UsesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	got := Dir()
	want := "/custom/config/recurly"
	if got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

func TestDir_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	got := Dir()
	want := filepath.Join(home, ".config", "recurly")
	if got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

func TestFilePath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	got := FilePath()
	want := "/custom/config/recurly/config.toml"
	if got != want {
		t.Errorf("FilePath() = %q, want %q", got, want)
	}
}

func TestInit_NoConfigFile(t *testing.T) {
	viper.Reset()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}
}

func TestInit_ReadsExistingConfig(t *testing.T) {
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir := filepath.Join(tmp, "recurly")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	content := "api_key = \"test-key\"\nregion = \"eu\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	if err := Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	if got := viper.GetString("api_key"); got != "test-key" {
		t.Errorf("api_key = %q, want %q", got, "test-key")
	}
	if got := viper.GetString("region"); got != "eu" {
		t.Errorf("region = %q, want %q", got, "eu")
	}
}

func TestWrite_CreatesDirectoryAndFile(t *testing.T) {
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := Write("api_key", "my-key"); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// Verify directory permissions
	dir := filepath.Join(tmp, "recurly")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("config dir not created: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0700 {
		t.Errorf("dir permissions = %o, want %o", perm, 0700)
	}

	// Verify file permissions
	path := filepath.Join(dir, "config.toml")
	finfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if perm := finfo.Mode().Perm(); perm != 0600 {
		t.Errorf("file permissions = %o, want %o", perm, 0600)
	}

	// Verify content
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "my-key") {
		t.Errorf("config file does not contain api_key value")
	}
}

func TestWrite_PreservesExistingKeys(t *testing.T) {
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := Write("api_key", "key1"); err != nil {
		t.Fatalf("Write(api_key) error: %v", err)
	}
	if err := Write("region", "eu"); err != nil {
		t.Fatalf("Write(region) error: %v", err)
	}

	// Read back with a fresh viper and verify both keys persist
	v := viper.New()
	v.SetConfigFile(FilePath())
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig() error: %v", err)
	}

	if got := v.GetString("api_key"); got != "key1" {
		t.Errorf("api_key = %q, want %q", got, "key1")
	}
	if got := v.GetString("region"); got != "eu" {
		t.Errorf("region = %q, want %q", got, "eu")
	}
}
