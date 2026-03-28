package output

import (
	"encoding/json"
	"strings"
	"testing"
)

type testItem struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

var testColumns = []Column{
	{Header: "Name", Extract: func(v any) string { return v.(testItem).Name }},
	{Header: "Email", Extract: func(v any) string { return v.(testItem).Email }},
}

func TestValidateFormat(t *testing.T) {
	for _, f := range []string{"table", "json", "json-pretty"} {
		if err := ValidateFormat(f); err != nil {
			t.Errorf("ValidateFormat(%q) returned error: %v", f, err)
		}
	}

	err := ValidateFormat("xml")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("error should mention the invalid value, got: %v", err)
	}
	if !strings.Contains(err.Error(), "table, json, json-pretty") {
		t.Errorf("error should list valid options, got: %v", err)
	}
}

func TestFormatListJSON(t *testing.T) {
	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}

	out, err := FormatList("json", testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be compact single-line JSON
	if strings.Contains(out, "\n") {
		t.Error("json format should be single-line")
	}

	var envelope struct {
		Object  string     `json:"object"`
		HasMore bool       `json:"has_more"`
		Data    []testItem `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if envelope.Object != "list" {
		t.Errorf("expected object=list, got %s", envelope.Object)
	}
	if envelope.HasMore {
		t.Error("expected has_more=false")
	}
	if len(envelope.Data) != 2 || envelope.Data[0].Name != "Alice" {
		t.Errorf("unexpected decoded data: %v", envelope.Data)
	}
}

func TestFormatListJSONHasMore(t *testing.T) {
	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
	}

	out, err := FormatList("json", testColumns, items, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var envelope struct {
		Object  string `json:"object"`
		HasMore bool   `json:"has_more"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if !envelope.HasMore {
		t.Error("expected has_more=true")
	}
}

func TestFormatListJSONPretty(t *testing.T) {
	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
	}

	out, err := FormatList("json-pretty", testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be indented with 2 spaces
	if !strings.Contains(out, "  ") {
		t.Error("json-pretty format should use 2-space indent")
	}
	if !strings.Contains(out, "\n") {
		t.Error("json-pretty format should be multi-line")
	}

	// Should have envelope structure
	var envelope struct {
		Object  string     `json:"object"`
		HasMore bool       `json:"has_more"`
		Data    []testItem `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if envelope.Object != "list" {
		t.Errorf("expected object=list, got %s", envelope.Object)
	}
}

func TestFormatListTable(t *testing.T) {
	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}

	out, err := FormatList("table", testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Table should contain headers and data
	if !strings.Contains(out, "Name") {
		t.Error("table should contain header 'Name'")
	}
	if !strings.Contains(out, "Email") {
		t.Error("table should contain header 'Email'")
	}
	if !strings.Contains(out, "Alice") {
		t.Error("table should contain data 'Alice'")
	}
	if !strings.Contains(out, "bob@example.com") {
		t.Error("table should contain data 'bob@example.com'")
	}
}

func TestFormatOneJSON(t *testing.T) {
	item := testItem{Name: "Alice", Email: "alice@example.com"}

	out, err := FormatOne("json", testColumns, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, "\n") {
		t.Error("json format should be single-line")
	}

	var decoded testItem
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if decoded.Name != "Alice" {
		t.Errorf("expected Name=Alice, got %s", decoded.Name)
	}
}

func TestFormatOneJSONPretty(t *testing.T) {
	item := testItem{Name: "Alice", Email: "alice@example.com"}

	out, err := FormatOne("json-pretty", testColumns, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "  ") {
		t.Error("json-pretty format should use 2-space indent")
	}
}

func TestFormatOneTable(t *testing.T) {
	item := testItem{Name: "Alice", Email: "alice@example.com"}

	out, err := FormatOne("table", testColumns, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Key-value table should have Field/Value headers
	if !strings.Contains(out, "Field") {
		t.Error("single-resource table should contain 'Field' header")
	}
	if !strings.Contains(out, "Value") {
		t.Error("single-resource table should contain 'Value' header")
	}
	if !strings.Contains(out, "Name") {
		t.Error("table should contain label 'Name'")
	}
	if !strings.Contains(out, "Alice") {
		t.Error("table should contain value 'Alice'")
	}
}

func TestFormatListInvalidFormat(t *testing.T) {
	_, err := FormatList("xml", testColumns, nil, false)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestFormatOneInvalidFormat(t *testing.T) {
	_, err := FormatOne("yaml", testColumns, testItem{})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}
