package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func newTestService(chart cdk8s.Chart, id string, port float64) plus.Service {
	return plus.NewService(chart, jsii.String(id), &plus.ServiceProps{Ports: &[]*plus.ServicePort{{Port: jsii.Number(port)}}})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L6
func TestStatefulSetDefaultChild(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	set := plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{Service: newTestService(chart, "TestService", 80)})
	child := set.Node().DefaultChild()
	if child == nil || stringValue(cdk8s.ApiObject_Of(child).Kind()) != "StatefulSet" {
		t.Fatal("StatefulSet default child is not a StatefulSet ApiObject")
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L17
func TestStatefulSetAutomaticallyAllocatesLabelSelector(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	set := plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{Service: newTestService(chart, "TestService", 80)})
	set.AddContainer(&plus.ContainerProps{Image: jsii.String("foobar")})
	wantManifest := map[string]interface{}{"cdk8s.io/metadata.addr": "test-StatefulSet-c809b559"}
	wantObject := map[string]string{"cdk8s.io/metadata.addr": "test-StatefulSet-c809b559"}
	spec := mapAt(t, manifestAt(t, chart, 1), "spec")
	requireDeepEqual(t, mapAt(t, spec, "selector", "matchLabels"), wantManifest)
	requireDeepEqual(t, mapAt(t, spec, "template", "metadata", "labels"), wantManifest)
	requireDeepEqual(t, plainStringMap(set.MatchLabels()), wantObject)
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L39
func TestStatefulSetSelectFalseGeneratesNoSelector(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	set := plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{
		Select: jsii.Bool(false), Containers: &[]*plus.ContainerProps{{Image: jsii.String("foobar")}},
		Service: newTestService(chart, "TestService", 80),
	})
	if got := mapAt(t, manifestAt(t, chart, 1), "spec", "selector", "matchLabels"); len(got) != 0 {
		t.Fatalf("selector.matchLabels = %#v, want empty", got)
	}
	requireDeepEqual(t, plainStringMap(set.MatchLabels()), map[string]string{})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L59
func TestStatefulSetCanSelectByLabel(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	set := plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{
		Select: jsii.Bool(false), Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}},
		Service: newTestService(chart, "TestService", 80),
	})
	set.Select(plus.LabelSelector_Of(&plus.LabelSelectorOptions{Labels: &map[string]*string{"foo": jsii.String("bar")}}))
	requireDeepEqual(t, mapAt(t, manifestAt(t, chart, 1), "spec", "selector", "matchLabels"), map[string]interface{}{"foo": "bar"})
	requireDeepEqual(t, plainStringMap(set.MatchLabels()), map[string]string{"foo": "bar"})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L87
func TestStatefulSetGetsDefaults(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), PortNumber: jsii.Number(6000)}},
	})
	requireSnapshotHash(t, synth(t, chart), "008e11f7f5fd1b551b46bc69cf153bcb8774a38646e218cb46eae4ce075617ee")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L99
func TestStatefulSetAllowsOverrides(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	service := plus.NewService(chart, jsii.String("TestService"), &plus.ServiceProps{
		Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String("test-srv")}, Type: plus.ServiceType_NODE_PORT,
		Ports: &[]*plus.ServicePort{{Port: jsii.Number(9000), TargetPort: jsii.Number(9900)}},
	})
	plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}}, Replicas: jsii.Number(5),
		PodManagementPolicy: plus.PodManagementPolicy_PARALLEL,
		MinReady:            cdk8s.Duration_Seconds(jsii.Number(30)), Service: service,
	})
	spec := mapAt(t, manifestAt(t, chart, 1), "spec")
	for key, want := range map[string]interface{}{
		"replicas": float64(5), "serviceName": "test-srv", "podManagementPolicy": "Parallel", "minReadySeconds": float64(30),
	} {
		if got := spec[key]; got != want {
			t.Errorf("spec.%s = %#v, want %#v", key, got, want)
		}
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L125
func TestStatefulSetSynthesizesSpecLazily(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	set := plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{Service: newTestService(chart, "TestService", 9300)})
	set.AddContainer(&plus.ContainerProps{Image: jsii.String("image"), Port: jsii.Number(9300)})
	containers := sliceAt(t, manifestAt(t, chart, 1), "spec", "template", "spec", "containers")
	container := mapAt(t, containers[0])
	if got := container["image"]; got != "image" {
		t.Errorf("image = %#v, want image", got)
	}
	ports := sliceAt(t, container, "ports")
	if got := mapAt(t, ports[0])["containerPort"]; got != float64(9300) {
		t.Errorf("containerPort = %#v, want 9300", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L144
func TestStatefulSetDefaultUpdateStrategy(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	set := plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{Service: newTestService(chart, "Service", 80)})
	set.AddContainer(&plus.ContainerProps{Image: jsii.String("image")})
	if set.Strategy() == nil {
		t.Fatal("default update strategy is nil")
	}
	requireDeepEqual(t, mapAt(t, manifestAt(t, chart, 1), "spec")["updateStrategy"], map[string]interface{}{
		"type": "RollingUpdate", "rollingUpdate": map[string]interface{}{"partition": float64(0)},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L167
func TestStatefulSetCustomUpdateStrategy(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	strategy := plus.StatefulSetUpdateStrategy_OnDelete()
	set := plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{
		Service: newTestService(chart, "Service", 80), Strategy: strategy,
	})
	set.AddContainer(&plus.ContainerProps{Image: jsii.String("image")})
	if set.Strategy() != strategy {
		t.Fatal("StatefulSet did not retain the configured strategy")
	}
	requireDeepEqual(t, mapAt(t, manifestAt(t, chart, 1), "spec")["updateStrategy"], map[string]interface{}{"type": "OnDelete"})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L188
func TestStatefulSetCanBeIsolated(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("foobar")}},
		Service:    newTestService(chart, "Service", 80), Isolate: jsii.Bool(true),
	})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "1b816756413eaef8e5e1d328d9de9130e8025d0ca66e4673914c5387b22040fa")
	policy := mapAt(t, manifests[2], "spec")
	if labels := mapAt(t, policy, "podSelector", "matchLabels"); len(labels) == 0 {
		t.Fatal("isolating StatefulSet produced no pod selector labels")
	}
	requireDeepEqual(t, policy["policyTypes"], []interface{}{"Egress", "Ingress"})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L210
func TestStatefulSetVolumeClaimTemplates(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	secret := plus.NewSecret(chart, jsii.String("AwsSecret"), nil)
	data := plus.Volume_FromName(chart, jsii.String("data"), jsii.String("data"))
	temp := plus.Volume_FromName(chart, jsii.String("temp"), jsii.String("temp"))
	secretVolume := plus.Volume_FromSecret(chart, jsii.String("secret"), secret, nil)
	set := plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{
		Containers: &[]*plus.ContainerProps{{
			Image: jsii.String("foobar"), PortNumber: jsii.Number(80),
			VolumeMounts: &[]*plus.VolumeMount{
				{Volume: data, Path: jsii.String("/mnt/data")},
				{Volume: temp, Path: jsii.String("/mnt/temp")},
				{Volume: secretVolume, Path: jsii.String("/mnt/secret")},
			},
		}},
		VolumeClaimTemplates: &[]*plus.PersistentVolumeClaimTemplateProps{{
			Name: jsii.String("data"), Storage: cdk8s.Size_Gibibytes(jsii.Number(20)),
			AccessModes:      &[]plus.PersistentVolumeAccessMode{plus.PersistentVolumeAccessMode_READ_WRITE_ONCE_POD},
			StorageClassName: jsii.String("standard"),
		}},
	})
	set.AddVolumeClaimTemplate(&plus.PersistentVolumeClaimTemplateProps{
		Name: jsii.String("temp"), Storage: cdk8s.Size_Gibibytes(jsii.Number(20)),
		AccessModes:      &[]plus.PersistentVolumeAccessMode{plus.PersistentVolumeAccessMode_READ_WRITE_ONCE_POD},
		StorageClassName: jsii.String("standard"),
	})
	manifests := synth(t, chart)
	spec := mapAt(t, manifests[2], "spec")
	requireDeepEqual(t, mapAt(t, spec, "template", "spec")["volumes"], []interface{}{
		map[string]interface{}{"name": "secret-test-awssecret-c8d3e80d", "secret": map[string]interface{}{"secretName": "test-awssecret-c8d3e80d"}},
	})
	requireDeepEqual(t, spec["volumeClaimTemplates"], []interface{}{
		map[string]interface{}{"metadata": map[string]interface{}{"name": "data"}, "spec": map[string]interface{}{
			"accessModes": []interface{}{"ReadWriteOncePod"}, "resources": map[string]interface{}{"requests": map[string]interface{}{"storage": "20Gi"}}, "storageClassName": "standard",
		}},
		map[string]interface{}{"metadata": map[string]interface{}{"name": "temp"}, "spec": map[string]interface{}{
			"accessModes": []interface{}{"ReadWriteOncePod"}, "resources": map[string]interface{}{"requests": map[string]interface{}{"storage": "20Gi"}}, "storageClassName": "standard",
		}},
	})
	requireSnapshotHash(t, manifests, "c0ef8e3b255ed3cfbbf01f978562bb8af2634870d4cc1ac0e98896975022ab55")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/statefulset.test.ts#L273
func TestStatefulSetRejectsUnusedVolumeClaimTemplate(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewStatefulSet(chart, jsii.String("StatefulSet"), &plus.StatefulSetProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("foobar"), PortNumber: jsii.Number(80)}},
		VolumeClaimTemplates: &[]*plus.PersistentVolumeClaimTemplateProps{{
			Name: jsii.String("data"), Storage: cdk8s.Size_Gibibytes(jsii.Number(20)),
			AccessModes:      &[]plus.PersistentVolumeAccessMode{plus.PersistentVolumeAccessMode_READ_WRITE_ONCE_POD},
			StorageClassName: jsii.String("standard"),
		}},
	})
	requirePanicContains(t, `Volume claim template with name "data" is not used by any container mount`, func() {
		cdk8s.Testing_Synth(chart)
	})
}
