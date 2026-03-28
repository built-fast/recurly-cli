package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

// mockAccountInvoiceLister implements recurly.InvoiceLister for testing.
type mockAccountInvoiceLister struct {
	invoices []recurly.Invoice
	fetched  bool
}

func (m *mockAccountInvoiceLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockAccountInvoiceLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockAccountInvoiceLister) Count() (*int64, error) {
	n := int64(len(m.invoices))
	return &n, nil
}

func (m *mockAccountInvoiceLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockAccountInvoiceLister) Data() []recurly.Invoice {
	return m.invoices
}

func (m *mockAccountInvoiceLister) HasMore() bool {
	return !m.fetched
}

func (m *mockAccountInvoiceLister) Next() string {
	return ""
}

func sampleAccountInvoice() recurly.Invoice {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return recurly.Invoice{
		Id:        "inv-123",
		Number:    "1001",
		State:     "paid",
		Type:      "charge",
		Currency:  "USD",
		Total:     49.99,
		Account:   recurly.AccountMini{Code: "acct-456"},
		CreatedAt: &now,
	}
}

// --- accounts invoices list ---

func TestAccountInvoices_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "invoices") {
		t.Error("expected accounts help to show 'invoices' subcommand")
	}
}

func TestAccountInvoicesList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("accounts", "invoices", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected invoices help to show 'list' subcommand")
	}
}

func TestAccountInvoicesListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand("accounts", "invoices", "list", "--help")
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
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "invoices", "list")
	if err == nil {
		t.Fatal("expected error when account_id is missing")
	}
}

func TestAccountInvoicesList_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand("accounts", "invoices", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountInvoicesList_PassesAccountID(t *testing.T) {
	var capturedAccountID string

	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			capturedAccountID = accountId
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "invoices", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAccountID != "acct-456" {
		t.Errorf("expected accountId=acct-456, got %q", capturedAccountID)
	}
}

func TestAccountInvoicesList_PaginationParams(t *testing.T) {
	var capturedParams *recurly.ListAccountInvoicesParams

	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			capturedParams = params
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "invoices", "list", "acct-456", "--limit", "50", "--order", "desc", "--sort", "updated_at")
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
	var capturedParams *recurly.ListAccountInvoicesParams

	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			capturedParams = params
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "invoices", "list", "acct-456", "--state", "paid", "--type", "charge")
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
	var capturedParams *recurly.ListAccountInvoicesParams

	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			capturedParams = params
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "invoices", "list", "acct-456")
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
	inv := sampleAccountInvoice()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{inv}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("accounts", "invoices", "list", "acct-456")
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
	inv := sampleAccountInvoice()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{inv}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "invoices", "list", "acct-456")
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
	inv := sampleAccountInvoice()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{inv}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json-pretty")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "invoices", "list", "acct-456")
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
	inv := sampleAccountInvoice()
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{inv}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	viper.Set("jq", ".data[0].number")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "invoices", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "1001") {
		t.Errorf("expected jq output to contain '1001', got: %s", out)
	}
}

func TestAccountInvoicesList_SDKError(t *testing.T) {
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "invoices", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestAccountInvoicesList_EmptyResults(t *testing.T) {
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("accounts", "invoices", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ID (Number)") {
		t.Error("expected table headers even for empty results")
	}
}

func TestAccountInvoicesList_EmptyResults_JSON(t *testing.T) {
	mock := &mockAccountNestedAPI{
		listAccountInvoicesFn: func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
			return &mockAccountInvoiceLister{invoices: []recurly.Invoice{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "invoices", "list", "acct-456")
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
