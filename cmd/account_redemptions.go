package cmd

import (
	"fmt"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
)

func newAccountRedemptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redemptions",
		Short: "Manage account coupon redemptions",
	}
	cmd.AddCommand(newAccountRedemptionsListCmd())
	cmd.AddCommand(newAccountRedemptionsListActiveCmd())
	cmd.AddCommand(withFromFile(withInteractive(newAccountRedemptionsCreateCmd())))
	cmd.AddCommand(newAccountRedemptionsRemoveCmd())
	return cmd
}

func redemptionListColumns() []output.Column {
	type R = recurly.CouponRedemption
	return output.ToColumns([]output.TypedColumn[R]{
		output.StringColumn[R]("ID", func(r R) string { return r.Id }),
		output.StringColumn[R]("Coupon Code", func(r R) string { return r.Coupon.Code }),
		output.StringColumn[R]("State", func(r R) string { return r.State }),
		output.StringColumn[R]("Currency", func(r R) string { return r.Currency }),
		output.FloatColumn[R]("Discounted", func(r R) float64 { return r.Discounted }),
		output.TimeColumn[R]("Created At", func(r R) *time.Time { return r.CreatedAt }),
	})
}

func redemptionDetailColumns() []output.Column {
	type R = *recurly.CouponRedemption
	return output.ToColumns([]output.TypedColumn[R]{
		output.StringColumn[R]("ID", func(r R) string { return r.Id }),
		output.StringColumn[R]("Account Code", func(r R) string { return r.Account.Code }),
		output.StringColumn[R]("Coupon Code", func(r R) string { return r.Coupon.Code }),
		output.StringColumn[R]("State", func(r R) string { return r.State }),
		output.StringColumn[R]("Currency", func(r R) string { return r.Currency }),
		output.FloatColumn[R]("Discounted", func(r R) float64 { return r.Discounted }),
		output.StringColumn[R]("Subscription ID", func(r R) string { return r.SubscriptionId }),
		output.TimeColumn[R]("Created At", func(r R) *time.Time { return r.CreatedAt }),
		output.TimeColumn[R]("Updated At", func(r R) *time.Time { return r.UpdatedAt }),
	})
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
			c, err := AppFromContext(cmd.Context()).NewAccountRedemptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			formatted, err := output.FormatOne(cfg, columns, redemption)
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

			c, err := AppFromContext(cmd.Context()).NewAccountRedemptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			formatted, err := output.FormatOne(cfg, columns, redemption)
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
			c, err := AppFromContext(cmd.Context()).NewAccountRedemptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			formatted, err := output.FormatList(cfg, columns, items, result.HasMore)
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
			c, err := AppFromContext(cmd.Context()).NewAccountRedemptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			formatted, err := output.FormatList(cfg, columns, items, result.HasMore)
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
