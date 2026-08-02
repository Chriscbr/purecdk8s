package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Properties used to configure the target of an Autoscaler.
type ScalingTarget struct {
	// The object's API version (e.g. "authorization.k8s.io/v1").
	ApiVersion *string `field:"required" json:"apiVersion" yaml:"apiVersion"`
	// Container definitions associated with the target.
	Containers *[]Container `field:"required" json:"containers" yaml:"containers"`
	// The object kind (e.g. "Deployment").
	Kind *string `field:"required" json:"kind" yaml:"kind"`
	// The Kubernetes name of this resource.
	Name *string `field:"required" json:"name" yaml:"name"`
	// The fixed number of replicas defined on the target.
	//
	// This is used for validation purposes as Scalable targets should not have a fixed number of replicas.
	Replicas *float64 `field:"optional" json:"replicas" yaml:"replicas"`
}

// Represents a scalable workload.
type IScalable interface {
	// Called on all IScalable targets when they are associated with an autoscaler.
	MarkHasAutoscaler()
	// Return the target spec properties of this Scalable.
	ToScalingTarget() *ScalingTarget
	// If this is a target of an autoscaler.
	HasAutoscaler() *bool
	SetHasAutoscaler(value *bool)
}

// Properties for HorizontalPodAutoscaler.
type HorizontalPodAutoscalerProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// The maximum number of replicas that can be scaled up to.
	MaxReplicas *float64 `field:"required" json:"maxReplicas" yaml:"maxReplicas"`
	// The workload to scale up or down.
	//
	// Scalable workload types: * Deployment * StatefulSet.
	Target IScalable `field:"required" json:"target" yaml:"target"`
	// The metric conditions that trigger a scale up or scale down. Default: - If metrics are not provided, then the target resource constraints (e.g. cpu limit) will be used as scaling metrics.
	Metrics *[]Metric `field:"optional" json:"metrics" yaml:"metrics"`
	// The minimum number of replicas that can be scaled down to.
	//
	// Can be set to 0 if the alpha feature gate `HPAScaleToZero` is enabled and at least one Object or External metric is configured. Default: 1.
	MinReplicas *float64 `field:"optional" json:"minReplicas" yaml:"minReplicas"`
	// The scaling behavior when scaling down. Default: - Scale down to minReplica count with a 5 minute stabilization window.
	ScaleDown *ScalingRules `field:"optional" json:"scaleDown" yaml:"scaleDown"`
	// The scaling behavior when scaling up. Default: - Is the higher of: * Increase no more than 4 pods per 60 seconds * Double the number of pods per 60 seconds.
	ScaleUp *ScalingRules `field:"optional" json:"scaleUp" yaml:"scaleUp"`
}

// A HorizontalPodAutoscaler scales a workload up or down in response to a metric change.
//
// This allows your services to scale up when demand is high and scale down when they are no longer needed.
//
// Typical use cases for HorizontalPodAutoscaler:
//
//   - When Memory usage is above 70%, scale up the number of replicas to meet the demand.
//   - When CPU usage is below 30%, scale down the number of replicas to save resources.
//   - When a service is experiencing a spike in traffic, scale up the number of replicas to meet the demand. Then, when the traffic subsides, scale down the number of replicas to save resources.
//
// The autoscaler uses the following algorithm to determine the number of replicas to scale:
//
// `desiredReplicas = ceil[currentReplicas * ( currentMetricValue / desiredMetricValue )]`
//
// HorizontalPodAutoscaler's can be used to with any `Scalable` workload: * Deployment * StatefulSet
//
// **Targets that already have a replica count defined:**
//
// Remove any replica counts from the target resource before associating with a HorizontalPodAutoscaler. If this isn't done, then any time a change to that object is applied, Kubernetes will scale the current number of Pods to the value of the target.replicas key. This may not be desired and could lead to unexpected behavior.
//
// Example:
//
//	const backend = new kplus.Deployment(this, 'Backend', ...);
//
//	const hpa = new kplus.HorizontalPodAutoscaler(chart, 'Hpa', {
//	 target: backend,
//	 maxReplicas: 10,
//	 scaleUp: {
//	   policies: [
//	     {
//	       replicas: kplus.Replicas.absolute(3),
//	       duration: Duration.minutes(5),
//	     },
//	   ],
//	 },
//	});
//
// See: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#implicit-maintenance-mode-deactivation
type HorizontalPodAutoscaler interface {
	Resource
	// The maximum number of replicas that can be scaled up to.
	MaxReplicas() *float64
	// The minimum number of replicas that can be scaled down to.
	MinReplicas() *float64
	// The metric conditions that trigger a scale up or scale down.
	Metrics() *[]Metric
	// The scaling behavior when scaling down.
	ScaleDown() *ScalingRules
	// The scaling behavior when scaling up.
	ScaleUp() *ScalingRules
	// The workload to scale up or down.
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

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
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

// A metric condition that HorizontalPodAutoscaler's scale on.
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

// A metric condition that will trigger scaling behavior when satisfied.
//
// Example:
//
//	MetricTarget.averageUtilization(70); // 70% average utilization
type MetricTarget interface{ toManifest() interface{} }

type metricTarget struct{ manifest map[string]interface{} }

func (m *metricTarget) toManifest() interface{} {
	return m.manifest
}

// Target a percentage value across all relevant pods.
func MetricTarget_AverageUtilization(value *float64) MetricTarget {
	if value == nil {
		panic("averageUtilization is required")
	}
	return &metricTarget{manifest: map[string]interface{}{"type": "Utilization", "averageUtilization": value}}
}

// Target the average value across all relevant pods.
func MetricTarget_AverageValue(value *float64) MetricTarget {
	if value == nil {
		panic("averageValue is required")
	}
	return &metricTarget{manifest: map[string]interface{}{"type": "AverageValue", "averageValue": value}}
}

// Target a specific target value.
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

// Tracks the available CPU of the pods in a target.
//
// Note: Since the resource usages of all the containers are summed up the total pod utilization may not accurately represent the individual container resource usage. This could lead to situations where a single container might be running with high usage and the HPA will not scale out because the overall pod usage is still within acceptable limits.
//
// Use case: * Scale up when CPU is above 40%.
func Metric_ResourceCpu(target MetricTarget) Metric {
	return metricResource("cpu", target)
}

// Tracks the available Memory of the pods in a target.
//
// Note: Since the resource usages of all the containers are summed up the total pod utilization may not accurately represent the individual container resource usage. This could lead to situations where a single container might be running with high usage and the HPA will not scale out because the overall pod usage is still within acceptable limits.
//
// Use case: * Scale up when Memory is above 512MB.
func Metric_ResourceMemory(target MetricTarget) Metric {
	return metricResource("memory", target)
}

// Tracks the available Storage of the pods in a target.
//
// Note: Since the resource usages of all the containers are summed up the total pod utilization may not accurately represent the individual container resource usage. This could lead to situations where a single container might be running with high usage and the HPA will not scale out because the overall pod usage is still within acceptable limits.
func Metric_ResourceStorage(target MetricTarget) Metric {
	return metricResource("storage", target)
}

// Tracks the available Ephemeral Storage of the pods in a target.
//
// Note: Since the resource usages of all the containers are summed up the total pod utilization may not accurately represent the individual container resource usage. This could lead to situations where a single container might be running with high usage and the HPA will not scale out because the overall pod usage is still within acceptable limits.
func Metric_ResourceEphemeralStorage(target MetricTarget) Metric {
	return metricResource("ephemeral-storage", target)
}

// Base options for a Metric.
type MetricOptions struct {
	// The name of the metric to scale on.
	Name *string `field:"required" json:"name" yaml:"name"`
	// The target metric value that will trigger scaling.
	Target MetricTarget `field:"required" json:"target" yaml:"target"`
	// A selector to find a metric by label.
	//
	// When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. Default: - Just the metric 'name' will be used to gather metrics.
	LabelSelector LabelSelector `field:"optional" json:"labelSelector" yaml:"labelSelector"`
}

// Options for `Metric.containerResource()`.
type MetricContainerResourceOptions struct {
	// Container where the metric can be found.
	Container Container `field:"required" json:"container" yaml:"container"`
	// Target metric value that will trigger scaling.
	Target MetricTarget `field:"required" json:"target" yaml:"target"`
}

// Options for `Metric.object()`.
type MetricObjectOptions struct {
	// The name of the metric to scale on.
	Name *string `field:"required" json:"name" yaml:"name"`
	// The target metric value that will trigger scaling.
	Target MetricTarget `field:"required" json:"target" yaml:"target"`
	// A selector to find a metric by label.
	//
	// When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. Default: - Just the metric 'name' will be used to gather metrics.
	LabelSelector LabelSelector `field:"optional" json:"labelSelector" yaml:"labelSelector"`
	// Resource where the metric can be found.
	Object IResource `field:"required" json:"object" yaml:"object"`
}

func metricContainerResource(name string, options *MetricContainerResourceOptions) Metric {
	if options == nil || options.Container == nil || options.Target == nil {
		panic("container and target are required")
	}
	return &metricImpl{type_: "ContainerResource", manifest: map[string]interface{}{"type": "ContainerResource", "containerResource": map[string]interface{}{"name": name, "container": options.Container.Name(), "target": options.Target.toManifest()}}}
}

// Metric that tracks the CPU of a container.
//
// This metric will be tracked across all pods of the current scale target.
func Metric_ContainerCpu(options *MetricContainerResourceOptions) Metric {
	return metricContainerResource("cpu", options)
}

// Metric that tracks the Memory of a container.
//
// This metric will be tracked across all pods of the current scale target.
func Metric_ContainerMemory(options *MetricContainerResourceOptions) Metric {
	return metricContainerResource("memory", options)
}

// Metric that tracks the volume size of a container.
//
// This metric will be tracked across all pods of the current scale target.
func Metric_ContainerStorage(options *MetricContainerResourceOptions) Metric {
	return metricContainerResource("storage", options)
}

// Metric that tracks the local ephemeral storage of a container.
//
// This metric will be tracked across all pods of the current scale target.
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

// A global metric that is not associated with any Kubernetes object.
//
// Allows for autoscaling based on information coming from components running outside of the cluster.
//
// Use case: * Scale up when the length of an SQS queue is greater than 10 messages. * Scale down when an outside load balancer's queries are less than 10000 per second.
func Metric_External(options *MetricOptions) Metric {
	metric := metricNamedManifest(options)
	return &metricImpl{type_: "External", manifest: map[string]interface{}{"type": "External", "external": map[string]interface{}{"metric": metric, "target": options.Target.toManifest()}}}
}

// A pod metric that will be averaged across all pods of the current scale target.
//
// Use case: * Average CPU utilization across all pods * Transactions processed per second across all pods.
func Metric_Pods(options *MetricOptions) Metric {
	metric := metricNamedManifest(options)
	return &metricImpl{type_: "Pods", manifest: map[string]interface{}{"type": "Pods", "pods": map[string]interface{}{"metric": metric, "target": options.Target.toManifest()}}}
}

// Metric that describes a metric of a kubernetes object.
//
// Use case: * Scale on a Kubernetes Ingress's hits-per-second metric.
func Metric_Object(options *MetricObjectOptions) Metric {
	if options == nil || options.Object == nil {
		panic("object is required")
	}
	base := metricNamedManifest(&MetricOptions{Name: options.Name, Target: options.Target, LabelSelector: options.LabelSelector})
	return &metricImpl{type_: "Object", manifest: map[string]interface{}{"type": "Object", "object": map[string]interface{}{"describedObject": map[string]interface{}{"apiVersion": options.Object.ApiVersion(), "kind": options.Object.Kind(), "name": options.Object.Name()}, "metric": base, "target": options.Target.toManifest()}}}
}

type ScalingStrategy string

const (
	// Use the policy that provisions the most changes.
	ScalingStrategy_MAX_CHANGE ScalingStrategy = "MAX_CHANGE"
	// Use the policy that provisions the least amount of changes.
	ScalingStrategy_MIN_CHANGE ScalingStrategy = "MIN_CHANGE"
	// Disables scaling in this direction. Deprecated: - Omit the ScalingRule instead.
	ScalingStrategy_DISABLED ScalingStrategy = "DISABLED"
)

type ScalingPolicy struct {
	// The type and quantity of replicas to change.
	Replicas Replicas `field:"required" json:"replicas" yaml:"replicas"`
	// The amount of time the scaling policy has to continue scaling before the target metric must be revalidated.
	//
	// Must be greater than 0 seconds and no longer than 30 minutes. Default: - 15 seconds.
	Duration cdk8s.Duration `field:"optional" json:"duration" yaml:"duration"`
}

// Defines the scaling behavior for one direction.
type ScalingRules struct {
	// The scaling policies. Default: * Scale up
	//   - Increase no more than 4 pods per 60 seconds
	//   - Double the number of pods per 60 seconds
	//
	// * Scale down * Decrease to minReplica count.
	Policies *[]*ScalingPolicy `field:"optional" json:"policies" yaml:"policies"`
	// Defines the window of past metrics that the autoscaler should consider when calculating wether or not autoscaling should occur.
	//
	// Minimum duration is 1 second, max is 1 hour.
	//
	// Example:
	//
	//	stabilizationWindow: Duration.minutes(30)
	//	// Autoscaler considers the last 30 minutes of metrics when deciding whether to scale.
	//
	// Default: * On scale down no stabilization is performed. * On scale up stabilization is performed for 5 minutes.
	StabilizationWindow cdk8s.Duration `field:"optional" json:"stabilizationWindow" yaml:"stabilizationWindow"`
	// The strategy to use when scaling. Default: MAX_CHANGE.
	Strategy ScalingStrategy `field:"optional" json:"strategy" yaml:"strategy"`
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
