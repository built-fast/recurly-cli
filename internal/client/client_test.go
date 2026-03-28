package client

import (
	"strings"
	"testing"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

func TestValidateRegion_ValidValues(t *testing.T) {
	for _, region := range []string{"us", "eu", "US", "EU", "Us", "Eu"} {
		if err := ValidateRegion(region); err != nil {
			t.Errorf("ValidateRegion(%q) returned unexpected error: %v", region, err)
		}
	}
}

func TestValidateRegion_InvalidValue(t *testing.T) {
	err := ValidateRegion("asia")
	if err == nil {
		t.Fatal("expected error for invalid region")
	}
	if !strings.Contains(err.Error(), `invalid region "asia"`) {
		t.Errorf("expected error to mention invalid region, got: %v", err)
	}
	if !strings.Contains(err.Error(), "us, eu") {
		t.Errorf("expected error to list valid options, got: %v", err)
	}
}

func TestValidateRegion_EmptyString(t *testing.T) {
	err := ValidateRegion("")
	if err == nil {
		t.Fatal("expected error for empty region")
	}
}

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

func TestNewClient_EURegionCaseInsensitive(t *testing.T) {
	viper.Reset()
	viper.Set("api_key", "test-key")
	viper.Set("region", "EU")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_InvalidRegion_ReturnsError(t *testing.T) {
	viper.Reset()
	viper.Set("api_key", "test-key")
	viper.Set("region", "asia")

	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error for invalid region")
	}
	if !strings.Contains(err.Error(), "invalid region") {
		t.Errorf("expected invalid region error, got: %v", err)
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

func TestNewClient_APIURLOverride(t *testing.T) {
	viper.Reset()
	viper.Set("api_key", "test-key")
	recurly.APIHost = ""
	t.Setenv("RECURLY_API_URL", "http://localhost:4010")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if recurly.APIHost != "http://localhost:4010" {
		t.Errorf("expected APIHost to be %q, got %q", "http://localhost:4010", recurly.APIHost)
	}
	// Clean up
	recurly.APIHost = ""
}

func TestNewClient_APIURLOverride_NotSet(t *testing.T) {
	viper.Reset()
	viper.Set("api_key", "test-key")
	recurly.APIHost = ""
	// Ensure RECURLY_API_URL is not set
	t.Setenv("RECURLY_API_URL", "")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if recurly.APIHost != "" {
		t.Errorf("expected APIHost to remain empty, got %q", recurly.APIHost)
	}
}
