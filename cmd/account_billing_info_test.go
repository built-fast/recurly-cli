package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

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

func setMockAccountBillingInfoAPI(mock *mockAccountBillingInfoAPI) func() {
	orig := newAccountBillingInfoAPI
	newAccountBillingInfoAPI = func() (AccountBillingInfoAPI, error) {
		return mock, nil
	}
	return func() { newAccountBillingInfoAPI = orig }
}

func sampleBillingInfo() *recurly.BillingInfo {
	created := time.Date(2025, 2, 10, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2025, 3, 15, 14, 0, 0, 0, time.UTC)
	return &recurly.BillingInfo{
		Id:        "bill1234",
		AccountId: "code-acct1",
		FirstName: "John",
		LastName:  "Doe",
		Company:   "Acme Inc",
		Valid:     true,
		PaymentMethod: recurly.PaymentMethod{
			CardType: "Visa",
		},
		PrimaryPaymentMethod: true,
		CreatedAt:            &created,
		UpdatedAt:            &updated,
	}
}

// --- billing-info get ---

func TestAccountBillingInfoGet_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("accounts", "billing-info", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "get") {
		t.Error("expected billing-info help to show 'get' subcommand")
	}
}

func TestAccountBillingInfoGet_RequiresAccountID(t *testing.T) {
	_, _, err := executeCommand("accounts", "billing-info", "get")
	if err == nil {
		t.Fatal("expected error when no account_id provided")
	}
}

func TestAccountBillingInfoGet_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("RECURLY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand("accounts", "billing-info", "get", "acct1")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountBillingInfoGet_PositionalArg(t *testing.T) {
	var capturedAccountID string
	bi := sampleBillingInfo()

	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			capturedAccountID = accountId
			return bi, nil
		},
	}
	cleanup := setMockAccountBillingInfoAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "billing-info", "get", "code-acct1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedAccountID != "code-acct1" {
		t.Errorf("expected accountId=code-acct1, got %q", capturedAccountID)
	}
}

func TestAccountBillingInfoGet_TableOutput(t *testing.T) {
	bi := sampleBillingInfo()
	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return bi, nil
		},
	}
	cleanup := setMockAccountBillingInfoAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("accounts", "billing-info", "get", "code-acct1")
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
	bi := sampleBillingInfo()
	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return bi, nil
		},
	}
	cleanup := setMockAccountBillingInfoAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "billing-info", "get", "code-acct1")
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
	bi := sampleBillingInfo()
	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return bi, nil
		},
	}
	cleanup := setMockAccountBillingInfoAPI(mock)
	defer cleanup()

	viper.Set("output", "json-pretty")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "billing-info", "get", "code-acct1")
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
	bi := sampleBillingInfo()
	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return bi, nil
		},
	}
	cleanup := setMockAccountBillingInfoAPI(mock)
	defer cleanup()

	viper.Set("output", "json")
	viper.Set("jq", ".first_name")
	defer viper.Reset()

	out, _, err := executeCommand("accounts", "billing-info", "get", "code-acct1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "John") {
		t.Errorf("expected jq output to contain 'John', got: %s", out)
	}
}

func TestAccountBillingInfoGet_SDKError(t *testing.T) {
	mock := &mockAccountBillingInfoAPI{
		getBillingInfoFn: func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
			return nil, &recurly.Error{Message: "not found"}
		},
	}
	cleanup := setMockAccountBillingInfoAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("accounts", "billing-info", "get", "code-acct1")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}
