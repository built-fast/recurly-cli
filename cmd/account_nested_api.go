package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AccountNestedAPI abstracts the Recurly SDK methods used by account nested listing commands,
// allowing tests to inject mocks without making real API calls.
type AccountNestedAPI interface {
	ListAccountSubscriptions(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error)
	ListAccountInvoices(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error)
	ListAccountTransactions(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error)
}

// newAccountNestedAPI is the factory function used by account nested listing commands to get an API client.
// Tests override this to inject mocks.
var newAccountNestedAPI = func(cmd *cobra.Command) (AccountNestedAPI, error) {
	cfg := output.FromContext(cmd.Context())
	return client.NewClient(client.ClientConfig{
		APIKey: viper.GetString("api_key"),
		Region: viper.GetString("region"),
		IsJSON: func() bool { return isJSONFormat(cfg.Format) },
	})
}
