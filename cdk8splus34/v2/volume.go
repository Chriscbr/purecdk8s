package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// MountOptions controls one volume mount.
type MountOptions struct {
	ReadOnly    *bool   `field:"optional" json:"readOnly" yaml:"readOnly"`
	MountPath   *string `field:"optional" json:"mountPath" yaml:"mountPath"`
	SubPath     *string `field:"optional" json:"subPath" yaml:"subPath"`
	SubPathExpr *string `field:"optional" json:"subPathExpr" yaml:"subPathExpr"`
}

// VolumeMount is an attached volume.
type VolumeMount struct {
	Path      *string `field:"required" json:"path" yaml:"path"`
	Volume    Volume  `field:"required" json:"volume" yaml:"volume"`
	MountPath *string `field:"optional" json:"mountPath" yaml:"mountPath"`
	ReadOnly  *bool   `field:"optional" json:"readOnly" yaml:"readOnly"`
}

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

func (v *volumeImpl) Node() constructs.Node                { return v.node }
func (v *volumeImpl) SetNodeInternal(node constructs.Node) { v.node = node }
func (v *volumeImpl) ToString() *string                    { return v.node.Path() }
func (v *volumeImpl) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return v.node.With(mixins...)
}
func (v *volumeImpl) Name() *string    { return v.name }
func (v *volumeImpl) AsVolume() Volume { return v }

type ConfigMapVolumeOptions struct {
	DefaultMode *float64 `field:"optional" json:"defaultMode" yaml:"defaultMode"`
	Name        *string  `field:"optional" json:"name" yaml:"name"`
	Optional    *bool    `field:"optional" json:"optional" yaml:"optional"`
}
type EmptyDirMedium string

const (
	EmptyDirMedium_DEFAULT EmptyDirMedium = ""
	EmptyDirMedium_MEMORY  EmptyDirMedium = "Memory"
)

type EmptyDirVolumeOptions struct {
	Medium EmptyDirMedium `field:"optional" json:"medium" yaml:"medium"`
	Name   *string        `field:"optional" json:"name" yaml:"name"`
}

func newVolume(scope constructs.Construct, id, name *string, spec map[string]interface{}) Volume {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &volumeImpl{name: name, spec: spec}
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
	return newVolume(scope, id, name, map[string]interface{}{"configMap": config})
}
func Volume_FromEmptyDir(scope constructs.Construct, id, name *string, options *EmptyDirVolumeOptions) Volume {
	actualName := name
	if options != nil && options.Name != nil {
		actualName = options.Name
	}
	spec := map[string]interface{}{"emptyDir": map[string]interface{}{}}
	if options != nil && options.Medium != "" {
		spec["emptyDir"].(map[string]interface{})["medium"] = string(options.Medium)
	}
	return newVolume(scope, id, actualName, spec)
}
func Volume_FromName(scope constructs.Construct, id, name *string) Volume {
	return newVolume(scope, id, name, map[string]interface{}{})
}
func Volume_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }
