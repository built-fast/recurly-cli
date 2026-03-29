package cmd

import (
	recurly "github.com/recurly/recurly-client-go/v5"
)

// ItemAPI abstracts the Recurly SDK methods used by item commands,
// allowing tests to inject mocks without making real API calls.
type ItemAPI interface {
	ListItems(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error)
	GetItem(itemId string, opts ...recurly.Option) (*recurly.Item, error)
	CreateItem(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error)
	UpdateItem(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error)
	DeactivateItem(itemId string, opts ...recurly.Option) (*recurly.Item, error)
	ReactivateItem(itemId string, opts ...recurly.Option) (*recurly.Item, error)
}
