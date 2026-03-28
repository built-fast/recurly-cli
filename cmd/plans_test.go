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
	t.Setenv("RECURLY_API_KEY", "")
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
	t.Setenv("RECURLY_API_KEY", "")
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
	t.Setenv("RECURLY_API_KEY", "")
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
