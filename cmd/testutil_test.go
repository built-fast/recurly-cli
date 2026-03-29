package cmd

import (
	"context"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
)

// Compile-time interface assertions for all concrete lister types used in tests.
var (
	_ recurly.AccountLister          = (*mockLister[recurly.Account])(nil)
	_ recurly.PlanLister             = (*mockLister[recurly.Plan])(nil)
	_ recurly.SubscriptionLister     = (*mockLister[recurly.Subscription])(nil)
	_ recurly.InvoiceLister          = (*mockLister[recurly.Invoice])(nil)
	_ recurly.LineItemLister         = (*mockLister[recurly.LineItem])(nil)
	_ recurly.TransactionLister      = (*mockLister[recurly.Transaction])(nil)
	_ recurly.CouponLister           = (*mockLister[recurly.Coupon])(nil)
	_ recurly.UniqueCouponCodeLister = (*mockLister[recurly.UniqueCouponCode])(nil)
	_ recurly.ItemLister             = (*mockLister[recurly.Item])(nil)
	_ recurly.AddOnLister            = (*mockLister[recurly.AddOn])(nil)
	_ recurly.CouponRedemptionLister = (*mockLister[recurly.CouponRedemption])(nil)
)

// mockLister is a generic mock that satisfies any recurly.*Lister interface
// (e.g., recurly.AccountLister, recurly.PlanLister) whose Data() returns []T.
//
// The only lister mock NOT replaced by this generic is mockLineItemListerWithMore
// in invoices_test.go, which overrides HasMore() to always return true in order
// to simulate a multi-page response. That custom behavior cannot be expressed
// with a single generic struct, so it remains as a targeted mock.
type mockLister[T any] struct {
	items   []T
	fetched bool
}

func (m *mockLister[T]) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockLister[T]) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockLister[T]) Count() (*int64, error) {
	n := int64(len(m.items))
	return &n, nil
}

func (m *mockLister[T]) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockLister[T]) Data() []T {
	return m.items
}

func (m *mockLister[T]) HasMore() bool {
	return !m.fetched
}

func (m *mockLister[T]) Next() string {
	return ""
}

// ---------------------------------------------------------------------------
// Per-API helpers: wrap a mock into an *App with a single line.
//
//   app := newTestAccountApp(mock)
//
// replaces the verbose:
//
//   app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}
// ---------------------------------------------------------------------------

func newTestAccountApp(api AccountAPI) *App {
	return &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return api, nil }}
}

func newTestPlanApp(api PlanAPI) *App {
	return &App{NewPlanAPI: func(_ *cobra.Command) (PlanAPI, error) { return api, nil }}
}

func newTestItemApp(api ItemAPI) *App {
	return &App{NewItemAPI: func(_ *cobra.Command) (ItemAPI, error) { return api, nil }}
}

func newTestSubscriptionApp(api SubscriptionAPI) *App {
	return &App{NewSubscriptionAPI: func(_ *cobra.Command) (SubscriptionAPI, error) { return api, nil }}
}

func newTestInvoiceApp(api InvoiceAPI) *App {
	return &App{NewInvoiceAPI: func(_ *cobra.Command) (InvoiceAPI, error) { return api, nil }}
}

func newTestTransactionApp(api TransactionAPI) *App {
	return &App{NewTransactionAPI: func(_ *cobra.Command) (TransactionAPI, error) { return api, nil }}
}

func newTestCouponApp(api CouponAPI) *App {
	return &App{NewCouponAPI: func(_ *cobra.Command) (CouponAPI, error) { return api, nil }}
}

func newTestPlanAddOnApp(api PlanAddOnAPI) *App {
	return &App{NewPlanAddOnAPI: func(_ *cobra.Command) (PlanAddOnAPI, error) { return api, nil }}
}

func newTestAccountBillingInfoApp(api AccountBillingInfoAPI) *App {
	return &App{NewAccountBillingInfoAPI: func(_ *cobra.Command) (AccountBillingInfoAPI, error) { return api, nil }}
}

func newTestAccountNestedApp(api AccountNestedAPI) *App {
	return &App{NewAccountNestedAPI: func(_ *cobra.Command) (AccountNestedAPI, error) { return api, nil }}
}

func newTestAccountRedemptionApp(api AccountRedemptionAPI) *App {
	return &App{NewAccountRedemptionAPI: func(_ *cobra.Command) (AccountRedemptionAPI, error) { return api, nil }}
}
