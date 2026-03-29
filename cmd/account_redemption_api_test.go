package cmd

import (
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
)

// mockAccountRedemptionAPI implements AccountRedemptionAPI for testing.
type mockAccountRedemptionAPI struct {
	listAccountCouponRedemptionsFn func(accountId string, params *recurly.ListAccountCouponRedemptionsParams, opts ...recurly.Option) (recurly.CouponRedemptionLister, error)
	listActiveCouponRedemptionsFn  func(accountId string, opts ...recurly.Option) (recurly.CouponRedemptionLister, error)
	createCouponRedemptionFn       func(accountId string, body *recurly.CouponRedemptionCreate, opts ...recurly.Option) (*recurly.CouponRedemption, error)
	removeCouponRedemptionFn       func(accountId string, opts ...recurly.Option) (*recurly.CouponRedemption, error)
	removeCouponRedemptionByIdFn   func(accountId string, couponRedemptionId string, opts ...recurly.Option) (*recurly.CouponRedemption, error)
	getCouponRedemptionFn          func(accountId string, couponRedemptionId string, opts ...recurly.Option) (*recurly.CouponRedemption, error)
}

func (m *mockAccountRedemptionAPI) ListAccountCouponRedemptions(accountId string, params *recurly.ListAccountCouponRedemptionsParams, opts ...recurly.Option) (recurly.CouponRedemptionLister, error) {
	if m.listAccountCouponRedemptionsFn != nil {
		return m.listAccountCouponRedemptionsFn(accountId, params, opts...)
	}
	return nil, nil
}

func (m *mockAccountRedemptionAPI) ListActiveCouponRedemptions(accountId string, opts ...recurly.Option) (recurly.CouponRedemptionLister, error) {
	if m.listActiveCouponRedemptionsFn != nil {
		return m.listActiveCouponRedemptionsFn(accountId, opts...)
	}
	return nil, nil
}

func (m *mockAccountRedemptionAPI) CreateCouponRedemption(accountId string, body *recurly.CouponRedemptionCreate, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
	if m.createCouponRedemptionFn != nil {
		return m.createCouponRedemptionFn(accountId, body, opts...)
	}
	return nil, nil
}

func (m *mockAccountRedemptionAPI) RemoveCouponRedemption(accountId string, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
	if m.removeCouponRedemptionFn != nil {
		return m.removeCouponRedemptionFn(accountId, opts...)
	}
	return nil, nil
}

func (m *mockAccountRedemptionAPI) RemoveCouponRedemptionById(accountId string, couponRedemptionId string, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
	if m.removeCouponRedemptionByIdFn != nil {
		return m.removeCouponRedemptionByIdFn(accountId, couponRedemptionId, opts...)
	}
	return nil, nil
}

func (m *mockAccountRedemptionAPI) GetCouponRedemption(accountId string, couponRedemptionId string, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
	if m.getCouponRedemptionFn != nil {
		return m.getCouponRedemptionFn(accountId, couponRedemptionId, opts...)
	}
	return nil, nil
}

func setMockAccountRedemptionAPI(mock *mockAccountRedemptionAPI) func() {
	orig := newAccountRedemptionAPI
	newAccountRedemptionAPI = func(_ *cobra.Command) (AccountRedemptionAPI, error) {
		return mock, nil
	}
	return func() { newAccountRedemptionAPI = orig }
}
