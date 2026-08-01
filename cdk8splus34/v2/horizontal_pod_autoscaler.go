package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// ScalingTarget describes a workload that can be scaled by an autoscaler.
type ScalingTarget struct {
	ApiVersion *string      `field:"required" json:"apiVersion" yaml:"apiVersion"`
	Containers *[]Container `field:"required" json:"containers" yaml:"containers"`
	Kind       *string      `field:"required" json:"kind" yaml:"kind"`
	Name       *string      `field:"required" json:"name" yaml:"name"`
	Replicas   *float64     `field:"optional" json:"replicas" yaml:"replicas"`
}

// IScalable is implemented by workload resources supported by an HPA.
type IScalable interface {
	MarkHasAutoscaler()
	ToScalingTarget() *ScalingTarget
	HasAutoscaler() *bool
	SetHasAutoscaler(value *bool)
}

// HorizontalPodAutoscalerProps configures an autoscaler for a workload.
type HorizontalPodAutoscalerProps struct {
	Metadata    *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	MaxReplicas *float64                 `field:"required" json:"maxReplicas" yaml:"maxReplicas"`
	Target      IScalable                `field:"required" json:"target" yaml:"target"`
	Metrics     *[]Metric                `field:"optional" json:"metrics" yaml:"metrics"`
	MinReplicas *float64                 `field:"optional" json:"minReplicas" yaml:"minReplicas"`
}

// HorizontalPodAutoscaler is a Kubernetes autoscaling/v2 HPA resource.
type HorizontalPodAutoscaler interface {
	Resource
	MaxReplicas() *float64
	MinReplicas() *float64
	Metrics() *[]Metric
	Target() IScalable
}

type horizontalPodAutoscalerImpl struct {
	resourceBase
	maxReplicas *float64
	minReplicas *float64
	metrics     []Metric
	target      IScalable
}

func NewHorizontalPodAutoscaler(scope constructs.Construct, id *string, props *HorizontalPodAutoscalerProps) HorizontalPodAutoscaler {
	if props == nil || props.Target == nil || props.MaxReplicas == nil {
		panic("target and maxReplicas are required")
	}
	minReplicas := props.MinReplicas
	if minReplicas == nil {
		minReplicas = jsii.Number(1)
	}
	if *minReplicas > *props.MaxReplicas {
		panic("minReplicas must be less than or equal to maxReplicas")
	}
	result := &horizontalPodAutoscalerImpl{
		maxReplicas: props.MaxReplicas,
		minReplicas: minReplicas,
		target:      props.Target,
	}
	if props.Metrics != nil {
		result.metrics = append(result.metrics, (*props.Metrics)...)
	}
	props.Target.MarkHasAutoscaler()
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "autoscaling/v2", "HorizontalPodAutoscaler", "horizontalpodautoscalers", props.Metadata, manifest)
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: result.toManifest})
	return result
}

func NewHorizontalPodAutoscaler_Override(autoscaler HorizontalPodAutoscaler, scope constructs.Construct, id *string, props *HorizontalPodAutoscalerProps) {
	panic("native cdk8splus34 overrides are not implemented")
}

func HorizontalPodAutoscaler_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}
func (h *horizontalPodAutoscalerImpl) MaxReplicas() *float64 { return h.maxReplicas }
func (h *horizontalPodAutoscalerImpl) MinReplicas() *float64 { return h.minReplicas }
func (h *horizontalPodAutoscalerImpl) Target() IScalable     { return h.target }
func (h *horizontalPodAutoscalerImpl) Metrics() *[]Metric {
	metrics := append([]Metric(nil), h.metrics...)
	return &metrics
}

func (h *horizontalPodAutoscalerImpl) toManifest() interface{} {
	target := h.target.ToScalingTarget()
	if target == nil || target.ApiVersion == nil || target.Kind == nil || target.Name == nil {
		panic("autoscaler target is invalid")
	}
	metrics := make([]interface{}, 0, len(h.metrics))
	for _, metric := range h.metrics {
		if metric == nil {
			panic("autoscaler metric is required")
		}
		metrics = append(metrics, metric.toManifest())
	}
	result := map[string]interface{}{
		"behavior": map[string]interface{}{
			"scaleDown": map[string]interface{}{
				"policies":                   []interface{}{map[string]interface{}{"periodSeconds": 300, "type": "Pods", "value": h.minReplicas}},
				"selectPolicy":               "Max",
				"stabilizationWindowSeconds": 300,
			},
			"scaleUp": map[string]interface{}{
				"policies": []interface{}{
					map[string]interface{}{"periodSeconds": 60, "type": "Pods", "value": 4},
					map[string]interface{}{"periodSeconds": 60, "type": "Percent", "value": 200},
				},
				"selectPolicy":               "Max",
				"stabilizationWindowSeconds": 0,
			},
		},
		"maxReplicas": h.maxReplicas,
		"minReplicas": h.minReplicas,
		"scaleTargetRef": map[string]interface{}{
			"apiVersion": target.ApiVersion,
			"kind":       target.Kind,
			"name":       target.Name,
		},
	}
	if len(metrics) > 0 {
		result["metrics"] = metrics
	}
	return result
}

// Metric is an autoscaling metric specification.
type Metric interface {
	Type() *string
	toManifest() interface{}
}

type metricImpl struct {
	type_    string
	manifest map[string]interface{}
}

func (m *metricImpl) Type() *string           { return jsii.String(m.type_) }
func (m *metricImpl) toManifest() interface{} { return m.manifest }

// MetricTarget describes the desired value for an autoscaling metric.
type MetricTarget interface{ toManifest() interface{} }

type metricTarget struct{ manifest map[string]interface{} }

func (m *metricTarget) toManifest() interface{} { return m.manifest }

func MetricTarget_AverageUtilization(value *float64) MetricTarget {
	if value == nil {
		panic("averageUtilization is required")
	}
	return &metricTarget{manifest: map[string]interface{}{"type": "Utilization", "averageUtilization": value}}
}

func MetricTarget_AverageValue(value *float64) MetricTarget {
	if value == nil {
		panic("averageValue is required")
	}
	return &metricTarget{manifest: map[string]interface{}{"type": "AverageValue", "averageValue": value}}
}

func MetricTarget_Value(value *float64) MetricTarget {
	if value == nil {
		panic("value is required")
	}
	return &metricTarget{manifest: map[string]interface{}{"type": "Value", "value": value}}
}

func metricResource(name string, target MetricTarget) Metric {
	if target == nil {
		panic("target is required")
	}
	return &metricImpl{type_: "Resource", manifest: map[string]interface{}{
		"type":     "Resource",
		"resource": map[string]interface{}{"name": name, "target": target.toManifest()},
	}}
}

func Metric_ResourceCpu(target MetricTarget) Metric     { return metricResource("cpu", target) }
func Metric_ResourceMemory(target MetricTarget) Metric  { return metricResource("memory", target) }
func Metric_ResourceStorage(target MetricTarget) Metric { return metricResource("storage", target) }
func Metric_ResourceEphemeralStorage(target MetricTarget) Metric {
	return metricResource("ephemeral-storage", target)
}
