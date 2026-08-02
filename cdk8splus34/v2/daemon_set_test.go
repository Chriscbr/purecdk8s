package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/daemon-set.test.ts#L5
func TestDaemonSetDefaultChild(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	daemonSet := plus.NewDaemonSet(chart, jsii.String("DaemonSet"), nil)
	child := daemonSet.Node().DefaultChild()
	if child == nil || stringValue(cdk8s.ApiObject_Of(child).Kind()) != "DaemonSet" {
		t.Fatal("DaemonSet default child is not a DaemonSet ApiObject")
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/daemon-set.test.ts#L15
func TestDaemonSetDefaults(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewDaemonSet(chart, jsii.String("DaemonSet"), &plus.DaemonSetProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}}})
	requireSnapshotHash(t, synth(t, chart), "d7ffc0b25cfc8a98401da70b0d8521e6d9c3c073b59233afd9653867d03d9e7e")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/daemon-set.test.ts#L26
func TestDaemonSetCustom(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewDaemonSet(chart, jsii.String("DaemonSet"), &plus.DaemonSetProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}}, MinReadySeconds: jsii.Number(5),
	})
	requireSnapshotHash(t, synth(t, chart), "97b775eeadeaa878c332ee159de2fe1360d7743e4b4b39124a6518b5794a212f")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/daemon-set.test.ts#L38
func TestDaemonSetAutomaticallyAllocatesLabelSelector(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	daemonSet := plus.NewDaemonSet(chart, jsii.String("DaemonSet"), nil)
	daemonSet.AddContainer(&plus.ContainerProps{Image: jsii.String("foobar")})
	want := map[string]string{"cdk8s.io/metadata.addr": "test-DaemonSet-c8f77186"}
	spec := mapAt(t, manifestAt(t, chart, 0), "spec")
	requireDeepEqual(t, mapAt(t, spec, "selector", "matchLabels"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-DaemonSet-c8f77186"})
	requireDeepEqual(t, mapAt(t, spec, "template", "metadata", "labels"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-DaemonSet-c8f77186"})
	requireDeepEqual(t, plainStringMap(daemonSet.MatchLabels()), want)
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/daemon-set.test.ts#L58
func TestDaemonSetSelectFalseGeneratesNoSelector(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	daemonSet := plus.NewDaemonSet(chart, jsii.String("DaemonSet"), &plus.DaemonSetProps{
		Select: jsii.Bool(false), Containers: &[]*plus.ContainerProps{{Image: jsii.String("foobar")}},
	})
	matchLabels := mapAt(t, manifestAt(t, chart, 0), "spec", "selector", "matchLabels")
	if len(matchLabels) != 0 {
		t.Fatalf("selector.matchLabels = %#v, want empty", matchLabels)
	}
	requireDeepEqual(t, plainStringMap(daemonSet.MatchLabels()), map[string]string{})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/daemon-set.test.ts#L76
func TestDaemonSetCanSelectByLabel(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	daemonSet := plus.NewDaemonSet(chart, jsii.String("DaemonSet"), &plus.DaemonSetProps{
		Select: jsii.Bool(false), Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}},
	})
	daemonSet.Select(plus.LabelSelector_Of(&plus.LabelSelectorOptions{Labels: &map[string]*string{"foo": jsii.String("bar")}}))
	want := map[string]interface{}{"foo": "bar"}
	requireDeepEqual(t, mapAt(t, manifestAt(t, chart, 0), "spec", "selector", "matchLabels"), want)
	requireDeepEqual(t, plainStringMap(daemonSet.MatchLabels()), map[string]string{"foo": "bar"})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/daemon-set.test.ts#L98
func TestDaemonSetCanBeIsolated(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewDaemonSet(chart, jsii.String("DaemonSet"), &plus.DaemonSetProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("foobar")}}, Isolate: jsii.Bool(true),
	})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "34a58ed14ab51f40eb167173dc6058d939a4c6d2d8230dc2b956e104d574b44f")
	policy := mapAt(t, manifests[1], "spec")
	if labels := mapAt(t, policy, "podSelector", "matchLabels"); len(labels) == 0 {
		t.Fatal("isolating DaemonSet produced no pod selector labels")
	}
	requireDeepEqual(t, policy["policyTypes"], []interface{}{"Egress", "Ingress"})
}
