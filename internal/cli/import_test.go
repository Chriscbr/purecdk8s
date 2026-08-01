package cli

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.example.com
spec:
  group: example.com
  names:
    kind: Widget
    plural: widgets
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required: [image]
              properties:
                image:
                  type: string
`

func TestImportHelpDocumentsExcludeArray(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"import", "--help"}, Options{
		Stdout:  &stdout,
		Stderr:  &stderr,
		WorkDir: t.TempDir(),
	})
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("code = %d\nstderr: %s", code, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "--exclude") || !strings.Contains(output, "repeatable") {
		t.Fatalf("import help does not document repeatable --exclude:\n%s", output)
	}
	if !strings.Contains(output, "--check-upgrade") {
		t.Fatalf("import help does not document --check-upgrade:\n%s", output)
	}
}

func TestImportLocalCRD(t *testing.T) {
	workDir := t.TempDir()
	source := filepath.Join(workDir, "widget.yaml")
	if err := os.WriteFile(source, []byte(testCRD), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"import", "widgets:=" + source,
		"--output", "imports",
		"--check-upgrade=false",
		"--no-save",
	}, Options{
		Stdout:  &stdout,
		Stderr:  &stderr,
		WorkDir: workDir,
	})
	if code != 0 {
		t.Fatalf("code = %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
	generated := filepath.Join(workDir, "imports", "widgets_examplecom", "generated.go")
	data, err := os.ReadFile(generated)
	if err != nil {
		t.Fatalf("generated source: %v", err)
	}
	if strings.Contains(string(data), "aws/jsii") || !strings.Contains(string(data), "func NewWidget(") {
		t.Fatalf("unexpected generated source:\n%s", data)
	}
	if _, err := os.Stat(filepath.Join(workDir, ConfigFileName)); !os.IsNotExist(err) {
		t.Fatalf("--no-save wrote cdk8s.yaml: %v", err)
	}
}

func TestImportKubernetesExcludes(t *testing.T) {
	workDir := t.TempDir()
	schema, err := os.ReadFile(filepath.Join("..", "importer", "testdata", "k8s-definitions.json"))
	if err != nil {
		t.Fatal(err)
	}
	requests := 0
	client := &http.Client{Transport: cliRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(schema)),
			Request:    request,
		}, nil
	})}
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"import", "k8s@1.25.0",
		"--exclude",
		`^#/definitions/io\.k8s\.apimachinery\.pkg\.apis\.meta\.v1\.ObjectMeta$`,
		`Recursive$`,
		"--exclude", `Quantity$`,
		"--output", "imports",
		"--no-save",
	}, Options{
		Stdout:     &stdout,
		Stderr:     &stderr,
		WorkDir:    workDir,
		HTTPClient: client,
	})
	if code != 0 {
		t.Fatalf("code = %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
	if requests != 1 {
		t.Fatalf("schema requests = %d, want 1", requests)
	}
	generated, err := os.ReadFile(filepath.Join(workDir, "imports", "k8s", "generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(generated)
	for _, excluded := range []string{"type ObjectMeta struct", "type Recursive struct", "type Quantity interface"} {
		if strings.Contains(source, excluded) {
			t.Errorf("generated source retained excluded type %q", excluded)
		}
	}
	canonical := strings.Join(strings.Fields(source), " ")
	for _, replacement := range []string{"Metadata interface{}", "Recursive interface{}", "Limits *map[string]interface{}"} {
		if !strings.Contains(canonical, replacement) {
			t.Errorf("generated source does not contain excluded-reference replacement %q", replacement)
		}
	}
}

func TestImportKubernetesRejectsInvalidExcludeBeforeDownload(t *testing.T) {
	workDir := t.TempDir()
	requests := 0
	client := &http.Client{Transport: cliRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		return nil, nil
	})}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"import", "k8s", "--exclude", "[", "--no-save"}, Options{
		Stdout:     &stdout,
		Stderr:     &stderr,
		WorkDir:    workDir,
		HTTPClient: client,
	})
	if code == 0 || !strings.Contains(stderr.String(), `invalid Kubernetes exclude regular expression "["`) {
		t.Fatalf("code = %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
	if requests != 0 {
		t.Fatalf("invalid exclude made %d schema requests", requests)
	}
}

func TestImportSavesConfig(t *testing.T) {
	workDir := t.TempDir()
	source := filepath.Join(workDir, "widget.yaml")
	if err := os.WriteFile(source, []byte(testCRD), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"gen", source}, Options{Stdout: &stdout, Stderr: &stderr, WorkDir: workDir})
	if code != 0 {
		t.Fatalf("code = %d: %s", code, stderr.String())
	}
	config, err := ReadConfig(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if config == nil || len(config.Imports) != 1 || config.Imports[0] != source {
		t.Fatalf("saved config = %#v", config)
	}
}

type cliRoundTripFunc func(*http.Request) (*http.Response, error)

func (function cliRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
