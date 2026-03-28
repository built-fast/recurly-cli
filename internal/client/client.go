package client

import (
	"fmt"
	"os"
	"strings"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

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
	if apiURL := os.Getenv("RECURLY_API_URL"); apiURL != "" {
		recurly.APIHost = apiURL
	}

	opts := recurly.ClientOptions{Region: recurly.US}
	if strings.EqualFold(regionStr, "eu") {
		opts.Region = recurly.EU
	}

	return recurly.NewClientWithOptions(apiKey, opts)
}
