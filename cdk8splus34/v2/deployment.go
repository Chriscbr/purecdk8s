package cdk8splus34

import (
	"fmt"
	"sort"

	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// DeploymentProps configures a Deployment workload.
type DeploymentProps struct {
	Metadata    *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Containers  *[]*ContainerProps       `field:"optional" json:"containers" yaml:"containers"`
	PodMetadata *cdk8s.ApiObjectMetadata `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	Replicas    *float64                 `field:"optional" json:"replicas" yaml:"replicas"`
	Select      *bool                    `field:"optional" json:"select" yaml:"select"`
	Spread      *bool                    `field:"optional" json:"spread" yaml:"spread"`
	Isolate     *bool                    `field:"optional" json:"isolate" yaml:"isolate"`
}
type DeploymentExposeViaServiceOptions struct {
	Ports       *[]*ServicePort `field:"optional" json:"ports" yaml:"ports"`
	ServiceType ServiceType     `field:"optional" json:"serviceType" yaml:"serviceType"`
	Name        *string         `field:"optional" json:"name" yaml:"name"`
}
type ExposeDeploymentViaIngressOptions struct {
	Ports       *[]*ServicePort     `field:"optional" json:"ports" yaml:"ports"`
	ServiceType ServiceType         `field:"optional" json:"serviceType" yaml:"serviceType"`
	Name        *string             `field:"optional" json:"name" yaml:"name"`
	Ingress     Ingress             `field:"optional" json:"ingress" yaml:"ingress"`
	PathType    HttpIngressPathType `field:"optional" json:"pathType" yaml:"pathType"`
}
type Deployment interface {
	Resource
	IScalable
	Containers() *[]Container
	Replicas() *float64
	AddContainer(cont *ContainerProps) Container
	ExposeViaService(options *DeploymentExposeViaServiceOptions) Service
	ExposeViaIngress(path *string, options *ExposeDeploymentViaIngressOptions) Ingress
	ToPodSelectorConfig() *PodSelectorConfig
	Scheduling() WorkloadScheduling
	Connections() PodConnections
}
type deploymentImpl struct {
	resourceBase
	containers    []Container
	replicas      *float64
	hasAutoscaler bool
	selector      map[string]*string
	podMetadata   *cdk8s.ApiObjectMetadata
	scheduling    *podSchedulingImpl
	connections   *podConnectionsImpl
	spread        bool
}

func NewDeployment(scope constructs.Construct, id *string, props *DeploymentProps) Deployment {
	if props == nil {
		props = &DeploymentProps{}
	}
	result := &deploymentImpl{
		replicas: props.Replicas, selector: map[string]*string{}, podMetadata: props.PodMetadata,
		scheduling: &podSchedulingImpl{}, spread: props.Spread != nil && *props.Spread,
	}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "apps/v1", "Deployment", "deployments", props.Metadata, manifest)
	selectPods := true
	if props.Select != nil {
		selectPods = *props.Select
	}
	if selectPods {
		matcher := cdk8s.Names_ToLabelValue(result, nil)
		result.selector[podAddressLabel] = matcher
	}
	if props.Containers != nil {
		for _, container := range *props.Containers {
			if container != nil {
				result.AddContainer(container)
			}
		}
	}
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toManifest() }})
	result.connections = &podConnectionsImpl{workload: result}
	if props.Isolate != nil && *props.Isolate {
		result.connections.Isolate()
	}
	return result
}
func NewDeployment_Override(d Deployment, scope constructs.Construct, id *string, props *DeploymentProps) {
	panic("native cdk8splus34 overrides are not implemented")
}
func Deployment_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }
func (d *deploymentImpl) Containers() *[]Container {
	values := append([]Container(nil), d.containers...)
	return &values
}
func (d *deploymentImpl) Replicas() *float64   { return d.replicas }
func (d *deploymentImpl) HasAutoscaler() *bool { return jsii.Bool(d.hasAutoscaler) }
func (d *deploymentImpl) SetHasAutoscaler(value *bool) {
	d.hasAutoscaler = value != nil && *value
}
func (d *deploymentImpl) MarkHasAutoscaler() { d.hasAutoscaler = true }
func (d *deploymentImpl) ToScalingTarget() *ScalingTarget {
	containers := d.Containers()
	return &ScalingTarget{
		ApiVersion: d.ApiVersion(),
		Containers: containers,
		Kind:       d.Kind(),
		Name:       d.Name(),
		Replicas:   d.replicas,
	}
}
func (d *deploymentImpl) AddContainer(props *ContainerProps) Container {
	container := NewContainer(props)
	d.containers = append(d.containers, container)
	return container
}
func (d *deploymentImpl) ToPodSelectorConfig() *PodSelectorConfig {
	labels := map[string]*string{}
	for k, v := range d.selector {
		labels[k] = v
	}
	return &PodSelectorConfig{LabelSelector: &LabelSelector{Labels: &labels}}
}
func (d *deploymentImpl) Scheduling() WorkloadScheduling { return d.scheduling }
func (d *deploymentImpl) Connections() PodConnections    { return d.connections }
func (d *deploymentImpl) toManifest() interface{} {
	if len(d.containers) == 0 {
		panic("PodSpec must have at least 1 container")
	}
	containers := make([]interface{}, 0, len(d.containers))
	volumes := map[string]Volume{}
	for _, container := range d.containers {
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
	if d.podMetadata != nil {
		labels = copyLabels(d.podMetadata.Labels)
	}
	for key, value := range d.selector {
		labels[key] = value
	}
	podMetadata := map[string]interface{}{"labels": labels}
	if d.podMetadata != nil {
		if d.podMetadata.Annotations != nil {
			podMetadata["annotations"] = *d.podMetadata.Annotations
		}
		if d.podMetadata.Name != nil {
			podMetadata["name"] = d.podMetadata.Name
		}
		if d.podMetadata.Namespace != nil {
			podMetadata["namespace"] = d.podMetadata.Namespace
		}
	}
	spec := map[string]interface{}{
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
		spec["volumes"] = volumeValues
	}
	if affinity := d.scheduling.toManifest(d, d.spread); affinity != nil {
		spec["affinity"] = affinity
	}
	replicas := d.replicas
	if replicas == nil && !d.hasAutoscaler {
		replicas = jsii.Number(2)
	}
	result := map[string]interface{}{"selector": map[string]interface{}{"matchLabels": d.selector}, "template": map[string]interface{}{"metadata": podMetadata, "spec": spec}, "strategy": map[string]interface{}{"type": "RollingUpdate", "rollingUpdate": map[string]interface{}{"maxSurge": "25%", "maxUnavailable": "25%"}}, "minReadySeconds": 0, "progressDeadlineSeconds": 600, "revisionHistoryLimit": 10}
	if replicas != nil {
		result["replicas"] = replicas
	}
	return result
}
func (d *deploymentImpl) ExposeViaService(options *DeploymentExposeViaServiceOptions) Service {
	if options == nil {
		options = &DeploymentExposeViaServiceOptions{}
	}
	ports := options.Ports
	if ports == nil {
		derived := []*ServicePort{}
		for _, container := range d.containers {
			for _, port := range *container.Ports() {
				derived = append(derived, &ServicePort{Port: port.Number, TargetPort: port.Number, Protocol: port.Protocol, Name: port.Name})
			}
		}
		ports = &derived
	}
	if len(*ports) == 0 {
		panic(fmt.Sprintf("Unable to expose deployment %s via a service: Deployment port cannot be determined. Either pass 'ports', or configure ports on the containers of the deployment", stringValue(d.Name())))
	}
	metadata := &cdk8s.ApiObjectMetadata{Namespace: d.Metadata().Namespace()}
	if options.Name != nil {
		metadata.Name = options.Name
	}
	serviceType := options.ServiceType
	if serviceType == "" {
		serviceType = ServiceType_CLUSTER_IP
	}
	return NewService(d, jsii.String(stringValue(options.Name)+"Service"), &ServiceProps{Metadata: metadata, Selector: d, Ports: ports, Type: serviceType})
}
func (d *deploymentImpl) ExposeViaIngress(path *string, options *ExposeDeploymentViaIngressOptions) Ingress {
	if path == nil {
		panic("path is required")
	}
	service := d.ExposeViaService(&DeploymentExposeViaServiceOptions{Ports: optionsPorts(options), ServiceType: optionsServiceType(options), Name: optionsName(options)})
	return service.ExposeViaIngress(path, &ExposeServiceViaIngressOptions{Ingress: optionsIngress(options), PathType: optionsPathType(options)})
}
func optionsPorts(options *ExposeDeploymentViaIngressOptions) *[]*ServicePort {
	if options == nil {
		return nil
	}
	return options.Ports
}
func optionsServiceType(options *ExposeDeploymentViaIngressOptions) ServiceType {
	if options == nil {
		return ""
	}
	return options.ServiceType
}
func optionsName(options *ExposeDeploymentViaIngressOptions) *string {
	if options == nil {
		return nil
	}
	return options.Name
}
func optionsIngress(options *ExposeDeploymentViaIngressOptions) Ingress {
	if options == nil {
		return nil
	}
	return options.Ingress
}
func optionsPathType(options *ExposeDeploymentViaIngressOptions) HttpIngressPathType {
	if options == nil {
		return ""
	}
	return options.PathType
}
