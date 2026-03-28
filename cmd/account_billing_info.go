package cmd

import (
	"fmt"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newAccountBillingInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing-info",
		Short: "Manage account billing info",
	}
	cmd.AddCommand(newAccountBillingInfoGetCmd())
	return cmd
}

func billingInfoDetailColumns() []output.Column {
	return []output.Column{
		{Header: "ID", Extract: func(v any) string { return v.(*recurly.BillingInfo).Id }},
		{Header: "Account ID", Extract: func(v any) string { return v.(*recurly.BillingInfo).AccountId }},
		{Header: "First Name", Extract: func(v any) string { return v.(*recurly.BillingInfo).FirstName }},
		{Header: "Last Name", Extract: func(v any) string { return v.(*recurly.BillingInfo).LastName }},
		{Header: "Company", Extract: func(v any) string { return v.(*recurly.BillingInfo).Company }},
		{Header: "Valid", Extract: func(v any) string {
			return fmt.Sprintf("%t", v.(*recurly.BillingInfo).Valid)
		}},
		{Header: "Payment Method", Extract: func(v any) string {
			return v.(*recurly.BillingInfo).PaymentMethod.CardType
		}},
		{Header: "Primary Payment Method", Extract: func(v any) string {
			return fmt.Sprintf("%t", v.(*recurly.BillingInfo).PrimaryPaymentMethod)
		}},
		{Header: "Created At", Extract: func(v any) string {
			b := v.(*recurly.BillingInfo)
			if b.CreatedAt != nil {
				return b.CreatedAt.Format(time.RFC3339)
			}
			return ""
		}},
		{Header: "Updated At", Extract: func(v any) string {
			b := v.(*recurly.BillingInfo)
			if b.UpdatedAt != nil {
				return b.UpdatedAt.Format(time.RFC3339)
			}
			return ""
		}},
	}
}

func newAccountBillingInfoGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <account_id>",
		Short: "Get billing info for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAccountBillingInfoAPI()
			if err != nil {
				return err
			}

			format := viper.GetString("output")

			billingInfo, err := c.GetBillingInfo(args[0])
			if err != nil {
				return err
			}

			columns := billingInfoDetailColumns()

			formatted, err := output.FormatOne(format, columns, billingInfo)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}
