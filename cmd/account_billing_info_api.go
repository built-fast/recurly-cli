package cmd

import (
	recurly "github.com/recurly/recurly-client-go/v5"
)

// AccountBillingInfoAPI abstracts the Recurly SDK methods used by billing info commands,
// allowing tests to inject mocks without making real API calls.
type AccountBillingInfoAPI interface {
	GetBillingInfo(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error)
	UpdateBillingInfo(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error)
	RemoveBillingInfo(accountId string, opts ...recurly.Option) (*recurly.Empty, error)
}
