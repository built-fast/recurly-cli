package cmd

import (
	recurly "github.com/recurly/recurly-client-go/v5"
)

// AccountRedemptionAPI abstracts the Recurly SDK methods used by account coupon redemption commands,
// allowing tests to inject mocks without making real API calls.
type AccountRedemptionAPI interface {
	ListAccountCouponRedemptions(accountId string, params *recurly.ListAccountCouponRedemptionsParams, opts ...recurly.Option) (recurly.CouponRedemptionLister, error)
	ListActiveCouponRedemptions(accountId string, opts ...recurly.Option) (recurly.CouponRedemptionLister, error)
	CreateCouponRedemption(accountId string, body *recurly.CouponRedemptionCreate, opts ...recurly.Option) (*recurly.CouponRedemption, error)
	RemoveCouponRedemption(accountId string, opts ...recurly.Option) (*recurly.CouponRedemption, error)
	RemoveCouponRedemptionById(accountId string, couponRedemptionId string, opts ...recurly.Option) (*recurly.CouponRedemption, error)
	GetCouponRedemption(accountId string, couponRedemptionId string, opts ...recurly.Option) (*recurly.CouponRedemption, error)
}
