package cdk8s_test

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/Chriscbr/purecdk8s/constructs/v10"
)

func coreString(value string) *string {
	return &value
}

func coreFloat(value float64) *float64 {
	return &value
}

func coreBool(value bool) *bool {
	return &value
}

func coreStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func coreAssertEqual(t *testing.T, got, want interface{}) {
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

func coreAssertJSONEqual(t *testing.T, got, want interface{}) {
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

func coreRequirePanicContains(t *testing.T, want string, callback func()) {
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

func coreRequireClose(t *testing.T, got, want float64) {
	t.Helper()
	const tolerance = 1e-9
	if math.Abs(got-want) > tolerance*math.Max(1, math.Abs(want)) {
		t.Fatalf("value = %.16g, want %.16g", got, want)
	}
}

func coreCreateTree(path string) constructs.Construct {
	var current constructs.Construct = constructs.NewRootConstruct(nil)
	for _, component := range strings.Split(path, "/") {
		current = constructs.NewConstruct(current, coreString(component))
	}
	return current
}
