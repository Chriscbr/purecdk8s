package cdk8s_test

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	constructs "github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
	"gopkg.in/yaml.v3"
)

func appTestingApp(t *testing.T, props *cdk8s.AppProps) cdk8s.App {
	t.Helper()
	app := cdk8s.Testing_App(props)
	t.Cleanup(func() { _ = os.RemoveAll(appStringValue(app.Outdir())) })
	return app
}

func appStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func appNewObject(scope constructs.Construct, id, kind string) cdk8s.ApiObject {
	return cdk8s.NewApiObject(scope, jsii.String(id), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String(kind),
	})
}

func appDirectoryNames(t *testing.T, directory string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.Name())
	}
	return result
}

func appFilesAndFolders(t *testing.T, directory string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			result = append(result, entry.Name())
			continue
		}
		children, err := os.ReadDir(filepath.Join(directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, child := range children {
			result = append(result, entry.Name()+"/"+child.Name())
		}
	}
	return result
}

func appReadFile(t *testing.T, directory, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func appManifestKinds(t *testing.T, manifest string) []string {
	t.Helper()
	decoder := yaml.NewDecoder(new(appStringReader).reset(manifest))
	result := make([]string, 0)
	for {
		var document map[string]interface{}
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(document) != 0 {
			result = append(result, document["kind"].(string))
		}
	}
	return result
}

// appStringReader keeps this file's helpers uniquely prefixed while
// avoiding an otherwise unnecessary bytes/string reader dependency.
type appStringReader struct {
	value string
	off   int
}

func (r *appStringReader) reset(value string) *appStringReader {
	r.value = value
	r.off = 0
	return r
}

func (r *appStringReader) Read(p []byte) (int, error) {
	if r.off >= len(r.value) {
		return 0, io.EOF
	}
	n := copy(p, r.value[r.off:])
	r.off += n
	return n, nil
}

type appHookChart struct{ cdk8s.Chart }

func appNewHookChart(scope constructs.Construct, id string) *appHookChart {
	chart := &appHookChart{}
	cdk8s.NewChart_Override(chart, scope, jsii.String(id), nil)
	appNewObject(chart, "ApiObject1", "Kind1")
	appNewObject(chart, "ApiObject2", "Kind2")
	return chart
}

func (c *appHookChart) ToJson() *[]interface{} {
	c.Node().TryRemoveChild(jsii.String("ApiObject1"))
	return c.Chart.ToJson()
}

type appCustomConstruct struct {
	constructs.Construct
	object cdk8s.ApiObject
}

func appNewCustomConstruct(scope constructs.Construct, id string) *appCustomConstruct {
	result := &appCustomConstruct{}
	constructs.NewConstruct_Override(result, scope, jsii.String(id))
	result.object = appNewObject(result, id+"obj", "CustomConstruct")
	return result
}

type appChildChart1 struct{ cdk8s.Chart }

func appNewChildChart1(scope constructs.Construct, id string) *appChildChart1 {
	result := &appChildChart1{}
	cdk8s.NewChart_Override(result, scope, jsii.String(id), nil)
	appNewObject(result, "namespace1", "Namespace")
	return result
}

type appChildChart2 struct{ cdk8s.Chart }

func appNewChildChart2(scope constructs.Construct, id string) *appChildChart2 {
	result := &appChildChart2{}
	cdk8s.NewChart_Override(result, scope, jsii.String(id), nil)
	appNewObject(result, "namespace2", "Namespace")
	return result
}

type appParentChart struct{ cdk8s.Chart }

func appNewParentChart(scope constructs.Construct, id string) *appParentChart {
	result := &appParentChart{}
	cdk8s.NewChart_Override(result, scope, jsii.String(id), nil)
	appNewChildChart1(result, "child1")
	appNewChildChart2(result, "child2")
	appNewObject(result, "namespace3", "Namespace")
	return result
}

type appValidatingConstruct struct {
	constructs.Construct
	invoked bool
}

func appNewValidatingConstruct(scope constructs.Construct, id string) *appValidatingConstruct {
	result := &appValidatingConstruct{}
	constructs.NewConstruct_Override(result, scope, jsii.String(id))
	result.Node().AddValidation(result)
	return result
}

func (c *appValidatingConstruct) Validate() *[]*string {
	c.invoked = true
	return &[]*string{}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L8
func TestAppSynthYamlConsidersDependencies(t *testing.T) {
	app := appTestingApp(t, nil)
	chart := cdk8s.NewChart(app, jsii.String("Chart"), nil)
	c1 := constructs.NewConstruct(chart, jsii.String("C1"))
	c2 := constructs.NewConstruct(chart, jsii.String("C2"))
	appNewObject(c1, "ApiObject1", "Kind1")
	appNewObject(c2, "ApiObject2", "Kind2")
	c1.Node().AddDependency(c2)

	want := "apiVersion: v1\nkind: Kind2\nmetadata:\n  name: chart-c2-apiobject2-c8a49d62\n---\napiVersion: v1\nkind: Kind1\nmetadata:\n  name: chart-c1-apiobject1-c8f49fa2\n"
	if got := appStringValue(app.SynthYaml()); got != want {
		t.Fatalf("manifest mismatch\n--- got ---\n%s--- want ---\n%s", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L26
func TestAppCanHookChartSynthesisDuringSynthYaml(t *testing.T) {
	app := appTestingApp(t, nil)
	appNewHookChart(app, "Chart")
	kinds := appManifestKinds(t, appStringValue(app.SynthYaml()))
	if want := []string{"Kind2"}; !reflect.DeepEqual(kinds, want) {
		t.Fatalf("manifest kinds = %#v, want %#v", kinds, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L55
func TestAppCanHookChartSynthesisDuringSynth(t *testing.T) {
	app := appTestingApp(t, nil)
	appNewHookChart(app, "Chart")
	app.Synth()

	wantFiles := []string{"chart-c86185a7.k8s.yaml"}
	if got := appDirectoryNames(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, wantFiles) {
		t.Fatalf("files = %#v, want %#v", got, wantFiles)
	}
	want := "apiVersion: v1\nkind: Kind2\nmetadata:\n  name: chart-apiobject2-c81f1ca2\n"
	if got := appReadFile(t, appStringValue(app.Outdir()), wantFiles[0]); got != want {
		t.Fatalf("manifest mismatch\n--- got ---\n%s--- want ---\n%s", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L86
func TestAppEmptyAppEmitsNoFiles(t *testing.T) {
	app := appTestingApp(t, nil)
	app.Synth()
	if got := appDirectoryNames(t, appStringValue(app.Outdir())); len(got) != 0 {
		t.Fatalf("files = %#v, want none", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L97
func TestAppWithTwoCharts(t *testing.T) {
	app := appTestingApp(t, nil)
	cdk8s.NewChart(app, jsii.String("chart1"), nil)
	cdk8s.NewChart(app, jsii.String("chart2"), nil)
	app.Synth()
	want := []string{"chart1.k8s.yaml", "chart2.k8s.yaml"}
	if got := appDirectoryNames(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L113
func TestAppWithChartsDirectlyDependant(t *testing.T) {
	app := appTestingApp(t, nil)
	chart1 := cdk8s.NewChart(app, jsii.String("chart1"), nil)
	chart2 := cdk8s.NewChart(app, jsii.String("chart2"), nil)
	chart3 := cdk8s.NewChart(app, jsii.String("chart3"), nil)
	chart1.Node().AddDependency(chart2)
	chart2.Node().AddDependency(chart3)
	app.Synth()
	want := []string{"0000-chart3.k8s.yaml", "0001-chart2.k8s.yaml", "0002-chart1.k8s.yaml"}
	if got := appDirectoryNames(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L137
func TestAppWithChartsIndirectlyDependant(t *testing.T) {
	app := appTestingApp(t, nil)
	chart1 := cdk8s.NewChart(app, jsii.String("chart1"), nil)
	chart2 := cdk8s.NewChart(app, jsii.String("chart2"), nil)
	chart3 := cdk8s.NewChart(app, jsii.String("chart3"), nil)
	obj1 := appNewObject(chart1, "obj1", "Kind1")
	obj2 := appNewObject(chart2, "obj2", "Kind2")
	obj3 := appNewObject(chart3, "obj3", "Kind3")
	obj1.Node().AddDependency(obj2)
	obj2.Node().AddDependency(obj3)
	app.Synth()
	want := []string{"0000-chart3.k8s.yaml", "0001-chart2.k8s.yaml", "0002-chart1.k8s.yaml"}
	if got := appDirectoryNames(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L165
func TestAppDefaultOutputDirectoryIsDist(t *testing.T) {
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	workdir := t.TempDir()
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	app := cdk8s.NewApp(nil)
	cdk8s.NewChart(app, jsii.String("chart1"), nil)
	app.Synth()
	if got := appStringValue(app.Outdir()); got != "dist" {
		t.Fatalf("outdir = %q, want dist", got)
	}
	info, err := os.Stat("dist")
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatal("dist is not a directory")
	}
	if _, err := os.Stat(filepath.Join("dist", "chart1.k8s.yaml")); err != nil {
		t.Fatal(err)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L189
func TestAppWithDependentAndIndependentCharts(t *testing.T) {
	app := appTestingApp(t, nil)
	chart1 := cdk8s.NewChart(app, jsii.String("chart1"), nil)
	cdk8s.NewChart(app, jsii.String("chart2"), nil)
	chart3 := cdk8s.NewChart(app, jsii.String("chart3"), nil)
	cdk8s.NewChart(app, jsii.String("chart4"), nil)
	chart1.Node().AddDependency(chart3)
	app.Synth()
	want := []string{"0000-chart3.k8s.yaml", "0001-chart1.k8s.yaml", "0002-chart2.k8s.yaml", "0003-chart4.k8s.yaml"}
	if got := appDirectoryNames(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L214
func TestAppWithChartDependenciesViaCustomConstructs(t *testing.T) {
	app := appTestingApp(t, nil)
	chart1 := cdk8s.NewChart(app, jsii.String("chart1"), nil)
	chart2 := cdk8s.NewChart(app, jsii.String("chart2"), nil)
	microService := appNewCustomConstruct(chart1, "MicroService")
	dataBase := appNewCustomConstruct(chart2, "DataBase")
	microService.Node().AddDependency(dataBase)
	app.Synth()
	want := []string{"0000-chart2.k8s.yaml", "0001-chart1.k8s.yaml"}
	if got := appDirectoryNames(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L245
func TestAppNestedChartsDeduplicateApiObjects(t *testing.T) {
	app := appTestingApp(t, nil)
	chart := cdk8s.NewChart(app, jsii.String("chart1"), nil)
	childChart := cdk8s.NewChart(chart, jsii.String("chart2"), nil)
	appNewCustomConstruct(chart, "child1")
	appNewCustomConstruct(childChart, "child2")
	app.Synth()

	wantFiles := []string{"0000-chart1-chart2-c883b207.k8s.yaml", "0001-chart1.k8s.yaml"}
	if got := appDirectoryNames(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, wantFiles) {
		t.Fatalf("files = %#v, want %#v", got, wantFiles)
	}
	wantManifests := []string{
		"apiVersion: v1\nkind: CustomConstruct\nmetadata:\n  name: chart1-chart2-child2-child2obj-c828dca6\n",
		"apiVersion: v1\nkind: CustomConstruct\nmetadata:\n  name: chart1-child1-child1obj-c868628e\n",
	}
	for index, file := range wantFiles {
		if got := appReadFile(t, appStringValue(app.Outdir()), file); got != wantManifests[index] {
			t.Errorf("%s mismatch\n--- got ---\n%s--- want ---\n%s", file, got, wantManifests[index])
		}
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L275
func TestAppNestedChartClassesDeduplicateApiObjects(t *testing.T) {
	app := appTestingApp(t, nil)
	appNewParentChart(app, "parent")
	app.Synth()

	wantFiles := []string{"0000-parent-child1-c8e38b2d.k8s.yaml", "0001-parent-child2-c882caae.k8s.yaml", "0002-parent.k8s.yaml"}
	if got := appDirectoryNames(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, wantFiles) {
		t.Fatalf("files = %#v, want %#v", got, wantFiles)
	}
	wantManifests := []string{
		"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: parent-child1-namespace1-c871643e\n",
		"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: parent-child2-namespace2-c806260b\n",
		"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: parent-namespace3-c8bf842a\n",
	}
	for index, file := range wantFiles {
		if got := appReadFile(t, appStringValue(app.Outdir()), file); got != wantManifests[index] {
			t.Errorf("%s mismatch\n--- got ---\n%s--- want ---\n%s", file, got, wantManifests[index])
		}
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L316
func TestAppSynthCallsValidate(t *testing.T) {
	output := t.TempDir()
	app := cdk8s.NewApp(&cdk8s.AppProps{Outdir: &output})
	construct := appNewValidatingConstruct(app, "ValidatingConstruct")
	app.Synth()
	if !construct.invoked {
		t.Fatal("validation was not invoked")
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L341
func TestAppOutputTypesWithTwoEmptyCharts(t *testing.T) {
	tests := []struct {
		name       string
		outputType cdk8s.YamlOutputType
		want       []string
	}{
		{"file per chart", cdk8s.YamlOutputType_FILE_PER_CHART, []string{"chart1.k8s.yaml", "chart2.k8s.yaml"}},
		{"file per app", cdk8s.YamlOutputType_FILE_PER_APP, []string{"app.k8s.yaml"}},
		{"file per resource", cdk8s.YamlOutputType_FILE_PER_RESOURCE, []string{}},
		{"folder per chart", cdk8s.YamlOutputType_FOLDER_PER_CHART_FILE_PER_RESOURCE, []string{"chart1", "chart2"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := appTestingApp(t, &cdk8s.AppProps{YamlOutputType: test.outputType})
			cdk8s.NewChart(app, jsii.String("chart1"), nil)
			cdk8s.NewChart(app, jsii.String("chart2"), nil)
			app.Synth()
			if got := appDirectoryNames(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("files = %#v, want %#v", got, test.want)
			}
		})
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L374
func TestAppReturnAppAsYamlString(t *testing.T) {
	app := appTestingApp(t, nil)
	chart1 := cdk8s.NewChart(app, jsii.String("chart1"), nil)
	appNewObject(chart1, "obj1", "Kind1")
	appNewObject(chart1, "obj2", "Kind2")
	cdk8s.NewChart(app, jsii.String("chart2"), nil)
	chart3 := cdk8s.NewChart(app, jsii.String("chart3"), nil)
	appNewObject(chart3, "obj3", "Kind3")
	appNewObject(chart3, "obj4", "Kind4")

	want := "apiVersion: v1\nkind: Kind1\nmetadata:\n  name: chart1-obj1-c818e77f\n---\napiVersion: v1\nkind: Kind2\nmetadata:\n  name: chart1-obj2-c87a5a2e\n---\napiVersion: v1\nkind: Kind3\nmetadata:\n  name: chart3-obj3-c8abbfb5\n---\napiVersion: v1\nkind: Kind4\nmetadata:\n  name: chart3-obj4-c8da728e\n"
	if got := appStringValue(app.SynthYaml()); got != want {
		t.Fatalf("manifest mismatch\n--- got ---\n%s--- want ---\n%s", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L396
func TestAppOutputTypesWithIndirectDependencies(t *testing.T) {
	tests := []struct {
		name       string
		outputType cdk8s.YamlOutputType
		want       []string
	}{
		{"file per chart", cdk8s.YamlOutputType_FILE_PER_CHART, []string{"0000-chart3.k8s.yaml", "0001-chart2.k8s.yaml", "0002-chart1.k8s.yaml"}},
		{"file per app", cdk8s.YamlOutputType_FILE_PER_APP, []string{"app.k8s.yaml"}},
		{"file per resource", cdk8s.YamlOutputType_FILE_PER_RESOURCE, []string{"Kind1.chart1-obj1-c818e77f.k8s.yaml", "Kind2.chart2-obj2-c8636f20.k8s.yaml", "Kind3.chart3-obj3-c8abbfb5.k8s.yaml"}},
		{"folder per chart", cdk8s.YamlOutputType_FOLDER_PER_CHART_FILE_PER_RESOURCE, []string{"0000-chart3/Kind3.chart3-obj3-c8abbfb5.k8s.yaml", "0001-chart2/Kind2.chart2-obj2-c8636f20.k8s.yaml", "0002-chart1/Kind1.chart1-obj1-c818e77f.k8s.yaml"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := appTestingApp(t, &cdk8s.AppProps{YamlOutputType: test.outputType})
			chart1 := cdk8s.NewChart(app, jsii.String("chart1"), nil)
			chart2 := cdk8s.NewChart(app, jsii.String("chart2"), nil)
			chart3 := cdk8s.NewChart(app, jsii.String("chart3"), nil)
			obj1 := appNewObject(chart1, "obj1", "Kind1")
			obj2 := appNewObject(chart2, "obj2", "Kind2")
			obj3 := appNewObject(chart3, "obj3", "Kind3")
			obj1.Node().AddDependency(obj2)
			obj2.Node().AddDependency(obj3)
			app.Synth()
			if got := appFilesAndFolders(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("files = %#v, want %#v", got, test.want)
			}
		})
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L451
func TestAppOutputTypesWithCustomConstructDependencies(t *testing.T) {
	tests := []struct {
		name       string
		outputType cdk8s.YamlOutputType
		want       []string
	}{
		{"file per chart", cdk8s.YamlOutputType_FILE_PER_CHART, []string{"0000-chart2.k8s.yaml", "0001-chart1.k8s.yaml"}},
		{"file per app", cdk8s.YamlOutputType_FILE_PER_APP, []string{"app.k8s.yaml"}},
		{"file per resource", cdk8s.YamlOutputType_FILE_PER_RESOURCE, []string{"CustomConstruct.chart1-microservice-microserviceobj-c8e1164f.k8s.yaml", "CustomConstruct.chart2-database-databaseobj-c8b5eba3.k8s.yaml"}},
		{"folder per chart", cdk8s.YamlOutputType_FOLDER_PER_CHART_FILE_PER_RESOURCE, []string{"0000-chart2/CustomConstruct.chart2-database-databaseobj-c8b5eba3.k8s.yaml", "0001-chart1/CustomConstruct.chart1-microservice-microserviceobj-c8e1164f.k8s.yaml"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := appTestingApp(t, &cdk8s.AppProps{YamlOutputType: test.outputType})
			chart1 := cdk8s.NewChart(app, jsii.String("chart1"), nil)
			chart2 := cdk8s.NewChart(app, jsii.String("chart2"), nil)
			microService := appNewCustomConstruct(chart1, "MicroService")
			dataBase := appNewCustomConstruct(chart2, "DataBase")
			microService.Node().AddDependency(dataBase)
			app.Synth()
			if got := appFilesAndFolders(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("files = %#v, want %#v", got, test.want)
			}
		})
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/app.test.ts#L505
func TestAppModifiedExtensionsWithOutputTypes(t *testing.T) {
	tests := []struct {
		name       string
		outputType cdk8s.YamlOutputType
		want       []string
	}{
		{"file per chart", cdk8s.YamlOutputType_FILE_PER_CHART, []string{"chart1.yaml", "chart2.yaml"}},
		{"file per app", cdk8s.YamlOutputType_FILE_PER_APP, []string{"app.yaml"}},
		{"file per resource", cdk8s.YamlOutputType_FILE_PER_RESOURCE, []string{}},
		{"folder per chart", cdk8s.YamlOutputType_FOLDER_PER_CHART_FILE_PER_RESOURCE, []string{"chart1", "chart2"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			extension := ".yaml"
			app := appTestingApp(t, &cdk8s.AppProps{YamlOutputType: test.outputType, OutputFileExtension: &extension})
			cdk8s.NewChart(app, jsii.String("chart1"), nil)
			cdk8s.NewChart(app, jsii.String("chart2"), nil)
			app.Synth()
			if got := appDirectoryNames(t, appStringValue(app.Outdir())); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("files = %#v, want %#v", got, test.want)
			}
		})
	}
}
