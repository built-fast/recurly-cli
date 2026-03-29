package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
)

// --- JSON input ---

func TestFromFile_JSONInput(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "account.json")
	if err := os.WriteFile(jsonFile, []byte(`{"code": "acct-json", "email": "json@test.com"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var captured *recurly.AccountCreate
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			captured = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "create", "--from-file", jsonFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected create to be called")
	}
	if *captured.Code != "acct-json" {
		t.Errorf("expected code acct-json, got %s", *captured.Code)
	}
	if *captured.Email != "json@test.com" {
		t.Errorf("expected email json@test.com, got %s", *captured.Email)
	}
}

// --- YAML input ---

func TestFromFile_YAMLInput(t *testing.T) {
	dir := t.TempDir()
	yamlFile := filepath.Join(dir, "account.yaml")
	if err := os.WriteFile(yamlFile, []byte("code: acct-yaml\nemail: yaml@test.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var captured *recurly.AccountCreate
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			captured = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "create", "--from-file", yamlFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected create to be called")
	}
	if *captured.Code != "acct-yaml" {
		t.Errorf("expected code acct-yaml, got %s", *captured.Code)
	}
	if *captured.Email != "yaml@test.com" {
		t.Errorf("expected email yaml@test.com, got %s", *captured.Email)
	}
}

// --- YML extension ---

func TestFromFile_YMLExtension(t *testing.T) {
	dir := t.TempDir()
	ymlFile := filepath.Join(dir, "account.yml")
	if err := os.WriteFile(ymlFile, []byte("code: acct-yml\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var captured *recurly.AccountCreate
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			captured = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "create", "--from-file", ymlFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil || captured.Code == nil {
		t.Fatal("expected code to be set")
	}
	if *captured.Code != "acct-yml" {
		t.Errorf("expected code acct-yml, got %s", *captured.Code)
	}
}

// --- Stdin input (JSON auto-detected) ---

func TestFromFile_StdinJSON(t *testing.T) {
	var captured *recurly.AccountCreate
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			captured = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	stdin := bytes.NewBufferString(`{"code": "acct-stdin", "email": "stdin@test.com"}`)
	_, _, err := executeCommandWithStdin(app, stdin, "accounts", "create", "--from-file", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected create to be called")
	}
	if *captured.Code != "acct-stdin" {
		t.Errorf("expected code acct-stdin, got %s", *captured.Code)
	}
}

// --- Stdin input (YAML auto-detected) ---

func TestFromFile_StdinYAML(t *testing.T) {
	var captured *recurly.AccountCreate
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			captured = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	stdin := bytes.NewBufferString("code: acct-stdin-yaml\nemail: yaml-stdin@test.com\n")
	_, _, err := executeCommandWithStdin(app, stdin, "accounts", "create", "--from-file", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected create to be called")
	}
	if *captured.Code != "acct-stdin-yaml" {
		t.Errorf("expected code acct-stdin-yaml, got %s", *captured.Code)
	}
}

// --- Flag override: CLI flags take precedence ---

func TestFromFile_FlagOverride(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "account.json")
	if err := os.WriteFile(jsonFile, []byte(`{"code": "file-code", "email": "file@test.com"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var captured *recurly.AccountCreate
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			captured = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	// CLI --code should override file code; email should come from file
	_, _, err := executeCommand(app, "accounts", "create", "--from-file", jsonFile, "--code", "cli-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected create to be called")
	}
	if *captured.Code != "cli-code" {
		t.Errorf("expected code cli-code (CLI override), got %s", *captured.Code)
	}
	if *captured.Email != "file@test.com" {
		t.Errorf("expected email from file, got %s", *captured.Email)
	}
}

// --- Underscore to hyphen key mapping ---

func TestFromFile_UnderscoreKeys(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "account.json")
	if err := os.WriteFile(jsonFile, []byte(`{"first_name": "Jane", "last_name": "Doe"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var captured *recurly.AccountCreate
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			captured = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "create", "--from-file", jsonFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected create to be called")
	}
	if *captured.FirstName != "Jane" {
		t.Errorf("expected first name Jane, got %s", *captured.FirstName)
	}
	if *captured.LastName != "Doe" {
		t.Errorf("expected last name Doe, got %s", *captured.LastName)
	}
}

// --- Nested objects (e.g., address on billing-info) ---

func TestFromFile_NestedObjects(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "billing.json")
	if err := os.WriteFile(jsonFile, []byte(`{
		"first_name": "Jane",
		"address": {
			"street1": "123 Main St",
			"city": "Springfield",
			"country": "US"
		}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var captured *recurly.BillingInfoCreate
	app := &App{
		NewAccountBillingInfoAPI: func(_ *cobra.Command) (AccountBillingInfoAPI, error) {
			return &mockBillingInfoAPI{
				updateBillingInfoFn: func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
					captured = body
					return &recurly.BillingInfo{Id: "bi-1", AccountId: accountId, FirstName: "Jane"}, nil
				},
			}, nil
		},
	}

	_, _, err := executeCommand(app, "accounts", "billing-info", "update", "acct-1", "--from-file", jsonFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected update to be called")
	}
	if *captured.FirstName != "Jane" {
		t.Errorf("expected first name Jane, got %s", *captured.FirstName)
	}
	if captured.Address == nil {
		t.Fatal("expected address to be set")
	}
	if *captured.Address.Street1 != "123 Main St" {
		t.Errorf("expected street1 '123 Main St', got %s", *captured.Address.Street1)
	}
	if *captured.Address.City != "Springfield" {
		t.Errorf("expected city Springfield, got %s", *captured.Address.City)
	}
	if *captured.Address.Country != "US" {
		t.Errorf("expected country US, got %s", *captured.Address.Country)
	}
}

// --- Error: invalid file path ---

func TestFromFile_InvalidPath(t *testing.T) {
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return &mockAccountAPI{}, nil }}

	_, stderr, err := executeCommand(app, "accounts", "create", "--from-file", "/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(stderr, "error reading file") {
		t.Errorf("expected error about reading file, got %q", stderr)
	}
}

// --- Error: unparseable content ---

func TestFromFile_UnparseableJSON(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(jsonFile, []byte(`{not valid json`), 0o644); err != nil {
		t.Fatal(err)
	}

	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return &mockAccountAPI{}, nil }}

	_, stderr, err := executeCommand(app, "accounts", "create", "--from-file", jsonFile)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(stderr, "invalid JSON") {
		t.Errorf("expected JSON parse error, got %q", stderr)
	}
}

// --- Error: unknown keys ---

func TestFromFile_UnknownKey(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "bad-keys.json")
	if err := os.WriteFile(jsonFile, []byte(`{"code": "x", "nonexistent_field": "y"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return &mockAccountAPI{}, nil }}

	_, stderr, err := executeCommand(app, "accounts", "create", "--from-file", jsonFile)
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
	if !strings.Contains(stderr, "unknown key") {
		t.Errorf("expected unknown key error, got %q", stderr)
	}
}

// --- Error: unsupported file extension ---

func TestFromFile_UnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	txtFile := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(txtFile, []byte(`code: x`), 0o644); err != nil {
		t.Fatal(err)
	}

	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return &mockAccountAPI{}, nil }}

	_, stderr, err := executeCommand(app, "accounts", "create", "--from-file", txtFile)
	if err == nil {
		t.Fatal("expected error for unsupported extension")
	}
	if !strings.Contains(stderr, "unsupported file extension") {
		t.Errorf("expected unsupported extension error, got %q", stderr)
	}
}

// --- Error: empty stdin ---

func TestFromFile_EmptyStdin(t *testing.T) {
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return &mockAccountAPI{}, nil }}

	stdin := bytes.NewBufferString("")
	_, stderr, err := executeCommandWithStdin(app, stdin, "accounts", "create", "--from-file", "-")
	if err == nil {
		t.Fatal("expected error for empty stdin")
	}
	if !strings.Contains(stderr, "stdin is empty") {
		t.Errorf("expected empty stdin error, got %q", stderr)
	}
}

// --- Short flag -F ---

func TestFromFile_ShortFlag(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "account.json")
	if err := os.WriteFile(jsonFile, []byte(`{"code": "short-flag"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var captured *recurly.AccountCreate
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			captured = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "create", "-F", jsonFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil || captured.Code == nil {
		t.Fatal("expected code to be set")
	}
	if *captured.Code != "short-flag" {
		t.Errorf("expected code short-flag, got %s", *captured.Code)
	}
}

// --- Boolean values from file ---

func TestFromFile_BooleanValue(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "account.json")
	if err := os.WriteFile(jsonFile, []byte(`{"code": "bool-test", "tax_exempt": true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var captured *recurly.AccountCreate
	mock := &mockAccountAPI{
		createAccountFn: func(body *recurly.AccountCreate, opts ...recurly.Option) (*recurly.Account, error) {
			captured = body
			return sampleAccount(), nil
		},
	}
	app := &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return mock, nil }}

	_, _, err := executeCommand(app, "accounts", "create", "--from-file", jsonFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected create to be called")
	}
	if captured.TaxExempt == nil || !*captured.TaxExempt {
		t.Error("expected tax_exempt to be true")
	}
}

// --- mockBillingInfoAPI for nested object test ---

type mockBillingInfoAPI struct {
	getBillingInfoFn    func(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error)
	updateBillingInfoFn func(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error)
	removeBillingInfoFn func(accountId string, opts ...recurly.Option) (*recurly.Empty, error)
}

func (m *mockBillingInfoAPI) GetBillingInfo(accountId string, opts ...recurly.Option) (*recurly.BillingInfo, error) {
	return m.getBillingInfoFn(accountId, opts...)
}

func (m *mockBillingInfoAPI) UpdateBillingInfo(accountId string, body *recurly.BillingInfoCreate, opts ...recurly.Option) (*recurly.BillingInfo, error) {
	return m.updateBillingInfoFn(accountId, body, opts...)
}

func (m *mockBillingInfoAPI) RemoveBillingInfo(accountId string, opts ...recurly.Option) (*recurly.Empty, error) {
	return m.removeBillingInfoFn(accountId, opts...)
}
