package importer

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestGenerateHelmWithTypedValuesCompilesAndSynthesizes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the generated synthesis smoke test uses a POSIX fake Helm executable")
	}
	chartDirectory := createTestHelmChart(t, true)
	generated, err := GenerateHelm(chartDirectory, GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if generated.PackageName != "webapp" {
		t.Fatalf("PackageName = %q, want webapp", generated.PackageName)
	}
	source := string(generated.Code)
	canonical := canonicalSource(source)
	for _, expected := range []string{
		"type WebappValues struct",
		"ReplicaCount *float64 `field:\"required\" json:\"replicaCount\" yaml:\"replicaCount\"`",
		"AdditionalValues *map[string]interface{} `field:\"optional\" json:\"additionalValues\" yaml:\"additionalValues\"`",
		"Global *map[string]interface{} `field:\"optional\" json:\"global\" yaml:\"global\"`",
		"Redis *map[string]interface{} `field:\"optional\" json:\"redis\" yaml:\"redis\"`",
		"type WebappImage struct",
		"type WebappMode string",
		"WebappMode_PROD WebappMode = \"PROD\"",
		"type WebappProps struct",
		"Values *WebappValues `field:\"required\" json:\"values\" yaml:\"values\"`",
		"HelmExecutable *string",
		"HelmFlags *[]*string",
		"Namespace *string",
		"ReleaseName *string",
		"type Webapp interface",
		"constructs.Construct",
		"Helm() cdk8s.Helm",
		"SetHelm(value cdk8s.Helm)",
		"func NewWebapp(",
		"func NewWebapp_Override(",
		"func Webapp_IsConstruct(",
		"constructs.NewConstruct(scope, id)",
		"cdk8s.NewHelm(value, &helmID, finalProps)",
		strconv.Quote(chartDirectory),
		"chart, version := " + strconv.Quote(chartDirectory) + ", \"1.2.3\"",
	} {
		if !strings.Contains(canonical, canonicalSource(expected)) {
			t.Errorf("generated Helm source does not contain %q", expected)
		}
	}

	compileGenerated(t, generated, generatedHelmSynthesisTest)
}

func TestGenerateHelmInlineTypesUseChartFQN(t *testing.T) {
	chartDirectory := createTestHelmChart(t, true)
	writeImporterTestFile(
		t,
		filepath.Join(chartDirectory, helmChartFileName),
		"apiVersion: v2\nname: sample-chart\nversion: 1.2.3\n",
	)

	generated, err := GenerateHelm(chartDirectory, GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	source := canonicalSource(string(generated.Code))
	for _, expected := range []string{
		"type SamplechartValues struct",
		"Image *SampleChartImage",
		"Mode SampleChartMode",
		"type SampleChartImage struct",
		"type SampleChartMode string",
	} {
		if !strings.Contains(source, canonicalSource(expected)) {
			t.Errorf("generated Helm source does not contain %q", expected)
		}
	}
	for _, staleName := range []string{"SamplechartValuesImage", "SamplechartValuesMode"} {
		if strings.Contains(source, staleName) {
			t.Errorf("generated Helm source contains non-upstream inline type %q", staleName)
		}
	}
	compileGenerated(t, generated)
}

func TestGenerateHelmWithoutValuesSchemaUsesUntypedMap(t *testing.T) {
	chartDirectory := createTestHelmChart(t, false)
	generated, err := GenerateHelm(chartDirectory, GenerateOptions{PackageName: "custom_chart"})
	if err != nil {
		t.Fatal(err)
	}
	if generated.PackageName != "custom_chart" {
		t.Fatalf("PackageName = %q, want custom_chart", generated.PackageName)
	}
	canonical := canonicalSource(string(generated.Code))
	if !strings.Contains(
		canonical,
		canonicalSource("Values *map[string]interface{} `field:\"optional\" json:\"values\" yaml:\"values\"`"),
	) {
		t.Fatalf("generated source does not expose untyped Values map")
	}
	if strings.Contains(canonical, "type WebappValues struct") {
		t.Fatalf("schema-less chart unexpectedly generated a typed values DTO")
	}
	if strings.Index(canonical, "ReleaseName *string") > strings.Index(canonical, "Values *map[string]interface{}") {
		t.Fatalf("schema-less optional props are not in upstream alphabetical order")
	}
	compileGenerated(t, generated)
}

func TestGenerateHelmValidation(t *testing.T) {
	t.Run("chart directory is required", func(t *testing.T) {
		if _, err := GenerateHelm("", GenerateOptions{}); err == nil {
			t.Fatal("GenerateHelm() error = nil")
		}
	})
	t.Run("Chart yaml requires name and version", func(t *testing.T) {
		directory := t.TempDir()
		writeImporterTestFile(t, filepath.Join(directory, helmChartFileName), "apiVersion: v2\nname: missing-version\n")
		if _, err := GenerateHelm(directory, GenerateOptions{}); err == nil ||
			!strings.Contains(err.Error(), "name and version are required") {
			t.Fatalf("GenerateHelm() error = %v", err)
		}
	})
	t.Run("values schema must be JSON", func(t *testing.T) {
		directory := createTestHelmChart(t, false)
		writeImporterTestFile(t, filepath.Join(directory, helmValuesFileName), "{")
		if _, err := GenerateHelm(directory, GenerateOptions{}); err == nil ||
			!strings.Contains(err.Error(), "decode Helm values.schema.json") {
			t.Fatalf("GenerateHelm() error = %v", err)
		}
	})
	t.Run("values schema root must be object", func(t *testing.T) {
		directory := createTestHelmChart(t, false)
		writeImporterTestFile(t, filepath.Join(directory, helmValuesFileName), `{"type":"array"}`)
		if _, err := GenerateHelm(directory, GenerateOptions{}); err == nil ||
			!strings.Contains(err.Error(), "root type must be object") {
			t.Fatalf("GenerateHelm() error = %v", err)
		}
	})
}

func TestImportLocalHelmCleansOnlyTargetPackage(t *testing.T) {
	chartDirectory := createTestHelmChart(t, true)
	output := t.TempDir()
	targetDirectory := filepath.Join(output, "custom_webapp")
	if err := os.MkdirAll(filepath.Join(targetDirectory, "jsii"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeImporterTestFile(t, filepath.Join(targetDirectory, "main.go"), "package webapp\n")
	writeImporterTestFile(t, filepath.Join(targetDirectory, "jsii", "proxy.go"), "package jsii\n")
	siblingFile := filepath.Join(output, "sibling", "keep.go")
	if err := os.MkdirAll(filepath.Dir(siblingFile), 0o755); err != nil {
		t.Fatal(err)
	}
	writeImporterTestFile(t, siblingFile, "package sibling\n")

	result, err := Import(context.Background(), Options{
		Source:      "helm:" + chartDirectory,
		OutputDir:   output,
		PackageName: "custom",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Version != "1.2.3" {
		t.Fatalf("Result.Version = %q, want 1.2.3", result.Version)
	}
	if len(result.Packages) != 1 {
		t.Fatalf("Import() produced %d packages, want 1", len(result.Packages))
	}
	generated := result.Packages[0]
	if generated.PackageName != "custom_webapp" ||
		len(generated.Resources) != 1 ||
		generated.Resources[0] != "Webapp" {
		t.Fatalf("generated package = %#v", generated)
	}
	if _, err := os.Stat(filepath.Join(targetDirectory, "generated.go")); err != nil {
		t.Fatalf("generated Helm package was not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDirectory, "main.go")); !os.IsNotExist(err) {
		t.Fatalf("stale upstream file was not removed; stat error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDirectory, "jsii")); !os.IsNotExist(err) {
		t.Fatalf("stale upstream directory was not removed; stat error = %v", err)
	}
	if got := readImporterTestFile(t, siblingFile); got != "package sibling\n" {
		t.Fatalf("sibling package changed: %q", got)
	}
}

func TestImportRemoteHelmDispatchesThroughAcquisition(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake Helm executable is a POSIX shell script")
	}
	executable, argsFile, _ := createFakeHelmPullExecutable(t, false)
	output := t.TempDir()
	result, err := Import(context.Background(), Options{
		Source:         "helm:https://charts.example.test/stable/webapp@1.2.3",
		OutputDir:      output,
		PackageName:    "vendor",
		HelmExecutable: executable,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Version != "1.2.3" || len(result.Packages) != 1 {
		t.Fatalf("Import() result = %#v", result)
	}
	generated := result.Packages[0]
	if generated.PackageName != "vendor_webapp" {
		t.Fatalf("PackageName = %q, want vendor_webapp", generated.PackageName)
	}
	source := canonicalSource(readImporterTestFile(t, generated.File))
	for _, expected := range []string{
		`chart, repo, version := "webapp", "https://charts.example.test/stable", "1.2.3"`,
		"Repo: &repo",
	} {
		if !strings.Contains(source, canonicalSource(expected)) {
			t.Errorf("generated source does not contain %q", expected)
		}
	}
	args := readImporterTestFile(t, argsFile)
	if !strings.Contains(args, "webapp\n--repo\nhttps://charts.example.test/stable\n") {
		t.Fatalf("helm pull args = %q", args)
	}
}

func TestParseLocalHelmSource(t *testing.T) {
	tests := []struct {
		source  string
		path    string
		match   bool
		wantErr bool
	}{
		{source: "k8s"},
		{source: "helm:./chart", path: "./chart", match: true},
		{source: "helm:/charts/webapp", path: "/charts/webapp", match: true},
		{source: "helm:", match: true, wantErr: true},
		{source: "helm:webapp", match: true, wantErr: true},
		{source: "helm:https://charts.example.test/webapp@1.2.3", match: true, wantErr: true},
	}
	for _, test := range tests {
		path, match, err := ParseLocalHelmSource(test.source)
		if (err != nil) != test.wantErr {
			t.Errorf("ParseLocalHelmSource(%q) error = %v, wantErr %t", test.source, err, test.wantErr)
		}
		if path != test.path || match != test.match {
			t.Errorf(
				"ParseLocalHelmSource(%q) = (%q, %t), want (%q, %t)",
				test.source,
				path,
				match,
				test.path,
				test.match,
			)
		}
	}
}

func TestParseHelmSource(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		match     bool
		want      *HelmSource
		errorText string
	}{
		{name: "other importer", source: "k8s"},
		{
			name:   "local relative path is preserved",
			source: "helm:./charts/webapp",
			match:  true,
			want: &HelmSource{
				Kind:      HelmSourceLocal,
				Source:    "helm:./charts/webapp",
				Chart:     "./charts/webapp",
				LocalPath: "./charts/webapp",
			},
		},
		{
			name:   "HTTP repository",
			source: "helm:https://charts.example.test/stable/webapp@1.2.3",
			match:  true,
			want: &HelmSource{
				Kind:         HelmSourceHTTP,
				Source:       "helm:https://charts.example.test/stable/webapp@1.2.3",
				ChartName:    "webapp",
				ChartVersion: "1.2.3",
				Chart:        "webapp",
				Repo:         "https://charts.example.test/stable",
			},
		},
		{
			name:   "OCI registry and full SemVer",
			source: "helm:oci://registry.example.test/team/webapp@v2.3.4-beta.1+build.7",
			match:  true,
			want: &HelmSource{
				Kind:         HelmSourceOCI,
				Source:       "helm:oci://registry.example.test/team/webapp@v2.3.4-beta.1+build.7",
				ChartName:    "webapp",
				ChartVersion: "v2.3.4-beta.1+build.7",
				Chart:        "oci://registry.example.test/team/webapp",
			},
		},
		{name: "empty", source: "helm:", match: true, errorText: "Invalid helm URL"},
		{name: "unqualified chart", source: "helm:webapp", match: true, errorText: "Invalid helm URL"},
		{
			name:      "HTTP missing version",
			source:    "helm:https://charts.example.test/stable/webapp",
			match:     true,
			errorText: "Invalid helm URL",
		},
		{
			name:      "OCI missing version",
			source:    "helm:oci://registry.example.test/team/webapp",
			match:     true,
			errorText: "Invalid helm URL",
		},
		{
			name:      "incomplete SemVer",
			source:    "helm:https://charts.example.test/webapp@1.2",
			match:     true,
			errorText: "Must follow SemVer-2",
		},
		{
			name:      "leading zero",
			source:    "helm:https://charts.example.test/webapp@01.2.3",
			match:     true,
			errorText: "Must follow SemVer-2",
		},
		{
			name:      "numeric prerelease leading zero",
			source:    "helm:https://charts.example.test/webapp@1.2.3-01",
			match:     true,
			errorText: "Must follow SemVer-2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, match, err := ParseHelmSource(test.source)
			if match != test.match {
				t.Fatalf("ParseHelmSource(%q) match = %t, want %t", test.source, match, test.match)
			}
			if test.errorText != "" {
				if err == nil || !strings.Contains(err.Error(), test.errorText) {
					t.Fatalf("ParseHelmSource(%q) error = %v, want containing %q", test.source, err, test.errorText)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ParseHelmSource(%q) = %#v, want %#v", test.source, got, test.want)
			}
		})
	}
}

func TestGenerateHelmSourcePreservesLocalRuntimePath(t *testing.T) {
	chartDirectory := createTestHelmChart(t, true)
	workDirectory := filepath.Dir(chartDirectory)
	runtimePath := "./" + filepath.Base(chartDirectory)
	generated, parsed, err := GenerateHelmSource(
		context.Background(),
		"helm:"+runtimePath,
		workDirectory,
		"",
		GenerateOptions{PackagePrefix: "vendor"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Kind != HelmSourceLocal ||
		parsed.ChartName != "webapp" ||
		parsed.ChartVersion != "1.2.3" {
		t.Fatalf("parsed Helm source = %#v", parsed)
	}
	if generated.PackageName != "vendor_webapp" {
		t.Fatalf("PackageName = %q, want vendor_webapp", generated.PackageName)
	}
	source := canonicalSource(string(generated.Code))
	expected := canonicalSource("chart, version := " + strconv.Quote(runtimePath) + ", \"1.2.3\"")
	if !strings.Contains(source, expected) {
		t.Fatalf("generated source does not preserve runtime chart path %q", runtimePath)
	}
	if strings.Contains(source, canonicalSource(strconv.Quote(chartDirectory))) {
		t.Fatalf("generated source leaked absolute chart path %q", chartDirectory)
	}
}

func TestGenerateHelmSourcePullsRemoteChartAndCleansTemporaryDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake Helm executable is a POSIX shell script")
	}
	tests := []struct {
		name         string
		source       string
		expectedArgs []string
		expectedCode []string
		forbidden    string
	}{
		{
			name:   "HTTP repository",
			source: "helm:https://charts.example.test/stable/webapp@1.2.3",
			expectedArgs: []string{
				"pull",
				"webapp",
				"--repo",
				"https://charts.example.test/stable",
				"--version",
				"1.2.3",
				"--untar",
				"--untardir",
			},
			expectedCode: []string{
				`chart, repo, version := "webapp", "https://charts.example.test/stable", "1.2.3"`,
				"Repo: &repo",
			},
		},
		{
			name:   "OCI registry",
			source: "helm:oci://registry.example.test/team/webapp@v2.3.4-beta.1+build.7",
			expectedArgs: []string{
				"pull",
				"oci://registry.example.test/team/webapp",
				"--version",
				"v2.3.4-beta.1+build.7",
				"--untar",
				"--untardir",
			},
			expectedCode: []string{
				`chart, version := "oci://registry.example.test/team/webapp", "v2.3.4-beta.1+build.7"`,
			},
			forbidden: "Repo: &repo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executable, argsFile, workDirectoryFile := createFakeHelmPullExecutable(t, false)
			generated, parsed, err := GenerateHelmSource(
				context.Background(),
				test.source,
				"",
				executable,
				GenerateOptions{},
			)
			if err != nil {
				t.Fatal(err)
			}
			if parsed.ChartName != "webapp" {
				t.Fatalf("ChartName = %q, want webapp", parsed.ChartName)
			}
			recordedArgs := strings.Split(strings.TrimSpace(readImporterTestFile(t, argsFile)), "\n")
			if len(recordedArgs) != len(test.expectedArgs)+1 {
				t.Fatalf("helm args = %#v, want prefix %#v and a temp directory", recordedArgs, test.expectedArgs)
			}
			if !reflect.DeepEqual(recordedArgs[:len(test.expectedArgs)], test.expectedArgs) {
				t.Fatalf("helm args = %#v, want prefix %#v", recordedArgs, test.expectedArgs)
			}
			if recordedArgs[len(recordedArgs)-1] == "" {
				t.Fatal("helm --untardir argument is empty")
			}
			acquisitionDirectory := strings.TrimSpace(readImporterTestFile(t, workDirectoryFile))
			if _, err := os.Stat(acquisitionDirectory); !os.IsNotExist(err) {
				t.Fatalf("temporary Helm acquisition directory still exists; stat error = %v", err)
			}
			source := canonicalSource(string(generated.Code))
			for _, expected := range test.expectedCode {
				if !strings.Contains(source, canonicalSource(expected)) {
					t.Errorf("generated source does not contain %q", expected)
				}
			}
			if test.forbidden != "" && strings.Contains(source, canonicalSource(test.forbidden)) {
				t.Errorf("generated source unexpectedly contains %q", test.forbidden)
			}
			compileGenerated(t, generated)
		})
	}
}

func TestGenerateHelmSourceCleansTemporaryDirectoryAfterPullFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake Helm executable is a POSIX shell script")
	}
	executable, _, workDirectoryFile := createFakeHelmPullExecutable(t, true)
	_, _, err := GenerateHelmSource(
		context.Background(),
		"helm:https://charts.example.test/stable/webapp@1.2.3",
		"",
		executable,
		GenerateOptions{},
	)
	if err == nil || !strings.Contains(err.Error(), "fake pull failed") {
		t.Fatalf("GenerateHelmSource() error = %v", err)
	}
	acquisitionDirectory := strings.TrimSpace(readImporterTestFile(t, workDirectoryFile))
	if _, err := os.Stat(acquisitionDirectory); !os.IsNotExist(err) {
		t.Fatalf("failed pull left temporary Helm directory behind; stat error = %v", err)
	}
}

func TestHelmConstructTypeNameMatchesUpstreamNormalization(t *testing.T) {
	tests := map[string]string{
		"webapp":       "Webapp",
		"web-app":      "Webapp",
		"wordpress_ha": "Wordpressha",
		"MY chart":     "MyChart",
	}
	for input, expected := range tests {
		if actual := helmConstructTypeName(input); actual != expected {
			t.Errorf("helmConstructTypeName(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func createFakeHelmPullExecutable(t *testing.T, fail bool) (executable, argsFile, workDirectoryFile string) {
	t.Helper()
	directory := t.TempDir()
	executable = filepath.Join(directory, "helm")
	argsFile = filepath.Join(directory, "args")
	workDirectoryFile = filepath.Join(directory, "workdir")
	status := ""
	if fail {
		status = `printf '%s\n' "fake pull failed" >&2
exit 7
`
	}
	script := `#!/bin/sh
set -eu
base=$(dirname "$0")
: > "$base/args"
for arg in "$@"; do
	printf '%s\n' "$arg" >> "$base/args"
done
target=
version=
untardir=
while [ "$#" -gt 0 ]; do
	case "$1" in
		pull|--untar)
			shift
			;;
		--repo)
			shift 2
			;;
		--version)
			version="$2"
			shift 2
			;;
		--untardir)
			untardir="$2"
			shift 2
			;;
		*)
			if [ -z "$target" ]; then
				target="$1"
			fi
			shift
			;;
	esac
done
printf '%s' "$untardir" > "$base/workdir"
` + status + `name=${target##*/}
mkdir -p "$untardir/$name"
printf '%s\n' \
	'apiVersion: v2' \
	"name: $name" \
	"version: $version" > "$untardir/$name/Chart.yaml"
printf '%s\n' '{"type":"object","properties":{"replicas":{"type":"integer"}}}' > "$untardir/$name/values.schema.json"
`
	if err := os.WriteFile(executable, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return executable, argsFile, workDirectoryFile
}

func createTestHelmChart(t *testing.T, withSchema bool) string {
	t.Helper()
	directory := t.TempDir()
	writeImporterTestFile(t, filepath.Join(directory, helmChartFileName), `apiVersion: v2
name: webapp
version: 1.2.3
dependencies:
  - name: redis
    version: 20.0.0
    repository: https://charts.example.test
`)
	if err := os.MkdirAll(filepath.Join(directory, "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeImporterTestFile(t, filepath.Join(directory, "templates", "config-map.yaml"), `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}
data:
  repository: {{ .Values.image.repository | quote }}
`)
	if withSchema {
		writeImporterTestFile(t, filepath.Join(directory, helmValuesFileName), `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["replicaCount"],
  "properties": {
    "replicaCount": {
      "type": "integer"
    },
    "mode": {
      "type": "string",
      "enum": ["dev", "prod"]
    },
    "image": {
      "type": "object",
      "properties": {
        "repository": {
          "type": "string"
        },
        "pullPolicy": {
          "type": "string",
          "enum": ["Always", "IfNotPresent"]
        }
      }
    },
    "workers": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          }
        }
      }
    }
  }
}`)
	}
	return directory
}

func writeImporterTestFile(t *testing.T, filename, contents string) {
	t.Helper()
	if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", filename, err)
	}
}

func readImporterTestFile(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	return string(data)
}

const generatedHelmSynthesisTest = `package webapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

type customWebapp struct {
	Webapp
}

func TestGeneratedHelmSynthesis(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "helm")
	script := ` + "`" + `#!/bin/sh
set -eu
values=
version=
chart=
while [ "$#" -gt 0 ]; do
	case "$1" in
		-f)
			values="$2"
			shift 2
			;;
		--version)
			version="$2"
			shift 2
			;;
		*)
			chart="$1"
			shift
			;;
	esac
done
test "$version" = "1.2.3"
test -f "$chart/Chart.yaml"
test -n "$values"
grep -q "replicaCount: 3" "$values"
grep -q "repository: nginx" "$values"
grep -q "mode: prod" "$values"
grep -q "customNested: nested" "$values"
grep -q "customTop: top" "$values"
if grep -q "additionalValues" "$values"; then
	exit 42
fi
printf '%s\n' \
	'apiVersion: v1' \
	'kind: ConfigMap' \
	'metadata:' \
	'  name: rendered' \
	'data:' \
	'  chart: webapp'
` + "`" + `
	if err := os.WriteFile(executable, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	outdir := filepath.Join(t.TempDir(), "dist")
	app := cdk8s.NewApp(&cdk8s.AppProps{Outdir: &outdir})
	chartID := "chart"
	chart := cdk8s.NewChart(app, &chartID, nil)
	replicas := 3.0
	repository := "nginx"
	nestedAdditional := map[string]interface{}{"customNested": "nested"}
	topAdditional := map[string]interface{}{"customTop": "top"}
	values := &WebappValues{
		ReplicaCount: &replicas,
		Mode: WebappMode_PROD,
		Image: &WebappImage{
			Repository: &repository,
			AdditionalValues: &nestedAdditional,
		},
		AdditionalValues: &topAdditional,
	}
	id := "webapp"
	generated := NewWebapp(chart, &id, &WebappProps{
		Values: values,
		HelmExecutable: &executable,
	})
	if generated.Helm() == nil {
		t.Fatal("Helm() = nil")
	}
	if isConstruct := Webapp_IsConstruct(generated); isConstruct == nil || !*isConstruct {
		t.Fatal("Webapp_IsConstruct() = false")
	}
	override := &customWebapp{}
	overrideID := "override"
	NewWebapp_Override(override, chart, &overrideID, &WebappProps{
		Values: values,
		HelmExecutable: &executable,
	})
	if override.Helm() == nil {
		t.Fatal("override Helm() = nil")
	}

	app.Synth()
	files, err := filepath.Glob(filepath.Join(outdir, "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("synthesis wrote %d yaml files, want 1: %v", len(files), files)
	}
	manifest, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), "kind: ConfigMap") ||
		!strings.Contains(string(manifest), "chart: webapp") {
		t.Fatalf("unexpected synthesized manifest:\n%s", manifest)
	}
}
`
