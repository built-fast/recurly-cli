package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// InvoiceAPI abstracts the Recurly SDK methods used by invoice commands,
// allowing tests to inject mocks without making real API calls.
type InvoiceAPI interface {
	ListInvoices(params *recurly.ListInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error)
	GetInvoice(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error)
	VoidInvoice(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error)
	CollectInvoice(invoiceId string, params *recurly.CollectInvoiceParams, opts ...recurly.Option) (*recurly.Invoice, error)
	MarkInvoiceFailed(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error)
	ListInvoiceLineItems(invoiceId string, params *recurly.ListInvoiceLineItemsParams, opts ...recurly.Option) (recurly.LineItemLister, error)
}

// newInvoiceAPI is the factory function used by invoice commands to get an API client.
// Tests override this to inject mocks.
var newInvoiceAPI = func(cmd *cobra.Command) (InvoiceAPI, error) {
	cfg := output.FromContext(cmd.Context())
	return client.NewClient(client.ClientConfig{
		APIKey: viper.GetString("api_key"),
		Region: viper.GetString("region"),
		IsJSON: func() bool { return isJSONFormat(cfg.Format) },
	})
}
