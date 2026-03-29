package cmd

import (
	"fmt"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
)

func newAccountBillingInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing-info",
		Short: "Manage account billing info",
	}
	cmd.AddCommand(withWatch(newAccountBillingInfoGetCmd()))
	cmd.AddCommand(withFromFile(newAccountBillingInfoUpdateCmd()))
	cmd.AddCommand(newAccountBillingInfoRemoveCmd())
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

			cfg := output.FromContext(cmd.Context())

			billingInfo, err := c.GetBillingInfo(args[0])
			if err != nil {
				return err
			}

			columns := billingInfoDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, billingInfo)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newAccountBillingInfoUpdateCmd() *cobra.Command {
	var (
		firstName            string
		lastName             string
		company              string
		vatNumber            string
		tokenId              string
		currency             string
		primaryPaymentMethod bool
		backupPaymentMethod  bool

		// Address fields
		addressStreet1    string
		addressStreet2    string
		addressCity       string
		addressRegion     string
		addressPostalCode string
		addressCountry    string
	)

	cmd := &cobra.Command{
		Use:   "update <account_id>",
		Short: "Update billing info for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newAccountBillingInfoAPI()
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())
			body := &recurly.BillingInfoCreate{}

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
			if cmd.Flags().Changed("token-id") {
				body.TokenId = recurly.String(tokenId)
			}
			if cmd.Flags().Changed("currency") {
				body.Currency = recurly.String(currency)
			}
			if cmd.Flags().Changed("primary-payment-method") {
				body.PrimaryPaymentMethod = recurly.Bool(primaryPaymentMethod)
			}
			if cmd.Flags().Changed("backup-payment-method") {
				body.BackupPaymentMethod = recurly.Bool(backupPaymentMethod)
			}

			// Address fields
			addressChanged := false
			addr := &recurly.AddressCreate{}
			if cmd.Flags().Changed("address-street1") {
				addr.Street1 = recurly.String(addressStreet1)
				addressChanged = true
			}
			if cmd.Flags().Changed("address-street2") {
				addr.Street2 = recurly.String(addressStreet2)
				addressChanged = true
			}
			if cmd.Flags().Changed("address-city") {
				addr.City = recurly.String(addressCity)
				addressChanged = true
			}
			if cmd.Flags().Changed("address-region") {
				addr.Region = recurly.String(addressRegion)
				addressChanged = true
			}
			if cmd.Flags().Changed("address-postal-code") {
				addr.PostalCode = recurly.String(addressPostalCode)
				addressChanged = true
			}
			if cmd.Flags().Changed("address-country") {
				addr.Country = recurly.String(addressCountry)
				addressChanged = true
			}
			if addressChanged {
				body.Address = addr
			}

			billingInfo, err := c.UpdateBillingInfo(args[0], body)
			if err != nil {
				return err
			}

			columns := billingInfoDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, billingInfo)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&firstName, "first-name", "", "First name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name")
	cmd.Flags().StringVar(&company, "company", "", "Company name")
	cmd.Flags().StringVar(&vatNumber, "vat-number", "", "VAT number")
	cmd.Flags().StringVar(&tokenId, "token-id", "", "Token ID from Recurly.js")
	cmd.Flags().StringVar(&currency, "currency", "", "3-letter ISO 4217 currency code")
	cmd.Flags().BoolVar(&primaryPaymentMethod, "primary-payment-method", false, "Designate as primary payment method")
	cmd.Flags().BoolVar(&backupPaymentMethod, "backup-payment-method", false, "Designate as backup payment method")

	// Address flags
	cmd.Flags().StringVar(&addressStreet1, "address-street1", "", "Street address line 1")
	cmd.Flags().StringVar(&addressStreet2, "address-street2", "", "Street address line 2")
	cmd.Flags().StringVar(&addressCity, "address-city", "", "City")
	cmd.Flags().StringVar(&addressRegion, "address-region", "", "State or province")
	cmd.Flags().StringVar(&addressPostalCode, "address-postal-code", "", "Zip or postal code")
	cmd.Flags().StringVar(&addressCountry, "address-country", "", "Country (2-letter ISO 3166-1 alpha-2 code)")

	return cmd
}

func newAccountBillingInfoRemoveCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "remove <account_id>",
		Short: "Remove billing info from an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Remove billing info from account %s? [y/N] ", accountID))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Removal cancelled.")
					return err
				}
			}

			c, err := newAccountBillingInfoAPI()
			if err != nil {
				return err
			}

			_, err = c.RemoveBillingInfo(accountID)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Billing info removed from account %s\n", accountID)
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}
