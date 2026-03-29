package cmd

import (
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

func sampleTransaction() *recurly.Transaction {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return &recurly.Transaction{
		Id:               "txn-001",
		Uuid:             "uuid-001",
		Type:             "purchase",
		Origin:           "api",
		Status:           "success",
		Success:          true,
		Amount:           99.99,
		Currency:         "USD",
		Account:          recurly.AccountMini{Id: "acct-id-1", Code: "acct-code-1"},
		Invoice:          recurly.InvoiceMini{Id: "inv-id-1", Number: "1001"},
		CollectionMethod: "automatic",
		PaymentMethod: recurly.PaymentMethod{
			Object:   "credit_card",
			CardType: "Visa",
			LastFour: "1234",
		},
		IpAddressV4:   "192.168.1.1",
		StatusCode:    "approved",
		StatusMessage: "Transaction approved",
		Refunded:      false,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
}

func sampleTransactions() []recurly.Transaction {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return []recurly.Transaction{
		{
			Id:        "txn-001",
			Type:      "purchase",
			Account:   recurly.AccountMini{Code: "acct-code-1"},
			Status:    "success",
			Currency:  "USD",
			Amount:    99.99,
			Success:   true,
			Origin:    "api",
			CreatedAt: &now,
		},
		{
			Id:        "txn-002",
			Type:      "refund",
			Account:   recurly.AccountMini{Code: "acct-code-2"},
			Status:    "declined",
			Currency:  "EUR",
			Amount:    49.50,
			Success:   false,
			Origin:    "recurly_admin",
			CreatedAt: &now,
		},
	}
}

// --- transactions list ---

func TestTransactionsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "transactions", "--help")
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
			return &mockLister[recurly.Transaction]{items: sampleTransactions()}, nil
		},
	}
	app := newTestTransactionApp(mock)

	out, _, err := executeCommand(app, "transactions", "list")
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
			return &mockLister[recurly.Transaction]{items: sampleTransactions()}, nil
		},
	}
	app := newTestTransactionApp(mock)

	out, _, err := executeCommand(app, "transactions", "list", "--output", "json")
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
			return &mockLister[recurly.Transaction]{items: sampleTransactions()}, nil
		},
	}
	app := newTestTransactionApp(mock)

	out, _, err := executeCommand(app, "transactions", "list", "--output", "json-pretty")
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
			return &mockLister[recurly.Transaction]{items: sampleTransactions()}, nil
		},
	}
	app := newTestTransactionApp(mock)

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand(app, "transactions", "list", "--jq", ".data[0].id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "txn-001" {
		t.Errorf("expected jq to extract first transaction id, got %q", trimmed)
	}
}

// --- transactions get ---

func TestTransactionsGet_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "transactions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "get") {
		t.Error("expected transactions help to show 'get' subcommand")
	}
}

func TestTransactionsGet_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "transactions", "get")
	if err == nil {
		t.Fatal("expected error for missing arg")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected missing arg error, got %q", stderr)
	}
}

func TestTransactionsGet_PositionalArg(t *testing.T) {
	var capturedID string
	mock := &mockTransactionAPI{
		getTransactionFn: func(transactionId string, opts ...recurly.Option) (*recurly.Transaction, error) {
			capturedID = transactionId
			return sampleTransaction(), nil
		},
	}
	app := newTestTransactionApp(mock)

	_, _, err := executeCommand(app, "transactions", "get", "txn-abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "txn-abc123" {
		t.Errorf("expected captured ID %q, got %q", "txn-abc123", capturedID)
	}
}

func TestTransactionsGet_TableOutput(t *testing.T) {
	mock := &mockTransactionAPI{
		getTransactionFn: func(transactionId string, opts ...recurly.Option) (*recurly.Transaction, error) {
			return sampleTransaction(), nil
		},
	}
	app := newTestTransactionApp(mock)

	out, _, err := executeCommand(app, "transactions", "get", "txn-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"txn-001",
		"uuid-001",
		"purchase",
		"api",
		"success",
		"true",
		"99.99",
		"USD",
		"acct-id-1",
		"acct-code-1",
		"inv-id-1",
		"1001",
		"automatic",
		"credit_card",
		"Visa",
		"1234",
		"192.168.1.1",
		"approved",
		"Transaction approved",
		"false",
		"2025-01-15T10:30:00Z",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestTransactionsGet_JSONOutput(t *testing.T) {
	mock := &mockTransactionAPI{
		getTransactionFn: func(transactionId string, opts ...recurly.Option) (*recurly.Transaction, error) {
			return sampleTransaction(), nil
		},
	}
	app := newTestTransactionApp(mock)

	out, _, err := executeCommand(app, "transactions", "get", "txn-001", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["id"] != "txn-001" {
		t.Errorf("expected id 'txn-001', got %v", result["id"])
	}
}

func TestTransactionsGet_JSONPrettyOutput(t *testing.T) {
	mock := &mockTransactionAPI{
		getTransactionFn: func(transactionId string, opts ...recurly.Option) (*recurly.Transaction, error) {
			return sampleTransaction(), nil
		},
	}
	app := newTestTransactionApp(mock)

	out, _, err := executeCommand(app, "transactions", "get", "txn-001", "--output", "json-pretty")
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

func TestTransactionsGet_JQFilter(t *testing.T) {
	mock := &mockTransactionAPI{
		getTransactionFn: func(transactionId string, opts ...recurly.Option) (*recurly.Transaction, error) {
			return sampleTransaction(), nil
		},
	}
	app := newTestTransactionApp(mock)

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand(app, "transactions", "get", "txn-001", "--jq", ".id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "txn-001" {
		t.Errorf("expected jq to extract id, got %q", trimmed)
	}
}

func TestTransactionsGet_APIError(t *testing.T) {
	mock := &mockTransactionAPI{
		getTransactionFn: func(transactionId string, opts ...recurly.Option) (*recurly.Transaction, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "transaction not found",
			}
		},
	}
	app := newTestTransactionApp(mock)

	_, stderr, err := executeCommand(app, "transactions", "get", "txn-nonexistent")
	if err == nil {
		t.Fatal("expected error for API failure")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected not found error, got %q", stderr)
	}
}

func TestTransactionsList_Filters(t *testing.T) {
	var capturedParams *recurly.ListTransactionsParams
	mock := &mockTransactionAPI{
		listTransactionsFn: func(params *recurly.ListTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
			capturedParams = params
			return &mockLister[recurly.Transaction]{items: sampleTransactions()}, nil
		},
	}
	app := newTestTransactionApp(mock)

	_, _, err := executeCommand(app, "transactions", "list",
		"--type", "purchase",
		"--success", "true",
		"--limit", "5",
		"--order", "asc",
		"--sort", "created_at",
		"--begin-time", "2025-01-01T00:00:00Z",
		"--end-time", "2025-12-31T23:59:59Z",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams == nil {
		t.Fatal("expected params to be captured")
	}
	if capturedParams.Type == nil || *capturedParams.Type != "purchase" {
		t.Error("expected type=purchase")
	}
	if capturedParams.Success == nil || *capturedParams.Success != "true" {
		t.Error("expected success=true")
	}
	if capturedParams.Limit == nil || *capturedParams.Limit != 5 {
		t.Error("expected limit=5")
	}
	if capturedParams.Order == nil || *capturedParams.Order != "asc" {
		t.Error("expected order=asc")
	}
	if capturedParams.Sort == nil || *capturedParams.Sort != "created_at" {
		t.Error("expected sort=created_at")
	}
	if capturedParams.BeginTime == nil {
		t.Error("expected begin-time to be set")
	}
	if capturedParams.EndTime == nil {
		t.Error("expected end-time to be set")
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
	app := newTestTransactionApp(mock)

	_, stderr, err := executeCommand(app, "transactions", "list")
	if err == nil {
		t.Fatal("expected error for API failure")
	}
	if !strings.Contains(stderr, "temporarily unavailable") {
		t.Errorf("expected service unavailable error, got %q", stderr)
	}
}
