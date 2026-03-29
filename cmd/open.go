package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// validResourceTypes lists all supported resource types for the open command.
var validResourceTypes = []string{"accounts", "plans", "subscriptions", "invoices", "transactions", "items", "coupons"}

// resourceNeedsUUID returns true for resource types whose dashboard URL requires a UUID
// rather than a code/number.
func resourceNeedsUUID(resource string) bool {
	return resource == "subscriptions" || resource == "transactions"
}

// openBrowserFunc is the function used to open a URL in the browser.
// Tests override this to capture the URL without launching a browser.
var openBrowserFunc = openBrowser

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Run()
}

// buildDashboardURL constructs the Recurly dashboard URL for the given site, resource, and identifier.
func buildDashboardURL(site, resource, identifier string) string {
	base := fmt.Sprintf("https://%s.recurly.com", site)
	if resource == "" {
		return base
	}
	if identifier == "" {
		return base + "/" + resource
	}
	return base + "/" + resource + "/" + identifier
}

// resolveIdentifier returns the identifier to use in the dashboard URL.
// For subscriptions and transactions, it fetches the resource to get the UUID.
// For other resource types, it returns the identifier as-is.
func resolveIdentifier(cmd *cobra.Command, resource, id string) (string, error) {
	if !resourceNeedsUUID(resource) {
		return id, nil
	}

	switch resource {
	case "subscriptions":
		c, err := newSubscriptionAPI()
		if err != nil {
			return "", err
		}
		sub, err := c.GetSubscription(id)
		if err != nil {
			return "", fmt.Errorf("fetching subscription %s: %w", id, err)
		}
		return sub.Uuid, nil
	case "transactions":
		c, err := newTransactionAPI()
		if err != nil {
			return "", err
		}
		txn, err := c.GetTransaction(id)
		if err != nil {
			return "", fmt.Errorf("fetching transaction %s: %w", id, err)
		}
		return txn.Uuid, nil
	}
	return id, nil
}

func newOpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [resource] [identifier]",
		Short: "Open a Recurly resource in the dashboard",
		Long: `Open a Recurly resource in the web dashboard.

With no arguments, opens the dashboard home page.
With a resource type, opens the resource list page.
With a resource type and identifier, opens the specific resource page.

Supported resource types: accounts, plans, subscriptions, invoices, transactions, items, coupons

For subscriptions and transactions, the CLI accepts the standard resource ID,
fetches the resource to retrieve the UUID, then constructs the dashboard URL.`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			site, _ := cmd.Flags().GetString("site")
			if site == "" {
				site = viper.GetString("site")
			}
			if site == "" {
				return fmt.Errorf("site subdomain is required: use --site flag or run 'recurly configure' to set it")
			}

			urlOnly, _ := cmd.Flags().GetBool("url")

			var resource, identifier string
			if len(args) >= 1 {
				resource = args[0]
				if !isValidResource(resource) {
					return fmt.Errorf("unrecognized resource type %q\nValid types: %s", resource, strings.Join(validResourceTypes, ", "))
				}
			}
			if len(args) >= 2 {
				var err error
				identifier, err = resolveIdentifier(cmd, resource, args[1])
				if err != nil {
					return err
				}
			}

			url := buildDashboardURL(site, resource, identifier)

			if urlOnly {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), url)
				return err
			}

			return openBrowserFunc(url)
		},
	}

	cmd.Flags().String("site", "", "Site subdomain (overrides configured site)")
	cmd.Flags().Bool("url", false, "Print URL instead of opening browser")

	return cmd
}

func isValidResource(resource string) bool {
	for _, r := range validResourceTypes {
		if r == resource {
			return true
		}
	}
	return false
}
