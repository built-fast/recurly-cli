package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

// mockPlanAddOnAPI implements PlanAddOnAPI for testing.
type mockPlanAddOnAPI struct {
	listPlanAddOnsFn  func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error)
	getPlanAddOnFn    func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error)
	createPlanAddOnFn func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error)
	updatePlanAddOnFn func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error)
	removePlanAddOnFn func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error)
}

func (m *mockPlanAddOnAPI) ListPlanAddOns(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
	return m.listPlanAddOnsFn(planId, params, opts...)
}

func (m *mockPlanAddOnAPI) GetPlanAddOn(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
	if m.getPlanAddOnFn != nil {
		return m.getPlanAddOnFn(planId, addOnId, opts...)
	}
	return nil, nil
}

func (m *mockPlanAddOnAPI) CreatePlanAddOn(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
	if m.createPlanAddOnFn != nil {
		return m.createPlanAddOnFn(planId, body, opts...)
	}
	return nil, nil
}

func (m *mockPlanAddOnAPI) UpdatePlanAddOn(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
	if m.updatePlanAddOnFn != nil {
		return m.updatePlanAddOnFn(planId, addOnId, body, opts...)
	}
	return nil, nil
}

func (m *mockPlanAddOnAPI) RemovePlanAddOn(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
	if m.removePlanAddOnFn != nil {
		return m.removePlanAddOnFn(planId, addOnId, opts...)
	}
	return nil, nil
}

func sampleAddOn() recurly.AddOn {
	now := time.Date(2025, 3, 10, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2025, 3, 15, 14, 0, 0, 0, time.UTC)
	return recurly.AddOn{
		Id:              "a1234",
		Code:            "extra-users",
		Name:            "Extra Users",
		State:           "active",
		AddOnType:       "fixed",
		DefaultQuantity: 1,
		Optional:        true,
		AccountingCode:  "extra-users-ac",
		TaxCode:         "SW054000",
		Currencies: []recurly.AddOnPricing{
			{Currency: "USD", UnitAmount: 5.00},
			{Currency: "EUR", UnitAmount: 4.50},
		},
		CreatedAt: &now,
		UpdatedAt: &updated,
	}
}

func TestPlanAddOnsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "plans", "add-ons", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected add-ons help to show 'list' subcommand")
	}
}

func TestPlanAddOnsListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "plans", "add-ons", "list", "--help")
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
			return &mockLister[recurly.AddOn]{items: []recurly.AddOn{}}, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "list")
	if err == nil {
		t.Fatal("expected error when plan_id is missing")
	}
}

func TestPlanAddOnsList_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "plans", "add-ons", "list", "code-plan1")
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
			return &mockLister[recurly.AddOn]{items: []recurly.AddOn{}}, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "list", "code-plan1", "--limit", "50", "--order", "desc", "--sort", "updated_at")
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
			return &mockLister[recurly.AddOn]{items: []recurly.AddOn{}}, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "list", "code-plan1", "--state", "active")
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
			return &mockLister[recurly.AddOn]{items: []recurly.AddOn{}}, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "list", "code-plan1")
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
			return &mockLister[recurly.AddOn]{items: []recurly.AddOn{addOn}}, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	out, _, err := executeCommand(app, "plans", "add-ons", "list", "code-plan1")
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
			return &mockLister[recurly.AddOn]{items: []recurly.AddOn{addOn}}, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "list", "code-plan1")
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
			return &mockLister[recurly.AddOn]{items: []recurly.AddOn{addOn}}, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json-pretty")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "list", "code-plan1")
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
			return &mockLister[recurly.AddOn]{items: []recurly.AddOn{addOn}}, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json")
	viper.Set("jq", ".data[0].code")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "list", "code-plan1")
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
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "list", "code-plan1")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestPlanAddOnsList_EmptyResults(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		listPlanAddOnsFn: func(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error) {
			return &mockLister[recurly.AddOn]{items: []recurly.AddOn{}}, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	out, _, err := executeCommand(app, "plans", "add-ons", "list", "code-plan1")
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
			return &mockLister[recurly.AddOn]{items: []recurly.AddOn{}}, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "list", "code-plan1")
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

// --- plan add-ons get ---

func TestPlanAddOnsGet_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "plans", "add-ons", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "get") {
		t.Error("expected add-ons help to show 'get' subcommand")
	}
}

func TestPlanAddOnsGet_RequiresBothArgs(t *testing.T) {
	_, _, err := executeCommand(nil, "plans", "add-ons", "get")
	if err == nil {
		t.Fatal("expected error when no args provided")
	}

	_, _, err = executeCommand(nil, "plans", "add-ons", "get", "plan1")
	if err == nil {
		t.Fatal("expected error when only plan_id provided")
	}
}

func TestPlanAddOnsGet_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "plans", "add-ons", "get", "plan1", "addon1")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestPlanAddOnsGet_PositionalArgs(t *testing.T) {
	var capturedPlanID, capturedAddOnID string
	addOn := sampleAddOn()

	mock := &mockPlanAddOnAPI{
		getPlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedPlanID = planId
			capturedAddOnID = addOnId
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "get", "code-plan1", "extra-users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPlanID != "code-plan1" {
		t.Errorf("expected planId=code-plan1, got %q", capturedPlanID)
	}
	if capturedAddOnID != "extra-users" {
		t.Errorf("expected addOnId=extra-users, got %q", capturedAddOnID)
	}
}

func TestPlanAddOnsGet_TableOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		getPlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	out, _, err := executeCommand(app, "plans", "add-ons", "get", "plan1", "addon1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"ID", "a1234",
		"Code", "extra-users",
		"Name", "Extra Users",
		"State", "active",
		"Add-On Type", "fixed",
		"Default Quantity", "1",
		"Optional", "true",
		"Accounting Code", "extra-users-ac",
		"Tax Code", "SW054000",
		"Currencies", "USD: 5.00",
		"EUR: 4.50",
		"Created At", "2025-03-10T12:00:00Z",
		"Updated At", "2025-03-15T14:00:00Z",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestPlanAddOnsGet_JSONOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		getPlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "get", "plan1", "addon1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\noutput: %s", err, out)
	}
	if result["code"] != "extra-users" {
		t.Errorf("expected code=extra-users in JSON, got %v", result["code"])
	}
	if result["id"] != "a1234" {
		t.Errorf("expected id=a1234 in JSON, got %v", result["id"])
	}
}

func TestPlanAddOnsGet_JSONPrettyOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		getPlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json-pretty")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "get", "plan1", "addon1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "\n") {
		t.Error("expected pretty-printed JSON with newlines")
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
}

func TestPlanAddOnsGet_JQFilter(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		getPlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json")
	viper.Set("jq", ".code")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "get", "plan1", "addon1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "extra-users") {
		t.Errorf("expected jq output to contain 'extra-users', got: %s", out)
	}
}

func TestPlanAddOnsGet_SDKError(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		getPlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "get", "plan1", "addon1")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- plan add-ons create ---

func TestPlanAddOnsCreate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "plans", "add-ons", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "create") {
		t.Error("expected add-ons help to show 'create' subcommand")
	}
}

func TestPlanAddOnsCreateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "plans", "add-ons", "create", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{
		"--code", "--name", "--add-on-type", "--default-quantity", "--optional",
		"--display-quantity", "--accounting-code", "--tax-code", "--revenue-schedule-type",
		"--usage-type", "--usage-calculation-type", "--measured-unit-id",
		"--currency", "--unit-amount",
	} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestPlanAddOnsCreate_RequiresPlanID(t *testing.T) {
	_, _, err := executeCommand(nil, "plans", "add-ons", "create")
	if err == nil {
		t.Fatal("expected error when plan_id is missing")
	}
}

func TestPlanAddOnsCreate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "plans", "add-ons", "create", "plan1", "--code", "test", "--name", "Test")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestPlanAddOnsCreate_CoreFlags(t *testing.T) {
	var capturedPlanID string
	var capturedBody *recurly.AddOnCreate
	addOn := sampleAddOn()

	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedPlanID = planId
			capturedBody = body
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "create", "code-plan1",
		"--code", "extra-users",
		"--name", "Extra Users",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPlanID != "code-plan1" {
		t.Errorf("expected planId=code-plan1, got %q", capturedPlanID)
	}
	if capturedBody.Code == nil || *capturedBody.Code != "extra-users" {
		t.Errorf("expected code=extra-users, got %v", capturedBody.Code)
	}
	if capturedBody.Name == nil || *capturedBody.Name != "Extra Users" {
		t.Errorf("expected name=Extra Users, got %v", capturedBody.Name)
	}
}

func TestPlanAddOnsCreate_AllOptionalFlags(t *testing.T) {
	var capturedBody *recurly.AddOnCreate
	addOn := sampleAddOn()

	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedBody = body
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "create", "plan1",
		"--code", "extra-users",
		"--name", "Extra Users",
		"--add-on-type", "usage",
		"--default-quantity", "5",
		"--optional",
		"--display-quantity",
		"--accounting-code", "ac-123",
		"--tax-code", "SW054000",
		"--revenue-schedule-type", "evenly",
		"--usage-type", "price",
		"--usage-calculation-type", "cumulative",
		"--measured-unit-id", "mu-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.AddOnType == nil || *capturedBody.AddOnType != "usage" {
		t.Errorf("expected add-on-type=usage, got %v", capturedBody.AddOnType)
	}
	if capturedBody.DefaultQuantity == nil || *capturedBody.DefaultQuantity != 5 {
		t.Errorf("expected default-quantity=5, got %v", capturedBody.DefaultQuantity)
	}
	if capturedBody.Optional == nil || !*capturedBody.Optional {
		t.Error("expected optional=true")
	}
	if capturedBody.DisplayQuantity == nil || !*capturedBody.DisplayQuantity {
		t.Error("expected display-quantity=true")
	}
	if capturedBody.AccountingCode == nil || *capturedBody.AccountingCode != "ac-123" {
		t.Errorf("expected accounting-code=ac-123, got %v", capturedBody.AccountingCode)
	}
	if capturedBody.TaxCode == nil || *capturedBody.TaxCode != "SW054000" {
		t.Errorf("expected tax-code=SW054000, got %v", capturedBody.TaxCode)
	}
	if capturedBody.RevenueScheduleType == nil || *capturedBody.RevenueScheduleType != "evenly" {
		t.Errorf("expected revenue-schedule-type=evenly, got %v", capturedBody.RevenueScheduleType)
	}
	if capturedBody.UsageType == nil || *capturedBody.UsageType != "price" {
		t.Errorf("expected usage-type=price, got %v", capturedBody.UsageType)
	}
	if capturedBody.UsageCalculationType == nil || *capturedBody.UsageCalculationType != "cumulative" {
		t.Errorf("expected usage-calculation-type=cumulative, got %v", capturedBody.UsageCalculationType)
	}
	if capturedBody.MeasuredUnitId == nil || *capturedBody.MeasuredUnitId != "mu-123" {
		t.Errorf("expected measured-unit-id=mu-123, got %v", capturedBody.MeasuredUnitId)
	}
}

func TestPlanAddOnsCreate_MultiCurrencyFlags(t *testing.T) {
	var capturedBody *recurly.AddOnCreate
	addOn := sampleAddOn()

	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedBody = body
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "create", "plan1",
		"--code", "extra-users",
		"--name", "Extra Users",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "5.00", "--unit-amount", "4.50",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Currencies == nil {
		t.Fatal("expected currencies to be set")
	}
	pricings := *capturedBody.Currencies
	if len(pricings) != 2 {
		t.Fatalf("expected 2 currencies, got %d", len(pricings))
	}
	if *pricings[0].Currency != "USD" || *pricings[0].UnitAmount != 5.00 {
		t.Errorf("expected USD: 5.00, got %s: %v", *pricings[0].Currency, *pricings[0].UnitAmount)
	}
	if *pricings[1].Currency != "EUR" || *pricings[1].UnitAmount != 4.50 {
		t.Errorf("expected EUR: 4.50, got %s: %v", *pricings[1].Currency, *pricings[1].UnitAmount)
	}
}

func TestPlanAddOnsCreate_SingleCurrency(t *testing.T) {
	var capturedBody *recurly.AddOnCreate
	addOn := sampleAddOn()

	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedBody = body
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "create", "plan1",
		"--code", "extra-users",
		"--name", "Extra Users",
		"--currency", "USD",
		"--unit-amount", "10.00",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Currencies == nil {
		t.Fatal("expected currencies to be set")
	}
	pricings := *capturedBody.Currencies
	if len(pricings) != 1 {
		t.Fatalf("expected 1 currency, got %d", len(pricings))
	}
	if *pricings[0].Currency != "USD" || *pricings[0].UnitAmount != 10.00 {
		t.Errorf("expected USD: 10.00, got %s: %v", *pricings[0].Currency, *pricings[0].UnitAmount)
	}
}

func TestPlanAddOnsCreate_CurrencyUnitAmountMismatch(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "create", "plan1",
		"--code", "test",
		"--name", "Test",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "5.00",
	)
	if err == nil {
		t.Fatal("expected error for mismatched currency/unit-amount counts")
	}
	if !strings.Contains(err.Error(), "must match") {
		t.Errorf("expected mismatch error, got: %v", err)
	}
}

func TestPlanAddOnsCreate_UnsetFlagsNotSent(t *testing.T) {
	var capturedBody *recurly.AddOnCreate
	addOn := sampleAddOn()

	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedBody = body
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "create", "plan1",
		"--code", "test",
		"--name", "Test",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.AddOnType != nil {
		t.Error("expected add-on-type to be nil when not set")
	}
	if capturedBody.DefaultQuantity != nil {
		t.Error("expected default-quantity to be nil when not set")
	}
	if capturedBody.Optional != nil {
		t.Error("expected optional to be nil when not set")
	}
	if capturedBody.DisplayQuantity != nil {
		t.Error("expected display-quantity to be nil when not set")
	}
	if capturedBody.AccountingCode != nil {
		t.Error("expected accounting-code to be nil when not set")
	}
	if capturedBody.TaxCode != nil {
		t.Error("expected tax-code to be nil when not set")
	}
	if capturedBody.RevenueScheduleType != nil {
		t.Error("expected revenue-schedule-type to be nil when not set")
	}
	if capturedBody.UsageType != nil {
		t.Error("expected usage-type to be nil when not set")
	}
	if capturedBody.UsageCalculationType != nil {
		t.Error("expected usage-calculation-type to be nil when not set")
	}
	if capturedBody.MeasuredUnitId != nil {
		t.Error("expected measured-unit-id to be nil when not set")
	}
	if capturedBody.Currencies != nil {
		t.Error("expected currencies to be nil when not set")
	}
}

func TestPlanAddOnsCreate_TableOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	out, _, err := executeCommand(app, "plans", "add-ons", "create", "plan1",
		"--code", "extra-users",
		"--name", "Extra Users",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"ID", "a1234",
		"Code", "extra-users",
		"Name", "Extra Users",
		"State", "active",
		"Add-On Type", "fixed",
		"USD: 5.00",
		"EUR: 4.50",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestPlanAddOnsCreate_JSONOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "create", "plan1",
		"--code", "extra-users",
		"--name", "Extra Users",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\noutput: %s", err, out)
	}
	if result["code"] != "extra-users" {
		t.Errorf("expected code=extra-users in JSON, got %v", result["code"])
	}
}

func TestPlanAddOnsCreate_JSONPrettyOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json-pretty")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "create", "plan1",
		"--code", "extra-users",
		"--name", "Extra Users",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "\n") {
		t.Error("expected pretty-printed JSON with newlines")
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
}

func TestPlanAddOnsCreate_JQFilter(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json")
	viper.Set("jq", ".code")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "create", "plan1",
		"--code", "extra-users",
		"--name", "Extra Users",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "extra-users") {
		t.Errorf("expected jq output to contain 'extra-users', got: %s", out)
	}
}

func TestPlanAddOnsCreate_SDKError(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		createPlanAddOnFn: func(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return nil, &recurly.Error{Message: "validation error"}
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "create", "plan1",
		"--code", "test",
		"--name", "Test",
	)
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- plan add-ons update ---

func TestPlanAddOnsUpdate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "plans", "add-ons", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "update") {
		t.Error("expected add-ons help to show 'update' subcommand")
	}
}

func TestPlanAddOnsUpdateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "plans", "add-ons", "update", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{
		"--code", "--name", "--default-quantity", "--optional",
		"--display-quantity", "--accounting-code", "--tax-code", "--revenue-schedule-type",
		"--usage-calculation-type", "--measured-unit-id", "--measured-unit-name",
		"--currency", "--unit-amount",
	} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestPlanAddOnsUpdate_RequiresBothArgs(t *testing.T) {
	_, _, err := executeCommand(nil, "plans", "add-ons", "update")
	if err == nil {
		t.Fatal("expected error when no args provided")
	}

	_, _, err = executeCommand(nil, "plans", "add-ons", "update", "plan1")
	if err == nil {
		t.Fatal("expected error when only plan_id provided")
	}
}

func TestPlanAddOnsUpdate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "plans", "add-ons", "update", "plan1", "addon1", "--name", "Updated")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestPlanAddOnsUpdate_PositionalArgs(t *testing.T) {
	var capturedPlanID, capturedAddOnID string
	addOn := sampleAddOn()

	mock := &mockPlanAddOnAPI{
		updatePlanAddOnFn: func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedPlanID = planId
			capturedAddOnID = addOnId
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "update", "code-plan1", "extra-users", "--name", "Updated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPlanID != "code-plan1" {
		t.Errorf("expected planId=code-plan1, got %q", capturedPlanID)
	}
	if capturedAddOnID != "extra-users" {
		t.Errorf("expected addOnId=extra-users, got %q", capturedAddOnID)
	}
}

func TestPlanAddOnsUpdate_AllOptionalFlags(t *testing.T) {
	var capturedBody *recurly.AddOnUpdate
	addOn := sampleAddOn()

	mock := &mockPlanAddOnAPI{
		updatePlanAddOnFn: func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedBody = body
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "update", "plan1", "addon1",
		"--code", "new-code",
		"--name", "New Name",
		"--default-quantity", "5",
		"--optional",
		"--display-quantity",
		"--accounting-code", "ac-123",
		"--tax-code", "SW054000",
		"--revenue-schedule-type", "evenly",
		"--usage-calculation-type", "cumulative",
		"--measured-unit-id", "mu-123",
		"--measured-unit-name", "API Calls",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Code == nil || *capturedBody.Code != "new-code" {
		t.Errorf("expected code=new-code, got %v", capturedBody.Code)
	}
	if capturedBody.Name == nil || *capturedBody.Name != "New Name" {
		t.Errorf("expected name=New Name, got %v", capturedBody.Name)
	}
	if capturedBody.DefaultQuantity == nil || *capturedBody.DefaultQuantity != 5 {
		t.Errorf("expected default-quantity=5, got %v", capturedBody.DefaultQuantity)
	}
	if capturedBody.Optional == nil || !*capturedBody.Optional {
		t.Error("expected optional=true")
	}
	if capturedBody.DisplayQuantity == nil || !*capturedBody.DisplayQuantity {
		t.Error("expected display-quantity=true")
	}
	if capturedBody.AccountingCode == nil || *capturedBody.AccountingCode != "ac-123" {
		t.Errorf("expected accounting-code=ac-123, got %v", capturedBody.AccountingCode)
	}
	if capturedBody.TaxCode == nil || *capturedBody.TaxCode != "SW054000" {
		t.Errorf("expected tax-code=SW054000, got %v", capturedBody.TaxCode)
	}
	if capturedBody.RevenueScheduleType == nil || *capturedBody.RevenueScheduleType != "evenly" {
		t.Errorf("expected revenue-schedule-type=evenly, got %v", capturedBody.RevenueScheduleType)
	}
	if capturedBody.UsageCalculationType == nil || *capturedBody.UsageCalculationType != "cumulative" {
		t.Errorf("expected usage-calculation-type=cumulative, got %v", capturedBody.UsageCalculationType)
	}
	if capturedBody.MeasuredUnitId == nil || *capturedBody.MeasuredUnitId != "mu-123" {
		t.Errorf("expected measured-unit-id=mu-123, got %v", capturedBody.MeasuredUnitId)
	}
	if capturedBody.MeasuredUnitName == nil || *capturedBody.MeasuredUnitName != "API Calls" {
		t.Errorf("expected measured-unit-name=API Calls, got %v", capturedBody.MeasuredUnitName)
	}
}

func TestPlanAddOnsUpdate_MultiCurrencyFlags(t *testing.T) {
	var capturedBody *recurly.AddOnUpdate
	addOn := sampleAddOn()

	mock := &mockPlanAddOnAPI{
		updatePlanAddOnFn: func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedBody = body
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "update", "plan1", "addon1",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "5.00", "--unit-amount", "4.50",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Currencies == nil {
		t.Fatal("expected currencies to be set")
	}
	pricings := *capturedBody.Currencies
	if len(pricings) != 2 {
		t.Fatalf("expected 2 currencies, got %d", len(pricings))
	}
	if *pricings[0].Currency != "USD" || *pricings[0].UnitAmount != 5.00 {
		t.Errorf("expected USD: 5.00, got %s: %v", *pricings[0].Currency, *pricings[0].UnitAmount)
	}
	if *pricings[1].Currency != "EUR" || *pricings[1].UnitAmount != 4.50 {
		t.Errorf("expected EUR: 4.50, got %s: %v", *pricings[1].Currency, *pricings[1].UnitAmount)
	}
}

func TestPlanAddOnsUpdate_CurrencyUnitAmountMismatch(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		updatePlanAddOnFn: func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "update", "plan1", "addon1",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "5.00",
	)
	if err == nil {
		t.Fatal("expected error for mismatched currency/unit-amount counts")
	}
	if !strings.Contains(err.Error(), "must match") {
		t.Errorf("expected mismatch error, got: %v", err)
	}
}

func TestPlanAddOnsUpdate_UnsetFlagsNotSent(t *testing.T) {
	var capturedBody *recurly.AddOnUpdate
	addOn := sampleAddOn()

	mock := &mockPlanAddOnAPI{
		updatePlanAddOnFn: func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedBody = body
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "update", "plan1", "addon1",
		"--name", "Updated Name",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Name == nil || *capturedBody.Name != "Updated Name" {
		t.Errorf("expected name=Updated Name, got %v", capturedBody.Name)
	}
	if capturedBody.Code != nil {
		t.Error("expected code to be nil when not set")
	}
	if capturedBody.DefaultQuantity != nil {
		t.Error("expected default-quantity to be nil when not set")
	}
	if capturedBody.Optional != nil {
		t.Error("expected optional to be nil when not set")
	}
	if capturedBody.DisplayQuantity != nil {
		t.Error("expected display-quantity to be nil when not set")
	}
	if capturedBody.AccountingCode != nil {
		t.Error("expected accounting-code to be nil when not set")
	}
	if capturedBody.TaxCode != nil {
		t.Error("expected tax-code to be nil when not set")
	}
	if capturedBody.RevenueScheduleType != nil {
		t.Error("expected revenue-schedule-type to be nil when not set")
	}
	if capturedBody.UsageCalculationType != nil {
		t.Error("expected usage-calculation-type to be nil when not set")
	}
	if capturedBody.MeasuredUnitId != nil {
		t.Error("expected measured-unit-id to be nil when not set")
	}
	if capturedBody.MeasuredUnitName != nil {
		t.Error("expected measured-unit-name to be nil when not set")
	}
	if capturedBody.Currencies != nil {
		t.Error("expected currencies to be nil when not set")
	}
}

func TestPlanAddOnsUpdate_TableOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		updatePlanAddOnFn: func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	out, _, err := executeCommand(app, "plans", "add-ons", "update", "plan1", "addon1",
		"--name", "Extra Users",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"ID", "a1234",
		"Code", "extra-users",
		"Name", "Extra Users",
		"State", "active",
		"Add-On Type", "fixed",
		"USD: 5.00",
		"EUR: 4.50",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestPlanAddOnsUpdate_JSONOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		updatePlanAddOnFn: func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "update", "plan1", "addon1",
		"--name", "Extra Users",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\noutput: %s", err, out)
	}
	if result["code"] != "extra-users" {
		t.Errorf("expected code=extra-users in JSON, got %v", result["code"])
	}
}

func TestPlanAddOnsUpdate_JSONPrettyOutput(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		updatePlanAddOnFn: func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json-pretty")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "update", "plan1", "addon1",
		"--name", "Extra Users",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "\n") {
		t.Error("expected pretty-printed JSON with newlines")
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
}

func TestPlanAddOnsUpdate_JQFilter(t *testing.T) {
	addOn := sampleAddOn()
	mock := &mockPlanAddOnAPI{
		updatePlanAddOnFn: func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return &addOn, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	viper.Set("output", "json")
	viper.Set("jq", ".code")
	defer viper.Reset()

	out, _, err := executeCommand(app, "plans", "add-ons", "update", "plan1", "addon1",
		"--name", "Extra Users",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "extra-users") {
		t.Errorf("expected jq output to contain 'extra-users', got: %s", out)
	}
}

func TestPlanAddOnsUpdate_SDKError(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		updatePlanAddOnFn: func(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error) {
			return nil, &recurly.Error{Message: "validation error"}
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "update", "plan1", "addon1",
		"--name", "Test",
	)
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- Delete tests ---

func TestPlanAddOnsDelete_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "plans", "add-ons", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "delete") {
		t.Error("expected 'delete' in add-ons help output")
	}
}

func TestPlanAddOnsDeleteHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "plans", "add-ons", "delete", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--yes") {
		t.Error("expected --yes flag in help output")
	}
}

func TestPlanAddOnsDelete_MissingArgs_ReturnsError(t *testing.T) {
	_, _, err := executeCommand(nil, "plans", "add-ons", "delete")
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}

	_, _, err = executeCommand(nil, "plans", "add-ons", "delete", "plan1")
	if err == nil {
		t.Fatal("expected error for missing add_on_id argument")
	}
}

func TestPlanAddOnsDelete_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "plans", "add-ons", "delete", "plan-123", "addon-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Delete add-on addon-456 from plan plan-123? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Deletion cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestPlanAddOnsDelete_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "plans", "add-ons", "delete", "plan-123", "addon-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestPlanAddOnsDelete_ConfirmYes_Succeeds(t *testing.T) {
	var capturedPlanID, capturedAddOnID string
	mock := &mockPlanAddOnAPI{
		removePlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedPlanID = planId
			capturedAddOnID = addOnId
			a := sampleAddOn()
			a.State = "inactive"
			return &a, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "plans", "add-ons", "delete", "plan-789", "addon-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPlanID != "plan-789" {
		t.Errorf("expected plan ID 'plan-789', got %q", capturedPlanID)
	}
	if capturedAddOnID != "addon-123" {
		t.Errorf("expected add-on ID 'addon-123', got %q", capturedAddOnID)
	}
	if !strings.Contains(stderr, "Delete add-on") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "extra-users") {
		t.Errorf("expected add-on details in output, got:\n%s", out)
	}
}

func TestPlanAddOnsDelete_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedPlanID, capturedAddOnID string
	mock := &mockPlanAddOnAPI{
		removePlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			capturedPlanID = planId
			capturedAddOnID = addOnId
			a := sampleAddOn()
			a.State = "inactive"
			return &a, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	out, stderr, err := executeCommand(app, "plans", "add-ons", "delete", "plan-456", "addon-789", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPlanID != "plan-456" {
		t.Errorf("expected plan ID 'plan-456', got %q", capturedPlanID)
	}
	if capturedAddOnID != "addon-789" {
		t.Errorf("expected add-on ID 'addon-789', got %q", capturedAddOnID)
	}
	if strings.Contains(stderr, "Delete add-on") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "extra-users") {
		t.Errorf("expected add-on details in output, got:\n%s", out)
	}
}

func TestPlanAddOnsDelete_DetailView(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		removePlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			a := sampleAddOn()
			return &a, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	out, _, err := executeCommand(app, "plans", "add-ons", "delete", "plan1", "addon1", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, expected := range []string{"a1234", "extra-users", "Extra Users", "active", "fixed"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected detail output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestPlanAddOnsDelete_JSONOutput(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		removePlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			a := sampleAddOn()
			return &a, nil
		},
	}
	app := newTestPlanAddOnApp(mock)

	out, _, err := executeCommand(app, "plans", "add-ons", "delete", "plan1", "addon1", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nOutput: %s", err, out)
	}
	if result["code"] != "extra-users" {
		t.Errorf("expected code 'extra-users', got %v", result["code"])
	}
}

func TestPlanAddOnsDelete_SDKError(t *testing.T) {
	mock := &mockPlanAddOnAPI{
		removePlanAddOnFn: func(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	app := newTestPlanAddOnApp(mock)

	_, _, err := executeCommand(app, "plans", "add-ons", "delete", "plan1", "addon1", "--yes")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}
