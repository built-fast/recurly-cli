package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CouponAPI abstracts the Recurly SDK methods used by coupon commands,
// allowing tests to inject mocks without making real API calls.
type CouponAPI interface {
	ListCoupons(params *recurly.ListCouponsParams, opts ...recurly.Option) (recurly.CouponLister, error)
	GetCoupon(couponId string, opts ...recurly.Option) (*recurly.Coupon, error)
	CreateCoupon(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error)
	UpdateCoupon(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error)
	DeactivateCoupon(couponId string, opts ...recurly.Option) (*recurly.Coupon, error)
	RestoreCoupon(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error)
	GenerateUniqueCouponCodes(couponId string, body *recurly.CouponBulkCreate, opts ...recurly.Option) (*recurly.UniqueCouponCodeParams, error)
	ListUniqueCouponCodes(couponId string, params *recurly.ListUniqueCouponCodesParams, opts ...recurly.Option) (recurly.UniqueCouponCodeLister, error)
}

// newCouponAPI is the factory function used by coupon commands to get an API client.
// Tests override this to inject mocks.
var newCouponAPI = func(cmd *cobra.Command) (CouponAPI, error) {
	cfg := output.FromContext(cmd.Context())
	return client.NewClient(client.ClientConfig{
		APIKey: viper.GetString("api_key"),
		Region: viper.GetString("region"),
		IsJSON: func() bool { return isJSONFormat(cfg.Format) },
	})
}
