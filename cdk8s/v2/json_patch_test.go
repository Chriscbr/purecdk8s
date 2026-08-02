package cdk8s_test

import (
	"reflect"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

func TestJsonPatchFactories(t *testing.T) {
	// The TypeScript tests inspect a private serialization hook. These black-box
	// ports verify the same operation fields through their public Apply behavior.

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/json-patch.test.ts#L4
	t.Run("add", func(t *testing.T) {
		gotObject := cdk8s.JsonPatch_Apply(
			map[string]interface{}{},
			cdk8s.JsonPatch_Add(coreString("/foo"), map[string]interface{}{"hello": float64(1234)}),
		)
		wantObject := map[string]interface{}{"foo": map[string]interface{}{"hello": float64(1234)}}
		if !reflect.DeepEqual(gotObject, wantObject) {
			t.Errorf("object-valued result = %#v, want %#v", gotObject, wantObject)
		}

		gotScalar := cdk8s.JsonPatch_Apply(
			map[string]interface{}{"foo": map[string]interface{}{}},
			cdk8s.JsonPatch_Add(coreString("/foo/bar"), float64(123)),
		)
		wantScalar := map[string]interface{}{"foo": map[string]interface{}{"bar": float64(123)}}
		if !reflect.DeepEqual(gotScalar, wantScalar) {
			t.Errorf("scalar-valued result = %#v, want %#v", gotScalar, wantScalar)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/json-patch.test.ts#L9
	t.Run("remove", func(t *testing.T) {
		got := cdk8s.JsonPatch_Apply(
			map[string]interface{}{"foo": map[string]interface{}{"hello": []interface{}{"first", "second"}}},
			cdk8s.JsonPatch_Remove(coreString("/foo/hello/0")),
		)
		want := map[string]interface{}{"foo": map[string]interface{}{"hello": []interface{}{"second"}}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("result = %#v, want %#v", got, want)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/json-patch.test.ts#L13
	t.Run("replace", func(t *testing.T) {
		got := cdk8s.JsonPatch_Apply(
			map[string]interface{}{"foo": map[string]interface{}{"hello": []interface{}{"old"}}},
			cdk8s.JsonPatch_Replace(coreString("/foo/hello/0"), map[string]interface{}{"value": float64(1234)}),
		)
		want := map[string]interface{}{"foo": map[string]interface{}{"hello": []interface{}{map[string]interface{}{"value": float64(1234)}}}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("result = %#v, want %#v", got, want)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/json-patch.test.ts#L17
	t.Run("copy", func(t *testing.T) {
		got := cdk8s.JsonPatch_Apply(
			map[string]interface{}{"from": map[string]interface{}{"value": float64(1)}},
			cdk8s.JsonPatch_Copy(coreString("/from"), coreString("/to")),
		)
		want := map[string]interface{}{
			"from": map[string]interface{}{"value": float64(1)},
			"to":   map[string]interface{}{"value": float64(1)},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("result = %#v, want %#v", got, want)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/json-patch.test.ts#L21
	t.Run("move", func(t *testing.T) {
		got := cdk8s.JsonPatch_Apply(
			map[string]interface{}{"from": map[string]interface{}{"value": float64(1)}},
			cdk8s.JsonPatch_Move(coreString("/from"), coreString("/to")),
		)
		want := map[string]interface{}{"to": map[string]interface{}{"value": float64(1)}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("result = %#v, want %#v", got, want)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/json-patch.test.ts#L25
	t.Run("test", func(t *testing.T) {
		document := map[string]interface{}{"path": "value"}
		if got := cdk8s.JsonPatch_Apply(document, cdk8s.JsonPatch_Test(coreString("/path"), "value")); !reflect.DeepEqual(got, document) {
			t.Fatalf("result = %#v, want %#v", got, document)
		}
		coreRequirePanicContains(t, "Test operation failed", func() {
			cdk8s.JsonPatch_Apply(document, cdk8s.JsonPatch_Test(coreString("/path"), "wrong"))
		})
	})
}

func TestJsonPatchApply(t *testing.T) {
	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/json-patch.test.ts#L30
	t.Run("apply", func(t *testing.T) {
		input := map[string]interface{}{
			"hello": float64(123),
			"world": map[string]interface{}{
				"foo": []interface{}{"bar", "baz"},
				"hi":  map[string]interface{}{"there": "hello-again"},
			},
		}
		got := cdk8s.JsonPatch_Apply(
			input,
			cdk8s.JsonPatch_Replace(coreString("/world/hi/there"), "goodbye"),
			cdk8s.JsonPatch_Add(coreString("/world/foo/"), "boom"),
			cdk8s.JsonPatch_Remove(coreString("/hello")),
		)
		want := map[string]interface{}{
			"world": map[string]interface{}{
				"foo": []interface{}{"boom", "bar", "baz"},
				"hi":  map[string]interface{}{"there": "goodbye"},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("result = %#v, want %#v", got, want)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/json-patch.test.ts#L56
	t.Run("apply does not mutate the patches", func(t *testing.T) {
		patches := []cdk8s.JsonPatch{
			cdk8s.JsonPatch_Add(coreString("/world/foo"), []interface{}{}),
			cdk8s.JsonPatch_Add(coreString("/world/foo/-"), "boom"),
		}
		want := map[string]interface{}{"world": map[string]interface{}{"foo": []interface{}{"boom"}}}
		for attempt := 0; attempt < 2; attempt++ {
			got := cdk8s.JsonPatch_Apply(map[string]interface{}{"world": map[string]interface{}{}}, patches...)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("application %d result = %#v, want %#v", attempt+1, got, want)
			}
		}
	})
}
