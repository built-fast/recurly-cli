package cmd

import (
	"context"
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

// mockItemLister implements recurly.ItemLister for testing.
type mockItemLister struct {
	items   []recurly.Item
	fetched bool
}

func (m *mockItemLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockItemLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockItemLister) Count() (*int64, error) {
	n := int64(len(m.items))
	return &n, nil
}

func (m *mockItemLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockItemLister) Data() []recurly.Item {
	return m.items
}

func (m *mockItemLister) HasMore() bool {
	return !m.fetched
}

func (m *mockItemLister) Next() string {
	return ""
}

// setMockItemAPI installs a mock and returns a cleanup function.
func setMockItemAPI(mock *mockItemAPI) func() {
	orig := newItemAPI
	newItemAPI = func() (ItemAPI, error) {
		return mock, nil
	}
	return func() { newItemAPI = orig }
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

// --- items list ---

func TestItemsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("items", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected items help to show 'list' subcommand")
	}
}

func TestItemsListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand("items", "list", "--help")
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
	_, stderr, err := executeCommand("items", "list")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestItemsList_InvalidBeginTime_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand("items", "list", "--begin-time", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid begin-time")
	}
	if !strings.Contains(stderr, "invalid --begin-time") {
		t.Errorf("expected 'invalid --begin-time' error, got %q", stderr)
	}
}

func TestItemsList_InvalidEndTime_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand("items", "list", "--end-time", "not-a-date")
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
			return &mockItemLister{items: []recurly.Item{}}, nil
		},
	}
	cleanup := setMockItemAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("items", "list", "--limit", "50", "--order", "desc", "--sort", "updated_at")
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
			return &mockItemLister{items: []recurly.Item{}}, nil
		},
	}
	cleanup := setMockItemAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("items", "list",
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
			return &mockItemLister{items: []recurly.Item{}}, nil
		},
	}
	cleanup := setMockItemAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("items", "list")
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
			return &mockItemLister{items: []recurly.Item{item}}, nil
		},
	}
	cleanup := setMockItemAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("items", "list")
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
			return &mockItemLister{items: []recurly.Item{item}}, nil
		},
	}
	cleanup := setMockItemAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("items", "list", "--output", "json")
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
			return &mockItemLister{items: []recurly.Item{item}}, nil
		},
	}
	cleanup := setMockItemAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("items", "list", "--output", "json-pretty")
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
			return &mockItemLister{items: []recurly.Item{item}}, nil
		},
	}
	cleanup := setMockItemAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("items", "list", "--jq", ".data[].code")
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
	cleanup := setMockItemAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("items", "list")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestItemsList_EmptyResults(t *testing.T) {
	mock := &mockItemAPI{
		listItemsFn: func(params *recurly.ListItemsParams, opts ...recurly.Option) (recurly.ItemLister, error) {
			return &mockItemLister{items: []recurly.Item{}}, nil
		},
	}
	cleanup := setMockItemAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("items", "list")
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
			return &mockItemLister{items: []recurly.Item{}}, nil
		},
	}
	cleanup := setMockItemAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("items", "list", "--output", "json")
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
