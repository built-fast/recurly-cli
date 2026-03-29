package cmd

import (
	"fmt"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/built-fast/recurly-cli/internal/pagination"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
)

func newAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Manage accounts",
	}
	cmd.AddCommand(newAccountsListCmd())
	cmd.AddCommand(withWatch(newAccountsGetCmd()))
	cmd.AddCommand(withFromFile(newAccountsCreateCmd()))
	cmd.AddCommand(withFromFile(newAccountsUpdateCmd()))
	cmd.AddCommand(newAccountsDeactivateCmd())
	cmd.AddCommand(newAccountsReactivateCmd())
	cmd.AddCommand(newAccountBillingInfoCmd())
	cmd.AddCommand(newAccountSubscriptionsCmd())
	cmd.AddCommand(newAccountInvoicesCmd())
	cmd.AddCommand(newAccountTransactionsCmd())
	cmd.AddCommand(newAccountRedemptionsCmd())
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
			c, err := newAccountAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

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

			result, err := pagination.Collect[recurly.Account](lister, limit, all)
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

			items := make([]any, len(result.Items))
			for i, a := range result.Items {
				items[i] = a
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
	cmd.Flags().StringVar(&email, "email", "", "Filter by exact email address")
	cmd.Flags().StringVar(&subscriber, "subscriber", "", "Filter for accounts with/without active subscriptions (true or false)")
	cmd.Flags().BoolVar(&pastDue, "past-due", false, "Filter for accounts with past-due invoices")
	cmd.Flags().StringVar(&beginTime, "begin-time", "", "Filter by begin time (ISO8601 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter by end time (ISO8601 format)")

	return cmd
}

func newAccountsCreateCmd() *cobra.Command {
	var (
		code            string
		email           string
		firstName       string
		lastName        string
		company         string
		vatNumber       string
		taxExempt       bool
		preferredLocale string
		billTo          string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an account",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAccountAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &recurly.AccountCreate{}

			if cmd.Flags().Changed("code") {
				body.Code = recurly.String(code)
			}
			if cmd.Flags().Changed("email") {
				body.Email = recurly.String(email)
			}
			if cmd.Flags().Changed("first-name") {
				body.FirstName = recurly.String(firstName)
			}
			if cmd.Flags().Changed("last-name") {
				body.LastName = recurly.String(lastName)
			}
			if cmd.Flags().Changed("company") {
				body.Company = recurly.String(company)
			}
			if cmd.Flags().Changed("vat-number") {
				body.VatNumber = recurly.String(vatNumber)
			}
			if cmd.Flags().Changed("tax-exempt") {
				body.TaxExempt = recurly.Bool(taxExempt)
			}
			if cmd.Flags().Changed("preferred-locale") {
				body.PreferredLocale = recurly.String(preferredLocale)
			}
			if cmd.Flags().Changed("bill-to") {
				body.BillTo = recurly.String(billTo)
			}

			account, err := c.CreateAccount(body)
			if err != nil {
				return err
			}

			columns := accountDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, account)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&code, "code", "", "Unique account code")
	cmd.Flags().StringVar(&email, "email", "", "Account email address")
	cmd.Flags().StringVar(&firstName, "first-name", "", "First name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name")
	cmd.Flags().StringVar(&company, "company", "", "Company name")
	cmd.Flags().StringVar(&vatNumber, "vat-number", "", "VAT number")
	cmd.Flags().BoolVar(&taxExempt, "tax-exempt", false, "Tax exempt status")
	cmd.Flags().StringVar(&preferredLocale, "preferred-locale", "", "Preferred locale (e.g. en-US)")
	cmd.Flags().StringVar(&billTo, "bill-to", "", "Billing target (self or parent)")

	return cmd
}

func accountDetailColumns() []output.Column {
	return []output.Column{
		{Header: "Code", Extract: func(v any) string { return v.(*recurly.Account).Code }},
		{Header: "Email", Extract: func(v any) string { return v.(*recurly.Account).Email }},
		{Header: "First Name", Extract: func(v any) string { return v.(*recurly.Account).FirstName }},
		{Header: "Last Name", Extract: func(v any) string { return v.(*recurly.Account).LastName }},
		{Header: "Company", Extract: func(v any) string { return v.(*recurly.Account).Company }},
		{Header: "State", Extract: func(v any) string { return v.(*recurly.Account).State }},
		{Header: "Created At", Extract: func(v any) string {
			a := v.(*recurly.Account)
			if a.CreatedAt != nil {
				return a.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Updated At", Extract: func(v any) string {
			a := v.(*recurly.Account)
			if a.UpdatedAt != nil {
				return a.UpdatedAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}

func newAccountsUpdateCmd() *cobra.Command {
	var (
		email           string
		firstName       string
		lastName        string
		company         string
		vatNumber       string
		taxExempt       bool
		preferredLocale string
		billTo          string
	)

	cmd := &cobra.Command{
		Use:   "update <account_id>",
		Short: "Update an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAccountAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &recurly.AccountUpdate{}

			if cmd.Flags().Changed("email") {
				body.Email = recurly.String(email)
			}
			if cmd.Flags().Changed("first-name") {
				body.FirstName = recurly.String(firstName)
			}
			if cmd.Flags().Changed("last-name") {
				body.LastName = recurly.String(lastName)
			}
			if cmd.Flags().Changed("company") {
				body.Company = recurly.String(company)
			}
			if cmd.Flags().Changed("vat-number") {
				body.VatNumber = recurly.String(vatNumber)
			}
			if cmd.Flags().Changed("tax-exempt") {
				body.TaxExempt = recurly.Bool(taxExempt)
			}
			if cmd.Flags().Changed("preferred-locale") {
				body.PreferredLocale = recurly.String(preferredLocale)
			}
			if cmd.Flags().Changed("bill-to") {
				body.BillTo = recurly.String(billTo)
			}

			account, err := c.UpdateAccount(args[0], body)
			if err != nil {
				return err
			}

			columns := accountDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, account)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Account email address")
	cmd.Flags().StringVar(&firstName, "first-name", "", "First name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name")
	cmd.Flags().StringVar(&company, "company", "", "Company name")
	cmd.Flags().StringVar(&vatNumber, "vat-number", "", "VAT number")
	cmd.Flags().BoolVar(&taxExempt, "tax-exempt", false, "Tax exempt status")
	cmd.Flags().StringVar(&preferredLocale, "preferred-locale", "", "Preferred locale (e.g. en-US)")
	cmd.Flags().StringVar(&billTo, "bill-to", "", "Billing target (self or parent)")

	return cmd
}

func newAccountsDeactivateCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "deactivate <account_id>",
		Short: "Deactivate an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, "Are you sure you want to deactivate this account? [y/N] ")
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deactivation cancelled.")
					return err
				}
			}

			c, err := newAccountAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			account, err := c.DeactivateAccount(accountID)
			if err != nil {
				return err
			}

			columns := accountDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, account)
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

func newAccountsReactivateCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "reactivate <account_id>",
		Short: "Reactivate an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, "Are you sure you want to reactivate this account? [y/N] ")
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Reactivation cancelled.")
					return err
				}
			}

			c, err := newAccountAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			account, err := c.ReactivateAccount(accountID)
			if err != nil {
				return err
			}

			columns := accountDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, account)
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

func newAccountsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <account_id>",
		Short: "Get account details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAccountAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			account, err := c.GetAccount(args[0])
			if err != nil {
				return err
			}

			columns := accountDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, account)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}
