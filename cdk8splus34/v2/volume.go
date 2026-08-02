package cdk8splus34

import (
	"sort"
	"strconv"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// MountOptions controls one volume mount.
type MountOptions struct {
	Propagation MountPropagation `field:"optional" json:"propagation" yaml:"propagation"`
	ReadOnly    *bool            `field:"optional" json:"readOnly" yaml:"readOnly"`
	SubPath     *string          `field:"optional" json:"subPath" yaml:"subPath"`
	SubPathExpr *string          `field:"optional" json:"subPathExpr" yaml:"subPathExpr"`
}

// VolumeMount is an attached volume.
type VolumeMount struct {
	Path        *string          `field:"required" json:"path" yaml:"path"`
	Volume      Volume           `field:"required" json:"volume" yaml:"volume"`
	Propagation MountPropagation `field:"optional" json:"propagation" yaml:"propagation"`
	ReadOnly    *bool            `field:"optional" json:"readOnly" yaml:"readOnly"`
	SubPath     *string          `field:"optional" json:"subPath" yaml:"subPath"`
	SubPathExpr *string          `field:"optional" json:"subPathExpr" yaml:"subPathExpr"`
}

type PathMapping struct {
	Path *string  `field:"required" json:"path" yaml:"path"`
	Mode *float64 `field:"optional" json:"mode" yaml:"mode"`
}

type AwsElasticBlockStoreVolumeOptions struct {
	FsType    *string  `field:"optional" json:"fsType" yaml:"fsType"`
	Name      *string  `field:"optional" json:"name" yaml:"name"`
	Partition *float64 `field:"optional" json:"partition" yaml:"partition"`
	ReadOnly  *bool    `field:"optional" json:"readOnly" yaml:"readOnly"`
}

type AzureDiskVolumeOptions struct {
	CachingMode AzureDiskPersistentVolumeCachingMode `field:"optional" json:"cachingMode" yaml:"cachingMode"`
	FsType      *string                              `field:"optional" json:"fsType" yaml:"fsType"`
	Kind        AzureDiskPersistentVolumeKind        `field:"optional" json:"kind" yaml:"kind"`
	Name        *string                              `field:"optional" json:"name" yaml:"name"`
	ReadOnly    *bool                                `field:"optional" json:"readOnly" yaml:"readOnly"`
}

type GCEPersistentDiskVolumeOptions struct {
	FsType    *string  `field:"optional" json:"fsType" yaml:"fsType"`
	Name      *string  `field:"optional" json:"name" yaml:"name"`
	Partition *float64 `field:"optional" json:"partition" yaml:"partition"`
	ReadOnly  *bool    `field:"optional" json:"readOnly" yaml:"readOnly"`
}

type HostPathVolumeOptions struct {
	Path *string            `field:"required" json:"path" yaml:"path"`
	Type HostPathVolumeType `field:"optional" json:"type" yaml:"type"`
}

type NfsVolumeOptions struct {
	Path     *string `field:"required" json:"path" yaml:"path"`
	Server   *string `field:"required" json:"server" yaml:"server"`
	ReadOnly *bool   `field:"optional" json:"readOnly" yaml:"readOnly"`
}

type CsiVolumeOptions struct {
	Attributes *map[string]*string `field:"optional" json:"attributes" yaml:"attributes"`
	FsType     *string             `field:"optional" json:"fsType" yaml:"fsType"`
	Name       *string             `field:"optional" json:"name" yaml:"name"`
	ReadOnly   *bool               `field:"optional" json:"readOnly" yaml:"readOnly"`
}

type SecretVolumeOptions struct {
	DefaultMode *float64                 `field:"optional" json:"defaultMode" yaml:"defaultMode"`
	Items       *map[string]*PathMapping `field:"optional" json:"items" yaml:"items"`
	Name        *string                  `field:"optional" json:"name" yaml:"name"`
	Optional    *bool                    `field:"optional" json:"optional" yaml:"optional"`
}

type PersistentVolumeClaimVolumeOptions struct {
	Name     *string `field:"optional" json:"name" yaml:"name"`
	ReadOnly *bool   `field:"optional" json:"readOnly" yaml:"readOnly"`
}

type AzureDiskPersistentVolumeKind string

const (
	AzureDiskPersistentVolumeKind_SHARED    AzureDiskPersistentVolumeKind = "SHARED"
	AzureDiskPersistentVolumeKind_DEDICATED AzureDiskPersistentVolumeKind = "DEDICATED"
	AzureDiskPersistentVolumeKind_MANAGED   AzureDiskPersistentVolumeKind = "MANAGED"
)

type AzureDiskPersistentVolumeCachingMode string

const (
	AzureDiskPersistentVolumeCachingMode_NONE       AzureDiskPersistentVolumeCachingMode = "NONE"
	AzureDiskPersistentVolumeCachingMode_READ_ONLY  AzureDiskPersistentVolumeCachingMode = "READ_ONLY"
	AzureDiskPersistentVolumeCachingMode_READ_WRITE AzureDiskPersistentVolumeCachingMode = "READ_WRITE"
)

type HostPathVolumeType string

const (
	HostPathVolumeType_DEFAULT             HostPathVolumeType = "DEFAULT"
	HostPathVolumeType_DIRECTORY_OR_CREATE HostPathVolumeType = "DIRECTORY_OR_CREATE"
	HostPathVolumeType_DIRECTORY           HostPathVolumeType = "DIRECTORY"
	HostPathVolumeType_FILE_OR_CREATE      HostPathVolumeType = "FILE_OR_CREATE"
	HostPathVolumeType_FILE                HostPathVolumeType = "FILE"
	HostPathVolumeType_SOCKET              HostPathVolumeType = "SOCKET"
	HostPathVolumeType_CHAR_DEVICE         HostPathVolumeType = "CHAR_DEVICE"
	HostPathVolumeType_BLOCK_DEVICE        HostPathVolumeType = "BLOCK_DEVICE"
)

// IStorage represents storage that can be mounted into a container.
type IStorage interface {
	constructs.IConstruct
	AsVolume() Volume
}

// Volume is a named pod volume.
type Volume interface {
	IStorage
	constructs.Construct
	Name() *string
}

type volumeImpl struct {
	node constructs.Node
	name *string
	spec map[string]interface{}
}

func (v *volumeImpl) Node() constructs.Node {
	return v.node
}

func (v *volumeImpl) SetNodeInternal(node constructs.Node) {
	v.node = node
}

func (v *volumeImpl) ToString() *string {
	return v.node.Path()
}

func (v *volumeImpl) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return v.node.With(mixins...)
}

func (v *volumeImpl) Name() *string {
	return v.name
}

func (v *volumeImpl) AsVolume() Volume {
	return v
}

type ConfigMapVolumeOptions struct {
	DefaultMode *float64                 `field:"optional" json:"defaultMode" yaml:"defaultMode"`
	Items       *map[string]*PathMapping `field:"optional" json:"items" yaml:"items"`
	Name        *string                  `field:"optional" json:"name" yaml:"name"`
	Optional    *bool                    `field:"optional" json:"optional" yaml:"optional"`
}

type EmptyDirMedium string

const (
	EmptyDirMedium_DEFAULT EmptyDirMedium = ""
	EmptyDirMedium_MEMORY  EmptyDirMedium = "Memory"
)

type EmptyDirVolumeOptions struct {
	Medium    EmptyDirMedium `field:"optional" json:"medium" yaml:"medium"`
	SizeLimit cdk8s.Size     `field:"optional" json:"sizeLimit" yaml:"sizeLimit"`
}

func newVolume(scope constructs.Construct, id, name *string, spec map[string]interface{}) Volume {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	actualName := *name
	if len(actualName) > 63 {
		actualName = actualName[:63]
	}
	result := &volumeImpl{name: jsii.String(actualName), spec: spec}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

func Volume_FromConfigMap(scope constructs.Construct, id *string, configMap IConfigMap, options *ConfigMapVolumeOptions) Volume {
	if configMap == nil {
		panic("configMap is required")
	}
	name := jsii.String("configmap-" + stringValue(configMap.Name()))
	if options != nil && options.Name != nil {
		name = options.Name
	}
	config := map[string]interface{}{"name": configMap.Name()}
	if options != nil && options.DefaultMode != nil {
		config["defaultMode"] = options.DefaultMode
	}
	if options != nil && options.Optional != nil {
		config["optional"] = options.Optional
	}
	if options != nil {
		config["items"] = volumePathMappings(options.Items)
	}
	return newVolume(scope, id, name, map[string]interface{}{"configMap": config})
}

func Volume_FromEmptyDir(scope constructs.Construct, id, name *string, options *EmptyDirVolumeOptions) Volume {
	spec := map[string]interface{}{"emptyDir": map[string]interface{}{}}
	if options != nil && options.Medium != "" {
		spec["emptyDir"].(map[string]interface{})["medium"] = string(options.Medium)
	}
	if options != nil && options.SizeLimit != nil {
		amount := options.SizeLimit.ToMebibytes(nil)
		spec["emptyDir"].(map[string]interface{})["sizeLimit"] = strconv.FormatFloat(*amount, 'f', -1, 64) + "Mi"
	}
	return newVolume(scope, id, name, spec)
}

func Volume_FromName(scope constructs.Construct, id, name *string) Volume {
	return newVolume(scope, id, name, map[string]interface{}{})
}

func Volume_FromAwsElasticBlockStore(scope constructs.Construct, id, volumeID *string, options *AwsElasticBlockStoreVolumeOptions) Volume {
	if volumeID == nil {
		panic("volumeId is required")
	}
	name, fsType, readOnly := jsii.String("ebs-"+*volumeID), jsii.String("ext4"), jsii.Bool(false)
	if options != nil {
		if options.Name != nil {
			name = options.Name
		}
		if options.FsType != nil {
			fsType = options.FsType
		}
		if options.ReadOnly != nil {
			readOnly = options.ReadOnly
		}
	}
	spec := map[string]interface{}{"volumeId": volumeID, "fsType": fsType, "readOnly": readOnly}
	if options != nil && options.Partition != nil {
		spec["partition"] = options.Partition
	}
	return newVolume(scope, id, name, map[string]interface{}{"awsElasticBlockStore": spec})
}

func Volume_FromAzureDisk(scope constructs.Construct, id, diskName, diskURI *string, options *AzureDiskVolumeOptions) Volume {
	if diskName == nil || diskURI == nil {
		panic("diskName and diskUri are required")
	}
	name, fsType, readOnly := jsii.String("azuredisk-"+*diskName), jsii.String("ext4"), jsii.Bool(false)
	caching, kind := AzureDiskPersistentVolumeCachingMode_NONE, AzureDiskPersistentVolumeKind_SHARED
	if options != nil {
		if options.Name != nil {
			name = options.Name
		}
		if options.FsType != nil {
			fsType = options.FsType
		}
		if options.ReadOnly != nil {
			readOnly = options.ReadOnly
		}
		if options.CachingMode != "" {
			caching = options.CachingMode
		}
		if options.Kind != "" {
			kind = options.Kind
		}
	}
	return newVolume(scope, id, name, map[string]interface{}{"azureDisk": map[string]interface{}{"diskName": diskName, "diskUri": diskURI, "cachingMode": azureCachingModeManifest(caching), "fsType": fsType, "kind": azureKindManifest(kind), "readOnly": readOnly}})
}

func Volume_FromGcePersistentDisk(scope constructs.Construct, id, pdName *string, options *GCEPersistentDiskVolumeOptions) Volume {
	if pdName == nil {
		panic("pdName is required")
	}
	name, fsType, readOnly := jsii.String("gcedisk-"+*pdName), jsii.String("ext4"), jsii.Bool(false)
	if options != nil {
		if options.Name != nil {
			name = options.Name
		}
		if options.FsType != nil {
			fsType = options.FsType
		}
		if options.ReadOnly != nil {
			readOnly = options.ReadOnly
		}
	}
	spec := map[string]interface{}{"pdName": pdName, "fsType": fsType, "readOnly": readOnly}
	if options != nil && options.Partition != nil {
		spec["partition"] = options.Partition
	}
	return newVolume(scope, id, name, map[string]interface{}{"gcePersistentDisk": spec})
}

func Volume_FromSecret(scope constructs.Construct, id *string, secret ISecret, options *SecretVolumeOptions) Volume {
	if secret == nil {
		panic("secret is required")
	}
	name := jsii.String("secret-" + stringValue(secret.Name()))
	if options != nil && options.Name != nil {
		name = options.Name
	}
	spec := map[string]interface{}{"secretName": secret.Name()}
	if options != nil {
		if options.DefaultMode != nil {
			spec["defaultMode"] = options.DefaultMode
		}
		if options.Optional != nil {
			spec["optional"] = options.Optional
		}
		spec["items"] = volumePathMappings(options.Items)
	}
	return newVolume(scope, id, name, map[string]interface{}{"secret": spec})
}

func Volume_FromPersistentVolumeClaim(scope constructs.Construct, id *string, claim IPersistentVolumeClaim, options *PersistentVolumeClaimVolumeOptions) Volume {
	if claim == nil {
		panic("claim is required")
	}
	name, readOnly := jsii.String("pvc-"+stringValue(claim.Name())), jsii.Bool(false)
	if options != nil {
		if options.Name != nil {
			name = options.Name
		}
		if options.ReadOnly != nil {
			readOnly = options.ReadOnly
		}
	}
	return newVolume(scope, id, name, map[string]interface{}{"persistentVolumeClaim": map[string]interface{}{"claimName": claim.Name(), "readOnly": readOnly}})
}

func Volume_FromHostPath(scope constructs.Construct, id, name *string, options *HostPathVolumeOptions) Volume {
	if options == nil || options.Path == nil {
		panic("path is required")
	}
	value := options.Type
	if value == "" {
		value = HostPathVolumeType_DEFAULT
	}
	return newVolume(scope, id, name, map[string]interface{}{"hostPath": map[string]interface{}{"path": options.Path, "type": hostPathTypeManifest(value)}})
}

func Volume_FromNfs(scope constructs.Construct, id, name *string, options *NfsVolumeOptions) Volume {
	if options == nil || options.Path == nil || options.Server == nil {
		panic("server and path are required")
	}
	spec := map[string]interface{}{"server": options.Server, "path": options.Path}
	if options.ReadOnly != nil {
		spec["readOnly"] = options.ReadOnly
	}
	return newVolume(scope, id, name, map[string]interface{}{"nfs": spec})
}

func Volume_FromCsi(scope constructs.Construct, id, driver *string, options *CsiVolumeOptions) Volume {
	if driver == nil {
		panic("driver is required")
	}
	name := cdk8s.Names_ToDnsLabel(scope, &cdk8s.NameOptions{Extra: &[]*string{id}})
	if options != nil && options.Name != nil {
		name = options.Name
	}
	spec := map[string]interface{}{"driver": driver}
	if options != nil {
		if options.FsType != nil {
			spec["fsType"] = options.FsType
		}
		if options.ReadOnly != nil {
			spec["readOnly"] = options.ReadOnly
		}
		if options.Attributes != nil {
			spec["volumeAttributes"] = *options.Attributes
		}
	}
	return newVolume(scope, id, name, map[string]interface{}{"csi": spec})
}

func volumePathMappings(values *map[string]*PathMapping) []interface{} {
	if values == nil {
		return nil
	}
	keys := make([]string, 0, len(*values))
	for key := range *values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		mapping := (*values)[key]
		if mapping == nil || mapping.Path == nil {
			panic("path mapping is required")
		}
		item := map[string]interface{}{"key": key, "path": mapping.Path}
		if mapping.Mode != nil {
			item["mode"] = mapping.Mode
		}
		result = append(result, item)
	}
	return result
}

func azureKindManifest(value AzureDiskPersistentVolumeKind) string {
	switch value {
	case AzureDiskPersistentVolumeKind_SHARED:
		return "Shared"
	case AzureDiskPersistentVolumeKind_DEDICATED:
		return "Dedicated"
	case AzureDiskPersistentVolumeKind_MANAGED:
		return "Managed"
	default:
		panic("invalid Azure disk kind")
	}
}

func azureCachingModeManifest(value AzureDiskPersistentVolumeCachingMode) string {
	switch value {
	case AzureDiskPersistentVolumeCachingMode_NONE:
		return "None"
	case AzureDiskPersistentVolumeCachingMode_READ_ONLY:
		return "ReadOnly"
	case AzureDiskPersistentVolumeCachingMode_READ_WRITE:
		return "ReadWrite"
	default:
		panic("invalid Azure disk caching mode")
	}
}

func hostPathTypeManifest(value HostPathVolumeType) string {
	switch value {
	case HostPathVolumeType_DEFAULT:
		return ""
	case HostPathVolumeType_DIRECTORY_OR_CREATE:
		return "DirectoryOrCreate"
	case HostPathVolumeType_DIRECTORY:
		return "Directory"
	case HostPathVolumeType_FILE_OR_CREATE:
		return "FileOrCreate"
	case HostPathVolumeType_FILE:
		return "File"
	case HostPathVolumeType_SOCKET:
		return "Socket"
	case HostPathVolumeType_CHAR_DEVICE:
		return "CharDevice"
	case HostPathVolumeType_BLOCK_DEVICE:
		return "BlockDevice"
	default:
		panic("invalid host path type")
	}
}

func Volume_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}
