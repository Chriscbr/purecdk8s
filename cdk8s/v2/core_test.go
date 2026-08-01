package cdk8s

import (
	"os"
	"path/filepath"
	"testing"

	purecdk8sserialization "github.com/purecdk8s/purecdk8s/serialization"
)

type testDeploymentProps struct {
	Metadata *ApiObjectMetadata  `json:"metadata,omitempty"`
	Spec     *testDeploymentSpec `json:"spec,omitempty"`
}

type testDeploymentSpec struct {
	Replicas *float64           `json:"replicas,omitempty"`
	Selector *testLabelSelector `json:"selector,omitempty"`
	Template *testPodTemplate   `json:"template,omitempty"`
}

type testLabelSelector struct {
	MatchLabels *map[string]*string `json:"matchLabels,omitempty"`
}

type testPodTemplate struct {
	Metadata *testObjectMeta `json:"metadata,omitempty"`
	Spec     *testPodSpec    `json:"spec,omitempty"`
}

type testObjectMeta struct {
	Labels *map[string]*string `json:"labels,omitempty"`
}

type testPodSpec struct {
	Containers *[]*testContainer `json:"containers,omitempty"`
}

type testContainer struct {
	Image *string               `json:"image,omitempty"`
	Name  *string               `json:"name,omitempty"`
	Ports *[]*testContainerPort `json:"ports,omitempty"`
}

type testContainerPort struct {
	ContainerPort *float64 `json:"containerPort,omitempty"`
}

func TestGettingStartedGolden(t *testing.T) {
	output := t.TempDir()
	app := NewApp(&AppProps{Outdir: &output})
	chartID, namespace := "getting-started", "default"
	chart := NewChart(app, &chartID, &ChartProps{Namespace: &namespace})
	resourceID := "deployment"
	apiVersion, kind := "apps/v1", "Deployment"
	name, image, label := "app-container", "nginx:1.19.10", "my-app"
	replicas, port := float64(3), float64(80)
	labels := map[string]*string{"app": &label}
	props := &testDeploymentProps{
		Spec: &testDeploymentSpec{
			Replicas: &replicas,
			Selector: &testLabelSelector{MatchLabels: &labels},
			Template: &testPodTemplate{
				Metadata: &testObjectMeta{Labels: &labels},
				Spec: &testPodSpec{Containers: &[]*testContainer{{
					Name:  &name,
					Image: &image,
					Ports: &[]*testContainerPort{{
						ContainerPort: &port,
					}},
				}}},
			},
		},
	}
	object := NewApiObjectWithManifest(chart, &resourceID, &ApiObjectProps{
		ApiVersion: &apiVersion,
		Kind:       &kind,
	}, props)

	if got, want := *object.Name(), "getting-started-deployment-c80c7257"; got != want {
		t.Fatalf("object name = %q, want %q", got, want)
	}

	app.Synth()
	data, err := os.ReadFile(filepath.Join(output, "getting-started.k8s.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	want := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: getting-started-deployment-c80c7257
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - image: nginx:1.19.10
          name: app-container
          ports:
            - containerPort: 80
`
	if got := string(data); got != want {
		t.Fatalf("manifest mismatch\n--- got ---\n%s--- want ---\n%s", got, want)
	}
}

type testUnion struct{ number *float64 }

func (u *testUnion) Value() interface{}    { return u.number }
func (u *testUnion) PureCDK8sScalarUnion() {}

type testCasing struct {
	HostIp *string    `json:"hostIp,omitempty" k8s:"hostIP"`
	Union  *testUnion `json:"union,omitempty"`
	Mode   testMode   `field:"optional" json:"mode"`
}

type testMode string

const testModeDeleteOptions testMode = "DELETE_OPTIONS"

func TestPlainValueUsesKubernetesCasingAndUnionValue(t *testing.T) {
	purecdk8sserialization.RegisterEnumWireValues(map[testMode]interface{}{
		testModeDeleteOptions: "DeleteOptions",
	})
	host := "127.0.0.1"
	number := float64(42)
	got := plainMap(&testCasing{HostIp: &host, Union: &testUnion{number: &number}})
	if got["hostIP"] != host {
		t.Fatalf("hostIP = %#v", got["hostIP"])
	}
	if got["union"] != number {
		t.Fatalf("union = %#v", got["union"])
	}
	if _, found := got["mode"]; found {
		t.Fatalf("optional zero enum was retained: %#v", got)
	}
	got = plainMap(&testCasing{Mode: testModeDeleteOptions})
	if got["mode"] != "DeleteOptions" {
		t.Fatalf("enum wire value = %#v", got["mode"])
	}
}

type testProducer struct{ value interface{} }

func (p *testProducer) Produce() interface{} { return p.value }

type dependencyObservingValidation struct {
	chart        Chart
	observations *[]bool
}

func (v *dependencyObservingValidation) Validate() *[]*string {
	*v.observations = append(*v.observations, len(*v.chart.Node().Dependencies()) != 0)
	result := make([]*string, 0)
	return &result
}

type testLazyManifest struct {
	Spec interface{} `json:"spec"`
}

func TestApiObjectRetainsLazyValuesUntilSynthesis(t *testing.T) {
	app := NewApp(nil)
	chartID, resourceID := "chart", "resource"
	chart := NewChart(app, &chartID, nil)
	apiVersion, kind := "v1", "ConfigMap"
	object := NewApiObjectWithManifest(chart, &resourceID, &ApiObjectProps{
		ApiVersion: &apiVersion,
		Kind:       &kind,
	}, &testLazyManifest{Spec: Lazy_Any(&testProducer{value: map[string]interface{}{"answer": 42}})})

	manifest := plainMap(object.ToJson())
	spec := plainMap(manifest["spec"])
	if got := spec["answer"]; got != float64(42) {
		t.Fatalf("lazy value = %#v", got)
	}
}

func TestNativeScalarUnionSynthesizesAsScalar(t *testing.T) {
	app := NewApp(nil)
	chartID, resourceID := "chart", "resource"
	chart := NewChart(app, &chartID, nil)
	apiVersion, kind := "v1", "Service"
	port := float64(8080)
	object := NewApiObjectWithManifest(chart, &resourceID, &ApiObjectProps{
		ApiVersion: &apiVersion,
		Kind:       &kind,
	}, map[string]interface{}{"spec": map[string]interface{}{"port": &testUnion{number: &port}}})

	spec := plainMap(plainMap(object.ToJson())["spec"])
	if got := spec["port"]; got != port {
		t.Fatalf("native union = %#v, want scalar %#v", got, port)
	}
}

func TestAppExplicitEmptyOutdirMatchesUpstream(t *testing.T) {
	outdir := ""
	app := NewApp(&AppProps{Outdir: &outdir})
	if got := *app.Outdir(); got != "" {
		t.Fatalf("outdir = %q, want explicit empty string", got)
	}
}

func TestAppSynthValidatesBeforeInferringDependencies(t *testing.T) {
	outdir := t.TempDir()
	app := NewApp(&AppProps{Outdir: &outdir})
	parentID, childID := "parent", "child"
	parent := NewChart(app, &parentID, nil)
	NewChart(parent, &childID, nil)

	observations := make([]bool, 0)
	app.Node().AddValidation(&dependencyObservingValidation{
		chart:        parent,
		observations: &observations,
	})
	app.Synth()
	if len(observations) == 0 || observations[0] {
		t.Fatalf("dependency observations = %#v, want first validation before inference", observations)
	}
}

func TestMetadataRequiredArgumentsMatchUpstream(t *testing.T) {
	chart := Testing_Chart()
	id, apiVersion, kind := "object", "v1", "ConfigMap"
	object := NewApiObject(chart, &id, &ApiObjectProps{ApiVersion: &apiVersion, Kind: &kind})

	assertCorePanics(t, func() { object.Metadata().Add(&id, nil) })
	assertCorePanics(t, func() { object.Metadata().GetLabel(nil) })
	assertCorePanics(t, func() { object.Metadata().AddFinalizers(nil) })
}

func TestDirectMetadataDefinitionDoesNotInheritChartDefaults(t *testing.T) {
	app := NewApp(nil)
	chartID, objectID := "chart", "object"
	namespace, labelKey, labelValue := "chart-ns", "chart-label", "value"
	labels := map[string]*string{labelKey: &labelValue}
	chart := NewChart(app, &chartID, &ChartProps{Namespace: &namespace, Labels: &labels})
	apiVersion, kind := "v1", "ConfigMap"
	object := NewApiObject(chart, &objectID, &ApiObjectProps{ApiVersion: &apiVersion, Kind: &kind})

	metadata := NewApiObjectMetadataDefinition(&ApiObjectMetadataDefinitionOptions{ApiObject: object})
	if metadata.Name() != nil {
		t.Fatalf("direct metadata name = %q, want nil", *metadata.Name())
	}
	if metadata.Namespace() != nil {
		t.Fatalf("direct metadata namespace = %q, want nil", *metadata.Namespace())
	}
	if got := plainMap(metadata.ToJson()); len(got) != 0 {
		t.Fatalf("direct metadata = %#v, want empty object", got)
	}
}

func TestIncludePreservesExplicitEmptyMetadataName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "objects.yaml")
	data := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: \"\"\n  namespace: ns\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	chart := Testing_Chart()
	id := "include"
	include := NewInclude(chart, &id, &IncludeProps{Url: &path})
	objects := *include.ApiObjects()
	if len(objects) != 1 {
		t.Fatalf("included objects = %d, want 1", len(objects))
	}
	if got := *objects[0].Node().Id(); got != "configmap-ns" {
		t.Fatalf("included object id = %q, want configmap-ns", got)
	}
	if got := *objects[0].Name(); got != "" {
		t.Fatalf("included object name = %q, want explicit empty name", got)
	}
}

func TestInvalidOutputTypePanicsDuringAppConstruction(t *testing.T) {
	assertCorePanics(t, func() {
		NewApp(&AppProps{YamlOutputType: YamlOutputType("UNKNOWN")})
	})
}

func assertCorePanics(t *testing.T, callback func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	callback()
}

type customApp struct{ App }
type customChart struct{ Chart }
type customApiObject struct{ ApiObject }

func TestOverrideConstructorsSupportEmbeddedInterfaces(t *testing.T) {
	app := &customApp{}
	NewApp_Override(app, nil)
	chart := &customChart{}
	chartID := "custom-chart"
	NewChart_Override(chart, app, &chartID, nil)
	object := &customApiObject{}
	resourceID, apiVersion, kind := "resource", "v1", "ConfigMap"
	NewApiObject_Override(object, chart, &resourceID, &ApiObjectProps{
		ApiVersion: &apiVersion,
		Kind:       &kind,
	})

	charts := *app.Charts()
	if len(charts) != 1 || charts[0] != chart {
		t.Fatalf("charts = %#v", charts)
	}
	objects := *chart.ApiObjects()
	if len(objects) != 1 || objects[0] != object {
		t.Fatalf("objects = %#v", objects)
	}
	if object.Chart() != chart {
		t.Fatalf("object chart did not retain custom chart host")
	}
	if manifests := chart.ToJson(); len(*manifests) != 1 {
		t.Fatalf("manifest count = %d", len(*manifests))
	}
}

func TestYamlKeepsKubernetesTopLevelKeyOrder(t *testing.T) {
	got := *Yaml_Stringify(map[string]interface{}{
		"zeta":       1,
		"metadata":   map[string]interface{}{"name": "example"},
		"aaa":        2,
		"kind":       "Thing",
		"apiVersion": "example.io/v1",
	})
	want := "apiVersion: example.io/v1\nkind: Thing\nmetadata:\n  name: example\naaa: 2\nzeta: 1\n"
	if got != want {
		t.Fatalf("yaml =\n%s\nwant:\n%s", got, want)
	}
}
