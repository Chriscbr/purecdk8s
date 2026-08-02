package cdk8splus34_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// requireSnapshotHash compares the complete public synthesis result with a
// canonical hash of the corresponding upstream Jest snapshot. JSON object key
// order is intentionally ignored, just as it is by Jest's structural matcher.
func requireSnapshotHash(t *testing.T, got interface{}, want string) {
	t.Helper()
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal snapshot value: %v", err)
	}
	actual := fmt.Sprintf("%x", sha256.Sum256(encoded))
	if actual != want {
		pretty, _ := json.MarshalIndent(got, "", "  ")
		t.Fatalf("snapshot hash = %s, want %s\nactual snapshot:\n%s", actual, want, pretty)
	}
}

func plainStringMap(values *map[string]*string) map[string]string {
	result := map[string]string{}
	if values == nil {
		return result
	}
	for key, value := range *values {
		result[key] = stringValue(value)
	}
	return result
}

func normalizeValue(t *testing.T, value interface{}) interface{} {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	var normalized interface{}
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		t.Fatalf("normalize value: %v", err)
	}
	return normalized
}

func attachedContainerManifest(t *testing.T, chart cdk8s.Chart, container plus.Container) map[string]interface{} {
	t.Helper()
	pod := plus.NewPod(chart, jsii.String("Pod"), nil)
	pod.AttachContainer(container)
	containers := sliceAt(t, manifestOfKind(t, chart, "Pod"), "spec", "containers")
	return mapAt(t, containers[0])
}
