package cmd

import (
	"fmt"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
)

func newAccountSubscriptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Manage account subscriptions",
	}
	cmd.AddCommand(newAccountSubscriptionsListCmd())
	return cmd
}

func newAccountSubscriptionsListCmd() *cobra.Command {
	var (
		limit int
		all   bool
		order string
		sort  string
		state string
	)

	cmd := &cobra.Command{
		Use:   "list <account_id>",
		Short: "List subscriptions for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewAccountNestedAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			params := &recurly.ListAccountSubscriptionsParams{}

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

			lister, err := c.ListAccountSubscriptions(args[0], params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.Subscription](lister, limit, all)
			if err != nil {
				return err
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
	cmd.Flags().StringVar(&sort, "sort", "", "Sort field (created_at or updated_at)")
	cmd.Flags().StringVar(&state, "state", "", "Filter by state (active, canceled, expired, future, in_trial, live)")

	return cmd
}

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
			c, err := AppFromContext(cmd.Context()).NewAccountNestedAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			type I = recurly.Invoice
			columns := output.ToColumns([]output.TypedColumn[I]{
				output.StringColumn[I]("ID (Number)", func(i I) string { return i.Number }),
				output.StringColumn[I]("State", func(i I) string { return i.State }),
				output.StringColumn[I]("Type", func(i I) string { return i.Type }),
				output.FloatColumn[I]("Total", func(i I) float64 { return i.Total }),
				output.StringColumn[I]("Currency", func(i I) string { return i.Currency }),
				output.TimeColumn[I]("Created At", func(i I) *time.Time { return i.CreatedAt }),
			})

			items := make([]any, len(result.Items))
			for i, inv := range result.Items {
				items[i] = inv
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
	cmd.Flags().StringVar(&sort, "sort", "", "Sort field (created_at or updated_at)")
	cmd.Flags().StringVar(&state, "state", "", "Filter by state")
	cmd.Flags().StringVar(&invoiceType, "type", "", "Filter by type (charge, credit, non-legacy, legacy)")

	return cmd
}

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
			c, err := AppFromContext(cmd.Context()).NewAccountNestedAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			type T = recurly.Transaction
			columns := output.ToColumns([]output.TypedColumn[T]{
				output.StringColumn[T]("ID", func(t T) string { return t.Id }),
				output.StringColumn[T]("Type", func(t T) string { return t.Type }),
				output.FloatColumn[T]("Amount", func(t T) float64 { return t.Amount }),
				output.StringColumn[T]("Currency", func(t T) string { return t.Currency }),
				output.StringColumn[T]("Status", func(t T) string { return t.Status }),
				output.BoolColumn[T]("Success", func(t T) bool { return t.Success }),
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
	cmd.Flags().StringVar(&sort, "sort", "", "Sort field (created_at or updated_at)")
	cmd.Flags().StringVar(&transactionType, "type", "", "Filter by type (e.g. payment, refund)")
	cmd.Flags().StringVar(&success, "success", "", "Filter by success (true or false)")

	return cmd
}
