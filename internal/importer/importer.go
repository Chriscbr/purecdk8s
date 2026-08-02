// Package importer implements the cdk8s schema importer in native Go.
//
// It consumes the same Kubernetes _definitions.json and
// CustomResourceDefinition inputs as the upstream CLI, but emits ordinary Go
// source that depends only on purecdk8s and constructs.
package importer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultKubernetesVersion matches the default used by the upstream cdk8s
	// importer.
	DefaultKubernetesVersion = "1.25.0"

	kubernetesSchemaURL = "https://raw.githubusercontent.com/cdk8s-team/cdk8s/master/kubernetes-schemas/v%s/_definitions.json"
)

var (
	kubernetesVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	githubCRDSourcePattern   = regexp.MustCompile(`^github:([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+)(?:@([0-9]+)\.([0-9]+)(?:\.([0-9]+))?)?$`)
)

// Options describes one cdk8s import entry. OutputDir is the imports base
// directory: a "k8s" source is written to OutputDir/k8s/generated.go, Helm
// charts are written to one package named after the chart, and CRDs are
// written to one child package per API group. PackageName is the
// upstream-compatible NAME prefix from NAME:=SOURCE, not an exact package
// override.
type Options struct {
	Source                 string
	OutputDir              string
	PackageName            string
	ClassNamePrefix        string
	DisableClassNamePrefix bool
	Excludes               []string
	CoreImport             string
	ConstructsImport       string
	HTTPClient             *http.Client
	WorkDir                string
	HelmExecutable         string
}

// PackageResult describes a package written by Import.
type PackageResult struct {
	PackageName string
	Group       string
	Directory   string
	File        string
	Resources   []string
}

// Result describes all packages generated for an import entry.
type Result struct {
	Source   string
	Version  string
	Packages []PackageResult
}

// Import fetches or reads an import source, generates ordinary Go code and
// writes it below Options.OutputDir.
func Import(ctx context.Context, options Options) (*Result, error) {
	if strings.TrimSpace(options.Source) == "" {
		return nil, fmt.Errorf("import source is required")
	}
	if strings.TrimSpace(options.OutputDir) == "" {
		return nil, fmt.Errorf("import output directory is required")
	}
	generateOptions := GenerateOptions{
		ClassNamePrefix:        options.ClassNamePrefix,
		DisableClassNamePrefix: options.DisableClassNamePrefix,
		Excludes:               append([]string(nil), options.Excludes...),
		CoreImport:             options.CoreImport,
		ConstructsImport:       options.ConstructsImport,
	}
	result := &Result{Source: options.Source}
	var generations []*Generation
	if parsed, ok, err := ParseHelmSource(options.Source); err != nil {
		return nil, err
	} else if ok {
		generateOptions.PackagePrefix = options.PackageName
		generated, acquired, err := GenerateHelmSource(
			ctx,
			options.Source,
			options.WorkDir,
			options.HelmExecutable,
			generateOptions,
		)
		if err != nil {
			return nil, err
		}
		if acquired == nil {
			acquired = parsed
		}
		result.Version = acquired.ChartVersion
		generations = []*Generation{generated}
	} else if version, ok, err := ParseKubernetesSource(options.Source); err != nil {
		return nil, err
	} else if ok {
		excludes, err := compileExcludePatterns(generateOptions.Excludes)
		if err != nil {
			return nil, err
		}
		generateOptions.excludeRegexps = excludes
		data, err := download(ctx, options.HTTPClient, fmt.Sprintf(kubernetesSchemaURL, version))
		if err != nil {
			return nil, fmt.Errorf("download k8s@%s schema: %w", version, err)
		}
		generateOptions.PackageName = prefixedPackageName(options.PackageName, "k8s")
		generated, err := GenerateKubernetes(data, generateOptions)
		if err != nil {
			return nil, err
		}
		result.Version = version
		generations = []*Generation{generated}
	} else {
		data, err := readSource(ctx, options.HTTPClient, NormalizeSource(options.Source))
		if err != nil {
			return nil, err
		}
		generateOptions.PackagePrefix = options.PackageName
		generations, err = GenerateCRDs(data, generateOptions)
		if err != nil {
			return nil, err
		}
	}

	for _, generated := range generations {
		directory, err := replacePackageDirectory(options.OutputDir, generated.PackageName)
		if err != nil {
			return nil, err
		}
		filename := filepath.Join(directory, "generated.go")
		if err := os.WriteFile(filename, generated.Code, 0o644); err != nil {
			return nil, fmt.Errorf("write generated package %s: %w", filename, err)
		}
		result.Packages = append(result.Packages, PackageResult{
			PackageName: generated.PackageName,
			Group:       generated.Group,
			Directory:   directory,
			File:        filename,
			Resources:   append([]string(nil), generated.Resources...),
		})
	}
	return result, nil
}

func prefixedPackageName(prefix, module string) string {
	module = sanitizePackageName(module)
	if strings.TrimSpace(prefix) == "" {
		return module
	}
	return sanitizePackagePrefix(prefix) + "_" + module
}

// NormalizeSource maps the upstream github:account/repository shorthand to
// doc.crds.dev and leaves every other source unchanged.
func NormalizeSource(source string) string {
	if normalized, ok := NormalizeGitHubSource(source); ok {
		return normalized
	}
	return source
}

// NormalizeGitHubSource maps a cdk8s github: CRD source to doc.crds.dev.
// Versions with no patch component are normalized to patch zero.
func NormalizeGitHubSource(source string) (string, bool) {
	match := githubCRDSourcePattern.FindStringSubmatch(strings.TrimSpace(source))
	if match == nil {
		return "", false
	}
	result := "https://doc.crds.dev/raw/github.com/" + match[1] + "/" + match[2]
	if match[3] != "" {
		patch := match[5]
		if patch == "" {
			patch = "0"
		}
		result += "@v" + match[3] + "." + match[4] + "." + patch
	}
	return result, true
}

// ParseKubernetesSource recognizes "k8s" and "k8s@X.Y.Z". The boolean is
// false for CRD paths and URLs.
func ParseKubernetesSource(source string) (version string, kubernetes bool, err error) {
	source = strings.TrimSpace(source)
	if source == "k8s" {
		return DefaultKubernetesVersion, true, nil
	}
	if !strings.HasPrefix(source, "k8s@") {
		return "", false, nil
	}
	version = strings.TrimPrefix(source, "k8s@")
	if !kubernetesVersionPattern.MatchString(version) {
		return "", true, fmt.Errorf("expected k8s version %q to match format <major>.<minor>.<patch>", version)
	}
	return version, true, nil
}

func readSource(ctx context.Context, client *http.Client, source string) ([]byte, error) {
	parsed, err := url.Parse(source)
	if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") {
		return download(ctx, client, source)
	}
	filename := source
	if err == nil && parsed.Scheme == "file" {
		filename = parsed.Path
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read import source %s: %w", source, err)
	}
	return data, nil
}

func download(ctx context.Context, client *http.Client, source string) ([]byte, error) {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("User-Agent", "purecdk8s-importer")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
		return nil, fmt.Errorf("GET %s: %s", source, response.Status)
	}
	const maximumSchemaSize = 64 << 20
	data, err := io.ReadAll(io.LimitReader(response.Body, maximumSchemaSize+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", source, err)
	}
	if len(data) > maximumSchemaSize {
		return nil, fmt.Errorf("read %s: response exceeds %d bytes", source, maximumSchemaSize)
	}
	return data, nil
}
