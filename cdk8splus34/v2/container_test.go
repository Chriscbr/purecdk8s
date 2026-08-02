package cdk8splus34_test

import (
	"os"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L8
func TestEnvValueCanBeCreatedFromValue(t *testing.T) {
	actual := plus.EnvValue_FromValue(jsii.String("value"))
	if got := normalizeValue(t, actual.Value()); got != "value" {
		t.Fatalf("Value() = %#v, want value", got)
	}
	if actual.ValueFrom() != nil {
		t.Fatalf("ValueFrom() = %#v, want nil", actual.ValueFrom())
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L17
func TestEnvValueCanBeCreatedFromConfigMapName(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	actual := plus.EnvValue_FromConfigMap(plus.ConfigMap_FromConfigMapName(chart, jsii.String("ConfigMap"), jsii.String("ConfigMap")), jsii.String("key"), nil)
	if actual.Value() != nil {
		t.Fatalf("Value() = %#v, want nil", actual.Value())
	}
	requireDeepEqual(t, normalizeValue(t, actual.ValueFrom()), map[string]interface{}{
		"configMapKeyRef": map[string]interface{}{"key": "key", "name": "ConfigMap"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L32
func TestEnvValueCanBeCreatedFromSecretValue(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	actual := plus.EnvValue_FromSecretValue(&plus.SecretValue{
		Secret: plus.Secret_FromSecretName(chart, jsii.String("Secret"), jsii.String("Secret")), Key: jsii.String("my-key"),
	}, nil)
	if actual.Value() != nil {
		t.Fatalf("Value() = %#v, want nil", actual.Value())
	}
	requireDeepEqual(t, normalizeValue(t, actual.ValueFrom()), map[string]interface{}{
		"secretKeyRef": map[string]interface{}{"key": "my-key", "name": "Secret"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L50
func TestEnvValueCanBeCreatedFromImportedSecretEnvValue(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	actual := plus.Secret_FromSecretName(chart, jsii.String("Secret"), jsii.String("Secret")).EnvValue(
		jsii.String("my-key"), &plus.EnvValueFromSecretOptions{Optional: jsii.Bool(false)},
	)
	if actual.Value() != nil {
		t.Fatalf("Value() = %#v, want nil", actual.Value())
	}
	requireDeepEqual(t, normalizeValue(t, actual.ValueFrom()), map[string]interface{}{
		"secretKeyRef": map[string]interface{}{"key": "my-key", "name": "Secret", "optional": false},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L65
func TestEnvValueCanBeCreatedFromNewSecretEnvValue(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	secret := plus.NewSecret(chart, jsii.String("Secret"), &plus.SecretProps{StringData: &map[string]*string{"my-key": jsii.String("my-value")}})
	actual := secret.EnvValue(jsii.String("my-key"), &plus.EnvValueFromSecretOptions{Optional: jsii.Bool(true)})
	if actual.Value() != nil {
		t.Fatalf("Value() = %#v, want nil", actual.Value())
	}
	requireDeepEqual(t, normalizeValue(t, actual.ValueFrom()), map[string]interface{}{
		"secretKeyRef": map[string]interface{}{"key": "my-key", "name": "test-secret-c837fa76", "optional": true},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L80
func TestEnvValueRejectsMissingRequiredProcessEnv(t *testing.T) {
	key := "cdk8s-plus.tests.container.env.fromProcess"
	original, found := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if found {
			_ = os.Setenv(key, original)
		} else {
			_ = os.Unsetenv(key)
		}
	})
	requirePanicContains(t, "Missing "+key+" env variable", func() {
		plus.EnvValue_FromProcess(jsii.String(key), &plus.EnvValueFromProcessOptions{Required: jsii.Bool(true)})
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L87
func TestEnvValueCanBeCreatedFromMissingOptionalProcessEnv(t *testing.T) {
	key := "cdk8s-plus.tests.container.env.fromProcess"
	original, found := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if found {
			_ = os.Setenv(key, original)
		}
	})
	actual := plus.EnvValue_FromProcess(jsii.String(key), nil)
	if actual.Value() != nil || actual.ValueFrom() != nil {
		t.Fatalf("optional missing env = (%#v, %#v), want nil values", actual.Value(), actual.ValueFrom())
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L97
func TestEnvValueCanBeCreatedFromExistingProcessEnv(t *testing.T) {
	key := "cdk8s-plus.tests.container.env.fromProcess"
	t.Setenv(key, "value")
	actual := plus.EnvValue_FromProcess(jsii.String(key), nil)
	if got := normalizeValue(t, actual.Value()); got != "value" {
		t.Fatalf("Value() = %#v, want value", got)
	}
	if actual.ValueFrom() != nil {
		t.Fatalf("ValueFrom() = %#v, want nil", actual.ValueFrom())
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L113
func TestEnvValueCanBeCreatedFromFieldRef(t *testing.T) {
	actual := plus.EnvValue_FromFieldRef(plus.EnvFieldPaths_POD_NAME, nil)
	if actual.Value() != nil {
		t.Fatalf("Value() = %#v, want nil", actual.Value())
	}
	requireDeepEqual(t, normalizeValue(t, actual.ValueFrom()), map[string]interface{}{
		"fieldRef": map[string]interface{}{"fieldPath": "metadata.name"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L124
func TestEnvValueCanBeCreatedFromFieldRefWithKey(t *testing.T) {
	actual := plus.EnvValue_FromFieldRef(plus.EnvFieldPaths_POD_LABEL, &plus.EnvValueFromFieldRefOptions{Key: jsii.String("someLabel")})
	if actual.Value() != nil {
		t.Fatalf("Value() = %#v, want nil", actual.Value())
	}
	requireDeepEqual(t, normalizeValue(t, actual.ValueFrom()), map[string]interface{}{
		"fieldRef": map[string]interface{}{"fieldPath": "metadata.labels['someLabel']"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L135
func TestEnvValueCannotBeCreatedFromFieldRefWithoutKey(t *testing.T) {
	requirePanicContains(t, "metadata.labels requires a key", func() {
		plus.EnvValue_FromFieldRef(plus.EnvFieldPaths_POD_LABEL, nil)
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L139
func TestEnvValueCanBeCreatedFromResourceFieldRef(t *testing.T) {
	actual := plus.EnvValue_FromResource(plus.ResourceFieldPaths_CPU_LIMIT, nil)
	if actual.Value() != nil {
		t.Fatalf("Value() = %#v, want nil", actual.Value())
	}
	requireDeepEqual(t, normalizeValue(t, actual.ValueFrom()), map[string]interface{}{
		"resourceFieldRef": map[string]interface{}{"resource": "limits.cpu"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L150
func TestEnvValueCanBeCreatedFromResourceFieldRefWithDivisor(t *testing.T) {
	actual := plus.EnvValue_FromResource(plus.ResourceFieldPaths_MEMORY_LIMIT, &plus.EnvValueFromResourceOptions{Divisor: jsii.String("1Mi")})
	if actual.Value() != nil {
		t.Fatalf("Value() = %#v, want nil", actual.Value())
	}
	requireDeepEqual(t, normalizeValue(t, actual.ValueFrom()), map[string]interface{}{
		"resourceFieldRef": map[string]interface{}{"resource": "limits.memory", "divisor": "1Mi"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L162
func TestEnvValueCanBeCreatedFromResourceFieldRefWithContainer(t *testing.T) {
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Name: jsii.String("name"), ImagePullPolicy: plus.ImagePullPolicy_NEVER})
	actual := plus.EnvValue_FromResource(plus.ResourceFieldPaths_CPU_LIMIT, &plus.EnvValueFromResourceOptions{Container: container})
	if actual.Value() != nil {
		t.Fatalf("Value() = %#v, want nil", actual.Value())
	}
	requireDeepEqual(t, normalizeValue(t, actual.ValueFrom()), map[string]interface{}{
		"resourceFieldRef": map[string]interface{}{"resource": "limits.cpu", "containerName": "name"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L183
func TestContainerRejectsIdenticalPortsAndProtocolsAtInstantiation(t *testing.T) {
	requirePanicContains(t, "Port with number 8080 and protocol TCP already exists", func() {
		plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Ports: &[]*plus.ContainerPort{
			{Number: jsii.Number(8080), Protocol: plus.Protocol_TCP},
			{Number: jsii.Number(8080), Protocol: plus.Protocol_TCP},
		}})
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L201
func TestContainerAllowsIdenticalPortsWithDifferentProtocolsAtInstantiation(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Ports: &[]*plus.ContainerPort{
		{Number: jsii.Number(8080), Protocol: plus.Protocol_TCP},
		{Number: jsii.Number(8080), Protocol: plus.Protocol_UDP},
	}})
	manifest := attachedContainerManifest(t, chart, container)
	requireDeepEqual(t, manifest["ports"], []interface{}{
		map[string]interface{}{"containerPort": float64(8080), "protocol": "TCP"},
		map[string]interface{}{"containerPort": float64(8080), "protocol": "UDP"},
	})
	ports := *container.Ports()
	if len(ports) != 2 || numberValue(ports[0].Number) != 8080 || ports[0].Protocol != plus.Protocol_TCP || numberValue(ports[1].Number) != 8080 || ports[1].Protocol != plus.Protocol_UDP {
		t.Fatalf("Ports() = %#v, want TCP and UDP port 8080", ports)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L232
func TestContainerRejectsExistingPortNumberWithIdenticalProtocol(t *testing.T) {
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Ports: &[]*plus.ContainerPort{{Number: jsii.Number(8080)}}})
	requirePanicContains(t, "Port with number 8080 and protocol TCP already exists", func() {
		container.AddPort(&plus.ContainerPort{Number: jsii.Number(8080)})
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L245
func TestContainerAllowsExistingPortNumberWithDifferentProtocol(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Ports: &[]*plus.ContainerPort{{Number: jsii.Number(8080), Protocol: plus.Protocol_TCP}}})
	container.AddPort(&plus.ContainerPort{Number: jsii.Number(8080), Protocol: plus.Protocol_UDP})
	manifest := attachedContainerManifest(t, chart, container)
	requireDeepEqual(t, manifest["ports"], []interface{}{
		map[string]interface{}{"containerPort": float64(8080), "protocol": "TCP"},
		map[string]interface{}{"containerPort": float64(8080), "protocol": "UDP"},
	})
	ports := *container.Ports()
	if len(ports) != 2 || ports[0].Protocol != plus.Protocol_TCP || ports[1].Protocol != plus.Protocol_UDP {
		t.Fatalf("Ports() = %#v, want TCP then UDP", ports)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L275
func TestContainerRejectsExistingPortName(t *testing.T) {
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Ports: &[]*plus.ContainerPort{{Number: jsii.Number(8080), Name: jsii.String("port1")}}})
	requirePanicContains(t, "Port with name port1 already exists", func() {
		container.AddPort(&plus.ContainerPort{Number: jsii.Number(9090), Name: jsii.String("port1")})
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L289
func TestContainerCanConfigureMultiplePorts(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Ports: &[]*plus.ContainerPort{{Number: jsii.Number(8080)}}})
	container.AddPort(&plus.ContainerPort{Number: jsii.Number(9090)})
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["ports"], []interface{}{
		map[string]interface{}{"containerPort": float64(8080)}, map[string]interface{}{"containerPort": float64(9090)},
	})
	ports := *container.Ports()
	if len(ports) != 2 || numberValue(ports[0].Number) != 8080 || numberValue(ports[1].Number) != 9090 {
		t.Fatalf("Ports() = %#v, want 8080 and 9090", ports)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L304
func TestContainerPortNumberIsEquivalentToPort(t *testing.T) {
	firstChart := cdk8s.Testing_Chart()
	first := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Port: jsii.Number(9000)})
	firstManifest := attachedContainerManifest(t, firstChart, first)
	secondChart := cdk8s.Testing_Chart()
	second := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), PortNumber: jsii.Number(9000)})
	secondManifest := attachedContainerManifest(t, secondChart, second)
	requireDeepEqual(t, firstManifest, secondManifest)
	if numberValue(first.PortNumber()) != numberValue(first.Port()) {
		t.Fatalf("PortNumber() = %v, Port() = %v", numberValue(first.PortNumber()), numberValue(first.Port()))
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L321
func TestContainerInstantiationPropertiesAreAllRespected(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{
		Image: jsii.String("image"), Name: jsii.String("name"), ImagePullPolicy: plus.ImagePullPolicy_NEVER,
		WorkingDir: jsii.String("workingDir"), Port: jsii.Number(9000), Command: &[]*string{jsii.String("command")},
		EnvVariables: &map[string]plus.EnvValue{"key": plus.EnvValue_FromValue(jsii.String("value"))},
		Startup:      plus.Probe_FromTcpSocket(&plus.TcpSocketProbeOptions{Port: jsii.Number(9000)}),
	}}})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "260ecfe628479063316d94041ae3bfefe4aa5dc0bfb8228304eb6cfb91d41e4a")
	container := mapAt(t, sliceAt(t, manifests[0], "spec", "containers")[0])
	for key, want := range map[string]interface{}{"name": "name", "imagePullPolicy": "Never", "image": "image", "workingDir": "workingDir"} {
		if got := container[key]; got != want {
			t.Errorf("container.%s = %#v, want %#v", key, got, want)
		}
	}
	if got := mapAt(t, sliceAt(t, container, "ports")[0])["containerPort"]; got != float64(9000) {
		t.Errorf("containerPort = %#v, want 9000", got)
	}
	if got := container["command"].([]interface{})[0]; got != "command" {
		t.Errorf("command[0] = %#v, want command", got)
	}
	env := mapAt(t, sliceAt(t, container, "env")[0])
	if env["name"] != "key" || env["value"] != "value" {
		t.Errorf("env[0] = %#v, want key=value", env)
	}
	security := mapAt(t, container, "securityContext")
	if security["privileged"] != false || security["readOnlyRootFilesystem"] != true || security["runAsNonRoot"] != true {
		t.Errorf("securityContext = %#v, want secure defaults", security)
	}
	startup := mapAt(t, container, "startupProbe")
	if startup["failureThreshold"] != float64(3) || mapAt(t, startup, "tcpSocket")["port"] != float64(9000) {
		t.Errorf("startupProbe = %#v, want TCP port 9000 and threshold 3", startup)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L365
func TestContainerMustUseContainerProps(t *testing.T) {
	requirePanicContains(t, "props.image is required", func() { plus.NewContainer(nil) })
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L376
func TestContainerCanAddEnvironmentVariable(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image")})
	container.Env().AddVariable(jsii.String("key"), plus.EnvValue_FromValue(jsii.String("value")))
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["env"], []interface{}{map[string]interface{}{"name": "key", "value": "value"}})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L395
func TestContainerCanAddEnvironmentVariablesFromSource(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	configMap := plus.NewConfigMap(chart, jsii.String("ConfigMap"), nil)
	secret := plus.NewSecret(chart, jsii.String("Secret"), nil)
	configMapSource := plus.Env_FromConfigMap(configMap, jsii.String("pref"))
	secretSource := plus.Env_FromSecret(secret)
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), EnvFrom: &[]plus.EnvFrom{configMapSource}})
	container.Env().CopyFrom(secretSource)
	manifest := attachedContainerManifest(t, chart, container)
	requireDeepEqual(t, manifest["envFrom"], []interface{}{
		map[string]interface{}{"configMapRef": map[string]interface{}{"name": stringValue(configMap.Name())}, "prefix": "pref"},
		map[string]interface{}{"secretRef": map[string]interface{}{"name": stringValue(secret.Name())}},
	})
	sources := *container.Env().Sources()
	if len(sources) != 2 || sources[0] != configMapSource || sources[1] != secretSource {
		t.Fatalf("Env().Sources() = %#v, want configured sources", sources)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L422
func TestContainerCanAddEnvironmentVariablesFromSecret(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	secret := plus.NewSecret(chart, jsii.String("Secret"), nil)
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image")})
	container.Env().CopyFrom(plus.Env_FromSecret(secret))

	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["envFrom"], []interface{}{
		map[string]interface{}{"secretRef": map[string]interface{}{"name": stringValue(secret.Name())}},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L424
func TestContainerCanMountToVolume(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image")})
	volume := plus.Volume_FromConfigMap(chart, jsii.String("Volume"), plus.ConfigMap_FromConfigMapName(chart, jsii.String("ConfigMap"), jsii.String("ConfigMap")), nil)
	container.Mount(jsii.String("/path/to/mount"), volume, nil)
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["volumeMounts"], []interface{}{
		map[string]interface{}{"mountPath": "/path/to/mount", "name": stringValue(volume.Name())},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L443
func TestContainerCanMountToPersistentVolume(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	pod := plus.NewPod(chart, jsii.String("Pod"), nil)
	volume := plus.NewAwsElasticBlockStorePersistentVolume(chart, jsii.String("PV"), &plus.AwsElasticBlockStorePersistentVolumeProps{VolumeId: jsii.String("vol")})
	container := pod.AddContainer(&plus.ContainerProps{Image: jsii.String("image")})
	container.Mount(jsii.String("/path/to/mount"), volume, nil)
	requireSnapshotHash(t, synth(t, chart), "310c50d9959d10118320c0a775ed710861a698846c00b073c00ea65770ec85c3")
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L457
func TestContainerMountOptions(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image")})
	volume := plus.Volume_FromConfigMap(chart, jsii.String("Volume"), plus.ConfigMap_FromConfigMapName(chart, jsii.String("ConfigMap"), jsii.String("ConfigMap")), nil)
	container.Mount(jsii.String("/path/to/mount"), volume, &plus.MountOptions{Propagation: plus.MountPropagation_BIDIRECTIONAL, ReadOnly: jsii.Bool(true)})
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["volumeMounts"], []interface{}{
		map[string]interface{}{"mountPath": "/path/to/mount", "name": stringValue(volume.Name()), "mountPropagation": "Bidirectional", "readOnly": true},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L480
func TestContainerMountFromConstructor(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	volume := plus.Volume_FromEmptyDir(chart, jsii.String("Volume"), jsii.String("empty"), nil)
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), VolumeMounts: &[]*plus.VolumeMount{{
		Path: jsii.String("/foo"), Volume: volume, SubPath: jsii.String("subPath"),
	}}})
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["volumeMounts"], []interface{}{
		map[string]interface{}{"mountPath": "/foo", "name": "empty", "subPath": "subPath"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L502
func TestContainerStartupProbeHasDefaultsWhenPortProvided(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("foo"), Port: jsii.Number(8080)}}})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "6f982faacce851810960ac5155ba9c0380306b2e88752ba6343c86cfbb309b5e")
	container := mapAt(t, sliceAt(t, manifests[0], "spec", "containers")[0])
	startup := mapAt(t, container, "startupProbe")
	if startup["failureThreshold"] != float64(3) || mapAt(t, startup, "tcpSocket")["port"] != float64(8080) {
		t.Fatalf("startupProbe = %#v, want threshold 3 and port 8080", startup)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L523
func TestContainerStartupProbeAbsentWhenPortNotProvided(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("foo")}}})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "709027d1015819dbf8d2abcf42012e15a7ee5295e4e9fdcd466e05f2a69877f8")
	container := mapAt(t, sliceAt(t, manifests[0], "spec", "containers")[0])
	if _, exists := container["startupProbe"]; exists {
		t.Fatalf("startupProbe = %#v, want absent", container["startupProbe"])
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L539
func TestContainerRestartPolicyCanBeConfigured(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{
		Containers:     &[]*plus.ContainerProps{{Image: jsii.String("foo")}},
		InitContainers: &[]*plus.ContainerProps{{Image: jsii.String("bar"), RestartPolicy: plus.ContainerRestartPolicy_ALWAYS}},
	})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "9e28ff44bbf05591df05aaa748fde84c7131edd8c78d2484d5bf67994e66375c")
	container := mapAt(t, sliceAt(t, manifests[0], "spec", "initContainers")[0])
	if got := container["restartPolicy"]; got != "Always" {
		t.Fatalf("restartPolicy = %#v, want Always", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L559
func TestContainerCanDefineReadinessLivenessAndStartupProbes(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{
		Image:     jsii.String("foo"),
		Readiness: plus.Probe_FromHttpGet(jsii.String("/ping"), &plus.HttpGetProbeOptions{TimeoutSeconds: cdk8s.Duration_Minutes(jsii.Number(2)), Scheme: plus.ConnectionScheme_HTTPS}),
		Liveness:  plus.Probe_FromHttpGet(jsii.String("/live"), &plus.HttpGetProbeOptions{TimeoutSeconds: cdk8s.Duration_Minutes(jsii.Number(3))}),
		Startup:   plus.Probe_FromHttpGet(jsii.String("/startup"), &plus.HttpGetProbeOptions{TimeoutSeconds: cdk8s.Duration_Minutes(jsii.Number(4)), Scheme: plus.ConnectionScheme_HTTP}),
	})
	manifest := attachedContainerManifest(t, chart, container)
	for key, want := range map[string]map[string]interface{}{
		"readinessProbe": {"failureThreshold": float64(3), "httpGet": map[string]interface{}{"path": "/ping", "port": float64(80), "scheme": "HTTPS"}, "timeoutSeconds": float64(120)},
		"livenessProbe":  {"failureThreshold": float64(3), "httpGet": map[string]interface{}{"path": "/live", "port": float64(80), "scheme": "HTTP"}, "timeoutSeconds": float64(180)},
		"startupProbe":   {"failureThreshold": float64(3), "httpGet": map[string]interface{}{"path": "/startup", "port": float64(80), "scheme": "HTTP"}, "timeoutSeconds": float64(240)},
	} {
		requireDeepEqual(t, manifest[key], want)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L603
func TestContainerCanAddResourceLimitsAndRequests(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Resources: &plus.ContainerResources{
		Cpu:              &plus.CpuResources{Request: plus.Cpu_Millis(jsii.Number(300)), Limit: plus.Cpu_Units(jsii.Number(0.5))},
		Memory:           &plus.MemoryResources{Request: cdk8s.Size_Mebibytes(jsii.Number(256)), Limit: cdk8s.Size_Mebibytes(jsii.Number(384))},
		EphemeralStorage: &plus.EphemeralStorageResources{Request: cdk8s.Size_Mebibytes(jsii.Number(1024)), Limit: cdk8s.Size_Mebibytes(jsii.Number(2048))},
	}})
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["resources"], map[string]interface{}{
		"limits":   map[string]interface{}{"cpu": "0.5", "memory": "384Mi", "ephemeral-storage": "2Gi"},
		"requests": map[string]interface{}{"cpu": "300m", "memory": "256Mi", "ephemeral-storage": "1Gi"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L639
func TestContainerCanAddOnlyResourceRequests(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Resources: &plus.ContainerResources{
		Cpu: &plus.CpuResources{Request: plus.Cpu_Millis(jsii.Number(300))}, Memory: &plus.MemoryResources{Request: cdk8s.Size_Mebibytes(jsii.Number(128))},
	}})
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["resources"], map[string]interface{}{
		"requests": map[string]interface{}{"cpu": "300m", "memory": "128Mi"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L660
func TestContainerCanAddOnlyResourceLimits(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Resources: &plus.ContainerResources{
		Cpu: &plus.CpuResources{Limit: plus.Cpu_Millis(jsii.Number(500))}, Memory: &plus.MemoryResources{Limit: cdk8s.Size_Mebibytes(jsii.Number(1024))},
	}})
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["resources"], map[string]interface{}{
		"limits": map[string]interface{}{"cpu": "500m", "memory": "1024Mi"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L681
func TestContainerCanAddOnlyMemoryLimitsAndRequests(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Resources: &plus.ContainerResources{
		Memory: &plus.MemoryResources{Limit: cdk8s.Size_Mebibytes(jsii.Number(1024)), Request: cdk8s.Size_Mebibytes(jsii.Number(512))},
	}})
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["resources"], map[string]interface{}{
		"limits": map[string]interface{}{"memory": "1024Mi"}, "requests": map[string]interface{}{"memory": "512Mi"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L702
func TestContainerCanAddOnlyCPULimitsAndRequests(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Resources: &plus.ContainerResources{
		Cpu: &plus.CpuResources{Limit: plus.Cpu_Units(jsii.Number(1)), Request: plus.Cpu_Millis(jsii.Number(250))},
	}})
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["resources"], map[string]interface{}{
		"limits": map[string]interface{}{"cpu": "1"}, "requests": map[string]interface{}{"cpu": "250m"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L723
func TestContainerCanAddOnlyEphemeralStorageLimitsAndRequests(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Resources: &plus.ContainerResources{
		EphemeralStorage: &plus.EphemeralStorageResources{Limit: cdk8s.Size_Gibibytes(jsii.Number(2)), Request: cdk8s.Size_Gibibytes(jsii.Number(1))},
	}})
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["resources"], map[string]interface{}{
		"limits": map[string]interface{}{"ephemeral-storage": "2Gi"}, "requests": map[string]interface{}{"ephemeral-storage": "1Gi"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L744
func TestContainerCanAddOnlyEphemeralStorageLimits(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Resources: &plus.ContainerResources{
		EphemeralStorage: &plus.EphemeralStorageResources{Limit: cdk8s.Size_Gibibytes(jsii.Number(2))},
	}})
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["resources"], map[string]interface{}{
		"limits": map[string]interface{}{"ephemeral-storage": "2Gi"},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L761
func TestContainerDefaultSecurityContext(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image")})
	security := container.SecurityContext()
	if !boolValue(security.EnsureNonRoot()) || boolValue(security.Privileged()) || !boolValue(security.ReadOnlyRootFilesystem()) || security.User() != nil || security.Group() != nil || boolValue(security.AllowPrivilegeEscalation()) {
		t.Fatalf("default security context has unexpected values")
	}
	requireDeepEqual(t, attachedContainerManifest(t, chart, container)["securityContext"], map[string]interface{}{
		"privileged": false, "readOnlyRootFilesystem": true, "runAsNonRoot": true, "allowPrivilegeEscalation": false,
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L785
func TestContainerCustomSecurityContext(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), SecurityContext: &plus.ContainerSecurityContextProps{
		EnsureNonRoot: jsii.Bool(true), ReadOnlyRootFilesystem: jsii.Bool(true), Privileged: jsii.Bool(true), User: jsii.Number(1000), Group: jsii.Number(2000),
		Capabilities:   &plus.ContainerSecutiryContextCapabilities{Add: &[]plus.Capability{plus.Capability_AUDIT_CONTROL}, Drop: &[]plus.Capability{plus.Capability_BPF}},
		SeccompProfile: &plus.SeccompProfile{Type: plus.SeccompProfileType_RUNTIME_DEFAULT},
	}})
	security := container.SecurityContext()
	if !boolValue(security.EnsureNonRoot()) || !boolValue(security.Privileged()) || !boolValue(security.ReadOnlyRootFilesystem()) || numberValue(security.User()) != 1000 || numberValue(security.Group()) != 2000 {
		t.Fatalf("custom security context getters have unexpected values")
	}
	if caps := security.Capabilities(); caps == nil || len(*caps.Add) != 1 || (*caps.Add)[0] != plus.Capability_AUDIT_CONTROL || len(*caps.Drop) != 1 || (*caps.Drop)[0] != plus.Capability_BPF {
		t.Fatalf("capabilities = %#v, want AUDIT_CONTROL added and BPF dropped", caps)
	}
	manifest := mapAt(t, attachedContainerManifest(t, chart, container), "securityContext")
	if mapAt(t, manifest, "seccompProfile")["type"] != "RuntimeDefault" {
		t.Fatalf("seccompProfile = %#v, want RuntimeDefault", manifest["seccompProfile"])
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L820
func TestContainerSeccompLocalhostProfileAllowedForLocalhostType(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), SecurityContext: &plus.ContainerSecurityContextProps{
		SeccompProfile: &plus.SeccompProfile{Type: plus.SeccompProfileType_LOCALHOST, LocalhostProfile: jsii.String("localhostProfile")},
	}})
	profile := mapAt(t, attachedContainerManifest(t, chart, container), "securityContext", "seccompProfile")
	if profile["localhostProfile"] != "localhostProfile" {
		t.Fatalf("localhostProfile = %#v, want localhostProfile", profile["localhostProfile"])
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L836
func TestContainerSeccompLocalhostProfileRejectedForOtherTypes(t *testing.T) {
	requirePanicContains(t, `localhostProfile must only be set if type is "Localhost"`, func() {
		plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), SecurityContext: &plus.ContainerSecurityContextProps{
			SeccompProfile: &plus.SeccompProfile{Type: plus.SeccompProfileType_UNCONFINED, LocalhostProfile: jsii.String("localhostProfile")},
		}})
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L849
func TestContainerCanConfigurePostStartLifecycleHook(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Lifecycle: &plus.ContainerLifecycle{
		PostStart: plus.Handler_FromCommand(&[]*string{jsii.String("hello")}),
	}})
	requireDeepEqual(t, mapAt(t, attachedContainerManifest(t, chart, container), "lifecycle")["postStart"], map[string]interface{}{
		"exec": map[string]interface{}{"command": []interface{}{"hello"}},
	})
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/container.test.ts#L863
func TestContainerCanConfigurePreStopLifecycleHook(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	container := plus.NewContainer(&plus.ContainerProps{Image: jsii.String("image"), Lifecycle: &plus.ContainerLifecycle{
		PreStop: plus.Handler_FromCommand(&[]*string{jsii.String("hello")}),
	}})
	requireDeepEqual(t, mapAt(t, attachedContainerManifest(t, chart, container), "lifecycle")["preStop"], map[string]interface{}{
		"exec": map[string]interface{}{"command": []interface{}{"hello"}},
	})
}
