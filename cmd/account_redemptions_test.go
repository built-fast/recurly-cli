package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	recurly "github.com/recurly/recurly-client-go/v5"
)

// --- accounts redemptions list ---

func TestAccountRedemptionsList_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "accounts", "redemptions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected redemptions help to show 'list' subcommand")
	}
}

func TestAccountRedemptionsList_Success(t *testing.T) {
	t.Parallel()
	r := sampleRedemption()
	mock := &mockAccountRedemptionAPI{
		listAccountCouponRedemptionsFn: func(accountId string, params *recurly.ListAccountCouponRedemptionsParams, opts ...recurly.Option) (recurly.CouponRedemptionLister, error) {
			if accountId != "acct-1" {
				t.Errorf("expected accountId=acct-1, got %q", accountId)
			}
			return &mockLister[recurly.CouponRedemption]{items: []recurly.CouponRedemption{r}}, nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	out, _, err := executeCommand(app, "accounts", "redemptions", "list", "acct-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"redemption-abc123", "SAVE10", "active", "USD", "10.50"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	for _, header := range []string{"ID", "Coupon Code", "State", "Currency", "Discounted", "Created At"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestAccountRedemptionsList_EmptyResults(t *testing.T) {
	t.Parallel()
	mock := &mockAccountRedemptionAPI{
		listAccountCouponRedemptionsFn: func(accountId string, params *recurly.ListAccountCouponRedemptionsParams, opts ...recurly.Option) (recurly.CouponRedemptionLister, error) {
			return &mockLister[recurly.CouponRedemption]{items: []recurly.CouponRedemption{}}, nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	out, _, err := executeCommand(app, "accounts", "redemptions", "list", "acct-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, "redemption-abc123") {
		t.Error("expected no redemption data in empty results")
	}
}

func TestAccountRedemptionsList_SDKError(t *testing.T) {
	t.Parallel()
	mock := &mockAccountRedemptionAPI{
		listAccountCouponRedemptionsFn: func(accountId string, params *recurly.ListAccountCouponRedemptionsParams, opts ...recurly.Option) (recurly.CouponRedemptionLister, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	app := newTestAccountRedemptionApp(mock)

	_, _, err := executeCommand(app, "accounts", "redemptions", "list", "acct-1")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestAccountRedemptionsList_FlagPassthrough(t *testing.T) {
	t.Parallel()
	var capturedParams *recurly.ListAccountCouponRedemptionsParams

	mock := &mockAccountRedemptionAPI{
		listAccountCouponRedemptionsFn: func(accountId string, params *recurly.ListAccountCouponRedemptionsParams, opts ...recurly.Option) (recurly.CouponRedemptionLister, error) {
			capturedParams = params
			return &mockLister[recurly.CouponRedemption]{items: []recurly.CouponRedemption{}}, nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	_, _, err := executeCommand(app, "accounts", "redemptions", "list", "acct-1", "--sort", "created_at")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedParams == nil {
		t.Fatal("expected params to be captured")
	}
	if capturedParams.Sort == nil || *capturedParams.Sort != "created_at" {
		t.Errorf("expected sort=created_at, got %v", capturedParams.Sort)
	}
}

func TestAccountRedemptionsList_MissingArg_ReturnsError(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "accounts", "redemptions", "list")
	if err == nil {
		t.Fatal("expected error when no account ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

// --- accounts redemptions list-active ---

func TestAccountRedemptionsListActive_Success(t *testing.T) {
	t.Parallel()
	r := sampleRedemption()
	mock := &mockAccountRedemptionAPI{
		listActiveCouponRedemptionsFn: func(accountId string, opts ...recurly.Option) (recurly.CouponRedemptionLister, error) {
			if accountId != "acct-1" {
				t.Errorf("expected accountId=acct-1, got %q", accountId)
			}
			return &mockLister[recurly.CouponRedemption]{items: []recurly.CouponRedemption{r}}, nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	out, _, err := executeCommand(app, "accounts", "redemptions", "list-active", "acct-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"redemption-abc123", "SAVE10", "active", "USD", "10.50"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestAccountRedemptionsListActive_EmptyResults(t *testing.T) {
	t.Parallel()
	mock := &mockAccountRedemptionAPI{
		listActiveCouponRedemptionsFn: func(accountId string, opts ...recurly.Option) (recurly.CouponRedemptionLister, error) {
			return &mockLister[recurly.CouponRedemption]{items: []recurly.CouponRedemption{}}, nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	out, _, err := executeCommand(app, "accounts", "redemptions", "list-active", "acct-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, "redemption-abc123") {
		t.Error("expected no redemption data in empty results")
	}
}

func TestAccountRedemptionsListActive_SDKError(t *testing.T) {
	t.Parallel()
	mock := &mockAccountRedemptionAPI{
		listActiveCouponRedemptionsFn: func(accountId string, opts ...recurly.Option) (recurly.CouponRedemptionLister, error) {
			return nil, fmt.Errorf("service unavailable")
		},
	}
	app := newTestAccountRedemptionApp(mock)

	_, _, err := executeCommand(app, "accounts", "redemptions", "list-active", "acct-1")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestAccountRedemptionsListActive_MissingArg_ReturnsError(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "accounts", "redemptions", "list-active")
	if err == nil {
		t.Fatal("expected error when no account ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

// --- accounts redemptions create ---

func TestAccountRedemptionsCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedAccountID string
	var capturedBody *recurly.CouponRedemptionCreate
	mock := &mockAccountRedemptionAPI{
		createCouponRedemptionFn: func(accountId string, body *recurly.CouponRedemptionCreate, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			capturedAccountID = accountId
			capturedBody = body
			return sampleRedemptionDetail(), nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	out, _, err := executeCommand(app, "accounts", "redemptions", "create", "acct-1", "--coupon-id", "SAVE10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAccountID != "acct-1" {
		t.Errorf("expected accountId=acct-1, got %q", capturedAccountID)
	}
	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if *capturedBody.CouponId != "SAVE10" {
		t.Errorf("expected coupon_id=SAVE10, got %v", *capturedBody.CouponId)
	}

	for _, expected := range []string{"redemption-abc123", "acct-1", "SAVE10", "active", "USD", "10.50"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestAccountRedemptionsCreate_AllOptionalFlags(t *testing.T) {
	t.Parallel()
	var capturedBody *recurly.CouponRedemptionCreate
	mock := &mockAccountRedemptionAPI{
		createCouponRedemptionFn: func(accountId string, body *recurly.CouponRedemptionCreate, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			capturedBody = body
			return sampleRedemptionDetail(), nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	_, _, err := executeCommand(app, "accounts", "redemptions", "create", "acct-1",
		"--coupon-id", "SAVE10",
		"--currency", "EUR",
		"--subscription-id", "sub-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if *capturedBody.CouponId != "SAVE10" {
		t.Errorf("expected coupon_id=SAVE10, got %v", *capturedBody.CouponId)
	}
	if capturedBody.Currency == nil || *capturedBody.Currency != "EUR" {
		t.Errorf("expected currency=EUR, got %v", capturedBody.Currency)
	}
	if capturedBody.SubscriptionId == nil || *capturedBody.SubscriptionId != "sub-123" {
		t.Errorf("expected subscription_id=sub-123, got %v", capturedBody.SubscriptionId)
	}
}

func TestAccountRedemptionsCreate_MissingCouponID_ReturnsError(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "accounts", "redemptions", "create", "acct-1", "--no-input")
	if err == nil {
		t.Fatal("expected error when --coupon-id is not provided")
	}
	if !strings.Contains(stderr, "coupon-id") {
		t.Errorf("expected error about missing coupon-id flag, got %q", stderr)
	}
}

func TestAccountRedemptionsCreate_MissingArg_ReturnsError(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "accounts", "redemptions", "create")
	if err == nil {
		t.Fatal("expected error when no account ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestAccountRedemptionsCreate_SDKError(t *testing.T) {
	t.Parallel()
	mock := &mockAccountRedemptionAPI{
		createCouponRedemptionFn: func(accountId string, body *recurly.CouponRedemptionCreate, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			return nil, fmt.Errorf("invalid coupon")
		},
	}
	app := newTestAccountRedemptionApp(mock)

	_, _, err := executeCommand(app, "accounts", "redemptions", "create", "acct-1", "--coupon-id", "BAD")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestAccountRedemptionsCreate_UnsetOptionalFlagsNil(t *testing.T) {
	t.Parallel()
	var capturedBody *recurly.CouponRedemptionCreate
	mock := &mockAccountRedemptionAPI{
		createCouponRedemptionFn: func(accountId string, body *recurly.CouponRedemptionCreate, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			capturedBody = body
			return sampleRedemptionDetail(), nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	_, _, err := executeCommand(app, "accounts", "redemptions", "create", "acct-1", "--coupon-id", "SAVE10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Currency != nil {
		t.Error("expected currency to be nil when not set")
	}
	if capturedBody.SubscriptionId != nil {
		t.Error("expected subscription_id to be nil when not set")
	}
}

// --- accounts redemptions remove ---

func TestAccountRedemptionsRemove_AccountOnly(t *testing.T) {
	t.Parallel()
	var capturedAccountID string
	mock := &mockAccountRedemptionAPI{
		removeCouponRedemptionFn: func(accountId string, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			capturedAccountID = accountId
			return sampleRedemptionDetail(), nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	out, _, err := executeCommand(app, "accounts", "redemptions", "remove", "acct-1", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAccountID != "acct-1" {
		t.Errorf("expected accountId=acct-1, got %q", capturedAccountID)
	}

	for _, expected := range []string{"redemption-abc123", "acct-1", "SAVE10", "active", "USD", "10.50"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestAccountRedemptionsRemove_WithRedemptionID(t *testing.T) {
	t.Parallel()
	var capturedAccountID, capturedRedemptionID string
	mock := &mockAccountRedemptionAPI{
		removeCouponRedemptionByIdFn: func(accountId string, couponRedemptionId string, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			capturedAccountID = accountId
			capturedRedemptionID = couponRedemptionId
			return sampleRedemptionDetail(), nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	out, _, err := executeCommand(app, "accounts", "redemptions", "remove", "acct-1", "redemption-abc123", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAccountID != "acct-1" {
		t.Errorf("expected accountId=acct-1, got %q", capturedAccountID)
	}
	if capturedRedemptionID != "redemption-abc123" {
		t.Errorf("expected redemptionId=redemption-abc123, got %q", capturedRedemptionID)
	}

	if !strings.Contains(out, "redemption-abc123") {
		t.Errorf("expected output to contain redemption ID, got:\n%s", out)
	}
}

func TestAccountRedemptionsRemove_MissingArg_ReturnsError(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "accounts", "redemptions", "remove")
	if err == nil {
		t.Fatal("expected error when no account ID is provided")
	}
	if !strings.Contains(stderr, "accepts between 1 and 2 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

func TestAccountRedemptionsRemove_SDKError(t *testing.T) {
	t.Parallel()
	mock := &mockAccountRedemptionAPI{
		removeCouponRedemptionFn: func(accountId string, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	app := newTestAccountRedemptionApp(mock)

	_, _, err := executeCommand(app, "accounts", "redemptions", "remove", "acct-1", "--yes")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestAccountRedemptionsRemove_SDKErrorById(t *testing.T) {
	t.Parallel()
	mock := &mockAccountRedemptionAPI{
		removeCouponRedemptionByIdFn: func(accountId string, couponRedemptionId string, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	app := newTestAccountRedemptionApp(mock)

	_, _, err := executeCommand(app, "accounts", "redemptions", "remove", "acct-1", "bad-id", "--yes")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- accounts redemptions remove confirmation tests ---

func TestAccountRedemptionsRemove_ConfirmationNo(t *testing.T) {
	t.Parallel()
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "accounts", "redemptions", "remove", "acct-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "[y/N]") {
		t.Error("expected confirmation prompt in stderr")
	}
	if !strings.Contains(stderr, "Removal canceled.") {
		t.Error("expected cancellation message in stderr")
	}
}

func TestAccountRedemptionsRemove_ConfirmationEmpty(t *testing.T) {
	t.Parallel()
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(nil, stdin, "accounts", "redemptions", "remove", "acct-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Removal canceled.") {
		t.Error("expected cancellation message for empty input")
	}
}

func TestAccountRedemptionsRemove_ConfirmationYes(t *testing.T) {
	t.Parallel()
	mock := &mockAccountRedemptionAPI{
		removeCouponRedemptionFn: func(accountId string, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			return sampleRedemptionDetail(), nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(app, stdin, "accounts", "redemptions", "remove", "acct-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "[y/N]") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "redemption-abc123") {
		t.Errorf("expected redemption data in output, got:\n%s", out)
	}
}

func TestAccountRedemptionsRemove_YesFlagSkipsConfirmation(t *testing.T) {
	t.Parallel()
	mock := &mockAccountRedemptionAPI{
		removeCouponRedemptionFn: func(accountId string, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			return sampleRedemptionDetail(), nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	out, stderr, err := executeCommand(app, "accounts", "redemptions", "remove", "acct-1", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stderr, "[y/N]") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "redemption-abc123") {
		t.Errorf("expected redemption data in output, got:\n%s", out)
	}
}

// --- JSON output tests ---

func TestAccountRedemptionsList_JSONOutput(t *testing.T) {
	t.Parallel()
	r := sampleRedemption()
	mock := &mockAccountRedemptionAPI{
		listAccountCouponRedemptionsFn: func(accountId string, params *recurly.ListAccountCouponRedemptionsParams, opts ...recurly.Option) (recurly.CouponRedemptionLister, error) {
			return &mockLister[recurly.CouponRedemption]{items: []recurly.CouponRedemption{r}}, nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	out, _, err := executeCommand(app, "--output", "json", "accounts", "redemptions", "list", "acct-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "redemption-abc123") {
		t.Errorf("expected JSON output with redemption data, got:\n%s", out)
	}
	if !strings.Contains(out, "\"id\"") {
		t.Errorf("expected JSON field 'id' in output, got:\n%s", out)
	}
}

func TestAccountRedemptionsCreate_JSONOutput(t *testing.T) {
	t.Parallel()
	mock := &mockAccountRedemptionAPI{
		createCouponRedemptionFn: func(accountId string, body *recurly.CouponRedemptionCreate, opts ...recurly.Option) (*recurly.CouponRedemption, error) {
			return sampleRedemptionDetail(), nil
		},
	}
	app := newTestAccountRedemptionApp(mock)

	out, _, err := executeCommand(app, "--output", "json", "accounts", "redemptions", "create", "acct-1", "--coupon-id", "SAVE10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "redemption-abc123") {
		t.Errorf("expected JSON output with redemption data, got:\n%s", out)
	}
	if !strings.Contains(out, "\"id\"") {
		t.Errorf("expected JSON field 'id' in output, got:\n%s", out)
	}
}
