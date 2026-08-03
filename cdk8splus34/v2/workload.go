package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Properties for `Workload`.
type WorkloadProps struct {
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
	// List of volumes that can be mounted by containers belonging to the pod.
	//
	// You can also add volumes later using `podSpec.addVolume()` See: https://kubernetes.io/docs/concepts/storage/volumes
	//
	// Default: - No volumes.
	Volumes *[]Volume `field:"optional" json:"volumes" yaml:"volumes"`
	// The pod metadata of this workload.
	PodMetadata *cdk8s.ApiObjectMetadata `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	// Automatically allocates a pod label selector for this workload and add it to the pod metadata.
	//
	// This ensures this workload manages pods created by its pod template. Default: true.
	Select *bool `field:"optional" json:"select" yaml:"select"`
	// Automatically spread pods across hostname and zones. See: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/#internal-default-constraints
	//
	// Default: false.
	Spread *bool `field:"optional" json:"spread" yaml:"spread"`
}

func (p *WorkloadProps) podProps() *PodProps {
	if p == nil {
		return &PodProps{}
	}
	return &PodProps{
		Metadata: p.Metadata, AutomountServiceAccountToken: p.AutomountServiceAccountToken,
		Containers: p.Containers, Dns: p.Dns, DockerRegistryAuth: p.DockerRegistryAuth,
		EnableServiceLinks: p.EnableServiceLinks, HostAliases: p.HostAliases, HostNetwork: p.HostNetwork,
		InitContainers: p.InitContainers, Isolate: p.Isolate, RestartPolicy: p.RestartPolicy,
		SecurityContext: p.SecurityContext, ServiceAccount: p.ServiceAccount,
		ShareProcessNamespace: p.ShareProcessNamespace, TerminationGracePeriod: p.TerminationGracePeriod,
		Volumes: p.Volumes,
	}
}

type workloadOwner interface {
	Resource
	IPodSelector
}

type workloadState struct {
	owner            workloadOwner
	podMetadata      *cdk8s.ApiObjectMetadata
	selector         map[string]*string
	matchExpressions []*LabelSelectorRequirement
	omitEmptyLabels  bool
}

func newWorkloadState(podMetadata *cdk8s.ApiObjectMetadata, omitEmptyLabels bool) workloadState {
	return workloadState{podMetadata: podMetadata, selector: map[string]*string{}, omitEmptyLabels: omitEmptyLabels}
}

func (w *workloadState) PodMetadata() cdk8s.ApiObjectMetadataDefinition {
	metadata := w.podMetadata
	if metadata == nil {
		metadata = &cdk8s.ApiObjectMetadata{}
	}
	result := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{ApiObject: w.owner.ApiObject(), Name: metadata.Name, Namespace: metadata.Namespace, Labels: metadata.Labels, Annotations: metadata.Annotations})
	for key, value := range w.selector {
		result.AddLabel(jsii.String(key), value)
	}
	return result
}

func (w *workloadState) ToPodSelectorConfig() *PodSelectorConfig {
	labels := map[string]*string{}
	for key, value := range w.selector {
		labels[key] = value
	}
	return &PodSelectorConfig{LabelSelector: newLabelSelectorFromRequirements(w.matchExpressions, &labels)}
}

func (w *workloadState) ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig {
	return &NetworkPolicyPeerConfig{PodSelector: w.ToPodSelectorConfig()}
}

func (w *workloadState) ToPodSelector() IPodSelector {
	return w.owner
}

func (w *workloadState) Select(selectors ...LabelSelector) {
	for _, selector := range selectors {
		if selector == nil {
			panic("selector is required")
		}
		for key, value := range labelSelectorLabels(selector) {
			w.selector[key] = value
		}
		w.matchExpressions = append(w.matchExpressions, labelSelectorRequirements(selector)...)
	}
}

func (w *workloadState) MatchLabels() *map[string]*string {
	values := map[string]*string{}
	for key, value := range w.selector {
		values[key] = value
	}
	return &values
}

func (w *workloadState) MatchExpressions() *[]*LabelSelectorRequirement {
	values := append([]*LabelSelectorRequirement(nil), w.matchExpressions...)
	return &values
}

func (w *workloadState) workloadSelector() map[string]interface{} {
	result := map[string]interface{}{}
	if !w.omitEmptyLabels || len(w.selector) > 0 {
		result["matchLabels"] = w.selector
	}
	if len(w.matchExpressions) > 0 {
		result["matchExpressions"] = w.matchExpressions
	}
	return result
}

type scalableOwner interface {
	Resource
	Containers() *[]Container
}

type scalableState struct {
	owner         scalableOwner
	replicas      *float64
	hasAutoscaler bool
}

func (s *scalableState) Replicas() *float64 {
	return s.replicas
}

func (s *scalableState) HasAutoscaler() *bool {
	return jsii.Bool(s.hasAutoscaler)
}

func (s *scalableState) SetHasAutoscaler(value *bool) {
	s.hasAutoscaler = value != nil && *value
}

func (s *scalableState) MarkHasAutoscaler() {
	s.hasAutoscaler = true
}

func (s *scalableState) ToScalingTarget() *ScalingTarget {
	return &ScalingTarget{
		ApiVersion: s.owner.ApiVersion(),
		Containers: s.owner.Containers(),
		Kind:       s.owner.Kind(),
		Name:       s.owner.Name(),
		Replicas:   s.replicas,
	}
}
