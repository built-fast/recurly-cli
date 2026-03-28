package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	recurly "github.com/recurly/recurly-client-go/v5"
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
var newAccountNestedAPI = func() (AccountNestedAPI, error) {
	return client.NewClient()
}
