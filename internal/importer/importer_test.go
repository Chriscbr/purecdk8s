package importer

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestParseKubernetesSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source  string
		version string
		match   bool
		wantErr bool
	}{
		{source: "k8s", version: DefaultKubernetesVersion, match: true},
		{source: "k8s@1.25.0", version: "1.25.0", match: true},
		{source: "./crd.yaml"},
		{source: "k8s@1.25", match: true, wantErr: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.source, func(t *testing.T) {
			version, match, err := ParseKubernetesSource(test.source)
			if (err != nil) != test.wantErr {
				t.Fatalf("ParseKubernetesSource() error = %v, wantErr %v", err, test.wantErr)
			}
			if version != test.version || match != test.match {
				t.Fatalf("ParseKubernetesSource() = (%q, %v), want (%q, %v)", version, match, test.version, test.match)
			}
		})
	}
}

func TestNormalizeGitHubSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source   string
		expected string
		match    bool
	}{
		{source: "github:crossplane/crossplane", expected: "https://doc.crds.dev/raw/github.com/crossplane/crossplane", match: true},
		{source: "github:crossplane/crossplane@0.14", expected: "https://doc.crds.dev/raw/github.com/crossplane/crossplane@v0.14.0", match: true},
		{source: "github:crossplane/crossplane@0.14.2", expected: "https://doc.crds.dev/raw/github.com/crossplane/crossplane@v0.14.2", match: true},
		{source: "github:account/repo@1"},
		{source: "gitlab:account/repo@1.2.3"},
	}
	for _, test := range tests {
		actual, match := NormalizeGitHubSource(test.source)
		if actual != test.expected || match != test.match {
			t.Errorf("NormalizeGitHubSource(%q) = (%q, %v), want (%q, %v)", test.source, actual, match, test.expected, test.match)
		}
	}
}

func TestGenerateKubernetesCompatibility(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("testdata", "k8s-definitions.json"))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := GenerateKubernetes(data, GenerateOptions{PackageName: "k8s"})
	if err != nil {
		t.Fatal(err)
	}
	source := string(generated.Code)
	canonical := canonicalSource(source)
	for _, expected := range []string{
		"type DeploymentSpec struct",
		"HostIp *string `field:\"optional\" json:\"hostIp\" yaml:\"hostIp\" k8s:\"hostIP\"`",
		"Limits *map[string]Quantity `field:\"optional\" json:\"limits\" yaml:\"limits\"`",
		"Selector *string `field:\"required\" json:\"selector\" yaml:\"selector\"`",
		"TargetPort IntOrString `field:\"optional\" json:\"targetPort\" yaml:\"targetPort\"`",
		"type Recursive struct",
		"Next **Recursive `field:\"optional\" json:\"next\" yaml:\"next\"`",
		"type IoK8SApimachineryPkgApisMetaV1DeleteOptionsKind string",
		"IoK8SApimachineryPkgApisMetaV1DeleteOptionsKind_DELETE_OPTIONS IoK8SApimachineryPkgApisMetaV1DeleteOptionsKind = \"DELETE_OPTIONS\"",
		"purecdk8sserialization.RegisterEnumWireValues(map[IoK8SApimachineryPkgApisMetaV1DeleteOptionsKind]interface{}",
		"type KubeDeployment interface",
		"func NewKubeDeployment(scope constructs.Construct, id *string, props *KubeDeploymentProps) KubeDeployment",
		"func KubeDeployment_Manifest(props *KubeDeploymentProps) interface{}",
		"func KubeDeployment_GVK() *cdk8s.GroupVersionKind",
		"Items *[]*KubeDeploymentProps `field:\"required\" json:\"items\" yaml:\"items\"`",
		"func Quantity_FromNumber(value *float64) Quantity",
		"func Quantity_FromString(value *string) Quantity",
	} {
		if !strings.Contains(canonical, canonicalSource(expected)) {
			t.Errorf("generated source does not contain %q", expected)
		}
	}
	if strings.Contains(source, "PureCDK8sEnumValue") {
		t.Fatal("generated enum has an extra exported serialization method")
	}
	orderingBody := between(source, "type Ordering struct {", "\n}")
	lastIndex := -1
	for _, field := range []string{"Middle ", "Zulu ", "Alpha ", "Beta "} {
		index := strings.Index(orderingBody, field)
		if index < 0 {
			t.Fatalf("Ordering is missing field %q:\n%s", field, orderingBody)
		}
		if index <= lastIndex {
			t.Fatalf("Ordering fields are not required-first/alphabetical:\n%s", orderingBody)
		}
		lastIndex = index
	}
	props := between(source, "type KubeDeploymentProps struct {", "\n}")
	if strings.Contains(props, "Status ") || strings.Contains(props, "ApiVersion ") || strings.Contains(props, "Kind ") {
		t.Errorf("resource props retained server/GVK fields:\n%s", props)
	}
	compileGenerated(t, generated, `package k8s

import (
	"reflect"
	"testing"

	cdk8s "github.com/purecdk8s/purecdk8s/cdk8s/v2"
)

func TestGeneratedRequiredValidation(t *testing.T) {
	if got := string(IoK8SApimachineryPkgApisMetaV1DeleteOptionsKind_DELETE_OPTIONS); got != "DELETE_OPTIONS" {
		t.Fatalf("public enum value = %q", got)
	}
	if methods := reflect.TypeOf(IoK8SApimachineryPkgApisMetaV1DeleteOptionsKind("")).NumMethod(); methods != 0 {
		t.Fatalf("generated enum has %d methods, want upstream-compatible zero", methods)
	}
	ordering := reflect.TypeOf(Ordering{})
	for index, want := range []string{"Middle", "Zulu", "Alpha", "Beta"} {
		if got := ordering.Field(index).Name; got != want {
			t.Fatalf("Ordering field %d = %q, want %q", index, got, want)
		}
	}
	selector := "app=test"
	props := &KubeDeploymentProps{Spec: &DeploymentSpec{
		Selector: &selector,
		DeleteOptions: &DeleteOptions{
			Kind: IoK8SApimachineryPkgApisMetaV1DeleteOptionsKind_DELETE_OPTIONS,
		},
	}}
	staticManifest := KubeDeployment_Manifest(props).(map[string]interface{})
	staticKind := staticManifest["spec"].(map[string]interface{})["deleteOptions"].(map[string]interface{})["kind"]
	if staticKind != "DeleteOptions" {
		t.Fatalf("static manifest enum = %#v", staticKind)
	}
	app := cdk8s.NewApp(nil)
	chartID, resourceID := "chart", "resource"
	chart := cdk8s.NewChart(app, &chartID, nil)
	resource := NewKubeDeployment(chart, &resourceID, props)
	resourceManifest := resource.ToJson().(map[string]interface{})
	resourceKind := resourceManifest["spec"].(map[string]interface{})["deleteOptions"].(map[string]interface{})["kind"]
	if resourceKind != "DeleteOptions" {
		t.Fatalf("resource manifest enum = %#v", resourceKind)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("missing required selector did not panic")
		}
	}()
	KubeDeployment_Manifest(&KubeDeploymentProps{Spec: &DeploymentSpec{}})
}
`)
}

func TestGenerateKubernetesWithoutClassPrefix(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("testdata", "k8s-definitions.json"))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := GenerateKubernetes(data, GenerateOptions{
		PackageName:            "k8s",
		DisableClassNamePrefix: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	source := string(generated.Code)
	if !strings.Contains(source, "type Deployment interface") || !strings.Contains(source, "func NewDeployment(") {
		t.Fatalf("generated source did not omit the Kube prefix")
	}
	if strings.Contains(source, "type KubeDeployment interface") {
		t.Fatalf("generated source retained the Kube prefix")
	}
}

func TestGenerateKubernetesExcludesRawReferences(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("testdata", "k8s-definitions.json"))
	if err != nil {
		t.Fatal(err)
	}
	notRaw, err := GenerateKubernetes(data, GenerateOptions{
		PackageName: "k8s",
		Excludes: []string{
			`^io\.k8s\.apimachinery\.pkg\.apis\.meta\.v1\.ObjectMeta$`,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(notRaw.Code), "type ObjectMeta struct") {
		t.Fatal("exclude pattern unexpectedly matched the dereferenced FQN instead of the raw $ref")
	}

	generated, err := GenerateKubernetes(data, GenerateOptions{
		PackageName: "k8s",
		Excludes: []string{
			`^#/definitions/io\.k8s\.apimachinery\.pkg\.apis\.meta\.v1\.ObjectMeta$`,
			`Recursive$`,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	source := string(generated.Code)
	for _, excludedType := range []string{"type ObjectMeta struct", "type Recursive struct"} {
		if strings.Contains(source, excludedType) {
			t.Errorf("generated source retained excluded definition %q", excludedType)
		}
	}
	canonical := canonicalSource(source)
	for _, expected := range []string{
		"Metadata interface{} `field:\"optional\" json:\"metadata\" yaml:\"metadata\"`",
		"Recursive interface{} `field:\"optional\" json:\"recursive\" yaml:\"recursive\"`",
	} {
		if !strings.Contains(canonical, canonicalSource(expected)) {
			t.Errorf("excluded reference was not emitted as interface{}: %q", expected)
		}
	}
	compileGenerated(t, generated)
}

func TestGenerateKubernetesRejectsInvalidExcludeRegex(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("testdata", "k8s-definitions.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = GenerateKubernetes(data, GenerateOptions{PackageName: "k8s", Excludes: []string{"["}})
	if err == nil || !strings.Contains(err.Error(), `invalid Kubernetes exclude regular expression "["`) {
		t.Fatalf("invalid exclude error = %v", err)
	}
}

func TestGenerateCRD(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("testdata", "widgets-crd.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := GenerateCRDs(data, GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(generated) != 1 {
		t.Fatalf("GenerateCRDs() produced %d packages, want 1", len(generated))
	}
	source := string(generated[0].Code)
	canonical := canonicalSource(source)
	for _, expected := range []string{
		"package exampletest",
		"type WidgetProps struct",
		"Metadata *cdk8s.ApiObjectMetadata `field:\"optional\" json:\"metadata\" yaml:\"metadata\"`",
		"type WidgetSpec struct",
		"Image *string `field:\"required\" json:\"image\" yaml:\"image\"`",
		"PodIp *string `field:\"optional\" json:\"podIp\" yaml:\"podIp\" k8s:\"podIP\"`",
		"type WidgetV2Alpha1Props struct",
		"func Widget_GVK() *cdk8s.GroupVersionKind",
		"func WidgetV2Alpha1_GVK() *cdk8s.GroupVersionKind",
	} {
		if !strings.Contains(canonical, canonicalSource(expected)) {
			t.Errorf("generated CRD source does not contain %q", expected)
		}
	}
	compileGenerated(t, generated[0])
	if _, err := GenerateCRDs(data, GenerateOptions{Excludes: []string{"["}}); err != nil {
		t.Fatalf("CRD generation should ignore Kubernetes-only exclude patterns: %v", err)
	}
}

func TestGenerateCRDPackageNaming(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("testdata", "widgets-crd.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	exact, err := GenerateCRDs(data, GenerateOptions{PackageName: "exact_name"})
	if err != nil {
		t.Fatal(err)
	}
	if len(exact) != 1 || exact[0].PackageName != "exact_name" {
		t.Fatalf("exact generator package = %#v", exact)
	}

	second := strings.ReplaceAll(string(data), "example.test", "other_test.example")
	second = strings.ReplaceAll(second, "Widget", "Gadget")
	second = strings.ReplaceAll(second, "widget", "gadget")
	multiple, err := GenerateCRDs(append(append(append([]byte(nil), data...), []byte("\n---\n")...), []byte(second)...), GenerateOptions{
		PackagePrefix: "my_widgets",
	})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, generated := range multiple {
		names = append(names, generated.PackageName)
	}
	if got := strings.Join(names, ","); got != "my_widgets_exampletest,my_widgets_othertestexample" {
		t.Fatalf("prefixed multi-group package names = %q", got)
	}
}

func TestImportLocalCRD(t *testing.T) {
	t.Parallel()
	output := t.TempDir()
	staleDirectory := filepath.Join(output, "widgets_exampletest")
	if err := os.MkdirAll(staleDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	staleFile := filepath.Join(staleDirectory, "zz_jsii_proxy.go")
	if err := os.WriteFile(staleFile, []byte("package exampletest\n\nimport _ \"github.com/aws/jsii-runtime-go\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Import(context.Background(), Options{
		Source:      filepath.Join("testdata", "widgets-crd.yaml"),
		OutputDir:   output,
		PackageName: "widgets",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Packages) != 1 {
		t.Fatalf("Import() produced %d packages, want 1", len(result.Packages))
	}
	if _, err := os.Stat(filepath.Join(output, "widgets_exampletest", "generated.go")); err != nil {
		t.Fatalf("generated package not written: %v", err)
	}
	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Fatalf("stale JSII file was not removed: %v", err)
	}
}

func TestImportNamedKubernetesUsesModulePrefix(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("testdata", "k8s-definitions.json"))
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if !strings.Contains(request.URL.Path, "/v1.25.0/_definitions.json") {
			t.Fatalf("unexpected schema URL: %s", request.URL)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(data)),
			Request:    request,
		}, nil
	})}
	output := t.TempDir()
	result, err := Import(context.Background(), Options{
		Source:      "k8s@1.25.0",
		OutputDir:   output,
		PackageName: "custom",
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Packages) != 1 || result.Packages[0].PackageName != "custom_k8s" {
		t.Fatalf("generated packages = %#v", result.Packages)
	}
	if _, err := os.Stat(filepath.Join(output, "custom_k8s", "generated.go")); err != nil {
		t.Fatalf("named Kubernetes package not written: %v", err)
	}
}

func TestKubernetes125Smoke(t *testing.T) {
	schemaFile := os.Getenv("PURECDK8S_K8S_SCHEMA")
	if schemaFile == "" {
		t.Skip("set PURECDK8S_K8S_SCHEMA to run the full-schema generation and compile smoke test")
	}
	data, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	generated, err := GenerateKubernetes(data, GenerateOptions{PackageName: "k8s"})
	if err != nil {
		t.Fatal(err)
	}
	generationDuration := time.Since(started)
	if outputFile := os.Getenv("PURECDK8S_K8S_OUTPUT"); outputFile != "" {
		if err := os.WriteFile(outputFile, generated.Code, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	compileGenerated(t, generated)
	t.Logf("generated %d DTO/props types and %d resources (%d bytes) in %s; generation+compile %s",
		len(generated.Types), len(generated.Resources), len(generated.Code), generationDuration, time.Since(started))
}

func TestUpstreamNameNormalization(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"CSIDriver":               "CsiDriver",
		"APIService":              "ApiService",
		"ClusterCIDRSpecV1Alpha1": "ClusterCidrSpecV1Alpha1",
		"KubeClusterCIDRV1Alpha1": "KubeClusterCidrv1Alpha1",
	}
	for input, expected := range tests {
		if actual := normalizeTypeName(input); actual != expected {
			t.Errorf("normalizeTypeName(%q) = %q, want %q", input, actual, expected)
		}
	}
	properties := map[string]string{
		"hostIP":            "hostIp",
		"clusterIPs":        "clusterIPs",
		"openAPIV3Schema":   "openApiv3Schema",
		"setHostnameAsFQDN": "setHostnameAsFqdn",
	}
	for input, expected := range properties {
		if actual := lowerCamel(input); actual != expected {
			t.Errorf("lowerCamel(%q) = %q, want %q", input, actual, expected)
		}
	}
	if got := sanitizePackageName("foo_bar.example-test"); got != "foobarexampletest" {
		t.Errorf("sanitizePackageName() = %q, want %q", got, "foobarexampletest")
	}
	if got := prefixedPackageName("my_widgets", "foo_bar.example"); got != "my_widgets_foobarexample" {
		t.Errorf("prefixedPackageName() = %q, want %q", got, "my_widgets_foobarexample")
	}
}

func compileGenerated(t *testing.T, generated *Generation, testSource ...string) {
	t.Helper()
	temp := t.TempDir()
	root := moduleRoot(t)
	goMod := "module example.com/generatedtest\n\ngo 1.23.0\n\nrequire github.com/purecdk8s/purecdk8s v0.0.0\n\nreplace github.com/purecdk8s/purecdk8s => " + root + "\n"
	if err := os.WriteFile(filepath.Join(temp, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(temp, generated.PackageName)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "generated.go"), generated.Code, 0o644); err != nil {
		t.Fatal(err)
	}
	if len(testSource) > 0 {
		if err := os.WriteFile(filepath.Join(directory, "generated_test.go"), []byte(testSource[0]), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command("go", "test", "-mod=mod", "./...")
	command.Dir = temp
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("generated package does not compile: %v\n%s", err, output)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func between(value, start, end string) string {
	index := strings.Index(value, start)
	if index < 0 {
		return ""
	}
	value = value[index+len(start):]
	index = strings.Index(value, end)
	if index < 0 {
		return value
	}
	return value[:index]
}

func canonicalSource(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
