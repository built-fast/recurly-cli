package cmd

import (
	"fmt"
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
)

func newItemsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items",
		Short: "Manage items",
	}
	cmd.AddCommand(newItemsListCmd())
	cmd.AddCommand(withWatch(newItemsGetCmd()))
	cmd.AddCommand(withFromFile(newItemsCreateCmd()))
	cmd.AddCommand(withFromFile(newItemsUpdateCmd()))
	cmd.AddCommand(newItemsDeactivateCmd())
	cmd.AddCommand(newItemsReactivateCmd())
	return cmd
}

func itemDetailColumns() []output.Column {
	type I = *recurly.Item
	return output.ToColumns([]output.TypedColumn[I]{
		output.StringColumn[I]("Code", func(i I) string { return i.Code }),
		output.StringColumn[I]("Name", func(i I) string { return i.Name }),
		output.StringColumn[I]("Description", func(i I) string { return i.Description }),
		output.StringColumn[I]("External SKU", func(i I) string { return i.ExternalSku }),
		output.StringColumn[I]("Accounting Code", func(i I) string { return i.AccountingCode }),
		output.StringColumn[I]("Revenue Schedule Type", func(i I) string { return i.RevenueScheduleType }),
		output.StringColumn[I]("Tax Code", func(i I) string { return i.TaxCode }),
		output.BoolColumn[I]("Tax Exempt", func(i I) bool { return i.TaxExempt }),
		output.IntColumn[I]("Avalara Transaction Type", func(i I) int { return i.AvalaraTransactionType }),
		output.IntColumn[I]("Avalara Service Type", func(i I) int { return i.AvalaraServiceType }),
		output.StringColumn[I]("Harmonized System Code", func(i I) string { return i.HarmonizedSystemCode }),
		output.StringColumn[I]("State", func(i I) string { return i.State }),
		output.TimeColumn[I]("Created At", func(i I) *time.Time { return i.CreatedAt }),
		output.TimeColumn[I]("Updated At", func(i I) *time.Time { return i.UpdatedAt }),
	})
}

func newItemsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <item_id>",
		Short: "Get item details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewItemAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			item, err := c.GetItem(args[0])
			if err != nil {
				return err
			}

			columns := itemDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, item)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newItemsCreateCmd() *cobra.Command {
	var (
		code                 string
		name                 string
		description          string
		externalSku          string
		accountingCode       string
		revenueScheduleType  string
		taxCode              string
		taxExempt            bool
		avalaraTransType     int
		avalaraServiceType   int
		harmonizedSystemCode string
		currencies           []string
		unitAmounts          []float64
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an item",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate currency/unit-amount pairing
			if cmd.Flags().Changed("currency") || cmd.Flags().Changed("unit-amount") {
				if len(currencies) != len(unitAmounts) {
					return fmt.Errorf("number of --currency values must match --unit-amount values")
				}
			}

			c, err := AppFromContext(cmd.Context()).NewItemAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
			body := &recurly.ItemCreate{}

			if cmd.Flags().Changed("code") {
				body.Code = recurly.String(code)
			}
			if cmd.Flags().Changed("name") {
				body.Name = recurly.String(name)
			}
			if cmd.Flags().Changed("description") {
				body.Description = recurly.String(description)
			}
			if cmd.Flags().Changed("external-sku") {
				body.ExternalSku = recurly.String(externalSku)
			}
			if cmd.Flags().Changed("accounting-code") {
				body.AccountingCode = recurly.String(accountingCode)
			}
			if cmd.Flags().Changed("revenue-schedule-type") {
				body.RevenueScheduleType = recurly.String(revenueScheduleType)
			}
			if cmd.Flags().Changed("tax-code") {
				body.TaxCode = recurly.String(taxCode)
			}
			if cmd.Flags().Changed("tax-exempt") {
				body.TaxExempt = recurly.Bool(taxExempt)
			}
			if cmd.Flags().Changed("avalara-transaction-type") {
				body.AvalaraTransactionType = recurly.Int(avalaraTransType)
			}
			if cmd.Flags().Changed("avalara-service-type") {
				body.AvalaraServiceType = recurly.Int(avalaraServiceType)
			}
			if cmd.Flags().Changed("harmonized-system-code") {
				body.HarmonizedSystemCode = recurly.String(harmonizedSystemCode)
			}

			// Multi-currency pricing
			if cmd.Flags().Changed("currency") {
				pricings := make([]recurly.PricingCreate, len(currencies))
				for i, cur := range currencies {
					pricings[i] = recurly.PricingCreate{
						Currency:   recurly.String(cur),
						UnitAmount: float64Ptr(unitAmounts[i]),
					}
				}
				body.Currencies = &pricings
			}

			item, err := c.CreateItem(body)
			if err != nil {
				return err
			}

			columns := itemDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, item)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&code, "code", "", "Unique item code")
	cmd.Flags().StringVar(&name, "name", "", "Item name")
	cmd.Flags().StringVar(&description, "description", "", "Item description")
	cmd.Flags().StringVar(&externalSku, "external-sku", "", "External stock keeping unit")
	cmd.Flags().StringVar(&accountingCode, "accounting-code", "", "Accounting code")
	cmd.Flags().StringVar(&revenueScheduleType, "revenue-schedule-type", "", "Revenue schedule type")
	cmd.Flags().StringVar(&taxCode, "tax-code", "", "Tax code")
	cmd.Flags().BoolVar(&taxExempt, "tax-exempt", false, "Tax exempt status")
	cmd.Flags().IntVar(&avalaraTransType, "avalara-transaction-type", 0, "Avalara transaction type")
	cmd.Flags().IntVar(&avalaraServiceType, "avalara-service-type", 0, "Avalara service type")
	cmd.Flags().StringVar(&harmonizedSystemCode, "harmonized-system-code", "", "Harmonized System (HS) code")
	cmd.Flags().StringSliceVar(&currencies, "currency", nil, "Currency code (repeatable, positionally matched with --unit-amount)")
	cmd.Flags().Float64SliceVar(&unitAmounts, "unit-amount", nil, "Unit amount (repeatable, positionally matched with --currency)")

	return cmd
}

func newItemsUpdateCmd() *cobra.Command {
	var (
		code                 string
		name                 string
		description          string
		externalSku          string
		accountingCode       string
		revenueScheduleType  string
		taxCode              string
		taxExempt            bool
		avalaraTransType     int
		avalaraServiceType   int
		harmonizedSystemCode string
		currencies           []string
		unitAmounts          []float64
	)

	cmd := &cobra.Command{
		Use:   "update <item_id>",
		Short: "Update an item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate currency/unit-amount pairing
			if cmd.Flags().Changed("currency") || cmd.Flags().Changed("unit-amount") {
				if len(currencies) != len(unitAmounts) {
					return fmt.Errorf("number of --currency values must match --unit-amount values")
				}
			}

			c, err := AppFromContext(cmd.Context()).NewItemAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
			body := &recurly.ItemUpdate{}

			if cmd.Flags().Changed("code") {
				body.Code = recurly.String(code)
			}
			if cmd.Flags().Changed("name") {
				body.Name = recurly.String(name)
			}
			if cmd.Flags().Changed("description") {
				body.Description = recurly.String(description)
			}
			if cmd.Flags().Changed("external-sku") {
				body.ExternalSku = recurly.String(externalSku)
			}
			if cmd.Flags().Changed("accounting-code") {
				body.AccountingCode = recurly.String(accountingCode)
			}
			if cmd.Flags().Changed("revenue-schedule-type") {
				body.RevenueScheduleType = recurly.String(revenueScheduleType)
			}
			if cmd.Flags().Changed("tax-code") {
				body.TaxCode = recurly.String(taxCode)
			}
			if cmd.Flags().Changed("tax-exempt") {
				body.TaxExempt = recurly.Bool(taxExempt)
			}
			if cmd.Flags().Changed("avalara-transaction-type") {
				body.AvalaraTransactionType = recurly.Int(avalaraTransType)
			}
			if cmd.Flags().Changed("avalara-service-type") {
				body.AvalaraServiceType = recurly.Int(avalaraServiceType)
			}
			if cmd.Flags().Changed("harmonized-system-code") {
				body.HarmonizedSystemCode = recurly.String(harmonizedSystemCode)
			}

			// Multi-currency pricing
			if cmd.Flags().Changed("currency") {
				pricings := make([]recurly.PricingCreate, len(currencies))
				for i, cur := range currencies {
					pricings[i] = recurly.PricingCreate{
						Currency:   recurly.String(cur),
						UnitAmount: float64Ptr(unitAmounts[i]),
					}
				}
				body.Currencies = &pricings
			}

			item, err := c.UpdateItem(args[0], body)
			if err != nil {
				return err
			}

			columns := itemDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, item)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&code, "code", "", "Unique item code")
	cmd.Flags().StringVar(&name, "name", "", "Item name")
	cmd.Flags().StringVar(&description, "description", "", "Item description")
	cmd.Flags().StringVar(&externalSku, "external-sku", "", "External stock keeping unit")
	cmd.Flags().StringVar(&accountingCode, "accounting-code", "", "Accounting code")
	cmd.Flags().StringVar(&revenueScheduleType, "revenue-schedule-type", "", "Revenue schedule type")
	cmd.Flags().StringVar(&taxCode, "tax-code", "", "Tax code")
	cmd.Flags().BoolVar(&taxExempt, "tax-exempt", false, "Tax exempt status")
	cmd.Flags().IntVar(&avalaraTransType, "avalara-transaction-type", 0, "Avalara transaction type")
	cmd.Flags().IntVar(&avalaraServiceType, "avalara-service-type", 0, "Avalara service type")
	cmd.Flags().StringVar(&harmonizedSystemCode, "harmonized-system-code", "", "Harmonized System (HS) code")
	cmd.Flags().StringSliceVar(&currencies, "currency", nil, "Currency code (repeatable, positionally matched with --unit-amount)")
	cmd.Flags().Float64SliceVar(&unitAmounts, "unit-amount", nil, "Unit amount (repeatable, positionally matched with --currency)")

	return cmd
}

func newItemsListCmd() *cobra.Command {
	var (
		limit     int
		all       bool
		order     string
		sort      string
		state     string
		beginTime string
		endTime   string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List items",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewItemAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			params := &recurly.ListItemsParams{}

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

			lister, err := c.ListItems(params)
			if err != nil {
				return err
			}

			result, err := pagination.Collect[recurly.Item](lister, limit, all)
			if err != nil {
				return err
			}

			columns := []output.Column{
				{Header: "Code", Extract: func(v any) string { return v.(recurly.Item).Code }},
				{Header: "Name", Extract: func(v any) string { return v.(recurly.Item).Name }},
				{Header: "External SKU", Extract: func(v any) string { return v.(recurly.Item).ExternalSku }},
				{Header: "State", Extract: func(v any) string { return v.(recurly.Item).State }},
				{Header: "Created At", Extract: func(v any) string {
					item := v.(recurly.Item)
					if item.CreatedAt != nil {
						return item.CreatedAt.Format(time.RFC3339)
					}
					return ""
				}},
			}

			items := make([]any, len(result.Items))
			for i, item := range result.Items {
				items[i] = item
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
	cmd.Flags().StringVar(&state, "state", "", "Filter by state (active or inactive)")
	cmd.Flags().StringVar(&beginTime, "begin-time", "", "Filter by begin time (ISO8601 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter by end time (ISO8601 format)")

	return cmd
}

func newItemsDeactivateCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "deactivate <item_id>",
		Short: "Deactivate an item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			itemID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to deactivate item %s? [y/N] ", itemID))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deactivation cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewItemAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			item, err := c.DeactivateItem(itemID)
			if err != nil {
				return err
			}

			columns := itemDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, item)
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

func newItemsReactivateCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "reactivate <item_id>",
		Short: "Reactivate an item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			itemID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to reactivate item %s? [y/N] ", itemID))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Reactivation cancelled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewItemAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			item, err := c.ReactivateItem(itemID)
			if err != nil {
				return err
			}

			columns := itemDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, item)
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
