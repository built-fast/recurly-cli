package output

import (
	"context"

	"github.com/itchyny/gojq"
)

// contextKey is an unexported type used as the key for storing Config in a context.
type contextKey struct{}

// NewContext returns a new context that carries the given Config.
func NewContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, contextKey{}, cfg)
}

// FromContext returns the Config stored in ctx. If no Config is set,
// it returns a non-nil zero-value Config.
func FromContext(ctx context.Context) *Config {
	if cfg, ok := ctx.Value(contextKey{}).(*Config); ok && cfg != nil {
		return cfg
	}
	return &Config{}
}

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
