package cmd

import (
	recurly "github.com/recurly/recurly-client-go/v5"
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
