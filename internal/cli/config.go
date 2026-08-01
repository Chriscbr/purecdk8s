package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = "cdk8s.yaml"

// SynthesisFormat is the output layout requested for a synthesis.
type SynthesisFormat string

const (
	SynthesisFormatPlain SynthesisFormat = "plain"
	SynthesisFormatHelm  SynthesisFormat = "helm"
)

// HelmChartAPIVersion is the apiVersion used in a Helm Chart.yaml file.
type HelmChartAPIVersion string

const (
	HelmChartAPIVersionV1 HelmChartAPIVersion = "v1"
	HelmChartAPIVersionV2 HelmChartAPIVersion = "v2"
)

// ValidationConfig describes a validation plugin from cdk8s.yaml. Validation
// plugins are parsed so existing project files remain readable, even though the
// pure-Go CLI does not execute JavaScript validation plugins.
type ValidationConfig struct {
	Package    string         `yaml:"package"`
	Version    string         `yaml:"version"`
	Class      string         `yaml:"class"`
	InstallEnv map[string]any `yaml:"installEnv,omitempty"`
	Properties map[string]any `yaml:"properties,omitempty"`
}

// SynthConfig contains the synthesis-specific cdk8s.yaml settings.
type SynthConfig struct {
	Format          *SynthesisFormat     `yaml:"format,omitempty"`
	ChartAPIVersion *HelmChartAPIVersion `yaml:"chartApiVersion,omitempty"`
	ChartVersion    *string              `yaml:"chartVersion,omitempty"`
}

// Config is the cdk8s.yaml format understood by cdk8s-cli. Fields are pointers
// where an explicitly empty value is observably different from an omitted one.
type Config struct {
	App              *string      `yaml:"app,omitempty"`
	Language         *string      `yaml:"language,omitempty"`
	Output           *string      `yaml:"output,omitempty"`
	ImportDirectory  *string      `yaml:"importDirectory,omitempty"`
	Imports          []string     `yaml:"imports,omitempty"`
	PluginsDirectory *string      `yaml:"pluginsDirectory,omitempty"`
	Validations      any          `yaml:"validations,omitempty"`
	SynthConfig      *SynthConfig `yaml:"synthConfig,omitempty"`
}

// ReadConfig reads cdk8s.yaml from dir. A missing config file is not an error.
func ReadConfig(dir string) (*Config, error) {
	return ReadConfigFile(filepath.Join(dir, ConfigFileName))
}

// ReadConfigFile reads a cdk8s configuration file. A missing file returns
// (nil, nil), matching cdk8s-cli's optional configuration behavior.
func ReadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	config := new(Config)
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return config, nil
}
