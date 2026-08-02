package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func TestNamespace(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/namespace.test.ts#L5
	t.Run("defaultChild", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		namespace := plus.NewNamespace(chart, jsii.String("Namespace"), nil)
		if got := stringValue(cdk8s.ApiObject_Of(namespace).Kind()); got != "Namespace" {
			t.Fatalf("default child kind = %q, want Namespace", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/namespace.test.ts#L16
	t.Run("defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewNamespace(chart, jsii.String("Namespace"), nil)
		requireDeepEqual(t, synth(t, chart), []interface{}{map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata":   map[string]interface{}{"name": "test-namespace-c83f04e1"},
			"spec":       map[string]interface{}{},
		}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/namespace.test.ts#L25
	t.Run("can select namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		namespaces := plus.Namespaces_Select(chart, jsii.String("Namespaces"), &plus.NamespacesSelectOptions{
			Labels:      &map[string]*string{"foo": jsii.String("bar")},
			Expressions: &[]plus.LabelExpression{plus.LabelExpression_Exists(jsii.String("web"))},
			Names:       &[]*string{jsii.String("web")},
		})
		config := namespaces.ToNamespaceSelectorConfig()
		if config.Names == nil || len(*config.Names) != 1 || stringValue((*config.Names)[0]) != "web" {
			t.Fatalf("namespace names = %#v, want [web]", config.Names)
		}
		policy := plus.NewNetworkPolicy(chart, jsii.String("Policy"), nil)
		policy.AddIngressRule(namespaces, nil)
		from := sliceAt(t, manifestOfKind(t, chart, "NetworkPolicy"), "spec", "ingress")
		peer := from[0].(map[string]interface{})["from"].([]interface{})[0].(map[string]interface{})
		selector := peer["namespaceSelector"].(map[string]interface{})
		labels := selector["matchLabels"].(map[string]interface{})
		if labels["foo"] != "bar" || labels["kubernetes.io/metadata.name"] != "web" {
			t.Fatalf("namespace selector labels = %#v", labels)
		}
		expressions := selector["matchExpressions"].([]interface{})
		requireDeepEqual(t, expressions, []interface{}{map[string]interface{}{"key": "web", "operator": "Exists"}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/namespace.test.ts#L36
	t.Run("can select all namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		namespaces := plus.Namespaces_All(chart, jsii.String("All"))
		config := namespaces.ToNamespaceSelectorConfig()
		if config.Names != nil {
			t.Fatalf("namespace names = %#v, want nil", config.Names)
		}
		if config.LabelSelector == nil || !boolValue(config.LabelSelector.IsEmpty()) {
			t.Fatal("all-namespaces label selector is not empty")
		}
		policy := plus.NewNetworkPolicy(chart, jsii.String("Policy"), nil)
		policy.AddIngressRule(namespaces, nil)
		ingress := sliceAt(t, manifestOfKind(t, chart, "NetworkPolicy"), "spec", "ingress")
		peer := ingress[0].(map[string]interface{})["from"].([]interface{})[0].(map[string]interface{})
		requireDeepEqual(t, peer["namespaceSelector"], map[string]interface{}{})
	})
}
