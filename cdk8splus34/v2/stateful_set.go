package cdk8splus34

import (
	"fmt"
	"sort"

	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// StatefulSetProps configures a StatefulSet workload.
type StatefulSetProps struct {
	Metadata    *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Containers  *[]*ContainerProps       `field:"optional" json:"containers" yaml:"containers"`
	PodMetadata *cdk8s.ApiObjectMetadata `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	Replicas    *float64                 `field:"optional" json:"replicas" yaml:"replicas"`
	Select      *bool                    `field:"optional" json:"select" yaml:"select"`
	Spread      *bool                    `field:"optional" json:"spread" yaml:"spread"`
	Isolate     *bool                    `field:"optional" json:"isolate" yaml:"isolate"`
	Service     Service                  `field:"optional" json:"service" yaml:"service"`
}

// StatefulSet manages a set of pods with stable identities.
type StatefulSet interface {
	Resource
	IScalable
	Containers() *[]Container
	Replicas() *float64
	AddContainer(cont *ContainerProps) Container
	Service() Service
	Scheduling() WorkloadScheduling
	Connections() PodConnections
	ToPodSelectorConfig() *PodSelectorConfig
}

type statefulSetImpl struct {
	resourceBase
	containers    []Container
	replicas      *float64
	hasAutoscaler bool
	selector      map[string]*string
	podMetadata   *cdk8s.ApiObjectMetadata
	service       Service
	scheduling    *podSchedulingImpl
	connections   *podConnectionsImpl
	spread        bool
}

// NewStatefulSet creates a native StatefulSet and its default headless Service.
func NewStatefulSet(scope constructs.Construct, id *string, props *StatefulSetProps) StatefulSet {
	if scope == nil || id == nil {
		panic("scope and id are required")
	}
	if props == nil {
		props = &StatefulSetProps{}
	}
	result := &statefulSetImpl{
		replicas: props.Replicas, selector: map[string]*string{}, podMetadata: props.PodMetadata,
		scheduling: &podSchedulingImpl{}, spread: props.Spread != nil && *props.Spread,
	}
	constructs.NewConstruct_Override(result, scope, id)
	selectPods := true
	if props.Select != nil {
		selectPods = *props.Select
	}
	if selectPods {
		result.selector[podAddressLabel] = cdk8s.Names_ToLabelValue(result, nil)
	}
	if props.Containers != nil {
		for _, container := range *props.Containers {
			if container != nil {
				result.AddContainer(container)
			}
		}
	}
	if props.Service != nil {
		result.service = props.Service
		result.service.Select(result)
	} else {
		result.service = result.createHeadlessService(props.Metadata)
	}
	manifest := map[string]interface{}{}
	result.resourceBase.initializeApiObject(result, "apps/v1", "StatefulSet", "statefulsets", props.Metadata, manifest)
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toManifest() }})
	result.connections = &podConnectionsImpl{workload: result}
	if props.Isolate != nil && *props.Isolate {
		result.connections.Isolate()
	}
	return result
}

func NewStatefulSet_Override(s StatefulSet, scope constructs.Construct, id *string, props *StatefulSetProps) {
	panic("native cdk8splus34 overrides are not implemented")
}

func StatefulSet_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }

func (s *statefulSetImpl) Containers() *[]Container {
	values := append([]Container(nil), s.containers...)
	return &values
}
func (s *statefulSetImpl) Replicas() *float64   { return s.replicas }
func (s *statefulSetImpl) HasAutoscaler() *bool { return jsii.Bool(s.hasAutoscaler) }
func (s *statefulSetImpl) SetHasAutoscaler(value *bool) {
	s.hasAutoscaler = value != nil && *value
}
func (s *statefulSetImpl) MarkHasAutoscaler() { s.hasAutoscaler = true }
func (s *statefulSetImpl) ToScalingTarget() *ScalingTarget {
	containers := s.Containers()
	return &ScalingTarget{
		ApiVersion: s.ApiVersion(),
		Containers: containers,
		Kind:       s.Kind(),
		Name:       s.Name(),
		Replicas:   s.replicas,
	}
}
func (s *statefulSetImpl) AddContainer(props *ContainerProps) Container {
	container := NewContainer(props)
	s.containers = append(s.containers, container)
	return container
}
func (s *statefulSetImpl) Service() Service               { return s.service }
func (s *statefulSetImpl) Scheduling() WorkloadScheduling { return s.scheduling }
func (s *statefulSetImpl) Connections() PodConnections    { return s.connections }
func (s *statefulSetImpl) ToPodSelectorConfig() *PodSelectorConfig {
	labels := map[string]*string{}
	for key, value := range s.selector {
		labels[key] = value
	}
	return &PodSelectorConfig{LabelSelector: &LabelSelector{Labels: &labels}}
}

func (s *statefulSetImpl) createHeadlessService(metadata *cdk8s.ApiObjectMetadata) Service {
	ports := []*ServicePort{}
	for _, container := range s.containers {
		for _, port := range *container.Ports() {
			ports = append(ports, &ServicePort{Port: port.Number, TargetPort: port.Number, Protocol: port.Protocol, Name: port.Name})
		}
	}
	if len(ports) == 0 {
		panic(fmt.Sprintf("Unable to create a service for the stateful set %s: StatefulSet ports cannot be determined.", stringValue(s.Name())))
	}
	serviceMetadata := &cdk8s.ApiObjectMetadata{}
	if metadata != nil {
		serviceMetadata.Namespace = metadata.Namespace
	}
	return NewService(s, jsii.String("Service"), &ServiceProps{
		Metadata:  serviceMetadata,
		Selector:  s,
		Ports:     &ports,
		ClusterIP: jsii.String("None"),
		Type:      ServiceType_CLUSTER_IP,
	})
}

func (s *statefulSetImpl) toManifest() interface{} {
	if len(s.containers) == 0 {
		panic("PodSpec must have at least 1 container")
	}
	containers := make([]interface{}, 0, len(s.containers))
	volumes := map[string]Volume{}
	for _, container := range s.containers {
		native, ok := container.(*containerImpl)
		if !ok {
			panic("unsupported native container")
		}
		containers = append(containers, native.toManifest())
		for _, mount := range *native.Mounts() {
			volumes[stringValue(mount.Volume.Name())] = mount.Volume
		}
	}
	volumeValues := make([]interface{}, 0, len(volumes))
	names := make([]string, 0, len(volumes))
	for name := range volumes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		volume := volumes[name].(*volumeImpl)
		entry := map[string]interface{}{"name": volume.Name()}
		for key, value := range volume.spec {
			entry[key] = value
		}
		volumeValues = append(volumeValues, entry)
	}
	labels := copyLabels(nil)
	if s.podMetadata != nil {
		labels = copyLabels(s.podMetadata.Labels)
	}
	for key, value := range s.selector {
		labels[key] = value
	}
	podMetadata := map[string]interface{}{"labels": labels}
	if s.podMetadata != nil {
		if s.podMetadata.Annotations != nil {
			podMetadata["annotations"] = *s.podMetadata.Annotations
		}
		if s.podMetadata.Name != nil {
			podMetadata["name"] = s.podMetadata.Name
		}
		if s.podMetadata.Namespace != nil {
			podMetadata["namespace"] = s.podMetadata.Namespace
		}
	}
	podSpec := map[string]interface{}{
		"automountServiceAccountToken": false,
		"containers":                   containers,
		"dnsPolicy":                    "ClusterFirst",
		"hostNetwork":                  false,
		"restartPolicy":                "Always",
		"securityContext": map[string]interface{}{
			"fsGroupChangePolicy": "Always",
			"runAsNonRoot":        true,
		},
		"setHostnameAsFQDN":             false,
		"shareProcessNamespace":         false,
		"terminationGracePeriodSeconds": 30,
	}
	if len(volumeValues) > 0 {
		podSpec["volumes"] = volumeValues
	}
	if affinity := s.scheduling.toManifest(s, s.spread); affinity != nil {
		podSpec["affinity"] = affinity
	}
	replicas := s.replicas
	if replicas == nil && !s.hasAutoscaler {
		replicas = jsii.Number(1)
	}
	result := map[string]interface{}{
		"minReadySeconds":     0,
		"podManagementPolicy": "OrderedReady",
		"selector":            map[string]interface{}{"matchLabels": s.selector},
		"serviceName":         s.service.Name(),
		"template":            map[string]interface{}{"metadata": podMetadata, "spec": podSpec},
		"updateStrategy":      map[string]interface{}{"type": "RollingUpdate", "rollingUpdate": map[string]interface{}{"partition": 0}},
	}
	if replicas != nil {
		result["replicas"] = replicas
	}
	return result
}
