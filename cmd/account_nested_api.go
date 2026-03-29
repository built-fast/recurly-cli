package cmd

import (
	recurly "github.com/recurly/recurly-client-go/v5"
)

// AccountNestedAPI abstracts the Recurly SDK methods used by account nested listing commands,
// allowing tests to inject mocks without making real API calls.
type AccountNestedAPI interface {
	ListAccountSubscriptions(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error)
	ListAccountInvoices(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error)
	ListAccountTransactions(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error)
}
