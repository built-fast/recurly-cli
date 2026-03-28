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
			c, err := newAccountNestedAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

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

			columns := []output.Column{
				{Header: "ID", Extract: func(v any) string { return v.(recurly.Subscription).Id }},
				{Header: "Account Code", Extract: func(v any) string { return v.(recurly.Subscription).Account.Code }},
				{Header: "Plan Code", Extract: func(v any) string { return v.(recurly.Subscription).Plan.Code }},
				{Header: "State", Extract: func(v any) string { return v.(recurly.Subscription).State }},
				{Header: "Currency", Extract: func(v any) string { return v.(recurly.Subscription).Currency }},
				{Header: "Unit Amount", Extract: func(v any) string {
					return fmt.Sprintf("%.2f", v.(recurly.Subscription).UnitAmount)
				}},
				{Header: "Current Period Ends At", Extract: func(v any) string {
					s := v.(recurly.Subscription)
					if s.CurrentPeriodEndsAt != nil {
						return s.CurrentPeriodEndsAt.Format(time.RFC3339)
					}
					return ""
				}},
			}

			items := make([]any, len(result.Items))
			for i, s := range result.Items {
				items[i] = s
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
	cmd.Flags().StringVar(&state, "state", "", "Filter by state (active, canceled, expired, future, in_trial, live)")

	return cmd
}
