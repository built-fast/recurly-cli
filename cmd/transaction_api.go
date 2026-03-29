package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TransactionAPI abstracts the Recurly SDK methods used by transaction commands,
// allowing tests to inject mocks without making real API calls.
type TransactionAPI interface {
	ListTransactions(params *recurly.ListTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error)
	GetTransaction(transactionId string, opts ...recurly.Option) (*recurly.Transaction, error)
}

// newTransactionAPI is the factory function used by transaction commands to get an API client.
// Tests override this to inject mocks.
var newTransactionAPI = func(cmd *cobra.Command) (TransactionAPI, error) {
	cfg := output.FromContext(cmd.Context())
	return client.NewClient(client.ClientConfig{
		APIKey: viper.GetString("api_key"),
		Region: viper.GetString("region"),
		IsJSON: func() bool { return isJSONFormat(cfg.Format) },
	})
}
