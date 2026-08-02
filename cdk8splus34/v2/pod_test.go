package cdk8splus34_test

import (
	"sort"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func podContainer(image string) *plus.ContainerProps {
	return &plus.ContainerProps{Image: jsii.String(image)}
}

func podContainers(images ...string) *[]*plus.ContainerProps {
	containers := make([]*plus.ContainerProps, 0, len(images))
	for _, image := range images {
		containers = append(containers, podContainer(image))
	}
	return &containers
}

func podSpec(t *testing.T, chart cdk8s.Chart) map[string]interface{} {
	t.Helper()
	return mapAt(t, manifestOfKind(t, chart, "Pod"), "spec")
}

func podFirstContainer(t *testing.T, spec map[string]interface{}, field string) map[string]interface{} {
	t.Helper()
	containers := sliceAt(t, spec, field)
	if len(containers) == 0 {
		t.Fatalf("%s is empty", field)
	}
	container, ok := containers[0].(map[string]interface{})
	if !ok {
		t.Fatalf("%s[0] has type %T, want map[string]interface{}", field, containers[0])
	}
	return container
}

func TestPodCore(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L6
	t.Run("defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})

		manifests := synth(t, chart)
		if len(manifests) != 1 {
			t.Fatalf("manifest count = %d, want 1", len(manifests))
		}
		requireDeepEqual(t, manifests[0], map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{"cdk8s.io/metadata.addr": "test-Pod-c815bc91"},
				"name":   "test-pod-c890e1b8",
			},
			"spec": map[string]interface{}{
				"automountServiceAccountToken": false,
				"containers": []interface{}{map[string]interface{}{
					"image":           "image",
					"imagePullPolicy": "Always",
					"name":            "main",
					"resources": map[string]interface{}{
						"limits":   map[string]interface{}{"cpu": "1500m", "memory": "2048Mi"},
						"requests": map[string]interface{}{"cpu": "1000m", "memory": "512Mi"},
					},
					"securityContext": map[string]interface{}{
						"allowPrivilegeEscalation": false,
						"privileged":               false,
						"readOnlyRootFilesystem":   true,
						"runAsNonRoot":             true,
					},
				}},
				"dnsPolicy":                     "ClusterFirst",
				"hostNetwork":                   false,
				"restartPolicy":                 "Always",
				"securityContext":               map[string]interface{}{"fsGroupChangePolicy": "Always", "runAsNonRoot": true},
				"setHostnameAsFQDN":             false,
				"shareProcessNamespace":         false,
				"terminationGracePeriodSeconds": float64(30),
			},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L21
	t.Run("fails with two volumes with the same name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		cm1 := plus.NewConfigMap(chart, jsii.String("cm1"), &plus.ConfigMapProps{Data: &map[string]*string{"f1": jsii.String("f1-content")}})
		cm2 := plus.NewConfigMap(chart, jsii.String("cm2"), &plus.ConfigMapProps{Data: &map[string]*string{"f2": jsii.String("f2-content")}})
		v1 := plus.Volume_FromConfigMap(chart, jsii.String("Volume1"), cm1, &plus.ConfigMapVolumeOptions{Name: jsii.String("v")})
		v2 := plus.Volume_FromConfigMap(chart, jsii.String("Volume2"), cm2, &plus.ConfigMapVolumeOptions{Name: jsii.String("v")})
		requirePanicContains(t, "Volume with name v already exists", func() {
			plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Volumes: &[]plus.Volume{v1, v2}})
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L35
	t.Run("fails adding a volume with the same name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		cm1 := plus.NewConfigMap(chart, jsii.String("cm1"), &plus.ConfigMapProps{Data: &map[string]*string{"f1": jsii.String("f1-content")}})
		cm2 := plus.NewConfigMap(chart, jsii.String("cm2"), &plus.ConfigMapProps{Data: &map[string]*string{"f2": jsii.String("f2-content")}})
		v1 := plus.Volume_FromConfigMap(chart, jsii.String("Volume1"), cm1, &plus.ConfigMapVolumeOptions{Name: jsii.String("v")})
		v2 := plus.Volume_FromConfigMap(chart, jsii.String("Volume2"), cm2, &plus.ConfigMapVolumeOptions{Name: jsii.String("v")})
		pod := plus.NewPod(chart, jsii.String("Pod"), nil)
		pod.AddVolume(v1)
		requirePanicContains(t, "Volume with name v already exists", func() { pod.AddVolume(v2) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L52
	t.Run("fails with a container that has mounts with different volumes of the same name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		cm1 := plus.NewConfigMap(chart, jsii.String("cm1"), &plus.ConfigMapProps{Data: &map[string]*string{"f1": jsii.String("f1-content")}})
		cm2 := plus.NewConfigMap(chart, jsii.String("cm2"), &plus.ConfigMapProps{Data: &map[string]*string{"f2": jsii.String("f2-content")}})
		v1 := plus.Volume_FromConfigMap(chart, jsii.String("Volume1"), cm1, &plus.ConfigMapVolumeOptions{Name: jsii.String("v")})
		v2 := plus.Volume_FromConfigMap(chart, jsii.String("Volume2"), cm2, &plus.ConfigMapVolumeOptions{Name: jsii.String("v")})
		mounts := []*plus.VolumeMount{
			{Volume: v1, Path: jsii.String("f1"), SubPath: jsii.String("f1")},
			{Volume: v2, Path: jsii.String("f2"), SubPath: jsii.String("f2")},
		}
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("nginx"), VolumeMounts: &mounts}}})
		requirePanicContains(t, "Invalid mount configuration. At least two different volumes have the same name: v", func() {
			cdk8s.Testing_Synth(chart)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L84
	t.Run("can configure multiple mounts with the same volume", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		config := plus.NewConfigMap(chart, jsii.String("Config"), &plus.ConfigMapProps{Data: &map[string]*string{
			"f1": jsii.String("f1-content"),
			"f2": jsii.String("f2-content"),
		}})
		volume := plus.Volume_FromConfigMap(chart, jsii.String("Volume"), config, nil)
		mounts := []*plus.VolumeMount{
			{Volume: volume, Path: jsii.String("f1"), SubPath: jsii.String("f1")},
			{Volume: volume, Path: jsii.String("f2"), SubPath: jsii.String("f2")},
		}
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), VolumeMounts: &mounts}}})
		requireDeepEqual(t, sliceAt(t, podSpec(t, chart), "volumes"), []interface{}{map[string]interface{}{
			"configMap": map[string]interface{}{"name": "test-config-c8c927dd"},
			"name":      "configmap-test-config-c8c927dd",
		}})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L128
	t.Run("defaultChild", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), nil)
		child, ok := pod.Node().DefaultChild().(cdk8s.ApiObject)
		if !ok {
			t.Fatalf("default child has type %T, want cdk8s.ApiObject", pod.Node().DefaultChild())
		}
		if got := stringValue(child.Kind()); got != "Pod" {
			t.Fatalf("default child kind = %q, want Pod", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L138
	t.Run("Can add container post instantiation", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), nil)
		pod.AddContainer(podContainer("image"))
		if got := podFirstContainer(t, podSpec(t, chart), "containers")["image"]; got != "image" {
			t.Fatalf("container image = %#v, want image", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L151
	t.Run("Can attach an existing container post instantiation", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), nil)
		pod.AttachContainer(plus.NewContainer(podContainer("image")))
		if got := podFirstContainer(t, podSpec(t, chart), "containers")["image"]; got != "image" {
			t.Fatalf("container image = %#v, want image", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L167
	t.Run("Must have at least one container", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), nil)
		requirePanicContains(t, "PodSpec must have at least 1 container", func() { cdk8s.Testing_Synth(chart) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L179
	t.Run("Can add volume post instantiation", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		pod.AddVolume(plus.Volume_FromEmptyDir(chart, jsii.String("Volume"), jsii.String("volume"), nil))
		volumes := sliceAt(t, podSpec(t, chart), "volumes")
		requireDeepEqual(t, volumes, []interface{}{map[string]interface{}{"name": "volume", "emptyDir": map[string]interface{}{}}})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L198
	t.Run("Automatically adds volumes from container mounts", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), nil)
		volume := plus.Volume_FromEmptyDir(chart, jsii.String("Volume"), jsii.String("volume"), nil)
		container := pod.AddContainer(podContainer("image"))
		container.Mount(jsii.String("/path/to/mount"), volume, nil)
		requireDeepEqual(t, sliceAt(t, podSpec(t, chart), "volumes"), []interface{}{map[string]interface{}{"name": "volume", "emptyDir": map[string]interface{}{}}})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L216
	t.Run("Synthesizes spec lazily", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{})
		pod.AddContainer(podContainer("image"))
		if got := podFirstContainer(t, podSpec(t, chart), "containers")["image"]; got != "image" {
			t.Fatalf("container image = %#v, want image", got)
		}
	})
}

func TestPodInitSecurityDNSAndOptions(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L230
	t.Run("init containers cannot have liveness probe", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		requirePanicContains(t, "Init containers must not have a liveness probe", func() {
			pod.AddInitContainer(&plus.ContainerProps{Image: jsii.String("image"), Liveness: plus.Probe_FromTcpSocket(nil)})
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L239
	t.Run("init containers cannot have readiness probe", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		requirePanicContains(t, "Init containers must not have a readiness probe", func() {
			pod.AddInitContainer(&plus.ContainerProps{Image: jsii.String("image"), Readiness: plus.Probe_FromTcpSocket(nil)})
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L248
	t.Run("init containers cannot have startup probe", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		requirePanicContains(t, "Init containers must not have a startup probe", func() {
			pod.AddInitContainer(&plus.ContainerProps{Image: jsii.String("image"), Startup: plus.Probe_FromTcpSocket(nil)})
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L257
	t.Run("sidecar containers can have liveness probe", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		pod.AddInitContainer(&plus.ContainerProps{Image: jsii.String("image"), Liveness: plus.Probe_FromTcpSocket(nil), RestartPolicy: plus.ContainerRestartPolicy_ALWAYS})
		if podFirstContainer(t, podSpec(t, chart), "initContainers")["livenessProbe"] == nil {
			t.Fatal("sidecar livenessProbe is absent")
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L270
	t.Run("sidecar containers can have readiness probe", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		pod.AddInitContainer(&plus.ContainerProps{Image: jsii.String("image"), Readiness: plus.Probe_FromTcpSocket(nil), RestartPolicy: plus.ContainerRestartPolicy_ALWAYS})
		if podFirstContainer(t, podSpec(t, chart), "initContainers")["readinessProbe"] == nil {
			t.Fatal("sidecar readinessProbe is absent")
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L283
	t.Run("sidecar containers can have startup probe", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		pod.AddInitContainer(&plus.ContainerProps{Image: jsii.String("image"), Startup: plus.Probe_FromTcpSocket(nil), RestartPolicy: plus.ContainerRestartPolicy_ALWAYS})
		if podFirstContainer(t, podSpec(t, chart), "initContainers")["startupProbe"] == nil {
			t.Fatal("sidecar startupProbe is absent")
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L296
	t.Run("can specify init containers at instantiation", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), InitContainers: podContainers("image")})
		if got := podFirstContainer(t, podSpec(t, chart), "initContainers")["image"]; got != "image" {
			t.Fatalf("init container image = %#v, want image", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L311
	t.Run("can add init container post instantiation", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		pod.AddInitContainer(podContainer("image"))
		if got := podFirstContainer(t, podSpec(t, chart), "initContainers")["image"]; got != "image" {
			t.Fatalf("init container image = %#v, want image", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L326
	t.Run("init container names are indexed", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		pod.AddInitContainer(podContainer("image1"))
		pod.AddInitContainer(podContainer("image2"))
		containers := sliceAt(t, podSpec(t, chart), "initContainers")
		if got := containers[0].(map[string]interface{})["name"]; got != "init-0" {
			t.Fatalf("first init container name = %#v, want init-0", got)
		}
		if got := containers[1].(map[string]interface{})["name"]; got != "init-1" {
			t.Fatalf("second init container name = %#v, want init-1", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L344
	t.Run("automatically adds volumes from init container mounts", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		volume := plus.Volume_FromEmptyDir(chart, jsii.String("Volume"), jsii.String("volume"), nil)
		container := pod.AddInitContainer(podContainer("image"))
		container.Mount(jsii.String("/path/to/mount"), volume, nil)
		requireDeepEqual(t, sliceAt(t, podSpec(t, chart), "volumes"), []interface{}{map[string]interface{}{"name": "volume", "emptyDir": map[string]interface{}{}}})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L364
	t.Run("default security context", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		security := pod.SecurityContext()
		if !boolValue(security.EnsureNonRoot()) {
			t.Fatal("ensureNonRoot = false, want true")
		}
		if got := len(*security.Sysctls()); got != 0 {
			t.Fatalf("sysctls length = %d, want 0", got)
		}
		if security.FsGroup() != nil || security.User() != nil || security.Group() != nil {
			t.Fatalf("default numeric security fields are present: fsGroup=%v user=%v group=%v", security.FsGroup(), security.User(), security.Group())
		}
		if got := security.FsGroupChangePolicy(); got != plus.FsGroupChangePolicy_ALWAYS {
			t.Fatalf("fsGroupChangePolicy = %q, want ALWAYS", got)
		}
		requireDeepEqual(t, mapAt(t, podSpec(t, chart), "securityContext"), map[string]interface{}{
			"fsGroupChangePolicy": "Always",
			"runAsNonRoot":        true,
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L390
	t.Run("custom security context", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		sysctls := []*plus.Sysctl{{Name: jsii.String("s1"), Value: jsii.String("v1")}}
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{
			Containers: podContainers("image"),
			SecurityContext: &plus.PodSecurityContextProps{
				EnsureNonRoot:       jsii.Bool(true),
				FsGroup:             jsii.Number(5000),
				FsGroupChangePolicy: plus.FsGroupChangePolicy_ON_ROOT_MISMATCH,
				Group:               jsii.Number(2000),
				User:                jsii.Number(1000),
				Sysctls:             &sysctls,
			},
		})
		security := pod.SecurityContext()
		if !boolValue(security.EnsureNonRoot()) || numberValue(security.FsGroup()) != 5000 || numberValue(security.User()) != 1000 || numberValue(security.Group()) != 2000 {
			t.Fatalf("custom security getters = ensureNonRoot:%v fsGroup:%v user:%v group:%v", security.EnsureNonRoot(), security.FsGroup(), security.User(), security.Group())
		}
		if got := security.FsGroupChangePolicy(); got != plus.FsGroupChangePolicy_ON_ROOT_MISMATCH {
			t.Fatalf("fsGroupChangePolicy = %q, want ON_ROOT_MISMATCH", got)
		}
		gotSysctls := *security.Sysctls()
		if len(gotSysctls) != 1 || stringValue(gotSysctls[0].Name) != "s1" || stringValue(gotSysctls[0].Value) != "v1" {
			t.Fatalf("sysctls = %#v", gotSysctls)
		}
		requireDeepEqual(t, mapAt(t, podSpec(t, chart), "securityContext"), map[string]interface{}{
			"fsGroup":             float64(5000),
			"fsGroupChangePolicy": "OnRootMismatch",
			"runAsGroup":          float64(2000),
			"runAsNonRoot":        true,
			"runAsUser":           float64(1000),
			"sysctls":             []interface{}{map[string]interface{}{"name": "s1", "value": "v1"}},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L415
	t.Run("custom host aliases", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		aliases := []*plus.HostAlias{{Ip: jsii.String("127.0.0.1"), Hostnames: &[]*string{jsii.String("foo.local"), jsii.String("bar.local")}}}
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), HostAliases: &aliases})
		pod.AddHostAlias(&plus.HostAlias{Ip: jsii.String("10.1.2.3"), Hostnames: &[]*string{jsii.String("foo.remote"), jsii.String("bar.remote")}})
		requireDeepEqual(t, sliceAt(t, podSpec(t, chart), "hostAliases"), []interface{}{
			map[string]interface{}{"ip": "127.0.0.1", "hostnames": []interface{}{"foo.local", "bar.local"}},
			map[string]interface{}{"ip": "10.1.2.3", "hostnames": []interface{}{"foo.remote", "bar.remote"}},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L434
	t.Run("default dns settings", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		dns := pod.Dns()
		if dns.Hostname() != nil || dns.Subdomain() != nil || boolValue(dns.HostnameAsFQDN()) || dns.Policy() != plus.DnsPolicy_CLUSTER_FIRST || len(*dns.Searches()) != 0 || len(*dns.Nameservers()) != 0 || len(*dns.Options()) != 0 {
			t.Fatalf("unexpected default DNS getters: hostname=%v subdomain=%v fqdn=%v policy=%q searches=%v nameservers=%v options=%v", dns.Hostname(), dns.Subdomain(), dns.HostnameAsFQDN(), dns.Policy(), dns.Searches(), dns.Nameservers(), dns.Options())
		}
		spec := podSpec(t, chart)
		if _, ok := spec["hostname"]; ok {
			t.Fatal("default spec contains hostname")
		}
		if _, ok := spec["subdomain"]; ok {
			t.Fatal("default spec contains subdomain")
		}
		if _, ok := spec["dnsConfig"]; ok {
			t.Fatal("default spec contains dnsConfig")
		}
		if spec["setHostnameAsFQDN"] != false || spec["dnsPolicy"] != "ClusterFirst" {
			t.Fatalf("default DNS spec = %#v", spec)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L459
	t.Run("custom dns settings", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		options := []*plus.DnsOption{{Name: jsii.String("opt1"), Value: jsii.String("opt1-value")}}
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{
			Containers: podContainers("image"),
			Dns: &plus.PodDnsProps{
				Hostname:       jsii.String("hostname"),
				Subdomain:      jsii.String("subdomain"),
				HostnameAsFQDN: jsii.Bool(true),
				Nameservers:    &[]*string{jsii.String("n1")},
				Searches:       &[]*string{jsii.String("s1")},
				Options:        &options,
				Policy:         plus.DnsPolicy_DEFAULT,
			},
		})
		dns := pod.Dns()
		dns.AddNameserver(jsii.String("n2"))
		dns.AddSearch(jsii.String("s2"))
		dns.AddOption(&plus.DnsOption{Name: jsii.String("opt2")})
		if stringValue(dns.Hostname()) != "hostname" || stringValue(dns.Subdomain()) != "subdomain" || !boolValue(dns.HostnameAsFQDN()) || dns.Policy() != plus.DnsPolicy_DEFAULT {
			t.Fatalf("custom DNS getters = hostname:%v subdomain:%v fqdn:%v policy:%q", dns.Hostname(), dns.Subdomain(), dns.HostnameAsFQDN(), dns.Policy())
		}
		if got := []string{stringValue((*dns.Searches())[0]), stringValue((*dns.Searches())[1])}; got[0] != "s1" || got[1] != "s2" {
			t.Fatalf("searches = %#v", got)
		}
		if got := []string{stringValue((*dns.Nameservers())[0]), stringValue((*dns.Nameservers())[1])}; got[0] != "n1" || got[1] != "n2" {
			t.Fatalf("nameservers = %#v", got)
		}
		dnsConfig := mapAt(t, podSpec(t, chart), "dnsConfig")
		requireDeepEqual(t, dnsConfig, map[string]interface{}{
			"searches":    []interface{}{"s1", "s2"},
			"nameservers": []interface{}{"n1", "n2"},
			"options": []interface{}{
				map[string]interface{}{"name": "opt1", "value": "opt1-value"},
				map[string]interface{}{"name": "opt2"},
			},
		})
		spec := podSpec(t, chart)
		if spec["hostname"] != "hostname" || spec["subdomain"] != "subdomain" || spec["setHostnameAsFQDN"] != true || spec["dnsPolicy"] != "Default" {
			t.Fatalf("custom DNS spec = %#v", spec)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L499
	t.Run("throws if more than 3 nameservers are configured", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), Dns: &plus.PodDnsProps{Nameservers: &[]*string{jsii.String("n1"), jsii.String("n2"), jsii.String("n3"), jsii.String("n4")}}})
		requirePanicContains(t, "There can be at most 3 nameservers specified", func() { cdk8s.Testing_Synth(chart) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L513
	t.Run("throws if more than 6 search domains are configured", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), Dns: &plus.PodDnsProps{Searches: &[]*string{jsii.String("s1"), jsii.String("s2"), jsii.String("s3"), jsii.String("s4"), jsii.String("s5"), jsii.String("s6"), jsii.String("s7")}}})
		requirePanicContains(t, "There can be at most 6 search domains specified", func() { cdk8s.Testing_Synth(chart) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L528
	t.Run("throws if no nameservers are given when dns policy is set to NONE", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), Dns: &plus.PodDnsProps{Policy: plus.DnsPolicy_NONE}})
		requirePanicContains(t, "When dns policy is set to NONE, at least one nameserver is required", func() { cdk8s.Testing_Synth(chart) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L543
	t.Run("can configure auth to docker registry", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		auth := plus.NewDockerConfigSecret(chart, jsii.String("Secret"), &plus.DockerConfigSecretProps{Data: &map[string]interface{}{"foo": "bar"}})
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), DockerRegistryAuth: auth})
		secrets := sliceAt(t, podSpec(t, chart), "imagePullSecrets")
		requireDeepEqual(t, secrets, []interface{}{map[string]interface{}{"name": stringValue(auth.Name())}})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L562
	t.Run("auto mounting token defaults to true", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		if boolValue(pod.AutomountServiceAccountToken()) || podSpec(t, chart)["automountServiceAccountToken"] != false {
			t.Fatalf("automountServiceAccountToken getter/spec = %v/%#v, want false/false", pod.AutomountServiceAccountToken(), podSpec(t, chart)["automountServiceAccountToken"])
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L576
	t.Run("auto mounting token can be disabled", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), AutomountServiceAccountToken: jsii.Bool(false)})
		if boolValue(pod.AutomountServiceAccountToken()) || podSpec(t, chart)["automountServiceAccountToken"] != false {
			t.Fatalf("automountServiceAccountToken getter/spec = %v/%#v, want false/false", pod.AutomountServiceAccountToken(), podSpec(t, chart)["automountServiceAccountToken"])
		}
	})
}

func podSpecByImage(t *testing.T, chart cdk8s.Chart, image string) map[string]interface{} {
	t.Helper()
	for _, candidate := range synth(t, chart) {
		manifest, ok := candidate.(map[string]interface{})
		if !ok || manifest["kind"] != "Pod" {
			continue
		}
		spec, ok := manifest["spec"].(map[string]interface{})
		if !ok {
			continue
		}
		containers, ok := spec["containers"].([]interface{})
		if !ok || len(containers) == 0 {
			continue
		}
		container, ok := containers[0].(map[string]interface{})
		if ok && container["image"] == image {
			return spec
		}
	}
	t.Fatalf("no Pod with first container image %q", image)
	return nil
}

func podAffinityTerm(t *testing.T, chart cdk8s.Chart, image, affinityKind, schedulingKind string) map[string]interface{} {
	t.Helper()
	affinity := mapAt(t, podSpecByImage(t, chart, image), "affinity", affinityKind)
	entries := sliceAt(t, affinity, schedulingKind)
	if len(entries) != 1 {
		t.Fatalf("%s.%s entry count = %d, want 1", affinityKind, schedulingKind, len(entries))
	}
	entry, ok := entries[0].(map[string]interface{})
	if !ok {
		t.Fatalf("affinity entry has type %T, want map[string]interface{}", entries[0])
	}
	return entry
}

func TestPodScheduling(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L593
	t.Run("only NO_EXECUTE taint queries can specify eviction", func(t *testing.T) {
		requirePanicContains(t, "Only 'NO_EXECUTE' effects can specify 'evictAfter'", func() {
			plus.NodeTaintQuery_Is(jsii.String("key"), jsii.String("value"), &plus.NodeTaintQueryOptions{
				Effect:     plus.TaintEffect_NO_SCHEDULE,
				EvictAfter: cdk8s.Duration_Hours(jsii.Number(1)),
			})
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L602
	t.Run("can tolerate tainted nodes", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		devNodes := plus.Node_Tainted(
			plus.NodeTaintQuery_Is(jsii.String("key1"), jsii.String("value1"), nil),
			plus.NodeTaintQuery_Is(jsii.String("key2"), jsii.String("value2"), &plus.NodeTaintQueryOptions{Effect: plus.TaintEffect_PREFER_NO_SCHEDULE}),
			plus.NodeTaintQuery_Exists(jsii.String("key3"), nil),
			plus.NodeTaintQuery_Exists(jsii.String("key4"), &plus.NodeTaintQueryOptions{Effect: plus.TaintEffect_NO_SCHEDULE}),
			plus.NodeTaintQuery_Is(jsii.String("key5"), jsii.String("value5"), &plus.NodeTaintQueryOptions{Effect: plus.TaintEffect_NO_EXECUTE, EvictAfter: cdk8s.Duration_Hours(jsii.Number(1))}),
			plus.NodeTaintQuery_Any(),
		)
		redis := plus.NewPod(chart, jsii.String("Redis"), &plus.PodProps{Containers: podContainers("redis")})
		redis.Scheduling().Tolerate(devNodes)
		requireDeepEqual(t, sliceAt(t, podSpecByImage(t, chart, "redis"), "tolerations"), []interface{}{
			map[string]interface{}{"key": "key1", "operator": "Equal", "value": "value1"},
			map[string]interface{}{"effect": "PreferNoSchedule", "key": "key2", "operator": "Equal", "value": "value2"},
			map[string]interface{}{"key": "key3", "operator": "Exists"},
			map[string]interface{}{"effect": "NoSchedule", "key": "key4", "operator": "Exists"},
			map[string]interface{}{"effect": "NoExecute", "key": "key5", "operator": "Equal", "tolerationSeconds": float64(3600), "value": "value5"},
			map[string]interface{}{"operator": "Exists"},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L624
	t.Run("can be assigned to a node by name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		redis := plus.NewPod(chart, jsii.String("Redis"), &plus.PodProps{Containers: podContainers("redis")})
		redis.Scheduling().Assign(plus.Node_Named(jsii.String("node1")))
		if got := podSpecByImage(t, chart, "redis")["nodeName"]; got != "node1" {
			t.Fatalf("nodeName = %#v, want node1", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L635
	t.Run("can be attracted to a node by selector - default", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		redis := plus.NewPod(chart, jsii.String("Redis"), &plus.PodProps{Containers: podContainers("redis")})
		redis.Scheduling().Attract(plus.Node_Labeled(plus.NodeLabelQuery_Is(jsii.String("memory"), jsii.String("high"))), nil)
		nodeAffinity := mapAt(t, podSpecByImage(t, chart, "redis"), "affinity", "nodeAffinity", "requiredDuringSchedulingIgnoredDuringExecution")
		requireDeepEqual(t, sliceAt(t, nodeAffinity, "nodeSelectorTerms"), []interface{}{map[string]interface{}{
			"matchExpressions": []interface{}{map[string]interface{}{"key": "memory", "operator": "In", "values": []interface{}{"high"}}},
		}})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L646
	t.Run("can be attracted to a node by selector - custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		redis := plus.NewPod(chart, jsii.String("Redis"), &plus.PodProps{Containers: podContainers("redis")})
		redis.Scheduling().Attract(plus.Node_Labeled(plus.NodeLabelQuery_Is(jsii.String("memory"), jsii.String("high"))), &plus.PodSchedulingAttractOptions{Weight: jsii.Number(1)})
		entries := sliceAt(t, mapAt(t, podSpecByImage(t, chart, "redis"), "affinity", "nodeAffinity"), "preferredDuringSchedulingIgnoredDuringExecution")
		requireDeepEqual(t, entries, []interface{}{map[string]interface{}{
			"preference": map[string]interface{}{"matchExpressions": []interface{}{map[string]interface{}{"key": "memory", "operator": "In", "values": []interface{}{"high"}}}},
			"weight":     float64(1),
		}})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L661
	t.Run("can be co-located with a managed pod - default", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		redis := plus.NewPod(chart, jsii.String("Redis"), &plus.PodProps{Containers: podContainers("redis")})
		web := plus.NewPod(chart, jsii.String("Web"), &plus.PodProps{Containers: podContainers("web")})
		web.Scheduling().Colocate(redis, nil)
		term := podAffinityTerm(t, chart, "web", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution")
		if term["topologyKey"] != "kubernetes.io/hostname" {
			t.Fatalf("topologyKey = %#v", term["topologyKey"])
		}
		requireDeepEqual(t, mapAt(t, term, "labelSelector", "matchLabels"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-Redis-c8b1633b"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L678
	t.Run("can be co-located with a managed pod - custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		redis := plus.NewPod(chart, jsii.String("Redis"), &plus.PodProps{Containers: podContainers("redis")})
		web := plus.NewPod(chart, jsii.String("Web"), &plus.PodProps{Containers: podContainers("web")})
		web.Scheduling().Colocate(redis, &plus.PodSchedulingColocateOptions{Topology: plus.Topology_ZONE(), Weight: jsii.Number(1)})
		entry := podAffinityTerm(t, chart, "web", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution")
		if entry["weight"] != float64(1) {
			t.Fatalf("weight = %#v, want 1", entry["weight"])
		}
		term := mapAt(t, entry, "podAffinityTerm")
		if term["topologyKey"] != "topology.kubernetes.io/zone" {
			t.Fatalf("topologyKey = %#v", term["topologyKey"])
		}
		requireDeepEqual(t, mapAt(t, term, "labelSelector", "matchLabels"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-Redis-c8b1633b"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L698
	t.Run("can be co-located with an unmanaged pod - default", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		redis := plus.Pods_Select(chart, jsii.String("Redis"), &plus.PodsSelectOptions{
			Labels:     &map[string]*string{"app": jsii.String("store")},
			Namespaces: plus.Namespaces_All(chart, jsii.String("All")),
		})
		web := plus.NewPod(chart, jsii.String("Web"), &plus.PodProps{Containers: podContainers("web")})
		web.Scheduling().Colocate(redis, nil)
		term := podAffinityTerm(t, chart, "web", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution")
		requireDeepEqual(t, mapAt(t, term, "labelSelector", "matchLabels"), map[string]interface{}{"app": "store"})
		requireDeepEqual(t, mapAt(t, term, "namespaceSelector"), map[string]interface{}{})
		if term["topologyKey"] != "kubernetes.io/hostname" {
			t.Fatalf("topologyKey = %#v", term["topologyKey"])
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L717
	t.Run("can be co-located with an unmanaged pod - custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		redis := plus.Pods_Select(chart, jsii.String("Redis"), &plus.PodsSelectOptions{Labels: &map[string]*string{"app": jsii.String("store")}})
		web := plus.NewPod(chart, jsii.String("Web"), &plus.PodProps{Containers: podContainers("web")})
		web.Scheduling().Colocate(redis, &plus.PodSchedulingColocateOptions{Topology: plus.Topology_ZONE(), Weight: jsii.Number(1)})
		entry := podAffinityTerm(t, chart, "web", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution")
		term := mapAt(t, entry, "podAffinityTerm")
		if entry["weight"] != float64(1) || term["topologyKey"] != "topology.kubernetes.io/zone" {
			t.Fatalf("preferred affinity = %#v", entry)
		}
		requireDeepEqual(t, mapAt(t, term, "labelSelector", "matchLabels"), map[string]interface{}{"app": "store"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L736
	t.Run("can be separated from a managed pod - default", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		redis := plus.NewPod(chart, jsii.String("Redis"), &plus.PodProps{Containers: podContainers("redis")})
		web := plus.NewPod(chart, jsii.String("Web"), &plus.PodProps{Containers: podContainers("web")})
		web.Scheduling().Separate(redis, nil)
		term := podAffinityTerm(t, chart, "web", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution")
		if term["topologyKey"] != "kubernetes.io/hostname" {
			t.Fatalf("topologyKey = %#v", term["topologyKey"])
		}
		requireDeepEqual(t, mapAt(t, term, "labelSelector", "matchLabels"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-Redis-c8b1633b"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L753
	t.Run("can be separated from a managed pod - custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		redis := plus.NewPod(chart, jsii.String("Redis"), &plus.PodProps{Containers: podContainers("redis")})
		web := plus.NewPod(chart, jsii.String("Web"), &plus.PodProps{Containers: podContainers("web")})
		web.Scheduling().Separate(redis, &plus.PodSchedulingSeparateOptions{Topology: plus.Topology_ZONE(), Weight: jsii.Number(1)})
		entry := podAffinityTerm(t, chart, "web", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution")
		term := mapAt(t, entry, "podAffinityTerm")
		if entry["weight"] != float64(1) || term["topologyKey"] != "topology.kubernetes.io/zone" {
			t.Fatalf("preferred anti-affinity = %#v", entry)
		}
		requireDeepEqual(t, mapAt(t, term, "labelSelector", "matchLabels"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-Redis-c8b1633b"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L773
	t.Run("can be separated from an unmanaged pod - default", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		namespaces := plus.Namespaces_Select(chart, jsii.String("WebNamespace"), &plus.NamespacesSelectOptions{
			Labels: &map[string]*string{"net": jsii.String("1")},
			Names:  &[]*string{jsii.String("web")},
		})
		redis := plus.Pods_Select(chart, jsii.String("Redis"), &plus.PodsSelectOptions{Labels: &map[string]*string{"app": jsii.String("store")}, Namespaces: namespaces})
		web := plus.NewPod(chart, jsii.String("Web"), &plus.PodProps{Containers: podContainers("web")})
		web.Scheduling().Separate(redis, nil)
		term := podAffinityTerm(t, chart, "web", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution")
		requireDeepEqual(t, mapAt(t, term, "labelSelector", "matchLabels"), map[string]interface{}{"app": "store"})
		requireDeepEqual(t, mapAt(t, term, "namespaceSelector", "matchLabels"), map[string]interface{}{"net": "1"})
		requireDeepEqual(t, term["namespaces"], []interface{}{"web"})
		if term["topologyKey"] != "kubernetes.io/hostname" {
			t.Fatalf("topologyKey = %#v", term["topologyKey"])
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L792
	t.Run("can be separated from an unmanaged pod - custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		redis := plus.Pods_Select(chart, jsii.String("Redis"), &plus.PodsSelectOptions{Labels: &map[string]*string{"app": jsii.String("store")}})
		web := plus.NewPod(chart, jsii.String("Web"), &plus.PodProps{Containers: podContainers("web")})
		web.Scheduling().Separate(redis, &plus.PodSchedulingSeparateOptions{Topology: plus.Topology_ZONE(), Weight: jsii.Number(1)})
		entry := podAffinityTerm(t, chart, "web", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution")
		term := mapAt(t, entry, "podAffinityTerm")
		if entry["weight"] != float64(1) || term["topologyKey"] != "topology.kubernetes.io/zone" {
			t.Fatalf("preferred anti-affinity = %#v", entry)
		}
		requireDeepEqual(t, mapAt(t, term, "labelSelector", "matchLabels"), map[string]interface{}{"app": "store"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L813
	t.Run("can select pods", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		expressions := []plus.LabelExpression{plus.LabelExpression_Exists(jsii.String("key"))}
		selected := plus.Pods_Select(chart, jsii.String("Selected"), &plus.PodsSelectOptions{
			Labels:      &map[string]*string{"foo": jsii.String("bar")},
			Expressions: &expressions,
			Namespaces:  plus.Namespaces_Select(chart, jsii.String("Bar"), &plus.NamespacesSelectOptions{Labels: &map[string]*string{"foo": jsii.String("bar")}}),
		})
		consumer := plus.NewPod(chart, jsii.String("Consumer"), &plus.PodProps{Containers: podContainers("consumer")})
		consumer.Scheduling().Colocate(selected, nil)
		term := podAffinityTerm(t, chart, "consumer", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution")
		requireDeepEqual(t, mapAt(t, term, "labelSelector"), map[string]interface{}{
			"matchExpressions": []interface{}{map[string]interface{}{"key": "key", "operator": "Exists"}},
			"matchLabels":      map[string]interface{}{"foo": "bar"},
		})
		requireDeepEqual(t, mapAt(t, term, "namespaceSelector"), map[string]interface{}{"matchLabels": map[string]interface{}{"foo": "bar"}})
		if _, exists := term["namespaces"]; exists {
			t.Fatalf("selector unexpectedly includes namespace names: %#v", term["namespaces"])
		}
	})
}

func podConnectionPod(chart cdk8s.Chart, id, namespace string, port float64) plus.Pod {
	props := &plus.PodProps{Containers: podContainers("pod")}
	if namespace != "" {
		props.Metadata = &cdk8s.ApiObjectMetadata{Namespace: jsii.String(namespace)}
	}
	if port != 0 {
		props.Containers = &[]*plus.ContainerProps{{Image: jsii.String("pod"), Port: jsii.Number(port)}}
	}
	return plus.NewPod(chart, jsii.String(id), props)
}

func podPolicySpecs(t *testing.T, chart cdk8s.Chart) []map[string]interface{} {
	t.Helper()
	policies := podManifestsOfKind(t, chart, "NetworkPolicy")
	result := make([]map[string]interface{}, 0, len(policies))
	for _, policy := range policies {
		result = append(result, mapAt(t, policy, "spec"))
	}
	return result
}

func podPoliciesForDirection(t *testing.T, chart cdk8s.Chart, direction string) []map[string]interface{} {
	t.Helper()
	result := make([]map[string]interface{}, 0)
	for _, policy := range podManifestsOfKind(t, chart, "NetworkPolicy") {
		spec := mapAt(t, policy, "spec")
		if _, exists := spec[direction]; exists {
			result = append(result, policy)
		}
	}
	return result
}

func podPolicyRule(t *testing.T, policy map[string]interface{}, direction string) map[string]interface{} {
	t.Helper()
	rules := sliceAt(t, policy, "spec", direction)
	if len(rules) != 1 {
		t.Fatalf("%s rule count = %d, want 1", direction, len(rules))
	}
	rule, ok := rules[0].(map[string]interface{})
	if !ok {
		t.Fatalf("%s rule has type %T, want map[string]interface{}", direction, rules[0])
	}
	return rule
}

func podRulePeers(t *testing.T, rule map[string]interface{}, direction string) []interface{} {
	t.Helper()
	key := "to"
	if direction == "ingress" {
		key = "from"
	}
	peers, ok := rule[key].([]interface{})
	if !ok {
		t.Fatalf("%s peer list has type %T, want []interface{}", direction, rule[key])
	}
	return peers
}

func podPolicyMetadataNamespace(t *testing.T, policy map[string]interface{}) string {
	t.Helper()
	namespace, _ := mapAt(t, policy, "metadata")["namespace"].(string)
	return namespace
}

func TestPodConnectionsAllowTo(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L827
	t.Run("can allow to ip block", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		pod.Connections().AllowTo(plus.NetworkPolicyIpBlock_AnyIpv4(chart, jsii.String("AnyIpv4")), nil)
		policies := podPoliciesForDirection(t, chart, "egress")
		if len(policies) != 1 || len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 1 {
			t.Fatalf("egress/total policy counts = %d/%d, want 1/1", len(policies), len(podManifestsOfKind(t, chart, "NetworkPolicy")))
		}
		rule := podPolicyRule(t, policies[0], "egress")
		requireDeepEqual(t, podRulePeers(t, rule, "egress"), []interface{}{map[string]interface{}{"ipBlock": map[string]interface{}{"cidr": "0.0.0.0/0"}}})
		requireDeepEqual(t, rule["ports"], []interface{}{})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L839
	t.Run("can isolate pod", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("pod"), Isolate: jsii.Bool(true)})
		policies := podManifestsOfKind(t, chart, "NetworkPolicy")
		if len(policies) != 1 {
			t.Fatalf("NetworkPolicy count = %d, want 1", len(policies))
		}
		spec := mapAt(t, policies[0], "spec")
		requireDeepEqual(t, spec["policyTypes"], []interface{}{"Egress", "Ingress"})
		labels := mapAt(t, spec, "podSelector", "matchLabels")
		if labels["cdk8s.io/metadata.addr"] == nil {
			t.Fatalf("isolating policy selector = %#v", labels)
		}
		if _, exists := spec["egress"]; exists {
			t.Fatalf("default-deny policy unexpectedly has egress rules: %#v", spec["egress"])
		}
		if _, exists := spec["ingress"]; exists {
			t.Fatalf("default-deny policy unexpectedly has ingress rules: %#v", spec["ingress"])
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L855
	t.Run("can allow to managed pod", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 0)
		ports := []plus.NetworkPolicyPort{plus.NetworkPolicyPort_Tcp(jsii.Number(4444))}
		pod1.Connections().AllowTo(pod2, &plus.PodConnectionsAllowToOptions{Ports: &ports})
		egress := podPoliciesForDirection(t, chart, "egress")
		ingress := podPoliciesForDirection(t, chart, "ingress")
		if len(egress) != 1 || len(ingress) != 1 {
			t.Fatalf("egress/ingress policy counts = %d/%d, want 1/1", len(egress), len(ingress))
		}
		egressRule := podPolicyRule(t, egress[0], "egress")
		ingressRule := podPolicyRule(t, ingress[0], "ingress")
		wantPorts := []interface{}{map[string]interface{}{"port": float64(4444), "protocol": "TCP"}}
		requireDeepEqual(t, egressRule["ports"], wantPorts)
		requireDeepEqual(t, ingressRule["ports"], wantPorts)
		requireDeepEqual(t, mapAt(t, podRulePeers(t, egressRule, "egress")[0], "podSelector", "matchLabels"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-Pod2-c82dc44e"})
		requireDeepEqual(t, mapAt(t, podRulePeers(t, ingressRule, "ingress")[0], "podSelector", "matchLabels"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-Pod1-c8591188"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L871
	t.Run("can allow to managed workload resource", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{Containers: podContainers("pod")})
		pod.Connections().AllowTo(deployment, nil)
		if len(podPoliciesForDirection(t, chart, "egress")) != 1 || len(podPoliciesForDirection(t, chart, "ingress")) != 1 {
			t.Fatalf("managed workload did not create paired policies: %#v", podPolicySpecs(t, chart))
		}
		rule := podPolicyRule(t, podPoliciesForDirection(t, chart, "egress")[0], "egress")
		selector := mapAt(t, podRulePeers(t, rule, "egress")[0], "podSelector", "matchLabels")
		if selector["cdk8s.io/metadata.addr"] == nil {
			t.Fatalf("deployment selector = %#v", selector)
		}
		if len(podManifestsOfKind(t, chart, "Deployment")) != 1 {
			t.Fatal("managed Deployment was not synthesized")
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L887
	t.Run("can allow to pods selected without namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		selected := plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{Labels: &map[string]*string{"type": jsii.String("selected")}})
		pod.Connections().AllowTo(selected, nil)
		if len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 2 {
			t.Fatalf("NetworkPolicy count = %d, want 2", len(podManifestsOfKind(t, chart, "NetworkPolicy")))
		}
		egressRule := podPolicyRule(t, podPoliciesForDirection(t, chart, "egress")[0], "egress")
		requireDeepEqual(t, mapAt(t, podRulePeers(t, egressRule, "egress")[0], "podSelector", "matchLabels"), map[string]interface{}{"type": "selected"})
		opposite := mapAt(t, podPoliciesForDirection(t, chart, "ingress")[0], "spec", "podSelector", "matchLabels")
		requireDeepEqual(t, opposite, map[string]interface{}{"type": "selected"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L901
	t.Run("can allow to pods selected with namespaces selected by names", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		namespaces := plus.Namespaces_Select(chart, jsii.String("Namespaces"), &plus.NamespacesSelectOptions{Names: &[]*string{jsii.String("selected1"), jsii.String("selected2")}})
		selected := plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{Labels: &map[string]*string{"type": jsii.String("selected")}, Namespaces: namespaces})
		pod.Connections().AllowTo(selected, nil)
		policies := podManifestsOfKind(t, chart, "NetworkPolicy")
		if len(policies) != 3 {
			t.Fatalf("NetworkPolicy count = %d, want 3", len(policies))
		}
		egressRule := podPolicyRule(t, podPoliciesForDirection(t, chart, "egress")[0], "egress")
		peers := podRulePeers(t, egressRule, "egress")
		if len(peers) != 2 {
			t.Fatalf("egress peer count = %d, want 2", len(peers))
		}
		names := make([]string, 0, 2)
		for _, peer := range peers {
			names = append(names, mapAt(t, peer, "namespaceSelector", "matchLabels")["kubernetes.io/metadata.name"].(string))
		}
		sort.Strings(names)
		requireDeepEqual(t, names, []string{"selected1", "selected2"})
		ingress := podPoliciesForDirection(t, chart, "ingress")
		gotNamespaces := []string{podPolicyMetadataNamespace(t, ingress[0]), podPolicyMetadataNamespace(t, ingress[1])}
		sort.Strings(gotNamespaces)
		requireDeepEqual(t, gotNamespaces, []string{"selected1", "selected2"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L918
	t.Run("cannot allow to pods selected with namespaces selected by labels", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		namespaces := plus.Namespaces_Select(chart, jsii.String("Namespaces"), &plus.NamespacesSelectOptions{Labels: &map[string]*string{"type": jsii.String("selected")}})
		selected := plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{Labels: &map[string]*string{"type": jsii.String("selected")}, Namespaces: namespaces})
		requirePanicContains(t, "Unable to create an Ingress policy for peer 'test/Pods' (pod=test-pod-c890e1b8). Peer must specify namespaces only by name", func() {
			pod.Connections().AllowTo(selected, nil)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L934
	t.Run("cannot allow to pods selected in all namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		selected := plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{
			Labels:     &map[string]*string{"type": jsii.String("selected")},
			Namespaces: plus.Namespaces_All(chart, jsii.String("AllNamespaces")),
		})
		requirePanicContains(t, "Unable to create an Ingress policy for peer 'test/Pods' (pod=test-pod-c890e1b8). Peer must specify namespace names", func() {
			pod.Connections().AllowTo(selected, nil)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L950
	t.Run("can allow to all pods", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		pod.Connections().AllowTo(plus.Pods_All(chart, jsii.String("AllPods"), nil), nil)
		if len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 2 {
			t.Fatalf("NetworkPolicy count = %d, want 2", len(podManifestsOfKind(t, chart, "NetworkPolicy")))
		}
		rule := podPolicyRule(t, podPoliciesForDirection(t, chart, "egress")[0], "egress")
		requireDeepEqual(t, mapAt(t, podRulePeers(t, rule, "egress")[0], "podSelector"), map[string]interface{}{})
		requireDeepEqual(t, mapAt(t, podPoliciesForDirection(t, chart, "ingress")[0], "spec", "podSelector"), map[string]interface{}{})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L964
	t.Run("can allow to managed namespace", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		namespace := plus.NewNamespace(chart, jsii.String("Namespace"), nil)
		pod.Connections().AllowTo(namespace, nil)
		if len(podManifestsOfKind(t, chart, "Namespace")) != 1 || len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 2 {
			t.Fatalf("Namespace/NetworkPolicy counts = %d/%d, want 1/2", len(podManifestsOfKind(t, chart, "Namespace")), len(podManifestsOfKind(t, chart, "NetworkPolicy")))
		}
		rule := podPolicyRule(t, podPoliciesForDirection(t, chart, "egress")[0], "egress")
		peer := podRulePeers(t, rule, "egress")[0]
		requireDeepEqual(t, mapAt(t, peer, "podSelector"), map[string]interface{}{})
		if got := mapAt(t, peer, "namespaceSelector", "matchLabels")["kubernetes.io/metadata.name"]; got != stringValue(namespace.Name()) {
			t.Fatalf("namespace selector = %#v, want %q", got, stringValue(namespace.Name()))
		}
		if got := podPolicyMetadataNamespace(t, podPoliciesForDirection(t, chart, "ingress")[0]); got != stringValue(namespace.Name()) {
			t.Fatalf("opposite policy namespace = %q, want %q", got, stringValue(namespace.Name()))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L978
	t.Run("can allow to namespaces selected by name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		namespace := plus.Namespaces_Select(chart, jsii.String("Namespaces"), &plus.NamespacesSelectOptions{Names: &[]*string{jsii.String("n1")}})
		pod.Connections().AllowTo(namespace, nil)
		if len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 2 {
			t.Fatalf("NetworkPolicy count = %d, want 2", len(podManifestsOfKind(t, chart, "NetworkPolicy")))
		}
		rule := podPolicyRule(t, podPoliciesForDirection(t, chart, "egress")[0], "egress")
		if got := mapAt(t, podRulePeers(t, rule, "egress")[0], "namespaceSelector", "matchLabels")["kubernetes.io/metadata.name"]; got != "n1" {
			t.Fatalf("namespace selector = %#v, want n1", got)
		}
		if got := podPolicyMetadataNamespace(t, podPoliciesForDirection(t, chart, "ingress")[0]); got != "n1" {
			t.Fatalf("opposite policy namespace = %q, want n1", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L992
	t.Run("cannot allow to namespaces selected by labels", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		namespace := plus.Namespaces_Select(chart, jsii.String("Namespaces"), &plus.NamespacesSelectOptions{Labels: &map[string]*string{"type": jsii.String("selected")}})
		requirePanicContains(t, "Unable to create an Ingress policy for peer 'test/Namespaces' (pod=test-pod-c890e1b8). Peer must specify namespaces only by name", func() {
			pod.Connections().AllowTo(namespace, nil)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1005
	t.Run("can allow to peer across namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "n1", 0)
		pod2 := podConnectionPod(chart, "Pod2", "n2", 0)
		pod1.Connections().AllowTo(pod2, nil)
		egress := podPoliciesForDirection(t, chart, "egress")
		ingress := podPoliciesForDirection(t, chart, "ingress")
		if len(egress) != 1 || len(ingress) != 1 {
			t.Fatalf("egress/ingress counts = %d/%d, want 1/1", len(egress), len(ingress))
		}
		if got := podPolicyMetadataNamespace(t, egress[0]); got != "n1" {
			t.Fatalf("egress policy namespace = %q, want n1", got)
		}
		if got := podPolicyMetadataNamespace(t, ingress[0]); got != "n2" {
			t.Fatalf("ingress policy namespace = %q, want n2", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1023
	t.Run("can allow to multiple peers", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 0)
		pod3 := podConnectionPod(chart, "Pod3", "", 0)
		pod1.Connections().AllowTo(pod2, nil)
		pod1.Connections().AllowTo(pod3, nil)
		if len(podPoliciesForDirection(t, chart, "egress")) != 2 || len(podPoliciesForDirection(t, chart, "ingress")) != 2 {
			t.Fatalf("paired policies = %#v", podPolicySpecs(t, chart))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1044
	t.Run("cannot allow to the same peer twice", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 0)
		pod1.Connections().AllowTo(pod2, nil)
		requirePanicContains(t, "There is already a Construct with name", func() { pod1.Connections().AllowTo(pod2, nil) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1060
	t.Run("allow to create an ingress policy in source namespace when peer doesnt define namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "n1", 0)
		redis := plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{Labels: &map[string]*string{"role": jsii.String("redis")}})
		pod.Connections().AllowTo(redis, nil)
		ingress := podPoliciesForDirection(t, chart, "ingress")
		if len(ingress) != 1 || podPolicyMetadataNamespace(t, ingress[0]) != "n1" {
			t.Fatalf("opposite ingress policies = %#v, want one in n1", ingress)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1079
	t.Run("allow to with peer isolation creates only ingress policy on peer", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 0)
		pod1.Connections().AllowTo(pod2, &plus.PodConnectionsAllowToOptions{Isolation: plus.PodConnectionsIsolation_PEER})
		if len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 1 || len(podPoliciesForDirection(t, chart, "ingress")) != 1 {
			t.Fatalf("policies = %#v, want one ingress policy", podPolicySpecs(t, chart))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1096
	t.Run("allow to with pod isolation creates only egress policy on pod", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 0)
		pod1.Connections().AllowTo(pod2, &plus.PodConnectionsAllowToOptions{Isolation: plus.PodConnectionsIsolation_POD})
		if len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 1 || len(podPoliciesForDirection(t, chart, "egress")) != 1 {
			t.Fatalf("policies = %#v, want one egress policy", podPolicySpecs(t, chart))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1113
	t.Run("allow to defaults to peer container ports", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 6739)
		pod1.Connections().AllowTo(pod2, nil)
		want := []interface{}{map[string]interface{}{"port": float64(6739), "protocol": "TCP"}}
		requireDeepEqual(t, podPolicyRule(t, podPoliciesForDirection(t, chart, "egress")[0], "egress")["ports"], want)
		requireDeepEqual(t, podPolicyRule(t, podPoliciesForDirection(t, chart, "ingress")[0], "ingress")["ports"], want)
	})
}

func TestPodConnectionsAllowFrom(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1130
	t.Run("can allow from ip block", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		pod.Connections().AllowFrom(plus.NetworkPolicyIpBlock_AnyIpv4(chart, jsii.String("AnyIpv4")), nil)
		policies := podPoliciesForDirection(t, chart, "ingress")
		if len(policies) != 1 || len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 1 {
			t.Fatalf("ingress/total policy counts = %d/%d, want 1/1", len(policies), len(podManifestsOfKind(t, chart, "NetworkPolicy")))
		}
		rule := podPolicyRule(t, policies[0], "ingress")
		requireDeepEqual(t, podRulePeers(t, rule, "ingress"), []interface{}{map[string]interface{}{"ipBlock": map[string]interface{}{"cidr": "0.0.0.0/0"}}})
		requireDeepEqual(t, rule["ports"], []interface{}{})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1142
	t.Run("can allow from managed pod", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 0)
		pod1.Connections().AllowFrom(pod2, nil)
		ingress := podPoliciesForDirection(t, chart, "ingress")
		egress := podPoliciesForDirection(t, chart, "egress")
		if len(ingress) != 1 || len(egress) != 1 {
			t.Fatalf("ingress/egress policy counts = %d/%d, want 1/1", len(ingress), len(egress))
		}
		ingressRule := podPolicyRule(t, ingress[0], "ingress")
		egressRule := podPolicyRule(t, egress[0], "egress")
		requireDeepEqual(t, mapAt(t, podRulePeers(t, ingressRule, "ingress")[0], "podSelector", "matchLabels"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-Pod2-c82dc44e"})
		requireDeepEqual(t, mapAt(t, podRulePeers(t, egressRule, "egress")[0], "podSelector", "matchLabels"), map[string]interface{}{"cdk8s.io/metadata.addr": "test-Pod1-c8591188"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1158
	t.Run("can allow from managed workload resource", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{Containers: podContainers("pod")})
		pod.Connections().AllowFrom(deployment, nil)
		if len(podPoliciesForDirection(t, chart, "ingress")) != 1 || len(podPoliciesForDirection(t, chart, "egress")) != 1 {
			t.Fatalf("managed workload did not create paired policies: %#v", podPolicySpecs(t, chart))
		}
		rule := podPolicyRule(t, podPoliciesForDirection(t, chart, "ingress")[0], "ingress")
		selector := mapAt(t, podRulePeers(t, rule, "ingress")[0], "podSelector", "matchLabels")
		if selector["cdk8s.io/metadata.addr"] == nil {
			t.Fatalf("deployment selector = %#v", selector)
		}
		if len(podManifestsOfKind(t, chart, "Deployment")) != 1 {
			t.Fatal("managed Deployment was not synthesized")
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1174
	t.Run("can allow from pods selected without namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		selected := plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{Labels: &map[string]*string{"type": jsii.String("selected")}})
		pod.Connections().AllowFrom(selected, nil)
		if len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 2 {
			t.Fatalf("NetworkPolicy count = %d, want 2", len(podManifestsOfKind(t, chart, "NetworkPolicy")))
		}
		ingressRule := podPolicyRule(t, podPoliciesForDirection(t, chart, "ingress")[0], "ingress")
		requireDeepEqual(t, mapAt(t, podRulePeers(t, ingressRule, "ingress")[0], "podSelector", "matchLabels"), map[string]interface{}{"type": "selected"})
		opposite := mapAt(t, podPoliciesForDirection(t, chart, "egress")[0], "spec", "podSelector", "matchLabels")
		requireDeepEqual(t, opposite, map[string]interface{}{"type": "selected"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1188
	t.Run("can allow from pods selected with namespaces selected by names", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		namespaces := plus.Namespaces_Select(chart, jsii.String("Namespaces"), &plus.NamespacesSelectOptions{Names: &[]*string{jsii.String("selected1"), jsii.String("selected2")}})
		selected := plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{Labels: &map[string]*string{"type": jsii.String("selected")}, Namespaces: namespaces})
		pod.Connections().AllowFrom(selected, nil)
		policies := podManifestsOfKind(t, chart, "NetworkPolicy")
		if len(policies) != 3 {
			t.Fatalf("NetworkPolicy count = %d, want 3", len(policies))
		}
		ingressRule := podPolicyRule(t, podPoliciesForDirection(t, chart, "ingress")[0], "ingress")
		peers := podRulePeers(t, ingressRule, "ingress")
		if len(peers) != 2 {
			t.Fatalf("ingress peer count = %d, want 2", len(peers))
		}
		names := make([]string, 0, 2)
		for _, peer := range peers {
			names = append(names, mapAt(t, peer, "namespaceSelector", "matchLabels")["kubernetes.io/metadata.name"].(string))
		}
		sort.Strings(names)
		requireDeepEqual(t, names, []string{"selected1", "selected2"})
		egress := podPoliciesForDirection(t, chart, "egress")
		gotNamespaces := []string{podPolicyMetadataNamespace(t, egress[0]), podPolicyMetadataNamespace(t, egress[1])}
		sort.Strings(gotNamespaces)
		requireDeepEqual(t, gotNamespaces, []string{"selected1", "selected2"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1205
	t.Run("cannot allow from pods selected with namespaces selected by labels", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		namespaces := plus.Namespaces_Select(chart, jsii.String("Namespaces"), &plus.NamespacesSelectOptions{Labels: &map[string]*string{"type": jsii.String("selected")}})
		selected := plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{Labels: &map[string]*string{"type": jsii.String("selected")}, Namespaces: namespaces})
		requirePanicContains(t, "Unable to create an Egress policy for peer 'test/Pods' (pod=test-pod-c890e1b8). Peer must specify namespaces only by name", func() {
			pod.Connections().AllowFrom(selected, nil)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1221
	t.Run("cannot allow from pods selected in all namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		selected := plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{
			Labels:     &map[string]*string{"type": jsii.String("selected")},
			Namespaces: plus.Namespaces_All(chart, jsii.String("AllNamespaces")),
		})
		requirePanicContains(t, "Unable to create an Egress policy for peer 'test/Pods' (pod=test-pod-c890e1b8). Peer must specify namespace names", func() {
			pod.Connections().AllowFrom(selected, nil)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1237
	t.Run("can allow from all pods", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		pod.Connections().AllowFrom(plus.Pods_All(chart, jsii.String("AllPods"), nil), nil)
		if len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 2 {
			t.Fatalf("NetworkPolicy count = %d, want 2", len(podManifestsOfKind(t, chart, "NetworkPolicy")))
		}
		rule := podPolicyRule(t, podPoliciesForDirection(t, chart, "ingress")[0], "ingress")
		requireDeepEqual(t, mapAt(t, podRulePeers(t, rule, "ingress")[0], "podSelector"), map[string]interface{}{})
		requireDeepEqual(t, mapAt(t, podPoliciesForDirection(t, chart, "egress")[0], "spec", "podSelector"), map[string]interface{}{})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1251
	t.Run("can allow from managed namespace", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		namespace := plus.NewNamespace(chart, jsii.String("Namespace"), nil)
		pod.Connections().AllowFrom(namespace, nil)
		if len(podManifestsOfKind(t, chart, "Namespace")) != 1 || len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 2 {
			t.Fatalf("Namespace/NetworkPolicy counts = %d/%d, want 1/2", len(podManifestsOfKind(t, chart, "Namespace")), len(podManifestsOfKind(t, chart, "NetworkPolicy")))
		}
		rule := podPolicyRule(t, podPoliciesForDirection(t, chart, "ingress")[0], "ingress")
		peer := podRulePeers(t, rule, "ingress")[0]
		if got := mapAt(t, peer, "namespaceSelector", "matchLabels")["kubernetes.io/metadata.name"]; got != stringValue(namespace.Name()) {
			t.Fatalf("namespace selector = %#v, want %q", got, stringValue(namespace.Name()))
		}
		if got := podPolicyMetadataNamespace(t, podPoliciesForDirection(t, chart, "egress")[0]); got != stringValue(namespace.Name()) {
			t.Fatalf("opposite policy namespace = %q, want %q", got, stringValue(namespace.Name()))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1265
	t.Run("can allow from namespaces selected by name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		namespace := plus.Namespaces_Select(chart, jsii.String("Namespaces"), &plus.NamespacesSelectOptions{Names: &[]*string{jsii.String("n1")}})
		pod.Connections().AllowFrom(namespace, nil)
		if len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 2 {
			t.Fatalf("NetworkPolicy count = %d, want 2", len(podManifestsOfKind(t, chart, "NetworkPolicy")))
		}
		rule := podPolicyRule(t, podPoliciesForDirection(t, chart, "ingress")[0], "ingress")
		if got := mapAt(t, podRulePeers(t, rule, "ingress")[0], "namespaceSelector", "matchLabels")["kubernetes.io/metadata.name"]; got != "n1" {
			t.Fatalf("namespace selector = %#v, want n1", got)
		}
		if got := podPolicyMetadataNamespace(t, podPoliciesForDirection(t, chart, "egress")[0]); got != "n1" {
			t.Fatalf("opposite policy namespace = %q, want n1", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1279
	t.Run("cannot allow from namespaces selected by labels", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "", 0)
		namespace := plus.Namespaces_Select(chart, jsii.String("Namespaces"), &plus.NamespacesSelectOptions{Labels: &map[string]*string{"type": jsii.String("selected")}})
		requirePanicContains(t, "Unable to create an Egress policy for peer 'test/Namespaces' (pod=test-pod-c890e1b8). Peer must specify namespaces only by name", func() {
			pod.Connections().AllowFrom(namespace, nil)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1292
	t.Run("can allow from peer across namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "n1", 0)
		pod2 := podConnectionPod(chart, "Pod2", "n2", 0)
		pod1.Connections().AllowFrom(pod2, nil)
		ingress := podPoliciesForDirection(t, chart, "ingress")
		egress := podPoliciesForDirection(t, chart, "egress")
		if len(ingress) != 1 || len(egress) != 1 {
			t.Fatalf("ingress/egress counts = %d/%d, want 1/1", len(ingress), len(egress))
		}
		if got := podPolicyMetadataNamespace(t, ingress[0]); got != "n1" {
			t.Fatalf("ingress policy namespace = %q, want n1", got)
		}
		if got := podPolicyMetadataNamespace(t, egress[0]); got != "n2" {
			t.Fatalf("egress policy namespace = %q, want n2", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1310
	t.Run("can allow from multiple peers", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 0)
		pod3 := podConnectionPod(chart, "Pod3", "", 0)
		pod1.Connections().AllowFrom(pod2, nil)
		pod1.Connections().AllowFrom(pod3, nil)
		if len(podPoliciesForDirection(t, chart, "ingress")) != 2 || len(podPoliciesForDirection(t, chart, "egress")) != 2 {
			t.Fatalf("paired policies = %#v", podPolicySpecs(t, chart))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1331
	t.Run("cannot allow from the same peer twice", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 0)
		pod1.Connections().AllowFrom(pod2, nil)
		requirePanicContains(t, "There is already a Construct with name", func() { pod1.Connections().AllowFrom(pod2, nil) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1347
	t.Run("allow from create an ingress policy in source namespace when peer doesnt define namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := podConnectionPod(chart, "Pod", "n1", 0)
		redis := plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{Labels: &map[string]*string{"role": jsii.String("redis")}})
		pod.Connections().AllowFrom(redis, nil)
		egress := podPoliciesForDirection(t, chart, "egress")
		if len(egress) != 1 || podPolicyMetadataNamespace(t, egress[0]) != "n1" {
			t.Fatalf("opposite egress policies = %#v, want one in n1", egress)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1366
	t.Run("allow from with peer isolation creates only ingress policy on peer", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 0)
		pod1.Connections().AllowFrom(pod2, &plus.PodConnectionsAllowFromOptions{Isolation: plus.PodConnectionsIsolation_PEER})
		if len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 1 || len(podPoliciesForDirection(t, chart, "egress")) != 1 {
			t.Fatalf("policies = %#v, want one peer egress policy", podPolicySpecs(t, chart))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1383
	t.Run("allow from with pod isolation creates only egress policy on pod", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 0)
		pod1.Connections().AllowFrom(pod2, &plus.PodConnectionsAllowFromOptions{Isolation: plus.PodConnectionsIsolation_POD})
		if len(podManifestsOfKind(t, chart, "NetworkPolicy")) != 1 || len(podPoliciesForDirection(t, chart, "ingress")) != 1 {
			t.Fatalf("policies = %#v, want one pod ingress policy", podPolicySpecs(t, chart))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1400
	t.Run("allow from defaults to peer container ports", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod1 := podConnectionPod(chart, "Pod1", "", 0)
		pod2 := podConnectionPod(chart, "Pod2", "", 6739)
		pod1.Connections().AllowFrom(pod2, nil)
		requireDeepEqual(t, podPolicyRule(t, podPoliciesForDirection(t, chart, "ingress")[0], "ingress")["ports"], []interface{}{})
		requireDeepEqual(t, podPolicyRule(t, podPoliciesForDirection(t, chart, "egress")[0], "egress")["ports"], []interface{}{})
	})
}

func podManifestsOfKind(t *testing.T, chart cdk8s.Chart, kind string) []map[string]interface{} {
	t.Helper()
	result := make([]map[string]interface{}, 0)
	for _, candidate := range synth(t, chart) {
		manifest, ok := candidate.(map[string]interface{})
		if ok && manifest["kind"] == kind {
			result = append(result, manifest)
		}
	}
	return result
}

func podRequireReadGrant(t *testing.T, chart cdk8s.Chart, resourceType, resourceName string, subject map[string]interface{}) {
	t.Helper()
	roles := podManifestsOfKind(t, chart, "Role")
	bindings := podManifestsOfKind(t, chart, "RoleBinding")
	if len(roles) != 1 || len(bindings) != 1 {
		t.Fatalf("Role/RoleBinding counts = %d/%d, want 1/1", len(roles), len(bindings))
	}
	rules := sliceAt(t, roles[0], "rules")
	if len(rules) != 1 {
		t.Fatalf("role rule count = %d, want 1", len(rules))
	}
	requireDeepEqual(t, rules[0], map[string]interface{}{
		"apiGroups":     []interface{}{""},
		"resourceNames": []interface{}{resourceName},
		"resources":     []interface{}{resourceType},
		"verbs":         []interface{}{"get", "list", "watch"},
	})
	subjects := sliceAt(t, bindings[0], "subjects")
	requireDeepEqual(t, subjects, []interface{}{subject})
	roleName := mapAt(t, roles[0], "metadata")["name"]
	roleRef := mapAt(t, bindings[0], "roleRef")
	if roleRef["apiGroup"] != "rbac.authorization.k8s.io" || roleRef["kind"] != "Role" || roleRef["name"] != roleName {
		t.Fatalf("roleRef = %#v, want generated Role %#v", roleRef, roleName)
	}
}

func TestPodPermissions(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1422
	t.Run("can grant read permissions to a user", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		pod.Permissions().GrantRead(plus.User_FromName(chart, jsii.String("User"), jsii.String("bob")))
		podRequireReadGrant(t, chart, "pods", stringValue(pod.Name()), map[string]interface{}{
			"apiGroup": "rbac.authorization.k8s.io",
			"kind":     "User",
			"name":     "bob",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1434
	t.Run("can grant read permissions to a group", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		pod.Permissions().GrantRead(plus.Group_FromName(chart, jsii.String("Group"), jsii.String("manager")))
		podRequireReadGrant(t, chart, "pods", stringValue(pod.Name()), map[string]interface{}{
			"apiGroup": "rbac.authorization.k8s.io",
			"kind":     "Group",
			"name":     "manager",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1446
	t.Run("can grant read permissions to a service account", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		account := plus.NewServiceAccount(chart, jsii.String("ServiceAccount"), nil)
		pod.Permissions().GrantRead(account)
		podRequireReadGrant(t, chart, "pods", stringValue(pod.Name()), map[string]interface{}{
			"apiGroup": "",
			"kind":     "ServiceAccount",
			"name":     stringValue(account.Name()),
		})
		if got := len(podManifestsOfKind(t, chart, "ServiceAccount")); got != 1 {
			t.Fatalf("ServiceAccount manifest count = %d, want 1", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1460
	t.Run("can grant read permissions to another pod", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod1"), &plus.PodProps{Containers: podContainers("image")})
		scraper := plus.NewPod(chart, jsii.String("Pod2"), &plus.PodProps{Containers: podContainers("scraper"), AutomountServiceAccountToken: jsii.Bool(true)})
		pod.Permissions().GrantRead(scraper)
		podRequireReadGrant(t, chart, "pods", stringValue(pod.Name()), map[string]interface{}{
			"apiGroup": "",
			"kind":     "ServiceAccount",
			"name":     "default",
		})
		if podSpecByImage(t, chart, "scraper")["automountServiceAccountToken"] != true {
			t.Fatal("scraper does not automount its service account token")
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1476
	t.Run("can grant read permissions to workload", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		scraper := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{Containers: podContainers("scraper"), AutomountServiceAccountToken: jsii.Bool(true)})
		pod.Permissions().GrantRead(scraper)
		podRequireReadGrant(t, chart, "pods", stringValue(pod.Name()), map[string]interface{}{
			"apiGroup": "",
			"kind":     "ServiceAccount",
			"name":     "default",
		})
		deployments := podManifestsOfKind(t, chart, "Deployment")
		if len(deployments) != 1 || mapAt(t, deployments[0], "spec", "template", "spec")["automountServiceAccountToken"] != true {
			t.Fatalf("deployment manifest = %#v", deployments)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1492
	t.Run("can grant read permissions twice with different subjects", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		ports := []*plus.ServicePort{{Port: jsii.Number(8080)}}
		service := plus.NewService(chart, jsii.String("Service"), &plus.ServiceProps{Ports: &ports})
		service.Permissions().GrantRead(plus.Group_FromName(chart, jsii.String("Manager"), jsii.String("manager")))
		service.Permissions().GrantRead(plus.Group_FromName(chart, jsii.String("Support"), jsii.String("support")))
		roles := podManifestsOfKind(t, chart, "Role")
		bindings := podManifestsOfKind(t, chart, "RoleBinding")
		if len(roles) != 2 || len(bindings) != 2 {
			t.Fatalf("Role/RoleBinding counts = %d/%d, want 2/2", len(roles), len(bindings))
		}
		names := make([]string, 0, 2)
		for _, binding := range bindings {
			subjects := sliceAt(t, binding, "subjects")
			if len(subjects) != 1 {
				t.Fatalf("subjects = %#v", subjects)
			}
			names = append(names, subjects[0].(map[string]interface{})["name"].(string))
		}
		sort.Strings(names)
		requireDeepEqual(t, names, []string{"manager", "support"})
		for _, role := range roles {
			rule := sliceAt(t, role, "rules")[0].(map[string]interface{})
			requireDeepEqual(t, rule["resources"], []interface{}{"services"})
			requireDeepEqual(t, rule["resourceNames"], []interface{}{stringValue(service.Name())})
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1505
	t.Run("cannot grant permissions twice with same subject", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		ports := []*plus.ServicePort{{Port: jsii.Number(8080)}}
		service := plus.NewService(chart, jsii.String("Service"), &plus.ServiceProps{Ports: &ports})
		manager := plus.Group_FromName(chart, jsii.String("Manager"), jsii.String("manager"))
		service.Permissions().GrantRead(manager)
		requirePanicContains(t, "There is already a Construct with name", func() { service.Permissions().GrantRead(manager) })
	})
}

func TestPodFinalOptions(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1521
	t.Run("can pass an existing secret as the docker auth", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.Secret_FromSecretName(chart, jsii.String("RegistrySecret"), jsii.String("scw-registry-secret"))
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), DockerRegistryAuth: secret})
		requireDeepEqual(t, sliceAt(t, podSpec(t, chart), "imagePullSecrets"), []interface{}{map[string]interface{}{"name": "scw-registry-secret"}})
		if got := len(synth(t, chart)); got != 1 {
			t.Fatalf("manifest count = %d, want referenced secret to synthesize no manifest", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1536
	t.Run("can add hostNetwork to pod", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), HostNetwork: jsii.Bool(true)})
		if got := podSpec(t, chart)["hostNetwork"]; got != true {
			t.Fatalf("hostNetwork = %#v, want true", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1551
	t.Run("pod hostNetwork is not added by default", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		if got := podSpec(t, chart)["hostNetwork"]; got != false {
			t.Fatalf("hostNetwork = %#v, want false", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1565
	t.Run("default termination grace period", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image")})
		if got := podSpec(t, chart)["terminationGracePeriodSeconds"]; got != float64(30) {
			t.Fatalf("terminationGracePeriodSeconds = %#v, want 30", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1576
	t.Run("custom termination grace period", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), TerminationGracePeriod: cdk8s.Duration_Seconds(jsii.Number(60))})
		if got := podSpec(t, chart)["terminationGracePeriodSeconds"]; got != float64(60) {
			t.Fatalf("terminationGracePeriodSeconds = %#v, want 60", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1588
	t.Run("custom termination grace period - minutes", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), TerminationGracePeriod: cdk8s.Duration_Minutes(jsii.Number(2))})
		if got := podSpec(t, chart)["terminationGracePeriodSeconds"]; got != float64(120) {
			t.Fatalf("terminationGracePeriodSeconds = %#v, want 120", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1600
	t.Run("Containers should not specify restartPolicy field", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("image"), RestartPolicy: plus.ContainerRestartPolicy_ALWAYS}}})
		requirePanicContains(t, "restartPolicy", func() { cdk8s.Testing_Synth(chart) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pod.test.ts#L1608
	t.Run("enableServiceLinks can be disabled", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: podContainers("image"), EnableServiceLinks: jsii.Bool(false)})
		if got := podSpec(t, chart)["enableServiceLinks"]; got != false {
			t.Fatalf("enableServiceLinks = %#v, want false", got)
		}
	})
}
