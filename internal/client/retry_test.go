package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestRetryTransport(handler http.Handler, stderr io.Writer, isJSON bool) (*retryTransport, *httptest.Server) {
	server := httptest.NewServer(handler)
	rt := &retryTransport{
		base:    server.Client().Transport,
		stderr:  stderr,
		sleepFn: func(time.Duration) {}, // no-op sleep for tests
		isJSON:  func() bool { return isJSON },
	}
	return rt, server
}

func TestRetryTransport_NoRetryOn200(t *testing.T) {
	t.Parallel()
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})

	rt, server := newTestRetryTransport(handler, io.Discard, false)
	defer server.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryTransport_RetriesOn429(t *testing.T) {
	t.Parallel()
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	var stderr bytes.Buffer
	rt, server := newTestRetryTransport(handler, &stderr, true)
	defer server.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after retries, got %d", resp.StatusCode)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls (1 original + 2 retries), got %d", calls)
	}
	if !strings.Contains(stderr.String(), "Rate limited") {
		t.Errorf("expected retry message on stderr, got: %q", stderr.String())
	}
}

func TestRetryTransport_MaxRetriesExhausted(t *testing.T) {
	t.Parallel()
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusTooManyRequests)
	})

	var stderr bytes.Buffer
	rt, server := newTestRetryTransport(handler, &stderr, true)
	defer server.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 after max retries, got %d", resp.StatusCode)
	}
	// 1 original + 3 retries = 4 total
	if calls != 4 {
		t.Errorf("expected 4 calls (1 original + 3 retries), got %d", calls)
	}

	output := stderr.String()
	if strings.Count(output, "attempt 1/3") != 1 {
		t.Errorf("expected exactly one attempt 1/3 message, got: %q", output)
	}
	if strings.Count(output, "attempt 2/3") != 1 {
		t.Errorf("expected exactly one attempt 2/3 message, got: %q", output)
	}
	if strings.Count(output, "attempt 3/3") != 1 {
		t.Errorf("expected exactly one attempt 3/3 message, got: %q", output)
	}
}

func TestRetryTransport_RespectsRetryAfterHeader(t *testing.T) {
	t.Parallel()
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	var stderr bytes.Buffer
	rt, server := newTestRetryTransport(handler, &stderr, true)
	defer server.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(stderr.String(), "retrying in 5s") {
		t.Errorf("expected Retry-After of 5s in message, got: %q", stderr.String())
	}
}

func TestRetryTransport_ExponentialBackoffDelays(t *testing.T) {
	t.Parallel()
	var delays []time.Duration
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusTooManyRequests)
	})

	rt, server := newTestRetryTransport(handler, io.Discard, true)
	rt.sleepFn = func(d time.Duration) {
		delays = append(delays, d)
	}
	defer server.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()

	expected := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	if len(delays) != len(expected) {
		t.Fatalf("expected %d delays, got %d: %v", len(expected), len(delays), delays)
	}
	for i, exp := range expected {
		if delays[i] != exp {
			t.Errorf("delay[%d] = %v, want %v", i, delays[i], exp)
		}
	}
}

func TestRetryTransport_RetryAfterCappedAt30s(t *testing.T) {
	t.Parallel()
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "120")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	var delays []time.Duration
	rt, server := newTestRetryTransport(handler, io.Discard, true)
	rt.sleepFn = func(d time.Duration) {
		delays = append(delays, d)
	}
	defer server.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()

	if len(delays) != 1 {
		t.Fatalf("expected 1 delay, got %d", len(delays))
	}
	if delays[0] != 30*time.Second {
		t.Errorf("expected delay capped at 30s, got %v", delays[0])
	}
}

func TestRetryTransport_JSONOutputNoANSI(t *testing.T) {
	t.Parallel()
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	var stderr bytes.Buffer
	rt, server := newTestRetryTransport(handler, &stderr, true)
	defer server.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()

	output := stderr.String()
	if strings.Contains(output, "\033") {
		t.Errorf("JSON mode should not contain ANSI escape sequences, got: %q", output)
	}
	if !strings.Contains(output, "Rate limited") {
		t.Errorf("expected rate limit message, got: %q", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("JSON mode message should end with newline, got: %q", output)
	}
}

func TestRetryTransport_NonJSONOutputUsesANSI(t *testing.T) {
	t.Parallel()
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	var stderr bytes.Buffer
	rt, server := newTestRetryTransport(handler, &stderr, false)
	defer server.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()

	output := stderr.String()
	if !strings.Contains(output, "\033[K") {
		t.Errorf("non-JSON mode should use ANSI escape sequences, got: %q", output)
	}
}

func TestRetryTransport_RetryWithRequestBody(t *testing.T) {
	t.Parallel()
	calls := 0
	var bodies []string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	rt, server := newTestRetryTransport(handler, io.Discard, true)
	defer server.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "POST", server.URL, strings.NewReader(`{"code":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()

	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
	if len(bodies) != 2 {
		t.Fatalf("expected 2 bodies, got %d", len(bodies))
	}
	for i, b := range bodies {
		if b != `{"code":"test"}` {
			t.Errorf("body[%d] = %q, want %q", i, b, `{"code":"test"}`)
		}
	}
}

func TestRetryTransport_NoRetryOnOtherErrors(t *testing.T) {
	t.Parallel()
	for _, status := range []int{400, 401, 403, 404, 500, 502, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()
			calls := 0
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(status)
			})

			rt, server := newTestRetryTransport(handler, io.Discard, true)
			defer server.Close()

			req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
			resp, err := rt.RoundTrip(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			_ = resp.Body.Close()

			if calls != 1 {
				t.Errorf("expected 1 call for status %d, got %d", status, calls)
			}
		})
	}
}

func TestRetryDelay_ExponentialBackoff(t *testing.T) {
	t.Parallel()
	resp := &http.Response{Header: http.Header{}}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{10, 30 * time.Second}, // capped at max
	}

	for _, tt := range tests {
		got := retryDelay(resp, tt.attempt)
		if got != tt.expected {
			t.Errorf("retryDelay(attempt=%d) = %v, want %v", tt.attempt, got, tt.expected)
		}
	}
}

func TestRetryDelay_RetryAfterHeader(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		header   string
		expected time.Duration
		attempt  int
	}{
		{"valid seconds", "3", 3 * time.Second, 0},
		{"large value capped", "60", 30 * time.Second, 0},
		{"invalid falls back to backoff", "invalid", 1 * time.Second, 0},
		{"zero falls back to backoff", "0", 1 * time.Second, 0},
		{"negative falls back to backoff", "-1", 1 * time.Second, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := &http.Response{Header: http.Header{}}
			resp.Header.Set("Retry-After", tt.header)

			got := retryDelay(resp, tt.attempt)
			if got != tt.expected {
				t.Errorf("retryDelay(Retry-After=%q, attempt=%d) = %v, want %v",
					tt.header, tt.attempt, got, tt.expected)
			}
		})
	}
}

func TestRetryTransport_StderrOutput(t *testing.T) {
	t.Parallel()
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	var stderr bytes.Buffer
	rt, server := newTestRetryTransport(handler, &stderr, true)
	defer server.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()

	output := stderr.String()
	if !strings.Contains(output, "attempt 1/3") {
		t.Errorf("missing attempt 1/3 in stderr: %q", output)
	}
	if !strings.Contains(output, "attempt 2/3") {
		t.Errorf("missing attempt 2/3 in stderr: %q", output)
	}
}
