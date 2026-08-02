package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// DaemonSetProps configures a DaemonSet workload.
type DaemonSetProps struct {
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
	MinReadySeconds              *float64                 `field:"optional" json:"minReadySeconds" yaml:"minReadySeconds"`
	PodMetadata                  *cdk8s.ApiObjectMetadata `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	RestartPolicy                RestartPolicy            `field:"optional" json:"restartPolicy" yaml:"restartPolicy"`
	SecurityContext              *PodSecurityContextProps `field:"optional" json:"securityContext" yaml:"securityContext"`
	Select                       *bool                    `field:"optional" json:"select" yaml:"select"`
	ServiceAccount               IServiceAccount          `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ShareProcessNamespace        *bool                    `field:"optional" json:"shareProcessNamespace" yaml:"shareProcessNamespace"`
	Spread                       *bool                    `field:"optional" json:"spread" yaml:"spread"`
	TerminationGracePeriod       cdk8s.Duration           `field:"optional" json:"terminationGracePeriod" yaml:"terminationGracePeriod"`
	Volumes                      *[]Volume                `field:"optional" json:"volumes" yaml:"volumes"`
}

// DaemonSet ensures matching nodes each run a copy of its pod template.
type DaemonSet interface {
	Workload
	MinReadySeconds() *float64
}

type daemonSetImpl struct {
	resourceBase
	podState
	minReadySeconds  *float64
	podMetadata      *cdk8s.ApiObjectMetadata
	selector         map[string]*string
	matchExpressions []*LabelSelectorRequirement
	scheduling       WorkloadScheduling
	connections      PodConnections
}

func NewDaemonSet(scope constructs.Construct, id *string, props *DaemonSetProps) DaemonSet {
	if props == nil {
		props = &DaemonSetProps{}
	}
	result := &daemonSetImpl{podState: newPodState(daemonSetPodProps(props)), minReadySeconds: props.MinReadySeconds, podMetadata: props.PodMetadata, selector: map[string]*string{}}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "apps/v1", "DaemonSet", "daemonsets", props.Metadata, manifest)
	selectPods := true
	if props.Select != nil {
		selectPods = *props.Select
	}
	if selectPods {
		result.selector[podAddressLabel] = cdk8s.Names_ToLabelValue(result, nil)
	}
	result.scheduling = NewWorkloadScheduling(result)
	if props.Spread != nil && *props.Spread {
		result.scheduling.Spread(&WorkloadSchedulingSpreadOptions{Topology: Topology_HOSTNAME()})
		result.scheduling.Spread(&WorkloadSchedulingSpreadOptions{Topology: Topology_ZONE()})
	}
	result.connections = NewPodConnections(result)
	if *result.Isolate() {
		result.connections.Isolate()
	}
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toManifest() }})
	return result
}

func NewDaemonSet_Override(daemonSet DaemonSet, scope constructs.Construct, id *string, props *DaemonSetProps) {
	applyOverride(daemonSet, NewDaemonSet(scope, id, props), "DaemonSet")
}

func DaemonSet_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (d *daemonSetImpl) MinReadySeconds() *float64 {
	if d.minReadySeconds == nil {
		return jsii.Number(0)
	}
	return d.minReadySeconds
}

func (d *daemonSetImpl) Containers() *[]Container {
	values := append([]Container(nil), d.containers...)
	return &values
}

func (d *daemonSetImpl) InitContainers() *[]Container {
	values := append([]Container(nil), d.initContainers...)
	return &values
}

func (d *daemonSetImpl) Volumes() *[]Volume {
	values := append([]Volume(nil), d.volumes...)
	return &values
}

func (d *daemonSetImpl) AddContainer(props *ContainerProps) Container {
	return d.addContainer(props)
}

func (d *daemonSetImpl) AddInitContainer(props *ContainerProps) Container {
	return d.addInitContainer(props)
}

func (d *daemonSetImpl) AddVolume(volume Volume) {
	d.addVolume(volume)
}

func (d *daemonSetImpl) AddHostAlias(alias *HostAlias) {
	if alias == nil || alias.Ip == nil || alias.Hostnames == nil {
		panic("host alias IP and hostnames are required")
	}
	d.hostAliases = append(d.hostAliases, alias)
}

func (d *daemonSetImpl) AttachContainer(container Container) {
	if container == nil {
		panic("container is required")
	}
	d.containers = append(d.containers, container)
}

func (d *daemonSetImpl) PodMetadata() cdk8s.ApiObjectMetadataDefinition {
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

func (d *daemonSetImpl) ToPodSelectorConfig() *PodSelectorConfig {
	labels := map[string]*string{}
	for key, value := range d.selector {
		labels[key] = value
	}
	return &PodSelectorConfig{LabelSelector: newLabelSelectorFromRequirements(d.matchExpressions, &labels)}
}

func (d *daemonSetImpl) ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig {
	return &NetworkPolicyPeerConfig{PodSelector: d.ToPodSelectorConfig()}
}

func (d *daemonSetImpl) ToPodSelector() IPodSelector {
	return d
}

func (d *daemonSetImpl) Select(selectors ...LabelSelector) {
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

func (d *daemonSetImpl) MatchLabels() *map[string]*string {
	values := map[string]*string{}
	for key, value := range d.selector {
		values[key] = value
	}
	return &values
}

func (d *daemonSetImpl) MatchExpressions() *[]*LabelSelectorRequirement {
	values := append([]*LabelSelectorRequirement(nil), d.matchExpressions...)
	return &values
}

func (d *daemonSetImpl) Connections() PodConnections {
	return d.connections
}

func (d *daemonSetImpl) Scheduling() WorkloadScheduling {
	return d.scheduling
}

func (d *daemonSetImpl) AutomountServiceAccountToken() *bool {
	if d.props.AutomountServiceAccountToken == nil {
		return jsii.Bool(false)
	}
	return d.props.AutomountServiceAccountToken
}

func (d *daemonSetImpl) Dns() PodDns {
	return d.dns
}

func (d *daemonSetImpl) DockerRegistryAuth() ISecret {
	return d.props.DockerRegistryAuth
}

func (d *daemonSetImpl) EnableServiceLinks() *bool {
	return d.props.EnableServiceLinks
}

func (d *daemonSetImpl) HostAliases() *[]*HostAlias {
	values := append([]*HostAlias(nil), d.hostAliases...)
	return &values
}

func (d *daemonSetImpl) HostNetwork() *bool {
	if d.props.HostNetwork == nil {
		return jsii.Bool(false)
	}
	return d.props.HostNetwork
}

func (d *daemonSetImpl) Isolate() *bool {
	if d.props.Isolate == nil {
		return jsii.Bool(false)
	}
	return d.props.Isolate
}

func (d *daemonSetImpl) RestartPolicy() RestartPolicy {
	if d.props.RestartPolicy == "" {
		return RestartPolicy_ALWAYS
	}
	return d.props.RestartPolicy
}

func (d *daemonSetImpl) SecurityContext() PodSecurityContext {
	return d.security
}

func (d *daemonSetImpl) ServiceAccount() IServiceAccount {
	return d.props.ServiceAccount
}

func (d *daemonSetImpl) ShareProcessNamespace() *bool {
	if d.props.ShareProcessNamespace == nil {
		return jsii.Bool(false)
	}
	return d.props.ShareProcessNamespace
}

func (d *daemonSetImpl) TerminationGracePeriod() cdk8s.Duration {
	if d.props.TerminationGracePeriod == nil {
		return cdk8s.Duration_Seconds(jsii.Number(30))
	}
	return d.props.TerminationGracePeriod
}

func (d *daemonSetImpl) ToSubjectConfiguration() *SubjectConfiguration {
	if d.props.ServiceAccount == nil && !*d.AutomountServiceAccountToken() {
		panic(stringValue(d.Name()) + " cannot be converted to a role binding subject: You must either assign a service account to it, or use 'automountServiceAccountToken: true'")
	}
	name := jsii.String("default")
	if d.props.ServiceAccount != nil {
		name = d.props.ServiceAccount.ResourceName()
	}
	return &SubjectConfiguration{ApiGroup: jsii.String(""), Kind: jsii.String("ServiceAccount"), Name: name}
}

func (d *daemonSetImpl) toManifest() map[string]interface{} {
	spec := d.podState.manifest(d.RestartPolicy())
	for key, value := range d.scheduling.toManifest() {
		spec[key] = value
	}
	return map[string]interface{}{"minReadySeconds": d.MinReadySeconds(), "selector": d.workloadSelector(), "template": map[string]interface{}{"metadata": d.PodMetadata().ToJson(), "spec": spec}}
}

func (d *daemonSetImpl) workloadSelector() map[string]interface{} {
	result := map[string]interface{}{"matchLabels": d.selector}
	if len(d.matchExpressions) > 0 {
		result["matchExpressions"] = d.matchExpressions
	}
	return result
}

func daemonSetPodProps(p *DaemonSetProps) *PodProps {
	return &PodProps{Metadata: p.Metadata, AutomountServiceAccountToken: p.AutomountServiceAccountToken, Containers: p.Containers, Dns: p.Dns, DockerRegistryAuth: p.DockerRegistryAuth, EnableServiceLinks: p.EnableServiceLinks, HostAliases: p.HostAliases, HostNetwork: p.HostNetwork, InitContainers: p.InitContainers, Isolate: p.Isolate, RestartPolicy: p.RestartPolicy, SecurityContext: p.SecurityContext, ServiceAccount: p.ServiceAccount, ShareProcessNamespace: p.ShareProcessNamespace, TerminationGracePeriod: p.TerminationGracePeriod, Volumes: p.Volumes}
}
