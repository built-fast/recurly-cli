package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newPlansCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Short: "Manage plans",
	}
	cmd.AddCommand(newPlansListCmd())
	cmd.AddCommand(withWatch(newPlansGetCmd()))
	cmd.AddCommand(withFromFile(newPlansCreateCmd()))
	cmd.AddCommand(withFromFile(newPlansUpdateCmd()))
	cmd.AddCommand(newPlansDeactivateCmd())
	cmd.AddCommand(newPlanAddOnsCmd())
	return cmd
}

func planDetailColumns() []output.Column {
	return []output.Column{
		{Header: "Id", Extract: func(v any) string { return v.(*recurly.Plan).Id }},
		{Header: "Code", Extract: func(v any) string { return v.(*recurly.Plan).Code }},
		{Header: "Name", Extract: func(v any) string { return v.(*recurly.Plan).Name }},
		{Header: "State", Extract: func(v any) string { return v.(*recurly.Plan).State }},
		{Header: "Pricing Model", Extract: func(v any) string { return v.(*recurly.Plan).PricingModel }},
		{Header: "Interval Unit", Extract: func(v any) string { return v.(*recurly.Plan).IntervalUnit }},
		{Header: "Interval Length", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(*recurly.Plan).IntervalLength)
		}},
		{Header: "Description", Extract: func(v any) string { return v.(*recurly.Plan).Description }},
		{Header: "Currencies", Extract: func(v any) string {
			p := v.(*recurly.Plan)
			if len(p.Currencies) == 0 {
				return ""
			}
			var parts []string
			for _, c := range p.Currencies {
				parts = append(parts, fmt.Sprintf("%s: %.2f (setup: %.2f)", c.Currency, c.UnitAmount, c.SetupFee))
			}
			return strings.Join(parts, ", ")
		}},
		{Header: "Trial Unit", Extract: func(v any) string { return v.(*recurly.Plan).TrialUnit }},
		{Header: "Trial Length", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(*recurly.Plan).TrialLength)
		}},
		{Header: "Auto Renew", Extract: func(v any) string {
			return fmt.Sprintf("%t", v.(*recurly.Plan).AutoRenew)
		}},
		{Header: "Total Billing Cycles", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(*recurly.Plan).TotalBillingCycles)
		}},
		{Header: "Tax Code", Extract: func(v any) string { return v.(*recurly.Plan).TaxCode }},
		{Header: "Tax Exempt", Extract: func(v any) string {
			return fmt.Sprintf("%t", v.(*recurly.Plan).TaxExempt)
		}},
		{Header: "Created At", Extract: func(v any) string {
			p := v.(*recurly.Plan)
			if p.CreatedAt != nil {
				return p.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Updated At", Extract: func(v any) string {
			p := v.(*recurly.Plan)
			if p.UpdatedAt != nil {
				return p.UpdatedAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}

func newPlansCreateCmd() *cobra.Command {
	var (
		// Core
		code         string
		name         string
		intervalUnit string
		intervalLen  int
		description  string
		pricingModel string

		// Multi-currency (repeatable slices, positionally matched)
		currencies  []string
		unitAmounts []float64
		setupFees   []float64

		// Trial
		trialUnit                string
		trialLength              int
		trialRequiresBillingInfo bool

		// Billing
		autoRenew          bool
		totalBillingCycles int

		// Tax
		taxCode              string
		taxExempt            bool
		avalaraTransType     int
		avalaraServiceType   int
		vertexTransType      string
		harmonizedSystemCode string

		// Accounting
		accountingCode                  string
		revenueScheduleType             string
		liabilityGlAccountId            string
		revenueGlAccountId              string
		performanceObligationId         string
		setupFeeAccountingCode          string
		setupFeeRevenueScheduleType     string
		setupFeeLiabilityGlAccountId    string
		setupFeeRevenueGlAccountId      string
		setupFeePerformanceObligationId string

		// Hosted pages
		successUrl         string
		cancelUrl          string
		bypassConfirmation bool
		displayQuantity    bool

		// Other
		allowAnyItemOnSubscriptions bool
		dunningCampaignId           string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate currency/unit-amount pairing
			if cmd.Flags().Changed("currency") || cmd.Flags().Changed("unit-amount") {
				if len(currencies) != len(unitAmounts) {
					return fmt.Errorf("number of --currency values must match --unit-amount values")
				}
			}
			if cmd.Flags().Changed("setup-fee") {
				if len(setupFees) != len(currencies) {
					return fmt.Errorf("number of --setup-fee values must match --currency values")
				}
			}

			c, err := newPlanAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")
			body := &recurly.PlanCreate{}

			// Core flags
			if cmd.Flags().Changed("code") {
				body.Code = recurly.String(code)
			}
			if cmd.Flags().Changed("name") {
				body.Name = recurly.String(name)
			}
			if cmd.Flags().Changed("interval-unit") {
				body.IntervalUnit = recurly.String(intervalUnit)
			}
			if cmd.Flags().Changed("interval-length") {
				body.IntervalLength = recurly.Int(intervalLen)
			}
			if cmd.Flags().Changed("description") {
				body.Description = recurly.String(description)
			}
			if cmd.Flags().Changed("pricing-model") {
				body.PricingModel = recurly.String(pricingModel)
			}

			// Multi-currency
			if cmd.Flags().Changed("currency") {
				pricings := make([]recurly.PlanPricingCreate, len(currencies))
				for i, cur := range currencies {
					pricings[i] = recurly.PlanPricingCreate{
						Currency:   recurly.String(cur),
						UnitAmount: float64Ptr(unitAmounts[i]),
					}
				}
				body.Currencies = &pricings
			}

			// Setup fees (separate top-level field)
			if cmd.Flags().Changed("setup-fee") {
				fees := make([]recurly.PlanSetupPricingCreate, len(currencies))
				for i, cur := range currencies {
					fees[i] = recurly.PlanSetupPricingCreate{
						Currency:   recurly.String(cur),
						UnitAmount: float64Ptr(setupFees[i]),
					}
				}
				body.SetupFees = &fees
			}

			// Trial flags
			if cmd.Flags().Changed("trial-unit") {
				body.TrialUnit = recurly.String(trialUnit)
			}
			if cmd.Flags().Changed("trial-length") {
				body.TrialLength = recurly.Int(trialLength)
			}
			if cmd.Flags().Changed("trial-requires-billing-info") {
				body.TrialRequiresBillingInfo = recurly.Bool(trialRequiresBillingInfo)
			}

			// Billing flags
			if cmd.Flags().Changed("auto-renew") {
				body.AutoRenew = recurly.Bool(autoRenew)
			}
			if cmd.Flags().Changed("total-billing-cycles") {
				body.TotalBillingCycles = recurly.Int(totalBillingCycles)
			}

			// Tax flags
			if cmd.Flags().Changed("tax-code") {
				body.TaxCode = recurly.String(taxCode)
			}
			if cmd.Flags().Changed("tax-exempt") {
				body.TaxExempt = recurly.Bool(taxExempt)
			}
			if cmd.Flags().Changed("avalara-transaction-type") {
				body.AvalaraTransactionType = recurly.Int(avalaraTransType)
			}
			if cmd.Flags().Changed("avalara-service-type") {
				body.AvalaraServiceType = recurly.Int(avalaraServiceType)
			}
			if cmd.Flags().Changed("vertex-transaction-type") {
				body.VertexTransactionType = recurly.String(vertexTransType)
			}
			if cmd.Flags().Changed("harmonized-system-code") {
				body.HarmonizedSystemCode = recurly.String(harmonizedSystemCode)
			}

			// Accounting flags
			if cmd.Flags().Changed("accounting-code") {
				body.AccountingCode = recurly.String(accountingCode)
			}
			if cmd.Flags().Changed("revenue-schedule-type") {
				body.RevenueScheduleType = recurly.String(revenueScheduleType)
			}
			if cmd.Flags().Changed("liability-gl-account-id") {
				body.LiabilityGlAccountId = recurly.String(liabilityGlAccountId)
			}
			if cmd.Flags().Changed("revenue-gl-account-id") {
				body.RevenueGlAccountId = recurly.String(revenueGlAccountId)
			}
			if cmd.Flags().Changed("performance-obligation-id") {
				body.PerformanceObligationId = recurly.String(performanceObligationId)
			}
			if cmd.Flags().Changed("setup-fee-accounting-code") {
				body.SetupFeeAccountingCode = recurly.String(setupFeeAccountingCode)
			}
			if cmd.Flags().Changed("setup-fee-revenue-schedule-type") {
				body.SetupFeeRevenueScheduleType = recurly.String(setupFeeRevenueScheduleType)
			}
			if cmd.Flags().Changed("setup-fee-liability-gl-account-id") {
				body.SetupFeeLiabilityGlAccountId = recurly.String(setupFeeLiabilityGlAccountId)
			}
			if cmd.Flags().Changed("setup-fee-revenue-gl-account-id") {
				body.SetupFeeRevenueGlAccountId = recurly.String(setupFeeRevenueGlAccountId)
			}
			if cmd.Flags().Changed("setup-fee-performance-obligation-id") {
				body.SetupFeePerformanceObligationId = recurly.String(setupFeePerformanceObligationId)
			}

			// Hosted pages flags
			hasHostedPages := cmd.Flags().Changed("success-url") || cmd.Flags().Changed("cancel-url") ||
				cmd.Flags().Changed("bypass-confirmation") || cmd.Flags().Changed("display-quantity")
			if hasHostedPages {
				hp := &recurly.PlanHostedPagesCreate{}
				if cmd.Flags().Changed("success-url") {
					hp.SuccessUrl = recurly.String(successUrl)
				}
				if cmd.Flags().Changed("cancel-url") {
					hp.CancelUrl = recurly.String(cancelUrl)
				}
				if cmd.Flags().Changed("bypass-confirmation") {
					hp.BypassConfirmation = recurly.Bool(bypassConfirmation)
				}
				if cmd.Flags().Changed("display-quantity") {
					hp.DisplayQuantity = recurly.Bool(displayQuantity)
				}
				body.HostedPages = hp
			}

			// Other flags
			if cmd.Flags().Changed("allow-any-item-on-subscriptions") {
				body.AllowAnyItemOnSubscriptions = recurly.Bool(allowAnyItemOnSubscriptions)
			}
			if cmd.Flags().Changed("dunning-campaign-id") {
				body.DunningCampaignId = recurly.String(dunningCampaignId)
			}

			plan, err := c.CreatePlan(body)
			if err != nil {
				return err
			}

			columns := planDetailColumns()

			formatted, err := output.FormatOne(format, columns, plan)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	// Core flags
	cmd.Flags().StringVar(&code, "code", "", "Unique plan code")
	cmd.Flags().StringVar(&name, "name", "", "Plan name")
	cmd.Flags().StringVar(&intervalUnit, "interval-unit", "", "Billing interval unit (day, week, month)")
	cmd.Flags().IntVar(&intervalLen, "interval-length", 0, "Billing interval length")
	cmd.Flags().StringVar(&description, "description", "", "Plan description")
	cmd.Flags().StringVar(&pricingModel, "pricing-model", "", "Pricing model (fixed or ramp)")

	// Multi-currency flags
	cmd.Flags().StringSliceVar(&currencies, "currency", nil, "Currency code (repeatable, positionally matched with --unit-amount)")
	cmd.Flags().Float64SliceVar(&unitAmounts, "unit-amount", nil, "Unit amount (repeatable, positionally matched with --currency)")
	cmd.Flags().Float64SliceVar(&setupFees, "setup-fee", nil, "Setup fee (repeatable, positionally matched with --currency)")

	// Trial flags
	cmd.Flags().StringVar(&trialUnit, "trial-unit", "", "Trial period unit (day, week, month)")
	cmd.Flags().IntVar(&trialLength, "trial-length", 0, "Trial period length")
	cmd.Flags().BoolVar(&trialRequiresBillingInfo, "trial-requires-billing-info", false, "Require billing info for trial")

	// Billing flags
	cmd.Flags().BoolVar(&autoRenew, "auto-renew", false, "Auto-renew subscriptions")
	cmd.Flags().IntVar(&totalBillingCycles, "total-billing-cycles", 0, "Total billing cycles before auto-termination")

	// Tax flags
	cmd.Flags().StringVar(&taxCode, "tax-code", "", "Tax code")
	cmd.Flags().BoolVar(&taxExempt, "tax-exempt", false, "Tax exempt status")
	cmd.Flags().IntVar(&avalaraTransType, "avalara-transaction-type", 0, "Avalara transaction type")
	cmd.Flags().IntVar(&avalaraServiceType, "avalara-service-type", 0, "Avalara service type")
	cmd.Flags().StringVar(&vertexTransType, "vertex-transaction-type", "", "Vertex transaction type (sale, rental, lease)")
	cmd.Flags().StringVar(&harmonizedSystemCode, "harmonized-system-code", "", "Harmonized System (HS) code")

	// Accounting flags
	cmd.Flags().StringVar(&accountingCode, "accounting-code", "", "Accounting code")
	cmd.Flags().StringVar(&revenueScheduleType, "revenue-schedule-type", "", "Revenue schedule type")
	cmd.Flags().StringVar(&liabilityGlAccountId, "liability-gl-account-id", "", "Liability GL account ID")
	cmd.Flags().StringVar(&revenueGlAccountId, "revenue-gl-account-id", "", "Revenue GL account ID")
	cmd.Flags().StringVar(&performanceObligationId, "performance-obligation-id", "", "Performance obligation ID")
	cmd.Flags().StringVar(&setupFeeAccountingCode, "setup-fee-accounting-code", "", "Setup fee accounting code")
	cmd.Flags().StringVar(&setupFeeRevenueScheduleType, "setup-fee-revenue-schedule-type", "", "Setup fee revenue schedule type")
	cmd.Flags().StringVar(&setupFeeLiabilityGlAccountId, "setup-fee-liability-gl-account-id", "", "Setup fee liability GL account ID")
	cmd.Flags().StringVar(&setupFeeRevenueGlAccountId, "setup-fee-revenue-gl-account-id", "", "Setup fee revenue GL account ID")
	cmd.Flags().StringVar(&setupFeePerformanceObligationId, "setup-fee-performance-obligation-id", "", "Setup fee performance obligation ID")

	// Hosted pages flags
	cmd.Flags().StringVar(&successUrl, "success-url", "", "Hosted page success redirect URL")
	cmd.Flags().StringVar(&cancelUrl, "cancel-url", "", "Hosted page cancel redirect URL")
	cmd.Flags().BoolVar(&bypassConfirmation, "bypass-confirmation", false, "Bypass hosted page confirmation")
	cmd.Flags().BoolVar(&displayQuantity, "display-quantity", false, "Display quantity on hosted pages")

	// Other flags
	cmd.Flags().BoolVar(&allowAnyItemOnSubscriptions, "allow-any-item-on-subscriptions", false, "Allow any item as add-on")
	cmd.Flags().StringVar(&dunningCampaignId, "dunning-campaign-id", "", "Dunning campaign ID")

	return cmd
}

func newPlansUpdateCmd() *cobra.Command {
	var (
		// Core
		code        string
		name        string
		description string

		// Multi-currency (repeatable slices, positionally matched)
		currencies  []string
		unitAmounts []float64
		setupFees   []float64

		// Trial
		trialUnit                string
		trialLength              int
		trialRequiresBillingInfo bool

		// Billing
		autoRenew          bool
		totalBillingCycles int

		// Tax
		taxCode              string
		taxExempt            bool
		avalaraTransType     int
		avalaraServiceType   int
		vertexTransType      string
		harmonizedSystemCode string

		// Accounting
		accountingCode                  string
		revenueScheduleType             string
		liabilityGlAccountId            string
		revenueGlAccountId              string
		performanceObligationId         string
		setupFeeAccountingCode          string
		setupFeeRevenueScheduleType     string
		setupFeeLiabilityGlAccountId    string
		setupFeeRevenueGlAccountId      string
		setupFeePerformanceObligationId string

		// Hosted pages
		successUrl         string
		cancelUrl          string
		bypassConfirmation bool
		displayQuantity    bool

		// Other
		allowAnyItemOnSubscriptions bool
		dunningCampaignId           string
	)

	cmd := &cobra.Command{
		Use:   "update <plan_id>",
		Short: "Update a plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate currency/unit-amount pairing
			if cmd.Flags().Changed("currency") || cmd.Flags().Changed("unit-amount") {
				if len(currencies) != len(unitAmounts) {
					return fmt.Errorf("number of --currency values must match --unit-amount values")
				}
			}
			if cmd.Flags().Changed("setup-fee") {
				if len(setupFees) != len(currencies) {
					return fmt.Errorf("number of --setup-fee values must match --currency values")
				}
			}

			c, err := newPlanAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")
			body := &recurly.PlanUpdate{}

			// Core flags
			if cmd.Flags().Changed("code") {
				body.Code = recurly.String(code)
			}
			if cmd.Flags().Changed("name") {
				body.Name = recurly.String(name)
			}
			if cmd.Flags().Changed("description") {
				body.Description = recurly.String(description)
			}

			// Multi-currency
			if cmd.Flags().Changed("currency") {
				pricings := make([]recurly.PlanPricingCreate, len(currencies))
				for i, cur := range currencies {
					pricings[i] = recurly.PlanPricingCreate{
						Currency:   recurly.String(cur),
						UnitAmount: float64Ptr(unitAmounts[i]),
					}
				}
				body.Currencies = &pricings
			}

			// Setup fees (separate top-level field)
			if cmd.Flags().Changed("setup-fee") {
				fees := make([]recurly.PlanSetupPricingCreate, len(currencies))
				for i, cur := range currencies {
					fees[i] = recurly.PlanSetupPricingCreate{
						Currency:   recurly.String(cur),
						UnitAmount: float64Ptr(setupFees[i]),
					}
				}
				body.SetupFees = &fees
			}

			// Trial flags
			if cmd.Flags().Changed("trial-unit") {
				body.TrialUnit = recurly.String(trialUnit)
			}
			if cmd.Flags().Changed("trial-length") {
				body.TrialLength = recurly.Int(trialLength)
			}
			if cmd.Flags().Changed("trial-requires-billing-info") {
				body.TrialRequiresBillingInfo = recurly.Bool(trialRequiresBillingInfo)
			}

			// Billing flags
			if cmd.Flags().Changed("auto-renew") {
				body.AutoRenew = recurly.Bool(autoRenew)
			}
			if cmd.Flags().Changed("total-billing-cycles") {
				body.TotalBillingCycles = recurly.Int(totalBillingCycles)
			}

			// Tax flags
			if cmd.Flags().Changed("tax-code") {
				body.TaxCode = recurly.String(taxCode)
			}
			if cmd.Flags().Changed("tax-exempt") {
				body.TaxExempt = recurly.Bool(taxExempt)
			}
			if cmd.Flags().Changed("avalara-transaction-type") {
				body.AvalaraTransactionType = recurly.Int(avalaraTransType)
			}
			if cmd.Flags().Changed("avalara-service-type") {
				body.AvalaraServiceType = recurly.Int(avalaraServiceType)
			}
			if cmd.Flags().Changed("vertex-transaction-type") {
				body.VertexTransactionType = recurly.String(vertexTransType)
			}
			if cmd.Flags().Changed("harmonized-system-code") {
				body.HarmonizedSystemCode = recurly.String(harmonizedSystemCode)
			}

			// Accounting flags
			if cmd.Flags().Changed("accounting-code") {
				body.AccountingCode = recurly.String(accountingCode)
			}
			if cmd.Flags().Changed("revenue-schedule-type") {
				body.RevenueScheduleType = recurly.String(revenueScheduleType)
			}
			if cmd.Flags().Changed("liability-gl-account-id") {
				body.LiabilityGlAccountId = recurly.String(liabilityGlAccountId)
			}
			if cmd.Flags().Changed("revenue-gl-account-id") {
				body.RevenueGlAccountId = recurly.String(revenueGlAccountId)
			}
			if cmd.Flags().Changed("performance-obligation-id") {
				body.PerformanceObligationId = recurly.String(performanceObligationId)
			}
			if cmd.Flags().Changed("setup-fee-accounting-code") {
				body.SetupFeeAccountingCode = recurly.String(setupFeeAccountingCode)
			}
			if cmd.Flags().Changed("setup-fee-revenue-schedule-type") {
				body.SetupFeeRevenueScheduleType = recurly.String(setupFeeRevenueScheduleType)
			}
			if cmd.Flags().Changed("setup-fee-liability-gl-account-id") {
				body.SetupFeeLiabilityGlAccountId = recurly.String(setupFeeLiabilityGlAccountId)
			}
			if cmd.Flags().Changed("setup-fee-revenue-gl-account-id") {
				body.SetupFeeRevenueGlAccountId = recurly.String(setupFeeRevenueGlAccountId)
			}
			if cmd.Flags().Changed("setup-fee-performance-obligation-id") {
				body.SetupFeePerformanceObligationId = recurly.String(setupFeePerformanceObligationId)
			}

			// Hosted pages flags
			hasHostedPages := cmd.Flags().Changed("success-url") || cmd.Flags().Changed("cancel-url") ||
				cmd.Flags().Changed("bypass-confirmation") || cmd.Flags().Changed("display-quantity")
			if hasHostedPages {
				hp := &recurly.PlanHostedPagesCreate{}
				if cmd.Flags().Changed("success-url") {
					hp.SuccessUrl = recurly.String(successUrl)
				}
				if cmd.Flags().Changed("cancel-url") {
					hp.CancelUrl = recurly.String(cancelUrl)
				}
				if cmd.Flags().Changed("bypass-confirmation") {
					hp.BypassConfirmation = recurly.Bool(bypassConfirmation)
				}
				if cmd.Flags().Changed("display-quantity") {
					hp.DisplayQuantity = recurly.Bool(displayQuantity)
				}
				body.HostedPages = hp
			}

			// Other flags
			if cmd.Flags().Changed("allow-any-item-on-subscriptions") {
				body.AllowAnyItemOnSubscriptions = recurly.Bool(allowAnyItemOnSubscriptions)
			}
			if cmd.Flags().Changed("dunning-campaign-id") {
				body.DunningCampaignId = recurly.String(dunningCampaignId)
			}

			plan, err := c.UpdatePlan(args[0], body)
			if err != nil {
				return err
			}

			columns := planDetailColumns()

			formatted, err := output.FormatOne(format, columns, plan)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	// Core flags (no --interval-unit, --interval-length, --pricing-model — immutable after creation)
	cmd.Flags().StringVar(&code, "code", "", "Plan code")
	cmd.Flags().StringVar(&name, "name", "", "Plan name")
	cmd.Flags().StringVar(&description, "description", "", "Plan description")

	// Multi-currency flags
	cmd.Flags().StringSliceVar(&currencies, "currency", nil, "Currency code (repeatable, positionally matched with --unit-amount)")
	cmd.Flags().Float64SliceVar(&unitAmounts, "unit-amount", nil, "Unit amount (repeatable, positionally matched with --currency)")
	cmd.Flags().Float64SliceVar(&setupFees, "setup-fee", nil, "Setup fee (repeatable, positionally matched with --currency)")

	// Trial flags
	cmd.Flags().StringVar(&trialUnit, "trial-unit", "", "Trial period unit (day, week, month)")
	cmd.Flags().IntVar(&trialLength, "trial-length", 0, "Trial period length")
	cmd.Flags().BoolVar(&trialRequiresBillingInfo, "trial-requires-billing-info", false, "Require billing info for trial")

	// Billing flags
	cmd.Flags().BoolVar(&autoRenew, "auto-renew", false, "Auto-renew subscriptions")
	cmd.Flags().IntVar(&totalBillingCycles, "total-billing-cycles", 0, "Total billing cycles before auto-termination")

	// Tax flags
	cmd.Flags().StringVar(&taxCode, "tax-code", "", "Tax code")
	cmd.Flags().BoolVar(&taxExempt, "tax-exempt", false, "Tax exempt status")
	cmd.Flags().IntVar(&avalaraTransType, "avalara-transaction-type", 0, "Avalara transaction type")
	cmd.Flags().IntVar(&avalaraServiceType, "avalara-service-type", 0, "Avalara service type")
	cmd.Flags().StringVar(&vertexTransType, "vertex-transaction-type", "", "Vertex transaction type (sale, rental, lease)")
	cmd.Flags().StringVar(&harmonizedSystemCode, "harmonized-system-code", "", "Harmonized System (HS) code")

	// Accounting flags
	cmd.Flags().StringVar(&accountingCode, "accounting-code", "", "Accounting code")
	cmd.Flags().StringVar(&revenueScheduleType, "revenue-schedule-type", "", "Revenue schedule type")
	cmd.Flags().StringVar(&liabilityGlAccountId, "liability-gl-account-id", "", "Liability GL account ID")
	cmd.Flags().StringVar(&revenueGlAccountId, "revenue-gl-account-id", "", "Revenue GL account ID")
	cmd.Flags().StringVar(&performanceObligationId, "performance-obligation-id", "", "Performance obligation ID")
	cmd.Flags().StringVar(&setupFeeAccountingCode, "setup-fee-accounting-code", "", "Setup fee accounting code")
	cmd.Flags().StringVar(&setupFeeRevenueScheduleType, "setup-fee-revenue-schedule-type", "", "Setup fee revenue schedule type")
	cmd.Flags().StringVar(&setupFeeLiabilityGlAccountId, "setup-fee-liability-gl-account-id", "", "Setup fee liability GL account ID")
	cmd.Flags().StringVar(&setupFeeRevenueGlAccountId, "setup-fee-revenue-gl-account-id", "", "Setup fee revenue GL account ID")
	cmd.Flags().StringVar(&setupFeePerformanceObligationId, "setup-fee-performance-obligation-id", "", "Setup fee performance obligation ID")

	// Hosted pages flags
	cmd.Flags().StringVar(&successUrl, "success-url", "", "Hosted page success redirect URL")
	cmd.Flags().StringVar(&cancelUrl, "cancel-url", "", "Hosted page cancel redirect URL")
	cmd.Flags().BoolVar(&bypassConfirmation, "bypass-confirmation", false, "Bypass hosted page confirmation")
	cmd.Flags().BoolVar(&displayQuantity, "display-quantity", false, "Display quantity on hosted pages")

	// Other flags
	cmd.Flags().BoolVar(&allowAnyItemOnSubscriptions, "allow-any-item-on-subscriptions", false, "Allow any item as add-on")
	cmd.Flags().StringVar(&dunningCampaignId, "dunning-campaign-id", "", "Dunning campaign ID")

	return cmd
}

func newPlansGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <plan_id>",
		Short: "Get plan details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newPlanAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			plan, err := c.GetPlan(args[0])
			if err != nil {
				return err
			}

			columns := planDetailColumns()

			formatted, err := output.FormatOne(format, columns, plan)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newPlansListCmd() *cobra.Command {
	var (
		limit int
		all   bool
		order string
		sort  string
		state string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List plans",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newPlanAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			params := &recurly.ListPlansParams{}

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

			lister, err := c.ListPlans(params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.Plan](lister, limit, all)
			if err != nil {
				return err
			}

			columns := []output.Column{
				{Header: "Code", Extract: func(v any) string { return v.(recurly.Plan).Code }},
				{Header: "Name", Extract: func(v any) string { return v.(recurly.Plan).Name }},
				{Header: "State", Extract: func(v any) string { return v.(recurly.Plan).State }},
				{Header: "Interval", Extract: func(v any) string {
					p := v.(recurly.Plan)
					if p.IntervalUnit != "" {
						return fmt.Sprintf("%d %s", p.IntervalLength, p.IntervalUnit)
					}
					return ""
				}},
				{Header: "Price", Extract: func(v any) string {
					p := v.(recurly.Plan)
					if len(p.Currencies) > 0 {
						c := p.Currencies[0]
						return fmt.Sprintf("%.2f %s", c.UnitAmount, c.Currency)
					}
					return ""
				}},
				{Header: "Created At", Extract: func(v any) string {
					p := v.(recurly.Plan)
					if p.CreatedAt != nil {
						return p.CreatedAt.Format(time.RFC3339)
					}
					return ""
				}},
			}

			items := make([]any, len(result.Items))
			for i, p := range result.Items {
				items[i] = p
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
	cmd.Flags().StringVar(&state, "state", "", "Filter by state (active or inactive)")

	return cmd
}

func newPlansDeactivateCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "deactivate <plan_id>",
		Short: "Deactivate a plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, "Are you sure you want to deactivate this plan? [y/N] ")
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deactivation cancelled.")
					return err
				}
			}

			c, err := newPlanAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			plan, err := c.RemovePlan(planID)
			if err != nil {
				return err
			}

			columns := planDetailColumns()

			formatted, err := output.FormatOne(format, columns, plan)
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
