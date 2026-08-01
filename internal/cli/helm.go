package cli

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/purecdk8s/purecdk8s/internal/importer"
	"gopkg.in/yaml.v3"
)

const helmReadme = "This Helm chart is generated using cdk8s. Any manual changes to the chart would be discarded once cdk8s app is synthesized again with `--format helm`."

var semanticVersionPattern = regexp.MustCompile(
	`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`,
)

type helmChartMetadata struct {
	APIVersion  string `yaml:"apiVersion"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
}

func isValidSemVer(value string) bool {
	if len(value) > 256 {
		return false
	}
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "v")

	matches := semanticVersionPattern.FindStringSubmatch(value)
	if matches == nil {
		return false
	}
	const maxSafeInteger = uint64(9_007_199_254_740_991)
	for _, component := range matches[1:4] {
		number, err := strconv.ParseUint(component, 10, 64)
		if err != nil || number > maxSafeInteger {
			return false
		}
	}
	if prerelease := matches[4]; prerelease != "" {
		for _, identifier := range strings.Split(prerelease, ".") {
			if isNumericIdentifier(identifier) && len(identifier) > 1 && identifier[0] == '0' {
				return false
			}
		}
	}
	return true
}

func isNumericIdentifier(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return value != ""
}

func configImports(config *Config) []string {
	if config == nil {
		return nil
	}
	return config.Imports
}

func isK8sImport(value string) bool {
	return value == "k8s" || strings.HasPrefix(value, "k8s@")
}

func isHelmImport(value string) bool {
	return strings.HasPrefix(value, "helm:") || strings.Contains(value, ":=helm:")
}

func crdsArePresent(imports []string) bool {
	for _, importSpec := range imports {
		if !isK8sImport(importSpec) && !isHelmImport(importSpec) {
			return true
		}
	}
	return false
}

func parseImportSource(specification string) (string, error) {
	parts := strings.Split(specification, ":=")
	switch len(parts) {
	case 1:
		return parts[0], nil
	case 2:
		return parts[1], nil
	default:
		return "", fmt.Errorf("Unable to parse import specification. Syntax is [NAME:=]SPEC")
	}
}

func (r *runner) createHelmScaffolding(
	outputDirectory string,
	options resolvedSynthOptions,
	config *Config,
) error {
	if options.chartAPIVersion == nil || options.chartVersion == nil {
		return fmt.Errorf("chart API version and chart version are required")
	}
	if err := os.MkdirAll(filepath.Join(outputDirectory, "templates"), 0o755); err != nil {
		return fmt.Errorf("create templates directory: %w", err)
	}

	metadata := helmChartMetadata{
		APIVersion:  string(*options.chartAPIVersion),
		Name:        filepath.Base(filepath.Clean(r.workDir)),
		Version:     *options.chartVersion,
		Description: "Generated chart for " + filepath.Base(filepath.Clean(r.workDir)),
		Type:        "application",
	}
	chartYAML, err := yaml.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode Chart.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outputDirectory, "Chart.yaml"), chartYAML, 0o644); err != nil {
		return fmt.Errorf("write Chart.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outputDirectory, "README.md"), []byte(helmReadme), 0o644); err != nil {
		return fmt.Errorf("write README.md: %w", err)
	}

	if *options.chartAPIVersion == HelmChartAPIVersionV2 {
		if err := r.copyHelmCRDs(outputDirectory, configImports(config)); err != nil {
			return err
		}
	}
	return nil
}

func (r *runner) copyHelmCRDs(chartDirectory string, imports []string) error {
	for _, specification := range imports {
		if isK8sImport(specification) || isHelmImport(specification) {
			continue
		}
		source, err := parseImportSource(specification)
		if err != nil {
			return err
		}
		manifest, err := r.readCRDSource(source)
		if err != nil {
			return fmt.Errorf("read CRD import %q: %w", source, err)
		}
		if err := os.MkdirAll(filepath.Join(chartDirectory, "crds"), 0o755); err != nil {
			return fmt.Errorf("create crds directory: %w", err)
		}
		filename := deriveCRDFilename(source, r.workDir)
		if err := os.WriteFile(
			filepath.Join(chartDirectory, "crds", filename+".yaml"),
			manifest,
			0o644,
		); err != nil {
			return fmt.Errorf("write CRD import %q: %w", source, err)
		}
	}
	return nil
}

func (r *runner) readCRDSource(source string) ([]byte, error) {
	resolvedSource := importer.NormalizeSource(source)
	parsed, err := url.Parse(resolvedSource)
	if err != nil {
		return nil, err
	}
	switch parsed.Scheme {
	case "http", "https":
		client := r.httpClient
		if client == nil {
			client = &http.Client{Timeout: 30 * time.Second}
		}
		response, err := client.Get(resolvedSource)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			_, _ = io.Copy(io.Discard, response.Body)
			return nil, fmt.Errorf("%s: %s", response.Status, resolvedSource)
		}
		return io.ReadAll(response.Body)
	case "", "file":
		path := resolvedSource
		if parsed.Scheme == "file" {
			path = parsed.Path
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(r.workDir, path)
		}
		return os.ReadFile(path)
	default:
		return nil, fmt.Errorf("unsupported protocol %s", parsed.Scheme+":")
	}
}

func deriveCRDFilename(source, workDirectory string) string {
	if normalized, ok := importer.NormalizeGitHubSource(source); ok {
		repository := normalized
		if at := strings.LastIndex(repository, "@"); at >= 0 {
			repository = repository[:at]
		}
		if slash := strings.LastIndex(repository, "/"); slash >= 0 && slash < len(repository)-1 {
			return repository[slash+1:]
		}
	}

	lastSlash := strings.LastIndex(source, "/")
	lastYAML := strings.LastIndex(source, ".yaml")
	if lastYAML <= 0 {
		lastYAML = strings.LastIndex(source, ".yml")
	}
	if lastSlash > 0 && lastYAML > 0 {
		filename := source[lastSlash+1 : lastYAML]
		if filename != "" {
			return filename
		}
	}

	hashInput := source
	localPath := source
	if !filepath.IsAbs(localPath) {
		localPath = filepath.Join(workDirectory, localPath)
	}
	if _, err := os.Stat(localPath); err == nil {
		hashInput = filepath.Base(source)
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(hashInput)))
}
