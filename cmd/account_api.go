package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AccountAPI abstracts the Recurly SDK methods used by account commands,
// allowing tests to inject mocks without making real API calls.
type AccountAPI interface {
	ListAccounts(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error)
	GetAccount(accountId string, opts ...recurly.Option) (*recurly.Account, error)
	CreateAccount(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error)
	UpdateAccount(accountId string, body *recurly.AccountUpdate, opts ...recurly.Option) (*recurly.Account, error)
	DeactivateAccount(accountId string, opts ...recurly.Option) (*recurly.Account, error)
	ReactivateAccount(accountId string, opts ...recurly.Option) (*recurly.Account, error)
}

// newAccountAPI is the factory function used by account commands to get an API client.
// Tests override this to inject mocks.
var newAccountAPI = func(cmd *cobra.Command) (AccountAPI, error) {
	cfg := output.FromContext(cmd.Context())
	return client.NewClient(client.ClientConfig{
		APIKey: viper.GetString("api_key"),
		Region: viper.GetString("region"),
		IsJSON: func() bool { return isJSONFormat(cfg.Format) },
	})
}
