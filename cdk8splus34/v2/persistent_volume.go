package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

type (
	IPersistentVolume      interface{ IResource }
	IPersistentVolumeClaim interface{ IResource }
)

type PersistentVolumeAccessMode string

const (
	PersistentVolumeAccessMode_READ_WRITE_ONCE     PersistentVolumeAccessMode = "READ_WRITE_ONCE"
	PersistentVolumeAccessMode_READ_ONLY_MANY      PersistentVolumeAccessMode = "READ_ONLY_MANY"
	PersistentVolumeAccessMode_READ_WRITE_MANY     PersistentVolumeAccessMode = "READ_WRITE_MANY"
	PersistentVolumeAccessMode_READ_WRITE_ONCE_POD PersistentVolumeAccessMode = "READ_WRITE_ONCE_POD"
)

type PersistentVolumeMode string

const (
	PersistentVolumeMode_FILE_SYSTEM PersistentVolumeMode = "FILE_SYSTEM"
	PersistentVolumeMode_BLOCK       PersistentVolumeMode = "BLOCK"
)

type PersistentVolumeReclaimPolicy string

const (
	PersistentVolumeReclaimPolicy_RETAIN PersistentVolumeReclaimPolicy = "RETAIN"
	PersistentVolumeReclaimPolicy_DELETE PersistentVolumeReclaimPolicy = "DELETE"
)

type PersistentVolumeProps struct {
	Metadata         *cdk8s.ApiObjectMetadata      `field:"optional" json:"metadata" yaml:"metadata"`
	AccessModes      *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	Claim            IPersistentVolumeClaim        `field:"optional" json:"claim" yaml:"claim"`
	MountOptions     *[]*string                    `field:"optional" json:"mountOptions" yaml:"mountOptions"`
	ReclaimPolicy    PersistentVolumeReclaimPolicy `field:"optional" json:"reclaimPolicy" yaml:"reclaimPolicy"`
	Storage          cdk8s.Size                    `field:"optional" json:"storage" yaml:"storage"`
	StorageClassName *string                       `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	VolumeMode       PersistentVolumeMode          `field:"optional" json:"volumeMode" yaml:"volumeMode"`
}

type PersistentVolumeClaimProps struct {
	Metadata         *cdk8s.ApiObjectMetadata      `field:"optional" json:"metadata" yaml:"metadata"`
	AccessModes      *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	Storage          cdk8s.Size                    `field:"optional" json:"storage" yaml:"storage"`
	StorageClassName *string                       `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	Volume           IPersistentVolume             `field:"optional" json:"volume" yaml:"volume"`
	VolumeMode       PersistentVolumeMode          `field:"optional" json:"volumeMode" yaml:"volumeMode"`
}

type PersistentVolumeClaimTemplateProps struct {
	Metadata         *cdk8s.ApiObjectMetadata      `field:"optional" json:"metadata" yaml:"metadata"`
	AccessModes      *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	Storage          cdk8s.Size                    `field:"optional" json:"storage" yaml:"storage"`
	StorageClassName *string                       `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	Volume           IPersistentVolume             `field:"optional" json:"volume" yaml:"volume"`
	VolumeMode       PersistentVolumeMode          `field:"optional" json:"volumeMode" yaml:"volumeMode"`
	Name             *string                       `field:"required" json:"name" yaml:"name"`
}

type PersistentVolume interface {
	Resource
	IPersistentVolume
	IStorage
	AccessModes() *[]PersistentVolumeAccessMode
	Claim() IPersistentVolumeClaim
	Mode() PersistentVolumeMode
	MountOptions() *[]*string
	ReclaimPolicy() PersistentVolumeReclaimPolicy
	Storage() cdk8s.Size
	StorageClassName() *string
	Bind(claim IPersistentVolumeClaim)
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

func PersistentVolume_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (p *persistentVolumeImpl) AccessModes() *[]PersistentVolumeAccessMode {
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

type PersistentVolumeClaim interface {
	Resource
	IPersistentVolumeClaim
	AccessModes() *[]PersistentVolumeAccessMode
	Storage() cdk8s.Size
	StorageClassName() *string
	Volume() IPersistentVolume
	VolumeMode() PersistentVolumeMode
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

func PersistentVolumeClaim_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (p *persistentVolumeClaimImpl) AccessModes() *[]PersistentVolumeAccessMode {
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

func PersistentVolumeClaim_FromClaimName(scope constructs.Construct, id, name *string) IPersistentVolumeClaim {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and claimName are required")
	}
	result := &importedPersistentVolumeClaim{name: name}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

type AwsElasticBlockStorePersistentVolumeProps struct {
	Metadata         *cdk8s.ApiObjectMetadata      `field:"optional" json:"metadata" yaml:"metadata"`
	AccessModes      *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	Claim            IPersistentVolumeClaim        `field:"optional" json:"claim" yaml:"claim"`
	MountOptions     *[]*string                    `field:"optional" json:"mountOptions" yaml:"mountOptions"`
	ReclaimPolicy    PersistentVolumeReclaimPolicy `field:"optional" json:"reclaimPolicy" yaml:"reclaimPolicy"`
	Storage          cdk8s.Size                    `field:"optional" json:"storage" yaml:"storage"`
	StorageClassName *string                       `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	VolumeMode       PersistentVolumeMode          `field:"optional" json:"volumeMode" yaml:"volumeMode"`
	VolumeId         *string                       `field:"required" json:"volumeId" yaml:"volumeId"`
	FsType           *string                       `field:"optional" json:"fsType" yaml:"fsType"`
	Partition        *float64                      `field:"optional" json:"partition" yaml:"partition"`
	ReadOnly         *bool                         `field:"optional" json:"readOnly" yaml:"readOnly"`
}

type AwsElasticBlockStorePersistentVolume interface {
	PersistentVolume
	FsType() *string
	Partition() *float64
	ReadOnly() *bool
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
	result := newPersistentVolume(scope, id, persistentVolumePropsFromAws(props), map[string]interface{}{"awsElasticBlockStore": map[string]interface{}{"volumeId": props.VolumeId, "fsType": fs, "partition": props.Partition, "readOnly": ro}})
	result.volumeID, result.fsType, result.partition, result.readOnly = props.VolumeId, fs, props.Partition, ro
	return result
}

func NewAwsElasticBlockStorePersistentVolume_Override(volume AwsElasticBlockStorePersistentVolume, scope constructs.Construct, id *string, props *AwsElasticBlockStorePersistentVolumeProps) {
	applyOverride(volume, NewAwsElasticBlockStorePersistentVolume(scope, id, props), "AwsElasticBlockStorePersistentVolume")
}

func AwsElasticBlockStorePersistentVolume_FromPersistentVolumeName(scope constructs.Construct, id, name *string) IPersistentVolume {
	return PersistentVolume_FromPersistentVolumeName(scope, id, name)
}

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

type AzureDiskPersistentVolumeProps struct {
	Metadata         *cdk8s.ApiObjectMetadata             `field:"optional" json:"metadata" yaml:"metadata"`
	AccessModes      *[]PersistentVolumeAccessMode        `field:"optional" json:"accessModes" yaml:"accessModes"`
	Claim            IPersistentVolumeClaim               `field:"optional" json:"claim" yaml:"claim"`
	MountOptions     *[]*string                           `field:"optional" json:"mountOptions" yaml:"mountOptions"`
	ReclaimPolicy    PersistentVolumeReclaimPolicy        `field:"optional" json:"reclaimPolicy" yaml:"reclaimPolicy"`
	Storage          cdk8s.Size                           `field:"optional" json:"storage" yaml:"storage"`
	StorageClassName *string                              `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	VolumeMode       PersistentVolumeMode                 `field:"optional" json:"volumeMode" yaml:"volumeMode"`
	DiskName         *string                              `field:"required" json:"diskName" yaml:"diskName"`
	DiskUri          *string                              `field:"required" json:"diskUri" yaml:"diskUri"`
	CachingMode      AzureDiskPersistentVolumeCachingMode `field:"optional" json:"cachingMode" yaml:"cachingMode"`
	FsType           *string                              `field:"optional" json:"fsType" yaml:"fsType"`
	Kind             AzureDiskPersistentVolumeKind        `field:"optional" json:"kind" yaml:"kind"`
	ReadOnly         *bool                                `field:"optional" json:"readOnly" yaml:"readOnly"`
}

type AzureDiskPersistentVolume interface {
	PersistentVolume
	AzureKind() AzureDiskPersistentVolumeKind
	CachingMode() AzureDiskPersistentVolumeCachingMode
	DiskName() *string
	DiskUri() *string
	FsType() *string
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
	result := newPersistentVolume(scope, id, persistentVolumePropsFromAzure(props), map[string]interface{}{"azureDisk": map[string]interface{}{"diskName": props.DiskName, "diskUri": props.DiskUri, "cachingMode": azureCachingModeManifest(cache), "fsType": fs, "kind": azureKindManifest(kind), "readOnly": ro}})
	result.diskName, result.diskURI, result.fsType, result.readOnly, result.cachingMode, result.azureKind = props.DiskName, props.DiskUri, fs, ro, cache, kind
	return result
}

func NewAzureDiskPersistentVolume_Override(volume AzureDiskPersistentVolume, scope constructs.Construct, id *string, props *AzureDiskPersistentVolumeProps) {
	applyOverride(volume, NewAzureDiskPersistentVolume(scope, id, props), "AzureDiskPersistentVolume")
}

func AzureDiskPersistentVolume_FromPersistentVolumeName(scope constructs.Construct, id, name *string) IPersistentVolume {
	return PersistentVolume_FromPersistentVolumeName(scope, id, name)
}

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

type GCEPersistentDiskPersistentVolumeProps struct {
	Metadata         *cdk8s.ApiObjectMetadata      `field:"optional" json:"metadata" yaml:"metadata"`
	AccessModes      *[]PersistentVolumeAccessMode `field:"optional" json:"accessModes" yaml:"accessModes"`
	Claim            IPersistentVolumeClaim        `field:"optional" json:"claim" yaml:"claim"`
	MountOptions     *[]*string                    `field:"optional" json:"mountOptions" yaml:"mountOptions"`
	ReclaimPolicy    PersistentVolumeReclaimPolicy `field:"optional" json:"reclaimPolicy" yaml:"reclaimPolicy"`
	Storage          cdk8s.Size                    `field:"optional" json:"storage" yaml:"storage"`
	StorageClassName *string                       `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	VolumeMode       PersistentVolumeMode          `field:"optional" json:"volumeMode" yaml:"volumeMode"`
	PdName           *string                       `field:"required" json:"pdName" yaml:"pdName"`
	FsType           *string                       `field:"optional" json:"fsType" yaml:"fsType"`
	Partition        *float64                      `field:"optional" json:"partition" yaml:"partition"`
	ReadOnly         *bool                         `field:"optional" json:"readOnly" yaml:"readOnly"`
}

type GCEPersistentDiskPersistentVolume interface {
	PersistentVolume
	FsType() *string
	Partition() *float64
	PdName() *string
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

func GCEPersistentDiskPersistentVolume_FromPersistentVolumeName(scope constructs.Construct, id, name *string) IPersistentVolume {
	return PersistentVolume_FromPersistentVolumeName(scope, id, name)
}

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
