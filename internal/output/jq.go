package output

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/itchyny/gojq"
)

// applyJQ runs the compiled jq expression against v and returns the formatted
// output. The format parameter controls how object/array results are rendered:
// "json-pretty" uses indented JSON, anything else uses compact JSON.
func applyJQ(code *gojq.Code, v any, format string) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("jq: marshaling input: %w", err)
	}

	var input any
	if err := json.Unmarshal(b, &input); err != nil {
		return "", fmt.Errorf("jq: unmarshaling input: %w", err)
	}

	iter := code.Run(input)
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
		case bool:
			lines = append(lines, strconv.FormatBool(val))
		case int:
			lines = append(lines, strconv.Itoa(val))
		case float64:
			lines = append(lines, strconv.FormatFloat(val, 'f', -1, 64))
		default:
			var out []byte
			if format == "json-pretty" {
				out, err = json.MarshalIndent(val, "", "  ")
			} else {
				out, err = json.Marshal(val)
			}
			if err != nil {
				return "", fmt.Errorf("jq: encoding result: %w", err)
			}
			lines = append(lines, string(out))
		}
	}
	return strings.Join(lines, "\n"), nil
}
