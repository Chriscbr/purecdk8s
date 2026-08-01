package cdk8splus34

import (
	"sort"

	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
)

// RestartPolicy controls what Kubernetes does when a container exits.
type RestartPolicy string

const (
	RestartPolicy_ALWAYS     RestartPolicy = "Always"
	RestartPolicy_NEVER      RestartPolicy = "Never"
	RestartPolicy_ON_FAILURE RestartPolicy = "OnFailure"
)

// PodProps contains the common pod fields used by native workload constructs.
type PodProps struct {
	Metadata                     *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Containers                   *[]*ContainerProps       `field:"optional" json:"containers" yaml:"containers"`
	InitContainers               *[]*ContainerProps       `field:"optional" json:"initContainers" yaml:"initContainers"`
	Volumes                      *[]Volume                `field:"optional" json:"volumes" yaml:"volumes"`
	AutomountServiceAccountToken *bool                    `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	EnableServiceLinks           *bool                    `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	HostNetwork                  *bool                    `field:"optional" json:"hostNetwork" yaml:"hostNetwork"`
	RestartPolicy                RestartPolicy            `field:"optional" json:"restartPolicy" yaml:"restartPolicy"`
	ServiceAccount               IServiceAccount          `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ShareProcessNamespace        *bool                    `field:"optional" json:"shareProcessNamespace" yaml:"shareProcessNamespace"`
}

type podState struct {
	containers     []Container
	initContainers []Container
	volumes        []Volume
	props          *PodProps
	selector       map[string]*string
}

func newPodState(props *PodProps) podState {
	if props == nil {
		props = &PodProps{}
	}
	state := podState{props: props, selector: map[string]*string{}}
	if props.Containers != nil {
		for _, value := range *props.Containers {
			if value != nil {
				state.containers = append(state.containers, NewContainer(value))
			}
		}
	}
	if props.InitContainers != nil {
		for _, value := range *props.InitContainers {
			if value != nil {
				state.initContainers = append(state.initContainers, NewContainer(value))
			}
		}
	}
	if props.Volumes != nil {
		state.volumes = append(state.volumes, (*props.Volumes)...)
	}
	return state
}

func (p *podState) addContainer(props *ContainerProps) Container {
	container := NewContainer(props)
	p.containers = append(p.containers, container)
	return container
}

func (p *podState) addInitContainer(props *ContainerProps) Container {
	container := NewContainer(props)
	p.initContainers = append(p.initContainers, container)
	return container
}

func (p *podState) addVolume(volume Volume) { p.volumes = append(p.volumes, volume) }

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
	for _, volume := range p.volumes {
		volumes[stringValue(volume.Name())] = volume
	}
	for _, container := range append(append([]Container{}, p.containers...), p.initContainers...) {
		for _, mount := range *container.Mounts() {
			volumes[stringValue(mount.Volume.Name())] = mount.Volume
		}
	}
	spec := map[string]interface{}{
		"automountServiceAccountToken": false,
		"containers":                   containerValues(p.containers),
		"dnsPolicy":                    "ClusterFirst",
		"hostNetwork":                  false,
		"restartPolicy":                restartPolicy,
		"securityContext": map[string]interface{}{
			"fsGroupChangePolicy": "Always",
			"runAsNonRoot":        true,
		},
		"setHostnameAsFQDN":             false,
		"shareProcessNamespace":         false,
		"terminationGracePeriodSeconds": 30,
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
	if len(p.initContainers) != 0 {
		spec["initContainers"] = containerValues(p.initContainers)
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
	Resource
	Containers() *[]Container
	InitContainers() *[]Container
	Volumes() *[]Volume
	AddContainer(props *ContainerProps) Container
	AddInitContainer(props *ContainerProps) Container
	AddVolume(volume Volume)
	ToPodSelectorConfig() *PodSelectorConfig
}

type podImpl struct {
	resourceBase
	podState
}

func NewPod(scope constructs.Construct, id *string, props *PodProps) Pod {
	if props == nil {
		props = &PodProps{}
	}
	result := &podImpl{podState: newPodState(props)}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "v1", "Pod", "pods", props.Metadata, manifest)
	result.selector[podAddressLabel] = cdk8s.Names_ToLabelValue(result, nil)
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} {
		policy := props.RestartPolicy
		if policy == "" {
			policy = RestartPolicy_ALWAYS
		}
		return result.podState.manifest(policy)
	}})
	return result
}

func NewPod_Override(pod Pod, scope constructs.Construct, id *string, props *PodProps) {
	panic("native cdk8splus34 overrides are not implemented")
}

func Pod_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }
func (p *podImpl) Containers() *[]Container {
	values := append([]Container(nil), p.containers...)
	return &values
}
func (p *podImpl) InitContainers() *[]Container {
	values := append([]Container(nil), p.initContainers...)
	return &values
}
func (p *podImpl) Volumes() *[]Volume                               { values := append([]Volume(nil), p.volumes...); return &values }
func (p *podImpl) AddContainer(props *ContainerProps) Container     { return p.addContainer(props) }
func (p *podImpl) AddInitContainer(props *ContainerProps) Container { return p.addInitContainer(props) }
func (p *podImpl) AddVolume(volume Volume)                          { p.addVolume(volume) }
func (p *podImpl) ToPodSelectorConfig() *PodSelectorConfig {
	labels := map[string]*string{}
	for key, value := range p.selector {
		labels[key] = value
	}
	return &PodSelectorConfig{LabelSelector: &LabelSelector{Labels: &labels}}
}
