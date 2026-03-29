package output

import (
	"testing"
)

func TestToColumnsExtract(t *testing.T) {
	t.Parallel()
	typed := []TypedColumn[testItem]{
		{Header: "Name", Extract: func(v testItem) string { return v.Name }},
		{Header: "Email", Extract: func(v testItem) string { return v.Email }},
	}

	cols := ToColumns(typed)

	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if cols[0].Header != "Name" {
		t.Errorf("expected header 'Name', got %q", cols[0].Header)
	}
	if cols[1].Header != "Email" {
		t.Errorf("expected header 'Email', got %q", cols[1].Header)
	}

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	if got := cols[0].Extract(item); got != "Alice" {
		t.Errorf("expected 'Alice', got %q", got)
	}
	if got := cols[1].Extract(item); got != "alice@example.com" {
		t.Errorf("expected 'alice@example.com', got %q", got)
	}
}

func TestToColumnsPanicsOnTypeMismatch(t *testing.T) {
	t.Parallel()
	typed := []TypedColumn[testItem]{
		{Header: "Name", Extract: func(v testItem) string { return v.Name }},
	}

	cols := ToColumns(typed)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on type mismatch, got none")
		}
	}()

	// Passing a string instead of testItem should panic.
	cols[0].Extract("not a testItem")
}

func TestToColumnsEmpty(t *testing.T) {
	t.Parallel()
	cols := ToColumns([]TypedColumn[testItem]{})
	if len(cols) != 0 {
		t.Fatalf("expected 0 columns, got %d", len(cols))
	}
}

func TestToColumnsWorksWithFormatList(t *testing.T) {
	t.Parallel()
	typed := []TypedColumn[testItem]{
		{Header: "Name", Extract: func(v testItem) string { return v.Name }},
		{Header: "Email", Extract: func(v testItem) string { return v.Email }},
	}

	cols := ToColumns(typed)
	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
	}

	out, err := FormatList(&Config{Format: "json"}, cols, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}
