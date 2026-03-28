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
	cmd.AddCommand(newPlansGetCmd())
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
