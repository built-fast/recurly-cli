package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

// mockAccountNestedAPI implements AccountNestedAPI for testing.
type mockAccountNestedAPI struct {
	listAccountSubscriptionsFn func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error)
	listAccountInvoicesFn      func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error)
	listAccountTransactionsFn  func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error)
}

func (m *mockAccountNestedAPI) ListAccountSubscriptions(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
	if m.listAccountSubscriptionsFn != nil {
		return m.listAccountSubscriptionsFn(accountId, params, opts...)
	}
	return nil, nil
}

func (m *mockAccountNestedAPI) ListAccountInvoices(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
	if m.listAccountInvoicesFn != nil {
		return m.listAccountInvoicesFn(accountId, params, opts...)
	}
	return nil, nil
}

func (m *mockAccountNestedAPI) ListAccountTransactions(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
	if m.listAccountTransactionsFn != nil {
		return m.listAccountTransactionsFn(accountId, params, opts...)
	}
	return nil, nil
}

// --- accounts subscriptions list ---

func TestAccountSubscriptions_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "subscriptions") {
		t.Error("expected accounts help to show 'subscriptions' subcommand")
	}
}

func TestAccountSubscriptionsList_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected subscriptions help to show 'list' subcommand")
	}
}

func TestAccountSubscriptionsListHelp_ShowsFlags(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "subscriptions", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--order", "--sort", "--state"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestAccountSubscriptionsList_RequiresAccountID(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "subscriptions", "list")
	if err == nil {
		t.Fatal("expected error when account_id is missing")
	}
}

func TestAccountSubscriptionsList_NoAPIKey_ReturnsError(t *testing.T) {
	// Cannot use t.Parallel() — uses t.Setenv which is incompatible
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "subscriptions", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountSubscriptionsList_PassesAccountID(t *testing.T) {
	t.Parallel()
	var capturedAccountID string

	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedAccountID = accountId
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAccountID != "acct-456" {
		t.Errorf("expected accountId=acct-456, got %q", capturedAccountID)
	}
}

func TestAccountSubscriptionsList_PaginationParams(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountSubscriptionsParams

	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedParams = params
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456", "--limit", "50", "--order", "desc", "--sort", "updated_at")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams == nil {
		t.Fatal("expected params to be captured")
	}
	if capturedParams.Limit == nil || *capturedParams.Limit != 50 {
		t.Errorf("expected limit=50, got %v", capturedParams.Limit)
	}
	if capturedParams.Order == nil || *capturedParams.Order != "desc" {
		t.Errorf("expected order=desc, got %v", capturedParams.Order)
	}
	if capturedParams.Sort == nil || *capturedParams.Sort != "updated_at" {
		t.Errorf("expected sort=updated_at, got %v", capturedParams.Sort)
	}
}

func TestAccountSubscriptionsList_FilterParams(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountSubscriptionsParams

	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedParams = params
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456", "--state", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.State == nil || *capturedParams.State != "active" {
		t.Errorf("expected state=active, got %v", capturedParams.State)
	}
}

func TestAccountSubscriptionsList_UnsetFlagsNotSent(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountSubscriptionsParams

	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedParams = params
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.Limit != nil {
		t.Error("expected Limit to be nil when not set")
	}
	if capturedParams.Order != nil {
		t.Error("expected Order to be nil when not set")
	}
	if capturedParams.Sort != nil {
		t.Error("expected Sort to be nil when not set")
	}
	if capturedParams.State != nil {
		t.Error("expected State to be nil when not set")
	}
}

func TestAccountSubscriptionsList_TableOutput(t *testing.T) {
	t.Parallel()
	sub := sampleSubscription()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{sub}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"sub-123", "acct-456", "gold", "active", "USD", "19.99"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	for _, header := range []string{"ID", "Account Code", "Plan Code", "State", "Currency", "Unit Amount", "Current Period Ends At"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestAccountSubscriptionsList_JSONOutput(t *testing.T) {
	t.Parallel()
	sub := sampleSubscription()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{sub}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Data    []json.RawMessage `json:"data"`
		HasMore bool              `json:"has_more"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\noutput: %s", err, out)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 item in data, got %d", len(result.Data))
	}
}

func TestAccountSubscriptionsList_JSONPrettyOutput(t *testing.T) {
	t.Parallel()
	sub := sampleSubscription()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{sub}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456", "--output", "json-pretty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "sub-123") {
		t.Error("expected JSON-pretty output to contain subscription ID")
	}
	if !strings.Contains(out, "\n") {
		t.Error("expected JSON-pretty output to be indented with newlines")
	}
}

func TestAccountSubscriptionsList_JQFilter(t *testing.T) {
	t.Parallel()
	sub := sampleSubscription()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{sub}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456", "--output", "json", "--jq", ".data[0].id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected jq output to contain 'sub-123', got: %s", out)
	}
}

func TestAccountSubscriptionsList_SDKError(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestAccountSubscriptionsList_EmptyResults(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ID") {
		t.Error("expected table headers even for empty results")
	}
}

func TestAccountSubscriptionsList_EmptyResults_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "subscriptions", "list", "acct-456", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Data    []json.RawMessage `json:"data"`
		HasMore bool              `json:"has_more"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 items in data, got %d", len(result.Data))
	}
}

// --- accounts invoices list ---

func TestAccountInvoices_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "invoices") {
		t.Error("expected accounts help to show 'invoices' subcommand")
	}
}

func TestAccountInvoicesList_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "invoices", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected invoices help to show 'list' subcommand")
	}
}

func TestAccountInvoicesListHelp_ShowsFlags(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "invoices", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--order", "--sort", "--state", "--type"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestAccountInvoicesList_RequiresAccountID(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "invoices", "list")
	if err == nil {
		t.Fatal("expected error when account_id is missing")
	}
}

func TestAccountInvoicesList_NoAPIKey_ReturnsError(t *testing.T) {
	// Cannot use t.Parallel() — uses t.Setenv which is incompatible
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "invoices", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountInvoicesList_PassesAccountID(t *testing.T) {
	t.Parallel()
	var capturedAccountID string

	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			capturedAccountID = accountId
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAccountID != "acct-456" {
		t.Errorf("expected accountId=acct-456, got %q", capturedAccountID)
	}
}

func TestAccountInvoicesList_PaginationParams(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountInvoicesParams

	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			capturedParams = params
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456", "--limit", "50", "--order", "desc", "--sort", "updated_at")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams == nil {
		t.Fatal("expected params to be captured")
	}
	if capturedParams.Limit == nil || *capturedParams.Limit != 50 {
		t.Errorf("expected limit=50, got %v", capturedParams.Limit)
	}
	if capturedParams.Order == nil || *capturedParams.Order != "desc" {
		t.Errorf("expected order=desc, got %v", capturedParams.Order)
	}
	if capturedParams.Sort == nil || *capturedParams.Sort != "updated_at" {
		t.Errorf("expected sort=updated_at, got %v", capturedParams.Sort)
	}
}

func TestAccountInvoicesList_FilterParams(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountInvoicesParams

	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			capturedParams = params
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456", "--state", "paid", "--type", "charge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.State == nil || *capturedParams.State != "paid" {
		t.Errorf("expected state=paid, got %v", capturedParams.State)
	}
	if capturedParams.Type == nil || *capturedParams.Type != "charge" {
		t.Errorf("expected type=charge, got %v", capturedParams.Type)
	}
}

func TestAccountInvoicesList_UnsetFlagsNotSent(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountInvoicesParams

	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			capturedParams = params
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.Limit != nil {
		t.Error("expected Limit to be nil when not set")
	}
	if capturedParams.Order != nil {
		t.Error("expected Order to be nil when not set")
	}
	if capturedParams.Sort != nil {
		t.Error("expected Sort to be nil when not set")
	}
	if capturedParams.State != nil {
		t.Error("expected State to be nil when not set")
	}
	if capturedParams.Type != nil {
		t.Error("expected Type to be nil when not set")
	}
}

func TestAccountInvoicesList_TableOutput(t *testing.T) {
	t.Parallel()
	inv := sampleAccountInvoice()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{inv}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"1001", "paid", "charge", "49.99", "USD"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	for _, header := range []string{"ID (Number)", "State", "Type", "Total", "Currency", "Created At"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestAccountInvoicesList_JSONOutput(t *testing.T) {
	t.Parallel()
	inv := sampleAccountInvoice()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{inv}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Data    []json.RawMessage `json:"data"`
		HasMore bool              `json:"has_more"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\noutput: %s", err, out)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 item in data, got %d", len(result.Data))
	}
}

func TestAccountInvoicesList_JSONPrettyOutput(t *testing.T) {
	t.Parallel()
	inv := sampleAccountInvoice()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{inv}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456", "--output", "json-pretty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "1001") {
		t.Error("expected JSON-pretty output to contain invoice number")
	}
	if !strings.Contains(out, "\n") {
		t.Error("expected JSON-pretty output to be indented with newlines")
	}
}

func TestAccountInvoicesList_JQFilter(t *testing.T) {
	t.Parallel()
	inv := sampleAccountInvoice()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{inv}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456", "--output", "json", "--jq", ".data[0].number")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "1001") {
		t.Errorf("expected jq output to contain '1001', got: %s", out)
	}
}

func TestAccountInvoicesList_SDKError(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestAccountInvoicesList_EmptyResults(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ID (Number)") {
		t.Error("expected table headers even for empty results")
	}
}

func TestAccountInvoicesList_EmptyResults_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockLister[recurly.Invoice]{items: []recurly.Invoice{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "invoices", "list", "acct-456", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Data    []json.RawMessage `json:"data"`
		HasMore bool              `json:"has_more"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 items in data, got %d", len(result.Data))
	}
}

// --- accounts transactions list ---

func TestAccountTransactions_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "transactions") {
		t.Error("expected accounts help to show 'transactions' subcommand")
	}
}

func TestAccountTransactionsList_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "transactions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected transactions help to show 'list' subcommand")
	}
}

func TestAccountTransactionsListHelp_ShowsFlags(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "transactions", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--order", "--sort", "--type", "--success"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestAccountTransactionsList_RequiresAccountID(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "transactions", "list")
	if err == nil {
		t.Fatal("expected error when account_id is missing")
	}
}

func TestAccountTransactionsList_NoAPIKey_ReturnsError(t *testing.T) {
	// Cannot use t.Parallel() — uses t.Setenv which is incompatible
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "transactions", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountTransactionsList_PassesAccountID(t *testing.T) {
	t.Parallel()
	var capturedAccountID string

	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			capturedAccountID = accountId
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAccountID != "acct-456" {
		t.Errorf("expected accountId=acct-456, got %q", capturedAccountID)
	}
}

func TestAccountTransactionsList_PaginationParams(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountTransactionsParams

	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			capturedParams = params
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456", "--limit", "50", "--order", "desc", "--sort", "updated_at")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams == nil {
		t.Fatal("expected params to be captured")
	}
	if capturedParams.Limit == nil || *capturedParams.Limit != 50 {
		t.Errorf("expected limit=50, got %v", capturedParams.Limit)
	}
	if capturedParams.Order == nil || *capturedParams.Order != "desc" {
		t.Errorf("expected order=desc, got %v", capturedParams.Order)
	}
	if capturedParams.Sort == nil || *capturedParams.Sort != "updated_at" {
		t.Errorf("expected sort=updated_at, got %v", capturedParams.Sort)
	}
}

func TestAccountTransactionsList_FilterParams(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountTransactionsParams

	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			capturedParams = params
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456", "--type", "payment", "--success", "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.Type == nil || *capturedParams.Type != "payment" {
		t.Errorf("expected type=payment, got %v", capturedParams.Type)
	}
	if capturedParams.Success == nil || *capturedParams.Success != "true" {
		t.Errorf("expected success=true, got %v", capturedParams.Success)
	}
}

func TestAccountTransactionsList_UnsetFlagsNotSent(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountTransactionsParams

	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			capturedParams = params
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.Limit != nil {
		t.Error("expected Limit to be nil when not set")
	}
	if capturedParams.Order != nil {
		t.Error("expected Order to be nil when not set")
	}
	if capturedParams.Sort != nil {
		t.Error("expected Sort to be nil when not set")
	}
	if capturedParams.Type != nil {
		t.Error("expected Type to be nil when not set")
	}
	if capturedParams.Success != nil {
		t.Error("expected Success to be nil when not set")
	}
}

func TestAccountTransactionsList_TableOutput(t *testing.T) {
	t.Parallel()
	txn := sampleAccountTransaction()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{txn}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"txn-123", "purchase", "29.99", "USD", "success", "true"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	for _, header := range []string{"ID", "Type", "Amount", "Currency", "Status", "Success", "Created At"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestAccountTransactionsList_JSONOutput(t *testing.T) {
	t.Parallel()
	txn := sampleAccountTransaction()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{txn}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Data    []json.RawMessage `json:"data"`
		HasMore bool              `json:"has_more"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\noutput: %s", err, out)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 item in data, got %d", len(result.Data))
	}
}

func TestAccountTransactionsList_JSONPrettyOutput(t *testing.T) {
	t.Parallel()
	txn := sampleAccountTransaction()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{txn}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456", "--output", "json-pretty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "txn-123") {
		t.Error("expected JSON-pretty output to contain transaction ID")
	}
	if !strings.Contains(out, "\n") {
		t.Error("expected JSON-pretty output to be indented with newlines")
	}
}

func TestAccountTransactionsList_JQFilter(t *testing.T) {
	t.Parallel()
	txn := sampleAccountTransaction()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{txn}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456", "--output", "json", "--jq", ".data[0].id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "txn-123") {
		t.Errorf("expected jq output to contain 'txn-123', got: %s", out)
	}
}

func TestAccountTransactionsList_SDKError(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	app := newTestAccountNestedApp(mock)

	_, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestAccountTransactionsList_EmptyResults(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ID") {
		t.Error("expected table headers even for empty results")
	}
}

func TestAccountTransactionsList_EmptyResults_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockLister[recurly.Transaction]{items: []recurly.Transaction{}}, nil
		},
	}
	app := newTestAccountNestedApp(mock)

	out, _, err := executeCommand(app, "accounts", "transactions", "list", "acct-456", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Data    []json.RawMessage `json:"data"`
		HasMore bool              `json:"has_more"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 items in data, got %d", len(result.Data))
	}
}
