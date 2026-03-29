package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// normalizeFieldName converts a field name to a canonical form for matching.
// "First Name" → "first_name", "first-name" → "first_name", "Code" → "code".
func normalizeFieldName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

// ValidateFields checks that all field names match available columns.
// Returns an error listing available fields if any are invalid.
func ValidateFields(columns []Column, fields []string) error {
	available := make(map[string]bool, len(columns))
	for _, c := range columns {
		available[normalizeFieldName(c.Header)] = true
	}

	var invalid []string
	for _, f := range fields {
		if !available[normalizeFieldName(f)] {
			invalid = append(invalid, f)
		}
	}

	if len(invalid) > 0 {
		names := make([]string, len(columns))
		for i, c := range columns {
			names[i] = c.Header
		}
		return fmt.Errorf("unknown field(s): %s\navailable fields: %s",
			strings.Join(invalid, ", "), strings.Join(names, ", "))
	}
	return nil
}

// SelectColumns filters and reorders columns to match the user-specified field
// order. The returned slice contains only columns whose headers match (case-
// insensitive) the given fields, in the order the fields were specified.
func SelectColumns(columns []Column, fields []string) []Column {
	fieldOrder := make(map[string]int, len(fields))
	for i, f := range fields {
		fieldOrder[normalizeFieldName(f)] = i
	}

	type indexed struct {
		order int
		col   Column
	}
	var matched []indexed
	for _, c := range columns {
		if order, ok := fieldOrder[normalizeFieldName(c.Header)]; ok {
			matched = append(matched, indexed{order, c})
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].order < matched[j].order
	})

	result := make([]Column, len(matched))
	for i, m := range matched {
		result[i] = m.col
	}
	return result
}

// filterOneJSON marshals an item to JSON, then filters it to only include keys
// whose normalized names match the selected columns' headers.
func filterOneJSON(item any, columns []Column) (any, error) {
	b, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("filtering fields: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("filtering fields: %w", err)
	}

	allowed := make(map[string]bool, len(columns))
	for _, c := range columns {
		allowed[normalizeFieldName(c.Header)] = true
	}

	filtered := make(map[string]any, len(allowed))
	for key, val := range raw {
		if allowed[normalizeFieldName(key)] {
			filtered[key] = val
		}
	}
	return filtered, nil
}

// filterItemsJSON applies filterOneJSON to each item in the slice.
func filterItemsJSON(items []any, columns []Column) ([]any, error) {
	result := make([]any, len(items))
	for i, item := range items {
		f, err := filterOneJSON(item, columns)
		if err != nil {
			return nil, err
		}
		result[i] = f
	}
	return result, nil
}
