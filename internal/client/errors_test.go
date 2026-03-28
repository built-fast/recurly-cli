package client

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"testing"

	recurly "github.com/recurly/recurly-client-go/v5"
)

func TestFormatError_Unauthorized(t *testing.T) {
	err := &recurly.Error{
		Message: "Invalid API key",
		Class:   recurly.ErrorClassClient,
		Type:    recurly.ErrorTypeUnauthorized,
	}
	got := FormatError(err)
	expected := "Error: Invalid API key. Run 'recurly configure' to update your credentials."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_InvalidApiKey(t *testing.T) {
	err := &recurly.Error{
		Message: "Invalid API key",
		Class:   recurly.ErrorClassClient,
		Type:    recurly.ErrorTypeInvalidApiKey,
	}
	got := FormatError(err)
	expected := "Error: Invalid API key. Run 'recurly configure' to update your credentials."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_NotFound(t *testing.T) {
	err := &recurly.Error{
		Message: "Couldn't find Account with id = abc123",
		Class:   recurly.ErrorClassClient,
		Type:    recurly.ErrorTypeNotFound,
	}
	got := FormatError(err)
	expected := "Error: Resource not found."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_RateLimited(t *testing.T) {
	err := &recurly.Error{
		Message: "Too many requests",
		Class:   recurly.ErrorClassClient,
		Type:    recurly.ErrorTypeRateLimited,
	}
	got := FormatError(err)
	expected := "Error: Rate limit exceeded. Try again in a moment."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_TooManyRequests(t *testing.T) {
	err := &recurly.Error{
		Message: "Too many requests",
		Class:   recurly.ErrorClassClient,
		Type:    recurly.ErrorTypeTooManyRequests,
	}
	got := FormatError(err)
	expected := "Error: Rate limit exceeded. Try again in a moment."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_ServerError(t *testing.T) {
	err := &recurly.Error{
		Message: "Internal server error",
		Class:   recurly.ErrorClassServer,
		Type:    recurly.ErrorTypeInternalServer,
	}
	got := FormatError(err)
	expected := "Error: Recurly API server error. Please try again later."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_ServerErrorVariants(t *testing.T) {
	serverTypes := []recurly.ErrorType{
		recurly.ErrorTypeBadGateway,
		recurly.ErrorTypeServiceUnavailable,
		recurly.ErrorTypeTimeout,
	}
	expected := "Error: Recurly API server error. Please try again later."
	for _, errType := range serverTypes {
		err := &recurly.Error{
			Message: "Server error",
			Class:   recurly.ErrorClassServer,
			Type:    errType,
		}
		got := FormatError(err)
		if got != expected {
			t.Errorf("type %q: expected %q, got %q", errType, expected, got)
		}
	}
}

func TestFormatError_Validation_NoParams(t *testing.T) {
	err := &recurly.Error{
		Message: "The record is invalid.",
		Class:   recurly.ErrorClassClient,
		Type:    recurly.ErrorTypeValidation,
	}
	got := FormatError(err)
	expected := "Error: The record is invalid."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_Validation_WithParams(t *testing.T) {
	err := &recurly.Error{
		Message: "The record is invalid.",
		Class:   recurly.ErrorClassClient,
		Type:    recurly.ErrorTypeValidation,
		Params: []recurly.ErrorParam{
			{Property: "email", Message: "is not valid"},
			{Property: "code", Message: "is already taken"},
		},
	}
	got := FormatError(err)
	if !strings.HasPrefix(got, "Error: The record is invalid.") {
		t.Errorf("expected error prefix, got %q", got)
	}
	if !strings.Contains(got, "\n  - email: is not valid") {
		t.Errorf("expected email field error, got %q", got)
	}
	if !strings.Contains(got, "\n  - code: is already taken") {
		t.Errorf("expected code field error, got %q", got)
	}
}

func TestFormatError_GenericClientError(t *testing.T) {
	err := &recurly.Error{
		Message: "Something went wrong with the request.",
		Class:   recurly.ErrorClassClient,
		Type:    recurly.ErrorTypeBadRequest,
	}
	got := FormatError(err)
	expected := "Error: Something went wrong with the request."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_NetworkError_URLError(t *testing.T) {
	err := &url.Error{
		Op:  "Get",
		URL: "https://v3.recurly.com/accounts",
		Err: &net.OpError{
			Op:  "dial",
			Net: "tcp",
			Err: fmt.Errorf("connection refused"),
		},
	}
	got := FormatError(err)
	expected := "Error: Unable to connect to Recurly API. Check your network connection."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_NetworkError_DNSError(t *testing.T) {
	err := &url.Error{
		Op:  "Get",
		URL: "https://v3.recurly.com/accounts",
		Err: &net.DNSError{
			Name: "v3.recurly.com",
			Err:  "no such host",
		},
	}
	got := FormatError(err)
	expected := "Error: Unable to connect to Recurly API. Check your network connection."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_GenericError(t *testing.T) {
	err := fmt.Errorf("API key not configured. Run 'recurly configure' or set RECURLY_API_KEY.")
	got := FormatError(err)
	expected := "Error: API key not configured. Run 'recurly configure' or set RECURLY_API_KEY."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
