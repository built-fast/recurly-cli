package cmd

import (
	"fmt"
	"time"

	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Manage accounts",
	}
	cmd.AddCommand(newAccountsListCmd())
	return cmd
}

func newAccountsListCmd() *cobra.Command {
	var (
		limit      int
		all        bool
		order      string
		sort       string
		email      string
		subscriber string
		pastDue    bool
		beginTime  string
		endTime    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.NewClient()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			params := &recurly.ListAccountsParams{}

			if limit > 0 {
				params.Limit = recurly.Int(limit)
			}
			if cmd.Flags().Changed("order") {
				params.Order = recurly.String(order)
			}
			if cmd.Flags().Changed("sort") {
				params.Sort = recurly.String(sort)
			}
			if cmd.Flags().Changed("email") {
				params.Email = recurly.String(email)
			}
			if cmd.Flags().Changed("subscriber") {
				b := subscriber == "true"
				params.Subscriber = recurly.Bool(b)
			}
			if pastDue {
				params.PastDue = recurly.String("true")
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

			lister, err := c.ListAccounts(params)
			if err != nil {
				return err
			}

			accounts, err := pagination.Collect[recurly.Account](lister, limit, all)
			if err != nil {
				return err
			}

			columns := []output.Column{
				{Header: "Code", Extract: func(v any) string { return v.(recurly.Account).Code }},
				{Header: "Email", Extract: func(v any) string { return v.(recurly.Account).Email }},
				{Header: "First Name", Extract: func(v any) string { return v.(recurly.Account).FirstName }},
				{Header: "Last Name", Extract: func(v any) string { return v.(recurly.Account).LastName }},
				{Header: "Company", Extract: func(v any) string { return v.(recurly.Account).Company }},
				{Header: "State", Extract: func(v any) string { return v.(recurly.Account).State }},
				{Header: "Created At", Extract: func(v any) string {
					a := v.(recurly.Account)
					if a.CreatedAt != nil {
						return a.CreatedAt.Format(time.RFC3339)
					}
					return ""
				}},
			}

			items := make([]any, len(accounts))
			for i, a := range accounts {
				items[i] = a
			}

			formatted, err := output.FormatList(format, columns, items)
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
	cmd.Flags().StringVar(&email, "email", "", "Filter by exact email address")
	cmd.Flags().StringVar(&subscriber, "subscriber", "", "Filter for accounts with/without active subscriptions (true or false)")
	cmd.Flags().BoolVar(&pastDue, "past-due", false, "Filter for accounts with past-due invoices")
	cmd.Flags().StringVar(&beginTime, "begin-time", "", "Filter by begin time (ISO8601 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter by end time (ISO8601 format)")

	return cmd
}
