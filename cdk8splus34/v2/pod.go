package cdk8splus34

import (
	"sort"
	"strconv"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Restart policy for all containers within the pod.
type RestartPolicy string

const (
	// Always restart the pod after it exits.
	RestartPolicy_ALWAYS RestartPolicy = "ALWAYS"
	// Never restart the pod.
	RestartPolicy_NEVER RestartPolicy = "NEVER"
	// Only restart if the pod exits with a non-zero exit code.
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

// Properties for `Pod`.
type PodProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// List of containers belonging to the pod.
	//
	// Containers cannot currently be added or removed. There must be at least one container in a Pod.
	//
	// You can add additionnal containers using `podSpec.addContainer()` Default: - No containers. Note that a pod spec must include at least one container.
	Containers *[]*ContainerProps `field:"optional" json:"containers" yaml:"containers"`
	// DNS settings for the pod. See: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/
	//
	// Default: policy: DnsPolicy.CLUSTER_FIRST hostnameAsFQDN: false.
	Dns *PodDnsProps `field:"optional" json:"dns" yaml:"dns"`
	// A secret containing docker credentials for authenticating to a registry. Default: - No auth. Images are assumed to be publicly available.
	DockerRegistryAuth ISecret `field:"optional" json:"dockerRegistryAuth" yaml:"dockerRegistryAuth"`
	// List of initialization containers belonging to the pod.
	//
	// Init containers are executed in order prior to containers being started. If any init container fails, the pod is considered to have failed and is handled according to its restartPolicy. The name for an init container or normal container must be unique among all containers. Init containers may not have Lifecycle actions, Readiness probes, Liveness probes, or Startup probes. The resourceRequirements of an init container are taken into account during scheduling by finding the highest request/limit for each resource type, and then using the max of of that value or the sum of the normal containers. Limits are applied to init containers in a similar fashion.
	//
	// Init containers cannot currently be added ,removed or updated. See: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	//
	// Default: - No init containers.
	InitContainers *[]*ContainerProps `field:"optional" json:"initContainers" yaml:"initContainers"`
	// List of volumes that can be mounted by containers belonging to the pod.
	//
	// You can also add volumes later using `podSpec.addVolume()` See: https://kubernetes.io/docs/concepts/storage/volumes
	//
	// Default: - No volumes.
	Volumes *[]Volume `field:"optional" json:"volumes" yaml:"volumes"`
	// Indicates whether a service account token should be automatically mounted. See: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#use-the-default-service-account-to-access-the-api-server
	//
	// Default: false.
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	// Indicates whether information about services should be injected into pod's environment variables, matching the syntax of Docker links. See: https://kubernetes.io/docs/concepts/services-networking/connect-applications-service/#accessing-the-service
	//
	// Default: true.
	EnableServiceLinks *bool `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	// HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the pod's hosts file.
	HostAliases *[]*HostAlias `field:"optional" json:"hostAliases" yaml:"hostAliases"`
	// Host network for the pod. Default: false.
	HostNetwork *bool `field:"optional" json:"hostNetwork" yaml:"hostNetwork"`
	// Isolates the pod.
	//
	// This will prevent any ingress or egress connections to / from this pod. You can however allow explicit connections post instantiation by using the `.connections` property. Default: false.
	Isolate *bool `field:"optional" json:"isolate" yaml:"isolate"`
	// Restart policy for all containers within the pod. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy
	//
	// Default: RestartPolicy.ALWAYS
	RestartPolicy RestartPolicy `field:"optional" json:"restartPolicy" yaml:"restartPolicy"`
	// SecurityContext holds pod-level security attributes and common container settings. Default: fsGroupChangePolicy: FsGroupChangePolicy.FsGroupChangePolicy.ALWAYS ensureNonRoot: true.
	SecurityContext *PodSecurityContextProps `field:"optional" json:"securityContext" yaml:"securityContext"`
	// A service account provides an identity for processes that run in a Pod.
	//
	// When you (a human) access the cluster (for example, using kubectl), you are authenticated by the apiserver as a particular User Account (currently this is usually admin, unless your cluster administrator has customized your cluster). Processes in containers inside pods can also contact the apiserver. When they do, they are authenticated as a particular Service Account (for example, default). See: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	//
	// Default: - No service account.
	ServiceAccount IServiceAccount `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	// When process namespace sharing is enabled, processes in a container are visible to all other containers in the same pod. See: https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/
	//
	// Default: false.
	ShareProcessNamespace *bool `field:"optional" json:"shareProcessNamespace" yaml:"shareProcessNamespace"`
	// Grace period until the pod is terminated. Default: Duration.seconds(30)
	TerminationGracePeriod cdk8s.Duration `field:"optional" json:"terminationGracePeriod" yaml:"terminationGracePeriod"`
}

// Properties for `AbstractPod`.
type AbstractPodProps = PodProps

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
	// Return the configuration of this peer. See: INetworkPolicyPeer.toNetworkPolicyPeerConfig ()
	ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig
	// Convert the peer into a pod selector, if possible. See: INetworkPolicyPeer.toPodSelector ()
	ToPodSelector() IPodSelector
	// Return the configuration of this selector. See: IPodSelector.toPodSelectorConfig ()
	ToPodSelectorConfig() *PodSelectorConfig
	// Return the subject configuration. See: ISubect.toSubjectConfiguration ()
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
	if props.RestartPolicy != ContainerRestartPolicy_ALWAYS {
		if props.Readiness != nil {
			panic("Init containers must not have a readiness probe")
		}
		if props.Liveness != nil {
			panic("Init containers must not have a liveness probe")
		}
		if props.Startup != nil {
			panic("Init containers must not have a startup probe")
		}
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
	for _, container := range p.containers {
		if container.RestartPolicy() != "" {
			panic("Invalid container spec: " + stringValue(container.Name()) + " has non-empty restartPolicy field. The field can only be specified for initContainers")
		}
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

// Pod is a collection of containers that can run on a host.
//
// This resource is created by clients and scheduled onto hosts.
type Pod interface {
	AbstractPod
	Containers() *[]Container
	InitContainers() *[]Container
	Volumes() *[]Volume
	AddContainer(props *ContainerProps) Container
	AddInitContainer(props *ContainerProps) Container
	AddVolume(volume Volume)
	// Return the configuration of this selector. See: IPodSelector.toPodSelectorConfig ()
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
	podAddress := cdk8s.Names_ToLabelValue(result, nil)
	result.selector[podAddressLabel] = podAddress
	result.Metadata().AddLabel(jsii.String(podAddressLabel), podAddress)
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

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
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
	config := &PodSelectorConfig{LabelSelector: labelSelectorFromLabels(&labels)}
	if namespace := p.Metadata().Namespace(); namespace != nil {
		config.Namespaces = &NamespaceSelectorConfig{Names: &[]*string{namespace}}
	}
	return config
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

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func AbstractPod_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func NewAbstractPod_Override(_ AbstractPod, _ constructs.Construct, _ *string, _ *AbstractPodProps) {
	panic("AbstractPod is an abstract base; use NewPod, NewDeployment, NewStatefulSet, or NewDaemonSet")
}
