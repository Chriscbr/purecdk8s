package cdk8splus34

import (
	"fmt"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Properties for `Deployment`.
type DeploymentProps struct {
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
	// Zero means the pod will be considered available as soon as it is ready. See: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#min-ready-seconds
	//
	// Default: Duration.seconds(0)
	MinReady cdk8s.Duration `field:"optional" json:"minReady" yaml:"minReady"`
	// The pod metadata of this workload.
	PodMetadata *cdk8s.ApiObjectMetadata `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	// The maximum duration for a deployment to make progress before it is considered to be failed.
	//
	// The deployment controller will continue to process failed deployments and a condition with a ProgressDeadlineExceeded reason will be surfaced in the deployment status.
	//
	// Note that progress will not be estimated during the time a deployment is paused. See: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#progress-deadline-seconds
	//
	// Default: Duration.seconds(600)
	ProgressDeadline cdk8s.Duration `field:"optional" json:"progressDeadline" yaml:"progressDeadline"`
	// Number of desired pods. Default: 2.
	Replicas *float64 `field:"optional" json:"replicas" yaml:"replicas"`
	// Restart policy for all containers within the pod. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy
	//
	// Default: RestartPolicy.ALWAYS
	RestartPolicy RestartPolicy `field:"optional" json:"restartPolicy" yaml:"restartPolicy"`
	// Specify how many old ReplicaSets for this Deployment you want to retain.
	//
	// The rest will be garbage-collected in the background. By default, it is 10. Default: 10.
	RevisionHistoryLimit *float64 `field:"optional" json:"revisionHistoryLimit" yaml:"revisionHistoryLimit"`
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
	// Specifies the strategy used to replace old Pods by new ones. Default: - RollingUpdate with maxSurge and maxUnavailable set to 25%.
	Strategy DeploymentStrategy `field:"optional" json:"strategy" yaml:"strategy"`
	// Grace period until the pod is terminated. Default: Duration.seconds(30)
	TerminationGracePeriod cdk8s.Duration `field:"optional" json:"terminationGracePeriod" yaml:"terminationGracePeriod"`
	// List of volumes that can be mounted by containers belonging to the pod.
	//
	// You can also add volumes later using `podSpec.addVolume()` See: https://kubernetes.io/docs/concepts/storage/volumes
	//
	// Default: - No volumes.
	Volumes *[]Volume `field:"optional" json:"volumes" yaml:"volumes"`
}

// Options for `Deployment.exposeViaService`.
type DeploymentExposeViaServiceOptions struct {
	// The ports that the service should bind to. Default: - extracted from the deployment.
	Ports *[]*ServicePort `field:"optional" json:"ports" yaml:"ports"`
	// The type of the exposed service. Default: - ClusterIP.
	ServiceType ServiceType `field:"optional" json:"serviceType" yaml:"serviceType"`
	// The name of the service to expose.
	//
	// If you'd like to expose the deployment multiple times, you must explicitly set a name starting from the second expose call. Default: - auto generated.
	Name *string `field:"optional" json:"name" yaml:"name"`
}

// Options for exposing a deployment via an ingress.
type ExposeDeploymentViaIngressOptions struct {
	// The ports that the service should bind to. Default: - extracted from the deployment.
	Ports *[]*ServicePort `field:"optional" json:"ports" yaml:"ports"`
	// The type of the exposed service. Default: - ClusterIP.
	ServiceType ServiceType `field:"optional" json:"serviceType" yaml:"serviceType"`
	// The name of the service to expose.
	//
	// If you'd like to expose the deployment multiple times, you must explicitly set a name starting from the second expose call. Default: - auto generated.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// The ingress to add rules to. Default: - An ingress will be automatically created.
	Ingress Ingress `field:"optional" json:"ingress" yaml:"ingress"`
	// The type of the path. Default: HttpIngressPathType.PREFIX
	PathType HttpIngressPathType `field:"optional" json:"pathType" yaml:"pathType"`
}

// A Deployment provides declarative updates for Pods and ReplicaSets.
//
// You describe a desired state in a Deployment, and the Deployment Controller changes the actual state to the desired state at a controlled rate. You can define Deployments to create new ReplicaSets, or to remove existing Deployments and adopt all their resources with new Deployments.
//
// > Note: Do not manage ReplicaSets owned by a Deployment. Consider opening an issue in the main Kubernetes repository if your use case is not covered below.
//
// # Use Case
//
// The following are typical use cases for Deployments:
//
//   - Create a Deployment to rollout a ReplicaSet. The ReplicaSet creates Pods in the background. Check the status of the rollout to see if it succeeds or not.
//   - Declare the new state of the Pods by updating the PodTemplateSpec of the Deployment. A new ReplicaSet is created and the Deployment manages moving the Pods from the old ReplicaSet to the new one at a controlled rate. Each new ReplicaSet updates the revision of the Deployment.
//   - Rollback to an earlier Deployment revision if the current state of the Deployment is not stable. Each rollback updates the revision of the Deployment.
//   - Scale up the Deployment to facilitate more load.
//   - Pause the Deployment to apply multiple fixes to its PodTemplateSpec and then resume it to start a new rollout.
//   - Use the status of the Deployment as an indicator that a rollout has stuck.
//   - Clean up older ReplicaSets that you don't need anymore.
type Deployment interface {
	Workload
	IScalable
	// Number of desired pods.
	Replicas() *float64
	// Minimum duration for which a newly created pod should be ready without any of its container crashing, for it to be considered available.
	MinReady() cdk8s.Duration
	// The maximum duration for a deployment to make progress before it is considered to be failed.
	ProgressDeadline() cdk8s.Duration
	// Number of desired replicasets history. Default: 10.
	RevisionHistoryLimit() *float64
	Strategy() DeploymentStrategy
	// Expose a deployment via a service.
	//
	// This is equivalent to running `kubectl expose deployment <deployment-name>`.
	ExposeViaService(options *DeploymentExposeViaServiceOptions) Service
	// Expose a deployment via an ingress.
	//
	// This will first expose the deployment with a service, and then expose the service via an ingress.
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

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
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
