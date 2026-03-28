package client

import (
	"fmt"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

// NewClient creates a configured Recurly SDK client.
// API key precedence: --api-key flag > RECURLY_API_KEY env var > config file api_key.
// Region precedence: --region flag > RECURLY_REGION env var > config file region > default "us".
func NewClient() (*recurly.Client, error) {
	apiKey := viper.GetString("api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("API key not configured. Run 'recurly configure' or set RECURLY_API_KEY.")
	}

	regionStr := viper.GetString("region")

	opts := recurly.ClientOptions{Region: recurly.US}
	if regionStr == "eu" {
		opts.Region = recurly.EU
	}

	return recurly.NewClientWithOptions(apiKey, opts)
}
