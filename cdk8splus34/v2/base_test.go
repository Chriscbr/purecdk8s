package cdk8splus34_test

import (
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	constructs "github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/base.test.ts#L5
func TestResourceCanMutateMetadata(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	custom := constructs.NewConstruct(chart, jsii.String("Custom"))
	resource := plus.NewConfigMap(custom, jsii.String("ConfigMap"), nil)
	resource.Metadata().AddLabel(jsii.String("key"), jsii.String("value"))

	manifest := manifestOfKind(t, chart, "ConfigMap")
	if manifest["apiVersion"] != "v1" || manifest["kind"] != "ConfigMap" {
		t.Fatalf("resource identity = %#v/%#v, want v1/ConfigMap", manifest["apiVersion"], manifest["kind"])
	}
	metadata := mapAt(t, manifest, "metadata")
	name, ok := metadata["name"].(string)
	if !ok || !strings.HasPrefix(name, "test-custom-configmap-") {
		t.Fatalf("metadata.name = %#v, want generated test-custom-configmap-* name", metadata["name"])
	}
	requireDeepEqual(t, mapAt(t, metadata, "labels"), map[string]interface{}{"key": "value"})
}
