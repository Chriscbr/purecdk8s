package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Properties for `DaemonSet`.
type DaemonSetProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Indicates whether a service account token should be automatically mounted. See: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#use-the-default-service-account-to-access-the-api-server
	//
	// Default: false.
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
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
	// Indicates whether information about services should be injected into pod's environment variables, matching the syntax of Docker links. See: https://kubernetes.io/docs/concepts/services-networking/connect-applications-service/#accessing-the-service
	//
	// Default: true.
	EnableServiceLinks *bool `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	// HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the pod's hosts file.
	HostAliases *[]*HostAlias `field:"optional" json:"hostAliases" yaml:"hostAliases"`
	// Host network for the pod. Default: false.
	HostNetwork *bool `field:"optional" json:"hostNetwork" yaml:"hostNetwork"`
	// List of initialization containers belonging to the pod.
	//
	// Init containers are executed in order prior to containers being started. If any init container fails, the pod is considered to have failed and is handled according to its restartPolicy. The name for an init container or normal container must be unique among all containers. Init containers may not have Lifecycle actions, Readiness probes, Liveness probes, or Startup probes. The resourceRequirements of an init container are taken into account during scheduling by finding the highest request/limit for each resource type, and then using the max of of that value or the sum of the normal containers. Limits are applied to init containers in a similar fashion.
	//
	// Init containers cannot currently be added ,removed or updated. See: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	//
	// Default: - No init containers.
	InitContainers *[]*ContainerProps `field:"optional" json:"initContainers" yaml:"initContainers"`
	// Isolates the pod.
	//
	// This will prevent any ingress or egress connections to / from this pod. You can however allow explicit connections post instantiation by using the `.connections` property. Default: false.
	Isolate *bool `field:"optional" json:"isolate" yaml:"isolate"`
	// Minimum number of seconds for which a newly created pod should be ready without any of its container crashing, for it to be considered available. Default: 0.
	MinReadySeconds *float64 `field:"optional" json:"minReadySeconds" yaml:"minReadySeconds"`
	// The pod metadata of this workload.
	PodMetadata *cdk8s.ApiObjectMetadata `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	// Restart policy for all containers within the pod. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy
	//
	// Default: RestartPolicy.ALWAYS
	RestartPolicy RestartPolicy `field:"optional" json:"restartPolicy" yaml:"restartPolicy"`
	// SecurityContext holds pod-level security attributes and common container settings. Default: fsGroupChangePolicy: FsGroupChangePolicy.FsGroupChangePolicy.ALWAYS ensureNonRoot: true.
	SecurityContext *PodSecurityContextProps `field:"optional" json:"securityContext" yaml:"securityContext"`
	// Automatically allocates a pod label selector for this workload and add it to the pod metadata.
	//
	// This ensures this workload manages pods created by its pod template. Default: true.
	Select *bool `field:"optional" json:"select" yaml:"select"`
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
	// Automatically spread pods across hostname and zones. See: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/#internal-default-constraints
	//
	// Default: false.
	Spread *bool `field:"optional" json:"spread" yaml:"spread"`
	// Grace period until the pod is terminated. Default: Duration.seconds(30)
	TerminationGracePeriod cdk8s.Duration `field:"optional" json:"terminationGracePeriod" yaml:"terminationGracePeriod"`
	// List of volumes that can be mounted by containers belonging to the pod.
	//
	// You can also add volumes later using `podSpec.addVolume()` See: https://kubernetes.io/docs/concepts/storage/volumes
	//
	// Default: - No volumes.
	Volumes *[]Volume `field:"optional" json:"volumes" yaml:"volumes"`
}

// A DaemonSet ensures that all (or some) Nodes run a copy of a Pod.
//
// As nodes are added to the cluster, Pods are added to them. As nodes are removed from the cluster, those Pods are garbage collected. Deleting a DaemonSet will clean up the Pods it created.
//
// Some typical uses of a DaemonSet are:
//
// - running a cluster storage daemon on every node - running a logs collection daemon on every node - running a node monitoring daemon on every node
//
// In a simple case, one DaemonSet, covering all nodes, would be used for each type of daemon. A more complex setup might use multiple DaemonSets for a single type of daemon, but with different flags and/or different memory and cpu requests for different hardware types.
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

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
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
