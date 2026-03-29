package cmd

import (
	"fmt"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
)

func newInvoicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoices",
		Short: "Manage invoices",
	}
	cmd.AddCommand(newInvoicesListCmd())
	cmd.AddCommand(withWatch(newInvoicesGetCmd()))
	cmd.AddCommand(newInvoicesVoidCmd())
	cmd.AddCommand(newInvoicesCollectCmd())
	cmd.AddCommand(newInvoicesMarkFailedCmd())
	cmd.AddCommand(newInvoicesLineItemsCmd())
	return cmd
}

func newInvoicesListCmd() *cobra.Command {
	var (
		limit       int
		all         bool
		order       string
		sort        string
		state       string
		invoiceType string
		beginTime   string
		endTime     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoices",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewInvoiceAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			params := &recurly.ListInvoicesParams{}

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

			lister, err := c.ListInvoices(params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.Invoice](lister, limit, all)
			if err != nil {
				return err
			}

			type I = recurly.Invoice
			columns := output.ToColumns([]output.TypedColumn[I]{
				output.StringColumn[I]("ID", func(i I) string { return i.Id }),
				output.StringColumn[I]("Number", func(i I) string { return i.Number }),
				output.StringColumn[I]("Type", func(i I) string { return i.Type }),
				output.StringColumn[I]("Account", func(i I) string { return i.Account.Code }),
				output.StringColumn[I]("State", func(i I) string { return i.State }),
				output.StringColumn[I]("Currency", func(i I) string { return i.Currency }),
				output.FloatColumn[I]("Total", func(i I) float64 { return i.Total }),
				output.FloatColumn[I]("Balance", func(i I) float64 { return i.Balance }),
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
	cmd.Flags().StringVar(&sort, "sort", "", "Sort field (e.g. created_at, updated_at)")
	cmd.Flags().StringVar(&state, "state", "", "Filter by state")
	cmd.Flags().StringVar(&invoiceType, "type", "", "Filter by type (charge, credit, non-legacy, legacy)")
	cmd.Flags().StringVar(&beginTime, "begin-time", "", "Filter by begin time (ISO8601 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter by end time (ISO8601 format)")

	return cmd
}

func newInvoicesGetCmd() *cobra.Command {
	var lineItemsCount int

	cmd := &cobra.Command{
		Use:   "get <invoice_id>",
		Short: "Get invoice details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewInvoiceAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			invoice, err := c.GetInvoice(args[0])
			if err != nil {
				return err
			}

			showLineItems := cmd.Flags().Changed("line-items")

			if showLineItems {
				params := &recurly.ListInvoiceLineItemsParams{}
				if lineItemsCount > 0 {
					params.Limit = recurly.Int(lineItemsCount)
				}

				lister, err := c.ListInvoiceLineItems(args[0], params)
				if err != nil {
					return err
				}

				result, err := pagination.Collect[recurly.LineItem](lister, lineItemsCount, false)
				if err != nil {
					return err
				}

				invoice.LineItems = result.Items
				invoice.HasMoreLineItems = result.HasMore
			}

			// For JSON/jq output, format the whole invoice object
			if cfg.Format == "json" || cfg.Format == "json-pretty" || cfg.HasJQ() {
				formatted, err := output.FormatOne(cfg, nil, invoice)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
				return err
			}

			// Table output: detail view
			columns := invoiceDetailColumns()
			formatted, err := output.FormatOne(cfg, columns, invoice)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			if err != nil {
				return err
			}

			if showLineItems && len(invoice.LineItems) > 0 {
				fmt.Fprintln(cmd.OutOrStdout())
				fmt.Fprintln(cmd.OutOrStdout(), "Line Items:")

				lineItemCols := lineItemColumns()
				items := make([]any, len(invoice.LineItems))
				for i, li := range invoice.LineItems {
					items[i] = li
				}

				tableOut, err := output.FormatList(&output.Config{Format: "table"}, lineItemCols, items, false)
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), tableOut)

				if invoice.HasMoreLineItems {
					fmt.Fprintf(cmd.OutOrStdout(),
						"Showing %d line items (more available). Use `recurly invoices line-items %s` to view all.\n",
						len(invoice.LineItems), args[0])
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&lineItemsCount, "line-items", 20, "Show line items (optional count, default: 20)")
	cmd.Flags().Lookup("line-items").NoOptDefVal = "20"

	return cmd
}

func newInvoicesVoidCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "void <invoice_id>",
		Short: "Void a credit invoice",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			invoiceID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to void invoice %s? [y/N] ", invoiceID))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Void cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewInvoiceAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			invoice, err := c.VoidInvoice(invoiceID)
			if err != nil {
				return err
			}

			// For JSON/jq output, format the whole invoice object
			if cfg.Format == "json" || cfg.Format == "json-pretty" || cfg.HasJQ() {
				formatted, err := output.FormatOne(cfg, nil, invoice)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
				return err
			}

			// Table output: detail view
			columns := invoiceDetailColumns()
			formatted, err := output.FormatOne(cfg, columns, invoice)
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

func newInvoicesCollectCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "collect <invoice_id>",
		Short: "Attempt collection on an invoice",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			invoiceID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to collect invoice %s? [y/N] ", invoiceID))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Collection cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewInvoiceAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			invoice, err := c.CollectInvoice(invoiceID, &recurly.CollectInvoiceParams{})
			if err != nil {
				return err
			}

			// For JSON/jq output, format the whole invoice object
			if cfg.Format == "json" || cfg.Format == "json-pretty" || cfg.HasJQ() {
				formatted, err := output.FormatOne(cfg, nil, invoice)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
				return err
			}

			// Table output: detail view
			columns := invoiceDetailColumns()
			formatted, err := output.FormatOne(cfg, columns, invoice)
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

func newInvoicesMarkFailedCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "mark-failed <invoice_id>",
		Short: "Mark an open invoice as failed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			invoiceID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to mark invoice %s as failed? [y/N] ", invoiceID))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Mark failed cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewInvoiceAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			invoice, err := c.MarkInvoiceFailed(invoiceID)
			if err != nil {
				return err
			}

			// For JSON/jq output, format the whole invoice object
			if cfg.Format == "json" || cfg.Format == "json-pretty" || cfg.HasJQ() {
				formatted, err := output.FormatOne(cfg, nil, invoice)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
				return err
			}

			// Table output: detail view
			columns := invoiceDetailColumns()
			formatted, err := output.FormatOne(cfg, columns, invoice)
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

func newInvoicesLineItemsCmd() *cobra.Command {
	var (
		limit     int
		all       bool
		order     string
		sort      string
		beginTime string
		endTime   string
	)

	cmd := &cobra.Command{
		Use:   "line-items <invoice_id>",
		Short: "List line items for an invoice",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewInvoiceAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			params := &recurly.ListInvoiceLineItemsParams{}

			if limit > 0 {
				params.Limit = recurly.Int(limit)
			}
			if cmd.Flags().Changed("order") {
				params.Order = recurly.String(order)
			}
			if cmd.Flags().Changed("sort") {
				params.Sort = recurly.String(sort)
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

			lister, err := c.ListInvoiceLineItems(args[0], params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.LineItem](lister, limit, all)
			if err != nil {
				return err
			}

			columns := lineItemColumns()

			items := make([]any, len(result.Items))
			for i, li := range result.Items {
				items[i] = li
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
	cmd.Flags().StringVar(&beginTime, "begin-time", "", "Filter by begin time (ISO8601 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter by end time (ISO8601 format)")

	return cmd
}

func invoiceDetailColumns() []output.Column {
	type I = *recurly.Invoice
	return output.ToColumns([]output.TypedColumn[I]{
		output.StringColumn[I]("ID", func(i I) string { return i.Id }),
		output.StringColumn[I]("UUID", func(i I) string { return i.Uuid }),
		output.StringColumn[I]("Number", func(i I) string { return i.Number }),
		output.StringColumn[I]("Type", func(i I) string { return i.Type }),
		output.StringColumn[I]("Origin", func(i I) string { return i.Origin }),
		output.StringColumn[I]("State", func(i I) string { return i.State }),
		output.StringColumn[I]("Account ID", func(i I) string { return i.Account.Id }),
		output.StringColumn[I]("Account Code", func(i I) string { return i.Account.Code }),
		output.StringColumn[I]("Collection Method", func(i I) string { return i.CollectionMethod }),
		output.StringColumn[I]("Currency", func(i I) string { return i.Currency }),
		output.FloatColumn[I]("Subtotal", func(i I) float64 { return i.Subtotal }),
		output.FloatColumn[I]("Discount", func(i I) float64 { return i.Discount }),
		output.FloatColumn[I]("Tax", func(i I) float64 { return i.Tax }),
		output.FloatColumn[I]("Total", func(i I) float64 { return i.Total }),
		output.FloatColumn[I]("Paid", func(i I) float64 { return i.Paid }),
		output.FloatColumn[I]("Balance", func(i I) float64 { return i.Balance }),
		output.FloatColumn[I]("Refundable Amount", func(i I) float64 { return i.RefundableAmount }),
		output.StringColumn[I]("PO Number", func(i I) string { return i.PoNumber }),
		output.IntColumn[I]("Net Terms", func(i I) int { return i.NetTerms }),
		output.StringColumn[I]("Net Terms Type", func(i I) string { return i.NetTermsType }),
		output.TimeColumn[I]("Created At", func(i I) *time.Time { return i.CreatedAt }),
		output.TimeColumn[I]("Updated At", func(i I) *time.Time { return i.UpdatedAt }),
		output.TimeColumn[I]("Due At", func(i I) *time.Time { return i.DueAt }),
		output.TimeColumn[I]("Closed At", func(i I) *time.Time { return i.ClosedAt }),
	})
}

func lineItemColumns() []output.Column {
	type L = recurly.LineItem
	return output.ToColumns([]output.TypedColumn[L]{
		output.StringColumn[L]("ID", func(l L) string { return l.Id }),
		output.StringColumn[L]("Type", func(l L) string { return l.Type }),
		output.StringColumn[L]("Description", func(l L) string { return l.Description }),
		output.StringColumn[L]("Currency", func(l L) string { return l.Currency }),
		output.FloatColumn[L]("Unit Amount", func(l L) float64 { return l.UnitAmount }),
		output.IntColumn[L]("Quantity", func(l L) int { return l.Quantity }),
		output.FloatColumn[L]("Subtotal", func(l L) float64 { return l.Subtotal }),
		output.FloatColumn[L]("Tax", func(l L) float64 { return l.Tax }),
		output.FloatColumn[L]("Amount", func(l L) float64 { return l.Amount }),
	})
}
