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

func newAccountTransactionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "Manage account transactions",
	}
	cmd.AddCommand(newAccountTransactionsListCmd())
	return cmd
}

func newAccountTransactionsListCmd() *cobra.Command {
	var (
		limit           int
		all             bool
		order           string
		sort            string
		transactionType string
		success         string
	)

	cmd := &cobra.Command{
		Use:   "list <account_id>",
		Short: "List transactions for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAccountNestedAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			params := &recurly.ListAccountTransactionsParams{}

			if limit > 0 {
				params.Limit = recurly.Int(limit)
			}
			if cmd.Flags().Changed("order") {
				params.Order = recurly.String(order)
			}
			if cmd.Flags().Changed("sort") {
				params.Sort = recurly.String(sort)
			}
			if cmd.Flags().Changed("type") {
				params.Type = recurly.String(transactionType)
			}
			if cmd.Flags().Changed("success") {
				params.Success = recurly.String(success)
			}

			lister, err := c.ListAccountTransactions(args[0], params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.Transaction](lister, limit, all)
			if err != nil {
				return err
			}

			columns := []output.Column{
				{Header: "ID", Extract: func(v any) string { return v.(recurly.Transaction).Id }},
				{Header: "Type", Extract: func(v any) string { return v.(recurly.Transaction).Type }},
				{Header: "Amount", Extract: func(v any) string {
					return fmt.Sprintf("%.2f", v.(recurly.Transaction).Amount)
				}},
				{Header: "Currency", Extract: func(v any) string { return v.(recurly.Transaction).Currency }},
				{Header: "Status", Extract: func(v any) string { return v.(recurly.Transaction).Status }},
				{Header: "Success", Extract: func(v any) string {
					return fmt.Sprintf("%t", v.(recurly.Transaction).Success)
				}},
				{Header: "Created At", Extract: func(v any) string {
					txn := v.(recurly.Transaction)
					if txn.CreatedAt != nil {
						return txn.CreatedAt.Format(time.RFC3339)
					}
					return ""
				}},
			}

			items := make([]any, len(result.Items))
			for i, txn := range result.Items {
				items[i] = txn
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
	cmd.Flags().StringVar(&transactionType, "type", "", "Filter by type (e.g. payment, refund)")
	cmd.Flags().StringVar(&success, "success", "", "Filter by success (true or false)")

	return cmd
}
