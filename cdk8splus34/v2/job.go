package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
)

// JobProps configures a batch Job.
type JobProps struct {
	Metadata       *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Containers     *[]*ContainerProps       `field:"optional" json:"containers" yaml:"containers"`
	InitContainers *[]*ContainerProps       `field:"optional" json:"initContainers" yaml:"initContainers"`
	Volumes        *[]Volume                `field:"optional" json:"volumes" yaml:"volumes"`
	BackoffLimit   *float64                 `field:"optional" json:"backoffLimit" yaml:"backoffLimit"`
	Completions    *float64                 `field:"optional" json:"completions" yaml:"completions"`
	Parallelism    *float64                 `field:"optional" json:"parallelism" yaml:"parallelism"`
}

// Job is a native Kubernetes Job construct.
type Job interface {
	Resource
	Containers() *[]Container
	AddContainer(props *ContainerProps) Container
}

type jobImpl struct {
	resourceBase
	podState
	props *JobProps
}

func NewJob(scope constructs.Construct, id *string, props *JobProps) Job {
	if props == nil {
		props = &JobProps{}
	}
	base := &PodProps{Metadata: props.Metadata, Containers: props.Containers, InitContainers: props.InitContainers, Volumes: props.Volumes}
	result := &jobImpl{podState: newPodState(base), props: props}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "batch/v1", "Job", "jobs", props.Metadata, manifest)
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} {
		spec := map[string]interface{}{"template": map[string]interface{}{"spec": result.podState.manifest(RestartPolicy_NEVER)}}
		if props.BackoffLimit != nil {
			spec["backoffLimit"] = props.BackoffLimit
		}
		if props.Completions != nil {
			spec["completions"] = props.Completions
		}
		if props.Parallelism != nil {
			spec["parallelism"] = props.Parallelism
		}
		return spec
	}})
	return result
}
func NewJob_Override(job Job, scope constructs.Construct, id *string, props *JobProps) {
	panic("native cdk8splus34 overrides are not implemented")
}
func Job_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }
func (j *jobImpl) Containers() *[]Container {
	values := append([]Container(nil), j.containers...)
	return &values
}
func (j *jobImpl) AddContainer(props *ContainerProps) Container { return j.addContainer(props) }
