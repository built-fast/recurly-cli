package cmd

import (
	"fmt"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
)

func newTransactionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "Manage transactions",
	}
	cmd.AddCommand(newTransactionsListCmd())
	cmd.AddCommand(withWatch(newTransactionsGetCmd()))
	return cmd
}

func newTransactionsListCmd() *cobra.Command {
	var (
		limit           int
		all             bool
		order           string
		sort            string
		transactionType string
		success         string
		beginTime       string
		endTime         string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewTransactionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			params := &recurly.ListTransactionsParams{}

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

			lister, err := c.ListTransactions(params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.Transaction](lister, limit, all)
			if err != nil {
				return err
			}

			type T = recurly.Transaction
			columns := output.ToColumns([]output.TypedColumn[T]{
				output.StringColumn[T]("ID", func(t T) string { return t.Id }),
				output.StringColumn[T]("Type", func(t T) string { return t.Type }),
				output.StringColumn[T]("Account", func(t T) string { return t.Account.Code }),
				output.StringColumn[T]("Status", func(t T) string { return t.Status }),
				output.StringColumn[T]("Currency", func(t T) string { return t.Currency }),
				output.FloatColumn[T]("Amount", func(t T) float64 { return t.Amount }),
				output.BoolColumn[T]("Success", func(t T) bool { return t.Success }),
				output.StringColumn[T]("Origin", func(t T) string { return t.Origin }),
				output.TimeColumn[T]("Created At", func(t T) *time.Time { return t.CreatedAt }),
			})

			items := make([]any, len(result.Items))
			for i, txn := range result.Items {
				items[i] = txn
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
	cmd.Flags().StringVar(&transactionType, "type", "", "Filter by type (purchase, capture, refund, verify, payment)")
	cmd.Flags().StringVar(&success, "success", "", "Filter by success (true or false)")
	cmd.Flags().StringVar(&beginTime, "begin-time", "", "Filter by begin time (ISO8601 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter by end time (ISO8601 format)")

	return cmd
}

func newTransactionsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <transaction_id>",
		Short: "Get transaction details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewTransactionAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			txn, err := c.GetTransaction(args[0])
			if err != nil {
				return err
			}

			// For JSON/jq output, format the whole transaction object
			if cfg.Format == "json" || cfg.Format == "json-pretty" || cfg.HasJQ() {
				formatted, err := output.FormatOne(cfg, nil, txn)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
				return err
			}

			// Table output: detail view
			columns := transactionDetailColumns()
			formatted, err := output.FormatOne(cfg, columns, txn)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func transactionDetailColumns() []output.Column {
	type T = *recurly.Transaction
	return output.ToColumns([]output.TypedColumn[T]{
		output.StringColumn[T]("ID", func(t T) string { return t.Id }),
		output.StringColumn[T]("UUID", func(t T) string { return t.Uuid }),
		output.StringColumn[T]("Type", func(t T) string { return t.Type }),
		output.StringColumn[T]("Origin", func(t T) string { return t.Origin }),
		output.StringColumn[T]("Status", func(t T) string { return t.Status }),
		output.BoolColumn[T]("Success", func(t T) bool { return t.Success }),
		output.FloatColumn[T]("Amount", func(t T) float64 { return t.Amount }),
		output.StringColumn[T]("Currency", func(t T) string { return t.Currency }),
		output.StringColumn[T]("Account ID", func(t T) string { return t.Account.Id }),
		output.StringColumn[T]("Account Code", func(t T) string { return t.Account.Code }),
		output.StringColumn[T]("Invoice ID", func(t T) string { return t.Invoice.Id }),
		output.StringColumn[T]("Invoice Number", func(t T) string { return t.Invoice.Number }),
		output.StringColumn[T]("Collection Method", func(t T) string { return t.CollectionMethod }),
		output.StringColumn[T]("Payment Method Type", func(t T) string { return t.PaymentMethod.Object }),
		output.StringColumn[T]("Payment Method Card Type", func(t T) string { return t.PaymentMethod.CardType }),
		output.StringColumn[T]("Payment Method Last Four", func(t T) string { return t.PaymentMethod.LastFour }),
		output.StringColumn[T]("IP Address", func(t T) string { return t.IpAddressV4 }),
		output.StringColumn[T]("Status Code", func(t T) string { return t.StatusCode }),
		output.StringColumn[T]("Status Message", func(t T) string { return t.StatusMessage }),
		output.BoolColumn[T]("Refunded", func(t T) bool { return t.Refunded }),
		output.TimeColumn[T]("Created At", func(t T) *time.Time { return t.CreatedAt }),
		output.TimeColumn[T]("Updated At", func(t T) *time.Time { return t.UpdatedAt }),
		output.TimeColumn[T]("Voided At", func(t T) *time.Time { return t.VoidedAt }),
	})
}
