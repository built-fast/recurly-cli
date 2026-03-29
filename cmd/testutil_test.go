package cmd

import (
	"context"

	recurly "github.com/recurly/recurly-client-go/v5"
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
