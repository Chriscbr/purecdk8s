package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
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
	ScaleDown   *ScalingRules            `field:"optional" json:"scaleDown" yaml:"scaleDown"`
	ScaleUp     *ScalingRules            `field:"optional" json:"scaleUp" yaml:"scaleUp"`
}

// HorizontalPodAutoscaler is a Kubernetes autoscaling/v2 HPA resource.
type HorizontalPodAutoscaler interface {
	Resource
	MaxReplicas() *float64
	MinReplicas() *float64
	Metrics() *[]Metric
	ScaleDown() *ScalingRules
	ScaleUp() *ScalingRules
	Target() IScalable
}

type horizontalPodAutoscalerImpl struct {
	resourceBase
	maxReplicas *float64
	minReplicas *float64
	metrics     []Metric
	target      IScalable
	scaleDown   *ScalingRules
	scaleUp     *ScalingRules
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
		scaleDown:   normalizedScalingRules(props.ScaleDown, false, minReplicas),
		scaleUp:     normalizedScalingRules(props.ScaleUp, true, minReplicas),
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
	applyOverride(autoscaler, NewHorizontalPodAutoscaler(scope, id, props), "HorizontalPodAutoscaler")
}

func HorizontalPodAutoscaler_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (h *horizontalPodAutoscalerImpl) MaxReplicas() *float64 {
	return h.maxReplicas
}

func (h *horizontalPodAutoscalerImpl) MinReplicas() *float64 {
	return h.minReplicas
}

func (h *horizontalPodAutoscalerImpl) Target() IScalable {
	return h.target
}

func (h *horizontalPodAutoscalerImpl) ScaleDown() *ScalingRules {
	return h.scaleDown
}

func (h *horizontalPodAutoscalerImpl) ScaleUp() *ScalingRules {
	return h.scaleUp
}

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
			"scaleDown": scalingRulesManifest(h.scaleDown),
			"scaleUp":   scalingRulesManifest(h.scaleUp),
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

func (m *metricImpl) Type() *string {
	return jsii.String(m.type_)
}

func (m *metricImpl) toManifest() interface{} {
	return m.manifest
}

// MetricTarget describes the desired value for an autoscaling metric.
type MetricTarget interface{ toManifest() interface{} }

type metricTarget struct{ manifest map[string]interface{} }

func (m *metricTarget) toManifest() interface{} {
	return m.manifest
}

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

func Metric_ResourceCpu(target MetricTarget) Metric {
	return metricResource("cpu", target)
}

func Metric_ResourceMemory(target MetricTarget) Metric {
	return metricResource("memory", target)
}

func Metric_ResourceStorage(target MetricTarget) Metric {
	return metricResource("storage", target)
}

func Metric_ResourceEphemeralStorage(target MetricTarget) Metric {
	return metricResource("ephemeral-storage", target)
}

type MetricOptions struct {
	Name          *string       `field:"required" json:"name" yaml:"name"`
	Target        MetricTarget  `field:"required" json:"target" yaml:"target"`
	LabelSelector LabelSelector `field:"optional" json:"labelSelector" yaml:"labelSelector"`
}

type MetricContainerResourceOptions struct {
	Container Container    `field:"required" json:"container" yaml:"container"`
	Target    MetricTarget `field:"required" json:"target" yaml:"target"`
}

type MetricObjectOptions struct {
	Name          *string       `field:"required" json:"name" yaml:"name"`
	Target        MetricTarget  `field:"required" json:"target" yaml:"target"`
	LabelSelector LabelSelector `field:"optional" json:"labelSelector" yaml:"labelSelector"`
	Object        IResource     `field:"required" json:"object" yaml:"object"`
}

func metricContainerResource(name string, options *MetricContainerResourceOptions) Metric {
	if options == nil || options.Container == nil || options.Target == nil {
		panic("container and target are required")
	}
	return &metricImpl{type_: "ContainerResource", manifest: map[string]interface{}{"type": "ContainerResource", "containerResource": map[string]interface{}{"name": name, "container": options.Container.Name(), "target": options.Target.toManifest()}}}
}

func Metric_ContainerCpu(options *MetricContainerResourceOptions) Metric {
	return metricContainerResource("cpu", options)
}

func Metric_ContainerMemory(options *MetricContainerResourceOptions) Metric {
	return metricContainerResource("memory", options)
}

func Metric_ContainerStorage(options *MetricContainerResourceOptions) Metric {
	return metricContainerResource("storage", options)
}

func Metric_ContainerEphemeralStorage(options *MetricContainerResourceOptions) Metric {
	return metricContainerResource("ephemeral-storage", options)
}

func metricNamedManifest(options *MetricOptions) map[string]interface{} {
	if options == nil || options.Name == nil || options.Target == nil {
		panic("name and target are required")
	}
	metric := map[string]interface{}{"name": options.Name}
	if options.LabelSelector != nil {
		metric["selector"] = labelSelectorManifest(options.LabelSelector)
	}
	return metric
}

func Metric_External(options *MetricOptions) Metric {
	metric := metricNamedManifest(options)
	return &metricImpl{type_: "External", manifest: map[string]interface{}{"type": "External", "external": map[string]interface{}{"metric": metric, "target": options.Target.toManifest()}}}
}

func Metric_Pods(options *MetricOptions) Metric {
	metric := metricNamedManifest(options)
	return &metricImpl{type_: "Pods", manifest: map[string]interface{}{"type": "Pods", "pods": map[string]interface{}{"metric": metric, "target": options.Target.toManifest()}}}
}

func Metric_Object(options *MetricObjectOptions) Metric {
	if options == nil || options.Object == nil {
		panic("object is required")
	}
	base := metricNamedManifest(&MetricOptions{Name: options.Name, Target: options.Target, LabelSelector: options.LabelSelector})
	return &metricImpl{type_: "Object", manifest: map[string]interface{}{"type": "Object", "object": map[string]interface{}{"describedObject": map[string]interface{}{"apiVersion": options.Object.ApiVersion(), "kind": options.Object.Kind(), "name": options.Object.Name()}, "metric": base, "target": options.Target.toManifest()}}}
}

type ScalingStrategy string

const (
	ScalingStrategy_MAX_CHANGE ScalingStrategy = "MAX_CHANGE"
	ScalingStrategy_MIN_CHANGE ScalingStrategy = "MIN_CHANGE"
	ScalingStrategy_DISABLED   ScalingStrategy = "DISABLED"
)

type ScalingPolicy struct {
	Replicas Replicas       `field:"required" json:"replicas" yaml:"replicas"`
	Duration cdk8s.Duration `field:"optional" json:"duration" yaml:"duration"`
}

type ScalingRules struct {
	Policies            *[]*ScalingPolicy `field:"optional" json:"policies" yaml:"policies"`
	StabilizationWindow cdk8s.Duration    `field:"optional" json:"stabilizationWindow" yaml:"stabilizationWindow"`
	Strategy            ScalingStrategy   `field:"optional" json:"strategy" yaml:"strategy"`
}

func normalizedScalingRules(rules *ScalingRules, up bool, minReplicas *float64) *ScalingRules {
	if rules == nil {
		rules = &ScalingRules{}
	}
	result := &ScalingRules{StabilizationWindow: rules.StabilizationWindow, Strategy: rules.Strategy}
	if result.Strategy == "" {
		result.Strategy = ScalingStrategy_MAX_CHANGE
	}
	if result.StabilizationWindow == nil {
		if up {
			result.StabilizationWindow = cdk8s.Duration_Seconds(jsii.Number(0))
		} else {
			result.StabilizationWindow = cdk8s.Duration_Minutes(jsii.Number(5))
		}
	}
	if rules.Policies != nil {
		result.Policies = rules.Policies
	} else if up {
		result.Policies = &[]*ScalingPolicy{{Replicas: Replicas_Absolute(jsii.Number(4)), Duration: cdk8s.Duration_Minutes(jsii.Number(1))}, {Replicas: Replicas_Percent(jsii.Number(200)), Duration: cdk8s.Duration_Minutes(jsii.Number(1))}}
	} else {
		result.Policies = &[]*ScalingPolicy{{Replicas: Replicas_Absolute(minReplicas), Duration: cdk8s.Duration_Minutes(jsii.Number(5))}}
	}
	return result
}

func scalingRulesManifest(rules *ScalingRules) map[string]interface{} {
	if rules == nil {
		return map[string]interface{}{}
	}
	result := map[string]interface{}{"selectPolicy": scalingStrategyManifest(rules.Strategy)}
	if rules.StabilizationWindow != nil {
		result["stabilizationWindowSeconds"] = rules.StabilizationWindow.ToSeconds(nil)
	}
	if rules.Policies != nil {
		policies := make([]interface{}, 0, len(*rules.Policies))
		for _, policy := range *rules.Policies {
			if policy == nil || policy.Replicas == nil {
				panic("scaling policy replicas are required")
			}
			duration := jsii.Number(15)
			if policy.Duration != nil {
				duration = policy.Duration.ToSeconds(nil)
			}
			entry := policy.Replicas.toManifest()
			entry["periodSeconds"] = duration
			policies = append(policies, entry)
		}
		result["policies"] = policies
	}
	return result
}

func scalingStrategyManifest(value ScalingStrategy) string {
	switch value {
	case ScalingStrategy_MAX_CHANGE:
		return "Max"
	case ScalingStrategy_MIN_CHANGE:
		return "Min"
	case ScalingStrategy_DISABLED:
		return "Disabled"
	default:
		panic("invalid scaling strategy")
	}
}
