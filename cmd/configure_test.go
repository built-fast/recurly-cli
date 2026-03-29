package cmd

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestConfigure_NewConfig(t *testing.T) {
	// Cannot use t.Parallel() — uses t.Setenv which is incompatible
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	out := new(bytes.Buffer)
	p := &configPrompter{
		reader: bufio.NewReader(strings.NewReader("us\nmysite\n")),
		writer: out,
		readPassword: func() (string, error) {
			return "test-api-key-12345", nil
		},
	}

	if err := runConfigure(p); err != nil {
		t.Fatalf("runConfigure() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "API Key:") {
		t.Errorf("expected API Key prompt, got: %s", output)
	}
	if !strings.Contains(output, "Region (us/eu) [us]:") {
		t.Errorf("expected Region prompt with us default, got: %s", output)
	}
	if !strings.Contains(output, "Site subdomain") {
		t.Errorf("expected Site subdomain prompt, got: %s", output)
	}
	if !strings.Contains(output, "Configuration saved to") {
		t.Errorf("expected confirmation message, got: %s", output)
	}

	// Verify config was written
	v := viper.New()
	v.SetConfigFile(filepath.Join(tmp, "recurly", "config.toml"))
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("reading config: %v", err)
	}
	if got := v.GetString("api_key"); got != "test-api-key-12345" {
		t.Errorf("api_key = %q, want %q", got, "test-api-key-12345")
	}
	if got := v.GetString("region"); got != "us" {
		t.Errorf("region = %q, want %q", got, "us")
	}
	if got := v.GetString("site"); got != "mysite" {
		t.Errorf("site = %q, want %q", got, "mysite")
	}
}

func TestConfigure_ExistingConfigShowsDefaults(t *testing.T) {
	// Cannot use t.Parallel() — uses t.Setenv which is incompatible
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	// Write existing config
	dir := filepath.Join(tmp, "recurly")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"),
		[]byte("api_key = \"existing-key-abcd\"\nregion = \"eu\"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	out := new(bytes.Buffer)
	p := &configPrompter{
		reader: bufio.NewReader(strings.NewReader("\n\n")), // Enter to keep existing region, Enter for site
		writer: out,
		readPassword: func() (string, error) {
			return "", nil // Enter to keep existing API key
		},
	}

	if err := runConfigure(p); err != nil {
		t.Fatalf("runConfigure() error: %v", err)
	}

	output := out.String()
	// Should show masked existing key in prompt
	if !strings.Contains(output, "API Key [****abcd]:") {
		t.Errorf("expected masked API key in prompt, got: %s", output)
	}
	// Should show existing region as default
	if !strings.Contains(output, "Region (us/eu) [eu]:") {
		t.Errorf("expected eu as default region, got: %s", output)
	}

	// Verify existing values preserved
	v := viper.New()
	v.SetConfigFile(filepath.Join(tmp, "recurly", "config.toml"))
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if got := v.GetString("api_key"); got != "existing-key-abcd" {
		t.Errorf("api_key = %q, want %q", got, "existing-key-abcd")
	}
	if got := v.GetString("region"); got != "eu" {
		t.Errorf("region = %q, want %q", got, "eu")
	}
}

func TestConfigure_InvalidRegionRePrompts(t *testing.T) {
	// Cannot use t.Parallel() — uses t.Setenv which is incompatible
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	out := new(bytes.Buffer)
	p := &configPrompter{
		reader: bufio.NewReader(strings.NewReader("invalid\neu\n\n")),
		writer: out,
		readPassword: func() (string, error) {
			return "test-key", nil
		},
	}

	if err := runConfigure(p); err != nil {
		t.Fatalf("runConfigure() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `invalid region "invalid"`) {
		t.Errorf("expected error for invalid region, got: %s", output)
	}
	// Should have prompted twice for region
	if strings.Count(output, "Region (us/eu)") != 2 {
		t.Errorf("expected two region prompts, got: %s", output)
	}

	// Verify eu was used
	v := viper.New()
	v.SetConfigFile(filepath.Join(tmp, "recurly", "config.toml"))
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if got := v.GetString("region"); got != "eu" {
		t.Errorf("region = %q, want %q", got, "eu")
	}
}

func TestConfigure_EmptyAPIKeyRequired(t *testing.T) {
	// Cannot use t.Parallel() — uses t.Setenv which is incompatible
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	out := new(bytes.Buffer)
	p := &configPrompter{
		reader: bufio.NewReader(strings.NewReader("us\n")),
		writer: out,
		readPassword: func() (string, error) {
			return "", nil
		},
	}

	err := runConfigure(p)
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
	if !strings.Contains(err.Error(), "API key is required") {
		t.Errorf("expected 'API key is required' error, got: %v", err)
	}
}

func TestConfigure_RegionDefaultEnter(t *testing.T) {
	// Cannot use t.Parallel() — uses t.Setenv which is incompatible
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	out := new(bytes.Buffer)
	p := &configPrompter{
		reader: bufio.NewReader(strings.NewReader("\n\n")), // Enter for region, Enter for site
		writer: out,
		readPassword: func() (string, error) {
			return "my-key", nil
		},
	}

	if err := runConfigure(p); err != nil {
		t.Fatalf("runConfigure() error: %v", err)
	}

	// Verify region defaults to "us"
	v := viper.New()
	v.SetConfigFile(filepath.Join(tmp, "recurly", "config.toml"))
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if got := v.GetString("region"); got != "us" {
		t.Errorf("region = %q, want %q", got, "us")
	}
}

func TestConfigure_ConfirmationMessage(t *testing.T) {
	// Cannot use t.Parallel() — uses t.Setenv which is incompatible
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	out := new(bytes.Buffer)
	p := &configPrompter{
		reader: bufio.NewReader(strings.NewReader("eu\n\n")),
		writer: out,
		readPassword: func() (string, error) {
			return "my-key", nil
		},
	}

	if err := runConfigure(p); err != nil {
		t.Fatalf("runConfigure() error: %v", err)
	}

	expected := "Configuration saved to " + filepath.Join(tmp, "recurly", "config.toml")
	if !strings.Contains(out.String(), expected) {
		t.Errorf("expected confirmation %q, got: %s", expected, out.String())
	}
}

func TestMaskKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"abc", "****"},
		{"abcd", "****"},
		{"abcde", "****bcde"},
		{"test-api-key-12345", "****2345"},
	}
	for _, tt := range tests {
		got := maskKey(tt.input)
		if got != tt.want {
			t.Errorf("maskKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
