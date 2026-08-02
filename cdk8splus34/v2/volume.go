package cdk8splus34

import (
	"sort"
	"strconv"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Options for mounts.
type MountOptions struct {
	// Determines how mounts are propagated from the host to container and the other way around.
	//
	// When not set, MountPropagationNone is used.
	//
	// Mount propagation allows for sharing volumes mounted by a Container to other Containers in the same Pod, or even to other Pods on the same node. Default: MountPropagation.NONE
	Propagation MountPropagation `field:"optional" json:"propagation" yaml:"propagation"`
	// Mounted read-only if true, read-write otherwise (false or unspecified).
	//
	// Defaults to false. Default: false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
	// Path within the volume from which the container's volume should be mounted.). Default: "" the volume's root.
	SubPath *string `field:"optional" json:"subPath" yaml:"subPath"`
	// Expanded path within the volume from which the container's volume should be mounted.
	//
	// Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment. Defaults to "" (volume's root).
	//
	// `subPathExpr` and `subPath` are mutually exclusive. Default: "" volume's root.
	SubPathExpr *string `field:"optional" json:"subPathExpr" yaml:"subPathExpr"`
}

// Mount a volume from the pod to the container.
type VolumeMount struct {
	// Path within the container at which the volume should be mounted.
	//
	// Must not contain ':'.
	Path *string `field:"required" json:"path" yaml:"path"`
	// The volume to mount.
	Volume Volume `field:"required" json:"volume" yaml:"volume"`
	// Determines how mounts are propagated from the host to container and the other way around.
	//
	// When not set, MountPropagationNone is used.
	//
	// Mount propagation allows for sharing volumes mounted by a Container to other Containers in the same Pod, or even to other Pods on the same node. Default: MountPropagation.NONE
	Propagation MountPropagation `field:"optional" json:"propagation" yaml:"propagation"`
	// Mounted read-only if true, read-write otherwise (false or unspecified).
	//
	// Defaults to false. Default: false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
	// Path within the volume from which the container's volume should be mounted.). Default: "" the volume's root.
	SubPath *string `field:"optional" json:"subPath" yaml:"subPath"`
	// Expanded path within the volume from which the container's volume should be mounted.
	//
	// Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment. Defaults to "" (volume's root).
	//
	// `subPathExpr` and `subPath` are mutually exclusive. Default: "" volume's root.
	SubPathExpr *string `field:"optional" json:"subPathExpr" yaml:"subPathExpr"`
}

// Maps a string key to a path within a volume.
type PathMapping struct {
	// The relative path of the file to map the key to.
	//
	// May not be an absolute path. May not contain the path element '..'. May not start with the string '..'.
	Path *string `field:"required" json:"path" yaml:"path"`
	// Optional: mode bits to use on this file, must be a value between 0 and 0777.
	//
	// If not specified, the volume defaultMode will be used. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.
	Mode *float64 `field:"optional" json:"mode" yaml:"mode"`
}

// Options of `Volume.fromAwsElasticBlockStore`.
type AwsElasticBlockStoreVolumeOptions struct {
	// Filesystem type of the volume that you want to mount.
	//
	// Tip: Ensure that the filesystem type is supported by the host operating system. See: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
	//
	// Default: 'ext4'.
	FsType *string `field:"optional" json:"fsType" yaml:"fsType"`
	// The volume name. Default: - auto-generated.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// The partition in the volume that you want to mount.
	//
	// If omitted, the default is to mount by volume name. Examples: For volume /dev/sda1, you specify the partition as "1". Similarly, the volume partition for /dev/sda is "0" (or you can leave the property empty). Default: - No partition.
	Partition *float64 `field:"optional" json:"partition" yaml:"partition"`
	// Specify "true" to force and set the ReadOnly property in VolumeMounts to "true". See: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
	//
	// Default: false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
}

// Options of `Volume.fromAzureDisk`.
type AzureDiskVolumeOptions struct {
	// Host Caching mode. Default: - AzureDiskPersistentVolumeCachingMode.NONE.
	CachingMode AzureDiskPersistentVolumeCachingMode `field:"optional" json:"cachingMode" yaml:"cachingMode"`
	// Filesystem type to mount.
	//
	// Must be a filesystem type supported by the host operating system. Default: 'ext4'.
	FsType *string `field:"optional" json:"fsType" yaml:"fsType"`
	// Kind of disk. Default: AzureDiskPersistentVolumeKind.SHARED
	Kind AzureDiskPersistentVolumeKind `field:"optional" json:"kind" yaml:"kind"`
	// The volume name. Default: - auto-generated.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// Force the ReadOnly setting in VolumeMounts. Default: false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
}

// Options of `Volume.fromGcePersistentDisk`.
type GCEPersistentDiskVolumeOptions struct {
	// Filesystem type of the volume that you want to mount.
	//
	// Tip: Ensure that the filesystem type is supported by the host operating system. See: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
	//
	// Default: 'ext4'.
	FsType *string `field:"optional" json:"fsType" yaml:"fsType"`
	// The volume name. Default: - auto-generated.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// The partition in the volume that you want to mount.
	//
	// If omitted, the default is to mount by volume name. Examples: For volume /dev/sda1, you specify the partition as "1". Similarly, the volume partition for /dev/sda is "0" (or you can leave the property empty). Default: - No partition.
	Partition *float64 `field:"optional" json:"partition" yaml:"partition"`
	// Specify "true" to force and set the ReadOnly property in VolumeMounts to "true". See: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
	//
	// Default: false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
}

// Options for a HostPathVolume-based volume.
type HostPathVolumeOptions struct {
	// The path of the directory on the host.
	Path *string `field:"required" json:"path" yaml:"path"`
	// The expected type of the path found on the host. Default: HostPathVolumeType.DEFAULT
	Type HostPathVolumeType `field:"optional" json:"type" yaml:"type"`
}

// Options for the NFS based volume.
type NfsVolumeOptions struct {
	// Path that is exported by the NFS server.
	Path *string `field:"required" json:"path" yaml:"path"`
	// Server is the hostname or IP address of the NFS server.
	Server *string `field:"required" json:"server" yaml:"server"`
	// If set to true, will force the NFS export to be mounted with read-only permissions. Default: - false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
}

// Options for the CSI driver based volume.
type CsiVolumeOptions struct {
	// Any driver-specific attributes to pass to the CSI volume builder. Default: - undefined.
	Attributes *map[string]*string `field:"optional" json:"attributes" yaml:"attributes"`
	// The filesystem type to mount.
	//
	// Ex. "ext4", "xfs", "ntfs". If not provided, the empty value is passed to the associated CSI driver, which will determine the default filesystem to apply. Default: - driver-dependent.
	FsType *string `field:"optional" json:"fsType" yaml:"fsType"`
	// The volume name. Default: - auto-generated.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// Whether the mounted volume should be read-only or not. Default: - false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
}

// Options for the Secret-based volume.
type SecretVolumeOptions struct {
	// Mode bits to use on created files by default.
	//
	// Must be a value between 0 and 0777. Defaults to 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set. Default: 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.
	DefaultMode *float64 `field:"optional" json:"defaultMode" yaml:"defaultMode"`
	// If unspecified, each key-value pair in the Data field of the referenced secret will be projected into the volume as a file whose name is the key and content is the value.
	//
	// If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the secret, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the '..' path or start with '..'. Default: - no mapping.
	Items *map[string]*PathMapping `field:"optional" json:"items" yaml:"items"`
	// The volume name. Default: - auto-generated.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// Specify whether the secret or its keys must be defined. Default: - undocumented.
	Optional *bool `field:"optional" json:"optional" yaml:"optional"`
}

// Options for a PersistentVolumeClaim-based volume.
type PersistentVolumeClaimVolumeOptions struct {
	// The volume name. Default: - Derived from the PVC name.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// Will force the ReadOnly setting in VolumeMounts. Default: false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
}

// Azure Disk kinds.
type AzureDiskPersistentVolumeKind string

const (
	// Multiple blob disks per storage account.
	AzureDiskPersistentVolumeKind_SHARED AzureDiskPersistentVolumeKind = "SHARED"
	// Single blob disk per storage account.
	AzureDiskPersistentVolumeKind_DEDICATED AzureDiskPersistentVolumeKind = "DEDICATED"
	// Azure managed data disk.
	AzureDiskPersistentVolumeKind_MANAGED AzureDiskPersistentVolumeKind = "MANAGED"
)

// Azure disk caching modes.
type AzureDiskPersistentVolumeCachingMode string

const (
	// None.
	AzureDiskPersistentVolumeCachingMode_NONE AzureDiskPersistentVolumeCachingMode = "NONE"
	// ReadOnly.
	AzureDiskPersistentVolumeCachingMode_READ_ONLY AzureDiskPersistentVolumeCachingMode = "READ_ONLY"
	// ReadWrite.
	AzureDiskPersistentVolumeCachingMode_READ_WRITE AzureDiskPersistentVolumeCachingMode = "READ_WRITE"
)

// Host path types.
type HostPathVolumeType string

const (
	// Empty string (default) is for backward compatibility, which means that no checks will be performed before mounting the hostPath volume.
	HostPathVolumeType_DEFAULT HostPathVolumeType = "DEFAULT"
	// If nothing exists at the given path, an empty directory will be created there as needed with permission set to 0755, having the same group and ownership with Kubelet.
	HostPathVolumeType_DIRECTORY_OR_CREATE HostPathVolumeType = "DIRECTORY_OR_CREATE"
	// A directory must exist at the given path.
	HostPathVolumeType_DIRECTORY HostPathVolumeType = "DIRECTORY"
	// If nothing exists at the given path, an empty file will be created there as needed with permission set to 0644, having the same group and ownership with Kubelet.
	HostPathVolumeType_FILE_OR_CREATE HostPathVolumeType = "FILE_OR_CREATE"
	// A file must exist at the given path.
	HostPathVolumeType_FILE HostPathVolumeType = "FILE"
	// A UNIX socket must exist at the given path.
	HostPathVolumeType_SOCKET HostPathVolumeType = "SOCKET"
	// A character device must exist at the given path.
	HostPathVolumeType_CHAR_DEVICE HostPathVolumeType = "CHAR_DEVICE"
	// A block device must exist at the given path.
	HostPathVolumeType_BLOCK_DEVICE HostPathVolumeType = "BLOCK_DEVICE"
)

// Represents a piece of storage in the cluster.
type IStorage interface {
	constructs.IConstruct
	// Convert the piece of storage into a concrete volume.
	AsVolume() Volume
}

// Volume represents a named volume in a pod that may be accessed by any container in the pod.
//
// Docker also has a concept of volumes, though it is somewhat looser and less managed. In Docker, a volume is simply a directory on disk or in another Container. Lifetimes are not managed and until very recently there were only local-disk-backed volumes. Docker now provides volume drivers, but the functionality is very limited for now (e.g. as of Docker 1.7 only one volume driver is allowed per Container and there is no way to pass parameters to volumes).
//
// A Kubernetes volume, on the other hand, has an explicit lifetime - the same as the Pod that encloses it. Consequently, a volume outlives any Containers that run within the Pod, and data is preserved across Container restarts. Of course, when a Pod ceases to exist, the volume will cease to exist, too. Perhaps more importantly than this, Kubernetes supports many types of volumes, and a Pod can use any number of them simultaneously.
//
// At its core, a volume is just a directory, possibly with some data in it, which is accessible to the Containers in a Pod. How that directory comes to be, the medium that backs it, and the contents of it are determined by the particular volume type used.
//
// To use a volume, a Pod specifies what volumes to provide for the Pod (the .spec.volumes field) and where to mount those into Containers (the .spec.containers[*].volumeMounts field).
//
// A process in a container sees a filesystem view composed from their Docker image and volumes. The Docker image is at the root of the filesystem hierarchy, and any volumes are mounted at the specified paths within the image. Volumes can not mount onto other volumes
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

// Options for the ConfigMap-based volume.
type ConfigMapVolumeOptions struct {
	// Mode bits to use on created files by default.
	//
	// Must be a value between 0 and 0777. Defaults to 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set. Default: 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.
	DefaultMode *float64 `field:"optional" json:"defaultMode" yaml:"defaultMode"`
	// If unspecified, each key-value pair in the Data field of the referenced ConfigMap will be projected into the volume as a file whose name is the key and content is the value.
	//
	// If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the ConfigMap, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the '..' path or start with '..'. Default: - no mapping.
	Items *map[string]*PathMapping `field:"optional" json:"items" yaml:"items"`
	// The volume name. Default: - auto-generated.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// Specify whether the ConfigMap or its keys must be defined. Default: - undocumented.
	Optional *bool `field:"optional" json:"optional" yaml:"optional"`
}

// The medium on which to store the volume.
type EmptyDirMedium string

const (
	// The default volume of the backing node.
	EmptyDirMedium_DEFAULT EmptyDirMedium = "DEFAULT"
	// Mount a tmpfs (RAM-backed filesystem) for you instead.
	//
	// While tmpfs is very fast, be aware that unlike disks, tmpfs is cleared on node reboot and any files you write will count against your Container's memory limit.
	EmptyDirMedium_MEMORY EmptyDirMedium = "MEMORY"
)

func emptyDirMediumManifestValue(value EmptyDirMedium) string {
	switch value {
	case EmptyDirMedium_DEFAULT:
		return ""
	case EmptyDirMedium_MEMORY:
		return "Memory"
	default:
		return string(value)
	}
}

// Options for volumes populated with an empty directory.
type EmptyDirVolumeOptions struct {
	// By default, emptyDir volumes are stored on whatever medium is backing the node - that might be disk or SSD or network storage, depending on your environment.
	//
	// However, you can set the emptyDir.medium field to `EmptyDirMedium.MEMORY` to tell Kubernetes to mount a tmpfs (RAM-backed filesystem) for you instead. While tmpfs is very fast, be aware that unlike disks, tmpfs is cleared on node reboot and any files you write will count against your Container's memory limit. Default: EmptyDirMedium.DEFAULT
	Medium EmptyDirMedium `field:"optional" json:"medium" yaml:"medium"`
	// Total amount of local storage required for this EmptyDir volume.
	//
	// The size limit is also applicable for memory medium. The maximum usage on memory medium EmptyDir would be the minimum value between the SizeLimit specified here and the sum of memory limits of all containers in a pod. Default: - limit is undefined.
	SizeLimit cdk8s.Size `field:"optional" json:"sizeLimit" yaml:"sizeLimit"`
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

// Populate the volume from a ConfigMap.
//
// The configMap resource provides a way to inject configuration data into Pods. The data stored in a ConfigMap object can be referenced in a volume of type configMap and then consumed by containerized applications running in a Pod.
//
// When referencing a configMap object, you can simply provide its name in the volume to reference it. You can also customize the path to use for a specific entry in the ConfigMap.
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
	if options != nil && options.Items != nil {
		config["items"] = volumePathMappings(options.Items)
	}
	return newVolume(scope, id, name, map[string]interface{}{"configMap": config})
}

// An emptyDir volume is first created when a Pod is assigned to a Node, and exists as long as that Pod is running on that node.
//
// As the name says, it is initially empty. Containers in the Pod can all read and write the same files in the emptyDir volume, though that volume can be mounted at the same or different paths in each Container. When a Pod is removed from a node for any reason, the data in the emptyDir is deleted forever. See: http://kubernetes.io/docs/user-guide/volumes#emptydir
func Volume_FromEmptyDir(scope constructs.Construct, id, name *string, options *EmptyDirVolumeOptions) Volume {
	spec := map[string]interface{}{"emptyDir": map[string]interface{}{}}
	if options != nil && options.Medium != "" {
		spec["emptyDir"].(map[string]interface{})["medium"] = emptyDirMediumManifestValue(options.Medium)
	}
	if options != nil && options.SizeLimit != nil {
		amount := options.SizeLimit.ToMebibytes(nil)
		spec["emptyDir"].(map[string]interface{})["sizeLimit"] = strconv.FormatFloat(*amount, 'f', -1, 64) + "Mi"
	}
	return newVolume(scope, id, name, spec)
}

// Create a volume with an arbitrary name and no configuration.
func Volume_FromName(scope constructs.Construct, id, name *string) Volume {
	return newVolume(scope, id, name, map[string]interface{}{})
}

// Mounts an Amazon Web Services (AWS) EBS volume into your pod.
//
// Unlike emptyDir, which is erased when a pod is removed, the contents of an EBS volume are persisted and the volume is unmounted. This means that an EBS volume can be pre-populated with data, and that data can be shared between pods.
//
// There are some restrictions when using an awsElasticBlockStore volume:
//
// - the nodes on which pods are running must be AWS EC2 instances. - those instances need to be in the same region and availability zone as the EBS volume. - EBS only supports a single EC2 instance mounting a volume.
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

// Mounts a Microsoft Azure Data Disk into a pod.
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

// Mounts a Google Compute Engine (GCE) persistent disk (PD) into your Pod.
//
// Unlike emptyDir, which is erased when a pod is removed, the contents of a PD are preserved and the volume is merely unmounted. This means that a PD can be pre-populated with data, and that data can be shared between pods.
//
// There are some restrictions when using a gcePersistentDisk:
//
// - the nodes on which Pods are running must be GCE VMs - those VMs need to be in the same GCE project and zone as the persistent disk.
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

// Populate the volume from a Secret.
//
// A secret volume is used to pass sensitive information, such as passwords, to Pods. You can store secrets in the Kubernetes API and mount them as files for use by pods without coupling to Kubernetes directly.
//
// secret volumes are backed by tmpfs (a RAM-backed filesystem) so they are never written to non-volatile storage. See: https://kubernetes.io/docs/concepts/storage/volumes/#secret
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
		if options.Items != nil {
			spec["items"] = volumePathMappings(options.Items)
		}
	}
	return newVolume(scope, id, name, map[string]interface{}{"secret": spec})
}

// Used to mount a PersistentVolume into a Pod.
//
// PersistentVolumeClaims are a way for users to "claim" durable storage (such as a GCE PersistentDisk or an iSCSI volume) without knowing the details of the particular cloud environment. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes/
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

// Used to mount a file or directory from the host node's filesystem into a Pod.
//
// This is not something that most Pods will need, but it offers a powerful escape hatch for some applications. See: https://kubernetes.io/docs/concepts/storage/volumes/#hostpath
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

// Used to mount an NFS share into a Pod. See: https://kubernetes.io/docs/concepts/storage/volumes/#nfs
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

// Populate the volume from a CSI driver, for example the Secrets Store CSI Driver: https://secrets-store-csi-driver.sigs.k8s.io/introduction.html. Which in turn needs an associated provider to source the secrets, such as the AWS Secrets Manager and Systems Manager Parameter Store provider: https://aws.github.io/secrets-store-csi-driver-provider-aws/.
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

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Volume_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}
