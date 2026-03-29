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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mockAccountAPI implements AccountAPI for testing.
type mockAccountAPI struct {
	listAccountsFn      func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error)
	getAccountFn        func(accountId string, opts ...recurly.Option) (*recurly.Account, error)
	createAccountFn     func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error)
	updateAccountFn     func(accountId string, body *recurly.AccountUpdate, opts ...recurly.Option) (*recurly.Account, error)
	deactivateAccountFn func(accountId string, opts ...recurly.Option) (*recurly.Account, error)
	reactivateAccountFn func(accountId string, opts ...recurly.Option) (*recurly.Account, error)
}

func (m *mockAccountAPI) ListAccounts(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
	return m.listAccountsFn(params, opts...)
}

func (m *mockAccountAPI) GetAccount(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
	return m.getAccountFn(accountId, opts...)
}

func (m *mockAccountAPI) CreateAccount(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
	return m.createAccountFn(body, opts...)
}

func (m *mockAccountAPI) UpdateAccount(accountId string, body *recurly.AccountUpdate, opts ...recurly.Option) (*recurly.Account, error) {
	return m.updateAccountFn(accountId, body, opts...)
}

func (m *mockAccountAPI) DeactivateAccount(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
	return m.deactivateAccountFn(accountId, opts...)
}

func (m *mockAccountAPI) ReactivateAccount(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
	return m.reactivateAccountFn(accountId, opts...)
}

// mockAccountLister implements recurly.AccountLister for testing.
type mockAccountLister struct {
	accounts []recurly.Account
	fetched  bool
}

func (m *mockAccountLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockAccountLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockAccountLister) Count() (*int64, error) {
	n := int64(len(m.accounts))
	return &n, nil
}

func (m *mockAccountLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockAccountLister) Data() []recurly.Account {
	return m.accounts
}

func (m *mockAccountLister) HasMore() bool {
	return !m.fetched
}

func (m *mockAccountLister) Next() string {
	return ""
}


// sampleAccount returns a test account with predictable fields.
func sampleAccount() *recurly.Account {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return &recurly.Account{
		Code:      "acct-123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Company:   "Acme Inc",
		State:     "active",
		CreatedAt: &now,
		UpdatedAt: &now,
	}
}

// --- accounts list ---

func TestAccountsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected accounts help to show 'list' subcommand")
	}
}

func TestAccountsListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--order", "--sort", "--email", "--subscriber", "--past-due", "--begin-time", "--end-time"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestAccountsList_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "list")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountsList_InvalidBeginTime_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand(nil, "accounts", "list", "--begin-time", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid begin-time")
	}
	if !strings.Contains(stderr, "invalid --begin-time") {
		t.Errorf("expected 'invalid --begin-time' error, got %q", stderr)
	}
}

func TestAccountsList_InvalidEndTime_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand(nil, "accounts", "list", "--end-time", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid end-time")
	}
	if !strings.Contains(stderr, "invalid --end-time") {
		t.Errorf("expected 'invalid --end-time' error, got %q", stderr)
	}
}

func TestAccountsList_PaginationParams(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountsParams

	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			capturedParams = params
			return &mockAccountLister{accounts: []recurly.Account{}}, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "list", "--limit", "50", "--order", "desc", "--sort", "updated_at")
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

func TestAccountsList_FilterParams(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountsParams

	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			capturedParams = params
			return &mockAccountLister{accounts: []recurly.Account{}}, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "list",
		"--email", "user@example.com",
		"--subscriber", "true",
		"--past-due",
		"--begin-time", "2025-01-01T00:00:00Z",
		"--end-time", "2025-12-31T23:59:59Z",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams.Email == nil || *capturedParams.Email != "user@example.com" {
		t.Errorf("expected email=user@example.com, got %v", capturedParams.Email)
	}
	if capturedParams.Subscriber == nil || *capturedParams.Subscriber != true {
		t.Errorf("expected subscriber=true, got %v", capturedParams.Subscriber)
	}
	if capturedParams.PastDue == nil || *capturedParams.PastDue != "true" {
		t.Errorf("expected past_due=true, got %v", capturedParams.PastDue)
	}
	if capturedParams.BeginTime == nil {
		t.Error("expected begin_time to be set")
	}
	if capturedParams.EndTime == nil {
		t.Error("expected end_time to be set")
	}
}

func TestAccountsList_UnsetFlagsNotSent(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountsParams

	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			capturedParams = params
			return &mockAccountLister{accounts: []recurly.Account{}}, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "list")
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
	if capturedParams.Email != nil {
		t.Error("expected email to be nil when not set")
	}
	if capturedParams.Subscriber != nil {
		t.Error("expected subscriber to be nil when not set")
	}
	if capturedParams.PastDue != nil {
		t.Error("expected past_due to be nil when not set")
	}
}

func TestAccountsList_TableOutput(t *testing.T) {
	acct := sampleAccount()
	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			return &mockAccountLister{accounts: []recurly.Account{*acct}}, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"acct-123", "test@example.com", "John", "Doe", "Acme Inc", "active"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	// Table should have column headers
	for _, header := range []string{"Code", "Email", "First Name", "Last Name", "Company", "State"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestAccountsList_JSONOutput(t *testing.T) {
	acct := sampleAccount()
	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			return &mockAccountLister{accounts: []recurly.Account{*acct}}, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "list", "--output", "json")
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
		t.Fatalf("expected 1 account in JSON output, got %d", len(envelope.Data))
	}
	if envelope.Data[0]["code"] != "acct-123" {
		t.Errorf("expected code=acct-123 in JSON, got %v", envelope.Data[0]["code"])
	}
}

func TestAccountsList_SDKError(t *testing.T) {
	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "list")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- accounts get ---

func TestAccountsGet_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "get") {
		t.Error("expected accounts help to show 'get' subcommand")
	}
}

func TestAccountsGet_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "accounts", "get")
	if err == nil {
		t.Fatal("expected error when no account ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestAccountsGet_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "get", "abc123")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountsGet_PositionalArg(t *testing.T) {
	var capturedID string
	mock := &mockAccountAPI{
		getAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			capturedID = accountId
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "get", "my-account-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "my-account-code" {
		t.Errorf("expected account ID 'my-account-code', got %q", capturedID)
	}
}

func TestAccountsGet_TableOutput(t *testing.T) {
	mock := &mockAccountAPI{
		getAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "get", "acct-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// FormatOne renders key-value table with "Field" and "Value" headers
	for _, expected := range []string{"Code", "acct-123", "Email", "test@example.com", "First Name", "John", "Last Name", "Doe", "Company", "Acme Inc", "State", "active"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestAccountsGet_JSONOutput(t *testing.T) {
	mock := &mockAccountAPI{
		getAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "get", "acct-123", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if result["code"] != "acct-123" {
		t.Errorf("expected code=acct-123 in JSON, got %v", result["code"])
	}
}

func TestAccountsGet_NotFound(t *testing.T) {
	mock := &mockAccountAPI{
		getAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Account with id = nonexistent",
			}
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, stderr, err := executeCommand(app, "accounts", "get", "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found account")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

// --- accounts create ---

func TestAccountsCreate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "create") {
		t.Error("expected accounts help to show 'create' subcommand")
	}
}

func TestAccountsCreateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "create", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--code", "--email", "--first-name", "--last-name", "--company", "--vat-number", "--tax-exempt", "--preferred-locale", "--bill-to"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestAccountsCreate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "create", "--code", "test")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountsCreate_FlagToStructMapping(t *testing.T) {
	var capturedBody *recurly.AccountCreate

	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			capturedBody = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "create",
		"--code", "new-acct",
		"--email", "new@example.com",
		"--first-name", "Jane",
		"--last-name", "Smith",
		"--company", "NewCo",
		"--vat-number", "VAT123",
		"--tax-exempt",
		"--preferred-locale", "en-US",
		"--bill-to", "self",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if capturedBody.Code == nil || *capturedBody.Code != "new-acct" {
		t.Errorf("expected code=new-acct, got %v", capturedBody.Code)
	}
	if capturedBody.Email == nil || *capturedBody.Email != "new@example.com" {
		t.Errorf("expected email=new@example.com, got %v", capturedBody.Email)
	}
	if capturedBody.FirstName == nil || *capturedBody.FirstName != "Jane" {
		t.Errorf("expected first_name=Jane, got %v", capturedBody.FirstName)
	}
	if capturedBody.LastName == nil || *capturedBody.LastName != "Smith" {
		t.Errorf("expected last_name=Smith, got %v", capturedBody.LastName)
	}
	if capturedBody.Company == nil || *capturedBody.Company != "NewCo" {
		t.Errorf("expected company=NewCo, got %v", capturedBody.Company)
	}
	if capturedBody.VatNumber == nil || *capturedBody.VatNumber != "VAT123" {
		t.Errorf("expected vat_number=VAT123, got %v", capturedBody.VatNumber)
	}
	if capturedBody.TaxExempt == nil || *capturedBody.TaxExempt != true {
		t.Errorf("expected tax_exempt=true, got %v", capturedBody.TaxExempt)
	}
	if capturedBody.PreferredLocale == nil || *capturedBody.PreferredLocale != "en-US" {
		t.Errorf("expected preferred_locale=en-US, got %v", capturedBody.PreferredLocale)
	}
	if capturedBody.BillTo == nil || *capturedBody.BillTo != "self" {
		t.Errorf("expected bill_to=self, got %v", capturedBody.BillTo)
	}
}

func TestAccountsCreate_OnlySetFlagsAreSent(t *testing.T) {
	var capturedBody *recurly.AccountCreate

	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			capturedBody = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "create", "--code", "minimal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Code == nil || *capturedBody.Code != "minimal" {
		t.Errorf("expected code=minimal, got %v", capturedBody.Code)
	}
	if capturedBody.Email != nil {
		t.Error("expected email to be nil when not set")
	}
	if capturedBody.FirstName != nil {
		t.Error("expected first_name to be nil when not set")
	}
	if capturedBody.LastName != nil {
		t.Error("expected last_name to be nil when not set")
	}
	if capturedBody.Company != nil {
		t.Error("expected company to be nil when not set")
	}
	if capturedBody.TaxExempt != nil {
		t.Error("expected tax_exempt to be nil when not set")
	}
}

func TestAccountsCreate_SuccessOutput(t *testing.T) {
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "create", "--code", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"acct-123", "test@example.com", "active"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestAccountsCreate_ValidationError(t *testing.T) {
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeValidation,
				Message: "The account could not be created",
				Params: []recurly.ErrorParam{
					{Property: "code", Message: "is already taken"},
				},
			}
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, stderr, err := executeCommand(app, "accounts", "create", "--code", "existing")
	if err == nil {
		t.Fatal("expected error for validation failure")
	}
	if !strings.Contains(stderr, "code") || !strings.Contains(stderr, "is already taken") {
		t.Errorf("expected validation error with field details, got %q", stderr)
	}
}

// --- accounts update ---

func TestAccountsUpdate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "update") {
		t.Error("expected accounts help to show 'update' subcommand")
	}
}

func TestAccountsUpdateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "update", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--email", "--first-name", "--last-name", "--company", "--vat-number", "--tax-exempt", "--preferred-locale", "--bill-to"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestAccountsUpdate_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "accounts", "update")
	if err == nil {
		t.Fatal("expected error when no account ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestAccountsUpdate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "update", "abc123", "--email", "new@example.com")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountsUpdate_PositionalArgAndFlagMapping(t *testing.T) {
	var capturedID string
	var capturedBody *recurly.AccountUpdate

	mock := &mockAccountAPI{
		updateAccountFn: func(accountId string, body *recurly.AccountUpdate, opts ...recurly.Option) (*recurly.Account, error) {
			capturedID = accountId
			capturedBody = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "update", "acct-456",
		"--email", "updated@example.com",
		"--first-name", "Updated",
		"--last-name", "Name",
		"--company", "UpdatedCo",
		"--vat-number", "VAT456",
		"--tax-exempt",
		"--preferred-locale", "fr-FR",
		"--bill-to", "parent",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedID != "acct-456" {
		t.Errorf("expected account ID 'acct-456', got %q", capturedID)
	}
	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if capturedBody.Email == nil || *capturedBody.Email != "updated@example.com" {
		t.Errorf("expected email=updated@example.com, got %v", capturedBody.Email)
	}
	if capturedBody.FirstName == nil || *capturedBody.FirstName != "Updated" {
		t.Errorf("expected first_name=Updated, got %v", capturedBody.FirstName)
	}
	if capturedBody.LastName == nil || *capturedBody.LastName != "Name" {
		t.Errorf("expected last_name=Name, got %v", capturedBody.LastName)
	}
	if capturedBody.Company == nil || *capturedBody.Company != "UpdatedCo" {
		t.Errorf("expected company=UpdatedCo, got %v", capturedBody.Company)
	}
	if capturedBody.VatNumber == nil || *capturedBody.VatNumber != "VAT456" {
		t.Errorf("expected vat_number=VAT456, got %v", capturedBody.VatNumber)
	}
	if capturedBody.TaxExempt == nil || *capturedBody.TaxExempt != true {
		t.Errorf("expected tax_exempt=true, got %v", capturedBody.TaxExempt)
	}
	if capturedBody.PreferredLocale == nil || *capturedBody.PreferredLocale != "fr-FR" {
		t.Errorf("expected preferred_locale=fr-FR, got %v", capturedBody.PreferredLocale)
	}
	if capturedBody.BillTo == nil || *capturedBody.BillTo != "parent" {
		t.Errorf("expected bill_to=parent, got %v", capturedBody.BillTo)
	}
}

func TestAccountsUpdate_OnlySetFlagsAreSent(t *testing.T) {
	var capturedBody *recurly.AccountUpdate

	mock := &mockAccountAPI{
		updateAccountFn: func(accountId string, body *recurly.AccountUpdate, opts ...recurly.Option) (*recurly.Account, error) {
			capturedBody = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "update", "acct-456", "--email", "only-email@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Email == nil || *capturedBody.Email != "only-email@example.com" {
		t.Errorf("expected email to be set, got %v", capturedBody.Email)
	}
	if capturedBody.FirstName != nil {
		t.Error("expected first_name to be nil when not set")
	}
	if capturedBody.Company != nil {
		t.Error("expected company to be nil when not set")
	}
	if capturedBody.TaxExempt != nil {
		t.Error("expected tax_exempt to be nil when not set")
	}
}

func TestAccountsUpdate_SuccessOutput(t *testing.T) {
	mock := &mockAccountAPI{
		updateAccountFn: func(accountId string, body *recurly.AccountUpdate, opts ...recurly.Option) (*recurly.Account, error) {
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "update", "acct-123", "--email", "new@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "acct-123") {
		t.Errorf("expected output to contain account code, got:\n%s", out)
	}
}

func TestAccountsUpdate_ValidationError(t *testing.T) {
	mock := &mockAccountAPI{
		updateAccountFn: func(accountId string, body *recurly.AccountUpdate, opts ...recurly.Option) (*recurly.Account, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeValidation,
				Message: "The account could not be updated",
				Params: []recurly.ErrorParam{
					{Property: "email", Message: "is not a valid email"},
				},
			}
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, stderr, err := executeCommand(app, "accounts", "update", "acct-123", "--email", "bad-email")
	if err == nil {
		t.Fatal("expected error for validation failure")
	}
	if !strings.Contains(stderr, "email") || !strings.Contains(stderr, "is not a valid email") {
		t.Errorf("expected validation error with field details, got %q", stderr)
	}
}

// --- accounts deactivate ---

func TestAccountsDeactivate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "deactivate") {
		t.Error("expected accounts help to show 'deactivate' subcommand")
	}
}

func TestAccountsDeactivateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "deactivate", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--yes") {
		t.Error("expected help output to contain --yes flag")
	}
}

func TestAccountsDeactivate_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "accounts", "deactivate")
	if err == nil {
		t.Fatal("expected error when no account ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestAccountsDeactivate_NoAPIKey_WithYes_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "deactivate", "abc123", "--yes")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountsDeactivate_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "accounts", "deactivate", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to deactivate this account? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Deactivation cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestAccountsDeactivate_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "accounts", "deactivate", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deactivation cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestAccountsDeactivate_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockAccountAPI{
		deactivateAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			capturedID = accountId
			acct := sampleAccount()
			acct.State = "closed"
			return acct, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "accounts", "deactivate", "acct-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "acct-789" {
		t.Errorf("expected account ID 'acct-789', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "acct-123") {
		t.Errorf("expected account details in output, got:\n%s", out)
	}
}

func TestAccountsDeactivate_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedID string
	mock := &mockAccountAPI{
		deactivateAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			capturedID = accountId
			acct := sampleAccount()
			acct.State = "closed"
			return acct, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, stderr, err := executeCommand(app, "accounts", "deactivate", "acct-789", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "acct-789" {
		t.Errorf("expected account ID 'acct-789', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "acct-123") {
		t.Errorf("expected account details in output, got:\n%s", out)
	}
}

func TestAccountsDeactivate_SDKError(t *testing.T) {
	mock := &mockAccountAPI{
		deactivateAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Account",
			}
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, stderr, err := executeCommand(app, "accounts", "deactivate", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

// --- accounts reactivate ---

func TestAccountsReactivate_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "reactivate") {
		t.Error("expected accounts help to show 'reactivate' subcommand")
	}
}

func TestAccountsReactivateHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "reactivate", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--yes") {
		t.Error("expected help output to contain --yes flag")
	}
}

func TestAccountsReactivate_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand(nil, "accounts", "reactivate")
	if err == nil {
		t.Fatal("expected error when no account ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestAccountsReactivate_NoAPIKey_WithYes_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "reactivate", "abc123", "--yes")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountsReactivate_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "accounts", "reactivate", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to reactivate this account? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Reactivation cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestAccountsReactivate_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "accounts", "reactivate", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Reactivation cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestAccountsReactivate_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockAccountAPI{
		reactivateAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			capturedID = accountId
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	stdin := bytes.NewBufferString("yes\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "accounts", "reactivate", "acct-closed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "acct-closed" {
		t.Errorf("expected account ID 'acct-closed', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "acct-123") {
		t.Errorf("expected account details in output, got:\n%s", out)
	}
}

func TestAccountsReactivate_YesFlag_SkipsPrompt(t *testing.T) {
	var capturedID string
	mock := &mockAccountAPI{
		reactivateAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			capturedID = accountId
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, stderr, err := executeCommand(app, "accounts", "reactivate", "acct-closed", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "acct-closed" {
		t.Errorf("expected account ID 'acct-closed', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "acct-123") {
		t.Errorf("expected account details in output, got:\n%s", out)
	}
}

func TestAccountsReactivate_SDKError(t *testing.T) {
	mock := &mockAccountAPI{
		reactivateAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Account",
			}
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, stderr, err := executeCommand(app, "accounts", "reactivate", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

// --- --field flag ---

func TestAccountsList_FieldFlag_TableOutput(t *testing.T) {
	acct := sampleAccount()
	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			return &mockAccountLister{accounts: []recurly.Account{*acct}}, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "list", "--field", "Code,Email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Code") {
		t.Error("expected table to contain 'Code' header")
	}
	if !strings.Contains(out, "Email") {
		t.Error("expected table to contain 'Email' header")
	}
	if !strings.Contains(out, "acct-123") {
		t.Error("expected table to contain account code")
	}
	if !strings.Contains(out, "test@example.com") {
		t.Error("expected table to contain email")
	}
	// Fields not selected should be absent
	if strings.Contains(out, "First Name") {
		t.Error("table should NOT contain 'First Name' when not selected")
	}
	if strings.Contains(out, "Company") {
		t.Error("table should NOT contain 'Company' when not selected")
	}
}

func TestAccountsList_FieldFlag_JSONOutput(t *testing.T) {
	acct := sampleAccount()
	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			return &mockAccountLister{accounts: []recurly.Account{*acct}}, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "list", "--output", "json", "--field", "Code,State")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &envelope); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(envelope.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(envelope.Data))
	}
	item := envelope.Data[0]
	if item["code"] != "acct-123" {
		t.Errorf("expected code=acct-123, got %v", item["code"])
	}
	if item["state"] != "active" {
		t.Errorf("expected state=active, got %v", item["state"])
	}
	if _, ok := item["email"]; ok {
		t.Error("JSON should NOT include 'email' when not selected")
	}
	if _, ok := item["first_name"]; ok {
		t.Error("JSON should NOT include 'first_name' when not selected")
	}
}

func TestAccountsList_FieldFlag_ShortFlag(t *testing.T) {
	acct := sampleAccount()
	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			return &mockAccountLister{accounts: []recurly.Account{*acct}}, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "list", "-f", "Code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "acct-123") {
		t.Error("expected -f short flag to work")
	}
}

func TestAccountsList_FieldFlag_InvalidField(t *testing.T) {
	acct := sampleAccount()
	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			return &mockAccountLister{accounts: []recurly.Account{*acct}}, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, stderr, err := executeCommand(app, "accounts", "list", "--field", "Code,bogus_field")
	if err == nil {
		t.Fatal("expected error for invalid field")
	}
	if !strings.Contains(stderr, "bogus_field") {
		t.Errorf("error should mention invalid field, got: %s", stderr)
	}
	if !strings.Contains(stderr, "available fields") {
		t.Errorf("error should list available fields, got: %s", stderr)
	}
}

func TestAccountsGet_FieldFlag_JSONOutput(t *testing.T) {
	acct := sampleAccount()
	mock := &mockAccountAPI{
		getAccountFn: func(accountId string, opts ...recurly.Option) (*recurly.Account, error) {
			return acct, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "get", "acct-123", "--output", "json", "--field", "Code,Email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if decoded["code"] != "acct-123" {
		t.Errorf("expected code=acct-123, got %v", decoded["code"])
	}
	if decoded["email"] != "test@example.com" {
		t.Errorf("expected email=test@example.com, got %v", decoded["email"])
	}
	if _, ok := decoded["state"]; ok {
		t.Error("JSON should NOT include 'state' when not selected")
	}
}

func TestAccountsList_FieldFlag_CaseInsensitive(t *testing.T) {
	acct := sampleAccount()
	mock := &mockAccountAPI{
		listAccountsFn: func(params *recurly.ListAccountsParams, opts ...recurly.Option) (recurly.AccountLister, error) {
			return &mockAccountLister{accounts: []recurly.Account{*acct}}, nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	out, _, err := executeCommand(app, "accounts", "list", "--field", "code,email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "acct-123") {
		t.Error("lowercase field names should match")
	}
	if !strings.Contains(out, "test@example.com") {
		t.Error("lowercase field names should match")
	}
}

func TestAccountsList_FieldFlag_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand(nil, "accounts", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--field") {
		t.Error("expected help to show --field flag")
	}
	if !strings.Contains(out, "-f") {
		t.Error("expected help to show -f short flag")
	}
}
