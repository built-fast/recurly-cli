package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AccountBillingInfoAPI abstracts the Recurly SDK methods used by billing info commands,
// allowing tests to inject mocks without making real API calls.
type AccountBillingInfoAPI interface {
	GetBillingInfo(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error)
	UpdateBillingInfo(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error)
	RemoveBillingInfo(accountId string, opts ...recurly.Option) (*recurly.Empty, error)
}

// newAccountBillingInfoAPI is the factory function used by billing info commands to get an API client.
// Tests override this to inject mocks.
var newAccountBillingInfoAPI = func(cmd *cobra.Command) (AccountBillingInfoAPI, error) {
	cfg := output.FromContext(cmd.Context())
	return client.NewClient(client.ClientConfig{
		APIKey: viper.GetString("api_key"),
		Region: viper.GetString("region"),
		IsJSON: func() bool { return isJSONFormat(cfg.Format) },
	})
}
