package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

// withFromFile wraps a command to add --from-file/-F support.
// File values are applied to flags that were not explicitly set on the CLI,
// so CLI flags always take precedence over file values.
func withFromFile(cmd *cobra.Command) *cobra.Command {
	var fromFile string
	cmd.Flags().StringVarP(&fromFile, "from-file", "F", "", "Path to JSON or YAML file (use '-' for stdin)")

	origRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if fromFile != "" {
			if err := applyFromFile(cmd, fromFile); err != nil {
				return err
			}
		}
		return origRunE(cmd, args)
	}

	return cmd
}

// applyFromFile reads a file (or stdin) and sets flag values for keys
// whose corresponding flags were not already set on the command line.
func applyFromFile(cmd *cobra.Command, path string) error {
	data, err := readInputFile(cmd, path)
	if err != nil {
		return err
	}

	flat := flattenMap(data, "")

	for key, val := range flat {
		flagName := keyToFlagName(key)
		f := cmd.Flags().Lookup(flagName)
		if f == nil {
			return fmt.Errorf("unknown key in file: %q (no matching flag --%s)", key, flagName)
		}

		// CLI flags take precedence over file values
		if cmd.Flags().Changed(flagName) {
			continue
		}

		strVal, err := valueToString(val)
		if err != nil {
			return fmt.Errorf("invalid value for key %q: %w", key, err)
		}

		if err := cmd.Flags().Set(flagName, strVal); err != nil {
			return fmt.Errorf("error setting flag --%s from file: %w", flagName, err)
		}
	}

	return nil
}

// readInputFile reads and parses a JSON or YAML file.
// If path is "-", reads from stdin with auto-detection (JSON first, then YAML).
func readInputFile(cmd *cobra.Command, path string) (map[string]any, error) {
	if path == "-" {
		content, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return nil, fmt.Errorf("error reading stdin: %w", err)
		}
		if len(content) == 0 {
			return nil, fmt.Errorf("stdin is empty")
		}
		return parseAutoDetect(content)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return parseJSON(content)
	case ".yaml", ".yml":
		return parseYAML(content)
	default:
		return nil, fmt.Errorf("unsupported file extension %q (use .json, .yaml, or .yml)", ext)
	}
}

func parseJSON(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return result, nil
}

func parseYAML(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}
	return result, nil
}

// parseAutoDetect tries JSON first, falls back to YAML.
func parseAutoDetect(data []byte) (map[string]any, error) {
	result, err := parseJSON(data)
	if err == nil {
		return result, nil
	}
	result, yamlErr := parseYAML(data)
	if yamlErr != nil {
		return nil, fmt.Errorf("unable to parse input as JSON or YAML:\n  JSON: %s\n  YAML: %s", err, yamlErr)
	}
	return result, nil
}

// keyToFlagName converts a file key to a cobra flag name.
// Underscores become hyphens (e.g., "first_name" -> "first-name").
func keyToFlagName(key string) string {
	return strings.ReplaceAll(key, "_", "-")
}

// flattenMap recursively flattens nested maps into a single level with
// hyphen-joined keys. For example: {"address": {"street1": "x"}} becomes
// {"address-street1": "x"}.
func flattenMap(m map[string]any, prefix string) map[string]any {
	result := make(map[string]any)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "-" + k
		}

		if nested, ok := v.(map[string]any); ok {
			for fk, fv := range flattenMap(nested, key) {
				result[fk] = fv
			}
		} else {
			result[key] = v
		}
	}
	return result
}

// valueToString converts a parsed value to a string suitable for cobra's
// Flag.Set() method.
func valueToString(val any) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case float64:
		return fmt.Sprintf("%v", v), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case []any:
		// Convert array to comma-separated string for slice flags
		parts := make([]string, len(v))
		for i, elem := range v {
			s, err := valueToString(elem)
			if err != nil {
				return "", err
			}
			parts[i] = s
		}
		return strings.Join(parts, ","), nil
	case nil:
		return "", nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}
