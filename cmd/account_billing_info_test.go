package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/viper"
)

// mockAccountBillingInfoAPI implements AccountBillingInfoAPI for testing.
type mockAccountBillingInfoAPI struct {
	getBillingInfoFn    func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error)
	updateBillingInfoFn func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error)
	removeBillingInfoFn func(accountId string, opts ...recurly.Option) (*recurly.Empty, error)
}

func (m *mockAccountBillingInfoAPI) GetBillingInfo(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
	if m.getBillingInfoFn != nil {
		return m.getBillingInfoFn(accountId, opts...)
	}
	return nil, nil
}

func (m *mockAccountBillingInfoAPI) UpdateBillingInfo(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
	if m.updateBillingInfoFn != nil {
		return m.updateBillingInfoFn(accountId, body, opts...)
	}
	return nil, nil
}

func (m *mockAccountBillingInfoAPI) RemoveBillingInfo(accountId string, opts ...recurly.Option) (*recurly.Empty, error) {
	if m.removeBillingInfoFn != nil {
		return m.removeBillingInfoFn(accountId, opts...)
	}
	return nil, nil
}

// --- billing-info get ---

func TestAccountBillingInfoGet_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "billing-info", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "get") {
		t.Error("expected billing-info help to show 'get' subcommand")
	}
}

func TestAccountBillingInfoGet_RequiresAccountID(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "accounts", "billing-info", "get")
	if err == nil {
		t.Fatal("expected error when no account_id provided")
	}
}

// Cannot use t.Parallel() — uses t.Setenv which is incompatible
func TestAccountBillingInfoGet_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "billing-info", "get", "acct1")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountBillingInfoGet_PositionalArg(t *testing.T) {
	t.Parallel()
	var capturedAccountID string
	bi := sampleBillingInfo()

	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			capturedAccountID = accountId
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	_, _, err := executeCommand(app, "accounts", "billing-info", "get", "code-acct1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedAccountID != "code-acct1" {
		t.Errorf("expected accountId=code-acct1, got %q", capturedAccountID)
	}
}

func TestAccountBillingInfoGet_TableOutput(t *testing.T) {
	t.Parallel()
	bi := sampleBillingInfo()
	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	out, _, err := executeCommand(app, "accounts", "billing-info", "get", "code-acct1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"ID", "bill1234",
		"Account ID", "code-acct1",
		"First Name", "John",
		"Last Name", "Doe",
		"Company", "Acme Inc",
		"Valid", "true",
		"Payment Method", "Visa",
		"Primary Payment Method", "true",
		"Created At", "2025-02-10T12:00:00Z",
		"Updated At", "2025-03-15T14:00:00Z",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestAccountBillingInfoGet_JSONOutput(t *testing.T) {
	t.Parallel()
	bi := sampleBillingInfo()
	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	out, _, err := executeCommand(app, "accounts", "billing-info", "get", "code-acct1", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\noutput: %s", err, out)
	}
	if result["id"] != "bill1234" {
		t.Errorf("expected id=bill1234 in JSON, got %v", result["id"])
	}
	if result["account_id"] != "code-acct1" {
		t.Errorf("expected account_id=code-acct1 in JSON, got %v", result["account_id"])
	}
}

func TestAccountBillingInfoGet_JSONPrettyOutput(t *testing.T) {
	t.Parallel()
	bi := sampleBillingInfo()
	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	out, _, err := executeCommand(app, "accounts", "billing-info", "get", "code-acct1", "--output", "json-pretty")
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

func TestAccountBillingInfoGet_JQFilter(t *testing.T) {
	t.Parallel()
	bi := sampleBillingInfo()
	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	out, _, err := executeCommand(app, "accounts", "billing-info", "get", "code-acct1", "--output", "json", "--jq", ".first_name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "John") {
		t.Errorf("expected jq output to contain 'John', got: %s", out)
	}
}

func TestAccountBillingInfoGet_SDKError(t *testing.T) {
	t.Parallel()
	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	_, _, err := executeCommand(app, "accounts", "billing-info", "get", "code-acct1")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- billing-info update ---

func TestAccountBillingInfoUpdate_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "billing-info", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "update") {
		t.Error("expected billing-info help to show 'update' subcommand")
	}
}

func TestAccountBillingInfoUpdate_RequiresAccountID(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "accounts", "billing-info", "update")
	if err == nil {
		t.Fatal("expected error when no account_id provided")
	}
}

// Cannot use t.Parallel() — uses t.Setenv which is incompatible
func TestAccountBillingInfoUpdate_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "accounts", "billing-info", "update", "acct1", "--first-name", "Jane")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountBillingInfoUpdate_PositionalArg(t *testing.T) {
	t.Parallel()
	var capturedAccountID string
	bi := sampleBillingInfo()

	mock := &mockAccountBillingInfoAPI{
		updateBillingInfoFn: func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			capturedAccountID = accountId
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	_, _, err := executeCommand(app, "accounts", "billing-info", "update", "code-acct1", "--first-name", "Jane")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedAccountID != "code-acct1" {
		t.Errorf("expected accountId=code-acct1, got %q", capturedAccountID)
	}
}

func TestAccountBillingInfoUpdate_OnlyChangedFields(t *testing.T) {
	t.Parallel()
	var capturedBody *recurly.BillingInfoCreate
	bi := sampleBillingInfo()

	mock := &mockAccountBillingInfoAPI{
		updateBillingInfoFn: func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			capturedBody = body
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	_, _, err := executeCommand(app, "accounts", "billing-info", "update", "code-acct1", "--first-name", "Jane", "--company", "NewCo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.FirstName == nil || *capturedBody.FirstName != "Jane" {
		t.Errorf("expected FirstName=Jane, got %v", capturedBody.FirstName)
	}
	if capturedBody.Company == nil || *capturedBody.Company != "NewCo" {
		t.Errorf("expected Company=NewCo, got %v", capturedBody.Company)
	}
	// Fields not set should be nil
	if capturedBody.LastName != nil {
		t.Errorf("expected LastName=nil (not changed), got %v", *capturedBody.LastName)
	}
	if capturedBody.VatNumber != nil {
		t.Errorf("expected VatNumber=nil (not changed), got %v", *capturedBody.VatNumber)
	}
	if capturedBody.TokenId != nil {
		t.Errorf("expected TokenId=nil (not changed), got %v", *capturedBody.TokenId)
	}
}

func TestAccountBillingInfoUpdate_BoolFlags(t *testing.T) {
	t.Parallel()
	var capturedBody *recurly.BillingInfoCreate
	bi := sampleBillingInfo()

	mock := &mockAccountBillingInfoAPI{
		updateBillingInfoFn: func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			capturedBody = body
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	_, _, err := executeCommand(app, "accounts", "billing-info", "update", "code-acct1", "--primary-payment-method", "--backup-payment-method")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.PrimaryPaymentMethod == nil || !*capturedBody.PrimaryPaymentMethod {
		t.Error("expected PrimaryPaymentMethod=true")
	}
	if capturedBody.BackupPaymentMethod == nil || !*capturedBody.BackupPaymentMethod {
		t.Error("expected BackupPaymentMethod=true")
	}
}

func TestAccountBillingInfoUpdate_AddressFlags(t *testing.T) {
	t.Parallel()
	var capturedBody *recurly.BillingInfoCreate
	bi := sampleBillingInfo()

	mock := &mockAccountBillingInfoAPI{
		updateBillingInfoFn: func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			capturedBody = body
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	_, _, err := executeCommand(app, "accounts", "billing-info", "update", "code-acct1",
		"--address-street1", "123 Main St",
		"--address-street2", "Apt 4",
		"--address-city", "San Francisco",
		"--address-region", "CA",
		"--address-postal-code", "94105",
		"--address-country", "US",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Address == nil {
		t.Fatal("expected Address to be set")
	}
	if *capturedBody.Address.Street1 != "123 Main St" {
		t.Errorf("expected Street1=123 Main St, got %v", *capturedBody.Address.Street1)
	}
	if *capturedBody.Address.Street2 != "Apt 4" {
		t.Errorf("expected Street2=Apt 4, got %v", *capturedBody.Address.Street2)
	}
	if *capturedBody.Address.City != "San Francisco" {
		t.Errorf("expected City=San Francisco, got %v", *capturedBody.Address.City)
	}
	if *capturedBody.Address.Region != "CA" {
		t.Errorf("expected Region=CA, got %v", *capturedBody.Address.Region)
	}
	if *capturedBody.Address.PostalCode != "94105" {
		t.Errorf("expected PostalCode=94105, got %v", *capturedBody.Address.PostalCode)
	}
	if *capturedBody.Address.Country != "US" {
		t.Errorf("expected Country=US, got %v", *capturedBody.Address.Country)
	}
}

func TestAccountBillingInfoUpdate_NoAddressWhenNotSet(t *testing.T) {
	t.Parallel()
	var capturedBody *recurly.BillingInfoCreate
	bi := sampleBillingInfo()

	mock := &mockAccountBillingInfoAPI{
		updateBillingInfoFn: func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			capturedBody = body
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	_, _, err := executeCommand(app, "accounts", "billing-info", "update", "code-acct1", "--first-name", "Jane")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Address != nil {
		t.Error("expected Address=nil when no address flags set")
	}
}

func TestAccountBillingInfoUpdate_TableOutput(t *testing.T) {
	t.Parallel()
	bi := sampleBillingInfo()
	mock := &mockAccountBillingInfoAPI{
		updateBillingInfoFn: func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	out, _, err := executeCommand(app, "accounts", "billing-info", "update", "code-acct1", "--first-name", "Jane")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"ID", "bill1234",
		"Account ID", "code-acct1",
		"First Name", "John",
		"Last Name", "Doe",
		"Company", "Acme Inc",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestAccountBillingInfoUpdate_JSONOutput(t *testing.T) {
	t.Parallel()
	bi := sampleBillingInfo()
	mock := &mockAccountBillingInfoAPI{
		updateBillingInfoFn: func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return bi, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	out, _, err := executeCommand(app, "accounts", "billing-info", "update", "code-acct1", "--first-name", "Jane", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\noutput: %s", err, out)
	}
	if result["id"] != "bill1234" {
		t.Errorf("expected id=bill1234 in JSON, got %v", result["id"])
	}
}

func TestAccountBillingInfoUpdate_SDKError(t *testing.T) {
	t.Parallel()
	mock := &mockAccountBillingInfoAPI{
		updateBillingInfoFn: func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return nil, &recurly.Error{Message: "validation error"}
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	_, _, err := executeCommand(app, "accounts", "billing-info", "update", "code-acct1", "--first-name", "Jane")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- billing-info remove ---

func TestAccountBillingInfoRemove_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "billing-info", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "remove") {
		t.Error("expected billing-info help to show 'remove' subcommand")
	}
}

func TestAccountBillingInfoRemove_RequiresAccountID(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "accounts", "billing-info", "remove")
	if err == nil {
		t.Fatal("expected error when no account_id provided")
	}
}

// Cannot use t.Parallel() — uses t.Setenv which is incompatible
func TestAccountBillingInfoRemove_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommandWithStdin(nil,
		bytes.NewBufferString("y\n"),
		"accounts", "billing-info", "remove", "acct1",
	)
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountBillingInfoRemove_ConfirmNo_Cancels(t *testing.T) {
	t.Parallel()
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "accounts", "billing-info", "remove", "acct-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Remove billing info from account acct-123? [y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Removal cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestAccountBillingInfoRemove_ConfirmDefault_Cancels(t *testing.T) {
	t.Parallel()
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "accounts", "billing-info", "remove", "acct-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Removal cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestAccountBillingInfoRemove_ConfirmYes_Succeeds(t *testing.T) {
	t.Parallel()
	var capturedAccountID string
	mock := &mockAccountBillingInfoAPI{
		removeBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.Empty, error) {
			capturedAccountID = accountId
			return &recurly.Empty{}, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "accounts", "billing-info", "remove", "acct-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedAccountID != "acct-789" {
		t.Errorf("expected accountId=acct-789, got %q", capturedAccountID)
	}
	if !strings.Contains(stderr, "Remove billing info from account acct-789?") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "Billing info removed from account acct-789") {
		t.Errorf("expected success message, got:\n%s", out)
	}
}

func TestAccountBillingInfoRemove_YesFlag_SkipsPrompt(t *testing.T) {
	t.Parallel()
	var capturedAccountID string
	mock := &mockAccountBillingInfoAPI{
		removeBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.Empty, error) {
			capturedAccountID = accountId
			return &recurly.Empty{}, nil
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	out, stderr, err := executeCommand(app, "accounts", "billing-info", "remove", "acct-456", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedAccountID != "acct-456" {
		t.Errorf("expected accountId=acct-456, got %q", capturedAccountID)
	}
	if strings.Contains(stderr, "Remove billing info") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "Billing info removed from account acct-456") {
		t.Errorf("expected success message, got:\n%s", out)
	}
}

func TestAccountBillingInfoRemove_SDKError(t *testing.T) {
	t.Parallel()
	mock := &mockAccountBillingInfoAPI{
		removeBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.Empty, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	app := newTestAccountBillingInfoApp(mock)

	_, _, err := executeCommand(app, "accounts", "billing-info", "remove", "acct1", "--yes")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}
