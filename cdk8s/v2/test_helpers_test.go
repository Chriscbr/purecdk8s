package cdk8s_test

import (
	"fmt"
	"math"
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
