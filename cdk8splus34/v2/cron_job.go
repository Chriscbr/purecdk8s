package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
)

// CronJobProps configures a CronJob. Schedule is a standard cron expression.
type CronJobProps struct {
	Metadata       *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Schedule       *string                  `field:"required" json:"schedule" yaml:"schedule"`
	Containers     *[]*ContainerProps       `field:"optional" json:"containers" yaml:"containers"`
	InitContainers *[]*ContainerProps       `field:"optional" json:"initContainers" yaml:"initContainers"`
	Volumes        *[]Volume                `field:"optional" json:"volumes" yaml:"volumes"`
	BackoffLimit   *float64                 `field:"optional" json:"backoffLimit" yaml:"backoffLimit"`
}

// CronJob is a native Kubernetes CronJob construct.
type CronJob interface {
	Resource
	Containers() *[]Container
	AddContainer(props *ContainerProps) Container
}
type cronJobImpl struct {
	resourceBase
	podState
	props *CronJobProps
}

func NewCronJob(scope constructs.Construct, id *string, props *CronJobProps) CronJob {
	if props == nil || props.Schedule == nil || *props.Schedule == "" {
		panic("schedule is required")
	}
	base := &PodProps{Metadata: props.Metadata, Containers: props.Containers, InitContainers: props.InitContainers, Volumes: props.Volumes}
	result := &cronJobImpl{podState: newPodState(base), props: props}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "batch/v1", "CronJob", "cronjobs", props.Metadata, manifest)
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} {
		job := map[string]interface{}{"template": map[string]interface{}{"spec": result.podState.manifest(RestartPolicy_NEVER)}}
		if props.BackoffLimit != nil {
			job["backoffLimit"] = props.BackoffLimit
		}
		return map[string]interface{}{"schedule": props.Schedule, "jobTemplate": map[string]interface{}{"spec": job}}
	}})
	return result
}
func NewCronJob_Override(job CronJob, scope constructs.Construct, id *string, props *CronJobProps) {
	panic("native cdk8splus34 overrides are not implemented")
}
func CronJob_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }
func (j *cronJobImpl) Containers() *[]Container {
	values := append([]Container(nil), j.containers...)
	return &values
}
func (j *cronJobImpl) AddContainer(props *ContainerProps) Container { return j.addContainer(props) }
