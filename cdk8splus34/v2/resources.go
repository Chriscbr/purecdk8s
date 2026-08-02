package cdk8splus34

import (
	"strconv"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Represents the amount of CPU.
//
// The amount can be passed as millis or units.
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

func Cpu_Millis(amount *float64) Cpu {
	if amount == nil {
		panic("amount is required")
	}
	return &cpuImpl{amount: jsii.String(numberString(*amount) + "m")}
}

func Cpu_Units(amount *float64) Cpu {
	if amount == nil {
		panic("amount is required")
	}
	return &cpuImpl{amount: jsii.String(numberString(*amount))}
}

func numberString(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// CPU request and limit.
type CpuResources struct {
	Limit   Cpu `field:"optional" json:"limit" yaml:"limit"`
	Request Cpu `field:"optional" json:"request" yaml:"request"`
}

// Memory request and limit.
type MemoryResources struct {
	Limit   cdk8s.Size `field:"optional" json:"limit" yaml:"limit"`
	Request cdk8s.Size `field:"optional" json:"request" yaml:"request"`
}

// Emphemeral storage request and limit.
type EphemeralStorageResources struct {
	Limit   cdk8s.Size `field:"optional" json:"limit" yaml:"limit"`
	Request cdk8s.Size `field:"optional" json:"request" yaml:"request"`
}

// CPU and memory compute resources.
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
			limits["memory"] = jsii.String(numberString(*resources.Memory.Limit.ToMebibytes(nil)) + "Mi")
		}
		if resources.Memory.Request != nil {
			requests["memory"] = jsii.String(numberString(*resources.Memory.Request.ToMebibytes(nil)) + "Mi")
		}
	}
	if resources.EphemeralStorage != nil {
		if resources.EphemeralStorage.Limit != nil {
			limits["ephemeral-storage"] = jsii.String(numberString(*resources.EphemeralStorage.Limit.ToGibibytes(nil)) + "Gi")
		}
		if resources.EphemeralStorage.Request != nil {
			requests["ephemeral-storage"] = jsii.String(numberString(*resources.EphemeralStorage.Request.ToGibibytes(nil)) + "Gi")
		}
	}
	result := map[string]interface{}{}
	if len(limits) > 0 {
		result["limits"] = limits
	}
	if len(requests) > 0 {
		result["requests"] = requests
	}
	return result
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

// Union like class repsenting either a ration in percents or an absolute number.
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

// Absolute number.
func PercentOrAbsolute_Absolute(value *float64) PercentOrAbsolute {
	if value == nil {
		panic("value is required")
	}
	return &percentOrAbsoluteImpl{value: value}
}

// Percent ratio.
func PercentOrAbsolute_Percent(value *float64) PercentOrAbsolute {
	if value == nil {
		panic("percent is required")
	}
	return &percentOrAbsoluteImpl{value: jsii.String(numberString(*value) + "%")}
}

// Options for `DeploymentStrategy.rollingUpdate`.
type DeploymentStrategyRollingUpdateOptions struct {
	// The maximum number of pods that can be scheduled above the desired number of pods.
	//
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%). Absolute number is calculated from percentage by rounding up. This can not be 0 if `maxUnavailable` is 0.
	//
	// Example: when this is set to 30%, the new ReplicaSet can be scaled up immediately when the rolling update starts, such that the total number of old and new pods do not exceed 130% of desired pods. Once old pods have been killed, new ReplicaSet can be scaled up further, ensuring that total number of pods running at any time during the update is at most 130% of desired pods. Default: '25%'.
	MaxSurge PercentOrAbsolute `field:"optional" json:"maxSurge" yaml:"maxSurge"`
	// The maximum number of pods that can be unavailable during the update.
	//
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%). Absolute number is calculated from percentage by rounding down. This can not be 0 if `maxSurge` is 0.
	//
	// Example: when this is set to 30%, the old ReplicaSet can be scaled down to 70% of desired pods immediately when the rolling update starts. Once new pods are ready, old ReplicaSet can be scaled down further, followed by scaling up the new ReplicaSet, ensuring that the total number of pods available at all times during the update is at least 70% of desired pods. Default: '25%'.
	MaxUnavailable PercentOrAbsolute `field:"optional" json:"maxUnavailable" yaml:"maxUnavailable"`
}

// Deployment strategies.
type DeploymentStrategy interface{ toManifest() map[string]interface{} }

type deploymentStrategyImpl struct{ manifest map[string]interface{} }

func (s *deploymentStrategyImpl) toManifest() map[string]interface{} {
	return s.manifest
}

// All existing Pods are killed before new ones are created. See: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#recreate-deployment
func DeploymentStrategy_Recreate() DeploymentStrategy {
	return &deploymentStrategyImpl{manifest: map[string]interface{}{"type": "Recreate"}}
}

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
		panic("'maxSurge' and 'maxUnavailable' cannot be both zero")
	}
	return &deploymentStrategyImpl{manifest: map[string]interface{}{
		"type":          "RollingUpdate",
		"rollingUpdate": map[string]interface{}{"maxSurge": maxSurge.toManifest(), "maxUnavailable": maxUnavailable.toManifest()},
	}}
}

// Options for `StatefulSetUpdateStrategy.rollingUpdate`.
type StatefulSetUpdateStrategyRollingUpdateOptions struct {
	// If specified, all Pods with an ordinal that is greater than or equal to the partition will be updated when the StatefulSet's .spec.template is updated. All Pods with an ordinal that is less than the partition will not be updated, and, even if they are deleted, they will be recreated at the previous version.
	//
	// If the partition is greater than replicas, updates to the pod template will not be propagated to Pods. In most cases you will not need to use a partition, but they are useful if you want to stage an update, roll out a canary, or perform a phased roll out. See: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#partitions
	//
	// Default: 0.
	Partition *float64 `field:"optional" json:"partition" yaml:"partition"`
}

// StatefulSet update strategies.
type StatefulSetUpdateStrategy interface{ toManifest() map[string]interface{} }

type statefulSetUpdateStrategyImpl struct{ manifest map[string]interface{} }

func (s *statefulSetUpdateStrategyImpl) toManifest() map[string]interface{} {
	return s.manifest
}

// The controller will not automatically update the Pods in a StatefulSet.
//
// Users must manually delete Pods to cause the controller to create new Pods that reflect modifications.
func StatefulSetUpdateStrategy_OnDelete() StatefulSetUpdateStrategy {
	return &statefulSetUpdateStrategyImpl{manifest: map[string]interface{}{"type": "OnDelete"}}
}

// The controller will delete and recreate each Pod in the StatefulSet.
//
// It will proceed in the same order as Pod termination (from the largest ordinal to the smallest), updating each Pod one at a time. The Kubernetes control plane waits until an updated Pod is Running and Ready prior to updating its predecessor.
func StatefulSetUpdateStrategy_RollingUpdate(options *StatefulSetUpdateStrategyRollingUpdateOptions) StatefulSetUpdateStrategy {
	partition := jsii.Number(0)
	if options != nil && options.Partition != nil {
		partition = options.Partition
	}
	return &statefulSetUpdateStrategyImpl{manifest: map[string]interface{}{
		"type": "RollingUpdate", "rollingUpdate": map[string]interface{}{"partition": partition},
	}}
}

// The amount of replicas that will change.
type Replicas interface{ toManifest() map[string]interface{} }

type replicasImpl struct{ manifest map[string]interface{} }

func (r *replicasImpl) toManifest() map[string]interface{} {
	return r.manifest
}

// Changes the pods by a percentage of the it's current value.
func Replicas_Absolute(value *float64) Replicas {
	if value == nil || *value <= 0 {
		panic("replica value must be greater than zero")
	}
	return &replicasImpl{manifest: map[string]interface{}{"type": "Pods", "value": value}}
}

// Changes the pods by a percentage of the it's current value.
func Replicas_Percent(value *float64) Replicas {
	if value == nil || *value <= 0 {
		panic("replica value must be greater than zero")
	}
	return &replicasImpl{manifest: map[string]interface{}{"type": "Percent", "value": value}}
}
