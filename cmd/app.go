package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// appContextKey is an unexported type used as the key for storing App in a context.
type appContextKey struct{}

// App holds all API factory functions, allowing commands to obtain API clients
// without relying on package-level variables.
type App struct {
	NewAccountAPI          func(cmd *cobra.Command) (AccountAPI, error)
	NewPlanAPI             func(cmd *cobra.Command) (PlanAPI, error)
	NewItemAPI             func(cmd *cobra.Command) (ItemAPI, error)
	NewSubscriptionAPI     func(cmd *cobra.Command) (SubscriptionAPI, error)
	NewInvoiceAPI          func(cmd *cobra.Command) (InvoiceAPI, error)
	NewTransactionAPI      func(cmd *cobra.Command) (TransactionAPI, error)
	NewCouponAPI           func(cmd *cobra.Command) (CouponAPI, error)
	NewPlanAddOnAPI        func(cmd *cobra.Command) (PlanAddOnAPI, error)
	NewAccountBillingInfoAPI  func(cmd *cobra.Command) (AccountBillingInfoAPI, error)
	NewAccountNestedAPI       func(cmd *cobra.Command) (AccountNestedAPI, error)
	NewAccountRedemptionAPI   func(cmd *cobra.Command) (AccountRedemptionAPI, error)
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
