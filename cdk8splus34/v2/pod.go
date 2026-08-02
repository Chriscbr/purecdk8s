package cdk8splus34

import (
	"sort"
	"strconv"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// RestartPolicy controls what Kubernetes does when a container exits.
type RestartPolicy string

const (
	RestartPolicy_ALWAYS     RestartPolicy = "ALWAYS"
	RestartPolicy_NEVER      RestartPolicy = "NEVER"
	RestartPolicy_ON_FAILURE RestartPolicy = "ON_FAILURE"
)

func restartPolicyManifestValue(value RestartPolicy) string {
	switch value {
	case RestartPolicy_ALWAYS:
		return "Always"
	case RestartPolicy_NEVER:
		return "Never"
	case RestartPolicy_ON_FAILURE:
		return "OnFailure"
	default:
		return string(value)
	}
}

// PodProps contains the common pod fields used by native workload constructs.
type PodProps struct {
	Metadata                     *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Containers                   *[]*ContainerProps       `field:"optional" json:"containers" yaml:"containers"`
	Dns                          *PodDnsProps             `field:"optional" json:"dns" yaml:"dns"`
	DockerRegistryAuth           ISecret                  `field:"optional" json:"dockerRegistryAuth" yaml:"dockerRegistryAuth"`
	InitContainers               *[]*ContainerProps       `field:"optional" json:"initContainers" yaml:"initContainers"`
	Volumes                      *[]Volume                `field:"optional" json:"volumes" yaml:"volumes"`
	AutomountServiceAccountToken *bool                    `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	EnableServiceLinks           *bool                    `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	HostAliases                  *[]*HostAlias            `field:"optional" json:"hostAliases" yaml:"hostAliases"`
	HostNetwork                  *bool                    `field:"optional" json:"hostNetwork" yaml:"hostNetwork"`
	Isolate                      *bool                    `field:"optional" json:"isolate" yaml:"isolate"`
	RestartPolicy                RestartPolicy            `field:"optional" json:"restartPolicy" yaml:"restartPolicy"`
	SecurityContext              *PodSecurityContextProps `field:"optional" json:"securityContext" yaml:"securityContext"`
	ServiceAccount               IServiceAccount          `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ShareProcessNamespace        *bool                    `field:"optional" json:"shareProcessNamespace" yaml:"shareProcessNamespace"`
	TerminationGracePeriod       cdk8s.Duration           `field:"optional" json:"terminationGracePeriod" yaml:"terminationGracePeriod"`
}

// AbstractPodProps configures the shared pod portion of Pod and workload
// resources. PodProps is its concrete counterpart for a standalone Pod.
type AbstractPodProps = PodProps

// AbstractPod is the common behavior shared by Pods and workload pod
// templates. It mirrors the cdk8s+ abstract base without relying on JSII.
type AbstractPod interface {
	Resource
	AutomountServiceAccountToken() *bool
	Containers() *[]Container
	Dns() PodDns
	DockerRegistryAuth() ISecret
	EnableServiceLinks() *bool
	HostAliases() *[]*HostAlias
	HostNetwork() *bool
	InitContainers() *[]Container
	Isolate() *bool
	PodMetadata() cdk8s.ApiObjectMetadataDefinition
	RestartPolicy() RestartPolicy
	SecurityContext() PodSecurityContext
	ServiceAccount() IServiceAccount
	ShareProcessNamespace() *bool
	TerminationGracePeriod() cdk8s.Duration
	Volumes() *[]Volume
	AddContainer(*ContainerProps) Container
	AddHostAlias(*HostAlias)
	AddInitContainer(*ContainerProps) Container
	AddVolume(Volume)
	AttachContainer(Container)
	ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig
	ToPodSelector() IPodSelector
	ToPodSelectorConfig() *PodSelectorConfig
	ToSubjectConfiguration() *SubjectConfiguration
}

type podState struct {
	containers     []Container
	initContainers []Container
	volumes        []Volume
	props          *PodProps
	selector       map[string]*string
	dns            PodDns
	security       PodSecurityContext
	hostAliases    []*HostAlias
}

func newPodState(props *PodProps) podState {
	if props == nil {
		props = &PodProps{}
	}
	state := podState{props: props, selector: map[string]*string{}, dns: NewPodDns(props.Dns), security: NewPodSecurityContext(props.SecurityContext)}
	if props.Containers != nil {
		for _, value := range *props.Containers {
			if value != nil {
				state.addContainer(value)
			}
		}
	}
	if props.InitContainers != nil {
		for _, value := range *props.InitContainers {
			if value != nil {
				state.addInitContainer(value)
			}
		}
	}
	if props.Volumes != nil {
		for _, volume := range *props.Volumes {
			state.addVolume(volume)
		}
	}
	if props.HostAliases != nil {
		state.hostAliases = append(state.hostAliases, (*props.HostAliases)...)
	}
	return state
}

func (p *podState) addContainer(props *ContainerProps) Container {
	container := NewContainer(props)
	p.containers = append(p.containers, container)
	return container
}

func (p *podState) addInitContainer(props *ContainerProps) Container {
	if props == nil {
		panic("container props are required")
	}
	if props.RestartPolicy != ContainerRestartPolicy_ALWAYS && (props.Readiness != nil || props.Liveness != nil || props.Startup != nil) {
		panic("Init containers must not have readiness, liveness, or startup probes")
	}
	if props.Name == nil {
		propsCopy := *props
		propsCopy.Name = jsii.String("init-" + strconv.Itoa(len(p.initContainers)))
		props = &propsCopy
	}
	container := NewContainer(props)
	p.initContainers = append(p.initContainers, container)
	return container
}

func (p *podState) addVolume(volume Volume) {
	if volume == nil {
		panic("volume is required")
	}
	for _, existing := range p.volumes {
		if stringValue(existing.Name()) == stringValue(volume.Name()) && existing != volume {
			panic("Volume with name " + stringValue(volume.Name()) + " already exists")
		}
	}
	p.volumes = append(p.volumes, volume)
}

func containerValues(values []Container) []interface{} {
	result := make([]interface{}, 0, len(values))
	for _, container := range values {
		native, ok := container.(*containerImpl)
		if !ok {
			panic("unsupported native container")
		}
		result = append(result, native.toManifest())
	}
	return result
}

func (p *podState) manifest(restartPolicy RestartPolicy) map[string]interface{} {
	if len(p.containers) == 0 {
		panic("PodSpec must have at least 1 container")
	}
	volumes := map[string]Volume{}
	addVolume := func(volume Volume) {
		name := stringValue(volume.Name())
		if existing, exists := volumes[name]; exists && existing != volume {
			panic("Invalid mount configuration. At least two different volumes have the same name: " + name)
		}
		volumes[name] = volume
	}
	for _, volume := range p.volumes {
		addVolume(volume)
	}
	allContainers := append(append([]Container{}, p.containers...), p.initContainers...)
	for _, container := range allContainers {
		for _, mount := range *container.Mounts() {
			addVolume(mount.Volume)
		}
	}
	dns := p.dns.toManifest()
	spec := map[string]interface{}{
		"automountServiceAccountToken":  false,
		"containers":                    containerValues(p.containers),
		"dnsPolicy":                     dns["dnsPolicy"],
		"hostNetwork":                   false,
		"restartPolicy":                 restartPolicyManifestValue(restartPolicy),
		"securityContext":               p.security.toManifest(),
		"setHostnameAsFQDN":             dns["setHostnameAsFQDN"],
		"shareProcessNamespace":         false,
		"terminationGracePeriodSeconds": 30,
	}
	for _, key := range []string{"dnsConfig", "hostname", "subdomain"} {
		if value, ok := dns[key]; ok {
			spec[key] = value
		}
	}
	if p.props.AutomountServiceAccountToken != nil {
		spec["automountServiceAccountToken"] = p.props.AutomountServiceAccountToken
	}
	if p.props.EnableServiceLinks != nil {
		spec["enableServiceLinks"] = p.props.EnableServiceLinks
	}
	if p.props.HostNetwork != nil {
		spec["hostNetwork"] = p.props.HostNetwork
	}
	if p.props.ShareProcessNamespace != nil {
		spec["shareProcessNamespace"] = p.props.ShareProcessNamespace
	}
	if p.props.ServiceAccount != nil {
		spec["serviceAccountName"] = p.props.ServiceAccount.ResourceName()
	}
	if p.props.DockerRegistryAuth != nil {
		spec["imagePullSecrets"] = []interface{}{map[string]interface{}{"name": p.props.DockerRegistryAuth.ResourceName()}}
	}
	if p.props.TerminationGracePeriod != nil {
		spec["terminationGracePeriodSeconds"] = p.props.TerminationGracePeriod.ToSeconds(nil)
	}
	if len(p.initContainers) != 0 {
		spec["initContainers"] = containerValues(p.initContainers)
	}
	if len(p.hostAliases) != 0 {
		spec["hostAliases"] = p.hostAliases
	}
	if len(volumes) != 0 {
		names := make([]string, 0, len(volumes))
		for name := range volumes {
			names = append(names, name)
		}
		sort.Strings(names)
		entries := make([]interface{}, 0, len(names))
		for _, name := range names {
			volume, ok := volumes[name].(*volumeImpl)
			if !ok {
				panic("unsupported native volume")
			}
			entry := map[string]interface{}{"name": volume.Name()}
			for key, value := range volume.spec {
				entry[key] = value
			}
			entries = append(entries, entry)
		}
		spec["volumes"] = entries
	}
	return spec
}

// Pod is a native Kubernetes Pod construct.
type Pod interface {
	AbstractPod
	Containers() *[]Container
	InitContainers() *[]Container
	Volumes() *[]Volume
	AddContainer(props *ContainerProps) Container
	AddInitContainer(props *ContainerProps) Container
	AddVolume(volume Volume)
	ToPodSelectorConfig() *PodSelectorConfig
	Connections() PodConnections
	Scheduling() PodScheduling
}

type podImpl struct {
	resourceBase
	podState
	connections PodConnections
	scheduling  PodScheduling
}

func NewPod(scope constructs.Construct, id *string, props *PodProps) Pod {
	if props == nil {
		props = &PodProps{}
	}
	result := &podImpl{podState: newPodState(props)}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "v1", "Pod", "pods", props.Metadata, manifest)
	result.selector[podAddressLabel] = cdk8s.Names_ToLabelValue(result, nil)
	result.scheduling = NewPodScheduling(result)
	result.connections = NewPodConnections(result)
	if *result.Isolate() {
		result.connections.Isolate()
	}
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} {
		policy := props.RestartPolicy
		if policy == "" {
			policy = RestartPolicy_ALWAYS
		}
		spec := result.podState.manifest(policy)
		for key, value := range result.scheduling.toManifest() {
			spec[key] = value
		}
		return spec
	}})
	return result
}

func NewPod_Override(pod Pod, scope constructs.Construct, id *string, props *PodProps) {
	applyOverride(pod, NewPod(scope, id, props), "Pod")
}

func Pod_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func Pod_ADDRESS_LABEL() *string {
	return jsii.String(podAddressLabel)
}

func (p *podImpl) Containers() *[]Container {
	values := append([]Container(nil), p.containers...)
	return &values
}

func (p *podImpl) InitContainers() *[]Container {
	values := append([]Container(nil), p.initContainers...)
	return &values
}

func (p *podImpl) Volumes() *[]Volume {
	values := append([]Volume(nil), p.volumes...)
	return &values
}

func (p *podImpl) AddContainer(props *ContainerProps) Container {
	return p.addContainer(props)
}

func (p *podImpl) AddInitContainer(props *ContainerProps) Container {
	return p.addInitContainer(props)
}

func (p *podImpl) AddVolume(volume Volume) {
	p.addVolume(volume)
}

func (p *podImpl) AddHostAlias(alias *HostAlias) {
	if alias == nil || alias.Ip == nil || alias.Hostnames == nil {
		panic("host alias IP and hostnames are required")
	}
	p.hostAliases = append(p.hostAliases, alias)
}

func (p *podImpl) AttachContainer(container Container) {
	if container == nil {
		panic("container is required")
	}
	p.containers = append(p.containers, container)
}

func (p *podImpl) ToPodSelectorConfig() *PodSelectorConfig {
	labels := map[string]*string{}
	for key, value := range p.selector {
		labels[key] = value
	}
	return &PodSelectorConfig{LabelSelector: labelSelectorFromLabels(&labels)}
}

func (p *podImpl) ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig {
	return &NetworkPolicyPeerConfig{PodSelector: p.ToPodSelectorConfig()}
}

func (p *podImpl) ToPodSelector() IPodSelector {
	return p
}

func (p *podImpl) PodMetadata() cdk8s.ApiObjectMetadataDefinition {
	return p.Metadata()
}

func (p *podImpl) ToSubjectConfiguration() *SubjectConfiguration {
	if p.props.ServiceAccount == nil && (p.props.AutomountServiceAccountToken == nil || !*p.props.AutomountServiceAccountToken) {
		panic(stringValue(p.Name()) + " cannot be converted to a role binding subject: You must either assign a service account to it, or use 'automountServiceAccountToken: true'")
	}
	name := jsii.String("default")
	if p.props.ServiceAccount != nil {
		name = p.props.ServiceAccount.ResourceName()
	}
	return &SubjectConfiguration{ApiGroup: jsii.String(""), Kind: jsii.String("ServiceAccount"), Name: name}
}

func (p *podImpl) AutomountServiceAccountToken() *bool {
	if p.props.AutomountServiceAccountToken == nil {
		return jsii.Bool(false)
	}
	return p.props.AutomountServiceAccountToken
}

func (p *podImpl) Dns() PodDns {
	return p.dns
}

func (p *podImpl) DockerRegistryAuth() ISecret {
	return p.props.DockerRegistryAuth
}

func (p *podImpl) EnableServiceLinks() *bool {
	return p.props.EnableServiceLinks
}

func (p *podImpl) HostAliases() *[]*HostAlias {
	values := append([]*HostAlias(nil), p.hostAliases...)
	return &values
}

func (p *podImpl) HostNetwork() *bool {
	if p.props.HostNetwork == nil {
		return jsii.Bool(false)
	}
	return p.props.HostNetwork
}

func (p *podImpl) Isolate() *bool {
	if p.props.Isolate == nil {
		return jsii.Bool(false)
	}
	return p.props.Isolate
}

func (p *podImpl) RestartPolicy() RestartPolicy {
	if p.props.RestartPolicy == "" {
		return RestartPolicy_ALWAYS
	}
	return p.props.RestartPolicy
}

func (p *podImpl) SecurityContext() PodSecurityContext {
	return p.security
}

func (p *podImpl) ServiceAccount() IServiceAccount {
	return p.props.ServiceAccount
}

func (p *podImpl) ShareProcessNamespace() *bool {
	if p.props.ShareProcessNamespace == nil {
		return jsii.Bool(false)
	}
	return p.props.ShareProcessNamespace
}

func (p *podImpl) TerminationGracePeriod() cdk8s.Duration {
	if p.props.TerminationGracePeriod == nil {
		return cdk8s.Duration_Seconds(jsii.Number(30))
	}
	return p.props.TerminationGracePeriod
}

func (p *podImpl) Connections() PodConnections {
	return p.connections
}

func (p *podImpl) Scheduling() PodScheduling {
	return p.scheduling
}

func AbstractPod_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func NewAbstractPod_Override(_ AbstractPod, _ constructs.Construct, _ *string, _ *AbstractPodProps) {
	panic("AbstractPod is an abstract base; use NewPod, NewDeployment, NewStatefulSet, or NewDaemonSet")
}
