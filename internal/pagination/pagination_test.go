package pagination

import (
	"fmt"
	"testing"
)

// mockLister simulates a paginated SDK lister with configurable pages.
type mockLister struct {
	pages   [][]string
	current int
	fetched bool
}

func newMockLister(pages [][]string) *mockLister {
	return &mockLister{pages: pages, current: -1}
}

func (m *mockLister) Fetch() error {
	m.current++
	if m.current >= len(m.pages) {
		return fmt.Errorf("no more pages")
	}
	m.fetched = true
	return nil
}

func (m *mockLister) Data() []string {
	if m.current < 0 || m.current >= len(m.pages) {
		return nil
	}
	return m.pages[m.current]
}

func (m *mockLister) HasMore() bool {
	return m.current < len(m.pages)-1
}

// errorLister returns an error on Fetch.
type errorLister struct{}

func (e *errorLister) Fetch() error   { return fmt.Errorf("api error") }
func (e *errorLister) Data() []string { return nil }
func (e *errorLister) HasMore() bool  { return true }

func TestCollect_AllPages(t *testing.T) {
	lister := newMockLister([][]string{
		{"a", "b", "c"},
		{"d", "e"},
		{"f"},
	})

	result, err := Collect[string](lister, 0, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 6 {
		t.Fatalf("expected 6 results, got %d", len(result.Items))
	}
	expected := []string{"a", "b", "c", "d", "e", "f"}
	for i, v := range result.Items {
		if v != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, v, expected[i])
		}
	}
	if result.HasMore {
		t.Error("expected HasMore=false when all=true")
	}
}

func TestCollect_LimitWithinFirstPage(t *testing.T) {
	lister := newMockLister([][]string{
		{"a", "b", "c", "d", "e"},
		{"f", "g"},
	})

	result, err := Collect[string](lister, 3, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result.Items))
	}
	expected := []string{"a", "b", "c"}
	for i, v := range result.Items {
		if v != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, v, expected[i])
		}
	}
	if !result.HasMore {
		t.Error("expected HasMore=true when results were truncated")
	}
}

func TestCollect_LimitAcrossPages(t *testing.T) {
	lister := newMockLister([][]string{
		{"a", "b"},
		{"c", "d"},
		{"e", "f"},
	})

	result, err := Collect[string](lister, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 5 {
		t.Fatalf("expected 5 results, got %d", len(result.Items))
	}
	expected := []string{"a", "b", "c", "d", "e"}
	for i, v := range result.Items {
		if v != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, v, expected[i])
		}
	}
	if !result.HasMore {
		t.Error("expected HasMore=true when results were truncated")
	}
}

func TestCollect_DefaultLimit(t *testing.T) {
	// 30 items across pages, default limit should return 20
	page := make([]string, 30)
	for i := range page {
		page[i] = fmt.Sprintf("item-%d", i)
	}
	lister := newMockLister([][]string{page})

	result, err := Collect[string](lister, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 20 {
		t.Fatalf("expected 20 results (default limit), got %d", len(result.Items))
	}
	if !result.HasMore {
		t.Error("expected HasMore=true when default limit truncates results")
	}
}

func TestCollect_LimitExceedsAvailable(t *testing.T) {
	lister := newMockLister([][]string{
		{"a", "b"},
	})

	result, err := Collect[string](lister, 10, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Items))
	}
	if result.HasMore {
		t.Error("expected HasMore=false when all items fit within limit")
	}
}

func TestCollect_EmptyLister(t *testing.T) {
	lister := newMockLister([][]string{{}})

	result, err := Collect[string](lister, 0, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result.Items))
	}
	if result.HasMore {
		t.Error("expected HasMore=false for empty lister")
	}
}

func TestCollect_FetchError(t *testing.T) {
	lister := &errorLister{}

	result, err := Collect[string](lister, 10, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "api error" {
		t.Errorf("expected 'api error', got %q", err.Error())
	}
	if result.Items != nil {
		t.Errorf("expected nil results on error, got %v", result.Items)
	}
}

func TestCollect_AllIgnoresLimit(t *testing.T) {
	lister := newMockLister([][]string{
		{"a", "b", "c"},
		{"d", "e"},
	})

	// When all=true, limit should be ignored
	result, err := Collect[string](lister, 2, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 5 {
		t.Fatalf("expected 5 results (all=true ignores limit), got %d", len(result.Items))
	}
	if result.HasMore {
		t.Error("expected HasMore=false when all=true")
	}
}

func TestCollect_SinglePage(t *testing.T) {
	lister := newMockLister([][]string{
		{"only"},
	})

	result, err := Collect[string](lister, 10, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Items))
	}
	if result.Items[0] != "only" {
		t.Errorf("expected 'only', got %q", result.Items[0])
	}
	if result.HasMore {
		t.Error("expected HasMore=false for single page within limit")
	}
}
