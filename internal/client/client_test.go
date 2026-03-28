package client

import (
	"testing"

	"github.com/spf13/viper"
)

func TestNewClient_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()

	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error when no API key configured")
	}
	expected := "API key not configured. Run 'recurly configure' or set RECURLY_API_KEY."
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestNewClient_WithAPIKey_DefaultRegion(t *testing.T) {
	viper.Reset()
	viper.Set("api_key", "test-key")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithAPIKey_USRegion(t *testing.T) {
	viper.Reset()
	viper.Set("api_key", "test-key")
	viper.Set("region", "us")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithAPIKey_EURegion(t *testing.T) {
	viper.Reset()
	viper.Set("api_key", "test-key")
	viper.Set("region", "eu")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_EnvVarAPIKey(t *testing.T) {
	viper.Reset()
	_ = viper.BindEnv("api_key", "RECURLY_API_KEY")
	t.Setenv("RECURLY_API_KEY", "env-key")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_EnvVarRegion(t *testing.T) {
	viper.Reset()
	_ = viper.BindEnv("api_key", "RECURLY_API_KEY")
	_ = viper.BindEnv("region", "RECURLY_REGION")
	t.Setenv("RECURLY_API_KEY", "env-key")
	t.Setenv("RECURLY_REGION", "eu")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_FlagOverridesEnv(t *testing.T) {
	viper.Reset()
	_ = viper.BindEnv("api_key", "RECURLY_API_KEY")
	t.Setenv("RECURLY_API_KEY", "env-key")
	viper.Set("api_key", "flag-key")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}
