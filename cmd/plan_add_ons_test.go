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

// mockPlanAddOnAPI implements PlanAddOnAPI for testing.
type mockPlanAddOnAPI struct {
	listPlanAddOnsFn func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error)
}

func (m *mockPlanAddOnAPI) ListPlanAddOns(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
	return m.listPlanAddOnsFn(planId, params, opts...)
}

func (m *mockPlanAddOnAPI) GetPlanAddOn(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
	return nil, nil
}

func (m *mockPlanAddOnAPI) CreatePlanAddOn(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
	return nil, nil
}

func (m *mockPlanAddOnAPI) UpdatePlanAddOn(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
	return nil, nil
}

func (m *mockPlanAddOnAPI) RemovePlanAddOn(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
	return nil, nil
}

// mockAddOnLister implements recurly.AddOnLister for testing.
type mockAddOnLister struct {
	addOns  []recurly.AddOn
	fetched bool
}

func (m *mockAddOnLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockAddOnLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockAddOnLister) Count() (*int64, error) {
	n := int64(len(m.addOns))
	return &n, nil
}

func (m *mockAddOnLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockAddOnLister) Data() []recurly.AddOn {
	return m.addOns
}

func (m *mockAddOnLister) HasMore() bool {
	return !m.fetched
}

func (m *mockAddOnLister) Next() string {
	return ""
}

func setMockPlanAddOnAPI(mock *mockPlanAddOnAPI) func() {
	orig := newPlanAddOnAPI
	newPlanAddOnAPI = func() (PlanAddOnAPI, error) {
		return mock, nil
	}
	return func() { newPlanAddOnAPI = orig }
}

func sampleAddOn() recurly.AddOn {
	now := time.Date(2025, 3, 10, 12, 0, 0, 0, time.UTC)
	return recurly.AddOn{
		Id:        "a1234",
		Code:      "extra-users",
		Name:      "Extra Users",
		State:     "active",
		AddOnType: "fixed",
		CreatedAt: &now,
	}
}

func TestPlanAddOnsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("plans", "add-ons", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected add-ons help to show 'list' subcommand")
	}
}

func TestPlanAddOnsListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand("plans", "add-ons", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--order", "--sort", "--state"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestPlanAddOnsList_RequiresPlanID(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			return &mockAddOnLister{addOns: []recurly.AddOn{}}, nil
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "add-ons", "list")
	if err == nil {
		t.Fatal("expected error when plan_id is missing")
	}
}

func TestPlanAddOnsList_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand("plans", "add-ons", "list", "code-plan1")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestPlanAddOnsList_PaginationParams(t *testing.T) {
	var capturedPlanID string
	var capturedParams *recurly.ListPlanAddOnsParams

	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			capturedPlanID = planId
			capturedParams = params
			return &mockAddOnLister{addOns: []recurly.AddOn{}}, nil
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "add-ons", "list", "code-plan1", "--limit", "50", "--order", "desc", "--sort", "updated_at")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPlanID != "code-plan1" {
		t.Errorf("expected planId=code-plan1, got %q", capturedPlanID)
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

func TestPlanAddOnsList_FilterParams(t *testing.T) {
	var capturedParams *recurly.ListPlanAddOnsParams

	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			capturedParams = params
			return &mockAddOnLister{addOns: []recurly.AddOn{}}, nil
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "add-ons", "list", "code-plan1", "--state", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.State == nil || *capturedParams.State != "active" {
		t.Errorf("expected state=active, got %v", capturedParams.State)
	}
}

func TestPlanAddOnsList_UnsetFlagsNotSent(t *testing.T) {
	var capturedParams *recurly.ListPlanAddOnsParams

	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			capturedParams = params
			return &mockAddOnLister{addOns: []recurly.AddOn{}}, nil
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "add-ons", "list", "code-plan1")
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
}

func TestPlanAddOnsList_TableOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			return &mockAddOnLister{addOns: []recurly.AddOn{addOn}}, nil
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "add-ons", "list", "code-plan1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"a1234", "extra-users", "Extra Users", "active", "fixed"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	for _, header := range []string{"ID", "Code", "Name", "State", "Add-On Type", "Created At"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestPlanAddOnsList_JSONOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			return &mockAddOnLister{addOns: []recurly.AddOn{addOn}}, nil
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("plans", "add-ons", "list", "code-plan1")
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
		t.Errorf("expected 1 item in JSON output, got %d", len(result.Data))
	}
}

func TestPlanAddOnsList_JSONPrettyOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			return &mockAddOnLister{addOns: []recurly.AddOn{addOn}}, nil
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	viper.Set("output", "json-pretty")
	defer viper.Reset()

	out, _, err := executeCommand("plans", "add-ons", "list", "code-plan1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "\n") {
		t.Error("expected pretty-printed JSON with newlines")
	}

	var result struct {
		Data    []json.RawMessage `json:"data"`
		HasMore bool              `json:"has_more"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
}

func TestPlanAddOnsList_JQFilter(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			return &mockAddOnLister{addOns: []recurly.AddOn{addOn}}, nil
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	viper.Set("jq", ".data[0].code")
	defer viper.Reset()

	out, _, err := executeCommand("plans", "add-ons", "list", "code-plan1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "extra-users") {
		t.Errorf("expected jq output to contain 'extra-users', got: %s", out)
	}
}

func TestPlanAddOnsList_SDKError(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "add-ons", "list", "code-plan1")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestPlanAddOnsList_EmptyResults(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			return &mockAddOnLister{addOns: []recurly.AddOn{}}, nil
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "add-ons", "list", "code-plan1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Table output should still have headers but no data rows
	if strings.Contains(out, "extra-users") {
		t.Error("expected no add-on data in empty results")
	}
}

func TestPlanAddOnsList_EmptyResults_JSON(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			return &mockAddOnLister{addOns: []recurly.AddOn{}}, nil
		},
	}
	cleanup := setMockPlanAddOnAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("plans", "add-ons", "list", "code-plan1")
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
		t.Errorf("expected empty data array, got %d items", len(result.Data))
	}
}
