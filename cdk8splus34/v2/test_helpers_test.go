package cdk8splus34_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func numberValue(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func requirePanicContains(t *testing.T, want string, callback func()) {
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

func requireDeepEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(want, "", "  ")
	t.Fatalf("value mismatch\ngot:  %s\nwant: %s", gotJSON, wantJSON)
}

// synth converts a chart's public synthesis result into ordinary JSON values.
// This keeps every assertion on the same observable boundary as a consumer of
// the package and makes pointer-backed JSII values straightforward to compare.
func synth(t *testing.T, chart cdk8s.Chart) []interface{} {
	t.Helper()
	raw := cdk8s.Testing_Synth(chart)
	encoded, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal synthesized manifests: %v", err)
	}
	var manifests []interface{}
	if err := json.Unmarshal(encoded, &manifests); err != nil {
		t.Fatalf("normalize synthesized manifests: %v", err)
	}
	return manifests
}

func manifestAt(t *testing.T, chart cdk8s.Chart, index int) map[string]interface{} {
	t.Helper()
	manifests := synth(t, chart)
	if index < 0 || index >= len(manifests) {
		t.Fatalf("manifest index %d out of range for %d manifests", index, len(manifests))
	}
	manifest, ok := manifests[index].(map[string]interface{})
	if !ok {
		t.Fatalf("manifest %d has type %T, want map[string]interface{}", index, manifests[index])
	}
	return manifest
}

func manifestOfKind(t *testing.T, chart cdk8s.Chart, kind string) map[string]interface{} {
	t.Helper()
	for _, candidate := range synth(t, chart) {
		manifest, ok := candidate.(map[string]interface{})
		if ok && manifest["kind"] == kind {
			return manifest
		}
	}
	t.Fatalf("no synthesized manifest has kind %q", kind)
	return nil
}

func mapAt(t *testing.T, value interface{}, keys ...string) map[string]interface{} {
	t.Helper()
	current := value
	for _, key := range keys {
		mapping, ok := current.(map[string]interface{})
		if !ok {
			t.Fatalf("value at %q has type %T, want map[string]interface{}", strings.Join(keys, "."), current)
		}
		current, ok = mapping[key]
		if !ok {
			t.Fatalf("missing key %q in path %q", key, strings.Join(keys, "."))
		}
	}
	result, ok := current.(map[string]interface{})
	if !ok {
		t.Fatalf("value at %q has type %T, want map[string]interface{}", strings.Join(keys, "."), current)
	}
	return result
}

func sliceAt(t *testing.T, value interface{}, keys ...string) []interface{} {
	t.Helper()
	current := value
	for _, key := range keys {
		mapping, ok := current.(map[string]interface{})
		if !ok {
			t.Fatalf("value at %q has type %T, want map[string]interface{}", strings.Join(keys, "."), current)
		}
		current, ok = mapping[key]
		if !ok {
			t.Fatalf("missing key %q in path %q", key, strings.Join(keys, "."))
		}
	}
	result, ok := current.([]interface{})
	if !ok {
		t.Fatalf("value at %q has type %T, want []interface{}", strings.Join(keys, "."), current)
	}
	return result
}
