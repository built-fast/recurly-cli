package cmd

import (
	"fmt"
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
