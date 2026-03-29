package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// newAccountRedemptionAPI is the factory function used by account coupon redemption commands to get an API client.
// Tests override this to inject mocks.
var newAccountRedemptionAPI = func(cmd *cobra.Command) (AccountRedemptionAPI, error) {
	cfg := output.FromContext(cmd.Context())
	return client.NewClient(client.ClientConfig{
		APIKey: viper.GetString("api_key"),
		Region: viper.GetString("region"),
		IsJSON: func() bool { return isJSONFormat(cfg.Format) },
	})
}
