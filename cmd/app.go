package cmd

import (
	"context"

	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// appContextKey is an unexported type used as the key for storing App in a context.
type appContextKey struct{}

// App holds all API factory functions, allowing commands to obtain API clients
// without relying on package-level variables.
type App struct {
	NewAccountAPI            func(cmd *cobra.Command) (AccountAPI, error)
	NewPlanAPI               func(cmd *cobra.Command) (PlanAPI, error)
	NewItemAPI               func(cmd *cobra.Command) (ItemAPI, error)
	NewSubscriptionAPI       func(cmd *cobra.Command) (SubscriptionAPI, error)
	NewInvoiceAPI            func(cmd *cobra.Command) (InvoiceAPI, error)
	NewTransactionAPI        func(cmd *cobra.Command) (TransactionAPI, error)
	NewCouponAPI             func(cmd *cobra.Command) (CouponAPI, error)
	NewPlanAddOnAPI          func(cmd *cobra.Command) (PlanAddOnAPI, error)
	NewAccountBillingInfoAPI func(cmd *cobra.Command) (AccountBillingInfoAPI, error)
	NewAccountNestedAPI      func(cmd *cobra.Command) (AccountNestedAPI, error)
	NewAccountRedemptionAPI  func(cmd *cobra.Command) (AccountRedemptionAPI, error)
}

// NewAppContext returns a new context that carries the given App.
func NewAppContext(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, appContextKey{}, app)
}

// AppFromContext returns the App stored in ctx. If no App is set,
// it returns a non-nil zero-value App to prevent nil panics.
func AppFromContext(ctx context.Context) *App {
	if app, ok := ctx.Value(appContextKey{}).(*App); ok && app != nil {
		return app
	}
	return &App{}
}

// newAPIFactory returns a factory function that creates an API client using
// the standard viper/output config pattern.
func newAPIFactory[T any](fn func(client.ClientConfig) (T, error)) func(cmd *cobra.Command) (T, error) {
	return func(cmd *cobra.Command) (T, error) {
		cfg := output.FromContext(cmd.Context())
		return fn(client.ClientConfig{
			APIKey: viper.GetString("api_key"),
			Region: viper.GetString("region"),
			IsJSON: func() bool { return isJSONFormat(cfg.Format) },
		})
	}
}

// DefaultApp returns an App with all factory functions wired to the
// production implementations (viper config + real Recurly client).
func DefaultApp() *App {
	return &App{
		NewAccountAPI:            newAPIFactory(func(c client.ClientConfig) (AccountAPI, error) { return client.NewClient(c) }),
		NewPlanAPI:               newAPIFactory(func(c client.ClientConfig) (PlanAPI, error) { return client.NewClient(c) }),
		NewItemAPI:               newAPIFactory(func(c client.ClientConfig) (ItemAPI, error) { return client.NewClient(c) }),
		NewSubscriptionAPI:       newAPIFactory(func(c client.ClientConfig) (SubscriptionAPI, error) { return client.NewClient(c) }),
		NewInvoiceAPI:            newAPIFactory(func(c client.ClientConfig) (InvoiceAPI, error) { return client.NewClient(c) }),
		NewTransactionAPI:        newAPIFactory(func(c client.ClientConfig) (TransactionAPI, error) { return client.NewClient(c) }),
		NewCouponAPI:             newAPIFactory(func(c client.ClientConfig) (CouponAPI, error) { return client.NewClient(c) }),
		NewPlanAddOnAPI:          newAPIFactory(func(c client.ClientConfig) (PlanAddOnAPI, error) { return client.NewClient(c) }),
		NewAccountBillingInfoAPI: newAPIFactory(func(c client.ClientConfig) (AccountBillingInfoAPI, error) { return client.NewClient(c) }),
		NewAccountNestedAPI:      newAPIFactory(func(c client.ClientConfig) (AccountNestedAPI, error) { return client.NewClient(c) }),
		NewAccountRedemptionAPI:  newAPIFactory(func(c client.ClientConfig) (AccountRedemptionAPI, error) { return client.NewClient(c) }),
	}
}
