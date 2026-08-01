package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportLocalHelmChart(t *testing.T) {
	workDir := t.TempDir()
	chartDirectory := filepath.Join(workDir, "charts", "webapp")
	if err := os.MkdirAll(chartDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(chartDirectory, "Chart.yaml"), `apiVersion: v2
name: webapp
version: 1.2.3
`)
	writeTestFile(t, filepath.Join(chartDirectory, "values.schema.json"), `{
  "type": "object",
  "properties": {
    "replicas": {
      "type": "integer"
    }
  }
}`)

	code, stdout, stderr := runTestCLI(
		t,
		workDir,
		[]string{"import", "helm:./charts/webapp", "--output", "imports", "--no-save"},
		nil,
	)
	if code != 0 {
		t.Fatalf("Run() code = %d; stdout = %q; stderr = %q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, filepath.Join("imports", "webapp", "generated.go")) {
		t.Fatalf("stdout = %q, want generated Helm package path", stdout)
	}
	generated := readTestFile(t, filepath.Join(workDir, "imports", "webapp", "generated.go"))
	for _, expected := range []string{
		"type WebappValues struct",
		"type WebappProps struct",
		"func NewWebapp(",
		"cdk8s.NewHelm(",
		`chart, version := "./charts/webapp", "1.2.3"`,
	} {
		if !strings.Contains(generated, expected) {
			t.Errorf("generated Helm package does not contain %q", expected)
		}
	}
	if strings.Contains(generated, chartDirectory) {
		t.Fatalf("generated Helm package leaked absolute chart path %q", chartDirectory)
	}
}
