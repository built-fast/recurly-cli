package output

import (
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
	SetJQ(compileJQ(t, ".email"))
	defer SetJQ(nil)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %q", out)
	}
}

func TestApplyJQ_ArrayIteration(t *testing.T) {
	SetJQ(compileJQ(t, ".[].name"))
	defer SetJQ(nil)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 2 || lines[0] != "Alice" || lines[1] != "Bob" {
		t.Errorf("expected Alice\\nBob, got %q", out)
	}
}

func TestApplyJQ_SelectFilter(t *testing.T) {
	SetJQ(compileJQ(t, `[.[] | select(.name == "Bob")] | length`))
	defer SetJQ(nil)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "1" {
		t.Errorf("expected 1, got %q", out)
	}
}

func TestApplyJQ_Length(t *testing.T) {
	SetJQ(compileJQ(t, "length"))
	defer SetJQ(nil)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "2" {
		t.Errorf("expected 2, got %q", out)
	}
}

func TestApplyJQ_PipeChain(t *testing.T) {
	SetJQ(compileJQ(t, `[.[] | .name] | sort`))
	defer SetJQ(nil)

	items := []any{
		testItem{Name: "Bob", Email: "bob@example.com"},
		testItem{Name: "Alice", Email: "alice@example.com"},
	}
	out, err := applyJQ(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Errorf("expected sorted array with Alice and Bob, got %q", out)
	}
}

func TestApplyJQ_Base64(t *testing.T) {
	SetJQ(compileJQ(t, ".name | @base64"))
	defer SetJQ(nil)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "QWxpY2U=" {
		t.Errorf("expected QWxpY2U=, got %q", out)
	}
}

func TestApplyJQ_CSV(t *testing.T) {
	SetJQ(compileJQ(t, `[.name, .email] | @csv`))
	defer SetJQ(nil)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "alice@example.com") {
		t.Errorf("expected CSV with Alice and email, got %q", out)
	}
}

func TestApplyJQ_NullResult(t *testing.T) {
	SetJQ(compileJQ(t, ".nonexistent"))
	defer SetJQ(nil)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "null" {
		t.Errorf("expected null, got %q", out)
	}
}

func TestApplyJQ_ObjectResult(t *testing.T) {
	SetJQ(compileJQ(t, `{name: .name}`))
	defer SetJQ(nil)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"name"`) || !strings.Contains(out, `"Alice"`) {
		t.Errorf("expected JSON object with name Alice, got %q", out)
	}
}

func TestApplyJQ_RuntimeError(t *testing.T) {
	SetJQ(compileJQ(t, ".name / 0"))
	defer SetJQ(nil)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	_, err := applyJQ(item)
	if err == nil {
		t.Fatal("expected runtime error")
	}
	if !strings.Contains(err.Error(), "jq:") {
		t.Errorf("expected error prefixed with jq:, got %q", err.Error())
	}
}

func TestFormatList_WithJQ(t *testing.T) {
	SetJQ(compileJQ(t, ".[].name"))
	defer SetJQ(nil)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := FormatList("json", testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 2 || lines[0] != "Alice" || lines[1] != "Bob" {
		t.Errorf("expected Alice\\nBob, got %q", out)
	}
}

func TestFormatList_WithJQ_NilItems(t *testing.T) {
	SetJQ(compileJQ(t, "length"))
	defer SetJQ(nil)

	out, err := FormatList("json", testColumns, nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "0" {
		t.Errorf("expected 0 for nil items, got %q", out)
	}
}

func TestFormatOne_WithJQ(t *testing.T) {
	SetJQ(compileJQ(t, ".email"))
	defer SetJQ(nil)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := FormatOne("json", testColumns, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %q", out)
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
