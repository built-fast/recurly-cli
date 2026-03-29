package cmd

import (
	"fmt"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newAccountRedemptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redemptions",
		Short: "Manage account coupon redemptions",
	}
	cmd.AddCommand(newAccountRedemptionsListCmd())
	cmd.AddCommand(newAccountRedemptionsListActiveCmd())
	cmd.AddCommand(withFromFile(newAccountRedemptionsCreateCmd()))
	cmd.AddCommand(newAccountRedemptionsRemoveCmd())
	return cmd
}

func redemptionListColumns() []output.Column {
	return []output.Column{
		{Header: "ID", Extract: func(v any) string { return v.(recurly.CouponRedemption).Id }},
		{Header: "Coupon Code", Extract: func(v any) string { return v.(recurly.CouponRedemption).Coupon.Code }},
		{Header: "State", Extract: func(v any) string { return v.(recurly.CouponRedemption).State }},
		{Header: "Currency", Extract: func(v any) string { return v.(recurly.CouponRedemption).Currency }},
		{Header: "Discounted", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(recurly.CouponRedemption).Discounted)
		}},
		{Header: "Created At", Extract: func(v any) string {
			r := v.(recurly.CouponRedemption)
			if r.CreatedAt != nil {
				return r.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}

func redemptionDetailColumns() []output.Column {
	return []output.Column{
		{Header: "ID", Extract: func(v any) string { return v.(*recurly.CouponRedemption).Id }},
		{Header: "Account Code", Extract: func(v any) string { return v.(*recurly.CouponRedemption).Account.Code }},
		{Header: "Coupon Code", Extract: func(v any) string { return v.(*recurly.CouponRedemption).Coupon.Code }},
		{Header: "State", Extract: func(v any) string { return v.(*recurly.CouponRedemption).State }},
		{Header: "Currency", Extract: func(v any) string { return v.(*recurly.CouponRedemption).Currency }},
		{Header: "Discounted", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.CouponRedemption).Discounted)
		}},
		{Header: "Subscription ID", Extract: func(v any) string { return v.(*recurly.CouponRedemption).SubscriptionId }},
		{Header: "Created At", Extract: func(v any) string {
			r := v.(*recurly.CouponRedemption)
			if r.CreatedAt != nil {
				return r.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Updated At", Extract: func(v any) string {
			r := v.(*recurly.CouponRedemption)
			if r.UpdatedAt != nil {
				return r.UpdatedAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}

func newAccountRedemptionsCreateCmd() *cobra.Command {
	var (
		couponID       string
		currency       string
		subscriptionID string
	)

	cmd := &cobra.Command{
		Use:   "create <account_id>",
		Short: "Redeem a coupon on an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAccountRedemptionAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			body := &recurly.CouponRedemptionCreate{
				CouponId: recurly.String(couponID),
			}

			if cmd.Flags().Changed("currency") {
				body.Currency = recurly.String(currency)
			}
			if cmd.Flags().Changed("subscription-id") {
				body.SubscriptionId = recurly.String(subscriptionID)
			}

			redemption, err := c.CreateCouponRedemption(args[0], body)
			if err != nil {
				return err
			}

			columns := redemptionDetailColumns()

			formatted, err := output.FormatOne(format, columns, redemption)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&couponID, "coupon-id", "", "Coupon ID to redeem (required)")
	cmd.Flags().StringVar(&currency, "currency", "", "3-letter ISO 4217 currency code")
	cmd.Flags().StringVar(&subscriptionID, "subscription-id", "", "Subscription ID to apply the coupon to")
	_ = cmd.MarkFlagRequired("coupon-id")

	return cmd
}

func newAccountRedemptionsRemoveCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "remove <account_id> [redemption_id]",
		Short: "Remove a coupon redemption from an account",
		Long:  "Remove the most recent coupon redemption from an account, or remove a specific redemption by ID.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				confirmed, err := confirm(cmd, "Are you sure you want to remove this coupon redemption? [y/N] ")
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Removal cancelled.")
					return err
				}
			}

			c, err := newAccountRedemptionAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			var redemption *recurly.CouponRedemption
			if len(args) == 2 {
				redemption, err = c.RemoveCouponRedemptionById(args[0], args[1])
			} else {
				redemption, err = c.RemoveCouponRedemption(args[0])
			}
			if err != nil {
				return err
			}

			columns := redemptionDetailColumns()

			formatted, err := output.FormatOne(format, columns, redemption)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func newAccountRedemptionsListCmd() *cobra.Command {
	var (
		limit int
		all   bool
		order string
		sort  string
	)

	cmd := &cobra.Command{
		Use:   "list <account_id>",
		Short: "List coupon redemptions for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAccountRedemptionAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			params := &recurly.ListAccountCouponRedemptionsParams{}

			if cmd.Flags().Changed("sort") {
				params.Sort = recurly.String(sort)
			}

			_ = order // order not supported by this API endpoint

			lister, err := c.ListAccountCouponRedemptions(args[0], params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.CouponRedemption](lister, limit, all)
			if err != nil {
				return err
			}

			columns := redemptionListColumns()

			items := make([]any, len(result.Items))
			for i, r := range result.Items {
				items[i] = r
			}

			formatted, err := output.FormatList(format, columns, items, result.HasMore)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results to return (default 20)")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all pages of results")
	cmd.Flags().StringVar(&order, "order", "", "Sort order (asc or desc)")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort field (e.g. created_at, updated_at)")

	return cmd
}

func newAccountRedemptionsListActiveCmd() *cobra.Command {
	var (
		limit int
		all   bool
		order string
		sort  string
	)

	cmd := &cobra.Command{
		Use:   "list-active <account_id>",
		Short: "List active coupon redemptions for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAccountRedemptionAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			_ = order // order not supported by this API endpoint
			_ = sort  // sort not supported by this API endpoint

			lister, err := c.ListActiveCouponRedemptions(args[0])
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.CouponRedemption](lister, limit, all)
			if err != nil {
				return err
			}

			columns := redemptionListColumns()

			items := make([]any, len(result.Items))
			for i, r := range result.Items {
				items[i] = r
			}

			formatted, err := output.FormatList(format, columns, items, result.HasMore)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results to return (default 20)")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all pages of results")
	cmd.Flags().StringVar(&order, "order", "", "Sort order (asc or desc)")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort field (e.g. created_at, updated_at)")

	return cmd
}
