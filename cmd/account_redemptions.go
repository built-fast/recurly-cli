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
