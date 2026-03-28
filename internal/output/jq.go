package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
)

var jqCode *gojq.Code

// SetJQ stores compiled jq bytecode for use by the output layer.
func SetJQ(code *gojq.Code) {
	jqCode = code
}

// HasJQ reports whether a jq filter is active.
func HasJQ() bool {
	return jqCode != nil
}

// applyJQ runs the compiled jq expression against v and returns the formatted output.
func applyJQ(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("jq: marshaling input: %w", err)
	}

	var input any
	if err := json.Unmarshal(b, &input); err != nil {
		return "", fmt.Errorf("jq: unmarshaling input: %w", err)
	}

	iter := jqCode.Run(input)
	var lines []string
	for {
		result, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := result.(error); isErr {
			return "", fmt.Errorf("jq: %w", err)
		}

		switch val := result.(type) {
		case nil:
			lines = append(lines, "null")
		case string:
			lines = append(lines, val)
		default:
			out, err := json.MarshalIndent(val, "", "  ")
			if err != nil {
				return "", fmt.Errorf("jq: encoding result: %w", err)
			}
			lines = append(lines, string(out))
		}
	}
	return strings.Join(lines, "\n"), nil
}
