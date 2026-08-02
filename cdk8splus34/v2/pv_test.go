package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func newDefaultAwsVolume(chart cdk8s.Chart) plus.AwsElasticBlockStorePersistentVolume {
	return plus.NewAwsElasticBlockStorePersistentVolume(chart, jsii.String("Volume"), &plus.AwsElasticBlockStorePersistentVolumeProps{VolumeId: jsii.String("vol1")})
}

func TestPersistentVolume(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L7
	t.Run("can grant permissions on imported", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.PersistentVolume_FromPersistentVolumeName(chart, jsii.String("Vol"), jsii.String("vol"))
		role := plus.NewRole(chart, jsii.String("Role"), nil)
		role.AllowRead(volume)
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{rbacRule("", "persistentvolumes", []interface{}{"vol"}, "get", "list", "watch")})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L19
	t.Run("defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := newDefaultAwsVolume(chart)
		if volume.AccessModes() != nil || volume.ReclaimPolicy() != plus.PersistentVolumeReclaimPolicy_RETAIN || volume.Storage() != nil || volume.StorageClassName() != nil || volume.Mode() != plus.PersistentVolumeMode_FILE_SYSTEM || volume.MountOptions() != nil {
			t.Fatal("unexpected persistent-volume defaults")
		}
		spec := mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec")
		requireDeepEqual(t, spec, map[string]interface{}{
			"awsElasticBlockStore":          map[string]interface{}{"fsType": "ext4", "readOnly": false, "volumeID": "vol1"},
			"persistentVolumeReclaimPolicy": "Retain", "volumeMode": "Filesystem",
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L38
	t.Run("custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.NewAwsElasticBlockStorePersistentVolume(chart, jsii.String("Volume"), &plus.AwsElasticBlockStorePersistentVolumeProps{
			AccessModes:  &[]plus.PersistentVolumeAccessMode{plus.PersistentVolumeAccessMode_READ_ONLY_MANY, plus.PersistentVolumeAccessMode_READ_WRITE_MANY},
			MountOptions: &[]*string{jsii.String("opt1")}, ReclaimPolicy: plus.PersistentVolumeReclaimPolicy_DELETE,
			VolumeMode: plus.PersistentVolumeMode_BLOCK, StorageClassName: jsii.String("storage-class"),
			Storage: cdk8s.Size_Gibibytes(jsii.Number(50)), VolumeId: jsii.String("vol1"),
		})
		modes := *volume.AccessModes()
		if len(modes) != 2 || modes[0] != plus.PersistentVolumeAccessMode_READ_ONLY_MANY || modes[1] != plus.PersistentVolumeAccessMode_READ_WRITE_MANY || volume.ReclaimPolicy() != plus.PersistentVolumeReclaimPolicy_DELETE || numberValue(volume.Storage().ToGibibytes(nil)) != 50 || stringValue(volume.StorageClassName()) != "storage-class" || volume.Mode() != plus.PersistentVolumeMode_BLOCK || stringValue((*volume.MountOptions())[0]) != "opt1" {
			t.Fatal("unexpected custom persistent-volume properties")
		}
		spec := mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec")
		requireDeepEqual(t, spec, map[string]interface{}{
			"accessModes":          []interface{}{"ReadOnlyMany", "ReadWriteMany"},
			"awsElasticBlockStore": map[string]interface{}{"fsType": "ext4", "readOnly": false, "volumeID": "vol1"},
			"capacity":             map[string]interface{}{"storage": "50Gi"}, "mountOptions": []interface{}{"opt1"},
			"persistentVolumeReclaimPolicy": "Delete", "storageClassName": "storage-class", "volumeMode": "Block",
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L64
	t.Run("can be imported", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.PersistentVolume_FromPersistentVolumeName(chart, jsii.String("Vol"), jsii.String("vol"))
		if got := stringValue(volume.Name()); got != "vol" {
			t.Fatalf("volume name = %q", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L72
	t.Run("can reserve with default storage class", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := newDefaultAwsVolume(chart)
		claim := volume.Reserve()
		if claim.AccessModes() != nil || claim.Storage() != nil || claim.StorageClassName() != volume.StorageClassName() || claim.Volume() != volume || volume.Claim() != claim {
			t.Fatal("reserved claim is not bidirectionally bound with defaults")
		}
		volumeSpec := mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec")
		claimSpec := mapAt(t, manifestOfKind(t, chart, "PersistentVolumeClaim"), "spec")
		if mapAt(t, volumeSpec, "claimRef")["name"] != "pvc-test-volume-c8db061e" || claimSpec["volumeName"] != "test-volume-c8db061e" {
			t.Fatalf("reservation manifests = %#v / %#v", volumeSpec, claimSpec)
		}
		if _, ok := claimSpec["storageClassName"]; ok {
			t.Fatal("default reservation has storageClassName")
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L94
	t.Run("can reserve with custom storage class", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.NewAwsElasticBlockStorePersistentVolume(chart, jsii.String("Volume"), &plus.AwsElasticBlockStorePersistentVolumeProps{VolumeId: jsii.String("vol1"), StorageClassName: jsii.String("storage-class")})
		claim := volume.Reserve()
		if stringValue(claim.StorageClassName()) != stringValue(volume.StorageClassName()) || claim.Volume() != volume || volume.Claim() != claim {
			t.Fatal("reserved claim did not inherit custom storage class")
		}
		if got := mapAt(t, manifestOfKind(t, chart, "PersistentVolumeClaim"), "spec")["storageClassName"]; got != "storage-class" {
			t.Fatalf("storageClassName = %#v", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L117
	t.Run("reserved claim uses volume namespace", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.NewAwsElasticBlockStorePersistentVolume(chart, jsii.String("Volume"), &plus.AwsElasticBlockStorePersistentVolumeProps{
			Metadata: &cdk8s.ApiObjectMetadata{Namespace: jsii.String("non-default")}, VolumeId: jsii.String("vol1"),
		})
		claim := volume.Reserve()
		if stringValue(claim.Metadata().Namespace()) != stringValue(volume.Metadata().Namespace()) {
			t.Fatalf("claim namespace = %q, volume namespace = %q", stringValue(claim.Metadata().Namespace()), stringValue(volume.Metadata().Namespace()))
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L130
	t.Run("throws if reserved twice", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := newDefaultAwsVolume(chart)
		volume.Reserve()
		requirePanicContains(t, "There is already a Construct with name 'test-volume-c8db061ePVC'", func() { volume.Reserve() })
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L143
	t.Run("can bind to a claim at instantiation", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim := plus.PersistentVolumeClaim_FromClaimName(chart, jsii.String("Claim"), jsii.String("claim"))
		volume := plus.NewAwsElasticBlockStorePersistentVolume(chart, jsii.String("Volume"), &plus.AwsElasticBlockStorePersistentVolumeProps{VolumeId: jsii.String("vol1"), Claim: claim})
		if volume.Claim() != claim || mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec", "claimRef")["name"] != "claim" {
			t.Fatal("volume was not bound to claim")
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L159
	t.Run("can bind to a claim after instantiation", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim := plus.PersistentVolumeClaim_FromClaimName(chart, jsii.String("Claim"), jsii.String("claim"))
		volume := newDefaultAwsVolume(chart)
		volume.Bind(claim)
		if volume.Claim() != claim || mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec", "claimRef")["name"] != "claim" {
			t.Fatal("volume was not bound to claim")
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L176
	t.Run("no-op binding twice to same claim", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim := plus.PersistentVolumeClaim_FromClaimName(chart, jsii.String("Claim"), jsii.String("claim"))
		volume := newDefaultAwsVolume(chart)
		volume.Bind(claim)
		volume.Bind(claim)
		if volume.Claim() != claim {
			t.Fatal("volume did not retain claim")
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L191
	t.Run("throws binding twice to different claims", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim1 := plus.PersistentVolumeClaim_FromClaimName(chart, jsii.String("Claim1"), jsii.String("claim1"))
		claim2 := plus.PersistentVolumeClaim_FromClaimName(chart, jsii.String("Claim2"), jsii.String("claim2"))
		volume := newDefaultAwsVolume(chart)
		volume.Bind(claim1)
		requirePanicContains(t, "Cannot bind volume 'test-volume-c8db061e' to claim 'claim2' since it is already bound to claim 'claim1'", func() { volume.Bind(claim2) })
	})
}

func TestAwsElasticBlockStorePersistentVolume(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L210
	t.Run("defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := newDefaultAwsVolume(chart)
		if stringValue(volume.VolumeId()) != "vol1" || stringValue(volume.FsType()) != "ext4" || volume.Partition() != nil || boolValue(volume.ReadOnly()) {
			t.Fatal("unexpected AWS volume defaults")
		}
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec", "awsElasticBlockStore"), map[string]interface{}{
			"volumeID": "vol1", "fsType": "ext4", "readOnly": false,
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L227
	t.Run("custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.NewAwsElasticBlockStorePersistentVolume(chart, jsii.String("Volume"), &plus.AwsElasticBlockStorePersistentVolumeProps{
			VolumeId: jsii.String("vol1"), Partition: jsii.Number(1), ReadOnly: jsii.Bool(true), FsType: jsii.String("ntfs"),
		})
		if stringValue(volume.VolumeId()) != "vol1" || stringValue(volume.FsType()) != "ntfs" || numberValue(volume.Partition()) != 1 || !boolValue(volume.ReadOnly()) {
			t.Fatal("unexpected custom AWS volume properties")
		}
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec", "awsElasticBlockStore"), map[string]interface{}{
			"volumeID": "vol1", "fsType": "ntfs", "partition": float64(1), "readOnly": true,
		})
	})
}

func TestAzureDiskPersistentVolume(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L251
	t.Run("defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.NewAzureDiskPersistentVolume(chart, jsii.String("Volume"), &plus.AzureDiskPersistentVolumeProps{DiskName: jsii.String("name"), DiskUri: jsii.String("uri")})
		if stringValue(volume.DiskName()) != "name" || stringValue(volume.DiskUri()) != "uri" || volume.CachingMode() != plus.AzureDiskPersistentVolumeCachingMode_NONE || boolValue(volume.ReadOnly()) || stringValue(volume.FsType()) != "ext4" || volume.AzureKind() != plus.AzureDiskPersistentVolumeKind_SHARED {
			t.Fatal("unexpected Azure volume defaults")
		}
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec", "azureDisk"), map[string]interface{}{
			"diskName": "name", "diskURI": "uri", "cachingMode": "None", "readOnly": false, "fsType": "ext4", "kind": "Shared",
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L271
	t.Run("custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.NewAzureDiskPersistentVolume(chart, jsii.String("Volume"), &plus.AzureDiskPersistentVolumeProps{
			DiskName: jsii.String("name"), DiskUri: jsii.String("uri"), CachingMode: plus.AzureDiskPersistentVolumeCachingMode_READ_ONLY,
			ReadOnly: jsii.Bool(true), FsType: jsii.String("ntfs"), Kind: plus.AzureDiskPersistentVolumeKind_DEDICATED,
		})
		if volume.CachingMode() != plus.AzureDiskPersistentVolumeCachingMode_READ_ONLY || !boolValue(volume.ReadOnly()) || stringValue(volume.FsType()) != "ntfs" || volume.AzureKind() != plus.AzureDiskPersistentVolumeKind_DEDICATED {
			t.Fatal("unexpected custom Azure volume properties")
		}
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec", "azureDisk"), map[string]interface{}{
			"diskName": "name", "diskURI": "uri", "cachingMode": "ReadOnly", "readOnly": true, "fsType": "ntfs", "kind": "Dedicated",
		})
	})
}

func TestGCEPersistentDiskPersistentVolume(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L299
	t.Run("defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.NewGCEPersistentDiskPersistentVolume(chart, jsii.String("Volume"), &plus.GCEPersistentDiskPersistentVolumeProps{PdName: jsii.String("name")})
		if stringValue(volume.PdName()) != "name" || volume.Partition() != nil || boolValue(volume.ReadOnly()) || stringValue(volume.FsType()) != "ext4" {
			t.Fatal("unexpected GCE volume defaults")
		}
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec", "gcePersistentDisk"), map[string]interface{}{
			"pdName": "name", "fsType": "ext4", "readOnly": false,
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pv.test.ts#L316
	t.Run("custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.NewGCEPersistentDiskPersistentVolume(chart, jsii.String("Volume"), &plus.GCEPersistentDiskPersistentVolumeProps{
			PdName: jsii.String("name"), Partition: jsii.Number(1), ReadOnly: jsii.Bool(true), FsType: jsii.String("ntfs"),
		})
		if stringValue(volume.PdName()) != "name" || numberValue(volume.Partition()) != 1 || !boolValue(volume.ReadOnly()) || stringValue(volume.FsType()) != "ntfs" {
			t.Fatal("unexpected custom GCE volume properties")
		}
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "PersistentVolume"), "spec", "gcePersistentDisk"), map[string]interface{}{
			"pdName": "name", "partition": float64(1), "readOnly": true, "fsType": "ntfs",
		})
	})
}
