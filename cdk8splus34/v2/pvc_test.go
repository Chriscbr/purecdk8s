package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func TestPersistentVolumeClaim(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pvc.test.ts#L5
	t.Run("can grant permissions on imported", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim := plus.PersistentVolumeClaim_FromClaimName(chart, jsii.String("Claim"), jsii.String("claim"))
		role := plus.NewRole(chart, jsii.String("Role"), nil)
		role.AllowRead(claim)
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{rbacRule("", "persistentvolumeclaims", []interface{}{"claim"}, "get", "list", "watch")})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pvc.test.ts#L17
	t.Run("defaults", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim := plus.NewPersistentVolumeClaim(chart, jsii.String("PersistentVolumeClaim"), nil)
		if claim.AccessModes() != nil || claim.Storage() != nil || claim.StorageClassName() != nil || claim.VolumeMode() != plus.PersistentVolumeMode_FILE_SYSTEM {
			t.Fatalf("unexpected default claim properties")
		}
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "PersistentVolumeClaim"), "spec"), map[string]interface{}{"volumeMode": "Filesystem"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pvc.test.ts#L32
	t.Run("custom", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		storage := cdk8s.Size_Gibibytes(jsii.Number(50))
		claim := plus.NewPersistentVolumeClaim(chart, jsii.String("PersistentVolumeClaim"), &plus.PersistentVolumeClaimProps{
			AccessModes: &[]plus.PersistentVolumeAccessMode{plus.PersistentVolumeAccessMode_READ_WRITE_MANY}, Storage: storage,
			StorageClassName: jsii.String("storage-class"), VolumeMode: plus.PersistentVolumeMode_BLOCK,
		})
		if modes := *claim.AccessModes(); len(modes) != 1 || modes[0] != plus.PersistentVolumeAccessMode_READ_WRITE_MANY {
			t.Fatalf("access modes = %#v", modes)
		}
		if numberValue(claim.Storage().ToGibibytes(nil)) != 50 || stringValue(claim.StorageClassName()) != "storage-class" || claim.VolumeMode() != plus.PersistentVolumeMode_BLOCK {
			t.Fatalf("unexpected custom claim properties")
		}
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "PersistentVolumeClaim"), "spec"), map[string]interface{}{
			"accessModes":      []interface{}{"ReadWriteMany"},
			"resources":        map[string]interface{}{"requests": map[string]interface{}{"storage": "50Gi"}},
			"storageClassName": "storage-class", "volumeMode": "Block",
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pvc.test.ts#L52
	t.Run("small size", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim := plus.NewPersistentVolumeClaim(chart, jsii.String("PersistentVolumeClaim"), &plus.PersistentVolumeClaimProps{Storage: cdk8s.Size_Mebibytes(jsii.Number(512))})
		if numberValue(claim.Storage().ToMebibytes(nil)) != 512 {
			t.Fatalf("storage = %#v", claim.Storage())
		}
		storage := mapAt(t, manifestOfKind(t, chart, "PersistentVolumeClaim"), "spec", "resources", "requests")["storage"]
		if storage != "0.5Gi" {
			t.Fatalf("storage request = %#v, want 0.5Gi", storage)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pvc.test.ts#L65
	t.Run("can be imported", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		claim := plus.PersistentVolumeClaim_FromClaimName(chart, jsii.String("Claim"), jsii.String("claim"))
		if got := stringValue(claim.Name()); got != "claim" {
			t.Fatalf("claim name = %q", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pvc.test.ts#L73
	t.Run("can bind to a volume at instantiation", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.PersistentVolume_FromPersistentVolumeName(chart, jsii.String("Vol"), jsii.String("vol"))
		claim := plus.NewPersistentVolumeClaim(chart, jsii.String("PersistentVolumeClaim"), &plus.PersistentVolumeClaimProps{Volume: volume})
		if claim.Volume() != volume || stringValue(claim.Volume().Name()) != "vol" {
			t.Fatal("claim did not retain its volume")
		}
		if got := mapAt(t, manifestOfKind(t, chart, "PersistentVolumeClaim"), "spec")["volumeName"]; got != "vol" {
			t.Fatalf("volumeName = %#v", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pvc.test.ts#L88
	t.Run("can bind to a volume after instantiation", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.PersistentVolume_FromPersistentVolumeName(chart, jsii.String("Vol"), jsii.String("vol"))
		claim := plus.NewPersistentVolumeClaim(chart, jsii.String("PersistentVolumeClaim"), nil)
		claim.Bind(volume)
		if claim.Volume() != volume || mapAt(t, manifestOfKind(t, chart, "PersistentVolumeClaim"), "spec")["volumeName"] != "vol" {
			t.Fatal("claim was not bound to volume")
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pvc.test.ts#L103
	t.Run("no-op binding twice to same volume", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume := plus.PersistentVolume_FromPersistentVolumeName(chart, jsii.String("Vole"), jsii.String("vol"))
		claim := plus.NewPersistentVolumeClaim(chart, jsii.String("PersistentVolumeClaim"), nil)
		claim.Bind(volume)
		claim.Bind(volume)
		if claim.Volume() != volume {
			t.Fatal("claim did not retain volume")
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/pvc.test.ts#L116
	t.Run("throws binding twice to different volumes", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		volume1 := plus.PersistentVolume_FromPersistentVolumeName(chart, jsii.String("Vol1"), jsii.String("vol1"))
		volume2 := plus.PersistentVolume_FromPersistentVolumeName(chart, jsii.String("Vol2"), jsii.String("vol2"))
		claim := plus.NewPersistentVolumeClaim(chart, jsii.String("PersistentVolumeClaim"), nil)
		claim.Bind(volume1)
		requirePanicContains(t, "Cannot bind claim 'test-persistentvolumeclaim-c8af0974' to volume 'vol2' since it is already bound to volume 'vol1'", func() { claim.Bind(volume2) })
	})
}
