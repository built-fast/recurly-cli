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

// mockAccountNestedAPI implements AccountNestedAPI for testing.
type mockAccountNestedAPI struct {
	listAccountSubscriptionsFn func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error)
	listAccountInvoicesFn      func(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error)
	listAccountTransactionsFn  func(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error)
}

func (m *mockAccountNestedAPI) ListAccountSubscriptions(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
	if m.listAccountSubscriptionsFn != nil {
		return m.listAccountSubscriptionsFn(accountId, params, opts...)
	}
	return nil, nil
}

func (m *mockAccountNestedAPI) ListAccountInvoices(accountId string, params *recurly.ListAccountInvoicesParams, opts ...recurly.Option) (recurly.InvoiceLister, error) {
	if m.listAccountInvoicesFn != nil {
		return m.listAccountInvoicesFn(accountId, params, opts...)
	}
	return nil, nil
}

func (m *mockAccountNestedAPI) ListAccountTransactions(accountId string, params *recurly.ListAccountTransactionsParams, opts ...recurly.Option) (recurly.TransactionLister, error) {
	if m.listAccountTransactionsFn != nil {
		return m.listAccountTransactionsFn(accountId, params, opts...)
	}
	return nil, nil
}

// mockAccountSubscriptionLister implements recurly.SubscriptionLister for testing.
type mockAccountSubscriptionLister struct {
	subscriptions []recurly.Subscription
	fetched       bool
}

func (m *mockAccountSubscriptionLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockAccountSubscriptionLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockAccountSubscriptionLister) Count() (*int64, error) {
	n := int64(len(m.subscriptions))
	return &n, nil
}

func (m *mockAccountSubscriptionLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockAccountSubscriptionLister) Data() []recurly.Subscription {
	return m.subscriptions
}

func (m *mockAccountSubscriptionLister) HasMore() bool {
	return !m.fetched
}

func (m *mockAccountSubscriptionLister) Next() string {
	return ""
}

func setMockAccountNestedAPI(mock *mockAccountNestedAPI) func() {
	orig := newAccountNestedAPI
	newAccountNestedAPI = func() (AccountNestedAPI, error) {
		return mock, nil
	}
	return func() { newAccountNestedAPI = orig }
}

func sampleAccountSubscription() recurly.Subscription {
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

// --- accounts subscriptions list ---

func TestAccountSubscriptions_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "subscriptions") {
		t.Error("expected accounts help to show 'subscriptions' subcommand")
	}
}

func TestAccountSubscriptionsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("accounts", "subscriptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected subscriptions help to show 'list' subcommand")
	}
}

func TestAccountSubscriptionsListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand("accounts", "subscriptions", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--order", "--sort", "--state"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestAccountSubscriptionsList_RequiresAccountID(t *testing.T) {
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "subscriptions", "list")
	if err == nil {
		t.Fatal("expected error when account_id is missing")
	}
}

func TestAccountSubscriptionsList_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand("accounts", "subscriptions", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountSubscriptionsList_PassesAccountID(t *testing.T) {
	var capturedAccountID string

	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedAccountID = accountId
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAccountID != "acct-456" {
		t.Errorf("expected accountId=acct-456, got %q", capturedAccountID)
	}
}

func TestAccountSubscriptionsList_PaginationParams(t *testing.T) {
	var capturedParams *recurly.ListAccountSubscriptionsParams

	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedParams = params
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456", "--limit", "50", "--order", "desc", "--sort", "updated_at")
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

func TestAccountSubscriptionsList_FilterParams(t *testing.T) {
	var capturedParams *recurly.ListAccountSubscriptionsParams

	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedParams = params
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456", "--state", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.State == nil || *capturedParams.State != "active" {
		t.Errorf("expected state=active, got %v", capturedParams.State)
	}
}

func TestAccountSubscriptionsList_UnsetFlagsNotSent(t *testing.T) {
	var capturedParams *recurly.ListAccountSubscriptionsParams

	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			capturedParams = params
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.Limit != nil {
		t.Error("expected Limit to be nil when not set")
	}
	if capturedParams.Order != nil {
		t.Error("expected Order to be nil when not set")
	}
	if capturedParams.Sort != nil {
		t.Error("expected Sort to be nil when not set")
	}
	if capturedParams.State != nil {
		t.Error("expected State to be nil when not set")
	}
}

func TestAccountSubscriptionsList_TableOutput(t *testing.T) {
	sub := sampleAccountSubscription()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{sub}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"sub-123", "acct-456", "gold", "active", "USD", "19.99"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	for _, header := range []string{"ID", "Account Code", "Plan Code", "State", "Currency", "Unit Amount", "Current Period Ends At"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestAccountSubscriptionsList_JSONOutput(t *testing.T) {
	sub := sampleAccountSubscription()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{sub}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456")
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
		t.Errorf("expected 1 item in data, got %d", len(result.Data))
	}
}

func TestAccountSubscriptionsList_JSONPrettyOutput(t *testing.T) {
	sub := sampleAccountSubscription()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{sub}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json-pretty")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "sub-123") {
		t.Error("expected JSON-pretty output to contain subscription ID")
	}
	if !strings.Contains(out, "\n") {
		t.Error("expected JSON-pretty output to be indented with newlines")
	}
}

func TestAccountSubscriptionsList_JQFilter(t *testing.T) {
	sub := sampleAccountSubscription()
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{sub}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	viper.Set("jq", ".data[0].id")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "sub-123") {
		t.Errorf("expected jq output to contain 'sub-123', got: %s", out)
	}
}

func TestAccountSubscriptionsList_SDKError(t *testing.T) {
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestAccountSubscriptionsList_EmptyResults(t *testing.T) {
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ID") {
		t.Error("expected table headers even for empty results")
	}
}

func TestAccountSubscriptionsList_EmptyResults_JSON(t *testing.T) {
	mock := &mockAccountNestedAPI{
		listAccountSubscriptionsFn: func(accountId string, params *recurly.ListAccountSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error) {
			return &mockAccountSubscriptionLister{subscriptions: []recurly.Subscription{}}, nil
		},
	}
	cleanup := setMockAccountNestedAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "subscriptions", "list", "acct-456")
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
		t.Errorf("expected 0 items in data, got %d", len(result.Data))
	}
}
