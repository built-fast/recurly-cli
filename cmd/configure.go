package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/built-fast/recurly-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// configPrompter handles interactive input for the configure command.
type configPrompter struct {
	reader       *bufio.Reader
	writer       io.Writer
	readPassword func() (string, error)
}

func newConfigureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure Recurly CLI settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := &configPrompter{
				reader: bufio.NewReader(cmd.InOrStdin()),
				writer: cmd.OutOrStdout(),
				readPassword: func() (string, error) {
					b, err := term.ReadPassword(int(os.Stdin.Fd()))
					return string(b), err
				},
			}
			return runConfigure(p)
		},
	}
}

func runConfigure(p *configPrompter) error {
	existing := readExistingConfig()

	apiKey, err := p.promptAPIKey(existing.apiKey)
	if err != nil {
		return err
	}

	region, err := p.promptRegion(existing.region)
	if err != nil {
		return err
	}

	site, err := p.promptSite(existing.site)
	if err != nil {
		return err
	}

	if err := config.Write("api_key", apiKey); err != nil {
		return err
	}
	if err := config.Write("region", region); err != nil {
		return err
	}
	if site != "" {
		if err := config.Write("site", site); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintf(p.writer, "Configuration saved to %s\n", config.FilePath())
	return nil
}

type existingConfig struct {
	apiKey string
	region string
	site   string
}

// readExistingConfig reads values directly from the config file,
// bypassing the global viper instance to avoid flag default interference.
func readExistingConfig() existingConfig {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(config.FilePath())
	var ec existingConfig
	if err := v.ReadInConfig(); err == nil {
		ec.apiKey = v.GetString("api_key")
		ec.region = v.GetString("region")
		ec.site = v.GetString("site")
	}
	return ec
}

func (p *configPrompter) promptAPIKey(existing string) (string, error) {
	if existing != "" {
		_, _ = fmt.Fprintf(p.writer, "API Key [%s]: ", maskKey(existing))
	} else {
		_, _ = fmt.Fprint(p.writer, "API Key: ")
	}

	key, err := p.readPassword()
	if err != nil {
		return "", fmt.Errorf("reading API key: %w", err)
	}
	_, _ = fmt.Fprintln(p.writer) // newline after masked input

	key = strings.TrimSpace(key)
	if key == "" {
		if existing != "" {
			return existing, nil
		}
		return "", fmt.Errorf("API key is required")
	}
	return key, nil
}

func (p *configPrompter) promptRegion(existing string) (string, error) {
	dflt := existing
	if dflt == "" {
		dflt = "us"
	}

	for {
		_, _ = fmt.Fprintf(p.writer, "Region (us/eu) [%s]: ", dflt)

		line, err := p.reader.ReadString('\n')
		if err != nil && line == "" {
			return "", fmt.Errorf("reading region: %w", err)
		}

		input := strings.TrimSpace(line)
		if input == "" {
			return dflt, nil
		}

		input = strings.ToLower(input)
		if input == "us" || input == "eu" {
			return input, nil
		}

		_, _ = fmt.Fprintf(p.writer, "Error: invalid region %q, must be us or eu\n", input)
	}
}

func (p *configPrompter) promptSite(existing string) (string, error) {
	if existing != "" {
		_, _ = fmt.Fprintf(p.writer, "Site subdomain [%s]: ", existing)
	} else {
		_, _ = fmt.Fprint(p.writer, "Site subdomain (e.g. mysite from mysite.recurly.com): ")
	}

	line, err := p.reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("reading site: %w", err)
	}

	input := strings.TrimSpace(line)
	if input == "" {
		return existing, nil
	}
	return input, nil
}

func maskKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return "****" + key[len(key)-4:]
}
