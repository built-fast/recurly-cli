package output

import (
	"testing"
	"time"
)

type helperTestItem struct {
	Name      string
	Active    bool
	Count     int
	Amount    float64
	CreatedAt *time.Time
}

func TestStringColumn(t *testing.T) {
	col := StringColumn("Name", func(v helperTestItem) string { return v.Name })
	if col.Header != "Name" {
		t.Errorf("expected header 'Name', got %q", col.Header)
	}
	if got := col.Extract(helperTestItem{Name: "Alice"}); got != "Alice" {
		t.Errorf("expected 'Alice', got %q", got)
	}
	if got := col.Extract(helperTestItem{}); got != "" {
		t.Errorf("expected empty string for zero value, got %q", got)
	}
}

func TestTimeColumnNonNil(t *testing.T) {
	ts := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	col := TimeColumn("Created At", func(v helperTestItem) *time.Time { return v.CreatedAt })
	got := col.Extract(helperTestItem{CreatedAt: &ts})
	want := "2024-06-15T12:30:00Z"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestTimeColumnNil(t *testing.T) {
	col := TimeColumn("Created At", func(v helperTestItem) *time.Time { return v.CreatedAt })
	got := col.Extract(helperTestItem{CreatedAt: nil})
	if got != "" {
		t.Errorf("expected empty string for nil time, got %q", got)
	}
}

func TestBoolColumn(t *testing.T) {
	col := BoolColumn("Active", func(v helperTestItem) bool { return v.Active })
	if col.Header != "Active" {
		t.Errorf("expected header 'Active', got %q", col.Header)
	}
	if got := col.Extract(helperTestItem{Active: true}); got != "true" {
		t.Errorf("expected 'true', got %q", got)
	}
	if got := col.Extract(helperTestItem{Active: false}); got != "false" {
		t.Errorf("expected 'false', got %q", got)
	}
}

func TestIntColumn(t *testing.T) {
	col := IntColumn("Count", func(v helperTestItem) int { return v.Count })
	if col.Header != "Count" {
		t.Errorf("expected header 'Count', got %q", col.Header)
	}
	if got := col.Extract(helperTestItem{Count: 42}); got != "42" {
		t.Errorf("expected '42', got %q", got)
	}
	if got := col.Extract(helperTestItem{Count: 0}); got != "0" {
		t.Errorf("expected '0', got %q", got)
	}
	if got := col.Extract(helperTestItem{Count: -7}); got != "-7" {
		t.Errorf("expected '-7', got %q", got)
	}
}

func TestFloatColumn(t *testing.T) {
	col := FloatColumn("Amount", func(v helperTestItem) float64 { return v.Amount })
	if col.Header != "Amount" {
		t.Errorf("expected header 'Amount', got %q", col.Header)
	}
	if got := col.Extract(helperTestItem{Amount: 19.99}); got != "19.99" {
		t.Errorf("expected '19.99', got %q", got)
	}
	if got := col.Extract(helperTestItem{Amount: 100.0}); got != "100.00" {
		t.Errorf("expected '100.00', got %q", got)
	}
	if got := col.Extract(helperTestItem{Amount: 0}); got != "0.00" {
		t.Errorf("expected '0.00', got %q", got)
	}
}

func TestHelperColumnsWorkWithToColumns(t *testing.T) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	typed := []TypedColumn[helperTestItem]{
		StringColumn("Name", func(v helperTestItem) string { return v.Name }),
		TimeColumn("Created At", func(v helperTestItem) *time.Time { return v.CreatedAt }),
		BoolColumn("Active", func(v helperTestItem) bool { return v.Active }),
		IntColumn("Count", func(v helperTestItem) int { return v.Count }),
		FloatColumn("Amount", func(v helperTestItem) float64 { return v.Amount }),
	}

	cols := ToColumns(typed)
	item := helperTestItem{
		Name:      "Test",
		Active:    true,
		Count:     5,
		Amount:    9.99,
		CreatedAt: &ts,
	}

	expected := []string{"Test", "2024-01-01T00:00:00Z", "true", "5", "9.99"}
	for i, col := range cols {
		got := col.Extract(item)
		if got != expected[i] {
			t.Errorf("column %d (%s): expected %q, got %q", i, col.Header, expected[i], got)
		}
	}
}
