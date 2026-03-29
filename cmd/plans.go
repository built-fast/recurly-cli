package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
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
	type P = *recurly.Plan
	return output.ToColumns([]output.TypedColumn[P]{
		output.StringColumn[P]("Id", func(p P) string { return p.Id }),
		output.StringColumn[P]("Code", func(p P) string { return p.Code }),
		output.StringColumn[P]("Name", func(p P) string { return p.Name }),
		output.StringColumn[P]("State", func(p P) string { return p.State }),
		output.StringColumn[P]("Pricing Model", func(p P) string { return p.PricingModel }),
		output.StringColumn[P]("Interval Unit", func(p P) string { return p.IntervalUnit }),
		output.IntColumn[P]("Interval Length", func(p P) int { return p.IntervalLength }),
		output.StringColumn[P]("Description", func(p P) string { return p.Description }),
		{Header: "Currencies", Extract: func(p P) string {
			if len(p.Currencies) == 0 {
				return ""
			}
			var parts []string
			for _, c := range p.Currencies {
				parts = append(parts, fmt.Sprintf("%s: %.2f (setup: %.2f)", c.Currency, c.UnitAmount, c.SetupFee))
			}
			return strings.Join(parts, ", ")
		}},
		output.StringColumn[P]("Trial Unit", func(p P) string { return p.TrialUnit }),
		output.IntColumn[P]("Trial Length", func(p P) int { return p.TrialLength }),
		output.BoolColumn[P]("Auto Renew", func(p P) bool { return p.AutoRenew }),
		output.IntColumn[P]("Total Billing Cycles", func(p P) int { return p.TotalBillingCycles }),
		output.StringColumn[P]("Tax Code", func(p P) string { return p.TaxCode }),
		output.BoolColumn[P]("Tax Exempt", func(p P) bool { return p.TaxExempt }),
		output.TimeColumn[P]("Created At", func(p P) *time.Time { return p.CreatedAt }),
		output.TimeColumn[P]("Updated At", func(p P) *time.Time { return p.UpdatedAt }),
	})
}

func float64Ptr(v float64) *float64 {
	return &v
}

// planFlags holds all flag variables shared between plan create and update commands.
type planFlags struct {
	code, name, description, pricingModel, intervalUnit string
	intervalLen                                         int
	currencies                                          []string
	unitAmounts, setupFees                              []float64
	trialUnit                                           string
	trialLength                                         int
	trialRequiresBillingInfo                             bool
	autoRenew                                           bool
	totalBillingCycles                                  int
	taxCode                                             string
	taxExempt                                           bool
	avalaraTransType, avalaraServiceType                int
	vertexTransType, harmonizedSystemCode               string
	accountingCode, revenueScheduleType                 string
	liabilityGlAccountId, revenueGlAccountId            string
	performanceObligationId                             string
	setupFeeAccountingCode, setupFeeRevenueScheduleType string
	setupFeeLiabilityGlAccountId                        string
	setupFeeRevenueGlAccountId                          string
	setupFeePerformanceObligationId                     string
	successUrl, cancelUrl                               string
	bypassConfirmation, displayQuantity                 bool
	allowAnyItemOnSubscriptions                         bool
	dunningCampaignId                                   string
}

func (f *planFlags) validateCurrencies(cmd *cobra.Command) error {
	if cmd.Flags().Changed("currency") || cmd.Flags().Changed("unit-amount") {
		if len(f.currencies) != len(f.unitAmounts) {
			return fmt.Errorf("number of --currency values must match --unit-amount values")
		}
	}
	if cmd.Flags().Changed("setup-fee") {
		if len(f.setupFees) != len(f.currencies) {
			return fmt.Errorf("number of --setup-fee values must match --currency values")
		}
	}
	return nil
}

func (f *planFlags) buildCurrencies(cmd *cobra.Command) *[]recurly.PlanPricingCreate {
	if !cmd.Flags().Changed("currency") {
		return nil
	}
	pricings := make([]recurly.PlanPricingCreate, len(f.currencies))
	for i, cur := range f.currencies {
		pricings[i] = recurly.PlanPricingCreate{
			Currency:   recurly.String(cur),
			UnitAmount: float64Ptr(f.unitAmounts[i]),
		}
	}
	return &pricings
}

func (f *planFlags) buildSetupFees(cmd *cobra.Command) *[]recurly.PlanSetupPricingCreate {
	if !cmd.Flags().Changed("setup-fee") {
		return nil
	}
	fees := make([]recurly.PlanSetupPricingCreate, len(f.currencies))
	for i, cur := range f.currencies {
		fees[i] = recurly.PlanSetupPricingCreate{
			Currency:   recurly.String(cur),
			UnitAmount: float64Ptr(f.setupFees[i]),
		}
	}
	return &fees
}

func (f *planFlags) buildHostedPages(cmd *cobra.Command) *recurly.PlanHostedPagesCreate {
	hasHostedPages := cmd.Flags().Changed("success-url") || cmd.Flags().Changed("cancel-url") ||
		cmd.Flags().Changed("bypass-confirmation") || cmd.Flags().Changed("display-quantity")
	if !hasHostedPages {
		return nil
	}
	hp := &recurly.PlanHostedPagesCreate{}
	if cmd.Flags().Changed("success-url") {
		hp.SuccessUrl = recurly.String(f.successUrl)
	}
	if cmd.Flags().Changed("cancel-url") {
		hp.CancelUrl = recurly.String(f.cancelUrl)
	}
	if cmd.Flags().Changed("bypass-confirmation") {
		hp.BypassConfirmation = recurly.Bool(f.bypassConfirmation)
	}
	if cmd.Flags().Changed("display-quantity") {
		hp.DisplayQuantity = recurly.Bool(f.displayQuantity)
	}
	return hp
}

func (f *planFlags) buildCreateBody(cmd *cobra.Command) *recurly.PlanCreate {
	body := &recurly.PlanCreate{}

	if cmd.Flags().Changed("code") {
		body.Code = recurly.String(f.code)
	}
	if cmd.Flags().Changed("name") {
		body.Name = recurly.String(f.name)
	}
	if cmd.Flags().Changed("interval-unit") {
		body.IntervalUnit = recurly.String(f.intervalUnit)
	}
	if cmd.Flags().Changed("interval-length") {
		body.IntervalLength = recurly.Int(f.intervalLen)
	}
	if cmd.Flags().Changed("description") {
		body.Description = recurly.String(f.description)
	}
	if cmd.Flags().Changed("pricing-model") {
		body.PricingModel = recurly.String(f.pricingModel)
	}

	body.Currencies = f.buildCurrencies(cmd)
	body.SetupFees = f.buildSetupFees(cmd)

	if cmd.Flags().Changed("trial-unit") {
		body.TrialUnit = recurly.String(f.trialUnit)
	}
	if cmd.Flags().Changed("trial-length") {
		body.TrialLength = recurly.Int(f.trialLength)
	}
	if cmd.Flags().Changed("trial-requires-billing-info") {
		body.TrialRequiresBillingInfo = recurly.Bool(f.trialRequiresBillingInfo)
	}
	if cmd.Flags().Changed("auto-renew") {
		body.AutoRenew = recurly.Bool(f.autoRenew)
	}
	if cmd.Flags().Changed("total-billing-cycles") {
		body.TotalBillingCycles = recurly.Int(f.totalBillingCycles)
	}
	if cmd.Flags().Changed("tax-code") {
		body.TaxCode = recurly.String(f.taxCode)
	}
	if cmd.Flags().Changed("tax-exempt") {
		body.TaxExempt = recurly.Bool(f.taxExempt)
	}
	if cmd.Flags().Changed("avalara-transaction-type") {
		body.AvalaraTransactionType = recurly.Int(f.avalaraTransType)
	}
	if cmd.Flags().Changed("avalara-service-type") {
		body.AvalaraServiceType = recurly.Int(f.avalaraServiceType)
	}
	if cmd.Flags().Changed("vertex-transaction-type") {
		body.VertexTransactionType = recurly.String(f.vertexTransType)
	}
	if cmd.Flags().Changed("harmonized-system-code") {
		body.HarmonizedSystemCode = recurly.String(f.harmonizedSystemCode)
	}

	f.setCreateAccountingFields(cmd, body)
	body.HostedPages = f.buildHostedPages(cmd)

	if cmd.Flags().Changed("allow-any-item-on-subscriptions") {
		body.AllowAnyItemOnSubscriptions = recurly.Bool(f.allowAnyItemOnSubscriptions)
	}
	if cmd.Flags().Changed("dunning-campaign-id") {
		body.DunningCampaignId = recurly.String(f.dunningCampaignId)
	}

	return body
}

func (f *planFlags) setCreateAccountingFields(cmd *cobra.Command, body *recurly.PlanCreate) {
	if cmd.Flags().Changed("accounting-code") {
		body.AccountingCode = recurly.String(f.accountingCode)
	}
	if cmd.Flags().Changed("revenue-schedule-type") {
		body.RevenueScheduleType = recurly.String(f.revenueScheduleType)
	}
	if cmd.Flags().Changed("liability-gl-account-id") {
		body.LiabilityGlAccountId = recurly.String(f.liabilityGlAccountId)
	}
	if cmd.Flags().Changed("revenue-gl-account-id") {
		body.RevenueGlAccountId = recurly.String(f.revenueGlAccountId)
	}
	if cmd.Flags().Changed("performance-obligation-id") {
		body.PerformanceObligationId = recurly.String(f.performanceObligationId)
	}
	if cmd.Flags().Changed("setup-fee-accounting-code") {
		body.SetupFeeAccountingCode = recurly.String(f.setupFeeAccountingCode)
	}
	if cmd.Flags().Changed("setup-fee-revenue-schedule-type") {
		body.SetupFeeRevenueScheduleType = recurly.String(f.setupFeeRevenueScheduleType)
	}
	if cmd.Flags().Changed("setup-fee-liability-gl-account-id") {
		body.SetupFeeLiabilityGlAccountId = recurly.String(f.setupFeeLiabilityGlAccountId)
	}
	if cmd.Flags().Changed("setup-fee-revenue-gl-account-id") {
		body.SetupFeeRevenueGlAccountId = recurly.String(f.setupFeeRevenueGlAccountId)
	}
	if cmd.Flags().Changed("setup-fee-performance-obligation-id") {
		body.SetupFeePerformanceObligationId = recurly.String(f.setupFeePerformanceObligationId)
	}
}

func (f *planFlags) buildUpdateBody(cmd *cobra.Command) *recurly.PlanUpdate {
	body := &recurly.PlanUpdate{}

	if cmd.Flags().Changed("code") {
		body.Code = recurly.String(f.code)
	}
	if cmd.Flags().Changed("name") {
		body.Name = recurly.String(f.name)
	}
	if cmd.Flags().Changed("description") {
		body.Description = recurly.String(f.description)
	}

	body.Currencies = f.buildCurrencies(cmd)
	body.SetupFees = f.buildSetupFees(cmd)

	if cmd.Flags().Changed("trial-unit") {
		body.TrialUnit = recurly.String(f.trialUnit)
	}
	if cmd.Flags().Changed("trial-length") {
		body.TrialLength = recurly.Int(f.trialLength)
	}
	if cmd.Flags().Changed("trial-requires-billing-info") {
		body.TrialRequiresBillingInfo = recurly.Bool(f.trialRequiresBillingInfo)
	}
	if cmd.Flags().Changed("auto-renew") {
		body.AutoRenew = recurly.Bool(f.autoRenew)
	}
	if cmd.Flags().Changed("total-billing-cycles") {
		body.TotalBillingCycles = recurly.Int(f.totalBillingCycles)
	}
	if cmd.Flags().Changed("tax-code") {
		body.TaxCode = recurly.String(f.taxCode)
	}
	if cmd.Flags().Changed("tax-exempt") {
		body.TaxExempt = recurly.Bool(f.taxExempt)
	}
	if cmd.Flags().Changed("avalara-transaction-type") {
		body.AvalaraTransactionType = recurly.Int(f.avalaraTransType)
	}
	if cmd.Flags().Changed("avalara-service-type") {
		body.AvalaraServiceType = recurly.Int(f.avalaraServiceType)
	}
	if cmd.Flags().Changed("vertex-transaction-type") {
		body.VertexTransactionType = recurly.String(f.vertexTransType)
	}
	if cmd.Flags().Changed("harmonized-system-code") {
		body.HarmonizedSystemCode = recurly.String(f.harmonizedSystemCode)
	}

	f.setUpdateAccountingFields(cmd, body)
	body.HostedPages = f.buildHostedPages(cmd)

	if cmd.Flags().Changed("allow-any-item-on-subscriptions") {
		body.AllowAnyItemOnSubscriptions = recurly.Bool(f.allowAnyItemOnSubscriptions)
	}
	if cmd.Flags().Changed("dunning-campaign-id") {
		body.DunningCampaignId = recurly.String(f.dunningCampaignId)
	}

	return body
}

func (f *planFlags) setUpdateAccountingFields(cmd *cobra.Command, body *recurly.PlanUpdate) {
	if cmd.Flags().Changed("accounting-code") {
		body.AccountingCode = recurly.String(f.accountingCode)
	}
	if cmd.Flags().Changed("revenue-schedule-type") {
		body.RevenueScheduleType = recurly.String(f.revenueScheduleType)
	}
	if cmd.Flags().Changed("liability-gl-account-id") {
		body.LiabilityGlAccountId = recurly.String(f.liabilityGlAccountId)
	}
	if cmd.Flags().Changed("revenue-gl-account-id") {
		body.RevenueGlAccountId = recurly.String(f.revenueGlAccountId)
	}
	if cmd.Flags().Changed("performance-obligation-id") {
		body.PerformanceObligationId = recurly.String(f.performanceObligationId)
	}
	if cmd.Flags().Changed("setup-fee-accounting-code") {
		body.SetupFeeAccountingCode = recurly.String(f.setupFeeAccountingCode)
	}
	if cmd.Flags().Changed("setup-fee-revenue-schedule-type") {
		body.SetupFeeRevenueScheduleType = recurly.String(f.setupFeeRevenueScheduleType)
	}
	if cmd.Flags().Changed("setup-fee-liability-gl-account-id") {
		body.SetupFeeLiabilityGlAccountId = recurly.String(f.setupFeeLiabilityGlAccountId)
	}
	if cmd.Flags().Changed("setup-fee-revenue-gl-account-id") {
		body.SetupFeeRevenueGlAccountId = recurly.String(f.setupFeeRevenueGlAccountId)
	}
	if cmd.Flags().Changed("setup-fee-performance-obligation-id") {
		body.SetupFeePerformanceObligationId = recurly.String(f.setupFeePerformanceObligationId)
	}
}

func (f *planFlags) registerSharedFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.code, "code", "", "Plan code")
	cmd.Flags().StringVar(&f.name, "name", "", "Plan name")
	cmd.Flags().StringVar(&f.description, "description", "", "Plan description")

	cmd.Flags().StringSliceVar(&f.currencies, "currency", nil, "Currency code (repeatable, positionally matched with --unit-amount)")
	cmd.Flags().Float64SliceVar(&f.unitAmounts, "unit-amount", nil, "Unit amount (repeatable, positionally matched with --currency)")
	cmd.Flags().Float64SliceVar(&f.setupFees, "setup-fee", nil, "Setup fee (repeatable, positionally matched with --currency)")

	cmd.Flags().StringVar(&f.trialUnit, "trial-unit", "", "Trial period unit (day, week, month)")
	cmd.Flags().IntVar(&f.trialLength, "trial-length", 0, "Trial period length")
	cmd.Flags().BoolVar(&f.trialRequiresBillingInfo, "trial-requires-billing-info", false, "Require billing info for trial")

	cmd.Flags().BoolVar(&f.autoRenew, "auto-renew", false, "Auto-renew subscriptions")
	cmd.Flags().IntVar(&f.totalBillingCycles, "total-billing-cycles", 0, "Total billing cycles before auto-termination")

	cmd.Flags().StringVar(&f.taxCode, "tax-code", "", "Tax code")
	cmd.Flags().BoolVar(&f.taxExempt, "tax-exempt", false, "Tax exempt status")
	cmd.Flags().IntVar(&f.avalaraTransType, "avalara-transaction-type", 0, "Avalara transaction type")
	cmd.Flags().IntVar(&f.avalaraServiceType, "avalara-service-type", 0, "Avalara service type")
	cmd.Flags().StringVar(&f.vertexTransType, "vertex-transaction-type", "", "Vertex transaction type (sale, rental, lease)")
	cmd.Flags().StringVar(&f.harmonizedSystemCode, "harmonized-system-code", "", "Harmonized System (HS) code")

	cmd.Flags().StringVar(&f.accountingCode, "accounting-code", "", "Accounting code")
	cmd.Flags().StringVar(&f.revenueScheduleType, "revenue-schedule-type", "", "Revenue schedule type")
	cmd.Flags().StringVar(&f.liabilityGlAccountId, "liability-gl-account-id", "", "Liability GL account ID")
	cmd.Flags().StringVar(&f.revenueGlAccountId, "revenue-gl-account-id", "", "Revenue GL account ID")
	cmd.Flags().StringVar(&f.performanceObligationId, "performance-obligation-id", "", "Performance obligation ID")
	cmd.Flags().StringVar(&f.setupFeeAccountingCode, "setup-fee-accounting-code", "", "Setup fee accounting code")
	cmd.Flags().StringVar(&f.setupFeeRevenueScheduleType, "setup-fee-revenue-schedule-type", "", "Setup fee revenue schedule type")
	cmd.Flags().StringVar(&f.setupFeeLiabilityGlAccountId, "setup-fee-liability-gl-account-id", "", "Setup fee liability GL account ID")
	cmd.Flags().StringVar(&f.setupFeeRevenueGlAccountId, "setup-fee-revenue-gl-account-id", "", "Setup fee revenue GL account ID")
	cmd.Flags().StringVar(&f.setupFeePerformanceObligationId, "setup-fee-performance-obligation-id", "", "Setup fee performance obligation ID")

	cmd.Flags().StringVar(&f.successUrl, "success-url", "", "Hosted page success redirect URL")
	cmd.Flags().StringVar(&f.cancelUrl, "cancel-url", "", "Hosted page cancel redirect URL")
	cmd.Flags().BoolVar(&f.bypassConfirmation, "bypass-confirmation", false, "Bypass hosted page confirmation")
	cmd.Flags().BoolVar(&f.displayQuantity, "display-quantity", false, "Display quantity on hosted pages")

	cmd.Flags().BoolVar(&f.allowAnyItemOnSubscriptions, "allow-any-item-on-subscriptions", false, "Allow any item as add-on")
	cmd.Flags().StringVar(&f.dunningCampaignId, "dunning-campaign-id", "", "Dunning campaign ID")
}

func newPlansCreateCmd() *cobra.Command {
	f := &planFlags{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.validateCurrencies(cmd); err != nil {
				return err
			}

			c, err := AppFromContext(cmd.Context()).NewPlanAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
			body := f.buildCreateBody(cmd)

			plan, err := c.CreatePlan(body)
			if err != nil {
				return err
			}

			formatted, err := output.FormatOne(cfg, planDetailColumns(), plan)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	f.registerSharedFlags(cmd)
	cmd.Flags().StringVar(&f.intervalUnit, "interval-unit", "", "Billing interval unit (day, week, month)")
	cmd.Flags().IntVar(&f.intervalLen, "interval-length", 0, "Billing interval length")
	cmd.Flags().StringVar(&f.pricingModel, "pricing-model", "", "Pricing model (fixed or ramp)")

	return cmd
}

func newPlansUpdateCmd() *cobra.Command {
	f := &planFlags{}

	cmd := &cobra.Command{
		Use:   "update <plan_id>",
		Short: "Update a plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.validateCurrencies(cmd); err != nil {
				return err
			}

			c, err := AppFromContext(cmd.Context()).NewPlanAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
			body := f.buildUpdateBody(cmd)

			plan, err := c.UpdatePlan(args[0], body)
			if err != nil {
				return err
			}

			formatted, err := output.FormatOne(cfg, planDetailColumns(), plan)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	// No --interval-unit, --interval-length, --pricing-model — immutable after creation
	f.registerSharedFlags(cmd)

	return cmd
}

func newPlansGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <plan_id>",
		Short: "Get plan details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewPlanAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			plan, err := c.GetPlan(args[0])
			if err != nil {
				return err
			}

			columns := planDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, plan)
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
			c, err := AppFromContext(cmd.Context()).NewPlanAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			type P = recurly.Plan
			columns := output.ToColumns([]output.TypedColumn[P]{
				output.StringColumn[P]("Code", func(p P) string { return p.Code }),
				output.StringColumn[P]("Name", func(p P) string { return p.Name }),
				output.StringColumn[P]("State", func(p P) string { return p.State }),
				{Header: "Interval", Extract: func(p P) string {
					if p.IntervalUnit != "" {
						return fmt.Sprintf("%d %s", p.IntervalLength, p.IntervalUnit)
					}
					return ""
				}},
				{Header: "Price", Extract: func(p P) string {
					if len(p.Currencies) > 0 {
						c := p.Currencies[0]
						return fmt.Sprintf("%.2f %s", c.UnitAmount, c.Currency)
					}
					return ""
				}},
				output.TimeColumn[P]("Created At", func(p P) *time.Time { return p.CreatedAt }),
			})

			items := make([]any, len(result.Items))
			for i, p := range result.Items {
				items[i] = p
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

			c, err := AppFromContext(cmd.Context()).NewPlanAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			plan, err := c.RemovePlan(planID)
			if err != nil {
				return err
			}

			columns := planDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, plan)
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
