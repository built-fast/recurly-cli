package cmd

import (
	"fmt"
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
)

func newSubscriptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Manage subscriptions",
	}
	cmd.AddCommand(newSubscriptionsListCmd())
	cmd.AddCommand(newSubscriptionsGetCmd())
	cmd.AddCommand(newSubscriptionsCreateCmd())
	cmd.AddCommand(newSubscriptionsUpdateCmd())
	return cmd
}

func subscriptionDetailColumns() []output.Column {
	return []output.Column{
		{Header: "ID", Extract: func(v any) string { return v.(*recurly.Subscription).Id }},
		{Header: "UUID", Extract: func(v any) string { return v.(*recurly.Subscription).Uuid }},
		{Header: "Account Code", Extract: func(v any) string { return v.(*recurly.Subscription).Account.Code }},
		{Header: "Plan Code", Extract: func(v any) string { return v.(*recurly.Subscription).Plan.Code }},
		{Header: "Plan Name", Extract: func(v any) string { return v.(*recurly.Subscription).Plan.Name }},
		{Header: "State", Extract: func(v any) string { return v.(*recurly.Subscription).State }},
		{Header: "Currency", Extract: func(v any) string { return v.(*recurly.Subscription).Currency }},
		{Header: "Unit Amount", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Subscription).UnitAmount)
		}},
		{Header: "Quantity", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(*recurly.Subscription).Quantity)
		}},
		{Header: "Subtotal", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Subscription).Subtotal)
		}},
		{Header: "Tax", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Subscription).Tax)
		}},
		{Header: "Total", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Subscription).Total)
		}},
		{Header: "Collection Method", Extract: func(v any) string { return v.(*recurly.Subscription).CollectionMethod }},
		{Header: "Auto Renew", Extract: func(v any) string {
			return fmt.Sprintf("%t", v.(*recurly.Subscription).AutoRenew)
		}},
		{Header: "Current Period Started At", Extract: func(v any) string {
			s := v.(*recurly.Subscription)
			if s.CurrentPeriodStartedAt != nil {
				return s.CurrentPeriodStartedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Current Period Ends At", Extract: func(v any) string {
			s := v.(*recurly.Subscription)
			if s.CurrentPeriodEndsAt != nil {
				return s.CurrentPeriodEndsAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Trial Started At", Extract: func(v any) string {
			s := v.(*recurly.Subscription)
			if s.TrialStartedAt != nil {
				return s.TrialStartedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Trial Ends At", Extract: func(v any) string {
			s := v.(*recurly.Subscription)
			if s.TrialEndsAt != nil {
				return s.TrialEndsAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Paused At", Extract: func(v any) string {
			s := v.(*recurly.Subscription)
			if s.PausedAt != nil {
				return s.PausedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Remaining Pause Cycles", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(*recurly.Subscription).RemainingPauseCycles)
		}},
		{Header: "Net Terms", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(*recurly.Subscription).NetTerms)
		}},
		{Header: "Net Terms Type", Extract: func(v any) string { return v.(*recurly.Subscription).NetTermsType }},
		{Header: "PO Number", Extract: func(v any) string { return v.(*recurly.Subscription).PoNumber }},
		{Header: "Gateway Code", Extract: func(v any) string { return v.(*recurly.Subscription).GatewayCode }},
		{Header: "Billing Info ID", Extract: func(v any) string { return v.(*recurly.Subscription).BillingInfoId }},
		{Header: "Created At", Extract: func(v any) string {
			s := v.(*recurly.Subscription)
			if s.CreatedAt != nil {
				return s.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Updated At", Extract: func(v any) string {
			s := v.(*recurly.Subscription)
			if s.UpdatedAt != nil {
				return s.UpdatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Activated At", Extract: func(v any) string {
			s := v.(*recurly.Subscription)
			if s.ActivatedAt != nil {
				return s.ActivatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Canceled At", Extract: func(v any) string {
			s := v.(*recurly.Subscription)
			if s.CanceledAt != nil {
				return s.CanceledAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Expires At", Extract: func(v any) string {
			s := v.(*recurly.Subscription)
			if s.ExpiresAt != nil {
				return s.ExpiresAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}

func newSubscriptionsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <subscription_id>",
		Short: "Get subscription details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newSubscriptionAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			subscription, err := c.GetSubscription(args[0])
			if err != nil {
				return err
			}

			columns := subscriptionDetailColumns()

			formatted, err := output.FormatOne(format, columns, subscription)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newSubscriptionsListCmd() *cobra.Command {
	var (
		limit     int
		all       bool
		order     string
		sort      string
		state     string
		planID    string
		beginTime string
		endTime   string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newSubscriptionAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			params := &recurly.ListSubscriptionsParams{}

			if limit > 0 {
				params.Limit = recurly.Int(limit)
			}
			if cmd.Flags().Changed("order") {
				params.Order = recurly.String(order)
			}
			if cmd.Flags().Changed("sort") {
				params.Sort = recurly.String(sort)
			}
			if cmd.Flags().Changed("state") {
				params.State = recurly.String(state)
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

			lister, err := c.ListSubscriptions(params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.Subscription](lister, limit, all)
			if err != nil {
				return err
			}

			// Client-side filter by plan ID (not supported by SDK params)
			if cmd.Flags().Changed("plan-id") {
				filtered := make([]recurly.Subscription, 0, len(result.Items))
				for _, s := range result.Items {
					if s.Plan.Id == planID {
						filtered = append(filtered, s)
					}
				}
				result.Items = filtered
			}

			columns := []output.Column{
				{Header: "ID", Extract: func(v any) string { return v.(recurly.Subscription).Id }},
				{Header: "Account Code", Extract: func(v any) string { return v.(recurly.Subscription).Account.Code }},
				{Header: "Plan Code", Extract: func(v any) string { return v.(recurly.Subscription).Plan.Code }},
				{Header: "State", Extract: func(v any) string { return v.(recurly.Subscription).State }},
				{Header: "Currency", Extract: func(v any) string { return v.(recurly.Subscription).Currency }},
				{Header: "Unit Amount", Extract: func(v any) string {
					return fmt.Sprintf("%.2f", v.(recurly.Subscription).UnitAmount)
				}},
				{Header: "Current Period Ends At", Extract: func(v any) string {
					s := v.(recurly.Subscription)
					if s.CurrentPeriodEndsAt != nil {
						return s.CurrentPeriodEndsAt.Format(time.RFC3339)
					}
					return ""
				}},
			}

			items := make([]any, len(result.Items))
			for i, s := range result.Items {
				items[i] = s
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
	cmd.Flags().StringVar(&state, "state", "", "Filter by state (active, canceled, expired, future, in_trial, live)")
	cmd.Flags().StringVar(&planID, "plan-id", "", "Filter by plan ID")
	cmd.Flags().StringVar(&beginTime, "begin-time", "", "Filter by begin time (ISO8601 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter by end time (ISO8601 format)")

	return cmd
}

func newSubscriptionsCreateCmd() *cobra.Command {
	var (
		planCode             string
		accountCode          string
		currency             string
		quantity             int
		unitAmount           float64
		autoRenew            bool
		trialEndsAt          string
		startsAt             string
		nextBillDate         string
		collectionMethod     string
		poNumber             string
		netTerms             int
		netTermsType         string
		totalBillingCycles   int
		renewalBillingCycles int
		couponCode           string
		gatewayCode          string
		billingInfoID        string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newSubscriptionAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")
			body := &recurly.SubscriptionCreate{}

			if cmd.Flags().Changed("plan-code") {
				body.PlanCode = recurly.String(planCode)
			}
			if cmd.Flags().Changed("account-code") {
				body.Account = &recurly.AccountCreate{
					Code: recurly.String(accountCode),
				}
			}
			if cmd.Flags().Changed("currency") {
				body.Currency = recurly.String(currency)
			}
			if cmd.Flags().Changed("quantity") {
				body.Quantity = recurly.Int(quantity)
			}
			if cmd.Flags().Changed("unit-amount") {
				body.UnitAmount = float64Ptr(unitAmount)
			}
			if cmd.Flags().Changed("auto-renew") {
				body.AutoRenew = recurly.Bool(autoRenew)
			}

			// Timestamp flags
			if cmd.Flags().Changed("trial-ends-at") {
				t, err := time.Parse(time.RFC3339, trialEndsAt)
				if err != nil {
					return fmt.Errorf("invalid --trial-ends-at: must be RFC3339 format (e.g. 2025-01-01T00:00:00Z): %w", err)
				}
				body.TrialEndsAt = &t
			}
			if cmd.Flags().Changed("starts-at") {
				t, err := time.Parse(time.RFC3339, startsAt)
				if err != nil {
					return fmt.Errorf("invalid --starts-at: must be RFC3339 format (e.g. 2025-01-01T00:00:00Z): %w", err)
				}
				body.StartsAt = &t
			}
			if cmd.Flags().Changed("next-bill-date") {
				t, err := time.Parse(time.RFC3339, nextBillDate)
				if err != nil {
					return fmt.Errorf("invalid --next-bill-date: must be RFC3339 format (e.g. 2025-01-01T00:00:00Z): %w", err)
				}
				body.NextBillDate = &t
			}

			// Billing flags
			if cmd.Flags().Changed("collection-method") {
				body.CollectionMethod = recurly.String(collectionMethod)
			}
			if cmd.Flags().Changed("po-number") {
				body.PoNumber = recurly.String(poNumber)
			}
			if cmd.Flags().Changed("net-terms") {
				body.NetTerms = recurly.Int(netTerms)
			}
			if cmd.Flags().Changed("net-terms-type") {
				body.NetTermsType = recurly.String(netTermsType)
			}
			if cmd.Flags().Changed("total-billing-cycles") {
				body.TotalBillingCycles = recurly.Int(totalBillingCycles)
			}
			if cmd.Flags().Changed("renewal-billing-cycles") {
				body.RenewalBillingCycles = recurly.Int(renewalBillingCycles)
			}

			// Additional flags
			if cmd.Flags().Changed("coupon-code") {
				codes := []string{couponCode}
				body.CouponCodes = &codes
			}
			if cmd.Flags().Changed("gateway-code") {
				body.GatewayCode = recurly.String(gatewayCode)
			}
			if cmd.Flags().Changed("billing-info-id") {
				body.BillingInfoId = recurly.String(billingInfoID)
			}

			subscription, err := c.CreateSubscription(body)
			if err != nil {
				return err
			}

			columns := subscriptionDetailColumns()

			formatted, err := output.FormatOne(format, columns, subscription)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	// Core flags
	cmd.Flags().StringVar(&planCode, "plan-code", "", "Plan code for the subscription")
	cmd.Flags().StringVar(&accountCode, "account-code", "", "Account code for the subscription")
	cmd.Flags().StringVar(&currency, "currency", "", "3-letter ISO 4217 currency code")
	cmd.Flags().IntVar(&quantity, "quantity", 0, "Subscription quantity")
	cmd.Flags().Float64Var(&unitAmount, "unit-amount", 0, "Override the plan's unit amount")
	cmd.Flags().BoolVar(&autoRenew, "auto-renew", false, "Whether the subscription renews at the end of its term")

	// Timing flags
	cmd.Flags().StringVar(&trialEndsAt, "trial-ends-at", "", "Trial end date (RFC3339 format)")
	cmd.Flags().StringVar(&startsAt, "starts-at", "", "Subscription start date (RFC3339 format)")
	cmd.Flags().StringVar(&nextBillDate, "next-bill-date", "", "Next bill date (RFC3339 format)")

	// Billing flags
	cmd.Flags().StringVar(&collectionMethod, "collection-method", "", "Collection method (automatic or manual)")
	cmd.Flags().StringVar(&poNumber, "po-number", "", "PO number for manual invoicing")
	cmd.Flags().IntVar(&netTerms, "net-terms", 0, "Net terms in days")
	cmd.Flags().StringVar(&netTermsType, "net-terms-type", "", "Net terms type (net or eom)")
	cmd.Flags().IntVar(&totalBillingCycles, "total-billing-cycles", 0, "Total billing cycles in a term")
	cmd.Flags().IntVar(&renewalBillingCycles, "renewal-billing-cycles", 0, "Billing cycles for renewal terms")

	// Additional flags
	cmd.Flags().StringVar(&couponCode, "coupon-code", "", "Coupon code to apply")
	cmd.Flags().StringVar(&gatewayCode, "gateway-code", "", "Payment gateway code")
	cmd.Flags().StringVar(&billingInfoID, "billing-info-id", "", "Billing info ID")

	return cmd
}

func newSubscriptionsUpdateCmd() *cobra.Command {
	var (
		collectionMethod     string
		remainingBillingCycles int
		renewalBillingCycles int
		autoRenew            bool
		nextBillDate         string
		revenueScheduleType  string
		termsAndConditions   string
		customerNotes        string
		poNumber             string
		netTerms             int
		netTermsType         string
		gatewayCode          string
		billingInfoID        string
	)

	cmd := &cobra.Command{
		Use:   "update <subscription_id>",
		Short: "Update a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newSubscriptionAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")
			body := &recurly.SubscriptionUpdate{}

			if cmd.Flags().Changed("collection-method") {
				body.CollectionMethod = recurly.String(collectionMethod)
			}
			if cmd.Flags().Changed("remaining-billing-cycles") {
				body.RemainingBillingCycles = recurly.Int(remainingBillingCycles)
			}
			if cmd.Flags().Changed("renewal-billing-cycles") {
				body.RenewalBillingCycles = recurly.Int(renewalBillingCycles)
			}
			if cmd.Flags().Changed("auto-renew") {
				body.AutoRenew = recurly.Bool(autoRenew)
			}
			if cmd.Flags().Changed("next-bill-date") {
				t, err := time.Parse(time.RFC3339, nextBillDate)
				if err != nil {
					return fmt.Errorf("invalid --next-bill-date: must be RFC3339 format (e.g. 2025-01-01T00:00:00Z): %w", err)
				}
				body.NextBillDate = &t
			}
			if cmd.Flags().Changed("revenue-schedule-type") {
				body.RevenueScheduleType = recurly.String(revenueScheduleType)
			}
			if cmd.Flags().Changed("terms-and-conditions") {
				body.TermsAndConditions = recurly.String(termsAndConditions)
			}
			if cmd.Flags().Changed("customer-notes") {
				body.CustomerNotes = recurly.String(customerNotes)
			}
			if cmd.Flags().Changed("po-number") {
				body.PoNumber = recurly.String(poNumber)
			}
			if cmd.Flags().Changed("net-terms") {
				body.NetTerms = recurly.Int(netTerms)
			}
			if cmd.Flags().Changed("net-terms-type") {
				body.NetTermsType = recurly.String(netTermsType)
			}
			if cmd.Flags().Changed("gateway-code") {
				body.GatewayCode = recurly.String(gatewayCode)
			}
			if cmd.Flags().Changed("billing-info-id") {
				body.BillingInfoId = recurly.String(billingInfoID)
			}

			subscription, err := c.UpdateSubscription(args[0], body)
			if err != nil {
				return err
			}

			columns := subscriptionDetailColumns()

			formatted, err := output.FormatOne(format, columns, subscription)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&collectionMethod, "collection-method", "", "Collection method (automatic or manual)")
	cmd.Flags().IntVar(&remainingBillingCycles, "remaining-billing-cycles", 0, "Remaining billing cycles in the current term")
	cmd.Flags().IntVar(&renewalBillingCycles, "renewal-billing-cycles", 0, "Billing cycles for renewal terms")
	cmd.Flags().BoolVar(&autoRenew, "auto-renew", false, "Whether the subscription renews at the end of its term")
	cmd.Flags().StringVar(&nextBillDate, "next-bill-date", "", "Next bill date (RFC3339 format)")
	cmd.Flags().StringVar(&revenueScheduleType, "revenue-schedule-type", "", "Revenue schedule type")
	cmd.Flags().StringVar(&termsAndConditions, "terms-and-conditions", "", "Custom terms and conditions")
	cmd.Flags().StringVar(&customerNotes, "customer-notes", "", "Custom customer notes")
	cmd.Flags().StringVar(&poNumber, "po-number", "", "PO number for manual invoicing")
	cmd.Flags().IntVar(&netTerms, "net-terms", 0, "Net terms in days")
	cmd.Flags().StringVar(&netTermsType, "net-terms-type", "", "Net terms type (net or eom)")
	cmd.Flags().StringVar(&gatewayCode, "gateway-code", "", "Payment gateway code")
	cmd.Flags().StringVar(&billingInfoID, "billing-info-id", "", "Billing info ID")

	return cmd
}
