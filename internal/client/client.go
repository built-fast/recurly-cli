package client

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

// isJSONOutput reports whether the current output format is JSON-based.
func isJSONOutput() bool {
	return strings.Contains(viper.GetString("output"), "json")
}

// ClientConfig holds the configuration needed to create a Recurly API client.
// It replaces direct reads from global state (e.g. viper) so that callers can
// inject values explicitly.
type ClientConfig struct {
	// APIKey is the Recurly API key used for authentication.
	APIKey string
	// Region is the Recurly data-center region (e.g. "us" or "eu").
	Region string
	// IsJSON reports whether the current output format is JSON-based.
	// It is used to decide retry/logging behaviour.
	IsJSON func() bool
}

// acceptRewriteTransport wraps an http.RoundTripper and normalizes the Accept
// header to application/json. This is needed when running against mock servers
// (e.g. Prism) that don't understand the SDK's vendor Accept header.
type acceptRewriteTransport struct {
	base http.RoundTripper
}

func (t *acceptRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/json")
	return t.base.RoundTrip(req)
}

// ValidRegions lists the accepted region values.
var ValidRegions = []string{"us", "eu"}

// ValidateRegion checks that region is a valid value (case-insensitive).
// Returns an error listing valid options if the value is invalid.
func ValidateRegion(region string) error {
	lower := strings.ToLower(region)
	for _, v := range ValidRegions {
		if lower == v {
			return nil
		}
	}
	return fmt.Errorf("invalid region %q: valid options are %s", region, strings.Join(ValidRegions, ", "))
}

// NewClient creates a configured Recurly SDK client.
// API key precedence: --api-key flag > RECURLY_API_KEY env var > config file api_key.
// Region precedence: --region flag > RECURLY_REGION env var > config file region > default "us".
func NewClient() (*recurly.Client, error) {
	apiKey := viper.GetString("api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("API key not configured. Run 'recurly configure' or set RECURLY_API_KEY.")
	}

	regionStr := viper.GetString("region")
	if regionStr == "" {
		regionStr = "us"
	}
	if err := ValidateRegion(regionStr); err != nil {
		return nil, err
	}

	// Allow overriding the API base URL (e.g., for local Prism mock server in e2e tests).
	apiURL := os.Getenv("RECURLY_API_URL")
	if apiURL != "" {
		recurly.APIHost = apiURL
	}

	opts := recurly.ClientOptions{Region: recurly.US}
	if strings.EqualFold(regionStr, "eu") {
		opts.Region = recurly.EU
	}

	c, err := recurly.NewClientWithOptions(apiKey, opts)
	if err != nil {
		return nil, err
	}

	// Build transport chain: retryTransport -> [acceptRewriteTransport] -> base
	base := c.HTTPClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}

	// When using a mock server, rewrite the vendor Accept header to
	// application/json so Prism can serve the response.
	if apiURL != "" {
		base = &acceptRewriteTransport{base: base}
	}

	c.HTTPClient = &http.Client{
		Transport: newRetryTransport(base, os.Stderr, isJSONOutput),
	}

	return c, nil
}
