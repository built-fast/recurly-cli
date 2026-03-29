package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mockCouponAPI implements CouponAPI for testing.
type mockCouponAPI struct {
	listCouponsFn               func(params *recurly.ListCouponsParams, opts ...recurly.Option) (recurly.CouponLister, error)
	getCouponFn                 func(couponId string, opts ...recurly.Option) (*recurly.Coupon, error)
	createCouponFn              func(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error)
	updateCouponFn              func(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error)
	deactivateCouponFn          func(couponId string, opts ...recurly.Option) (*recurly.Coupon, error)
	restoreCouponFn             func(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error)
	generateUniqueCouponCodesFn func(couponId string, body *recurly.CouponBulkCreate, opts ...recurly.Option) (*recurly.UniqueCouponCodeParams, error)
	listUniqueCouponCodesFn     func(couponId string, params *recurly.ListUniqueCouponCodesParams, opts ...recurly.Option) (recurly.UniqueCouponCodeLister, error)
}

func (m *mockCouponAPI) ListCoupons(params *recurly.ListCouponsParams, opts ...recurly.Option) (recurly.CouponLister, error) {
	return m.listCouponsFn(params, opts...)
}

func (m *mockCouponAPI) GetCoupon(couponId string, opts ...recurly.Option) (*recurly.Coupon, error) {
	return m.getCouponFn(couponId, opts...)
}

func (m *mockCouponAPI) CreateCoupon(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error) {
	return m.createCouponFn(body, opts...)
}

func (m *mockCouponAPI) UpdateCoupon(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error) {
	return m.updateCouponFn(couponId, body, opts...)
}

func (m *mockCouponAPI) DeactivateCoupon(couponId string, opts ...recurly.Option) (*recurly.Coupon, error) {
	return m.deactivateCouponFn(couponId, opts...)
}

func (m *mockCouponAPI) RestoreCoupon(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error) {
	return m.restoreCouponFn(couponId, body, opts...)
}

func (m *mockCouponAPI) GenerateUniqueCouponCodes(couponId string, body *recurly.CouponBulkCreate, opts ...recurly.Option) (*recurly.UniqueCouponCodeParams, error) {
	return m.generateUniqueCouponCodesFn(couponId, body, opts...)
}

func (m *mockCouponAPI) ListUniqueCouponCodes(couponId string, params *recurly.ListUniqueCouponCodesParams, opts ...recurly.Option) (recurly.UniqueCouponCodeLister, error) {
	return m.listUniqueCouponCodesFn(couponId, params, opts...)
}

// mockCouponLister implements recurly.CouponLister for testing.
type mockCouponLister struct {
	coupons []recurly.Coupon
	fetched bool
}

func (m *mockCouponLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockCouponLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockCouponLister) Count() (*int64, error) {
	n := int64(len(m.coupons))
	return &n, nil
}

func (m *mockCouponLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockCouponLister) Data() []recurly.Coupon {
	return m.coupons
}

func (m *mockCouponLister) HasMore() bool {
	return !m.fetched
}

func (m *mockCouponLister) Next() string {
	return ""
}

// mockUniqueCouponCodeLister implements recurly.UniqueCouponCodeLister for testing.
type mockUniqueCouponCodeLister struct {
	codes   []recurly.UniqueCouponCode
	fetched bool
}

func (m *mockUniqueCouponCodeLister) Fetch() error {
	m.fetched = true
	return nil
}

func (m *mockUniqueCouponCodeLister) FetchWithContext(_ context.Context) error {
	return m.Fetch()
}

func (m *mockUniqueCouponCodeLister) Count() (*int64, error) {
	n := int64(len(m.codes))
	return &n, nil
}

func (m *mockUniqueCouponCodeLister) CountWithContext(_ context.Context) (*int64, error) {
	return m.Count()
}

func (m *mockUniqueCouponCodeLister) Data() []recurly.UniqueCouponCode {
	return m.codes
}

func (m *mockUniqueCouponCodeLister) HasMore() bool {
	return !m.fetched
}

func (m *mockUniqueCouponCodeLister) Next() string {
	return ""
}

// setMockCouponAPI installs a mock and returns a cleanup function.
func setMockCouponAPI(mock *mockCouponAPI) func() {
	orig := newCouponAPI
	newCouponAPI = func(_ *cobra.Command) (CouponAPI, error) {
		return mock, nil
	}
	return func() { newCouponAPI = orig }
}

// sampleCoupon returns a test coupon with predictable fields for list tests.
func sampleCoupon() recurly.Coupon {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return recurly.Coupon{
		Id:        "coupon-abc123",
		Code:      "SAVE25",
		Name:      "Save 25%",
		State:     "redeemable",
		Discount:  recurly.CouponDiscount{Type: "percent", Percent: 25},
		CreatedAt: &now,
	}
}

// sampleCouponDetail returns a test coupon with all detail fields populated.
func sampleCouponDetail() *recurly.Coupon {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC)
	redeemBy := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	return &recurly.Coupon{
		Id:                       "coupon-abc123",
		Code:                     "SAVE25",
		Name:                     "Save 25%",
		State:                    "redeemable",
		Discount:                 recurly.CouponDiscount{Type: "percent", Percent: 25},
		Duration:                 "forever",
		CouponType:               "single_code",
		MaxRedemptions:           100,
		MaxRedemptionsPerAccount: 1,
		RedeemBy:                 &redeemBy,
		AppliesToAllPlans:        true,
		AppliesToAllItems:        false,
		CreatedAt:                &now,
		UpdatedAt:                &updated,
	}
}

// --- coupons list ---

func TestCouponsList_Success(t *testing.T) {
	coupon := sampleCoupon()
	mock := &mockCouponAPI{
		listCouponsFn: func(params *recurly.ListCouponsParams, opts ...recurly.Option) (recurly.CouponLister, error) {
			return &mockCouponLister{coupons: []recurly.Coupon{coupon}}, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"SAVE25", "Save 25%", "percent", "redeemable"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected table output to contain %q, got:\n%s", expected, out)
		}
	}
	for _, header := range []string{"Code", "Name", "Discount Type", "State", "Created At"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestCouponsList_EmptyResults(t *testing.T) {
	mock := &mockCouponAPI{
		listCouponsFn: func(params *recurly.ListCouponsParams, opts ...recurly.Option) (recurly.CouponLister, error) {
			return &mockCouponLister{coupons: []recurly.Coupon{}}, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, "SAVE25") {
		t.Error("expected no coupon data in empty results")
	}
}

func TestCouponsList_SDKError(t *testing.T) {
	mock := &mockCouponAPI{
		listCouponsFn: func(params *recurly.ListCouponsParams, opts ...recurly.Option) (recurly.CouponLister, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "list")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestCouponsList_FlagPassthrough(t *testing.T) {
	var capturedParams *recurly.ListCouponsParams

	mock := &mockCouponAPI{
		listCouponsFn: func(params *recurly.ListCouponsParams, opts ...recurly.Option) (recurly.CouponLister, error) {
			capturedParams = params
			return &mockCouponLister{coupons: []recurly.Coupon{}}, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "list", "--limit", "50", "--order", "desc", "--sort", "updated_at")
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

func TestCouponsList_UnsetFlagsNotSent(t *testing.T) {
	var capturedParams *recurly.ListCouponsParams

	mock := &mockCouponAPI{
		listCouponsFn: func(params *recurly.ListCouponsParams, opts ...recurly.Option) (recurly.CouponLister, error) {
			capturedParams = params
			return &mockCouponLister{coupons: []recurly.Coupon{}}, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "list")
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
	if capturedParams.BeginTime != nil {
		t.Error("expected begin_time to be nil when not set")
	}
	if capturedParams.EndTime != nil {
		t.Error("expected end_time to be nil when not set")
	}
}

// --- coupons get ---

func TestCouponsGet_Success(t *testing.T) {
	mock := &mockCouponAPI{
		getCouponFn: func(couponId string, opts ...recurly.Option) (*recurly.Coupon, error) {
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "get", "SAVE25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{
		"coupon-abc123", "SAVE25", "Save 25%", "redeemable",
		"percent", "25%", "forever", "single_code",
		"100", "1",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestCouponsGet_PositionalArg(t *testing.T) {
	var capturedID string
	mock := &mockCouponAPI{
		getCouponFn: func(couponId string, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedID = couponId
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "get", "my-coupon-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "my-coupon-id" {
		t.Errorf("expected coupon ID 'my-coupon-id', got %q", capturedID)
	}
}

func TestCouponsGet_NotFoundError(t *testing.T) {
	mock := &mockCouponAPI{
		getCouponFn: func(couponId string, opts ...recurly.Option) (*recurly.Coupon, error) {
			return nil, &recurly.Error{
				Type:    recurly.ErrorTypeNotFound,
				Message: "Couldn't find Coupon with id = nonexistent",
			}
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("coupons", "get", "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found coupon")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got %q", stderr)
	}
}

func TestCouponsGet_MissingArg_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand("coupons", "get")
	if err == nil {
		t.Fatal("expected error when no coupon ID is provided")
	}
	if !strings.Contains(stderr, "accepts 1 arg") {
		t.Errorf("expected usage error about missing argument, got %q", stderr)
	}
}

// --- coupons create-percent ---

func TestCouponsCreatePercent_Success(t *testing.T) {
	var capturedBody *recurly.CouponCreate
	mock := &mockCouponAPI{
		createCouponFn: func(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedBody = body
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "create-percent",
		"--code", "SAVE25",
		"--name", "Save 25%",
		"--discount-percent", "25",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if *capturedBody.Code != "SAVE25" {
		t.Errorf("expected code=SAVE25, got %v", *capturedBody.Code)
	}
	if *capturedBody.Name != "Save 25%" {
		t.Errorf("expected name=Save 25%%, got %v", *capturedBody.Name)
	}
	if *capturedBody.DiscountType != "percent" {
		t.Errorf("expected discount_type=percent, got %v", *capturedBody.DiscountType)
	}
	if *capturedBody.DiscountPercent != 25 {
		t.Errorf("expected discount_percent=25, got %v", *capturedBody.DiscountPercent)
	}

	if !strings.Contains(out, "SAVE25") {
		t.Errorf("expected output to contain coupon code, got:\n%s", out)
	}
}

func TestCouponsCreatePercent_AllOptionalFlags(t *testing.T) {
	var capturedBody *recurly.CouponCreate
	mock := &mockCouponAPI{
		createCouponFn: func(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedBody = body
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "create-percent",
		"--code", "FULL",
		"--name", "Full Coupon",
		"--discount-percent", "50",
		"--max-redemptions", "100",
		"--max-redemptions-per-account", "1",
		"--duration", "temporal",
		"--temporal-amount", "3",
		"--temporal-unit", "month",
		"--coupon-type", "bulk",
		"--unique-code-template", "TMPL",
		"--applies-to-all-plans",
		"--applies-to-all-items",
		"--applies-to-non-plan-charges",
		"--plan-codes", "plan1,plan2",
		"--item-codes", "item1,item2",
		"--redemption-resource", "subscription",
		"--hosted-page-description", "Hosted desc",
		"--invoice-description", "Invoice desc",
		"--redeem-by", "2025-12-31",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedBody.MaxRedemptions != 100 {
		t.Errorf("expected max_redemptions=100, got %v", *capturedBody.MaxRedemptions)
	}
	if *capturedBody.MaxRedemptionsPerAccount != 1 {
		t.Errorf("expected max_redemptions_per_account=1, got %v", *capturedBody.MaxRedemptionsPerAccount)
	}
	if *capturedBody.Duration != "temporal" {
		t.Errorf("expected duration=temporal, got %v", *capturedBody.Duration)
	}
	if *capturedBody.TemporalAmount != 3 {
		t.Errorf("expected temporal_amount=3, got %v", *capturedBody.TemporalAmount)
	}
	if *capturedBody.TemporalUnit != "month" {
		t.Errorf("expected temporal_unit=month, got %v", *capturedBody.TemporalUnit)
	}
	if *capturedBody.CouponType != "bulk" {
		t.Errorf("expected coupon_type=bulk, got %v", *capturedBody.CouponType)
	}
	if *capturedBody.UniqueCodeTemplate != "TMPL" {
		t.Errorf("expected unique_code_template=TMPL, got %v", *capturedBody.UniqueCodeTemplate)
	}
	if *capturedBody.AppliesToAllPlans != true {
		t.Error("expected applies_to_all_plans=true")
	}
	if *capturedBody.AppliesToAllItems != true {
		t.Error("expected applies_to_all_items=true")
	}
	if *capturedBody.AppliesToNonPlanCharges != true {
		t.Error("expected applies_to_non_plan_charges=true")
	}
	if capturedBody.PlanCodes == nil || len(*capturedBody.PlanCodes) != 2 {
		t.Errorf("expected 2 plan codes, got %v", capturedBody.PlanCodes)
	}
	if capturedBody.ItemCodes == nil || len(*capturedBody.ItemCodes) != 2 {
		t.Errorf("expected 2 item codes, got %v", capturedBody.ItemCodes)
	}
	if *capturedBody.RedemptionResource != "subscription" {
		t.Errorf("expected redemption_resource=subscription, got %v", *capturedBody.RedemptionResource)
	}
	if *capturedBody.HostedDescription != "Hosted desc" {
		t.Errorf("expected hosted_description=Hosted desc, got %v", *capturedBody.HostedDescription)
	}
	if *capturedBody.InvoiceDescription != "Invoice desc" {
		t.Errorf("expected invoice_description=Invoice desc, got %v", *capturedBody.InvoiceDescription)
	}
	if *capturedBody.RedeemByDate != "2025-12-31" {
		t.Errorf("expected redeem_by_date=2025-12-31, got %v", *capturedBody.RedeemByDate)
	}
}

func TestCouponsCreatePercent_MissingRequiredFlags(t *testing.T) {
	mock := &mockCouponAPI{
		createCouponFn: func(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error) {
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	// Missing --code
	_, stderr, err := executeCommand("coupons", "create-percent",
		"--no-input",
		"--name", "Test",
		"--discount-percent", "10",
	)
	if err == nil {
		t.Fatal("expected error for missing --code")
	}
	if !strings.Contains(stderr, "code") {
		t.Errorf("expected error about missing code flag, got %q", stderr)
	}

	// Missing --name
	_, stderr, err = executeCommand("coupons", "create-percent",
		"--no-input",
		"--code", "TEST",
		"--discount-percent", "10",
	)
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
	if !strings.Contains(stderr, "name") {
		t.Errorf("expected error about missing name flag, got %q", stderr)
	}

	// Missing --discount-percent
	_, stderr, err = executeCommand("coupons", "create-percent",
		"--no-input",
		"--code", "TEST",
		"--name", "Test",
	)
	if err == nil {
		t.Fatal("expected error for missing --discount-percent")
	}
	if !strings.Contains(stderr, "discount-percent") {
		t.Errorf("expected error about missing discount-percent flag, got %q", stderr)
	}
}

func TestCouponsCreatePercent_UnsetOptionalFlagsNotSent(t *testing.T) {
	var capturedBody *recurly.CouponCreate
	mock := &mockCouponAPI{
		createCouponFn: func(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedBody = body
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "create-percent",
		"--code", "MINIMAL",
		"--name", "Minimal",
		"--discount-percent", "10",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.MaxRedemptions != nil {
		t.Error("expected max_redemptions to be nil when not set")
	}
	if capturedBody.MaxRedemptionsPerAccount != nil {
		t.Error("expected max_redemptions_per_account to be nil when not set")
	}
	if capturedBody.Duration != nil {
		t.Error("expected duration to be nil when not set")
	}
	if capturedBody.TemporalAmount != nil {
		t.Error("expected temporal_amount to be nil when not set")
	}
	if capturedBody.TemporalUnit != nil {
		t.Error("expected temporal_unit to be nil when not set")
	}
	if capturedBody.CouponType != nil {
		t.Error("expected coupon_type to be nil when not set")
	}
	if capturedBody.UniqueCodeTemplate != nil {
		t.Error("expected unique_code_template to be nil when not set")
	}
	if capturedBody.AppliesToAllPlans != nil {
		t.Error("expected applies_to_all_plans to be nil when not set")
	}
	if capturedBody.AppliesToAllItems != nil {
		t.Error("expected applies_to_all_items to be nil when not set")
	}
	if capturedBody.AppliesToNonPlanCharges != nil {
		t.Error("expected applies_to_non_plan_charges to be nil when not set")
	}
	if capturedBody.PlanCodes != nil {
		t.Error("expected plan_codes to be nil when not set")
	}
	if capturedBody.ItemCodes != nil {
		t.Error("expected item_codes to be nil when not set")
	}
	if capturedBody.RedemptionResource != nil {
		t.Error("expected redemption_resource to be nil when not set")
	}
	if capturedBody.HostedDescription != nil {
		t.Error("expected hosted_description to be nil when not set")
	}
	if capturedBody.InvoiceDescription != nil {
		t.Error("expected invoice_description to be nil when not set")
	}
	if capturedBody.RedeemByDate != nil {
		t.Error("expected redeem_by_date to be nil when not set")
	}
}

// --- coupons create-fixed ---

func TestCouponsCreateFixed_Success(t *testing.T) {
	var capturedBody *recurly.CouponCreate
	mock := &mockCouponAPI{
		createCouponFn: func(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedBody = body
			detail := sampleCouponDetail()
			detail.Discount = recurly.CouponDiscount{
				Type: "fixed",
				Currencies: []recurly.CouponDiscountPricing{
					{Amount: 10.00, Currency: "USD"},
				},
			}
			return detail, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "create-fixed",
		"--code", "FIXED10",
		"--name", "Fixed $10",
		"--currency", "USD",
		"--discount-amount", "10.00",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if *capturedBody.Code != "FIXED10" {
		t.Errorf("expected code=FIXED10, got %v", *capturedBody.Code)
	}
	if *capturedBody.DiscountType != "fixed" {
		t.Errorf("expected discount_type=fixed, got %v", *capturedBody.DiscountType)
	}
	if capturedBody.Currencies == nil {
		t.Fatal("expected currencies to be set")
	}
	currencies := *capturedBody.Currencies
	if len(currencies) != 1 {
		t.Fatalf("expected 1 currency, got %d", len(currencies))
	}
	if *currencies[0].Currency != "USD" {
		t.Errorf("expected currency=USD, got %v", *currencies[0].Currency)
	}
	if *currencies[0].Discount != 10.00 {
		t.Errorf("expected discount=10.00, got %v", *currencies[0].Discount)
	}

	if !strings.Contains(out, "fixed") {
		t.Errorf("expected output to contain discount type 'fixed', got:\n%s", out)
	}
	if !strings.Contains(out, "10.00 USD") {
		t.Errorf("expected output to contain '10.00 USD', got:\n%s", out)
	}
}

func TestCouponsCreateFixed_CurrencyAmountMismatch(t *testing.T) {
	mock := &mockCouponAPI{
		createCouponFn: func(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error) {
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("coupons", "create-fixed",
		"--code", "BAD",
		"--name", "Bad",
		"--currency", "USD", "--currency", "EUR",
		"--discount-amount", "10.00",
	)
	if err == nil {
		t.Fatal("expected error for mismatched currency/discount-amount")
	}
	if !strings.Contains(stderr, "--currency and --discount-amount must be specified in pairs") {
		t.Errorf("expected mismatch error, got %q", stderr)
	}
}

func TestCouponsCreateFixed_MultiCurrency(t *testing.T) {
	var capturedBody *recurly.CouponCreate
	mock := &mockCouponAPI{
		createCouponFn: func(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedBody = body
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "create-fixed",
		"--code", "MULTI",
		"--name", "Multi Currency",
		"--currency", "USD", "--currency", "EUR",
		"--discount-amount", "10.00", "--discount-amount", "9.00",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	currencies := *capturedBody.Currencies
	if len(currencies) != 2 {
		t.Fatalf("expected 2 currencies, got %d", len(currencies))
	}
	if *currencies[0].Currency != "USD" || *currencies[0].Discount != 10.00 {
		t.Errorf("expected USD/10.00, got %s/%.2f", *currencies[0].Currency, *currencies[0].Discount)
	}
	if *currencies[1].Currency != "EUR" || *currencies[1].Discount != 9.00 {
		t.Errorf("expected EUR/9.00, got %s/%.2f", *currencies[1].Currency, *currencies[1].Discount)
	}
}

// --- coupons create-free-trial ---

func TestCouponsCreateFreeTrial_Success(t *testing.T) {
	var capturedBody *recurly.CouponCreate
	mock := &mockCouponAPI{
		createCouponFn: func(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedBody = body
			detail := sampleCouponDetail()
			detail.Discount = recurly.CouponDiscount{
				Type:  "free_trial",
				Trial: recurly.CouponDiscountTrial{Length: 30, Unit: "day"},
			}
			return detail, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "create-free-trial",
		"--code", "FREETRIAL",
		"--name", "Free Trial 30 Days",
		"--free-trial-amount", "30",
		"--free-trial-unit", "day",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("expected body to be captured")
	}
	if *capturedBody.Code != "FREETRIAL" {
		t.Errorf("expected code=FREETRIAL, got %v", *capturedBody.Code)
	}
	if *capturedBody.DiscountType != "free_trial" {
		t.Errorf("expected discount_type=free_trial, got %v", *capturedBody.DiscountType)
	}
	if *capturedBody.FreeTrialAmount != 30 {
		t.Errorf("expected free_trial_amount=30, got %v", *capturedBody.FreeTrialAmount)
	}
	if *capturedBody.FreeTrialUnit != "day" {
		t.Errorf("expected free_trial_unit=day, got %v", *capturedBody.FreeTrialUnit)
	}

	if !strings.Contains(out, "30 day") {
		t.Errorf("expected output to contain '30 day', got:\n%s", out)
	}
}

func TestCouponsCreateFreeTrial_MissingRequiredFlags(t *testing.T) {
	mock := &mockCouponAPI{
		createCouponFn: func(body *recurly.CouponCreate, opts ...recurly.Option) (*recurly.Coupon, error) {
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	// Missing --free-trial-amount and --free-trial-unit
	_, stderr, err := executeCommand("coupons", "create-free-trial",
		"--no-input",
		"--code", "TEST",
		"--name", "Test",
	)
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
	if !strings.Contains(stderr, "free-trial-amount") || !strings.Contains(stderr, "free-trial-unit") {
		t.Errorf("expected error about missing free trial flags, got %q", stderr)
	}
}

// --- coupons update ---

func TestCouponsUpdate_SuccessWithChangedFlags(t *testing.T) {
	var capturedID string
	var capturedBody *recurly.CouponUpdate
	mock := &mockCouponAPI{
		updateCouponFn: func(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedID = couponId
			capturedBody = body
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "update", "SAVE25",
		"--name", "Updated Name",
		"--max-redemptions", "200",
		"--max-redemptions-per-account", "5",
		"--hosted-description", "New hosted desc",
		"--invoice-description", "New invoice desc",
		"--redeem-by-date", "2026-12-31",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedID != "SAVE25" {
		t.Errorf("expected coupon ID 'SAVE25', got %q", capturedID)
	}
	if *capturedBody.Name != "Updated Name" {
		t.Errorf("expected name=Updated Name, got %v", *capturedBody.Name)
	}
	if *capturedBody.MaxRedemptions != 200 {
		t.Errorf("expected max_redemptions=200, got %v", *capturedBody.MaxRedemptions)
	}
	if *capturedBody.MaxRedemptionsPerAccount != 5 {
		t.Errorf("expected max_redemptions_per_account=5, got %v", *capturedBody.MaxRedemptionsPerAccount)
	}
	if *capturedBody.HostedDescription != "New hosted desc" {
		t.Errorf("expected hosted_description=New hosted desc, got %v", *capturedBody.HostedDescription)
	}
	if *capturedBody.InvoiceDescription != "New invoice desc" {
		t.Errorf("expected invoice_description=New invoice desc, got %v", *capturedBody.InvoiceDescription)
	}
	if *capturedBody.RedeemByDate != "2026-12-31" {
		t.Errorf("expected redeem_by_date=2026-12-31, got %v", *capturedBody.RedeemByDate)
	}

	if !strings.Contains(out, "SAVE25") {
		t.Errorf("expected output to contain coupon details, got:\n%s", out)
	}
}

func TestCouponsUpdate_NoFlags_EmptyBody(t *testing.T) {
	var capturedBody *recurly.CouponUpdate
	mock := &mockCouponAPI{
		updateCouponFn: func(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedBody = body
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "update", "SAVE25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Name != nil {
		t.Error("expected name to be nil when not set")
	}
	if capturedBody.MaxRedemptions != nil {
		t.Error("expected max_redemptions to be nil when not set")
	}
	if capturedBody.MaxRedemptionsPerAccount != nil {
		t.Error("expected max_redemptions_per_account to be nil when not set")
	}
	if capturedBody.HostedDescription != nil {
		t.Error("expected hosted_description to be nil when not set")
	}
	if capturedBody.InvoiceDescription != nil {
		t.Error("expected invoice_description to be nil when not set")
	}
	if capturedBody.RedeemByDate != nil {
		t.Error("expected redeem_by_date to be nil when not set")
	}
}

func TestCouponsUpdate_SDKError(t *testing.T) {
	mock := &mockCouponAPI{
		updateCouponFn: func(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error) {
			return nil, fmt.Errorf("validation failed")
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "update", "SAVE25", "--name", "bad")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

func TestCouponsUpdate_MissingArg_ReturnsError(t *testing.T) {
	_, _, err := executeCommand("coupons", "update")
	if err == nil {
		t.Fatal("expected error when coupon_id is missing")
	}
}

// --- coupons deactivate ---

func TestCouponsDeactivate_YesFlag_Success(t *testing.T) {
	var capturedID string
	mock := &mockCouponAPI{
		deactivateCouponFn: func(couponId string, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedID = couponId
			detail := sampleCouponDetail()
			detail.State = "inactive"
			return detail, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, stderr, err := executeCommand("coupons", "deactivate", "SAVE25", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "SAVE25" {
		t.Errorf("expected coupon ID 'SAVE25', got %q", capturedID)
	}
	if strings.Contains(stderr, "Are you sure") {
		t.Error("expected no confirmation prompt with --yes flag")
	}
	if !strings.Contains(out, "SAVE25") {
		t.Errorf("expected coupon details in output, got:\n%s", out)
	}
}

func TestCouponsDeactivate_ConfirmYes_Succeeds(t *testing.T) {
	var capturedID string
	mock := &mockCouponAPI{
		deactivateCouponFn: func(couponId string, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedID = couponId
			detail := sampleCouponDetail()
			detail.State = "inactive"
			return detail, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	stdin := bytes.NewBufferString("y\n")
	out, stderr, err := executeCommandWithStdin(stdin, "coupons", "deactivate", "SAVE25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "SAVE25" {
		t.Errorf("expected coupon ID 'SAVE25', got %q", capturedID)
	}
	if !strings.Contains(stderr, "Are you sure") {
		t.Error("expected confirmation prompt")
	}
	if !strings.Contains(out, "SAVE25") {
		t.Errorf("expected coupon details in output, got:\n%s", out)
	}
}

func TestCouponsDeactivate_ConfirmNo_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(stdin, "coupons", "deactivate", "SAVE25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Are you sure you want to deactivate coupon SAVE25? [y/N]") {
		t.Error("expected confirmation prompt with coupon ID in stderr")
	}
	if !strings.Contains(stderr, "Deactivation cancelled.") {
		t.Error("expected cancellation message")
	}
}

func TestCouponsDeactivate_ConfirmDefault_Cancels(t *testing.T) {
	stdin := bytes.NewBufferString("\n")
	_, stderr, err := executeCommandWithStdin(stdin, "coupons", "deactivate", "SAVE25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deactivation cancelled.") {
		t.Error("expected cancellation message when pressing Enter without input")
	}
}

func TestCouponsDeactivate_MissingArg_ReturnsError(t *testing.T) {
	_, _, err := executeCommand("coupons", "deactivate")
	if err == nil {
		t.Fatal("expected error for missing argument")
	}
}

func TestCouponsDeactivate_SDKError(t *testing.T) {
	mock := &mockCouponAPI{
		deactivateCouponFn: func(couponId string, opts ...recurly.Option) (*recurly.Coupon, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "deactivate", "SAVE25", "--yes")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- coupons restore ---

func TestCouponsRestore_Success(t *testing.T) {
	var capturedID string
	var capturedBody *recurly.CouponUpdate
	mock := &mockCouponAPI{
		restoreCouponFn: func(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedID = couponId
			capturedBody = body
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "restore", "SAVE25",
		"--name", "Restored Coupon",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedID != "SAVE25" {
		t.Errorf("expected coupon ID 'SAVE25', got %q", capturedID)
	}
	if *capturedBody.Name != "Restored Coupon" {
		t.Errorf("expected name=Restored Coupon, got %v", *capturedBody.Name)
	}
	if !strings.Contains(out, "SAVE25") {
		t.Errorf("expected coupon details in output, got:\n%s", out)
	}
}

func TestCouponsRestore_NoFlags(t *testing.T) {
	var capturedBody *recurly.CouponUpdate
	mock := &mockCouponAPI{
		restoreCouponFn: func(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error) {
			capturedBody = body
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "restore", "SAVE25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.Name != nil {
		t.Error("expected name to be nil when not set")
	}
	if capturedBody.MaxRedemptions != nil {
		t.Error("expected max_redemptions to be nil when not set")
	}
}

func TestCouponsRestore_MissingArg_ReturnsError(t *testing.T) {
	_, _, err := executeCommand("coupons", "restore")
	if err == nil {
		t.Fatal("expected error when coupon_id is missing")
	}
}

func TestCouponsRestore_SDKError(t *testing.T) {
	mock := &mockCouponAPI{
		restoreCouponFn: func(couponId string, body *recurly.CouponUpdate, opts ...recurly.Option) (*recurly.Coupon, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "restore", "SAVE25")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- coupons generate-codes ---

func TestCouponsGenerateCodes_Success(t *testing.T) {
	var capturedID string
	var capturedBody *recurly.CouponBulkCreate
	beginTime := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	mock := &mockCouponAPI{
		generateUniqueCouponCodesFn: func(couponId string, body *recurly.CouponBulkCreate, opts ...recurly.Option) (*recurly.UniqueCouponCodeParams, error) {
			capturedID = couponId
			capturedBody = body
			return &recurly.UniqueCouponCodeParams{
				Limit:     200,
				Order:     "asc",
				Sort:      "created_at",
				BeginTime: &beginTime,
			}, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "generate-codes", "BULK-COUPON",
		"--number-of-codes", "50",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedID != "BULK-COUPON" {
		t.Errorf("expected coupon ID 'BULK-COUPON', got %q", capturedID)
	}
	if *capturedBody.NumberOfUniqueCodes != 50 {
		t.Errorf("expected number_of_unique_codes=50, got %v", *capturedBody.NumberOfUniqueCodes)
	}

	if !strings.Contains(out, "200") {
		t.Errorf("expected output to contain limit '200', got:\n%s", out)
	}
	if !strings.Contains(out, "asc") {
		t.Errorf("expected output to contain order 'asc', got:\n%s", out)
	}
}

func TestCouponsGenerateCodes_MissingRequiredFlag(t *testing.T) {
	_, stderr, err := executeCommand("coupons", "generate-codes", "BULK-COUPON", "--no-input")
	if err == nil {
		t.Fatal("expected error for missing --number-of-codes")
	}
	if !strings.Contains(stderr, "number-of-codes") {
		t.Errorf("expected error about missing flag, got %q", stderr)
	}
}

func TestCouponsGenerateCodes_ZeroCodes_ReturnsError(t *testing.T) {
	mock := &mockCouponAPI{
		generateUniqueCouponCodesFn: func(couponId string, body *recurly.CouponBulkCreate, opts ...recurly.Option) (*recurly.UniqueCouponCodeParams, error) {
			return nil, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, stderr, err := executeCommand("coupons", "generate-codes", "BULK-COUPON",
		"--number-of-codes", "0",
	)
	if err == nil {
		t.Fatal("expected error for zero codes")
	}
	if !strings.Contains(stderr, "must be at least 1") {
		t.Errorf("expected validation error, got %q", stderr)
	}
}

func TestCouponsGenerateCodes_MissingArg_ReturnsError(t *testing.T) {
	_, _, err := executeCommand("coupons", "generate-codes")
	if err == nil {
		t.Fatal("expected error when coupon_id is missing")
	}
}

func TestCouponsGenerateCodes_SDKError(t *testing.T) {
	mock := &mockCouponAPI{
		generateUniqueCouponCodesFn: func(couponId string, body *recurly.CouponBulkCreate, opts ...recurly.Option) (*recurly.UniqueCouponCodeParams, error) {
			return nil, fmt.Errorf("server error")
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "generate-codes", "BULK-COUPON",
		"--number-of-codes", "10",
	)
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- coupons list-codes ---

func TestCouponsListCodes_Success(t *testing.T) {
	now := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	codes := []recurly.UniqueCouponCode{
		{
			Code:           "ABC123",
			State:          "redeemable",
			BulkCouponCode: "BULK-COUPON",
			CreatedAt:      &now,
		},
		{
			Code:           "DEF456",
			State:          "redeemed",
			BulkCouponCode: "BULK-COUPON",
			RedeemedAt:     &now,
			CreatedAt:      &now,
		},
	}
	mock := &mockCouponAPI{
		listUniqueCouponCodesFn: func(couponId string, params *recurly.ListUniqueCouponCodesParams, opts ...recurly.Option) (recurly.UniqueCouponCodeLister, error) {
			return &mockUniqueCouponCodeLister{codes: codes}, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "list-codes", "BULK-COUPON")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"ABC123", "DEF456", "redeemable", "redeemed", "BULK-COUPON"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
	for _, header := range []string{"Code", "State", "Bulk Coupon Code", "Created At"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestCouponsListCodes_Pagination(t *testing.T) {
	var capturedParams *recurly.ListUniqueCouponCodesParams
	var capturedID string
	mock := &mockCouponAPI{
		listUniqueCouponCodesFn: func(couponId string, params *recurly.ListUniqueCouponCodesParams, opts ...recurly.Option) (recurly.UniqueCouponCodeLister, error) {
			capturedID = couponId
			capturedParams = params
			return &mockUniqueCouponCodeLister{codes: []recurly.UniqueCouponCode{}}, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "list-codes", "BULK-COUPON",
		"--limit", "25",
		"--order", "desc",
		"--sort", "created_at",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedID != "BULK-COUPON" {
		t.Errorf("expected coupon ID 'BULK-COUPON', got %q", capturedID)
	}
	if capturedParams.Limit == nil || *capturedParams.Limit != 25 {
		t.Errorf("expected limit=25, got %v", capturedParams.Limit)
	}
	if capturedParams.Order == nil || *capturedParams.Order != "desc" {
		t.Errorf("expected order=desc, got %v", capturedParams.Order)
	}
	if capturedParams.Sort == nil || *capturedParams.Sort != "created_at" {
		t.Errorf("expected sort=created_at, got %v", capturedParams.Sort)
	}
}

func TestCouponsListCodes_EmptyResults(t *testing.T) {
	mock := &mockCouponAPI{
		listUniqueCouponCodesFn: func(couponId string, params *recurly.ListUniqueCouponCodesParams, opts ...recurly.Option) (recurly.UniqueCouponCodeLister, error) {
			return &mockUniqueCouponCodeLister{codes: []recurly.UniqueCouponCode{}}, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "list-codes", "BULK-COUPON")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, "ABC123") {
		t.Error("expected no code data in empty results")
	}
}

func TestCouponsListCodes_MissingArg_ReturnsError(t *testing.T) {
	_, _, err := executeCommand("coupons", "list-codes")
	if err == nil {
		t.Fatal("expected error when coupon_id is missing")
	}
}

func TestCouponsListCodes_SDKError(t *testing.T) {
	mock := &mockCouponAPI{
		listUniqueCouponCodesFn: func(couponId string, params *recurly.ListUniqueCouponCodesParams, opts ...recurly.Option) (recurly.UniqueCouponCodeLister, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	_, _, err := executeCommand("coupons", "list-codes", "BULK-COUPON")
	if err == nil {
		t.Fatal("expected error from SDK")
	}
}

// --- JSON output tests ---

func TestCouponsGet_JSONOutput(t *testing.T) {
	viper.Reset()
	mock := &mockCouponAPI{
		getCouponFn: func(couponId string, opts ...recurly.Option) (*recurly.Coupon, error) {
			return sampleCouponDetail(), nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "get", "SAVE25", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON
	if !strings.Contains(strings.TrimSpace(out), "{") {
		t.Errorf("expected JSON output, got:\n%s", out)
	}
}

func TestCouponsList_JSONOutput(t *testing.T) {
	viper.Reset()
	coupon := sampleCoupon()
	mock := &mockCouponAPI{
		listCouponsFn: func(params *recurly.ListCouponsParams, opts ...recurly.Option) (recurly.CouponLister, error) {
			return &mockCouponLister{coupons: []recurly.Coupon{coupon}}, nil
		},
	}
	cleanup := setMockCouponAPI(mock)
	defer cleanup()

	out, _, err := executeCommand("coupons", "list", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(strings.TrimSpace(out), "\"object\"") {
		t.Errorf("expected JSON list envelope, got:\n%s", out)
	}
}
