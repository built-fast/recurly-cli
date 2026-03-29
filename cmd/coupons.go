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
	cmd.AddCommand(newCouponsCreatePercentCmd())
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

func newCouponsCreatePercentCmd() *cobra.Command {
	var (
		code                     string
		name                     string
		discountPercent          int
		maxRedemptions           int
		maxRedemptionsPerAccount int
		duration                 string
		temporalAmount           int
		temporalUnit             string
		couponType               string
		uniqueCodeTemplate       string
		appliesToAllPlans        bool
		appliesToAllItems        bool
		appliesToNonPlanCharges  bool
		planCodes                []string
		itemCodes                []string
		redemptionResource       string
		hostedPageDescription    string
		invoiceDescription       string
		redeemBy                 string
	)

	cmd := &cobra.Command{
		Use:   "create-percent",
		Short: "Create a percentage-based coupon",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newCouponAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			body := &recurly.CouponCreate{
				Code:            recurly.String(code),
				Name:            recurly.String(name),
				DiscountType:    recurly.String("percent"),
				DiscountPercent: recurly.Int(discountPercent),
			}

			if cmd.Flags().Changed("max-redemptions") {
				body.MaxRedemptions = recurly.Int(maxRedemptions)
			}
			if cmd.Flags().Changed("max-redemptions-per-account") {
				body.MaxRedemptionsPerAccount = recurly.Int(maxRedemptionsPerAccount)
			}
			if cmd.Flags().Changed("duration") {
				body.Duration = recurly.String(duration)
			}
			if cmd.Flags().Changed("temporal-amount") {
				body.TemporalAmount = recurly.Int(temporalAmount)
			}
			if cmd.Flags().Changed("temporal-unit") {
				body.TemporalUnit = recurly.String(temporalUnit)
			}
			if cmd.Flags().Changed("coupon-type") {
				body.CouponType = recurly.String(couponType)
			}
			if cmd.Flags().Changed("unique-code-template") {
				body.UniqueCodeTemplate = recurly.String(uniqueCodeTemplate)
			}
			if cmd.Flags().Changed("applies-to-all-plans") {
				body.AppliesToAllPlans = recurly.Bool(appliesToAllPlans)
			}
			if cmd.Flags().Changed("applies-to-all-items") {
				body.AppliesToAllItems = recurly.Bool(appliesToAllItems)
			}
			if cmd.Flags().Changed("applies-to-non-plan-charges") {
				body.AppliesToNonPlanCharges = recurly.Bool(appliesToNonPlanCharges)
			}
			if cmd.Flags().Changed("plan-codes") {
				body.PlanCodes = &planCodes
			}
			if cmd.Flags().Changed("item-codes") {
				body.ItemCodes = &itemCodes
			}
			if cmd.Flags().Changed("redemption-resource") {
				body.RedemptionResource = recurly.String(redemptionResource)
			}
			if cmd.Flags().Changed("hosted-page-description") {
				body.HostedDescription = recurly.String(hostedPageDescription)
			}
			if cmd.Flags().Changed("invoice-description") {
				body.InvoiceDescription = recurly.String(invoiceDescription)
			}
			if cmd.Flags().Changed("redeem-by") {
				body.RedeemByDate = recurly.String(redeemBy)
			}

			coupon, err := c.CreateCoupon(body)
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

	cmd.Flags().StringVar(&code, "code", "", "Coupon code (required)")
	cmd.Flags().StringVar(&name, "name", "", "Coupon name (required)")
	cmd.Flags().IntVar(&discountPercent, "discount-percent", 0, "Discount percentage (required)")
	cmd.Flags().IntVar(&maxRedemptions, "max-redemptions", 0, "Maximum number of redemptions")
	cmd.Flags().IntVar(&maxRedemptionsPerAccount, "max-redemptions-per-account", 0, "Maximum redemptions per account")
	cmd.Flags().StringVar(&duration, "duration", "", "Duration: single_use, temporal, or forever")
	cmd.Flags().IntVar(&temporalAmount, "temporal-amount", 0, "Temporal duration amount")
	cmd.Flags().StringVar(&temporalUnit, "temporal-unit", "", "Temporal duration unit: day, week, month, or year")
	cmd.Flags().StringVar(&couponType, "coupon-type", "", "Coupon type: single_code or bulk")
	cmd.Flags().StringVar(&uniqueCodeTemplate, "unique-code-template", "", "Template for bulk coupon unique codes")
	cmd.Flags().BoolVar(&appliesToAllPlans, "applies-to-all-plans", false, "Applies to all plans")
	cmd.Flags().BoolVar(&appliesToAllItems, "applies-to-all-items", false, "Applies to all items")
	cmd.Flags().BoolVar(&appliesToNonPlanCharges, "applies-to-non-plan-charges", false, "Applies to non-plan charges")
	cmd.Flags().StringSliceVar(&planCodes, "plan-codes", nil, "Plan codes this coupon applies to")
	cmd.Flags().StringSliceVar(&itemCodes, "item-codes", nil, "Item codes this coupon applies to")
	cmd.Flags().StringVar(&redemptionResource, "redemption-resource", "", "Redemption resource: account or subscription")
	cmd.Flags().StringVar(&hostedPageDescription, "hosted-page-description", "", "Hosted page description")
	cmd.Flags().StringVar(&invoiceDescription, "invoice-description", "", "Invoice description")
	cmd.Flags().StringVar(&redeemBy, "redeem-by", "", "Coupon expiration date (YYYY-MM-DD)")

	_ = cmd.MarkFlagRequired("code")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("discount-percent")

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
