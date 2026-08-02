package cdk8splus34

import (
	"fmt"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Properties for initialization of `StatefulSet`.
type StatefulSetProps struct {
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
	// Minimum duration for which a newly created pod should be ready without any of its container crashing, for it to be considered available.
	//
	// Zero means the pod will be considered available as soon as it is ready.
	//
	// This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate. See: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#min-ready-seconds
	//
	// Default: Duration.seconds(0)
	MinReady cdk8s.Duration `field:"optional" json:"minReady" yaml:"minReady"`
	// The pod metadata of this workload.
	PodMetadata *cdk8s.ApiObjectMetadata `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	// Pod management policy to use for this statefulset. Default: PodManagementPolicy.ORDERED_READY
	PodManagementPolicy PodManagementPolicy `field:"optional" json:"podManagementPolicy" yaml:"podManagementPolicy"`
	// Number of desired pods. Default: 1.
	Replicas *float64 `field:"optional" json:"replicas" yaml:"replicas"`
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
	// Service to associate with the statefulset. Default: - A new headless service will be created.
	Service Service `field:"optional" json:"service" yaml:"service"`
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
	// Indicates the StatefulSetUpdateStrategy that will be employed to update Pods in the StatefulSet when a revision is made to Template. Default: - RollingUpdate with partition set to 0.
	Strategy StatefulSetUpdateStrategy `field:"optional" json:"strategy" yaml:"strategy"`
	// Grace period until the pod is terminated. Default: Duration.seconds(30)
	TerminationGracePeriod cdk8s.Duration `field:"optional" json:"terminationGracePeriod" yaml:"terminationGracePeriod"`
	// A list of PersistentVolumeClaim templates that will be created for each pod in the StatefulSet.
	//
	// The StatefulSet controller creates a PVC and a PV for each template based on the pod's ordinal index, ensuring stable storage across pod restarts and rescheduling.
	//
	// Each claim in this list must have at least one matching (by name) volumeMount in one of the containers. Default: - No volume claim templates will be created.
	VolumeClaimTemplates *[]*PersistentVolumeClaimTemplateProps `field:"optional" json:"volumeClaimTemplates" yaml:"volumeClaimTemplates"`
	// List of volumes that can be mounted by containers belonging to the pod.
	//
	// You can also add volumes later using `podSpec.addVolume()` See: https://kubernetes.io/docs/concepts/storage/volumes
	//
	// Default: - No volumes.
	Volumes *[]Volume `field:"optional" json:"volumes" yaml:"volumes"`
}

// StatefulSet is the workload API object used to manage stateful applications.
//
// Manages the deployment and scaling of a set of Pods, and provides guarantees about the ordering and uniqueness of these Pods.
//
// Like a Deployment, a StatefulSet manages Pods that are based on an identical container spec. Unlike a Deployment, a StatefulSet maintains a sticky identity for each of their Pods. These pods are created from the same spec, but are not interchangeable: each has a persistent identifier that it maintains across any rescheduling.
//
// If you want to use storage volumes to provide persistence for your workload, you can use a StatefulSet as part of the solution. Although individual Pods in a StatefulSet are susceptible to failure, the persistent Pod identifiers make it easier to match existing volumes to the new Pods that replace any that have failed.
//
// Using StatefulSets ------------------ StatefulSets are valuable for applications that require one or more of the following.
//
// - Stable, unique network identifiers. - Stable, persistent storage. - Ordered, graceful deployment and scaling. - Ordered, automated rolling updates.
type StatefulSet interface {
	Workload
	IScalable
	// Number of desired pods.
	Replicas() *float64
	Service() Service
	// Management policy to use for the set.
	PodManagementPolicy() PodManagementPolicy
	// Minimum duration for which a newly created pod should be ready without any of its container crashing, for it to be considered available.
	MinReady() cdk8s.Duration
	// The update startegy of this stateful set.
	Strategy() StatefulSetUpdateStrategy
	VolumeClaimTemplates() *[]*PersistentVolumeClaimTemplateProps
	SetVolumeClaimTemplates(*[]*PersistentVolumeClaimTemplateProps)
	AddVolumeClaimTemplate(*PersistentVolumeClaimTemplateProps)
}

type statefulSetImpl struct {
	resourceBase
	podState
	replicas             *float64
	hasAutoscaler        bool
	selector             map[string]*string
	matchExpressions     []*LabelSelectorRequirement
	podMetadata          *cdk8s.ApiObjectMetadata
	service              Service
	scheduling           WorkloadScheduling
	connections          PodConnections
	spread               bool
	strategy             StatefulSetUpdateStrategy
	podManagementPolicy  PodManagementPolicy
	minReady             cdk8s.Duration
	volumeClaimTemplates []*PersistentVolumeClaimTemplateProps
}

func NewStatefulSet(scope constructs.Construct, id *string, props *StatefulSetProps) StatefulSet {
	if scope == nil || id == nil {
		panic("scope and id are required")
	}
	if props == nil {
		props = &StatefulSetProps{}
	}
	result := &statefulSetImpl{
		podState: newPodState(statefulSetPodProps(props)), replicas: props.Replicas, selector: map[string]*string{}, podMetadata: props.PodMetadata,
		spread:              props.Spread != nil && *props.Spread,
		strategy:            props.Strategy,
		podManagementPolicy: props.PodManagementPolicy,
		minReady:            props.MinReady,
	}
	constructs.NewConstruct_Override(result, scope, id)
	selectPods := true
	if props.Select != nil {
		selectPods = *props.Select
	}
	if selectPods {
		result.selector[podAddressLabel] = cdk8s.Names_ToLabelValue(result, nil)
	}
	result.scheduling = NewWorkloadScheduling(result)
	if result.spread {
		result.scheduling.Spread(&WorkloadSchedulingSpreadOptions{Topology: Topology_HOSTNAME()})
		result.scheduling.Spread(&WorkloadSchedulingSpreadOptions{Topology: Topology_ZONE()})
	}
	if props.Service != nil {
		result.service = props.Service
		result.service.Select(result)
	} else {
		result.service = result.createHeadlessService(props.Metadata)
	}
	if props.VolumeClaimTemplates != nil {
		result.SetVolumeClaimTemplates(props.VolumeClaimTemplates)
	}
	manifest := map[string]interface{}{}
	result.resourceBase.initializeApiObject(result, "apps/v1", "StatefulSet", "statefulsets", props.Metadata, manifest)
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toManifest() }})
	result.connections = NewPodConnections(result)
	if props.Isolate != nil && *props.Isolate {
		result.connections.Isolate()
	}
	return result
}

func NewStatefulSet_Override(s StatefulSet, scope constructs.Construct, id *string, props *StatefulSetProps) {
	applyOverride(s, NewStatefulSet(scope, id, props), "StatefulSet")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func StatefulSet_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (s *statefulSetImpl) Containers() *[]Container {
	values := append([]Container(nil), s.podState.containers...)
	return &values
}

func (s *statefulSetImpl) Replicas() *float64 {
	return s.replicas
}

func (s *statefulSetImpl) HasAutoscaler() *bool {
	return jsii.Bool(s.hasAutoscaler)
}

func (s *statefulSetImpl) SetHasAutoscaler(value *bool) {
	s.hasAutoscaler = value != nil && *value
}

func (s *statefulSetImpl) MarkHasAutoscaler() {
	s.hasAutoscaler = true
}

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
	return s.addContainer(props)
}

func (s *statefulSetImpl) Service() Service {
	return s.service
}

func (s *statefulSetImpl) PodManagementPolicy() PodManagementPolicy {
	if s.podManagementPolicy == "" {
		return PodManagementPolicy_ORDERED_READY
	}
	return s.podManagementPolicy
}

func (s *statefulSetImpl) MinReady() cdk8s.Duration {
	if s.minReady == nil {
		return cdk8s.Duration_Seconds(jsii.Number(0))
	}
	return s.minReady
}

func (s *statefulSetImpl) Strategy() StatefulSetUpdateStrategy {
	if s.strategy == nil {
		return StatefulSetUpdateStrategy_RollingUpdate(nil)
	}
	return s.strategy
}

func (s *statefulSetImpl) VolumeClaimTemplates() *[]*PersistentVolumeClaimTemplateProps {
	values := append([]*PersistentVolumeClaimTemplateProps(nil), s.volumeClaimTemplates...)
	return &values
}

func (s *statefulSetImpl) SetVolumeClaimTemplates(templates *[]*PersistentVolumeClaimTemplateProps) {
	s.volumeClaimTemplates = nil
	if templates != nil {
		for _, template := range *templates {
			s.AddVolumeClaimTemplate(template)
		}
	}
}

func (s *statefulSetImpl) AddVolumeClaimTemplate(template *PersistentVolumeClaimTemplateProps) {
	if template == nil || template.Name == nil {
		panic("volume claim template name is required")
	}
	for _, existing := range s.volumeClaimTemplates {
		if stringValue(existing.Name) == stringValue(template.Name) {
			panic("A volume claim template with name \"" + stringValue(template.Name) + "\" already exists")
		}
	}
	s.volumeClaimTemplates = append(s.volumeClaimTemplates, template)
}

func (s *statefulSetImpl) Scheduling() WorkloadScheduling {
	return s.scheduling
}

func (s *statefulSetImpl) Connections() PodConnections {
	return s.connections
}

func (s *statefulSetImpl) ToPodSelectorConfig() *PodSelectorConfig {
	labels := map[string]*string{}
	for key, value := range s.selector {
		labels[key] = value
	}
	return &PodSelectorConfig{LabelSelector: newLabelSelectorFromRequirements(s.matchExpressions, &labels)}
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
	podSpec := s.podState.manifest(s.RestartPolicy())
	for key, value := range s.scheduling.toManifest() {
		podSpec[key] = value
	}
	replicas := s.replicas
	if replicas == nil && !s.hasAutoscaler {
		replicas = jsii.Number(1)
	}
	if len(s.volumeClaimTemplates) > 0 {
		s.filterTemplateVolumes(podSpec)
	}
	result := map[string]interface{}{
		"minReadySeconds":     s.MinReady().ToSeconds(nil),
		"podManagementPolicy": podManagementPolicyManifestValue(s.PodManagementPolicy()),
		"selector":            s.workloadSelector(),
		"serviceName":         s.service.Name(),
		"template":            map[string]interface{}{"metadata": s.PodMetadata().ToJson(), "spec": podSpec},
		"updateStrategy":      s.Strategy().toManifest(),
	}
	if replicas != nil {
		result["replicas"] = replicas
	}
	if len(s.volumeClaimTemplates) > 0 {
		result["volumeClaimTemplates"] = s.claimTemplateManifests()
	}
	return result
}

func (s *statefulSetImpl) filterTemplateVolumes(podSpec map[string]interface{}) {
	values, ok := podSpec["volumes"].([]interface{})
	if !ok {
		return
	}
	templates := map[string]bool{}
	for _, template := range s.volumeClaimTemplates {
		templates[stringValue(template.Name)] = true
	}
	filtered := make([]interface{}, 0, len(values))
	for _, value := range values {
		if volume, ok := value.(map[string]interface{}); ok && templates[stringValue(volume["name"].(*string))] {
			continue
		}
		filtered = append(filtered, value)
	}
	if len(filtered) == 0 {
		delete(podSpec, "volumes")
	} else {
		podSpec["volumes"] = filtered
	}
}

func (s *statefulSetImpl) claimTemplateManifests() []interface{} {
	mounted := map[string]bool{}
	for _, container := range s.containers {
		for _, mount := range *container.Mounts() {
			mounted[stringValue(mount.Volume.Name())] = true
		}
	}
	result := make([]interface{}, 0, len(s.volumeClaimTemplates))
	for _, template := range s.volumeClaimTemplates {
		if !mounted[stringValue(template.Name)] {
			panic("Volume claim template with name \"" + stringValue(template.Name) + "\" is not used by any container mount")
		}
		spec := map[string]interface{}{}
		if template.AccessModes != nil {
			spec["accessModes"] = persistentVolumeAccessModesManifest(*template.AccessModes)
		}
		if template.StorageClassName != nil {
			spec["storageClassName"] = template.StorageClassName
		}
		if template.Storage != nil {
			spec["resources"] = map[string]interface{}{"requests": map[string]interface{}{"storage": sizeGibibytes(template.Storage)}}
		}
		result = append(result, map[string]interface{}{"metadata": map[string]interface{}{"name": template.Name}, "spec": spec})
	}
	return result
}

// Controls how pods are created during initial scale up, when replacing pods on nodes, or when scaling down.
//
// The default policy is `OrderedReady`, where pods are created in increasing order (pod-0, then pod-1, etc) and the controller will wait until each pod is ready before continuing. When scaling down, the pods are removed in the opposite order.
//
// The alternative policy is `Parallel` which will create pods in parallel to match the desired scale without waiting, and on scale down will delete all pods at once.
type PodManagementPolicy string

const (
	PodManagementPolicy_ORDERED_READY PodManagementPolicy = "ORDERED_READY"
	PodManagementPolicy_PARALLEL      PodManagementPolicy = "PARALLEL"
)

func podManagementPolicyManifestValue(value PodManagementPolicy) string {
	switch value {
	case PodManagementPolicy_ORDERED_READY:
		return "OrderedReady"
	case PodManagementPolicy_PARALLEL:
		return "Parallel"
	default:
		return string(value)
	}
}

func statefulSetPodProps(p *StatefulSetProps) *PodProps {
	return &PodProps{Metadata: p.Metadata, AutomountServiceAccountToken: p.AutomountServiceAccountToken, Containers: p.Containers, Dns: p.Dns, DockerRegistryAuth: p.DockerRegistryAuth, EnableServiceLinks: p.EnableServiceLinks, HostAliases: p.HostAliases, HostNetwork: p.HostNetwork, InitContainers: p.InitContainers, Isolate: p.Isolate, RestartPolicy: p.RestartPolicy, SecurityContext: p.SecurityContext, ServiceAccount: p.ServiceAccount, ShareProcessNamespace: p.ShareProcessNamespace, TerminationGracePeriod: p.TerminationGracePeriod, Volumes: p.Volumes}
}

func (s *statefulSetImpl) PodMetadata() cdk8s.ApiObjectMetadataDefinition {
	metadata := s.podMetadata
	if metadata == nil {
		metadata = &cdk8s.ApiObjectMetadata{}
	}
	result := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{ApiObject: s.ApiObject(), Name: metadata.Name, Namespace: metadata.Namespace, Labels: metadata.Labels, Annotations: metadata.Annotations})
	for key, value := range s.selector {
		result.AddLabel(jsii.String(key), value)
	}
	return result
}

func (s *statefulSetImpl) workloadSelector() map[string]interface{} {
	result := map[string]interface{}{"matchLabels": s.selector}
	if len(s.matchExpressions) > 0 {
		result["matchExpressions"] = s.matchExpressions
	}
	return result
}

func (s *statefulSetImpl) MatchLabels() *map[string]*string {
	values := map[string]*string{}
	for key, value := range s.selector {
		values[key] = value
	}
	return &values
}

func (s *statefulSetImpl) MatchExpressions() *[]*LabelSelectorRequirement {
	values := append([]*LabelSelectorRequirement(nil), s.matchExpressions...)
	return &values
}

func (s *statefulSetImpl) Select(selectors ...LabelSelector) {
	for _, selector := range selectors {
		if selector == nil {
			panic("selector is required")
		}
		for key, value := range labelSelectorLabels(selector) {
			s.selector[key] = value
		}
		s.matchExpressions = append(s.matchExpressions, labelSelectorRequirements(selector)...)
	}
}

func (s *statefulSetImpl) InitContainers() *[]Container {
	values := append([]Container(nil), s.podState.initContainers...)
	return &values
}

func (s *statefulSetImpl) Volumes() *[]Volume {
	values := append([]Volume(nil), s.podState.volumes...)
	return &values
}

func (s *statefulSetImpl) AddInitContainer(props *ContainerProps) Container {
	return s.addInitContainer(props)
}

func (s *statefulSetImpl) AddVolume(volume Volume) {
	s.addVolume(volume)
}

func (s *statefulSetImpl) AddHostAlias(alias *HostAlias) {
	if alias == nil || alias.Ip == nil || alias.Hostnames == nil {
		panic("host alias IP and hostnames are required")
	}
	s.hostAliases = append(s.hostAliases, alias)
}

func (s *statefulSetImpl) AttachContainer(container Container) {
	if container == nil {
		panic("container is required")
	}
	s.containers = append(s.containers, container)
}

func (s *statefulSetImpl) ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig {
	return &NetworkPolicyPeerConfig{PodSelector: s.ToPodSelectorConfig()}
}

func (s *statefulSetImpl) ToPodSelector() IPodSelector {
	return s
}

func (s *statefulSetImpl) AutomountServiceAccountToken() *bool {
	if s.props.AutomountServiceAccountToken == nil {
		return jsii.Bool(false)
	}
	return s.props.AutomountServiceAccountToken
}

func (s *statefulSetImpl) Dns() PodDns {
	return s.dns
}

func (s *statefulSetImpl) DockerRegistryAuth() ISecret {
	return s.props.DockerRegistryAuth
}

func (s *statefulSetImpl) EnableServiceLinks() *bool {
	return s.props.EnableServiceLinks
}

func (s *statefulSetImpl) HostAliases() *[]*HostAlias {
	values := append([]*HostAlias(nil), s.hostAliases...)
	return &values
}

func (s *statefulSetImpl) HostNetwork() *bool {
	if s.props.HostNetwork == nil {
		return jsii.Bool(false)
	}
	return s.props.HostNetwork
}

func (s *statefulSetImpl) Isolate() *bool {
	if s.props.Isolate == nil {
		return jsii.Bool(false)
	}
	return s.props.Isolate
}

func (s *statefulSetImpl) RestartPolicy() RestartPolicy {
	if s.props.RestartPolicy == "" {
		return RestartPolicy_ALWAYS
	}
	return s.props.RestartPolicy
}

func (s *statefulSetImpl) SecurityContext() PodSecurityContext {
	return s.security
}

func (s *statefulSetImpl) ServiceAccount() IServiceAccount {
	return s.props.ServiceAccount
}

func (s *statefulSetImpl) ShareProcessNamespace() *bool {
	if s.props.ShareProcessNamespace == nil {
		return jsii.Bool(false)
	}
	return s.props.ShareProcessNamespace
}

func (s *statefulSetImpl) TerminationGracePeriod() cdk8s.Duration {
	if s.props.TerminationGracePeriod == nil {
		return cdk8s.Duration_Seconds(jsii.Number(30))
	}
	return s.props.TerminationGracePeriod
}

func (s *statefulSetImpl) ToSubjectConfiguration() *SubjectConfiguration {
	if s.props.ServiceAccount == nil && !*s.AutomountServiceAccountToken() {
		panic(stringValue(s.Name()) + " cannot be converted to a role binding subject: You must either assign a service account to it, or use 'automountServiceAccountToken: true'")
	}
	name := jsii.String("default")
	if s.props.ServiceAccount != nil {
		name = s.props.ServiceAccount.ResourceName()
	}
	return &SubjectConfiguration{ApiGroup: jsii.String(""), Kind: jsii.String("ServiceAccount"), Name: name}
}
