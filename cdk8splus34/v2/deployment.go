package cdk8splus34

import (
	"fmt"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// DeploymentProps configures a Deployment workload.
type DeploymentProps struct {
	Metadata                     *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	AutomountServiceAccountToken *bool                    `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	Containers                   *[]*ContainerProps       `field:"optional" json:"containers" yaml:"containers"`
	Dns                          *PodDnsProps             `field:"optional" json:"dns" yaml:"dns"`
	DockerRegistryAuth           ISecret                  `field:"optional" json:"dockerRegistryAuth" yaml:"dockerRegistryAuth"`
	EnableServiceLinks           *bool                    `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	HostAliases                  *[]*HostAlias            `field:"optional" json:"hostAliases" yaml:"hostAliases"`
	HostNetwork                  *bool                    `field:"optional" json:"hostNetwork" yaml:"hostNetwork"`
	InitContainers               *[]*ContainerProps       `field:"optional" json:"initContainers" yaml:"initContainers"`
	Isolate                      *bool                    `field:"optional" json:"isolate" yaml:"isolate"`
	MinReady                     cdk8s.Duration           `field:"optional" json:"minReady" yaml:"minReady"`
	PodMetadata                  *cdk8s.ApiObjectMetadata `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	ProgressDeadline             cdk8s.Duration           `field:"optional" json:"progressDeadline" yaml:"progressDeadline"`
	Replicas                     *float64                 `field:"optional" json:"replicas" yaml:"replicas"`
	RestartPolicy                RestartPolicy            `field:"optional" json:"restartPolicy" yaml:"restartPolicy"`
	RevisionHistoryLimit         *float64                 `field:"optional" json:"revisionHistoryLimit" yaml:"revisionHistoryLimit"`
	SecurityContext              *PodSecurityContextProps `field:"optional" json:"securityContext" yaml:"securityContext"`
	Select                       *bool                    `field:"optional" json:"select" yaml:"select"`
	ServiceAccount               IServiceAccount          `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ShareProcessNamespace        *bool                    `field:"optional" json:"shareProcessNamespace" yaml:"shareProcessNamespace"`
	Spread                       *bool                    `field:"optional" json:"spread" yaml:"spread"`
	Strategy                     DeploymentStrategy       `field:"optional" json:"strategy" yaml:"strategy"`
	TerminationGracePeriod       cdk8s.Duration           `field:"optional" json:"terminationGracePeriod" yaml:"terminationGracePeriod"`
	Volumes                      *[]Volume                `field:"optional" json:"volumes" yaml:"volumes"`
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
	Workload
	IScalable
	Replicas() *float64
	MinReady() cdk8s.Duration
	ProgressDeadline() cdk8s.Duration
	RevisionHistoryLimit() *float64
	Strategy() DeploymentStrategy
	ExposeViaService(options *DeploymentExposeViaServiceOptions) Service
	ExposeViaIngress(path *string, options *ExposeDeploymentViaIngressOptions) Ingress
}

type deploymentImpl struct {
	resourceBase
	podState
	replicas             *float64
	hasAutoscaler        bool
	selector             map[string]*string
	matchExpressions     []*LabelSelectorRequirement
	podMetadata          *cdk8s.ApiObjectMetadata
	scheduling           WorkloadScheduling
	connections          PodConnections
	spread               bool
	strategy             DeploymentStrategy
	minReady             cdk8s.Duration
	progressDeadline     cdk8s.Duration
	revisionHistoryLimit *float64
}

func NewDeployment(scope constructs.Construct, id *string, props *DeploymentProps) Deployment {
	if props == nil {
		props = &DeploymentProps{}
	}
	result := &deploymentImpl{
		podState: newPodState(deploymentPodProps(props)), replicas: props.Replicas, selector: map[string]*string{}, podMetadata: props.PodMetadata,
		spread:   props.Spread != nil && *props.Spread,
		strategy: props.Strategy,
		minReady: props.MinReady, progressDeadline: props.ProgressDeadline, revisionHistoryLimit: props.RevisionHistoryLimit,
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
	result.scheduling = NewWorkloadScheduling(result)
	if result.spread {
		result.scheduling.Spread(&WorkloadSchedulingSpreadOptions{Topology: Topology_HOSTNAME()})
		result.scheduling.Spread(&WorkloadSchedulingSpreadOptions{Topology: Topology_ZONE()})
	}
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toManifest() }})
	result.connections = NewPodConnections(result)
	if props.Isolate != nil && *props.Isolate {
		result.connections.Isolate()
	}
	return result
}

func NewDeployment_Override(d Deployment, scope constructs.Construct, id *string, props *DeploymentProps) {
	applyOverride(d, NewDeployment(scope, id, props), "Deployment")
}

func Deployment_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (d *deploymentImpl) Containers() *[]Container {
	values := append([]Container(nil), d.podState.containers...)
	return &values
}

func (d *deploymentImpl) Replicas() *float64 {
	return d.replicas
}

func (d *deploymentImpl) MinReady() cdk8s.Duration {
	if d.minReady == nil {
		return cdk8s.Duration_Seconds(jsii.Number(0))
	}
	return d.minReady
}

func (d *deploymentImpl) ProgressDeadline() cdk8s.Duration {
	if d.progressDeadline == nil {
		return cdk8s.Duration_Seconds(jsii.Number(600))
	}
	return d.progressDeadline
}

func (d *deploymentImpl) RevisionHistoryLimit() *float64 {
	if d.revisionHistoryLimit == nil {
		return jsii.Number(10)
	}
	return d.revisionHistoryLimit
}

func (d *deploymentImpl) Strategy() DeploymentStrategy {
	if d.strategy == nil {
		return DeploymentStrategy_RollingUpdate(nil)
	}
	return d.strategy
}

func (d *deploymentImpl) HasAutoscaler() *bool {
	return jsii.Bool(d.hasAutoscaler)
}

func (d *deploymentImpl) SetHasAutoscaler(value *bool) {
	d.hasAutoscaler = value != nil && *value
}

func (d *deploymentImpl) MarkHasAutoscaler() {
	d.hasAutoscaler = true
}

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
	return d.addContainer(props)
}

func (d *deploymentImpl) ToPodSelectorConfig() *PodSelectorConfig {
	labels := map[string]*string{}
	for k, v := range d.selector {
		labels[k] = v
	}
	return &PodSelectorConfig{LabelSelector: newLabelSelectorFromRequirements(d.matchExpressions, &labels)}
}

func (d *deploymentImpl) Scheduling() WorkloadScheduling {
	return d.scheduling
}

func (d *deploymentImpl) Connections() PodConnections {
	return d.connections
}

func (d *deploymentImpl) toManifest() interface{} {
	spec := d.podState.manifest(d.RestartPolicy())
	for key, value := range d.scheduling.toManifest() {
		spec[key] = value
	}
	replicas := d.replicas
	if replicas == nil && !d.hasAutoscaler {
		replicas = jsii.Number(2)
	}
	result := map[string]interface{}{"selector": d.workloadSelector(), "template": map[string]interface{}{"metadata": d.PodMetadata().ToJson(), "spec": spec}, "strategy": d.Strategy().toManifest(), "minReadySeconds": d.MinReady().ToSeconds(nil), "progressDeadlineSeconds": d.ProgressDeadline().ToSeconds(nil), "revisionHistoryLimit": d.RevisionHistoryLimit()}
	if replicas != nil {
		result["replicas"] = replicas
	}
	return result
}

func deploymentPodProps(p *DeploymentProps) *PodProps {
	return &PodProps{Metadata: p.Metadata, AutomountServiceAccountToken: p.AutomountServiceAccountToken, Containers: p.Containers, Dns: p.Dns, DockerRegistryAuth: p.DockerRegistryAuth, EnableServiceLinks: p.EnableServiceLinks, HostAliases: p.HostAliases, HostNetwork: p.HostNetwork, InitContainers: p.InitContainers, Isolate: p.Isolate, RestartPolicy: p.RestartPolicy, SecurityContext: p.SecurityContext, ServiceAccount: p.ServiceAccount, ShareProcessNamespace: p.ShareProcessNamespace, TerminationGracePeriod: p.TerminationGracePeriod, Volumes: p.Volumes}
}

func (d *deploymentImpl) PodMetadata() cdk8s.ApiObjectMetadataDefinition {
	metadata := d.podMetadata
	if metadata == nil {
		metadata = &cdk8s.ApiObjectMetadata{}
	}
	result := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{ApiObject: d.ApiObject(), Name: metadata.Name, Namespace: metadata.Namespace, Labels: metadata.Labels, Annotations: metadata.Annotations})
	for key, value := range d.selector {
		result.AddLabel(jsii.String(key), value)
	}
	return result
}

func (d *deploymentImpl) workloadSelector() map[string]interface{} {
	result := map[string]interface{}{"matchLabels": d.selector}
	if len(d.matchExpressions) > 0 {
		result["matchExpressions"] = d.matchExpressions
	}
	return result
}

func (d *deploymentImpl) MatchLabels() *map[string]*string {
	values := map[string]*string{}
	for key, value := range d.selector {
		values[key] = value
	}
	return &values
}

func (d *deploymentImpl) MatchExpressions() *[]*LabelSelectorRequirement {
	values := append([]*LabelSelectorRequirement(nil), d.matchExpressions...)
	return &values
}

func (d *deploymentImpl) Select(selectors ...LabelSelector) {
	for _, selector := range selectors {
		if selector == nil {
			panic("selector is required")
		}
		for key, value := range labelSelectorLabels(selector) {
			d.selector[key] = value
		}
		d.matchExpressions = append(d.matchExpressions, labelSelectorRequirements(selector)...)
	}
}

func (d *deploymentImpl) InitContainers() *[]Container {
	values := append([]Container(nil), d.podState.initContainers...)
	return &values
}

func (d *deploymentImpl) Volumes() *[]Volume {
	values := append([]Volume(nil), d.podState.volumes...)
	return &values
}

func (d *deploymentImpl) AddInitContainer(props *ContainerProps) Container {
	return d.addInitContainer(props)
}

func (d *deploymentImpl) AddVolume(volume Volume) {
	d.addVolume(volume)
}

func (d *deploymentImpl) AddHostAlias(alias *HostAlias) {
	if alias == nil || alias.Ip == nil || alias.Hostnames == nil {
		panic("host alias IP and hostnames are required")
	}
	d.hostAliases = append(d.hostAliases, alias)
}

func (d *deploymentImpl) AttachContainer(container Container) {
	if container == nil {
		panic("container is required")
	}
	d.containers = append(d.containers, container)
}

func (d *deploymentImpl) ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig {
	return &NetworkPolicyPeerConfig{PodSelector: d.ToPodSelectorConfig()}
}

func (d *deploymentImpl) ToPodSelector() IPodSelector {
	return d
}

func (d *deploymentImpl) AutomountServiceAccountToken() *bool {
	if d.props.AutomountServiceAccountToken == nil {
		return jsii.Bool(false)
	}
	return d.props.AutomountServiceAccountToken
}

func (d *deploymentImpl) Dns() PodDns {
	return d.dns
}

func (d *deploymentImpl) DockerRegistryAuth() ISecret {
	return d.props.DockerRegistryAuth
}

func (d *deploymentImpl) EnableServiceLinks() *bool {
	return d.props.EnableServiceLinks
}

func (d *deploymentImpl) HostAliases() *[]*HostAlias {
	values := append([]*HostAlias(nil), d.hostAliases...)
	return &values
}

func (d *deploymentImpl) HostNetwork() *bool {
	if d.props.HostNetwork == nil {
		return jsii.Bool(false)
	}
	return d.props.HostNetwork
}

func (d *deploymentImpl) Isolate() *bool {
	if d.props.Isolate == nil {
		return jsii.Bool(false)
	}
	return d.props.Isolate
}

func (d *deploymentImpl) RestartPolicy() RestartPolicy {
	if d.props.RestartPolicy == "" {
		return RestartPolicy_ALWAYS
	}
	return d.props.RestartPolicy
}

func (d *deploymentImpl) SecurityContext() PodSecurityContext {
	return d.security
}

func (d *deploymentImpl) ServiceAccount() IServiceAccount {
	return d.props.ServiceAccount
}

func (d *deploymentImpl) ShareProcessNamespace() *bool {
	if d.props.ShareProcessNamespace == nil {
		return jsii.Bool(false)
	}
	return d.props.ShareProcessNamespace
}

func (d *deploymentImpl) TerminationGracePeriod() cdk8s.Duration {
	if d.props.TerminationGracePeriod == nil {
		return cdk8s.Duration_Seconds(jsii.Number(30))
	}
	return d.props.TerminationGracePeriod
}

func (d *deploymentImpl) ToSubjectConfiguration() *SubjectConfiguration {
	if d.props.ServiceAccount == nil && !*d.AutomountServiceAccountToken() {
		panic(stringValue(d.Name()) + " cannot be converted to a role binding subject: You must either assign a service account to it, or use 'automountServiceAccountToken: true'")
	}
	name := jsii.String("default")
	if d.props.ServiceAccount != nil {
		name = d.props.ServiceAccount.ResourceName()
	}
	return &SubjectConfiguration{ApiGroup: jsii.String(""), Kind: jsii.String("ServiceAccount"), Name: name}
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
