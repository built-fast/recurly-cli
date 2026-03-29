package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

// mockItemAPI implements ItemAPI for testing.
type mockItemAPI struct {
	listItemsFn      func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error)
	getItemFn        func(itemId string, opts ...recurly.Option) (*recurly.Item, error)
	createItemFn     func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error)
	updateItemFn     func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error)
	deactivateItemFn func(itemId string, opts ...recurly.Option) (*recurly.Item, error)
	reactivateItemFn func(itemId string, opts ...recurly.Option) (*recurly.Item, error)
}

func (m *mockItemAPI) ListItems(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
	return m.listItemsFn(params, opts...)
}

func (m *mockItemAPI) GetItem(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
	return m.getItemFn(itemId, opts...)
}

func (m *mockItemAPI) CreateItem(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
	return m.createItemFn(body, opts...)
}

func (m *mockItemAPI) UpdateItem(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
	return m.updateItemFn(itemId, body, opts...)
}

func (m *mockItemAPI) DeactivateItem(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
	return m.deactivateItemFn(itemId, opts...)
}

func (m *mockItemAPI) ReactivateItem(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
	return m.reactivateItemFn(itemId, opts...)
}

// sampleItem returns a test item with predictable fields.
func sampleItem() recurly.Item {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return recurly.Item{
		Code:        "widget-1",
		Name:        "Premium Widget",
		ExternalSku: "SKU-001",
		State:       "active",
		CreatedAt:   &now,
	}
}

// sampleItemDetail returns a test item with all detail fields populated.
func sampleItemDetail() *recurly.Item {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC)
	return &recurly.Item{
		Code:                   "widget-1",
		Name:                   "Premium Widget",
		Description:            "A high-quality widget",
		ExternalSku:            "SKU-001",
		AccountingCode:         "ACC-100",
		RevenueScheduleType:    "evenly",
		TaxCode:                "digital",
		TaxExempt:              false,
		AvalaraTransactionType: 3,
		AvalaraServiceType:     6,
		HarmonizedSystemCode:   "8471.30",
		State:                  "active",
		CreatedAt:              &now,
		UpdatedAt:              &updated,
	}
}

// --- items get ---

func TestItemsGet_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "get") {
		t.Error("expected items help to show 'get' subcommand")
	}
}

func TestItemsGet_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "items", "get")
	if err == nil {
		t.Fatal("expected error when no item ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestItemsGet_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "items", "get", "item123")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestItemsGet_PositionalArg(t *testing.T) {
	var capturedID string
	mock := &mockItemAPI{
		getItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			capturedID = itemId
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "get", "my-item-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "my-item-id" {
		t.Errorf("expected item ID 'my-item-id', got %q", capturedID)
	}
}

func TestItemsGet_TableOutput(t *testing.T) {
	viper.Reset()
	mock := &mockItemAPI{
		getItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "get", "item123", "--output", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"Code", "widget-1",
		"Name", "Premium Widget",
		"Description", "A high-quality widget",
		"External SKU", "SKU-001",
		"Accounting Code", "ACC-100",
		"Revenue Schedule Type", "evenly",
		"Tax Code", "digital",
		"Tax Exempt", "false",
		"Avalara Transaction Type", "3",
		"Avalara Service Type", "6",
		"Harmonized System Code", "8471.30",
		"State", "active",
		"Created At",
		"Updated At",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestItemsGet_JSONOutput(t *testing.T) {
	viper.Reset()
	mock := &mockItemAPI{
		getItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "get", "item123", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["code"] != "widget-1" {
		t.Errorf("expected code=widget-1 in JSON, got %v", result["code"])
	}
	if _, ok := result["object"]; ok {
		if result["object"] == "list" {
			t.Error("expected single item object, not list envelope")
		}
	}
}

func TestItemsGet_JSONPrettyOutput(t *testing.T) {
	viper.Reset()
	mock := &mockItemAPI{
		getItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "get", "item123", "--output", "json-pretty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "  ") {
		t.Error("expected indented JSON output for json-pretty format")
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["code"] != "widget-1" {
		t.Errorf("expected code=widget-1, got %v", result["code"])
	}
}

func TestItemsGet_JQFilter(t *testing.T) {
	viper.Reset()
	mock := &mockItemAPI{
		getItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "get", "item123", "--jq", ".code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "widget-1" {
		t.Errorf("expected jq output 'widget-1', got %q", trimmed)
	}
}

func TestItemsGet_SDKError(t *testing.T) {
	mock := &mockItemAPI{
		getItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "get", "item123")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestItemsGet_NotFound(t *testing.T) {
	mock := &mockItemAPI{
		getItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Item with id = nonexistent",
			}
		},
	}
	app := newTestItemApp(mock)

	_, stderr, err := executeCommand(app, "items", "get", "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found item")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

// --- items list ---

func TestItemsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected items help to show 'list' subcommand")
	}
}

func TestItemsListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--order", "--sort", "--state", "--begin-time", "--end-time"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestItemsList_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "items", "list")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestItemsList_InvalidBeginTime_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand(nil, "items", "list", "--begin-time", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid begin-time")
	}
	if !strings.Contains(stderr, "invalid --begin-time") {
		t.Errorf("expected 'invalid --begin-time' error, got %q", stderr)
	}
}

func TestItemsList_InvalidEndTime_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand(nil, "items", "list", "--end-time", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid end-time")
	}
	if !strings.Contains(stderr, "invalid --end-time") {
		t.Errorf("expected 'invalid --end-time' error, got %q", stderr)
	}
}

func TestItemsList_PaginationParams(t *testing.T) {
	var capturedParams *recurly.ListItemsParams

	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			capturedParams = params
			return &mockLister[recurly.Item]{items: []recurly.Item{}}, nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "list", "--limit", "50", "--order", "desc", "--sort", "updated_at")
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

func TestItemsList_FilterParams(t *testing.T) {
	var capturedParams *recurly.ListItemsParams

	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			capturedParams = params
			return &mockLister[recurly.Item]{items: []recurly.Item{}}, nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "list",
		"--state", "active",
		"--begin-time", "2025-01-01T00:00:00Z",
		"--end-time", "2025-12-31T23:59:59Z",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.State == nil || *capturedParams.State != "active" {
		t.Errorf("expected state=active, got %v", capturedParams.State)
	}
	if capturedParams.BeginTime == nil {
		t.Error("expected begin_time to be set")
	}
	if capturedParams.EndTime == nil {
		t.Error("expected end_time to be set")
	}
}

func TestItemsList_UnsetFlagsNotSent(t *testing.T) {
	var capturedParams *recurly.ListItemsParams

	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			capturedParams = params
			return &mockLister[recurly.Item]{items: []recurly.Item{}}, nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.Limit != nil {
		t.Error("expected limit to be nil when not set")
	}
	if capturedParams.Order != nil {
		t.Error("expected order to be nil when not set")
	}
	if capturedParams.Sort != nil {
		t.Error("expected sort to be nil when not set")
	}
	if capturedParams.State != nil {
		t.Error("expected state to be nil when not set")
	}
	if capturedParams.BeginTime != nil {
		t.Error("expected begin_time to be nil when not set")
	}
	if capturedParams.EndTime != nil {
		t.Error("expected end_time to be nil when not set")
	}
}

func TestItemsList_TableOutput(t *testing.T) {
	item := sampleItem()
	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			return &mockLister[recurly.Item]{items: []recurly.Item{item}}, nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"widget-1", "Premium Widget", "SKU-001", "active"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	// Table should have column headers
	for _, header := range []string{"Code", "Name", "External SKU", "State", "Created At"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestItemsList_JSONOutput(t *testing.T) {
	item := sampleItem()
	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			return &mockLister[recurly.Item]{items: []recurly.Item{item}}, nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "list", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var envelope struct {
		Object  string                   `json:"object"`
		HasMore bool                     `json:"has_more"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &envelope); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v\noutput: %s", err, out)
	}
	if envelope.Object != "list" {
		t.Errorf("expected object=list, got %s", envelope.Object)
	}
	if envelope.HasMore {
		t.Error("expected has_more=false")
	}
	if len(envelope.Data) != 1 {
		t.Fatalf("expected 1 item in JSON output, got %d", len(envelope.Data))
	}
	if envelope.Data[0]["code"] != "widget-1" {
		t.Errorf("expected code=widget-1 in JSON, got %v", envelope.Data[0]["code"])
	}
}

func TestItemsList_JSONPrettyOutput(t *testing.T) {
	item := sampleItem()
	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			return &mockLister[recurly.Item]{items: []recurly.Item{item}}, nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "list", "--output", "json-pretty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// json-pretty should contain indentation
	if !strings.Contains(out, "  ") {
		t.Error("expected indented JSON output for json-pretty format")
	}

	var envelope struct {
		Object  string                   `json:"object"`
		HasMore bool                     `json:"has_more"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &envelope); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v\noutput: %s", err, out)
	}
	if envelope.Object != "list" {
		t.Errorf("expected object=list, got %s", envelope.Object)
	}
}

func TestItemsList_JQFilter(t *testing.T) {
	item := sampleItem()
	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			return &mockLister[recurly.Item]{items: []recurly.Item{item}}, nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "list", "--jq", ".data[].code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if !strings.Contains(trimmed, "widget-1") {
		t.Errorf("expected jq output to contain 'widget-1', got: %s", trimmed)
	}
}

func TestItemsList_SDKError(t *testing.T) {
	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "list")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestItemsList_EmptyResults(t *testing.T) {
	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			return &mockLister[recurly.Item]{items: []recurly.Item{}}, nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, "widget-1") {
		t.Error("expected no item data in empty results")
	}
}

func TestItemsList_EmptyResults_JSON(t *testing.T) {
	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			return &mockLister[recurly.Item]{items: []recurly.Item{}}, nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "list", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var envelope struct {
		Object  string        `json:"object"`
		HasMore bool          `json:"has_more"`
		Data    []interface{} `json:"data"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &envelope); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if len(envelope.Data) != 0 {
		t.Errorf("expected empty data array, got %d items", len(envelope.Data))
	}
}

// --- items create ---

func TestItemsCreate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "create") {
		t.Error("expected items help to show 'create' subcommand")
	}
}

func TestItemsCreateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "create", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{
		"--code", "--name", "--description", "--external-sku",
		"--accounting-code", "--revenue-schedule-type", "--tax-code",
		"--tax-exempt", "--avalara-transaction-type", "--avalara-service-type",
		"--harmonized-system-code", "--currency", "--unit-amount",
	} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestItemsCreate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "items", "create", "--code", "test")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestItemsCreate_CoreFlags(t *testing.T) {
	var capturedBody *recurly.ItemCreate
	mock := &mockItemAPI{
		createItemFn: func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedBody = body
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "create",
		"--code", "widget-1",
		"--name", "Premium Widget",
		"--description", "A high-quality widget",
		"--external-sku", "SKU-001",
		"--accounting-code", "ACC-100",
		"--revenue-schedule-type", "evenly",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if *capturedBody.Code != "widget-1" {
		t.Errorf("expected code=widget-1, got %v", *capturedBody.Code)
	}
	if *capturedBody.Name != "Premium Widget" {
		t.Errorf("expected name=Premium Widget, got %v", *capturedBody.Name)
	}
	if *capturedBody.Description != "A high-quality widget" {
		t.Errorf("expected description, got %v", *capturedBody.Description)
	}
	if *capturedBody.ExternalSku != "SKU-001" {
		t.Errorf("expected external-sku=SKU-001, got %v", *capturedBody.ExternalSku)
	}
	if *capturedBody.AccountingCode != "ACC-100" {
		t.Errorf("expected accounting-code=ACC-100, got %v", *capturedBody.AccountingCode)
	}
	if *capturedBody.RevenueScheduleType != "evenly" {
		t.Errorf("expected revenue-schedule-type=evenly, got %v", *capturedBody.RevenueScheduleType)
	}
}

func TestItemsCreate_TaxFlags(t *testing.T) {
	var capturedBody *recurly.ItemCreate
	mock := &mockItemAPI{
		createItemFn: func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedBody = body
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "create",
		"--code", "tax-item",
		"--tax-code", "digital",
		"--tax-exempt",
		"--avalara-transaction-type", "3",
		"--avalara-service-type", "6",
		"--harmonized-system-code", "8471.30",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.TaxCode != "digital" {
		t.Errorf("expected tax-code=digital, got %v", *capturedBody.TaxCode)
	}
	if *capturedBody.TaxExempt != true {
		t.Error("expected tax-exempt=true")
	}
	if *capturedBody.AvalaraTransactionType != 3 {
		t.Errorf("expected avalara-transaction-type=3, got %v", *capturedBody.AvalaraTransactionType)
	}
	if *capturedBody.AvalaraServiceType != 6 {
		t.Errorf("expected avalara-service-type=6, got %v", *capturedBody.AvalaraServiceType)
	}
	if *capturedBody.HarmonizedSystemCode != "8471.30" {
		t.Errorf("expected harmonized-system-code=8471.30, got %v", *capturedBody.HarmonizedSystemCode)
	}
}

func TestItemsCreate_MultiCurrencyFlags(t *testing.T) {
	var capturedBody *recurly.ItemCreate
	mock := &mockItemAPI{
		createItemFn: func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedBody = body
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "create",
		"--code", "multi",
		"--name", "Multi Currency Item",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "10.00", "--unit-amount", "9.00",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Currencies == nil {
		t.Fatal("expected currencies to be set")
	}
	currencies := *capturedBody.Currencies
	if len(currencies) != 2 {
		t.Fatalf("expected 2 currencies, got %d", len(currencies))
	}
	if *currencies[0].Currency != "USD" || *currencies[0].UnitAmount != 10.00 {
		t.Errorf("expected USD/10.00, got %s/%.2f", *currencies[0].Currency, *currencies[0].UnitAmount)
	}
	if *currencies[1].Currency != "EUR" || *currencies[1].UnitAmount != 9.00 {
		t.Errorf("expected EUR/9.00, got %s/%.2f", *currencies[1].Currency, *currencies[1].UnitAmount)
	}
}

func TestItemsCreate_SingleCurrency(t *testing.T) {
	var capturedBody *recurly.ItemCreate
	mock := &mockItemAPI{
		createItemFn: func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedBody = body
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "create",
		"--code", "single",
		"--name", "Single Currency Item",
		"--currency", "USD",
		"--unit-amount", "19.99",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Currencies == nil {
		t.Fatal("expected currencies to be set")
	}
	currencies := *capturedBody.Currencies
	if len(currencies) != 1 {
		t.Fatalf("expected 1 currency, got %d", len(currencies))
	}
	if *currencies[0].Currency != "USD" || *currencies[0].UnitAmount != 19.99 {
		t.Errorf("expected USD/19.99, got %s/%.2f", *currencies[0].Currency, *currencies[0].UnitAmount)
	}
}

func TestItemsCreate_CurrencyUnitAmountMismatch(t *testing.T) {
	mock := &mockItemAPI{
		createItemFn: func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, stderr, err := executeCommand(app, "items", "create",
		"--code", "bad",
		"--name", "Bad Item",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "10.00",
	)
	if err == nil {
		t.Fatal("expected error for mismatched currency/unit-amount")
	}
	if !strings.Contains(stderr, "number of --currency values must match --unit-amount values") {
		t.Errorf("expected mismatch error, got %q", stderr)
	}
}

func TestItemsCreate_UnsetFlagsNotSent(t *testing.T) {
	var capturedBody *recurly.ItemCreate
	mock := &mockItemAPI{
		createItemFn: func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedBody = body
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "create", "--code", "minimal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Name != nil {
		t.Error("expected name to be nil when not set")
	}
	if capturedBody.Description != nil {
		t.Error("expected description to be nil when not set")
	}
	if capturedBody.ExternalSku != nil {
		t.Error("expected external-sku to be nil when not set")
	}
	if capturedBody.AccountingCode != nil {
		t.Error("expected accounting-code to be nil when not set")
	}
	if capturedBody.RevenueScheduleType != nil {
		t.Error("expected revenue-schedule-type to be nil when not set")
	}
	if capturedBody.TaxCode != nil {
		t.Error("expected tax-code to be nil when not set")
	}
	if capturedBody.TaxExempt != nil {
		t.Error("expected tax-exempt to be nil when not set")
	}
	if capturedBody.AvalaraTransactionType != nil {
		t.Error("expected avalara-transaction-type to be nil when not set")
	}
	if capturedBody.AvalaraServiceType != nil {
		t.Error("expected avalara-service-type to be nil when not set")
	}
	if capturedBody.HarmonizedSystemCode != nil {
		t.Error("expected harmonized-system-code to be nil when not set")
	}
	if capturedBody.Currencies != nil {
		t.Error("expected currencies to be nil when not set")
	}
}

func TestItemsCreate_AllFlagsPopulated(t *testing.T) {
	var capturedBody *recurly.ItemCreate
	mock := &mockItemAPI{
		createItemFn: func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedBody = body
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "create",
		"--code", "full",
		"--name", "Full Item",
		"--description", "All flags test",
		"--external-sku", "SKU-FULL",
		"--accounting-code", "ACC-FULL",
		"--revenue-schedule-type", "evenly",
		"--tax-code", "digital",
		"--tax-exempt",
		"--avalara-transaction-type", "3",
		"--avalara-service-type", "6",
		"--harmonized-system-code", "8471.30",
		"--currency", "USD",
		"--unit-amount", "29.99",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.Code != "full" {
		t.Errorf("expected code=full, got %v", *capturedBody.Code)
	}
	if *capturedBody.Name != "Full Item" {
		t.Errorf("expected name=Full Item, got %v", *capturedBody.Name)
	}
	if *capturedBody.Description != "All flags test" {
		t.Errorf("expected description, got %v", *capturedBody.Description)
	}
	if *capturedBody.ExternalSku != "SKU-FULL" {
		t.Errorf("expected external-sku=SKU-FULL, got %v", *capturedBody.ExternalSku)
	}
	if *capturedBody.AccountingCode != "ACC-FULL" {
		t.Errorf("expected accounting-code=ACC-FULL, got %v", *capturedBody.AccountingCode)
	}
	if *capturedBody.RevenueScheduleType != "evenly" {
		t.Errorf("expected revenue-schedule-type=evenly, got %v", *capturedBody.RevenueScheduleType)
	}
	if *capturedBody.TaxCode != "digital" {
		t.Errorf("expected tax-code=digital, got %v", *capturedBody.TaxCode)
	}
	if *capturedBody.TaxExempt != true {
		t.Error("expected tax-exempt=true")
	}
	if *capturedBody.AvalaraTransactionType != 3 {
		t.Errorf("expected avalara-transaction-type=3, got %v", *capturedBody.AvalaraTransactionType)
	}
	if *capturedBody.AvalaraServiceType != 6 {
		t.Errorf("expected avalara-service-type=6, got %v", *capturedBody.AvalaraServiceType)
	}
	if *capturedBody.HarmonizedSystemCode != "8471.30" {
		t.Errorf("expected harmonized-system-code=8471.30, got %v", *capturedBody.HarmonizedSystemCode)
	}
	if capturedBody.Currencies == nil || len(*capturedBody.Currencies) != 1 {
		t.Fatal("expected 1 currency")
	}
}

func TestItemsCreate_TableOutput(t *testing.T) {
	viper.Reset()
	mock := &mockItemAPI{
		createItemFn: func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "create", "--code", "widget-1", "--output", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"Code", "widget-1",
		"Name", "Premium Widget",
		"Description", "A high-quality widget",
		"External SKU", "SKU-001",
		"State", "active",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestItemsCreate_JSONOutput(t *testing.T) {
	viper.Reset()
	mock := &mockItemAPI{
		createItemFn: func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "create", "--code", "widget-1", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["code"] != "widget-1" {
		t.Errorf("expected code=widget-1 in JSON, got %v", result["code"])
	}
}

func TestItemsCreate_SDKError(t *testing.T) {
	mock := &mockItemAPI{
		createItemFn: func(body *recurly.ItemCreate, opts ...recurly.Option) (*recurly.Item, error) {
			return nil, fmt.Errorf("validation failed")
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "create", "--code", "bad")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- items update ---

func TestItemsUpdate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "update") {
		t.Error("expected items help to show 'update' subcommand")
	}
}

func TestItemsUpdateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "update", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{
		"--code", "--name", "--description", "--external-sku",
		"--accounting-code", "--revenue-schedule-type", "--tax-code",
		"--tax-exempt", "--avalara-transaction-type", "--avalara-service-type",
		"--harmonized-system-code", "--currency", "--unit-amount",
	} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestItemsUpdate_MissingArg_ReturnsError(t *testing.T) {
	mock := &mockItemAPI{
		updateItemFn: func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "update")
	if err == nil {
		t.Fatal("expected error when item_id is missing")
	}
}

func TestItemsUpdate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "items", "update", "item-123", "--name", "test")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestItemsUpdate_PositionalArg(t *testing.T) {
	var capturedID string
	mock := &mockItemAPI{
		updateItemFn: func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedID = itemId
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "update", "item-abc123", "--name", "Updated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "item-abc123" {
		t.Errorf("expected item_id=item-abc123, got %q", capturedID)
	}
}

func TestItemsUpdate_CoreFlags(t *testing.T) {
	var capturedBody *recurly.ItemUpdate
	mock := &mockItemAPI{
		updateItemFn: func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedBody = body
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "update", "item-123",
		"--code", "widget-1",
		"--name", "Premium Widget",
		"--description", "A high-quality widget",
		"--external-sku", "SKU-001",
		"--accounting-code", "ACC-100",
		"--revenue-schedule-type", "evenly",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if *capturedBody.Code != "widget-1" {
		t.Errorf("expected code=widget-1, got %v", *capturedBody.Code)
	}
	if *capturedBody.Name != "Premium Widget" {
		t.Errorf("expected name=Premium Widget, got %v", *capturedBody.Name)
	}
	if *capturedBody.Description != "A high-quality widget" {
		t.Errorf("expected description, got %v", *capturedBody.Description)
	}
	if *capturedBody.ExternalSku != "SKU-001" {
		t.Errorf("expected external-sku=SKU-001, got %v", *capturedBody.ExternalSku)
	}
	if *capturedBody.AccountingCode != "ACC-100" {
		t.Errorf("expected accounting-code=ACC-100, got %v", *capturedBody.AccountingCode)
	}
	if *capturedBody.RevenueScheduleType != "evenly" {
		t.Errorf("expected revenue-schedule-type=evenly, got %v", *capturedBody.RevenueScheduleType)
	}
}

func TestItemsUpdate_TaxFlags(t *testing.T) {
	var capturedBody *recurly.ItemUpdate
	mock := &mockItemAPI{
		updateItemFn: func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedBody = body
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "update", "item-123",
		"--tax-code", "digital",
		"--tax-exempt",
		"--avalara-transaction-type", "3",
		"--avalara-service-type", "6",
		"--harmonized-system-code", "8471.30",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.TaxCode != "digital" {
		t.Errorf("expected tax-code=digital, got %v", *capturedBody.TaxCode)
	}
	if *capturedBody.TaxExempt != true {
		t.Error("expected tax-exempt=true")
	}
	if *capturedBody.AvalaraTransactionType != 3 {
		t.Errorf("expected avalara-transaction-type=3, got %v", *capturedBody.AvalaraTransactionType)
	}
	if *capturedBody.AvalaraServiceType != 6 {
		t.Errorf("expected avalara-service-type=6, got %v", *capturedBody.AvalaraServiceType)
	}
	if *capturedBody.HarmonizedSystemCode != "8471.30" {
		t.Errorf("expected harmonized-system-code=8471.30, got %v", *capturedBody.HarmonizedSystemCode)
	}
}

func TestItemsUpdate_MultiCurrencyFlags(t *testing.T) {
	var capturedBody *recurly.ItemUpdate
	mock := &mockItemAPI{
		updateItemFn: func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedBody = body
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "update", "item-123",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "10.00", "--unit-amount", "9.00",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Currencies == nil {
		t.Fatal("expected currencies to be set")
	}
	pricings := *capturedBody.Currencies
	if len(pricings) != 2 {
		t.Fatalf("expected 2 pricings, got %d", len(pricings))
	}
	if *pricings[0].Currency != "USD" {
		t.Errorf("expected first currency=USD, got %v", *pricings[0].Currency)
	}
	if *pricings[0].UnitAmount != 10.00 {
		t.Errorf("expected first unit-amount=10.00, got %v", *pricings[0].UnitAmount)
	}
	if *pricings[1].Currency != "EUR" {
		t.Errorf("expected second currency=EUR, got %v", *pricings[1].Currency)
	}
	if *pricings[1].UnitAmount != 9.00 {
		t.Errorf("expected second unit-amount=9.00, got %v", *pricings[1].UnitAmount)
	}
}

func TestItemsUpdate_CurrencyUnitAmountMismatch(t *testing.T) {
	mock := &mockItemAPI{
		updateItemFn: func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, stderr, err := executeCommand(app, "items", "update", "item-123",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "10.00",
	)
	if err == nil {
		t.Fatal("expected error for mismatched currency/unit-amount")
	}
	if !strings.Contains(stderr, "number of --currency values must match --unit-amount values") {
		t.Errorf("expected mismatch error, got %q", stderr)
	}
}

func TestItemsUpdate_UnsetFlagsNotSent(t *testing.T) {
	var capturedBody *recurly.ItemUpdate
	mock := &mockItemAPI{
		updateItemFn: func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
			capturedBody = body
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "update", "item-123", "--name", "Only Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Code != nil {
		t.Error("expected code to be nil when not set")
	}
	if capturedBody.Description != nil {
		t.Error("expected description to be nil when not set")
	}
	if capturedBody.ExternalSku != nil {
		t.Error("expected external-sku to be nil when not set")
	}
	if capturedBody.AccountingCode != nil {
		t.Error("expected accounting-code to be nil when not set")
	}
	if capturedBody.RevenueScheduleType != nil {
		t.Error("expected revenue-schedule-type to be nil when not set")
	}
	if capturedBody.TaxCode != nil {
		t.Error("expected tax-code to be nil when not set")
	}
	if capturedBody.TaxExempt != nil {
		t.Error("expected tax-exempt to be nil when not set")
	}
	if capturedBody.AvalaraTransactionType != nil {
		t.Error("expected avalara-transaction-type to be nil when not set")
	}
	if capturedBody.AvalaraServiceType != nil {
		t.Error("expected avalara-service-type to be nil when not set")
	}
	if capturedBody.HarmonizedSystemCode != nil {
		t.Error("expected harmonized-system-code to be nil when not set")
	}
	if capturedBody.Currencies != nil {
		t.Error("expected currencies to be nil when not set")
	}
}

func TestItemsUpdate_TableOutput(t *testing.T) {
	viper.Reset()
	mock := &mockItemAPI{
		updateItemFn: func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "update", "item-123", "--name", "Updated", "--output", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"Code", "widget-1",
		"Name", "Premium Widget",
		"Description", "A high-quality widget",
		"External SKU", "SKU-001",
		"State", "active",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestItemsUpdate_JSONOutput(t *testing.T) {
	viper.Reset()
	mock := &mockItemAPI{
		updateItemFn: func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "update", "item-123", "--name", "Updated", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["code"] != "widget-1" {
		t.Errorf("expected code=widget-1 in JSON, got %v", result["code"])
	}
}

func TestItemsUpdate_SDKError(t *testing.T) {
	mock := &mockItemAPI{
		updateItemFn: func(itemId string, body *recurly.ItemUpdate, opts ...recurly.Option) (*recurly.Item, error) {
			return nil, fmt.Errorf("validation failed")
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "update", "item-123", "--name", "bad")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- Deactivate tests ---

func TestItemsDeactivate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "deactivate") {
		t.Error("expected 'deactivate' in items help output")
	}
}

func TestItemsDeactivateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "deactivate", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--yes") {
		t.Error("expected --yes flag in help output")
	}
}

func TestItemsDeactivate_MissingArg_ReturnsError(t *testing.T) {
	_, _, err := executeCommand(nil, "items", "deactivate")
	if err == nil {
		t.Fatal("expected error for missing argument")
	}
}

func TestItemsDeactivate_NoAPIKey_WithYes_ReturnsError(t *testing.T) {
	viper.Set("api_key", "")
	defer viper.Set("api_key", "")

	_, _, err := executeCommand(nil, "items", "deactivate", "item-123", "--yes")
	if err == nil {
		t.Fatal("expected error when no API key is set")
	}
}

func TestItemsDeactivate_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "items", "deactivate", "item-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to deactivate item item-123? [y/N]") {
		t.Error("expected confirmation prompt with item ID in stderr")
	}
	if !strings.Contains(stderr, "Deactivation cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestItemsDeactivate_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "items", "deactivate", "item-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deactivation cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestItemsDeactivate_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockItemAPI{
		deactivateItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			capturedID = itemId
			item := sampleItemDetail()
			item.State = "inactive"
			return item, nil
		},
	}
	app := newTestItemApp(mock)

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "items", "deactivate", "item-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "item-789" {
		t.Errorf("expected item ID 'item-789', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "widget-1") {
		t.Errorf("expected item details in output, got:\n%s", out)
	}
}

func TestItemsDeactivate_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedID string
	mock := &mockItemAPI{
		deactivateItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			capturedID = itemId
			item := sampleItemDetail()
			item.State = "inactive"
			return item, nil
		},
	}
	app := newTestItemApp(mock)

	out, stderr, err := executeCommand(app, "items", "deactivate", "item-456", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "item-456" {
		t.Errorf("expected item ID 'item-456', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "widget-1") {
		t.Errorf("expected item details in output, got:\n%s", out)
	}
}

func TestItemsDeactivate_TableOutput(t *testing.T) {
	mock := &mockItemAPI{
		deactivateItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			item := sampleItemDetail()
			item.State = "inactive"
			return item, nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "deactivate", "item-123", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "inactive") {
		t.Errorf("expected 'inactive' state in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Premium Widget") {
		t.Errorf("expected item name in output, got:\n%s", out)
	}
}

func TestItemsDeactivate_JSONOutput(t *testing.T) {
	mock := &mockItemAPI{
		deactivateItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			item := sampleItemDetail()
			item.State = "inactive"
			return item, nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "deactivate", "item-123", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nOutput: %s", err, out)
	}
	if result["code"] != "widget-1" {
		t.Errorf("expected code=widget-1 in JSON, got %v", result["code"])
	}
}

func TestItemsDeactivate_SDKError(t *testing.T) {
	mock := &mockItemAPI{
		deactivateItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "deactivate", "item-123", "--yes")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- Reactivate command tests ---

func TestItemsReactivate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "reactivate") {
		t.Error("expected 'reactivate' in items help output")
	}
}

func TestItemsReactivateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "items", "reactivate", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--yes") {
		t.Error("expected --yes flag in help output")
	}
}

func TestItemsReactivate_MissingArg_ReturnsError(t *testing.T) {
	_, _, err := executeCommand(nil, "items", "reactivate")
	if err == nil {
		t.Fatal("expected error for missing argument")
	}
}

func TestItemsReactivate_NoAPIKey_WithYes_ReturnsError(t *testing.T) {
	viper.Set("api_key", "")
	defer viper.Set("api_key", "")

	_, _, err := executeCommand(nil, "items", "reactivate", "item-123", "--yes")
	if err == nil {
		t.Fatal("expected error when no API key is set")
	}
}

func TestItemsReactivate_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "items", "reactivate", "item-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to reactivate item item-123? [y/N]") {
		t.Error("expected confirmation prompt with item ID in stderr")
	}
	if !strings.Contains(stderr, "Reactivation cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestItemsReactivate_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "items", "reactivate", "item-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Reactivation cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestItemsReactivate_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockItemAPI{
		reactivateItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			capturedID = itemId
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "items", "reactivate", "item-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "item-789" {
		t.Errorf("expected item ID 'item-789', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "widget-1") {
		t.Errorf("expected item details in output, got:\n%s", out)
	}
}

func TestItemsReactivate_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedID string
	mock := &mockItemAPI{
		reactivateItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			capturedID = itemId
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, stderr, err := executeCommand(app, "items", "reactivate", "item-456", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "item-456" {
		t.Errorf("expected item ID 'item-456', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "widget-1") {
		t.Errorf("expected item details in output, got:\n%s", out)
	}
}

func TestItemsReactivate_TableOutput(t *testing.T) {
	mock := &mockItemAPI{
		reactivateItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "reactivate", "item-123", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "active") {
		t.Errorf("expected 'active' state in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Premium Widget") {
		t.Errorf("expected item name in output, got:\n%s", out)
	}
}

func TestItemsReactivate_JSONOutput(t *testing.T) {
	mock := &mockItemAPI{
		reactivateItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			return sampleItemDetail(), nil
		},
	}
	app := newTestItemApp(mock)

	out, _, err := executeCommand(app, "items", "reactivate", "item-123", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nOutput: %s", err, out)
	}
	if result["code"] != "widget-1" {
		t.Errorf("expected code=widget-1 in JSON, got %v", result["code"])
	}
}

func TestItemsReactivate_SDKError(t *testing.T) {
	mock := &mockItemAPI{
		reactivateItemFn: func(itemId string, opts ...recurly.Option) (*recurly.Item, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	app := newTestItemApp(mock)

	_, _, err := executeCommand(app, "items", "reactivate", "item-123", "--yes")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}
