package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2/table"
)

var validFormats = []string{"table", "json", "json-pretty"}

// Column defines a table column with a header label and a function to extract
// the cell value from a data item.
type Column struct {
	Header  string
	Extract func(any) string
}

// ValidateFormat returns an error if format is not one of the accepted values.
func ValidateFormat(format string) error {
	for _, v := range validFormats {
		if format == v {
			return nil
		}
	}
	return fmt.Errorf("invalid output format %q: valid options are %s", format, strings.Join(validFormats, ", "))
}

// listEnvelope wraps list data in a Recurly-style JSON envelope.
type listEnvelope struct {
	Object  string `json:"object"`
	HasMore bool   `json:"has_more"`
	Data    []any  `json:"data"`
}

// FormatList formats a slice of items as a columnar table (with headers),
// compact JSON, or indented JSON depending on format. JSON output is wrapped
// in a Recurly-style list envelope with pagination metadata.
func FormatList(format string, columns []Column, items []any, hasMore bool) (string, error) {
	if err := ValidateFormat(format); err != nil {
		return "", err
	}

	if HasJQ() {
		data := items
		if data == nil {
			data = []any{}
		}
		return applyJQ(data)
	}

	switch format {
	case "json", "json-pretty":
		data := items
		if data == nil {
			data = []any{}
		}
		envelope := listEnvelope{
			Object:  "list",
			HasMore: hasMore,
			Data:    data,
		}
		if format == "json-pretty" {
			return marshalJSONPretty(envelope)
		}
		return marshalJSON(envelope)
	default:
		return renderListTable(columns, items), nil
	}
}

// FormatOne formats a single item as a key-value table (label on left, value
// on right), compact JSON, or indented JSON depending on format.
func FormatOne(format string, columns []Column, item any) (string, error) {
	if err := ValidateFormat(format); err != nil {
		return "", err
	}

	if HasJQ() {
		return applyJQ(item)
	}

	switch format {
	case "json":
		return marshalJSON(item)
	case "json-pretty":
		return marshalJSONPretty(item)
	default:
		return renderOneTable(columns, item), nil
	}
}

func marshalJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encoding JSON: %w", err)
	}
	return string(b), nil
}

func marshalJSONPretty(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encoding JSON: %w", err)
	}
	return string(b), nil
}

func renderListTable(columns []Column, items []any) string {
	headers := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = c.Header
	}

	t := table.New().Headers(headers...)

	for _, item := range items {
		row := make([]string, len(columns))
		for i, c := range columns {
			row[i] = c.Extract(item)
		}
		t.Row(row...)
	}

	return t.String()
}

func renderOneTable(columns []Column, item any) string {
	t := table.New().Headers("Field", "Value")

	for _, c := range columns {
		t.Row(c.Header, c.Extract(item))
	}

	return t.String()
}
