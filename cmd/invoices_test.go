package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

// mockInvoiceAPI implements InvoiceAPI for testing.
type mockInvoiceAPI struct {
	listInvoicesFn        func(params *recurly.ListInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error)
	getInvoiceFn          func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error)
	voidInvoiceFn         func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error)
	collectInvoiceFn      func(invoiceId string, params *recurly.CollectInvoiceParams, opts ...recurly.Option) (*recurly.Invoice, error)
	markInvoiceFailedFn   func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error)
	listInvoiceLineItemsFn func(invoiceId string, params *recurly.ListInvoiceLineItemsParams, opts ...recurly.Option) (recurly.LineItemLister, error)
}

func (m *mockInvoiceAPI) ListInvoices(params *recurly.ListInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
	return m.listInvoicesFn(params, opts...)
}

func (m *mockInvoiceAPI) GetInvoice(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
	return m.getInvoiceFn(invoiceId, opts...)
}

func (m *mockInvoiceAPI) VoidInvoice(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
	return m.voidInvoiceFn(invoiceId, opts...)
}

func (m *mockInvoiceAPI) CollectInvoice(invoiceId string, params *recurly.CollectInvoiceParams, opts ...recurly.Option) (*recurly.Invoice, error) {
	return m.collectInvoiceFn(invoiceId, params, opts...)
}

func (m *mockInvoiceAPI) MarkInvoiceFailed(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
	return m.markInvoiceFailedFn(invoiceId, opts...)
}

func (m *mockInvoiceAPI) ListInvoiceLineItems(invoiceId string, params *recurly.ListInvoiceLineItemsParams, opts ...recurly.Option) (recurly.LineItemLister, error) {
	return m.listInvoiceLineItemsFn(invoiceId, params, opts...)
}

func setMockInvoiceAPI(mock *mockInvoiceAPI) func() {
	orig := newInvoiceAPI
	newInvoiceAPI = func() (InvoiceAPI, error) {
		return mock, nil
	}
	return func() { newInvoiceAPI = orig }
}

// mockLineItemLister implements recurly.LineItemLister for testing.
type mockLineItemLister struct {
	lineItems []recurly.LineItem
	fetched   bool
}

func (m *mockLineItemLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockLineItemLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockLineItemLister) Count() (*int64, error) {
	n := int64(len(m.lineItems))
	return &n, nil
}

func (m *mockLineItemLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockLineItemLister) Data() []recurly.LineItem {
	return m.lineItems
}

func (m *mockLineItemLister) HasMore() bool {
	return !m.fetched
}

func (m *mockLineItemLister) Next() string {
	return ""
}

// mockInvoiceLister implements recurly.InvoiceLister for testing.
type mockInvoiceLister struct {
	invoices []recurly.Invoice
	fetched  bool
}

func (m *mockInvoiceLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockInvoiceLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockInvoiceLister) Count() (*int64, error) {
	n := int64(len(m.invoices))
	return &n, nil
}

func (m *mockInvoiceLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockInvoiceLister) Data() []recurly.Invoice {
	return m.invoices
}

func (m *mockInvoiceLister) HasMore() bool {
	return !m.fetched
}

func (m *mockInvoiceLister) Next() string {
	return ""
}

func sampleInvoice() *recurly.Invoice {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2025, 1, 16, 8, 0, 0, 0, time.UTC)
	due := time.Date(2025, 2, 14, 10, 30, 0, 0, time.UTC)
	closed := time.Date(2025, 1, 20, 12, 0, 0, 0, time.UTC)

	return &recurly.Invoice{
		Id:               "inv-abc123",
		Uuid:             "uuid-abc123",
		Number:           "1001",
		Type:             "charge",
		Origin:           "purchase",
		State:            "paid",
		Account:          recurly.AccountMini{Id: "acct-id-1", Code: "acct-code-1"},
		CollectionMethod: "automatic",
		Currency:         "USD",
		Subtotal:         100.00,
		Discount:         10.00,
		Tax:              8.10,
		Total:            98.10,
		Paid:             98.10,
		Balance:          0.00,
		RefundableAmount: 98.10,
		PoNumber:         "PO-12345",
		NetTerms:         30,
		NetTermsType:     "net",
		CreatedAt:        &now,
		UpdatedAt:        &updated,
		DueAt:            &due,
		ClosedAt:         &closed,
	}
}

func sampleLineItems() []recurly.LineItem {
	return []recurly.LineItem{
		{
			Id:          "li-001",
			Type:        "charge",
			Description: "Monthly subscription",
			Currency:    "USD",
			UnitAmount:  50.00,
			Quantity:    1,
			Subtotal:    50.00,
			Tax:         4.05,
			Amount:      54.05,
		},
		{
			Id:          "li-002",
			Type:        "charge",
			Description: "Add-on feature",
			Currency:    "USD",
			UnitAmount:  50.00,
			Quantity:    1,
			Subtotal:    50.00,
			Tax:         4.05,
			Amount:      54.05,
		},
	}
}

// --- invoices get ---

func TestInvoicesGet_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("invoices", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "get") {
		t.Error("expected invoices help to show 'get' subcommand")
	}
}

func TestInvoicesGet_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand("invoices", "get")
	if err == nil {
		t.Fatal("expected error when no invoice ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestInvoicesGet_NoAPIKey_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "")
	_, stderr, err := executeCommand("invoices", "get", "inv-abc123")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestInvoicesGet_PositionalArg(t *testing.T) {
	var capturedID string
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			capturedID = invoiceId
			return sampleInvoice(), nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("invoices", "get", "my-invoice-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "my-invoice-id" {
		t.Errorf("expected invoice ID 'my-invoice-id', got %q", capturedID)
	}
}

func TestInvoicesGet_TableOutput(t *testing.T) {
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return sampleInvoice(), nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "get", "inv-abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"ID", "inv-abc123",
		"UUID", "uuid-abc123",
		"Number", "1001",
		"Type", "charge",
		"Origin", "purchase",
		"State", "paid",
		"Account ID", "acct-id-1",
		"Account Code", "acct-code-1",
		"Collection Method", "automatic",
		"Currency", "USD",
		"Subtotal", "100.00",
		"Discount", "10.00",
		"Tax", "8.10",
		"Total", "98.10",
		"Paid", "98.10",
		"Balance", "0.00",
		"Refundable Amount", "98.10",
		"PO Number", "PO-12345",
		"Net Terms Type", "net",
		"Created At", "2025-01-15T10:30:00Z",
		"Updated At", "2025-01-16T08:00:00Z",
		"Due At", "2025-02-14T10:30:00Z",
		"Closed At", "2025-01-20T12:00:00Z",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestInvoicesGet_TableOutput_NoLineItemsByDefault(t *testing.T) {
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return sampleInvoice(), nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "get", "inv-abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, "Line Items:") {
		t.Error("expected no line items section when --line-items is not passed")
	}
}

func TestInvoicesGet_JSONOutput(t *testing.T) {
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return sampleInvoice(), nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "get", "inv-abc123", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["id"] != "inv-abc123" {
		t.Errorf("expected id=inv-abc123 in JSON, got %v", result["id"])
	}
}

func TestInvoicesGet_JSONPrettyOutput(t *testing.T) {
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return sampleInvoice(), nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "get", "inv-abc123", "--output", "json-pretty")
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

func TestInvoicesGet_NotFound(t *testing.T) {
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Invoice with id = nonexistent",
			}
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("invoices", "get", "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found invoice")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

func TestInvoicesGet_SDKError(t *testing.T) {
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeServiceNotAvailable,
				Message: "Service not available",
			}
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("invoices", "get", "inv-abc123")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- invoices get --line-items ---

func TestInvoicesGet_LineItems_Shown(t *testing.T) {
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return sampleInvoice(), nil
		},
		listInvoiceLineItemsFn: func(invoiceId string, params *recurly.ListInvoiceLineItemsParams, opts ...recurly.Option) (recurly.LineItemLister, error) {
			return &mockLineItemLister{lineItems: sampleLineItems()}, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "get", "inv-abc123", "--line-items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Line Items:") {
		t.Error("expected 'Line Items:' header in output")
	}

	for _, expected := range []string{
		"li-001", "li-002",
		"Monthly subscription", "Add-on feature",
		"charge", "50.00", "54.05",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestInvoicesGet_LineItems_WithCount(t *testing.T) {
	var capturedParams *recurly.ListInvoiceLineItemsParams
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return sampleInvoice(), nil
		},
		listInvoiceLineItemsFn: func(invoiceId string, params *recurly.ListInvoiceLineItemsParams, opts ...recurly.Option) (recurly.LineItemLister, error) {
			capturedParams = params
			return &mockLineItemLister{lineItems: sampleLineItems()}, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("invoices", "get", "inv-abc123", "--line-items=50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams == nil || capturedParams.Limit == nil || *capturedParams.Limit != 50 {
		t.Error("expected limit=50 to be passed to ListInvoiceLineItems")
	}
}

func TestInvoicesGet_LineItems_HasMoreMessage(t *testing.T) {
	// Use a lister that reports hasMore=true after first fetch
	lister := &mockLineItemListerWithMore{
		lineItems: sampleLineItems(),
	}
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return sampleInvoice(), nil
		},
		listInvoiceLineItemsFn: func(invoiceId string, params *recurly.ListInvoiceLineItemsParams, opts ...recurly.Option) (recurly.LineItemLister, error) {
			return lister, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "get", "inv-abc123", "--line-items=2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Showing 2 line items (more available)") {
		t.Errorf("expected 'more available' message, got output:\n%s", out)
	}
	if !strings.Contains(out, "recurly invoices line-items inv-abc123") {
		t.Errorf("expected usage hint for line-items command, got output:\n%s", out)
	}
}

func TestInvoicesGet_LineItems_JSONOutput(t *testing.T) {
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return sampleInvoice(), nil
		},
		listInvoiceLineItemsFn: func(invoiceId string, params *recurly.ListInvoiceLineItemsParams, opts ...recurly.Option) (recurly.LineItemLister, error) {
			return &mockLineItemLister{lineItems: sampleLineItems()}, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "get", "inv-abc123", "--line-items", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	lineItems, ok := result["line_items"].([]interface{})
	if !ok {
		t.Fatal("expected line_items array in JSON output")
	}
	if len(lineItems) != 2 {
		t.Errorf("expected 2 line items in JSON, got %d", len(lineItems))
	}
}

func TestInvoicesGet_LineItems_NotShownWithoutFlag(t *testing.T) {
	lineItemsCalled := false
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return sampleInvoice(), nil
		},
		listInvoiceLineItemsFn: func(invoiceId string, params *recurly.ListInvoiceLineItemsParams, opts ...recurly.Option) (recurly.LineItemLister, error) {
			lineItemsCalled = true
			return &mockLineItemLister{lineItems: sampleLineItems()}, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("invoices", "get", "inv-abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lineItemsCalled {
		t.Error("expected ListInvoiceLineItems NOT to be called without --line-items flag")
	}
}

func TestInvoicesGet_JQFilter(t *testing.T) {
	mock := &mockInvoiceAPI{
		getInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return sampleInvoice(), nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("invoices", "get", "inv-abc123", "--jq", ".id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "inv-abc123" {
		t.Errorf("expected jq to extract id, got %q", trimmed)
	}
}

// mockLineItemListerWithMore simulates a lister that has more items
// beyond the first page — after pagination.Collect takes what it needs,
// hasMore remains true.
type mockLineItemListerWithMore struct {
	lineItems []recurly.LineItem
	fetched   bool
}

func (m *mockLineItemListerWithMore) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockLineItemListerWithMore) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockLineItemListerWithMore) Count() (*int64, error) {
	n := int64(len(m.lineItems))
	return &n, nil
}

func (m *mockLineItemListerWithMore) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockLineItemListerWithMore) Data() []recurly.LineItem {
	return m.lineItems
}

func (m *mockLineItemListerWithMore) HasMore() bool {
	// Always has more (simulates a lister with more pages)
	return !m.fetched || true
}

func (m *mockLineItemListerWithMore) Next() string {
	return ""
}

// --- invoices void ---

func TestInvoicesVoid_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("invoices", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "void") {
		t.Error("expected invoices help to show 'void' subcommand")
	}
}

func TestInvoicesVoid_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand("invoices", "void")
	if err == nil {
		t.Fatal("expected error when no invoice ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestInvoicesVoid_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(stdin, "invoices", "void", "inv-abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to void invoice inv-abc123? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Void cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestInvoicesVoid_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockInvoiceAPI{
		voidInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			capturedID = invoiceId
			inv := sampleInvoice()
			inv.State = "voided"
			return inv, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(stdin, "invoices", "void", "inv-abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "inv-abc123" {
		t.Errorf("expected invoice ID 'inv-abc123', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "voided") {
		t.Errorf("expected updated invoice detail with 'voided' state, got:\n%s", out)
	}
}

func TestInvoicesVoid_YesFlag_SkipsConfirmation(t *testing.T) {
	mock := &mockInvoiceAPI{
		voidInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			inv := sampleInvoice()
			inv.State = "voided"
			return inv, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, stderr, err := executeCommand("invoices", "void", "inv-abc123", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected --yes to skip confirmation prompt")
	}
	if !strings.Contains(out, "voided") {
		t.Errorf("expected updated invoice detail, got:\n%s", out)
	}
}

func TestInvoicesVoid_JSONOutput(t *testing.T) {
	mock := &mockInvoiceAPI{
		voidInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			inv := sampleInvoice()
			inv.State = "voided"
			return inv, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "void", "inv-abc123", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["state"] != "voided" {
		t.Errorf("expected state=voided in JSON, got %v", result["state"])
	}
}

func TestInvoicesVoid_JSONPrettyOutput(t *testing.T) {
	mock := &mockInvoiceAPI{
		voidInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			inv := sampleInvoice()
			inv.State = "voided"
			return inv, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "void", "inv-abc123", "--yes", "--output", "json-pretty")
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

func TestInvoicesVoid_JQFilter(t *testing.T) {
	mock := &mockInvoiceAPI{
		voidInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			inv := sampleInvoice()
			inv.State = "voided"
			return inv, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("invoices", "void", "inv-abc123", "--yes", "--jq", ".state")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "voided" {
		t.Errorf("expected jq to extract state, got %q", trimmed)
	}
}

func TestInvoicesVoid_APIError(t *testing.T) {
	mock := &mockInvoiceAPI{
		voidInvoiceFn: func(invoiceId string, opts ...recurly.Option) (*recurly.Invoice, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Invoice with id = nonexistent",
			}
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("invoices", "void", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for not found invoice")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

// --- invoices collect ---

func TestInvoicesCollect_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("invoices", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "collect") {
		t.Error("expected invoices help to show 'collect' subcommand")
	}
}

func TestInvoicesCollect_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand("invoices", "collect")
	if err == nil {
		t.Fatal("expected error when no invoice ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestInvoicesCollect_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(stdin, "invoices", "collect", "inv-abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to collect invoice inv-abc123? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Collection cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestInvoicesCollect_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockInvoiceAPI{
		collectInvoiceFn: func(invoiceId string, params *recurly.CollectInvoiceParams, opts ...recurly.Option) (*recurly.Invoice, error) {
			capturedID = invoiceId
			inv := sampleInvoice()
			inv.State = "paid"
			return inv, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(stdin, "invoices", "collect", "inv-abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "inv-abc123" {
		t.Errorf("expected invoice ID 'inv-abc123', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "paid") {
		t.Errorf("expected updated invoice detail with 'paid' state, got:\n%s", out)
	}
}

func TestInvoicesCollect_YesFlag_SkipsConfirmation(t *testing.T) {
	mock := &mockInvoiceAPI{
		collectInvoiceFn: func(invoiceId string, params *recurly.CollectInvoiceParams, opts ...recurly.Option) (*recurly.Invoice, error) {
			inv := sampleInvoice()
			inv.State = "paid"
			return inv, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, stderr, err := executeCommand("invoices", "collect", "inv-abc123", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected --yes to skip confirmation prompt")
	}
	if !strings.Contains(out, "paid") {
		t.Errorf("expected updated invoice detail, got:\n%s", out)
	}
}

func TestInvoicesCollect_JSONOutput(t *testing.T) {
	mock := &mockInvoiceAPI{
		collectInvoiceFn: func(invoiceId string, params *recurly.CollectInvoiceParams, opts ...recurly.Option) (*recurly.Invoice, error) {
			inv := sampleInvoice()
			inv.State = "paid"
			return inv, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "collect", "inv-abc123", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["state"] != "paid" {
		t.Errorf("expected state=paid in JSON, got %v", result["state"])
	}
}

func TestInvoicesCollect_JSONPrettyOutput(t *testing.T) {
	mock := &mockInvoiceAPI{
		collectInvoiceFn: func(invoiceId string, params *recurly.CollectInvoiceParams, opts ...recurly.Option) (*recurly.Invoice, error) {
			inv := sampleInvoice()
			inv.State = "paid"
			return inv, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("invoices", "collect", "inv-abc123", "--yes", "--output", "json-pretty")
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

func TestInvoicesCollect_JQFilter(t *testing.T) {
	mock := &mockInvoiceAPI{
		collectInvoiceFn: func(invoiceId string, params *recurly.CollectInvoiceParams, opts ...recurly.Option) (*recurly.Invoice, error) {
			inv := sampleInvoice()
			inv.State = "paid"
			return inv, nil
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("invoices", "collect", "inv-abc123", "--yes", "--jq", ".state")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "paid" {
		t.Errorf("expected jq to extract state, got %q", trimmed)
	}
}

func TestInvoicesCollect_APIError(t *testing.T) {
	mock := &mockInvoiceAPI{
		collectInvoiceFn: func(invoiceId string, params *recurly.CollectInvoiceParams, opts ...recurly.Option) (*recurly.Invoice, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Invoice with id = nonexistent",
			}
		},
	}
	cleanup := setMockInvoiceAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("invoices", "collect", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for not found invoice")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}
