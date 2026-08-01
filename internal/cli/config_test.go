package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadConfig(t *testing.T) {
	t.Run("missing file is optional", func(t *testing.T) {
		config, err := ReadConfig(t.TempDir())
		if err != nil {
			t.Fatalf("ReadConfig() error = %v", err)
		}
		if config != nil {
			t.Fatalf("ReadConfig() = %#v, want nil", config)
		}
	})

	t.Run("all supported fields", func(t *testing.T) {
		dir := t.TempDir()
		writeTestFile(t, filepath.Join(dir, ConfigFileName), `
app: go run .
language: go
output: manifests
importDirectory: generated
imports:
  - k8s
  - example.com/crd.yaml
pluginsDirectory: .plugins
validations:
  - package: example-validator
    version: 1.2.3
    class: Validator
synthConfig:
  format: plain
  chartApiVersion: v2
  chartVersion: 1.0.0
`)

		config, err := ReadConfig(dir)
		if err != nil {
			t.Fatalf("ReadConfig() error = %v", err)
		}
		assertStringPointer(t, "App", config.App, "go run .")
		assertStringPointer(t, "Language", config.Language, "go")
		assertStringPointer(t, "Output", config.Output, "manifests")
		assertStringPointer(t, "ImportDirectory", config.ImportDirectory, "generated")
		assertStringPointer(t, "PluginsDirectory", config.PluginsDirectory, ".plugins")
		if got, want := strings.Join(config.Imports, ","), "k8s,example.com/crd.yaml"; got != want {
			t.Fatalf("Imports = %q, want %q", got, want)
		}
		if config.Validations == nil {
			t.Fatal("Validations = nil, want parsed validation configuration")
		}
		if config.SynthConfig == nil {
			t.Fatal("SynthConfig = nil")
		}
		if config.SynthConfig.Format == nil || *config.SynthConfig.Format != SynthesisFormatPlain {
			t.Fatalf("SynthConfig.Format = %#v, want %q", config.SynthConfig.Format, SynthesisFormatPlain)
		}
		if config.SynthConfig.ChartAPIVersion == nil || *config.SynthConfig.ChartAPIVersion != HelmChartAPIVersionV2 {
			t.Fatalf("SynthConfig.ChartAPIVersion = %#v, want %q", config.SynthConfig.ChartAPIVersion, HelmChartAPIVersionV2)
		}
		assertStringPointer(t, "SynthConfig.ChartVersion", config.SynthConfig.ChartVersion, "1.0.0")
	})

	t.Run("invalid yaml is reported", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ConfigFileName)
		writeTestFile(t, path, "app: [\n")

		config, err := ReadConfigFile(path)
		if err == nil {
			t.Fatalf("ReadConfigFile() = %#v, nil; want an error", config)
		}
		if !strings.Contains(err.Error(), "parse "+path) {
			t.Fatalf("ReadConfigFile() error = %q, want path and parse context", err)
		}
	})
}

func assertStringPointer(t *testing.T, name string, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Fatalf("%s = %#v, want %q", name, got, want)
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
