package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func newTestDeployment(chart cdk8s.Chart, id, image string) plus.Deployment {
	return plus.NewDeployment(chart, jsii.String(id), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String(image)}},
	})
}

func unmanagedTestPods(chart cdk8s.Chart) plus.Pods {
	namespaces := plus.Namespaces_Select(chart, jsii.String("Net"), &plus.NamespacesSelectOptions{
		Labels: &map[string]*string{"net": jsii.String("1")},
	})
	return plus.Pods_Select(chart, jsii.String("Redis"), &plus.PodsSelectOptions{
		Labels: &map[string]*string{"app": jsii.String("store")}, Namespaces: namespaces,
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L6
func TestDeploymentDefaultChild(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), nil)
	child := deployment.Node().DefaultChild()
	if child == nil || stringValue(cdk8s.ApiObject_Of(child).Kind()) != "Deployment" {
		t.Fatal("Deployment default child is not a Deployment ApiObject")
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L16
func TestDeploymentAutomaticallyAllocatesLabelSelector(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), nil)
	deployment.AddContainer(&plus.ContainerProps{Image: jsii.String("foobar")})
	wantManifest := map[string]interface{}{"cdk8s.io/metadata.addr": "test-Deployment-c83f5e59"}
	wantObject := map[string]string{"cdk8s.io/metadata.addr": "test-Deployment-c83f5e59"}
	spec := mapAt(t, manifestAt(t, chart, 0), "spec")
	requireDeepEqual(t, mapAt(t, spec, "selector", "matchLabels"), wantManifest)
	requireDeepEqual(t, mapAt(t, spec, "template", "metadata", "labels"), wantManifest)
	requireDeepEqual(t, plainStringMap(deployment.MatchLabels()), wantObject)
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L37
func TestDeploymentCanBeIsolated(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("foobar")}}, Replicas: jsii.Number(5), Isolate: jsii.Bool(true),
	})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "3013a9e0bc92416c3354beec0b30e43b5639f66023d2804bd2b33c93c99de172")
	policy := mapAt(t, manifests[1], "spec")
	if labels := mapAt(t, policy, "podSelector", "matchLabels"); len(labels) == 0 {
		t.Fatal("isolating Deployment produced no pod selector labels")
	}
	requireDeepEqual(t, policy["policyTypes"], []interface{}{"Egress", "Ingress"})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L55
func TestDeploymentSelectFalseGeneratesNoSelector(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Select: jsii.Bool(false), Containers: &[]*plus.ContainerProps{{Image: jsii.String("foobar")}},
	})
	selector := mapAt(t, manifestAt(t, chart, 0), "spec", "selector")
	if value, exists := selector["matchLabels"]; exists {
		t.Fatalf("selector.matchLabels = %#v, want absent", value)
	}
	requireDeepEqual(t, plainStringMap(deployment.MatchLabels()), map[string]string{})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L73
func TestDeploymentCanSelectByLabel(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Select: jsii.Bool(false), Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}},
	})
	deployment.Select(plus.LabelSelector_Of(&plus.LabelSelectorOptions{Labels: &map[string]*string{"foo": jsii.String("bar")}}))
	requireDeepEqual(t, mapAt(t, manifestAt(t, chart, 0), "spec", "selector", "matchLabels"), map[string]interface{}{"foo": "bar"})
	requireDeepEqual(t, plainStringMap(deployment.MatchLabels()), map[string]string{"foo": "bar"})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L99
func TestDeploymentCanBeExposedViaIngress(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), Port: jsii.Number(9300)}},
	})
	deployment.ExposeViaIngress(jsii.String("/hello"), nil)
	requireSnapshotHash(t, synth(t, chart), "2cc4a0829a803be16f7d02648143a609475ad497eeae8304456f382d4f9d08f2")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L117
func TestDeploymentExposeUsesCorrectDefaults(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), Port: jsii.Number(9300)}},
	})
	deployment.ExposeViaService(nil)
	serviceSpec := mapAt(t, manifestAt(t, chart, 1), "spec")
	if got := serviceSpec["type"]; got != "ClusterIP" {
		t.Errorf("service type = %#v, want ClusterIP", got)
	}
	ports := sliceAt(t, serviceSpec, "ports")
	port := mapAt(t, ports[0])
	if got := port["targetPort"]; got != float64(9300) {
		t.Errorf("targetPort = %#v, want 9300", got)
	}
	if got := port["port"]; got != float64(9300) {
		t.Errorf("port = %#v, want 9300", got)
	}
	if got := mapAt(t, manifestAt(t, chart, 0), "spec")["replicas"]; got != float64(2) {
		t.Errorf("replicas = %#v, want 2", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L141
func TestDeploymentExposeCanSetServiceAndPortDetails(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), Port: jsii.Number(9500)}},
	})
	deployment.ExposeViaService(&plus.DeploymentExposeViaServiceOptions{
		Ports: &[]*plus.ServicePort{{Port: jsii.Number(9200), Protocol: plus.Protocol_UDP, TargetPort: jsii.Number(9500)}},
		Name:  jsii.String("test-srv"), ServiceType: plus.ServiceType_CLUSTER_IP,
	})
	service := manifestAt(t, chart, 1)
	if got := mapAt(t, service, "metadata")["name"]; got != "test-srv" {
		t.Errorf("service name = %#v, want test-srv", got)
	}
	spec := mapAt(t, service, "spec")
	if got := spec["type"]; got != "ClusterIP" {
		t.Errorf("service type = %#v, want ClusterIP", got)
	}
	requireDeepEqual(t, mapAt(t, spec, "selector"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-Deployment-c83f5e59"})
	port := mapAt(t, sliceAt(t, spec, "ports")[0])
	requireDeepEqual(t, port, map[string]interface{}{"port": float64(9200), "protocol": "UDP", "targetPort": float64(9500)})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L173
func TestDeploymentCannotBeExposedWithoutContainers(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), nil)
	requirePanicContains(t, "Unable to expose deployment", func() { deployment.ExposeViaService(nil) })
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L182
func TestDeploymentSynthesizesSpecLazily(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), nil)
	deployment.AddContainer(&plus.ContainerProps{Image: jsii.String("image"), Port: jsii.Number(9300)})
	container := mapAt(t, sliceAt(t, manifestAt(t, chart, 0), "spec", "template", "spec", "containers")[0])
	if got := container["image"]; got != "image" {
		t.Errorf("image = %#v, want image", got)
	}
	if got := mapAt(t, sliceAt(t, container, "ports")[0])["containerPort"]; got != float64(9300) {
		t.Errorf("containerPort = %#v, want 9300", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L201
func TestDeploymentDefaultStrategy(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), nil)
	deployment.AddContainer(&plus.ContainerProps{Image: jsii.String("image")})
	if deployment.Strategy() == nil {
		t.Fatal("default deployment strategy is nil")
	}
	requireDeepEqual(t, mapAt(t, manifestAt(t, chart, 0), "spec")["strategy"], map[string]interface{}{
		"type": "RollingUpdate", "rollingUpdate": map[string]interface{}{"maxSurge": "25%", "maxUnavailable": "25%"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L221
func TestDeploymentCustomStrategy(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	strategy := plus.DeploymentStrategy_Recreate()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{Strategy: strategy})
	deployment.AddContainer(&plus.ContainerProps{Image: jsii.String("image")})
	if deployment.Strategy() != strategy {
		t.Fatal("Deployment did not retain the configured strategy")
	}
	requireDeepEqual(t, mapAt(t, manifestAt(t, chart, 0), "spec")["strategy"], map[string]interface{}{"type": "Recreate"})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L237
func TestDeploymentRollingUpdateStrategyCanBeCustomized(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Strategy: plus.DeploymentStrategy_RollingUpdate(&plus.DeploymentStrategyRollingUpdateOptions{
			MaxSurge: plus.PercentOrAbsolute_Percent(jsii.Number(50)), MaxUnavailable: plus.PercentOrAbsolute_Absolute(jsii.Number(1)),
		}),
	})
	deployment.AddContainer(&plus.ContainerProps{Image: jsii.String("image")})
	requireDeepEqual(t, mapAt(t, manifestAt(t, chart, 0), "spec")["strategy"], map[string]interface{}{
		"type": "RollingUpdate", "rollingUpdate": map[string]interface{}{"maxSurge": "50%", "maxUnavailable": float64(1)},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L260
func TestDeploymentRollingUpdateRejectsBothValuesZero(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	requirePanicContains(t, "'maxSurge' and 'maxUnavailable' cannot be both zero", func() {
		plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
			Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}},
			Strategy: plus.DeploymentStrategy_RollingUpdate(&plus.DeploymentStrategyRollingUpdateOptions{
				MaxSurge: plus.PercentOrAbsolute_Absolute(jsii.Number(0)), MaxUnavailable: plus.PercentOrAbsolute_Percent(jsii.Number(0)),
			}),
		})
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L270
func TestPercentOrAbsoluteZero(t *testing.T) {
	for name, test := range map[string]struct {
		value plus.PercentOrAbsolute
		want  bool
	}{
		"zero percent":  {plus.PercentOrAbsolute_Percent(jsii.Number(0)), true},
		"zero absolute": {plus.PercentOrAbsolute_Absolute(jsii.Number(0)), true},
		"percent":       {plus.PercentOrAbsolute_Percent(jsii.Number(1)), false},
		"absolute":      {plus.PercentOrAbsolute_Absolute(jsii.Number(1)), false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := boolValue(test.value.IsZero()); got != test.want {
				t.Errorf("IsZero() = %v, want %v", got, test.want)
			}
		})
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L279
func TestDeploymentDefaultMinReady(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := newTestDeployment(chart, "Deployment", "image")
	if got := numberValue(deployment.MinReady().ToSeconds(nil)); got != 0 {
		t.Errorf("minReady = %v, want 0s", got)
	}
	if got := mapAt(t, manifestAt(t, chart, 0), "spec")["minReadySeconds"]; got != float64(0) {
		t.Errorf("minReadySeconds = %#v, want 0", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L293
func TestDeploymentDefaultProgressDeadline(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := newTestDeployment(chart, "Deployment", "image")
	if got := numberValue(deployment.ProgressDeadline().ToSeconds(nil)); got != 600 {
		t.Errorf("progressDeadline = %v, want 600s", got)
	}
	if got := mapAt(t, manifestAt(t, chart, 0), "spec")["progressDeadlineSeconds"]; got != float64(600) {
		t.Errorf("progressDeadlineSeconds = %#v, want 600", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L307
func TestDeploymentDefaultRevisionHistoryLimit(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := newTestDeployment(chart, "Deployment", "image")
	if got := numberValue(deployment.RevisionHistoryLimit()); got != 10 {
		t.Errorf("revisionHistoryLimit = %v, want 10", got)
	}
	if got := mapAt(t, manifestAt(t, chart, 0), "spec")["revisionHistoryLimit"]; got != float64(10) {
		t.Errorf("revisionHistoryLimit = %#v, want 10", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L321
func TestDeploymentCanConfigureMinReady(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}}, MinReady: cdk8s.Duration_Seconds(jsii.Number(60)),
	})
	if got := numberValue(deployment.MinReady().ToSeconds(nil)); got != 60 {
		t.Errorf("minReady = %v, want 60s", got)
	}
	if got := mapAt(t, manifestAt(t, chart, 0), "spec")["minReadySeconds"]; got != float64(60) {
		t.Errorf("minReadySeconds = %#v, want 60", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L336
func TestDeploymentCanConfigureProgressDeadline(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}}, ProgressDeadline: cdk8s.Duration_Seconds(jsii.Number(60)),
	})
	if got := numberValue(deployment.ProgressDeadline().ToSeconds(nil)); got != 60 {
		t.Errorf("progressDeadline = %v, want 60s", got)
	}
	if got := mapAt(t, manifestAt(t, chart, 0), "spec")["progressDeadlineSeconds"]; got != float64(60) {
		t.Errorf("progressDeadlineSeconds = %#v, want 60", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L351
func TestDeploymentRejectsMinReadyGreaterThanProgressDeadline(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	requirePanicContains(t, "'progressDeadline' (30s) must be greater than 'minReady' (60s)", func() {
		plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
			Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}},
			MinReady:   cdk8s.Duration_Seconds(jsii.Number(60)), ProgressDeadline: cdk8s.Duration_Seconds(jsii.Number(30)),
		})
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L363
func TestDeploymentRejectsMinReadyEqualToProgressDeadline(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	requirePanicContains(t, "'progressDeadline' (60s) must be greater than 'minReady' (60s)", func() {
		plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
			Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}},
			MinReady:   cdk8s.Duration_Seconds(jsii.Number(60)), ProgressDeadline: cdk8s.Duration_Seconds(jsii.Number(60)),
		})
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L375
func TestDeploymentCanSelectWithExpressions(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}}, Select: jsii.Bool(false),
	})
	values := &[]*string{jsii.String("v1"), jsii.String("v2")}
	deployment.Select(plus.LabelSelector_Of(&plus.LabelSelectorOptions{Expressions: &[]plus.LabelExpression{
		plus.LabelExpression_In(jsii.String("foo"), values),
		plus.LabelExpression_NotIn(jsii.String("foo"), values),
		plus.LabelExpression_Exists(jsii.String("foo")),
		plus.LabelExpression_DoesNotExist(jsii.String("foo")),
	}}))
	requireDeepEqual(t, sliceAt(t, manifestAt(t, chart, 0), "spec", "selector", "matchExpressions"), []interface{}{
		map[string]interface{}{"key": "foo", "operator": "In", "values": []interface{}{"v1", "v2"}},
		map[string]interface{}{"key": "foo", "operator": "NotIn", "values": []interface{}{"v1", "v2"}},
		map[string]interface{}{"key": "foo", "operator": "Exists"},
		map[string]interface{}{"key": "foo", "operator": "DoesNotExist"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L407
func TestDeploymentSchedulingCanTolerateTaintedNodes(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	devNodes := plus.Node_Tainted(
		plus.NodeTaintQuery_Is(jsii.String("key1"), jsii.String("value1"), nil),
		plus.NodeTaintQuery_Is(jsii.String("key2"), jsii.String("value2"), &plus.NodeTaintQueryOptions{Effect: plus.TaintEffect_PREFER_NO_SCHEDULE}),
		plus.NodeTaintQuery_Exists(jsii.String("key3"), nil),
		plus.NodeTaintQuery_Exists(jsii.String("key4"), &plus.NodeTaintQueryOptions{Effect: plus.TaintEffect_NO_SCHEDULE}),
		plus.NodeTaintQuery_Is(jsii.String("key5"), jsii.String("value5"), &plus.NodeTaintQueryOptions{Effect: plus.TaintEffect_NO_EXECUTE, EvictAfter: cdk8s.Duration_Hours(jsii.Number(1))}),
		plus.NodeTaintQuery_Any(),
	)
	redis := plus.NewPod(chart, jsii.String("Redis"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("redis")}}})
	redis.Scheduling().Tolerate(devNodes)
	requireSnapshotHash(t, synth(t, chart), "4a1bee84468d9108ce55babff3f5cbdb6b36d3d08d2a22fca03170fb4dd8961e")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L429
func TestDeploymentSchedulingCanAssignNodeByName(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := newTestDeployment(chart, "Redis", "redis")
	redis.Scheduling().Assign(plus.Node_Named(jsii.String("node1")))
	requireSnapshotHash(t, synth(t, chart), "f3928b48665ca10ceae3ba5a0f86de02d21a980889dd97f9f6144c75731a32df")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L440
func TestDeploymentSchedulingCanAttractNodeBySelectorDefault(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := newTestDeployment(chart, "Redis", "redis")
	redis.Scheduling().Attract(plus.Node_Labeled(plus.NodeLabelQuery_Is(jsii.String("memory"), jsii.String("high"))), nil)
	requireSnapshotHash(t, synth(t, chart), "c0df5192f389dbdd081a6b169b764cf9e939d76328eb0547359e1cd109187b05")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L451
func TestDeploymentSchedulingCanAttractNodeBySelectorCustom(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := newTestDeployment(chart, "Redis", "redis")
	redis.Scheduling().Attract(plus.Node_Labeled(plus.NodeLabelQuery_Is(jsii.String("memory"), jsii.String("high"))), &plus.PodSchedulingAttractOptions{Weight: jsii.Number(1)})
	requireSnapshotHash(t, synth(t, chart), "3ada2f23b83f2d3956ba9b53c6f5376f0f951e3c0bd922c2929c00f09b859ddf")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L466
func TestDeploymentSchedulingCanColocateManagedDefault(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := newTestDeployment(chart, "Redis", "redis")
	web := newTestDeployment(chart, "Web", "web")
	web.Scheduling().Colocate(redis, nil)
	requireSnapshotHash(t, synth(t, chart), "646ef24e9c30cb286a550ca66b96e2319b0edc752265b241add326efb2213a43")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L483
func TestDeploymentSchedulingCanColocateManagedCustom(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := newTestDeployment(chart, "Redis", "redis")
	web := newTestDeployment(chart, "Web", "web")
	web.Scheduling().Colocate(redis, &plus.PodSchedulingColocateOptions{Topology: plus.Topology_ZONE(), Weight: jsii.Number(1)})
	requireSnapshotHash(t, synth(t, chart), "670f059db3015f1dc01fefbf520cd7dea53225e9b5cdc96e5b03e1e93edf24ac")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L503
func TestDeploymentSchedulingCanColocateUnmanagedDefault(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := unmanagedTestPods(chart)
	web := newTestDeployment(chart, "Web", "web")
	web.Scheduling().Colocate(redis, nil)
	requireSnapshotHash(t, synth(t, chart), "3d80179a8bd3178a3b1fa18779e91db89f7deaa10098bf70e50c0cbc88c64891")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L522
func TestDeploymentSchedulingCanColocateUnmanagedCustom(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := unmanagedTestPods(chart)
	web := newTestDeployment(chart, "Web", "web")
	web.Scheduling().Colocate(redis, &plus.PodSchedulingColocateOptions{Topology: plus.Topology_ZONE(), Weight: jsii.Number(1)})
	requireSnapshotHash(t, synth(t, chart), "1626427ae928ce90d7a68cafe3871d59221f9f9f80ddc562fe440877c33b07de")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L544
func TestDeploymentSchedulingCanSpreadDefault(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := newTestDeployment(chart, "Deployment", "redis")
	deployment.Scheduling().Spread(nil)
	requireSnapshotHash(t, synth(t, chart), "d4eb4c7b9ecff9c8c4f3250889f3700683e9ca23039706389c4965a32c7b9b5e")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L558
func TestDeploymentSchedulingCanSpreadCustom(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := newTestDeployment(chart, "Deployment", "redis")
	deployment.Scheduling().Spread(&plus.WorkloadSchedulingSpreadOptions{Weight: jsii.Number(1)})
	requireSnapshotHash(t, synth(t, chart), "5a55e695bc4ee6c95790c427a90e933835e994423971001d7e1a8fa613e4df8d")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L572
func TestDeploymentSchedulingSpreadProp(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("redis")}}, Spread: jsii.Bool(true),
	})
	requireSnapshotHash(t, synth(t, chart), "31b125ddcf0cdb92f680666cba34dc6d024ef0f8f3b314ec7f76b9242b754b28")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L584
func TestDeploymentSchedulingCanSeparateManagedDefault(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := newTestDeployment(chart, "Redis", "redis")
	web := newTestDeployment(chart, "Web", "web")
	web.Scheduling().Separate(redis, nil)
	requireSnapshotHash(t, synth(t, chart), "52e8d46c45fde2d1b93a7105e685232c4720f329ec34a860b9e0b834811fa314")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L601
func TestDeploymentSchedulingCanSeparateManagedCustom(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := newTestDeployment(chart, "Redis", "redis")
	web := newTestDeployment(chart, "Web", "web")
	web.Scheduling().Separate(redis, &plus.PodSchedulingSeparateOptions{Topology: plus.Topology_ZONE(), Weight: jsii.Number(1)})
	requireSnapshotHash(t, synth(t, chart), "53b2e99605807ad2437ff805cc4c2deb13f3d8b6f8fbad92ac2a716c22d96316")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L621
func TestDeploymentSchedulingCanSeparateUnmanagedDefault(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := unmanagedTestPods(chart)
	web := newTestDeployment(chart, "Web", "web")
	web.Scheduling().Separate(redis, nil)
	requireSnapshotHash(t, synth(t, chart), "b8f867ed29378e7d43b8326aa90c3644f711a00f14cbc372eb37e4e10df972d4")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L640
func TestDeploymentSchedulingCanSeparateUnmanagedCustom(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	redis := unmanagedTestPods(chart)
	web := newTestDeployment(chart, "Web", "web")
	web.Scheduling().Separate(redis, &plus.PodSchedulingSeparateOptions{Topology: plus.Topology_ZONE(), Weight: jsii.Number(1)})
	requireSnapshotHash(t, synth(t, chart), "681273dd93421c8b6f67302ec9da5856943e0ccc92973142e468c52c61c7b928")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L664
func TestDeploymentExposePreservesNamespace(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), Port: jsii.Number(9300)}},
		Metadata:   &cdk8s.ApiObjectMetadata{Namespace: jsii.String("custom")},
	})
	deployment.ExposeViaService(nil)
	requireSnapshotHash(t, synth(t, chart), "cfbef92533522ae7ac858cfb70e7c346960b0252c9428c965009d7bfefe9f8cf")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L683
func TestDeploymentExposeCapturesAllContainerPorts(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), Ports: &[]*plus.ContainerPort{
			{Number: jsii.Number(8080), Name: jsii.String("port1")},
			{Number: jsii.Number(9090), Name: jsii.String("port2")},
		}}},
	})
	deployment.ExposeViaService(nil)
	requireSnapshotHash(t, synth(t, chart), "ad1785d7941ab4b034386dbc6d297875820a363c9068c2fc553de43d8d4be75e")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L708
func TestDeploymentCannotExposePortNotOwnedByContainer(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), Ports: &[]*plus.ContainerPort{
			{Number: jsii.Number(8080)}, {Number: jsii.Number(9090)},
		}}},
	})
	requirePanicContains(t, "Unable to expose deployment", func() {
		deployment.ExposeViaService(&plus.DeploymentExposeViaServiceOptions{Ports: &[]*plus.ServicePort{{Port: jsii.Number(2020)}}})
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/deployment.test.ts#L734
func TestDeploymentExposeMultiplePortsRequiresNames(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), Ports: &[]*plus.ContainerPort{
			{Number: jsii.Number(8080), Name: jsii.String("port1")}, {Number: jsii.Number(9090)},
		}}},
	})
	requirePanicContains(t, "Unable to expose deployment", func() { deployment.ExposeViaService(nil) })

	deployment2 := plus.NewDeployment(chart, jsii.String("Deployment2"), &plus.DeploymentProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), Ports: &[]*plus.ContainerPort{
			{Number: jsii.Number(8080)}, {Number: jsii.Number(9090)},
		}}},
	})
	requirePanicContains(t, "Unable to expose deployment", func() { deployment2.ExposeViaService(nil) })
}
