package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSynthHelmV2ScaffoldsChartAndCopiesLocalCRDs(t *testing.T) {
	requireUnixShell(t)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	const crd = "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\n"
	writeTestFile(t, filepath.Join(dir, "specs", "widgets.yaml"), crd)
	writeTestFile(t, filepath.Join(dir, ConfigFileName), `
imports:
  - k8s
  - k8s@1.31.0
  - helm:./vendor/example
  - example:=helm:https://charts.example.test/example@1.0.0
  - widgets:=specs/widgets.yaml
`)

	code, stdout, stderr := runTestCLI(
		t,
		dir,
		[]string{
			"synth",
			"--app", helperCommand(t),
			"--output", "helm-dist",
			"--format", "helm",
			"--chart-version", "1.2.3",
		},
		helperEnvironment(),
	)
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stdout = %q; stderr = %q", code, stdout, stderr)
	}
	wantStdout := "Synthesizing application\n" +
		"  - " + filepath.Join("helm-dist", "templates", "z.yaml") + "\n" +
		"  - " + filepath.Join("helm-dist", "templates", "sub", "a.yaml") + "\n"
	if stdout != wantStdout {
		t.Fatalf("stdout = %q, want %q", stdout, wantStdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	chartDirectory := filepath.Join(dir, "helm-dist")
	var metadata helmChartMetadata
	chartYAML := readTestFile(t, filepath.Join(chartDirectory, "Chart.yaml"))
	if err := yaml.Unmarshal([]byte(chartYAML), &metadata); err != nil {
		t.Fatalf("parse Chart.yaml: %v", err)
	}
	wantMetadata := helmChartMetadata{
		APIVersion:  "v2",
		Name:        filepath.Base(dir),
		Version:     "1.2.3",
		Description: "Generated chart for " + filepath.Base(dir),
		Type:        "application",
	}
	if metadata != wantMetadata {
		t.Fatalf("Chart.yaml metadata = %#v, want %#v", metadata, wantMetadata)
	}
	if got := readTestFile(t, filepath.Join(chartDirectory, "README.md")); got != helmReadme {
		t.Fatalf("README.md = %q, want %q", got, helmReadme)
	}
	if got := readTestFile(t, filepath.Join(chartDirectory, "crds", "widgets.yaml")); got != crd {
		t.Fatalf("copied CRD = %q, want %q", got, crd)
	}
	if got := readTestFile(t, filepath.Join(chartDirectory, "templates", "z.yaml")); got != "kind: Root\n" {
		t.Fatalf("root template = %q", got)
	}
	if _, err := os.Stat(filepath.Join(chartDirectory, "z.yaml")); !os.IsNotExist(err) {
		t.Fatalf("manifest was written outside templates; stat error = %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(chartDirectory, "crds"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "widgets.yaml" {
		t.Fatalf("crds entries = %v, want only widgets.yaml", entryNames(entries))
	}
}

func TestSynthHelmV2CopiesHTTPCRD(t *testing.T) {
	requireUnixShell(t)
	const crd = "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n  name: gadgets.example.test\n"
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/gadgets.yaml" {
			http.NotFound(response, request)
			return
		}
		_, _ = response.Write([]byte(crd))
	}))
	defer server.Close()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ConfigFileName), fmt.Sprintf(
		"imports:\n  - %s\n",
		strconv.Quote(server.URL+"/gadgets.yaml"),
	))
	code, stdout, stderr := runTestCLI(
		t,
		dir,
		[]string{
			"synth",
			"--app", helperCommand(t),
			"--format", "helm",
			"--chart-version", "2.0.0-beta.1+build.7",
		},
		helperEnvironment(),
	)
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stdout = %q; stderr = %q", code, stdout, stderr)
	}
	if got := readTestFile(t, filepath.Join(dir, "dist", "crds", "gadgets.yaml")); got != crd {
		t.Fatalf("copied HTTP CRD = %q, want %q", got, crd)
	}
}

func TestSynthHelmV2CopiesGitHubCRDShorthand(t *testing.T) {
	const crd = "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n  name: composites.example.test\n"
	tests := []struct {
		name       string
		importSpec string
		wantURL    string
	}{
		{
			name:       "default branch",
			importSpec: "github:crossplane/crossplane",
			wantURL:    "https://doc.crds.dev/raw/github.com/crossplane/crossplane",
		},
		{
			name:       "named import with minor version",
			importSpec: "platform:=github:crossplane/crossplane@1.15",
			wantURL:    "https://doc.crds.dev/raw/github.com/crossplane/crossplane@v1.15.0",
		},
		{
			name:       "patch version",
			importSpec: "github:crossplane/crossplane@1.15.2",
			wantURL:    "https://doc.crds.dev/raw/github.com/crossplane/crossplane@v1.15.2",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var requestedURL string
			client := &http.Client{Transport: cliRoundTripFunc(func(request *http.Request) (*http.Response, error) {
				requestedURL = request.URL.String()
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Header:     make(http.Header),
					Body:       io.NopCloser(bytes.NewBufferString(crd)),
					Request:    request,
				}, nil
			})}
			chartDirectory := t.TempDir()
			r := &runner{workDir: t.TempDir(), httpClient: client}
			if err := r.copyHelmCRDs(chartDirectory, []string{test.importSpec}); err != nil {
				t.Fatal(err)
			}
			if requestedURL != test.wantURL {
				t.Fatalf("requested URL = %q, want %q", requestedURL, test.wantURL)
			}
			filename := filepath.Join(chartDirectory, "crds", "crossplane.yaml")
			if got := readTestFile(t, filename); got != crd {
				t.Fatalf("copied GitHub CRD = %q, want %q", got, crd)
			}
		})
	}
}

func TestSynthHelmV1WithoutCRDs(t *testing.T) {
	requireUnixShell(t)
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ConfigFileName), `
imports:
  - k8s
  - chart:=helm:./vendor/chart
`)
	code, stdout, stderr := runTestCLI(
		t,
		dir,
		[]string{
			"synth",
			"--app", helperCommand(t),
			"--format", "helm",
			"--chart-api-version", "v1",
			"--chart-version", "1.0.0",
		},
		helperEnvironment(),
	)
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stdout = %q; stderr = %q", code, stdout, stderr)
	}
	if got := readTestFile(t, filepath.Join(dir, "dist", "Chart.yaml")); !strings.HasPrefix(got, "apiVersion: v1\n") {
		t.Fatalf("Chart.yaml = %q, want apiVersion v1", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "dist", "crds")); !os.IsNotExist(err) {
		t.Fatalf("v1 chart unexpectedly has crds directory; stat error = %v", err)
	}
}

func TestSynthHelmValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		args      []string
		wantError string
	}{
		{
			name:      "chart version is required",
			args:      []string{"synth", "--app", "exit 99", "--format", "helm"},
			wantError: "You need to specify '--chart-version'",
		},
		{
			name:      "chart version must be SemVer",
			args:      []string{"synth", "--app", "exit 99", "--format", "helm", "--chart-version", "1.2"},
			wantError: "does not follow SemVer-2",
		},
		{
			name:      "chart API version must be supported",
			args:      []string{"synth", "--app", "exit 99", "--format", "helm", "--chart-api-version", "v3", "--chart-version", "1.2.3"},
			wantError: "helm chart api version either as v1 or v2",
		},
		{
			name:      "stdout is unsupported",
			args:      []string{"synth", "--app", "exit 99", "--stdout", "--format", "helm", "--chart-version", "1.2.3"},
			wantError: "Helm format synthesis does not support 'stdout'",
		},
		{
			name:      "v1 rejects CRD imports",
			config:    "imports:\n  - specs/widgets.yaml\n",
			args:      []string{"synth", "--app", "exit 99", "--format", "helm", "--chart-api-version", "v1", "--chart-version", "1.2.3"},
			wantError: "Your application uses CRDs, which are not supported",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			if test.config != "" {
				writeTestFile(t, filepath.Join(dir, ConfigFileName), test.config)
			}
			code, stdout, stderr := runTestCLI(t, dir, test.args, nil)
			if code != 1 {
				t.Fatalf("Run() code = %d, want 1; stdout = %q; stderr = %q", code, stdout, stderr)
			}
			if !strings.Contains(stderr, test.wantError) {
				t.Fatalf("stderr = %q, want substring %q", stderr, test.wantError)
			}
			if strings.Contains(stderr, "non-zero exit code 99") {
				t.Fatalf("application ran before Helm validation: %q", stderr)
			}
		})
	}
}

func TestIsValidSemVer(t *testing.T) {
	tests := map[string]bool{
		"1.2.3":                  true,
		"v1.2.3":                 true,
		" 1.2.3 ":                true,
		"0.0.0":                  true,
		"1.2.3-alpha.1":          true,
		"1.2.3+build.01":         true,
		"1.2.3-alpha.1+build.7":  true,
		"9007199254740991.2.3":   true,
		"":                       false,
		"V1.2.3":                 false,
		"1.2":                    false,
		"01.2.3":                 false,
		"1.2.3-alpha.01":         false,
		"1.2.3-alpha..1":         false,
		"1.2.3+":                 false,
		"9007199254740992.2.3":   false,
		strings.Repeat("a", 257): false,
	}
	for version, want := range tests {
		t.Run(version, func(t *testing.T) {
			if got := isValidSemVer(version); got != want {
				t.Fatalf("isValidSemVer(%q) = %t, want %t", version, got, want)
			}
		})
	}
}

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}
