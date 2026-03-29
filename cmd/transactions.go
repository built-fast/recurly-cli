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

			columns := []output.Column{
				{Header: "ID", Extract: func(v any) string { return v.(recurly.Transaction).Id }},
				{Header: "Type", Extract: func(v any) string { return v.(recurly.Transaction).Type }},
				{Header: "Account", Extract: func(v any) string { return v.(recurly.Transaction).Account.Code }},
				{Header: "Status", Extract: func(v any) string { return v.(recurly.Transaction).Status }},
				{Header: "Currency", Extract: func(v any) string { return v.(recurly.Transaction).Currency }},
				{Header: "Amount", Extract: func(v any) string {
					return fmt.Sprintf("%.2f", v.(recurly.Transaction).Amount)
				}},
				{Header: "Success", Extract: func(v any) string {
					return fmt.Sprintf("%t", v.(recurly.Transaction).Success)
				}},
				{Header: "Origin", Extract: func(v any) string { return v.(recurly.Transaction).Origin }},
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
	return []output.Column{
		{Header: "ID", Extract: func(v any) string { return v.(*recurly.Transaction).Id }},
		{Header: "UUID", Extract: func(v any) string { return v.(*recurly.Transaction).Uuid }},
		{Header: "Type", Extract: func(v any) string { return v.(*recurly.Transaction).Type }},
		{Header: "Origin", Extract: func(v any) string { return v.(*recurly.Transaction).Origin }},
		{Header: "Status", Extract: func(v any) string { return v.(*recurly.Transaction).Status }},
		{Header: "Success", Extract: func(v any) string {
			return fmt.Sprintf("%t", v.(*recurly.Transaction).Success)
		}},
		{Header: "Amount", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Transaction).Amount)
		}},
		{Header: "Currency", Extract: func(v any) string { return v.(*recurly.Transaction).Currency }},
		{Header: "Account ID", Extract: func(v any) string { return v.(*recurly.Transaction).Account.Id }},
		{Header: "Account Code", Extract: func(v any) string { return v.(*recurly.Transaction).Account.Code }},
		{Header: "Invoice ID", Extract: func(v any) string { return v.(*recurly.Transaction).Invoice.Id }},
		{Header: "Invoice Number", Extract: func(v any) string { return v.(*recurly.Transaction).Invoice.Number }},
		{Header: "Collection Method", Extract: func(v any) string { return v.(*recurly.Transaction).CollectionMethod }},
		{Header: "Payment Method Type", Extract: func(v any) string { return v.(*recurly.Transaction).PaymentMethod.Object }},
		{Header: "Payment Method Card Type", Extract: func(v any) string { return v.(*recurly.Transaction).PaymentMethod.CardType }},
		{Header: "Payment Method Last Four", Extract: func(v any) string { return v.(*recurly.Transaction).PaymentMethod.LastFour }},
		{Header: "IP Address", Extract: func(v any) string { return v.(*recurly.Transaction).IpAddressV4 }},
		{Header: "Status Code", Extract: func(v any) string { return v.(*recurly.Transaction).StatusCode }},
		{Header: "Status Message", Extract: func(v any) string { return v.(*recurly.Transaction).StatusMessage }},
		{Header: "Refunded", Extract: func(v any) string {
			return fmt.Sprintf("%t", v.(*recurly.Transaction).Refunded)
		}},
		{Header: "Created At", Extract: func(v any) string {
			txn := v.(*recurly.Transaction)
			if txn.CreatedAt != nil {
				return txn.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Updated At", Extract: func(v any) string {
			txn := v.(*recurly.Transaction)
			if txn.UpdatedAt != nil {
				return txn.UpdatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Voided At", Extract: func(v any) string {
			txn := v.(*recurly.Transaction)
			if txn.VoidedAt != nil {
				return txn.VoidedAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}
