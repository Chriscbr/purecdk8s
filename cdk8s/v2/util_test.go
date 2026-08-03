package cdk8s_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func utilNewObject(t *testing.T, manifest map[string]interface{}) cdk8s.ApiObject {
	t.Helper()
	outdir := t.TempDir()
	app := cdk8s.NewApp(&cdk8s.AppProps{Outdir: &outdir})
	chart := cdk8s.NewChart(app, jsii.String("chart"), nil)
	return cdk8s.NewApiObjectWithManifest(chart, jsii.String("object"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Example"),
	}, manifest)
}

func utilMap(t *testing.T, value interface{}) map[string]interface{} {
	t.Helper()
	result, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("value has type %T, want map[string]interface{}", value)
	}
	return result
}

func utilMetadata(t *testing.T, values map[string]interface{}) map[string]interface{} {
	t.Helper()
	object := utilNewObject(t, map[string]interface{}{})
	metadata := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{ApiObject: object})
	for key, value := range values {
		key := key
		metadata.Add(&key, value)
	}
	return utilMap(t, metadata.ToJson())
}

func utilRequirePanicContains(t *testing.T, want string, callback func()) {
	t.Helper()
	defer func() {
		value := recover()
		if value == nil {
			t.Fatalf("expected panic containing %q", want)
		}
		if got := fmt.Sprint(value); !strings.Contains(got, want) {
			t.Fatalf("panic = %q, want it to contain %q", got, want)
		}
	}()
	callback()
}

type utilDummy func()

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/util.test.ts#L5
func TestUtilSanitizeValueDefaultOptions(t *testing.T) {
	var undefined *string
	object := utilNewObject(t, map[string]interface{}{
		"values": map[string]interface{}{
			"null":        nil,
			"undefined":   undefined,
			"emptyObject": map[string]interface{}{},
			"emptyArray":  []interface{}{},
			"number":      1,
			"object":      map[string]interface{}{"hello": 123},
			"array":       []interface{}{1, 2, 3},
			"arrayValue":  map[string]interface{}{"xoo": 123, "foo": []interface{}{}},
			"nestedObject": map[string]interface{}{
				"xoo": map[string]interface{}{},
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"zoo": undefined,
						"hey": map[string]interface{}{},
						"me":  123,
					},
				},
			},
			"nestedArray": map[string]interface{}{
				"xoo": 123,
				"foo": []interface{}{1, 2, map[string]interface{}{
					"foo": 123,
					"bar": undefined,
					"zoo": []interface{}{},
				}, 3},
			},
			"mixedEmptyValues": []interface{}{1, 2, 3, []interface{}{}, map[string]interface{}{}, 4},
		},
	})
	values := utilMap(t, utilMap(t, object.ToJson())["values"])

	if _, found := values["null"]; found {
		t.Fatal("null was retained")
	}
	if _, found := values["undefined"]; found {
		t.Fatal("typed nil was retained")
	}
	want := map[string]interface{}{
		"emptyObject": map[string]interface{}{},
		"emptyArray":  []interface{}{},
		"number":      float64(1),
		"object":      map[string]interface{}{"hello": float64(123)},
		"array":       []interface{}{float64(1), float64(2), float64(3)},
		"arrayValue":  map[string]interface{}{"xoo": float64(123), "foo": []interface{}{}},
		"nestedObject": map[string]interface{}{
			"xoo": map[string]interface{}{},
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"hey": map[string]interface{}{},
					"me":  float64(123),
				},
			},
		},
		"nestedArray": map[string]interface{}{
			"xoo": float64(123),
			"foo": []interface{}{float64(1), float64(2), map[string]interface{}{
				"foo": float64(123),
				"zoo": []interface{}{},
			}, float64(3)},
		},
		"mixedEmptyValues": []interface{}{float64(1), float64(2), float64(3), []interface{}{}, map[string]interface{}{}, float64(4)},
	}
	delete(values, "null")
	delete(values, "undefined")
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("sanitized values = %#v, want %#v", values, want)
	}

	bad := utilNewObject(t, map[string]interface{}{"bad": utilDummy(func() {})})
	utilRequirePanicContains(t, "can't render non-simple object of type 'cdk8s_test.utilDummy'", func() {
		bad.ToJson()
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/util.test.ts#L23
func TestUtilSanitizeValueFiltersEmptyArrays(t *testing.T) {
	if got := utilMetadata(t, map[string]interface{}{"value": []interface{}{}}); len(got) != 0 {
		t.Fatalf("empty array metadata = %#v, want empty", got)
	}

	got := utilMetadata(t, map[string]interface{}{
		"foo": []interface{}{},
		"bar": []interface{}{1, 2},
	})
	want := map[string]interface{}{"bar": []interface{}{float64(1), float64(2)}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("metadata = %#v, want %#v", got, want)
	}

	// Metadata is the public caller of array filtering and intentionally also
	// filters empty objects, so the now-empty outer object is omitted as well.
	if got := utilMetadata(t, map[string]interface{}{
		"foo": map[string]interface{}{"bar": []interface{}{}},
	}); len(got) != 0 {
		t.Fatalf("nested empty array metadata = %#v, want empty", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/util.test.ts#L33
func TestUtilSanitizeValueFiltersEmptyObjects(t *testing.T) {
	if got := utilMetadata(t, map[string]interface{}{"value": map[string]interface{}{}}); len(got) != 0 {
		t.Fatalf("empty object metadata = %#v, want empty", got)
	}

	got := utilMetadata(t, map[string]interface{}{
		"foo": map[string]interface{}{},
		"bar": map[string]interface{}{"hey": "there"},
	})
	want := map[string]interface{}{"bar": map[string]interface{}{"hey": "there"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("metadata = %#v, want %#v", got, want)
	}

	if got := utilMetadata(t, map[string]interface{}{
		"foo": map[string]interface{}{"bar": map[string]interface{}{}},
	}); len(got) != 0 {
		t.Fatalf("nested empty object metadata = %#v, want empty", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/util.test.ts#L43
func TestUtilSanitizeValueSortsKeys(t *testing.T) {
	input := map[string]interface{}{
		"zzz": 999,
		"aaa": 111,
		"nested": map[string]interface{}{
			"foo": map[string]interface{}{
				"zag": []interface{}{1, 2, 3},
				"bar": "1111",
			},
		},
	}

	render := func() string {
		manifest := utilMap(t, utilNewObject(t, input).ToJson())
		selected := map[string]interface{}{
			"zzz":    manifest["zzz"],
			"aaa":    manifest["aaa"],
			"nested": manifest["nested"],
		}
		data, err := json.MarshalIndent(selected, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}

	want := "{\n  \"aaa\": 111,\n  \"nested\": {\n    \"foo\": {\n      \"bar\": \"1111\",\n      \"zag\": [\n        1,\n        2,\n        3\n      ]\n    }\n  },\n  \"zzz\": 999\n}"
	if got := render(); got != want {
		t.Fatalf("sorted JSON = %s, want %s", got, want)
	}

	// Go maps have deliberately unspecified iteration order. Disabling cdk8s'
	// JavaScript object-key sorting therefore preserves the same Go value; a
	// deterministic encoder still emits its canonical order.
	t.Setenv("CDK8S_DISABLE_SORT", "1")
	if got := render(); got != want {
		t.Fatalf("JSON with sorting disabled = %s, want the same Go-map value %s", got, want)
	}
}
