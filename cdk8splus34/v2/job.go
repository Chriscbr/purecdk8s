package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Properties for `Job`.
type JobProps struct {
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
	// Specifies the duration the job may be active before the system tries to terminate it. Default: - If unset, then there is no deadline.
	ActiveDeadline cdk8s.Duration `field:"optional" json:"activeDeadline" yaml:"activeDeadline"`
	// Specifies the number of retries before marking this job failed. Default: - If not set, system defaults to 6.
	BackoffLimit *float64 `field:"optional" json:"backoffLimit" yaml:"backoffLimit"`
	// Limits the lifetime of a Job that has finished execution (either Complete or Failed).
	//
	// If this field is set, after the Job finishes, it is eligible to be automatically deleted. When the Job is being deleted, its lifecycle guarantees (e.g. finalizers) will be honored. If this field is set to zero, the Job becomes eligible to be deleted immediately after it finishes. This field is alpha-level and is only honored by servers that enable the `TTLAfterFinished` feature. Default: - If this field is unset, the Job won't be automatically deleted.
	TtlAfterFinished cdk8s.Duration `field:"optional" json:"ttlAfterFinished" yaml:"ttlAfterFinished"`
}

// A Job creates one or more Pods and ensures that a specified number of them successfully terminate.
//
// As pods successfully complete, the Job tracks the successful completions. When a specified number of successful completions is reached, the task (ie, Job) is complete. Deleting a Job will clean up the Pods it created. A simple case is to create one Job object in order to reliably run one Pod to completion. The Job object will start a new Pod if the first Pod fails or is deleted (for example due to a node hardware failure or a node reboot). You can also use a Job to run multiple Pods in parallel.
type Job interface {
	Workload
	// Duration before job is terminated.
	//
	// If undefined, there is no deadline.
	ActiveDeadline() cdk8s.Duration
	// Number of retries before marking failed.
	BackoffLimit() *float64
	// TTL before the job is deleted after it is finished.
	TtlAfterFinished() cdk8s.Duration
}

type jobImpl struct {
	resourceBase
	podState
	podMetadata      *cdk8s.ApiObjectMetadata
	selector         map[string]*string
	matchExpressions []*LabelSelectorRequirement
	scheduling       WorkloadScheduling
	connections      PodConnections
	activeDeadline   cdk8s.Duration
	backoffLimit     *float64
	ttlAfterFinished cdk8s.Duration
}

func NewJob(scope constructs.Construct, id *string, props *JobProps) Job {
	if props == nil {
		props = &JobProps{}
	}
	result := &jobImpl{podState: newPodState(jobPodProps(props)), podMetadata: props.PodMetadata, selector: map[string]*string{}, activeDeadline: props.ActiveDeadline, backoffLimit: props.BackoffLimit, ttlAfterFinished: props.TtlAfterFinished}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "batch/v1", "Job", "jobs", props.Metadata, manifest)
	selectPods := false
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

func NewJob_Override(job Job, scope constructs.Construct, id *string, props *JobProps) {
	applyOverride(job, NewJob(scope, id, props), "Job")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Job_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (j *jobImpl) ActiveDeadline() cdk8s.Duration {
	return j.activeDeadline
}

func (j *jobImpl) BackoffLimit() *float64 {
	return j.backoffLimit
}

func (j *jobImpl) TtlAfterFinished() cdk8s.Duration {
	return j.ttlAfterFinished
}

func (j *jobImpl) Containers() *[]Container {
	values := append([]Container(nil), j.containers...)
	return &values
}

func (j *jobImpl) InitContainers() *[]Container {
	values := append([]Container(nil), j.initContainers...)
	return &values
}

func (j *jobImpl) Volumes() *[]Volume {
	values := append([]Volume(nil), j.volumes...)
	return &values
}

func (j *jobImpl) AddContainer(props *ContainerProps) Container {
	return j.addContainer(props)
}

func (j *jobImpl) AddInitContainer(props *ContainerProps) Container {
	return j.addInitContainer(props)
}

func (j *jobImpl) AddVolume(volume Volume) {
	j.addVolume(volume)
}

func (j *jobImpl) AddHostAlias(alias *HostAlias) {
	if alias == nil || alias.Ip == nil || alias.Hostnames == nil {
		panic("host alias IP and hostnames are required")
	}
	j.hostAliases = append(j.hostAliases, alias)
}

func (j *jobImpl) AttachContainer(container Container) {
	if container == nil {
		panic("container is required")
	}
	j.containers = append(j.containers, container)
}

func (j *jobImpl) PodMetadata() cdk8s.ApiObjectMetadataDefinition {
	metadata := j.podMetadata
	if metadata == nil {
		metadata = &cdk8s.ApiObjectMetadata{}
	}
	result := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{ApiObject: j.ApiObject(), Name: metadata.Name, Namespace: metadata.Namespace, Labels: metadata.Labels, Annotations: metadata.Annotations})
	result.AddLabel(jsii.String(podAddressLabel), cdk8s.Names_ToLabelValue(j, nil))
	return result
}

func (j *jobImpl) ToPodSelectorConfig() *PodSelectorConfig {
	labels := map[string]*string{podAddressLabel: cdk8s.Names_ToLabelValue(j, nil)}
	config := &PodSelectorConfig{LabelSelector: labelSelectorFromLabels(&labels)}
	if namespace := j.Metadata().Namespace(); namespace != nil {
		config.Namespaces = &NamespaceSelectorConfig{Names: &[]*string{namespace}}
	}
	return config
}

func (j *jobImpl) ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig {
	return &NetworkPolicyPeerConfig{PodSelector: j.ToPodSelectorConfig()}
}

func (j *jobImpl) ToPodSelector() IPodSelector {
	return j
}

func (j *jobImpl) Select(selectors ...LabelSelector) {
	for _, selector := range selectors {
		if selector == nil {
			panic("selector is required")
		}
		for key, value := range labelSelectorLabels(selector) {
			j.selector[key] = value
		}
		j.matchExpressions = append(j.matchExpressions, labelSelectorRequirements(selector)...)
	}
}

func (j *jobImpl) MatchLabels() *map[string]*string {
	values := map[string]*string{}
	for key, value := range j.selector {
		values[key] = value
	}
	return &values
}

func (j *jobImpl) MatchExpressions() *[]*LabelSelectorRequirement {
	values := append([]*LabelSelectorRequirement(nil), j.matchExpressions...)
	return &values
}

func (j *jobImpl) Connections() PodConnections {
	return j.connections
}

func (j *jobImpl) Scheduling() WorkloadScheduling {
	return j.scheduling
}

func (j *jobImpl) AutomountServiceAccountToken() *bool {
	if j.props.AutomountServiceAccountToken == nil {
		return jsii.Bool(false)
	}
	return j.props.AutomountServiceAccountToken
}

func (j *jobImpl) Dns() PodDns {
	return j.dns
}

func (j *jobImpl) DockerRegistryAuth() ISecret {
	return j.props.DockerRegistryAuth
}

func (j *jobImpl) EnableServiceLinks() *bool {
	return j.props.EnableServiceLinks
}

func (j *jobImpl) HostAliases() *[]*HostAlias {
	values := append([]*HostAlias(nil), j.hostAliases...)
	return &values
}

func (j *jobImpl) HostNetwork() *bool {
	if j.props.HostNetwork == nil {
		return jsii.Bool(false)
	}
	return j.props.HostNetwork
}

func (j *jobImpl) Isolate() *bool {
	if j.props.Isolate == nil {
		return jsii.Bool(false)
	}
	return j.props.Isolate
}

func (j *jobImpl) RestartPolicy() RestartPolicy {
	if j.props.RestartPolicy != "" {
		return j.props.RestartPolicy
	}
	return RestartPolicy_NEVER
}

func (j *jobImpl) SecurityContext() PodSecurityContext {
	return j.security
}

func (j *jobImpl) ServiceAccount() IServiceAccount {
	return j.props.ServiceAccount
}

func (j *jobImpl) ShareProcessNamespace() *bool {
	if j.props.ShareProcessNamespace == nil {
		return jsii.Bool(false)
	}
	return j.props.ShareProcessNamespace
}

func (j *jobImpl) TerminationGracePeriod() cdk8s.Duration {
	if j.props.TerminationGracePeriod == nil {
		return cdk8s.Duration_Seconds(jsii.Number(30))
	}
	return j.props.TerminationGracePeriod
}

func (j *jobImpl) ToSubjectConfiguration() *SubjectConfiguration {
	if j.props.ServiceAccount == nil && !*j.AutomountServiceAccountToken() {
		panic(stringValue(j.Name()) + " cannot be converted to a role binding subject: You must either assign a service account to it, or use 'automountServiceAccountToken: true'")
	}
	name := jsii.String("default")
	if j.props.ServiceAccount != nil {
		name = j.props.ServiceAccount.ResourceName()
	}
	return &SubjectConfiguration{ApiGroup: jsii.String(""), Kind: jsii.String("ServiceAccount"), Name: name}
}

func (j *jobImpl) toManifest() map[string]interface{} {
	spec := j.podState.manifest(j.RestartPolicy())
	for key, value := range j.scheduling.toManifest() {
		spec[key] = value
	}
	result := map[string]interface{}{"template": map[string]interface{}{"metadata": j.PodMetadata().ToJson(), "spec": spec}}
	if j.backoffLimit != nil {
		result["backoffLimit"] = j.backoffLimit
	}
	if j.activeDeadline != nil {
		result["activeDeadlineSeconds"] = j.activeDeadline.ToSeconds(nil)
	}
	if j.ttlAfterFinished != nil {
		result["ttlSecondsAfterFinished"] = j.ttlAfterFinished.ToSeconds(nil)
	}
	return result
}

func jobPodProps(p *JobProps) *PodProps {
	restartPolicy := p.RestartPolicy
	if restartPolicy == "" {
		restartPolicy = RestartPolicy_NEVER
	}
	return &PodProps{Metadata: p.Metadata, AutomountServiceAccountToken: p.AutomountServiceAccountToken, Containers: p.Containers, Dns: p.Dns, DockerRegistryAuth: p.DockerRegistryAuth, EnableServiceLinks: p.EnableServiceLinks, HostAliases: p.HostAliases, HostNetwork: p.HostNetwork, InitContainers: p.InitContainers, Isolate: p.Isolate, RestartPolicy: restartPolicy, SecurityContext: p.SecurityContext, ServiceAccount: p.ServiceAccount, ShareProcessNamespace: p.ShareProcessNamespace, TerminationGracePeriod: p.TerminationGracePeriod, Volumes: p.Volumes}
}
