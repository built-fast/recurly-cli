package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCouponsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coupons",
		Short: "Manage coupons",
	}
	cmd.AddCommand(newCouponsListCmd())
	cmd.AddCommand(newCouponsGetCmd())
	return cmd
}

func newCouponsListCmd() *cobra.Command {
	var (
		limit     int
		all       bool
		order     string
		sort      string
		beginTime string
		endTime   string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List coupons",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newCouponAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			params := &recurly.ListCouponsParams{}

			if limit > 0 {
				params.Limit = recurly.Int(limit)
			}
			if cmd.Flags().Changed("order") {
				params.Order = recurly.String(order)
			}
			if cmd.Flags().Changed("sort") {
				params.Sort = recurly.String(sort)
			}
			if cmd.Flags().Changed("begin-time") {
				t, err := time.Parse(time.RFC3339, beginTime)
				if err != nil {
					return fmt.Errorf("invalid --begin-time: %w", err)
				}
				params.BeginTime = &t
			}
			if cmd.Flags().Changed("end-time") {
				t, err := time.Parse(time.RFC3339, endTime)
				if err != nil {
					return fmt.Errorf("invalid --end-time: %w", err)
				}
				params.EndTime = &t
			}

			lister, err := c.ListCoupons(params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.Coupon](lister, limit, all)
			if err != nil {
				return err
			}

			columns := []output.Column{
				{Header: "Code", Extract: func(v any) string { return v.(recurly.Coupon).Code }},
				{Header: "Name", Extract: func(v any) string { return v.(recurly.Coupon).Name }},
				{Header: "Discount Type", Extract: func(v any) string { return v.(recurly.Coupon).Discount.Type }},
				{Header: "State", Extract: func(v any) string { return v.(recurly.Coupon).State }},
				{Header: "Created At", Extract: func(v any) string {
					c := v.(recurly.Coupon)
					if c.CreatedAt != nil {
						return c.CreatedAt.Format(time.RFC3339)
					}
					return ""
				}},
			}

			items := make([]any, len(result.Items))
			for i, coupon := range result.Items {
				items[i] = coupon
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
	cmd.Flags().StringVar(&beginTime, "begin-time", "", "Filter by begin time (ISO8601 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter by end time (ISO8601 format)")

	return cmd
}

func newCouponsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <coupon_id>",
		Short: "Get coupon details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newCouponAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			coupon, err := c.GetCoupon(args[0])
			if err != nil {
				return err
			}

			columns := couponDetailColumns()

			formatted, err := output.FormatOne(format, columns, coupon)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func couponDetailColumns() []output.Column {
	return []output.Column{
		{Header: "ID", Extract: func(v any) string { return v.(*recurly.Coupon).Id }},
		{Header: "Code", Extract: func(v any) string { return v.(*recurly.Coupon).Code }},
		{Header: "Name", Extract: func(v any) string { return v.(*recurly.Coupon).Name }},
		{Header: "State", Extract: func(v any) string { return v.(*recurly.Coupon).State }},
		{Header: "Discount Type", Extract: func(v any) string { return v.(*recurly.Coupon).Discount.Type }},
		{Header: "Discount Value", Extract: func(v any) string {
			d := v.(*recurly.Coupon).Discount
			switch d.Type {
			case "percent":
				return fmt.Sprintf("%d%%", d.Percent)
			case "fixed":
				var parts []string
				for _, c := range d.Currencies {
					parts = append(parts, fmt.Sprintf("%.2f %s", c.Amount, c.Currency))
				}
				return fmt.Sprintf("%v", parts)
			case "free_trial":
				return fmt.Sprintf("%d %s", d.Trial.Length, d.Trial.Unit)
			default:
				return ""
			}
		}},
		{Header: "Duration", Extract: func(v any) string { return v.(*recurly.Coupon).Duration }},
		{Header: "Coupon Type", Extract: func(v any) string { return v.(*recurly.Coupon).CouponType }},
		{Header: "Max Redemptions", Extract: func(v any) string {
			return strconv.Itoa(v.(*recurly.Coupon).MaxRedemptions)
		}},
		{Header: "Max Redemptions Per Account", Extract: func(v any) string {
			return strconv.Itoa(v.(*recurly.Coupon).MaxRedemptionsPerAccount)
		}},
		{Header: "Redeem By", Extract: func(v any) string {
			c := v.(*recurly.Coupon)
			if c.RedeemBy != nil {
				return c.RedeemBy.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Applies To All Plans", Extract: func(v any) string {
			return strconv.FormatBool(v.(*recurly.Coupon).AppliesToAllPlans)
		}},
		{Header: "Applies To All Items", Extract: func(v any) string {
			return strconv.FormatBool(v.(*recurly.Coupon).AppliesToAllItems)
		}},
		{Header: "Created At", Extract: func(v any) string {
			c := v.(*recurly.Coupon)
			if c.CreatedAt != nil {
				return c.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Updated At", Extract: func(v any) string {
			c := v.(*recurly.Coupon)
			if c.UpdatedAt != nil {
				return c.UpdatedAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}
