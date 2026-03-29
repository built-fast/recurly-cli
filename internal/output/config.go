package output

import "github.com/itchyny/gojq"

// Config holds all output configuration: format, jq filter, and field selection.
type Config struct {
	Format string
	JQ     *gojq.Code
	Fields []string
}

// HasJQ reports whether a jq filter is configured.
func (c *Config) HasJQ() bool {
	return c.JQ != nil
}

// HasFields reports whether field selection is active.
func (c *Config) HasFields() bool {
	return len(c.Fields) > 0
}
