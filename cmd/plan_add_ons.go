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

func newPlanAddOnsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-ons",
		Short: "Manage plan add-ons",
	}
	cmd.AddCommand(newPlanAddOnsListCmd())
	cmd.AddCommand(newPlanAddOnsGetCmd())
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
			c, err := newPlanAddOnAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

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
			c, err := newPlanAddOnAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			addOn, err := c.GetPlanAddOn(args[0], args[1])
			if err != nil {
				return err
			}

			columns := planAddOnDetailColumns()

			formatted, err := output.FormatOne(format, columns, addOn)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}
