package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

// mockPlanAPI implements PlanAPI for testing.
type mockPlanAPI struct {
	listPlansFn  func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error)
	getPlanFn    func(planId string, opts ...recurly.Option) (*recurly.Plan, error)
	createPlanFn func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error)
	updatePlanFn func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error)
	removePlanFn func(planId string, opts ...recurly.Option) (*recurly.Plan, error)
}

func (m *mockPlanAPI) ListPlans(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
	return m.listPlansFn(params, opts...)
}

func (m *mockPlanAPI) GetPlan(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
	return m.getPlanFn(planId, opts...)
}

func (m *mockPlanAPI) CreatePlan(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
	return m.createPlanFn(body, opts...)
}

func (m *mockPlanAPI) UpdatePlan(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
	return m.updatePlanFn(planId, body, opts...)
}

func (m *mockPlanAPI) RemovePlan(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
	return m.removePlanFn(planId, opts...)
}

// mockPlanLister implements recurly.PlanLister for testing.
type mockPlanLister struct {
	plans   []recurly.Plan
	fetched bool
}

func (m *mockPlanLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockPlanLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockPlanLister) Count() (*int64, error) {
	n := int64(len(m.plans))
	return &n, nil
}

func (m *mockPlanLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockPlanLister) Data() []recurly.Plan {
	return m.plans
}

func (m *mockPlanLister) HasMore() bool {
	return !m.fetched
}

func (m *mockPlanLister) Next() string {
	return ""
}

// setMockPlanAPI installs a mock and returns a cleanup function.
func setMockPlanAPI(mock *mockPlanAPI) func() {
	orig := newPlanAPI
	newPlanAPI = func() (PlanAPI, error) {
		return mock, nil
	}
	return func() { newPlanAPI = orig }
}

// samplePlan returns a test plan with predictable fields.
func samplePlan() recurly.Plan {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return recurly.Plan{
		Code:           "gold",
		Name:           "Gold Plan",
		State:          "active",
		IntervalLength: 1,
		IntervalUnit:   "month",
		Currencies: []recurly.PlanPricing{
			{Currency: "USD", UnitAmount: 10.00},
		},
		CreatedAt: &now,
	}
}

// samplePlanDetail returns a test plan pointer with all detail fields populated.
func samplePlanDetail() *recurly.Plan {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC)
	return &recurly.Plan{
		Id:             "p1234",
		Code:           "gold",
		Name:           "Gold Plan",
		State:          "active",
		PricingModel:   "fixed",
		IntervalUnit:   "month",
		IntervalLength: 1,
		Description:    "A premium plan",
		Currencies: []recurly.PlanPricing{
			{Currency: "USD", UnitAmount: 10.00, SetupFee: 5.00},
			{Currency: "EUR", UnitAmount: 9.00, SetupFee: 4.50},
		},
		TrialUnit:          "day",
		TrialLength:        14,
		AutoRenew:          true,
		TotalBillingCycles: 12,
		TaxCode:            "digital",
		TaxExempt:          false,
		CreatedAt:          &now,
		UpdatedAt:          &updated,
	}
}

// --- plans list ---

func TestPlansList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("plans", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected plans help to show 'list' subcommand")
	}
}

func TestPlansListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand("plans", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--order", "--sort", "--state"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestPlansList_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand("plans", "list")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestPlansList_PaginationParams(t *testing.T) {
	var capturedParams *recurly.ListPlansParams

	mock := &mockPlanAPI{
		listPlansFn: func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
			capturedParams = params
			return &mockPlanLister{plans: []recurly.Plan{}}, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "list", "--limit", "50", "--order", "desc", "--sort", "updated_at")
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

func TestPlansList_FilterParams(t *testing.T) {
	var capturedParams *recurly.ListPlansParams

	mock := &mockPlanAPI{
		listPlansFn: func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
			capturedParams = params
			return &mockPlanLister{plans: []recurly.Plan{}}, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "list", "--state", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.State == nil || *capturedParams.State != "active" {
		t.Errorf("expected state=active, got %v", capturedParams.State)
	}
}

func TestPlansList_UnsetFlagsNotSent(t *testing.T) {
	var capturedParams *recurly.ListPlansParams

	mock := &mockPlanAPI{
		listPlansFn: func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
			capturedParams = params
			return &mockPlanLister{plans: []recurly.Plan{}}, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "list")
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

func TestPlansList_TableOutput(t *testing.T) {
	plan := samplePlan()
	mock := &mockPlanAPI{
		listPlansFn: func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
			return &mockPlanLister{plans: []recurly.Plan{plan}}, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"gold", "Gold Plan", "active", "1 month", "10.00 USD"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	// Table should have column headers
	for _, header := range []string{"Code", "Name", "State", "Interval", "Price", "Created At"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestPlansList_JSONOutput(t *testing.T) {
	plan := samplePlan()
	mock := &mockPlanAPI{
		listPlansFn: func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
			return &mockPlanLister{plans: []recurly.Plan{plan}}, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "list", "--output", "json")
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
		t.Fatalf("expected 1 plan in JSON output, got %d", len(envelope.Data))
	}
	if envelope.Data[0]["code"] != "gold" {
		t.Errorf("expected code=gold in JSON, got %v", envelope.Data[0]["code"])
	}
}

func TestPlansList_JSONPrettyOutput(t *testing.T) {
	plan := samplePlan()
	mock := &mockPlanAPI{
		listPlansFn: func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
			return &mockPlanLister{plans: []recurly.Plan{plan}}, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "list", "--output", "json-pretty")
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

func TestPlansList_JQFilter(t *testing.T) {
	plan := samplePlan()
	mock := &mockPlanAPI{
		listPlansFn: func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
			return &mockPlanLister{plans: []recurly.Plan{plan}}, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "list", "--jq", ".data[].code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if !strings.Contains(trimmed, "gold") {
		t.Errorf("expected jq output to contain 'gold', got: %s", trimmed)
	}
}

func TestPlansList_SDKError(t *testing.T) {
	mock := &mockPlanAPI{
		listPlansFn: func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "list")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestPlansList_EmptyResults(t *testing.T) {
	mock := &mockPlanAPI{
		listPlansFn: func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
			return &mockPlanLister{plans: []recurly.Plan{}}, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Table output should still have headers but no data rows
	if strings.Contains(out, "gold") {
		t.Error("expected no plan data in empty results")
	}
}

func TestPlansList_EmptyResults_JSON(t *testing.T) {
	mock := &mockPlanAPI{
		listPlansFn: func(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error) {
			return &mockPlanLister{plans: []recurly.Plan{}}, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "list", "--output", "json")
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
	if envelope.Object != "list" {
		t.Errorf("expected object=list, got %s", envelope.Object)
	}
	if len(envelope.Data) != 0 {
		t.Errorf("expected empty data array, got %d items", len(envelope.Data))
	}
}

// --- plans get ---

func TestPlansGet_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("plans", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "get") {
		t.Error("expected plans help to show 'get' subcommand")
	}
}

func TestPlansGet_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand("plans", "get")
	if err == nil {
		t.Fatal("expected error when no plan ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestPlansGet_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand("plans", "get", "p1234")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestPlansGet_PositionalArg(t *testing.T) {
	var capturedID string
	mock := &mockPlanAPI{
		getPlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedID = planId
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "get", "my-plan-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "my-plan-id" {
		t.Errorf("expected plan ID 'my-plan-id', got %q", capturedID)
	}
}

func TestPlansGet_TableOutput(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		getPlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "get", "p1234", "--output", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"Id", "p1234",
		"Code", "gold",
		"Name", "Gold Plan",
		"State", "active",
		"Pricing Model", "fixed",
		"Interval Unit", "month",
		"Interval Length", "1",
		"Description", "A premium plan",
		"Currencies", "USD: 10.00 (setup: 5.00)",
		"EUR: 9.00 (setup: 4.50)",
		"Trial Unit", "day",
		"Trial Length", "14",
		"Auto Renew", "true",
		"Total Billing Cycles", "12",
		"Tax Code", "digital",
		"Tax Exempt", "false",
		"Created At",
		"Updated At",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestPlansGet_JSONOutput(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		getPlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "get", "p1234", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["code"] != "gold" {
		t.Errorf("expected code=gold in JSON, got %v", result["code"])
	}
	// JSON output should be bare object, no envelope
	if _, ok := result["object"]; ok {
		// object field may exist from the Plan struct but should not be "list"
		if result["object"] == "list" {
			t.Error("expected single plan object, not list envelope")
		}
	}
}

func TestPlansGet_JQFilter(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		getPlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "get", "p1234", "--jq", ".code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "gold" {
		t.Errorf("expected jq output 'gold', got %q", trimmed)
	}
}

func TestPlansGet_SDKError(t *testing.T) {
	mock := &mockPlanAPI{
		getPlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "get", "p1234")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestPlansGet_NotFound(t *testing.T) {
	mock := &mockPlanAPI{
		getPlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Plan with id = nonexistent",
			}
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("plans", "get", "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found plan")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

func TestPlansGet_CurrenciesFormatting(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		getPlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "get", "p1234", "--output", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain both currencies formatted with setup fees
	if !strings.Contains(out, "USD: 10.00 (setup: 5.00)") {
		t.Errorf("expected USD currency formatting, got:\n%s", out)
	}
	if !strings.Contains(out, "EUR: 9.00 (setup: 4.50)") {
		t.Errorf("expected EUR currency formatting, got:\n%s", out)
	}
}

// --- plans create ---

func TestPlansCreate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("plans", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "create") {
		t.Error("expected plans help to show 'create' subcommand")
	}
}

func TestPlansCreateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand("plans", "create", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{
		"--code", "--name", "--interval-unit", "--interval-length", "--description", "--pricing-model",
		"--currency", "--unit-amount", "--setup-fee",
		"--trial-unit", "--trial-length", "--trial-requires-billing-info",
		"--auto-renew", "--total-billing-cycles",
		"--tax-code", "--tax-exempt", "--avalara-transaction-type", "--avalara-service-type",
		"--vertex-transaction-type", "--harmonized-system-code",
		"--accounting-code", "--revenue-schedule-type", "--liability-gl-account-id", "--revenue-gl-account-id",
		"--performance-obligation-id", "--setup-fee-accounting-code", "--setup-fee-revenue-schedule-type",
		"--setup-fee-liability-gl-account-id", "--setup-fee-revenue-gl-account-id", "--setup-fee-performance-obligation-id",
		"--success-url", "--cancel-url", "--bypass-confirmation", "--display-quantity",
		"--allow-any-item-on-subscriptions", "--dunning-campaign-id",
	} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestPlansCreate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand("plans", "create", "--code", "test")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestPlansCreate_CoreFlags(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		"--code", "gold",
		"--name", "Gold Plan",
		"--interval-unit", "month",
		"--interval-length", "1",
		"--description", "A premium plan",
		"--pricing-model", "fixed",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if *capturedBody.Code != "gold" {
		t.Errorf("expected code=gold, got %v", *capturedBody.Code)
	}
	if *capturedBody.Name != "Gold Plan" {
		t.Errorf("expected name=Gold Plan, got %v", *capturedBody.Name)
	}
	if *capturedBody.IntervalUnit != "month" {
		t.Errorf("expected interval-unit=month, got %v", *capturedBody.IntervalUnit)
	}
	if *capturedBody.IntervalLength != 1 {
		t.Errorf("expected interval-length=1, got %v", *capturedBody.IntervalLength)
	}
	if *capturedBody.Description != "A premium plan" {
		t.Errorf("expected description, got %v", *capturedBody.Description)
	}
	if *capturedBody.PricingModel != "fixed" {
		t.Errorf("expected pricing-model=fixed, got %v", *capturedBody.PricingModel)
	}
}

func TestPlansCreate_MultiCurrencyFlags(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		"--code", "multi",
		"--name", "Multi Currency",
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

func TestPlansCreate_SingleCurrency(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		"--code", "single",
		"--name", "Single Currency",
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

func TestPlansCreate_AllFlagsPopulated(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		// Core
		"--code", "full",
		"--name", "Full Plan",
		"--interval-unit", "month",
		"--interval-length", "1",
		"--description", "All flags test",
		"--pricing-model", "fixed",
		// Currency
		"--currency", "USD",
		"--unit-amount", "29.99",
		"--setup-fee", "10.00",
		// Trial
		"--trial-unit", "day",
		"--trial-length", "7",
		"--trial-requires-billing-info",
		// Billing
		"--auto-renew",
		"--total-billing-cycles", "24",
		// Tax
		"--tax-code", "digital",
		"--tax-exempt",
		"--avalara-transaction-type", "100",
		"--avalara-service-type", "200",
		"--vertex-transaction-type", "sale",
		"--harmonized-system-code", "8471.30",
		// Accounting
		"--accounting-code", "PLAN-FULL",
		"--revenue-schedule-type", "evenly",
		"--liability-gl-account-id", "gl-1",
		"--revenue-gl-account-id", "gl-2",
		"--performance-obligation-id", "po-1",
		"--setup-fee-accounting-code", "SF-FULL",
		"--setup-fee-revenue-schedule-type", "at_range_start",
		"--setup-fee-liability-gl-account-id", "gl-3",
		"--setup-fee-revenue-gl-account-id", "gl-4",
		"--setup-fee-performance-obligation-id", "po-2",
		// Hosted pages
		"--success-url", "https://example.com/ok",
		"--cancel-url", "https://example.com/cancel",
		"--bypass-confirmation",
		"--display-quantity",
		// Other
		"--allow-any-item-on-subscriptions",
		"--dunning-campaign-id", "dc-full",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all flag groups were captured
	if *capturedBody.Code != "full" {
		t.Errorf("expected code=full, got %v", *capturedBody.Code)
	}
	if *capturedBody.Name != "Full Plan" {
		t.Errorf("expected name=Full Plan, got %v", *capturedBody.Name)
	}
	if *capturedBody.IntervalUnit != "month" {
		t.Errorf("expected interval-unit=month, got %v", *capturedBody.IntervalUnit)
	}
	if *capturedBody.IntervalLength != 1 {
		t.Errorf("expected interval-length=1, got %v", *capturedBody.IntervalLength)
	}
	if *capturedBody.Description != "All flags test" {
		t.Errorf("expected description, got %v", *capturedBody.Description)
	}
	if *capturedBody.PricingModel != "fixed" {
		t.Errorf("expected pricing-model=fixed, got %v", *capturedBody.PricingModel)
	}
	if capturedBody.Currencies == nil || len(*capturedBody.Currencies) != 1 {
		t.Fatal("expected 1 currency")
	}
	if capturedBody.SetupFees == nil || len(*capturedBody.SetupFees) != 1 {
		t.Fatal("expected 1 setup fee")
	}
	if *capturedBody.TrialUnit != "day" {
		t.Errorf("expected trial-unit=day, got %v", *capturedBody.TrialUnit)
	}
	if *capturedBody.TrialLength != 7 {
		t.Errorf("expected trial-length=7, got %v", *capturedBody.TrialLength)
	}
	if *capturedBody.TrialRequiresBillingInfo != true {
		t.Error("expected trial-requires-billing-info=true")
	}
	if *capturedBody.AutoRenew != true {
		t.Error("expected auto-renew=true")
	}
	if *capturedBody.TotalBillingCycles != 24 {
		t.Errorf("expected total-billing-cycles=24, got %v", *capturedBody.TotalBillingCycles)
	}
	if *capturedBody.TaxCode != "digital" {
		t.Errorf("expected tax-code=digital, got %v", *capturedBody.TaxCode)
	}
	if *capturedBody.TaxExempt != true {
		t.Error("expected tax-exempt=true")
	}
	if *capturedBody.AccountingCode != "PLAN-FULL" {
		t.Errorf("expected accounting-code=PLAN-FULL, got %v", *capturedBody.AccountingCode)
	}
	if capturedBody.HostedPages == nil {
		t.Fatal("expected hosted pages to be set")
	}
	if *capturedBody.HostedPages.SuccessUrl != "https://example.com/ok" {
		t.Errorf("expected success-url, got %v", *capturedBody.HostedPages.SuccessUrl)
	}
	if *capturedBody.AllowAnyItemOnSubscriptions != true {
		t.Error("expected allow-any-item-on-subscriptions=true")
	}
	if *capturedBody.DunningCampaignId != "dc-full" {
		t.Errorf("expected dunning-campaign-id=dc-full, got %v", *capturedBody.DunningCampaignId)
	}
}

func TestPlansCreate_CurrencyUnitAmountMismatch(t *testing.T) {
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("plans", "create",
		"--code", "bad",
		"--name", "Bad Plan",
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

func TestPlansCreate_SetupFeeFlags(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		"--code", "fees",
		"--name", "Fee Plan",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "10.00", "--unit-amount", "9.00",
		"--setup-fee", "5.00", "--setup-fee", "4.50",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.SetupFees == nil {
		t.Fatal("expected setup fees to be set")
	}
	fees := *capturedBody.SetupFees
	if len(fees) != 2 {
		t.Fatalf("expected 2 setup fees, got %d", len(fees))
	}
	if *fees[0].Currency != "USD" || *fees[0].UnitAmount != 5.00 {
		t.Errorf("expected USD/5.00, got %s/%.2f", *fees[0].Currency, *fees[0].UnitAmount)
	}
	if *fees[1].Currency != "EUR" || *fees[1].UnitAmount != 4.50 {
		t.Errorf("expected EUR/4.50, got %s/%.2f", *fees[1].Currency, *fees[1].UnitAmount)
	}
}

func TestPlansCreate_SetupFeeCurrencyMismatch(t *testing.T) {
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("plans", "create",
		"--code", "bad",
		"--name", "Bad Plan",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "10.00", "--unit-amount", "9.00",
		"--setup-fee", "5.00",
	)
	if err == nil {
		t.Fatal("expected error for mismatched setup-fee/currency")
	}
	if !strings.Contains(stderr, "number of --setup-fee values must match --currency values") {
		t.Errorf("expected mismatch error, got %q", stderr)
	}
}

func TestPlansCreate_TrialFlags(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		"--code", "trial",
		"--name", "Trial Plan",
		"--trial-unit", "day",
		"--trial-length", "14",
		"--trial-requires-billing-info",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.TrialUnit != "day" {
		t.Errorf("expected trial-unit=day, got %v", *capturedBody.TrialUnit)
	}
	if *capturedBody.TrialLength != 14 {
		t.Errorf("expected trial-length=14, got %v", *capturedBody.TrialLength)
	}
	if *capturedBody.TrialRequiresBillingInfo != true {
		t.Errorf("expected trial-requires-billing-info=true")
	}
}

func TestPlansCreate_BillingFlags(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		"--code", "billing",
		"--name", "Billing Plan",
		"--auto-renew",
		"--total-billing-cycles", "12",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.AutoRenew != true {
		t.Error("expected auto-renew=true")
	}
	if *capturedBody.TotalBillingCycles != 12 {
		t.Errorf("expected total-billing-cycles=12, got %v", *capturedBody.TotalBillingCycles)
	}
}

func TestPlansCreate_TaxFlags(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		"--code", "tax",
		"--name", "Tax Plan",
		"--tax-code", "digital",
		"--tax-exempt",
		"--avalara-transaction-type", "100",
		"--avalara-service-type", "200",
		"--vertex-transaction-type", "sale",
		"--harmonized-system-code", "1234.56",
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
	if *capturedBody.AvalaraTransactionType != 100 {
		t.Errorf("expected avalara-transaction-type=100, got %v", *capturedBody.AvalaraTransactionType)
	}
	if *capturedBody.AvalaraServiceType != 200 {
		t.Errorf("expected avalara-service-type=200, got %v", *capturedBody.AvalaraServiceType)
	}
	if *capturedBody.VertexTransactionType != "sale" {
		t.Errorf("expected vertex-transaction-type=sale, got %v", *capturedBody.VertexTransactionType)
	}
	if *capturedBody.HarmonizedSystemCode != "1234.56" {
		t.Errorf("expected harmonized-system-code=1234.56, got %v", *capturedBody.HarmonizedSystemCode)
	}
}

func TestPlansCreate_AccountingFlags(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		"--code", "acct",
		"--name", "Acct Plan",
		"--accounting-code", "plan-acct",
		"--revenue-schedule-type", "evenly",
		"--liability-gl-account-id", "gl-liab",
		"--revenue-gl-account-id", "gl-rev",
		"--performance-obligation-id", "po-1",
		"--setup-fee-accounting-code", "sf-acct",
		"--setup-fee-revenue-schedule-type", "at_range_start",
		"--setup-fee-liability-gl-account-id", "sf-gl-liab",
		"--setup-fee-revenue-gl-account-id", "sf-gl-rev",
		"--setup-fee-performance-obligation-id", "sf-po-1",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.AccountingCode != "plan-acct" {
		t.Errorf("expected accounting-code=plan-acct, got %v", *capturedBody.AccountingCode)
	}
	if *capturedBody.RevenueScheduleType != "evenly" {
		t.Errorf("expected revenue-schedule-type=evenly, got %v", *capturedBody.RevenueScheduleType)
	}
	if *capturedBody.LiabilityGlAccountId != "gl-liab" {
		t.Errorf("expected liability-gl-account-id=gl-liab, got %v", *capturedBody.LiabilityGlAccountId)
	}
	if *capturedBody.RevenueGlAccountId != "gl-rev" {
		t.Errorf("expected revenue-gl-account-id=gl-rev, got %v", *capturedBody.RevenueGlAccountId)
	}
	if *capturedBody.PerformanceObligationId != "po-1" {
		t.Errorf("expected performance-obligation-id=po-1, got %v", *capturedBody.PerformanceObligationId)
	}
	if *capturedBody.SetupFeeAccountingCode != "sf-acct" {
		t.Errorf("expected setup-fee-accounting-code=sf-acct, got %v", *capturedBody.SetupFeeAccountingCode)
	}
	if *capturedBody.SetupFeeRevenueScheduleType != "at_range_start" {
		t.Errorf("expected setup-fee-revenue-schedule-type=at_range_start, got %v", *capturedBody.SetupFeeRevenueScheduleType)
	}
	if *capturedBody.SetupFeeLiabilityGlAccountId != "sf-gl-liab" {
		t.Errorf("expected setup-fee-liability-gl-account-id=sf-gl-liab, got %v", *capturedBody.SetupFeeLiabilityGlAccountId)
	}
	if *capturedBody.SetupFeeRevenueGlAccountId != "sf-gl-rev" {
		t.Errorf("expected setup-fee-revenue-gl-account-id=sf-gl-rev, got %v", *capturedBody.SetupFeeRevenueGlAccountId)
	}
	if *capturedBody.SetupFeePerformanceObligationId != "sf-po-1" {
		t.Errorf("expected setup-fee-performance-obligation-id=sf-po-1, got %v", *capturedBody.SetupFeePerformanceObligationId)
	}
}

func TestPlansCreate_HostedPagesFlags(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		"--code", "hosted",
		"--name", "Hosted Plan",
		"--success-url", "https://example.com/success",
		"--cancel-url", "https://example.com/cancel",
		"--bypass-confirmation",
		"--display-quantity",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.HostedPages == nil {
		t.Fatal("expected hosted pages to be set")
	}
	hp := capturedBody.HostedPages
	if *hp.SuccessUrl != "https://example.com/success" {
		t.Errorf("expected success-url, got %v", *hp.SuccessUrl)
	}
	if *hp.CancelUrl != "https://example.com/cancel" {
		t.Errorf("expected cancel-url, got %v", *hp.CancelUrl)
	}
	if *hp.BypassConfirmation != true {
		t.Error("expected bypass-confirmation=true")
	}
	if *hp.DisplayQuantity != true {
		t.Error("expected display-quantity=true")
	}
}

func TestPlansCreate_OtherFlags(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create",
		"--code", "other",
		"--name", "Other Plan",
		"--allow-any-item-on-subscriptions",
		"--dunning-campaign-id", "dc-1",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.AllowAnyItemOnSubscriptions != true {
		t.Error("expected allow-any-item-on-subscriptions=true")
	}
	if *capturedBody.DunningCampaignId != "dc-1" {
		t.Errorf("expected dunning-campaign-id=dc-1, got %v", *capturedBody.DunningCampaignId)
	}
}

func TestPlansCreate_UnsetFlagsNotSent(t *testing.T) {
	var capturedBody *recurly.PlanCreate
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create", "--code", "minimal", "--name", "Minimal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.IntervalUnit != nil {
		t.Error("expected interval-unit to be nil when not set")
	}
	if capturedBody.IntervalLength != nil {
		t.Error("expected interval-length to be nil when not set")
	}
	if capturedBody.Description != nil {
		t.Error("expected description to be nil when not set")
	}
	if capturedBody.Currencies != nil {
		t.Error("expected currencies to be nil when not set")
	}
	if capturedBody.SetupFees != nil {
		t.Error("expected setup-fees to be nil when not set")
	}
	if capturedBody.TrialUnit != nil {
		t.Error("expected trial-unit to be nil when not set")
	}
	if capturedBody.AutoRenew != nil {
		t.Error("expected auto-renew to be nil when not set")
	}
	if capturedBody.TaxCode != nil {
		t.Error("expected tax-code to be nil when not set")
	}
	if capturedBody.HostedPages != nil {
		t.Error("expected hosted-pages to be nil when not set")
	}
	if capturedBody.AllowAnyItemOnSubscriptions != nil {
		t.Error("expected allow-any-item-on-subscriptions to be nil when not set")
	}
}

func TestPlansCreate_TableOutput(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "create", "--code", "gold", "--name", "Gold Plan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should display using planDetailColumns (same as get)
	for _, expected := range []string{"Id", "p1234", "Code", "gold", "Name", "Gold Plan"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestPlansCreate_JSONOutput(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "create", "--code", "gold", "--name", "Gold", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["code"] != "gold" {
		t.Errorf("expected code=gold, got %v", result["code"])
	}
}

func TestPlansCreate_SDKError(t *testing.T) {
	mock := &mockPlanAPI{
		createPlanFn: func(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error) {
			return nil, fmt.Errorf("validation error")
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "create", "--code", "bad", "--name", "Bad")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestPlansGet_EmptyCurrencies(t *testing.T) {
	viper.Reset()
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	mock := &mockPlanAPI{
		getPlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			return &recurly.Plan{
				Id:        "p999",
				Code:      "basic",
				Name:      "Basic Plan",
				CreatedAt: &now,
			}, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "get", "p999", "--output", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "basic") {
		t.Errorf("expected output to contain 'basic', got:\n%s", out)
	}
}

// --- plans update ---

func TestPlansUpdate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("plans", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "update") {
		t.Error("expected plans help to show 'update' subcommand")
	}
}

func TestPlansUpdateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand("plans", "update", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{
		"--code", "--name", "--description",
		"--currency", "--unit-amount", "--setup-fee",
		"--trial-unit", "--trial-length", "--trial-requires-billing-info",
		"--auto-renew", "--total-billing-cycles",
		"--tax-code", "--tax-exempt", "--avalara-transaction-type", "--avalara-service-type",
		"--vertex-transaction-type", "--harmonized-system-code",
		"--accounting-code", "--revenue-schedule-type", "--liability-gl-account-id", "--revenue-gl-account-id",
		"--performance-obligation-id", "--setup-fee-accounting-code", "--setup-fee-revenue-schedule-type",
		"--setup-fee-liability-gl-account-id", "--setup-fee-revenue-gl-account-id", "--setup-fee-performance-obligation-id",
		"--success-url", "--cancel-url", "--bypass-confirmation", "--display-quantity",
		"--allow-any-item-on-subscriptions", "--dunning-campaign-id",
	} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestPlansUpdateHelp_NoImmutableFlags(t *testing.T) {
	out, _, err := executeCommand("plans", "update", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--interval-unit", "--interval-length", "--pricing-model"} {
		if strings.Contains(out, flag) {
			t.Errorf("expected help output NOT to contain immutable flag %q", flag)
		}
	}
}

func TestPlansUpdate_MissingArg(t *testing.T) {
	_, _, err := executeCommand("plans", "update")
	if err == nil {
		t.Fatal("expected error when plan_id argument is missing")
	}
}

func TestPlansUpdate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand("plans", "update", "p1234", "--name", "New Name")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestPlansUpdate_PlanIdArg(t *testing.T) {
	var capturedPlanId string
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedPlanId = planId
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234", "--name", "Updated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPlanId != "p1234" {
		t.Errorf("expected planId=p1234, got %v", capturedPlanId)
	}
}

func TestPlansUpdate_CoreFlags(t *testing.T) {
	var capturedBody *recurly.PlanUpdate
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--code", "gold-v2",
		"--name", "Gold Plan V2",
		"--description", "Updated plan",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if *capturedBody.Code != "gold-v2" {
		t.Errorf("expected code=gold-v2, got %v", *capturedBody.Code)
	}
	if *capturedBody.Name != "Gold Plan V2" {
		t.Errorf("expected name=Gold Plan V2, got %v", *capturedBody.Name)
	}
	if *capturedBody.Description != "Updated plan" {
		t.Errorf("expected description=Updated plan, got %v", *capturedBody.Description)
	}
}

func TestPlansUpdate_MultiCurrencyFlags(t *testing.T) {
	var capturedBody *recurly.PlanUpdate
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "15.00", "--unit-amount", "13.00",
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
	if *pricings[0].Currency != "USD" || *pricings[0].UnitAmount != 15.00 {
		t.Errorf("unexpected first currency: %v", pricings[0])
	}
	if *pricings[1].Currency != "EUR" || *pricings[1].UnitAmount != 13.00 {
		t.Errorf("unexpected second currency: %v", pricings[1])
	}
}

func TestPlansUpdate_CurrencyUnitAmountMismatch(t *testing.T) {
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "10.00",
	)
	if err == nil {
		t.Fatal("expected error for currency/unit-amount mismatch")
	}
	if !strings.Contains(err.Error(), "must match") {
		t.Errorf("expected mismatch error, got: %v", err)
	}
}

func TestPlansUpdate_SetupFeeFlags(t *testing.T) {
	var capturedBody *recurly.PlanUpdate
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--currency", "USD", "--currency", "EUR",
		"--unit-amount", "10.00", "--unit-amount", "9.00",
		"--setup-fee", "5.00", "--setup-fee", "4.50",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.SetupFees == nil {
		t.Fatal("expected setup-fees to be set")
	}
	fees := *capturedBody.SetupFees
	if len(fees) != 2 {
		t.Fatalf("expected 2 setup fees, got %d", len(fees))
	}
	if *fees[0].Currency != "USD" || *fees[0].UnitAmount != 5.00 {
		t.Errorf("unexpected first setup fee: %v", fees[0])
	}
	if *fees[1].Currency != "EUR" || *fees[1].UnitAmount != 4.50 {
		t.Errorf("unexpected second setup fee: %v", fees[1])
	}
}

func TestPlansUpdate_SetupFeeCurrencyMismatch(t *testing.T) {
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--currency", "USD",
		"--unit-amount", "10.00",
		"--setup-fee", "5.00", "--setup-fee", "4.50",
	)
	if err == nil {
		t.Fatal("expected error for setup-fee/currency mismatch")
	}
	if !strings.Contains(err.Error(), "must match") {
		t.Errorf("expected mismatch error, got: %v", err)
	}
}

func TestPlansUpdate_TrialFlags(t *testing.T) {
	var capturedBody *recurly.PlanUpdate
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--trial-unit", "day",
		"--trial-length", "14",
		"--trial-requires-billing-info",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.TrialUnit != "day" {
		t.Errorf("expected trial-unit=day, got %v", *capturedBody.TrialUnit)
	}
	if *capturedBody.TrialLength != 14 {
		t.Errorf("expected trial-length=14, got %v", *capturedBody.TrialLength)
	}
	if *capturedBody.TrialRequiresBillingInfo != true {
		t.Errorf("expected trial-requires-billing-info=true, got %v", *capturedBody.TrialRequiresBillingInfo)
	}
}

func TestPlansUpdate_BillingFlags(t *testing.T) {
	var capturedBody *recurly.PlanUpdate
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--auto-renew",
		"--total-billing-cycles", "12",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.AutoRenew != true {
		t.Errorf("expected auto-renew=true, got %v", *capturedBody.AutoRenew)
	}
	if *capturedBody.TotalBillingCycles != 12 {
		t.Errorf("expected total-billing-cycles=12, got %v", *capturedBody.TotalBillingCycles)
	}
}

func TestPlansUpdate_TaxFlags(t *testing.T) {
	var capturedBody *recurly.PlanUpdate
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--tax-code", "digital",
		"--tax-exempt",
		"--avalara-transaction-type", "600",
		"--avalara-service-type", "100",
		"--vertex-transaction-type", "sale",
		"--harmonized-system-code", "1234",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.TaxCode != "digital" {
		t.Errorf("expected tax-code=digital, got %v", *capturedBody.TaxCode)
	}
	if *capturedBody.TaxExempt != true {
		t.Errorf("expected tax-exempt=true, got %v", *capturedBody.TaxExempt)
	}
	if *capturedBody.AvalaraTransactionType != 600 {
		t.Errorf("expected avalara-transaction-type=600, got %v", *capturedBody.AvalaraTransactionType)
	}
	if *capturedBody.AvalaraServiceType != 100 {
		t.Errorf("expected avalara-service-type=100, got %v", *capturedBody.AvalaraServiceType)
	}
	if *capturedBody.VertexTransactionType != "sale" {
		t.Errorf("expected vertex-transaction-type=sale, got %v", *capturedBody.VertexTransactionType)
	}
	if *capturedBody.HarmonizedSystemCode != "1234" {
		t.Errorf("expected harmonized-system-code=1234, got %v", *capturedBody.HarmonizedSystemCode)
	}
}

func TestPlansUpdate_AccountingFlags(t *testing.T) {
	var capturedBody *recurly.PlanUpdate
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--accounting-code", "PLAN-001",
		"--revenue-schedule-type", "evenly",
		"--liability-gl-account-id", "gl-001",
		"--revenue-gl-account-id", "gl-002",
		"--performance-obligation-id", "po-001",
		"--setup-fee-accounting-code", "SETUP-001",
		"--setup-fee-revenue-schedule-type", "at_invoice",
		"--setup-fee-liability-gl-account-id", "gl-003",
		"--setup-fee-revenue-gl-account-id", "gl-004",
		"--setup-fee-performance-obligation-id", "po-002",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.AccountingCode != "PLAN-001" {
		t.Errorf("expected accounting-code=PLAN-001, got %v", *capturedBody.AccountingCode)
	}
	if *capturedBody.RevenueScheduleType != "evenly" {
		t.Errorf("expected revenue-schedule-type=evenly, got %v", *capturedBody.RevenueScheduleType)
	}
	if *capturedBody.LiabilityGlAccountId != "gl-001" {
		t.Errorf("expected liability-gl-account-id=gl-001, got %v", *capturedBody.LiabilityGlAccountId)
	}
	if *capturedBody.RevenueGlAccountId != "gl-002" {
		t.Errorf("expected revenue-gl-account-id=gl-002, got %v", *capturedBody.RevenueGlAccountId)
	}
	if *capturedBody.PerformanceObligationId != "po-001" {
		t.Errorf("expected performance-obligation-id=po-001, got %v", *capturedBody.PerformanceObligationId)
	}
	if *capturedBody.SetupFeeAccountingCode != "SETUP-001" {
		t.Errorf("expected setup-fee-accounting-code=SETUP-001, got %v", *capturedBody.SetupFeeAccountingCode)
	}
	if *capturedBody.SetupFeeRevenueScheduleType != "at_invoice" {
		t.Errorf("expected setup-fee-revenue-schedule-type=at_invoice, got %v", *capturedBody.SetupFeeRevenueScheduleType)
	}
	if *capturedBody.SetupFeeLiabilityGlAccountId != "gl-003" {
		t.Errorf("expected setup-fee-liability-gl-account-id=gl-003, got %v", *capturedBody.SetupFeeLiabilityGlAccountId)
	}
	if *capturedBody.SetupFeeRevenueGlAccountId != "gl-004" {
		t.Errorf("expected setup-fee-revenue-gl-account-id=gl-004, got %v", *capturedBody.SetupFeeRevenueGlAccountId)
	}
	if *capturedBody.SetupFeePerformanceObligationId != "po-002" {
		t.Errorf("expected setup-fee-performance-obligation-id=po-002, got %v", *capturedBody.SetupFeePerformanceObligationId)
	}
}

func TestPlansUpdate_HostedPagesFlags(t *testing.T) {
	var capturedBody *recurly.PlanUpdate
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--success-url", "https://example.com/success",
		"--cancel-url", "https://example.com/cancel",
		"--bypass-confirmation",
		"--display-quantity",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.HostedPages == nil {
		t.Fatal("expected hosted-pages to be set")
	}
	hp := capturedBody.HostedPages
	if *hp.SuccessUrl != "https://example.com/success" {
		t.Errorf("expected success-url, got %v", *hp.SuccessUrl)
	}
	if *hp.CancelUrl != "https://example.com/cancel" {
		t.Errorf("expected cancel-url, got %v", *hp.CancelUrl)
	}
	if *hp.BypassConfirmation != true {
		t.Errorf("expected bypass-confirmation=true, got %v", *hp.BypassConfirmation)
	}
	if *hp.DisplayQuantity != true {
		t.Errorf("expected display-quantity=true, got %v", *hp.DisplayQuantity)
	}
}

func TestPlansUpdate_OtherFlags(t *testing.T) {
	var capturedBody *recurly.PlanUpdate
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234",
		"--allow-any-item-on-subscriptions",
		"--dunning-campaign-id", "dc-001",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.AllowAnyItemOnSubscriptions != true {
		t.Errorf("expected allow-any-item=true, got %v", *capturedBody.AllowAnyItemOnSubscriptions)
	}
	if *capturedBody.DunningCampaignId != "dc-001" {
		t.Errorf("expected dunning-campaign-id=dc-001, got %v", *capturedBody.DunningCampaignId)
	}
}

func TestPlansUpdate_UnsetFlagsNotSent(t *testing.T) {
	var capturedBody *recurly.PlanUpdate
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedBody = body
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234", "--name", "Updated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Code != nil {
		t.Error("expected code to be nil when not set")
	}
	if capturedBody.Description != nil {
		t.Error("expected description to be nil when not set")
	}
	if capturedBody.Currencies != nil {
		t.Error("expected currencies to be nil when not set")
	}
	if capturedBody.SetupFees != nil {
		t.Error("expected setup-fees to be nil when not set")
	}
	if capturedBody.TrialUnit != nil {
		t.Error("expected trial-unit to be nil when not set")
	}
	if capturedBody.AutoRenew != nil {
		t.Error("expected auto-renew to be nil when not set")
	}
	if capturedBody.TaxCode != nil {
		t.Error("expected tax-code to be nil when not set")
	}
	if capturedBody.HostedPages != nil {
		t.Error("expected hosted-pages to be nil when not set")
	}
	if capturedBody.AllowAnyItemOnSubscriptions != nil {
		t.Error("expected allow-any-item-on-subscriptions to be nil when not set")
	}
}

func TestPlansUpdate_TableOutput(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "update", "p1234", "--name", "Gold Plan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"Id", "p1234", "Code", "gold", "Name", "Gold Plan"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestPlansUpdate_JSONOutput(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "update", "p1234", "--name", "Gold", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["code"] != "gold" {
		t.Errorf("expected code=gold, got %v", result["code"])
	}
}

func TestPlansUpdate_SDKError(t *testing.T) {
	mock := &mockPlanAPI{
		updatePlanFn: func(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("plans", "update", "p1234", "--name", "Bad")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- plans deactivate ---

func TestPlansDeactivate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("plans", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "deactivate") {
		t.Error("expected plans help to show 'deactivate' subcommand")
	}
}

func TestPlansDeactivateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand("plans", "deactivate", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--yes") {
		t.Error("expected help output to contain --yes flag")
	}
}

func TestPlansDeactivate_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand("plans", "deactivate")
	if err == nil {
		t.Fatal("expected error when no plan ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestPlansDeactivate_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(stdin, "plans", "deactivate", "plan-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to deactivate this plan? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Deactivation cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestPlansDeactivate_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(stdin, "plans", "deactivate", "plan-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deactivation cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestPlansDeactivate_ConfirmYes_Succeeds(t *testing.T) {
	viper.Reset()
	var capturedID string
	mock := &mockPlanAPI{
		removePlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedID = planId
			p := samplePlanDetail()
			p.State = "inactive"
			return p, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(stdin, "plans", "deactivate", "plan-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "plan-789" {
		t.Errorf("expected plan ID 'plan-789', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "inactive") {
		t.Errorf("expected plan details in output, got:\n%s", out)
	}
}

func TestPlansDeactivate_ConfirmYES_CaseInsensitive(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		removePlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			p := samplePlanDetail()
			p.State = "inactive"
			return p, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	stdin := bytes.NewBufferString("YES\n")
	out, _, err := executeCommandWithStdin(stdin, "plans", "deactivate", "plan-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "inactive") {
		t.Errorf("expected plan details in output, got:\n%s", out)
	}
}

func TestPlansDeactivate_YesFlag_SkipsPrompt(t *testing.T) {
	viper.Reset()
	var capturedID string
	mock := &mockPlanAPI{
		removePlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			capturedID = planId
			p := samplePlanDetail()
			p.State = "inactive"
			return p, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, stderr, err := executeCommand("plans", "deactivate", "plan-789", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "plan-789" {
		t.Errorf("expected plan ID 'plan-789', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "inactive") {
		t.Errorf("expected plan details in output, got:\n%s", out)
	}
}

func TestPlansDeactivate_SDKError(t *testing.T) {
	mock := &mockPlanAPI{
		removePlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Plan",
			}
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("plans", "deactivate", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

func TestPlansDeactivate_JSON_Output(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		removePlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			p := samplePlanDetail()
			p.State = "inactive"
			return p, nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "deactivate", "plan-789", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nOutput: %s", err, out)
	}
	if result["state"] != "inactive" {
		t.Errorf("expected state 'inactive', got %v", result["state"])
	}
}

func TestPlansDeactivate_UsesDetailColumns(t *testing.T) {
	viper.Reset()
	mock := &mockPlanAPI{
		removePlanFn: func(planId string, opts ...recurly.Option) (*recurly.Plan, error) {
			return samplePlanDetail(), nil
		},
	}
	cleanup := setMockPlanAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("plans", "deactivate", "plan-789", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify planDetailColumns() fields are present
	for _, field := range []string{"Id", "Code", "Name", "State", "Pricing Model", "Interval Unit"} {
		if !strings.Contains(out, field) {
			t.Errorf("expected output to contain %q, got:\n%s", field, out)
		}
	}
}
