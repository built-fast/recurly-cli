package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
)

func newCouponsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coupons",
		Short: "Manage coupons",
	}
	cmd.AddCommand(newCouponsListCmd())
	cmd.AddCommand(withWatch(newCouponsGetCmd()))
	cmd.AddCommand(withFromFile(withInteractive(newCouponsCreatePercentCmd())))
	cmd.AddCommand(withFromFile(withInteractive(newCouponsCreateFixedCmd())))
	cmd.AddCommand(withFromFile(withInteractive(newCouponsCreateFreeTrialCmd())))
	cmd.AddCommand(withFromFile(newCouponsUpdateCmd()))
	cmd.AddCommand(newCouponsDeactivateCmd())
	cmd.AddCommand(newCouponsRestoreCmd())
	cmd.AddCommand(withInteractive(newCouponsGenerateCodesCmd()))
	cmd.AddCommand(newCouponsListCodesCmd())
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
			c, err := AppFromContext(cmd.Context()).NewCouponAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			type C = recurly.Coupon
			columns := output.ToColumns([]output.TypedColumn[C]{
				output.StringColumn[C]("Code", func(c C) string { return c.Code }),
				output.StringColumn[C]("Name", func(c C) string { return c.Name }),
				output.StringColumn[C]("Discount Type", func(c C) string { return c.Discount.Type }),
				output.StringColumn[C]("State", func(c C) string { return c.State }),
				output.TimeColumn[C]("Created At", func(c C) *time.Time { return c.CreatedAt }),
			})

			items := make([]any, len(result.Items))
			for i, coupon := range result.Items {
				items[i] = coupon
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
			c, err := AppFromContext(cmd.Context()).NewCouponAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			coupon, err := c.GetCoupon(args[0])
			if err != nil {
				return err
			}

			columns := couponDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, coupon)
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
			c, err := AppFromContext(cmd.Context()).NewCouponAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			formatted, err := output.FormatOne(cfg, columns, coupon)
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

func newCouponsCreateFixedCmd() *cobra.Command {
	var (
		code                     string
		name                     string
		currencies               []string
		discountAmounts          []float64
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
		Use:   "create-fixed",
		Short: "Create a fixed-amount coupon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(currencies) != len(discountAmounts) {
				return fmt.Errorf("--currency and --discount-amount must be specified in pairs (got %d currencies and %d amounts)", len(currencies), len(discountAmounts))
			}

			c, err := AppFromContext(cmd.Context()).NewCouponAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			couponPricings := make([]recurly.CouponPricing, len(currencies))
			for i := range currencies {
				couponPricings[i] = recurly.CouponPricing{
					Currency: recurly.String(currencies[i]),
					Discount: recurly.Float(discountAmounts[i]),
				}
			}

			body := &recurly.CouponCreate{
				Code:         recurly.String(code),
				Name:         recurly.String(name),
				DiscountType: recurly.String("fixed"),
				Currencies:   &couponPricings,
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

			formatted, err := output.FormatOne(cfg, columns, coupon)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&code, "code", "", "Coupon code (required)")
	cmd.Flags().StringVar(&name, "name", "", "Coupon name (required)")
	cmd.Flags().StringSliceVar(&currencies, "currency", nil, "Currency code (e.g. USD); specify once per currency/amount pair (required)")
	cmd.Flags().Float64SliceVar(&discountAmounts, "discount-amount", nil, "Discount amount; specify once per currency/amount pair (required)")
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
	_ = cmd.MarkFlagRequired("currency")
	_ = cmd.MarkFlagRequired("discount-amount")

	return cmd
}

func newCouponsCreateFreeTrialCmd() *cobra.Command {
	var (
		code                     string
		name                     string
		freeTrialAmount          int
		freeTrialUnit            string
		maxRedemptions           int
		maxRedemptionsPerAccount int
		couponType               string
		uniqueCodeTemplate       string
		appliesToAllPlans        bool
		planCodes                []string
		redemptionResource       string
		hostedPageDescription    string
		invoiceDescription       string
		redeemBy                 string
	)

	cmd := &cobra.Command{
		Use:   "create-free-trial",
		Short: "Create a free trial coupon",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewCouponAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &recurly.CouponCreate{
				Code:            recurly.String(code),
				Name:            recurly.String(name),
				DiscountType:    recurly.String("free_trial"),
				FreeTrialAmount: recurly.Int(freeTrialAmount),
				FreeTrialUnit:   recurly.String(freeTrialUnit),
			}

			if cmd.Flags().Changed("max-redemptions") {
				body.MaxRedemptions = recurly.Int(maxRedemptions)
			}
			if cmd.Flags().Changed("max-redemptions-per-account") {
				body.MaxRedemptionsPerAccount = recurly.Int(maxRedemptionsPerAccount)
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
			if cmd.Flags().Changed("plan-codes") {
				body.PlanCodes = &planCodes
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

			formatted, err := output.FormatOne(cfg, columns, coupon)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&code, "code", "", "Coupon code (required)")
	cmd.Flags().StringVar(&name, "name", "", "Coupon name (required)")
	cmd.Flags().IntVar(&freeTrialAmount, "free-trial-amount", 0, "Free trial duration amount (required)")
	cmd.Flags().StringVar(&freeTrialUnit, "free-trial-unit", "", "Free trial duration unit: day, week, month, or year (required)")
	cmd.Flags().IntVar(&maxRedemptions, "max-redemptions", 0, "Maximum number of redemptions")
	cmd.Flags().IntVar(&maxRedemptionsPerAccount, "max-redemptions-per-account", 0, "Maximum redemptions per account")
	cmd.Flags().StringVar(&couponType, "coupon-type", "", "Coupon type: single_code or bulk")
	cmd.Flags().StringVar(&uniqueCodeTemplate, "unique-code-template", "", "Template for bulk coupon unique codes")
	cmd.Flags().BoolVar(&appliesToAllPlans, "applies-to-all-plans", false, "Applies to all plans")
	cmd.Flags().StringSliceVar(&planCodes, "plan-codes", nil, "Plan codes this coupon applies to")
	cmd.Flags().StringVar(&redemptionResource, "redemption-resource", "", "Redemption resource: account or subscription")
	cmd.Flags().StringVar(&hostedPageDescription, "hosted-page-description", "", "Hosted page description")
	cmd.Flags().StringVar(&invoiceDescription, "invoice-description", "", "Invoice description")
	cmd.Flags().StringVar(&redeemBy, "redeem-by", "", "Coupon expiration date (YYYY-MM-DD)")

	_ = cmd.MarkFlagRequired("code")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("free-trial-amount")
	_ = cmd.MarkFlagRequired("free-trial-unit")

	setFlagOptions(cmd, "free-trial-unit", []string{"day", "week", "month", "year"})

	return cmd
}

func newCouponsUpdateCmd() *cobra.Command {
	var (
		name                     string
		maxRedemptions           int
		maxRedemptionsPerAccount int
		hostedDescription        string
		invoiceDescription       string
		redeemByDate             string
	)

	cmd := &cobra.Command{
		Use:   "update <coupon_id>",
		Short: "Update a coupon",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewCouponAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
			body := &recurly.CouponUpdate{}

			if cmd.Flags().Changed("name") {
				body.Name = recurly.String(name)
			}
			if cmd.Flags().Changed("max-redemptions") {
				body.MaxRedemptions = recurly.Int(maxRedemptions)
			}
			if cmd.Flags().Changed("max-redemptions-per-account") {
				body.MaxRedemptionsPerAccount = recurly.Int(maxRedemptionsPerAccount)
			}
			if cmd.Flags().Changed("hosted-description") {
				body.HostedDescription = recurly.String(hostedDescription)
			}
			if cmd.Flags().Changed("invoice-description") {
				body.InvoiceDescription = recurly.String(invoiceDescription)
			}
			if cmd.Flags().Changed("redeem-by-date") {
				body.RedeemByDate = recurly.String(redeemByDate)
			}

			coupon, err := c.UpdateCoupon(args[0], body)
			if err != nil {
				return err
			}

			columns := couponDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, coupon)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Coupon name")
	cmd.Flags().IntVar(&maxRedemptions, "max-redemptions", 0, "Maximum number of redemptions")
	cmd.Flags().IntVar(&maxRedemptionsPerAccount, "max-redemptions-per-account", 0, "Maximum redemptions per account")
	cmd.Flags().StringVar(&hostedDescription, "hosted-description", "", "Hosted page description")
	cmd.Flags().StringVar(&invoiceDescription, "invoice-description", "", "Invoice description")
	cmd.Flags().StringVar(&redeemByDate, "redeem-by-date", "", "Coupon expiration date (YYYY-MM-DD)")

	return cmd
}

func couponDetailColumns() []output.Column {
	type C = *recurly.Coupon
	return output.ToColumns([]output.TypedColumn[C]{
		output.StringColumn[C]("ID", func(c C) string { return c.Id }),
		output.StringColumn[C]("Code", func(c C) string { return c.Code }),
		output.StringColumn[C]("Name", func(c C) string { return c.Name }),
		output.StringColumn[C]("State", func(c C) string { return c.State }),
		output.StringColumn[C]("Discount Type", func(c C) string { return c.Discount.Type }),
		{Header: "Discount Value", Extract: func(c C) string {
			d := c.Discount
			switch d.Type {
			case "percent":
				return fmt.Sprintf("%d%%", d.Percent)
			case "fixed":
				var parts []string
				for _, cur := range d.Currencies {
					parts = append(parts, fmt.Sprintf("%.2f %s", cur.Amount, cur.Currency))
				}
				return fmt.Sprintf("%v", parts)
			case "free_trial":
				return fmt.Sprintf("%d %s", d.Trial.Length, d.Trial.Unit)
			default:
				return ""
			}
		}},
		output.StringColumn[C]("Duration", func(c C) string { return c.Duration }),
		output.StringColumn[C]("Coupon Type", func(c C) string { return c.CouponType }),
		output.IntColumn[C]("Max Redemptions", func(c C) int { return c.MaxRedemptions }),
		output.IntColumn[C]("Max Redemptions Per Account", func(c C) int { return c.MaxRedemptionsPerAccount }),
		output.TimeColumn[C]("Redeem By", func(c C) *time.Time { return c.RedeemBy }),
		output.BoolColumn[C]("Applies To All Plans", func(c C) bool { return c.AppliesToAllPlans }),
		output.BoolColumn[C]("Applies To All Items", func(c C) bool { return c.AppliesToAllItems }),
		output.TimeColumn[C]("Created At", func(c C) *time.Time { return c.CreatedAt }),
		output.TimeColumn[C]("Updated At", func(c C) *time.Time { return c.UpdatedAt }),
	})
}

func newCouponsDeactivateCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "deactivate <coupon_id>",
		Short: "Deactivate a coupon",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			couponID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to deactivate coupon %s? [y/N] ", couponID))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deactivation cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewCouponAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			coupon, err := c.DeactivateCoupon(couponID)
			if err != nil {
				return err
			}

			columns := couponDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, coupon)
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

func newCouponsRestoreCmd() *cobra.Command {
	var (
		name                     string
		maxRedemptions           int
		maxRedemptionsPerAccount int
		hostedDescription        string
		invoiceDescription       string
		redeemByDate             string
	)

	cmd := &cobra.Command{
		Use:   "restore <coupon_id>",
		Short: "Restore an inactive coupon",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewCouponAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
			body := &recurly.CouponUpdate{}

			if cmd.Flags().Changed("name") {
				body.Name = recurly.String(name)
			}
			if cmd.Flags().Changed("max-redemptions") {
				body.MaxRedemptions = recurly.Int(maxRedemptions)
			}
			if cmd.Flags().Changed("max-redemptions-per-account") {
				body.MaxRedemptionsPerAccount = recurly.Int(maxRedemptionsPerAccount)
			}
			if cmd.Flags().Changed("hosted-description") {
				body.HostedDescription = recurly.String(hostedDescription)
			}
			if cmd.Flags().Changed("invoice-description") {
				body.InvoiceDescription = recurly.String(invoiceDescription)
			}
			if cmd.Flags().Changed("redeem-by-date") {
				body.RedeemByDate = recurly.String(redeemByDate)
			}

			coupon, err := c.RestoreCoupon(args[0], body)
			if err != nil {
				return err
			}

			columns := couponDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, coupon)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Coupon name")
	cmd.Flags().IntVar(&maxRedemptions, "max-redemptions", 0, "Maximum number of redemptions")
	cmd.Flags().IntVar(&maxRedemptionsPerAccount, "max-redemptions-per-account", 0, "Maximum redemptions per account")
	cmd.Flags().StringVar(&hostedDescription, "hosted-description", "", "Hosted page description")
	cmd.Flags().StringVar(&invoiceDescription, "invoice-description", "", "Invoice description")
	cmd.Flags().StringVar(&redeemByDate, "redeem-by-date", "", "Coupon expiration date (YYYY-MM-DD)")

	return cmd
}

func newCouponsGenerateCodesCmd() *cobra.Command {
	var numberOfCodes int

	cmd := &cobra.Command{
		Use:   "generate-codes <coupon_id>",
		Short: "Generate unique coupon codes for a bulk coupon",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if numberOfCodes < 1 {
				return fmt.Errorf("--number-of-codes must be at least 1")
			}

			c, err := AppFromContext(cmd.Context()).NewCouponAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &recurly.CouponBulkCreate{
				NumberOfUniqueCodes: recurly.Int(numberOfCodes),
			}

			result, err := c.GenerateUniqueCouponCodes(args[0], body)
			if err != nil {
				return err
			}

			columns := []output.Column{
				{Header: "Limit", Extract: func(v any) string {
					return strconv.Itoa(v.(*recurly.UniqueCouponCodeParams).Limit)
				}},
				{Header: "Order", Extract: func(v any) string {
					return v.(*recurly.UniqueCouponCodeParams).Order
				}},
				{Header: "Sort", Extract: func(v any) string {
					return v.(*recurly.UniqueCouponCodeParams).Sort
				}},
				{Header: "Begin Time", Extract: func(v any) string {
					p := v.(*recurly.UniqueCouponCodeParams)
					if p.BeginTime != nil {
						return p.BeginTime.Format(time.RFC3339)
					}
					return ""
				}},
			}

			formatted, err := output.FormatOne(cfg, columns, result)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().IntVar(&numberOfCodes, "number-of-codes", 0, "Number of unique codes to generate (required, minimum 1)")
	_ = cmd.MarkFlagRequired("number-of-codes")

	return cmd
}

func newCouponsListCodesCmd() *cobra.Command {
	var (
		limit    int
		all      bool
		order    string
		sort     string
		redeemed string
	)

	cmd := &cobra.Command{
		Use:   "list-codes <coupon_id>",
		Short: "List unique coupon codes for a bulk coupon",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewCouponAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			params := &recurly.ListUniqueCouponCodesParams{}

			if limit > 0 {
				params.Limit = recurly.Int(limit)
			}
			if cmd.Flags().Changed("order") {
				params.Order = recurly.String(order)
			}
			if cmd.Flags().Changed("sort") {
				params.Sort = recurly.String(sort)
			}
			if cmd.Flags().Changed("redeemed") {
				params.Redeemed = recurly.String(redeemed)
			}

			lister, err := c.ListUniqueCouponCodes(args[0], params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.UniqueCouponCode](lister, limit, all)
			if err != nil {
				return err
			}

			type U = recurly.UniqueCouponCode
			columns := output.ToColumns([]output.TypedColumn[U]{
				output.StringColumn[U]("Code", func(u U) string { return u.Code }),
				output.StringColumn[U]("State", func(u U) string { return u.State }),
				output.StringColumn[U]("Bulk Coupon Code", func(u U) string { return u.BulkCouponCode }),
				output.TimeColumn[U]("Redeemed At", func(u U) *time.Time { return u.RedeemedAt }),
				output.TimeColumn[U]("Created At", func(u U) *time.Time { return u.CreatedAt }),
			})

			items := make([]any, len(result.Items))
			for i, code := range result.Items {
				items[i] = code
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
	cmd.Flags().StringVar(&redeemed, "redeemed", "", "Filter by redemption status (true or false)")

	return cmd
}
