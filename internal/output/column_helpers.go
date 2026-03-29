package output

import (
	"fmt"
	"strconv"
	"time"
)

// StringColumn creates a TypedColumn that extracts a string value directly.
func StringColumn[T any](header string, fn func(T) string) TypedColumn[T] {
	return TypedColumn[T]{
		Header:  header,
		Extract: fn,
	}
}

// TimeColumn creates a TypedColumn that formats a *time.Time as RFC3339,
// or returns an empty string when the pointer is nil.
func TimeColumn[T any](header string, fn func(T) *time.Time) TypedColumn[T] {
	return TypedColumn[T]{
		Header: header,
		Extract: func(v T) string {
			t := fn(v)
			if t != nil {
				return t.Format(time.RFC3339)
			}
			return ""
		},
	}
}

// BoolColumn creates a TypedColumn that formats a bool as "true" or "false".
func BoolColumn[T any](header string, fn func(T) bool) TypedColumn[T] {
	return TypedColumn[T]{
		Header: header,
		Extract: func(v T) string {
			return fmt.Sprintf("%t", fn(v))
		},
	}
}

// IntColumn creates a TypedColumn that formats an int via strconv.Itoa.
func IntColumn[T any](header string, fn func(T) int) TypedColumn[T] {
	return TypedColumn[T]{
		Header: header,
		Extract: func(v T) string {
			return strconv.Itoa(fn(v))
		},
	}
}

// FloatColumn creates a TypedColumn that formats a float64 with two decimal
// places, matching the currency formatting convention used throughout the CLI.
func FloatColumn[T any](header string, fn func(T) float64) TypedColumn[T] {
	return TypedColumn[T]{
		Header: header,
		Extract: func(v T) string {
			return strconv.FormatFloat(fn(v), 'f', 2, 64)
		},
	}
}
