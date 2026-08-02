package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

type (
	// Contract of a `PersistentVolumeClaim`.
	IPersistentVolume interface{ IResource }
	// Contract of a `PersistentVolumeClaim`.
	IPersistentVolumeClaim interface{ IResource }
)

// Access Modes.
type PersistentVolumeAccessMode string

const (
	// The volume can be mounted as read-write by a single node.
	//
	// ReadWriteOnce access mode still can allow multiple pods to access the volume when the pods are running on the same node.
	PersistentVolumeAccessMode_READ_WRITE_ONCE PersistentVolumeAccessMode = "READ_WRITE_ONCE"
	// The volume can be mounted as read-only by many nodes.
	PersistentVolumeAccessMode_READ_ONLY_MANY PersistentVolumeAccessMode = "READ_ONLY_MANY"
	// The volume can be mounted as read-write by many nodes.
	PersistentVolumeAccessMode_READ_WRITE_MANY PersistentVolumeAccessMode = "READ_WRITE_MANY"
	// The volume can be mounted as read-write by a single Pod.
	//
	// Use ReadWriteOncePod access mode if you want to ensure that only one pod across whole cluster can read that PVC or write to it. This is only supported for CSI volumes and Kubernetes version 1.22+.
	PersistentVolumeAccessMode_READ_WRITE_ONCE_POD PersistentVolumeAccessMode = "READ_WRITE_ONCE_POD"
)

// Volume Modes.
type PersistentVolumeMode string

const (
	// Volume is ounted into Pods into a directory.
	//
	// If the volume is backed by a block device and the device is empty, Kubernetes creates a filesystem on the device before mounting it for the first time.
	PersistentVolumeMode_FILE_SYSTEM PersistentVolumeMode = "FILE_SYSTEM"
	// Use a volume as a raw block device.
	//
	// Such volume is presented into a Pod as a block device, without any filesystem on it. This mode is useful to provide a Pod the fastest possible way to access a volume, without any filesystem layer between the Pod and the volume. On the other hand, the application running in the Pod must know how to handle a raw block device.
	PersistentVolumeMode_BLOCK PersistentVolumeMode = "BLOCK"
)

// Reclaim Policies.
type PersistentVolumeReclaimPolicy string

const (
	// The Retain reclaim policy allows for manual reclamation of the resource.
	//
	// When the PersistentVolumeClaim is deleted, the PersistentVolume still exists and the volume is considered "released". But it is not yet available for another claim because the previous claimant's data remains on the volume. An administrator can manually reclaim the volume with the following steps:
	//
	//  1. Delete the PersistentVolume. The associated storage asset in external infrastructure (such as an AWS EBS, GCE PD, Azure Disk, or Cinder volume) still exists after the PV is deleted.
	//  2. Manually clean up the data on the associated storage asset accordingly.
	//  3. Manually delete the associated storage asset.
	//
	// If you want to reuse the same storage asset, create a new PersistentVolume with the same storage asset definition.
	PersistentVolumeReclaimPolicy_RETAIN PersistentVolumeReclaimPolicy = "RETAIN"
	// For volume plugins that support the Delete reclaim policy, deletion removes both the PersistentVolume object from Kubernetes, as well as the associated storage asset in the external infrastructure, such as an AWS EBS, GCE PD, Azure Disk, or Cinder volume.
	//
	// Volumes that were dynamically provisioned inherit the reclaim policy of their StorageClass, which defaults to Delete. The administrator should configure the StorageClass according to users' expectations; otherwise, the PV must be edited or patched after it is created.
	PersistentVolumeReclaimPolicy_DELETE PersistentVolumeReclaimPolicy = "DELETE"
)

// Properties for `PersistentVolume`.
type PersistentVolumeProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Contains all ways the volume can be mounted. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes
	//
	// Default: - No access modes.
	AccessModes *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	// Part of a bi-directional binding between PersistentVolume and PersistentVolumeClaim.
	//
	// Expected to be non-nil when bound. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#binding
	//
	// Default: - Not bound to a specific claim.
	Claim IPersistentVolumeClaim `field:"optional" json:"claim" yaml:"claim"`
	// A list of mount options, e.g. ["ro", "soft"]. Not validated - mount will simply fail if one is invalid. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#mount-options
	//
	// Default: - No options.
	MountOptions *[]*string `field:"optional" json:"mountOptions" yaml:"mountOptions"`
	// When a user is done with their volume, they can delete the PVC objects from the API that allows reclamation of the resource.
	//
	// The reclaim policy tells the cluster what to do with the volume after it has been released of its claim. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#reclaiming
	//
	// Default: PersistentVolumeReclaimPolicy.RETAIN
	ReclaimPolicy PersistentVolumeReclaimPolicy `field:"optional" json:"reclaimPolicy" yaml:"reclaimPolicy"`
	// What is the storage capacity of this volume. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources
	//
	// Default: - No specified.
	Storage cdk8s.Size `field:"optional" json:"storage" yaml:"storage"`
	// Name of StorageClass to which this persistent volume belongs. Default: - Volume does not belong to any storage class.
	StorageClassName *string `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	// Defines what type of volume is required by the claim. Default: VolumeMode.FILE_SYSTEM
	VolumeMode PersistentVolumeMode `field:"optional" json:"volumeMode" yaml:"volumeMode"`
}

// Properties for `PersistentVolumeClaim`.
type PersistentVolumeClaimProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Contains the access modes the volume should support. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1
	//
	// Default: - No access modes requirement.
	AccessModes *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	// Minimum storage size the volume should have. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources
	//
	// Default: - No storage requirement.
	Storage cdk8s.Size `field:"optional" json:"storage" yaml:"storage"`
	// Name of the StorageClass required by the claim. When this property is not set, the behavior is as follows:.
	//
	// - If the admission plugin is turned on, the storage class marked as default will be used. - If the admission plugin is turned off, the pvc can only be bound to volumes without a storage class. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	//
	// Default: - Not set.
	StorageClassName *string `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	// The PersistentVolume backing this claim.
	//
	// The control plane still checks that storage class, access modes, and requested storage size on the volume are valid.
	//
	// Note that in order to guarantee a proper binding, the volume should also define a `claimRef` referring to this claim. Otherwise, the volume may be claimed be other pvc's before it gets a chance to bind to this one.
	//
	// If the volume is managed (i.e not imported), you can use `pv.claim()` to easily create a bi-directional bounded claim. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#binding.
	//
	// Default: - No specific volume binding.
	Volume IPersistentVolume `field:"optional" json:"volume" yaml:"volume"`
	// Defines what type of volume is required by the claim. Default: VolumeMode.FILE_SYSTEM
	VolumeMode PersistentVolumeMode `field:"optional" json:"volumeMode" yaml:"volumeMode"`
}

// A PersistentVolumeClaim template for StatefulSets.
type PersistentVolumeClaimTemplateProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Contains the access modes the volume should support. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1
	//
	// Default: - No access modes requirement.
	AccessModes *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	// Minimum storage size the volume should have. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources
	//
	// Default: - No storage requirement.
	Storage cdk8s.Size `field:"optional" json:"storage" yaml:"storage"`
	// Name of the StorageClass required by the claim. When this property is not set, the behavior is as follows:.
	//
	// - If the admission plugin is turned on, the storage class marked as default will be used. - If the admission plugin is turned off, the pvc can only be bound to volumes without a storage class. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	//
	// Default: - Not set.
	StorageClassName *string `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	// The PersistentVolume backing this claim.
	//
	// The control plane still checks that storage class, access modes, and requested storage size on the volume are valid.
	//
	// Note that in order to guarantee a proper binding, the volume should also define a `claimRef` referring to this claim. Otherwise, the volume may be claimed be other pvc's before it gets a chance to bind to this one.
	//
	// If the volume is managed (i.e not imported), you can use `pv.claim()` to easily create a bi-directional bounded claim. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#binding.
	//
	// Default: - No specific volume binding.
	Volume IPersistentVolume `field:"optional" json:"volume" yaml:"volume"`
	// Defines what type of volume is required by the claim. Default: VolumeMode.FILE_SYSTEM
	VolumeMode PersistentVolumeMode `field:"optional" json:"volumeMode" yaml:"volumeMode"`
	// The name of the claim that the StatefulSet controller will create for each pod.
	//
	// This will be used to name the created PVC in the format <claim-name>-<pod-name>
	//
	// This name should match the name of a volume mount in one of the containers.
	Name *string `field:"required" json:"name" yaml:"name"`
}

// A PersistentVolume (PV) is a piece of storage in the cluster that has been provisioned by an administrator or dynamically provisioned using Storage Classes.
//
// It is a resource in the cluster just like a node is a cluster resource. PVs are volume plugins like Volumes, but have a lifecycle independent of any individual Pod that uses the PV. This API object captures the details of the implementation of the storage, be that NFS, iSCSI, or a cloud-provider-specific storage system.
type PersistentVolume interface {
	Resource
	IPersistentVolume
	IStorage
	// Access modes requirement of this claim.
	AccessModes() *[]PersistentVolumeAccessMode
	// PVC this volume is bound to.
	//
	// Undefined means this volume is not yet claimed by any PVC.
	Claim() IPersistentVolumeClaim
	// Volume mode of this volume.
	Mode() PersistentVolumeMode
	// Mount options of this volume.
	MountOptions() *[]*string
	// Reclaim policy of this volume.
	ReclaimPolicy() PersistentVolumeReclaimPolicy
	// Storage size of this volume.
	Storage() cdk8s.Size
	// Storage class this volume belongs to.
	StorageClassName() *string
	// Bind a volume to a specific claim.
	//
	// Note that you must also bind the claim to the volume. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#binding
	Bind(claim IPersistentVolumeClaim)
	// Reserve a `PersistentVolume` by creating a `PersistentVolumeClaim` that is wired to claim this volume.
	//
	// Note that this method will throw in case the volume is already claimed. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#reserving-a-persistentvolume
	Reserve() PersistentVolumeClaim
}

type persistentVolumeImpl struct {
	resourceBase
	accessModes                                 []PersistentVolumeAccessMode
	claim                                       IPersistentVolumeClaim
	mode                                        PersistentVolumeMode
	mountOptions                                []*string
	reclaimPolicy                               PersistentVolumeReclaimPolicy
	storage                                     cdk8s.Size
	storageClassName                            *string
	source                                      map[string]interface{}
	volumeID, fsType, diskName, diskURI, pdName *string
	partition                                   *float64
	readOnly                                    *bool
	cachingMode                                 AzureDiskPersistentVolumeCachingMode
	azureKind                                   AzureDiskPersistentVolumeKind
}

func NewPersistentVolume(scope constructs.Construct, id *string, props *PersistentVolumeProps) PersistentVolume {
	return newPersistentVolume(scope, id, props, nil)
}

func newPersistentVolume(scope constructs.Construct, id *string, props *PersistentVolumeProps, source map[string]interface{}) *persistentVolumeImpl {
	if props == nil {
		props = &PersistentVolumeProps{}
	}
	result := &persistentVolumeImpl{mode: props.VolumeMode, reclaimPolicy: props.ReclaimPolicy, storage: props.Storage, storageClassName: props.StorageClassName, source: source}
	if result.mode == "" {
		result.mode = PersistentVolumeMode_FILE_SYSTEM
	}
	if result.reclaimPolicy == "" {
		result.reclaimPolicy = PersistentVolumeReclaimPolicy_RETAIN
	}
	if props.AccessModes != nil {
		result.accessModes = append(result.accessModes, (*props.AccessModes)...)
	}
	if props.MountOptions != nil {
		result.mountOptions = append(result.mountOptions, (*props.MountOptions)...)
	}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "v1", "PersistentVolume", "persistentvolumes", props.Metadata, manifest)
	if props.Claim != nil {
		result.Bind(props.Claim)
	}
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toManifest() }})
	return result
}

func NewPersistentVolume_Override(volume PersistentVolume, scope constructs.Construct, id *string, props *PersistentVolumeProps) {
	applyOverride(volume, NewPersistentVolume(scope, id, props), "PersistentVolume")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func PersistentVolume_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (p *persistentVolumeImpl) AccessModes() *[]PersistentVolumeAccessMode {
	if len(p.accessModes) == 0 {
		return nil
	}
	values := append([]PersistentVolumeAccessMode(nil), p.accessModes...)
	return &values
}

func (p *persistentVolumeImpl) Claim() IPersistentVolumeClaim {
	return p.claim
}

func (p *persistentVolumeImpl) Mode() PersistentVolumeMode {
	return p.mode
}

func (p *persistentVolumeImpl) MountOptions() *[]*string {
	if len(p.mountOptions) == 0 {
		return nil
	}
	values := append([]*string(nil), p.mountOptions...)
	return &values
}

func (p *persistentVolumeImpl) ReclaimPolicy() PersistentVolumeReclaimPolicy {
	return p.reclaimPolicy
}

func (p *persistentVolumeImpl) Storage() cdk8s.Size {
	return p.storage
}

func (p *persistentVolumeImpl) StorageClassName() *string {
	return p.storageClassName
}

func (p *persistentVolumeImpl) Bind(claim IPersistentVolumeClaim) {
	if claim == nil {
		panic("claim is required")
	}
	if p.claim != nil && p.claim.Name() != claim.Name() {
		panic("Cannot bind volume '" + stringValue(p.Name()) + "' to claim '" + stringValue(claim.Name()) + "' since it is already bound to claim '" + stringValue(p.claim.Name()) + "'")
	}
	p.claim = claim
}

func (p *persistentVolumeImpl) Reserve() PersistentVolumeClaim {
	claim := NewPersistentVolumeClaim(p, jsii.String(stringValue(p.Name())+"PVC"), &PersistentVolumeClaimProps{Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String("pvc-" + stringValue(p.Name())), Namespace: p.Metadata().Namespace()}, StorageClassName: p.storageClassName})
	p.Bind(claim)
	claim.Bind(p)
	return claim
}

func (p *persistentVolumeImpl) AsVolume() Volume {
	return Volume_FromPersistentVolumeClaim(p, jsii.String("Volume"), p.Reserve(), nil)
}

func (p *persistentVolumeImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{"persistentVolumeReclaimPolicy": persistentVolumeReclaimPolicyManifest(p.reclaimPolicy), "volumeMode": persistentVolumeModeManifest(p.mode)}
	if len(p.accessModes) > 0 {
		result["accessModes"] = persistentVolumeAccessModesManifest(p.accessModes)
	}
	if p.claim != nil {
		result["claimRef"] = map[string]interface{}{"name": p.claim.Name()}
	}
	if p.storage != nil {
		result["capacity"] = map[string]interface{}{"storage": sizeGibibytes(p.storage)}
	}
	if len(p.mountOptions) > 0 {
		result["mountOptions"] = p.mountOptions
	}
	if p.storageClassName != nil {
		result["storageClassName"] = p.storageClassName
	}
	for key, value := range p.source {
		result[key] = value
	}
	return result
}

// A PersistentVolumeClaim (PVC) is a request for storage by a user.
//
// It is similar to a Pod. Pods consume node resources and PVCs consume PV resources. Pods can request specific levels of resources (CPU and Memory). Claims can request specific size and access modes.
type PersistentVolumeClaim interface {
	Resource
	IPersistentVolumeClaim
	// Access modes requirement of this claim.
	AccessModes() *[]PersistentVolumeAccessMode
	// Storage requirement of this claim.
	Storage() cdk8s.Size
	// Storage class requirment of this claim.
	StorageClassName() *string
	// PV this claim is bound to.
	//
	// Undefined means the claim is not bound to any specific volume.
	Volume() IPersistentVolume
	// Volume mode requirement of this claim.
	VolumeMode() PersistentVolumeMode
	// Bind a claim to a specific volume.
	//
	// Note that you must also bind the volume to the claim. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#binding
	Bind(volume IPersistentVolume)
}

type persistentVolumeClaimImpl struct {
	resourceBase
	accessModes      []PersistentVolumeAccessMode
	storage          cdk8s.Size
	storageClassName *string
	volume           IPersistentVolume
	volumeMode       PersistentVolumeMode
}

func NewPersistentVolumeClaim(scope constructs.Construct, id *string, props *PersistentVolumeClaimProps) PersistentVolumeClaim {
	if props == nil {
		props = &PersistentVolumeClaimProps{}
	}
	result := &persistentVolumeClaimImpl{storage: props.Storage, storageClassName: props.StorageClassName, volumeMode: props.VolumeMode}
	if result.volumeMode == "" {
		result.volumeMode = PersistentVolumeMode_FILE_SYSTEM
	}
	if props.AccessModes != nil {
		result.accessModes = append(result.accessModes, (*props.AccessModes)...)
	}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "v1", "PersistentVolumeClaim", "persistentvolumeclaims", props.Metadata, manifest)
	if props.Volume != nil {
		result.Bind(props.Volume)
	}
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toManifest() }})
	return result
}

func NewPersistentVolumeClaim_Override(claim PersistentVolumeClaim, scope constructs.Construct, id *string, props *PersistentVolumeClaimProps) {
	applyOverride(claim, NewPersistentVolumeClaim(scope, id, props), "PersistentVolumeClaim")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func PersistentVolumeClaim_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (p *persistentVolumeClaimImpl) AccessModes() *[]PersistentVolumeAccessMode {
	if len(p.accessModes) == 0 {
		return nil
	}
	values := append([]PersistentVolumeAccessMode(nil), p.accessModes...)
	return &values
}

func (p *persistentVolumeClaimImpl) Storage() cdk8s.Size {
	return p.storage
}

func (p *persistentVolumeClaimImpl) StorageClassName() *string {
	return p.storageClassName
}

func (p *persistentVolumeClaimImpl) Volume() IPersistentVolume {
	return p.volume
}

func (p *persistentVolumeClaimImpl) VolumeMode() PersistentVolumeMode {
	return p.volumeMode
}

func (p *persistentVolumeClaimImpl) Bind(volume IPersistentVolume) {
	if volume == nil {
		panic("volume is required")
	}
	if p.volume != nil && p.volume.Name() != volume.Name() {
		panic("Cannot bind claim '" + stringValue(p.Name()) + "' to volume '" + stringValue(volume.Name()) + "' since it is already bound to volume '" + stringValue(p.volume.Name()) + "'")
	}
	p.volume = volume
}

func (p *persistentVolumeClaimImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{"volumeMode": persistentVolumeModeManifest(p.volumeMode)}
	if p.volume != nil {
		result["volumeName"] = p.volume.Name()
	}
	if len(p.accessModes) > 0 {
		result["accessModes"] = persistentVolumeAccessModesManifest(p.accessModes)
	}
	if p.storage != nil {
		result["resources"] = map[string]interface{}{"requests": map[string]interface{}{"storage": sizeGibibytes(p.storage)}}
	}
	if p.storageClassName != nil {
		result["storageClassName"] = p.storageClassName
	}
	return result
}

type importedPersistentVolume struct {
	node constructs.Node
	name *string
}

func (p *importedPersistentVolume) Node() constructs.Node {
	return p.node
}

func (p *importedPersistentVolume) SetNodeInternal(node constructs.Node) {
	p.node = node
}

func (p *importedPersistentVolume) ToString() *string {
	return p.node.Path()
}

func (p *importedPersistentVolume) With(m ...constructs.IMixin) constructs.IConstruct {
	return p.node.With(m...)
}

func (p *importedPersistentVolume) ApiVersion() *string {
	return jsii.String("v1")
}

func (p *importedPersistentVolume) ApiGroup() *string {
	return jsii.String("")
}

func (p *importedPersistentVolume) Kind() *string {
	return jsii.String("PersistentVolume")
}

func (p *importedPersistentVolume) Name() *string {
	return p.name
}

func (p *importedPersistentVolume) ResourceName() *string {
	return p.name
}

func (p *importedPersistentVolume) ResourceType() *string {
	return jsii.String("persistentvolumes")
}

// Imports a pv from the cluster as a reference.
func PersistentVolume_FromPersistentVolumeName(scope constructs.Construct, id, name *string) IPersistentVolume {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and volumeName are required")
	}
	result := &importedPersistentVolume{name: name}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

type importedPersistentVolumeClaim struct {
	node constructs.Node
	name *string
}

func (p *importedPersistentVolumeClaim) Node() constructs.Node {
	return p.node
}

func (p *importedPersistentVolumeClaim) SetNodeInternal(node constructs.Node) {
	p.node = node
}

func (p *importedPersistentVolumeClaim) ToString() *string {
	return p.node.Path()
}

func (p *importedPersistentVolumeClaim) With(m ...constructs.IMixin) constructs.IConstruct {
	return p.node.With(m...)
}

func (p *importedPersistentVolumeClaim) ApiVersion() *string {
	return jsii.String("v1")
}

func (p *importedPersistentVolumeClaim) ApiGroup() *string {
	return jsii.String("")
}

func (p *importedPersistentVolumeClaim) Kind() *string {
	return jsii.String("PersistentVolumeClaim")
}

func (p *importedPersistentVolumeClaim) Name() *string {
	return p.name
}

func (p *importedPersistentVolumeClaim) ResourceName() *string {
	return p.name
}

func (p *importedPersistentVolumeClaim) ResourceType() *string {
	return jsii.String("persistentvolumeclaims")
}

// Imports a pvc from the cluster as a reference.
func PersistentVolumeClaim_FromClaimName(scope constructs.Construct, id, name *string) IPersistentVolumeClaim {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and claimName are required")
	}
	result := &importedPersistentVolumeClaim{name: name}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

// Properties for `AwsElasticBlockStorePersistentVolume`.
type AwsElasticBlockStorePersistentVolumeProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Contains all ways the volume can be mounted. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes
	//
	// Default: - No access modes.
	AccessModes *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	// Part of a bi-directional binding between PersistentVolume and PersistentVolumeClaim.
	//
	// Expected to be non-nil when bound. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#binding
	//
	// Default: - Not bound to a specific claim.
	Claim IPersistentVolumeClaim `field:"optional" json:"claim" yaml:"claim"`
	// A list of mount options, e.g. ["ro", "soft"]. Not validated - mount will simply fail if one is invalid. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#mount-options
	//
	// Default: - No options.
	MountOptions *[]*string `field:"optional" json:"mountOptions" yaml:"mountOptions"`
	// When a user is done with their volume, they can delete the PVC objects from the API that allows reclamation of the resource.
	//
	// The reclaim policy tells the cluster what to do with the volume after it has been released of its claim. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#reclaiming
	//
	// Default: PersistentVolumeReclaimPolicy.RETAIN
	ReclaimPolicy PersistentVolumeReclaimPolicy `field:"optional" json:"reclaimPolicy" yaml:"reclaimPolicy"`
	// What is the storage capacity of this volume. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources
	//
	// Default: - No specified.
	Storage cdk8s.Size `field:"optional" json:"storage" yaml:"storage"`
	// Name of StorageClass to which this persistent volume belongs. Default: - Volume does not belong to any storage class.
	StorageClassName *string `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	// Defines what type of volume is required by the claim. Default: VolumeMode.FILE_SYSTEM
	VolumeMode PersistentVolumeMode `field:"optional" json:"volumeMode" yaml:"volumeMode"`
	// Unique ID of the persistent disk resource in AWS (Amazon EBS volume).
	//
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore See: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
	VolumeId *string `field:"required" json:"volumeId" yaml:"volumeId"`
	// Filesystem type of the volume that you want to mount.
	//
	// Tip: Ensure that the filesystem type is supported by the host operating system. See: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
	//
	// Default: 'ext4'.
	FsType *string `field:"optional" json:"fsType" yaml:"fsType"`
	// The partition in the volume that you want to mount.
	//
	// If omitted, the default is to mount by volume name. Examples: For volume /dev/sda1, you specify the partition as "1". Similarly, the volume partition for /dev/sda is "0" (or you can leave the property empty). Default: - No partition.
	Partition *float64 `field:"optional" json:"partition" yaml:"partition"`
	// Specify "true" to force and set the ReadOnly property in VolumeMounts to "true". See: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
	//
	// Default: false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
}

// Represents an AWS Disk resource that is attached to a kubelet's host machine and then exposed to the pod. See: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
type AwsElasticBlockStorePersistentVolume interface {
	PersistentVolume
	// File system type of this volume.
	FsType() *string
	// Partition of this volume.
	Partition() *float64
	// Whether or not it is mounted as a read-only volume.
	ReadOnly() *bool
	// Volume id of this volume.
	VolumeId() *string
}

func NewAwsElasticBlockStorePersistentVolume(scope constructs.Construct, id *string, props *AwsElasticBlockStorePersistentVolumeProps) AwsElasticBlockStorePersistentVolume {
	if props == nil || props.VolumeId == nil {
		panic("volumeId is required")
	}
	fs, ro := jsii.String("ext4"), jsii.Bool(false)
	if props.FsType != nil {
		fs = props.FsType
	}
	if props.ReadOnly != nil {
		ro = props.ReadOnly
	}
	result := newPersistentVolume(scope, id, persistentVolumePropsFromAws(props), map[string]interface{}{"awsElasticBlockStore": map[string]interface{}{"volumeID": props.VolumeId, "fsType": fs, "partition": props.Partition, "readOnly": ro}})
	result.volumeID, result.fsType, result.partition, result.readOnly = props.VolumeId, fs, props.Partition, ro
	return result
}

func NewAwsElasticBlockStorePersistentVolume_Override(volume AwsElasticBlockStorePersistentVolume, scope constructs.Construct, id *string, props *AwsElasticBlockStorePersistentVolumeProps) {
	applyOverride(volume, NewAwsElasticBlockStorePersistentVolume(scope, id, props), "AwsElasticBlockStorePersistentVolume")
}

// Imports a pv from the cluster as a reference.
func AwsElasticBlockStorePersistentVolume_FromPersistentVolumeName(scope constructs.Construct, id, name *string) IPersistentVolume {
	return PersistentVolume_FromPersistentVolumeName(scope, id, name)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func AwsElasticBlockStorePersistentVolume_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (p *persistentVolumeImpl) FsType() *string {
	return p.fsType
}

func (p *persistentVolumeImpl) Partition() *float64 {
	return p.partition
}

func (p *persistentVolumeImpl) ReadOnly() *bool {
	return p.readOnly
}

func (p *persistentVolumeImpl) VolumeId() *string {
	return p.volumeID
}

func persistentVolumePropsFromAws(p *AwsElasticBlockStorePersistentVolumeProps) *PersistentVolumeProps {
	return &PersistentVolumeProps{Metadata: p.Metadata, AccessModes: p.AccessModes, Claim: p.Claim, MountOptions: p.MountOptions, ReclaimPolicy: p.ReclaimPolicy, Storage: p.Storage, StorageClassName: p.StorageClassName, VolumeMode: p.VolumeMode}
}

// Properties for `AzureDiskPersistentVolume`.
type AzureDiskPersistentVolumeProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Contains all ways the volume can be mounted. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes
	//
	// Default: - No access modes.
	AccessModes *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	// Part of a bi-directional binding between PersistentVolume and PersistentVolumeClaim.
	//
	// Expected to be non-nil when bound. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#binding
	//
	// Default: - Not bound to a specific claim.
	Claim IPersistentVolumeClaim `field:"optional" json:"claim" yaml:"claim"`
	// A list of mount options, e.g. ["ro", "soft"]. Not validated - mount will simply fail if one is invalid. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#mount-options
	//
	// Default: - No options.
	MountOptions *[]*string `field:"optional" json:"mountOptions" yaml:"mountOptions"`
	// When a user is done with their volume, they can delete the PVC objects from the API that allows reclamation of the resource.
	//
	// The reclaim policy tells the cluster what to do with the volume after it has been released of its claim. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#reclaiming
	//
	// Default: PersistentVolumeReclaimPolicy.RETAIN
	ReclaimPolicy PersistentVolumeReclaimPolicy `field:"optional" json:"reclaimPolicy" yaml:"reclaimPolicy"`
	// What is the storage capacity of this volume. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources
	//
	// Default: - No specified.
	Storage cdk8s.Size `field:"optional" json:"storage" yaml:"storage"`
	// Name of StorageClass to which this persistent volume belongs. Default: - Volume does not belong to any storage class.
	StorageClassName *string `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	// Defines what type of volume is required by the claim. Default: VolumeMode.FILE_SYSTEM
	VolumeMode PersistentVolumeMode `field:"optional" json:"volumeMode" yaml:"volumeMode"`
	// The Name of the data disk in the blob storage.
	DiskName *string `field:"required" json:"diskName" yaml:"diskName"`
	// The URI the data disk in the blob storage.
	DiskUri *string `field:"required" json:"diskUri" yaml:"diskUri"`
	// Host Caching mode. Default: - AzureDiskPersistentVolumeCachingMode.NONE.
	CachingMode AzureDiskPersistentVolumeCachingMode `field:"optional" json:"cachingMode" yaml:"cachingMode"`
	// Filesystem type to mount.
	//
	// Must be a filesystem type supported by the host operating system. Default: 'ext4'.
	FsType *string `field:"optional" json:"fsType" yaml:"fsType"`
	// Kind of disk. Default: AzureDiskPersistentVolumeKind.SHARED
	Kind AzureDiskPersistentVolumeKind `field:"optional" json:"kind" yaml:"kind"`
	// Force the ReadOnly setting in VolumeMounts. Default: false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
}

// AzureDisk represents an Azure Data Disk mount on the host and bind mount to the pod.
type AzureDiskPersistentVolume interface {
	PersistentVolume
	// Azure kind of this volume.
	AzureKind() AzureDiskPersistentVolumeKind
	// Caching mode of this volume.
	CachingMode() AzureDiskPersistentVolumeCachingMode
	// Disk name of this volume.
	DiskName() *string
	// Disk URI of this volume.
	DiskUri() *string
	// File system type of this volume.
	FsType() *string
	// Whether or not it is mounted as a read-only volume.
	ReadOnly() *bool
}

func NewAzureDiskPersistentVolume(scope constructs.Construct, id *string, props *AzureDiskPersistentVolumeProps) AzureDiskPersistentVolume {
	if props == nil || props.DiskName == nil || props.DiskUri == nil {
		panic("diskName and diskUri are required")
	}
	fs, ro := jsii.String("ext4"), jsii.Bool(false)
	cache, kind := AzureDiskPersistentVolumeCachingMode_NONE, AzureDiskPersistentVolumeKind_SHARED
	if props.FsType != nil {
		fs = props.FsType
	}
	if props.ReadOnly != nil {
		ro = props.ReadOnly
	}
	if props.CachingMode != "" {
		cache = props.CachingMode
	}
	if props.Kind != "" {
		kind = props.Kind
	}
	result := newPersistentVolume(scope, id, persistentVolumePropsFromAzure(props), map[string]interface{}{"azureDisk": map[string]interface{}{"diskName": props.DiskName, "diskURI": props.DiskUri, "cachingMode": azureCachingModeManifest(cache), "fsType": fs, "kind": azureKindManifest(kind), "readOnly": ro}})
	result.diskName, result.diskURI, result.fsType, result.readOnly, result.cachingMode, result.azureKind = props.DiskName, props.DiskUri, fs, ro, cache, kind
	return result
}

func NewAzureDiskPersistentVolume_Override(volume AzureDiskPersistentVolume, scope constructs.Construct, id *string, props *AzureDiskPersistentVolumeProps) {
	applyOverride(volume, NewAzureDiskPersistentVolume(scope, id, props), "AzureDiskPersistentVolume")
}

// Imports a pv from the cluster as a reference.
func AzureDiskPersistentVolume_FromPersistentVolumeName(scope constructs.Construct, id, name *string) IPersistentVolume {
	return PersistentVolume_FromPersistentVolumeName(scope, id, name)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func AzureDiskPersistentVolume_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (p *persistentVolumeImpl) AzureKind() AzureDiskPersistentVolumeKind {
	return p.azureKind
}

func (p *persistentVolumeImpl) CachingMode() AzureDiskPersistentVolumeCachingMode {
	return p.cachingMode
}

func (p *persistentVolumeImpl) DiskName() *string {
	return p.diskName
}

func (p *persistentVolumeImpl) DiskUri() *string {
	return p.diskURI
}

func persistentVolumePropsFromAzure(p *AzureDiskPersistentVolumeProps) *PersistentVolumeProps {
	return &PersistentVolumeProps{Metadata: p.Metadata, AccessModes: p.AccessModes, Claim: p.Claim, MountOptions: p.MountOptions, ReclaimPolicy: p.ReclaimPolicy, Storage: p.Storage, StorageClassName: p.StorageClassName, VolumeMode: p.VolumeMode}
}

// Properties for `GCEPersistentDiskPersistentVolume`.
type GCEPersistentDiskPersistentVolumeProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Contains all ways the volume can be mounted. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes
	//
	// Default: - No access modes.
	AccessModes *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	// Part of a bi-directional binding between PersistentVolume and PersistentVolumeClaim.
	//
	// Expected to be non-nil when bound. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#binding
	//
	// Default: - Not bound to a specific claim.
	Claim IPersistentVolumeClaim `field:"optional" json:"claim" yaml:"claim"`
	// A list of mount options, e.g. ["ro", "soft"]. Not validated - mount will simply fail if one is invalid. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#mount-options
	//
	// Default: - No options.
	MountOptions *[]*string `field:"optional" json:"mountOptions" yaml:"mountOptions"`
	// When a user is done with their volume, they can delete the PVC objects from the API that allows reclamation of the resource.
	//
	// The reclaim policy tells the cluster what to do with the volume after it has been released of its claim. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#reclaiming
	//
	// Default: PersistentVolumeReclaimPolicy.RETAIN
	ReclaimPolicy PersistentVolumeReclaimPolicy `field:"optional" json:"reclaimPolicy" yaml:"reclaimPolicy"`
	// What is the storage capacity of this volume. See: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources
	//
	// Default: - No specified.
	Storage cdk8s.Size `field:"optional" json:"storage" yaml:"storage"`
	// Name of StorageClass to which this persistent volume belongs. Default: - Volume does not belong to any storage class.
	StorageClassName *string `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	// Defines what type of volume is required by the claim. Default: VolumeMode.FILE_SYSTEM
	VolumeMode PersistentVolumeMode `field:"optional" json:"volumeMode" yaml:"volumeMode"`
	// Unique name of the PD resource in GCE.
	//
	// Used to identify the disk in GCE. See: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk
	PdName *string `field:"required" json:"pdName" yaml:"pdName"`
	// Filesystem type of the volume that you want to mount.
	//
	// Tip: Ensure that the filesystem type is supported by the host operating system. See: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
	//
	// Default: 'ext4'.
	FsType *string `field:"optional" json:"fsType" yaml:"fsType"`
	// The partition in the volume that you want to mount.
	//
	// If omitted, the default is to mount by volume name. Examples: For volume /dev/sda1, you specify the partition as "1". Similarly, the volume partition for /dev/sda is "0" (or you can leave the property empty). Default: - No partition.
	Partition *float64 `field:"optional" json:"partition" yaml:"partition"`
	// Specify "true" to force and set the ReadOnly property in VolumeMounts to "true". See: https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore
	//
	// Default: false.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
}

// GCEPersistentDisk represents a GCE Disk resource that is attached to a kubelet's host machine and then exposed to the pod.
//
// Provisioned by an admin. See: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk
type GCEPersistentDiskPersistentVolume interface {
	PersistentVolume
	// File system type of this volume.
	FsType() *string
	// Partition of this volume.
	Partition() *float64
	// PD resource in GCE of this volume.
	PdName() *string
	// Whether or not it is mounted as a read-only volume.
	ReadOnly() *bool
}

func NewGCEPersistentDiskPersistentVolume(scope constructs.Construct, id *string, props *GCEPersistentDiskPersistentVolumeProps) GCEPersistentDiskPersistentVolume {
	if props == nil || props.PdName == nil {
		panic("pdName is required")
	}
	fs, ro := jsii.String("ext4"), jsii.Bool(false)
	if props.FsType != nil {
		fs = props.FsType
	}
	if props.ReadOnly != nil {
		ro = props.ReadOnly
	}
	result := newPersistentVolume(scope, id, persistentVolumePropsFromGce(props), map[string]interface{}{"gcePersistentDisk": map[string]interface{}{"pdName": props.PdName, "fsType": fs, "partition": props.Partition, "readOnly": ro}})
	result.pdName, result.fsType, result.partition, result.readOnly = props.PdName, fs, props.Partition, ro
	return result
}

func NewGCEPersistentDiskPersistentVolume_Override(volume GCEPersistentDiskPersistentVolume, scope constructs.Construct, id *string, props *GCEPersistentDiskPersistentVolumeProps) {
	applyOverride(volume, NewGCEPersistentDiskPersistentVolume(scope, id, props), "GCEPersistentDiskPersistentVolume")
}

// Imports a pv from the cluster as a reference.
func GCEPersistentDiskPersistentVolume_FromPersistentVolumeName(scope constructs.Construct, id, name *string) IPersistentVolume {
	return PersistentVolume_FromPersistentVolumeName(scope, id, name)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func GCEPersistentDiskPersistentVolume_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (p *persistentVolumeImpl) PdName() *string {
	return p.pdName
}

func persistentVolumePropsFromGce(p *GCEPersistentDiskPersistentVolumeProps) *PersistentVolumeProps {
	return &PersistentVolumeProps{Metadata: p.Metadata, AccessModes: p.AccessModes, Claim: p.Claim, MountOptions: p.MountOptions, ReclaimPolicy: p.ReclaimPolicy, Storage: p.Storage, StorageClassName: p.StorageClassName, VolumeMode: p.VolumeMode}
}

func persistentVolumeAccessModesManifest(values []PersistentVolumeAccessMode) []interface{} {
	result := make([]interface{}, 0, len(values))
	for _, value := range values {
		switch value {
		case PersistentVolumeAccessMode_READ_WRITE_ONCE:
			result = append(result, "ReadWriteOnce")
		case PersistentVolumeAccessMode_READ_ONLY_MANY:
			result = append(result, "ReadOnlyMany")
		case PersistentVolumeAccessMode_READ_WRITE_MANY:
			result = append(result, "ReadWriteMany")
		case PersistentVolumeAccessMode_READ_WRITE_ONCE_POD:
			result = append(result, "ReadWriteOncePod")
		default:
			panic("invalid persistent volume access mode")
		}
	}
	return result
}

func persistentVolumeModeManifest(value PersistentVolumeMode) string {
	if value == PersistentVolumeMode_FILE_SYSTEM {
		return "Filesystem"
	}
	if value == PersistentVolumeMode_BLOCK {
		return "Block"
	}
	panic("invalid persistent volume mode")
}

func persistentVolumeReclaimPolicyManifest(value PersistentVolumeReclaimPolicy) string {
	if value == PersistentVolumeReclaimPolicy_RETAIN {
		return "Retain"
	}
	if value == PersistentVolumeReclaimPolicy_DELETE {
		return "Delete"
	}
	panic("invalid persistent volume reclaim policy")
}

func sizeGibibytes(value cdk8s.Size) *string {
	return jsii.String(numberString(*value.ToGibibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_NONE})) + "Gi")
}
