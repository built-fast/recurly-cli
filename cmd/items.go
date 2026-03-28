package cmd

import (
	"fmt"
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
)

func newItemsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items",
		Short: "Manage items",
	}
	cmd.AddCommand(newItemsListCmd())
	cmd.AddCommand(newItemsGetCmd())
	return cmd
}

func itemDetailColumns() []output.Column {
	return []output.Column{
		{Header: "Code", Extract: func(v any) string { return v.(*recurly.Item).Code }},
		{Header: "Name", Extract: func(v any) string { return v.(*recurly.Item).Name }},
		{Header: "Description", Extract: func(v any) string { return v.(*recurly.Item).Description }},
		{Header: "External SKU", Extract: func(v any) string { return v.(*recurly.Item).ExternalSku }},
		{Header: "Accounting Code", Extract: func(v any) string { return v.(*recurly.Item).AccountingCode }},
		{Header: "Revenue Schedule Type", Extract: func(v any) string { return v.(*recurly.Item).RevenueScheduleType }},
		{Header: "Tax Code", Extract: func(v any) string { return v.(*recurly.Item).TaxCode }},
		{Header: "Tax Exempt", Extract: func(v any) string {
			return fmt.Sprintf("%t", v.(*recurly.Item).TaxExempt)
		}},
		{Header: "Avalara Transaction Type", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(*recurly.Item).AvalaraTransactionType)
		}},
		{Header: "Avalara Service Type", Extract: func(v any) string {
			return fmt.Sprintf("%d", v.(*recurly.Item).AvalaraServiceType)
		}},
		{Header: "Harmonized System Code", Extract: func(v any) string { return v.(*recurly.Item).HarmonizedSystemCode }},
		{Header: "State", Extract: func(v any) string { return v.(*recurly.Item).State }},
		{Header: "Created At", Extract: func(v any) string {
			item := v.(*recurly.Item)
			if item.CreatedAt != nil {
				return item.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Updated At", Extract: func(v any) string {
			item := v.(*recurly.Item)
			if item.UpdatedAt != nil {
				return item.UpdatedAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}

func newItemsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <item_id>",
		Short: "Get item details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newItemAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			item, err := c.GetItem(args[0])
			if err != nil {
				return err
			}

			columns := itemDetailColumns()

			formatted, err := output.FormatOne(format, columns, item)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

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
			c, err := newItemAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

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
	cmd.Flags().StringVar(&beginTime, "begin-time", "", "Filter by begin time (ISO8601 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter by end time (ISO8601 format)")

	return cmd
}
