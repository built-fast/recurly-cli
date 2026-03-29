package cmd

import (
	"fmt"
	"strings"
	"testing"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

func TestBuildDashboardURL_NoResource(t *testing.T) {
	url := buildDashboardURL("mysite", "", "")
	expected := "https://mysite.recurly.com"
	if url != expected {
		t.Errorf("expected %q, got %q", expected, url)
	}
}

func TestBuildDashboardURL_ResourceOnly(t *testing.T) {
	tests := []struct {
		resource string
		expected string
	}{
		{"accounts", "https://mysite.recurly.com/accounts"},
		{"plans", "https://mysite.recurly.com/plans"},
		{"subscriptions", "https://mysite.recurly.com/subscriptions"},
		{"invoices", "https://mysite.recurly.com/invoices"},
		{"transactions", "https://mysite.recurly.com/transactions"},
		{"items", "https://mysite.recurly.com/items"},
		{"coupons", "https://mysite.recurly.com/coupons"},
	}
	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			url := buildDashboardURL("mysite", tt.resource, "")
			if url != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, url)
			}
		})
	}
}

func TestBuildDashboardURL_ResourceWithIdentifier(t *testing.T) {
	tests := []struct {
		resource   string
		identifier string
		expected   string
	}{
		{"accounts", "acct123", "https://mysite.recurly.com/accounts/acct123"},
		{"plans", "gold", "https://mysite.recurly.com/plans/gold"},
		{"subscriptions", "uuid-abc", "https://mysite.recurly.com/subscriptions/uuid-abc"},
		{"invoices", "1001", "https://mysite.recurly.com/invoices/1001"},
		{"transactions", "uuid-xyz", "https://mysite.recurly.com/transactions/uuid-xyz"},
		{"items", "item-code", "https://mysite.recurly.com/items/item-code"},
		{"coupons", "SAVE10", "https://mysite.recurly.com/coupons/SAVE10"},
	}
	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			url := buildDashboardURL("mysite", tt.resource, tt.identifier)
			if url != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, url)
			}
		})
	}
}

func TestOpenCmd_NoArgs_PrintsURL(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	out, _, err := executeCommand("open", "--url")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://testsite.recurly.com\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestOpenCmd_ResourceOnly_PrintsURL(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	out, _, err := executeCommand("open", "--url", "accounts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://testsite.recurly.com/accounts\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestOpenCmd_AccountWithCode_PrintsURL(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	out, _, err := executeCommand("open", "--url", "accounts", "acct123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://testsite.recurly.com/accounts/acct123\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestOpenCmd_PlanWithCode_PrintsURL(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	out, _, err := executeCommand("open", "--url", "plans", "gold")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://testsite.recurly.com/plans/gold\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestOpenCmd_ItemWithCode_PrintsURL(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	out, _, err := executeCommand("open", "--url", "items", "widget")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://testsite.recurly.com/items/widget\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestOpenCmd_CouponWithCode_PrintsURL(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	out, _, err := executeCommand("open", "--url", "coupons", "SAVE10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://testsite.recurly.com/coupons/SAVE10\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestOpenCmd_InvoiceWithNumber_PrintsURL(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	out, _, err := executeCommand("open", "--url", "invoices", "1001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://testsite.recurly.com/invoices/1001\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestOpenCmd_SubscriptionFetchesUUID(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	orig := newSubscriptionAPI
	newSubscriptionAPI = func() (SubscriptionAPI, error) {
		return &mockSubscriptionAPI{
			getSubscriptionFn: func(id string, opts ...recurly.Option) (*recurly.Subscription, error) {
				return &recurly.Subscription{Uuid: "sub-uuid-123"}, nil
			},
		}, nil
	}
	defer func() { newSubscriptionAPI = orig }()

	out, _, err := executeCommand("open", "--url", "subscriptions", "sxyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://testsite.recurly.com/subscriptions/sub-uuid-123\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestOpenCmd_TransactionFetchesUUID(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	orig := newTransactionAPI
	newTransactionAPI = func() (TransactionAPI, error) {
		return &mockTransactionAPI{
			getTransactionFn: func(id string, opts ...recurly.Option) (*recurly.Transaction, error) {
				return &recurly.Transaction{Uuid: "txn-uuid-456"}, nil
			},
		}, nil
	}
	defer func() { newTransactionAPI = orig }()

	out, _, err := executeCommand("open", "--url", "transactions", "tabc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://testsite.recurly.com/transactions/txn-uuid-456\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestOpenCmd_SiteFlagOverridesConfig(t *testing.T) {
	viper.Set("site", "configsite")
	defer viper.Set("site", "")

	out, _, err := executeCommand("open", "--url", "--site", "override")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://override.recurly.com\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestOpenCmd_NoSiteConfigured_ReturnsError(t *testing.T) {
	viper.Set("site", "")

	_, stderr, err := executeCommand("open", "--url")
	if err == nil {
		t.Fatal("expected error when no site configured")
	}
	if !strings.Contains(stderr, "site subdomain is required") {
		t.Errorf("expected error about site subdomain, got %q", stderr)
	}
}

func TestOpenCmd_InvalidResource_ReturnsError(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	_, stderr, err := executeCommand("open", "--url", "widgets")
	if err == nil {
		t.Fatal("expected error for invalid resource type")
	}
	if !strings.Contains(stderr, "unrecognized resource type") {
		t.Errorf("expected 'unrecognized resource type' error, got %q", stderr)
	}
	for _, rt := range validResourceTypes {
		if !strings.Contains(stderr, rt) {
			t.Errorf("expected error to list valid type %q, got %q", rt, stderr)
		}
	}
}

func TestOpenCmd_TooManyArgs_ReturnsError(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	_, _, err := executeCommand("open", "accounts", "code", "extra")
	if err == nil {
		t.Fatal("expected error for too many args")
	}
}

func TestOpenCmd_OpenssBrowser(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	var openedURL string
	orig := openBrowserFunc
	openBrowserFunc = func(url string) error {
		openedURL = url
		return nil
	}
	defer func() { openBrowserFunc = orig }()

	_, _, err := executeCommand("open", "accounts", "acct123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://testsite.recurly.com/accounts/acct123"
	if openedURL != expected {
		t.Errorf("expected browser to open %q, got %q", expected, openedURL)
	}
}

func TestOpenCmd_BrowserError_ReturnsError(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	orig := openBrowserFunc
	openBrowserFunc = func(url string) error {
		return fmt.Errorf("browser not found")
	}
	defer func() { openBrowserFunc = orig }()

	_, _, err := executeCommand("open")
	if err == nil {
		t.Fatal("expected error when browser fails")
	}
	if !strings.Contains(err.Error(), "browser not found") {
		t.Errorf("expected 'browser not found' error, got %v", err)
	}
}

func TestOpenCmd_SubscriptionFetchError_ReturnsError(t *testing.T) {
	viper.Set("site", "testsite")
	defer viper.Set("site", "")

	orig := newSubscriptionAPI
	newSubscriptionAPI = func() (SubscriptionAPI, error) {
		return &mockSubscriptionAPI{
			getSubscriptionFn: func(id string, opts ...recurly.Option) (*recurly.Subscription, error) {
				return nil, fmt.Errorf("not found")
			},
		}, nil
	}
	defer func() { newSubscriptionAPI = orig }()

	_, stderr, err := executeCommand("open", "--url", "subscriptions", "bad-id")
	if err == nil {
		t.Fatal("expected error when subscription fetch fails")
	}
	if !strings.Contains(stderr, "fetching subscription") {
		t.Errorf("expected error about fetching subscription, got %q", stderr)
	}
}

func TestOpenCmd_Help_ShowsUsage(t *testing.T) {
	out, _, err := executeCommand("open", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"--site", "--url", "accounts", "plans", "subscriptions"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected help to contain %q", want)
		}
	}
}

func TestResourceNeedsUUID(t *testing.T) {
	if !resourceNeedsUUID("subscriptions") {
		t.Error("expected subscriptions to need UUID")
	}
	if !resourceNeedsUUID("transactions") {
		t.Error("expected transactions to need UUID")
	}
	for _, r := range []string{"accounts", "plans", "items", "coupons", "invoices"} {
		if resourceNeedsUUID(r) {
			t.Errorf("expected %s to not need UUID", r)
		}
	}
}

func TestIsValidResource(t *testing.T) {
	for _, r := range validResourceTypes {
		if !isValidResource(r) {
			t.Errorf("expected %q to be valid", r)
		}
	}
	if isValidResource("widgets") {
		t.Error("expected 'widgets' to be invalid")
	}
}
