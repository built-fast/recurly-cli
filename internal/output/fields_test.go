package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/itchyny/gojq"
)

func TestNormalizeFieldName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Code", "code"},
		{"code", "code"},
		{"CODE", "code"},
		{"First Name", "first_name"},
		{"first_name", "first_name"},
		{"first-name", "first_name"},
		{"Created At", "created_at"},
		{" Email ", "email"},
	}
	for _, tt := range tests {
		got := normalizeFieldName(tt.input)
		if got != tt.want {
			t.Errorf("normalizeFieldName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValidateFields_AllValid(t *testing.T) {
	err := ValidateFields(testColumns, []string{"Name", "Email"})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateFields_CaseInsensitive(t *testing.T) {
	err := ValidateFields(testColumns, []string{"name", "EMAIL"})
	if err != nil {
		t.Errorf("expected no error for case-insensitive match, got: %v", err)
	}
}

func TestValidateFields_Invalid(t *testing.T) {
	err := ValidateFields(testColumns, []string{"Name", "bogus"})
	if err == nil {
		t.Fatal("expected error for invalid field")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error should mention invalid field name, got: %v", err)
	}
	if !strings.Contains(err.Error(), "available fields") {
		t.Errorf("error should list available fields, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Name") || !strings.Contains(err.Error(), "Email") {
		t.Errorf("error should list available field names, got: %v", err)
	}
}

func TestSelectColumns_OrderMatchesFieldOrder(t *testing.T) {
	cols := SelectColumns(testColumns, []string{"Email", "Name"})
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if cols[0].Header != "Email" {
		t.Errorf("expected first column to be Email, got %s", cols[0].Header)
	}
	if cols[1].Header != "Name" {
		t.Errorf("expected second column to be Name, got %s", cols[1].Header)
	}
}

func TestSelectColumns_CaseInsensitive(t *testing.T) {
	cols := SelectColumns(testColumns, []string{"email"})
	if len(cols) != 1 {
		t.Fatalf("expected 1 column, got %d", len(cols))
	}
	if cols[0].Header != "Email" {
		t.Errorf("expected Email column, got %s", cols[0].Header)
	}
}

// --- Integration tests via FormatList / FormatOne ---

func TestFieldSelection_ListTable(t *testing.T) {
	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}

	out, err := FormatList(&Config{Format: "table", Fields: []string{"Name"}}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Name") {
		t.Error("table should contain header 'Name'")
	}
	if !strings.Contains(out, "Alice") {
		t.Error("table should contain data 'Alice'")
	}
	if strings.Contains(out, "Email") {
		t.Error("table should NOT contain header 'Email' when not selected")
	}
	if strings.Contains(out, "alice@example.com") {
		t.Error("table should NOT contain email data when not selected")
	}
}

func TestFieldSelection_ListJSON(t *testing.T) {
	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
	}

	out, err := FormatList(&Config{Format: "json", Fields: []string{"Name"}}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(envelope.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(envelope.Data))
	}
	item := envelope.Data[0]
	if _, ok := item["name"]; !ok {
		t.Error("JSON should include 'name' field")
	}
	if _, ok := item["email"]; ok {
		t.Error("JSON should NOT include 'email' field when not selected")
	}
}

func TestFieldSelection_OneTable(t *testing.T) {
	item := testItem{Name: "Alice", Email: "alice@example.com"}

	out, err := FormatOne(&Config{Format: "table", Fields: []string{"Email"}}, testColumns, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Email") {
		t.Error("table should contain 'Email' row")
	}
	if !strings.Contains(out, "alice@example.com") {
		t.Error("table should contain email value")
	}
	// "Name" should not appear as a field label (but "Field" header row exists)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		// Check for a row that has "Name" as field label (not the header row)
		if strings.Contains(line, "Name") && !strings.Contains(line, "Field") {
			t.Error("table should NOT contain 'Name' row when not selected")
		}
	}
}

func TestFieldSelection_OneJSON(t *testing.T) {
	item := testItem{Name: "Alice", Email: "alice@example.com"}

	out, err := FormatOne(&Config{Format: "json", Fields: []string{"Email"}}, testColumns, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := decoded["email"]; !ok {
		t.Error("JSON should include 'email'")
	}
	if _, ok := decoded["name"]; ok {
		t.Error("JSON should NOT include 'name' when not selected")
	}
}

func TestFieldSelection_InvalidField_ReturnsError(t *testing.T) {
	items := []any{testItem{Name: "Alice", Email: "alice@example.com"}}

	_, err := FormatList(&Config{Format: "table", Fields: []string{"bogus"}}, testColumns, items, false)
	if err == nil {
		t.Fatal("expected error for invalid field")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error should mention 'bogus', got: %v", err)
	}
	if !strings.Contains(err.Error(), "available fields") {
		t.Errorf("error should list available fields, got: %v", err)
	}
}

func TestFieldSelection_InvalidField_FormatOne(t *testing.T) {
	_, err := FormatOne(&Config{Format: "json", Fields: []string{"nonexistent"}}, testColumns, testItem{Name: "Alice"})
	if err == nil {
		t.Fatal("expected error for invalid field")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention 'nonexistent', got: %v", err)
	}
}

func TestFieldSelection_WithJQ(t *testing.T) {
	// Field selection applied first, then jq
	query, err := gojq.Parse(".data[].name")
	if err != nil {
		t.Fatalf("parse jq: %v", err)
	}
	code, err := gojq.Compile(query)
	if err != nil {
		t.Fatalf("compile jq: %v", err)
	}

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}

	out, fmtErr := FormatList(&Config{Format: "json", Fields: []string{"Name"}, JQ: code}, testColumns, items, false)
	if fmtErr != nil {
		t.Fatalf("unexpected error: %v", fmtErr)
	}

	// jq should extract names from field-filtered data
	if !strings.Contains(out, "Alice") {
		t.Error("output should contain 'Alice'")
	}
	if !strings.Contains(out, "Bob") {
		t.Error("output should contain 'Bob'")
	}
}

func TestFieldSelection_WithJQ_FilteredFieldNotAccessible(t *testing.T) {
	// Select only Name, then try to access email via jq — should get null
	query, err := gojq.Parse(".data[].email")
	if err != nil {
		t.Fatalf("parse jq: %v", err)
	}
	code, err := gojq.Compile(query)
	if err != nil {
		t.Fatalf("compile jq: %v", err)
	}

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
	}

	out, fmtErr := FormatList(&Config{Format: "json", Fields: []string{"Name"}, JQ: code}, testColumns, items, false)
	if fmtErr != nil {
		t.Fatalf("unexpected error: %v", fmtErr)
	}

	// email was filtered out, so jq should produce null
	if !strings.Contains(out, "null") {
		t.Errorf("expected null for filtered-out field, got: %s", out)
	}
	if strings.Contains(out, "alice@example.com") {
		t.Error("filtered-out email should not appear in output")
	}
}

func TestFieldSelection_CaseInsensitive_Integration(t *testing.T) {
	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
	}

	out, err := FormatList(&Config{Format: "json", Fields: []string{"name", "EMAIL"}}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(envelope.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(envelope.Data))
	}
	item := envelope.Data[0]
	if item["name"] != "Alice" {
		t.Errorf("expected name=Alice, got %v", item["name"])
	}
	if item["email"] != "alice@example.com" {
		t.Errorf("expected email=alice@example.com, got %v", item["email"])
	}
}

func TestFieldSelection_ColumnOrder_Table(t *testing.T) {
	// Verify that table columns appear in the order specified by --field
	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
	}

	out, err := FormatList(&Config{Format: "table", Fields: []string{"Email", "Name"}}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	emailIdx := strings.Index(out, "Email")
	nameIdx := strings.Index(out, "Name")
	if emailIdx < 0 || nameIdx < 0 {
		t.Fatalf("expected both Email and Name in output:\n%s", out)
	}
	if emailIdx > nameIdx {
		t.Error("Email should appear before Name when specified first in --field")
	}
}
