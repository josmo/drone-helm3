package helm

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"io"
	"regexp"
)

var justNumbers = regexp.MustCompile(`^\d+$`)

// The Config struct captures the `settings` and `environment` blocks in the application's drone
// config. Configuration in drone's `settings` block arrives as uppercase env vars matching the
// config key, prefixed with `PLUGIN_`. Config from the `environment` block is uppercased, but does
// not have the `PLUGIN_` prefix. It may, however, be prefixed with the value in `$PLUGIN_PREFIX`.
type Config struct {
	// Configuration for drone-helm itself
	Command            string   `envconfig:"HELM_COMMAND"`           // Helm command to run
	DroneEvent         string   `envconfig:"DRONE_BUILD_EVENT"`      // Drone event that invoked this plugin.
	UpdateDependencies bool     `split_words:"true"`                 // Call `helm dependency update` before the main command
	AddRepos           []string `envconfig:"HELM_REPOS"`             // Call `helm repo add` before the main command
	Prefix             string   ``                                   // Prefix to use when looking up secret env vars
	Debug              bool     ``                                   // Generate debug output and pass --debug to all helm commands
	Values             string   ``                                   // Argument to pass to --set in applicable helm commands
	StringValues       string   `split_words:"true"`                 // Argument to pass to --set-string in applicable helm commands
	ValuesFiles        []string `split_words:"true"`                 // Arguments to pass to --values in applicable helm commands
	Namespace          string   ``                                   // Kubernetes namespace for all helm commands
	KubeToken          string   `envconfig:"KUBERNETES_TOKEN"`       // Kubernetes authentication token to put in .kube/config
	SkipTLSVerify      bool     `envconfig:"SKIP_TLS_VERIFY"`        // Put insecure-skip-tls-verify in .kube/config
	Certificate        string   `envconfig:"KUBERNETES_CERTIFICATE"` // The Kubernetes cluster CA's self-signed certificate (must be base64-encoded)
	APIServer          string   `envconfig:"API_SERVER"`             // The Kubernetes cluster's API endpoint
	ServiceAccount     string   `split_words:"true"`                 // Account to use for connecting to the Kubernetes cluster
	ChartVersion       string   `split_words:"true"`                 // Specific chart version to use in `helm upgrade`
	DryRun             bool     `split_words:"true"`                 // Pass --dry-run to applicable helm commands
	Wait               bool     ``                                   // Pass --wait to applicable helm commands
	ReuseValues        bool     `split_words:"true"`                 // Pass --reuse-values to `helm upgrade`
	Timeout            string   ``                                   // Argument to pass to --timeout in applicable helm commands
	Chart              string   ``                                   // Chart argument to use in applicable helm commands
	Release            string   ``                                   // Release argument to use in applicable helm commands
	Force              bool     ``                                   // Pass --force to applicable helm commands

	Stdout io.Writer `ignored:"true"`
	Stderr io.Writer `ignored:"true"`
}

// NewConfig creates a Config and reads environment variables into it, accounting for several possible formats.
func NewConfig(stdout, stderr io.Writer) (*Config, error) {
	cfg := Config{
		Stdout: stdout,
		Stderr: stderr,
	}
	if err := envconfig.Process("plugin", &cfg); err != nil {
		return nil, err
	}

	prefix := cfg.Prefix

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	if prefix != "" {
		if err := envconfig.Process(cfg.Prefix, &cfg); err != nil {
			return nil, err
		}
	}

	if justNumbers.MatchString(cfg.Timeout) {
		cfg.Timeout = fmt.Sprintf("%ss", cfg.Timeout)
	}

	if cfg.Debug && cfg.Stderr != nil {
		cfg.logDebug()
	}

	return &cfg, nil
}

func (cfg Config) logDebug() {
	if cfg.KubeToken != "" {
		cfg.KubeToken = "(redacted)"
	}
	fmt.Fprintf(cfg.Stderr, "Generated config: %+v\n", cfg)
}
