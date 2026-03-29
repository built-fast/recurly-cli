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

// mockSubscriptionAPI implements SubscriptionAPI for testing.
type mockSubscriptionAPI struct {
	listSubscriptionsFn      func(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error)
	getSubscriptionFn        func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error)
	createSubscriptionFn     func(body *recurly.SubscriptionCreate, opts ...recurly.Option) (*recurly.Subscription, error)
	updateSubscriptionFn     func(subscriptionId string, body *recurly.SubscriptionUpdate, opts ...recurly.Option) (*recurly.Subscription, error)
	cancelSubscriptionFn     func(subscriptionId string, params *recurly.CancelSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error)
	reactivateSubscriptionFn func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error)
	pauseSubscriptionFn      func(subscriptionId string, body *recurly.SubscriptionPause, opts ...recurly.Option) (*recurly.Subscription, error)
	resumeSubscriptionFn     func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error)
	terminateSubscriptionFn  func(subscriptionId string, params *recurly.TerminateSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error)
	convertTrialFn           func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error)
}

func (m *mockSubscriptionAPI) ListSubscriptions(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
	return m.listSubscriptionsFn(params, opts...)
}

func (m *mockSubscriptionAPI) GetSubscription(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
	return m.getSubscriptionFn(subscriptionId, opts...)
}

func (m *mockSubscriptionAPI) CreateSubscription(body *recurly.SubscriptionCreate, opts ...recurly.Option) (*recurly.Subscription, error) {
	return m.createSubscriptionFn(body, opts...)
}

func (m *mockSubscriptionAPI) UpdateSubscription(subscriptionId string, body *recurly.SubscriptionUpdate, opts ...recurly.Option) (*recurly.Subscription, error) {
	return m.updateSubscriptionFn(subscriptionId, body, opts...)
}

func (m *mockSubscriptionAPI) CancelSubscription(subscriptionId string, params *recurly.CancelSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
	return m.cancelSubscriptionFn(subscriptionId, params, opts...)
}

func (m *mockSubscriptionAPI) ReactivateSubscription(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
	return m.reactivateSubscriptionFn(subscriptionId, opts...)
}

func (m *mockSubscriptionAPI) PauseSubscription(subscriptionId string, body *recurly.SubscriptionPause, opts ...recurly.Option) (*recurly.Subscription, error) {
	return m.pauseSubscriptionFn(subscriptionId, body, opts...)
}

func (m *mockSubscriptionAPI) ResumeSubscription(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
	return m.resumeSubscriptionFn(subscriptionId, opts...)
}

func (m *mockSubscriptionAPI) TerminateSubscription(subscriptionId string, params *recurly.TerminateSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
	return m.terminateSubscriptionFn(subscriptionId, params, opts...)
}

func (m *mockSubscriptionAPI) ConvertTrial(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
	return m.convertTrialFn(subscriptionId, opts...)
}

// sampleSubscription returns a test subscription with predictable fields for list output.
func sampleSubscription() recurly.Subscription {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	periodEnd := time.Date(2025, 2, 15, 10, 30, 0, 0, time.UTC)
	return recurly.Subscription{
		Id:                  "sub-123",
		Uuid:                "uuid-abc",
		Account:             recurly.AccountMini{Code: "acct-456"},
		Plan:                recurly.PlanMini{Id: "plan-789", Code: "gold", Name: "Gold Plan"},
		State:               "active",
		Currency:            "USD",
		UnitAmount:          19.99,
		Quantity:            1,
		Subtotal:            19.99,
		CollectionMethod:    "automatic",
		CurrentPeriodEndsAt: &periodEnd,
		CreatedAt:           &now,
	}
}

// sampleSubscriptionDetail returns a test subscription pointer with detail fields populated.
func sampleSubscriptionDetail() *recurly.Subscription {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	periodStart := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	periodEnd := time.Date(2025, 2, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC)
	return &recurly.Subscription{
		Id:                     "sub-123",
		Uuid:                   "uuid-abc",
		Account:                recurly.AccountMini{Code: "acct-456"},
		Plan:                   recurly.PlanMini{Id: "plan-789", Code: "gold", Name: "Gold Plan"},
		State:                  "active",
		Currency:               "USD",
		UnitAmount:             19.99,
		Quantity:               1,
		Subtotal:               19.99,
		Tax:                    1.60,
		Total:                  21.59,
		CollectionMethod:       "automatic",
		AutoRenew:              true,
		NetTerms:               0,
		CurrentPeriodStartedAt: &periodStart,
		CurrentPeriodEndsAt:    &periodEnd,
		CreatedAt:              &now,
		UpdatedAt:              &updated,
		ActivatedAt:            &now,
	}
}

// --- subscriptions list ---

func TestSubscriptionsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected subscriptions help to show 'list' subcommand")
	}
}

func TestSubscriptionsListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--order", "--sort", "--state", "--plan-id", "--begin-time", "--end-time"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestSubscriptionsList_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "subscriptions", "list")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestSubscriptionsList_InvalidBeginTime_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand(nil, "subscriptions", "list", "--begin-time", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid begin-time")
	}
	if !strings.Contains(stderr, "invalid --begin-time") {
		t.Errorf("expected 'invalid --begin-time' error, got %q", stderr)
	}
}

func TestSubscriptionsList_InvalidEndTime_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand(nil, "subscriptions", "list", "--end-time", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid end-time")
	}
	if !strings.Contains(stderr, "invalid --end-time") {
		t.Errorf("expected 'invalid --end-time' error, got %q", stderr)
	}
}

func TestSubscriptionsList_PaginationParams(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListSubscriptionsParams

	mock := &mockSubscriptionAPI{
		listSubscriptionsFn: func(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedParams = params
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "list", "--limit", "50", "--order", "desc", "--sort", "updated_at")
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

func TestSubscriptionsList_FilterParams(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListSubscriptionsParams

	mock := &mockSubscriptionAPI{
		listSubscriptionsFn: func(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedParams = params
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "list",
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

func TestSubscriptionsList_UnsetFlagsNotSent(t *testing.T) {
	var capturedParams *recurly.ListSubscriptionsParams

	mock := &mockSubscriptionAPI{
		listSubscriptionsFn: func(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedParams = params
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "list")
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

func TestSubscriptionsList_TableOutput(t *testing.T) {
	sub := sampleSubscription()
	mock := &mockSubscriptionAPI{
		listSubscriptionsFn: func(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{sub}}, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"sub-123", "acct-456", "gold", "active", "USD", "19.99"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	for _, header := range []string{"ID", "Account Code", "Plan Code", "State", "Currency", "Unit Amount"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestSubscriptionsList_JSONOutput(t *testing.T) {
	sub := sampleSubscription()
	mock := &mockSubscriptionAPI{
		listSubscriptionsFn: func(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{sub}}, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "list", "--output", "json")
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
		t.Fatalf("expected 1 subscription in JSON output, got %d", len(envelope.Data))
	}
	if envelope.Data[0]["id"] != "sub-123" {
		t.Errorf("expected id=sub-123 in JSON, got %v", envelope.Data[0]["id"])
	}
}

func TestSubscriptionsList_JSONPrettyOutput(t *testing.T) {
	sub := sampleSubscription()
	mock := &mockSubscriptionAPI{
		listSubscriptionsFn: func(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{sub}}, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "list", "--output", "json-pretty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

func TestSubscriptionsList_SDKError(t *testing.T) {
	mock := &mockSubscriptionAPI{
		listSubscriptionsFn: func(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "list")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestSubscriptionsList_EmptyResults(t *testing.T) {
	mock := &mockSubscriptionAPI{
		listSubscriptionsFn: func(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockLister[recurly.Subscription]{items: []recurly.Subscription{}}, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "sub-") {
		t.Error("expected empty results, but found subscription data")
	}
}

// --- subscriptions get ---

func TestSubscriptionsGet_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "get") {
		t.Error("expected subscriptions help to show 'get' subcommand")
	}
}

func TestSubscriptionsGet_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "subscriptions", "get")
	if err == nil {
		t.Fatal("expected error when no subscription ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestSubscriptionsGet_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "subscriptions", "get", "sub-123")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestSubscriptionsGet_PositionalArg(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		getSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "get", "my-sub-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "my-sub-id" {
		t.Errorf("expected subscription ID 'my-sub-id', got %q", capturedID)
	}
}

func TestSubscriptionsGet_TableOutput(t *testing.T) {
	viper.Reset()
	mock := &mockSubscriptionAPI{
		getSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "get", "sub-123", "--output", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"ID", "sub-123",
		"UUID", "uuid-abc",
		"Account Code", "acct-456",
		"Plan Code", "gold",
		"State", "active",
		"Currency", "USD",
		"Unit Amount", "19.99",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestSubscriptionsGet_JSONOutput(t *testing.T) {
	mock := &mockSubscriptionAPI{
		getSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "get", "sub-123", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["id"] != "sub-123" {
		t.Errorf("expected id=sub-123 in JSON, got %v", result["id"])
	}
}

func TestSubscriptionsGet_NotFound(t *testing.T) {
	mock := &mockSubscriptionAPI{
		getSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Subscription with id = nonexistent",
			}
		},
	}
	app := newTestSubscriptionApp(mock)

	_, stderr, err := executeCommand(app, "subscriptions", "get", "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found subscription")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

// --- subscriptions create ---

func TestSubscriptionsCreate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "create") {
		t.Error("expected subscriptions help to show 'create' subcommand")
	}
}

func TestSubscriptionsCreateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "create", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--plan-code", "--account-code", "--currency", "--quantity", "--unit-amount", "--auto-renew", "--trial-ends-at", "--starts-at", "--next-bill-date", "--collection-method", "--po-number", "--net-terms", "--net-terms-type", "--total-billing-cycles", "--renewal-billing-cycles", "--coupon-code", "--gateway-code", "--billing-info-id"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestSubscriptionsCreate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "subscriptions", "create", "--plan-code", "gold")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestSubscriptionsCreate_FlagToStructMapping(t *testing.T) {
	var capturedBody *recurly.SubscriptionCreate

	mock := &mockSubscriptionAPI{
		createSubscriptionFn: func(body *recurly.SubscriptionCreate, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedBody = body
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "create",
		"--plan-code", "gold",
		"--account-code", "acct-new",
		"--currency", "USD",
		"--quantity", "2",
		"--unit-amount", "29.99",
		"--collection-method", "manual",
		"--po-number", "PO-123",
		"--net-terms", "30",
		"--net-terms-type", "net",
		"--total-billing-cycles", "12",
		"--renewal-billing-cycles", "6",
		"--coupon-code", "SAVE10",
		"--gateway-code", "gw-abc",
		"--billing-info-id", "bi-xyz",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if capturedBody.PlanCode == nil || *capturedBody.PlanCode != "gold" {
		t.Errorf("expected plan_code=gold, got %v", capturedBody.PlanCode)
	}
	if capturedBody.Account == nil || capturedBody.Account.Code == nil || *capturedBody.Account.Code != "acct-new" {
		t.Error("expected account.code=acct-new")
	}
	if capturedBody.Currency == nil || *capturedBody.Currency != "USD" {
		t.Errorf("expected currency=USD, got %v", capturedBody.Currency)
	}
	if capturedBody.Quantity == nil || *capturedBody.Quantity != 2 {
		t.Errorf("expected quantity=2, got %v", capturedBody.Quantity)
	}
	if capturedBody.UnitAmount == nil || *capturedBody.UnitAmount != 29.99 {
		t.Errorf("expected unit_amount=29.99, got %v", capturedBody.UnitAmount)
	}
	if capturedBody.CollectionMethod == nil || *capturedBody.CollectionMethod != "manual" {
		t.Errorf("expected collection_method=manual, got %v", capturedBody.CollectionMethod)
	}
	if capturedBody.PoNumber == nil || *capturedBody.PoNumber != "PO-123" {
		t.Errorf("expected po_number=PO-123, got %v", capturedBody.PoNumber)
	}
	if capturedBody.NetTerms == nil || *capturedBody.NetTerms != 30 {
		t.Errorf("expected net_terms=30, got %v", capturedBody.NetTerms)
	}
	if capturedBody.NetTermsType == nil || *capturedBody.NetTermsType != "net" {
		t.Errorf("expected net_terms_type=net, got %v", capturedBody.NetTermsType)
	}
	if capturedBody.TotalBillingCycles == nil || *capturedBody.TotalBillingCycles != 12 {
		t.Errorf("expected total_billing_cycles=12, got %v", capturedBody.TotalBillingCycles)
	}
	if capturedBody.RenewalBillingCycles == nil || *capturedBody.RenewalBillingCycles != 6 {
		t.Errorf("expected renewal_billing_cycles=6, got %v", capturedBody.RenewalBillingCycles)
	}
	if capturedBody.CouponCodes == nil || len(*capturedBody.CouponCodes) != 1 || (*capturedBody.CouponCodes)[0] != "SAVE10" {
		t.Errorf("expected coupon_codes=[SAVE10], got %v", capturedBody.CouponCodes)
	}
	if capturedBody.GatewayCode == nil || *capturedBody.GatewayCode != "gw-abc" {
		t.Errorf("expected gateway_code=gw-abc, got %v", capturedBody.GatewayCode)
	}
	if capturedBody.BillingInfoId == nil || *capturedBody.BillingInfoId != "bi-xyz" {
		t.Errorf("expected billing_info_id=bi-xyz, got %v", capturedBody.BillingInfoId)
	}
}

func TestSubscriptionsCreate_OnlySetFlagsAreSent(t *testing.T) {
	var capturedBody *recurly.SubscriptionCreate

	mock := &mockSubscriptionAPI{
		createSubscriptionFn: func(body *recurly.SubscriptionCreate, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedBody = body
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "create", "--plan-code", "gold")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.PlanCode == nil || *capturedBody.PlanCode != "gold" {
		t.Errorf("expected plan_code=gold, got %v", capturedBody.PlanCode)
	}
	if capturedBody.Account != nil {
		t.Error("expected account to be nil when not set")
	}
	if capturedBody.Currency != nil {
		t.Error("expected currency to be nil when not set")
	}
	if capturedBody.CollectionMethod != nil {
		t.Error("expected collection_method to be nil when not set")
	}
	if capturedBody.CouponCodes != nil {
		t.Error("expected coupon_codes to be nil when not set")
	}
}

func TestSubscriptionsCreate_TimestampFlags(t *testing.T) {
	var capturedBody *recurly.SubscriptionCreate

	mock := &mockSubscriptionAPI{
		createSubscriptionFn: func(body *recurly.SubscriptionCreate, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedBody = body
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "create",
		"--plan-code", "gold",
		"--trial-ends-at", "2025-06-01T00:00:00Z",
		"--starts-at", "2025-05-01T00:00:00Z",
		"--next-bill-date", "2025-07-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.TrialEndsAt == nil {
		t.Error("expected trial_ends_at to be set")
	}
	if capturedBody.StartsAt == nil {
		t.Error("expected starts_at to be set")
	}
	if capturedBody.NextBillDate == nil {
		t.Error("expected next_bill_date to be set")
	}
}

func TestSubscriptionsCreate_InvalidTimestamp_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand(nil, "subscriptions", "create", "--plan-code", "gold", "--trial-ends-at", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid timestamp")
	}
	if !strings.Contains(stderr, "invalid --trial-ends-at") {
		t.Errorf("expected 'invalid --trial-ends-at' error, got %q", stderr)
	}
}

func TestSubscriptionsCreate_SuccessOutput(t *testing.T) {
	mock := &mockSubscriptionAPI{
		createSubscriptionFn: func(body *recurly.SubscriptionCreate, opts ...recurly.Option) (*recurly.Subscription, error) {
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "create", "--plan-code", "gold")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"sub-123", "gold", "active"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestSubscriptionsCreate_ValidationError(t *testing.T) {
	mock := &mockSubscriptionAPI{
		createSubscriptionFn: func(body *recurly.SubscriptionCreate, opts ...recurly.Option) (*recurly.Subscription, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeValidation,
				Message: "The subscription could not be created",
				Params: []recurly.ErrorParam{
					{Property: "plan_code", Message: "is required"},
				},
			}
		},
	}
	app := newTestSubscriptionApp(mock)

	_, stderr, err := executeCommand(app, "subscriptions", "create")
	if err == nil {
		t.Fatal("expected error for validation failure")
	}
	if !strings.Contains(stderr, "plan_code") || !strings.Contains(stderr, "is required") {
		t.Errorf("expected validation error with field details, got %q", stderr)
	}
}

// --- subscriptions update ---

func TestSubscriptionsUpdate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "update") {
		t.Error("expected subscriptions help to show 'update' subcommand")
	}
}

func TestSubscriptionsUpdateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "update", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--collection-method", "--remaining-billing-cycles", "--renewal-billing-cycles", "--auto-renew", "--next-bill-date", "--revenue-schedule-type", "--terms-and-conditions", "--customer-notes", "--po-number", "--net-terms", "--net-terms-type", "--gateway-code", "--billing-info-id"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestSubscriptionsUpdate_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "subscriptions", "update")
	if err == nil {
		t.Fatal("expected error when no subscription ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestSubscriptionsUpdate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "subscriptions", "update", "sub-123", "--collection-method", "manual")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestSubscriptionsUpdate_PositionalArgAndFlagMapping(t *testing.T) {
	var capturedID string
	var capturedBody *recurly.SubscriptionUpdate

	mock := &mockSubscriptionAPI{
		updateSubscriptionFn: func(subscriptionId string, body *recurly.SubscriptionUpdate, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			capturedBody = body
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "update", "sub-456",
		"--collection-method", "manual",
		"--remaining-billing-cycles", "5",
		"--renewal-billing-cycles", "3",
		"--auto-renew",
		"--revenue-schedule-type", "evenly",
		"--terms-and-conditions", "New terms",
		"--customer-notes", "VIP customer",
		"--po-number", "PO-999",
		"--net-terms", "60",
		"--net-terms-type", "eom",
		"--gateway-code", "gw-def",
		"--billing-info-id", "bi-abc",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedID != "sub-456" {
		t.Errorf("expected subscription ID 'sub-456', got %q", capturedID)
	}
	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if capturedBody.CollectionMethod == nil || *capturedBody.CollectionMethod != "manual" {
		t.Errorf("expected collection_method=manual, got %v", capturedBody.CollectionMethod)
	}
	if capturedBody.RemainingBillingCycles == nil || *capturedBody.RemainingBillingCycles != 5 {
		t.Errorf("expected remaining_billing_cycles=5, got %v", capturedBody.RemainingBillingCycles)
	}
	if capturedBody.RenewalBillingCycles == nil || *capturedBody.RenewalBillingCycles != 3 {
		t.Errorf("expected renewal_billing_cycles=3, got %v", capturedBody.RenewalBillingCycles)
	}
	if capturedBody.AutoRenew == nil || *capturedBody.AutoRenew != true {
		t.Errorf("expected auto_renew=true, got %v", capturedBody.AutoRenew)
	}
	if capturedBody.RevenueScheduleType == nil || *capturedBody.RevenueScheduleType != "evenly" {
		t.Errorf("expected revenue_schedule_type=evenly, got %v", capturedBody.RevenueScheduleType)
	}
	if capturedBody.TermsAndConditions == nil || *capturedBody.TermsAndConditions != "New terms" {
		t.Errorf("expected terms_and_conditions='New terms', got %v", capturedBody.TermsAndConditions)
	}
	if capturedBody.CustomerNotes == nil || *capturedBody.CustomerNotes != "VIP customer" {
		t.Errorf("expected customer_notes='VIP customer', got %v", capturedBody.CustomerNotes)
	}
	if capturedBody.PoNumber == nil || *capturedBody.PoNumber != "PO-999" {
		t.Errorf("expected po_number=PO-999, got %v", capturedBody.PoNumber)
	}
	if capturedBody.NetTerms == nil || *capturedBody.NetTerms != 60 {
		t.Errorf("expected net_terms=60, got %v", capturedBody.NetTerms)
	}
	if capturedBody.NetTermsType == nil || *capturedBody.NetTermsType != "eom" {
		t.Errorf("expected net_terms_type=eom, got %v", capturedBody.NetTermsType)
	}
	if capturedBody.GatewayCode == nil || *capturedBody.GatewayCode != "gw-def" {
		t.Errorf("expected gateway_code=gw-def, got %v", capturedBody.GatewayCode)
	}
	if capturedBody.BillingInfoId == nil || *capturedBody.BillingInfoId != "bi-abc" {
		t.Errorf("expected billing_info_id=bi-abc, got %v", capturedBody.BillingInfoId)
	}
}

func TestSubscriptionsUpdate_OnlySetFlagsAreSent(t *testing.T) {
	var capturedBody *recurly.SubscriptionUpdate

	mock := &mockSubscriptionAPI{
		updateSubscriptionFn: func(subscriptionId string, body *recurly.SubscriptionUpdate, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedBody = body
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "update", "sub-456", "--collection-method", "manual")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.CollectionMethod == nil || *capturedBody.CollectionMethod != "manual" {
		t.Errorf("expected collection_method to be set, got %v", capturedBody.CollectionMethod)
	}
	if capturedBody.RemainingBillingCycles != nil {
		t.Error("expected remaining_billing_cycles to be nil when not set")
	}
	if capturedBody.RevenueScheduleType != nil {
		t.Error("expected revenue_schedule_type to be nil when not set")
	}
	if capturedBody.AutoRenew != nil {
		t.Error("expected auto_renew to be nil when not set")
	}
}

func TestSubscriptionsUpdate_NextBillDate(t *testing.T) {
	var capturedBody *recurly.SubscriptionUpdate

	mock := &mockSubscriptionAPI{
		updateSubscriptionFn: func(subscriptionId string, body *recurly.SubscriptionUpdate, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedBody = body
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "update", "sub-456", "--next-bill-date", "2025-06-01T00:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.NextBillDate == nil {
		t.Error("expected next_bill_date to be set")
	}
}

func TestSubscriptionsUpdate_InvalidNextBillDate_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand(nil, "subscriptions", "update", "sub-456", "--next-bill-date", "bad-date")
	if err == nil {
		t.Fatal("expected error for invalid next-bill-date")
	}
	if !strings.Contains(stderr, "invalid --next-bill-date") {
		t.Errorf("expected 'invalid --next-bill-date' error, got %q", stderr)
	}
}

func TestSubscriptionsUpdate_SuccessOutput(t *testing.T) {
	mock := &mockSubscriptionAPI{
		updateSubscriptionFn: func(subscriptionId string, body *recurly.SubscriptionUpdate, opts ...recurly.Option) (*recurly.Subscription, error) {
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "update", "sub-123", "--collection-method", "manual")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected output to contain subscription ID, got:\n%s", out)
	}
}

// --- subscriptions cancel ---

func TestSubscriptionsCancel_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cancel") {
		t.Error("expected subscriptions help to show 'cancel' subcommand")
	}
}

func TestSubscriptionsCancelHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "cancel", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--yes", "--timeframe"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestSubscriptionsCancel_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "subscriptions", "cancel")
	if err == nil {
		t.Fatal("expected error when no subscription ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestSubscriptionsCancel_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "cancel", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to cancel this subscription? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Cancellation cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestSubscriptionsCancel_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "cancel", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Cancellation cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestSubscriptionsCancel_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		cancelSubscriptionFn: func(subscriptionId string, params *recurly.CancelSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			sub := sampleSubscriptionDetail()
			sub.State = "canceled"
			return sub, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "subscriptions", "cancel", "sub-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-789" {
		t.Errorf("expected subscription ID 'sub-789', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsCancel_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		cancelSubscriptionFn: func(subscriptionId string, params *recurly.CancelSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			sub := sampleSubscriptionDetail()
			sub.State = "canceled"
			return sub, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, stderr, err := executeCommand(app, "subscriptions", "cancel", "sub-789", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-789" {
		t.Errorf("expected subscription ID 'sub-789', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsCancel_TimeframeFlag(t *testing.T) {
	var capturedParams *recurly.CancelSubscriptionParams

	mock := &mockSubscriptionAPI{
		cancelSubscriptionFn: func(subscriptionId string, params *recurly.CancelSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedParams = params
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "cancel", "sub-123", "--yes", "--timeframe", "term_end")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams == nil || capturedParams.Body == nil {
		t.Fatal("expected params with body to be captured")
	}
	if capturedParams.Body.Timeframe == nil || *capturedParams.Body.Timeframe != "term_end" {
		t.Errorf("expected timeframe=term_end, got %v", capturedParams.Body.Timeframe)
	}
}

func TestSubscriptionsCancel_SDKError(t *testing.T) {
	mock := &mockSubscriptionAPI{
		cancelSubscriptionFn: func(subscriptionId string, params *recurly.CancelSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Subscription",
			}
		},
	}
	app := newTestSubscriptionApp(mock)

	_, stderr, err := executeCommand(app, "subscriptions", "cancel", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

func TestSubscriptionsCancel_JSONOutput(t *testing.T) {
	mock := &mockSubscriptionAPI{
		cancelSubscriptionFn: func(subscriptionId string, params *recurly.CancelSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "cancel", "sub-123", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["id"] != "sub-123" {
		t.Errorf("expected id=sub-123 in JSON, got %v", result["id"])
	}
}

// --- subscriptions reactivate ---

func TestSubscriptionsReactivate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "reactivate") {
		t.Error("expected subscriptions help to show 'reactivate' subcommand")
	}
}

func TestSubscriptionsReactivateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "reactivate", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--yes") {
		t.Error("expected help output to contain --yes flag")
	}
}

func TestSubscriptionsReactivate_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "subscriptions", "reactivate")
	if err == nil {
		t.Fatal("expected error when no subscription ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestSubscriptionsReactivate_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "reactivate", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to reactivate this subscription? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Reactivation cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestSubscriptionsReactivate_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "reactivate", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Reactivation cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestSubscriptionsReactivate_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		reactivateSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	stdin := bytes.NewBufferString("yes\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "subscriptions", "reactivate", "sub-closed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-closed" {
		t.Errorf("expected subscription ID 'sub-closed', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsReactivate_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		reactivateSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, stderr, err := executeCommand(app, "subscriptions", "reactivate", "sub-closed", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-closed" {
		t.Errorf("expected subscription ID 'sub-closed', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsReactivate_SDKError(t *testing.T) {
	mock := &mockSubscriptionAPI{
		reactivateSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Subscription",
			}
		},
	}
	app := newTestSubscriptionApp(mock)

	_, stderr, err := executeCommand(app, "subscriptions", "reactivate", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

func TestSubscriptionsReactivate_JSONOutput(t *testing.T) {
	mock := &mockSubscriptionAPI{
		reactivateSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "reactivate", "sub-123", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["id"] != "sub-123" {
		t.Errorf("expected id=sub-123 in JSON, got %v", result["id"])
	}
}

// --- subscriptions pause ---

func TestSubscriptionsPause_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "pause") {
		t.Error("expected subscriptions help to show 'pause' subcommand")
	}
}

func TestSubscriptionsPauseHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "pause", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--yes", "--remaining-pause-cycles"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestSubscriptionsPause_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "subscriptions", "pause", "--remaining-pause-cycles", "3")
	if err == nil {
		t.Fatal("expected error when no subscription ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestSubscriptionsPause_MissingRequiredFlag_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "subscriptions", "pause", "sub-123", "--yes", "--no-input")
	if err == nil {
		t.Fatal("expected error when required flag is missing")
	}
	if !strings.Contains(stderr, "remaining-pause-cycles") {
		t.Errorf("expected error about required flag, got %q", stderr)
	}
}

func TestSubscriptionsPause_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "pause", "sub-123", "--remaining-pause-cycles", "3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to pause this subscription? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Pause cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestSubscriptionsPause_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "pause", "sub-123", "--remaining-pause-cycles", "3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Pause cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestSubscriptionsPause_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	var capturedBody *recurly.SubscriptionPause

	mock := &mockSubscriptionAPI{
		pauseSubscriptionFn: func(subscriptionId string, body *recurly.SubscriptionPause, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			capturedBody = body
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "subscriptions", "pause", "sub-active", "--remaining-pause-cycles", "3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-active" {
		t.Errorf("expected subscription ID 'sub-active', got %q", capturedID)
	}
	if capturedBody == nil || capturedBody.RemainingPauseCycles == nil || *capturedBody.RemainingPauseCycles != 3 {
		t.Error("expected remaining_pause_cycles=3")
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsPause_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		pauseSubscriptionFn: func(subscriptionId string, body *recurly.SubscriptionPause, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, stderr, err := executeCommand(app, "subscriptions", "pause", "sub-active", "--yes", "--remaining-pause-cycles", "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-active" {
		t.Errorf("expected subscription ID 'sub-active', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsPause_SDKError(t *testing.T) {
	mock := &mockSubscriptionAPI{
		pauseSubscriptionFn: func(subscriptionId string, body *recurly.SubscriptionPause, opts ...recurly.Option) (*recurly.Subscription, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Subscription",
			}
		},
	}
	app := newTestSubscriptionApp(mock)

	_, stderr, err := executeCommand(app, "subscriptions", "pause", "nonexistent", "--yes", "--remaining-pause-cycles", "1")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

// --- subscriptions resume ---

func TestSubscriptionsResume_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "resume") {
		t.Error("expected subscriptions help to show 'resume' subcommand")
	}
}

func TestSubscriptionsResumeHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "resume", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--yes") {
		t.Error("expected help output to contain --yes flag")
	}
}

func TestSubscriptionsResume_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "subscriptions", "resume")
	if err == nil {
		t.Fatal("expected error when no subscription ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestSubscriptionsResume_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "resume", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to resume this subscription? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Resume cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestSubscriptionsResume_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "resume", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Resume cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestSubscriptionsResume_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		resumeSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	stdin := bytes.NewBufferString("yes\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "subscriptions", "resume", "sub-paused")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-paused" {
		t.Errorf("expected subscription ID 'sub-paused', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsResume_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		resumeSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, stderr, err := executeCommand(app, "subscriptions", "resume", "sub-paused", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-paused" {
		t.Errorf("expected subscription ID 'sub-paused', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsResume_SDKError(t *testing.T) {
	mock := &mockSubscriptionAPI{
		resumeSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Subscription",
			}
		},
	}
	app := newTestSubscriptionApp(mock)

	_, stderr, err := executeCommand(app, "subscriptions", "resume", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

func TestSubscriptionsResume_JSONOutput(t *testing.T) {
	mock := &mockSubscriptionAPI{
		resumeSubscriptionFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "resume", "sub-123", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["id"] != "sub-123" {
		t.Errorf("expected id=sub-123 in JSON, got %v", result["id"])
	}
}

// --- subscriptions terminate ---

func TestSubscriptionsTerminate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "terminate") {
		t.Error("expected subscriptions help to show 'terminate' subcommand")
	}
}

func TestSubscriptionsTerminateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "terminate", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--yes", "--refund", "--charge"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestSubscriptionsTerminate_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "subscriptions", "terminate")
	if err == nil {
		t.Fatal("expected error when no subscription ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestSubscriptionsTerminate_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "terminate", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to terminate this subscription? This cannot be undone. [y/N]") {
		t.Error("expected confirmation prompt with warning in stderr")
	}
	if !strings.Contains(stderr, "Termination cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestSubscriptionsTerminate_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "terminate", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Termination cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestSubscriptionsTerminate_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		terminateSubscriptionFn: func(subscriptionId string, params *recurly.TerminateSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			sub := sampleSubscriptionDetail()
			sub.State = "expired"
			return sub, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "subscriptions", "terminate", "sub-active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-active" {
		t.Errorf("expected subscription ID 'sub-active', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsTerminate_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		terminateSubscriptionFn: func(subscriptionId string, params *recurly.TerminateSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			sub := sampleSubscriptionDetail()
			sub.State = "expired"
			return sub, nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, stderr, err := executeCommand(app, "subscriptions", "terminate", "sub-active", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-active" {
		t.Errorf("expected subscription ID 'sub-active', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsTerminate_RefundAndChargeFlags(t *testing.T) {
	var capturedParams *recurly.TerminateSubscriptionParams

	mock := &mockSubscriptionAPI{
		terminateSubscriptionFn: func(subscriptionId string, params *recurly.TerminateSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedParams = params
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	_, _, err := executeCommand(app, "subscriptions", "terminate", "sub-123", "--yes", "--refund", "full", "--charge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams == nil {
		t.Fatal("expected params to be captured")
	}
	if capturedParams.Refund == nil || *capturedParams.Refund != "full" {
		t.Errorf("expected refund=full, got %v", capturedParams.Refund)
	}
	if capturedParams.Charge == nil || *capturedParams.Charge != true {
		t.Errorf("expected charge=true, got %v", capturedParams.Charge)
	}
}

func TestSubscriptionsTerminate_SDKError(t *testing.T) {
	mock := &mockSubscriptionAPI{
		terminateSubscriptionFn: func(subscriptionId string, params *recurly.TerminateSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Subscription",
			}
		},
	}
	app := newTestSubscriptionApp(mock)

	_, stderr, err := executeCommand(app, "subscriptions", "terminate", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

func TestSubscriptionsTerminate_JSONOutput(t *testing.T) {
	mock := &mockSubscriptionAPI{
		terminateSubscriptionFn: func(subscriptionId string, params *recurly.TerminateSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error) {
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "terminate", "sub-123", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["id"] != "sub-123" {
		t.Errorf("expected id=sub-123 in JSON, got %v", result["id"])
	}
}

// --- subscriptions convert-trial ---

func TestSubscriptionsConvertTrial_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "convert-trial") {
		t.Error("expected subscriptions help to show 'convert-trial' subcommand")
	}
}

func TestSubscriptionsConvertTrialHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "subscriptions", "convert-trial", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--yes") {
		t.Error("expected help output to contain --yes flag")
	}
}

func TestSubscriptionsConvertTrial_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "subscriptions", "convert-trial")
	if err == nil {
		t.Fatal("expected error when no subscription ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestSubscriptionsConvertTrial_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "convert-trial", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to convert this trial subscription? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Conversion cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestSubscriptionsConvertTrial_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "subscriptions", "convert-trial", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Conversion cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestSubscriptionsConvertTrial_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		convertTrialFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "subscriptions", "convert-trial", "sub-trial")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-trial" {
		t.Errorf("expected subscription ID 'sub-trial', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsConvertTrial_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedID string
	mock := &mockSubscriptionAPI{
		convertTrialFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			capturedID = subscriptionId
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, stderr, err := executeCommand(app, "subscriptions", "convert-trial", "sub-trial", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "sub-trial" {
		t.Errorf("expected subscription ID 'sub-trial', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected subscription details in output, got:\n%s", out)
	}
}

func TestSubscriptionsConvertTrial_SDKError(t *testing.T) {
	mock := &mockSubscriptionAPI{
		convertTrialFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Subscription",
			}
		},
	}
	app := newTestSubscriptionApp(mock)

	_, stderr, err := executeCommand(app, "subscriptions", "convert-trial", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

func TestSubscriptionsConvertTrial_JSONOutput(t *testing.T) {
	mock := &mockSubscriptionAPI{
		convertTrialFn: func(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error) {
			return sampleSubscriptionDetail(), nil
		},
	}
	app := newTestSubscriptionApp(mock)

	out, _, err := executeCommand(app, "subscriptions", "convert-trial", "sub-123", "--yes", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["id"] != "sub-123" {
		t.Errorf("expected id=sub-123 in JSON, got %v", result["id"])
	}
}
