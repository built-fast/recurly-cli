package cmd

import (
	"time"

	recurly "github.com/recurly/recurly-client-go/v5"
)

// ---------------------------------------------------------------------------
// Shared sample data factories for tests.
//
// Each factory returns a fully populated resource struct suitable for both
// list and detail test scenarios. Factories live here so that test data is
// consistent and not duplicated across test files.
// ---------------------------------------------------------------------------

// --- Account ---

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

// --- Subscription ---

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

// --- Invoice ---

func sampleInvoice() *recurly.Invoice {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2025, 1, 16, 8, 0, 0, 0, time.UTC)
	due := time.Date(2025, 2, 14, 10, 30, 0, 0, time.UTC)
	closed := time.Date(2025, 1, 20, 12, 0, 0, 0, time.UTC)

	return &recurly.Invoice{
		Id:               "inv-abc123",
		Uuid:             "uuid-abc123",
		Number:           "1001",
		Type:             "charge",
		Origin:           "purchase",
		State:            "paid",
		Account:          recurly.AccountMini{Id: "acct-id-1", Code: "acct-code-1"},
		CollectionMethod: "automatic",
		Currency:         "USD",
		Subtotal:         100.00,
		Discount:         10.00,
		Tax:              8.10,
		Total:            98.10,
		Paid:             98.10,
		Balance:          0.00,
		RefundableAmount: 98.10,
		PoNumber:         "PO-12345",
		NetTerms:         30,
		NetTermsType:     "net",
		CreatedAt:        &now,
		UpdatedAt:        &updated,
		DueAt:            &due,
		ClosedAt:         &closed,
	}
}

func sampleInvoices() []recurly.Invoice {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return []recurly.Invoice{
		{
			Id:        "inv-001",
			Number:    "1001",
			Type:      "charge",
			Account:   recurly.AccountMini{Code: "acct-code-1"},
			State:     "paid",
			Currency:  "USD",
			Total:     100.00,
			Balance:   0.00,
			CreatedAt: &now,
		},
		{
			Id:        "inv-002",
			Number:    "1002",
			Type:      "credit",
			Account:   recurly.AccountMini{Code: "acct-code-2"},
			State:     "pending",
			Currency:  "EUR",
			Total:     50.00,
			Balance:   50.00,
			CreatedAt: &now,
		},
	}
}

func sampleLineItems() []recurly.LineItem {
	return []recurly.LineItem{
		{
			Id:          "li-001",
			Type:        "charge",
			Description: "Monthly subscription",
			Currency:    "USD",
			UnitAmount:  50.00,
			Quantity:    1,
			Subtotal:    50.00,
			Tax:         4.05,
			Amount:      54.05,
		},
		{
			Id:          "li-002",
			Type:        "charge",
			Description: "Add-on feature",
			Currency:    "USD",
			UnitAmount:  50.00,
			Quantity:    1,
			Subtotal:    50.00,
			Tax:         4.05,
			Amount:      54.05,
		},
	}
}

// sampleAccountInvoice returns a minimal invoice for account-nested list tests.
func sampleAccountInvoice() recurly.Invoice {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return recurly.Invoice{
		Id:        "inv-123",
		Number:    "1001",
		State:     "paid",
		Type:      "charge",
		Currency:  "USD",
		Total:     49.99,
		Account:   recurly.AccountMini{Code: "acct-456"},
		CreatedAt: &now,
	}
}

// --- Transaction ---

func sampleTransaction() *recurly.Transaction {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return &recurly.Transaction{
		Id:               "txn-001",
		Uuid:             "uuid-001",
		Type:             "purchase",
		Origin:           "api",
		Status:           "success",
		Success:          true,
		Amount:           99.99,
		Currency:         "USD",
		Account:          recurly.AccountMini{Id: "acct-id-1", Code: "acct-code-1"},
		Invoice:          recurly.InvoiceMini{Id: "inv-id-1", Number: "1001"},
		CollectionMethod: "automatic",
		PaymentMethod: recurly.PaymentMethod{
			Object:   "credit_card",
			CardType: "Visa",
			LastFour: "1234",
		},
		IpAddressV4:   "192.168.1.1",
		StatusCode:    "approved",
		StatusMessage: "Transaction approved",
		Refunded:      false,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
}

func sampleTransactions() []recurly.Transaction {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return []recurly.Transaction{
		{
			Id:        "txn-001",
			Type:      "purchase",
			Account:   recurly.AccountMini{Code: "acct-code-1"},
			Status:    "success",
			Currency:  "USD",
			Amount:    99.99,
			Success:   true,
			Origin:    "api",
			CreatedAt: &now,
		},
		{
			Id:        "txn-002",
			Type:      "refund",
			Account:   recurly.AccountMini{Code: "acct-code-2"},
			Status:    "declined",
			Currency:  "EUR",
			Amount:    49.50,
			Success:   false,
			Origin:    "recurly_admin",
			CreatedAt: &now,
		},
	}
}

// sampleAccountTransaction returns a minimal transaction for account-nested list tests.
func sampleAccountTransaction() recurly.Transaction {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return recurly.Transaction{
		Id:        "txn-123",
		Type:      "purchase",
		Amount:    29.99,
		Currency:  "USD",
		Status:    "success",
		Success:   true,
		Account:   recurly.AccountMini{Code: "acct-456"},
		CreatedAt: &now,
	}
}

// --- Plan ---

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

// --- Coupon ---

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

// --- Item ---

func sampleItem() recurly.Item {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	return recurly.Item{
		Code:        "widget-1",
		Name:        "Premium Widget",
		ExternalSku: "SKU-001",
		State:       "active",
		CreatedAt:   &now,
	}
}

func sampleItemDetail() *recurly.Item {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC)
	return &recurly.Item{
		Code:                   "widget-1",
		Name:                   "Premium Widget",
		Description:            "A high-quality widget",
		ExternalSku:            "SKU-001",
		AccountingCode:         "ACC-100",
		RevenueScheduleType:    "evenly",
		TaxCode:                "digital",
		TaxExempt:              false,
		AvalaraTransactionType: 3,
		AvalaraServiceType:     6,
		HarmonizedSystemCode:   "8471.30",
		State:                  "active",
		CreatedAt:              &now,
		UpdatedAt:              &updated,
	}
}

// --- Add-On ---

func sampleAddOn() recurly.AddOn {
	now := time.Date(2025, 3, 10, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2025, 3, 15, 14, 0, 0, 0, time.UTC)
	return recurly.AddOn{
		Id:              "a1234",
		Code:            "extra-users",
		Name:            "Extra Users",
		State:           "active",
		AddOnType:       "fixed",
		DefaultQuantity: 1,
		Optional:        true,
		AccountingCode:  "extra-users-ac",
		TaxCode:         "SW054000",
		Currencies: []recurly.AddOnPricing{
			{Currency: "USD", UnitAmount: 5.00},
			{Currency: "EUR", UnitAmount: 4.50},
		},
		CreatedAt: &now,
		UpdatedAt: &updated,
	}
}

// --- Billing Info ---

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

// --- Coupon Redemption ---

func sampleRedemption() recurly.CouponRedemption {
	now := time.Date(2025, 3, 10, 12, 0, 0, 0, time.UTC)
	return recurly.CouponRedemption{
		Id:         "redemption-abc123",
		State:      "active",
		Currency:   "USD",
		Discounted: 10.50,
		Account:    recurly.AccountMini{Code: "acct-1"},
		Coupon:     recurly.Coupon{Code: "SAVE10"},
		CreatedAt:  &now,
	}
}

func sampleRedemptionDetail() *recurly.CouponRedemption {
	now := time.Date(2025, 3, 10, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2025, 3, 15, 14, 0, 0, 0, time.UTC)
	return &recurly.CouponRedemption{
		Id:             "redemption-abc123",
		State:          "active",
		Currency:       "USD",
		Discounted:     10.50,
		SubscriptionId: "sub-xyz",
		Account:        recurly.AccountMini{Code: "acct-1"},
		Coupon:         recurly.Coupon{Code: "SAVE10"},
		CreatedAt:      &now,
		UpdatedAt:      &updated,
	}
}
