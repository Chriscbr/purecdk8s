package cdk8s

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/purecdk8s/purecdk8s/constructs/v10"
)

func TestCronParity(t *testing.T) {
	tests := []struct {
		name string
		cron Cron
		want string
	}{
		{"every minute", Cron_EveryMinute(), "* * * * *"},
		{"hourly", Cron_Hourly(), "0 * * * *"},
		{"daily", Cron_Daily(), "0 0 * * *"},
		{"weekly", Cron_Weekly(), "0 0 * * 0"},
		{"monthly", Cron_Monthly(), "0 0 1 * *"},
		{"annually", Cron_Annually(), "0 0 1 1 *"},
		{"defaults", NewCron(nil), "* * * * *"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := stringValue(test.cron.ExpressionString()); got != test.want {
				t.Fatalf("expression = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDurationParity(t *testing.T) {
	tests := []struct {
		duration Duration
		human    string
		iso      string
		unit     string
	}{
		{Duration_Seconds(numberPointer(1.5)), "1 second 500 millis", "PT1.5S", "seconds"},
		{Duration_Minutes(numberPointer(90)), "1 hour 30 minutes", "PT1H30M", "minutes"},
		// This rounding quirk is present in cdk8s 2.70.85.
		{Duration_Hours(numberPointer(25)), "1 day 60 minutes", "PT1D1H", "hours"},
		{Duration_Days(numberPointer(1.5)), "1 day 12 hours", "PT1.5D", "days"},
		{Duration_Seconds(numberPointer(3661)), "1 hour 1 minute", "PT1H1M1S", "seconds"},
	}
	for _, test := range tests {
		if got := stringValue(test.duration.ToHumanString()); got != test.human {
			t.Errorf("human string = %q, want %q", got, test.human)
		}
		if got := stringValue(test.duration.ToIsoString()); got != test.iso {
			t.Errorf("ISO string = %q, want %q", got, test.iso)
		}
		if got := stringValue(test.duration.UnitLabel()); got != test.unit {
			t.Errorf("unit = %q, want %q", got, test.unit)
		}
	}

	parsed := Duration_Parse(stringPointer("PT1D2H3M4S"))
	if got := *parsed.ToMilliseconds(&TimeConversionOptions{Integral: boolPointer(false)}); got != 93_784_000 {
		t.Fatalf("parsed milliseconds = %v", got)
	}
	if got := stringValue(parsed.ToHumanString()); got != "1 day 2 hours" {
		t.Fatalf("parsed human string = %q", got)
	}
	assertPanics(t, func() { parsed.ToIsoString() })
	assertPanics(t, func() { Duration_Seconds(numberPointer(1.5)).ToSeconds(nil) })
	if got := *Duration_Seconds(numberPointer(1.5)).ToSeconds(&TimeConversionOptions{Integral: boolPointer(false)}); got != 1.5 {
		t.Fatalf("non-integral seconds = %v", got)
	}
}

func TestSizeParity(t *testing.T) {
	size := Size_Mebibytes(numberPointer(1.5))
	if got := stringValue(size.AsString()); got != "1.5Mi" {
		t.Fatalf("AsString = %q", got)
	}
	if got := *size.ToKibibytes(nil); got != 1536 {
		t.Fatalf("kibibytes = %v", got)
	}
	assertPanics(t, func() { size.ToGibibytes(nil) })
	if got := *size.ToGibibytes(&SizeConversionOptions{Rounding: SizeRoundingBehavior_FLOOR}); got != 0 {
		t.Fatalf("floor gibibytes = %v", got)
	}
	if got := *size.ToGibibytes(&SizeConversionOptions{Rounding: SizeRoundingBehavior_NONE}); got != 1.5/1024 {
		t.Fatalf("fractional gibibytes = %v", got)
	}
	if got := *size.ToMebibytes(nil); got != 1.5 {
		t.Fatalf("same-unit conversion = %v", got)
	}
	assertPanics(t, func() { Size_Kibibytes(numberPointer(-1)) })
	assertPanics(t, func() {
		Size_Mebibytes(numberPointer(1)).ToMebibytes(
			&SizeConversionOptions{Rounding: SizeRoundingBehavior("UNKNOWN")},
		)
	})
}

func TestNamesEdgeCases(t *testing.T) {
	root := constructs.NewRootConstruct(stringPointer("valid"))
	if got := stringValue(Names_ToDnsLabel(root, nil)); got != "valid" {
		t.Fatalf("single valid component = %q", got)
	}

	parent := constructs.NewConstruct(root, stringPointer("parent"))
	resource := constructs.NewConstruct(parent, stringPointer("Resource"))
	name := stringValue(Names_ToDnsLabel(resource, nil))
	if strings.Contains(strings.ToLower(name), "resource") {
		t.Fatalf("default child component was not omitted: %q", name)
	}
	if len(name) > 63 {
		t.Fatalf("name too long: %q", name)
	}

	t.Setenv("CDK8S_LEGACY_HASH", "1")
	legacy := stringValue(Names_ToDnsLabel(parent, nil))
	if legacy != "valid-parent-4508fc0b" {
		t.Fatalf("legacy name = %q", legacy)
	}

	maxLen := float64(7)
	assertPanics(t, func() { Names_ToDnsLabel(parent, &NameOptions{MaxLen: &maxLen}) })
	delimiter := "/"
	assertPanics(t, func() { Names_ToLabelValue(parent, &NameOptions{Delimiter: &delimiter}) })
}

func TestJsonPatchParity(t *testing.T) {
	document := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"name": "first"},
			map[string]interface{}{"name": "second"},
		},
		"a/b":  "old",
		"drop": true,
	}
	result := JsonPatch_Apply(
		document,
		JsonPatch_Add(stringPointer("/items/1"), map[string]interface{}{"name": "inserted"}),
		JsonPatch_Replace(stringPointer("/a~1b"), "new"),
		JsonPatch_Copy(stringPointer("/items/0"), stringPointer("/best")),
		JsonPatch_Move(stringPointer("/items/2"), stringPointer("/last")),
		JsonPatch_Remove(stringPointer("/drop")),
		JsonPatch_Test(stringPointer("/best/name"), "first"),
	)
	got := result.(map[string]interface{})
	want := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"name": "first"},
			map[string]interface{}{"name": "inserted"},
		},
		"a/b":  "new",
		"best": map[string]interface{}{"name": "first"},
		"last": map[string]interface{}{"name": "second"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("patched document = %#v, want %#v", got, want)
	}
	if document["a/b"] != "old" {
		t.Fatal("JsonPatch_Apply mutated its Go caller's input")
	}
	assertPanics(t, func() {
		JsonPatch_Apply(document, JsonPatch_Test(stringPointer("/a~1b"), "wrong"))
	})
	assertPanics(t, func() {
		var null *string
		JsonPatch_Apply(document, JsonPatch_Test(stringPointer("/missing"), null))
	})
	missingMove := JsonPatch_Apply(
		map[string]interface{}{"kept": true},
		JsonPatch_Move(stringPointer("/missing"), stringPointer("/destination")),
	)
	if _, found := missingMove.(map[string]interface{})["destination"]; found {
		t.Fatal("move from a missing path should not materialize an undefined destination")
	}
	if got := JsonPatch_Apply(document, JsonPatch_Remove(stringPointer(""))); got != nil {
		t.Fatalf("remove root = %#v, want nil", got)
	}
}

func TestYAMLParity(t *testing.T) {
	document := map[string]interface{}{
		"arr":  []interface{}{float64(1), "2"},
		"date": "2020-01-01",
		"nil":  nil,
		"yes":  "yes",
	}
	got := stringValue(Yaml_Stringify(document, nil, map[string]interface{}{"x": "y"}))
	want := `arr:
  - 1
  - "2"
date: "2020-01-01"
"yes": "yes"
---

---
x: "y"
`
	if got != want {
		t.Fatalf("YAML mismatch\n--- got ---\n%s--- want ---\n%s", got, want)
	}

	path := filepath.Join(t.TempDir(), "input.yaml")
	input := `---
{}
---
enabled: yes
mode: 0775
date: 2020-01-01
---
[]
---
null
---
- x
`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	docs := Yaml_Load(&path)
	wantDocs := []interface{}{
		map[string]interface{}{"date": "2020-01-01T00:00:00.000Z", "enabled": true, "mode": float64(509)},
		[]interface{}{"x"},
	}
	if !reflect.DeepEqual(*docs, wantDocs) {
		t.Fatalf("loaded docs = %#v, want %#v", *docs, wantDocs)
	}

	temp := Yaml_Tmp(docs)
	if filepath.Base(*temp) != "temp.yaml" || !strings.HasPrefix(filepath.Base(filepath.Dir(*temp)), "cdk8s-") {
		t.Fatalf("unexpected temp path: %s", *temp)
	}
}

func TestDependencyGraphTopologyAndCycle(t *testing.T) {
	root := constructs.NewRootConstruct(stringPointer("root"))
	a := constructs.NewConstruct(root, stringPointer("a"))
	b := constructs.NewConstruct(root, stringPointer("b"))
	c := constructs.NewConstruct(root, stringPointer("c"))
	a.Node().AddDependency(b)

	topology := *NewDependencyGraph(root.Node()).Topology()
	if indexOfConstruct(topology, b) >= indexOfConstruct(topology, a) {
		t.Fatalf("dependency is not before dependant: %#v", constructPaths(topology))
	}
	if indexOfConstruct(topology, c) < 0 {
		t.Fatal("orphan construct missing from topology")
	}

	b.Node().AddDependency(a)
	assertPanics(t, func() { NewDependencyGraph(root.Node()) })
}

func TestDependencyGraphMatchesJavaScriptNumericKeyOrder(t *testing.T) {
	root := constructs.NewRootConstruct(nil)
	constructs.NewConstruct(root, stringPointer("10"))
	constructs.NewConstruct(root, stringPointer("2"))
	constructs.NewConstruct(root, stringPointer("01"))

	got := constructPaths(*NewDependencyGraph(root.Node()).Topology())
	want := []string{"2", "10", "", "01"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("topology paths = %#v, want %#v", got, want)
	}
}

type (
	testCronOverride    struct{ Cron }
	testVertexOverride  struct{ DependencyVertex }
	testGraphOverride   struct{ DependencyGraph }
	testContextOverride struct{ ResolutionContext }
)

func TestUtilityOverrideConstructorsPopulateEmbeddedInterfaces(t *testing.T) {
	cron := &testCronOverride{}
	NewCron_Override(cron, &CronOptions{Hour: stringPointer("2")})
	if got := stringValue(cron.ExpressionString()); got != "* 2 * * *" {
		t.Fatalf("overridden cron expression = %q", got)
	}

	root := constructs.NewRootConstruct(stringPointer("root"))
	child := constructs.NewConstruct(root, stringPointer("child"))
	parentVertex := &testVertexOverride{}
	childVertex := &testVertexOverride{}
	NewDependencyVertex_Override(parentVertex, root)
	NewDependencyVertex_Override(childVertex, child)
	parentVertex.AddChild(childVertex)
	if len(*childVertex.Inbound()) != 1 {
		t.Fatalf("overridden vertex inbound = %#v", *childVertex.Inbound())
	}
	if (*childVertex.Inbound())[0] != parentVertex {
		t.Fatalf("overridden vertex lost its outer identity: %#v", (*childVertex.Inbound())[0])
	}

	graph := &testGraphOverride{}
	NewDependencyGraph_Override(graph, root.Node())
	if graph.Root() == nil || len(*graph.Topology()) != 2 {
		t.Fatalf("overridden graph topology = %#v", graph.Topology())
	}

	app := NewApp(nil)
	chart := NewChart(app, stringPointer("chart"), nil)
	object := NewApiObject(chart, stringPointer("object"), &ApiObjectProps{
		ApiVersion: stringPointer("v1"),
		Kind:       stringPointer("Thing"),
	})
	key := []*string{stringPointer("spec")}
	context := &testContextOverride{}
	NewResolutionContext_Override(context, object, &key, "old")
	context.ReplaceValue("new")
	if !boolValue(context.Replaced()) || context.ReplacedValue() != "new" {
		t.Fatalf("overridden context replacement = %#v", context.ReplacedValue())
	}
}

type testReplacingResolver struct {
	seen [][]string
}

func (r *testReplacingResolver) Resolve(context ResolutionContext) {
	value, ok := context.Value().(string)
	if !ok || value != "replace-me" {
		return
	}
	key := make([]string, 0, len(*context.Key()))
	for _, item := range *context.Key() {
		key = append(key, stringValue(item))
	}
	r.seen = append(r.seen, key)
	context.ReplaceValue("replaced")
}

type testMetadataNameResolver struct{}

func (r *testMetadataNameResolver) Resolve(context ResolutionContext) {
	key := *context.Key()
	if len(key) != 2 || stringValue(key[0]) != "metadata" || stringValue(key[1]) != "name" {
		return
	}
	switch name := context.Value().(type) {
	case string:
		if !strings.HasPrefix(name, "prefix-") {
			context.ReplaceValue("prefix-" + name)
		}
	case *string:
		if name != nil && !strings.HasPrefix(*name, "prefix-") {
			context.ReplaceValue("prefix-" + *name)
		}
	}
}

type testAnyProducer struct{ value interface{} }

func (p *testAnyProducer) Produce() interface{} {
	return p.value
}

type testImplicitToken struct{ value string }

func (t *testImplicitToken) Resolve() *string {
	return &t.value
}

type testNumberStringUnion struct{ value *float64 }

func (u *testNumberStringUnion) Value() interface{} {
	return u.value
}

func TestResolverPrimitives(t *testing.T) {
	resolver := &testReplacingResolver{}
	resolvers := []IResolver{resolver}
	app := NewApp(&AppProps{Resolvers: &resolvers})
	chart := NewChart(app, stringPointer("chart"), nil)
	object := NewApiObjectWithManifest(chart, stringPointer("object"), &ApiObjectProps{
		ApiVersion: stringPointer("v1"),
		Kind:       stringPointer("Thing"),
	}, map[string]interface{}{"spec": map[string]interface{}{"value": "replace-me"}})

	rendered := object.ToJson().(map[string]interface{})
	if got := rendered["spec"].(map[string]interface{})["value"]; got != "replaced" {
		t.Fatalf("custom resolver result = %#v", got)
	}
	if len(resolver.seen) == 0 || !reflect.DeepEqual(resolver.seen[0], []string{"spec", "value"}) {
		t.Fatalf("resolver key = %#v", resolver.seen)
	}

	if got := resolveValue(Lazy_Any(&testAnyProducer{value: "lazy"}), object); got != "lazy" {
		t.Fatalf("lazy result = %#v", got)
	}
	if got := resolveValue(&testImplicitToken{value: "token"}, object); got != "token" {
		t.Fatalf("implicit token result = %#v", got)
	}
	union := resolveValue(&testNumberStringUnion{value: numberPointer(42)}, object)
	if !reflect.DeepEqual(union, map[string]interface{}{"value": float64(42)}) {
		t.Fatalf("union result = %#v", union)
	}
}

func TestResolverReceivesFullMetadataPath(t *testing.T) {
	resolvers := []IResolver{&testMetadataNameResolver{}}
	app := NewApp(&AppProps{Resolvers: &resolvers})
	chart := NewChart(app, stringPointer("chart"), nil)
	object := NewApiObject(chart, stringPointer("object"), &ApiObjectProps{
		ApiVersion: stringPointer("v1"),
		Kind:       stringPointer("Thing"),
		Metadata:   &ApiObjectMetadata{Name: stringPointer("settings")},
	})

	rendered := object.ToJson().(map[string]interface{})
	metadata := rendered["metadata"].(map[string]interface{})
	if got := metadata["name"]; got != "prefix-settings" {
		t.Fatalf("resolved metadata.name = %#v, want prefix-settings", got)
	}
}

func numberPointer(value float64) *float64 {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}

func stringPointer(value string) *string {
	return &value
}

func assertPanics(t *testing.T, callback func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	callback()
}

func indexOfConstruct(values []constructs.IConstruct, target constructs.IConstruct) int {
	for index, value := range values {
		if value == target {
			return index
		}
	}
	return -1
}

func constructPaths(values []constructs.IConstruct) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = stringValue(value.Node().Path())
	}
	return result
}
