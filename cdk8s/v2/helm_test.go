package cdk8s_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

func helmString(value string) *string { return &value }

type helmHarness struct {
	chartPath string
	argsLog   string
	callsLog  string
	valuesLog string
	output    string
}

func helmNewHarness(t *testing.T) helmHarness {
	t.Helper()
	directory := t.TempDir()
	executable := filepath.Join(directory, "helm")
	argsLog := filepath.Join(directory, "args.log")
	callsLog := filepath.Join(directory, "calls.log")
	valuesLog := filepath.Join(directory, "values.yaml")
	output := filepath.Join(directory, "output.yaml")
	chartPath := filepath.Join(directory, "helm-sample")
	if err := os.Mkdir(chartPath, 0o755); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
printf 'call\n' >> "$HELM_PORT_CALLS_LOG"
: > "$HELM_PORT_ARGS_LOG"
previous=
for argument in "$@"; do
  printf '%s\n' "$argument" >> "$HELM_PORT_ARGS_LOG"
  if [ "$previous" = "-f" ] && [ -n "$HELM_PORT_VALUES_LOG" ]; then
    cp "$argument" "$HELM_PORT_VALUES_LOG"
  fi
  previous="$argument"
done
if [ -n "$HELM_PORT_ERROR" ]; then
  printf '%s\n' "$HELM_PORT_ERROR" >&2
  exit 1
fi
if [ -n "$HELM_PORT_STDERR_OUTPUT" ]; then
  cat "$HELM_PORT_STDERR_OUTPUT" >&2
  exit 1
fi
if [ -n "$HELM_PORT_OUTPUT" ]; then
  cat "$HELM_PORT_OUTPUT"
fi
`
	if err := os.WriteFile(executable, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HELM_PORT_ARGS_LOG", argsLog)
	t.Setenv("HELM_PORT_CALLS_LOG", callsLog)
	t.Setenv("HELM_PORT_VALUES_LOG", valuesLog)
	t.Setenv("HELM_PORT_OUTPUT", "")
	t.Setenv("HELM_PORT_ERROR", "")
	t.Setenv("HELM_PORT_STDERR_OUTPUT", "")
	return helmHarness{
		chartPath: chartPath,
		argsLog:   argsLog,
		callsLog:  callsLog,
		valuesLog: valuesLog,
		output:    output,
	}
}

func helmCallCount(t *testing.T, harness helmHarness) int {
	t.Helper()
	data, err := os.ReadFile(harness.callsLog)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Count(string(data), "call\n")
}

func helmSetOutput(t *testing.T, harness helmHarness, objects []interface{}) {
	t.Helper()
	var documents strings.Builder
	for index, object := range objects {
		if index > 0 {
			documents.WriteString("\n---\n")
		}
		data, err := json.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		documents.Write(data)
	}
	documents.WriteByte('\n')
	if err := os.WriteFile(harness.output, []byte(documents.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HELM_PORT_OUTPUT", harness.output)
}

func helmArgs(t *testing.T, harness helmHarness) []string {
	t.Helper()
	data, err := os.ReadFile(harness.argsLog)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.TrimSuffix(string(data), "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

func helmExpectedObjects(release string, replicas int, nodeSelector map[string]interface{}) []interface{} {
	name := release + "-helm-sample"
	labels := func() map[string]interface{} {
		return map[string]interface{}{
			"app.kubernetes.io/instance":   release,
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/name":       "helm-sample",
			"app.kubernetes.io/version":    "1.16.0",
			"helm.sh/chart":                "helm-sample-0.1.0",
		}
	}
	selectorLabels := func() map[string]interface{} {
		return map[string]interface{}{
			"app.kubernetes.io/instance": release,
			"app.kubernetes.io/name":     "helm-sample",
		}
	}
	serviceAccount := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ServiceAccount",
		"metadata": map[string]interface{}{
			"labels": labels(),
			"name":   name,
		},
	}
	service := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"labels": labels(),
			"name":   name,
		},
		"spec": map[string]interface{}{
			"ports": []interface{}{map[string]interface{}{
				"name":       "http",
				"port":       80,
				"protocol":   "TCP",
				"targetPort": "http",
			}},
			"selector": selectorLabels(),
			"type":     "ClusterIP",
		},
	}
	podSpec := map[string]interface{}{
		"containers": []interface{}{map[string]interface{}{
			"image":           "nginx:1.16.0",
			"imagePullPolicy": "IfNotPresent",
			"livenessProbe": map[string]interface{}{
				"httpGet": map[string]interface{}{"path": "/", "port": "http"},
			},
			"name": "helm-sample",
			"ports": []interface{}{map[string]interface{}{
				"containerPort": 80,
				"name":          "http",
				"protocol":      "TCP",
			}},
			"readinessProbe": map[string]interface{}{
				"httpGet": map[string]interface{}{"path": "/", "port": "http"},
			},
			"resources":       map[string]interface{}{},
			"securityContext": map[string]interface{}{},
		}},
		"securityContext":    map[string]interface{}{},
		"serviceAccountName": name,
	}
	if nodeSelector != nil {
		podSpec["nodeSelector"] = nodeSelector
	}
	deployment := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"labels": labels(),
			"name":   name,
		},
		"spec": map[string]interface{}{
			"replicas": replicas,
			"selector": map[string]interface{}{
				"matchLabels": selectorLabels(),
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{"labels": selectorLabels()},
				"spec":     podSpec,
			},
		},
	}
	return []interface{}{serviceAccount, service, deployment}
}

func helmAssertJSONEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	normalize := func(value interface{}) interface{} {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal value: %v", err)
		}
		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("unmarshal value: %v", err)
		}
		return result
	}
	gotNormalized, wantNormalized := normalize(got), normalize(want)
	if !reflect.DeepEqual(gotNormalized, wantNormalized) {
		gotJSON, _ := json.MarshalIndent(gotNormalized, "", "  ")
		wantJSON, _ := json.MarshalIndent(wantNormalized, "", "  ")
		t.Fatalf("value mismatch\n--- got ---\n%s\n--- want ---\n%s", gotJSON, wantJSON)
	}
}

func helmRequirePanicContains(t *testing.T, want string, callback func()) {
	t.Helper()
	defer func() {
		panicValue := recover()
		if panicValue == nil {
			t.Fatalf("expected panic containing %q", want)
		}
		if got := fmt.Sprint(panicValue); !strings.Contains(got, want) {
			t.Fatalf("panic = %q, want it to contain %q", got, want)
		}
	}()
	callback()
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/helm.test.ts#L9
func TestHelmBasicUsage(t *testing.T) {
	harness := helmNewHarness(t)
	want := helmExpectedObjects("test-sample-c8e2763d", 1, nil)
	helmSetOutput(t, harness, want)
	chart := cdk8s.Testing_Chart()
	helm := cdk8s.NewHelm(chart, helmString("sample"), &cdk8s.HelmProps{
		Chart: &harness.chartPath,
	})

	if got, expected := *helm.ReleaseName(), "test-sample-c8e2763d"; got != expected {
		t.Fatalf("release name = %q, want %q", got, expected)
	}
	helmAssertJSONEqual(t, *cdk8s.Testing_Synth(chart), want)
	wantArgs := []string{"template", "test-sample-c8e2763d", harness.chartPath}
	if got := helmArgs(t, harness); !reflect.DeepEqual(got, wantArgs) {
		t.Fatalf("helm arguments = %#v, want %#v", got, wantArgs)
	}
	if got := helmCallCount(t, harness); got != 1 {
		t.Fatalf("helm invocation count = %d, want 1", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/helm.test.ts#L23
func TestHelmFailsIfExecutableIsNotFound(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	chartPath := filepath.Join(t.TempDir(), "helm-sample")
	missingExecutable := filepath.Join(t.TempDir(), "helm-port-does-not-exist")

	helmRequirePanicContains(t, "unable to execute '"+missingExecutable+"' to render Helm chart", func() {
		cdk8s.NewHelm(chart, helmString("sample"), &cdk8s.HelmProps{
			Chart:          &chartPath,
			HelmExecutable: &missingExecutable,
		})
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/helm.test.ts#L35
func TestHelmValuesCanBeSpecified(t *testing.T) {
	harness := helmNewHarness(t)
	want := helmExpectedObjects("test-sample-c8e2763d", 889, map[string]interface{}{
		"selectMe": "boomboom",
	})
	helmSetOutput(t, harness, want)
	chart := cdk8s.Testing_Chart()
	values := map[string]interface{}{
		"replicaCount": 889,
		"ingress":      map[string]interface{}{"enabled": false},
		"nodeSelector": map[string]interface{}{"selectMe": "boomboom"},
	}
	cdk8s.NewHelm(chart, helmString("sample"), &cdk8s.HelmProps{
		Chart:  &harness.chartPath,
		Values: &values,
	})

	helmAssertJSONEqual(t, *cdk8s.Testing_Synth(chart), want)
	loadedValues := cdk8s.Yaml_Load(&harness.valuesLog)
	helmAssertJSONEqual(t, *loadedValues, []interface{}{values})
	args := helmArgs(t, harness)
	if len(args) != 5 || args[0] != "template" || args[1] != "-f" || args[3] != "test-sample-c8e2763d" || args[4] != harness.chartPath {
		t.Fatalf("helm arguments = %#v, want template -f <values> release chart", args)
	}
	if got := helmCallCount(t, harness); got != 1 {
		t.Fatalf("helm invocation count = %d, want 1", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/helm.test.ts#L57
func TestHelmReleaseNameCanBeSpecified(t *testing.T) {
	harness := helmNewHarness(t)
	want := helmExpectedObjects("your-release", 1, nil)
	helmSetOutput(t, harness, want)
	chart := cdk8s.Testing_Chart()
	releaseName := "your-release"
	helm := cdk8s.NewHelm(chart, helmString("sample"), &cdk8s.HelmProps{
		Chart:       &harness.chartPath,
		ReleaseName: &releaseName,
	})

	if got := *helm.ReleaseName(); got != releaseName {
		t.Fatalf("release name = %q, want %q", got, releaseName)
	}
	objects := *cdk8s.Testing_Synth(chart)
	for _, object := range objects {
		manifest := object.(map[string]interface{})
		metadata := manifest["metadata"].(map[string]interface{})
		name := metadata["name"].(string)
		if !strings.HasPrefix(name, "your-release-") {
			t.Errorf("resource name %q does not start with your-release-", name)
		}
	}
	helmAssertJSONEqual(t, objects, want)
	wantArgs := []string{"template", releaseName, harness.chartPath}
	if got := helmArgs(t, harness); !reflect.DeepEqual(got, wantArgs) {
		t.Fatalf("helm arguments = %#v, want %#v", got, wantArgs)
	}
	if got := helmCallCount(t, harness); got != 1 {
		t.Fatalf("helm invocation count = %d, want 1", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/helm.test.ts#L75
func TestHelmCanInteractWithAPIObjects(t *testing.T) {
	harness := helmNewHarness(t)
	want := helmExpectedObjects("test-sample-c8e2763d", 1, nil)
	helmSetOutput(t, harness, want)
	chart := cdk8s.Testing_Chart()
	helm := cdk8s.NewHelm(chart, helmString("sample"), &cdk8s.HelmProps{
		Chart: &harness.chartPath,
	})

	var serviceAccount cdk8s.ApiObject
	kindsAndNames := make([]string, 0, len(*helm.ApiObjects()))
	for _, object := range *helm.ApiObjects() {
		kindsAndNames = append(kindsAndNames, *object.Kind()+"/"+*object.Name())
		if *object.Kind() == "ServiceAccount" && *object.Name() == "test-sample-c8e2763d-helm-sample" {
			serviceAccount = object
		}
	}
	if serviceAccount == nil {
		t.Fatal("ServiceAccount API object was not found")
	}
	serviceAccount.Metadata().AddAnnotation(helmString("my.annotation"), helmString("hey-there"))
	sort.Strings(kindsAndNames)
	wantKindsAndNames := []string{
		"Deployment/test-sample-c8e2763d-helm-sample",
		"Service/test-sample-c8e2763d-helm-sample",
		"ServiceAccount/test-sample-c8e2763d-helm-sample",
	}
	if !reflect.DeepEqual(kindsAndNames, wantKindsAndNames) {
		t.Fatalf("API objects = %#v, want %#v", kindsAndNames, wantKindsAndNames)
	}
	manifest := serviceAccount.ToJson().(map[string]interface{})
	metadata := manifest["metadata"].(map[string]interface{})
	helmAssertJSONEqual(t, metadata["annotations"], map[string]interface{}{
		"my.annotation": "hey-there",
	})
	wantArgs := []string{"template", "test-sample-c8e2763d", harness.chartPath}
	if got := helmArgs(t, harness); !reflect.DeepEqual(got, wantArgs) {
		t.Fatalf("helm arguments = %#v, want %#v", got, wantArgs)
	}
	if got := helmCallCount(t, harness); got != 1 {
		t.Fatalf("helm invocation count = %d, want 1", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/helm.test.ts#L99
func TestHelmFlagsSpecifyAdditionalOptions(t *testing.T) {
	harness := helmNewHarness(t)
	chart := cdk8s.Testing_Chart()
	flags := []*string{
		helmString("--description"),
		helmString("my custom description"),
		helmString("--no-hooks"),
	}
	cdk8s.NewHelm(chart, helmString("sample"), &cdk8s.HelmProps{
		Chart:     &harness.chartPath,
		HelmFlags: &flags,
	})

	want := []string{
		"template",
		"--description", "my custom description",
		"--no-hooks",
		"test-sample-c8e2763d",
		harness.chartPath,
	}
	if got := helmArgs(t, harness); !reflect.DeepEqual(got, want) {
		t.Fatalf("helm arguments = %#v, want %#v", got, want)
	}
	if got := helmCallCount(t, harness); got != 1 {
		t.Fatalf("helm invocation count = %d, want 1", got)
	}

	// The upstream mock also verifies its 10 MiB maxBuffer option. Exercise
	// the exact boundary and both observable overflow paths through the public
	// constructor.
	exactLimitHarness := helmNewHarness(t)
	exactLimitOutput := bytes.Repeat([]byte{'x'}, 10*1024*1024)
	exactLimitOutput[0] = '#'
	exactLimitOutput[len(exactLimitOutput)-1] = '\n'
	if err := os.WriteFile(exactLimitHarness.output, exactLimitOutput, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HELM_PORT_OUTPUT", exactLimitHarness.output)
	exactLimitChart := cdk8s.Testing_Chart()
	cdk8s.NewHelm(exactLimitChart, helmString("sample"), &cdk8s.HelmProps{
		Chart:     &exactLimitHarness.chartPath,
		HelmFlags: &flags,
	})
	if got := helmCallCount(t, exactLimitHarness); got != 1 {
		t.Fatalf("exact-limit helm invocation count = %d, want 1", got)
	}

	limitHarness := helmNewHarness(t)
	oversizedOutput := make([]byte, 10*1024*1024+1)
	if err := os.WriteFile(limitHarness.output, oversizedOutput, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HELM_PORT_OUTPUT", limitHarness.output)
	bufferChart := cdk8s.Testing_Chart()
	helmRequirePanicContains(t, "stdout maxBuffer length exceeded", func() {
		cdk8s.NewHelm(bufferChart, helmString("sample"), &cdk8s.HelmProps{
			Chart:     &limitHarness.chartPath,
			HelmFlags: &flags,
		})
	})
	if got := helmCallCount(t, limitHarness); got != 1 {
		t.Fatalf("stdout-limit helm invocation count = %d, want 1", got)
	}

	stderrHarness := helmNewHarness(t)
	if err := os.WriteFile(stderrHarness.output, oversizedOutput, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HELM_PORT_STDERR_OUTPUT", stderrHarness.output)
	stderrChart := cdk8s.Testing_Chart()
	helmRequirePanicContains(t, "stderr maxBuffer length exceeded", func() {
		cdk8s.NewHelm(stderrChart, helmString("sample"), &cdk8s.HelmProps{
			Chart:     &stderrHarness.chartPath,
			HelmFlags: &flags,
		})
	})
	if got := helmCallCount(t, stderrHarness); got != 1 {
		t.Fatalf("stderr-limit helm invocation count = %d, want 1", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/helm.test.ts#L134
func TestHelmRepoCanBeSpecified(t *testing.T) {
	harness := helmNewHarness(t)
	chart := cdk8s.Testing_Chart()
	repo, version, namespace := "foo-repo", "foo-version", "foo-namespace"
	cdk8s.NewHelm(chart, helmString("sample"), &cdk8s.HelmProps{
		Chart:     &harness.chartPath,
		Repo:      &repo,
		Version:   &version,
		Namespace: &namespace,
	})

	want := []string{
		"template",
		"--repo", "foo-repo",
		"--version", "foo-version",
		"--namespace", "foo-namespace",
		"test-sample-c8e2763d",
		harness.chartPath,
	}
	if got := helmArgs(t, harness); !reflect.DeepEqual(got, want) {
		t.Fatalf("helm arguments = %#v, want %#v", got, want)
	}
	if got := helmCallCount(t, harness); got != 1 {
		t.Fatalf("helm invocation count = %d, want 1", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/helm.test.ts#L169
func TestHelmPropagatesHelmFailures(t *testing.T) {
	harness := helmNewHarness(t)
	t.Setenv("HELM_PORT_ERROR", "Error: unknown flag: --invalid-argument-not-found-boom-boom")
	chart := cdk8s.Testing_Chart()
	flags := []*string{helmString("--invalid-argument-not-found-boom-boom")}

	helmRequirePanicContains(t, "unknown flag", func() {
		cdk8s.NewHelm(chart, helmString("my-chart"), &cdk8s.HelmProps{
			Chart:     &harness.chartPath,
			HelmFlags: &flags,
		})
	})
	wantArgs := []string{
		"template",
		"--invalid-argument-not-found-boom-boom",
		"test-my-chart-c8e2763d",
		harness.chartPath,
	}
	if got := helmArgs(t, harness); !reflect.DeepEqual(got, wantArgs) {
		t.Fatalf("helm arguments = %#v, want %#v", got, wantArgs)
	}
	if got := helmCallCount(t, harness); got != 1 {
		t.Fatalf("helm invocation count = %d, want 1", got)
	}
}
