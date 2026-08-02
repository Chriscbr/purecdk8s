package cdk8splus34

import (
	"strconv"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Cpu represents a CPU quantity expressed in Kubernetes CPU units.
type Cpu interface {
	Amount() *string
	SetAmount(value *string)
}

type cpuImpl struct{ amount *string }

func (c *cpuImpl) Amount() *string {
	return c.amount
}

func (c *cpuImpl) SetAmount(value *string) {
	if value == nil {
		panic("amount is required")
	}
	c.amount = value
}

// Cpu_Millis returns a CPU quantity in millicores.
func Cpu_Millis(amount *float64) Cpu {
	if amount == nil {
		panic("amount is required")
	}
	return &cpuImpl{amount: jsii.String(numberString(*amount) + "m")}
}

// Cpu_Units returns a CPU quantity in whole CPU units.
func Cpu_Units(amount *float64) Cpu {
	if amount == nil {
		panic("amount is required")
	}
	return &cpuImpl{amount: jsii.String(numberString(*amount))}
}

func numberString(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// CpuResources contains CPU resource requests and limits.
type CpuResources struct {
	Limit   Cpu `field:"optional" json:"limit" yaml:"limit"`
	Request Cpu `field:"optional" json:"request" yaml:"request"`
}

// MemoryResources contains memory resource requests and limits.
type MemoryResources struct {
	Limit   cdk8s.Size `field:"optional" json:"limit" yaml:"limit"`
	Request cdk8s.Size `field:"optional" json:"request" yaml:"request"`
}

// EphemeralStorageResources contains ephemeral-storage resource requests and limits.
type EphemeralStorageResources struct {
	Limit   cdk8s.Size `field:"optional" json:"limit" yaml:"limit"`
	Request cdk8s.Size `field:"optional" json:"request" yaml:"request"`
}

// ContainerResources contains compute resource requests and limits.
type ContainerResources struct {
	Cpu              *CpuResources              `field:"optional" json:"cpu" yaml:"cpu"`
	EphemeralStorage *EphemeralStorageResources `field:"optional" json:"ephemeralStorage" yaml:"ephemeralStorage"`
	Memory           *MemoryResources           `field:"optional" json:"memory" yaml:"memory"`
}

func containerResourcesManifest(resources *ContainerResources) map[string]interface{} {
	if resources == nil {
		resources = normalizedContainerResources(nil)
	}
	limits := map[string]interface{}{}
	requests := map[string]interface{}{}
	if resources.Cpu != nil {
		if resources.Cpu.Limit != nil {
			limits["cpu"] = resources.Cpu.Limit.Amount()
		}
		if resources.Cpu.Request != nil {
			requests["cpu"] = resources.Cpu.Request.Amount()
		}
	}
	if resources.Memory != nil {
		if resources.Memory.Limit != nil {
			limits["memory"] = resources.Memory.Limit.AsString()
		}
		if resources.Memory.Request != nil {
			requests["memory"] = resources.Memory.Request.AsString()
		}
	}
	if resources.EphemeralStorage != nil {
		if resources.EphemeralStorage.Limit != nil {
			limits["ephemeral-storage"] = resources.EphemeralStorage.Limit.AsString()
		}
		if resources.EphemeralStorage.Request != nil {
			requests["ephemeral-storage"] = resources.EphemeralStorage.Request.AsString()
		}
	}
	return map[string]interface{}{"limits": limits, "requests": requests}
}

func normalizedContainerResources(resources *ContainerResources) *ContainerResources {
	if resources != nil {
		return resources
	}
	return &ContainerResources{
		Cpu:    &CpuResources{Limit: Cpu_Millis(jsii.Number(1500)), Request: Cpu_Millis(jsii.Number(1000))},
		Memory: &MemoryResources{Limit: cdk8s.Size_Mebibytes(jsii.Number(2048)), Request: cdk8s.Size_Mebibytes(jsii.Number(512))},
	}
}

// PercentOrAbsolute represents either an absolute number or a percentage.
type PercentOrAbsolute interface {
	Value() interface{}
	IsZero() *bool
	toManifest() interface{}
}

type percentOrAbsoluteImpl struct{ value interface{} }

func (p *percentOrAbsoluteImpl) Value() interface{} {
	return p.value
}

func (p *percentOrAbsoluteImpl) toManifest() interface{} {
	return p.value
}

func (p *percentOrAbsoluteImpl) IsZero() *bool {
	zero := false
	switch value := p.value.(type) {
	case *float64:
		zero = value != nil && *value == 0
	case *string:
		zero = value != nil && (*value == "0%" || *value == "0")
	}
	return jsii.Bool(zero)
}

// PercentOrAbsolute_Absolute returns an absolute quantity.
func PercentOrAbsolute_Absolute(value *float64) PercentOrAbsolute {
	if value == nil {
		panic("value is required")
	}
	return &percentOrAbsoluteImpl{value: value}
}

// PercentOrAbsolute_Percent returns a percentage quantity.
func PercentOrAbsolute_Percent(value *float64) PercentOrAbsolute {
	if value == nil {
		panic("percent is required")
	}
	return &percentOrAbsoluteImpl{value: jsii.String(numberString(*value) + "%")}
}

// DeploymentStrategyRollingUpdateOptions configures a rolling update.
type DeploymentStrategyRollingUpdateOptions struct {
	MaxSurge       PercentOrAbsolute `field:"optional" json:"maxSurge" yaml:"maxSurge"`
	MaxUnavailable PercentOrAbsolute `field:"optional" json:"maxUnavailable" yaml:"maxUnavailable"`
}

// DeploymentStrategy controls how a Deployment replaces pods.
type DeploymentStrategy interface{ toManifest() map[string]interface{} }

type deploymentStrategyImpl struct{ manifest map[string]interface{} }

func (s *deploymentStrategyImpl) toManifest() map[string]interface{} {
	return s.manifest
}

// DeploymentStrategy_Recreate deletes all pods before creating replacements.
func DeploymentStrategy_Recreate() DeploymentStrategy {
	return &deploymentStrategyImpl{manifest: map[string]interface{}{"type": "Recreate"}}
}

// DeploymentStrategy_RollingUpdate creates a rolling-update strategy.
func DeploymentStrategy_RollingUpdate(options *DeploymentStrategyRollingUpdateOptions) DeploymentStrategy {
	maxSurge := PercentOrAbsolute_Percent(jsii.Number(25))
	maxUnavailable := PercentOrAbsolute_Percent(jsii.Number(25))
	if options != nil {
		if options.MaxSurge != nil {
			maxSurge = options.MaxSurge
		}
		if options.MaxUnavailable != nil {
			maxUnavailable = options.MaxUnavailable
		}
	}
	if *maxSurge.IsZero() && *maxUnavailable.IsZero() {
		panic("maxSurge and maxUnavailable cannot both be zero")
	}
	return &deploymentStrategyImpl{manifest: map[string]interface{}{
		"type":          "RollingUpdate",
		"rollingUpdate": map[string]interface{}{"maxSurge": maxSurge.toManifest(), "maxUnavailable": maxUnavailable.toManifest()},
	}}
}

// StatefulSetUpdateStrategyRollingUpdateOptions configures a StatefulSet rolling update.
type StatefulSetUpdateStrategyRollingUpdateOptions struct {
	Partition *float64 `field:"optional" json:"partition" yaml:"partition"`
}

// StatefulSetUpdateStrategy controls how a StatefulSet replaces pods.
type StatefulSetUpdateStrategy interface{ toManifest() map[string]interface{} }

type statefulSetUpdateStrategyImpl struct{ manifest map[string]interface{} }

func (s *statefulSetUpdateStrategyImpl) toManifest() map[string]interface{} {
	return s.manifest
}

// StatefulSetUpdateStrategy_OnDelete requires users to delete pods manually.
func StatefulSetUpdateStrategy_OnDelete() StatefulSetUpdateStrategy {
	return &statefulSetUpdateStrategyImpl{manifest: map[string]interface{}{"type": "OnDelete"}}
}

// StatefulSetUpdateStrategy_RollingUpdate creates a rolling-update strategy.
func StatefulSetUpdateStrategy_RollingUpdate(options *StatefulSetUpdateStrategyRollingUpdateOptions) StatefulSetUpdateStrategy {
	partition := jsii.Number(0)
	if options != nil && options.Partition != nil {
		partition = options.Partition
	}
	return &statefulSetUpdateStrategyImpl{manifest: map[string]interface{}{
		"type": "RollingUpdate", "rollingUpdate": map[string]interface{}{"partition": partition},
	}}
}

// Replicas describes the amount a horizontal autoscaler scaling policy changes.
type Replicas interface{ toManifest() map[string]interface{} }

type replicasImpl struct{ manifest map[string]interface{} }

func (r *replicasImpl) toManifest() map[string]interface{} {
	return r.manifest
}

// Replicas_Absolute creates a policy that changes an absolute number of pods.
func Replicas_Absolute(value *float64) Replicas {
	if value == nil || *value <= 0 {
		panic("replica value must be greater than zero")
	}
	return &replicasImpl{manifest: map[string]interface{}{"type": "Pods", "value": value}}
}

// Replicas_Percent creates a policy that changes a percentage of pods.
func Replicas_Percent(value *float64) Replicas {
	if value == nil || *value <= 0 {
		panic("replica value must be greater than zero")
	}
	return &replicasImpl{manifest: map[string]interface{}{"type": "Percent", "value": value}}
}
