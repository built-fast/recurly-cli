package client

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	recurly "github.com/recurly/recurly-client-go/v5"
)

// FormatError formats an error into a user-friendly message for stderr display.
// It handles Recurly SDK errors, network errors, and generic errors.
func FormatError(err error) string {
	var recurlyErr *recurly.Error
	if errors.As(err, &recurlyErr) {
		return formatRecurlyError(recurlyErr)
	}

	if isNetworkError(err) {
		return "Error: Unable to connect to Recurly API. Check your network connection."
	}

	return fmt.Sprintf("Error: %s", err.Error())
}

func formatRecurlyError(err *recurly.Error) string {
	switch err.Type {
	case recurly.ErrorTypeUnauthorized, recurly.ErrorTypeInvalidApiKey:
		return "Error: Invalid API key. Run 'recurly configure' to update your credentials."
	case recurly.ErrorTypeNotFound:
		return "Error: Resource not found."
	case recurly.ErrorTypeRateLimited, recurly.ErrorTypeTooManyRequests:
		return "Error: Rate limit exceeded. Try again in a moment."
	case recurly.ErrorTypeValidation:
		return formatValidationError(err)
	default:
		if err.Class == recurly.ErrorClassServer {
			return "Error: Recurly API server error. Please try again later."
		}
		return fmt.Sprintf("Error: %s", err.Message)
	}
}

func formatValidationError(err *recurly.Error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Error: %s", err.Message)
	for _, p := range err.Params {
		fmt.Fprintf(&b, "\n  - %s: %s", p.Property, p.Message)
	}
	return b.String()
}

func isNetworkError(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}
	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		return true
	}
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}
