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

// mockTransactionAPI implements TransactionAPI for testing.
type mockTransactionAPI struct {
	listTransactionsFn func(params *recurly.ListTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error)
	getTransactionFn   func(transactionId string, opts ...recurly.Option) (*recurly.Transaction, error)
}

func (m *mockTransactionAPI) ListTransactions(params *recurly.ListTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
	return m.listTransactionsFn(params, opts...)
}

func (m *mockTransactionAPI) GetTransaction(transactionId string, opts ...recurly.Option) (*recurly.Transaction, error) {
	return m.getTransactionFn(transactionId, opts...)
}

func setMockTransactionAPI(mock *mockTransactionAPI) func() {
	orig := newTransactionAPI
	newTransactionAPI = func() (TransactionAPI, error) {
		return mock, nil
	}
	return func() { newTransactionAPI = orig }
}

// mockTransactionLister implements recurly.TransactionLister for testing.
type mockTransactionLister struct {
	transactions []recurly.Transaction
	fetched      bool
}

func (m *mockTransactionLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockTransactionLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockTransactionLister) Count() (*int64, error) {
	n := int64(len(m.transactions))
	return &n, nil
}

func (m *mockTransactionLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockTransactionLister) Data() []recurly.Transaction {
	return m.transactions
}

func (m *mockTransactionLister) HasMore() bool {
	return !m.fetched
}

func (m *mockTransactionLister) Next() string {
	return ""
}

func sampleTransactions() []recurly.Transaction {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return []recurly.Transaction{
		{
			Id:       "txn-001",
			Type:     "purchase",
			Account:  recurly.AccountMini{Code: "acct-code-1"},
			Status:   "success",
			Currency: "USD",
			Amount:   99.99,
			Success:  true,
			Origin:   "api",
			CreatedAt: &now,
		},
		{
			Id:       "txn-002",
			Type:     "refund",
			Account:  recurly.AccountMini{Code: "acct-code-2"},
			Status:   "declined",
			Currency: "EUR",
			Amount:   49.50,
			Success:  false,
			Origin:   "recurly_admin",
			CreatedAt: &now,
		},
	}
}

// --- transactions list ---

func TestTransactionsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("transactions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected transactions help to show 'list' subcommand")
	}
}

func TestTransactionsList_TableOutput(t *testing.T) {
	mock := &mockTransactionAPI{
		listTransactionsFn: func(params *recurly.ListTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockTransactionLister{transactions: sampleTransactions()}, nil
		},
	}
	cleanup := setMockTransactionAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("transactions", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"txn-001", "txn-002",
		"purchase", "refund",
		"acct-code-1", "acct-code-2",
		"success", "declined",
		"USD", "EUR",
		"99.99", "49.50",
		"true", "false",
		"api", "recurly_admin",
		"2025-01-15T10:30:00Z",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestTransactionsList_JSONOutput(t *testing.T) {
	mock := &mockTransactionAPI{
		listTransactionsFn: func(params *recurly.ListTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockTransactionLister{transactions: sampleTransactions()}, nil
		},
	}
	cleanup := setMockTransactionAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("transactions", "list", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("expected 'data' array in JSON envelope")
	}
	if len(data) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(data))
	}
}

func TestTransactionsList_JSONPrettyOutput(t *testing.T) {
	mock := &mockTransactionAPI{
		listTransactionsFn: func(params *recurly.ListTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockTransactionLister{transactions: sampleTransactions()}, nil
		},
	}
	cleanup := setMockTransactionAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("transactions", "list", "--output", "json-pretty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if !strings.Contains(out, "\n") {
		t.Error("expected indented JSON output")
	}
}

func TestTransactionsList_JQFilter(t *testing.T) {
	mock := &mockTransactionAPI{
		listTransactionsFn: func(params *recurly.ListTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return &mockTransactionLister{transactions: sampleTransactions()}, nil
		},
	}
	cleanup := setMockTransactionAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("transactions", "list", "--jq", ".data[0].id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "txn-001" {
		t.Errorf("expected jq to extract first transaction id, got %q", trimmed)
	}
}

func TestTransactionsList_APIError(t *testing.T) {
	mock := &mockTransactionAPI{
		listTransactionsFn: func(params *recurly.ListTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeServiceNotAvailable,
				Message: "service temporarily unavailable",
			}
		},
	}
	cleanup := setMockTransactionAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("transactions", "list")
	if err == nil {
		t.Fatal("expected error for API failure")
	}
	if !strings.Contains(stderr, "temporarily unavailable") {
		t.Errorf("expected service unavailable error, got %q", stderr)
	}
}
