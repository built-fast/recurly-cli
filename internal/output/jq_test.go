package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/itchyny/gojq"
)

func compileJQ(t *testing.T, expr string) *gojq.Code {
	t.Helper()
	query, err := gojq.Parse(expr)
	if err != nil {
		t.Fatalf("failed to parse jq expression %q: %v", expr, err)
	}
	code, err := gojq.Compile(query)
	if err != nil {
		t.Fatalf("failed to compile jq expression %q: %v", expr, err)
	}
	return code
}

func TestApplyJQ_FieldAccess(t *testing.T) {
	code := compileJQ(t, ".email")

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %q", out)
	}
}

func TestApplyJQ_ArrayIteration(t *testing.T) {
	code := compileJQ(t, ".[].name")

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(code, items, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 2 || lines[0] != "Alice" || lines[1] != "Bob" {
		t.Errorf("expected Alice\\nBob, got %q", out)
	}
}

func TestApplyJQ_SelectFilter(t *testing.T) {
	code := compileJQ(t, `[.[] | select(.name == "Bob")] | length`)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(code, items, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "1" {
		t.Errorf("expected 1, got %q", out)
	}
}

func TestApplyJQ_Length(t *testing.T) {
	code := compileJQ(t, "length")

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(code, items, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "2" {
		t.Errorf("expected 2, got %q", out)
	}
}

func TestApplyJQ_PipeChain(t *testing.T) {
	code := compileJQ(t, `[.[] | .name] | sort`)

	items := []any{
		testItem{Name: "Bob", Email: "bob@example.com"},
		testItem{Name: "Alice", Email: "alice@example.com"},
	}
	out, err := applyJQ(code, items, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Errorf("expected sorted array with Alice and Bob, got %q", out)
	}
}

func TestApplyJQ_Base64(t *testing.T) {
	code := compileJQ(t, ".name | @base64")

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "QWxpY2U=" {
		t.Errorf("expected QWxpY2U=, got %q", out)
	}
}

func TestApplyJQ_CSV(t *testing.T) {
	code := compileJQ(t, `[.name, .email] | @csv`)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "alice@example.com") {
		t.Errorf("expected CSV with Alice and email, got %q", out)
	}
}

func TestApplyJQ_NullResult(t *testing.T) {
	code := compileJQ(t, ".nonexistent")

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "null" {
		t.Errorf("expected null, got %q", out)
	}
}

func TestApplyJQ_BoolResult_True(t *testing.T) {
	code := compileJQ(t, ".has_more")

	input := map[string]any{"has_more": true}
	out, err := applyJQ(code, input, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "true" {
		t.Errorf("expected true, got %q", out)
	}
}

func TestApplyJQ_BoolResult_False(t *testing.T) {
	code := compileJQ(t, ".has_more")

	input := map[string]any{"has_more": false}
	out, err := applyJQ(code, input, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "false" {
		t.Errorf("expected false, got %q", out)
	}
}

func TestApplyJQ_NumericResult_Int(t *testing.T) {
	code := compileJQ(t, ".data | length")

	input := map[string]any{"data": []any{1, 2, 3}}
	out, err := applyJQ(code, input, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "3" {
		t.Errorf("expected 3, got %q", out)
	}
}

func TestApplyJQ_NumericResult_Float(t *testing.T) {
	code := compileJQ(t, ".price")

	input := map[string]any{"price": 1.5}
	out, err := applyJQ(code, input, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "1.5" {
		t.Errorf("expected 1.5, got %q", out)
	}
}

func TestApplyJQ_ObjectResult_Compact(t *testing.T) {
	code := compileJQ(t, `{name: .name}`)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Compact JSON — no newlines
	if strings.Contains(out, "\n") {
		t.Errorf("expected compact JSON (no newlines), got %q", out)
	}
	var decoded map[string]string
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded["name"] != "Alice" {
		t.Errorf("expected name=Alice, got %q", decoded["name"])
	}
}

func TestApplyJQ_ObjectResult_Pretty(t *testing.T) {
	code := compileJQ(t, `{name: .name}`)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json-pretty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Pretty JSON — has newlines and indentation
	if !strings.Contains(out, "\n") {
		t.Errorf("expected indented JSON (multi-line), got %q", out)
	}
	if !strings.Contains(out, "  ") {
		t.Errorf("expected 2-space indentation, got %q", out)
	}
}

func TestApplyJQ_ArrayResult_Compact(t *testing.T) {
	code := compileJQ(t, `[.[] | .name]`)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(code, items, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "\n") {
		t.Errorf("expected compact JSON array, got %q", out)
	}
	if out != `["Alice","Bob"]` {
		t.Errorf("expected [\"Alice\",\"Bob\"], got %q", out)
	}
}

func TestApplyJQ_ArrayResult_Pretty(t *testing.T) {
	code := compileJQ(t, `[.[] | .name]`)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(code, items, "json-pretty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "\n") {
		t.Errorf("expected indented JSON array, got %q", out)
	}
}

func TestApplyJQ_EmptyResult(t *testing.T) {
	code := compileJQ(t, `select(.name == "Nobody")`)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty output, got %q", out)
	}
}

func TestApplyJQ_RuntimeError(t *testing.T) {
	code := compileJQ(t, ".name / 0")

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	_, err := applyJQ(code, item, "json")
	if err == nil {
		t.Fatal("expected runtime error")
	}
	if !strings.Contains(err.Error(), "jq:") {
		t.Errorf("expected error prefixed with jq:, got %q", err.Error())
	}
}

// --- FormatList with jq ---

func TestFormatList_WithJQ_SelectFilter(t *testing.T) {
	code := compileJQ(t, `.data | map(select(.state == "active"))`)

	items := []any{
		map[string]any{"code": "a1", "state": "active"},
		map[string]any{"code": "b2", "state": "expired"},
		map[string]any{"code": "c3", "state": "active"},
	}
	out, err := FormatList(&Config{Format: "json", JQ: code}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result []map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 active items, got %d", len(result))
	}
	for _, item := range result {
		if item["state"] != "active" {
			t.Errorf("expected state=active, got %v", item["state"])
		}
	}
}

func TestFormatList_WithJQ_DataLength(t *testing.T) {
	code := compileJQ(t, `.data | length`)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
		testItem{Name: "Carol", Email: "carol@example.com"},
	}
	out, err := FormatList(&Config{Format: "json", JQ: code}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "3" {
		t.Errorf("expected 3, got %q", out)
	}
}

func TestFormatList_WithJQ_PipeChainProjection(t *testing.T) {
	code := compileJQ(t, `.data[] | {name, email}`)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := FormatList(&Config{Format: "json", JQ: code}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 projected objects, got %d", len(lines))
	}
	for _, line := range lines {
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("invalid JSON line %q: %v", line, err)
		}
		if _, ok := obj["name"]; !ok {
			t.Errorf("expected projected object to have 'name', got %v", obj)
		}
		if _, ok := obj["email"]; !ok {
			t.Errorf("expected projected object to have 'email', got %v", obj)
		}
	}
}

func TestFormatList_WithJQ_EnvelopeInput(t *testing.T) {
	// jq input for FormatList is the full envelope, so .data[].name works
	code := compileJQ(t, ".data[].name")

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := FormatList(&Config{Format: "json", JQ: code}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 2 || lines[0] != "Alice" || lines[1] != "Bob" {
		t.Errorf("expected Alice\\nBob, got %q", out)
	}
}

func TestFormatList_WithJQ_EnvelopeFields(t *testing.T) {
	// Verify the envelope object field is accessible
	code := compileJQ(t, ".object")

	items := []any{testItem{Name: "Alice", Email: "alice@example.com"}}
	out, err := FormatList(&Config{Format: "json", JQ: code}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "list" {
		t.Errorf("expected 'list', got %q", out)
	}
}

func TestFormatList_WithJQ_HasMore(t *testing.T) {
	code := compileJQ(t, ".has_more")

	items := []any{testItem{Name: "Alice", Email: "alice@example.com"}}
	out, err := FormatList(&Config{Format: "json", JQ: code}, testColumns, items, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "true" {
		t.Errorf("expected 'true', got %q", out)
	}
}

func TestFormatList_WithJQ_NilItems(t *testing.T) {
	code := compileJQ(t, ".data | length")

	out, err := FormatList(&Config{Format: "json", JQ: code}, testColumns, nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "0" {
		t.Errorf("expected 0 for nil items, got %q", out)
	}
}

func TestFormatList_WithJQ_Pretty(t *testing.T) {
	code := compileJQ(t, `.data[0] | {name: .name}`)

	items := []any{testItem{Name: "Alice", Email: "alice@example.com"}}
	out, err := FormatList(&Config{Format: "json-pretty", JQ: code}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "\n") {
		t.Errorf("expected indented JSON with json-pretty, got %q", out)
	}
}

// --- FormatOne with jq ---

func TestFormatOne_WithJQ(t *testing.T) {
	code := compileJQ(t, ".email")

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := FormatOne(&Config{Format: "json", JQ: code}, testColumns, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %q", out)
	}
}

func TestFormatOne_WithJQ_Pretty(t *testing.T) {
	code := compileJQ(t, `{name: .name}`)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := FormatOne(&Config{Format: "json-pretty", JQ: code}, testColumns, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "\n") {
		t.Errorf("expected indented JSON with json-pretty, got %q", out)
	}
}

func TestSetJQ_HasJQ(t *testing.T) {
	SetJQ(nil)
	if HasJQ() {
		t.Error("expected HasJQ() == false after SetJQ(nil)")
	}

	SetJQ(compileJQ(t, "."))
	defer SetJQ(nil)
	if !HasJQ() {
		t.Error("expected HasJQ() == true after SetJQ(code)")
	}
}
