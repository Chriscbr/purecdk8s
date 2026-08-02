package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// JobProps configures a batch Job and its pod template.
type JobProps struct {
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
	PodMetadata                  *cdk8s.ApiObjectMetadata `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	RestartPolicy                RestartPolicy            `field:"optional" json:"restartPolicy" yaml:"restartPolicy"`
	SecurityContext              *PodSecurityContextProps `field:"optional" json:"securityContext" yaml:"securityContext"`
	Select                       *bool                    `field:"optional" json:"select" yaml:"select"`
	ServiceAccount               IServiceAccount          `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ShareProcessNamespace        *bool                    `field:"optional" json:"shareProcessNamespace" yaml:"shareProcessNamespace"`
	Spread                       *bool                    `field:"optional" json:"spread" yaml:"spread"`
	TerminationGracePeriod       cdk8s.Duration           `field:"optional" json:"terminationGracePeriod" yaml:"terminationGracePeriod"`
	Volumes                      *[]Volume                `field:"optional" json:"volumes" yaml:"volumes"`
	ActiveDeadline               cdk8s.Duration           `field:"optional" json:"activeDeadline" yaml:"activeDeadline"`
	BackoffLimit                 *float64                 `field:"optional" json:"backoffLimit" yaml:"backoffLimit"`
	TtlAfterFinished             cdk8s.Duration           `field:"optional" json:"ttlAfterFinished" yaml:"ttlAfterFinished"`
}

// Job is a batch workload that runs its pod template to completion.
type Job interface {
	Workload
	ActiveDeadline() cdk8s.Duration
	BackoffLimit() *float64
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

func NewJob_Override(job Job, scope constructs.Construct, id *string, props *JobProps) {
	applyOverride(job, NewJob(scope, id, props), "Job")
}

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
	for key, value := range j.selector {
		result.AddLabel(jsii.String(key), value)
	}
	return result
}

func (j *jobImpl) ToPodSelectorConfig() *PodSelectorConfig {
	labels := map[string]*string{}
	for key, value := range j.selector {
		labels[key] = value
	}
	return &PodSelectorConfig{LabelSelector: newLabelSelectorFromRequirements(j.matchExpressions, &labels)}
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
	spec := j.podState.manifest(RestartPolicy_NEVER)
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
	return &PodProps{Metadata: p.Metadata, AutomountServiceAccountToken: p.AutomountServiceAccountToken, Containers: p.Containers, Dns: p.Dns, DockerRegistryAuth: p.DockerRegistryAuth, EnableServiceLinks: p.EnableServiceLinks, HostAliases: p.HostAliases, HostNetwork: p.HostNetwork, InitContainers: p.InitContainers, Isolate: p.Isolate, RestartPolicy: RestartPolicy_NEVER, SecurityContext: p.SecurityContext, ServiceAccount: p.ServiceAccount, ShareProcessNamespace: p.ShareProcessNamespace, TerminationGracePeriod: p.TerminationGracePeriod, Volumes: p.Volumes}
}
