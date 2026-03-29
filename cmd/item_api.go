package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// newItemAPI is the factory function used by item commands to get an API client.
// Tests override this to inject mocks.
var newItemAPI = func(cmd *cobra.Command) (ItemAPI, error) {
	cfg := output.FromContext(cmd.Context())
	return client.NewClient(client.ClientConfig{
		APIKey: viper.GetString("api_key"),
		Region: viper.GetString("region"),
		IsJSON: func() bool { return isJSONFormat(cfg.Format) },
	})
}
