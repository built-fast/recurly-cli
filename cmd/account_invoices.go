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

func newAccountInvoicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoices",
		Short: "Manage account invoices",
	}
	cmd.AddCommand(newAccountInvoicesListCmd())
	return cmd
}

func newAccountInvoicesListCmd() *cobra.Command {
	var (
		limit       int
		all         bool
		order       string
		sort        string
		state       string
		invoiceType string
	)

	cmd := &cobra.Command{
		Use:   "list <account_id>",
		Short: "List invoices for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAccountNestedAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			params := &recurly.ListAccountInvoicesParams{}

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
			if cmd.Flags().Changed("type") {
				params.Type = recurly.String(invoiceType)
			}

			lister, err := c.ListAccountInvoices(args[0], params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.Invoice](lister, limit, all)
			if err != nil {
				return err
			}

			columns := []output.Column{
				{Header: "ID (Number)", Extract: func(v any) string { return v.(recurly.Invoice).Number }},
				{Header: "State", Extract: func(v any) string { return v.(recurly.Invoice).State }},
				{Header: "Type", Extract: func(v any) string { return v.(recurly.Invoice).Type }},
				{Header: "Total", Extract: func(v any) string {
					return fmt.Sprintf("%.2f", v.(recurly.Invoice).Total)
				}},
				{Header: "Currency", Extract: func(v any) string { return v.(recurly.Invoice).Currency }},
				{Header: "Created At", Extract: func(v any) string {
					inv := v.(recurly.Invoice)
					if inv.CreatedAt != nil {
						return inv.CreatedAt.Format(time.RFC3339)
					}
					return ""
				}},
			}

			items := make([]any, len(result.Items))
			for i, inv := range result.Items {
				items[i] = inv
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
	cmd.Flags().StringVar(&sort, "sort", "", "Sort field (created_at or updated_at)")
	cmd.Flags().StringVar(&state, "state", "", "Filter by state")
	cmd.Flags().StringVar(&invoiceType, "type", "", "Filter by type (charge, credit, non-legacy, legacy)")

	return cmd
}
