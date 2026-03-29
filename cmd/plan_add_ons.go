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

func newPlanAddOnsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-ons",
		Short: "Manage plan add-ons",
	}
	cmd.AddCommand(newPlanAddOnsListCmd())
	cmd.AddCommand(withWatch(newPlanAddOnsGetCmd()))
	cmd.AddCommand(withFromFile(newPlanAddOnsCreateCmd()))
	cmd.AddCommand(withFromFile(newPlanAddOnsUpdateCmd()))
	cmd.AddCommand(newPlanAddOnsDeleteCmd())
	return cmd
}

func newPlanAddOnsListCmd() *cobra.Command {
	var (
		limit int
		all   bool
		order string
		sort  string
		state string
	)

	cmd := &cobra.Command{
		Use:   "list <plan_id>",
		Short: "List add-ons for a plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newPlanAddOnAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			params := &recurly.ListPlanAddOnsParams{}

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

			lister, err := c.ListPlanAddOns(args[0], params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.AddOn](lister, limit, all)
			if err != nil {
				return err
			}

			columns := []output.Column{
				{Header: "ID", Extract: func(v any) string { return v.(recurly.AddOn).Id }},
				{Header: "Code", Extract: func(v any) string { return v.(recurly.AddOn).Code }},
				{Header: "Name", Extract: func(v any) string { return v.(recurly.AddOn).Name }},
				{Header: "State", Extract: func(v any) string { return v.(recurly.AddOn).State }},
				{Header: "Add-On Type", Extract: func(v any) string { return v.(recurly.AddOn).AddOnType }},
				{Header: "Created At", Extract: func(v any) string {
					a := v.(recurly.AddOn)
					if a.CreatedAt != nil {
						return a.CreatedAt.Format(time.RFC3339)
					}
					return ""
				}},
			}

			items := make([]any, len(result.Items))
			for i, a := range result.Items {
				items[i] = a
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

func planAddOnDetailColumns() []output.Column {
	return []output.Column{
		{Header: "ID", Extract: func(v any) string { return v.(*recurly.AddOn).Id }},
		{Header: "Code", Extract: func(v any) string { return v.(*recurly.AddOn).Code }},
		{Header: "Name", Extract: func(v any) string { return v.(*recurly.AddOn).Name }},
		{Header: "State", Extract: func(v any) string { return v.(*recurly.AddOn).State }},
		{Header: "Add-On Type", Extract: func(v any) string { return v.(*recurly.AddOn).AddOnType }},
		{Header: "Default Quantity", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(*recurly.AddOn).DefaultQuantity)
		}},
		{Header: "Optional", Extract: func(v any) string {
			return fmt.Sprintf("%t", v.(*recurly.AddOn).Optional)
		}},
		{Header: "Accounting Code", Extract: func(v any) string { return v.(*recurly.AddOn).AccountingCode }},
		{Header: "Tax Code", Extract: func(v any) string { return v.(*recurly.AddOn).TaxCode }},
		{Header: "Currencies", Extract: func(v any) string {
			currencies := v.(*recurly.AddOn).Currencies
			if len(currencies) == 0 {
				return ""
			}
			parts := make([]string, len(currencies))
			for i, c := range currencies {
				parts[i] = fmt.Sprintf("%s: %.2f", c.Currency, c.UnitAmount)
			}
			return strings.Join(parts, ", ")
		}},
		{Header: "Created At", Extract: func(v any) string {
			a := v.(*recurly.AddOn)
			if a.CreatedAt != nil {
				return a.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Updated At", Extract: func(v any) string {
			a := v.(*recurly.AddOn)
			if a.UpdatedAt != nil {
				return a.UpdatedAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}

func newPlanAddOnsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <plan_id> <add_on_id>",
		Short: "Get details of a plan add-on",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newPlanAddOnAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			addOn, err := c.GetPlanAddOn(args[0], args[1])
			if err != nil {
				return err
			}

			columns := planAddOnDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, addOn)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newPlanAddOnsCreateCmd() *cobra.Command {
	var (
		code                 string
		name                 string
		addOnType            string
		defaultQuantity      int
		optional             bool
		displayQuantity      bool
		accountingCode       string
		taxCode              string
		revenueScheduleType  string
		usageType            string
		usageCalculationType string
		measuredUnitId       string

		// Multi-currency (repeatable slices, positionally matched)
		currencies  []string
		unitAmounts []float64
	)

	cmd := &cobra.Command{
		Use:   "create <plan_id>",
		Short: "Create a plan add-on",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate currency/unit-amount pairing
			if cmd.Flags().Changed("currency") || cmd.Flags().Changed("unit-amount") {
				if len(currencies) != len(unitAmounts) {
					return fmt.Errorf("number of --currency values must match --unit-amount values")
				}
			}

			c, err := newPlanAddOnAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
			body := &recurly.AddOnCreate{}

			if cmd.Flags().Changed("code") {
				body.Code = recurly.String(code)
			}
			if cmd.Flags().Changed("name") {
				body.Name = recurly.String(name)
			}
			if cmd.Flags().Changed("add-on-type") {
				body.AddOnType = recurly.String(addOnType)
			}
			if cmd.Flags().Changed("default-quantity") {
				body.DefaultQuantity = recurly.Int(defaultQuantity)
			}
			if cmd.Flags().Changed("optional") {
				body.Optional = recurly.Bool(optional)
			}
			if cmd.Flags().Changed("display-quantity") {
				body.DisplayQuantity = recurly.Bool(displayQuantity)
			}
			if cmd.Flags().Changed("accounting-code") {
				body.AccountingCode = recurly.String(accountingCode)
			}
			if cmd.Flags().Changed("tax-code") {
				body.TaxCode = recurly.String(taxCode)
			}
			if cmd.Flags().Changed("revenue-schedule-type") {
				body.RevenueScheduleType = recurly.String(revenueScheduleType)
			}
			if cmd.Flags().Changed("usage-type") {
				body.UsageType = recurly.String(usageType)
			}
			if cmd.Flags().Changed("usage-calculation-type") {
				body.UsageCalculationType = recurly.String(usageCalculationType)
			}
			if cmd.Flags().Changed("measured-unit-id") {
				body.MeasuredUnitId = recurly.String(measuredUnitId)
			}

			// Multi-currency
			if cmd.Flags().Changed("currency") {
				pricings := make([]recurly.AddOnPricingCreate, len(currencies))
				for i, cur := range currencies {
					pricings[i] = recurly.AddOnPricingCreate{
						Currency:   recurly.String(cur),
						UnitAmount: float64Ptr(unitAmounts[i]),
					}
				}
				body.Currencies = &pricings
			}

			addOn, err := c.CreatePlanAddOn(args[0], body)
			if err != nil {
				return err
			}

			columns := planAddOnDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, addOn)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&code, "code", "", "Add-on code (required)")
	cmd.Flags().StringVar(&name, "name", "", "Add-on name (required)")
	cmd.Flags().StringVar(&addOnType, "add-on-type", "fixed", "Add-on type (fixed or usage)")
	cmd.Flags().IntVar(&defaultQuantity, "default-quantity", 0, "Default quantity")
	cmd.Flags().BoolVar(&optional, "optional", false, "Whether the add-on is optional")
	cmd.Flags().BoolVar(&displayQuantity, "display-quantity", false, "Display quantity on hosted pages")
	cmd.Flags().StringVar(&accountingCode, "accounting-code", "", "Accounting code")
	cmd.Flags().StringVar(&taxCode, "tax-code", "", "Tax code")
	cmd.Flags().StringVar(&revenueScheduleType, "revenue-schedule-type", "", "Revenue schedule type")
	cmd.Flags().StringVar(&usageType, "usage-type", "", "Usage type (price or percentage)")
	cmd.Flags().StringVar(&usageCalculationType, "usage-calculation-type", "", "Usage calculation type (cumulative or last_in_period)")
	cmd.Flags().StringVar(&measuredUnitId, "measured-unit-id", "", "Measured unit ID")

	// Multi-currency flags
	cmd.Flags().StringSliceVar(&currencies, "currency", nil, "Currency code (repeatable, positionally matched with --unit-amount)")
	cmd.Flags().Float64SliceVar(&unitAmounts, "unit-amount", nil, "Unit amount (repeatable, positionally matched with --currency)")

	return cmd
}

func newPlanAddOnsUpdateCmd() *cobra.Command {
	var (
		code                 string
		name                 string
		defaultQuantity      int
		optional             bool
		displayQuantity      bool
		accountingCode       string
		taxCode              string
		revenueScheduleType  string
		usageCalculationType string
		measuredUnitId       string
		measuredUnitName     string

		// Multi-currency (repeatable slices, positionally matched)
		currencies  []string
		unitAmounts []float64
	)

	cmd := &cobra.Command{
		Use:   "update <plan_id> <add_on_id>",
		Short: "Update a plan add-on",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate currency/unit-amount pairing
			if cmd.Flags().Changed("currency") || cmd.Flags().Changed("unit-amount") {
				if len(currencies) != len(unitAmounts) {
					return fmt.Errorf("number of --currency values must match --unit-amount values")
				}
			}

			c, err := newPlanAddOnAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
			body := &recurly.AddOnUpdate{}

			if cmd.Flags().Changed("code") {
				body.Code = recurly.String(code)
			}
			if cmd.Flags().Changed("name") {
				body.Name = recurly.String(name)
			}
			if cmd.Flags().Changed("default-quantity") {
				body.DefaultQuantity = recurly.Int(defaultQuantity)
			}
			if cmd.Flags().Changed("optional") {
				body.Optional = recurly.Bool(optional)
			}
			if cmd.Flags().Changed("display-quantity") {
				body.DisplayQuantity = recurly.Bool(displayQuantity)
			}
			if cmd.Flags().Changed("accounting-code") {
				body.AccountingCode = recurly.String(accountingCode)
			}
			if cmd.Flags().Changed("tax-code") {
				body.TaxCode = recurly.String(taxCode)
			}
			if cmd.Flags().Changed("revenue-schedule-type") {
				body.RevenueScheduleType = recurly.String(revenueScheduleType)
			}
			if cmd.Flags().Changed("usage-calculation-type") {
				body.UsageCalculationType = recurly.String(usageCalculationType)
			}
			if cmd.Flags().Changed("measured-unit-id") {
				body.MeasuredUnitId = recurly.String(measuredUnitId)
			}
			if cmd.Flags().Changed("measured-unit-name") {
				body.MeasuredUnitName = recurly.String(measuredUnitName)
			}

			// Multi-currency
			if cmd.Flags().Changed("currency") {
				pricings := make([]recurly.AddOnPricingCreate, len(currencies))
				for i, cur := range currencies {
					pricings[i] = recurly.AddOnPricingCreate{
						Currency:   recurly.String(cur),
						UnitAmount: float64Ptr(unitAmounts[i]),
					}
				}
				body.Currencies = &pricings
			}

			addOn, err := c.UpdatePlanAddOn(args[0], args[1], body)
			if err != nil {
				return err
			}

			columns := planAddOnDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, addOn)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&code, "code", "", "Add-on code")
	cmd.Flags().StringVar(&name, "name", "", "Add-on name")
	cmd.Flags().IntVar(&defaultQuantity, "default-quantity", 0, "Default quantity")
	cmd.Flags().BoolVar(&optional, "optional", false, "Whether the add-on is optional")
	cmd.Flags().BoolVar(&displayQuantity, "display-quantity", false, "Display quantity on hosted pages")
	cmd.Flags().StringVar(&accountingCode, "accounting-code", "", "Accounting code")
	cmd.Flags().StringVar(&taxCode, "tax-code", "", "Tax code")
	cmd.Flags().StringVar(&revenueScheduleType, "revenue-schedule-type", "", "Revenue schedule type")
	cmd.Flags().StringVar(&usageCalculationType, "usage-calculation-type", "", "Usage calculation type (cumulative or last_in_period)")
	cmd.Flags().StringVar(&measuredUnitId, "measured-unit-id", "", "Measured unit ID")
	cmd.Flags().StringVar(&measuredUnitName, "measured-unit-name", "", "Measured unit name")

	// Multi-currency flags
	cmd.Flags().StringSliceVar(&currencies, "currency", nil, "Currency code (repeatable, positionally matched with --unit-amount)")
	cmd.Flags().Float64SliceVar(&unitAmounts, "unit-amount", nil, "Unit amount (repeatable, positionally matched with --currency)")

	return cmd
}

func newPlanAddOnsDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <plan_id> <add_on_id>",
		Short: "Delete a plan add-on",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID := args[0]
			addOnID := args[1]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Delete add-on %s from plan %s? [y/N] ", addOnID, planID))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion cancelled.")
					return err
				}
			}

			c, err := newPlanAddOnAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			addOn, err := c.RemovePlanAddOn(planID, addOnID)
			if err != nil {
				return err
			}

			columns := planAddOnDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, addOn)
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
