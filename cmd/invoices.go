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
			c, err := newInvoiceAPI()
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

			columns := []output.Column{
				{Header: "ID", Extract: func(v any) string { return v.(recurly.Invoice).Id }},
				{Header: "Number", Extract: func(v any) string { return v.(recurly.Invoice).Number }},
				{Header: "Type", Extract: func(v any) string { return v.(recurly.Invoice).Type }},
				{Header: "Account", Extract: func(v any) string { return v.(recurly.Invoice).Account.Code }},
				{Header: "State", Extract: func(v any) string { return v.(recurly.Invoice).State }},
				{Header: "Currency", Extract: func(v any) string { return v.(recurly.Invoice).Currency }},
				{Header: "Total", Extract: func(v any) string {
					return fmt.Sprintf("%.2f", v.(recurly.Invoice).Total)
				}},
				{Header: "Balance", Extract: func(v any) string {
					return fmt.Sprintf("%.2f", v.(recurly.Invoice).Balance)
				}},
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
			c, err := newInvoiceAPI()
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

			c, err := newInvoiceAPI()
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

			c, err := newInvoiceAPI()
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

			c, err := newInvoiceAPI()
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
			c, err := newInvoiceAPI()
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
	return []output.Column{
		{Header: "ID", Extract: func(v any) string { return v.(*recurly.Invoice).Id }},
		{Header: "UUID", Extract: func(v any) string { return v.(*recurly.Invoice).Uuid }},
		{Header: "Number", Extract: func(v any) string { return v.(*recurly.Invoice).Number }},
		{Header: "Type", Extract: func(v any) string { return v.(*recurly.Invoice).Type }},
		{Header: "Origin", Extract: func(v any) string { return v.(*recurly.Invoice).Origin }},
		{Header: "State", Extract: func(v any) string { return v.(*recurly.Invoice).State }},
		{Header: "Account ID", Extract: func(v any) string { return v.(*recurly.Invoice).Account.Id }},
		{Header: "Account Code", Extract: func(v any) string { return v.(*recurly.Invoice).Account.Code }},
		{Header: "Collection Method", Extract: func(v any) string { return v.(*recurly.Invoice).CollectionMethod }},
		{Header: "Currency", Extract: func(v any) string { return v.(*recurly.Invoice).Currency }},
		{Header: "Subtotal", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Invoice).Subtotal)
		}},
		{Header: "Discount", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Invoice).Discount)
		}},
		{Header: "Tax", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Invoice).Tax)
		}},
		{Header: "Total", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Invoice).Total)
		}},
		{Header: "Paid", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Invoice).Paid)
		}},
		{Header: "Balance", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Invoice).Balance)
		}},
		{Header: "Refundable Amount", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(*recurly.Invoice).RefundableAmount)
		}},
		{Header: "PO Number", Extract: func(v any) string { return v.(*recurly.Invoice).PoNumber }},
		{Header: "Net Terms", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(*recurly.Invoice).NetTerms)
		}},
		{Header: "Net Terms Type", Extract: func(v any) string { return v.(*recurly.Invoice).NetTermsType }},
		{Header: "Created At", Extract: func(v any) string {
			inv := v.(*recurly.Invoice)
			if inv.CreatedAt != nil {
				return inv.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Updated At", Extract: func(v any) string {
			inv := v.(*recurly.Invoice)
			if inv.UpdatedAt != nil {
				return inv.UpdatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Due At", Extract: func(v any) string {
			inv := v.(*recurly.Invoice)
			if inv.DueAt != nil {
				return inv.DueAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Closed At", Extract: func(v any) string {
			inv := v.(*recurly.Invoice)
			if inv.ClosedAt != nil {
				return inv.ClosedAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}

func lineItemColumns() []output.Column {
	return []output.Column{
		{Header: "ID", Extract: func(v any) string { return v.(recurly.LineItem).Id }},
		{Header: "Type", Extract: func(v any) string { return v.(recurly.LineItem).Type }},
		{Header: "Description", Extract: func(v any) string { return v.(recurly.LineItem).Description }},
		{Header: "Currency", Extract: func(v any) string { return v.(recurly.LineItem).Currency }},
		{Header: "Unit Amount", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(recurly.LineItem).UnitAmount)
		}},
		{Header: "Quantity", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(recurly.LineItem).Quantity)
		}},
		{Header: "Subtotal", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(recurly.LineItem).Subtotal)
		}},
		{Header: "Tax", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(recurly.LineItem).Tax)
		}},
		{Header: "Amount", Extract: func(v any) string {
			return fmt.Sprintf("%.2f", v.(recurly.LineItem).Amount)
		}},
	}
}
