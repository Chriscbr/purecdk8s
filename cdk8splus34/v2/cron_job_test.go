package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/cron-job.test.ts#L5
func TestCronJobDefaultChild(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	schedule := cdk8s.NewCron(&cdk8s.CronOptions{
		Minute: jsii.String("*"), Hour: jsii.String("*"), Day: jsii.String("*"),
		Month: jsii.String("*"), WeekDay: jsii.String("*"),
	})
	job := plus.NewCronJob(chart, jsii.String("CronJob"), &plus.CronJobProps{Schedule: schedule})
	child := job.Node().DefaultChild()
	if child == nil {
		t.Fatal("CronJob has no default child")
	}
	if got := stringValue(cdk8s.ApiObject_Of(child).Kind()); got != "CronJob" {
		t.Fatalf("default child kind = %q, want CronJob", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/cron-job.test.ts#L26
func TestCronJobDefaultConfiguration(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewCronJob(chart, jsii.String("CronJob"), &plus.CronJobProps{
		Schedule:   cdk8s.Cron_EveryMinute(),
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}},
	})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "cf4327bde417ac248e2bab2a4a31652bfebf9713a709944a1e79938066fe6d6e")
	spec := mapAt(t, manifests[0], "spec")
	for key, want := range map[string]interface{}{
		"schedule": "* * * * *", "concurrencyPolicy": "Forbid",
		"startingDeadlineSeconds": float64(10), "suspend": false,
		"successfulJobsHistoryLimit": float64(3), "failedJobsHistoryLimit": float64(1),
	} {
		if got := spec[key]; got != want {
			t.Errorf("spec.%s = %#v, want %#v", key, got, want)
		}
	}
	if _, exists := spec["timeZone"]; exists {
		t.Fatalf("spec.timeZone = %#v, want absent", spec["timeZone"])
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/cron-job.test.ts#L51
func TestCronJobCustomConfiguration(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	schedule := cdk8s.Cron_Schedule(&cdk8s.CronOptions{
		Minute: jsii.String("5"), Hour: jsii.String("*"), Day: jsii.String("*"),
		Month: jsii.String("*"), WeekDay: jsii.String("*"),
	})
	plus.NewCronJob(chart, jsii.String("CronJob"), &plus.CronJobProps{
		ActiveDeadline:         cdk8s.Duration_Seconds(jsii.Number(60)),
		BackoffLimit:           jsii.Number(4),
		Schedule:               schedule,
		TimeZone:               jsii.String("America/Los_Angeles"),
		ConcurrencyPolicy:      plus.ConcurrencyPolicy_ALLOW,
		StartingDeadline:       cdk8s.Duration_Seconds(jsii.Number(60)),
		Suspend:                jsii.Bool(false),
		SuccessfulJobsRetained: jsii.Number(3),
		FailedJobsRetained:     jsii.Number(3),
		Containers:             &[]*plus.ContainerProps{{Image: jsii.String("image")}},
	})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "3e2b1b57f105f0e88bfb25ce67c789f44096ecfbc3462996a7e1545889a47c5f")
	spec := mapAt(t, manifests[0], "spec")
	for key, want := range map[string]interface{}{
		"schedule": "5 * * * *", "concurrencyPolicy": "Allow",
		"timeZone": "America/Los_Angeles", "startingDeadlineSeconds": float64(60),
		"suspend": false, "successfulJobsHistoryLimit": float64(3),
		"failedJobsHistoryLimit": float64(3),
	} {
		if got := spec[key]; got != want {
			t.Errorf("spec.%s = %#v, want %#v", key, got, want)
		}
	}
	jobSpec := mapAt(t, spec, "jobTemplate", "spec")
	if got := jobSpec["activeDeadlineSeconds"]; got != float64(60) {
		t.Errorf("activeDeadlineSeconds = %#v, want 60", got)
	}
	if got := jobSpec["backoffLimit"]; got != float64(4) {
		t.Errorf("backoffLimit = %#v, want 4", got)
	}
	containers := sliceAt(t, jobSpec, "template", "spec", "containers")
	if got := mapAt(t, containers[0])["image"]; got != "image" {
		t.Errorf("container image = %#v, want image", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/cron-job.test.ts#L96
func TestCronJobCanBeIsolated(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewCronJob(chart, jsii.String("CronJob"), &plus.CronJobProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("foobar")}},
		Schedule:   cdk8s.Cron_Daily(), Isolate: jsii.Bool(true),
	})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "01059d57429891a0f6b846fabf14135a45b958efcc15ae064a03edacb834d64a")
	policy := mapAt(t, manifests[1], "spec")
	if labels := mapAt(t, policy, "podSelector", "matchLabels"); len(labels) == 0 {
		t.Fatal("isolating CronJob produced no pod selector labels")
	}
	requireDeepEqual(t, policy["policyTypes"], []interface{}{"Egress", "Ingress"})
}
