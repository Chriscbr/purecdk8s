package cdk8s_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	constructs "github.com/Chriscbr/purecdk8s/constructs/v10"
)

func chartString(value string) *string { return &value }

func chartBool(value bool) *bool { return &value }

func chartApp(t *testing.T) cdk8s.App {
	t.Helper()
	outdir := t.TempDir()
	return cdk8s.Testing_App(&cdk8s.AppProps{Outdir: &outdir})
}

func chartChart(t *testing.T) cdk8s.Chart {
	t.Helper()
	return cdk8s.NewChart(chartApp(t), chartString("test"), nil)
}

func chartNewObject(scope constructs.Construct, id, kind string) cdk8s.ApiObject {
	return cdk8s.NewApiObject(scope, chartString(id), &cdk8s.ApiObjectProps{
		ApiVersion: chartString("v1"),
		Kind:       chartString(kind),
	})
}

func chartManifest(apiVersion, kind, name string) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": name,
		},
	}
}

func chartAssertJSONEqual(t *testing.T, got, want interface{}) {
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

func chartRequirePanicContains(t *testing.T, want string, callback func()) {
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

type chartProducer struct{ value interface{} }

func (producer *chartProducer) Produce() interface{} { return producer.value }

type chartImplicitToken struct{ value interface{} }

func (token *chartImplicitToken) Resolve() interface{} { return token.value }

type chartValidation struct{ invoked *bool }

func (validation *chartValidation) Validate() *[]*string {
	*validation.invoked = true
	return &[]*string{}
}

type chartCustomChart struct{ cdk8s.Chart }

type chartCustomConstruct struct {
	construct constructs.Construct
	object    cdk8s.ApiObject
}

func chartNewCustomConstruct(scope constructs.Construct, id string) chartCustomConstruct {
	construct := constructs.NewConstruct(scope, chartString(id))
	object := chartNewObject(construct, id+"obj", "CustomConstruct")
	return chartCustomConstruct{construct: construct, object: object}
}

func chartNewNestedCustomConstruct(scope constructs.Construct, id string) chartCustomConstruct {
	outer := constructs.NewConstruct(scope, chartString(id))
	nested := chartNewCustomConstruct(outer, "nested")
	return chartCustomConstruct{construct: outer, object: nested.object}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L7
func TestChartEmptyStack(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("empty"), nil)

	chartAssertJSONEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L18
func TestChartDisablingResourceNameHashesAtChartLevel(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("test"), &cdk8s.ChartProps{
		DisableResourceNameHashes: chartBool(true),
	})
	object1 := chartNewObject(chart, "resource1", "Resource1")
	object2 := chartNewObject(chart, "resource2", "Resource3")

	if got, want := *object1.Name(), "test-resource1"; got != want {
		t.Fatalf("first name = %q, want %q", got, want)
	}
	if got, want := *object2.Name(), "test-resource2"; got != want {
		t.Fatalf("second name = %q, want %q", got, want)
	}
	chartAssertJSONEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		chartManifest("v1", "Resource1", "test-resource1"),
		chartManifest("v1", "Resource3", "test-resource2"),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L35
func TestChartResourceNameHashesWorkByDefault(t *testing.T) {
	chart := chartChart(t)
	object1 := chartNewObject(chart, "resource1", "Resource1")
	object2 := chartNewObject(chart, "resource2", "Resource3")

	if got, want := *object1.Name(), "test-resource1-c85cb0fc"; got != want {
		t.Fatalf("first name = %q, want %q", got, want)
	}
	if got, want := *object2.Name(), "test-resource2-c8c6bd27"; got != want {
		t.Fatalf("second name = %q, want %q", got, want)
	}
	chartAssertJSONEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		chartManifest("v1", "Resource1", "test-resource1-c85cb0fc"),
		chartManifest("v1", "Resource3", "test-resource2-c8c6bd27"),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L50
func TestChartOutputIncludesAllSynthesizedResources(t *testing.T) {
	chart := chartChart(t)
	chartNewObject(chart, "resource1", "Resource1")
	chartNewObject(chart, "resource2", "Resource2")
	chartNewObject(chart, "resource3", "Resource3")
	scope := constructs.NewConstruct(chart, chartString("scope"))
	chartNewObject(scope, "resource1", "Resource1")
	chartNewObject(scope, "resource2", "Resource2")

	chartAssertJSONEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		chartManifest("v1", "Resource1", "test-resource1-c85cb0fc"),
		chartManifest("v1", "Resource2", "test-resource2-c8c6bd27"),
		chartManifest("v1", "Resource3", "test-resource3-c8ccc739"),
		chartManifest("v1", "Resource1", "test-scope-resource1-c84ac5c2"),
		chartManifest("v1", "Resource2", "test-scope-resource2-c889750d"),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L69
func TestChartCronJobNamesAreAtMost52Characters(t *testing.T) {
	chart := chartChart(t)
	short := chartNewObject(chart, "cj1", "CronJob")
	long := chartNewObject(chart, "resourceNameThatIsLongerThan52Charactersssssssssssss", "CronJob")

	if got := len(*short.Name()); got > 52 {
		t.Fatalf("short CronJob name length = %d, want <= 52", got)
	}
	if got := len(*long.Name()); got > 52 {
		t.Fatalf("long CronJob name length = %d, want <= 52", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L83
func TestChartTokensAreResolvedDuringSynth(t *testing.T) {
	chart := chartChart(t)
	cdk8s.NewApiObjectWithManifest(chart, chartString("resource1"), &cdk8s.ApiObjectProps{
		ApiVersion: chartString("v1"),
		Kind:       chartString("Resource1"),
	}, map[string]interface{}{
		"spec": map[string]interface{}{
			"foo": cdk8s.Lazy_Any(&chartProducer{value: 123}),
			"implicitToken": &chartImplicitToken{value: map[string]interface{}{
				"foo": "bar",
			}},
		},
	})

	want := chartManifest("v1", "Resource1", "test-resource1-c85cb0fc")
	want["spec"] = map[string]interface{}{
		"foo":           123,
		"implicitToken": map[string]interface{}{"foo": "bar"},
	}
	chartAssertJSONEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{want})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L102
func TestChartOfReturnsFirstParentChart(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("MyFirst"), nil)
	direct := constructs.NewConstruct(chart, chartString("Direct"))
	indirect := constructs.NewConstruct(direct, chartString("Indirect"))
	childChart := cdk8s.NewChart(indirect, chartString("ChildChart"), nil)
	childChild := constructs.NewConstruct(childChart, chartString("ChildChild"))

	for description, test := range map[string]struct {
		construct constructs.IConstruct
		want      cdk8s.Chart
	}{
		"chart":       {chart, chart},
		"direct":      {direct, chart},
		"indirect":    {indirect, chart},
		"child chart": {childChart, childChart},
		"child child": {childChild, childChart},
	} {
		if got := cdk8s.Chart_Of(test.construct); got != test.want {
			t.Errorf("Chart_Of(%s) returned the wrong chart", description)
		}
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L121
func TestChartOfFailsWithoutParentChart(t *testing.T) {
	app := chartApp(t)
	child := constructs.NewConstruct(app, chartString("MyConstruct"))

	chartRequirePanicContains(t, "cannot find a parent chart (directly or indirectly)", func() {
		cdk8s.Chart_Of(child)
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L130
func TestChartToJsonSynthesizesSpecificChart(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("chart"), nil)
	chartNewObject(chart, "obj1", "Kind1")
	chartNewObject(chart, "obj2", "Kind2")

	chartAssertJSONEqual(t, *chart.ToJson(), []interface{}{
		chartManifest("v1", "Kind1", "chart-obj1-c80aa35c"),
		chartManifest("v1", "Kind2", "chart-obj2-c8016fab"),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L144
func TestChartAddDependency(t *testing.T) {
	app := chartApp(t)
	chart1 := cdk8s.NewChart(app, chartString("chart1"), nil)
	chart2 := cdk8s.NewChart(app, chartString("chart2"), nil)
	chart3 := cdk8s.NewChart(app, chartString("chart3"), nil)

	chart1.AddDependency(chart2, chart3)
	dependencies := *chart1.Node().Dependencies()
	if len(dependencies) != 2 || dependencies[0] != chart2 || dependencies[1] != chart3 {
		t.Fatalf("dependencies = %#v, want chart2 then chart3", dependencies)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L162
func TestChartIsChartDistinguishesChartsFromNonCharts(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("chart"), nil)
	childChart := &chartCustomChart{}
	cdk8s.NewChart_Override(childChart, app, chartString("my-chart"), nil)
	object := chartNewObject(chart, "api-object", "Foo")

	for description, test := range map[string]struct {
		value interface{}
		want  bool
	}{
		"chart":       {chart, true},
		"child chart": {childChart, true},
		"api object":  {object, false},
		"app":         {app, false},
	} {
		if got := *cdk8s.Chart_IsChart(test.value); got != test.want {
			t.Errorf("Chart_IsChart(%s) = %t, want %t", description, got, test.want)
		}
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L189
func TestChartToJsonValidatesTheChart(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("chart"), nil)
	construct := constructs.NewConstruct(chart, chartString("ValidatingConstruct"))
	invoked := false
	construct.Node().AddValidation(&chartValidation{invoked: &invoked})

	chart.ToJson()
	if !invoked {
		t.Fatal("chart validation was not invoked")
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L214
func TestChartToJsonReturnsOrderedList(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("chart1"), nil)
	object1 := chartNewObject(chart, "obj1", "Kind1")
	object2 := chartNewObject(chart, "obj2", "Kind2")
	object3 := chartNewObject(chart, "obj3", "Kind3")
	object1.AddDependency(object2)
	object2.AddDependency(object3)

	chartAssertJSONEqual(t, *chart.ToJson(), []interface{}{
		object3.ToJson(), object2.ToJson(), object1.ToJson(),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L234
func TestChartToJsonIgnoresObjectsFromDifferentChart(t *testing.T) {
	app := chartApp(t)
	chart1 := cdk8s.NewChart(app, chartString("chart1"), nil)
	chart2 := cdk8s.NewChart(app, chartString("chart2"), nil)
	object1 := chartNewObject(chart1, "obj1", "Kind1")
	object2 := chartNewObject(chart2, "obj2", "Kind2")
	object1.AddDependency(object2)

	chartAssertJSONEqual(t, *chart1.ToJson(), []interface{}{object1.ToJson()})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L251
func TestChartToJsonIgnoresChartObjects(t *testing.T) {
	app := chartApp(t)
	chart1 := cdk8s.NewChart(app, chartString("chart1"), nil)
	chart2 := cdk8s.NewChart(app, chartString("chart2"), nil)
	object1 := chartNewObject(chart1, "obj1", "Kind1")
	object2 := chartNewObject(chart2, "obj2", "Kind2")
	object1.AddDependency(object2)
	chart1.AddDependency(chart2)

	chartAssertJSONEqual(t, *chart1.ToJson(), []interface{}{object1.ToJson()})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L269
func TestChartToJsonOrdersCustomConstructs(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("chart"), nil)
	microService := chartNewCustomConstruct(chart, "MicroService")
	database := chartNewCustomConstruct(chart, "Database")
	microService.construct.Node().AddDependency(database.construct)

	chartAssertJSONEqual(t, *chart.ToJson(), []interface{}{
		database.object.ToJson(), microService.object.ToJson(),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L286
func TestChartToJsonOrdersTransitiveCustomConstructs(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("chart"), nil)
	microService := chartNewCustomConstruct(chart, "MicroService")
	database := chartNewNestedCustomConstruct(chart, "Database")
	microService.construct.Node().AddDependency(database.construct)

	chartAssertJSONEqual(t, *chart.ToJson(), []interface{}{
		database.object.ToJson(), microService.object.ToJson(),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L303
func TestChartToJsonApiObjectDependsOnCustomConstruct(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("chart"), nil)
	microService := chartNewObject(chart, "MicroService", "MicroService")
	database := chartNewCustomConstruct(chart, "Database")
	microService.AddDependency(database.construct)

	chartAssertJSONEqual(t, *chart.ToJson(), []interface{}{
		database.object.ToJson(), microService.ToJson(),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L320
func TestChartToJsonConstructDependsOnApiObject(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("chart"), nil)
	database := chartNewObject(chart, "MicroService", "MicroService")
	microService := chartNewCustomConstruct(chart, "Database")
	microService.construct.Node().AddDependency(database)

	chartAssertJSONEqual(t, *chart.ToJson(), []interface{}{
		database.ToJson(), microService.object.ToJson(),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L337
func TestChartParentExcludesObjectsFromChildCharts(t *testing.T) {
	app := chartApp(t)
	chart := cdk8s.NewChart(app, chartString("chart1"), nil)
	childChart := cdk8s.NewChart(chart, chartString("chart2"), nil)
	chartNewCustomConstruct(chart, "child1")
	chartNewCustomConstruct(childChart, "child2")

	chartAssertJSONEqual(t, *chart.ToJson(), []interface{}{
		chartManifest("v1", "CustomConstruct", "chart1-child1-child1obj-c868628e"),
	})
	chartAssertJSONEqual(t, *childChart.ToJson(), []interface{}{
		chartManifest("v1", "CustomConstruct", "chart1-chart2-child2-child2obj-c828dca6"),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L356
func TestChartConstructMetadataRecordedWhenRequestedByAPI(t *testing.T) {
	outdir := t.TempDir()
	app := cdk8s.Testing_App(&cdk8s.AppProps{
		Outdir:                  &outdir,
		RecordConstructMetadata: chartBool(true),
	})
	chart := cdk8s.NewChart(app, chartString("chart1"), nil)
	chartNewObject(chart, "obj1", "Deployment")
	app.Synth()

	data, err := os.ReadFile(filepath.Join(outdir, "construct-metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	chartAssertJSONEqual(t, got, map[string]interface{}{
		"version": "1.0.0",
		"resources": map[string]interface{}{
			"chart1-obj1-c818e77f": map[string]interface{}{"path": "chart1/obj1"},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L381
func TestChartConstructMetadataRecordedWhenRequestedByEnvironment(t *testing.T) {
	t.Setenv("CDK8S_RECORD_CONSTRUCT_METADATA", "true")
	outdir := t.TempDir()
	app := cdk8s.Testing_App(&cdk8s.AppProps{Outdir: &outdir})
	chart := cdk8s.NewChart(app, chartString("chart1"), nil)
	chartNewObject(chart, "obj1", "Deployment")
	app.Synth()

	data, err := os.ReadFile(filepath.Join(outdir, "construct-metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	chartAssertJSONEqual(t, got, map[string]interface{}{
		"version": "1.0.0",
		"resources": map[string]interface{}{
			"chart1-obj1-c818e77f": map[string]interface{}{"path": "chart1/obj1"},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L410
func TestChartConstructMetadataNotRecordedWhenNotRequested(t *testing.T) {
	t.Setenv("CDK8S_RECORD_CONSTRUCT_METADATA", "")
	outdir := t.TempDir()
	app := cdk8s.Testing_App(&cdk8s.AppProps{Outdir: &outdir})
	chart := cdk8s.NewChart(app, chartString("chart1"), nil)
	chartNewObject(chart, "obj1", "Deployment")
	app.Synth()

	if _, err := os.Stat(filepath.Join(outdir, "construct-metadata.json")); !os.IsNotExist(err) {
		t.Fatalf("construct-metadata.json exists or stat failed unexpectedly: %v", err)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/chart.test.ts#L454
func TestChartApiObjectsReturnsAllAPIObjects(t *testing.T) {
	chart := chartChart(t)
	chartNewObject(chart, "obj1", "Deployment")
	cdk8s.NewApiObject(chart, chartString("obj2"), &cdk8s.ApiObjectProps{
		ApiVersion: chartString("v1"),
		Kind:       chartString("Foo"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: chartString("resource1")},
	})
	cdk8s.NewApiObject(chart, chartString("obj3"), &cdk8s.ApiObjectProps{
		ApiVersion: chartString("v1"),
		Kind:       chartString("Bar"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: chartString("resource1")},
	})

	kinds := make([]string, 0, len(*chart.ApiObjects()))
	for _, object := range *chart.ApiObjects() {
		kinds = append(kinds, *object.Kind())
	}
	sort.Strings(kinds)
	if want := []string{"Bar", "Deployment", "Foo"}; !reflect.DeepEqual(kinds, want) {
		t.Fatalf("kinds = %#v, want %#v", kinds, want)
	}
}
