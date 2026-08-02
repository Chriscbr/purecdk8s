package cdk8splus34

import (
	"fmt"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// StatefulSetProps configures a StatefulSet workload.
type StatefulSetProps struct {
	Metadata                     *cdk8s.ApiObjectMetadata               `field:"optional" json:"metadata" yaml:"metadata"`
	AutomountServiceAccountToken *bool                                  `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	Containers                   *[]*ContainerProps                     `field:"optional" json:"containers" yaml:"containers"`
	Dns                          *PodDnsProps                           `field:"optional" json:"dns" yaml:"dns"`
	DockerRegistryAuth           ISecret                                `field:"optional" json:"dockerRegistryAuth" yaml:"dockerRegistryAuth"`
	EnableServiceLinks           *bool                                  `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	HostAliases                  *[]*HostAlias                          `field:"optional" json:"hostAliases" yaml:"hostAliases"`
	HostNetwork                  *bool                                  `field:"optional" json:"hostNetwork" yaml:"hostNetwork"`
	InitContainers               *[]*ContainerProps                     `field:"optional" json:"initContainers" yaml:"initContainers"`
	Isolate                      *bool                                  `field:"optional" json:"isolate" yaml:"isolate"`
	MinReady                     cdk8s.Duration                         `field:"optional" json:"minReady" yaml:"minReady"`
	PodMetadata                  *cdk8s.ApiObjectMetadata               `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	PodManagementPolicy          PodManagementPolicy                    `field:"optional" json:"podManagementPolicy" yaml:"podManagementPolicy"`
	Replicas                     *float64                               `field:"optional" json:"replicas" yaml:"replicas"`
	RestartPolicy                RestartPolicy                          `field:"optional" json:"restartPolicy" yaml:"restartPolicy"`
	SecurityContext              *PodSecurityContextProps               `field:"optional" json:"securityContext" yaml:"securityContext"`
	Select                       *bool                                  `field:"optional" json:"select" yaml:"select"`
	Service                      Service                                `field:"optional" json:"service" yaml:"service"`
	ServiceAccount               IServiceAccount                        `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ShareProcessNamespace        *bool                                  `field:"optional" json:"shareProcessNamespace" yaml:"shareProcessNamespace"`
	Spread                       *bool                                  `field:"optional" json:"spread" yaml:"spread"`
	Strategy                     StatefulSetUpdateStrategy              `field:"optional" json:"strategy" yaml:"strategy"`
	TerminationGracePeriod       cdk8s.Duration                         `field:"optional" json:"terminationGracePeriod" yaml:"terminationGracePeriod"`
	VolumeClaimTemplates         *[]*PersistentVolumeClaimTemplateProps `field:"optional" json:"volumeClaimTemplates" yaml:"volumeClaimTemplates"`
	Volumes                      *[]Volume                              `field:"optional" json:"volumes" yaml:"volumes"`
}

// StatefulSet manages a set of pods with stable identities.
type StatefulSet interface {
	Workload
	IScalable
	Replicas() *float64
	Service() Service
	PodManagementPolicy() PodManagementPolicy
	MinReady() cdk8s.Duration
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

// NewStatefulSet creates a native StatefulSet and its default headless Service.
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
		"podManagementPolicy": string(s.PodManagementPolicy()),
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

// PodManagementPolicy controls how StatefulSet pods are created and removed.
type PodManagementPolicy string

const (
	PodManagementPolicy_ORDERED_READY PodManagementPolicy = "OrderedReady"
	PodManagementPolicy_PARALLEL      PodManagementPolicy = "Parallel"
)

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
