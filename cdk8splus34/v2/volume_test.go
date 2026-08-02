package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func volumeManifest(t *testing.T, chart cdk8s.Chart, volume plus.Volume) map[string]interface{} {
	t.Helper()
	volumes := []plus.Volume{volume}
	containers := []*plus.ContainerProps{{Image: jsii.String("image")}}
	plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: &containers, Volumes: &volumes})

	entries := sliceAt(t, manifestOfKind(t, chart, "Pod"), "spec", "volumes")
	if len(entries) != 1 {
		t.Fatalf("synthesized volume count = %d, want 1", len(entries))
	}
	result, ok := entries[0].(map[string]interface{})
	if !ok {
		t.Fatalf("synthesized volume has type %T, want map[string]interface{}", entries[0])
	}
	return result
}

func TestVolume(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L5
	t.Run("fromSecret minimal definition", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewSecret(chart, jsii.String("my-secret"), nil)
		volume := plus.Volume_FromSecret(chart, jsii.String("Secret"), secret, nil)
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"name": "secret-test-my-secret-c8a71744",
			"secret": map[string]interface{}{
				"secretName": "test-my-secret-c8a71744",
			},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L27
	t.Run("fromSecret volume name is trimmed if needed", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewSecret(chart, jsii.String("my-secret"), &plus.SecretProps{Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("veryveryveryveryveryveryveryveryveryveryveryveryveryveryverylong"),
		}})
		volume := plus.Volume_FromSecret(chart, jsii.String("Secret"), secret, nil)
		want := "secret-veryveryveryveryveryveryveryveryveryveryveryveryveryvery"
		if got := stringValue(volume.Name()); got != want {
			t.Fatalf("volume name = %q, want %q", got, want)
		}
		if got := volumeManifest(t, chart, volume)["name"]; got != want {
			t.Fatalf("synthesized volume name = %#v, want %q", got, want)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L43
	t.Run("fromSecret custom volume name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewSecret(chart, jsii.String("my-secret"), nil)
		volume := plus.Volume_FromSecret(chart, jsii.String("Secret"), secret, &plus.SecretVolumeOptions{Name: jsii.String("filesystem")})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"name": "filesystem",
			"secret": map[string]interface{}{
				"secretName": "test-my-secret-c8a71744",
			},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L58
	t.Run("fromSecret default mode", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewSecret(chart, jsii.String("my-secret"), nil)
		volume := plus.Volume_FromSecret(chart, jsii.String("Secret"), secret, &plus.SecretVolumeOptions{DefaultMode: jsii.Number(0o777)})
		secretSpec := mapAt(t, volumeManifest(t, chart, volume), "secret")
		if got := secretSpec["defaultMode"]; got != float64(0o777) {
			t.Fatalf("defaultMode = %#v, want %d", got, 0o777)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L72
	t.Run("fromSecret optional", func(t *testing.T) {
		makeSecretSpec := func(options *plus.SecretVolumeOptions) map[string]interface{} {
			chart := cdk8s.Testing_Chart()
			secret := plus.NewSecret(chart, jsii.String("my-secret"), nil)
			volume := plus.Volume_FromSecret(chart, jsii.String("Secret"), secret, options)
			return mapAt(t, volumeManifest(t, chart, volume), "secret")
		}
		unset := makeSecretSpec(nil)
		if _, exists := unset["optional"]; exists {
			t.Fatalf("optional unexpectedly synthesized: %#v", unset["optional"])
		}
		if got := makeSecretSpec(&plus.SecretVolumeOptions{Optional: jsii.Bool(true)})["optional"]; got != true {
			t.Fatalf("optional = %#v, want true", got)
		}
		if got := makeSecretSpec(&plus.SecretVolumeOptions{Optional: jsii.Bool(false)})["optional"]; got != false {
			t.Fatalf("optional = %#v, want false", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L88
	t.Run("fromSecret items", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewSecret(chart, jsii.String("my-secret"), nil)
		items := map[string]*plus.PathMapping{
			"key1": {Path: jsii.String("path/to/key1")},
			"key2": {Path: jsii.String("path/key2"), Mode: jsii.Number(0o100)},
		}
		volume := plus.Volume_FromSecret(chart, jsii.String("Secret"), secret, &plus.SecretVolumeOptions{Items: &items})
		requireDeepEqual(t, sliceAt(t, volumeManifest(t, chart, volume), "secret", "items"), []interface{}{
			map[string]interface{}{"key": "key1", "path": "path/to/key1"},
			map[string]interface{}{"key": "key2", "mode": float64(0o100), "path": "path/key2"},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L114
	t.Run("fromSecret items are sorted by key for deterministic synthesis", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewSecret(chart, jsii.String("my-secret"), nil)
		items := map[string]*plus.PathMapping{
			"key2": {Path: jsii.String("path2")},
			"key1": {Path: jsii.String("path1")},
		}
		volume := plus.Volume_FromSecret(chart, jsii.String("Secret"), secret, &plus.SecretVolumeOptions{Items: &items})
		requireDeepEqual(t, sliceAt(t, volumeManifest(t, chart, volume), "secret", "items"), []interface{}{
			map[string]interface{}{"key": "key1", "path": "path1"},
			map[string]interface{}{"key": "key2", "path": "path2"},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L143
	t.Run("fromConfigMap volume name is trimmed if needed", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), &plus.ConfigMapProps{Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("veryveryveryveryveryveryveryveryveryveryveryveryveryveryverylong"),
		}})
		volume := plus.Volume_FromConfigMap(chart, jsii.String("ConfigMap"), configMap, nil)
		want := "configmap-veryveryveryveryveryveryveryveryveryveryveryveryveryv"
		if got := stringValue(volume.Name()); got != want {
			t.Fatalf("volume name = %q, want %q", got, want)
		}
		if got := volumeManifest(t, chart, volume)["name"]; got != want {
			t.Fatalf("synthesized volume name = %#v, want %q", got, want)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L159
	t.Run("fromConfigMap minimal definition", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		volume := plus.Volume_FromConfigMap(chart, jsii.String("ConfigMap"), configMap, nil)
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"configMap": map[string]interface{}{
				"name": "test-my-config-map-c8eaefa4",
			},
			"name": "configmap-test-my-config-map-c8eaefa4",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L181
	t.Run("fromConfigMap custom volume name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		volume := plus.Volume_FromConfigMap(chart, jsii.String("ConfigMap"), configMap, &plus.ConfigMapVolumeOptions{Name: jsii.String("filesystem")})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"configMap": map[string]interface{}{
				"name": "test-my-config-map-c8eaefa4",
			},
			"name": "filesystem",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L196
	t.Run("fromConfigMap default mode", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		volume := plus.Volume_FromConfigMap(chart, jsii.String("ConfigMap"), configMap, &plus.ConfigMapVolumeOptions{DefaultMode: jsii.Number(0o777)})
		configMapSpec := mapAt(t, volumeManifest(t, chart, volume), "configMap")
		if got := configMapSpec["defaultMode"]; got != float64(0o777) {
			t.Fatalf("defaultMode = %#v, want %d", got, 0o777)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L210
	t.Run("fromConfigMap optional", func(t *testing.T) {
		makeConfigMapSpec := func(options *plus.ConfigMapVolumeOptions) map[string]interface{} {
			chart := cdk8s.Testing_Chart()
			configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
			volume := plus.Volume_FromConfigMap(chart, jsii.String("ConfigMap"), configMap, options)
			return mapAt(t, volumeManifest(t, chart, volume), "configMap")
		}
		unset := makeConfigMapSpec(nil)
		if _, exists := unset["optional"]; exists {
			t.Fatalf("optional unexpectedly synthesized: %#v", unset["optional"])
		}
		if got := makeConfigMapSpec(&plus.ConfigMapVolumeOptions{Optional: jsii.Bool(true)})["optional"]; got != true {
			t.Fatalf("optional = %#v, want true", got)
		}
		if got := makeConfigMapSpec(&plus.ConfigMapVolumeOptions{Optional: jsii.Bool(false)})["optional"]; got != false {
			t.Fatalf("optional = %#v, want false", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L226
	t.Run("fromConfigMap items", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		items := map[string]*plus.PathMapping{
			"key1": {Path: jsii.String("path/to/key1")},
			"key2": {Path: jsii.String("path/key2"), Mode: jsii.Number(0o100)},
		}
		volume := plus.Volume_FromConfigMap(chart, jsii.String("ConfigMap"), configMap, &plus.ConfigMapVolumeOptions{Items: &items})
		requireDeepEqual(t, sliceAt(t, volumeManifest(t, chart, volume), "configMap", "items"), []interface{}{
			map[string]interface{}{"key": "key1", "path": "path/to/key1"},
			map[string]interface{}{"key": "key2", "mode": float64(0o100), "path": "path/key2"},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L252
	t.Run("fromConfigMap items are sorted by key for determinstic synthesis", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		items := map[string]*plus.PathMapping{
			"key2": {Path: jsii.String("path2")},
			"key1": {Path: jsii.String("path1")},
		}
		volume := plus.Volume_FromConfigMap(chart, jsii.String("ConfigMap"), configMap, &plus.ConfigMapVolumeOptions{Items: &items})
		requireDeepEqual(t, sliceAt(t, volumeManifest(t, chart, volume), "configMap", "items"), []interface{}{
			map[string]interface{}{"key": "key1", "path": "path1"},
			map[string]interface{}{"key": "key2", "path": "path2"},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L280
	t.Run("fromEmptyDir minimal definition", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromEmptyDir(chart, jsii.String("Volume"), jsii.String("main"), nil)
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"emptyDir": map[string]interface{}{},
			"name":     "main",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L295
	t.Run("fromEmptyDir default medium", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromEmptyDir(chart, jsii.String("Volume"), jsii.String("main"), &plus.EmptyDirVolumeOptions{Medium: plus.EmptyDirMedium_DEFAULT})
		if got := mapAt(t, volumeManifest(t, chart, volume), "emptyDir")["medium"]; got != "" {
			t.Fatalf("medium = %#v, want empty string", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L301
	t.Run("fromEmptyDir memory medium", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromEmptyDir(chart, jsii.String("Volume"), jsii.String("main"), &plus.EmptyDirVolumeOptions{Medium: plus.EmptyDirMedium_MEMORY})
		if got := mapAt(t, volumeManifest(t, chart, volume), "emptyDir")["medium"]; got != "Memory" {
			t.Fatalf("medium = %#v, want Memory", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L307
	t.Run("fromEmptyDir size limit", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromEmptyDir(chart, jsii.String("Volume"), jsii.String("main"), &plus.EmptyDirVolumeOptions{
			SizeLimit: cdk8s.Size_Gibibytes(jsii.Number(20)),
		})
		if got := mapAt(t, volumeManifest(t, chart, volume), "emptyDir")["sizeLimit"]; got != "20480Mi" {
			t.Fatalf("sizeLimit = %#v, want 20480Mi", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L316
	t.Run("fromPersistentVolumeClaim volume name is trimmed if needed", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim := plus.PersistentVolumeClaim_FromClaimName(chart, jsii.String("Claim"), jsii.String("veryveryveryveryveryveryveryveryveryveryveryveryveryveryverylong"))
		volume := plus.Volume_FromPersistentVolumeClaim(chart, jsii.String("Volume"), claim, nil)
		want := "pvc-veryveryveryveryveryveryveryveryveryveryveryveryveryveryver"
		if got := stringValue(volume.Name()); got != want {
			t.Fatalf("volume name = %q, want %q", got, want)
		}
		if got := volumeManifest(t, chart, volume)["name"]; got != want {
			t.Fatalf("synthesized volume name = %#v, want %q", got, want)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L330
	t.Run("fromPersistentVolumeClaim defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim := plus.PersistentVolumeClaim_FromClaimName(chart, jsii.String("Claim"), jsii.String("claim"))
		volume := plus.Volume_FromPersistentVolumeClaim(chart, jsii.String("Volume"), claim, nil)
		if got := stringValue(volume.Name()); got != "pvc-claim" {
			t.Fatalf("volume name = %q, want pvc-claim", got)
		}
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"name": "pvc-claim",
			"persistentVolumeClaim": map[string]interface{}{
				"claimName": "claim",
				"readOnly":  false,
			},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L347
	t.Run("fromPersistentVolumeClaim custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim := plus.PersistentVolumeClaim_FromClaimName(chart, jsii.String("Claim"), jsii.String("claim"))
		volume := plus.Volume_FromPersistentVolumeClaim(chart, jsii.String("Volume"), claim, &plus.PersistentVolumeClaimVolumeOptions{
			Name: jsii.String("custom"), ReadOnly: jsii.Bool(true),
		})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"name": "custom",
			"persistentVolumeClaim": map[string]interface{}{
				"claimName": "claim",
				"readOnly":  true,
			},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L371
	t.Run("fromAwsElasticBlockStore volume name is trimmed if needed", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromAwsElasticBlockStore(chart, jsii.String("Volume"), jsii.String("veryveryveryveryveryveryveryveryveryveryveryveryveryveryverylong"), nil)
		want := "ebs-veryveryveryveryveryveryveryveryveryveryveryveryveryveryver"
		if got := stringValue(volume.Name()); got != want {
			t.Fatalf("volume name = %q, want %q", got, want)
		}
		if got := volumeManifest(t, chart, volume)["name"]; got != want {
			t.Fatalf("synthesized volume name = %#v, want %q", got, want)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L384
	t.Run("fromAwsElasticBlockStore defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromAwsElasticBlockStore(chart, jsii.String("Volume"), jsii.String("vol"), nil)
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"awsElasticBlockStore": map[string]interface{}{
				"fsType":   "ext4",
				"readOnly": false,
				"volumeId": "vol",
			},
			"name": "ebs-vol",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L400
	t.Run("fromAwsElasticBlockStore custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromAwsElasticBlockStore(chart, jsii.String("Volume"), jsii.String("vol"), &plus.AwsElasticBlockStoreVolumeOptions{
			FsType: jsii.String("fs"), Name: jsii.String("name"), Partition: jsii.Number(1), ReadOnly: jsii.Bool(true),
		})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"awsElasticBlockStore": map[string]interface{}{
				"fsType":    "fs",
				"partition": float64(1),
				"readOnly":  true,
				"volumeId":  "vol",
			},
			"name": "name",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L426
	t.Run("fromGcePersistentDisk volume name is trimmed if needed", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromGcePersistentDisk(chart, jsii.String("Volume"), jsii.String("veryveryveryveryveryveryveryveryveryveryveryveryveryveryverylong"), nil)
		want := "gcedisk-veryveryveryveryveryveryveryveryveryveryveryveryveryver"
		if got := stringValue(volume.Name()); got != want {
			t.Fatalf("volume name = %q, want %q", got, want)
		}
		if got := volumeManifest(t, chart, volume)["name"]; got != want {
			t.Fatalf("synthesized volume name = %#v, want %q", got, want)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L439
	t.Run("fromGcePersistentDisk defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromGcePersistentDisk(chart, jsii.String("Volume"), jsii.String("pd"), nil)
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"gcePersistentDisk": map[string]interface{}{
				"fsType":   "ext4",
				"pdName":   "pd",
				"readOnly": false,
			},
			"name": "gcedisk-pd",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L455
	t.Run("fromGcePersistentDisk custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromGcePersistentDisk(chart, jsii.String("Volume"), jsii.String("pd"), &plus.GCEPersistentDiskVolumeOptions{
			FsType: jsii.String("fs"), Name: jsii.String("name"), Partition: jsii.Number(1), ReadOnly: jsii.Bool(true),
		})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"gcePersistentDisk": map[string]interface{}{
				"fsType":    "fs",
				"partition": float64(1),
				"pdName":    "pd",
				"readOnly":  true,
			},
			"name": "name",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L481
	t.Run("fromAzureDisk volume name is trimmed if needed", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromAzureDisk(chart, jsii.String("Volume"), jsii.String("veryveryveryveryveryveryveryveryveryveryveryveryveryveryverylong"), jsii.String("uri"), nil)
		want := "azuredisk-veryveryveryveryveryveryveryveryveryveryveryveryveryv"
		if got := stringValue(volume.Name()); got != want {
			t.Fatalf("volume name = %q, want %q", got, want)
		}
		if got := volumeManifest(t, chart, volume)["name"]; got != want {
			t.Fatalf("synthesized volume name = %#v, want %q", got, want)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L494
	t.Run("fromAzureDisk defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromAzureDisk(chart, jsii.String("Volume"), jsii.String("disk"), jsii.String("uri"), nil)
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"azureDisk": map[string]interface{}{
				"cachingMode": "None",
				"diskName":    "disk",
				"diskUri":     "uri",
				"fsType":      "ext4",
				"kind":        "Shared",
				"readOnly":    false,
			},
			"name": "azuredisk-disk",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L513
	t.Run("fromAzureDisk custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromAzureDisk(chart, jsii.String("Volume"), jsii.String("disk"), jsii.String("uri"), &plus.AzureDiskVolumeOptions{
			CachingMode: plus.AzureDiskPersistentVolumeCachingMode_READ_ONLY,
			FsType:      jsii.String("fs"),
			Kind:        plus.AzureDiskPersistentVolumeKind_DEDICATED,
			Name:        jsii.String("name"),
			ReadOnly:    jsii.Bool(true),
		})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"azureDisk": map[string]interface{}{
				"cachingMode": "ReadOnly",
				"diskName":    "disk",
				"diskUri":     "uri",
				"fsType":      "fs",
				"kind":        "Dedicated",
				"readOnly":    true,
			},
			"name": "name",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L543
	t.Run("fromHostPath defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromHostPath(chart, jsii.String("Volume"), jsii.String("disk"), &plus.HostPathVolumeOptions{Path: jsii.String("/host/path")})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"hostPath": map[string]interface{}{
				"path": "/host/path",
				"type": "",
			},
			"name": "disk",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L560
	t.Run("fromHostPath custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromHostPath(chart, jsii.String("Volume"), jsii.String("disk"), &plus.HostPathVolumeOptions{
			Path: jsii.String("/host/path"), Type: plus.HostPathVolumeType_DIRECTORY,
		})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"hostPath": map[string]interface{}{
				"path": "/host/path",
				"type": "Directory",
			},
			"name": "disk",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L582
	t.Run("fromNfs defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromNfs(chart, jsii.String("Volume"), jsii.String("disk"), &plus.NfsVolumeOptions{
			Server: jsii.String("169.254.1.1"), Path: jsii.String("/nfs/path"),
		})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"name": "disk",
			"nfs": map[string]interface{}{
				"path":   "/nfs/path",
				"server": "169.254.1.1",
			},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L600
	t.Run("fromNfs custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromNfs(chart, jsii.String("Volume"), jsii.String("disk"), &plus.NfsVolumeOptions{
			Server: jsii.String("169.254.1.1"), Path: jsii.String("/nfs/path"), ReadOnly: jsii.Bool(true),
		})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"name": "disk",
			"nfs": map[string]interface{}{
				"path":     "/nfs/path",
				"readOnly": true,
				"server":   "169.254.1.1",
			},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L623
	t.Run("fromCsi minimal definition", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.Volume_FromCsi(chart, jsii.String("Csi"), jsii.String("my-csi-driver"), nil)
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"csi": map[string]interface{}{
				"driver": "my-csi-driver",
			},
			"name": "test-csi-c8e2763d",
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/volume.test.ts#L644
	t.Run("fromCsi custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		attributes := map[string]*string{"secretProviderClass": jsii.String("my-csi")}
		volume := plus.Volume_FromCsi(chart, jsii.String("Csi"), jsii.String("secrets-store.csi.k8s.io"), &plus.CsiVolumeOptions{
			Attributes: &attributes,
			FsType:     jsii.String("ext4"),
			Name:       jsii.String("filesystem"),
			ReadOnly:   jsii.Bool(true),
		})
		requireDeepEqual(t, volumeManifest(t, chart, volume), map[string]interface{}{
			"csi": map[string]interface{}{
				"driver":   "secrets-store.csi.k8s.io",
				"fsType":   "ext4",
				"readOnly": true,
				"volumeAttributes": map[string]interface{}{
					"secretProviderClass": "my-csi",
				},
			},
			"name": "filesystem",
		})
	})
}
