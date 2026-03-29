package client

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	maxRetries     = 3
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
)

// retryTransport wraps an http.RoundTripper and automatically retries
// requests that receive HTTP 429 (Too Many Requests) responses.
// It uses the Retry-After header if present, otherwise exponential backoff.
type retryTransport struct {
	base    http.RoundTripper
	stderr  io.Writer
	sleepFn func(time.Duration)
	isJSON  func() bool
}

// newRetryTransport creates a retryTransport wrapping the given base transport.
// stderr is the writer for retry status messages. isJSON controls whether
// ANSI escape sequences are suppressed in output.
func newRetryTransport(base http.RoundTripper, stderr io.Writer, isJSON func() bool) *retryTransport {
	return &retryTransport{
		base:    base,
		stderr:  stderr,
		sleepFn: time.Sleep,
		isJSON:  isJSON,
	}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		// Reset request body for retries
		if attempt > 0 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			req.Body = body
		}

		resp, err := t.base.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			if attempt > 0 {
				t.clearLine()
			}
			return resp, nil
		}

		// Exhausted all retries — return the 429 for the SDK to handle
		if attempt >= maxRetries {
			if attempt > 0 {
				t.clearLine()
			}
			return resp, nil
		}

		// Close the 429 response body before retrying
		_ = resp.Body.Close()

		delay := retryDelay(resp, attempt)
		t.printRetryMessage(delay, attempt)
		t.sleep(delay)
	}
}

// retryDelay computes the wait duration before the next retry.
// It uses the Retry-After header if present and valid, otherwise
// exponential backoff starting at 1s (1s, 2s, 4s) capped at 30s.
func retryDelay(resp *http.Response, attempt int) time.Duration {
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
			delay := time.Duration(seconds) * time.Second
			if delay > maxBackoff {
				delay = maxBackoff
			}
			return delay
		}
	}

	delay := time.Duration(math.Pow(2, float64(attempt))) * initialBackoff
	if delay > maxBackoff {
		delay = maxBackoff
	}
	return delay
}

func (t *retryTransport) printRetryMessage(delay time.Duration, attempt int) {
	w := t.stderr
	if w == nil {
		w = os.Stderr
	}

	seconds := int(delay.Seconds())
	if t.isJSON != nil && t.isJSON() {
		_, _ = fmt.Fprintf(w, "Rate limited, retrying in %ds... (attempt %d/%d)\n", seconds, attempt+1, maxRetries)
	} else {
		_, _ = fmt.Fprintf(w, "\r\033[KRate limited, retrying in %ds... (attempt %d/%d)", seconds, attempt+1, maxRetries)
	}
}

func (t *retryTransport) clearLine() {
	if t.isJSON != nil && t.isJSON() {
		return
	}
	w := t.stderr
	if w == nil {
		w = os.Stderr
	}
	_, _ = fmt.Fprintf(w, "\r\033[K")
}

func (t *retryTransport) sleep(d time.Duration) {
	if t.sleepFn != nil {
		t.sleepFn(d)
	} else {
		time.Sleep(d)
	}
}
