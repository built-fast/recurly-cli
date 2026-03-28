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

// mockAccountTransactionLister implements recurly.TransactionLister for testing.
type mockAccountTransactionLister struct {
	transactions []recurly.Transaction
	fetched      bool
}

func (m *mockAccountTransactionLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockAccountTransactionLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockAccountTransactionLister) Count() (*int64, error) {
	n := int64(len(m.transactions))
	return &n, nil
}

func (m *mockAccountTransactionLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockAccountTransactionLister) Data() []recurly.Transaction {
	return m.transactions
}

func (m *mockAccountTransactionLister) HasMore() bool {
	return !m.fetched
}

func (m *mockAccountTransactionLister) Next() string {
	return ""
}

func sampleAccountTransaction() recurly.Transaction {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return recurly.Transaction{
		Id:        "txn-123",
		Type:      "purchase",
		Amount:    29.99,
		Currency:  "USD",
		Status:    "success",
		Success:   true,
		Account:   recurly.AccountMini{Code: "acct-456"},
		CreatedAt: &now,
	}
}

// --- accounts transactions list ---

func TestAccountTransactions_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "transactions") {
		t.Error("expected accounts help to show 'transactions' subcommand")
	}
}

func TestAccountTransactionsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("accounts", "transactions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected transactions help to show 'list' subcommand")
	}
}

func TestAccountTransactionsListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand("accounts", "transactions", "list", "--help")
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
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "transactions", "list")
	if err == nil {
		t.Fatal("expected error when account_id is missing")
	}
}

func TestAccountTransactionsList_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand("accounts", "transactions", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountTransactionsList_PassesAccountID(t *testing.T) {
	var capturedAccountID string

	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			capturedAccountID = accountId
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "transactions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAccountID != "acct-456" {
		t.Errorf("expected accountId=acct-456, got %q", capturedAccountID)
	}
}

func TestAccountTransactionsList_PaginationParams(t *testing.T) {
	var capturedParams *recurly.ListAccountTransactionsParams

	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			capturedParams = params
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "transactions", "list", "acct-456", "--limit", "50", "--order", "desc", "--sort", "updated_at")
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
	var capturedParams *recurly.ListAccountTransactionsParams

	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			capturedParams = params
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "transactions", "list", "acct-456", "--type", "payment", "--success", "true")
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
	var capturedParams *recurly.ListAccountTransactionsParams

	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			capturedParams = params
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "transactions", "list", "acct-456")
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
	txn := sampleAccountTransaction()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{txn}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("accounts", "transactions", "list", "acct-456")
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
	txn := sampleAccountTransaction()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{txn}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "transactions", "list", "acct-456")
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
	txn := sampleAccountTransaction()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{txn}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json-pretty")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "transactions", "list", "acct-456")
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
	txn := sampleAccountTransaction()
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{txn}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	viper.Set("jq", ".data[0].id")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "transactions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "txn-123") {
		t.Errorf("expected jq output to contain 'txn-123', got: %s", out)
	}
}

func TestAccountTransactionsList_SDKError(t *testing.T) {
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "transactions", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestAccountTransactionsList_EmptyResults(t *testing.T) {
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("accounts", "transactions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ID") {
		t.Error("expected table headers even for empty results")
	}
}

func TestAccountTransactionsList_EmptyResults_JSON(t *testing.T) {
	mock := &mockAccountNestedAPI{
		listAccountTransactionsFn: func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockAccountTransactionLister{transactions: []recurly.Transaction{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "transactions", "list", "acct-456")
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
