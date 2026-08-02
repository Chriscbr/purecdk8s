package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

type ConcurrencyPolicy string

const (
	ConcurrencyPolicy_ALLOW   ConcurrencyPolicy = "ALLOW"
	ConcurrencyPolicy_FORBID  ConcurrencyPolicy = "FORBID"
	ConcurrencyPolicy_REPLACE ConcurrencyPolicy = "REPLACE"
)

func concurrencyPolicyManifestValue(value ConcurrencyPolicy) string {
	switch value {
	case ConcurrencyPolicy_ALLOW:
		return "Allow"
	case ConcurrencyPolicy_FORBID:
		return "Forbid"
	case ConcurrencyPolicy_REPLACE:
		return "Replace"
	default:
		return string(value)
	}
}

// CronJobProps configures a CronJob and the Job pod template it creates.
type CronJobProps struct {
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
	Schedule                     cdk8s.Cron               `field:"required" json:"schedule" yaml:"schedule"`
	ConcurrencyPolicy            ConcurrencyPolicy        `field:"optional" json:"concurrencyPolicy" yaml:"concurrencyPolicy"`
	FailedJobsRetained           *float64                 `field:"optional" json:"failedJobsRetained" yaml:"failedJobsRetained"`
	StartingDeadline             cdk8s.Duration           `field:"optional" json:"startingDeadline" yaml:"startingDeadline"`
	SuccessfulJobsRetained       *float64                 `field:"optional" json:"successfulJobsRetained" yaml:"successfulJobsRetained"`
	Suspend                      *bool                    `field:"optional" json:"suspend" yaml:"suspend"`
	TimeZone                     *string                  `field:"optional" json:"timeZone" yaml:"timeZone"`
}

type CronJob interface {
	Workload
	ConcurrencyPolicy() *string
	FailedJobsRetained() *float64
	Schedule() cdk8s.Cron
	StartingDeadline() cdk8s.Duration
	SuccessfulJobsRetained() *float64
	Suspend() *bool
	TimeZone() *string
}

type cronJobImpl struct {
	jobImpl
	schedule               cdk8s.Cron
	concurrencyPolicy      ConcurrencyPolicy
	failedJobsRetained     *float64
	startingDeadline       cdk8s.Duration
	successfulJobsRetained *float64
	suspend                *bool
	timeZone               *string
}

func NewCronJob(scope constructs.Construct, id *string, props *CronJobProps) CronJob {
	if props == nil || props.Schedule == nil {
		panic("schedule is required")
	}
	if props.StartingDeadline != nil && *props.StartingDeadline.ToSeconds(nil) < 10 {
		panic("The 'startingDeadline' property cannot be less than 10 seconds")
	}
	if props.TtlAfterFinished != nil && (props.SuccessfulJobsRetained != nil || props.FailedJobsRetained != nil) {
		panic("The 'ttlAfterFinished' property cannot be set if 'successfulJobsRetained' property or 'failedJobsRetained' property is set")
	}
	result := &cronJobImpl{jobImpl: jobImpl{podState: newPodState(cronJobPodProps(props)), podMetadata: props.PodMetadata, selector: map[string]*string{}, activeDeadline: props.ActiveDeadline, backoffLimit: props.BackoffLimit, ttlAfterFinished: props.TtlAfterFinished}, schedule: props.Schedule, concurrencyPolicy: props.ConcurrencyPolicy, failedJobsRetained: props.FailedJobsRetained, startingDeadline: props.StartingDeadline, successfulJobsRetained: props.SuccessfulJobsRetained, suspend: props.Suspend, timeZone: props.TimeZone}
	if result.concurrencyPolicy == "" {
		result.concurrencyPolicy = ConcurrencyPolicy_FORBID
	}
	if result.failedJobsRetained == nil {
		result.failedJobsRetained = jsii.Number(1)
	}
	if result.successfulJobsRetained == nil {
		result.successfulJobsRetained = jsii.Number(3)
	}
	if result.startingDeadline == nil {
		result.startingDeadline = cdk8s.Duration_Seconds(jsii.Number(10))
	}
	if result.suspend == nil {
		result.suspend = jsii.Bool(false)
	}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "batch/v1", "CronJob", "cronjobs", props.Metadata, manifest)
	// CronJobs do not automatically select their generated Pods.
	result.scheduling = NewWorkloadScheduling(result)
	if props.Spread != nil && *props.Spread {
		result.scheduling.Spread(&WorkloadSchedulingSpreadOptions{Topology: Topology_HOSTNAME()})
		result.scheduling.Spread(&WorkloadSchedulingSpreadOptions{Topology: Topology_ZONE()})
	}
	result.connections = NewPodConnections(result)
	if *result.Isolate() {
		result.connections.Isolate()
	}
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toCronManifest() }})
	return result
}

func NewCronJob_Override(job CronJob, scope constructs.Construct, id *string, props *CronJobProps) {
	applyOverride(job, NewCronJob(scope, id, props), "CronJob")
}

func CronJob_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (c *cronJobImpl) Schedule() cdk8s.Cron {
	return c.schedule
}

func (c *cronJobImpl) ConcurrencyPolicy() *string {
	return jsii.String(concurrencyPolicyManifestValue(c.concurrencyPolicy))
}

func (c *cronJobImpl) FailedJobsRetained() *float64 {
	return c.failedJobsRetained
}

func (c *cronJobImpl) StartingDeadline() cdk8s.Duration {
	return c.startingDeadline
}

func (c *cronJobImpl) SuccessfulJobsRetained() *float64 {
	return c.successfulJobsRetained
}

func (c *cronJobImpl) Suspend() *bool {
	return c.suspend
}

func (c *cronJobImpl) TimeZone() *string {
	return c.timeZone
}

func (c *cronJobImpl) toCronManifest() map[string]interface{} {
	podSpec := c.podState.manifest(RestartPolicy_NEVER)
	for key, value := range c.scheduling.toManifest() {
		podSpec[key] = value
	}
	job := map[string]interface{}{"template": map[string]interface{}{"metadata": c.PodMetadata().ToJson(), "spec": podSpec}}
	if c.activeDeadline != nil {
		job["activeDeadlineSeconds"] = c.activeDeadline.ToSeconds(nil)
	}
	if c.backoffLimit != nil {
		job["backoffLimit"] = c.backoffLimit
	}
	if c.ttlAfterFinished != nil {
		job["ttlSecondsAfterFinished"] = c.ttlAfterFinished.ToSeconds(nil)
	}
	result := map[string]interface{}{"concurrencyPolicy": concurrencyPolicyManifestValue(c.concurrencyPolicy), "failedJobsHistoryLimit": c.failedJobsRetained, "jobTemplate": map[string]interface{}{"spec": job}, "schedule": c.schedule.ExpressionString(), "startingDeadlineSeconds": c.startingDeadline.ToSeconds(nil), "successfulJobsHistoryLimit": c.successfulJobsRetained, "suspend": c.suspend}
	if c.timeZone != nil {
		result["timeZone"] = c.timeZone
	}
	return result
}

func cronJobPodProps(p *CronJobProps) *PodProps {
	return &PodProps{Metadata: p.Metadata, AutomountServiceAccountToken: p.AutomountServiceAccountToken, Containers: p.Containers, Dns: p.Dns, DockerRegistryAuth: p.DockerRegistryAuth, EnableServiceLinks: p.EnableServiceLinks, HostAliases: p.HostAliases, HostNetwork: p.HostNetwork, InitContainers: p.InitContainers, Isolate: p.Isolate, RestartPolicy: RestartPolicy_NEVER, SecurityContext: p.SecurityContext, ServiceAccount: p.ServiceAccount, ShareProcessNamespace: p.ShareProcessNamespace, TerminationGracePeriod: p.TerminationGracePeriod, Volumes: p.Volumes}
}
