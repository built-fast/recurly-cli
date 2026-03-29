package cmd

import (
	"fmt"
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
)

func newSubscriptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Manage subscriptions",
	}
	cmd.AddCommand(newSubscriptionsListCmd())
	cmd.AddCommand(withWatch(newSubscriptionsGetCmd()))
	cmd.AddCommand(withFromFile(newSubscriptionsCreateCmd()))
	cmd.AddCommand(withFromFile(newSubscriptionsUpdateCmd()))
	cmd.AddCommand(newSubscriptionsCancelCmd())
	cmd.AddCommand(newSubscriptionsReactivateCmd())
	cmd.AddCommand(withInteractive(newSubscriptionsPauseCmd()))
	cmd.AddCommand(newSubscriptionsResumeCmd())
	cmd.AddCommand(newSubscriptionsTerminateCmd())
	cmd.AddCommand(newSubscriptionsConvertTrialCmd())
	return cmd
}

func subscriptionDetailColumns() []output.Column {
	type S = *recurly.Subscription
	return output.ToColumns([]output.TypedColumn[S]{
		output.StringColumn[S]("ID", func(s S) string { return s.Id }),
		output.StringColumn[S]("UUID", func(s S) string { return s.Uuid }),
		output.StringColumn[S]("Account Code", func(s S) string { return s.Account.Code }),
		output.StringColumn[S]("Plan Code", func(s S) string { return s.Plan.Code }),
		output.StringColumn[S]("Plan Name", func(s S) string { return s.Plan.Name }),
		output.StringColumn[S]("State", func(s S) string { return s.State }),
		output.StringColumn[S]("Currency", func(s S) string { return s.Currency }),
		output.FloatColumn[S]("Unit Amount", func(s S) float64 { return s.UnitAmount }),
		output.IntColumn[S]("Quantity", func(s S) int { return s.Quantity }),
		output.FloatColumn[S]("Subtotal", func(s S) float64 { return s.Subtotal }),
		output.FloatColumn[S]("Tax", func(s S) float64 { return s.Tax }),
		output.FloatColumn[S]("Total", func(s S) float64 { return s.Total }),
		output.StringColumn[S]("Collection Method", func(s S) string { return s.CollectionMethod }),
		output.BoolColumn[S]("Auto Renew", func(s S) bool { return s.AutoRenew }),
		output.TimeColumn[S]("Current Period Started At", func(s S) *time.Time { return s.CurrentPeriodStartedAt }),
		output.TimeColumn[S]("Current Period Ends At", func(s S) *time.Time { return s.CurrentPeriodEndsAt }),
		output.TimeColumn[S]("Trial Started At", func(s S) *time.Time { return s.TrialStartedAt }),
		output.TimeColumn[S]("Trial Ends At", func(s S) *time.Time { return s.TrialEndsAt }),
		output.TimeColumn[S]("Paused At", func(s S) *time.Time { return s.PausedAt }),
		output.IntColumn[S]("Remaining Pause Cycles", func(s S) int { return s.RemainingPauseCycles }),
		output.IntColumn[S]("Net Terms", func(s S) int { return s.NetTerms }),
		output.StringColumn[S]("Net Terms Type", func(s S) string { return s.NetTermsType }),
		output.StringColumn[S]("PO Number", func(s S) string { return s.PoNumber }),
		output.StringColumn[S]("Gateway Code", func(s S) string { return s.GatewayCode }),
		output.StringColumn[S]("Billing Info ID", func(s S) string { return s.BillingInfoId }),
		output.TimeColumn[S]("Created At", func(s S) *time.Time { return s.CreatedAt }),
		output.TimeColumn[S]("Updated At", func(s S) *time.Time { return s.UpdatedAt }),
		output.TimeColumn[S]("Activated At", func(s S) *time.Time { return s.ActivatedAt }),
		output.TimeColumn[S]("Canceled At", func(s S) *time.Time { return s.CanceledAt }),
		output.TimeColumn[S]("Expires At", func(s S) *time.Time { return s.ExpiresAt }),
	})
}

func newSubscriptionsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <subscription_id>",
		Short: "Get subscription details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewSubscriptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			subscription, err := c.GetSubscription(args[0])
			if err != nil {
				return err
			}

			columns := subscriptionDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, subscription)
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
			c, err := AppFromContext(cmd.Context()).NewSubscriptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			type S = recurly.Subscription
			columns := output.ToColumns([]output.TypedColumn[S]{
				output.StringColumn[S]("ID", func(s S) string { return s.Id }),
				output.StringColumn[S]("Account Code", func(s S) string { return s.Account.Code }),
				output.StringColumn[S]("Plan Code", func(s S) string { return s.Plan.Code }),
				output.StringColumn[S]("State", func(s S) string { return s.State }),
				output.StringColumn[S]("Currency", func(s S) string { return s.Currency }),
				output.FloatColumn[S]("Unit Amount", func(s S) float64 { return s.UnitAmount }),
				output.TimeColumn[S]("Current Period Ends At", func(s S) *time.Time { return s.CurrentPeriodEndsAt }),
			})

			items := make([]any, len(result.Items))
			for i, s := range result.Items {
				items[i] = s
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
			c, err := AppFromContext(cmd.Context()).NewSubscriptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
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

			formatted, err := output.FormatOne(cfg, columns, subscription)
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
		collectionMethod       string
		remainingBillingCycles int
		renewalBillingCycles   int
		autoRenew              bool
		nextBillDate           string
		revenueScheduleType    string
		termsAndConditions     string
		customerNotes          string
		poNumber               string
		netTerms               int
		netTermsType           string
		gatewayCode            string
		billingInfoID          string
	)

	cmd := &cobra.Command{
		Use:   "update <subscription_id>",
		Short: "Update a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewSubscriptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
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

			formatted, err := output.FormatOne(cfg, columns, subscription)
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

func newSubscriptionsCancelCmd() *cobra.Command {
	var (
		yes       bool
		timeframe string
	)

	cmd := &cobra.Command{
		Use:   "cancel <subscription_id>",
		Short: "Cancel a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subscriptionID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, "Are you sure you want to cancel this subscription? [y/N] ")
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Cancellation cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewSubscriptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			params := &recurly.CancelSubscriptionParams{}
			if cmd.Flags().Changed("timeframe") {
				params.Body = &recurly.SubscriptionCancel{
					Timeframe: recurly.String(timeframe),
				}
			}

			subscription, err := c.CancelSubscription(subscriptionID, params)
			if err != nil {
				return err
			}

			columns := subscriptionDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, subscription)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&timeframe, "timeframe", "", "Cancellation timeframe (bill_date or term_end)")

	return cmd
}

func newSubscriptionsReactivateCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "reactivate <subscription_id>",
		Short: "Reactivate a canceled subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subscriptionID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, "Are you sure you want to reactivate this subscription? [y/N] ")
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Reactivation cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewSubscriptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			subscription, err := c.ReactivateSubscription(subscriptionID)
			if err != nil {
				return err
			}

			columns := subscriptionDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, subscription)
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

func newSubscriptionsPauseCmd() *cobra.Command {
	var (
		yes                  bool
		remainingPauseCycles int
	)

	cmd := &cobra.Command{
		Use:   "pause <subscription_id>",
		Short: "Pause a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subscriptionID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, "Are you sure you want to pause this subscription? [y/N] ")
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Pause cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewSubscriptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &recurly.SubscriptionPause{
				RemainingPauseCycles: recurly.Int(remainingPauseCycles),
			}

			subscription, err := c.PauseSubscription(subscriptionID, body)
			if err != nil {
				return err
			}

			columns := subscriptionDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, subscription)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().IntVar(&remainingPauseCycles, "remaining-pause-cycles", 0, "Number of billing cycles to pause")
	_ = cmd.MarkFlagRequired("remaining-pause-cycles")

	return cmd
}

func newSubscriptionsResumeCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "resume <subscription_id>",
		Short: "Resume a paused subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subscriptionID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, "Are you sure you want to resume this subscription? [y/N] ")
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Resume cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewSubscriptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			subscription, err := c.ResumeSubscription(subscriptionID)
			if err != nil {
				return err
			}

			columns := subscriptionDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, subscription)
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

func newSubscriptionsTerminateCmd() *cobra.Command {
	var (
		yes    bool
		refund string
		charge bool
	)

	cmd := &cobra.Command{
		Use:   "terminate <subscription_id>",
		Short: "Terminate a subscription immediately",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subscriptionID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, "Are you sure you want to terminate this subscription? This cannot be undone. [y/N] ")
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Termination cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewSubscriptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			params := &recurly.TerminateSubscriptionParams{}
			if cmd.Flags().Changed("refund") {
				params.Refund = recurly.String(refund)
			}
			if cmd.Flags().Changed("charge") {
				params.Charge = recurly.Bool(charge)
			}

			subscription, err := c.TerminateSubscription(subscriptionID, params)
			if err != nil {
				return err
			}

			columns := subscriptionDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, subscription)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&refund, "refund", "", "Refund type (full, partial, or none)")
	cmd.Flags().BoolVar(&charge, "charge", false, "Invoice unbilled usage on the final invoice")

	return cmd
}

func newSubscriptionsConvertTrialCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "convert-trial <subscription_id>",
		Short: "Convert a trial subscription to a paid subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subscriptionID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, "Are you sure you want to convert this trial subscription? [y/N] ")
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Conversion cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewSubscriptionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			subscription, err := c.ConvertTrial(subscriptionID)
			if err != nil {
				return err
			}

			columns := subscriptionDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, subscription)
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
