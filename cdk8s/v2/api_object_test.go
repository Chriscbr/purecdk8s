package cdk8s_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	constructs "github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func apiObjectAssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	gotJSON, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("marshal actual value: %v", err)
	}
	wantJSON, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatalf("marshal expected value: %v", err)
	}
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("value mismatch\n--- got ---\n%s\n--- want ---\n%s", gotJSON, wantJSON)
	}
}

func apiObjectRequirePanicContains(t *testing.T, want string, callback func()) {
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

type apiObjectProducer struct{ value interface{} }

func (p *apiObjectProducer) Produce() interface{} { return p.value }

type apiObjectImplicitToken struct{ value interface{} }

func (t *apiObjectImplicitToken) Resolve() interface{} { return t.value }

type apiObjectFixedNameChart struct{ cdk8s.Chart }

func (c *apiObjectFixedNameChart) GenerateObjectName(cdk8s.ApiObject) *string {
	return jsii.String("fixed!")
}

type apiObjectReplaceStringsResolver struct{}

func (r *apiObjectReplaceStringsResolver) Resolve(context cdk8s.ResolutionContext) {
	value, ok := context.Value().(string)
	if ok && value != "newValue" {
		context.ReplaceValue("newValue")
	}
}

type apiObjectTypeResolver struct{}

func (r *apiObjectTypeResolver) Resolve(context cdk8s.ResolutionContext) {
	for _, component := range *context.Key() {
		if component != nil && *component == "type" && context.Value() != "newValue1" {
			context.ReplaceValue("newValue1")
			return
		}
	}
}

type apiObjectArrayResolver struct{}

func (r *apiObjectArrayResolver) Resolve(context cdk8s.ResolutionContext) {
	for _, component := range *context.Key() {
		if component != nil && *component == "someArray" && context.Value() != "newValue2" {
			context.ReplaceValue("newValue2")
			return
		}
	}
}

type apiObjectResolvable struct{}

func (r *apiObjectResolvable) Resolver() interface{} { return "blah" }

type apiObjectAnonymousResolver struct{}

func (r *apiObjectAnonymousResolver) Resolve(context cdk8s.ResolutionContext) {
	if resolvable, ok := context.Value().(interface{ Resolver() interface{} }); ok {
		context.ReplaceValue(resolvable.Resolver())
	}
}

type apiObjectIntOrString struct{ value interface{} }

func (v *apiObjectIntOrString) Value() interface{} { return v.value }

type apiObjectL1 struct{ cdk8s.ApiObject }

func apiObjectNewL1(scope constructs.Construct, id string, surge *apiObjectIntOrString) *apiObjectL1 {
	object := &apiObjectL1{}
	apiVersion, kind := "v1", "Kind"
	cdk8s.NewApiObjectWithManifest_Override(object, scope, &id, &cdk8s.ApiObjectProps{
		ApiVersion: &apiVersion,
		Kind:       &kind,
	}, map[string]interface{}{
		"spec": map[string]interface{}{"surge": surge},
	})
	return object
}

func (l *apiObjectL1) ToJson() interface{} {
	resolved := l.ApiObject.ToJson().(map[string]interface{})
	spec := resolved["spec"].(map[string]interface{})
	surge := spec["surge"].(map[string]interface{})
	return map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Kind",
		"metadata":   l.Metadata().ToJson(),
		"spec":       map[string]interface{}{"surge": surge["value"]},
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L14
func TestApiObjectMinimalConfiguration(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := cdk8s.NewChart(app, jsii.String("test"), nil)
	cdk8s.NewApiObject(chart, jsii.String("my-resource"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("MyResource"),
	})

	apiObjectAssertEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "MyResource",
			"metadata": map[string]interface{}{
				"name": "test-my-c8487bf7",
			},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L26
func TestApiObjectPrintedYAMLIsAlphabetical(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := cdk8s.NewChart(app, jsii.String("test"), nil)
	cdk8s.NewApiObjectWithManifest(chart, jsii.String("my-resource"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("MyResource"),
	}, map[string]interface{}{
		"kind": "MyResource",
		"spec": map[string]interface{}{
			"secondProperty": map[string]interface{}{
				"innerThirdProperty":  "!",
				"beforeThirdProperty": "world",
			},
			"firstProperty": "hello",
		},
		"metadata": map[string]interface{}{
			"meta": map[string]interface{}{"zzz": "hello", "aaa": 123},
		},
		"apiVersion": "v1",
	})

	actual, err := json.MarshalIndent(*cdk8s.Testing_Synth(chart), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want := `[
  {
    "apiVersion": "v1",
    "kind": "MyResource",
    "metadata": {
      "meta": {
        "aaa": 123,
        "zzz": "hello"
      },
      "name": "test-my-c8487bf7"
    },
    "spec": {
      "firstProperty": "hello",
      "secondProperty": {
        "beforeThirdProperty": "world",
        "innerThirdProperty": "!"
      }
    }
  }
]`
	if got := string(actual); got != want {
		t.Fatalf("serialized manifest mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L54
func TestApiObjectDisableSortEnvironmentVariable(t *testing.T) {
	object := cdk8s.NewApiObjectWithManifest(cdk8s.Testing_Chart(), jsii.String("my-api-object"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Dummy"),
	}, map[string]interface{}{
		"hello": map[string]interface{}{
			"zzz": 123,
			"aaa": 333,
			"nested": map[string]interface{}{
				"yyy": "hello",
				"bbb": 123,
			},
		},
	})

	want := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Dummy",
		"metadata": map[string]interface{}{
			"name": "test-my-api-object-c8e6fbed",
		},
		"hello": map[string]interface{}{
			"aaa": float64(333),
			"nested": map[string]interface{}{
				"bbb": float64(123),
				"yyy": "hello",
			},
			"zzz": float64(123),
		},
	}
	apiObjectAssertEqual(t, object.ToJson(), want)

	// Go maps intentionally have no observable insertion order. The portable
	// assertion is that disabling sorting preserves every manifest value.
	t.Setenv("CDK8S_DISABLE_SORT", "1")
	apiObjectAssertEqual(t, object.ToJson(), want)
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L84
func TestApiObjectAddDependency(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := cdk8s.NewChart(app, jsii.String("chart1"), nil)
	object1 := cdk8s.NewApiObject(chart, jsii.String("obj1"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Kind1")})
	object2 := cdk8s.NewApiObject(chart, jsii.String("obj2"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Kind2")})
	object3 := cdk8s.NewApiObject(chart, jsii.String("obj3"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Kind3")})

	object1.AddDependency(object2, object3)
	dependencies := *object1.Node().Dependencies()
	if len(dependencies) != 2 || dependencies[0] != object2 || dependencies[1] != object3 {
		t.Fatalf("dependencies = %#v, want [%v, %v]", dependencies, object2, object3)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L108
func TestApiObjectSynthesizedResourceNameIsBasedOnPath(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := cdk8s.NewChart(app, jsii.String("test"), nil)
	cdk8s.NewApiObject(chart, jsii.String("my-resource"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("MyResource")})
	scope := constructs.NewConstruct(chart, jsii.String("scope"))
	cdk8s.NewApiObject(scope, jsii.String("my-resource"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("MyResource")})

	apiObjectAssertEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "MyResource",
			"metadata":   map[string]interface{}{"name": "test-my-c8487bf7"},
		},
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "MyResource",
			"metadata":   map[string]interface{}{"name": "test-scope-my-c8fafaf7"},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L129
func TestApiObjectExplicitNameIsRespected(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := cdk8s.NewChart(app, jsii.String("test"), nil)
	cdk8s.NewApiObject(chart, jsii.String("my-resource"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("MyResource"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("boom")},
	})

	apiObjectAssertEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "MyResource",
			"metadata":   map[string]interface{}{"name": "boom"},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L147
func TestApiObjectSpecIsSynthesizedAsIs(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := cdk8s.NewChart(app, jsii.String("test"), nil)
	cdk8s.NewApiObjectWithManifest(chart, jsii.String("resource"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("ResourceType")}, map[string]interface{}{
		"spec": map[string]interface{}{
			"prop1": "hello",
			"prop2": map[string]interface{}{"world": 123},
		},
	})

	apiObjectAssertEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ResourceType",
			"metadata":   map[string]interface{}{"name": "test-c8c5facf"},
			"spec": map[string]interface{}{
				"prop1": "hello",
				"prop2": map[string]interface{}{"world": float64(123)},
			},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L168
func TestApiObjectDataCanSpecifyResourceData(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := cdk8s.NewChart(app, jsii.String("test"), nil)
	cdk8s.NewApiObjectWithManifest(chart, jsii.String("resource"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("ResourceType")}, map[string]interface{}{
		"data": map[string]interface{}{"boom": 123},
	})

	apiObjectAssertEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"data":       map[string]interface{}{"boom": float64(123)},
			"kind":       "ResourceType",
			"metadata":   map[string]interface{}{"name": "test-c8c5facf"},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L186
func TestApiObjectNamingLogicCanBeOverridden(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := &apiObjectFixedNameChart{}
	cdk8s.NewChart_Override(chart, app, jsii.String("my-chart"), nil)
	object := cdk8s.NewApiObject(chart, jsii.String("my-object"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("MyKind")})

	if got := *object.Name(); got != "fixed!" {
		t.Fatalf("object name = %q, want %q", got, "fixed!")
	}
	apiObjectAssertEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "MyKind",
			"metadata":   map[string]interface{}{"name": "fixed!"},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L214
func TestApiObjectDefaultNamespaceAtChartLevel(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := cdk8s.NewChart(app, jsii.String("chart"), &cdk8s.ChartProps{Namespace: jsii.String("ns1")})
	group := constructs.NewConstruct(chart, jsii.String("group1"))
	cdk8s.NewApiObject(group, jsii.String("obj1"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Kind1")})
	cdk8s.NewApiObject(group, jsii.String("obj2"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v2"),
		Kind:       jsii.String("Kind2"),
		Metadata:   &cdk8s.ApiObjectMetadata{Namespace: jsii.String("foobar")},
	})

	apiObjectAssertEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Kind1",
			"metadata": map[string]interface{}{
				"name":      "chart-group1-obj1-c885aeec",
				"namespace": "ns1",
			},
		},
		map[string]interface{}{
			"apiVersion": "v2",
			"kind":       "Kind2",
			"metadata": map[string]interface{}{
				"name":      "chart-group1-obj2-c81931d8",
				"namespace": "foobar",
			},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L249
func TestApiObjectChartLabelsAppliedToAllObjects(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	labels := map[string]*string{"foo": jsii.String("ffffffffff"), "bar": jsii.String("bbbbbb")}
	chart := cdk8s.NewChart(app, jsii.String("my-chart"), &cdk8s.ChartProps{Labels: &labels})
	cdk8s.NewApiObject(chart, jsii.String("obj1"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Obj1")})
	group := constructs.NewConstruct(chart, jsii.String("group"))
	objectLabels := map[string]*string{"foo": jsii.String("override by object"), "zoo": jsii.String("zoo1")}
	cdk8s.NewApiObject(group, jsii.String("obj2"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v2"),
		Kind:       jsii.String("Obj2"),
		Metadata:   &cdk8s.ApiObjectMetadata{Labels: &objectLabels},
	})

	apiObjectAssertEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Obj1",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{"bar": "bbbbbb", "foo": "ffffffffff"},
				"name":   "my-chart-obj1-c880bc50",
			},
		},
		map[string]interface{}{
			"apiVersion": "v2",
			"kind":       "Obj2",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{"bar": "bbbbbb", "foo": "override by object", "zoo": "zoo1"},
				"name":   "my-chart-group-obj2-c824cfcd",
			},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L303
func TestApiObjectOfFailsWithoutDefaultChild(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	parent := constructs.NewConstruct(chart, jsii.String("hello"))
	apiObjectRequirePanicContains(t, "cannot find a (direct or indirect) child of type ApiObject", func() {
		cdk8s.ApiObject_Of(parent)
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L313
func TestApiObjectOfReturnsApiObject(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	object := cdk8s.NewApiObject(chart, jsii.String("my-obj"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Foo")})
	if got := cdk8s.ApiObject_Of(object); got != object {
		t.Fatalf("ApiObject_Of(object) = %v, want original object", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L325
func TestApiObjectOfReturnsDirectChild(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	parent := constructs.NewConstruct(chart, jsii.String("l2"))
	object := cdk8s.NewApiObject(parent, jsii.String("Default"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Foo")})
	if got := cdk8s.ApiObject_Of(parent); got != object {
		t.Fatalf("ApiObject_Of(parent) = %v, want direct child", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L340
func TestApiObjectOfReturnsIndirectChild(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	parent1 := constructs.NewConstruct(chart, jsii.String("l3"))
	parent2 := constructs.NewConstruct(parent1, jsii.String("Default"))
	object := cdk8s.NewApiObject(parent2, jsii.String("Default"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Foo")})
	if got := cdk8s.ApiObject_Of(parent1); got != object {
		t.Fatalf("ApiObject_Of(parent) = %v, want indirect child", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L358
func TestApiObjectJsonPatchAppliedAfterSynthesis(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	object := cdk8s.NewApiObjectWithManifest(chart, jsii.String("obj"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Obj")}, map[string]interface{}{
		"spec": map[string]interface{}{
			"foo": 1234,
			"bar": map[string]interface{}{"baz": []interface{}{1, 2, 3}},
		},
	})
	object.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec/bar/baz/3"), 4))
	object.AddJsonPatch(cdk8s.JsonPatch_Remove(jsii.String("/spec/foo")))
	object.AddJsonPatch(cdk8s.JsonPatch_Copy(jsii.String("/apiVersion"), jsii.String("/spec/apiVersion")))

	apiObjectAssertEqual(t, *cdk8s.Testing_Synth(chart), []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Obj",
			"metadata":   map[string]interface{}{"name": "test-obj-c8686f96"},
			"spec": map[string]interface{}{
				"apiVersion": "v1",
				"bar":        map[string]interface{}{"baz": []interface{}{float64(1), float64(2), float64(3), float64(4)}},
			},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L396
func TestApiObjectCompoundResolution(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := cdk8s.NewChart(app, jsii.String("test"), nil)
	object := cdk8s.NewApiObjectWithManifest(chart, jsii.String("resource1"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Resource1")}, map[string]interface{}{
		"spec": map[string]interface{}{
			"foo": cdk8s.Lazy_Any(&apiObjectProducer{value: &apiObjectImplicitToken{value: 123}}),
		},
	})

	apiObjectAssertEqual(t, object.ToJson(), map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Resource1",
		"metadata":   map[string]interface{}{"name": "test-resource1-c85cb0fc"},
		"spec":       map[string]interface{}{"foo": float64(123)},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L429
func TestApiObjectCustomResolver(t *testing.T) {
	resolvers := []cdk8s.IResolver{&apiObjectReplaceStringsResolver{}}
	app := cdk8s.Testing_App(&cdk8s.AppProps{Resolvers: &resolvers})
	chart := cdk8s.NewChart(app, jsii.String("Chart"), nil)
	object := cdk8s.NewApiObjectWithManifest(chart, jsii.String("ApiObject"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Service")}, map[string]interface{}{
		"metadata": map[string]interface{}{"foo": "bar"},
		"spec":     map[string]interface{}{"type": "LoadBalancer", "someArray": []interface{}{1, 2}},
	})

	apiObjectAssertEqual(t, object.ToJson(), map[string]interface{}{
		"apiVersion": "newValue",
		"kind":       "newValue",
		"metadata":   map[string]interface{}{"foo": "newValue", "name": "newValue"},
		"spec":       map[string]interface{}{"someArray": []interface{}{float64(1), float64(2)}, "type": "newValue"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L473
func TestApiObjectMultipleCustomResolvers(t *testing.T) {
	resolvers := []cdk8s.IResolver{&apiObjectTypeResolver{}, &apiObjectArrayResolver{}}
	app := cdk8s.Testing_App(&cdk8s.AppProps{Resolvers: &resolvers})
	chart := cdk8s.NewChart(app, jsii.String("Chart"), nil)
	object := cdk8s.NewApiObjectWithManifest(chart, jsii.String("ApiObject"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Service")}, map[string]interface{}{
		"metadata": map[string]interface{}{"foo": "bar"},
		"spec":     map[string]interface{}{"type": "LoadBalancer", "someArray": []interface{}{1, 2}},
	})

	apiObjectAssertEqual(t, object.ToJson(), map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata":   map[string]interface{}{"foo": "bar", "name": "chart-apiobject-c830d7bd"},
		"spec":       map[string]interface{}{"someArray": "newValue2", "type": "newValue1"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L521
func TestApiObjectAnonymousObjectCustomResolver(t *testing.T) {
	resolvers := []cdk8s.IResolver{&apiObjectAnonymousResolver{}}
	app := cdk8s.Testing_App(&cdk8s.AppProps{Resolvers: &resolvers})
	chart := cdk8s.NewChart(app, jsii.String("Chart"), nil)
	object := cdk8s.NewApiObjectWithManifest(chart, jsii.String("ApiObject"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("Service")}, map[string]interface{}{
		"metadata": map[string]interface{}{"foo": "bar"},
		"spec":     map[string]interface{}{"type": &apiObjectResolvable{}},
	})

	apiObjectAssertEqual(t, object.ToJson(), map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata":   map[string]interface{}{"foo": "bar", "name": "chart-apiobject-c830d7bd"},
		"spec":       map[string]interface{}{"type": "blah"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L566
func TestApiObjectCanResolveL1(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	object := apiObjectNewL1(chart, "L1", &apiObjectIntOrString{value: 500})

	apiObjectAssertEqual(t, object.ToJson(), map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Kind",
		"metadata":   map[string]interface{}{"name": "test-l1-c8c430b5"},
		"spec":       map[string]interface{}{"surge": float64(500)},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/api-object.test.ts#L630
func TestApiObjectToJSONErrorMessage(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	object := cdk8s.NewApiObjectWithManifest(chart, jsii.String("ConfigMap"), &cdk8s.ApiObjectProps{ApiVersion: jsii.String("v1"), Kind: jsii.String("ConfigMap")}, map[string]interface{}{
		"data": map[string]interface{}{"size": cdk8s.Size_Gibibytes(jsii.Number(5))},
	})
	want := fmt.Sprintf("Failed serializing construct at path '%s' with name '%s': Error: can't render non-simple object of type 'Size'", *object.Node().Path(), *object.Name())
	apiObjectRequirePanicContains(t, want, func() {
		object.ToJson()
	})
}
